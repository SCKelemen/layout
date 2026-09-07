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

// TestGridRound1AspectRatioVerticalWritingModes checks that an aspect-ratio
// item in a vertical writing mode is fitted to its grid area in logical
// terms. The area is one 200px column (physical height) by one 100px row
// (physical width); a 2:1 (width:height) item must fit inside it as 100x50
// and, in vertical-rl, sit against the right edge. Previously the physical
// width/height were fitted against the logical column/row sizes, producing a
// 100x200 box outside the area.
//
// https://www.w3.org/TR/css-sizing-4/#aspect-ratio
// https://www.w3.org/TR/css-writing-modes-3/#logical-to-physical
func TestGridRound1AspectRatioVerticalWritingModes(t *testing.T) {
	for _, mode := range []WritingMode{WritingModeVerticalRL, WritingModeVerticalLR} {
		item := &Node{Style: Style{AspectRatio: 2}}
		root := &Node{
			Style: Style{
				Display:             DisplayGrid,
				WritingMode:         mode,
				GridTemplateColumns: []GridTrack{FixedTrack(Px(200))},
				GridTemplateRows:    []GridTrack{FixedTrack(Px(100))},
			},
			Children: []*Node{item},
		}
		size := LayoutGrid(root, Loose(Unbounded, Unbounded), NewLayoutContext(800, 600, 16))
		name := "vertical-lr"
		if mode == WritingModeVerticalRL {
			name = "vertical-rl"
		}
		gridRound1Approx(t, name+" container width", size.Width, 100)
		gridRound1Approx(t, name+" container height", size.Height, 200)
		gridRound1Approx(t, name+" item width", item.Rect.Width, 100)
		gridRound1Approx(t, name+" item height", item.Rect.Height, 50)
		gridRound1Approx(t, name+" item X", item.Rect.X, 0)
		gridRound1Approx(t, name+" item Y", item.Rect.Y, 0)
		if item.Rect.X+item.Rect.Width > 100.01 || item.Rect.Y+item.Rect.Height > 200.01 {
			t.Errorf("%s: item %+v overflows its 100x200 area", name, item.Rect)
		}
	}
}

// gridRound1RepeatGrid returns a grid using a column repeat() pattern with the
// given number of auto-placed 20px-tall items. Width is explicit when > 0.
func gridRound1RepeatGrid(width float64, explicit []GridTrack, repeat RepeatTrack, items int) *Node {
	root := &Node{
		Style: Style{
			Display:                   DisplayGrid,
			GridTemplateColumns:       explicit,
			GridTemplateColumnsRepeat: []RepeatTrack{repeat},
			GridAutoRows:              FixedTrack(Px(20)),
			GridGap:                   Px(10),
		},
	}
	if width > 0 {
		root.Style.Width = Px(width)
	}
	for i := 0; i < items; i++ {
		root.Children = append(root.Children, &Node{})
	}
	return root
}

// TestGridRound1AutoFillColumns checks that repeat(auto-fill, 100px) in a
// 500px grid with a 10px gap produces floor((500+10)/(100+10)) = 4 columns
// (§7.2.3.2): the fourth item sits in the fourth column and the fifth wraps
// to the next row.
//
// https://www.w3.org/TR/css-grid-1/#auto-repeat
func TestGridRound1AutoFillColumns(t *testing.T) {
	root := gridRound1RepeatGrid(500, nil, RepeatTrack{Count: RepeatCountAutoFill, Tracks: []GridTrack{FixedTrack(Px(100))}}, 5)
	size := LayoutGrid(root, Loose(500, Unbounded), NewLayoutContext(800, 600, 16))

	gridRound1Approx(t, "item 3 X", root.Children[3].Rect.X, 330)
	gridRound1Approx(t, "item 3 Y", root.Children[3].Rect.Y, 0)
	gridRound1Approx(t, "item 4 X (wrapped)", root.Children[4].Rect.X, 0)
	gridRound1Approx(t, "item 4 Y (wrapped)", root.Children[4].Rect.Y, 30)
	gridRound1Approx(t, "column width", root.Children[0].Rect.Width, 100)
	gridRound1Approx(t, "container width", size.Width, 500)
	gridRound1Approx(t, "container height", size.Height, 50)
}

