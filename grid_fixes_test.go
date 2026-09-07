package layout

import (
	"math"
	"testing"
)

// Regression tests for the grid correctness fixes. Each test names the bug
// it guards against and links the governing CSS Grid Layout Module Level 1
// section.

func gridFixLayout(t *testing.T, root *Node, c Constraints) Size {
	t.Helper()
	ctx := NewLayoutContext(800, 600, 16)
	return LayoutGrid(root, c, ctx)
}

func gridFixExpectRect(t *testing.T, name string, n *Node, x, y, w, h float64) {
	t.Helper()
	if math.Abs(n.Rect.X-x) > 0.01 || math.Abs(n.Rect.Y-y) > 0.01 ||
		math.Abs(n.Rect.Width-w) > 0.01 || math.Abs(n.Rect.Height-h) > 0.01 {
		t.Errorf("%s: expected rect (%.2f,%.2f %.2fx%.2f), got (%.2f,%.2f %.2fx%.2f)",
			name, x, y, w, h, n.Rect.X, n.Rect.Y, n.Rect.Width, n.Rect.Height)
	}
}

// Fix 1: dense auto-placement used to set rowStart = len(rows) without
// growing the implicit grid and then indexed past rowHeights (panic).
// §8.5 step 4: the implicit grid grows as necessary.
// https://www.w3.org/TR/css-grid-1/#auto-placement-algo
func TestGridFixDenseFallbackGrowsImplicitRows(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridAutoRows:        FixedTrack(Px(50)),
			GridAutoFlow:        GridAutoFlowRowDense,
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridRowEnd: 2, GridColumnStart: 0, GridColumnEnd: 1, Height: Px(100)}},
			{Style: Style{GridRowStart: -1, GridRowEnd: -1, GridColumnStart: -1, GridColumnEnd: -1, Height: Px(50)}},
		},
	}
	size := gridFixLayout(t, root, Loose(200, Unbounded))

	// Item B lands in the new implicit row 2 (50px, from grid-auto-rows).
	gridFixExpectRect(t, "dense fallback item", root.Children[1], 0, 100, 100, 50)
	if math.Abs(size.Height-150) > 0.01 {
		t.Errorf("container height: expected 150 (3 x 50px rows), got %.2f", size.Height)
	}
}

// Fix 2: sparse auto-placement was index-based and ignored occupancy.
// §8.5: auto items go to the first cell not covered by a placed item.
// https://www.w3.org/TR/css-grid-1/#auto-placement-algo
func TestGridFixSparsePlacementSkipsOccupiedCells(t *testing.T) {
	newGrid := func(children ...*Node) *Node {
		return &Node{
			Style: Style{
				Display:             DisplayGrid,
				GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
				GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			},
			Children: children,
		}
	}

	t.Run("explicit at (0,1) then auto", func(t *testing.T) {
		root := newGrid(
			&Node{Style: Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 1, GridColumnEnd: 2}},
			&Node{},
		)
		gridFixLayout(t, root, Loose(200, 100))
		gridFixExpectRect(t, "auto item", root.Children[1], 0, 0, 100, 50)
	})

	t.Run("explicit at (1,0) then auto", func(t *testing.T) {
		root := newGrid(
			&Node{Style: Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 0, GridColumnEnd: 1}},
			&Node{},
		)
		gridFixLayout(t, root, Loose(200, 100))
		gridFixExpectRect(t, "auto item", root.Children[1], 0, 0, 100, 50)
	})

	t.Run("row-locked item takes first free cell in its row", func(t *testing.T) {
		root := newGrid(
			&Node{Style: Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 0, GridColumnEnd: 1}},
			&Node{Style: Style{GridRowStart: 1, GridRowEnd: 2}}, // auto column
		)
		gridFixLayout(t, root, Loose(200, 100))
		gridFixExpectRect(t, "row-locked item", root.Children[1], 100, 50, 100, 50)
	})

	t.Run("spanning explicit item blocks cells", func(t *testing.T) {
		root := &Node{
			Style: Style{
				Display:             DisplayGrid,
				GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50)), FixedTrack(Px(50))},
				GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			},
			Children: []*Node{
				{Style: Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 0, GridColumnEnd: 2}},
				{}, {}, {},
			},
		}
		gridFixLayout(t, root, Loose(150, 100))
		gridFixExpectRect(t, "auto 1", root.Children[1], 100, 0, 50, 50)
		gridFixExpectRect(t, "auto 2", root.Children[2], 0, 50, 50, 50)
		gridFixExpectRect(t, "auto 3", root.Children[3], 50, 50, 50, 50)
	})

	t.Run("column flow skips occupied cells", func(t *testing.T) {
		root := newGrid(
			&Node{Style: Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 0, GridColumnEnd: 1}},
			&Node{}, &Node{},
		)
		root.Style.GridAutoFlow = GridAutoFlowColumn
		gridFixLayout(t, root, Loose(200, 100))
		gridFixExpectRect(t, "auto 1", root.Children[1], 0, 50, 100, 50)
		gridFixExpectRect(t, "auto 2", root.Children[2], 100, 0, 100, 50)
	})
}

