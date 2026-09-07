package layout

import (
	"math"
	"testing"
)

// Regression tests for the text layout review fixes. All tests use
// fakeMetrics (10px per rune).

// TestTextUnboundedWidthIsAuto verifies that a Width resolving to the
// Unbounded sentinel is treated as auto rather than as an explicit inline
// size. Before the fix, text-align right/center aligned lines against
// math.MaxFloat64 and produced astronomically large offsets.
func TestTextUnboundedWidthIsAuto(t *testing.T) {
	setupFakeMetrics()

	cases := []struct {
		name      string
		align     TextAlign
		alignLast TextAlignLast
	}{
		{"right", TextAlignRight, TextAlignLastAuto},
		{"center", TextAlignCenter, TextAlignLastAuto},
		{"justify", TextAlignJustify, TextAlignLastJustify},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			node := &Node{
				Style: Style{
					Display: DisplayInlineText,
					Width:   PxUnbounded,
					TextStyle: &TextStyle{
						FontSize:      16,
						TextAlign:     tc.align,
						TextAlignLast: tc.alignLast,
						WhiteSpace:    WhiteSpaceNormal,
						Direction:     DirectionLTR,
					},
				},
				Text: "hello world",
			}
			size := LayoutText(node, Loose(300, 300), NewLayoutContext(800, 600, 16))

			if math.IsInf(size.Width, 0) || size.Width > 300 {
				t.Fatalf("size.Width = %v, want finite and <= 300", size.Width)
			}
			if node.TextLayout == nil || len(node.TextLayout.Lines) == 0 {
				t.Fatal("expected at least one laid out line")
			}
			for i, line := range node.TextLayout.Lines {
				if line.OffsetX < 0 || line.OffsetX > 300 {
					t.Errorf("line %d OffsetX = %v, want within [0, 300]", i, line.OffsetX)
				}
				if line.Width < 0 || line.Width > 300 {
					t.Errorf("line %d Width = %v, want within [0, 300]", i, line.Width)
				}
			}
		})
	}
}

// TestTextUnboundedHeightIsAuto verifies that an unbounded Height is treated
// as auto and the block size falls back to the line count.
func TestTextUnboundedHeightIsAuto(t *testing.T) {
	setupFakeMetrics()

	node := &Node{
		Style: Style{
			Display: DisplayInlineText,
			Height:  PxUnbounded,
			TextStyle: &TextStyle{
				FontSize:   16,
				WhiteSpace: WhiteSpaceNormal,
				Direction:  DirectionLTR,
			},
		},
		Text: "hello world",
	}
	size := LayoutText(node, Loose(300, 300), NewLayoutContext(800, 600, 16))

	// One line at line-height normal (1.2 * 16).
	if want := 16 * 1.2; math.Abs(size.Height-want) > 0.001 {
		t.Errorf("size.Height = %v, want %v", size.Height, want)
	}
}

// TestTextHelperLeavesSizeAuto verifies that Text() does not seed Px(0) for
// Width/Height. The library treats the zero-value Length (Unit == "") as
// auto and Px(0) as an explicit zero, so a Px(0) seed collapsed Text() nodes
// used as grid or block children to 0x0.
func TestTextHelperLeavesSizeAuto(t *testing.T) {
	node := Text("hello world")
	if node.Style.Width.Unit != "" || node.Style.Width.Value != 0 {
		t.Errorf("Text() Width = %+v, want zero-value (auto)", node.Style.Width)
	}
	if node.Style.Height.Unit != "" || node.Style.Height.Value != 0 {
		t.Errorf("Text() Height = %+v, want zero-value (auto)", node.Style.Height)
	}
	if node.Style.Display != DisplayInlineText {
		t.Errorf("Text() Display = %v, want DisplayInlineText", node.Style.Display)
	}
	if node.Style.TextStyle == nil {
		t.Error("Text() should set a default TextStyle")
	}
}

// TestTextHelperInBlockParent verifies that a Text() child of a fixed-width
// block gets a positive width.
func TestTextHelperInBlockParent(t *testing.T) {
	setupFakeMetrics()

	child := Text("hello world")
	root := &Node{
		Style:    Style{Display: DisplayBlock, Width: Px(200)},
		Children: []*Node{child},
	}
	Layout(root, Loose(400, 400), NewLayoutContext(800, 600, 16))

	if child.Rect.Width <= 0 {
		t.Errorf("Text() child width = %v, want > 0", child.Rect.Width)
	}
	if child.Rect.Height <= 0 {
		t.Errorf("Text() child height = %v, want > 0", child.Rect.Height)
	}
}

// TestTextHelperInGridCell verifies that a Text() grid item lays out in its
// cell like a raw DisplayInlineText node does.
func TestTextHelperInGridCell(t *testing.T) {
	setupFakeMetrics()

	layoutItem := func(item *Node) Rect {
		root := &Node{
			Style: Style{
				Display:             DisplayGrid,
				Width:               Px(200),
				Height:              Px(100),
				GridTemplateColumns: []GridTrack{FixedTrack(Px(200))},
				GridTemplateRows:    []GridTrack{FixedTrack(Px(100))},
			},
			Children: []*Node{item},
		}
		Layout(root, Loose(400, 400), NewLayoutContext(800, 600, 16))
		return item.Rect
	}

	helper := layoutItem(Text("hello world"))
	raw := layoutItem(&Node{
		Style: Style{Display: DisplayInlineText},
		Text:  "hello world",
	})

	if helper.Width <= 0 || helper.Height <= 0 {
		t.Fatalf("Text() grid item rect = %+v, want positive size", helper)
	}
	if helper != raw {
		t.Errorf("Text() grid item rect = %+v, raw inline-text item rect = %+v; want equal", helper, raw)
	}
}

// TestGridIntrinsicTrackUsesNodeFontSize verifies that em-sized grid tracks
// resolve against the grid container's font size during intrinsic sizing
// instead of a hard-coded 16px.
func TestGridIntrinsicTrackUsesNodeFontSize(t *testing.T) {
	node := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Em(2))},
			TextStyle:           &TextStyle{FontSize: 20},
		},
	}
	ctx := NewLayoutContext(800, 600, 16)

	if got := calculateGridMinContentWidth(node, Loose(800, 600), ctx); math.Abs(got-40) > 0.001 {
		t.Errorf("min-content width = %v, want 40 (2em at 20px)", got)
	}
	if got := calculateGridMaxContentWidth(node, Loose(800, 600), ctx); math.Abs(got-40) > 0.001 {
		t.Errorf("max-content width = %v, want 40 (2em at 20px)", got)
	}
}
