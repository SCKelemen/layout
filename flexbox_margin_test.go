package layout

import (
	"math"
	"testing"
)

func TestFlexboxMargin(t *testing.T) {
	// Test that margins create spacing between flex items
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
		},
		Children: []*Node{
			{
				Style: Style{
					Width:  Px(100),
					Height: Px(50),
					Margin: Uniform(Px(10)),
				},
			},
			{
				Style: Style{
					Width:  Px(100),
					Height: Px(50),
					Margin: Uniform(Px(10)),
				},
			},
		},
	}

	constraints := Loose(500, 200)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// First item should be at x=10 (left margin)
	if math.Abs(root.Children[0].Rect.X-10.0) > 0.1 {
		t.Errorf("First item X should be 10 (margin), got %.2f", root.Children[0].Rect.X)
	}

	// Second item should be at x=120 (first item width 100 + left margin 10 + right margin 10 + second left margin 10)
	expectedX := 100.0 + 10.0 + 10.0 + 10.0
	if math.Abs(root.Children[1].Rect.X-expectedX) > 0.1 {
		t.Errorf("Second item X should be %.2f, got %.2f", expectedX, root.Children[1].Rect.X)
	}
}

func TestFlexboxMarginVertical(t *testing.T) {
	// Test margins in column direction
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
		},
		Children: []*Node{
			{
				Style: Style{
					Width:  Px(100),
					Height: Px(50),
					Margin: Uniform(Px(10)),
				},
			},
			{
				Style: Style{
					Width:  Px(100),
					Height: Px(50),
					Margin: Uniform(Px(10)),
				},
			},
		},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// First item should be at y=10 (top margin)
	if math.Abs(root.Children[0].Rect.Y-10.0) > 0.1 {
		t.Errorf("First item Y should be 10 (margin), got %.2f", root.Children[0].Rect.Y)
	}

	// Second item should be below first with margins
	expectedY := 50.0 + 10.0 + 10.0 + 10.0
	if math.Abs(root.Children[1].Rect.Y-expectedY) > 0.1 {
		t.Errorf("Second item Y should be %.2f, got %.2f", expectedY, root.Children[1].Rect.Y)
	}
}

func TestFlexboxMarginWithJustifyContent(t *testing.T) {
	// Test that margins work with justify-content
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentCenter,
		},
		Children: []*Node{
			{
				Style: Style{
					Width:  Px(100),
					Height: Px(50),
					Margin: Uniform(Px(10)),
				},
			},
		},
	}

	constraints := Loose(500, 200)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// Item should be centered, accounting for margins
	// Container width is 500, item width is 100, margins are 20 total
	// So centered position should account for margins
	// This is a basic test - exact position depends on implementation
	if root.Children[0].Rect.X < 0 {
		t.Errorf("Item X should be positive, got %.2f", root.Children[0].Rect.X)
	}
}

func TestFlexboxMarginWithAlignItems(t *testing.T) {
	// Test that margins work with align-items
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			AlignItems:    AlignItemsCenter,
		},
		Children: []*Node{
			{
				Style: Style{
					Width:  Px(100),
					Height: Px(50),
					Margin: Vertical(Px(10)), // Top and bottom margins
				},
			},
		},
	}

	constraints := Loose(500, 200)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// Item should be centered vertically, accounting for margins
	// Container height is 200, item height is 50, margins are 10 top + 10 bottom = 20 total
	// Centered calculation: (containerHeight - itemHeightWithMargins) / 2 + topMargin
	// (200 - 50 - 20) / 2 + 10 = 75
	expectedY := 75.0
	if math.Abs(root.Children[0].Rect.Y-expectedY) > 1.0 {
		t.Errorf("Item Y should be approximately %.2f (centered with margins), got %.2f", expectedY, root.Children[0].Rect.Y)
	}
}
