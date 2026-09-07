package layout

import (
	"math"
	"strings"
	"testing"
)

// Round-1 correctness regression tests for text layout. Unless a test installs
// its own provider through the LayoutContext, glyphs are 10px wide
// (setupFakeMetrics).

func round1Layout(t *testing.T, node *Node, width float64, ctx *LayoutContext) Size {
	t.Helper()
	size := LayoutText(node, Loose(width, 1000), ctx)
	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}
	return size
}

func round1LineTexts(lines []TextLine) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		parts := make([]string, len(line.Boxes))
		for j, box := range line.Boxes {
			parts[j] = box.Text
		}
		out[i] = strings.Join(parts, "|")
	}
	return out
}

// TestRound1LineMetadataFields checks that TextLine.EndsWithForcedBreak and
// InlineBox.SpaceAfter are populated for soft wraps, forced breaks, and the
// last line (which behaves like a line before a forced break, css-text-3 Â§7.1).
// https://www.w3.org/TR/css-text-3/#forced-line-break
func TestRound1LineMetadataFields(t *testing.T) {
	setupFakeMetrics()
	ctx := NewLayoutContext(800, 600, 16)

	t.Run("soft wrap", func(t *testing.T) {
		node := Text("Hello big world", Style{Width: Px(100), TextStyle: &TextStyle{FontSize: 16}})
		round1Layout(t, node, 100, ctx)
		lines := node.TextLayout.Lines
		if got := round1LineTexts(lines); strings.Join(got, "/") != "Hello|big/world" {
			t.Fatalf("lines: got %q", got)
		}
		if lines[0].EndsWithForcedBreak {
			t.Error("a soft-wrapped line must not report a forced break")
		}
		if !lines[1].EndsWithForcedBreak {
			t.Error("the last line must report EndsWithForcedBreak")
		}
		if !lines[0].Boxes[0].SpaceAfter {
			t.Error("Hello is followed by an inter-word space: SpaceAfter should be true")
		}
		if lines[0].Boxes[1].SpaceAfter {
			t.Error("the trailing space of a wrapped line is removed: SpaceAfter should be false")
		}
		if lines[1].Boxes[0].SpaceAfter {
			t.Error("the last word has no space after it")
		}
	})

	t.Run("forced break in pre-line", func(t *testing.T) {
		node := Text("a b\nc", Style{Width: Px(200), TextStyle: &TextStyle{FontSize: 16, WhiteSpace: WhiteSpacePreLine}})
		round1Layout(t, node, 200, ctx)
		lines := node.TextLayout.Lines
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(lines))
		}
		if !lines[0].EndsWithForcedBreak || !lines[1].EndsWithForcedBreak {
			t.Errorf("both lines end at a forced break or the end of text: got %v, %v",
				lines[0].EndsWithForcedBreak, lines[1].EndsWithForcedBreak)
		}
		if !lines[0].Boxes[0].SpaceAfter || lines[0].Boxes[1].SpaceAfter {
			t.Errorf("SpaceAfter: got %v, %v; want true, false", lines[0].Boxes[0].SpaceAfter, lines[0].Boxes[1].SpaceAfter)
		}
	})

	t.Run("pre", func(t *testing.T) {
		node := Text("x\ny", Style{TextStyle: &TextStyle{FontSize: 16, WhiteSpace: WhiteSpacePre}})
		round1Layout(t, node, 200, ctx)
		for i, line := range node.TextLayout.Lines {
			if !line.EndsWithForcedBreak {
				t.Errorf("pre line %d should end with a forced break", i)
			}
		}
	})
}