// TestGridRound1AutoFitCollapsesEmptyTracks checks that with auto-fit the
// two empty trailing tracks (and their gutters) collapse: with
// justify-content: end the two occupied tracks are pushed against the end
// edge with no trailing gaps, so the track total is 210 rather than 430.
//
// https://www.w3.org/TR/css-grid-1/#collapsed-track
func TestGridRound1AutoFitCollapsesEmptyTracks(t *testing.T) {
	pattern := []GridTrack{FixedTrack(Px(100))}

	fit := gridRound1RepeatGrid(500, nil, RepeatTrack{Count: RepeatCountAutoFit, Tracks: pattern}, 2)
	fit.Style.JustifyContent = JustifyContentEnd
	LayoutGrid(fit, Loose(500, Unbounded), NewLayoutContext(800, 600, 16))
	gridRound1Approx(t, "auto-fit item 0 X", fit.Children[0].Rect.X, 290)
	gridRound1Approx(t, "auto-fit item 1 X", fit.Children[1].Rect.X, 400)

	// auto-fill keeps the empty tracks: 4 tracks + 3 gaps = 430, free 70.
	fill := gridRound1RepeatGrid(500, nil, RepeatTrack{Count: RepeatCountAutoFill, Tracks: pattern}, 2)
	fill.Style.JustifyContent = JustifyContentEnd
	LayoutGrid(fill, Loose(500, Unbounded), NewLayoutContext(800, 600, 16))
	gridRound1Approx(t, "auto-fill item 0 X", fill.Children[0].Rect.X, 70)
	gridRound1Approx(t, "auto-fill item 1 X", fill.Children[1].Rect.X, 180)

	// An unset width with a bounded constraint is definite too; the
	// container still fills it, but the track extent is 210.
	fitAuto := gridRound1RepeatGrid(0, nil, RepeatTrack{Count: RepeatCountAutoFit, Tracks: pattern}, 2)
	fitAuto.Style.JustifyContent = JustifyContentEnd
	LayoutGrid(fitAuto, Loose(500, Unbounded), NewLayoutContext(800, 600, 16))
	gridRound1Approx(t, "auto-width auto-fit item 1 X", fitAuto.Children[1].Rect.X, 400)
}

// TestGridRound1AutoFitInteriorCollapse checks that an empty auto-fit track
// between two occupied ones collapses together with its gutters, leaving a
// single gap between the neighbors, and that the items keep their tracks.
func TestGridRound1AutoFitInteriorCollapse(t *testing.T) {
	root := gridRound1RepeatGrid(500, nil, RepeatTrack{Count: RepeatCountAutoFit, Tracks: []GridTrack{FixedTrack(Px(100))}}, 2)
	root.Children[0].Style.GridColumnStart = 0
	root.Children[1].Style.GridColumnStart = 2 // leaves column 1 empty
	root.Children[0].Style.GridRowStart, root.Children[1].Style.GridRowStart = 0, 0
	root.Style.GridTemplateRows = []GridTrack{FixedTrack(Px(20))}
	LayoutGrid(root, Loose(500, Unbounded), NewLayoutContext(800, 600, 16))

	gridRound1Approx(t, "item 0 X", root.Children[0].Rect.X, 0)
	gridRound1Approx(t, "item 1 X (one gap after collapse)", root.Children[1].Rect.X, 110)
	gridRound1Approx(t, "item 1 width", root.Children[1].Rect.Width, 100)
}

// TestGridRound1AutoRepeatIndefiniteAxis checks that an auto-repeat against
// an indefinite axis size repeats exactly once (§7.2.3.2), so three items
// stack in a single 100px column.
func TestGridRound1AutoRepeatIndefiniteAxis(t *testing.T) {
	root := gridRound1RepeatGrid(0, nil, RepeatTrack{Count: RepeatCountAutoFill, Tracks: []GridTrack{FixedTrack(Px(100))}}, 3)
	size := LayoutGrid(root, Loose(Unbounded, Unbounded), NewLayoutContext(800, 600, 16))

	gridRound1Approx(t, "container width", size.Width, 100)
	gridRound1Approx(t, "item 1 X", root.Children[1].Rect.X, 0)
	gridRound1Approx(t, "item 1 Y", root.Children[1].Rect.Y, 30)
	gridRound1Approx(t, "item 2 Y", root.Children[2].Rect.Y, 60)
}