// Fix 3: dense packing treated default-constructed items (GridRowStart 0,
// GridRowEnd 0) as explicitly positioned, so they were never re-packed.
// §8.5 dense: the cursor restarts at the start for every auto item.
// https://www.w3.org/TR/css-grid-1/#grid-auto-flow-property
func TestGridFixDensePacksDefaultConstructedItems(t *testing.T) {
	build := func(flow GridAutoFlow) *Node {
		return &Node{
			Style: Style{
				Display:             DisplayGrid,
				GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50)), FixedTrack(Px(50))},
				GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
				GridAutoFlow:        flow,
			},
			Children: []*Node{
				{Style: Style{GridColumnStart: 2, GridColumnEnd: 3}}, // auto row, column 2
				{}, // both auto, default-constructed
			},
		}
	}

	sparse := build(GridAutoFlowRow)
	gridFixLayout(t, sparse, Loose(150, 100))
	// Sparse: the cursor is at (0,2) after the first item and only moves
	// forward, so the second item wraps to (1,0).
	gridFixExpectRect(t, "sparse item", sparse.Children[1], 0, 50, 50, 50)

	dense := build(GridAutoFlowRowDense)
	gridFixLayout(t, dense, Loose(150, 100))
	// Dense: the cursor restarts, so the hole at (0,0) is filled.
	gridFixExpectRect(t, "dense item", dense.Children[1], 0, 0, 50, 50)
}

// Fix 4: a zero-value GridAutoRows produced minmax(0,0) implicit rows.
// §7.6: the initial value of grid-auto-rows is auto.
// https://www.w3.org/TR/css-grid-1/#auto-tracks
func TestGridFixZeroValueAutoRowsAreAuto(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
		},
		Children: []*Node{
			{Style: Style{Height: Px(50)}}, {Style: Style{Height: Px(50)}},
			{Style: Style{Height: Px(50)}}, {Style: Style{Height: Px(50)}},
		},
	}
	size := gridFixLayout(t, root, Loose(200, Unbounded))
	gridFixExpectRect(t, "row 1 item", root.Children[2], 0, 50, 100, 50)
	if math.Abs(size.Height-100) > 0.01 {
		t.Errorf("container height: expected 100, got %.2f", size.Height)
	}

	// Also for a zero-value track inside a template.
	root = &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    make([]GridTrack, 2),
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridRowEnd: 1, Height: Px(30)}},
			{Style: Style{GridRowStart: 1, GridRowEnd: 2, Height: Px(70)}},
		},
	}
	size = gridFixLayout(t, root, Loose(100, Unbounded))
	if math.Abs(size.Height-100) > 0.01 {
		t.Errorf("zero-value template rows should be auto: expected height 100, got %.2f", size.Height)
	}
}

