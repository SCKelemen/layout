package layout

import (
	"math"
	"strings"
	"testing"
)

// Regression tests for text layout fixes. All tests use fakeMetrics
// (10px per rune, letter-spacing added between runes).

// layoutFixText lays out text with fakeMetrics and returns the node.
func layoutFixText(t *testing.T, text string, style Style, constraints Constraints) *Node {
	t.Helper()
	setupFakeMetrics()
	node := Text(text, style)
	LayoutText(node, constraints, NewLayoutContext(800, 600, 16))
	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}
	return node
}

// lineStrings renders each line as a string: boxes are joined with a space
// when the line tracks inter-word spaces (SpaceCount > 0), otherwise they are
// concatenated (pre-wrap keeps spaces as their own boxes).
func lineStrings(lines []TextLine) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		texts := make([]string, 0, len(line.Boxes))
		for _, box := range line.Boxes {
			texts = append(texts, box.Text)
		}
		sep := ""
		if line.SpaceCount > 0 {
			sep = " "
		}
		out = append(out, strings.Join(texts, sep))
	}
	return out
}

func assertLines(t *testing.T, got []TextLine, want ...string) {
	t.Helper()
	gotStrings := lineStrings(got)
	if len(gotStrings) != len(want) {
		t.Fatalf("expected %d lines %q, got %d lines %q", len(want), want, len(gotStrings), gotStrings)
	}
	for i := range want {
		if gotStrings[i] != want[i] {
			t.Errorf("line %d: expected %q, got %q", i, want[i], gotStrings[i])
		}
	}
}

// TestExplicitWidthControlsLineBreaking verifies that an explicit width is
// the inline size used for line breaking, not just for the resulting box.
// "Hello world" is 110px; with width 100 it must wrap.
// https://www.w3.org/TR/css-text-3/#line-breaking
func TestExplicitWidthControlsLineBreaking(t *testing.T) {
	node := layoutFixText(t, "Hello world", Style{Width: Px(100)}, Loose(800, 600))

	assertLines(t, node.TextLayout.Lines, "Hello", "world")
	if node.Rect.Width != 100 {
		t.Errorf("Rect.Width: expected 100, got %.2f", node.Rect.Width)
	}
}

// TestExplicitWidthBorderBoxControlsLineBreaking verifies that the explicit
// width is converted to a content-box size before line breaking.
// https://www.w3.org/TR/css-sizing-3/#box-sizing
func TestExplicitWidthBorderBoxControlsLineBreaking(t *testing.T) {
	node := layoutFixText(t, "Hello world", Style{
		Width:     Px(120),
		Padding:   Uniform(Px(10)),
		BoxSizing: BoxSizingBorderBox,
	}, Loose(800, 600))

	// Content width is 120 - 20 = 100, so "Hello world" (110px) wraps.
	assertLines(t, node.TextLayout.Lines, "Hello", "world")
	if node.Rect.Width != 120 {
		t.Errorf("Rect.Width: expected 120, got %.2f", node.Rect.Width)
	}
}

// TestMaxWidthClampsLineBreaking verifies that max-width limits the
// shrink-to-fit inline size used for line breaking.
// https://www.w3.org/TR/css-sizing-3/#min-max-widths
func TestMaxWidthClampsLineBreaking(t *testing.T) {
	node := layoutFixText(t, "Hello world", Style{MaxWidth: Px(100)}, Loose(800, 600))

	assertLines(t, node.TextLayout.Lines, "Hello", "world")
	if node.Rect.Width > 100 {
		t.Errorf("Rect.Width: expected <= 100, got %.2f", node.Rect.Width)
	}
}

// TestTrailingSpaceDoesNotCountTowardFit verifies that the space after a
// word does not decide whether the word fits: white space at the end of a
// line hangs (CSS Text 3 §4.1.3).
// https://www.w3.org/TR/css-text-3/#white-space-phase-2
func TestTrailingSpaceDoesNotCountTowardFit(t *testing.T) {
	// "aaaa bbbb" = 90px fits in 95px; only the trailing space would overflow.
	node := layoutFixText(t, "aaaa bbbb cc", Style{Width: Px(95)}, Loose(800, 600))

	assertLines(t, node.TextLayout.Lines, "aaaa bbbb", "cc")
	if node.TextLayout.Lines[0].Width != 90 {
		t.Errorf("line 0 width: expected 90 (trailing space removed), got %.2f", node.TextLayout.Lines[0].Width)
	}
	if node.TextLayout.Lines[0].SpaceCount != 1 {
		t.Errorf("line 0 SpaceCount: expected 1, got %d", node.TextLayout.Lines[0].SpaceCount)
	}
}