// TestRound1ContextTextMetricsHonored checks that LayoutContext.TextMetrics
// drives both text layout and intrinsic sizing, and that a nil context still
// falls back to the package-level provider.
func TestRound1ContextTextMetricsHonored(t *testing.T) {
	setupFakeMetrics() // global: 10px per glyph
	custom := &fakeMetrics{charWidth: 12}
	ctx := NewLayoutContext(800, 600, 16).WithTextMetrics(custom)

	node := Text("Hello", Style{TextStyle: &TextStyle{FontSize: 16}})
	size := round1Layout(t, node, 500, ctx)
	if size.Width != 60 {
		t.Errorf("laid-out width with a 12px/char context provider: got %.2f, want 60", size.Width)
	}
	if got := node.TextLayout.Lines[0].Boxes[0].Width; got != 60 {
		t.Errorf("box width: got %.2f, want 60", got)
	}

	if got := CalculateIntrinsicWidth(node, Unconstrained(), IntrinsicSizeMaxContent, ctx); got != 60 {
		t.Errorf("max-content with context provider: got %.2f, want 60", got)
	}
	if got := CalculateIntrinsicWidth(node, Unconstrained(), IntrinsicSizeMinContent, ctx); got != 60 {
		t.Errorf("min-content with context provider: got %.2f, want 60", got)
	}

	// nil context: the global provider is used.
	node = Text("Hello", Style{TextStyle: &TextStyle{FontSize: 16}})
	size = round1Layout(t, node, 500, nil)
	if size.Width != 50 {
		t.Errorf("laid-out width with nil ctx: got %.2f, want 50", size.Width)
	}
	if got := CalculateIntrinsicWidth(node, Unconstrained(), IntrinsicSizeMaxContent, nil); got != 50 {
		t.Errorf("max-content with nil ctx: got %.2f, want 50", got)
	}

	// A context whose TextMetrics is nil also falls back to the global.
	size = round1Layout(t, node, 500, &LayoutContext{RootFontSize: 16})
	if size.Width != 50 {
		t.Errorf("laid-out width with ctx.TextMetrics == nil: got %.2f, want 50", size.Width)
	}
}

// TestRound1FontSizeZeroFallsBackToRoot checks that a non-nil TextStyle with
// FontSize == 0 uses the root font size (or 16 without a context) instead of
// laying out as 0x0, and that the caller's TextStyle is not mutated.
func TestRound1FontSizeZeroFallsBackToRoot(t *testing.T) {
	setupFakeMetrics()

	style := &TextStyle{}
	node := Text("Hi", Style{TextStyle: style})
	size := round1Layout(t, node, 500, NewLayoutContext(800, 600, 20))
	if size.Width != 20 || math.Abs(size.Height-24) > 1e-9 {
		t.Errorf("FontSize 0 with root 20: got %.2fx%.2f, want 20x24", size.Width, size.Height)
	}
	if math.Abs(node.TextLayout.LineHeight-24) > 1e-9 {
		t.Errorf("line height: got %.2f, want 24", node.TextLayout.LineHeight)
	}
	if style.FontSize != 0 {
		t.Errorf("the caller's TextStyle must not be mutated, FontSize is now %.2f", style.FontSize)
	}

	node = Text("Hi", Style{TextStyle: &TextStyle{}})
	size = round1Layout(t, node, 500, nil)
	if size.Width != 20 || math.Abs(size.Height-19.2) > 1e-9 {
		t.Errorf("FontSize 0 with nil ctx: got %.2fx%.2f, want 20x19.2", size.Width, size.Height)
	}

	// Intrinsic sizing measures with the same fallback.
	node = Text("Hi", Style{TextStyle: &TextStyle{}})
	if got := CalculateIntrinsicWidth(node, Unconstrained(), IntrinsicSizeMaxContent, nil); got != 20 {
		t.Errorf("max-content with FontSize 0: got %.2f, want 20", got)
	}
}

