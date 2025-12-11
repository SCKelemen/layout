package layout

import (
	"math"
	"testing"
)

// TestFlexboxAlignContentImplementation tests align-content for multi-line flexbox
func TestFlexboxAlignContentImplementation(t *testing.T) {
	tests := []struct {
		name            string
		alignContent    AlignContent
		containerHeight float64
		expectedFirstY  float64
	}{
		{
			name:            "stretch (default)",
			alignContent:    AlignContentStretch,
			containerHeight: 300,
			expectedFirstY:  0, // Lines stretch to fill
		},
		{
			name:            "flex-start",
			alignContent:    AlignContentFlexStart,
			containerHeight: 300,
			expectedFirstY:  0,
		},
		{
			name:            "flex-end",
			alignContent:    AlignContentFlexEnd,
			containerHeight: 300,
			expectedFirstY:  200, // 300 - (50*2 + gap) ≈ 200 (with 50px lines)
		},
		{
			name:            "center",
			alignContent:    AlignContentCenter,
			containerHeight: 300,
			expectedFirstY:  100, // (300 - 100) / 2 = 100 (approximately)
		},
		{
			name:            "space-between",
			alignContent:    AlignContentSpaceBetween,
			containerHeight: 300,
			expectedFirstY:  0, // First line at start
		},
		{
			name:            "space-around",
			alignContent:    AlignContentSpaceAround,
			containerHeight: 300,
			expectedFirstY:  50, // (300 - 100) / 4 = 50 (approximately)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Node{
				Style: Style{
					Display:       DisplayFlex,
					FlexDirection: FlexDirectionRow,
					FlexWrap:      FlexWrapWrap,
					AlignContent:  tt.alignContent,
					Height:        tt.containerHeight,
					Width:         100, // Force wrapping
				},
				Children: []*Node{
					{Style: Style{Width: 60, Height: 50}}, // 60px to force wrapping
					{Style: Style{Width: 60, Height: 50}},
				},
			}

			constraints := Loose(100, tt.containerHeight)
			LayoutFlexbox(root, constraints)

			if len(root.Children) != 2 {
				t.Fatalf("Expected 2 children, got %d", len(root.Children))
			}

			firstChild := root.Children[0]
			actualY := firstChild.Rect.Y - root.Style.Padding.Top - root.Style.Border.Top

			if math.Abs(actualY-tt.expectedFirstY) > 1.0 {
				t.Errorf("Expected first child Y %.2f, got %.2f", tt.expectedFirstY, actualY)
			}
		})
	}
}

// TestFlexboxFlexDirectionReverse tests flex-direction reverse
func TestFlexboxFlexDirectionReverse(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRowReverse,
			Width:         300,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}}, // Item 1
			{Style: Style{Width: 50, Height: 50}}, // Item 2
			{Style: Style{Width: 50, Height: 50}}, // Item 3
		},
	}

	constraints := Loose(300, Unbounded)
	LayoutFlexbox(root, constraints)

	// With row-reverse, items are visually reversed and positioned from the end
	// In CSS, row-reverse reverses the main axis direction, so flex-start is on the right
	// Items are also reversed in order, so item1 (first in DOM) appears at rightmost position
	// With 3 items of 50px each in 300px container and flex-start (default):
	// Items are positioned from right: item1 at 250, item2 at 200, item3 at 150
	item1X := root.Children[0].Rect.X - root.Style.Padding.Left - root.Style.Border.Left
	item2X := root.Children[1].Rect.X - root.Style.Padding.Left - root.Style.Border.Left
	item3X := root.Children[2].Rect.X - root.Style.Padding.Left - root.Style.Border.Left

	// With row-reverse, items are positioned from right to left
	// Item 1 (first in original array) should be at rightmost position
	expectedItem1X := 250.0 // Rightmost (300 - 50)
	expectedItem2X := 200.0 // Middle (250 - 50)
	expectedItem3X := 150.0 // Leftmost (200 - 50)

	if math.Abs(item1X-expectedItem1X) > 1.0 {
		t.Errorf("Item 1 (first in array) should be at X %.2f (rightmost), got %.2f", expectedItem1X, item1X)
	}
	if math.Abs(item2X-expectedItem2X) > 1.0 {
		t.Errorf("Item 2 should be at X %.2f, got %.2f", expectedItem2X, item2X)
	}
	if math.Abs(item3X-expectedItem3X) > 1.0 {
		t.Errorf("Item 3 should be at X %.2f (leftmost), got %.2f", expectedItem3X, item3X)
	}
}