// TestTextIndentAppliesToWholeFirstLine verifies that text-indent reduces
// the space available for every word on the first line, not just the first
// word.
// https://www.w3.org/TR/css-text-3/#text-indent-property
func TestTextIndentAppliesToWholeFirstLine(t *testing.T) {
	node := layoutFixText(t, "aaaa bbbb c", Style{
		Width:     Px(100),
		TextStyle: &TextStyle{FontSize: 16, TextIndent: 30},
	}, Loose(800, 600))

	// 30 + "aaaa bbbb" (90) = 120 > 100, so "bbbb" moves to the second line.
	assertLines(t, node.TextLayout.Lines, "aaaa", "bbbb c")
	line0 := node.TextLayout.Lines[0]
	if line0.OffsetX != 30 {
		t.Errorf("line 0 OffsetX: expected 30, got %.2f", line0.OffsetX)
	}
	if end := line0.OffsetX + line0.Width; end > 100 {
		t.Errorf("line 0 must end within the 100px content box, ends at %.2f", end)
	}
}

// TestPreWrapTrailingSpacesHang verifies that preserved spaces at the end of
// a pre-wrap line hang instead of being pushed onto a line of their own, and
// are excluded from the line width.
// https://www.w3.org/TR/css-text-3/#white-space-phase-2
func TestPreWrapTrailingSpacesHang(t *testing.T) {
	node := layoutFixText(t, "aaaa bbbb", Style{
		Width:     Px(40),
		TextStyle: &TextStyle{FontSize: 16, WhiteSpace: WhiteSpacePreWrap},
	}, Loose(800, 600))

	assertLines(t, node.TextLayout.Lines, "aaaa ", "bbbb")
	for i, line := range node.TextLayout.Lines {
		if line.Width != 40 {
			t.Errorf("line %d width: expected 40, got %.2f", i, line.Width)
		}
		if strings.TrimSpace(lineStrings([]TextLine{line})[0]) == "" {
			t.Errorf("line %d is space-only", i)
		}
	}
}

// TestPreWrapTrailingSpacesBeforeForcedBreakHangConditionally verifies that
// spaces before a forced break count toward the width while they fit.
func TestPreWrapTrailingSpacesBeforeForcedBreakHangConditionally(t *testing.T) {
	node := layoutFixText(t, "aa  \nbb", Style{
		Width:     Px(200),
		TextStyle: &TextStyle{FontSize: 16, WhiteSpace: WhiteSpacePreWrap},
	}, Loose(800, 600))

	assertLines(t, node.TextLayout.Lines, "aa  ", "bb")
	if node.TextLayout.Lines[0].Width != 40 {
		t.Errorf("line 0 width: expected 40 (spaces fit, so they count), got %.2f", node.TextLayout.Lines[0].Width)
	}
}

// TestEllipsisAccountsForInterWordSpaces verifies that text-overflow
// truncation includes the inter-word space advance when fitting boxes and
// keeps SpaceCount for the retained gaps.
// https://www.w3.org/TR/css-overflow-3/#text-overflow
func TestEllipsisAccountsForInterWordSpaces(t *testing.T) {
	node := layoutFixText(t, "aaaa bbbb cccc dddd", Style{
		Width: Px(100),
		TextStyle: &TextStyle{
			FontSize:     16,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowEllipsis,
		},
	}, Loose(800, 600))

	if len(node.TextLayout.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(node.TextLayout.Lines))
	}
	line := node.TextLayout.Lines[0]
	var texts []string
	for _, box := range line.Boxes {
		texts = append(texts, box.Text)
	}
	// 100 - "..." (30) = 70 available: "aaaa" (40) + space (10) + "bb" (20)
	want := []string{"aaaa", "bb", "..."}
	if strings.Join(texts, "|") != strings.Join(want, "|") {
		t.Errorf("boxes: expected %q, got %q", want, texts)
	}
	if line.SpaceCount != 1 {
		t.Errorf("SpaceCount: expected 1 retained gap, got %d", line.SpaceCount)
	}
	if line.Width != 100 {
		t.Errorf("line width: expected 100 (visible run incl. space), got %.2f", line.Width)
	}
}

