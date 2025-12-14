package layout

import (
	"math"
	"testing"
)

// TestBoxSizingContentBox tests that content-box (default) behavior works correctly
// In content-box, width/height = content size only, padding and border are added
func TestBoxSizingContentBox(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:     Px(100),
			Height:    Px(100),
			Padding:   Uniform(Px(10)),
			Border:    Uniform(Px(5)),
			BoxSizing: BoxSizingContentBox, // Explicit, though this is the default
		},
		Children: []*Node{},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	// With content-box: width=100 means content=100, total = 100 + 10*2 + 5*2 = 130
	expectedWidth := 100.0 + 10*2 + 5*2
	expectedHeight := 100.0 + 10*2 + 5*2

	if math.Abs(size.Width-expectedWidth) > 0.01 {
		t.Errorf("Content-box width: expected %.2f, got %.2f", expectedWidth, size.Width)
	}
	if math.Abs(size.Height-expectedHeight) > 0.01 {
		t.Errorf("Content-box height: expected %.2f, got %.2f", expectedHeight, size.Height)
	}
}

// TestBoxSizingBorderBox tests that border-box behavior works correctly
// In border-box, width/height = content + padding + border, total size = width
func TestBoxSizingBorderBox(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:     Px(100),
			Height:    Px(100),
			Padding:   Uniform(Px(10)),
			Border:    Uniform(Px(5)),
			BoxSizing: BoxSizingBorderBox,
		},
		Children: []*Node{},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	// With border-box: width=100 means total=100, content = 100 - 10*2 - 5*2 = 70
	expectedWidth := 100.0  // Total size equals specified width
	expectedHeight := 100.0 // Total size equals specified height

	if math.Abs(size.Width-expectedWidth) > 0.01 {
		t.Errorf("Border-box width: expected %.2f, got %.2f", expectedWidth, size.Width)
	}
	if math.Abs(size.Height-expectedHeight) > 0.01 {
		t.Errorf("Border-box height: expected %.2f, got %.2f", expectedHeight, size.Height)
	}

	// Verify content area is correct
	contentWidth := size.Width - 10*2 - 5*2
	contentHeight := size.Height - 10*2 - 5*2
	expectedContentWidth := 70.0
	expectedContentHeight := 70.0

	if math.Abs(contentWidth-expectedContentWidth) > 0.01 {
		t.Errorf("Border-box content width: expected %.2f, got %.2f", expectedContentWidth, contentWidth)
	}
	if math.Abs(contentHeight-expectedContentHeight) > 0.01 {
		t.Errorf("Border-box content height: expected %.2f, got %.2f", expectedContentHeight, contentHeight)
	}
}

// TestBoxSizingContentBoxAuto tests content-box with auto width/height
func TestBoxSizingContentBoxAuto(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:     Px(-1), // auto
			Height:    Px(-1), // auto
			Padding:   Uniform(Px(10)),
			Border:    Uniform(Px(5)),
			BoxSizing: BoxSizingContentBox,
		},
		Children: []*Node{
			{Style: Style{Width: Px(100), Height: Px(50)}},
		},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	// Auto width should be child width (100) + padding + border = 100 + 20 + 10 = 130
	// Auto height should be child height (50) + padding + border = 50 + 20 + 10 = 80
	expectedWidth := 100.0 + 10*2 + 5*2
	expectedHeight := 50.0 + 10*2 + 5*2

	if math.Abs(size.Width-expectedWidth) > 0.01 {
		t.Errorf("Content-box auto width: expected %.2f, got %.2f", expectedWidth, size.Width)
	}
	if math.Abs(size.Height-expectedHeight) > 0.01 {
		t.Errorf("Content-box auto height: expected %.2f, got %.2f", expectedHeight, size.Height)
	}
}

