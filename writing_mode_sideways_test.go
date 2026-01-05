package layout

import (
	"testing"
)

// TestBlockLayoutSidewaysRL verifies that block layout works correctly in sideways-rl mode
// Sideways-rl has the same layout behavior as vertical-rl (blocks progress right-to-left)
// but differs in character orientation (all characters rotated 90° clockwise)
func TestBlockLayoutSidewaysRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a block container with sideways-rl writing mode
	root := &Node{
		Style: Style{
			Display:     DisplayBlock,
			Width:       Px(300),
			Height:      Px(200),
			WritingMode: WritingModeSidewaysRL,
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

	// In sideways-rl mode:
	// - Block axis is horizontal, right-to-left
	// - Inline axis is vertical, top-to-bottom
	// - Children should stack right-to-left (same as vertical-rl)
	child1 := root.Children[0]
	child2 := root.Children[1]
	child3 := root.Children[2]

	// Check that children stack right-to-left
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

// TestBlockLayoutSidewaysLR verifies that block layout works correctly in sideways-lr mode
// Sideways-lr has the same layout behavior as vertical-lr (blocks progress left-to-right)
// but differs in character orientation (all characters rotated 90° counter-clockwise)
func TestBlockLayoutSidewaysLR(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a block container with sideways-lr writing mode
	root := &Node{
		Style: Style{
			Display:     DisplayBlock,
			Width:       Px(300),
			Height:      Px(200),
			WritingMode: WritingModeSidewaysLR,
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

	// In sideways-lr mode:
	// - Block axis is horizontal, left-to-right
	// - Inline axis is vertical, top-to-bottom
	// - Children should stack left-to-right (same as vertical-lr)
	child1 := root.Children[0]
	child2 := root.Children[1]
	child3 := root.Children[2]

	// Check that children stack left-to-right
	if child1.Rect.X != 0 {
		t.Errorf("Child 1 X: expected 0, got %.2f", child1.Rect.X)
	}
	if child2.Rect.X != 50 {
		t.Errorf("Child 2 X: expected 50, got %.2f", child2.Rect.X)
	}
	if child3.Rect.X != 110 {
		t.Errorf("Child 3 X: expected 110, got %.2f", child3.Rect.X)
	}

	// Check that Y positions are all the same
	if child1.Rect.Y != 0 || child2.Rect.Y != 0 || child3.Rect.Y != 0 {
		t.Errorf("Children should have same Y position (0), got: child1=%.2f, child2=%.2f, child3=%.2f",
			child1.Rect.Y, child2.Rect.Y, child3.Rect.Y)
	}
}

// TestFlexboxColumnSidewaysRL verifies that flex-direction: column in sideways-rl mode
// creates a horizontal main axis that progresses right to left
func TestFlexboxColumnSidewaysRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with sideways-rl writing mode
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn, // column = block direction
			Width:         Px(300),
			Height:        Px(200),
			WritingMode:   WritingModeSidewaysRL,
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

	// In sideways-rl mode with flex-direction: column:
	// - Main axis is block direction = horizontal (right-to-left)
	// - Items should stack right to left
	child1 := root.Children[0]
	child2 := root.Children[1]
	child3 := root.Children[2]

	// Check that children stack right-to-left
	if child1.Rect.X != 250 {
		t.Errorf("Child 1 X: expected 250, got %.2f", child1.Rect.X)
	}
	if child2.Rect.X != 190 {
		t.Errorf("Child 2 X: expected 190, got %.2f", child2.Rect.X)
	}
	if child3.Rect.X != 120 {
		t.Errorf("Child 3 X: expected 120, got %.2f", child3.Rect.X)
	}

	// Check that Y positions are all the same
	if child1.Rect.Y != 0 || child2.Rect.Y != 0 || child3.Rect.Y != 0 {
		t.Errorf("Children should have same Y position (0), got: child1=%.2f, child2=%.2f, child3=%.2f",
			child1.Rect.Y, child2.Rect.Y, child3.Rect.Y)
	}
}

// TestFlexboxColumnSidewaysLR verifies that flex-direction: column in sideways-lr mode
// creates a horizontal main axis that progresses left to right
func TestFlexboxColumnSidewaysLR(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with sideways-lr writing mode
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn, // column = block direction
			Width:         Px(300),
			Height:        Px(200),
			WritingMode:   WritingModeSidewaysLR,
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

	// In sideways-lr mode with flex-direction: column:
	// - Main axis is block direction = horizontal (left-to-right)
	// - Items should stack left to right
	child1 := root.Children[0]
	child2 := root.Children[1]
	child3 := root.Children[2]

	// Check that children stack left-to-right
	if child1.Rect.X != 0 {
		t.Errorf("Child 1 X: expected 0, got %.2f", child1.Rect.X)
	}
	if child2.Rect.X != 50 {
		t.Errorf("Child 2 X: expected 50, got %.2f", child2.Rect.X)
	}
	if child3.Rect.X != 110 {
		t.Errorf("Child 3 X: expected 110, got %.2f", child3.Rect.X)
	}

	// Check that Y positions are all the same
	if child1.Rect.Y != 0 || child2.Rect.Y != 0 || child3.Rect.Y != 0 {
		t.Errorf("Children should have same Y position (0), got: child1=%.2f, child2=%.2f, child3=%.2f",
			child1.Rect.Y, child2.Rect.Y, child3.Rect.Y)
	}
}

// TestFlexboxRowSidewaysRL verifies that flex-direction: row in sideways-rl mode
// creates a vertical main axis (items stack top to bottom)
func TestFlexboxRowSidewaysRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with sideways-rl writing mode
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow, // row = inline direction
			Width:         Px(300),
			Height:        Px(300),
			WritingMode:   WritingModeSidewaysRL,
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

	// In sideways-rl mode with flex-direction: row:
	// - Main axis is inline direction = vertical (Y increases)
	// - Items should stack top to bottom
	child1 := root.Children[0]
	child2 := root.Children[1]
	child3 := root.Children[2]

	// Check that children stack vertically
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

// TestFlexboxRowSidewaysLR verifies that flex-direction: row in sideways-lr mode
// creates a vertical main axis (items stack top to bottom)
func TestFlexboxRowSidewaysLR(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with sideways-lr writing mode
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow, // row = inline direction
			Width:         Px(300),
			Height:        Px(300),
			WritingMode:   WritingModeSidewaysLR,
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

	// In sideways-lr mode with flex-direction: row:
	// - Main axis is inline direction = vertical (Y increases)
	// - Items should stack top to bottom
	child1 := root.Children[0]
	child2 := root.Children[1]
	child3 := root.Children[2]

	// Check that children stack vertically
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

// TestGridSidewaysRL verifies that grid layout works correctly in sideways-rl mode
func TestGridSidewaysRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a grid container with sideways-rl writing mode
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			Width:               Px(200),
			Height:              Px(200),
			WritingMode:         WritingModeSidewaysRL,
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

	// In sideways-rl mode (same as vertical-rl):
	// - Rows control horizontal positioning (X axis), right-to-left
	// - Columns control vertical positioning (Y axis)
	child0 := root.Children[0]
	child1 := root.Children[1]
	child2 := root.Children[2]
	child3 := root.Children[3]

	// Check positioning
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

// TestGridSidewaysLR verifies that grid layout works correctly in sideways-lr mode
func TestGridSidewaysLR(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a grid container with sideways-lr writing mode
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			Width:               Px(200),
			Height:              Px(200),
			WritingMode:         WritingModeSidewaysLR,
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

	// In sideways-lr mode (same as vertical-lr):
	// - Rows control horizontal positioning (X axis), left-to-right
	// - Columns control vertical positioning (Y axis)
	child0 := root.Children[0]
	child1 := root.Children[1]
	child2 := root.Children[2]
	child3 := root.Children[3]

	// Check positioning
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

// TestFlexboxJustifyContentCenterSidewaysRL verifies justify-content in sideways-rl
func TestFlexboxJustifyContentCenterSidewaysRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with sideways-rl and justify-content: center
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionColumn,
			JustifyContent: JustifyContentCenter,
			Width:          Px(300),
			Height:         Px(200),
			WritingMode:    WritingModeSidewaysRL,
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

	// Main axis is horizontal, right-to-left
	// Total item width = 100, free space = 200, offset = 100
	// Items positioned from right: 300 - 100 = 200 (start)
	expectedX1 := 150.0
	expectedX2 := 100.0

	if child1.Rect.X != expectedX1 {
		t.Errorf("Child 1 X: expected %.2f, got %.2f", expectedX1, child1.Rect.X)
	}
	if child2.Rect.X != expectedX2 {
		t.Errorf("Child 2 X: expected %.2f, got %.2f", expectedX2, child2.Rect.X)
	}
}

// TestFlexboxJustifyContentCenterSidewaysLR verifies justify-content in sideways-lr
func TestFlexboxJustifyContentCenterSidewaysLR(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with sideways-lr and justify-content: center
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionColumn,
			JustifyContent: JustifyContentCenter,
			Width:          Px(300),
			Height:         Px(200),
			WritingMode:    WritingModeSidewaysLR,
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

	// Main axis is horizontal, left-to-right
	// Total item width = 100, free space = 200, offset = 100
	expectedX1 := 100.0
	expectedX2 := 150.0

	if child1.Rect.X != expectedX1 {
		t.Errorf("Child 1 X: expected %.2f, got %.2f", expectedX1, child1.Rect.X)
	}
	if child2.Rect.X != expectedX2 {
		t.Errorf("Child 2 X: expected %.2f, got %.2f", expectedX2, child2.Rect.X)
	}
}
