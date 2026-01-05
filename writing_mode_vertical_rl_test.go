package layout

import (
	"testing"
)

// TestBlockLayoutVerticalRL verifies that block layout works correctly in vertical-rl mode
// where blocks progress from right to left
func TestBlockLayoutVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a block container with vertical-rl writing mode
	root := &Node{
		Style: Style{
			Display:     DisplayBlock,
			Width:       Px(300),
			Height:      Px(200),
			WritingMode: WritingModeVerticalRL,
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
					Width:   Px(60),
					Height:  Px(100),
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(70),
					Height:  Px(100),
				},
			},
		},
	}

	constraints := Tight(300, 200)
	LayoutBlock(root, constraints, ctx)

	// In vertical-rl mode:
	// - Block axis is horizontal, right-to-left
	// - Inline axis is vertical, top-to-bottom
	// - Children should stack right-to-left (X decreases)
	child1 := root.Children[0]
	child2 := root.Children[1]
	child3 := root.Children[2]

	// Check that children stack right-to-left
	// Content width is 300px
	// Child 1 (width 50) should be at X = 300 - 50 = 250
	// Child 2 (width 60) should be at X = 300 - 50 - 60 = 190
	// Child 3 (width 70) should be at X = 300 - 50 - 60 - 70 = 120
	if child1.Rect.X != 250 {
		t.Errorf("Child 1 X: expected 250, got %.2f", child1.Rect.X)
	}
	if child2.Rect.X != 190 {
		t.Errorf("Child 2 X: expected 190, got %.2f", child2.Rect.X)
	}
	if child3.Rect.X != 120 {
		t.Errorf("Child 3 X: expected 120, got %.2f", child3.Rect.X)
	}

	// Check that Y positions are all the same (inline start)
	if child1.Rect.Y != 0 || child2.Rect.Y != 0 || child3.Rect.Y != 0 {
		t.Errorf("Children should have same Y position (0), got: child1=%.2f, child2=%.2f, child3=%.2f",
			child1.Rect.Y, child2.Rect.Y, child3.Rect.Y)
	}
}

// TestFlexboxColumnVerticalRL verifies that flex-direction: column in vertical-rl mode
// creates a horizontal main axis that progresses right to left
func TestFlexboxColumnVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with vertical-rl writing mode
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn, // column = block direction
			Width:         Px(300),
			Height:        Px(200),
			WritingMode:   WritingModeVerticalRL,
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
					Width:   Px(60),
					Height:  Px(100),
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(70),
					Height:  Px(100),
				},
			},
		},
	}

	constraints := Tight(300, 200)
	LayoutFlexbox(root, constraints, ctx)

	// In vertical-rl mode with flex-direction: column:
	// - Main axis is block direction = horizontal (right-to-left)
	// - Cross axis is inline direction = vertical
	// - Items should stack right to left (X decreases)
	child1 := root.Children[0]
	child2 := root.Children[1]
	child3 := root.Children[2]

	// Check that children stack right-to-left (X decreases)
	// Child 1 should be at X = 300 - 50 = 250
	// Child 2 should be at X = 300 - 50 - 60 = 190
	// Child 3 should be at X = 300 - 50 - 60 - 70 = 120
	if child1.Rect.X != 250 {
		t.Errorf("Child 1 X: expected 250, got %.2f", child1.Rect.X)
	}
	if child2.Rect.X != 190 {
		t.Errorf("Child 2 X: expected 190, got %.2f", child2.Rect.X)
	}
	if child3.Rect.X != 120 {
		t.Errorf("Child 3 X: expected 120, got %.2f", child3.Rect.X)
	}

	// Check that Y positions are all the same (children don't stack vertically)
	if child1.Rect.Y != 0 || child2.Rect.Y != 0 || child3.Rect.Y != 0 {
		t.Errorf("Children should have same Y position (0), got: child1=%.2f, child2=%.2f, child3=%.2f",
			child1.Rect.Y, child2.Rect.Y, child3.Rect.Y)
	}
}

