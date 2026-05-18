package layout

import (
	"math"

	"github.com/SCKelemen/units"
)

// Length represents a CSS <length> value with its unit.
//
// As of Phase 1 of the units<->layout migration, Length is a type alias for
// github.com/SCKelemen/units.Length so layout shares the canonical type
// defined in the units package. This makes layout-side and units-side
// Length values trivially interoperable.
//
// Length values can be absolute (px) or relative (em, rem, ch, vh, vw, ...).
// Relative units are resolved to pixels during layout using a LayoutContext.
//
// Based on CSS Values and Units Module Level 4:
// https://www.w3.org/TR/css-values-4/
type Length = units.Length

// LengthUnit identifies the unit type of a Length value.
//
// LengthUnit is a type alias for github.com/SCKelemen/units.LengthUnit, which
// is defined as `type LengthUnit string` per the CSS Values L4 spec. Layout's
// historic int/iota-based LengthUnit has been replaced by the canonical
// units.LengthUnit.
type LengthUnit = units.LengthUnit

// Layout's historic named unit constants. These are preserved as re-exports
// of the corresponding github.com/SCKelemen/units constants so existing
// layout consumers (e.g. layout.Pixels, layout.EmUnit, ...) keep compiling.
//
// UnboundedUnit remains a layout-specific sentinel value with no CSS L4
// equivalent; it is used as the upper bound for unconstrained layout passes.
const (
	// Absolute length units (CSS reference pixel: 1in = 96px).
	Pixels  = units.PX // 1px = 1/96 of 1 inch (anchor unit)
	PtUnit  = units.PT // 1pt = 1/72 of 1 inch
	PcUnit  = units.PC // 1pc = 12pt = 1/6 inch
	InUnit  = units.IN // 1in = 96px
	CmUnit  = units.CM // 1cm = 96/2.54 px
	MmUnit  = units.MM // 1mm = 1/10 cm
	QUnit   = units.QQ // 1Q  = 1/40 cm

	// Font-relative units.
	EmUnit  = units.EM  // 1em  = current element font-size
	RemUnit = units.REM // 1rem = root element font-size
	ChUnit  = units.CH  // 1ch  = advance measure of '0' glyph

	// Viewport-relative units.
	VhUnit   = units.VH   // 1vh   = 1% of viewport height
	VwUnit   = units.VW   // 1vw   = 1% of viewport width
	VmaxUnit = units.VMAX // 1vmax = 1% of larger viewport dimension
	VminUnit = units.VMIN // 1vmin = 1% of smaller viewport dimension

	// UnboundedUnit represents an unbounded length (infinity).
	// Layout-specific sentinel; not part of CSS L4. Used for maximum sizes
	// that have no limit (e.g. unconstrained layout passes).
	UnboundedUnit LengthUnit = "unbounded"
)

// ─────────────────────────────────────────────────────────────────────────
// Constructors
// ─────────────────────────────────────────────────────────────────────────
//
// These re-export the units constructors so existing layout-side call sites
// (e.g. layout.Px(10), layout.Em(1.5)) keep compiling unchanged.

// Px creates a Length in pixels.
func Px(value float64) Length { return units.Px(value) }

// Pt creates a Length in points (1pt = 1/72 inch).
func Pt(value float64) Length { return units.Pt(value) }

// Pc creates a Length in picas (1pc = 12pt).
func Pc(value float64) Length { return units.Pc(value) }

// In creates a Length in inches (1in = 96px in CSS).
func In(value float64) Length { return units.In(value) }

// Cm creates a Length in centimeters (1cm ≈ 37.8px).
func Cm(value float64) Length { return units.Cm(value) }

// Mm creates a Length in millimeters (1mm ≈ 3.78px).
func Mm(value float64) Length { return units.Mm(value) }

// Q creates a Length in quarter-millimeters (1Q ≈ 0.945px).
func Q(value float64) Length { return units.Q(value) }

// Em creates a Length in em units (relative to element font size).
func Em(value float64) Length { return units.Em(value) }

// Rem creates a Length in rem units (relative to root font size).
func Rem(value float64) Length { return units.Rem(value) }

// Ch creates a Length in ch units (relative to '0' character width).
func Ch(value float64) Length { return units.Ch(value) }

// Vh creates a Length in vh units (relative to viewport height).
func Vh(value float64) Length { return units.Vh(value) }

// Vw creates a Length in vw units (relative to viewport width).
func Vw(value float64) Length { return units.Vw(value) }

// Vmax creates a Length in vmax units (relative to larger viewport dimension).
func Vmax(value float64) Length { return units.Vmax(value) }

