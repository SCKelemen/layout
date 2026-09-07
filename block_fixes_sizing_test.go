package layout

import (
	"math"
	"testing"
)

// Regression tests for block sizing fixes:
//   - border-box width smaller than padding+border (convertToContentSize clamp)
//   - max-height applied to auto-height blocks (CSS 2.1 §10.7)
//   - min wins over max when min > max (CSS 2.1 §10.4)
//   - intrinsic HeightSizing falls back to auto instead of storing -1
//   - zero-value Width/Height is auto, explicit Px(0) is zero
//   - nil LayoutContext and unbounded available width guardrails

func fixApproxEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

// TestBlockFixBorderBoxSmallerThanPaddingClampsToZero: a border-box width
// smaller than its padding must yield a 0 content box, not a negative one
// that is then misread as auto.
// CSS Sizing Level 3 §3.1: https://www.w3.org/TR/css-sizing-3/#box-sizing
func TestBlockFixBorderBoxSmallerThanPaddingClampsToZero(t *testing.T) {
	node := &Node{Style: Style{
		Display:   DisplayBlock,
		Width:     Px(10),
		Height:    Px(10),
		Padding:   Uniform(Px(20)),
		BoxSizing: BoxSizingBorderBox,
	}}
	size := Layout(node, Loose(500, 500), NewLayoutContext(500, 500, 16))
	if !fixApproxEqual(size.Width, 40) || !fixApproxEqual(size.Height, 40) {
		t.Fatalf("border-box 10px with 20px padding: expected 40x40 box, got %.2fx%.2f", size.Width, size.Height)
	}
}

// TestBlockFixMaxHeightAppliesToAutoHeight: max-height clamps a content-based
// height. CSS 2.1 §10.7: https://www.w3.org/TR/CSS21/visudet.html#min-max-heights
func TestBlockFixMaxHeightAppliesToAutoHeight(t *testing.T) {
	for _, height := range []Length{Px(-1), {}} {
		root := &Node{
			Style: Style{Display: DisplayBlock, Width: Px(100), Height: height, MaxHeight: Px(30)},
			Children: []*Node{
				{Style: Style{Display: DisplayBlock, Height: Px(100)}},
			},
		}
		size := Layout(root, Loose(500, 500), NewLayoutContext(500, 500, 16))
		if !fixApproxEqual(size.Height, 30) {
			t.Errorf("auto height %+v with max-height 30 and 100px child: expected 30, got %.2f", height, size.Height)
		}
	}
}

// TestBlockFixMinHeightBeatsMaxHeightOnAutoHeight: max is applied first, then
// min, so min wins (CSS 2.1 §10.7).
func TestBlockFixMinHeightBeatsMaxHeightOnAutoHeight(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayBlock, Width: Px(100), MinHeight: Px(80), MaxHeight: Px(30)},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Height: Px(100)}},
		},
	}
	size := Layout(root, Loose(500, 500), NewLayoutContext(500, 500, 16))
	if !fixApproxEqual(size.Height, 80) {
		t.Fatalf("min-height 80 > max-height 30: expected 80, got %.2f", size.Height)
	}
}

// TestBlockFixMinWidthWinsOverMaxWidth: CSS 2.1 §10.4 "If the computed value
// of min-width is greater than the value of max-width, max-width is set to the
// value of min-width." https://www.w3.org/TR/CSS21/visudet.html#min-max-widths
func TestBlockFixMinWidthWinsOverMaxWidth(t *testing.T) {
	node := &Node{Style: Style{
		Display: DisplayBlock, Width: Px(100), MinWidth: Px(300), MaxWidth: Px(200), Height: Px(10),
	}}
	size := Layout(node, Loose(500, 500), NewLayoutContext(500, 500, 16))
	if !fixApproxEqual(size.Width, 300) {
		t.Fatalf("width 100, min 300, max 200: expected 300, got %.2f", size.Width)
	}

	node = &Node{Style: Style{
		Display: DisplayBlock, Width: Px(10), Height: Px(100), MinHeight: Px(300), MaxHeight: Px(200),
	}}
	size = Layout(node, Loose(500, 500), NewLayoutContext(500, 500, 16))
	if !fixApproxEqual(size.Height, 300) {
		t.Fatalf("height 100, min 300, max 200: expected 300, got %.2f", size.Height)
	}
}

