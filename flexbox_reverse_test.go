package layout

import (
	"testing"
)

// TestFlexDirectionRowReverseWithGap tests row-reverse with gaps
func TestFlexDirectionRowReverseWithGap(t *testing.T) {
	// Test that gaps are correctly positioned in reverse direction
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRowReverse,
			FlexGap:       Px(10),
			Width:         Px(200),
			Height:        Px(100),
		},
		Children: []*Node{
			{Style: Style{Width: Px(40), Height: Px(40)}},
			{Style: Style{Width: Px(40), Height: Px(40)}},
			{Style: Style{Width: Px(40), Height: Px(40)}},
		},
	}

	constraints := Loose(200, 100)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// In row-reverse, items should be positioned from right to left
	// With justify-content: flex-start (default in reverse = flex-end in normal)
	// Last child (index 2) should be leftmost after reversal
	// Items are: [0][gap][1][gap][2]
	// After reverse: [2][gap][1][gap][0]
	// In RTL layout, this means item 0 should be rightmost

	// Check that items are in reverse visual order
	// Item 0 should be right of item 1
	if root.Children[0].Rect.X <= root.Children[1].Rect.X {
		t.Errorf("Row-reverse: Item 0 should be right of item 1. Item 0 X: %.2f, Item 1 X: %.2f",
			root.Children[0].Rect.X, root.Children[1].Rect.X)
	}

	// Item 1 should be right of item 2
	if root.Children[1].Rect.X <= root.Children[2].Rect.X {
		t.Errorf("Row-reverse: Item 1 should be right of item 2. Item 1 X: %.2f, Item 2 X: %.2f",
			root.Children[1].Rect.X, root.Children[2].Rect.X)
	}

	// Gaps should be present between items
	// Distance between items should be 40 (width) + 10 (gap) = 50
	gap1 := root.Children[1].Rect.X - root.Children[2].Rect.X
	if gap1 < 45 || gap1 > 55 {
		t.Errorf("Gap between item 2 and 1: expected ~50, got %.2f", gap1)
	}
}

// TestFlexDirectionColumnReverseWithGap tests column-reverse with gaps
func TestFlexDirectionColumnReverseWithGap(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumnReverse,
			FlexGap:       Px(15),
			Width:         Px(100),
			Height:        Px(300),
		},
		Children: []*Node{
			{Style: Style{Width: Px(40), Height: Px(50)}},
			{Style: Style{Width: Px(40), Height: Px(50)}},
			{Style: Style{Width: Px(40), Height: Px(50)}},
		},
	}

	constraints := Loose(100, 300)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// In column-reverse, items should be positioned from bottom to top
	// Item 0 should be below item 1
	if root.Children[0].Rect.Y <= root.Children[1].Rect.Y {
		t.Errorf("Column-reverse: Item 0 should be below item 1. Item 0 Y: %.2f, Item 1 Y: %.2f",
			root.Children[0].Rect.Y, root.Children[1].Rect.Y)
	}

	// Item 1 should be below item 2
	if root.Children[1].Rect.Y <= root.Children[2].Rect.Y {
		t.Errorf("Column-reverse: Item 1 should be below item 2. Item 1 Y: %.2f, Item 2 Y: %.2f",
			root.Children[1].Rect.Y, root.Children[2].Rect.Y)
	}
}

// TestFlexDirectionReverseWithJustifyContent tests reverse with different justify-content values
func TestFlexDirectionReverseWithJustifyContent(t *testing.T) {
	testCases := []struct {
		name    string
		justify JustifyContent
	}{
		{"flex-start", JustifyContentFlexStart},
		{"flex-end", JustifyContentFlexEnd},
		{"center", JustifyContentCenter},
		{"space-between", JustifyContentSpaceBetween},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := &Node{
				Style: Style{
					Display:        DisplayFlex,
					FlexDirection:  FlexDirectionRowReverse,
					JustifyContent: tc.justify,
					Width:          Px(300),
					Height:         Px(100),
				},
				Children: []*Node{
					{Style: Style{Width: Px(50), Height: Px(40)}},
					{Style: Style{Width: Px(50), Height: Px(40)}},
				},
			}

			constraints := Loose(300, 100)
			ctx := NewLayoutContext(1920, 1080, 16)
			LayoutFlexbox(root, constraints, ctx)

			// Should layout without errors
			// Item 0 should be to the right of item 1 (reversed order)
			if root.Children[0].Rect.X <= root.Children[1].Rect.X {
				t.Errorf("%s: Items should be in reverse order. Item 0 X: %.2f, Item 1 X: %.2f",
					tc.name, root.Children[0].Rect.X, root.Children[1].Rect.X)
			}
		})
	}
}

