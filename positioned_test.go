package layout

import (
	"math"
	"testing"
)

func TestPositionAbsolute(t *testing.T) {
	// Test absolute positioning
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
	ctx := NewLayoutContext(800, 600, 16)
	Layout(root, constraints, ctx)
	LayoutWithPositioning(root, constraints, root.Rect, ctx)

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
			Width:  Px(400),
			Height: Px(300),
		},
		Children: []*Node{
			{
				Style: Style{
					Position: PositionRelative,
					Left:     Px(20),
					Top:      Px(10),
					Width:    Px(100),
					Height:   Px(100),
				},
			},
		},
	}

	constraints := Loose(500, 400)
	ctx := NewLayoutContext(800, 600, 16)
	Layout(root, constraints, ctx)
	LayoutWithPositioning(root, constraints, root.Rect, ctx)

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
			Width:  Px(400),
			Height: Px(300),
		},
		Children: []*Node{
			{
				Style: Style{
					Position: PositionAbsolute,
					Right:    Px(50),
					Bottom:   Px(50),
					Width:    Px(100),
					Height:   Px(100),
				},
			},
		},
	}

	constraints := Loose(500, 400)
	ctx := NewLayoutContext(800, 600, 16)
	Layout(root, constraints, ctx)
	LayoutWithPositioning(root, constraints, root.Rect, ctx)

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
			Width:  Px(400),
			Height: Px(300),
		},
		Children: []*Node{
			{
				Style: Style{
					Position: PositionFixed,
					Right:    Px(10),
					Top:      Px(10),
					Width:    Px(100),
					Height:   Px(100),
				},
			},
		},
	}

	constraints := Loose(500, 400)
	ctx := NewLayoutContext(800, 600, 16)
	Layout(root, constraints, ctx)
	LayoutWithPositioning(root, constraints, viewport, ctx)

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
	// Absolute positioning with left, width, and right all set is
	// over-constrained. CSS 2.1 §10.3.7: "If the values are over-constrained,
	// ignore the value for 'left' (in case 'direction' of the containing block
	// is 'rtl') or 'right' (in case 'direction' is 'ltr')". The explicit width
	// is kept (overflowing the containing block) and the box sits at left.
	// https://www.w3.org/TR/CSS21/visudet.html#abs-non-replaced-width
	//
	// Previously the engine clamped the width to the space between the
	// offsets (400 - 50 - 50 = 300), which is the rule for width: auto only.
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
					Right:    Px(50),
					Top:      Px(50),
					Width:    Px(500), // This should be constrained
					Height:   Px(100),
				},
			},
		},
	}

	constraints := Loose(500, 400)
	ctx := NewLayoutContext(800, 600, 16)
	Layout(root, constraints, ctx)
	LayoutWithPositioning(root, constraints, root.Rect, ctx)

	child := root.Children[0]
	// Explicit width wins; right is ignored (ltr).
	if child.Rect.Width != 500 {
		t.Errorf("Expected explicit width 500 to be kept, got %.2f", child.Rect.Width)
	}
	if child.Rect.X != 50 {
		t.Errorf("Expected X=50 (left honored, right ignored), got %.2f", child.Rect.X)
	}

	// direction: rtl on the containing block ignores left instead
	// (CSS 2.1 §10.3.7 uses the containing block's direction):
	// X = 400 - 50 - 500 = -150.
	root.Style.Direction = DirectionRTL
	LayoutWithPositioning(root, constraints, root.Rect, ctx)
	if child.Rect.Width != 500 {
		t.Errorf("rtl: expected explicit width 500 to be kept, got %.2f", child.Rect.Width)
	}
	if child.Rect.X != -150 {
		t.Errorf("rtl: expected X=-150 (right honored, left ignored), got %.2f", child.Rect.X)
	}

	// With width: auto the box spans the space between the offsets.
	child.Style.Direction = DirectionLTR
	child.Style.Width = Length{}
	LayoutWithPositioning(root, constraints, root.Rect, ctx)
	if child.Rect.Width != 300 || child.Rect.X != 50 {
		t.Errorf("auto width: expected X=50 W=300, got X=%.2f W=%.2f", child.Rect.X, child.Rect.Width)
	}
}