// TestFlexboxRowVerticalRL verifies that flex-direction: row in vertical-rl mode
// creates a vertical main axis (items stack top to bottom, same as vertical-lr)
func TestFlexboxRowVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with vertical-rl writing mode
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow, // row = inline direction
			Width:         Px(300),
			Height:        Px(300),
			WritingMode:   WritingModeVerticalRL,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:    DisplayBlock,
					Width:      Px(50),
					Height:     Px(100),
					FlexShrink: 0,
				},
			},
			{
				Style: Style{
					Display:    DisplayBlock,
					Width:      Px(60),
					Height:     Px(100),
					FlexShrink: 0,
				},
			},
			{
				Style: Style{
					Display:    DisplayBlock,
					Width:      Px(70),
					Height:     Px(100),
					FlexShrink: 0,
				},
			},
		},
	}

	constraints := Tight(300, 300)
	LayoutFlexbox(root, constraints, ctx)

	// In vertical-rl mode with flex-direction: row:
	// - Main axis is inline direction = vertical (Y increases, same as vertical-lr)
	// - Cross axis is block direction = horizontal
	// - Items should stack top to bottom (Y increases)
	child1 := root.Children[0]
	child2 := root.Children[1]
	child3 := root.Children[2]

	// Check that children stack vertically (Y increases)
	if child1.Rect.Y != 0 {
		t.Errorf("Child 1 Y: expected 0, got %.2f", child1.Rect.Y)
	}
	if child2.Rect.Y != 100 {
		t.Errorf("Child 2 Y: expected 100, got %.2f", child2.Rect.Y)
	}
	if child3.Rect.Y != 200 {
		t.Errorf("Child 3 Y: expected 200, got %.2f", child3.Rect.Y)
	}

	// Check that X positions are all the same
	if child1.Rect.X != 0 || child2.Rect.X != 0 || child3.Rect.X != 0 {
		t.Errorf("Children should have same X position (0), got: child1=%.2f, child2=%.2f, child3=%.2f",
			child1.Rect.X, child2.Rect.X, child3.Rect.X)
	}
}

// TestGridVerticalRL verifies that grid layout works correctly in vertical-rl mode
func TestGridVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a grid container with vertical-rl writing mode
	// 2 rows (which in vertical-rl control horizontal positioning, right-to-left)
	// 2 columns (which in vertical-rl control vertical positioning)
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			Width:               Px(200),
			Height:              Px(200),
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
					GridRowStart:    0,
					GridRowEnd:      1,
					GridColumnStart: 1,
					GridColumnEnd:   2,
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
					GridRowStart:    1,
					GridRowEnd:      2,
					GridColumnStart: 1,
					GridColumnEnd:   2,
				},
			},
		},
	}

	constraints := Tight(200, 200)
	LayoutGrid(root, constraints, ctx)

	// In vertical-rl mode:
	// - Rows control horizontal positioning (X axis), right-to-left
	// - Columns control vertical positioning (Y axis)
	//
	// Grid should be (from user's perspective):
	// Row 0 (X=200-100=100 to 200): Col 0 (Y=0-50): child0   Col 1 (Y=50-100): child1
	// Row 1 (X=0-100):              Col 0 (Y=0-50): child2   Col 1 (Y=50-100): child3
	child0 := root.Children[0] // Row 0, Col 0
	child1 := root.Children[1] // Row 0, Col 1
	child2 := root.Children[2] // Row 1, Col 0
	child3 := root.Children[3] // Row 1, Col 1

	// Check child0 (row 0, col 0): X should be 100 (rightmost), Y should be 0
	if child0.Rect.X != 100 {
		t.Errorf("Child 0 X: expected 100, got %.2f", child0.Rect.X)
	}
	if child0.Rect.Y != 0 {
		t.Errorf("Child 0 Y: expected 0, got %.2f", child0.Rect.Y)
	}

	// Check child1 (row 0, col 1): X should be 100, Y should be 50
	if child1.Rect.X != 100 {
		t.Errorf("Child 1 X: expected 100, got %.2f", child1.Rect.X)
	}
	if child1.Rect.Y != 50 {
		t.Errorf("Child 1 Y: expected 50, got %.2f", child1.Rect.Y)
	}

	// Check child2 (row 1, col 0): X should be 0 (leftmost), Y should be 0
	if child2.Rect.X != 0 {
		t.Errorf("Child 2 X: expected 0, got %.2f", child2.Rect.X)
	}
	if child2.Rect.Y != 0 {
		t.Errorf("Child 2 Y: expected 0, got %.2f", child2.Rect.Y)
	}

	// Check child3 (row 1, col 1): X should be 0, Y should be 50
	if child3.Rect.X != 0 {
		t.Errorf("Child 3 X: expected 0, got %.2f", child3.Rect.X)
	}
	if child3.Rect.Y != 50 {
		t.Errorf("Child 3 Y: expected 50, got %.2f", child3.Rect.Y)
	}

	// Check sizes (should all be 100x50 in vertical mode, which is row x column size)
	expectedWidth := 100.0  // Row size
	expectedHeight := 50.0  // Column size
	if child0.Rect.Width != expectedWidth || child0.Rect.Height != expectedHeight {
		t.Errorf("Child 0 size: expected %.2fx%.2f, got %.2fx%.2f",
			expectedWidth, expectedHeight, child0.Rect.Width, child0.Rect.Height)
	}
}