// TestBoxSizingBorderBoxAuto tests border-box with auto width/height
func TestBoxSizingBorderBoxAuto(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:     Px(-1), // auto
			Height:    Px(-1), // auto
			Padding:   Uniform(Px(10)),
			Border:    Uniform(Px(5)),
			BoxSizing: BoxSizingBorderBox,
		},
		Children: []*Node{
			{Style: Style{Width: Px(100), Height: Px(50)}},
		},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	// Auto width/height in border-box should size to content + padding + border
	// Content width = child width = 100, so total = 100 + 20 + 10 = 130
	// Content height = child height = 50, so total = 50 + 20 + 10 = 80
	expectedWidth := 100.0 + 10*2 + 5*2
	expectedHeight := 50.0 + 10*2 + 5*2

	if math.Abs(size.Width-expectedWidth) > 0.01 {
		t.Errorf("Border-box auto width: expected %.2f, got %.2f", expectedWidth, size.Width)
	}
	if math.Abs(size.Height-expectedHeight) > 0.01 {
		t.Errorf("Border-box auto height: expected %.2f, got %.2f", expectedHeight, size.Height)
	}
}

// TestBoxSizingMinMaxContentBox tests min/max constraints with content-box
func TestBoxSizingMinMaxContentBox(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:     Px(300), // Will be clamped by MaxWidth
			Height:    Px(100),
			MinWidth:  Px(100),
			MaxWidth:  Px(200),
			Padding:   Uniform(Px(10)),
			Border:    Uniform(Px(5)),
			BoxSizing: BoxSizingContentBox,
		},
		Children: []*Node{},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	// MaxWidth=200 means content max = 200, total = 200 + 20 + 10 = 230
	expectedWidth := 200.0 + 10*2 + 5*2
	expectedHeight := 100.0 + 10*2 + 5*2

	if math.Abs(size.Width-expectedWidth) > 0.01 {
		t.Errorf("Content-box with MaxWidth: expected %.2f, got %.2f", expectedWidth, size.Width)
	}
	if math.Abs(size.Height-expectedHeight) > 0.01 {
		t.Errorf("Content-box height: expected %.2f, got %.2f", expectedHeight, size.Height)
	}
}

// TestBoxSizingMinMaxBorderBox tests min/max constraints with border-box
func TestBoxSizingMinMaxBorderBox(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:     Px(300), // Will be clamped by MaxWidth
			Height:    Px(100),
			MinWidth:  Px(100),
			MaxWidth:  Px(200),
			Padding:   Uniform(Px(10)),
			Border:    Uniform(Px(5)),
			BoxSizing: BoxSizingBorderBox,
		},
		Children: []*Node{},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	// MaxWidth=200 in border-box means total max = 200
	// Content max = 200 - 20 - 10 = 170
	expectedWidth := 200.0  // Total size equals MaxWidth
	expectedHeight := 100.0 // Total size equals specified height

	if math.Abs(size.Width-expectedWidth) > 0.01 {
		t.Errorf("Border-box with MaxWidth: expected %.2f, got %.2f", expectedWidth, size.Width)
	}
	if math.Abs(size.Height-expectedHeight) > 0.01 {
		t.Errorf("Border-box height: expected %.2f, got %.2f", expectedHeight, size.Height)
	}
}

// TestBoxSizingAspectRatioContentBox tests aspect ratio with content-box
func TestBoxSizingAspectRatioContentBox(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:       Px(200),
			Height:      Px(-1), // auto, will be calculated from aspect ratio
			AspectRatio: 16.0 / 9.0,
			Padding:     Uniform(Px(10)),
			Border:      Uniform(Px(5)),
			BoxSizing:   BoxSizingContentBox,
		},
		Children: []*Node{},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	// Content width = 200, content height = 200 / (16/9) = 112.5
	// Total width = 200 + 20 + 10 = 230
	// Total height = 112.5 + 20 + 10 = 142.5
	expectedContentHeight := 200.0 / (16.0 / 9.0)
	expectedWidth := 200.0 + 10*2 + 5*2
	expectedHeight := expectedContentHeight + 10*2 + 5*2

	if math.Abs(size.Width-expectedWidth) > 0.01 {
		t.Errorf("Content-box aspect ratio width: expected %.2f, got %.2f", expectedWidth, size.Width)
	}
	if math.Abs(size.Height-expectedHeight) > 0.01 {
		t.Errorf("Content-box aspect ratio height: expected %.2f, got %.2f", expectedHeight, size.Height)
	}
}

