package layout

import (
	"testing"
)

// TestFlexboxRowVerticalLR verifies that flex-direction: row in vertical-lr mode
// creates a vertical main axis (items stack top to bottom)
func TestFlexboxRowVerticalLR(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with vertical-lr writing mode
	// Use flex-shrink: 0 to prevent items from shrinking
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow, // row = inline direction
			Width:         Px(300),
			Height:        Px(300), // Increased to fit all children
			WritingMode:   WritingModeVerticalLR,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:    DisplayBlock,
					Width:      Px(50),
					Height:     Px(100),
					FlexShrink: 0, // Prevent shrinking
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

	// In vertical-lr mode with flex-direction: row:
	// - Main axis is inline direction = vertical (Y increases)
	// - Cross axis is block direction = horizontal (X direction)
	// - Items should stack top to bottom (Y increases)
	child1 := root.Children[0]
	child2 := root.Children[1]
	child3 := root.Children[2]

	// Check that children stack vertically (Y increases)
	if child1.Rect.Y != 0 {
		t.Errorf("Child 1 Y: expected 0, got %.2f", child1.Rect.Y)
	}
	if child2.Rect.Y != 100 { // Child1.Height
		t.Errorf("Child 2 Y: expected 100, got %.2f", child2.Rect.Y)
	}
	if child3.Rect.Y != 200 { // Child1.Height + Child2.Height
		t.Errorf("Child 3 Y: expected 200, got %.2f", child3.Rect.Y)
	}

	// Check that X positions are all the same (children don't stack horizontally)
	if child1.Rect.X != 0 || child2.Rect.X != 0 || child3.Rect.X != 0 {
		t.Errorf("Children should have same X position (0), got: child1=%.2f, child2=%.2f, child3=%.2f",
			child1.Rect.X, child2.Rect.X, child3.Rect.X)
	}
}

// TestFlexboxColumnVerticalLR verifies that flex-direction: column in vertical-lr mode
// creates a horizontal main axis (items stack left to right)
func TestFlexboxColumnVerticalLR(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with vertical-lr writing mode
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn, // column = block direction
			Width:         Px(300),
			Height:        Px(200),
			WritingMode:   WritingModeVerticalLR,
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

	// In vertical-lr mode with flex-direction: column:
	// - Main axis is block direction = horizontal (X increases)
	// - Cross axis is inline direction = vertical (Y direction)
	// - Items should stack left to right (X increases)
	child1 := root.Children[0]
	child2 := root.Children[1]
	child3 := root.Children[2]

	// Check that children stack horizontally (X increases)
	if child1.Rect.X != 0 {
		t.Errorf("Child 1 X: expected 0, got %.2f", child1.Rect.X)
	}
	if child2.Rect.X != 50 { // Child1.Width
		t.Errorf("Child 2 X: expected 50, got %.2f", child2.Rect.X)
	}
	if child3.Rect.X != 110 { // Child1.Width + Child2.Width
		t.Errorf("Child 3 X: expected 110, got %.2f", child3.Rect.X)
	}

	// Check that Y positions are all the same (children don't stack vertically)
	if child1.Rect.Y != 0 || child2.Rect.Y != 0 || child3.Rect.Y != 0 {
		t.Errorf("Children should have same Y position (0), got: child1=%.2f, child2=%.2f, child3=%.2f",
			child1.Rect.Y, child2.Rect.Y, child3.Rect.Y)
	}
}

