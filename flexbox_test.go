package layout

import (
	"math"
	"testing"
)

func TestFlexboxBasicRow(t *testing.T) {
	// Test basic row flexbox with two fixed-size children
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
		},
		Children: []*Node{
			{Style: Style{Width: Px(100), Height: Px(50)}},
			{Style: Style{Width: Px(100), Height: Px(50)}},
		},
	}

	constraints := Loose(400, 200)
	ctx := NewLayoutContext(800, 600, 16)
	size := LayoutFlexbox(root, constraints, ctx)

	// Container should be at least as wide as children
	if size.Width < 200 {
		t.Errorf("Expected container width >= 200, got %.2f", size.Width)
	}

	// Check children are positioned correctly
	if root.Children[0].Rect.X < 0 {
		t.Errorf("Child 0 X should be >= 0, got %.2f", root.Children[0].Rect.X)
	}
	if root.Children[0].Rect.Width != 100 {
		t.Errorf("Child 0 width should be 100, got %.2f", root.Children[0].Rect.Width)
	}
	if root.Children[1].Rect.Width != 100 {
		t.Errorf("Child 1 width should be 100, got %.2f", root.Children[1].Rect.Width)
	}
}

func TestFlexboxJustifyContentFlexStart(t *testing.T) {
	// Test justify-content: flex-start (default)
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentFlexStart,
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}},
			{Style: Style{Width: Px(50), Height: Px(50)}},
		},
	}

	constraints := Loose(200, 100)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutFlexbox(root, constraints, ctx)

	// First child should be at X=0
	if root.Children[0].Rect.X != 0 {
		t.Errorf("Expected first child at X=0, got %.2f", root.Children[0].Rect.X)
	}
}

func TestFlexboxJustifyContentSpaceBetween(t *testing.T) {
	// Test justify-content: space-between
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentSpaceBetween,
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}},
			{Style: Style{Width: Px(50), Height: Px(50)}},
		},
	}

	constraints := Tight(200, 100)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutFlexbox(root, constraints, ctx)

	// First child should be at start
	if root.Children[0].Rect.X != 0 {
		t.Errorf("Expected first child at X=0, got %.2f", root.Children[0].Rect.X)
	}

	// Second child should be at end (or close to it)
	expectedX := 200.0 - 50.0
	if math.Abs(root.Children[1].Rect.X-expectedX) > 1.0 {
		t.Errorf("Expected second child at X=%.2f, got %.2f", expectedX, root.Children[1].Rect.X)
	}
}

func TestFlexboxJustifyContentCenter(t *testing.T) {
	// Test justify-content: center
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentCenter,
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}},
		},
	}

	constraints := Tight(200, 100)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutFlexbox(root, constraints, ctx)

	// Child should be centered
	expectedX := (200.0 - 50.0) / 2.0
	if math.Abs(root.Children[0].Rect.X-expectedX) > 1.0 {
		t.Errorf("Expected child at X=%.2f (centered), got %.2f", expectedX, root.Children[0].Rect.X)
	}
}

func TestFlexboxFlexGrow(t *testing.T) {
	// Test flex-grow property
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50), FlexGrow: 1}},
			{Style: Style{Width: Px(50), Height: Px(50), FlexGrow: 2}},
		},
	}

	constraints := Tight(300, 100)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutFlexbox(root, constraints, ctx)

	// Calculate expected sizes
	// Available space: 300 - 50 - 50 = 200
	// Total flex-grow: 3
	// First child: 50 + (200 * 1 / 3) = 50 + 66.67 = 116.67
	// Second child: 50 + (200 * 2 / 3) = 50 + 133.33 = 183.33

	child0Width := root.Children[0].Rect.Width
	child1Width := root.Children[1].Rect.Width

	// First child should be smaller than second
	if child0Width >= child1Width {
		t.Errorf("First child (flex-grow: 1) should be smaller than second (flex-grow: 2). Got %.2f vs %.2f", child0Width, child1Width)
	}

	// Total width should be close to 300
	totalWidth := child0Width + child1Width
	if math.Abs(totalWidth-300.0) > 1.0 {
		t.Errorf("Total width should be ~300, got %.2f", totalWidth)
	}
}

