package layout

import "testing"

// Regression tests for vertical writing modes in block layout
// (CSS Writing Modes Level 3 §7, https://www.w3.org/TR/css-writing-modes-3/#vertical-layout).

// TestBlockFixVerticalInlineConstraintUsesHeight: in vertical-rl the inline
// axis is the physical height, so children are constrained by the container's
// height, not its width.
func TestBlockFixVerticalInlineConstraintUsesHeight(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayBlock, WritingMode: WritingModeVerticalRL, Width: Px(300), Height: Px(100)},
		Children: []*Node{
			{
				Style: Style{Display: DisplayBlock, Width: Px(50)},
				Children: []*Node{
					{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(250)}},
				},
			},
		},
	}
	Layout(root, Tight(300, 100), NewLayoutContext(500, 500, 16))
	if got := root.Children[0].Rect.Height; got != 100 {
		t.Fatalf("child auto height in vertical-rl 300x100 container: expected clamp to 100, got %.2f", got)
	}
}

// TestBlockFixVerticalRLPositionsAgainstFinalWidth: vertical-rl children are
// placed from the right edge of the container's resolved width, not from the
// available width handed down in the constraints.
func TestBlockFixVerticalRLPositionsAgainstFinalWidth(t *testing.T) {
	ctx := NewLayoutContext(1000, 1000, 16)
	root := &Node{
		Style: Style{Display: DisplayBlock, WritingMode: WritingModeVerticalRL, Width: Px(300), Height: Px(100)},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(50)}},
			{Style: Style{Display: DisplayBlock, Width: Px(60), Height: Px(50)}},
		},
	}
	Layout(root, Loose(1000, 1000), ctx)
	if got := root.Children[0].Rect.X; got != 250 {
		t.Errorf("first child X: expected 250 (300-50), got %.2f", got)
	}
	if got := root.Children[1].Rect.X; got != 190 {
		t.Errorf("second child X: expected 190 (300-50-60), got %.2f", got)
	}

	// Auto width: the block size is the sum of the children, and children
	// are mirrored against it.
	root = &Node{
		Style: Style{Display: DisplayBlock, WritingMode: WritingModeVerticalRL, Height: Px(100)},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(50)}},
			{Style: Style{Display: DisplayBlock, Width: Px(60), Height: Px(50)}},
		},
	}
	size := Layout(root, Loose(1000, 100), ctx)
	if size.Width != 110 {
		t.Fatalf("auto width vertical-rl: expected 110, got %.2f", size.Width)
	}
	if got := root.Children[0].Rect.X; got != 60 {
		t.Errorf("first child X: expected 60 (110-50), got %.2f", got)
	}
	if got := root.Children[1].Rect.X; got != 0 {
		t.Errorf("second child X: expected 0, got %.2f", got)
	}
}
