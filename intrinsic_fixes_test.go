package layout

import (
	"math"
	"testing"
)

// Regression tests for intrinsic sizing contributions
// (CSS Sizing Level 3 §5.1, https://www.w3.org/TR/css-sizing-3/#intrinsic-contribution).

func fixMaxContentBlock(children ...*Node) *Node {
	return &Node{
		Style:    Style{Display: DisplayBlock, WidthSizing: IntrinsicSizeMaxContent},
		Children: children,
	}
}

// TestIntrinsicFixChildLengthsAreResolved: child widths and margins in
// relative units are resolved to pixels; padding and border are added for
// content-box children and already included for border-box children.
func TestIntrinsicFixChildLengthsAreResolved(t *testing.T) {
	ctx := NewLayoutContext(1000, 1000, 16)
	tests := []struct {
		name     string
		child    *Node
		expected float64
	}{
		{"em width", &Node{Style: Style{Width: Em(10), Height: Px(10)}}, 160},
		{"content-box padding and border", &Node{Style: Style{Width: Px(100), Height: Px(10), Padding: Uniform(Px(10)), Border: Uniform(Px(2))}}, 124},
		{"border-box padding", &Node{Style: Style{Width: Px(100), Height: Px(10), Padding: Uniform(Px(10)), BoxSizing: BoxSizingBorderBox}}, 100},
		{"em margin", &Node{Style: Style{Width: Px(100), Height: Px(10), Margin: Spacing{Left: Em(1)}}}, 116},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := fixMaxContentBlock(tt.child)
			size := Layout(root, Loose(1000, 1000), ctx)
			if math.Abs(size.Width-tt.expected) > 0.01 {
				t.Errorf("max-content width: expected %.2f, got %.2f", tt.expected, size.Width)
			}
		})
	}
}

// TestIntrinsicFixTextChildrenContribute: text children contribute the width
// of the whole run to max-content and of the longest word to min-content
// (CSS Sizing Level 3 §4.1, https://www.w3.org/TR/css-sizing-3/#intrinsic-sizes).
func TestIntrinsicFixTextChildrenContribute(t *testing.T) {
	ctx := NewLayoutContext(1000, 1000, 16)
	text := Text("hello world")
	style := *text.Style.TextStyle
	fullWidth, _, _ := getTextMetrics().Measure("hello world", style)
	wordWidth, _, _ := getTextMetrics().Measure("hello", style)
	if fullWidth <= 0 || wordWidth <= 0 || wordWidth >= fullWidth {
		t.Fatalf("unexpected metrics: full=%v word=%v", fullWidth, wordWidth)
	}

	root := fixMaxContentBlock(text)
	size := Layout(root, Loose(1000, 1000), ctx)
	if math.Abs(size.Width-fullWidth) > 0.01 {
		t.Errorf("max-content with text child: expected %.2f, got %.2f", fullWidth, size.Width)
	}

	root = &Node{
		Style:    Style{Display: DisplayBlock, WidthSizing: IntrinsicSizeMinContent},
		Children: []*Node{Text("hello world")},
	}
	size = Layout(root, Loose(1000, 1000), ctx)
	if math.Abs(size.Width-wordWidth) > 0.01 {
		t.Errorf("min-content with text child: expected longest word %.2f, got %.2f", wordWidth, size.Width)
	}

	// nowrap: no soft wrap opportunities, min-content equals max-content.
	nowrap := Text("hello world")
	nowrap.Style.TextStyle.WhiteSpace = WhiteSpaceNowrap
	root = &Node{
		Style:    Style{Display: DisplayBlock, WidthSizing: IntrinsicSizeMinContent},
		Children: []*Node{nowrap},
	}
	size = Layout(root, Loose(1000, 1000), ctx)
	if math.Abs(size.Width-fullWidth) > 0.01 {
		t.Errorf("min-content with nowrap text child: expected %.2f, got %.2f", fullWidth, size.Width)
	}
}

// TestIntrinsicFixFlexGapSkipsDisplayNone: gaps are only placed between
// visible items (CSS Box Alignment Level 3 §8).
func TestIntrinsicFixFlexGapSkipsDisplayNone(t *testing.T) {
	ctx := NewLayoutContext(1000, 1000, 16)
	root := &Node{
		Style: Style{Display: DisplayFlex, FlexDirection: FlexDirectionRow, FlexGap: Px(10)},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(10)}},
			{Style: Style{Display: DisplayNone, Width: Px(50), Height: Px(10)}},
			{Style: Style{Width: Px(50), Height: Px(10)}},
		},
	}
	for _, sizing := range []IntrinsicSize{IntrinsicSizeMinContent, IntrinsicSizeMaxContent} {
		got := CalculateIntrinsicWidth(root, Unconstrained(), sizing, ctx)
		if math.Abs(got-110) > 0.01 {
			t.Errorf("flex row [50, none, 50] gap 10 (%v): expected 110, got %.2f", sizing, got)
		}
	}
}