func TestFlexboxFlexShrink(t *testing.T) {
	// Test flex-shrink property
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
		},
		Children: []*Node{
			{Style: Style{Width: Px(200), Height: Px(50), FlexShrink: 1}},
			{Style: Style{Width: Px(200), Height: Px(50), FlexShrink: 2}},
		},
	}

	constraints := Tight(300, 100)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutFlexbox(root, constraints, ctx)

	child0Width := root.Children[0].Rect.Width
	child1Width := root.Children[1].Rect.Width

	// Second child should shrink more (flex-shrink: 2)
	if child1Width >= child0Width {
		t.Errorf("Second child (flex-shrink: 2) should be smaller than first (flex-shrink: 1). Got %.2f vs %.2f", child1Width, child0Width)
	}

	// Total width should be close to 300
	totalWidth := child0Width + child1Width
	if math.Abs(totalWidth-300.0) > 1.0 {
		t.Errorf("Total width should be ~300, got %.2f", totalWidth)
	}
}

func TestFlexboxColumnDirection(t *testing.T) {
	// Test column direction
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
		},
		Children: []*Node{
			{Style: Style{Width: Px(100), Height: Px(50)}},
			{Style: Style{Width: Px(100), Height: Px(50)}},
		},
	}

	constraints := Loose(200, 200)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutFlexbox(root, constraints, ctx)

	// Children should be stacked vertically
	if root.Children[0].Rect.Y != 0 {
		t.Errorf("First child Y should be 0, got %.2f", root.Children[0].Rect.Y)
	}

	// Second child should be below first
	if root.Children[1].Rect.Y < root.Children[0].Rect.Y+root.Children[0].Rect.Height {
		t.Errorf("Second child should be below first. Got Y=%.2f, expected >= %.2f",
			root.Children[1].Rect.Y, root.Children[0].Rect.Y+root.Children[0].Rect.Height)
	}
}

func TestFlexboxAlignItemsStretch(t *testing.T) {
	// Test align-items: stretch
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			AlignItems:    AlignItemsStretch,
		},
		Children: []*Node{
			{Style: Style{Width: Px(50)}}, // No height specified
			{Style: Style{Width: Px(50)}}, // No height specified
		},
	}

	constraints := Tight(200, 100)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutFlexbox(root, constraints, ctx)

	// Children should stretch to container height
	expectedHeight := 100.0
	for i, child := range root.Children {
		if math.Abs(child.Rect.Height-expectedHeight) > 1.0 {
			t.Errorf("Child %d should stretch to height %.2f, got %.2f", i, expectedHeight, child.Rect.Height)
		}
	}
}

func TestFlexboxAlignItemsCenter(t *testing.T) {
	// Test align-items: center
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			AlignItems:    AlignItemsCenter,
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(30)}},
		},
	}

	constraints := Tight(200, 100)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutFlexbox(root, constraints, ctx)

	// Child should be vertically centered
	expectedY := (100.0 - 30.0) / 2.0
	if math.Abs(root.Children[0].Rect.Y-expectedY) > 1.0 {
		t.Errorf("Expected child at Y=%.2f (centered), got %.2f", expectedY, root.Children[0].Rect.Y)
	}
}

func TestFlexboxPadding(t *testing.T) {
	// Test padding affects container size
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			Padding:       Uniform(Px(10)),
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}},
		},
	}

	constraints := Loose(200, 100)
	ctx := NewLayoutContext(800, 600, 16)
	size := LayoutFlexbox(root, constraints, ctx)

	// Container should include padding
	// Content: 50, Padding: 20 (10*2), Total: 70
	minExpectedWidth := 70.0
	if size.Width < minExpectedWidth {
		t.Errorf("Expected width >= %.2f (content + padding), got %.2f", minExpectedWidth, size.Width)
	}
}

func TestFlexboxEmptyContainer(t *testing.T) {
	// Test empty flex container
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
		},
		Children: []*Node{},
	}

	constraints := Loose(200, 100)
	ctx := NewLayoutContext(800, 600, 16)
	size := LayoutFlexbox(root, constraints, ctx)

	// Empty container should have minimal size
	if size.Width < 0 {
		t.Errorf("Empty container width should be >= 0, got %.2f", size.Width)
	}
	if size.Height < 0 {
		t.Errorf("Empty container height should be >= 0, got %.2f", size.Height)
	}
}

func TestFlexboxNested(t *testing.T) {
	// Test nested flex containers
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:       DisplayFlex,
					FlexDirection: FlexDirectionRow,
				},
				Children: []*Node{
					{Style: Style{Width: Px(50), Height: Px(50)}},
					{Style: Style{Width: Px(50), Height: Px(50)}},
				},
			},
		},
	}

	constraints := Loose(200, 200)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutFlexbox(root, constraints, ctx)

	// Nested container should be laid out
	if len(root.Children[0].Children) != 2 {
		t.Errorf("Expected 2 children in nested container, got %d", len(root.Children[0].Children))
	}
}

