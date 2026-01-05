package layout

import (
	"testing"
)

// Edge case tests for vertical-rl and sideways writing modes

// TestFlexboxWrapVerticalRL tests flexbox wrapping in vertical-rl mode
func TestFlexboxWrapVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with vertical-rl and wrap
	// In vertical-rl with column (main=horizontal right-to-left), items should wrap to new rows
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn, // Main axis horizontal
			FlexWrap:      FlexWrapWrap,
			Width:         Px(150), // Main axis size
			Height:        Px(200),
			WritingMode:   WritingModeVerticalRL,
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(100), Height: Px(50), FlexShrink: 0}},
			{Style: Style{Display: DisplayBlock, Width: Px(100), Height: Px(50), FlexShrink: 0}},
			{Style: Style{Display: DisplayBlock, Width: Px(100), Height: Px(50), FlexShrink: 0}},
		},
	}

	constraints := Tight(150, 200)
	LayoutFlexbox(root, constraints, ctx)

	// With main axis=150, items width=100 each
	// Items wrap because 100+100 > 150
	// Line 1: Item 0 (starts from right, X = 150 - 100 = 50)
	// Line 2: Item 1 (new row, Y increased, X = 150 - 100 = 50)
	// Line 3: Item 2 (new row, Y increased more)

	child0 := root.Children[0]
	child1 := root.Children[1]
	child2 := root.Children[2]

	// All items should have same X position (right-aligned in main axis)
	if child0.Rect.X != 50 {
		t.Errorf("Child 0 X: expected 50, got %.2f", child0.Rect.X)
	}
	if child1.Rect.X != 50 {
		t.Errorf("Child 1 X: expected 50, got %.2f", child1.Rect.X)
	}
	if child2.Rect.X != 50 {
		t.Errorf("Child 2 X: expected 50, got %.2f", child2.Rect.X)
	}

	// Items should have different Y positions (wrapped to new lines)
	if child0.Rect.Y != 0 {
		t.Errorf("Child 0 Y: expected 0, got %.2f", child0.Rect.Y)
	}
	if child1.Rect.Y <= child0.Rect.Y {
		t.Errorf("Child 1 should be below Child 0, got Y: child0=%.2f, child1=%.2f",
			child0.Rect.Y, child1.Rect.Y)
	}
	if child2.Rect.Y <= child1.Rect.Y {
		t.Errorf("Child 2 should be below Child 1, got Y: child1=%.2f, child2=%.2f",
			child1.Rect.Y, child2.Rect.Y)
	}
}

// TestFlexboxGapVerticalRL tests flexbox gaps in vertical-rl mode
func TestFlexboxGapVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with vertical-rl and gaps
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn, // Main axis horizontal (right-to-left)
			FlexGap:       Px(10),
			Width:         Px(300),
			Height:        Px(200),
			WritingMode:   WritingModeVerticalRL,
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(100)}},
			{Style: Style{Display: DisplayBlock, Width: Px(60), Height: Px(100)}},
			{Style: Style{Display: DisplayBlock, Width: Px(70), Height: Px(100)}},
		},
	}

	constraints := Tight(300, 200)
	LayoutFlexbox(root, constraints, ctx)

	child0 := root.Children[0]
	child1 := root.Children[1]
	child2 := root.Children[2]

	// In vertical-rl with column direction:
	// Items stack right-to-left with gaps between
	// Child 0: X = 300 - 50 = 250
	// Child 1: X = 300 - 50 - 10 (gap) - 60 = 180
	// Child 2: X = 300 - 50 - 10 - 60 - 10 (gap) - 70 = 100

	if child0.Rect.X != 250 {
		t.Errorf("Child 0 X: expected 250, got %.2f", child0.Rect.X)
	}
	if child1.Rect.X != 180 {
		t.Errorf("Child 1 X: expected 180, got %.2f", child1.Rect.X)
	}
	if child2.Rect.X != 100 {
		t.Errorf("Child 2 X: expected 100, got %.2f", child2.Rect.X)
	}

	// Verify gap between items
	gap1 := child0.Rect.X - (child1.Rect.X + child1.Rect.Width)
	gap2 := child1.Rect.X - (child2.Rect.X + child2.Rect.Width)

	if gap1 != 10 {
		t.Errorf("Gap between child 0 and 1: expected 10, got %.2f", gap1)
	}
	if gap2 != 10 {
		t.Errorf("Gap between child 1 and 2: expected 10, got %.2f", gap2)
	}
}

