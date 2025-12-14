package layout

import (
	"testing"
)

// TestFlexboxBaselineAlignment tests baseline alignment in flexbox
func TestFlexboxBaselineAlignment(t *testing.T) {
	// CSS Flexbox §10.3.1: Baseline alignment
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			AlignItems:    AlignItemsBaseline,
			Width:         Px(300),
			Height:        Px(100),
		},
		Children: []*Node{
			{
				Style: Style{
					Width:  Px(50),
					Height: Px(30),
				},
				Baseline: 20, // Baseline at 20px from top
			},
			{
				Style: Style{
					Width:  Px(50),
					Height: Px(40),
				},
				Baseline: 30, // Baseline at 30px from top
			},
			{
				Style: Style{
					Width:  Px(50),
					Height: Px(50),
				},
				Baseline: 25, // Baseline at 25px from top
			},
		},
	}

	constraints := Loose(300, 100)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// All items should align their baselines to the maximum baseline (30px)
	// First item: baseline at 20, needs to be offset by (30-20) = 10
	expectedY1 := 10.0
	if root.Children[0].Rect.Y != expectedY1 {
		t.Errorf("First item Y: expected %.2f (baseline offset), got %.2f", expectedY1, root.Children[0].Rect.Y)
	}

	// Second item: baseline at 30, no offset needed
	expectedY2 := 0.0
	if root.Children[1].Rect.Y != expectedY2 {
		t.Errorf("Second item Y: expected %.2f (no offset), got %.2f", expectedY2, root.Children[1].Rect.Y)
	}

	// Third item: baseline at 25, needs to be offset by (30-25) = 5
	expectedY3 := 5.0
	if root.Children[2].Rect.Y != expectedY3 {
		t.Errorf("Third item Y: expected %.2f (baseline offset), got %.2f", expectedY3, root.Children[2].Rect.Y)
	}
}

// TestFlexboxBaselineAlignmentWithMargins tests baseline alignment with margins
func TestFlexboxBaselineAlignmentWithMargins(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			AlignItems:    AlignItemsBaseline,
			Width:         Px(300),
			Height:        Px(100),
		},
		Children: []*Node{
			{
				Style: Style{
					Width:  Px(50),
					Height: Px(30),
					Margin: Spacing{Top: Px(10)},
				},
				Baseline: 20, // Baseline at 20px from content top
			},
			{
				Style: Style{
					Width:  Px(50),
					Height: Px(40),
					Margin: Spacing{Top: Px(5)},
				},
				Baseline: 30,
			},
		},
	}

	constraints := Loose(300, 100)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// Baseline calculation includes top margin
	// First item: baseline with margin = 10 + 20 = 30
	// Second item: baseline with margin = 5 + 30 = 35
	// Max baseline = 35
	// First item should be offset by (35 - 30) = 5 additional pixels beyond its margin
	expectedY1 := 15.0 // 10 (margin) + 5 (baseline offset)
	if root.Children[0].Rect.Y != expectedY1 {
		t.Errorf("First item Y: expected %.2f, got %.2f", expectedY1, root.Children[0].Rect.Y)
	}
}

// TestFlexboxBaselineAlignmentNoBaseline tests baseline alignment with items that don't have baseline set
func TestFlexboxBaselineAlignmentNoBaseline(t *testing.T) {
	// Items without baseline should use their height as default baseline (align bottom edge)
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			AlignItems:    AlignItemsBaseline,
			Width:         Px(300),
			Height:        Px(100),
		},
		Children: []*Node{
			{
				Style: Style{
					Width:  Px(50),
					Height: Px(30),
				},
				// No baseline set, defaults to 30 (bottom of box)
			},
			{
				Style: Style{
					Width:  Px(50),
					Height: Px(50),
				},
				Baseline: 25,
			},
		},
	}

	constraints := Loose(300, 100)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// First item: default baseline at 30 (height)
	// Second item: baseline at 25
	// Max baseline = 30
	// Second item needs offset of (30 - 25) = 5
	expectedY2 := 5.0
	if root.Children[1].Rect.Y != expectedY2 {
		t.Errorf("Second item Y: expected %.2f, got %.2f", expectedY2, root.Children[1].Rect.Y)
	}
}

// TestFlexboxBaselineAlignmentColumn tests baseline alignment in column direction
func TestFlexboxBaselineAlignmentColumn(t *testing.T) {
	// In column direction, baseline alignment applies to the cross axis (horizontal)
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
			AlignItems:    AlignItemsBaseline,
			Width:         Px(100),
			Height:        Px(300),
		},
		Children: []*Node{
			{
				Style: Style{
					Width:  Px(30),
					Height: Px(50),
				},
				Baseline: 10, // For column, baseline is in horizontal direction
			},
			{
				Style: Style{
					Width:  Px(40),
					Height: Px(50),
				},
				Baseline: 15,
			},
		},
	}

	constraints := Loose(100, 300)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// In column direction, baseline affects X positioning (cross axis)
	// Max baseline = 15
	// First item needs offset of (15 - 10) = 5 in X direction
	expectedX1 := 5.0
	if root.Children[0].Rect.X != expectedX1 {
		t.Errorf("First item X: expected %.2f, got %.2f", expectedX1, root.Children[0].Rect.X)
	}

	// Second item should be at X = 0 (has max baseline)
	if root.Children[1].Rect.X != 0 {
		t.Errorf("Second item X: expected 0, got %.2f", root.Children[1].Rect.X)
	}
}

// TestFlexboxBaselineAlignmentMultiLine tests baseline alignment doesn't apply across lines
func TestFlexboxBaselineAlignmentMultiLine(t *testing.T) {
	// Each line in a multi-line flex container has its own baseline alignment
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
			FlexWrap:      FlexWrapWrap,
			AlignItems:    AlignItemsBaseline,
			Width:         Px(100),
			Height:        Px(200),
		},
		Children: []*Node{
			{
				Style: Style{
					Width:  Px(50),
					Height: Px(30),
				},
				Baseline: 20,
			},
			{
				Style: Style{
					Width:  Px(50),
					Height: Px(40),
				},
				Baseline: 30,
			},
			{
				Style: Style{
					Width:  Px(50), // This wraps to second line
					Height: Px(25),
				},
				Baseline: 15,
			},
		},
	}

	constraints := Loose(100, 200)
	ctx := NewLayoutContext(1920, 1080, 16)
	LayoutFlexbox(root, constraints, ctx)

	// First two items should be baseline aligned in first line
	// Third item should be in its own line with its own baseline

	// First item offset = max(20, 30) - 20 = 10
	expectedY1 := 10.0
	if root.Children[0].Rect.Y != expectedY1 {
		t.Errorf("First item Y: expected %.2f, got %.2f", expectedY1, root.Children[0].Rect.Y)
	}

	// Second item offset = 0 (has max baseline in line)
	if root.Children[1].Rect.Y != 0 {
		t.Errorf("Second item Y: expected 0, got %.2f", root.Children[1].Rect.Y)
	}

	// Third item should be on second line (Y > first line height)
	if root.Children[2].Rect.Y <= 40 {
		t.Errorf("Third item should be on second line, Y: %.2f", root.Children[2].Rect.Y)
	}
}
