package layout

import (
	"unicode"
	"unicode/utf8"
)

// UAX #14: Unicode Line Breaking Algorithm
// Based on: https://www.unicode.org/reports/tr14/
//
// This is a focused implementation of the pair rules LB1-LB31 for the
// character classes this package classifies. It is not a complete
// implementation (no tailoring, no emoji modifier or regional indicator
// handling), but every rule it applies follows the published algorithm.

// BreakClass represents a Unicode line breaking class.
type BreakClass int

const (
	// Mandatory breaks
	ClassBK BreakClass = iota // Mandatory Break
	ClassCR                   // Carriage Return
	ClassLF                   // Line Feed
	ClassNL                   // Next Line
	ClassSP                   // Space

	// Prohibited breaks
	ClassWJ BreakClass = iota + 5 // Word Joiner
	ClassZW                       // Zero Width Space

	// Break opportunities
	ClassBA BreakClass = iota + 10 // Break After
	ClassBB                        // Break Before
	ClassB2                        // Break Opportunity Before and After
	ClassHY                        // Hyphen
	ClassCB                        // Contingent Break Opportunity

	// Characters
	ClassAL BreakClass = iota + 20 // Alphabetic
	ClassHL                        // Hebrew Letter
	ClassID                        // Ideographic
	ClassIN                        // Inseparable
	ClassNU                        // Numeric
	ClassPR                        // Prefix Numeric
	ClassPO                        // Postfix Numeric
	ClassIS                        // Infix Numeric Separator
	ClassSY                        // Symbols Allowing Break After
	ClassAI                        // Ambiguous (Alphabetic or Ideographic)
	ClassCJ                        // Conditional Japanese Starter
	ClassSA                        // Complex Context Dependent (South East Asian)

	// Punctuation
	ClassOP BreakClass = iota + 40 // Open Punctuation
	ClassCL                        // Close Punctuation
	ClassCP                        // Close Parenthesis
	ClassQU                        // Quotation
	ClassGL                        // Non-breaking ("Glue")
	ClassNS                        // Nonstarter
	ClassEX                        // Exclamation/Interrogation

	// Combining marks
	ClassCM BreakClass = iota + 60 // Combining Mark

	// Hangul
	ClassJL BreakClass = iota + 70 // Hangul L Jamo
	ClassJV                        // Hangul V Jamo
	ClassJT                        // Hangul T Jamo
	ClassH2                        // Hangul LV Syllable
	ClassH3                        // Hangul LVT Syllable

	// Regional indicators
	ClassRI BreakClass = iota + 80 // Regional Indicator

	// Surrogates
	ClassSG BreakClass = iota + 90 // Surrogate

	// Unknown
	ClassXX BreakClass = iota + 100 // Unknown
)

// BreakAction represents the action to take at a line break opportunity.
type BreakAction int

const (
	// BreakProhibited means no line break is allowed
	BreakProhibited BreakAction = iota
	// BreakDirect means a line break is allowed
	BreakDirect
	// BreakIndirect means a line break is allowed only if preceded by space
	BreakIndirect
	// BreakMandatory means a line break is required
	BreakMandatory
)

// smallKana lists the kana that UAX #14 classifies as CJ (Conditional
// Japanese Starter), which LB1 resolves to NS for our purposes: small kana,
// the prolonged sound mark, and iteration marks.
// Reference: https://www.unicode.org/reports/tr14/#CJ
var smallKana = map[rune]bool{
	// Hiragana small letters
	'ぁ': true, 'ぃ': true, 'ぅ': true, 'ぇ': true, 'ぉ': true,
	'っ': true, 'ゃ': true, 'ゅ': true, 'ょ': true, 'ゎ': true,
	'ゕ': true, 'ゖ': true, 'ゝ': true, 'ゞ': true,
	// Katakana small letters
	'ァ': true, 'ィ': true, 'ゥ': true, 'ェ': true, 'ォ': true,
	'ッ': true, 'ャ': true, 'ュ': true, 'ョ': true, 'ヮ': true,
	'ヵ': true, 'ヶ': true, 'ヽ': true, 'ヾ': true,
	'ー': true, // U+30FC KATAKANA-HIRAGANA PROLONGED SOUND MARK
	// Halfwidth katakana small letters and prolonged sound mark
	'ｧ': true, 'ｨ': true, 'ｩ': true, 'ｪ': true, 'ｫ': true,
	'ｬ': true, 'ｭ': true, 'ｮ': true, 'ｯ': true, 'ｰ': true,
}

