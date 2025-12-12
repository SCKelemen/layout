package layout

import (
	"testing"
)

// TestBlockMarginCollapsingAdjacentSiblings tests margin collapsing between adjacent siblings
func TestBlockMarginCollapsingAdjacentSiblings(t *testing.T) {
	// CSS Box Model §8.3.1: Adjacent vertical margins collapse
	root := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   200,
			Height:  -1, // auto
		},
		Children: []*Node{
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   100,
					Height:  50,
					Margin:  Spacing{Bottom: 20}, // First child has 20px bottom margin
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   100,
					Height:  50,
					Margin:  Spacing{Top: 30}, // Second child has 30px top margin
				},
			},
		},
	}

	constraints := Loose(200, 200)
	LayoutBlock(root, constraints)

	// First child should be at Y = 0
	if root.Children[0].Rect.Y != 0 {
		t.Errorf("First child Y: expected 0, got %.2f", root.Children[0].Rect.Y)
	}

	// Second child should be at Y = 50 (first child height) + 30 (collapsed margin, max of 20 and 30)
	// Total: 80
	expectedY := 80.0
	if root.Children[1].Rect.Y != expectedY {
		t.Errorf("Second child Y: expected %.2f (50 + max(20,30)), got %.2f", expectedY, root.Children[1].Rect.Y)
	}

	// Container height should be 50 + 30 + 50 = 130 (child1 + collapsed margin + child2)
	expectedHeight := 130.0
	if root.Rect.Height != expectedHeight {
		t.Errorf("Container height: expected %.2f, got %.2f", expectedHeight, root.Rect.Height)
	}
}

// TestBlockMarginCollapsingLargerMarginWins tests that larger margin wins when collapsing
func TestBlockMarginCollapsingLargerMarginWins(t *testing.T) {
	root := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   200,
			Height:  -1,
		},
		Children: []*Node{
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   100,
					Height:  50,
					Margin:  Spacing{Bottom: 50}, // Larger bottom margin
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   100,
					Height:  50,
					Margin:  Spacing{Top: 10}, // Smaller top margin
				},
			},
		},
	}

	constraints := Loose(200, 200)
	LayoutBlock(root, constraints)

	// Second child should be at Y = 50 + max(50, 10) = 100
	expectedY := 100.0
	if root.Children[1].Rect.Y != expectedY {
		t.Errorf("Second child Y: expected %.2f (50 + max(50,10)), got %.2f", expectedY, root.Children[1].Rect.Y)
	}
}

// TestBlockMarginCollapsingThreeChildren tests margin collapsing with three children
func TestBlockMarginCollapsingThreeChildren(t *testing.T) {
	root := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   200,
			Height:  -1,
		},
		Children: []*Node{
			{
				Style: Style{
					Display: DisplayBlock,
					Height:  30,
					Margin:  Spacing{Bottom: 15},
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Height:  30,
					Margin:  Spacing{Top: 10, Bottom: 25},
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Height:  30,
					Margin:  Spacing{Top: 20},
				},
			},
		},
	}

	constraints := Loose(200, 200)
	LayoutBlock(root, constraints)

	// First child at Y = 0
	if root.Children[0].Rect.Y != 0 {
		t.Errorf("First child Y: expected 0, got %.2f", root.Children[0].Rect.Y)
	}

	// Second child at Y = 30 + max(15, 10) = 45
	expectedY2 := 45.0
	if root.Children[1].Rect.Y != expectedY2 {
		t.Errorf("Second child Y: expected %.2f, got %.2f", expectedY2, root.Children[1].Rect.Y)
	}

	// Third child at Y = 45 + 30 + max(25, 20) = 100
	expectedY3 := 100.0
	if root.Children[2].Rect.Y != expectedY3 {
		t.Errorf("Third child Y: expected %.2f, got %.2f", expectedY3, root.Children[2].Rect.Y)
	}
}

// TestBlockMarginCollapsingFirstChild tests first child's top margin
func TestBlockMarginCollapsingFirstChild(t *testing.T) {
	root := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   200,
			Height:  -1,
		},
		Children: []*Node{
			{
				Style: Style{
					Display: DisplayBlock,
					Height:  50,
					Margin:  Spacing{Top: 20}, // First child has top margin
				},
			},
		},
	}

	constraints := Loose(200, 200)
	LayoutBlock(root, constraints)

	// First child should be at Y = 20 (its top margin)
	expectedY := 20.0
	if root.Children[0].Rect.Y != expectedY {
		t.Errorf("First child Y: expected %.2f, got %.2f", expectedY, root.Children[0].Rect.Y)
	}
}

// TestBlockMarginCollapsingWithHorizontalMargins tests that horizontal margins don't collapse
func TestBlockMarginCollapsingWithHorizontalMargins(t *testing.T) {
	root := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   200,
			Height:  -1,
		},
		Children: []*Node{
			{
				Style: Style{
					Display: DisplayBlock,
					Height:  50,
					Margin:  Spacing{Left: 10, Right: 10, Bottom: 20},
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Height:  50,
					Margin:  Spacing{Left: 15, Right: 15, Top: 30},
				},
			},
		},
	}

	constraints := Loose(200, 200)
	LayoutBlock(root, constraints)

	// First child should have left margin applied
	if root.Children[0].Rect.X != 10 {
		t.Errorf("First child X: expected 10, got %.2f", root.Children[0].Rect.X)
	}

	// Second child should have left margin applied
	if root.Children[1].Rect.X != 15 {
		t.Errorf("Second child X: expected 15, got %.2f", root.Children[1].Rect.X)
	}

	// Vertical margins should still collapse
	expectedY := 80.0 // 50 + max(20, 30)
	if root.Children[1].Rect.Y != expectedY {
		t.Errorf("Second child Y: expected %.2f, got %.2f", expectedY, root.Children[1].Rect.Y)
	}
}

// TestBlockMarginCollapsingNestedFlex tests that flex items handle margins correctly
func TestBlockMarginCollapsingNestedFlex(t *testing.T) {
	// Margins should work correctly when block layout contains flex items
	root := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   200,
			Height:  -1,
		},
		Children: []*Node{
			{
				Style: Style{
					Display: DisplayFlex,
					Width:   100,
					Height:  50,
					Margin:  Spacing{Bottom: 20},
				},
				// Add a child to the flex container so it has content
				Children: []*Node{
					{Style: Style{Width: 50, Height: 30}},
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Height:  50,
					Margin:  Spacing{Top: 30},
				},
			},
		},
	}

	constraints := Loose(200, 200)
	LayoutBlock(root, constraints)

	// Margins should collapse even when one child is flex
	expectedY := 80.0 // 50 + max(20, 30)
	if root.Children[1].Rect.Y != expectedY {
		t.Errorf("Second child Y: expected %.2f, got %.2f", expectedY, root.Children[1].Rect.Y)
	}
}