// TestFlexboxColumnReverseVerticalRL tests column-reverse in vertical-rl mode
func TestFlexboxColumnReverseVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// In vertical-rl with column-reverse:
	// Main axis is horizontal, reversed means left-to-right (opposite of normal right-to-left)
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumnReverse,
			Width:         Px(300),
			Height:        Px(200),
			WritingMode:   WritingModeVerticalRL,
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(100)}},
			{Style: Style{Display: DisplayBlock, Width: Px(60), Height: Px(100)}},
			{Style: Style{Display: DisplayBlock, Width: Px(70), Height: Px(100)}},
		},
	}

	constraints := Tight(300, 200)
	LayoutFlexbox(root, constraints, ctx)

	child0 := root.Children[0]
	child1 := root.Children[1]
	child2 := root.Children[2]

	// With column-reverse in vertical-rl, items should stack left-to-right
	// Child 0: X = 0
	// Child 1: X = 50
	// Child 2: X = 110
	if child0.Rect.X != 0 {
		t.Errorf("Child 0 X: expected 0, got %.2f", child0.Rect.X)
	}
	if child1.Rect.X != 50 {
		t.Errorf("Child 1 X: expected 50, got %.2f", child1.Rect.X)
	}
	if child2.Rect.X != 110 {
		t.Errorf("Child 2 X: expected 110, got %.2f", child2.Rect.X)
	}
}

// TestFlexboxRowReverseVerticalRL tests row-reverse in vertical-rl mode
func TestFlexboxRowReverseVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// In vertical-rl with row-reverse:
	// Main axis is vertical (inline direction), reversed means bottom-to-top
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRowReverse,
			Width:         Px(300),
			Height:        Px(300),
			WritingMode:   WritingModeVerticalRL,
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(100), FlexShrink: 0}},
			{Style: Style{Display: DisplayBlock, Width: Px(60), Height: Px(100), FlexShrink: 0}},
			{Style: Style{Display: DisplayBlock, Width: Px(70), Height: Px(100), FlexShrink: 0}},
		},
	}

	constraints := Tight(300, 300)
	LayoutFlexbox(root, constraints, ctx)

	child0 := root.Children[0]
	child1 := root.Children[1]
	child2 := root.Children[2]

	// With row-reverse, items should stack bottom-to-top
	// Child 0: Y = 300 - 100 = 200
	// Child 1: Y = 200 - 100 = 100
	// Child 2: Y = 100 - 100 = 0
	if child0.Rect.Y != 200 {
		t.Errorf("Child 0 Y: expected 200, got %.2f", child0.Rect.Y)
	}
	if child1.Rect.Y != 100 {
		t.Errorf("Child 1 Y: expected 100, got %.2f", child1.Rect.Y)
	}
	if child2.Rect.Y != 0 {
		t.Errorf("Child 2 Y: expected 0, got %.2f", child2.Rect.Y)
	}
}

// TestGridSpanningVerticalRL tests grid spanning in vertical-rl mode
func TestGridSpanningVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a grid with spanning items in vertical-rl
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			Width:               Px(300),
			Height:              Px(100),
			WritingMode:         WritingModeVerticalRL,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:         DisplayBlock,
					GridRowStart:    0,
					GridRowEnd:      2, // Spans 2 rows
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

	constraints := Tight(300, 100)
	LayoutGrid(root, constraints, ctx)

	child0 := root.Children[0] // Spans rows 0-2
	child1 := root.Children[1] // Single cell

	// In vertical-rl:
	// - Rows control X (right-to-left)
	// - Columns control Y
	// Child 0 spans rows 0-2, so width should be 200 (2 rows × 100)
	// Child 0 should be at X = 300 - 200 = 100 (rightmost position)
	if child0.Rect.Width != 200 {
		t.Errorf("Child 0 width: expected 200, got %.2f", child0.Rect.Width)
	}
	if child0.Rect.X != 100 {
		t.Errorf("Child 0 X: expected 100, got %.2f", child0.Rect.X)
	}
	if child0.Rect.Y != 0 {
		t.Errorf("Child 0 Y: expected 0, got %.2f", child0.Rect.Y)
	}

	// Child 1 is in row 0, col 1
	if child1.Rect.X != 200 {
		t.Errorf("Child 1 X: expected 200, got %.2f", child1.Rect.X)
	}
	if child1.Rect.Y != 50 {
		t.Errorf("Child 1 Y: expected 50, got %.2f", child1.Rect.Y)
	}
}