// TestRound1StyleDirection checks that Style.Direction is honored (with
// TextStyle.Direction as the legacy fallback) where direction affects
// alignment: text-align: start resolves to the right in RTL.
// https://www.w3.org/TR/css-writing-modes-3/#propdef-direction
func TestRound1StyleDirection(t *testing.T) {
	setupFakeMetrics()
	ctx := NewLayoutContext(800, 600, 16)

	// Style.Direction alone (TextStyle.Direction left at its LTR zero value).
	node := Text("Hello", Style{Width: Px(200), Direction: DirectionRTL, TextStyle: &TextStyle{FontSize: 16}})
	round1Layout(t, node, 200, ctx)
	if got := node.TextLayout.Lines[0].OffsetX; got != 150 {
		t.Errorf("Style.Direction RTL: OffsetX got %.2f, want 150", got)
	}

	// Legacy TextStyle.Direction still works.
	node = Text("Hello", Style{Width: Px(200), TextStyle: &TextStyle{FontSize: 16, Direction: DirectionRTL}})
	round1Layout(t, node, 200, ctx)
	if got := node.TextLayout.Lines[0].OffsetX; got != 150 {
		t.Errorf("TextStyle.Direction RTL: OffsetX got %.2f, want 150", got)
	}

	// LTR from both sources stays start-aligned at 0.
	node = Text("Hello", Style{Width: Px(200), TextStyle: &TextStyle{FontSize: 16}})
	round1Layout(t, node, 200, ctx)
	if got := node.TextLayout.Lines[0].OffsetX; got != 0 {
		t.Errorf("LTR: OffsetX got %.2f, want 0", got)
	}

	// In RTL the start edge is the right edge, so text-indent shortens the
	// line there: start-aligned text moves left by the indent.
	node = Text("Hello", Style{Width: Px(200), Direction: DirectionRTL, TextStyle: &TextStyle{FontSize: 16, TextIndent: 20}})
	round1Layout(t, node, 200, ctx)
	if got := node.TextLayout.Lines[0].OffsetX; got != 130 {
		t.Errorf("RTL with indent 20: OffsetX got %.2f, want 130", got)
	}
}

// TestRound1TextIndentEndAligned checks that text-indent is a margin on the
// start edge only (css-text-3 Â§7.2.1): right-aligned text stays flush with
// the right edge, center stays centered within the shortened line box.
// https://www.w3.org/TR/css-text-3/#text-indent-property
func TestRound1TextIndentEndAligned(t *testing.T) {
	setupFakeMetrics()
	ctx := NewLayoutContext(800, 600, 16)

	cases := []struct {
		name  string
		style TextStyle
		want  float64
	}{
		{"right", TextStyle{FontSize: 16, TextAlign: TextAlignRight, TextIndent: 20}, 150},
		{"center", TextStyle{FontSize: 16, TextAlign: TextAlignCenter, TextIndent: 20}, 85},
		{"left", TextStyle{FontSize: 16, TextAlign: TextAlignLeft, TextIndent: 20}, 20},
		{"justify last line right", TextStyle{FontSize: 16, TextAlign: TextAlignJustify, TextAlignLast: TextAlignLastRight, TextIndent: 20}, 150},
		{"justify last line center", TextStyle{FontSize: 16, TextAlign: TextAlignJustify, TextAlignLast: TextAlignLastCenter, TextIndent: 20}, 85},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := tc.style
			node := Text("Hello", Style{Width: Px(200), TextStyle: &st})
			round1Layout(t, node, 200, ctx)
			if got := node.TextLayout.Lines[0].OffsetX; math.Abs(got-tc.want) > 1e-9 {
				t.Errorf("OffsetX: got %.2f, want %.2f", got, tc.want)
			}
		})
	}
}

// TestRound1PreTabStops checks that a preserved tab (white-space: pre and
// pre-wrap) advances to the next tab stop instead of counting as a fixed
// number of glyphs, and that the tab survives in the box text.
// https://www.w3.org/TR/css-text-3/#tab-size-property
func TestRound1PreTabStops(t *testing.T) {
	setupFakeMetrics()
	// 6px glyphs: a tab stop every 8 * 6 = 48px.
	ctx := NewLayoutContext(800, 600, 16).WithTextMetrics(&fakeMetrics{charWidth: 6})

	cases := []struct {
		name  string
		text  string
		ws    WhiteSpace
		tab   float64
		width float64
	}{
		{"pre default tab-size", "a\tb", WhiteSpacePre, 0, 54},
		{"pre tab-size 8", "a\tb", WhiteSpacePre, 8, 54},
		{"pre tab-size 4", "a\tb", WhiteSpacePre, 4, 30},
		{"pre tab at a stop advances a full stop", "\tb", WhiteSpacePre, 8, 54},
		{"pre two tabs", "a\t\tb", WhiteSpacePre, 8, 102},
		{"pre run past first stop", "abcdefghi\tb", WhiteSpacePre, 8, 102},
		{"pre-wrap default tab-size", "a\tb", WhiteSpacePreWrap, 0, 54},
		{"pre-wrap tab between words", "ab \tcd", WhiteSpacePreWrap, 8, 60},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			node := Text(tc.text, Style{TextStyle: &TextStyle{FontSize: 16, WhiteSpace: tc.ws, TabSize: tc.tab}})
			size := round1Layout(t, node, 1000, ctx)
			if math.Abs(size.Width-tc.width) > 1e-9 {
				t.Errorf("width: got %.2f, want %.2f", size.Width, tc.width)
			}
			joined := strings.Join(round1LineTexts(node.TextLayout.Lines), "")
			if !strings.Contains(joined, "\t") {
				t.Errorf("tab must survive in box text, got %q", joined)
			}
			if got := CalculateIntrinsicWidth(node, Unconstrained(), IntrinsicSizeMaxContent, ctx); math.Abs(got-tc.width) > 1e-9 {
				t.Errorf("max-content: got %.2f, want %.2f", got, tc.width)
			}
		})
	}

	// Sanity: a fixed-count expansion would have given 18px for "a\tb".
	node := Text("a\tb", Style{TextStyle: &TextStyle{FontSize: 16, WhiteSpace: WhiteSpacePre}})
	if size := round1Layout(t, node, 1000, ctx); size.Width == 18 {
		t.Error("tab measured as a single glyph")
	}
}