// TestFlexboxRowHorizontalTB verifies that horizontal mode still works (baseline behavior)
func TestFlexboxRowHorizontalTB(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with horizontal-tb writing mode (default)
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			Width:         Px(300),
			Height:        Px(200),
			WritingMode:   WritingModeHorizontalTB,
		},
		Children: []*Node{
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(100),
					Height:  Px(50),
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(100),
					Height:  Px(60),
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(100),
					Height:  Px(70),
				},
			},
		},
	}

	constraints := Tight(300, 200)
	LayoutFlexbox(root, constraints, ctx)

	// In horizontal mode with row, children should stack left-to-right (X increases)
	child1 := root.Children[0]
	child2 := root.Children[1]
	child3 := root.Children[2]

	// Check that children stack horizontally (X increases)
	if child1.Rect.X != 0 {
		t.Errorf("Child 1 X: expected 0, got %.2f", child1.Rect.X)
	}
	if child2.Rect.X != 100 { // Child1.Width
		t.Errorf("Child 2 X: expected 100, got %.2f", child2.Rect.X)
	}
	if child3.Rect.X != 200 { // Child1.Width + Child2.Width
		t.Errorf("Child 3 X: expected 200, got %.2f", child3.Rect.X)
	}

	// Check that Y positions are all the same (children don't stack vertically)
	if child1.Rect.Y != 0 || child2.Rect.Y != 0 || child3.Rect.Y != 0 {
		t.Errorf("Children should have same Y position (0), got: child1=%.2f, child2=%.2f, child3=%.2f",
			child1.Rect.Y, child2.Rect.Y, child3.Rect.Y)
	}
}

// TestFlexboxColumnHorizontalTB verifies that horizontal mode with column works
func TestFlexboxColumnHorizontalTB(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with horizontal-tb writing mode (default)
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
			Width:         Px(300),
			Height:        Px(200),
			WritingMode:   WritingModeHorizontalTB,
		},
		Children: []*Node{
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(100),
					Height:  Px(50),
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(100),
					Height:  Px(60),
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(100),
					Height:  Px(70),
				},
			},
		},
	}

	constraints := Tight(300, 200)
	LayoutFlexbox(root, constraints, ctx)

	// In horizontal mode with column, children should stack top-to-bottom (Y increases)
	child1 := root.Children[0]
	child2 := root.Children[1]
	child3 := root.Children[2]

	// Check that children stack vertically (Y increases)
	if child1.Rect.Y != 0 {
		t.Errorf("Child 1 Y: expected 0, got %.2f", child1.Rect.Y)
	}
	if child2.Rect.Y != 50 { // Child1.Height
		t.Errorf("Child 2 Y: expected 50, got %.2f", child2.Rect.Y)
	}
	if child3.Rect.Y != 110 { // Child1.Height + Child2.Height
		t.Errorf("Child 3 Y: expected 110, got %.2f", child3.Rect.Y)
	}

	// Check that X positions are all the same (children don't stack horizontally)
	if child1.Rect.X != 0 || child2.Rect.X != 0 || child3.Rect.X != 0 {
		t.Errorf("Children should have same X position (0), got: child1=%.2f, child2=%.2f, child3=%.2f",
			child1.Rect.X, child2.Rect.X, child3.Rect.X)
	}
}

// TestFlexboxJustifyContentVertical verifies justify-content in vertical writing mode
func TestFlexboxJustifyContentVertical(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with vertical-lr and justify-content: center
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow, // Main axis = vertical
			JustifyContent: JustifyContentCenter,
			Width:          Px(200),
			Height:         Px(300),
			WritingMode:    WritingModeVerticalLR,
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

	constraints := Tight(200, 300)
	LayoutFlexbox(root, constraints, ctx)

	child1 := root.Children[0]
	child2 := root.Children[1]

	// Main axis is vertical (height = 300)
	// Total item height = 100 + 100 = 200
	// Free space = 300 - 200 = 100
	// Center alignment: offset = 100 / 2 = 50
	expectedY1 := 50.0
	expectedY2 := 150.0 // 50 + 100

	if child1.Rect.Y != expectedY1 {
		t.Errorf("Child 1 Y: expected %.2f, got %.2f", expectedY1, child1.Rect.Y)
	}
	if child2.Rect.Y != expectedY2 {
		t.Errorf("Child 2 Y: expected %.2f, got %.2f", expectedY2, child2.Rect.Y)
	}

	// Cross axis (X) should be at 0
	if child1.Rect.X != 0 || child2.Rect.X != 0 {
		t.Errorf("Children X should be 0, got: child1=%.2f, child2=%.2f", child1.Rect.X, child2.Rect.X)
	}
}

