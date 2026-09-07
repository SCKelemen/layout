package layout

import (
	"math"
	"testing"
)

// Round-1 correctness regression tests for CSS Grid.
//
// Each test reproduces a specific bug against the relevant section of CSS Grid
// Layout Module Level 1 (https://www.w3.org/TR/css-grid-1/) or CSS Box
// Alignment Module Level 3 (https://www.w3.org/TR/css-align-3/).

func gridRound1Approx(t *testing.T, what string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Errorf("%s: got %.2f, want %.2f", what, got, want)
	}
}

// TestGridRound1ExplicitZeroGapOverridesShorthand checks that an explicit
// Px(0) row-gap or column-gap wins over the gap shorthand. Previously the
// fallback was keyed on the resolved value being 0, so Px(0) could never
// override GridGap.
//
// https://www.w3.org/TR/css-align-3/#gap-shorthand
func TestGridRound1ExplicitZeroGapOverridesShorthand(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridGap:             Px(10),
			GridRowGap:          Px(0), // explicit zero: no row gap
			// GridColumnGap unset: falls back to GridGap (10px)
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridColumnStart: 0}},
			{Style: Style{GridRowStart: 1, GridColumnStart: 1}},
		},
	}
	ctx := NewLayoutContext(800, 600, 16)
	size := LayoutGrid(root, Loose(Unbounded, Unbounded), ctx)

	gridRound1Approx(t, "second item Y (no row gap)", root.Children[1].Rect.Y, 50)
	gridRound1Approx(t, "second item X (column gap from shorthand)", root.Children[1].Rect.X, 110)
	gridRound1Approx(t, "container width", size.Width, 210)
	gridRound1Approx(t, "container height", size.Height, 100)
}

// TestGridRound1ItemFontSizeFromContext checks that an item's em margins
// resolve against the layout context's root font size when the item has no
// TextStyle, as getCurrentFontSize does everywhere else.
func TestGridRound1ItemFontSizeFromContext(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
		},
		Children: []*Node{
			{Style: Style{Margin: Uniform(Em(1))}},
		},
	}
	// Root font size 20px: Em(1) margins are 20px, not the 16px default.
	ctx := NewLayoutContext(800, 600, 20)
	LayoutGrid(root, Loose(Unbounded, Unbounded), ctx)

	gridRound1Approx(t, "item X", root.Children[0].Rect.X, 20)
	gridRound1Approx(t, "item Y", root.Children[0].Rect.Y, 20)
	gridRound1Approx(t, "item width", root.Children[0].Rect.Width, 60)
	gridRound1Approx(t, "item height", root.Children[0].Rect.Height, 60)
}

// gridRound1IntrinsicGrid returns a 200px-wide single-cell grid whose item has
// a single 50px child, so the item's min-content and max-content widths are
// both 50px.
func gridRound1IntrinsicGrid(item *Node) *Node {
	item.Children = []*Node{{Style: Style{Width: Px(50), Height: Px(20)}}}
	return &Node{
		Style: Style{
			Display:             DisplayGrid,
			Width:               Px(200),
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(200))},
		},
		Children: []*Node{item},
	}
}

// TestGridRound1ItemWidthSizingMinContent checks that a grid item with
// WidthSizing: IntrinsicSizeMinContent is sized to its min-content width and
// is not stretched to the cell (css-sizing-3 §4, css-align-3 §6.2).
//
// https://www.w3.org/TR/css-sizing-3/#sizing-values
func TestGridRound1ItemWidthSizingMinContent(t *testing.T) {
	item := &Node{Style: Style{WidthSizing: IntrinsicSizeMinContent}}
	root := gridRound1IntrinsicGrid(item)
	LayoutGrid(root, Loose(200, Unbounded), NewLayoutContext(800, 600, 16))

	gridRound1Approx(t, "item width", item.Rect.Width, 50)
	gridRound1Approx(t, "item X", item.Rect.X, 0)
	// align-items: stretch still applies in the row axis.
	gridRound1Approx(t, "item height", item.Rect.Height, 100)
}

// TestGridRound1ItemMinContentWidthSentinel checks that the deprecated
// MinContentWidth() helper (Width = SizeMinContent sentinel) is honored the
// same way as WidthSizing, instead of being read as auto and stretched.
func TestGridRound1ItemMinContentWidthSentinel(t *testing.T) {
	item := MinContentWidth(&Node{})
	root := gridRound1IntrinsicGrid(item)
	LayoutGrid(root, Loose(200, Unbounded), NewLayoutContext(800, 600, 16))

	gridRound1Approx(t, "item width", item.Rect.Width, 50)
	gridRound1Approx(t, "item X", item.Rect.X, 0)
}

