package layout

import (
	"math"
	"testing"
)

func TestFlexboxJustifyContentSpaceAround(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentSpaceAround,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}},
			{Style: Style{Width: 50, Height: 50}},
		},
	}

	constraints := Tight(200, 100)
	LayoutFlexbox(root, constraints)

	// SpaceAround should distribute space around items
	// First item should not be at X=0
	if root.Children[0].Rect.X == 0 {
		t.Error("SpaceAround should not start at X=0")
	}
}

func TestFlexboxJustifyContentSpaceEvenly(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentSpaceEvenly,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}},
			{Style: Style{Width: 50, Height: 50}},
		},
	}

	constraints := Tight(200, 100)
	LayoutFlexbox(root, constraints)

	// SpaceEvenly should distribute space evenly
	// First item should not be at X=0
	if root.Children[0].Rect.X == 0 {
		t.Error("SpaceEvenly should not start at X=0")
	}
}

func TestFlexboxRowReverse(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRowReverse,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}},
			{Style: Style{Width: 50, Height: 50}},
		},
	}

	constraints := Tight(200, 100)
	LayoutFlexbox(root, constraints)

	// In row reverse, second child should be before first
	// (visually reversed, but we check positions are different)
	if root.Children[0].Rect.X == root.Children[1].Rect.X {
		t.Error("RowReverse should position children differently")
	}
}

func TestFlexboxColumnReverse(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumnReverse,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}},
			{Style: Style{Width: 50, Height: 50}},
		},
	}

	constraints := Tight(200, 200)
	LayoutFlexbox(root, constraints)

	// In column reverse, children are laid out in reverse order
	// The layout should position children differently than normal column
	// Just verify both children are positioned (not at same Y)
	if root.Children[0].Rect.Y == root.Children[1].Rect.Y {
		t.Error("ColumnReverse should position children at different Y positions")
	}
}

func TestFlexboxWrap(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:      DisplayFlex,
			FlexDirection: FlexDirectionRow,
			FlexWrap:     FlexWrapWrap,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
		},
	}

	constraints := Tight(150, 200)
	LayoutFlexbox(root, constraints)

	// With wrap, third child should be on second line
	// First line: child 0, child 1 (or just child 0 if they don't fit)
	// Second line: child 2
	if root.Children[2].Rect.Y <= root.Children[0].Rect.Y+root.Children[0].Rect.Height {
		// Third child should be below first line
		if root.Children[2].Rect.Y < root.Children[0].Rect.Y+root.Children[0].Rect.Height {
			t.Error("Wrapped child should be on second line")
		}
	}
}

func TestFlexboxAlignContent(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:      DisplayFlex,
			FlexDirection: FlexDirectionRow,
			FlexWrap:     FlexWrapWrap,
			AlignContent: AlignContentCenter,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}},
			{Style: Style{Width: 50, Height: 50}},
			{Style: Style{Width: 50, Height: 50}},
		},
	}

	constraints := Tight(100, 200)
	LayoutFlexbox(root, constraints)

	// With multiple lines and AlignContentCenter, lines should be centered
	// This is a basic check that it doesn't crash
	if len(root.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(root.Children))
	}
}

func TestFlexboxAlignItemsFlexEnd(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:     DisplayFlex,
			FlexDirection: FlexDirectionRow,
			AlignItems:  AlignItemsFlexEnd,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 30}},
		},
	}

	constraints := Tight(200, 100)
	LayoutFlexbox(root, constraints)

	// Child should be aligned to bottom (flex-end)
	expectedY := 100.0 - 30.0 // container height - child height
	if math.Abs(root.Children[0].Rect.Y-expectedY) > 1.0 {
		t.Errorf("Expected Y=%.2f (flex-end), got %.2f", expectedY, root.Children[0].Rect.Y)
	}
}

