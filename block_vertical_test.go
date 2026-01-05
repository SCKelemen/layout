package layout

import (
	"testing"
)

// TestBlockLayoutVerticalLR verifies that block layout stacks children horizontally in vertical-lr mode
func TestBlockLayoutVerticalLR(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a block container with vertical-lr writing mode
	root := &Node{
		Style: Style{
			Display:     DisplayBlock,
			Width:       Px(300),
			Height:      Px(200),
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

	// In vertical-lr mode, children should stack left-to-right (X increases)
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

// TestBlockLayoutHorizontalTB verifies that horizontal mode still works (baseline behavior)
func TestBlockLayoutHorizontalTB(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a block container with horizontal-tb writing mode (default)
	root := &Node{
		Style: Style{
			Display:     DisplayBlock,
			Width:       Px(300),
			Height:      Px(200),
			WritingMode: WritingModeHorizontalTB,
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
	LayoutBlock(root, constraints, ctx)

	// In horizontal mode, children should stack top-to-bottom (Y increases)
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

// TestBlockLayoutVerticalWithMargins verifies margin collapsing in vertical mode
func TestBlockLayoutVerticalWithMargins(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	root := &Node{
		Style: Style{
			Display:     DisplayBlock,
			Width:       Px(300),
			Height:      Px(200),
			WritingMode: WritingModeVerticalLR,
		},
		Children: []*Node{
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(50),
					Height:  Px(100),
					Margin:  Uniform(Px(10)),
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(60),
					Height:  Px(100),
					Margin:  Uniform(Px(15)),
				},
			},
		},
	}

	constraints := Tight(300, 200)
	LayoutBlock(root, constraints, ctx)

	child1 := root.Children[0]
	child2 := root.Children[1]

	// In vertical-lr mode, block-axis margins are left/right
	// Child 1 starts at margin-left (10)
	if child1.Rect.X != 10 {
		t.Errorf("Child 1 X: expected 10 (margin-left), got %.2f", child1.Rect.X)
	}

	// Child 2 position: child1.X + child1.Width + collapsed margin
	// Collapsed margin = max(child1.marginRight, child2.marginLeft) = max(10, 15) = 15
	expectedX := 10.0 + 50.0 + 15.0 // child1.marginLeft + child1.Width + collapsed margin
	if child2.Rect.X != expectedX {
		t.Errorf("Child 2 X: expected %.2f, got %.2f", expectedX, child2.Rect.X)
	}

	// Y positions should include inline-axis margins (top)
	if child1.Rect.Y != 10 {
		t.Errorf("Child 1 Y: expected 10 (margin-top), got %.2f", child1.Rect.Y)
	}
	if child2.Rect.Y != 15 {
		t.Errorf("Child 2 Y: expected 15 (margin-top), got %.2f", child2.Rect.Y)
	}
}

// TestBlockLayoutAutoSizeVertical verifies auto-sizing in vertical mode
// Note: Full auto-sizing for vertical modes requires additional work in blockDetermineContainerSize.
// For now, we test with explicit sizes and verify positioning.
func TestBlockLayoutAutoSizeVertical(t *testing.T) {
	t.Skip("Auto-sizing for vertical writing modes requires additional refactoring of blockDetermineContainerSize")
}

// TestBlockLayoutMixedWritingModes verifies nested containers with different writing modes
func TestBlockLayoutMixedWritingModes(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Horizontal container with vertical child
	root := &Node{
		Style: Style{
			Display:     DisplayBlock,
			Width:       Px(300),
			Height:      Px(200),
			WritingMode: WritingModeHorizontalTB,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:     DisplayBlock,
					Width:       Px(100),
					Height:      Px(50),
					WritingMode: WritingModeVerticalLR,
				},
				Children: []*Node{
					{
						Style: Style{
							Display: DisplayBlock,
							Width:   Px(20),
							Height:  Px(40),
						},
					},
					{
						Style: Style{
							Display: DisplayBlock,
							Width:   Px(30),
							Height:  Px(40),
						},
					},
				},
			},
		},
	}

	constraints := Tight(300, 200)
	LayoutBlock(root, constraints, ctx)

	child := root.Children[0]
	grandchild1 := child.Children[0]
	grandchild2 := child.Children[1]

	// Child should be at (0, 0) in the horizontal container
	if child.Rect.X != 0 || child.Rect.Y != 0 {
		t.Errorf("Child position: expected (0, 0), got (%.2f, %.2f)", child.Rect.X, child.Rect.Y)
	}

	// Grandchildren should stack horizontally within the vertical child
	if grandchild1.Rect.X != 0 {
		t.Errorf("Grandchild 1 X: expected 0, got %.2f", grandchild1.Rect.X)
	}
	if grandchild2.Rect.X != 20 { // grandchild1.Width
		t.Errorf("Grandchild 2 X: expected 20, got %.2f", grandchild2.Rect.X)
	}

	// Grandchildren Y should both be 0 (same line in vertical mode)
	if grandchild1.Rect.Y != 0 || grandchild2.Rect.Y != 0 {
		t.Errorf("Grandchildren Y: expected both 0, got gc1=%.2f, gc2=%.2f",
			grandchild1.Rect.Y, grandchild2.Rect.Y)
	}
}
