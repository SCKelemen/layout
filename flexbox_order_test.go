package layout

import "testing"

// TestFlexboxOrderBasic tests basic order property functionality
// Items should be reordered visually by their order value
func TestFlexboxOrderBasic(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentFlexStart,
			Width:          300,
			Height:         100,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50, Order: 2}}, // Should be last
			{Style: Style{Width: 50, Height: 50, Order: 0}}, // Should be first (default)
			{Style: Style{Width: 50, Height: 50, Order: 1}}, // Should be middle
		},
	}

	LayoutFlexbox(container, Loose(300, 100))

	// Check that items are positioned in order: Order 0, Order 1, Order 2
	if container.Children[0].Rect.X != 100 {
		t.Errorf("Item with Order=2 should be at X=100, got %v", container.Children[0].Rect.X)
	}
	if container.Children[1].Rect.X != 0 {
		t.Errorf("Item with Order=0 should be at X=0, got %v", container.Children[1].Rect.X)
	}
	if container.Children[2].Rect.X != 50 {
		t.Errorf("Item with Order=1 should be at X=50, got %v", container.Children[2].Rect.X)
	}
}

// TestFlexboxOrderNegative tests negative order values
// Negative order values should place items before those with order 0 (default)
func TestFlexboxOrderNegative(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentFlexStart,
			Width:          300,
			Height:         100,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50, Order: 0}},  // Should be middle (default)
			{Style: Style{Width: 50, Height: 50, Order: -1}}, // Should be first
			{Style: Style{Width: 50, Height: 50, Order: 1}},  // Should be last
		},
	}

	LayoutFlexbox(container, Loose(300, 100))

	// Check that items are positioned in order: Order -1, Order 0, Order 1
	if container.Children[0].Rect.X != 50 {
		t.Errorf("Item with Order=0 should be at X=50, got %v", container.Children[0].Rect.X)
	}
	if container.Children[1].Rect.X != 0 {
		t.Errorf("Item with Order=-1 should be at X=0, got %v", container.Children[1].Rect.X)
	}
	if container.Children[2].Rect.X != 100 {
		t.Errorf("Item with Order=1 should be at X=100, got %v", container.Children[2].Rect.X)
	}
}

// TestFlexboxOrderSameValue tests items with the same order value
// Items with the same order should maintain source order (stable sort)
func TestFlexboxOrderSameValue(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentFlexStart,
			Width:          300,
			Height:         100,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50, Order: 1}}, // First with order=1
			{Style: Style{Width: 50, Height: 50, Order: 1}}, // Second with order=1
			{Style: Style{Width: 50, Height: 50, Order: 1}}, // Third with order=1
		},
	}

	LayoutFlexbox(container, Loose(300, 100))

	// All items have the same order, so they should maintain source order
	if container.Children[0].Rect.X != 0 {
		t.Errorf("First item should be at X=0, got %v", container.Children[0].Rect.X)
	}
	if container.Children[1].Rect.X != 50 {
		t.Errorf("Second item should be at X=50, got %v", container.Children[1].Rect.X)
	}
	if container.Children[2].Rect.X != 100 {
		t.Errorf("Third item should be at X=100, got %v", container.Children[2].Rect.X)
	}
}

// TestFlexboxOrderColumn tests order property in column direction
func TestFlexboxOrderColumn(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionColumn,
			JustifyContent: JustifyContentFlexStart,
			Width:          100,
			Height:         300,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50, Order: 2}}, // Should be last
			{Style: Style{Width: 50, Height: 50, Order: 0}}, // Should be first (default)
			{Style: Style{Width: 50, Height: 50, Order: 1}}, // Should be middle
		},
	}

	LayoutFlexbox(container, Loose(100, 300))

	// Check that items are positioned in order: Order 0, Order 1, Order 2
	if container.Children[0].Rect.Y != 100 {
		t.Errorf("Item with Order=2 should be at Y=100, got %v", container.Children[0].Rect.Y)
	}
	if container.Children[1].Rect.Y != 0 {
		t.Errorf("Item with Order=0 should be at Y=0, got %v", container.Children[1].Rect.Y)
	}
	if container.Children[2].Rect.Y != 50 {
		t.Errorf("Item with Order=1 should be at Y=50, got %v", container.Children[2].Rect.Y)
	}
}

// TestFlexboxOrderWithWrap tests order property with flex-wrap
func TestFlexboxOrderWithWrap(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			FlexWrap:       FlexWrapWrap,
			JustifyContent: JustifyContentFlexStart,
			Width:          150,
			Height:         200,
		},
		Children: []*Node{
			{Style: Style{Width: 60, Height: 50, Order: 3}}, // Should be last (second line)
			{Style: Style{Width: 60, Height: 50, Order: 0}}, // Should be first (first line)
			{Style: Style{Width: 60, Height: 50, Order: 1}}, // Should be second (first line)
			{Style: Style{Width: 60, Height: 50, Order: 2}}, // Should be third (second line)
		},
	}

	LayoutFlexbox(container, Loose(150, 200))

	// First line: Order 0, Order 1
	// Second line: Order 2, Order 3
	// With align-content defaulting to stretch, the 2 lines stretch to fill 200px height
	// So each line is 100px tall (200/2 = 100)
	if container.Children[1].Rect.X != 0 || container.Children[1].Rect.Y != 0 {
		t.Errorf("Item with Order=0 should be at (0,0), got (%v,%v)", container.Children[1].Rect.X, container.Children[1].Rect.Y)
	}
	if container.Children[2].Rect.X != 60 || container.Children[2].Rect.Y != 0 {
		t.Errorf("Item with Order=1 should be at (60,0), got (%v,%v)", container.Children[2].Rect.X, container.Children[2].Rect.Y)
	}
	if container.Children[3].Rect.X != 0 || container.Children[3].Rect.Y != 100 {
		t.Errorf("Item with Order=2 should be at (0,100), got (%v,%v)", container.Children[3].Rect.X, container.Children[3].Rect.Y)
	}
	if container.Children[0].Rect.X != 60 || container.Children[0].Rect.Y != 100 {
		t.Errorf("Item with Order=3 should be at (60,100), got (%v,%v)", container.Children[0].Rect.X, container.Children[0].Rect.Y)
	}
}

// TestFlexboxOrderWithGaps tests order property with gaps
func TestFlexboxOrderWithGaps(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentFlexStart,
			FlexColumnGap:  10,
			Width:          300,
			Height:         100,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50, Order: 2}}, // Should be last
			{Style: Style{Width: 50, Height: 50, Order: 0}}, // Should be first
			{Style: Style{Width: 50, Height: 50, Order: 1}}, // Should be middle
		},
	}

	LayoutFlexbox(container, Loose(300, 100))

	// Check that items are positioned with gaps in order: Order 0, Order 1, Order 2
	// Expected positions: 0, 60 (50+10), 120 (50+10+50+10)
	if container.Children[1].Rect.X != 0 {
		t.Errorf("Item with Order=0 should be at X=0, got %v", container.Children[1].Rect.X)
	}
	if container.Children[2].Rect.X != 60 {
		t.Errorf("Item with Order=1 should be at X=60, got %v", container.Children[2].Rect.X)
	}
	if container.Children[0].Rect.X != 120 {
		t.Errorf("Item with Order=2 should be at X=120, got %v", container.Children[0].Rect.X)
	}
}
