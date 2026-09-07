package layout

import (
	"strings"
	"testing"
)

// breakSegments splits text at the UAX #14 break opportunities found for the
// given hyphens setting. The returned segments concatenate back to text.
func breakSegments(t *testing.T, text string, hyphens Hyphens) []string {
	t.Helper()
	points := findLineBreakOpportunitiesWithHyphens(text, hyphens)
	if len(points) < 2 || points[0] != 0 || points[len(points)-1] != len(text) {
		t.Fatalf("break points for %q must start at 0 and end at len: %v", text, points)
	}
	segments := make([]string, 0, len(points)-1)
	for i := 0; i < len(points)-1; i++ {
		if points[i+1] <= points[i] {
			t.Fatalf("break points for %q are not strictly increasing: %v", text, points)
		}
		segments = append(segments, text[points[i]:points[i+1]])
	}
	if strings.Join(segments, "") != text {
		t.Fatalf("segments %q do not reassemble %q", segments, text)
	}
	return segments
}

func assertSegments(t *testing.T, text string, hyphens Hyphens, want ...string) {
	t.Helper()
	got := breakSegments(t, text, hyphens)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("%q: expected segments %q, got %q", text, want, got)
	}
}

// TestUAX14ZeroWidthSpaceBreaks verifies LB8: ZW SP* ÷.
// https://www.unicode.org/reports/tr14/#LB8
func TestUAX14ZeroWidthSpaceBreaks(t *testing.T) {
	assertSegments(t, "a\u200Bb", HyphensManual, "a\u200B", "b")
	// LB7: no break before ZW itself
	assertSegments(t, "ab\u200Bcd", HyphensManual, "ab\u200B", "cd")
}

// TestUAX14WordJoinerProhibitsBreak verifies LB11: × WJ, WJ ×, even after a space.
// https://www.unicode.org/reports/tr14/#LB11
func TestUAX14WordJoinerProhibitsBreak(t *testing.T) {
	assertSegments(t, "a \u2060b", HyphensManual, "a \u2060b")
	assertSegments(t, "a\u2060 b", HyphensManual, "a\u2060 ", "b")
}

// TestUAX14NoBreakBeforeClosingPunctuationAfterSpace verifies LB13/LB15d:
// no break before CL, CP, EX, IS, SY even after spaces.
// https://www.unicode.org/reports/tr14/#LB13
func TestUAX14NoBreakBeforeClosingPunctuationAfterSpace(t *testing.T) {
	assertSegments(t, "word !", HyphensManual, "word !")
	assertSegments(t, "word ?", HyphensManual, "word ?")
	assertSegments(t, "word )", HyphensManual, "word )")
	assertSegments(t, "word ,", HyphensManual, "word ,")
	assertSegments(t, "word .", HyphensManual, "word .")
	assertSegments(t, "word /", HyphensManual, "word /")
	// LB15c: a decimal mark after a space may be preceded by a break
	assertSegments(t, "a .5", HyphensManual, "a ", ".5")
}

// TestUAX14IdeographsAndPunctuation verifies that breaks around ideographs
// respect the neighbor's class: no break before closing punctuation, quotes,
// or combining marks, and no break after opening punctuation or glue.
// https://www.unicode.org/reports/tr14/#LB9
// https://www.unicode.org/reports/tr14/#LB12
// https://www.unicode.org/reports/tr14/#LB13
// https://www.unicode.org/reports/tr14/#LB14
// https://www.unicode.org/reports/tr14/#LB19
func TestUAX14IdeographsAndPunctuation(t *testing.T) {
	assertSegments(t, "漢字", HyphensManual, "漢", "字")
	assertSegments(t, "漢字。", HyphensManual, "漢", "字。")
	assertSegments(t, "漢字！", HyphensManual, "漢", "字！")
	assertSegments(t, "漢字!", HyphensManual, "漢", "字!")
	assertSegments(t, "漢字,", HyphensManual, "漢", "字,")
	assertSegments(t, "漢字、", HyphensManual, "漢", "字、")
	assertSegments(t, "漢字)", HyphensManual, "漢", "字)")
	assertSegments(t, "漢字」", HyphensManual, "漢", "字」")
	assertSegments(t, "漢字\"", HyphensManual, "漢", "字\"")
	assertSegments(t, "(漢字", HyphensManual, "(漢", "字")
	assertSegments(t, "「漢字", HyphensManual, "「漢", "字")
	assertSegments(t, "\"漢字", HyphensManual, "\"漢", "字")
	// LB9: combining mark attaches to the ideograph
	assertSegments(t, "漢\u0301字", HyphensManual, "漢\u0301", "字")
	// LB12/LB12a: NBSP glues both neighbors
	assertSegments(t, "漢\u00A0字", HyphensManual, "漢\u00A0字")
	// LB21: no break before a nonstarter (small kana, prolonged sound mark)
	assertSegments(t, "きゃく", HyphensManual, "きゃ", "く")
	assertSegments(t, "ラーメン", HyphensManual, "ラー", "メ", "ン")
}

