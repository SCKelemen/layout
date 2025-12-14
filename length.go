package layout

import (
	"fmt"
	"math"
)

// Length represents a CSS <length> value with its unit.
// Based on CSS Values and Units Module Level 4: https://www.w3.org/TR/css-values-4/
//
// Length values can be absolute (px) or relative (em, rem, ch, vh, vw).
// Relative units are resolved to pixels during layout using a LayoutContext.
type Length struct {
	Value float64
	Unit  LengthUnit
}

// LengthUnit represents the unit type for a Length value.
type LengthUnit int

const (
	// Pixels represents an absolute pixel unit (px).
	Pixels LengthUnit = iota

	// EmUnit represents a length relative to the element's font size.
	// 1em = current element's font size in points.
	EmUnit

	// RemUnit represents a length relative to the root element's font size.
	// 1rem = root font size in points.
	RemUnit

	// ChUnit represents a length relative to the width of the '0' character.
	// 1ch = width of the '0' glyph in the element's font.
	ChUnit

	// VhUnit represents a length relative to viewport height.
	// 1vh = 1% of viewport height.
	VhUnit

	// VwUnit represents a length relative to viewport width.
	// 1vw = 1% of viewport width.
	VwUnit

	// VmaxUnit represents a length relative to the larger viewport dimension.
	// 1vmax = 1% of max(viewport width, viewport height).
	VmaxUnit

	// VminUnit represents a length relative to the smaller viewport dimension.
	// 1vmin = 1% of min(viewport width, viewport height).
	VminUnit

	// UnboundedUnit represents an unbounded length (infinity).
	// Used for maximum sizes that have no limit.
	UnboundedUnit
)

// String returns a string representation of the LengthUnit.
func (u LengthUnit) String() string {
	switch u {
	case Pixels:
		return "px"
	case EmUnit:
		return "em"
	case RemUnit:
		return "rem"
	case ChUnit:
		return "ch"
	case VhUnit:
		return "vh"
	case VwUnit:
		return "vw"
	case VmaxUnit:
		return "vmax"
	case VminUnit:
		return "vmin"
	case UnboundedUnit:
		return "unbounded"
	default:
		return "unknown"
	}
}

// String returns a string representation of the Length.
func (l Length) String() string {
	return fmt.Sprintf("%.2f%s", l.Value, l.Unit)
}

// Px creates a Length in pixels.
func Px(value float64) Length {
	return Length{Value: value, Unit: Pixels}
}

// Em creates a Length in em units (relative to element font size).
func Em(value float64) Length {
	return Length{Value: value, Unit: EmUnit}
}

// Rem creates a Length in rem units (relative to root font size).
func Rem(value float64) Length {
	return Length{Value: value, Unit: RemUnit}
}

// Ch creates a Length in ch units (relative to '0' character width).
func Ch(value float64) Length {
	return Length{Value: value, Unit: ChUnit}
}

// Vh creates a Length in vh units (relative to viewport height).
func Vh(value float64) Length {
	return Length{Value: value, Unit: VhUnit}
}

// Vw creates a Length in vw units (relative to viewport width).
func Vw(value float64) Length {
	return Length{Value: value, Unit: VwUnit}
}

// Vmax creates a Length in vmax units (relative to larger viewport dimension).
func Vmax(value float64) Length {
	return Length{Value: value, Unit: VmaxUnit}
}

// Vmin creates a Length in vmin units (relative to smaller viewport dimension).
func Vmin(value float64) Length {
	return Length{Value: value, Unit: VminUnit}
}

// PxUnbounded is a pre-allocated unbounded pixel length for performance.
// Equivalent to Px(math.MaxFloat64) but avoids repeated allocations.
var PxUnbounded = Length{Value: math.MaxFloat64, Unit: Pixels}

// UnboundedLength creates an unbounded Length.
// This is more semantically clear than Px(math.MaxFloat64).
func UnboundedLength() Length {
	return Length{Value: 0, Unit: UnboundedUnit}
}

// ResolveLength converts a Length to pixels using the provided context.
//
// Parameters:
//   - l: The Length to resolve
//   - ctx: LayoutContext containing viewport size, root font size, and text metrics
//   - currentFontSize: The current element's font size in points (for em unit resolution)
//
// Returns the resolved length in pixels.
//
// Resolution rules:
//   - Pixels: returned as-is
//   - Em: multiplied by currentFontSize
//   - Rem: multiplied by ctx.RootFontSize
//   - Ch: multiplied by the width of ctx.ChReferenceChar
//   - Vh: (value / 100) * ctx.ViewportHeight
//   - Vw: (value / 100) * ctx.ViewportWidth
//   - Vmax: (value / 100) * max(ctx.ViewportWidth, ctx.ViewportHeight)
//   - Vmin: (value / 100) * min(ctx.ViewportWidth, ctx.ViewportHeight)
//   - UnboundedUnit: returns math.MaxFloat64
func ResolveLength(l Length, ctx *LayoutContext, currentFontSize float64) float64 {
	switch l.Unit {
	case Pixels:
		return l.Value

	case EmUnit:
		// Relative to current element's font size
		return l.Value * currentFontSize

	case RemUnit:
		// Relative to root font size
		return l.Value * ctx.RootFontSize

	case ChUnit:
		// Measure the reference character width
		charWidth := measureCharWidth(ctx.ChReferenceChar, currentFontSize, ctx.TextMetrics)
		return l.Value * charWidth

	case VhUnit:
		// 1vh = 1% of viewport height
		return (l.Value / 100.0) * ctx.ViewportHeight

	case VwUnit:
		// 1vw = 1% of viewport width
		return (l.Value / 100.0) * ctx.ViewportWidth

	case VmaxUnit:
		// 1vmax = 1% of larger viewport dimension
		maxDimension := math.Max(ctx.ViewportWidth, ctx.ViewportHeight)
		return (l.Value / 100.0) * maxDimension

	case VminUnit:
		// 1vmin = 1% of smaller viewport dimension
		minDimension := math.Min(ctx.ViewportWidth, ctx.ViewportHeight)
		return (l.Value / 100.0) * minDimension

	case UnboundedUnit:
		// Unbounded length resolves to infinity
		return math.MaxFloat64

	default:
		// Unknown unit, return value as-is
		return l.Value
	}
}

// measureCharWidth estimates the width of a character using text metrics.
// For now, uses monospace approximation via TextMetricsProvider.
// Can be swapped for true text measurement (HarfBuzz, FreeType) in the future.
func measureCharWidth(char rune, fontSize float64, metrics TextMetricsProvider) float64 {
	if metrics == nil {
		// Fallback: monospace approximation (60% of font size)
		return fontSize * 0.6
	}

	style := TextStyle{
		FontSize: fontSize,
		// Use defaults for other fields
	}

	// Measure a single character
	width, _, _ := metrics.Measure(string(char), style)
	return width
}