// getBreakClass returns the line breaking class for a rune.
// This is a simplified implementation focusing on common cases. LB1
// (resolution of AI, CJ, SA, SG, XX) is folded into this function so
// callers only ever see resolved classes.
// Reference: https://www.unicode.org/reports/tr14/#Table1
func getBreakClass(r rune) BreakClass {
	// Mandatory breaks (LB4, LB5)
	switch r {
	case '\n':
		return ClassLF
	case '\r':
		return ClassCR
	case '\u0085': // NEL (Next Line)
		return ClassNL
	case '\v', '\f', '\u2028', '\u2029': // VT, FF, LS, PS
		return ClassBK
	}

	// Space characters. TAB is BA in UAX #14, but the layout engine either
	// expands tabs before line breaking or preserves them without wrapping,
	// so treating it as SP here is harmless and keeps word splitting simple.
	if r == ' ' || r == '\t' {
		return ClassSP
	}

	switch r {
	case '\u00A0', '\u202F', '\u2007': // NBSP, NNBSP, FIGURE SPACE
		return ClassGL
	case '\u200B': // ZERO WIDTH SPACE
		return ClassZW
	case '\u2060', '\uFEFF': // WORD JOINER, ZWNBSP
		return ClassWJ
	case '\u00AD': // SOFT HYPHEN
		return ClassBA
	case '\u3000': // IDEOGRAPHIC SPACE
		return ClassBA
	case '‐', '‒', '–': // HYPHEN, FIGURE DASH, EN DASH
		return ClassBA
	case '—': // EM DASH
		return ClassB2
	}

	// Punctuation
	switch r {
	case '(', '[', '{', '⟨', '｟', '（', '［', '｛', '「', '『', '【', '〈', '《', '〔', '〖', '〘', '〚', '¿', '¡':
		return ClassOP
	case ')', ']', '}', '⟩', '｠', '）', '］', '｝':
		return ClassCP
	case '」', '』', '】', '〉', '》', '〕', '〗', '〙', '〛', '、', '。', '，', '．', '､', '｡':
		return ClassCL
	case '"', '\'', '«', '»', '„', '‚', '‹', '›', '“', '”', '‘', '’':
		return ClassQU
	case '!', '?', '！', '？':
		return ClassEX
	case '-':
		return ClassHY
	case '/':
		return ClassSY
	case ',', '.', ':', ';':
		return ClassIS
	case '…', '‥', '․': // HORIZONTAL ELLIPSIS, TWO/ONE DOT LEADER
		return ClassIN
	case '$', '£', '¥', '€', '+', '\\', '−': // currency and signs
		return ClassPR
	case '%', '°', '‰', '℃', '℉': // percent, degree signs
		return ClassPO
	case '々', '〻': // IDEOGRAPHIC ITERATION MARK, VERTICAL IDEOGRAPHIC ITERATION MARK
		return ClassNS
	}

	// Hangul (LB26/LB27 operate on these)
	switch {
	case r >= 0x1100 && r <= 0x115F, r >= 0xA960 && r <= 0xA97C:
		return ClassJL
	case r >= 0x1160 && r <= 0x11A7, r >= 0xD7B0 && r <= 0xD7C6:
		return ClassJV
	case r >= 0x11A8 && r <= 0x11FF, r >= 0xD7CB && r <= 0xD7FB:
		return ClassJT
	case r >= 0xAC00 && r <= 0xD7A3:
		if (r-0xAC00)%28 == 0 {
			return ClassH2
		}
		return ClassH3
	}

	// Numeric
	if unicode.Is(unicode.N, r) {
		return ClassNU
	}

	// Combining marks
	if unicode.Is(unicode.M, r) {
		return ClassCM
	}

	// Kana: small kana and the prolonged sound mark are CJ (resolved to NS
	// by LB1); all other kana are ID. The prolonged sound marks (U+30FC,
	// U+FF70) have Script=Common, so check the table before the script.
	// https://www.unicode.org/reports/tr14/#ID
	if smallKana[r] {
		return ClassNS
	}
	if unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) {
		return ClassID
	}

	// Ideographic (CJK)
	if unicode.Is(unicode.Ideographic, r) || unicode.Is(unicode.Han, r) {
		return ClassID
	}

	// Hebrew letters
	if unicode.Is(unicode.Hebrew, r) {
		return ClassHL
	}

	// Alphabetic (default for letters and everything else, including
	// symbols; UAX #14 maps most Sm/Sc/Sk/So to AL).
	return ClassAL
}

// isMandatoryBreakClass reports whether cls forces a break after it (LB4, LB5).
func isMandatoryBreakClass(cls BreakClass) bool {
	return cls == ClassBK || cls == ClassCR || cls == ClassLF || cls == ClassNL
}

// getBreakAction returns the break action between two adjacent character
// classes with no intervening spaces. It is a thin wrapper over the pair
// rules used by findLineBreakOpportunities; BreakIndirect is never returned
// because the space state is resolved by the caller.
func getBreakAction(before, after BreakClass) BreakAction {
	if isMandatoryBreakClass(before) {
		if before == ClassCR && after == ClassLF {
			return BreakProhibited // LB5: CR × LF
		}
		return BreakMandatory
	}
	if isMandatoryBreakClass(after) {
		return BreakProhibited // LB6
	}
	if pairBreakAllowed(before, after, false, ClassXX) {
		return BreakDirect
	}
	return BreakProhibited
}

