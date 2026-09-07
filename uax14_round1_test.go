package layout

import "testing"

// Round-1 regression tests for the UAX #14 plumbing that text layout relies
// on for hyphenation and preserved tabs.

// TestUAX14SoftHyphenSegments verifies that every soft hyphen (U+00AD, class
// BA) yields a break opportunity after it when hyphenation is allowed, with
// byte offsets that keep the two-byte U+00AD attached to the preceding
// segment, and that hyphens: none suppresses all of them.
// https://www.unicode.org/reports/tr14/#BA
// https://www.w3.org/TR/css-text-3/#valdef-hyphens-none
func TestUAX14SoftHyphenSegments(t *testing.T) {
	if got := getBreakClass('­'); got != ClassBA {
		t.Errorf("U+00AD class: got %v, want ClassBA", got)
	}
	assertSegments(t, "a­b­c", HyphensManual, "a­", "b­", "c")
	assertSegments(t, "a­b­c", HyphensAuto, "a­", "b­", "c")
	assertSegments(t, "a­b­c", HyphensNone, "a­b­c")
	// A soft hyphen next to a space does not create an extra opportunity
	// beyond the one the space already provides.
	assertSegments(t, "ab­ cd", HyphensManual, "ab­ ", "cd")
	// Multi-byte text before the soft hyphen keeps byte offsets aligned.
	assertSegments(t, "éé­üü", HyphensManual, "éé­", "üü")
}

// TestUAX14TabIsSpace verifies that a preserved tab is treated as SP for
// segmentation, so the pre-wrap breaker sees "a\t" and "b" as separate
// segments and can advance the tab to a tab stop.
func TestUAX14TabIsSpace(t *testing.T) {
	if got := getBreakClass('\t'); got != ClassSP {
		t.Errorf("TAB class: got %v, want ClassSP", got)
	}
	assertSegments(t, "a\tb", HyphensManual, "a\t", "b")
}
