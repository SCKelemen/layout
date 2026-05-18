package layout

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// ParseLength parses a CSS-style length token of the form `<number><unit>`
// (for example "10px", "1.5rem", "100vw", "50svmin") into a Length value.
//
// Spec: https://www.w3.org/TR/css-values-4/#lengths
//
// Identifiers are matched case-insensitively per the CSS spec
// (https://www.w3.org/TR/css-syntax-3/#ref-for-ident-token). The numeric
// component is parsed with strconv.ParseFloat and accepts an optional
// leading sign plus decimal digits. Scientific notation in the number
// component (e.g. "1e2px") is not accepted here because the 'e' would be
// ambiguous with the `em` and `ex`-family unit prefixes; callers needing
// that form can stringify the number themselves. Leading and trailing
// whitespace is trimmed.
//
// Returns an error for the empty string, a missing unit, an unrecognized
// unit identifier, or an unparsable numeric component.
//
// Recognized units: px, pt, pc, in, cm, mm, q, em, rem, ch, vh, vw, vmin,
// vmax, plus all Level 4 aliases declared in length_l4.go (lh, rlh, ic, ric,
// cap, rcap, rch, rex, vi, vb, svw, svh, svi, svb, svmin, svmax, lvw, lvh,
// lvi, lvb, lvmin, lvmax, dvw, dvh, dvi, dvb, dvmin, dvmax).
func ParseLength(s string) (Length, error) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return Length{}, fmt.Errorf("layout: empty length string")
	}

	// Split into numeric prefix and unit suffix at the first rune that
	// isn't part of a number. Numeric runes are digits, '.', and a single
	// leading '+' or '-'.
	split := len(raw)
	for i, r := range raw {
		if isNumberRune(r, i) {
			continue
		}
		split = i
		break
	}
	if split == 0 {
		return Length{}, fmt.Errorf("layout: %q has no numeric component", s)
	}
	numPart := raw[:split]
	unitPart := raw[split:]
	if unitPart == "" {
		return Length{}, fmt.Errorf("layout: %q has no unit", s)
	}

	v, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return Length{}, fmt.Errorf("layout: invalid number %q in length: %w", numPart, err)
	}

	unit, ok := lookupUnit(unitPart)
	if !ok {
		return Length{}, fmt.Errorf("layout: unknown length unit %q", unitPart)
	}
	return Length{Value: v, Unit: unit}, nil
}

// isNumberRune reports whether r is part of the numeric prefix of a length
// token. The first position may legitimately be '+' or '-'. 'e'/'E' is
// deliberately excluded so that "10em"/"10ex" parse cleanly rather than
// being misread as the start of a scientific-notation exponent.
func isNumberRune(r rune, pos int) bool {
	if unicode.IsDigit(r) || r == '.' {
		return true
	}
	if pos == 0 && (r == '+' || r == '-') {
		return true
	}
	return false
}

// lookupUnit maps a unit identifier (any case) to a LengthUnit. It returns
// false for unrecognized identifiers.
func lookupUnit(s string) (LengthUnit, bool) {
	switch strings.ToLower(s) {
	// Existing absolute units.
	case "px":
		return Pixels, true
	case "pt":
		return PtUnit, true
	case "pc":
		return PcUnit, true
	case "in":
		return InUnit, true
	case "cm":
		return CmUnit, true
	case "mm":
		return MmUnit, true
	case "q":
		return QUnit, true

	// Existing relative/font units.
	case "em":
		return EmUnit, true
	case "rem":
		return RemUnit, true
	case "ch":
		return ChUnit, true

	// Existing viewport units.
	case "vh":
		return VhUnit, true
	case "vw":
		return VwUnit, true
	case "vmin":
		return VminUnit, true
	case "vmax":
		return VmaxUnit, true

	// Level 4 font-relative additions.
	case "lh":
		return LhUnit, true
	case "rlh":
		return RlhUnit, true
	case "ic":
		return IcUnit, true
	case "ric":
		return RicUnit, true
	case "cap":
		return CapUnit, true
	case "rcap":
		return RcapUnit, true
	case "rch":
		return RchUnit, true
	case "rex":
		return RexUnit, true

	// Level 4 logical viewport units.
	case "vi":
		return ViUnit, true
	case "vb":
		return VbUnit, true

	// Level 4 small viewport variants.
	case "svw":
		return SvwUnit, true
	case "svh":
		return SvhUnit, true
	case "svi":
		return SviUnit, true
	case "svb":
		return SvbUnit, true
	case "svmin":
		return SvminUnit, true
	case "svmax":
		return SvmaxUnit, true

	// Level 4 large viewport variants.
	case "lvw":
		return LvwUnit, true
	case "lvh":
		return LvhUnit, true
	case "lvi":
		return LviUnit, true
	case "lvb":
		return LvbUnit, true
	case "lvmin":
		return LvminUnit, true
	case "lvmax":
		return LvmaxUnit, true

	// Level 4 dynamic viewport variants.
	case "dvw":
		return DvwUnit, true
	case "dvh":
		return DvhUnit, true
	case "dvi":
		return DviUnit, true
	case "dvb":
		return DvbUnit, true
	case "dvmin":
		return DvminUnit, true
	case "dvmax":
		return DvmaxUnit, true
	}
	return 0, false
}
