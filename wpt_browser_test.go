package layout

import (
	"math"
	"testing"
)

// Generated from WPT test: Simple Flex Row Test
// Source: simple-flex-row.json
// Browser: Chrome Headless
// Expected values extracted from actual browser layout

func TestWPTBrowser_1(t *testing.T) {
	// WPT test: Simple Flex Row Test
	// Browser expected values for #container

	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentSpaceBetween,
			AlignItems:     AlignItemsFlexStart,
			Width:          600,
			Height:         100,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
		},
	}

	constraints := Loose(800.00, 600.00)
	LayoutFlexbox(root, constraints)

	// Container dimensions (browser expected)
	if math.Abs(root.Rect.Width-600.00) > 1.0 {
		t.Errorf("Width: expected 600.00 (browser), got %f", root.Rect.Width)
	}
	if math.Abs(root.Rect.Height-100.00) > 1.0 {
		t.Errorf("Height: expected 100.00 (browser), got %f", root.Rect.Height)
	}

	// Child positions (browser expected, adjusted for no body margin in our engine)
	// Child 0
	if math.Abs(root.Children[0].Rect.X-0.00) > 1.0 {
		t.Errorf("Child 0 X: expected 0.00, got %f", root.Children[0].Rect.X)
	}
	if math.Abs(root.Children[0].Rect.Y-0.00) > 1.0 {
		t.Errorf("Child 0 Y: expected 0.00, got %f", root.Children[0].Rect.Y)
	}
	if math.Abs(root.Children[0].Rect.Width-100.00) > 1.0 {
		t.Errorf("Child 0 Width: expected 100.00, got %f", root.Children[0].Rect.Width)
	}
	if math.Abs(root.Children[0].Rect.Height-50.00) > 1.0 {
		t.Errorf("Child 0 Height: expected 50.00, got %f", root.Children[0].Rect.Height)
	}
	// Child 1
	if math.Abs(root.Children[1].Rect.X-250.00) > 1.0 {
		t.Errorf("Child 1 X: expected 250.00, got %f", root.Children[1].Rect.X)
	}
	if math.Abs(root.Children[1].Rect.Y-0.00) > 1.0 {
		t.Errorf("Child 1 Y: expected 0.00, got %f", root.Children[1].Rect.Y)
	}
	if math.Abs(root.Children[1].Rect.Width-100.00) > 1.0 {
		t.Errorf("Child 1 Width: expected 100.00, got %f", root.Children[1].Rect.Width)
	}
	if math.Abs(root.Children[1].Rect.Height-50.00) > 1.0 {
		t.Errorf("Child 1 Height: expected 50.00, got %f", root.Children[1].Rect.Height)
	}
	// Child 2
	if math.Abs(root.Children[2].Rect.X-500.00) > 1.0 {
		t.Errorf("Child 2 X: expected 500.00, got %f", root.Children[2].Rect.X)
	}
	if math.Abs(root.Children[2].Rect.Y-0.00) > 1.0 {
		t.Errorf("Child 2 Y: expected 0.00, got %f", root.Children[2].Rect.Y)
	}
	if math.Abs(root.Children[2].Rect.Width-100.00) > 1.0 {
		t.Errorf("Child 2 Width: expected 100.00, got %f", root.Children[2].Rect.Width)
	}
	if math.Abs(root.Children[2].Rect.Height-50.00) > 1.0 {
		t.Errorf("Child 2 Height: expected 50.00, got %f", root.Children[2].Rect.Height)
	}
}