// TestUAX14KanaAreIdeographic verifies that hiragana and katakana are class
// ID and therefore break between each other.
// https://www.unicode.org/reports/tr14/#ID
func TestUAX14KanaAreIdeographic(t *testing.T) {
	assertSegments(t, "こんにちは", HyphensManual, "こ", "ん", "に", "ち", "は")
	assertSegments(t, "カタカナ", HyphensManual, "カ", "タ", "カ", "ナ")
	if got := getBreakClass('あ'); got != ClassID {
		t.Errorf("あ: expected ClassID, got %v", got)
	}
	if got := getBreakClass('ー'); got != ClassNS {
		t.Errorf("ー: expected ClassNS, got %v", got)
	}
	if got := getBreakClass('漢'); got != ClassID {
		t.Errorf("漢: expected ClassID, got %v", got)
	}
}

// TestUAX14HardHyphensAlwaysBreak verifies LB21: a break after an ordinary
// hyphen is always an opportunity, regardless of the CSS hyphens property,
// which only controls hyphenation.
// https://www.unicode.org/reports/tr14/#LB21
// https://www.w3.org/TR/css-text-3/#hyphens-property
func TestUAX14HardHyphensAlwaysBreak(t *testing.T) {
	for _, h := range []Hyphens{HyphensNone, HyphensManual, HyphensAuto} {
		assertSegments(t, "well-known", h, "well-", "known")
		// LB25: HY × NU keeps numeric ranges together
		assertSegments(t, "1-2", h, "1-2")
		assertSegments(t, "-2", h, "-2")
	}
}

// TestUAX14SoftHyphenGatedByHyphens verifies that U+00AD is a break
// opportunity only when hyphenation is enabled.
// https://www.w3.org/TR/css-text-3/#valdef-hyphens-none
func TestUAX14SoftHyphenGatedByHyphens(t *testing.T) {
	assertSegments(t, "su\u00ADper", HyphensNone, "su\u00ADper")
	assertSegments(t, "su\u00ADper", HyphensManual, "su\u00AD", "per")
	assertSegments(t, "su\u00ADper", HyphensAuto, "su\u00AD", "per")
}

// TestUAX14Spaces verifies LB7 and LB18: breaks happen after a run of
// spaces, never inside or before it.
// https://www.unicode.org/reports/tr14/#LB18
func TestUAX14Spaces(t *testing.T) {
	assertSegments(t, "a b", HyphensManual, "a ", "b")
	assertSegments(t, "a  b", HyphensManual, "a  ", "b")
	assertSegments(t, "ab cd ef", HyphensManual, "ab ", "cd ", "ef")
	// Leading spaces belong to the first segment
	assertSegments(t, " a", HyphensManual, " ", "a")
}

// TestUAX14MandatoryBreaks verifies LB4-LB6 with CR LF treated as one break.
// https://www.unicode.org/reports/tr14/#LB5
func TestUAX14MandatoryBreaks(t *testing.T) {
	assertSegments(t, "a\nb", HyphensManual, "a\n", "b")
	assertSegments(t, "a\r\nb", HyphensManual, "a\r\n", "b")
	assertSegments(t, "a b", HyphensManual, "a ", "b")
}

// TestUAX14AlphabeticAndNumeric verifies that words and numbers stay intact
// (LB23, LB25, LB28, LB29, LB30) and that quotes and brackets attach to them.
// https://www.unicode.org/reports/tr14/#LB28
func TestUAX14AlphabeticAndNumeric(t *testing.T) {
	assertSegments(t, "hello", HyphensManual, "hello")
	assertSegments(t, "abc123", HyphensManual, "abc123")
	assertSegments(t, "3.14", HyphensManual, "3.14")
	assertSegments(t, "1,000", HyphensManual, "1,000")
	assertSegments(t, "$100", HyphensManual, "$100")
	assertSegments(t, "100%", HyphensManual, "100%")
	assertSegments(t, "(word)", HyphensManual, "(word)")
	assertSegments(t, "\"quoted\" text", HyphensManual, "\"quoted\" ", "text")
	assertSegments(t, "e.g. x", HyphensManual, "e.g. ", "x")
}

// TestUAX14Hangul verifies LB26/LB27 style breaking between Hangul syllables.
// https://www.unicode.org/reports/tr14/#LB26
func TestUAX14Hangul(t *testing.T) {
	assertSegments(t, "한국어", HyphensManual, "한", "국", "어")
}

// TestUAX14Progress verifies the guardrails: every input yields strictly
// increasing byte offsets that never split a multi-byte rune, and empty text
// is handled.
func TestUAX14Progress(t *testing.T) {
	if got := findLineBreakOpportunitiesWithHyphens("", HyphensManual); len(got) != 1 || got[0] != 0 {
		t.Errorf("empty text: got %v", got)
	}
	inputs := []string{"\u0301", "\u0301\u0301a", " ", "   ", "\n\n", "\u200B\u200B", "a\u00A0", "漢\u0301\u0301字 ", "\r\n\r\n"}
	for _, in := range inputs {
		for _, p := range findLineBreakOpportunitiesWithHyphens(in, HyphensAuto) {
			if p < 0 || p > len(in) {
				t.Errorf("%q: break offset %d out of range", in, p)
			} else if p < len(in) && !isRuneStart(in, p) {
				t.Errorf("%q: break offset %d is not a rune boundary", in, p)
			}
		}
		breakSegments(t, in, HyphensAuto) // asserts monotonic and reassembly
	}
}

func isRuneStart(s string, i int) bool {
	return i == 0 || i == len(s) || (s[i]&0xC0) != 0x80
}
