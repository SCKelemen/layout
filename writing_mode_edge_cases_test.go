package layout

import (
	"testing"
)

// TestFlexboxWrapVertical tests flexbox wrapping in vertical writing mode
func TestFlexboxWrapVertical(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with vertical-lr and wrap
	// In vertical-lr with row (main=vertical), items should wrap to new columns
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow, // Main axis vertical
			FlexWrap:      FlexWrapWrap,
			Width:         Px(200),
			Height:        Px(150), // Main axis size
			WritingMode:   WritingModeVerticalLR,
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(100), FlexShrink: 0}},
			{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(100), FlexShrink: 0}},
			{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(100), FlexShrink: 0}},
		},
	}

	constraints := Tight(200, 150)
	LayoutFlexbox(root, constraints, ctx)

	// With main axis=150, items height=100 each
	// First 2 items fit in first column (100+100 > 150, so actually just 1 item per column)
	// Actually: 100+100 = 200 > 150, so items wrap
	// Line 1: Item 0 (Y=0)
	// Line 2: Item 1 (Y=0, but X increased)
	// Line 3: Item 2 (Y=0, but X increased more)

	child0 := root.Children[0]
	child1 := root.Children[1]
	child2 := root.Children[2]

	// All items should start at Y=0 (main axis start)
	if child0.Rect.Y != 0 {
		t.Errorf("Child 0 Y: expected 0, got %.2f", child0.Rect.Y)
	}
	if child1.Rect.Y != 0 {
		t.Errorf("Child 1 Y: expected 0, got %.2f", child1.Rect.Y)
	}
	if child2.Rect.Y != 0 {
		t.Errorf("Child 2 Y: expected 0, got %.2f", child2.Rect.Y)
	}

	// Items should be on different lines (X increases)
	if child0.Rect.X >= child1.Rect.X {
		t.Errorf("Child 0 X (%.2f) should be < Child 1 X (%.2f)", child0.Rect.X, child1.Rect.X)
	}
	if child1.Rect.X >= child2.Rect.X {
		t.Errorf("Child 1 X (%.2f) should be < Child 2 X (%.2f)", child1.Rect.X, child2.Rect.X)
	}
}

// TestFlexboxGapVertical tests flexbox gaps in vertical writing mode
func TestFlexboxGapVertical(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow, // Main axis vertical
			FlexGap:       Px(10),
			Width:         Px(200),
			Height:        Px(300),
			WritingMode:   WritingModeVerticalLR,
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(100), FlexShrink: 0}},
			{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(100), FlexShrink: 0}},
		},
	}

	constraints := Tight(200, 300)
	LayoutFlexbox(root, constraints, ctx)

	child0 := root.Children[0]
	child1 := root.Children[1]

	// Gap should be in main axis direction (vertical = Y)
	// Child 0 at Y=0, height=100
	// Gap = 10
	// Child 1 at Y=110
	expectedGap := child1.Rect.Y - (child0.Rect.Y + child0.Rect.Height)
	if expectedGap != 10.0 {
		t.Errorf("Gap: expected 10, got %.2f (child1.Y=%.2f, child0.bottom=%.2f)",
			expectedGap, child1.Rect.Y, child0.Rect.Y+child0.Rect.Height)
	}
}

// TestFlexboxRowReverseVertical tests row-reverse in vertical writing mode
func TestFlexboxRowReverseVertical(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRowReverse, // Main axis vertical, reversed
			Width:         Px(200),
			Height:        Px(300),
			WritingMode:   WritingModeVerticalLR,
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(100), FlexShrink: 0}},
			{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(100), FlexShrink: 0}},
		},
	}

	constraints := Tight(200, 300)
	LayoutFlexbox(root, constraints, ctx)

	child0 := root.Children[0]
	child1 := root.Children[1]

	// In row-reverse with vertical main axis, items should be reversed
	// Child 0 should be below Child 1
	if child0.Rect.Y <= child1.Rect.Y {
		t.Errorf("Child 0 Y (%.2f) should be > Child 1 Y (%.2f) for row-reverse",
			child0.Rect.Y, child1.Rect.Y)
	}
}