// isOneOf reports whether cls is in the given set.
func isOneOf(cls BreakClass, set ...BreakClass) bool {
	for _, c := range set {
		if cls == c {
			return true
		}
	}
	return false
}

// pairBreakAllowed applies UAX #14 rules LB7-LB31 to a pair of classes.
//
//   - before is the class of the last non-space character (after LB9/LB10
//     combining-mark resolution), or ClassSP if the text so far is spaces.
//   - after is the class of the current character.
//   - hadSpace reports whether one or more SP separated the two.
//   - next is the class of the character following "after" (ClassXX if
//     none); it is only consulted by LB15c.
//
// The rules are evaluated in specification order; the first matching rule
// decides. Rules whose classes this package never produces (LB8a, LB21a,
// LB28a, LB30a, LB30b) are omitted.
// Reference: https://www.unicode.org/reports/tr14/#Algorithm
func pairBreakAllowed(before, after BreakClass, hadSpace bool, next BreakClass) bool {
	// LB7: × SP, × ZW
	if after == ClassSP || after == ClassZW {
		return false
	}
	// LB8: ZW SP* ÷
	if before == ClassZW {
		return true
	}
	// LB11: × WJ, WJ ×
	if after == ClassWJ || (before == ClassWJ && !hadSpace) {
		return false
	}
	// LB12: GL ×
	if before == ClassGL && !hadSpace {
		return false
	}
	// LB12a: [^SP BA HY] × GL
	if after == ClassGL && !hadSpace && !isOneOf(before, ClassSP, ClassBA, ClassHY) {
		return false
	}
	// LB13: × CL, × CP, × EX, × SY (even after spaces)
	if isOneOf(after, ClassCL, ClassCP, ClassEX, ClassSY) {
		return false
	}
	// LB14: OP SP* ×
	if before == ClassOP {
		return false
	}
	// LB15: QU SP* × OP
	if before == ClassQU && after == ClassOP {
		return false
	}
	// LB15c: SP ÷ IS NU (break before a decimal mark that follows a space)
	if after == ClassIS && hadSpace && next == ClassNU {
		return true
	}
	// LB15d: × IS
	if after == ClassIS {
		return false
	}
	// LB16: (CL | CP) SP* × NS
	if isOneOf(before, ClassCL, ClassCP) && after == ClassNS {
		return false
	}
	// LB17: B2 SP* × B2
	if before == ClassB2 && after == ClassB2 {
		return false
	}
	// LB18: SP ÷
	if hadSpace {
		return true
	}
	// LB19: × QU, QU ×
	if after == ClassQU || before == ClassQU {
		return false
	}
	// LB20: ÷ CB, CB ÷
	if after == ClassCB || before == ClassCB {
		return true
	}
	// LB21: × BA, × HY, × NS, BB ×
	if isOneOf(after, ClassBA, ClassHY, ClassNS) || before == ClassBB {
		return false
	}
	// LB21b: SY × HL
	if before == ClassSY && after == ClassHL {
		return false
	}
	// LB22: × IN
	if after == ClassIN {
		return false
	}
	// LB23: (AL | HL) × NU, NU × (AL | HL)
	if (isOneOf(before, ClassAL, ClassHL) && after == ClassNU) ||
		(before == ClassNU && isOneOf(after, ClassAL, ClassHL)) {
		return false
	}
	// LB23a: PR × ID, ID × PO
	if (before == ClassPR && after == ClassID) || (before == ClassID && after == ClassPO) {
		return false
	}
	// LB24: (PR | PO) × (AL | HL), (AL | HL) × (PR | PO)
	if (isOneOf(before, ClassPR, ClassPO) && isOneOf(after, ClassAL, ClassHL)) ||
		(isOneOf(before, ClassAL, ClassHL) && isOneOf(after, ClassPR, ClassPO)) {
		return false
	}
	// LB25: numeric expressions (simplified pair form)
	switch {
	case isOneOf(before, ClassCL, ClassCP) && isOneOf(after, ClassPO, ClassPR):
		return false
	case before == ClassNU && isOneOf(after, ClassPO, ClassPR, ClassNU):
		return false
	case isOneOf(before, ClassPO, ClassPR) && isOneOf(after, ClassOP, ClassNU):
		return false
	case isOneOf(before, ClassHY, ClassIS, ClassSY) && after == ClassNU:
		return false
	}
	// LB26: Korean syllable blocks
	switch {
	case before == ClassJL && isOneOf(after, ClassJL, ClassJV, ClassH2, ClassH3):
		return false
	case isOneOf(before, ClassJV, ClassH2) && isOneOf(after, ClassJV, ClassJT):
		return false
	case isOneOf(before, ClassJT, ClassH3) && after == ClassJT:
		return false
	}
	// LB27: (JL | JV | JT | H2 | H3) × PO, PR × (JL | JV | JT | H2 | H3)
	if (isOneOf(before, ClassJL, ClassJV, ClassJT, ClassH2, ClassH3) && after == ClassPO) ||
		(before == ClassPR && isOneOf(after, ClassJL, ClassJV, ClassJT, ClassH2, ClassH3)) {
		return false
	}
	// LB28: (AL | HL) × (AL | HL)
	if isOneOf(before, ClassAL, ClassHL) && isOneOf(after, ClassAL, ClassHL) {
		return false
	}
	// LB29: IS × (AL | HL)
	if before == ClassIS && isOneOf(after, ClassAL, ClassHL) {
		return false
	}
	// LB30: (AL | HL | NU) × OP, CP × (AL | HL | NU)
	if (isOneOf(before, ClassAL, ClassHL, ClassNU) && after == ClassOP) ||
		(before == ClassCP && isOneOf(after, ClassAL, ClassHL, ClassNU)) {
		return false
	}
	// LB31: ÷ everywhere else
	return true
}