// Fix 5: an unset Width/Height resolved to 0 and was treated as definite,
// collapsing fr tracks. An auto width fills the available inline size; an
// auto height is indefinite so fr rows size to content.
// https://www.w3.org/TR/css-grid-1/#algo-flex-tracks
func TestGridFixUnsetContainerSizeIsAuto(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FractionTrack(1), FractionTrack(1)},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50))},
		},
		Children: []*Node{{}, {}},
	}
	size := gridFixLayout(t, root, Loose(400, 400))
	gridFixExpectRect(t, "col 0", root.Children[0], 0, 0, 200, 50)
	gridFixExpectRect(t, "col 1", root.Children[1], 200, 0, 200, 50)
	if math.Abs(size.Width-400) > 0.01 {
		t.Errorf("container width: expected 400, got %.2f", size.Width)
	}

	// Unset height: fr rows are content-sized under a loose constraint...
	root = &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FractionTrack(1), FractionTrack(1)},
		},
		Children: []*Node{{Style: Style{Height: Px(50)}}, {Style: Style{Height: Px(30)}}},
	}
	size = gridFixLayout(t, root, Loose(100, 400))
	// §12.7.1 indefinite: flex fraction = max(base / factor) = 50 → 50 each.
	gridFixExpectRect(t, "fr row 1", root.Children[1], 0, 50, 100, 30)
	if math.Abs(size.Height-100) > 0.01 {
		t.Errorf("container height: expected 100, got %.2f", size.Height)
	}

	// ...and fill the space when the constraints force a definite height.
	root = &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FractionTrack(1), FractionTrack(1)},
		},
		Children: []*Node{{}, {}},
	}
	gridFixLayout(t, root, Tight(100, 400))
	gridFixExpectRect(t, "tight fr row 1", root.Children[1], 0, 200, 100, 200)
}

// Fix 6: align-content: stretch grew fixed tracks. §12.8 only stretches
// tracks whose max sizing function is auto.
// https://www.w3.org/TR/css-grid-1/#algo-stretch
func TestGridFixStretchOnlyGrowsAutoTracks(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			Height:              Px(500),
		},
		Children: []*Node{{}, {}},
	}
	gridFixLayout(t, root, Loose(100, 500))
	gridFixExpectRect(t, "fixed row 0", root.Children[0], 0, 0, 100, 100)
	gridFixExpectRect(t, "fixed row 1", root.Children[1], 0, 100, 100, 100)

	root = &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), AutoTrack()},
			Height:              Px(300),
		},
		Children: []*Node{{}, {}},
	}
	gridFixLayout(t, root, Loose(100, 300))
	gridFixExpectRect(t, "fixed row", root.Children[0], 0, 0, 100, 100)
	gridFixExpectRect(t, "stretched auto row", root.Children[1], 0, 100, 100, 200)
}

// Fix 7: vertical writing modes sized tracks against the wrong physical
// axis. Columns run along Y and rows along X in vertical-lr.
// https://www.w3.org/TR/css-writing-modes-3/#logical-to-physical
func TestGridFixVerticalWritingModeTrackAxes(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FractionTrack(1), FractionTrack(1)},
			GridTemplateRows:    []GridTrack{FractionTrack(1)},
			Width:               Px(400),
			Height:              Px(200),
			WritingMode:         WritingModeVerticalLR,
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 0, GridColumnEnd: 1}},
			{Style: Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 1, GridColumnEnd: 2}},
		},
	}
	gridFixLayout(t, root, Tight(400, 200))
	// Columns share the 200px height (100 each); the row spans the 400px width.
	gridFixExpectRect(t, "column 0", root.Children[0], 0, 0, 400, 100)
	gridFixExpectRect(t, "column 1", root.Children[1], 0, 100, 400, 100)

	// An item's explicit Height applies to the column axis in vertical mode.
	root.Children[0].Style.Height = Px(40)
	gridFixLayout(t, root, Tight(400, 200))
	gridFixExpectRect(t, "explicit column-axis size", root.Children[0], 0, 0, 400, 40)
}

// Fix 8: auto tracks were sized to 0 instead of their content.
// §11.5 / §12.5: auto max sizing function → max-content contributions.
// https://www.w3.org/TR/css-grid-1/#algo-content
func TestGridFixAutoTracksSizeToContent(t *testing.T) {
	root := GridAuto(1, 2)
	root.Children = []*Node{
		{Style: Style{Width: Px(100), Height: Px(50)}},
		{Style: Style{Width: Px(150), Height: Px(50)}},
	}
	size := gridFixLayout(t, root, Loose(500, 500))
	gridFixExpectRect(t, "auto column 0", root.Children[0], 0, 0, 100, 50)
	gridFixExpectRect(t, "auto column 1", root.Children[1], 100, 0, 150, 50)
	if math.Abs(size.Width-250) > 0.01 || math.Abs(size.Height-50) > 0.01 {
		t.Errorf("container: expected 250x50, got %.2fx%.2f", size.Width, size.Height)
	}
}

