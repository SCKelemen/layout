package layout

import (
	"math"
	"testing"
)

func TestPositionSticky(t *testing.T) {
	// Test sticky positioning (treated as relative without scroll context)
	root := &Node{
		Style: Style{
			Width:  Px(400),
			Height: Px(300),
		},
		Children: []*Node{
			{
				Style: Style{
					Position: PositionSticky,
					Top:      Px(10),
					Left:     Px(20),
					Width:    Px(100),
					Height:   Px(100),
				},
			},
		},
	}

	constraints := Loose(500, 400)
	ctx := NewLayoutContext(800, 600, 16)
	Layout(root, constraints, ctx)
	LayoutWithPositioning(root, constraints, root.Rect)

	child := root.Children[0]
	// Sticky should offset like relative
	if child.Rect.X < 20 {
		t.Errorf("Expected X >= 20, got %.2f", child.Rect.X)
	}
	if child.Rect.Y < 10 {
		t.Errorf("Expected Y >= 10, got %.2f", child.Rect.Y)
	}
}

func TestLayoutWithPositioning(t *testing.T) {
	// Test the LayoutWithPositioning helper function
	root := &Node{
		Style: Style{
			Width:  Px(400),
			Height: Px(300),
		},
		Children: []*Node{
			{
				Style: Style{
					Position: PositionAbsolute,
					Left:     Px(50),
					Top:      Px(50),
					Width:    Px(100),
					Height:   Px(100),
				},
			},
		},
	}

	constraints := Loose(500, 400)
	viewport := Rect{X: 0, Y: 0, Width: 500, Height: 400}
	size := LayoutWithPositioning(root, constraints, viewport)

	if size.Width <= 0 || size.Height <= 0 {
		t.Error("LayoutWithPositioning should return valid size")
	}

	child := root.Children[0]
	if math.Abs(child.Rect.X-50.0) > 1.0 {
		t.Errorf("Expected X=50, got %.2f", child.Rect.X)
	}
	if math.Abs(child.Rect.Y-50.0) > 1.0 {
		t.Errorf("Expected Y=50, got %.2f", child.Rect.Y)
	}
}