// TestRound1SoftHyphen checks that U+00AD is invisible (not measured, not in
// box text) and rendered as a hyphen only when a line breaks at it.
// https://www.w3.org/TR/css-text-3/#hyphens-property
func TestRound1SoftHyphen(t *testing.T) {
	setupFakeMetrics()
	ctx := NewLayoutContext(800, 600, 16).WithTextMetrics(&fakeMetrics{charWidth: 6})
	const word = "abcÂ\u00addef"

	t.Run("no break: invisible", func(t *testing.T) {
		node := Text(word, Style{TextStyle: &TextStyle{FontSize: 16, Hyphens: HyphensManual}})
		size := round1Layout(t, node, 200, ctx)
		if size.Width != 36 {
			t.Errorf("width: got %.2f, want 36 (6 glyphs)", size.Width)
		}
		// The hyphenation opportunity splits the word into two adjacent
		// boxes (no space between them); neither carries the U+00AD.
		got := round1LineTexts(node.TextLayout.Lines)
		if strings.Join(got, "/") != "abc|def" {
			t.Errorf("boxes: got %q, want [abc|def]", got)
		}
		if strings.Contains(strings.Join(got, ""), "Â\u00ad") {
			t.Errorf("box text must not contain U+00AD: %q", got)
		}
		if box := node.TextLayout.Lines[0].Boxes[0]; box.SpaceAfter {
			t.Error("no inter-word space follows a soft-hyphen segment that did not break")
		}
		if got := CalculateIntrinsicWidth(node, Unconstrained(), IntrinsicSizeMaxContent, ctx); got != 36 {
			t.Errorf("max-content: got %.2f, want 36", got)
		}
	})

	t.Run("break at soft hyphen renders a hyphen", func(t *testing.T) {
		node := Text(word, Style{Width: Px(30), TextStyle: &TextStyle{FontSize: 16, Hyphens: HyphensManual}})
		round1Layout(t, node, 30, ctx)
		lines := node.TextLayout.Lines
		if got := round1LineTexts(lines); strings.Join(got, "/") != "abc-/def" {
			t.Fatalf("lines: got %q, want [abc- def]", got)
		}
		if lines[0].Width != 24 || lines[0].Boxes[0].Width != 24 {
			t.Errorf("hyphenated line: line %.2f box %.2f, want 24 (abc + hyphen)", lines[0].Width, lines[0].Boxes[0].Width)
		}
		if lines[1].Width != 18 {
			t.Errorf("second line: got %.2f, want 18", lines[1].Width)
		}
	})

	t.Run("hyphen width participates in the fit test", func(t *testing.T) {
		// "xy ab-" is 30px with the hyphen; at 28px the break must happen
		// before "ab" so the hyphenated line does not overflow.
		node := Text("xy abÂ\u00adcd", Style{Width: Px(28), TextStyle: &TextStyle{FontSize: 16, Hyphens: HyphensManual}})
		round1Layout(t, node, 28, ctx)
		for i, line := range node.TextLayout.Lines {
			if line.Width > 28 {
				t.Errorf("line %d overflows: %.2f > 28 (%q)", i, line.Width, round1LineTexts(node.TextLayout.Lines)[i])
			}
		}
		if got := round1LineTexts(node.TextLayout.Lines); strings.Join(got, "/") != "xy/ab|cd" {
			t.Errorf("lines: got %q, want [xy ab|cd]", got)
		}
	})

	t.Run("hyphens none: no break, still invisible", func(t *testing.T) {
		node := Text(word, Style{Width: Px(30), TextStyle: &TextStyle{FontSize: 16, Hyphens: HyphensNone}})
		round1Layout(t, node, 30, ctx)
		got := round1LineTexts(node.TextLayout.Lines)
		if strings.Join(got, "/") != "abcdef" {
			t.Errorf("lines: got %q, want [abcdef]", got)
		}
		if node.TextLayout.Lines[0].Width != 36 {
			t.Errorf("width: got %.2f, want 36", node.TextLayout.Lines[0].Width)
		}
	})
}