// Fix 9: a spanning item's height was split equally over fixed and auto
// rows. §12.5 step 3: extra space goes only to intrinsic tracks.
// https://www.w3.org/TR/css-grid-1/#algo-content
func TestGridFixSpanningExtraSpaceGoesToIntrinsicRows(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), AutoTrack()},
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridRowEnd: 2, GridColumnStart: 0, GridColumnEnd: 1, Height: Px(300)}},
			{Style: Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 1, GridColumnEnd: 2}},
		},
	}
	size := gridFixLayout(t, root, Loose(200, Unbounded))
	// Auto row = 300 - 100 (fixed row) = 200; the fixed row stays 100.
	gridFixExpectRect(t, "auto row item", root.Children[1], 100, 100, 100, 200)
	if math.Abs(size.Height-300) > 0.01 {
		t.Errorf("container height: expected 300, got %.2f", size.Height)
	}
}

// Fix 10: fr tracks ignored their auto minimum. §7.2.4: 1fr is
// minmax(auto, 1fr), so a content-based base size larger than the fr share
// wins (§12.7.1 restarts with the track treated as inflexible).
// https://www.w3.org/TR/css-grid-1/#algo-find-fr-size
func TestGridFixFlexTracksHonorContentMinimum(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FractionTrack(1)},
			Height:              Px(100),
		},
		Children: []*Node{{Style: Style{Height: Px(300)}}},
	}
	gridFixLayout(t, root, Loose(100, 100))
	gridFixExpectRect(t, "overflowing fr row", root.Children[0], 0, 0, 100, 300)

	// No remaining space: the fr row still gets its content size, not 0.
	root = &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FractionTrack(1)},
			Height:              Px(100),
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridRowEnd: 1}},
			{Style: Style{GridRowStart: 1, GridRowEnd: 2, Height: Px(50)}},
		},
	}
	gridFixLayout(t, root, Loose(100, 100))
	gridFixExpectRect(t, "fr row without free space", root.Children[1], 0, 100, 100, 50)

	// Same for columns: 1fr in 100px with a 300px child.
	root = &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FractionTrack(1)},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50))},
			Width:               Px(100),
		},
		Children: []*Node{{Style: Style{Width: Px(300)}}},
	}
	gridFixLayout(t, root, Loose(100, 50))
	gridFixExpectRect(t, "overflowing fr column", root.Children[0], 0, 0, 300, 50)

	// Two fr tracks where only one is content-bound: the other takes the rest.
	root = &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FractionTrack(1), FractionTrack(1)},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50))},
			Width:               Px(400),
		},
		Children: []*Node{
			{Style: Style{GridColumnStart: 0, GridColumnEnd: 1, Width: Px(300)}},
			{Style: Style{GridColumnStart: 1, GridColumnEnd: 2}},
		},
	}
	gridFixLayout(t, root, Loose(400, 50))
	gridFixExpectRect(t, "content-bound fr column", root.Children[0], 0, 0, 300, 50)
	gridFixExpectRect(t, "remaining fr column", root.Children[1], 300, 0, 100, 50)
}

// Fix 11: start == end produced a zero-span item. §8.3.1: if the end line
// equals the start line the end is removed (span 1); if it is before the
// start line the two are swapped.
// https://www.w3.org/TR/css-grid-1/#grid-placement-errors
func TestGridFixPlacementConflictHandling(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 1, GridColumnEnd: 1}}, // start == end
			{Style: Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 2, GridColumnEnd: 1}}, // end < start
		},
	}
	gridFixLayout(t, root, Loose(300, 100))
	gridFixExpectRect(t, "start == end", root.Children[0], 100, 0, 100, 50)
	gridFixExpectRect(t, "end < start", root.Children[1], 100, 50, 100, 50)

	spans := []struct {
		start, end int
		want       gridSpan
	}{
		{1, 1, gridSpan{start: 1, span: 1}},
		{2, 1, gridSpan{start: 1, span: 1}},
		{0, 0, gridSpan{start: 0, span: 1, auto: true}},
		{-1, 0, gridSpan{start: 0, span: 1, auto: true}},
		{0, 2, gridSpan{start: 0, span: 2}},
		{-1, 3, gridSpan{start: 2, span: 1}},
		{3, 0, gridSpan{start: 3, span: 1}},
	}
	for _, tc := range spans {
		if got := gridNormalizeSpan(tc.start, tc.end); got != tc.want {
			t.Errorf("gridNormalizeSpan(%d, %d) = %+v, want %+v", tc.start, tc.end, got, tc.want)
		}
	}
}

