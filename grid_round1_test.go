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

// gridRound1JustifyGrid returns a 400px-wide grid with two columns (100px
// each unless given) and one 50px row, ready for a justify-content value.
func gridRound1JustifyGrid(justify JustifyContent, columns ...GridTrack) *Node {
	if len(columns) == 0 {
		columns = []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))}
	}
	return &Node{
		Style: Style{
			Display:             DisplayGrid,
			Width:               Px(400),
			GridTemplateColumns: columns,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50))},
			JustifyContent:      justify,
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridColumnStart: 0}},
			{Style: Style{GridRowStart: 0, GridColumnStart: 1}},
		},
	}
}

// TestGridRound1JustifyContentColumns checks that justify-content distributes
// the column-axis free space (css-grid-1 §10.4 / css-align-3 §5.3). Before
// this fix columns always started at 0.
//
// https://www.w3.org/TR/css-grid-1/#grid-align
// https://www.w3.org/TR/css-align-3/#distribution-values
func TestGridRound1JustifyContentColumns(t *testing.T) {
	cases := []struct {
		name    string
		justify JustifyContent
		x0, x1  float64
	}{
		{"start", JustifyContentStart, 0, 100},
		{"center", JustifyContentCenter, 100, 200},
		{"end", JustifyContentEnd, 200, 300},
		{"space-between", JustifyContentSpaceBetween, 0, 300},
		{"space-around", JustifyContentSpaceAround, 50, 250},
		{"space-evenly", JustifyContentSpaceEvenly, 200.0 / 3, 200.0/3*2 + 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := gridRound1JustifyGrid(tc.justify)
			size := LayoutGrid(root, Loose(400, Unbounded), NewLayoutContext(800, 600, 16))
			gridRound1Approx(t, "item 0 X", root.Children[0].Rect.X, tc.x0)
			gridRound1Approx(t, "item 1 X", root.Children[1].Rect.X, tc.x1)
			gridRound1Approx(t, "item width", root.Children[1].Rect.Width, 100)
			gridRound1Approx(t, "container width", size.Width, 400)
		})
	}
}

// TestGridRound1JustifyContentStretchColumns checks that
// JustifyContentStretch grows auto columns to fill the free space (§12.8)
// while leaving fixed columns alone.
//
// https://www.w3.org/TR/css-grid-1/#algo-stretch
func TestGridRound1JustifyContentStretchColumns(t *testing.T) {
	root := gridRound1JustifyGrid(JustifyContentStretch, FixedTrack(Px(100)), AutoTrack())
	LayoutGrid(root, Loose(400, Unbounded), NewLayoutContext(800, 600, 16))
	gridRound1Approx(t, "fixed column width", root.Children[0].Rect.Width, 100)
	gridRound1Approx(t, "auto column X", root.Children[1].Rect.X, 100)
	gridRound1Approx(t, "auto column width", root.Children[1].Rect.Width, 300)

	// Without stretch the auto column is content-sized (0 here) and the
	// columns are start-aligned.
	root = gridRound1JustifyGrid(JustifyContentStart, FixedTrack(Px(100)), AutoTrack())
	LayoutGrid(root, Loose(400, Unbounded), NewLayoutContext(800, 600, 16))
	gridRound1Approx(t, "unstretched auto column width", root.Children[1].Rect.Width, 0)
}

// TestGridRound1AlignContentSpaceEvenlyRows checks the new
// AlignContentSpaceEvenly value on the row axis.
//
// https://www.w3.org/TR/css-align-3/#valdef-align-content-space-evenly
func TestGridRound1AlignContentSpaceEvenlyRows(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			Height:              Px(300),
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			AlignContent:        AlignContentSpaceEvenly,
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridColumnStart: 0}},
			{Style: Style{GridRowStart: 1, GridColumnStart: 0}},
		},
	}
	LayoutGrid(root, Loose(100, 300), NewLayoutContext(800, 600, 16))
	// Free space 200 split into three equal portions of 66.67.
	gridRound1Approx(t, "row 0 Y", root.Children[0].Rect.Y, 200.0/3)
	gridRound1Approx(t, "row 1 Y", root.Children[1].Rect.Y, 200.0/3*2+50)
}

// TestGridRound1JustifyContentVerticalRL checks justify-content in a
// vertical-rl grid: the column axis is physical Y, so centering the columns
// offsets Y, while rows still progress from the right edge.
//
// https://www.w3.org/TR/css-writing-modes-3/#logical-to-physical
func TestGridRound1JustifyContentVerticalRL(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			WritingMode:         WritingModeVerticalRL,
			Width:               Px(100),
			Height:              Px(400),
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50))},
			JustifyContent:      JustifyContentCenter,
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridColumnStart: 0}},
			{Style: Style{GridRowStart: 0, GridColumnStart: 1}},
		},
	}
	LayoutGrid(root, Loose(100, 400), NewLayoutContext(800, 600, 16))
	// Columns run along Y: free space 200, centered offset 100.
	gridRound1Approx(t, "column 0 Y", root.Children[0].Rect.Y, 100)
	gridRound1Approx(t, "column 1 Y", root.Children[1].Rect.Y, 200)
	gridRound1Approx(t, "item height (column size)", root.Children[0].Rect.Height, 100)
	// Rows are start-aligned from the right edge (vertical-rl): the 50px row
	// occupies X 50..100 of the 100px row axis.
	gridRound1Approx(t, "item X (row from right)", root.Children[0].Rect.X, 50)
	gridRound1Approx(t, "item width (row size)", root.Children[0].Rect.Width, 50)
}

// TestGridRound1NamedAreasDoNotMutateStyle checks that resolving a named grid
// area (§7.3) no longer writes GridRow*/GridColumn* into the child's Style,
// so the same tree laid out again with a different template is placed by the
// new template.
//
// https://www.w3.org/TR/css-grid-1/#grid-template-areas-property
func TestGridRound1NamedAreasDoNotMutateStyle(t *testing.T) {
	item := PlaceInArea(&Node{}, "main")
	// Placement-related fields as set by the user; they must survive layout.
	type placementFields struct {
		rowStart, rowEnd, colStart, colEnd int
		area                               string
	}
	snapshot := func() placementFields {
		return placementFields{
			item.Style.GridRowStart, item.Style.GridRowEnd,
			item.Style.GridColumnStart, item.Style.GridColumnEnd,
			item.Style.GridArea,
		}
	}
	before := snapshot()

	first := NewGridTemplateAreas(2, 2)
	if err := first.DefineArea("main", 0, 1, 0, 1); err != nil {
		t.Fatal(err)
	}
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			GridTemplateAreas:   first,
		},
		Children: []*Node{item},
	}
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, Loose(200, 100), ctx)
	gridRound1Approx(t, "first layout X", item.Rect.X, 0)
	gridRound1Approx(t, "first layout Y", item.Rect.Y, 0)

	if got := snapshot(); got != before {
		t.Errorf("layout mutated the child's Style: before %+v, after %+v", before, got)
	}

	// Move "main" to the bottom-right cell and lay out again.
	second := NewGridTemplateAreas(2, 2)
	if err := second.DefineArea("main", 1, 2, 1, 2); err != nil {
		t.Fatal(err)
	}
	root.Style.GridTemplateAreas = second
	LayoutGrid(root, Loose(200, 100), ctx)
	gridRound1Approx(t, "second layout X", item.Rect.X, 100)
	gridRound1Approx(t, "second layout Y", item.Rect.Y, 50)
	if got := snapshot(); got != before {
		t.Errorf("second layout mutated the child's Style: %+v", got)
	}
}
