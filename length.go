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

	// Absolute length units (CSS spec)
	// Based on CSS Values and Units Module Level 4
	// See: https://www.w3.org/TR/css-values-4/#absolute-lengths

	// PtUnit represents points (1pt = 1/72 inch).
	// Used primarily for print media and font sizes.
	PtUnit

	// PcUnit represents picas (1pc = 12pt = 1/6 inch).
	// Traditional typography unit.
	PcUnit

	// InUnit represents inches (1in = 96px in CSS).
	// Physical measurement unit.
	InUnit

	// CmUnit represents centimeters (1cm = 96px/2.54 ≈ 37.8px).
	// Metric measurement unit.
	CmUnit

	// MmUnit represents millimeters (1mm = 1/10 cm ≈ 3.78px).
	// Metric measurement unit.
	MmUnit

	// QUnit represents quarter-millimeters (1Q = 1/40 cm ≈ 0.945px).
	// Fine-grained metric unit.
	QUnit

	// Relative font units

	// EmUnit represents a length relative to the element's font size.
	// 1em = current element's font size in points.
	EmUnit

	// RemUnit represents a length relative to the root element's font size.
	// 1rem = root font size in points.
	RemUnit

	// ChUnit represents a length relative to the width of the '0' character.
	// 1ch = width of the '0' glyph in the element's font.
	ChUnit

	// Viewport units

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

	// Special units

	// UnboundedUnit represents an unbounded length (infinity).
	// Used for maximum sizes that have no limit.
	UnboundedUnit

	// Container query units (CSS Values and Units Module Level 4).
	// See: https://www.w3.org/TR/css-values-4/#container-relative-lengths
	//
	// These units resolve against the nearest ancestor query container
	// (an element whose Style.ContainerType is not ContainerTypeNormal).
	// When no such ancestor exists they fall back to the viewport.

	// CQWUnit represents a length relative to the query container's inline
	// size in horizontal writing modes (1cqw = 1% of container width).
	CQWUnit

	// CQHUnit represents a length relative to the query container's block
	// size in horizontal writing modes (1cqh = 1% of container height).
	CQHUnit

	// CQIUnit represents a length relative to the query container's inline
	// size (writing-mode aware, 1cqi = 1% of container inline-size).
	CQIUnit

	// CQBUnit represents a length relative to the query container's block
	// size (writing-mode aware, 1cqb = 1% of container block-size).
	CQBUnit

	// CQMinUnit represents the smaller of cqi and cqb.
	CQMinUnit

	// CQMaxUnit represents the larger of cqi and cqb.
	CQMaxUnit
)

// IsContainerQuery reports whether the unit is a CSS L4 container query
// unit (one of cqw, cqh, cqi, cqb, cqmin, cqmax).
func (u LengthUnit) IsContainerQuery() bool {
	switch u {
	case CQWUnit, CQHUnit, CQIUnit, CQBUnit, CQMinUnit, CQMaxUnit:
		return true
	default:
		return false
	}
}