// Fix 12: an empty grid omitted row gaps from its height.
// https://www.w3.org/TR/css-grid-1/#gutters
func TestGridFixEmptyGridIncludesRowGaps(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridGap:             Px(10),
		},
	}
	size := gridFixLayout(t, root, Loose(300, 300))
	if math.Abs(size.Width-100) > 0.01 || math.Abs(size.Height-210) > 0.01 {
		t.Errorf("empty grid: expected 100x210, got %.2fx%.2f", size.Width, size.Height)
	}
}

// Fix 13: auto-repeat with an indefinite size produced ~MaxInt repetitions,
// and minmax() patterns counted with the min instead of the max sizing
// function. §7.2.3.2: indefinite → one repetition; use the max sizing
// function when definite. The count is also hard-capped.
// https://www.w3.org/TR/css-grid-1/#auto-repeat
func TestGridFixAutoRepeatCountBounds(t *testing.T) {
	fixed := RepeatTrack{Count: RepeatCountAutoFill, Tracks: []GridTrack{FixedTrack(Px(100))}}

	if got := calculateAutoRepeatCount(fixed, Unbounded, 10); got != 1 {
		t.Errorf("indefinite size: expected 1 repetition, got %d", got)
	}
	if got := calculateAutoRepeatCount(fixed, math.Inf(1), 10); got != 1 {
		t.Errorf("infinite size: expected 1 repetition, got %d", got)
	}
	if got := calculateAutoRepeatCount(fixed, math.NaN(), 10); got != 1 {
		t.Errorf("NaN size: expected 1 repetition, got %d", got)
	}
	if got := calculateAutoRepeatCount(fixed, 1e300, 0); got != gridMaxAutoRepeat {
		t.Errorf("huge size: expected cap %d, got %d", gridMaxAutoRepeat, got)
	}

	minmax := RepeatTrack{Count: RepeatCountAutoFill, Tracks: []GridTrack{MinMaxTrack(Px(50), Px(200))}}
	if got := calculateAutoRepeatCount(minmax, 400, 0); got != 2 {
		t.Errorf("minmax(50px, 200px) in 400px: expected 2 repetitions (max sizing function), got %d", got)
	}

	// Expansion with an indefinite size terminates with a single repetition.
	tracks := expandAutoRepeatTracks([]RepeatTrack{fixed}, nil, Unbounded, 10)
	if len(tracks) != 1 {
		t.Errorf("indefinite expansion: expected 1 track, got %d", len(tracks))
	}
	tracks = expandAutoRepeatTracks([]RepeatTrack{fixed}, nil, 1e300, 0)
	if len(tracks) != gridMaxAutoRepeat {
		t.Errorf("huge expansion: expected %d tracks, got %d", gridMaxAutoRepeat, len(tracks))
	}
}

// Guardrail: absurd explicit line numbers are clamped so the implicit grid
// can never grow (or allocate) without bound.
// https://www.w3.org/TR/css-grid-1/#overlarge-grids
func TestGridFixImplicitGridIsCapped(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 1 << 40, GridRowEnd: 1<<40 + 5, Height: Px(10)}},
			{Style: Style{GridColumnStart: 1 << 40, Height: Px(10)}},
			{Style: Style{Height: Px(10)}},
		},
	}
	rows := []GridTrack{AutoTrack()}
	cols := []GridTrack{FixedTrack(Px(100))}
	items := gridPlaceItems(root, &rows, &cols, GridAutoFlowRow)
	if len(rows) > gridMaxTracks || len(cols) > gridMaxTracks {
		t.Fatalf("implicit grid exceeded cap: %d rows, %d columns", len(rows), len(cols))
	}
	for i, item := range items {
		if item.rowEnd > len(rows) || item.colEnd > len(cols) || item.rowStart < 0 || item.colStart < 0 {
			t.Errorf("item %d placed outside the grid: rows %d-%d cols %d-%d", i, item.rowStart, item.rowEnd, item.colStart, item.colEnd)
		}
	}
	// The full layout must not panic either.
	gridFixLayout(t, root, Loose(100, Unbounded))
}