// TestFlexboxFlexWrapReverse tests flex-wrap reverse
func TestFlexboxFlexWrapReverse(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			FlexWrap:      FlexWrapWrapReverse,
			Width:         100, // Force wrapping
			Height:        200,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}}, // Line 1
			{Style: Style{Width: 50, Height: 50}}, // Line 1
			{Style: Style{Width: 50, Height: 50}}, // Line 2
		},
	}

	constraints := Loose(100, 200)
	LayoutFlexbox(root, constraints)

	// With wrap-reverse, lines are visually reversed
	// Original: Line 1 (items 1,2), Line 2 (item 3)
	// Reversed: Line 1 visual (item 3), Line 2 visual (items 1,2)
	// So item 3 should be at Y=0, items 1,2 should be at Y=50 (line height)
	item1Y := root.Children[0].Rect.Y - root.Style.Padding.Top - root.Style.Border.Top
	item3Y := root.Children[2].Rect.Y - root.Style.Padding.Top - root.Style.Border.Top

	// With wrap-reverse, last line (in original order) becomes first visually
	expectedItem3Y := 0.0  // First line visually (was last originally)
	expectedItem1Y := 50.0 // Second line visually (was first originally, 50px line height)

	if math.Abs(item3Y-expectedItem3Y) > 1.0 {
		t.Errorf("Item 3 (last in array, first line) should be at Y %.2f, got %.2f", expectedItem3Y, item3Y)
	}
	if math.Abs(item1Y-expectedItem1Y) > 1.0 {
		t.Errorf("Item 1 (first in array, second line) should be at Y %.2f, got %.2f", expectedItem1Y, item1Y)
	}
}

// TestFlexboxGap tests flex gap support
func TestFlexboxGap(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			FlexGap:       20, // 20px gap between items
			Width:         200,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}},
			{Style: Style{Width: 50, Height: 50}},
			{Style: Style{Width: 50, Height: 50}},
		},
	}

	constraints := Loose(200, Unbounded)
	LayoutFlexbox(root, constraints)

	// Items should have 20px gaps between them
	// Item 1 at X = 0, Item 2 at X = 70 (50 + 20), Item 3 at X = 140 (50 + 20 + 50 + 20)
	item1X := root.Children[0].Rect.X - root.Style.Padding.Left - root.Style.Border.Left
	item2X := root.Children[1].Rect.X - root.Style.Padding.Left - root.Style.Border.Left
	item3X := root.Children[2].Rect.X - root.Style.Padding.Left - root.Style.Border.Left

	expectedItem1X := 0.0
	expectedItem2X := 70.0  // 50 (item1) + 20 (gap)
	expectedItem3X := 140.0 // 50 + 20 + 50 + 20

	if math.Abs(item1X-expectedItem1X) > 1.0 {
		t.Errorf("Item 1 X should be %.2f, got %.2f", expectedItem1X, item1X)
	}
	if math.Abs(item2X-expectedItem2X) > 1.0 {
		t.Errorf("Item 2 X should be %.2f (with 20px gap), got %.2f", expectedItem2X, item2X)
	}
	if math.Abs(item3X-expectedItem3X) > 1.0 {
		t.Errorf("Item 3 X should be %.2f (with 20px gaps), got %.2f", expectedItem3X, item3X)
	}
}

// TestFlexboxRowGapAndColumnGap tests separate row and column gap
func TestFlexboxRowGapAndColumnGap(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			FlexWrap:      FlexWrapWrap,
			FlexRowGap:    30,  // 30px between rows
			FlexColumnGap: 40,  // 40px between columns
			Width:         100, // Force wrapping
			Height:        200,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}}, // Line 1
			{Style: Style{Width: 50, Height: 50}}, // Line 1
			{Style: Style{Width: 50, Height: 50}}, // Line 2
		},
	}

	constraints := Loose(100, 200)
	LayoutFlexbox(root, constraints)

	// Items in same line should have 40px gap (column gap)
	item2X := root.Children[1].Rect.X - root.Style.Padding.Left - root.Style.Border.Left

	// Items in different lines should have 30px gap (row gap)
	item3Y := root.Children[2].Rect.Y - root.Style.Padding.Top - root.Style.Border.Top

	expectedItem2X := 90.0 // 50 + 40 (column gap)
	expectedItem3Y := 80.0 // 50 + 30 (row gap)

	if math.Abs(item2X-expectedItem2X) > 1.0 {
		t.Errorf("Item 2 X should be %.2f (with 40px column gap), got %.2f", expectedItem2X, item2X)
	}
	if math.Abs(item3Y-expectedItem3Y) > 1.0 {
		t.Errorf("Item 3 Y should be %.2f (with 30px row gap), got %.2f", expectedItem3Y, item3Y)
	}
}