// String returns a string representation of the LengthUnit.
func (u LengthUnit) String() string {
	switch u {
	case Pixels:
		return "px"
	case PtUnit:
		return "pt"
	case PcUnit:
		return "pc"
	case InUnit:
		return "in"
	case CmUnit:
		return "cm"
	case MmUnit:
		return "mm"
	case QUnit:
		return "Q"
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
	case CQWUnit:
		return "cqw"
	case CQHUnit:
		return "cqh"
	case CQIUnit:
		return "cqi"
	case CQBUnit:
		return "cqb"
	case CQMinUnit:
		return "cqmin"
	case CQMaxUnit:
		return "cqmax"
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

// Pt creates a Length in points (1pt = 1/72 inch).
func Pt(value float64) Length {
	return Length{Value: value, Unit: PtUnit}
}

// Pc creates a Length in picas (1pc = 12pt).
func Pc(value float64) Length {
	return Length{Value: value, Unit: PcUnit}
}

// In creates a Length in inches (1in = 96px in CSS).
func In(value float64) Length {
	return Length{Value: value, Unit: InUnit}
}

// Cm creates a Length in centimeters (1cm ≈ 37.8px).
func Cm(value float64) Length {
	return Length{Value: value, Unit: CmUnit}
}

// Mm creates a Length in millimeters (1mm ≈ 3.78px).
func Mm(value float64) Length {
	return Length{Value: value, Unit: MmUnit}
}

// Q creates a Length in quarter-millimeters (1Q ≈ 0.945px).
func Q(value float64) Length {
	return Length{Value: value, Unit: QUnit}
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

// Cqw creates a Length in cqw units (1% of the nearest query container's
// inline size in horizontal writing modes).
// Based on CSS Values and Units Module Level 4:
// https://www.w3.org/TR/css-values-4/#container-relative-lengths
func Cqw(value float64) Length {
	return Length{Value: value, Unit: CQWUnit}
}

// Cqh creates a Length in cqh units (1% of the nearest query container's
// block size in horizontal writing modes).
func Cqh(value float64) Length {
	return Length{Value: value, Unit: CQHUnit}
}

// Cqi creates a Length in cqi units (1% of the nearest query container's
// inline size, writing-mode aware).
func Cqi(value float64) Length {
	return Length{Value: value, Unit: CQIUnit}
}

// Cqb creates a Length in cqb units (1% of the nearest query container's
// block size, writing-mode aware).
func Cqb(value float64) Length {
	return Length{Value: value, Unit: CQBUnit}
}

// Cqmin creates a Length in cqmin units (1% of the smaller of cqi and cqb).
func Cqmin(value float64) Length {
	return Length{Value: value, Unit: CQMinUnit}
}

// Cqmax creates a Length in cqmax units (1% of the larger of cqi and cqb).
func Cqmax(value float64) Length {
	return Length{Value: value, Unit: CQMaxUnit}
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
//   - Absolute units: converted using CSS reference pixel (1in = 96px)
//   - Pt: 1pt = 1/72 inch = 96/72 px ≈ 1.333px
//   - Pc: 1pc = 12pt = 16px
//   - In: 1in = 96px
//   - Cm: 1cm = 96/2.54 px ≈ 37.795px
//   - Mm: 1mm = 96/25.4 px ≈ 3.7795px
//   - Q: 1Q = 96/101.6 px ≈ 0.945px
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

	// Absolute length units (based on CSS reference pixel: 1in = 96px)
	case PtUnit:
		// 1pt = 1/72 inch
		return l.Value * (96.0 / 72.0)

	case PcUnit:
		// 1pc = 12pt = 1/6 inch
		return l.Value * 16.0

	case InUnit:
		// 1in = 96px (CSS reference pixel)
		return l.Value * 96.0

	case CmUnit:
		// 1cm = 1/2.54 inch
		return l.Value * (96.0 / 2.54)

	case MmUnit:
		// 1mm = 1/25.4 inch
		return l.Value * (96.0 / 25.4)

	case QUnit:
		// 1Q = 1/40 cm = 1/101.6 inch
		return l.Value * (96.0 / 101.6)

	// Relative font units
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

	// Viewport units
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

	case CQWUnit, CQIUnit:
		// Container query inline-axis units. Without a NodeContext we have
		// no ancestor chain to walk: fall back to the viewport per the L4
		// spec's "no query container" rule. See ResolveLengthInContext for
		// the container-aware resolver.
		return (l.Value / 100.0) * ctx.ViewportWidth

	case CQHUnit, CQBUnit:
		// Container query block-axis units. Same fallback as above.
		return (l.Value / 100.0) * ctx.ViewportHeight

	case CQMinUnit:
		minDimension := math.Min(ctx.ViewportWidth, ctx.ViewportHeight)
		return (l.Value / 100.0) * minDimension

	case CQMaxUnit:
		maxDimension := math.Max(ctx.ViewportWidth, ctx.ViewportHeight)
		return (l.Value / 100.0) * maxDimension

	default:
		// Unknown unit, return value as-is
		return l.Value
	}
}

// ResolveLengthInContext resolves a Length to pixels while honoring CSS
// container query units (cqw, cqh, cqi, cqb, cqmin, cqmax). The provided
// NodeContext is used to walk the ancestor chain to locate the nearest
// query container.
//
// Resolution rules for cq* units:
//   - cqw / cqi: nearest ancestor whose ContainerType is `size` or
//     `inline-size`. Resolved against that ancestor's measured inline-size
//     (width in horizontal writing modes, height in vertical writing modes).
//   - cqh / cqb: nearest ancestor whose ContainerType is `size`. Ancestors
//     with `inline-size` do not satisfy a block-axis query and the walk
//     continues. Resolved against the matching ancestor's measured
//     block-size.
//   - cqmin / cqmax: resolved as min/max of the cqi and cqb resolutions,
//     each computed independently (so a `size` container yields container
//     min/max, while a chain that only answers the inline axis effectively
//     falls back to viewport for the block side).
//   - If no matching ancestor is found, the unit falls back to the
//     corresponding viewport dimension.
//
// All other units defer to ResolveLength.
//
// Resolution timing note: cq* resolution reads the matching ancestor's
// measured Rect. Layout proceeds parent-first, so by the time a child's
// lengths are resolved the parent's Rect has been computed for the current
// pass. There is no special handling for circular dependencies (a child
// whose container-query-derived size influences a parent that container-
// queries back is resolved with "last-wins" semantics: subsequent layout
// passes will see updated sizes, but no fixed-point iteration is performed).
func ResolveLengthInContext(l Length, ctx *LayoutContext, currentFontSize float64, nctx *NodeContext) float64 {
	if !l.Unit.IsContainerQuery() {
		return ResolveLength(l, ctx, currentFontSize)
	}

	// Writing mode for axis disambiguation. We use the writing mode of the
	// node whose property is being resolved (its containing context's
	// writing mode dictates whether cqw/cqh map to inline/block).
	var mode WritingMode
	if nctx != nil && nctx.Node != nil {
		mode = nctx.Node.Style.WritingMode
	}

	switch l.Unit {
	case CQWUnit:
		// Physical: width-aligned. In CSS L4 cqw is defined relative to
		// the container's inline size when in a horizontal writing mode;
		// implementations align cqw with cqi here as the inline axis.
		return cqAxisSize(l.Value, ctx, nctx, ContainerAxisInline, mode)
	case CQIUnit:
		return cqAxisSize(l.Value, ctx, nctx, ContainerAxisInline, mode)
	case CQHUnit:
		return cqAxisSize(l.Value, ctx, nctx, ContainerAxisBlock, mode)
	case CQBUnit:
		return cqAxisSize(l.Value, ctx, nctx, ContainerAxisBlock, mode)
	case CQMinUnit:
		i := cqAxisSize(l.Value, ctx, nctx, ContainerAxisInline, mode)
		b := cqAxisSize(l.Value, ctx, nctx, ContainerAxisBlock, mode)
		return math.Min(i, b)
	case CQMaxUnit:
		i := cqAxisSize(l.Value, ctx, nctx, ContainerAxisInline, mode)
		b := cqAxisSize(l.Value, ctx, nctx, ContainerAxisBlock, mode)
		return math.Max(i, b)
	}
	return l.Value
}

// cqAxisSize resolves a container-query length value on the requested
// logical axis. percent is the literal numeric prefix (e.g. 50 for "50cqw").
func cqAxisSize(percent float64, ctx *LayoutContext, nctx *NodeContext, axis ContainerAxis, mode WritingMode) float64 {
	resolved := resolveContainerQuery(nctx, axis)
	var basis float64
	if resolved.Found {
		basis = resolved.Size
	} else {
		basis = containerQueryViewportSize(ctx, axis, mode)
	}
	return (percent / 100.0) * basis
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