// findLineBreakOpportunities finds all valid line break opportunities in text.
// Returns a slice of byte positions where breaks are allowed.
func findLineBreakOpportunities(text string) []int {
	return findLineBreakOpportunitiesWithHyphens(text, HyphensManual)
}

// findLineBreakOpportunitiesWithHyphens finds all valid line break
// opportunities in text. Returns a slice of byte positions where breaks are
// allowed, always starting with 0 and ending with len(text).
//
// The CSS hyphens property only controls hyphenation, i.e. breaks at soft
// hyphens (U+00AD) and dictionary-inserted hyphens. Break opportunities
// after ordinary hyphens (UAX #14 LB21) are always honored.
// https://www.w3.org/TR/css-text-3/#hyphens-property
func findLineBreakOpportunitiesWithHyphens(text string, hyphens Hyphens) []int {
	if text == "" {
		return []int{0}
	}

	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return []int{0, len(text)}
	}

	classes := make([]BreakClass, n)
	for i, r := range runes {
		classes[i] = getBreakClass(r)
	}

	breakPoints := make([]int, 0, n/4+2)
	breakPoints = append(breakPoints, 0) // Start is always a break point

	// before is the class of the last non-space character, after LB9/LB10
	// resolution. hadSpace tracks intervening SP characters.
	before := classes[0]
	if before == ClassCM {
		before = ClassAL // LB10
	}
	hadSpace := before == ClassSP

	// bytePos is the byte offset of runes[i]; it is advanced by the encoded
	// length of each rune so we never index runes with byte offsets.
	bytePos := utf8.RuneLen(runes[0])

	// The loop makes exactly one step per rune.
	for i := 1; i < n; i++ {
		cur := classes[i]
		prevImmediate := classes[i-1]

		allowed := false
		switch {
		case isMandatoryBreakClass(prevImmediate):
			// LB4, LB5: break after hard line breaks, except inside CR LF
			allowed = !(prevImmediate == ClassCR && cur == ClassLF)
		case isMandatoryBreakClass(cur):
			// LB6: never break before a hard line break
			allowed = false
		case cur == ClassCM && !hadSpace && !isOneOf(before, ClassSP, ClassZW) && !isMandatoryBreakClass(before):
			// LB9: X CM* → X; the mark attaches to the preceding character
			allowed = false
		default:
			next := ClassXX
			if i+1 < n {
				next = classes[i+1]
			}
			allowed = pairBreakAllowed(before, cur, hadSpace, next)
		}

		if allowed {
			// CSS hyphens: none disables breaks at soft hyphens (U+00AD).
			// https://www.w3.org/TR/css-text-3/#valdef-hyphens-none
			softHyphenBreak := runes[i-1] == '\u00AD'
			if !(softHyphenBreak && hyphens == HyphensNone) {
				breakPoints = append(breakPoints, bytePos)
			}
		}

		// Update state
		switch {
		case cur == ClassSP:
			hadSpace = true
		case cur == ClassCM && !hadSpace && !isOneOf(before, ClassSP, ClassZW) && !isMandatoryBreakClass(before):
			// LB9: combining mark does not change the base class
		default:
			if cur == ClassCM {
				cur = ClassAL // LB10
			}
			before = cur
			hadSpace = false
		}

		bytePos += utf8.RuneLen(runes[i])
	}

	// End of text is always a break point
	if breakPoints[len(breakPoints)-1] != len(text) {
		breakPoints = append(breakPoints, len(text))
	}

	return breakPoints
}