// TestNBSPNotTrimmedAtEdges verifies that U+00A0 is not stripped by white
// space trimming in normal and pre-line modes.
// https://www.w3.org/TR/css-text-3/#white-space-processing
func TestNBSPNotTrimmedAtEdges(t *testing.T) {
	text := "\u00A0Hello\u00A0"

	node := layoutFixText(t, text, Style{Width: Px(200)}, Loose(800, 600))
	assertLines(t, node.TextLayout.Lines, text)
	if node.TextLayout.Lines[0].Width != 70 {
		t.Errorf("normal: width expected 70, got %.2f", node.TextLayout.Lines[0].Width)
	}

	node = layoutFixText(t, "\u00A0Hi\nx", Style{
		Width:     Px(200),
		TextStyle: &TextStyle{FontSize: 16, WhiteSpace: WhiteSpacePreLine},
	}, Loose(800, 600))
	assertLines(t, node.TextLayout.Lines, "\u00A0Hi", "x")
}

// TestNBSPSurvivesCollapsibleTrim verifies the helper directly.
func TestNBSPSurvivesCollapsibleTrim(t *testing.T) {
	if got := trimCollapsibleSpace(" \t\u00A0a\u00A0 \n"); got != "\u00A0a\u00A0" {
		t.Errorf("trimCollapsibleSpace: got %q", got)
	}
	words := splitIntoWords("a\u00A0b c")
	if len(words) != 2 || words[0] != "a\u00A0b" {
		t.Errorf("splitIntoWords: got %q", words)
	}
}

// TestHangingPunctuationAllowEndIsEndOnly verifies that allow-end does not
// hang opening punctuation at the start of the line.
// https://www.w3.org/TR/css-text-3/#valdef-hanging-punctuation-allow-end
func TestHangingPunctuationAllowEndIsEndOnly(t *testing.T) {
	node := layoutFixText(t, "\"Hello", Style{
		Width:     Px(200),
		TextStyle: &TextStyle{FontSize: 16, HangingPunctuation: HangingPunctuationAllowEnd},
	}, Loose(800, 600))
	line := node.TextLayout.Lines[0]
	if line.OffsetX != 0 {
		t.Errorf("allow-end must not hang opening quote: OffsetX expected 0, got %.2f", line.OffsetX)
	}
	if line.Width != 60 {
		t.Errorf("allow-end must not change width of a line without closing punctuation: expected 60, got %.2f", line.Width)
	}

	// allow-end still hangs closing punctuation at the end.
	node = layoutFixText(t, "Hello.", Style{
		Width:     Px(200),
		TextStyle: &TextStyle{FontSize: 16, HangingPunctuation: HangingPunctuationAllowEnd},
	}, Loose(800, 600))
	if w := node.TextLayout.Lines[0].Width; w != 50 {
		t.Errorf("allow-end should hang the period: width expected 50, got %.2f", w)
	}

	// first still hangs opening punctuation.
	node = layoutFixText(t, "\"Hello", Style{
		Width:     Px(200),
		TextStyle: &TextStyle{FontSize: 16, HangingPunctuation: HangingPunctuationFirst},
	}, Loose(800, 600))
	if x := node.TextLayout.Lines[0].OffsetX; x != -10 {
		t.Errorf("first should hang opening quote: OffsetX expected -10, got %.2f", x)
	}
}

// TestBreakWordHonorsLetterSpacing verifies that overflow-wrap: break-word
// measures candidate pieces as runs so letter-spacing is included.
// https://www.w3.org/TR/css-text-3/#overflow-wrap-property
func TestBreakWordHonorsLetterSpacing(t *testing.T) {
	node := layoutFixText(t, "abcdefghij", Style{
		Width: Px(40),
		TextStyle: &TextStyle{
			FontSize:      16,
			LetterSpacing: 5,
			OverflowWrap:  OverflowWrapBreakWord,
		},
	}, Loose(800, 600))

	// 3 runes = 30 + 2*5 = 40 fits; 4 runes = 40 + 3*5 = 55 does not.
	assertLines(t, node.TextLayout.Lines, "abc", "def", "ghi", "j")
	for i, line := range node.TextLayout.Lines {
		if line.Width > 40 {
			t.Errorf("line %d width %.2f exceeds 40", i, line.Width)
		}
	}
}