// TestGridGapVerticalRL tests grid gaps in vertical-rl mode
func TestGridGapVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a grid with gaps in vertical-rl
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			GridRowGap:          Px(10), // Gap between rows (horizontal in vertical-rl)
			GridColumnGap:       Px(5),  // Gap between columns (vertical in vertical-rl)
			Width:               Px(250),
			Height:              Px(150),
			WritingMode:         WritingModeVerticalRL,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:         DisplayBlock,
					GridRowStart:    0,
					GridRowEnd:      1,
					GridColumnStart: 0,
					GridColumnEnd:   1,
				},
			},
			{
				Style: Style{
					Display:         DisplayBlock,
					GridRowStart:    1,
					GridRowEnd:      2,
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

	constraints := Tight(250, 150)
	LayoutGrid(root, constraints, ctx)

	child0 := root.Children[0] // Row 0, Col 0
	child1 := root.Children[1] // Row 1, Col 0
	child2 := root.Children[2] // Row 0, Col 1

	// In vertical-rl:
	// Row gap appears in X axis (horizontal)
	// Column gap appears in Y axis (vertical)
	// Child 0: X = 250 - 100 = 150 (rightmost), Y = 0
	// Child 1: X = 250 - 100 - 10 (gap) - 100 = 40, Y = 0
	// Child 2: X = 150, Y = 50 + 5 (gap) = 55

	if child0.Rect.X != 150 {
		t.Errorf("Child 0 X: expected 150, got %.2f", child0.Rect.X)
	}
	if child0.Rect.Y != 0 {
		t.Errorf("Child 0 Y: expected 0, got %.2f", child0.Rect.Y)
	}

	if child1.Rect.X != 40 {
		t.Errorf("Child 1 X: expected 40, got %.2f", child1.Rect.X)
	}
	if child1.Rect.Y != 0 {
		t.Errorf("Child 1 Y: expected 0, got %.2f", child1.Rect.Y)
	}

	if child2.Rect.X != 150 {
		t.Errorf("Child 2 X: expected 150, got %.2f", child2.Rect.X)
	}
	if child2.Rect.Y != 55 {
		t.Errorf("Child 2 Y: expected 55, got %.2f", child2.Rect.Y)
	}
}

// TestGridAutoPlacementVerticalRL tests grid auto-placement in vertical-rl mode
func TestGridAutoPlacementVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a grid with auto-placement in vertical-rl
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			Width:               Px(200),
			Height:              Px(100),
			WritingMode:         WritingModeVerticalRL,
			GridAutoFlow:        GridAutoFlowRow, // Default: fill rows first
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock}}, // Auto-placed
			{Style: Style{Display: DisplayBlock}}, // Auto-placed
			{Style: Style{Display: DisplayBlock}}, // Auto-placed
			{Style: Style{Display: DisplayBlock}}, // Auto-placed
		},
	}

	constraints := Tight(200, 100)
	LayoutGrid(root, constraints, ctx)

	// Grid has 2 rows × 2 columns = 4 cells
	// Items should auto-place row by row (right-to-left, then down)
	child0 := root.Children[0] // Row 0, Col 0 - rightmost, top
	child1 := root.Children[1] // Row 0, Col 1 - rightmost, bottom
	child2 := root.Children[2] // Row 1, Col 0 - left, top
	child3 := root.Children[3] // Row 1, Col 1 - left, bottom

	// In vertical-rl with 2 rows, 2 columns:
	// Row 0 is rightmost (X = 100-200), Row 1 is left (X = 0-100)
	// Col 0 is top (Y = 0-50), Col 1 is bottom (Y = 50-100)

	if child0.Rect.X != 100 {
		t.Errorf("Child 0 X: expected 100, got %.2f", child0.Rect.X)
	}
	if child0.Rect.Y != 0 {
		t.Errorf("Child 0 Y: expected 0, got %.2f", child0.Rect.Y)
	}

	if child1.Rect.X != 100 {
		t.Errorf("Child 1 X: expected 100, got %.2f", child1.Rect.X)
	}
	if child1.Rect.Y != 50 {
		t.Errorf("Child 1 Y: expected 50, got %.2f", child1.Rect.Y)
	}

	if child2.Rect.X != 0 {
		t.Errorf("Child 2 X: expected 0, got %.2f", child2.Rect.X)
	}
	if child2.Rect.Y != 0 {
		t.Errorf("Child 2 Y: expected 0, got %.2f", child2.Rect.Y)
	}

	if child3.Rect.X != 0 {
		t.Errorf("Child 3 X: expected 0, got %.2f", child3.Rect.X)
	}
	if child3.Rect.Y != 50 {
		t.Errorf("Child 3 Y: expected 50, got %.2f", child3.Rect.Y)
	}
}

