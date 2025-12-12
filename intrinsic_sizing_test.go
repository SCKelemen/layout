package layout

import "testing"

// TestMinContentWidthBlock tests min-content width on block layout
func TestMinContentWidthBlock(t *testing.T) {
	container := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   SizeMinContent,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 200, Height: 50}},
			{Style: Style{Width: 150, Height: 50}},
		},
	}

	size := LayoutBlock(container, Loose(500, 500))

	// Min-content for block should be max of children's widths
	// Expected: 200 (widest child)
	if size.Width < 200 || size.Width > 250 {
		t.Errorf("Min-content block width should be around 200, got %.2f", size.Width)
	}
}

// TestMaxContentWidthBlock tests max-content width on block layout
func TestMaxContentWidthBlock(t *testing.T) {
	container := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   SizeMaxContent,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 200, Height: 50}},
			{Style: Style{Width: 150, Height: 50}},
		},
	}

	size := LayoutBlock(container, Loose(500, 500))

	// Max-content for block should be max of children's widths
	// Expected: 200 (widest child)
	if size.Width < 200 || size.Width > 250 {
		t.Errorf("Max-content block width should be around 200, got %.2f", size.Width)
	}
}

// TestFitContentWidthBlock tests fit-content width on block layout
func TestFitContentWidthBlock(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:         DisplayBlock,
			Width:           SizeFitContent,
			FitContentWidth: 150,
		},
		Children: []*Node{
			{Style: Style{Width: 200, Height: 50}}, // Exceeds fit-content limit
		},
	}

	size := LayoutBlock(container, Loose(500, 500))

	// Fit-content should clamp to FitContentWidth (150)
	if size.Width < 145 || size.Width > 155 {
		t.Errorf("Fit-content block width should be around 150, got %.2f", size.Width)
	}
}

// TestMinContentWidthFlexRow tests min-content width on flex row
func TestMinContentWidthFlexRow(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			Width:         SizeMinContent,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 150, Height: 50}},
		},
	}

	size := LayoutFlexbox(container, Loose(500, 500))

	// Min-content for flex row should be sum of children's widths
	// Expected: 100 + 150 = 250
	if size.Width < 250 || size.Width > 260 {
		t.Errorf("Min-content flex row width should be around 250, got %.2f", size.Width)
	}
}

// TestMaxContentWidthFlexRow tests max-content width on flex row
func TestMaxContentWidthFlexRow(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			Width:         SizeMaxContent,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 150, Height: 50}},
		},
	}

	size := LayoutFlexbox(container, Loose(500, 500))

	// Max-content for flex row should be sum of children's widths
	// Expected: 100 + 150 = 250
	if size.Width < 250 || size.Width > 260 {
		t.Errorf("Max-content flex row width should be around 250, got %.2f", size.Width)
	}
}

// TestMinContentWidthFlexColumn tests min-content width on flex column
func TestMinContentWidthFlexColumn(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
			Width:         SizeMinContent,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 200, Height: 50}},
			{Style: Style{Width: 150, Height: 50}},
		},
	}

	size := LayoutFlexbox(container, Loose(500, 500))

	// Min-content for flex column should be max of children's widths
	// Expected: 200 (widest child)
	if size.Width < 200 || size.Width > 210 {
		t.Errorf("Min-content flex column width should be around 200, got %.2f", size.Width)
	}
}

// TestMinContentTrack tests MinContentTrack helper for grid
func TestMinContentTrack(t *testing.T) {
	track := MinContentTrack()

	if track.MaxSize != SizeMinContent {
		t.Errorf("MinContentTrack should have MaxSize = SizeMinContent, got %.2f", track.MaxSize)
	}
	if track.MinSize != 0 {
		t.Errorf("MinContentTrack should have MinSize = 0, got %.2f", track.MinSize)
	}
}

// TestMaxContentTrack tests MaxContentTrack helper for grid
func TestMaxContentTrack(t *testing.T) {
	track := MaxContentTrack()

	if track.MaxSize != SizeMaxContent {
		t.Errorf("MaxContentTrack should have MaxSize = SizeMaxContent, got %.2f", track.MaxSize)
	}
	if track.MinSize != 0 {
		t.Errorf("MaxContentTrack should have MinSize = 0, got %.2f", track.MinSize)
	}
}

// TestFitContentTrack tests FitContentTrack helper for grid
func TestFitContentTrack(t *testing.T) {
	track := FitContentTrack(300)

	if track.MaxSize != 300 {
		t.Errorf("FitContentTrack should have MaxSize = 300, got %.2f", track.MaxSize)
	}
	if track.Fraction != -1 {
		t.Errorf("FitContentTrack should have Fraction = -1, got %.2f", track.Fraction)
	}
}

// TestIntrinsicSizingAPIHelpers tests the API helper functions
func TestIntrinsicSizingAPIHelpers(t *testing.T) {
	node := &Node{}

	// Test MinContentWidth
	MinContentWidth(node)
	if node.Style.Width != SizeMinContent {
		t.Errorf("MinContentWidth should set Width to SizeMinContent")
	}

	// Test MaxContentWidth
	node2 := &Node{}
	MaxContentWidth(node2)
	if node2.Style.Width != SizeMaxContent {
		t.Errorf("MaxContentWidth should set Width to SizeMaxContent")
	}

	// Test FitContentWidth
	node3 := &Node{}
	FitContentWidth(node3, 500)
	if node3.Style.Width != SizeFitContent {
		t.Errorf("FitContentWidth should set Width to SizeFitContent")
	}
	if node3.Style.FitContentWidth != 500 {
		t.Errorf("FitContentWidth should set FitContentWidth to 500")
	}
}

// TestWidthSizingEnumField tests using the WidthSizing enum field
func TestWidthSizingEnumField(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:     DisplayBlock,
			WidthSizing: IntrinsicSizeMinContent,
		},
		Children: []*Node{
			{Style: Style{Width: 150, Height: 50}},
		},
	}

	size := LayoutBlock(container, Loose(500, 500))

	// Should use min-content sizing via enum field
	if size.Width < 145 || size.Width > 160 {
		t.Errorf("WidthSizing enum should work for min-content, got %.2f", size.Width)
	}
}

// TestNestedIntrinsicSizing tests intrinsic sizing with nested containers
func TestNestedIntrinsicSizing(t *testing.T) {
	innerContainer := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   SizeMaxContent,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
		},
	}

	outerContainer := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   SizeMinContent,
		},
		Children: []*Node{innerContainer},
	}

	size := LayoutBlock(outerContainer, Loose(500, 500))

	// Outer should size based on inner's max-content
	if size.Width < 95 || size.Width > 110 {
		t.Errorf("Nested intrinsic sizing should work, got %.2f", size.Width)
	}
}