// TestBoxSizingAspectRatioBorderBox tests aspect ratio with border-box
func TestBoxSizingAspectRatioBorderBox(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:       Px(200),
			Height:      Px(-1), // auto, will be calculated from aspect ratio
			AspectRatio: 16.0 / 9.0,
			Padding:     Uniform(Px(10)),
			Border:      Uniform(Px(5)),
			BoxSizing:   BoxSizingBorderBox,
		},
		Children: []*Node{},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutBlock(root, constraints, ctx)

	// Total width = 200 (border-box), content width = 200 - 20 - 10 = 170
	// Content height = 170 / (16/9) = 95.625
	// Total height = 95.625 + 20 + 10 = 125.625
	expectedContentWidth := 200.0 - 10*2 - 5*2
	expectedContentHeight := expectedContentWidth / (16.0 / 9.0)
	expectedWidth := 200.0
	expectedHeight := expectedContentHeight + 10*2 + 5*2

	if math.Abs(size.Width-expectedWidth) > 0.01 {
		t.Errorf("Border-box aspect ratio width: expected %.2f, got %.2f", expectedWidth, size.Width)
	}
	if math.Abs(size.Height-expectedHeight) > 0.01 {
		t.Errorf("Border-box aspect ratio height: expected %.2f, got %.2f", expectedHeight, size.Height)
	}
}

// TestBoxSizingNoPaddingBorder tests box-sizing with no padding or border
// Both should behave the same when there's no padding/border
func TestBoxSizingNoPaddingBorder(t *testing.T) {
	// Content-box
	root1 := &Node{
		Style: Style{
			Width:     Px(100),
			Height:    Px(100),
			BoxSizing: BoxSizingContentBox,
		},
		Children: []*Node{},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(1920, 1080, 16)
	size1 := LayoutBlock(root1, constraints, ctx)

	// Border-box
	root2 := &Node{
		Style: Style{
			Width:     Px(100),
			Height:    Px(100),
			BoxSizing: BoxSizingBorderBox,
		},
		Children: []*Node{},
	}

	size2 := LayoutBlock(root2, constraints, ctx)

	// Both should be the same when there's no padding/border
	if math.Abs(size1.Width-size2.Width) > 0.01 {
		t.Errorf("No padding/border: content-box width %.2f != border-box width %.2f", size1.Width, size2.Width)
	}
	if math.Abs(size1.Height-size2.Height) > 0.01 {
		t.Errorf("No padding/border: content-box height %.2f != border-box height %.2f", size1.Height, size2.Height)
	}
}

// TestBoxSizingChildItems tests that child items in flexbox/grid respect box-sizing
func TestBoxSizingChildItems(t *testing.T) {
	// Test with flexbox
	child1 := &Node{
		Style: Style{
			Width:     Px(100),
			Height:    Px(50),
			Padding:   Uniform(Px(5)),
			Border:    Uniform(Px(2)),
			BoxSizing: BoxSizingContentBox,
		},
	}

	child2 := &Node{
		Style: Style{
			Width:     Px(100),
			Height:    Px(50),
			Padding:   Uniform(Px(5)),
			Border:    Uniform(Px(2)),
			BoxSizing: BoxSizingBorderBox,
		},
	}

	root := HStack(child1, child2)
	root.Style.Padding = Uniform(Px(10))

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(1920, 1080, 16)
	Layout(root, constraints, ctx)

	// Child1 (content-box): content=100, total=100+10+4=114
	// Child2 (border-box): total=100, content=100-10-4=86
	expectedChild1Width := 100.0 + 5*2 + 2*2
	expectedChild2Width := 100.0

	if math.Abs(child1.Rect.Width-expectedChild1Width) > 0.01 {
		t.Errorf("Flexbox child1 (content-box) width: expected %.2f, got %.2f", expectedChild1Width, child1.Rect.Width)
	}
	if math.Abs(child2.Rect.Width-expectedChild2Width) > 0.01 {
		t.Errorf("Flexbox child2 (border-box) width: expected %.2f, got %.2f", expectedChild2Width, child2.Rect.Width)
	}
}