// TestBreakWordToFitDirect checks the helper with letter-spacing.
func TestBreakWordToFitDirect(t *testing.T) {
	setupFakeMetrics()
	style := TextStyle{FontSize: 16, LetterSpacing: 5}
	pieces := breakWordToFit("abcdefghij", 40, style)
	if strings.Join(pieces, "|") != "abc|def|ghi|j" {
		t.Errorf("pieces: got %q", pieces)
	}
	for _, p := range pieces {
		if w, _, _ := getTextMetrics().Measure(p, style); w > 40 {
			t.Errorf("piece %q measures %.2f > 40", p, w)
		}
	}
	// A single rune wider than the limit still produces a piece (progress guaranteed).
	if got := breakWordToFit("ab", 5, style); strings.Join(got, "|") != "a|b" {
		t.Errorf("narrow limit: got %q", got)
	}
}

// TestJustifyLineBeforeForcedBreakUsesTextAlignLast verifies that the line
// preceding a forced break is aligned like the last line (CSS Text 3 §7.1)
// instead of being justified.
// https://www.w3.org/TR/css-text-3/#text-align-property
func TestJustifyLineBeforeForcedBreakUsesTextAlignLast(t *testing.T) {
	style := Style{
		Width: Px(95),
		TextStyle: &TextStyle{
			FontSize:   16,
			WhiteSpace: WhiteSpacePreLine,
			TextAlign:  TextAlignJustify,
		},
	}
	node := layoutFixText(t, "aa bb cc dd ee ff\nzz", style, Loose(800, 600))

	assertLines(t, node.TextLayout.Lines, "aa bb cc", "dd ee ff", "zz")
	lines := node.TextLayout.Lines
	if math.Abs(lines[0].SpaceAdjustment-7.5) > 0.01 {
		t.Errorf("line 0 (soft wrapped) should be justified: SpaceAdjustment expected 7.5, got %.2f", lines[0].SpaceAdjustment)
	}
	if lines[1].SpaceAdjustment != 0 {
		t.Errorf("line 1 (before forced break) must not be justified: SpaceAdjustment expected 0, got %.2f", lines[1].SpaceAdjustment)
	}
	if lines[2].SpaceAdjustment != 0 {
		t.Errorf("last line must not be justified: got %.2f", lines[2].SpaceAdjustment)
	}

	// With text-align-last: justify, the line before the forced break is justified too.
	style.TextStyle.TextAlignLast = TextAlignLastJustify
	node = layoutFixText(t, "aa bb cc dd ee ff\nzz", style, Loose(800, 600))
	if math.Abs(node.TextLayout.Lines[1].SpaceAdjustment-7.5) > 0.01 {
		t.Errorf("text-align-last: justify should justify line before forced break: got %.2f", node.TextLayout.Lines[1].SpaceAdjustment)
	}
}

// TestPreLineFittingSegmentTracksSpaces verifies that a pre-line segment that
// fits on one line still records its inter-word spaces so it can be justified.
func TestPreLineFittingSegmentTracksSpaces(t *testing.T) {
	node := layoutFixText(t, "aa bb cc\nzz", Style{
		Width: Px(95),
		TextStyle: &TextStyle{
			FontSize:      16,
			WhiteSpace:    WhiteSpacePreLine,
			TextAlign:     TextAlignJustify,
			TextAlignLast: TextAlignLastJustify,
		},
	}, Loose(800, 600))

	assertLines(t, node.TextLayout.Lines, "aa bb cc", "zz")
	line0 := node.TextLayout.Lines[0]
	if line0.SpaceCount != 2 {
		t.Fatalf("line 0 SpaceCount: expected 2, got %d", line0.SpaceCount)
	}
	if math.Abs(line0.SpaceAdjustment-7.5) > 0.01 {
		t.Errorf("line 0 SpaceAdjustment: expected 7.5, got %.2f", line0.SpaceAdjustment)
	}
}

