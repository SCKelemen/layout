package layout

import "testing"

// Regression tests for margin collapsing (CSS 2.1 §8.3.1,
// https://www.w3.org/TR/CSS21/box.html#collapsing-margins) and for
// out-of-flow children not consuming block flow space (CSS 2.1 §9.3).

func fixMarginRoot(children ...*Node) *Node {
	return &Node{
		Style:    Style{Display: DisplayBlock, Width: Px(200)},
		Children: children,
	}
}

func fixBlock(height float64, margin Spacing) *Node {
	return &Node{Style: Style{Display: DisplayBlock, Height: Px(height), Margin: margin}}
}

// TestBlockFixNegativeMarginCollapsing: the collapsed margin is the largest
// positive margin plus the most negative margin; if all are negative, the most
// negative one wins.
func TestBlockFixNegativeMarginCollapsing(t *testing.T) {
	ctx := NewLayoutContext(500, 500, 16)

	// [mb:-10][mt:20] -> collapsed 10
	root := fixMarginRoot(
		fixBlock(50, Spacing{Bottom: Px(-10)}),
		fixBlock(50, Spacing{Top: Px(20)}),
	)
	Layout(root, Loose(500, 500), ctx)
	if got := root.Children[1].Rect.Y; got != 60 {
		t.Errorf("[mb:-10][mt:20]: expected second child Y=60 (50 + (20-10)), got %.2f", got)
	}
	if root.Rect.Height != 110 {
		t.Errorf("[mb:-10][mt:20]: expected container height 110, got %.2f", root.Rect.Height)
	}

	// [mb:-10][mt:-20] -> collapsed -20
	root = fixMarginRoot(
		fixBlock(50, Spacing{Bottom: Px(-10)}),
		fixBlock(50, Spacing{Top: Px(-20)}),
	)
	Layout(root, Loose(500, 500), ctx)
	if got := root.Children[1].Rect.Y; got != 30 {
		t.Errorf("[mb:-10][mt:-20]: expected second child Y=30 (50 - 20), got %.2f", got)
	}
	if root.Rect.Height != 80 {
		t.Errorf("[mb:-10][mt:-20]: expected container height 80, got %.2f", root.Rect.Height)
	}
}

// TestBlockFixEmptyBlockSelfCollapses: a zero-height block with no padding or
// border collapses its own top and bottom margins together with the adjoining
// sibling margins, so one 10px margin remains, not 20px.
func TestBlockFixEmptyBlockSelfCollapses(t *testing.T) {
	ctx := NewLayoutContext(500, 500, 16)

	root := fixMarginRoot(
		fixBlock(50, Spacing{}),
		fixBlock(0, Spacing{Top: Px(10), Bottom: Px(10)}),
		fixBlock(50, Spacing{}),
	)
	Layout(root, Loose(500, 500), ctx)
	if got := root.Children[2].Rect.Y; got != 60 {
		t.Errorf("empty block mt:10 mb:10: expected third child Y=60, got %.2f", got)
	}
	if got := root.Children[1].Rect.Y; got != 60 {
		t.Errorf("empty block should sit at the collapsed margin position (60), got %.2f", got)
	}
	if root.Rect.Height != 110 {
		t.Errorf("expected container height 110, got %.2f", root.Rect.Height)
	}

	// All four margins around an empty block collapse into one: max(5,10,10,8) = 10.
	root = fixMarginRoot(
		fixBlock(50, Spacing{Bottom: Px(5)}),
		fixBlock(0, Spacing{Top: Px(10), Bottom: Px(10)}),
		fixBlock(50, Spacing{Top: Px(8)}),
	)
	Layout(root, Loose(500, 500), ctx)
	if got := root.Children[2].Rect.Y; got != 60 {
		t.Errorf("four adjoining margins (5,10,10,8): expected third child Y=60, got %.2f", got)
	}

	// An empty block with padding is not self-collapsing.
	root = fixMarginRoot(
		fixBlock(50, Spacing{}),
		&Node{Style: Style{Display: DisplayBlock, Height: Px(0), Padding: Spacing{Top: Px(1)}, Margin: Spacing{Top: Px(10), Bottom: Px(10)}}},
		fixBlock(50, Spacing{}),
	)
	Layout(root, Loose(500, 500), ctx)
	if got := root.Children[2].Rect.Y; got != 71 {
		t.Errorf("empty block with 1px padding: expected third child Y=71 (50+10+1+10), got %.2f", got)
	}
}

// TestBlockFixAbsoluteChildTakesNoFlowSpace: absolutely positioned children
// are out of flow (CSS 2.1 §9.3) and must not advance the block position or
// contribute to the container's auto height.
func TestBlockFixAbsoluteChildTakesNoFlowSpace(t *testing.T) {
	ctx := NewLayoutContext(500, 500, 16)
	root := fixMarginRoot(
		fixBlock(50, Spacing{}),
		&Node{Style: Style{Display: DisplayBlock, Position: PositionAbsolute, Width: Px(30), Height: Px(30)}},
		fixBlock(50, Spacing{}),
	)
	Layout(root, Loose(500, 500), ctx)

	if got := root.Children[2].Rect.Y; got != 50 {
		t.Errorf("third child: expected Y=50, got %.2f", got)
	}
	if root.Rect.Height != 100 {
		t.Errorf("container height: expected 100, got %.2f", root.Rect.Height)
	}
	abs := root.Children[1]
	if abs.Rect.Height != 30 || abs.Rect.Width != 30 {
		t.Errorf("absolute child should still be laid out (30x30), got %.2fx%.2f", abs.Rect.Width, abs.Rect.Height)
	}
	if abs.Rect.Y != 50 {
		t.Errorf("absolute child static position: expected Y=50, got %.2f", abs.Rect.Y)
	}

	// Margins of an out-of-flow child do not collapse with siblings.
	root = fixMarginRoot(
		fixBlock(50, Spacing{Bottom: Px(10)}),
		&Node{Style: Style{Display: DisplayBlock, Position: PositionFixed, Width: Px(30), Height: Px(30), Margin: Spacing{Top: Px(100), Bottom: Px(100)}}},
		fixBlock(50, Spacing{Top: Px(20)}),
	)
	Layout(root, Loose(500, 500), ctx)
	if got := root.Children[2].Rect.Y; got != 70 {
		t.Errorf("fixed child margins must not affect siblings: expected Y=70, got %.2f", got)
	}
}