// TestRound1BreakWordIndentFillsFirstLine checks that when text-indent leaves
// no room on the first line, overflow-wrap: break-word starts the word on the
// next line instead of overflowing the first line by the indent.
// https://www.w3.org/TR/css-text-3/#text-indent-property
func TestRound1BreakWordIndentFillsFirstLine(t *testing.T) {
	setupFakeMetrics()
	ctx := NewLayoutContext(800, 600, 16)

	node := Text("abcdefgh", Style{Width: Px(50), TextStyle: &TextStyle{FontSize: 16, TextIndent: 50, OverflowWrap: OverflowWrapBreakWord}})
	round1Layout(t, node, 50, ctx)
	lines := node.TextLayout.Lines
	if got := round1LineTexts(lines); strings.Join(got, "/") != "/abcde/fgh" {
		t.Fatalf("lines: got %q, want [<empty> abcde fgh]", got)
	}
	for i, line := range lines {
		end := line.OffsetX + line.Width
		if end > 50+1e-9 {
			t.Errorf("line %d ends at %.2f, overflowing the 50px box", i, end)
		}
	}
	if lines[1].OffsetX != 0 {
		t.Errorf("second line must not be indented, OffsetX=%.2f", lines[1].OffsetX)
	}

	// With room on the first line the indent is honored: the first piece is
	// limited to the indented width, continuation pieces to the full width.
	node = Text("abcdefgh", Style{Width: Px(50), TextStyle: &TextStyle{FontSize: 16, TextIndent: 20, OverflowWrap: OverflowWrapBreakWord}})
	round1Layout(t, node, 50, ctx)
	if got := round1LineTexts(node.TextLayout.Lines); strings.Join(got, "/") != "abc/defgh" {
		t.Errorf("lines with indent 20: got %q, want [abc defgh]", got)
	}
}

// TestRound1TextOverflowUsesSpaceAfter checks that ellipsis truncation still
// accounts for inter-word spaces now that the information lives on the boxes.
func TestRound1TextOverflowUsesSpaceAfter(t *testing.T) {
	setupFakeMetrics()
	ctx := NewLayoutContext(800, 600, 16)

	node := Text("Hello big world", Style{Width: Px(100), TextStyle: &TextStyle{FontSize: 16, WhiteSpace: WhiteSpaceNowrap, TextOverflow: TextOverflowEllipsis}})
	round1Layout(t, node, 100, ctx)
	line := node.TextLayout.Lines[0]
	if line.Width > 100 {
		t.Errorf("truncated line overflows: %.2f", line.Width)
	}
	last := line.Boxes[len(line.Boxes)-1]
	if last.Text != "..." {
		t.Fatalf("last box should be the ellipsis, got %q", last.Text)
	}
	if len(line.Boxes) >= 2 && line.Boxes[len(line.Boxes)-2].SpaceAfter {
		t.Error("no space should separate the retained text from the ellipsis")
	}
	if line.Boxes[0].Text != "Hello" || !line.Boxes[0].SpaceAfter {
		t.Errorf("first box: got %q SpaceAfter=%v, want Hello with a space after", line.Boxes[0].Text, line.Boxes[0].SpaceAfter)
	}
}