// TestVerticalRLSizeAndLineOffsets verifies that vertical-rl text reports
// its size with the inline and block axes swapped, and stacks lines from the
// right edge of the content box.
// https://www.w3.org/TR/css-writing-modes-3/#block-flow
func TestVerticalRLSizeAndLineOffsets(t *testing.T) {
	// Inline size (height) 50 fits "aaaa" (40) per line: 6 lines.
	text := "aaaa bbbb cccc dddd eeee ffff"
	node := layoutFixText(t, text, Style{
		WritingMode: WritingModeVerticalRL,
		TextStyle:   &TextStyle{FontSize: 16, LineHeight: 20},
	}, Loose(800, 50))

	lines := node.TextLayout.Lines
	if len(lines) != 6 {
		t.Fatalf("expected 6 lines, got %d: %q", len(lines), lineStrings(lines))
	}
	if node.Rect.Width != 120 {
		t.Errorf("Rect.Width (block size): expected 120, got %.2f", node.Rect.Width)
	}
	if node.Rect.Height > 50 {
		t.Errorf("Rect.Height (inline size): expected <= 50, got %.2f", node.Rect.Height)
	}
	for i, line := range lines {
		wantX := 100 - float64(i)*20
		if line.OffsetX != wantX {
			t.Errorf("line %d OffsetX: expected %.2f, got %.2f", i, wantX, line.OffsetX)
		}
		if line.OffsetY != 0 {
			t.Errorf("line %d OffsetY: expected 0, got %.2f", i, line.OffsetY)
		}
	}

	// vertical-lr: same size, lines stack from the left edge.
	node = layoutFixText(t, text, Style{
		WritingMode: WritingModeVerticalLR,
		TextStyle:   &TextStyle{FontSize: 16, LineHeight: 20},
	}, Loose(800, 50))
	if node.Rect.Width != 120 || node.Rect.Height > 50 {
		t.Errorf("vertical-lr size: expected 120 x <=50, got %.2f x %.2f", node.Rect.Width, node.Rect.Height)
	}
	for i, line := range node.TextLayout.Lines {
		if wantX := float64(i) * 20; line.OffsetX != wantX {
			t.Errorf("vertical-lr line %d OffsetX: expected %.2f, got %.2f", i, wantX, line.OffsetX)
		}
	}
}

// TestVerticalExplicitHeightControlsLineBreaking verifies that in vertical
// modes the explicit height is the inline size used for line breaking.
func TestVerticalExplicitHeightControlsLineBreaking(t *testing.T) {
	node := layoutFixText(t, "Hello world", Style{
		WritingMode: WritingModeVerticalLR,
		Height:      Px(100),
		TextStyle:   &TextStyle{FontSize: 16, LineHeight: 20},
	}, Loose(800, 600))

	assertLines(t, node.TextLayout.Lines, "Hello", "world")
	if node.Rect.Height != 100 {
		t.Errorf("Rect.Height: expected 100, got %.2f", node.Rect.Height)
	}
	if node.Rect.Width != 40 {
		t.Errorf("Rect.Width: expected 40 (2 lines x 20), got %.2f", node.Rect.Width)
	}
}

// TestTabSizeZeroMeansDefault verifies that an unset tab-size (0) expands
// tabs to the default 8 spaces, matching -1.
// https://www.w3.org/TR/css-text-3/#tab-size-property
func TestTabSizeZeroMeansDefault(t *testing.T) {
	want := "a" + strings.Repeat(" ", 8) + "b"
	if got := expandTabs("a\tb", 0); got != want {
		t.Errorf("tab-size 0: got %q, want %q", got, want)
	}
	if got := expandTabs("a\tb", -1); got != want {
		t.Errorf("tab-size -1: got %q, want %q", got, want)
	}
	if got := expandTabs("a\tb", 4); got != "a    b" {
		t.Errorf("tab-size 4: got %q", got)
	}
}

// TestCJKClosingPunctuationStaysWithPrecedingCharacter verifies through the
// full layout that a line never starts with closing punctuation.
// https://www.unicode.org/reports/tr14/#LB13
func TestCJKClosingPunctuationStaysWithPrecedingCharacter(t *testing.T) {
	node := layoutFixText(t, "漢字。漢字。", Style{Width: Px(30)}, Loose(800, 600))
	for i, line := range node.TextLayout.Lines {
		if len(line.Boxes) == 0 {
			continue
		}
		if strings.HasPrefix(line.Boxes[0].Text, "。") {
			t.Errorf("line %d starts with closing punctuation: %q", i, lineStrings([]TextLine{line})[0])
		}
	}
}