// TestGridRound1ItemIntrinsicWidthAlignedEnd checks that an intrinsically
// sized item still follows justify-self.
func TestGridRound1ItemIntrinsicWidthAlignedEnd(t *testing.T) {
	item := &Node{Style: Style{WidthSizing: IntrinsicSizeMaxContent, JustifySelf: JustifyItemsEnd}}
	root := gridRound1IntrinsicGrid(item)
	LayoutGrid(root, Loose(200, Unbounded), NewLayoutContext(800, 600, 16))

	gridRound1Approx(t, "item width", item.Rect.Width, 50)
	gridRound1Approx(t, "item X", item.Rect.X, 150)
}

// TestGridRound1MarginsInTrackContributions checks that an item's margins
// count toward the intrinsic size of the tracks it occupies (§12.5 uses the
// item's outer, margin-box size). A 50px item with 10px margins in an auto
// column makes that column 70px, and the auto row 30px.
//
// https://www.w3.org/TR/css-grid-1/#algo-content
func TestGridRound1MarginsInTrackContributions(t *testing.T) {
	item := &Node{Style: Style{Width: Px(50), Height: Px(10), Margin: Uniform(Px(10))}}
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), AutoTrack()},
			GridTemplateRows:    []GridTrack{AutoTrack()},
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridColumnStart: 0}},
			item,
		},
	}
	size := LayoutGrid(root, Loose(Unbounded, Unbounded), NewLayoutContext(800, 600, 16))

	gridRound1Approx(t, "container width", size.Width, 120)
	gridRound1Approx(t, "container height", size.Height, 30)
	gridRound1Approx(t, "item X", item.Rect.X, 60)
	gridRound1Approx(t, "item Y", item.Rect.Y, 10)
	gridRound1Approx(t, "item width", item.Rect.Width, 50)
	gridRound1Approx(t, "item height", item.Rect.Height, 10)
}

// TestGridRound1MarginsLogicalInVerticalModes checks that margins follow the
// grid's logical axes in vertical writing modes: top/bottom margins lie along
// the column axis (physical Y) and left/right along the row axis, with the
// row-start margin being the right one in vertical-rl.
//
// https://www.w3.org/TR/css-writing-modes-3/#logical-to-physical
func TestGridRound1MarginsLogicalInVerticalModes(t *testing.T) {
	build := func(mode WritingMode) (*Node, *Node) {
		item := &Node{Style: Style{
			Width:     Px(20), // row axis (physical width) in vertical modes
			AlignSelf: AlignItemsStart,
			Margin:    Spacing{Top: Px(5), Bottom: Px(15), Left: Px(10), Right: Px(40)},
		}}
		root := &Node{
			Style: Style{
				Display:             DisplayGrid,
				WritingMode:         mode,
				GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
				GridTemplateRows:    []GridTrack{FixedTrack(Px(100))},
			},
			Children: []*Node{item},
		}
		return root, item
	}

	root, item := build(WritingModeVerticalLR)
	LayoutGrid(root, Loose(Unbounded, Unbounded), NewLayoutContext(800, 600, 16))
	// Column axis is Y: top margin 5, stretched height 100-5-15.
	gridRound1Approx(t, "vertical-lr item Y", item.Rect.Y, 5)
	gridRound1Approx(t, "vertical-lr item height", item.Rect.Height, 80)
	// Row axis is X, rows start at the left: left margin 10.
	gridRound1Approx(t, "vertical-lr item X", item.Rect.X, 10)
	gridRound1Approx(t, "vertical-lr item width", item.Rect.Width, 20)

	root, item = build(WritingModeVerticalRL)
	LayoutGrid(root, Loose(Unbounded, Unbounded), NewLayoutContext(800, 600, 16))
	gridRound1Approx(t, "vertical-rl item Y", item.Rect.Y, 5)
	gridRound1Approx(t, "vertical-rl item height", item.Rect.Height, 80)
	// Rows start at the right edge: right margin 40, so the item's right edge
	// is at 60 and its left at 40.
	gridRound1Approx(t, "vertical-rl item X", item.Rect.X, 40)
	gridRound1Approx(t, "vertical-rl item width", item.Rect.Width, 20)
}