// TestGridRound1ExplicitPlusAutoRepeat checks that explicit tracks come first
// and the auto-repeat fills the remaining space: [50px] + repeat(auto-fill,
// 100px) in 500px with 10px gaps leaves 450px for floor((450+10)/110) = 4
// repetitions, five columns in total.
func TestGridRound1ExplicitPlusAutoRepeat(t *testing.T) {
	root := gridRound1RepeatGrid(500, []GridTrack{FixedTrack(Px(50))}, RepeatTrack{Count: RepeatCountAutoFill, Tracks: []GridTrack{FixedTrack(Px(100))}}, 6)
	LayoutGrid(root, Loose(500, Unbounded), NewLayoutContext(800, 600, 16))

	gridRound1Approx(t, "item 0 width (explicit 50px)", root.Children[0].Rect.Width, 50)
	gridRound1Approx(t, "item 1 X", root.Children[1].Rect.X, 60)
	gridRound1Approx(t, "item 4 X (fifth column)", root.Children[4].Rect.X, 390)
	gridRound1Approx(t, "item 4 Y", root.Children[4].Rect.Y, 0)
	gridRound1Approx(t, "item 5 Y (wrapped)", root.Children[5].Rect.Y, 30)
}

// TestGridRound1FixedCountRepeat checks repeat(2, [100px 50px]) expands to
// four columns regardless of the available size.
func TestGridRound1FixedCountRepeat(t *testing.T) {
	root := gridRound1RepeatGrid(0, nil, RepeatTrack{Count: 2, Tracks: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(50))}}, 4)
	root.Style.GridGap = Px(0)
	size := LayoutGrid(root, Loose(Unbounded, Unbounded), NewLayoutContext(800, 600, 16))

	xs := []float64{0, 100, 150, 250}
	for i, want := range xs {
		gridRound1Approx(t, "item X", root.Children[i].Rect.X, want)
		gridRound1Approx(t, "item Y", root.Children[i].Rect.Y, 0)
	}
	gridRound1Approx(t, "container width", size.Width, 300)
}

// TestGridRound1InvalidAutoRepeatIgnored checks that an auto-repeat whose
// pattern is not fixed-size (1fr here) is ignored: the template stays empty
// and items stack in a single implicit column.
//
// https://www.w3.org/TR/css-grid-1/#auto-repeat (fixed-size requirement)
func TestGridRound1InvalidAutoRepeatIgnored(t *testing.T) {
	root := gridRound1RepeatGrid(500, nil, RepeatTrack{Count: RepeatCountAutoFill, Tracks: []GridTrack{FractionTrack(1)}}, 2)
	LayoutGrid(root, Loose(500, Unbounded), NewLayoutContext(800, 600, 16))

	gridRound1Approx(t, "item 1 X", root.Children[1].Rect.X, 0)
	gridRound1Approx(t, "item 1 Y (second row)", root.Children[1].Rect.Y, 30)
}

// TestGridRound1AutoFillRows checks the row axis: repeat(auto-fill, 50px)
// against a 200px-tall grid yields 4 rows, so with column flow the fifth
// item starts a second column.
func TestGridRound1AutoFillRows(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:                DisplayGrid,
			Height:                 Px(200),
			GridTemplateRowsRepeat: []RepeatTrack{{Count: RepeatCountAutoFill, Tracks: []GridTrack{FixedTrack(Px(50))}}},
			GridAutoColumns:        FixedTrack(Px(30)),
			GridAutoFlow:           GridAutoFlowColumn,
		},
	}
	for i := 0; i < 5; i++ {
		root.Children = append(root.Children, &Node{})
	}
	LayoutGrid(root, Loose(Unbounded, 200), NewLayoutContext(800, 600, 16))

	gridRound1Approx(t, "item 3 Y", root.Children[3].Rect.Y, 150)
	gridRound1Approx(t, "item 3 X", root.Children[3].Rect.X, 0)
	gridRound1Approx(t, "item 4 X (second column)", root.Children[4].Rect.X, 30)
	gridRound1Approx(t, "item 4 Y", root.Children[4].Rect.Y, 0)
}

// TestGridRound1AutoFitEmptyContainer checks that an auto-fit grid without
// items collapses every repeated track, so only the explicit tracks remain.
func TestGridRound1AutoFitEmptyContainer(t *testing.T) {
	root := gridRound1RepeatGrid(0, []GridTrack{FixedTrack(Px(50))}, RepeatTrack{Count: RepeatCountAutoFit, Tracks: []GridTrack{FixedTrack(Px(100))}}, 0)
	root.Style.GridTemplateRows = []GridTrack{FixedTrack(Px(20))}
	size := LayoutGrid(root, Loose(500, Unbounded), NewLayoutContext(800, 600, 16))
	gridRound1Approx(t, "container width (explicit track only)", size.Width, 50)
	gridRound1Approx(t, "container height", size.Height, 20)
}
