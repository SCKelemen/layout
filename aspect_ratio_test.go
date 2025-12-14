package layout

import (
	"math"
	"testing"
)

func TestAspectRatioWidthSet(t *testing.T) {
	// When width is set and height is auto, aspect ratio should calculate height
	node := &Node{
		Style: Style{
			Width:       Px(800),
			Height:      Px(-1), // auto
			AspectRatio: 16.0 / 9.0,
		},
	}

	constraints := Loose(1000, 1000)
	LayoutBlock(node, constraints, NewLayoutContext(1920, 1080, 16))

	expectedHeight := 800.0 / (16.0 / 9.0) // 450
	if math.Abs(node.Rect.Height-expectedHeight) > 0.01 {
		t.Errorf("Expected height %.2f, got %.2f", expectedHeight, node.Rect.Height)
	}
}

func TestAspectRatioHeightSet(t *testing.T) {
	// When height is set and width is auto, aspect ratio should calculate width
	node := &Node{
		Style: Style{
			Width:       Px(-1), // auto
			Height:      Px(450),
			AspectRatio: 16.0 / 9.0,
		},
	}

	constraints := Loose(1000, 1000)
	LayoutBlock(node, constraints, NewLayoutContext(1920, 1080, 16))

	expectedWidth := 450.0 * (16.0 / 9.0) // 800
	if math.Abs(node.Rect.Width-expectedWidth) > 0.01 {
		t.Errorf("Expected width %.2f, got %.2f", expectedWidth, node.Rect.Width)
	}
}

func TestAspectRatioBothAuto(t *testing.T) {
	// When both width and height are auto, aspect ratio should use available space
	node := &Node{
		Style: Style{
			Width:       Px(-1), // auto
			Height:      Px(-1), // auto
			AspectRatio: 16.0 / 9.0,
		},
	}

	constraints := Loose(800, 600)
	LayoutBlock(node, constraints, NewLayoutContext(1920, 1080, 16))

	// Should use available width (800) and calculate height
	expectedHeight := 800.0 / (16.0 / 9.0) // 450
	if math.Abs(node.Rect.Height-expectedHeight) > 0.01 {
		t.Errorf("Expected height %.2f, got %.2f", expectedHeight, node.Rect.Height)
	}
	if math.Abs(node.Rect.Width-800.0) > 0.01 {
		t.Errorf("Expected width 800.0, got %.2f", node.Rect.Width)
	}
}

func TestAspectRatioConstrainedByHeight(t *testing.T) {
	// When both are auto but calculated height exceeds available height
	node := &Node{
		Style: Style{
			Width:       Px(-1), // auto
			Height:      Px(-1), // auto
			AspectRatio: 16.0 / 9.0,
		},
	}

	constraints := Loose(800, 400) // Height constraint is smaller
	LayoutBlock(node, constraints, NewLayoutContext(1920, 1080, 16))

	// Should constrain to available height and recalculate width
	expectedWidth := 400.0 * (16.0 / 9.0) // 711.11...
	if math.Abs(node.Rect.Width-expectedWidth) > 0.01 {
		t.Errorf("Expected width %.2f, got %.2f", expectedWidth, node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-400.0) > 0.01 {
		t.Errorf("Expected height 400.0, got %.2f", node.Rect.Height)
	}
}

func TestAspectRatioBothSet(t *testing.T) {
	// When both width and height are explicitly set, aspect ratio should be ignored
	node := &Node{
		Style: Style{
			Width:       Px(800),
			Height:      Px(600),
			AspectRatio: 16.0 / 9.0, // Should be ignored
		},
	}

	constraints := Loose(1000, 1000)
	LayoutBlock(node, constraints, NewLayoutContext(1920, 1080, 16))

	// Should use explicit dimensions, not aspect ratio
	if math.Abs(node.Rect.Width-800.0) > 0.01 {
		t.Errorf("Expected width 800.0, got %.2f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-600.0) > 0.01 {
		t.Errorf("Expected height 600.0, got %.2f", node.Rect.Height)
	}
}

func TestAspectRatioRespectsMinMax(t *testing.T) {
	// Aspect ratio should respect min/max constraints
	node := &Node{
		Style: Style{
			Width:       Px(800),
			Height:      Px(-1), // auto
			AspectRatio: 16.0 / 9.0,
			MinHeight:   Px(500), // Should override calculated height
		},
	}

	constraints := Loose(1000, 1000)
	LayoutBlock(node, constraints, NewLayoutContext(1920, 1080, 16))

	// MinHeight should override aspect ratio calculation
	if node.Rect.Height < 500.0 {
		t.Errorf("Height should respect MinHeight: expected >= 500.0, got %.2f", node.Rect.Height)
	}
}

func TestAspectRatioSquare(t *testing.T) {
	// Test 1:1 aspect ratio (square)
	node := &Node{
		Style: Style{
			Width:       Px(200),
			Height:      Px(-1), // auto
			AspectRatio: 1.0,
		},
	}

	constraints := Loose(1000, 1000)
	LayoutBlock(node, constraints, NewLayoutContext(1920, 1080, 16))

	expectedHeight := 200.0
	if math.Abs(node.Rect.Height-expectedHeight) > 0.01 {
		t.Errorf("Expected height %.2f, got %.2f", expectedHeight, node.Rect.Height)
	}
}