// TestGridSpanningVertical tests grid spanning in vertical writing mode
func TestGridSpanningVertical(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			Width:               Px(200),
			Height:              Px(100),
			WritingMode:         WritingModeVerticalLR,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:         DisplayBlock,
					GridRowStart:    0,
					GridRowEnd:      2, // Span 2 rows
					GridColumnStart: 0,
					GridColumnEnd:   1,
				},
			},
			{
				Style: Style{
					Display:         DisplayBlock,
					GridRowStart:    0,
					GridRowEnd:      1,
					GridColumnStart: 1,
					GridColumnEnd:   2,
				},
			},
		},
	}

	constraints := Tight(200, 100)
	LayoutGrid(root, constraints, ctx)

	child0 := root.Children[0] // Spans 2 rows
	child1 := root.Children[1] // Single cell

	// In vertical-lr: rows control X, columns control Y
	// Child 0 spans rows 0-2, so width should be 200 (2 rows * 100)
	expectedWidth := 200.0
	if child0.Rect.Width != expectedWidth {
		t.Errorf("Spanning child width: expected %.2f, got %.2f", expectedWidth, child0.Rect.Width)
	}

	// Child 0 in column 0 (Y=0), Child 1 in column 1 (Y=50)
	if child0.Rect.Y != 0 {
		t.Errorf("Child 0 Y: expected 0, got %.2f", child0.Rect.Y)
	}
	if child1.Rect.Y != 50 {
		t.Errorf("Child 1 Y: expected 50, got %.2f", child1.Rect.Y)
	}

	// Both children start at row 0 (X=0)
	if child0.Rect.X != 0 || child1.Rect.X != 0 {
		t.Errorf("Children should have X=0, got child0=%.2f, child1=%.2f", child0.Rect.X, child1.Rect.X)
	}
}

// TestGridGapVertical tests grid gaps in vertical writing mode
func TestGridGapVertical(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			GridRowGap:          Px(10), // Gap between rows (horizontal in vertical-lr)
			GridColumnGap:       Px(5),  // Gap between columns (vertical in vertical-lr)
			Width:               Px(250),
			Height:              Px(150),
			WritingMode:         WritingModeVerticalLR,
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 0, GridColumnEnd: 1}},
			{Style: Style{Display: DisplayBlock, GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 0, GridColumnEnd: 1}},
			{Style: Style{Display: DisplayBlock, GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 1, GridColumnEnd: 2}},
		},
	}

	constraints := Tight(250, 150)
	LayoutGrid(root, constraints, ctx)

	child0 := root.Children[0] // Row 0, Col 0
	child1 := root.Children[1] // Row 1, Col 0
	child2 := root.Children[2] // Row 0, Col 1

	// In vertical-lr: row gap is horizontal (X), column gap is vertical (Y)
	// Child 0 and Child 1 are in same column, different rows
	// So X should differ by row size + row gap = 100 + 10 = 110
	xGap := child1.Rect.X - child0.Rect.X
	if xGap != 110.0 {
		t.Errorf("Row gap in X: expected 110, got %.2f", xGap)
	}

	// Child 0 and Child 2 are in same row, different columns
	// So Y should differ by column size + column gap = 50 + 5 = 55
	yGap := child2.Rect.Y - child0.Rect.Y
	if yGap != 55.0 {
		t.Errorf("Column gap in Y: expected 55, got %.2f", yGap)
	}
}

// TestNestedMixedWritingModes tests complex nesting with different writing modes
func TestNestedMixedWritingModes(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Horizontal block > Vertical flexbox > Horizontal grid
	root := &Node{
		Style: Style{
			Display:     DisplayBlock,
			Width:       Px(400),
			Height:      Px(300),
			WritingMode: WritingModeHorizontalTB,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:       DisplayFlex,
					FlexDirection: FlexDirectionRow,
					Width:         Px(200),
					Height:        Px(150),
					WritingMode:   WritingModeVerticalLR,
				},
				Children: []*Node{
					{
						Style: Style{
							Display:             DisplayGrid,
							GridTemplateRows:    []GridTrack{FixedTrack(Px(30)), FixedTrack(Px(30))},
							GridTemplateColumns: []GridTrack{FixedTrack(Px(40))},
							Width:               Px(40),
							Height:              Px(60),
							WritingMode:         WritingModeHorizontalTB,
						},
						Children: []*Node{
							{Style: Style{Display: DisplayBlock, GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 0, GridColumnEnd: 1}},
							{Style: Style{Display: DisplayBlock, GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 0, GridColumnEnd: 1}},
						},
					},
				},
			},
		},
	}

	constraints := Tight(400, 300)
	LayoutBlock(root, constraints, ctx)

	flexContainer := root.Children[0]
	gridContainer := flexContainer.Children[0]
	gridChild0 := gridContainer.Children[0]
	gridChild1 := gridContainer.Children[1]

	// Flex container should be positioned at (0, 0) in block parent
	if flexContainer.Rect.X != 0 || flexContainer.Rect.Y != 0 {
		t.Errorf("Flex container: expected (0, 0), got (%.2f, %.2f)",
			flexContainer.Rect.X, flexContainer.Rect.Y)
	}

	// Grid container is in vertical flexbox (row = vertical main axis)
	// Should be positioned at (0, 0) within flex container
	if gridContainer.Rect.X != 0 || gridContainer.Rect.Y != 0 {
		t.Errorf("Grid container: expected (0, 0), got (%.2f, %.2f)",
			gridContainer.Rect.X, gridContainer.Rect.Y)
	}

	// Grid children in horizontal-tb mode: rows control Y
	// Child 0 at row 0, child 1 at row 1
	if gridChild0.Rect.Y != 0 {
		t.Errorf("Grid child 0 Y: expected 0, got %.2f", gridChild0.Rect.Y)
	}
	if gridChild1.Rect.Y <= gridChild0.Rect.Y {
		t.Errorf("Grid child 1 Y (%.2f) should be > Grid child 0 Y (%.2f)",
			gridChild1.Rect.Y, gridChild0.Rect.Y)
	}
}

