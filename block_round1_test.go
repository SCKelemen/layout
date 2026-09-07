package layout

import (
	"math"
	"testing"
)

// Round-1 correctness regressions for block layout.

// assertFiniteRect fails when any Rect field is NaN, infinite, or the
// Unbounded sentinel. No layout result may carry an indefinite value.
func assertFiniteRect(t *testing.T, name string, r Rect) {
	t.Helper()
	for label, v := range map[string]float64{"X": r.X, "Y": r.Y, "Width": r.Width, "Height": r.Height} {
		if math.IsNaN(v) || math.IsInf(v, 0) || v >= Unbounded || v <= -Unbounded {
			t.Errorf("%s: %s = %v, expected a finite, definite value", name, label, v)
		}
	}
}

// TestBlockRound1AspectRatioUnboundedWidth: an auto-sized block with an
// aspect ratio and no definite available size has nothing to transfer
// through the ratio (css-sizing-4 §5.1, size transfers need a definite size),
// so it sizes to its (empty) content instead of MaxFloat64 x MaxFloat64/2.
// https://www.w3.org/TR/css-sizing-4/#aspect-ratio-size-transfers
func TestBlockRound1AspectRatioUnboundedWidth(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)
	node := &Node{Style: Style{AspectRatio: 2}}
	size := Layout(node, Unconstrained(), ctx)
	assertFiniteRect(t, "aspect-ratio block under Unconstrained()", node.Rect)
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("expected 0x0 for an empty auto block with no available size, got %vx%v", size.Width, size.Height)
	}

	// Unbounded width, bounded height: the height axis is the only definite
	// one, so the width is transferred from it.
	node = &Node{Style: Style{AspectRatio: 2}}
	size = Layout(node, Loose(Unbounded, 100), ctx)
	assertFiniteRect(t, "aspect-ratio block under Loose(Unbounded, 100)", node.Rect)
	if size.Width != 200 || size.Height != 100 {
		t.Errorf("expected 200x100 (height-based transfer), got %vx%v", size.Width, size.Height)
	}
}

// TestBlockRound1AspectRatioFlexChildUnboundedHeight: a flex container with
// an indefinite height measures its block child with unbounded constraints;
// the child's aspect ratio must not turn that into a 9e307 height.
func TestBlockRound1AspectRatioFlexChildUnboundedHeight(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)
	child := &Node{Style: Style{AspectRatio: 2}}
	root := &Node{Style: Style{Display: DisplayFlex}, Children: []*Node{child}}
	Layout(root, Loose(400, Unbounded), ctx)
	assertFiniteRect(t, "flex > block{AspectRatio: 2}", child.Rect)
	assertFiniteRect(t, "flex container", root.Rect)
}

// TestBlockRound1AspectRatioVerticalWritingModeUnboundedWidth: in a vertical
// writing mode the block axis is horizontal and children get an unbounded
// MaxWidth; an aspect-ratio descendant must not size itself to that bound.
func TestBlockRound1AspectRatioVerticalWritingModeUnboundedWidth(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)
	leaf := &Node{Style: Style{AspectRatio: 2}}
	root := &Node{Children: []*Node{
		{Style: Style{WritingMode: WritingModeSidewaysRL}, Children: []*Node{
			{Children: []*Node{leaf}},
		}},
	}}
	Layout(root, Tight(400, 300), ctx)
	assertFiniteRect(t, "aspect-ratio leaf under sideways-rl", leaf.Rect)
	assertFiniteRect(t, "root", root.Rect)
}

// TestBlockRound1NegativeWidthWithAspectRatioIsNotNaN: a negative explicit
// width is illegal in CSS (CSS 2.1 §10.2) and is treated as auto; combined
// with an aspect ratio inside a shrinking flex item it previously produced
// a NaN width.
// https://www.w3.org/TR/CSS21/visudet.html#propdef-width
func TestBlockRound1NegativeWidthWithAspectRatioIsNotNaN(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)
	child := &Node{Style: Style{FlexShrink: 2, Width: Px(-155), AspectRatio: 0.5}}
	root := &Node{Style: Style{Display: DisplayFlex, Width: Rem(9.19)}, Children: []*Node{child}}
	Layout(root, Unconstrained(), ctx)
	assertFiniteRect(t, "flex child with negative width and aspect ratio", child.Rect)
	assertFiniteRect(t, "flex container", root.Rect)

	// The same node as a plain block: negative width is auto, so the box
	// fills the available width and derives its height from the ratio.
	block := &Node{Style: Style{Width: Px(-155), AspectRatio: 0.5}}
	Layout(block, Loose(100, Unbounded), ctx)
	assertFiniteRect(t, "block with negative width and aspect ratio", block.Rect)
	if block.Rect.Width != 100 || block.Rect.Height != 200 {
		t.Errorf("expected 100x200 (auto width, ratio 0.5), got %vx%v", block.Rect.Width, block.Rect.Height)
	}
}