// TestNestedMixedWritingModesVerticalRL tests nested containers with vertical-rl and horizontal-tb
func TestNestedMixedWritingModesVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Vertical-RL container with horizontal-tb child
	root := &Node{
		Style: Style{
			Display:     DisplayFlex,
			FlexDirection: FlexDirectionColumn,
			Width:       Px(300),
			Height:      Px(200),
			WritingMode: WritingModeVerticalRL,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:       DisplayFlex,
					FlexDirection: FlexDirectionRow,
					Width:         Px(100),
					Height:        Px(200),
					WritingMode:   WritingModeHorizontalTB, // Switches back to horizontal
				},
				Children: []*Node{
					{
						Style: Style{
							Display: DisplayBlock,
							Width:   Px(50),
							Height:  Px(100),
						},
					},
					{
						Style: Style{
							Display: DisplayBlock,
							Width:   Px(50),
							Height:  Px(100),
						},
					},
				},
			},
		},
	}

	constraints := Tight(300, 200)
	LayoutFlexbox(root, constraints, ctx)

	child := root.Children[0]
	grandchild0 := child.Children[0]
	grandchild1 := child.Children[1]

	// Child should be positioned at right edge (vertical-rl)
	if child.Rect.X != 200 {
		t.Errorf("Child X: expected 200, got %.2f", child.Rect.X)
	}

	// Grandchildren should stack horizontally left-to-right (horizontal-tb)
	if grandchild0.Rect.X != 0 {
		t.Errorf("Grandchild 0 X: expected 0, got %.2f", grandchild0.Rect.X)
	}
	if grandchild1.Rect.X != 50 {
		t.Errorf("Grandchild 1 X: expected 50, got %.2f", grandchild1.Rect.X)
	}
}

// Sideways mode edge case tests

// TestFlexboxWrapSidewaysRL tests flexbox wrapping in sideways-rl mode
func TestFlexboxWrapSidewaysRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Sideways-RL has same layout as vertical-rl
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
			FlexWrap:      FlexWrapWrap,
			Width:         Px(150),
			Height:        Px(200),
			WritingMode:   WritingModeSidewaysRL,
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(100), Height: Px(50), FlexShrink: 0}},
			{Style: Style{Display: DisplayBlock, Width: Px(100), Height: Px(50), FlexShrink: 0}},
			{Style: Style{Display: DisplayBlock, Width: Px(100), Height: Px(50), FlexShrink: 0}},
		},
	}

	constraints := Tight(150, 200)
	LayoutFlexbox(root, constraints, ctx)

	child0 := root.Children[0]
	child1 := root.Children[1]
	child2 := root.Children[2]

	// All items should have same X position (right-aligned)
	if child0.Rect.X != 50 {
		t.Errorf("Child 0 X: expected 50, got %.2f", child0.Rect.X)
	}

	// Items should wrap to new lines (Y increases)
	if child0.Rect.Y != 0 {
		t.Errorf("Child 0 Y: expected 0, got %.2f", child0.Rect.Y)
	}
	if child1.Rect.Y <= child0.Rect.Y {
		t.Errorf("Child 1 should be below Child 0")
	}
	if child2.Rect.Y <= child1.Rect.Y {
		t.Errorf("Child 2 should be below Child 1")
	}
}

