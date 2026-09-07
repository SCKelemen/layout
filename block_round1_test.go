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
