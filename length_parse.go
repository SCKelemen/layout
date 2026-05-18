package layout

// Length value string parser. The layout library exposes a Go-typed API
// (Px, Em, Cqw, ...) for general use; this small parser is provided so
// callers can round-trip from CSS-like strings, primarily for tests, demos,
// and the new container-query property parsers (`container-type`, etc.)
// which accept string input.

import (
	"fmt"
	"strconv"
	"strings"
)

// unitTable maps lowercase CSS unit suffixes to their LengthUnit. Order
// here does not matter; lookups go through ParseLength's longest-match
// search below.
var unitTable = map[string]LengthUnit{
	"px":    Pixels,
	"pt":    PtUnit,
	"pc":    PcUnit,
	"in":    InUnit,
	"cm":    CmUnit,
	"mm":    MmUnit,
	"q":     QUnit,
	"em":    EmUnit,
	"rem":   RemUnit,
	"ch":    ChUnit,
	"vh":    VhUnit,
	"vw":    VwUnit,
	"vmax":  VmaxUnit,
	"vmin":  VminUnit,
	"cqw":   CQWUnit,
	"cqh":   CQHUnit,
	"cqi":   CQIUnit,
	"cqb":   CQBUnit,
	"cqmin": CQMinUnit,
	"cqmax": CQMaxUnit,
}

// ParseLength parses a CSS-style <length> string such as "50cqw", "1.5rem",
// or "100px" into a Length value. Unit identifiers are case-insensitive.
// A bare number (with no unit) is treated as pixels, matching the layout
// library's default unit. Whitespace surrounding the value is permitted.
//
// Examples:
//
//	ParseLength("50cqw")  -> Length{50, CQWUnit}
//	ParseLength("1.5rem") -> Length{1.5, RemUnit}
//	ParseLength("12")      -> Length{12, Pixels}
//	ParseLength("50CQW")  -> Length{50, CQWUnit}    (case-insensitive)
func ParseLength(s string) (Length, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return Length{}, fmt.Errorf("layout: empty length value")
	}

	// Find the boundary between number and unit. The numeric portion is
	// the longest prefix of digits, sign, decimal point, or 'e'/'E'.
	splitIdx := len(trimmed)
	for i, r := range trimmed {
		if r == '-' || r == '+' || r == '.' || (r >= '0' && r <= '9') {
			continue
		}
		if (r == 'e' || r == 'E') && i > 0 {
			// Permit exponent notation when preceded by a digit/sign.
			continue
		}
		splitIdx = i
		break
	}

	numStr := trimmed[:splitIdx]
	unitStr := strings.ToLower(strings.TrimSpace(trimmed[splitIdx:]))

	if numStr == "" {
		return Length{}, fmt.Errorf("layout: missing numeric value in %q", s)
	}
	value, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return Length{}, fmt.Errorf("layout: invalid number %q: %w", numStr, err)
	}

	if unitStr == "" {
		return Length{Value: value, Unit: Pixels}, nil
	}
	unit, ok := unitTable[unitStr]
	if !ok {
		return Length{}, fmt.Errorf("layout: unknown unit %q in %q", unitStr, s)
	}
	return Length{Value: value, Unit: unit}, nil
}