// Vmin creates a Length in vmin units (relative to smaller viewport dimension).
func Vmin(value float64) Length { return units.Vmin(value) }

// Container-query length constructors. These create container-relative
// lengths (cq*) that resolve against the nearest query container when
// passed to ResolveLengthInContext. See CSS Containment Module Level 3:
// https://www.w3.org/TR/css-contain-3/#container-lengths

// Cqw creates a Length in cqw units (1% of query container width).
func Cqw(value float64) Length { return units.Cqw(value) }

// Cqh creates a Length in cqh units (1% of query container height).
func Cqh(value float64) Length { return units.Cqh(value) }

// Cqi creates a Length in cqi units (1% of query container inline size).
func Cqi(value float64) Length { return units.Cqi(value) }

// Cqb creates a Length in cqb units (1% of query container block size).
func Cqb(value float64) Length { return units.Cqb(value) }

// Cqmin creates a Length in cqmin units (1% of the smaller of cqi and cqb).
func Cqmin(value float64) Length { return units.Cqmin(value) }

// Cqmax creates a Length in cqmax units (1% of the larger of cqi and cqb).
func Cqmax(value float64) Length { return units.Cqmax(value) }

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
// Most of the unit math is delegated to github.com/SCKelemen/units via
// units.Length.Resolve, which provides resolvers for the full CSS L4 unit
// set (absolute, font-relative, viewport-relative, container-relative).
// Two pieces of behavior remain layout-specific:
//
//   - UnboundedUnit short-circuits to math.MaxFloat64. It is a layout-only
//     sentinel; the units package has no concept of it.
//   - Unknown / unsupported units (e.g. cq*, vi/vb when the corresponding
//     context fields are unset) preserve the pre-migration default-case
//     behavior of returning l.Value unchanged.
//
// For ancestor-aware container query resolution (cq*), use
// ResolveLengthInContext, which is the only path that populates
// units.Context.ContainerWidth / ContainerHeight.
func ResolveLength(l Length, ctx *LayoutContext, currentFontSize float64) float64 {
	// Layout-specific sentinel: not in CSS, units pkg doesn't know it.
	if l.Unit == UnboundedUnit {
		return math.MaxFloat64
	}

	uctx := buildUnitsContext(ctx, currentFontSize)
	resolved, err := l.Resolve(uctx)
	if err != nil {
		// Resolution failure (unknown unit, missing context field).
		// Preserve pre-migration default-case behavior: "return value as-is".
		return l.Value
	}
	return resolved.Value
}

// buildUnitsContext maps a layout-side LayoutContext (plus the current
// element's font size) onto a units.Context.
//
// The mapping reflects layout's terminal-cell-grid simplifications:
//
//   - ex (x-height), cap (cap-height), and lh (line-height) all collapse
//     to the current font size — terminals do not provide distinct font
//     metrics for these.
//   - rlh (root line-height) collapses to the root font size for the same
//     reason.
//   - ic (CJK ideograph advance) is set to 2 * ch on the assumption that
//     full-width CJK glyphs occupy two terminal cells.
//
// ContainerWidth and ContainerHeight are intentionally left zero here.
// They are populated only by ResolveLengthInContext, which has the
// NodeContext required to walk ancestors and find a query container.
//
// A nil LayoutContext is accepted and produces a minimal units.Context
// populated only with currentFontSize. This matches the pre-migration
// behavior of ResolveLength, which short-circuited the absolute units
// (Pixels, Pt, Pc, In, Cm, Mm, Q) without touching ctx and would only
// panic on nil ctx for relative units. After delegation, relative units
// against a nil context cleanly fail in units.Length.Resolve and fall
// back to l.Value via the error path in ResolveLength.
func buildUnitsContext(ctx *LayoutContext, currentFontSize float64) *units.Context {
	if ctx == nil {
		return &units.Context{FontSize: currentFontSize}
	}
	chWidth := measureCharWidth(ctx.ChReferenceChar, currentFontSize, ctx.TextMetrics)
	return &units.Context{
		FontSize:       currentFontSize,
		RootFontSize:   ctx.RootFontSize,
		XHeight:        currentFontSize,
		CapHeight:      currentFontSize,
		ChWidth:        chWidth,
		IcWidth:        chWidth * 2,
		LineHeight:     currentFontSize,
		RootLineHeight: ctx.RootFontSize,
		ViewportWidth:  ctx.ViewportWidth,
		ViewportHeight: ctx.ViewportHeight,
		// ContainerWidth / ContainerHeight left zero; populated by
		// ResolveLengthInContext when a NodeContext is available.
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