// TestWrapReverseWithoutExplicitCrossSize tests wrap-reverse without explicit height
func TestWrapReverseWithoutExplicitCrossSize(t *testing.T) {
	// This should now work after removing the hasExplicitCrossSize gate
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			FlexWrap:      FlexWrapWrapReverse,
			Width:         Px(100),
			Height:        Px(-1), // Auto height (no explicit cross size)
		},
		Children: []*Node{
			{Style: Style{Width: Px(60), Height: Px(30)}},
			{Style: Style{Width: Px(60), Height: Px(30)}},
			{Style: Style{Width: Px(60), Height: Px(30)}},
		},
	}

	constraints := Loose(100, 300)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutFlexbox(root, constraints, ctx)

	// Should successfully layout with wrap-reverse
	// Container should have a height (sum of lines)
	if size.Height <= 0 {
		t.Errorf("Wrap-reverse without explicit cross size: expected height > 0, got %.2f", size.Height)
	}

	// Items should wrap (3 items of 60px width in 100px container = 2 lines minimum)
	// At least one item should be on a different line
	hasMultipleLines := false
	firstItemY := root.Children[0].Rect.Y
	for i := 1; i < len(root.Children); i++ {
		if root.Children[i].Rect.Y != firstItemY {
			hasMultipleLines = true
			break
		}
	}

	if !hasMultipleLines {
		t.Error("Wrap-reverse: items should wrap to multiple lines")
	}
}

// TestWrapReverseWithExplicitCrossSize tests wrap-reverse with explicit height
func TestWrapReverseWithExplicitCrossSize(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			FlexWrap:      FlexWrapWrapReverse,
			Width:         Px(100),
			Height:        Px(200), // Explicit height
		},
		Children: []*Node{
			{Style: Style{Width: Px(60), Height: Px(50)}},
			{Style: Style{Width: Px(60), Height: Px(50)}},
			{Style: Style{Width: Px(60), Height: Px(50)}},
		},
	}

	constraints := Loose(100, 200)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// In wrap-reverse with explicit cross size, lines should be reversed
	// Items should wrap and be positioned in reverse line order
	hasMultipleLines := false
	firstItemY := root.Children[0].Rect.Y
	for i := 1; i < len(root.Children); i++ {
		if root.Children[i].Rect.Y != firstItemY {
			hasMultipleLines = true
			break
		}
	}

	if !hasMultipleLines {
		t.Error("Wrap-reverse with explicit cross size: items should wrap")
	}
}

// TestWrapReverseColumnDirection tests wrap-reverse in column direction
func TestWrapReverseColumnDirection(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
			FlexWrap:      FlexWrapWrapReverse,
			Width:         Px(200),
			Height:        Px(100),
		},
		Children: []*Node{
			{Style: Style{Width: Px(40), Height: Px(60)}},
			{Style: Style{Width: Px(40), Height: Px(60)}},
			{Style: Style{Width: Px(40), Height: Px(60)}},
		},
	}

	constraints := Loose(200, 100)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// In column direction with wrap-reverse, items should wrap horizontally
	// and lines should be reversed
	hasMultipleLines := false
	firstItemX := root.Children[0].Rect.X
	for i := 1; i < len(root.Children); i++ {
		if root.Children[i].Rect.X != firstItemX {
			hasMultipleLines = true
			break
		}
	}

	if !hasMultipleLines {
		t.Error("Wrap-reverse column: items should wrap to multiple columns")
	}
}

// TestReverseDirectionWithMargins tests reverse direction with margins
func TestReverseDirectionWithMargins(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRowReverse,
			Width:         Px(300),
			Height:        Px(100),
		},
		Children: []*Node{
			{
				Style: Style{
					Width:  Px(50),
					Height: Px(40),
					Margin: Spacing{Left: Px(10), Right: Px(10)},
				},
			},
			{
				Style: Style{
					Width:  Px(50),
					Height: Px(40),
					Margin: Spacing{Left: Px(15), Right: Px(15)},
				},
			},
		},
	}

	constraints := Loose(300, 100)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// Margins should be respected in reverse direction
	// Item 0 should be to the right of item 1
	if root.Children[0].Rect.X <= root.Children[1].Rect.X {
		t.Errorf("Reverse with margins: Item 0 should be right of item 1. Item 0 X: %.2f, Item 1 X: %.2f",
			root.Children[0].Rect.X, root.Children[1].Rect.X)
	}
}