// TestFlexboxAlignItemsVertical verifies align-items in vertical writing mode
func TestFlexboxAlignItemsVertical(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a flexbox container with vertical-lr and align-items: center
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow, // Main axis = vertical
			AlignItems:    AlignItemsCenter,
			Width:         Px(200),
			Height:        Px(300),
			WritingMode:   WritingModeVerticalLR,
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
		},
	}

	constraints := Tight(200, 300)
	LayoutFlexbox(root, constraints, ctx)

	child1 := root.Children[0]
	child2 := root.Children[1]

	// Cross axis is horizontal (width = 200)
	// Child 1 width = 50, center alignment: X = (200 - 50) / 2 = 75
	// Child 2 width = 60, center alignment: X = (200 - 60) / 2 = 70
	expectedX1 := 75.0
	expectedX2 := 70.0

	if child1.Rect.X != expectedX1 {
		t.Errorf("Child 1 X: expected %.2f, got %.2f", expectedX1, child1.Rect.X)
	}
	if child2.Rect.X != expectedX2 {
		t.Errorf("Child 2 X: expected %.2f, got %.2f", expectedX2, child2.Rect.X)
	}

	// Main axis (Y) should stack normally
	if child1.Rect.Y != 0 {
		t.Errorf("Child 1 Y: expected 0, got %.2f", child1.Rect.Y)
	}
	if child2.Rect.Y != 100 {
		t.Errorf("Child 2 Y: expected 100, got %.2f", child2.Rect.Y)
	}
}

// TestFlexboxMixedWritingModes verifies nested containers with different writing modes
func TestFlexboxMixedWritingModes(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Horizontal container with vertical child
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			Width:         Px(300),
			Height:        Px(200),
			WritingMode:   WritingModeHorizontalTB,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:       DisplayFlex,
					FlexDirection: FlexDirectionRow, // In vertical-lr, row = vertical
					Width:         Px(100),
					Height:        Px(200),
					WritingMode:   WritingModeVerticalLR,
				},
				Children: []*Node{
					{
						Style: Style{
							Display: DisplayBlock,
							Width:   Px(50),
							Height:  Px(80),
						},
					},
					{
						Style: Style{
							Display: DisplayBlock,
							Width:   Px(50),
							Height:  Px(80),
						},
					},
				},
			},
		},
	}

	constraints := Tight(300, 200)
	LayoutFlexbox(root, constraints, ctx)

	child := root.Children[0]
	grandchild1 := child.Children[0]
	grandchild2 := child.Children[1]

	// Child should be at (0, 0) in the horizontal container
	if child.Rect.X != 0 || child.Rect.Y != 0 {
		t.Errorf("Child position: expected (0, 0), got (%.2f, %.2f)", child.Rect.X, child.Rect.Y)
	}

	// Grandchildren should stack vertically within the vertical child (row in vertical-lr = vertical)
	if grandchild1.Rect.Y != 0 {
		t.Errorf("Grandchild 1 Y: expected 0, got %.2f", grandchild1.Rect.Y)
	}
	if grandchild2.Rect.Y != 80 { // grandchild1.Height
		t.Errorf("Grandchild 2 Y: expected 80, got %.2f", grandchild2.Rect.Y)
	}

	// Grandchildren X should both be 0 (same line in vertical mode)
	if grandchild1.Rect.X != 0 || grandchild2.Rect.X != 0 {
		t.Errorf("Grandchildren X: expected both 0, got gc1=%.2f, gc2=%.2f",
			grandchild1.Rect.X, grandchild2.Rect.X)
	}
}