// TestBlockRound1SanitizeSize pins the NaN/Inf guard applied to every block
// layout result.
func TestBlockRound1SanitizeSize(t *testing.T) {
	cases := []struct{ in, want float64 }{
		{math.NaN(), 0},
		{math.Inf(1), Unbounded},
		{math.Inf(-1), 0},
		{Unbounded, Unbounded},
		{0, 0},
		{42.5, 42.5},
		{-3, -3},
	}
	for _, c := range cases {
		got := sanitizeSize(c.in)
		if got != c.want {
			t.Errorf("sanitizeSize(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

// Parent/child margin collapsing (CSS 2.1 §8.3.1,
// https://www.w3.org/TR/CSS21/box.html#collapsing-margins): the top margin of
// a box collapses with the top margin of its first in-flow child when no
// padding or border separates them, and the bottom margins collapse when the
// box additionally has an auto height and no min-height. The root and boxes
// that establish a new formatting context (flex, grid) never collapse through.

func round1Block(children ...*Node) *Node {
	return &Node{Style: Style{Display: DisplayBlock, Width: Px(200)}, Children: children}
}

func blockRound1Layout(root *Node) {
	Layout(root, Loose(500, 500), NewLayoutContext(500, 500, 16))
}

func round1Expect(t *testing.T, what string, got, want float64) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %.2f, want %.2f", what, got, want)
	}
}

// TestBlockRound1ParentChildMarginsCollapse: root > parent{} > child{H50,
// mt20, mb30}. The child's margins collapse through the parent and become the
// parent's margins: the child sits at the parent's origin, the parent is 50px
// tall and is offset by 20px inside the root, and the root is 100px tall.
func TestBlockRound1ParentChildMarginsCollapse(t *testing.T) {
	child := &Node{Style: Style{Display: DisplayBlock, Height: Px(50), Margin: Spacing{Top: Px(20), Bottom: Px(30)}}}
	parent := &Node{Style: Style{Display: DisplayBlock}, Children: []*Node{child}}
	root := round1Block(parent)
	blockRound1Layout(root)

	round1Expect(t, "child.Y (margin collapsed through parent)", child.Rect.Y, 0)
	round1Expect(t, "parent.Height", parent.Rect.Height, 50)
	round1Expect(t, "parent.Y (collapsed top margin)", parent.Rect.Y, 20)
	round1Expect(t, "root.Height", root.Rect.Height, 100)
}

// TestBlockRound1CollapsedThroughMarginMeetsSibling: a{ > {H10, mb30} } then
// b{H10, mt20}. The grandchild's bottom margin collapses through a and then
// with b's top margin: one 30px gap, not 30 + 20.
func TestBlockRound1CollapsedThroughMarginMeetsSibling(t *testing.T) {
	a := &Node{Style: Style{Display: DisplayBlock}, Children: []*Node{
		{Style: Style{Display: DisplayBlock, Height: Px(10), Margin: Spacing{Bottom: Px(30)}}},
	}}
	b := &Node{Style: Style{Display: DisplayBlock, Height: Px(10), Margin: Spacing{Top: Px(20)}}}
	root := round1Block(a, b)
	blockRound1Layout(root)

	round1Expect(t, "a.Height", a.Rect.Height, 10)
	round1Expect(t, "b.Y", b.Rect.Y, 40)
	round1Expect(t, "root.Height", root.Rect.Height, 50)
}

// TestBlockRound1PaddingBlocksCollapse: padding or border on the parent's edge
// separates the margins, so they stay inside the parent on that edge only.
func TestBlockRound1PaddingBlocksCollapse(t *testing.T) {
	child := &Node{Style: Style{Display: DisplayBlock, Height: Px(50), Margin: Spacing{Top: Px(20), Bottom: Px(30)}}}
	parent := &Node{Style: Style{Display: DisplayBlock, Padding: Spacing{Top: Px(1)}}, Children: []*Node{child}}
	root := round1Block(parent)
	blockRound1Layout(root)

	round1Expect(t, "child.Y (1px padding + 20px margin inside)", child.Rect.Y, 21)
	round1Expect(t, "parent.Y", parent.Rect.Y, 0)
	round1Expect(t, "parent.Height (1 + 20 + 50, bottom margin collapsed through)", parent.Rect.Height, 71)
	round1Expect(t, "root.Height (71 + 30)", root.Rect.Height, 101)

	// Border on the bottom edge keeps the bottom margin inside instead.
	parent.Style.Padding = Spacing{}
	parent.Style.Border = Spacing{Bottom: Px(2)}
	blockRound1Layout(root)
	round1Expect(t, "border-bottom: child.Y", child.Rect.Y, 0)
	round1Expect(t, "border-bottom: parent.Y (top collapsed through)", parent.Rect.Y, 20)
	round1Expect(t, "border-bottom: parent.Height (50 + 30 + 2)", parent.Rect.Height, 82)
	round1Expect(t, "border-bottom: root.Height", root.Rect.Height, 102)
}

// TestBlockRound1ExplicitHeightOrMinHeightBlocksBottomCollapse: the bottom
// margin collapses through only when the parent's height is auto and its
// min-height is 0; the top margin still collapses.
func TestBlockRound1ExplicitHeightOrMinHeightBlocksBottomCollapse(t *testing.T) {
	child := &Node{Style: Style{Display: DisplayBlock, Height: Px(50), Margin: Spacing{Top: Px(20), Bottom: Px(30)}}}
	parent := &Node{Style: Style{Display: DisplayBlock, Height: Px(100)}, Children: []*Node{child}}
	root := round1Block(parent)
	blockRound1Layout(root)
	round1Expect(t, "explicit height: child.Y", child.Rect.Y, 0)
	round1Expect(t, "explicit height: parent.Y", parent.Rect.Y, 20)
	round1Expect(t, "explicit height: parent.Height", parent.Rect.Height, 100)
	round1Expect(t, "explicit height: root.Height (bottom margin stays inside)", root.Rect.Height, 120)

	parent.Style.Height = Length{}
	parent.Style.MinHeight = Px(10)
	blockRound1Layout(root)
	round1Expect(t, "min-height: child.Y", child.Rect.Y, 0)
	round1Expect(t, "min-height: parent.Height (50 + 30 inside)", parent.Rect.Height, 80)
	round1Expect(t, "min-height: root.Height", root.Rect.Height, 100)
}

// TestBlockRound1FormattingContextRootsDoNotCollapseThrough: flex and grid
// containers establish independent formatting contexts (css-display-3 §2.1),
// so their children's margins never escape; the layout root does not collapse
// through either (TestBlockMarginCollapsingFirstChild pins that).
func TestBlockRound1FormattingContextRootsDoNotCollapseThrough(t *testing.T) {
	for _, display := range []Display{DisplayFlex, DisplayGrid} {
		item := &Node{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(50), Margin: Spacing{Top: Px(20)}}}
		container := &Node{Style: Style{Display: display, Width: Px(200)}, Children: []*Node{item}}
		root := round1Block(container)
		blockRound1Layout(root)
		if container.Rect.Y != 0 {
			t.Errorf("display %v: container.Y = %.2f, want 0 (item margin must stay inside)", display, container.Rect.Y)
		}
		if root.Rect.Height != container.Rect.Height {
			t.Errorf("display %v: root.Height = %.2f, container.Height = %.2f; no margin may escape the container", display, root.Rect.Height, container.Rect.Height)
		}
	}
}

// TestBlockRound1CollapseThroughSeveralLevels: margins collapse through any
// number of empty-edged ancestors and are combined as one set. With negative
// margins the set matters: collapse(25, -10, 30) = 30 - 10 = 20, whereas
// collapsing pairwise from the inside out would give collapse(25, 20) = 25.
func TestBlockRound1CollapseThroughSeveralLevels(t *testing.T) {
	leaf := &Node{Style: Style{Display: DisplayBlock, Height: Px(10), Margin: Spacing{Top: Px(30)}}}
	mid := &Node{Style: Style{Display: DisplayBlock, Margin: Spacing{Top: Px(-10)}}, Children: []*Node{leaf}}
	parent := &Node{Style: Style{Display: DisplayBlock, Margin: Spacing{Top: Px(25)}}, Children: []*Node{mid}}
	root := round1Block(parent)
	blockRound1Layout(root)

	round1Expect(t, "leaf.Y", leaf.Rect.Y, 0)
	round1Expect(t, "mid.Y", mid.Rect.Y, 0)
	round1Expect(t, "parent.Y (max positive 30 + most negative -10)", parent.Rect.Y, 20)
	round1Expect(t, "parent.Height", parent.Rect.Height, 10)
	round1Expect(t, "root.Height", root.Rect.Height, 30)
}

// TestBlockRound1EmptyContainerIsSelfCollapsing: a container whose only
// children are self-collapsing has zero height and all its margins adjoin;
// they collapse with the next sibling's top margin into a single gap.
func TestBlockRound1EmptyContainerIsSelfCollapsing(t *testing.T) {
	empty := &Node{Style: Style{Display: DisplayBlock}, Children: []*Node{
		{Style: Style{Display: DisplayBlock, Margin: Spacing{Top: Px(10), Bottom: Px(20)}}},
	}}
	b := &Node{Style: Style{Display: DisplayBlock, Height: Px(10), Margin: Spacing{Top: Px(5)}}}
	root := round1Block(empty, b)
	blockRound1Layout(root)

	round1Expect(t, "empty.Height", empty.Rect.Height, 0)
	round1Expect(t, "b.Y (max of 10, 20, 5)", b.Rect.Y, 20)
	round1Expect(t, "root.Height", root.Rect.Height, 30)
}

// TestBlockRound1OutOfFlowChildStaticPositionAtCollapsedStart: an absolutely
// positioned first child takes its static position at the parent's content
// origin when the following in-flow margins collapse through the parent.
func TestBlockRound1OutOfFlowChildStaticPositionAtCollapsedStart(t *testing.T) {
	abs := &Node{Style: Style{Display: DisplayBlock, Position: PositionAbsolute, Width: Px(10), Height: Px(10), Margin: Spacing{Top: Px(20)}}}
	child := &Node{Style: Style{Display: DisplayBlock, Height: Px(50), Margin: Spacing{Top: Px(20)}}}
	parent := &Node{Style: Style{Display: DisplayBlock}, Children: []*Node{abs, child}}
	root := round1Block(parent)
	blockRound1Layout(root)

	round1Expect(t, "abs.Y (static position)", abs.Rect.Y, 0)
	round1Expect(t, "child.Y", child.Rect.Y, 0)
	round1Expect(t, "parent.Y", parent.Rect.Y, 20)
	round1Expect(t, "parent.Height", parent.Rect.Height, 50)
}

// TestBlockRound1CollapseThroughVerticalWritingMode: in vertical-lr the block
// axis is horizontal, so left/right margins collapse through the parent.
// WritingMode is a per-node property in this engine (it is not inherited), so
// every container in the chain sets it.
// CSS Writing Modes Level 3 §7.1: https://www.w3.org/TR/css-writing-modes-3/#logical-to-physical
func TestBlockRound1CollapseThroughVerticalWritingMode(t *testing.T) {
	child := &Node{Style: Style{Display: DisplayBlock, WritingMode: WritingModeVerticalLR, Width: Px(50), Margin: Spacing{Left: Px(20), Right: Px(30)}}}
	parent := &Node{Style: Style{Display: DisplayBlock, WritingMode: WritingModeVerticalLR}, Children: []*Node{child}}
	root := &Node{Style: Style{Display: DisplayBlock, WritingMode: WritingModeVerticalLR, Height: Px(200)}, Children: []*Node{parent}}
	blockRound1Layout(root)

	round1Expect(t, "child.X", child.Rect.X, 0)
	round1Expect(t, "parent.X (collapsed start margin)", parent.Rect.X, 20)
	round1Expect(t, "parent.Width", parent.Rect.Width, 50)
	round1Expect(t, "root.Width", root.Rect.Width, 100)
}