// TestFlexboxGapSidewaysLR tests flexbox gaps in sideways-lr mode
func TestFlexboxGapSidewaysLR(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Sideways-LR has same layout as vertical-lr
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
			FlexGap:       Px(10),
			Width:         Px(300),
			Height:        Px(200),
			WritingMode:   WritingModeSidewaysLR,
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(100)}},
			{Style: Style{Display: DisplayBlock, Width: Px(60), Height: Px(100)}},
			{Style: Style{Display: DisplayBlock, Width: Px(70), Height: Px(100)}},
		},
	}

	constraints := Tight(300, 200)
	LayoutFlexbox(root, constraints, ctx)

	child0 := root.Children[0]
	child1 := root.Children[1]
	child2 := root.Children[2]

	// Items stack left-to-right with gaps
	if child0.Rect.X != 0 {
		t.Errorf("Child 0 X: expected 0, got %.2f", child0.Rect.X)
	}
	if child1.Rect.X != 60 {
		t.Errorf("Child 1 X: expected 60, got %.2f", child1.Rect.X)
	}
	if child2.Rect.X != 130 {
		t.Errorf("Child 2 X: expected 130, got %.2f", child2.Rect.X)
	}

	// Verify gaps
	gap1 := child1.Rect.X - (child0.Rect.X + child0.Rect.Width)
	gap2 := child2.Rect.X - (child1.Rect.X + child1.Rect.Width)

	if gap1 != 10 {
		t.Errorf("Gap between child 0 and 1: expected 10, got %.2f", gap1)
	}
	if gap2 != 10 {
		t.Errorf("Gap between child 1 and 2: expected 10, got %.2f", gap2)
	}
}

// TestGridSpanningSidewaysRL tests grid spanning in sideways-rl mode
func TestGridSpanningSidewaysRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			Width:               Px(300),
			Height:              Px(100),
			WritingMode:         WritingModeSidewaysRL,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:         DisplayBlock,
					GridRowStart:    0,
					GridRowEnd:      2, // Spans 2 rows
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

	constraints := Tight(300, 100)
	LayoutGrid(root, constraints, ctx)

	child0 := root.Children[0] // Spans rows 0-2
	child1 := root.Children[1] // Single cell

	// Same as vertical-rl
	if child0.Rect.Width != 200 {
		t.Errorf("Child 0 width: expected 200, got %.2f", child0.Rect.Width)
	}
	if child0.Rect.X != 100 {
		t.Errorf("Child 0 X: expected 100, got %.2f", child0.Rect.X)
	}
	if child1.Rect.X != 200 {
		t.Errorf("Child 1 X: expected 200, got %.2f", child1.Rect.X)
	}
	if child1.Rect.Y != 50 {
		t.Errorf("Child 1 Y: expected 50, got %.2f", child1.Rect.Y)
	}
}

// TestGridAutoPlacementSidewaysLR tests grid auto-placement in sideways-lr mode
func TestGridAutoPlacementSidewaysLR(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			Width:               Px(200),
			Height:              Px(100),
			WritingMode:         WritingModeSidewaysLR,
			GridAutoFlow:        GridAutoFlowRow,
		},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock}},
			{Style: Style{Display: DisplayBlock}},
			{Style: Style{Display: DisplayBlock}},
			{Style: Style{Display: DisplayBlock}},
		},
	}

	constraints := Tight(200, 100)
	LayoutGrid(root, constraints, ctx)

	// Same as vertical-lr: left-to-right progression
	child0 := root.Children[0] // Row 0, Col 0
	child1 := root.Children[1] // Row 0, Col 1
	child2 := root.Children[2] // Row 1, Col 0
	child3 := root.Children[3] // Row 1, Col 1

	if child0.Rect.X != 0 {
		t.Errorf("Child 0 X: expected 0, got %.2f", child0.Rect.X)
	}
	if child0.Rect.Y != 0 {
		t.Errorf("Child 0 Y: expected 0, got %.2f", child0.Rect.Y)
	}

	if child1.Rect.X != 0 {
		t.Errorf("Child 1 X: expected 0, got %.2f", child1.Rect.X)
	}
	if child1.Rect.Y != 50 {
		t.Errorf("Child 1 Y: expected 50, got %.2f", child1.Rect.Y)
	}

	if child2.Rect.X != 100 {
		t.Errorf("Child 2 X: expected 100, got %.2f", child2.Rect.X)
	}
	if child2.Rect.Y != 0 {
		t.Errorf("Child 2 Y: expected 0, got %.2f", child2.Rect.Y)
	}

	if child3.Rect.X != 100 {
		t.Errorf("Child 3 X: expected 100, got %.2f", child3.Rect.X)
	}
	if child3.Rect.Y != 50 {
		t.Errorf("Child 3 Y: expected 50, got %.2f", child3.Rect.Y)
	}
}