// TestFlexboxJustifyContentCenterVerticalRL verifies justify-content: center in vertical-rl
func TestFlexboxJustifyContentCenterVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with vertical-rl and justify-content: center
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionColumn, // Main axis = horizontal (right-to-left)
			JustifyContent: JustifyContentCenter,
			Width:          Px(300),
			Height:         Px(200),
			WritingMode:    WritingModeVerticalRL,
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
	}

	constraints := Tight(300, 200)
	LayoutFlexbox(root, constraints, ctx)

	child1 := root.Children[0]
	child2 := root.Children[1]

	// Main axis is horizontal (width = 300), right-to-left
	// Total item width = 50 + 50 = 100
	// Free space = 300 - 100 = 200
	// Center alignment: offset = 200 / 2 = 100
	// Items positioned from right: 300 - 100 = 200 (start position)
	// Child 1: X = 200 - 50 = 150
	// Child 2: X = 150 - 50 = 100
	expectedX1 := 150.0
	expectedX2 := 100.0

	if child1.Rect.X != expectedX1 {
		t.Errorf("Child 1 X: expected %.2f, got %.2f", expectedX1, child1.Rect.X)
	}
	if child2.Rect.X != expectedX2 {
		t.Errorf("Child 2 X: expected %.2f, got %.2f", expectedX2, child2.Rect.X)
	}

	// Cross axis (Y) should be at 0
	if child1.Rect.Y != 0 || child2.Rect.Y != 0 {
		t.Errorf("Children Y should be 0, got: child1=%.2f, child2=%.2f", child1.Rect.Y, child2.Rect.Y)
	}
}

// TestBlockLayoutWithMarginsVerticalRL verifies margin handling in vertical-rl
func TestBlockLayoutWithMarginsVerticalRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a block container with vertical-rl writing mode and margins
	root := &Node{
		Style: Style{
			Display:     DisplayBlock,
			Width:       Px(300),
			Height:      Px(200),
			WritingMode: WritingModeVerticalRL,
		},
		Children: []*Node{
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(50),
					Height:  Px(100),
					Margin:  Spacing{Right: Px(10), Left: Px(20)}, // Right is start, Left is end in vertical-rl
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(60),
					Height:  Px(100),
					Margin:  Spacing{Right: Px(15), Left: Px(25)},
				},
			},
		},
	}

	constraints := Tight(300, 200)
	LayoutBlock(root, constraints, ctx)

	child1 := root.Children[0]
	child2 := root.Children[1]

	// In vertical-rl, right margin is block-start, left margin is block-end
	// Child 1:
	//   - currentBlockPos starts at 10 (right margin, block-start)
	//   - X = 300 - 10 - 50 = 240
	//   - After: currentBlockPos = 10 + 50 + 20 = 80
	// Child 2:
	//   - Margin collapse: max(20, 15) = 20
	//   - currentBlockPos = 80 - 20 + 20 = 80 (no change due to collapse)
	//   - X = 300 - 80 - 60 = 160
	if child1.Rect.X != 240 {
		t.Errorf("Child 1 X: expected 240, got %.2f", child1.Rect.X)
	}
	if child2.Rect.X != 160 {
		t.Errorf("Child 2 X: expected 160, got %.2f", child2.Rect.X)
	}
}
