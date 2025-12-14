package layout

import (
	"math"
	"testing"
)

func TestBlockBasic(t *testing.T) {
	// Test basic block layout
	root := &Node{
		Style: Style{
			Width:  Px(200),
			Height: Px(100),
		},
		Children: []*Node{},
	}

	constraints := Loose(300, 300)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	if math.Abs(size.Width-200.0) > 1.0 {
		t.Errorf("Expected width 200, got %.2f", size.Width)
	}
	if math.Abs(size.Height-100.0) > 1.0 {
		t.Errorf("Expected height 100, got %.2f", size.Height)
	}
}

func TestBlockStackedChildren(t *testing.T) {
	// Test block layout stacks children vertically
	root := &Node{
		Style: Style{},
		Children: []*Node{
			{Style: Style{Width: Px(100), Height: Px(50)}},
			{Style: Style{Width: Px(100), Height: Px(50)}},
		},
	}

	constraints := Loose(200, 300)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutBlock(root, constraints, ctx)

	// First child should be at top
	if root.Children[0].Rect.Y != 0 {
		t.Errorf("First child Y should be 0, got %.2f", root.Children[0].Rect.Y)
	}

	// Second child should be below first
	expectedY := 50.0
	if math.Abs(root.Children[1].Rect.Y-expectedY) > 1.0 {
		t.Errorf("Second child Y should be %.2f, got %.2f", expectedY, root.Children[1].Rect.Y)
	}
}

func TestBlockAutoWidth(t *testing.T) {
	// Test auto width uses max child width
	root := &Node{
		Style: Style{
			Width: Px(-1), // auto
		},
		Children: []*Node{
			{Style: Style{Width: Px(150), Height: Px(50)}},
			{Style: Style{Width: Px(200), Height: Px(50)}},
			{Style: Style{Width: Px(100), Height: Px(50)}},
		},
	}

	constraints := Loose(300, 300)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	// Width should be max child width (200)
	expectedWidth := 200.0
	if math.Abs(size.Width-expectedWidth) > 1.0 {
		t.Errorf("Expected width %.2f (max child), got %.2f", expectedWidth, size.Width)
	}
}

func TestBlockAutoHeight(t *testing.T) {
	// Test auto height uses sum of children heights
	root := &Node{
		Style: Style{
			Height: Px(-1), // auto
		},
		Children: []*Node{
			{Style: Style{Width: Px(100), Height: Px(50)}},
			{Style: Style{Width: Px(100), Height: Px(75)}},
			{Style: Style{Width: Px(100), Height: Px(25)}},
		},
	}

	constraints := Loose(200, 300)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	// Height should be sum of children: 50 + 75 + 25 = 150
	expectedHeight := 150.0
	if math.Abs(size.Height-expectedHeight) > 1.0 {
		t.Errorf("Expected height %.2f (sum of children), got %.2f", expectedHeight, size.Height)
	}
}

func TestBlockPadding(t *testing.T) {
	// Test padding affects block size
	padding := 10.0
	root := &Node{
		Style: Style{
			Width:   Px(100),
			Height:  Px(100),
			Padding: Uniform(Px(padding)),
		},
		Children: []*Node{},
	}

	constraints := Loose(200, 200)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	// Size should include padding: 100 + 20 = 120
	expectedWidth := 100.0 + padding*2
	if math.Abs(size.Width-expectedWidth) > 1.0 {
		t.Errorf("Expected width with padding %.2f, got %.2f", expectedWidth, size.Width)
	}
}

func TestBlockMinMaxConstraints(t *testing.T) {
	// Test min/max width constraints
	root := &Node{
		Style: Style{
			Width:    Px(300),
			Height:   Px(100),
			MinWidth: Px(100),
			MaxWidth: Px(200),
		},
		Children: []*Node{},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	// Width should be clamped to max
	if size.Width > 200.1 {
		t.Errorf("Width should be clamped to max 200, got %.2f", size.Width)
	}

	// Test min constraint
	root2 := &Node{
		Style: Style{
			Width:    Px(50),
			Height:   Px(100),
			MinWidth: Px(100),
		},
		Children: []*Node{},
	}

	size2 := LayoutBlock(root2, constraints, ctx)
	if size2.Width < 99.9 {
		t.Errorf("Width should be at least min 100, got %.2f", size2.Width)
	}
}

func TestBlockConstraints(t *testing.T) {
	// Test that block respects constraints
	root := &Node{
		Style: Style{
			Width:  Px(500),
			Height: Px(500),
		},
		Children: []*Node{},
	}

	constraints := Tight(200, 200)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	// Size should be constrained to 200x200
	if size.Width > 200.1 {
		t.Errorf("Width should be constrained to 200, got %.2f", size.Width)
	}
	if size.Height > 200.1 {
		t.Errorf("Height should be constrained to 200, got %.2f", size.Height)
	}
}

func TestBlockEmpty(t *testing.T) {
	// Test empty block
	root := &Node{
		Style: Style{
			Width:  Px(100),
			Height: Px(100),
		},
		Children: []*Node{},
	}

	constraints := Loose(200, 200)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	if math.Abs(size.Width-100.0) > 1.0 {
		t.Errorf("Empty block width should be 100, got %.2f", size.Width)
	}
	if math.Abs(size.Height-100.0) > 1.0 {
		t.Errorf("Empty block height should be 100, got %.2f", size.Height)
	}
}
