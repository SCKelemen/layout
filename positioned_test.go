package layout

import (
	"math"
	"testing"
)

func TestPositionAbsolute(t *testing.T) {
	// Test absolute positioning
	root := &Node{
		Style: Style{
			Width:  400,
			Height: 300,
		},
		Children: []*Node{
			{
				Style: Style{
					Position: PositionAbsolute,
					Left:     50,
					Top:      50,
					Width:    100,
					Height:   100,
				},
			},
		},
	}

	constraints := Loose(500, 400)
	Layout(root, constraints)
	LayoutWithPositioning(root, constraints, root.Rect)

	child := root.Children[0]
	if math.Abs(child.Rect.X-50.0) > 1.0 {
		t.Errorf("Expected X=50, got %.2f", child.Rect.X)
	}
	if math.Abs(child.Rect.Y-50.0) > 1.0 {
		t.Errorf("Expected Y=50, got %.2f", child.Rect.Y)
	}
}

func TestPositionRelative(t *testing.T) {
	// Test relative positioning (offsets from normal flow)
	root := &Node{
		Style: Style{
			Width:  400,
			Height: 300,
		},
		Children: []*Node{
			{
				Style: Style{
					Position: PositionRelative,
					Left:     20,
					Top:      10,
					Width:    100,
					Height:   100,
				},
			},
		},
	}

	constraints := Loose(500, 400)
	Layout(root, constraints)
	LayoutWithPositioning(root, constraints, root.Rect)

	child := root.Children[0]
	// Should be offset from normal flow position (0,0) by (20, 10)
	if math.Abs(child.Rect.X-20.0) > 1.0 {
		t.Errorf("Expected X=20, got %.2f", child.Rect.X)
	}
	if math.Abs(child.Rect.Y-10.0) > 1.0 {
		t.Errorf("Expected Y=10, got %.2f", child.Rect.Y)
	}
}

func TestPositionAbsoluteWithRightBottom(t *testing.T) {
	// Test absolute positioning with right and bottom
	root := &Node{
		Style: Style{
			Width:  400,
			Height: 300,
		},
		Children: []*Node{
			{
				Style: Style{
					Position: PositionAbsolute,
					Right:    50,
					Bottom:   50,
					Width:    100,
					Height:   100,
				},
			},
		},
	}

	constraints := Loose(500, 400)
	Layout(root, constraints)
	LayoutWithPositioning(root, constraints, root.Rect)

	child := root.Children[0]
	// Should be positioned from right and bottom
	expectedX := 400.0 - 100.0 - 50.0 // container width - child width - right offset
	expectedY := 300.0 - 100.0 - 50.0 // container height - child height - bottom offset

	if math.Abs(child.Rect.X-expectedX) > 1.0 {
		t.Errorf("Expected X=%.2f, got %.2f", expectedX, child.Rect.X)
	}
	if math.Abs(child.Rect.Y-expectedY) > 1.0 {
		t.Errorf("Expected Y=%.2f, got %.2f", expectedY, child.Rect.Y)
	}
}

func TestPositionFixed(t *testing.T) {
	// Test fixed positioning (relative to viewport)
	viewport := Rect{X: 0, Y: 0, Width: 800, Height: 600}
	root := &Node{
		Style: Style{
			Width:  400,
			Height: 300,
		},
		Children: []*Node{
			{
				Style: Style{
					Position: PositionFixed,
					Right:    10,
					Top:      10,
					Width:    100,
					Height:   100,
				},
			},
		},
	}

	constraints := Loose(500, 400)
	Layout(root, constraints)
	LayoutWithPositioning(root, constraints, viewport)

	child := root.Children[0]
	// Fixed positioning uses viewport, not parent
	expectedX := 800.0 - 100.0 - 10.0 // viewport width - child width - right offset
	expectedY := 10.0                 // top offset

	if math.Abs(child.Rect.X-expectedX) > 1.0 {
		t.Errorf("Expected X=%.2f, got %.2f", expectedX, child.Rect.X)
	}
	if math.Abs(child.Rect.Y-expectedY) > 1.0 {
		t.Errorf("Expected Y=%.2f, got %.2f", expectedY, child.Rect.Y)
	}
}

func TestPositionAbsoluteConstrainedWidth(t *testing.T) {
	// Test absolute positioning with both left and right (constrains width)
	root := &Node{
		Style: Style{
			Width:  400,
			Height: 300,
		},
		Children: []*Node{
			{
				Style: Style{
					Position: PositionAbsolute,
					Left:     50,
					Right:    50,
					Top:      50,
					Width:    500, // This should be constrained
					Height:   100,
				},
			},
		},
	}

	constraints := Loose(500, 400)
	Layout(root, constraints)
	LayoutWithPositioning(root, constraints, root.Rect)

	child := root.Children[0]
	// Width should be constrained: 400 - 50 - 50 = 300
	expectedWidth := 300.0
	if math.Abs(child.Rect.Width-expectedWidth) > 1.0 {
		t.Errorf("Expected width %.2f, got %.2f", expectedWidth, child.Rect.Width)
	}
}