// TestBlockFixIntrinsicHeightSizingFallsBackToAuto: HeightSizing min-/max-/
// fit-content on a block container cannot be computed up front and must
// behave as auto, not produce "-1 + padding".
// CSS Sizing Level 3 §5.2: https://www.w3.org/TR/css-sizing-3/#intrinsic-contribution
func TestBlockFixIntrinsicHeightSizingFallsBackToAuto(t *testing.T) {
	for _, sizing := range []IntrinsicSize{IntrinsicSizeMinContent, IntrinsicSizeMaxContent, IntrinsicSizeFitContent} {
		root := &Node{
			Style: Style{Display: DisplayBlock, Width: Px(100), HeightSizing: sizing, Padding: Uniform(Px(10))},
			Children: []*Node{
				{Style: Style{Display: DisplayBlock, Height: Px(50)}},
			},
		}
		size := Layout(root, Loose(500, 500), NewLayoutContext(500, 500, 16))
		if !fixApproxEqual(size.Height, 70) {
			t.Errorf("HeightSizing %v with 50px child and 10px padding: expected 70, got %.2f", sizing, size.Height)
		}
	}

	// No children: only padding remains, never a negative content height.
	root := &Node{Style: Style{Display: DisplayBlock, Width: Px(100), HeightSizing: IntrinsicSizeMinContent, Padding: Uniform(Px(10))}}
	size := Layout(root, Loose(500, 500), NewLayoutContext(500, 500, 16))
	if !fixApproxEqual(size.Height, 20) {
		t.Errorf("HeightSizing min-content with no children: expected 20, got %.2f", size.Height)
	}
}

// TestBlockFixUnsetWidthIsAuto: a zero-value Width (never assigned) is auto,
// matching the flexbox and grid conventions and the CSS initial value
// (CSS 2.1 §10.3.3, https://www.w3.org/TR/CSS21/visudet.html#blockwidth).
func TestBlockFixUnsetWidthIsAuto(t *testing.T) {
	ctx := NewLayoutContext(500, 500, 16)

	// Style{} root: its width is auto, so its child can take its own width.
	root := &Node{
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(100), Height: Px(50)}},
		},
	}
	Layout(root, Loose(200, 200), ctx)
	if !fixApproxEqual(root.Children[0].Rect.Width, 100) {
		t.Errorf("child of Style{} root: expected width 100, got %.2f", root.Children[0].Rect.Width)
	}

	// A child with only Height set fills its parent's content width.
	root = &Node{
		Style: Style{Display: DisplayBlock, Width: Px(200)},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Height: Px(50)}},
		},
	}
	Layout(root, Loose(500, 500), ctx)
	if !fixApproxEqual(root.Children[0].Rect.Width, 200) {
		t.Errorf("child with unset width in 200px parent: expected 200, got %.2f", root.Children[0].Rect.Width)
	}
}

// TestBlockFixExplicitZeroWidthStaysZero: Px(0) is a real zero width.
func TestBlockFixExplicitZeroWidthStaysZero(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayBlock, Width: Px(200)},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Width: Px(0), Height: Px(50)}},
		},
	}
	Layout(root, Loose(500, 500), NewLayoutContext(500, 500, 16))
	if root.Children[0].Rect.Width != 0 {
		t.Errorf("child with Width Px(0): expected 0, got %.2f", root.Children[0].Rect.Width)
	}
}

// TestBlockFixNilContextDoesNotPanic: LayoutBlock with a nil LayoutContext
// falls back to the default 16px root font size.
func TestBlockFixNilContextDoesNotPanic(t *testing.T) {
	node := &Node{
		Style: Style{Display: DisplayBlock, Width: Em(10), Height: Px(10)},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Height: Px(5), Margin: Spacing{Top: Em(1)}}},
		},
	}
	size := LayoutBlock(node, Loose(500, 500), nil)
	if !fixApproxEqual(size.Width, 160) {
		t.Fatalf("nil ctx: Em(10) should resolve against 16px default, got width %.2f", size.Width)
	}
}

// TestBlockFixUnboundedAvailableWidthNeverBecomesSize: an unbounded available
// width is indefinite; an auto-width block with no children shrinks to 0
// (CSS 2.1 §10.3.5 shrink-to-fit with no content) instead of MaxFloat64.
func TestBlockFixUnboundedAvailableWidthNeverBecomesSize(t *testing.T) {
	node := &Node{Style: Style{Display: DisplayBlock, Height: Px(10)}}
	size := Layout(node, Unconstrained(), NewLayoutContext(500, 500, 16))
	if size.Width != 0 {
		t.Fatalf("auto width with unbounded available width and no children: expected 0, got %v", size.Width)
	}
	if node.Rect.Width >= Unbounded {
		t.Fatalf("Rect.Width must never be unbounded, got %v", node.Rect.Width)
	}
}