// TestFlexboxColumnReverseVertical tests column-reverse in vertical writing mode
func TestFlexboxColumnReverseVertical(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumnReverse, // Main axis horizontal, reversed
			Width:         Px(300),
			Height:        Px(200),
			WritingMode:   WritingModeVerticalLR,
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(100), Height: Px(50), FlexShrink: 0}},
			{Style: Style{Display: DisplayBlock, Width: Px(100), Height: Px(50), FlexShrink: 0}},
		},
	}

	constraints := Tight(300, 200)
	LayoutFlexbox(root, constraints, ctx)

	child0 := root.Children[0]
	child1 := root.Children[1]

	// In column-reverse with vertical-lr, main axis is horizontal (reversed)
	// Child 0 should be to the right of Child 1
	if child0.Rect.X <= child1.Rect.X {
		t.Errorf("Child 0 X (%.2f) should be > Child 1 X (%.2f) for column-reverse",
			child0.Rect.X, child1.Rect.X)
	}

	// Both should have same Y (cross axis)
	if child0.Rect.Y != child1.Rect.Y {
		t.Errorf("Children should have same Y, got child0=%.2f, child1=%.2f",
			child0.Rect.Y, child1.Rect.Y)
	}
}

// TestBlockPaddingVertical tests padding in vertical writing mode
func TestBlockPaddingVertical(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	root := &Node{
		Style: Style{
			Display:     DisplayBlock,
			Width:       Px(200),
			Height:      Px(150),
			Padding:     Uniform(Px(10)),
			WritingMode: WritingModeVerticalLR,
		},
		Children: []*Node{
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(50),
					Height:  Px(100),
				},
			},
		},
	}

	constraints := Tight(200, 150)
	LayoutBlock(root, constraints, ctx)

	child := root.Children[0]

	// In vertical-lr, child should be positioned at padding-left, padding-top
	expectedX := 10.0
	expectedY := 10.0
	if child.Rect.X != expectedX {
		t.Errorf("Child X: expected %.2f, got %.2f", expectedX, child.Rect.X)
	}
	if child.Rect.Y != expectedY {
		t.Errorf("Child Y: expected %.2f, got %.2f", expectedY, child.Rect.Y)
	}
}

// TestGridAutoPlacementVertical tests grid auto-placement in vertical writing mode
func TestGridAutoPlacementVertical(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			Width:               Px(200),
			Height:              Px(100),
			WritingMode:         WritingModeVerticalLR,
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock}}, // Auto-placed
			{Style: Style{Display: DisplayBlock}}, // Auto-placed
			{Style: Style{Display: DisplayBlock}}, // Auto-placed
		},
	}

	constraints := Tight(200, 100)
	LayoutGrid(root, constraints, ctx)

	// Auto-placement should fill grid cells in order
	// With 2 rows x 2 columns = 4 cells, 3 children
	child0 := root.Children[0]
	child2 := root.Children[2]

	// In vertical-lr with row-major auto-flow (default):
	// Cell (0,0): child0
	// Cell (0,1): child1
	// Cell (1,0): child2

	// Verify they have different positions
	if child0.Rect.X == child2.Rect.X && child0.Rect.Y == child2.Rect.Y {
		t.Errorf("Child 0 and Child 2 should not be at same position")
	}

	// All children should have non-zero dimensions
	for i, child := range root.Children {
		if child.Rect.Width == 0 || child.Rect.Height == 0 {
			t.Errorf("Child %d has zero dimension: %.2fx%.2f", i, child.Rect.Width, child.Rect.Height)
		}
	}
}
