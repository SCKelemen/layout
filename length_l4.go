package layout

import "math"

// CSS Values and Units Module Level 4 — additional length units.
//
// This file extends the base length unit set (px, em, rem, ch, vh, vw, vmin,
// vmax) with the Level 4 font-relative and logical viewport units, plus the
// small/large/dynamic viewport variants introduced for the same level.
//
// Spec: https://www.w3.org/TR/css-values-4/#font-relative-lengths
// Spec: https://www.w3.org/TR/css-values-4/#viewport-relative-lengths
//
// Terminal-context resolution rules
// ---------------------------------
// SCKelemen/layout targets monospace-terminal style layout. All cells in a
// terminal share a fixed grid, which collapses several distinctions that
// matter in proportional typography:
//
//   - The local and root line-box height are identical, so `lh` and `rlh`
//     both resolve to the current and root font size respectively (one row).
//   - The cap height, x-height, and ascender height all coincide with one
//     row, so `cap` / `rcap` / `rex` all resolve to one row as well.
//   - CJK ideographs occupy two cells horizontally in a typical terminal, so
//     `ic` / `ric` resolve to two character widths (two cells).
//   - The "small", "large", and "dynamic" viewport variants only differ on
//     browsers with collapsible UI chrome; a terminal has no such chrome, so
//     `sv*`, `lv*`, and `dv*` are exact aliases for the base viewport units.
//   - The logical viewport units `vi` and `vb` depend on writing-mode. The
//     resolver does not currently receive a writing-mode parameter, so they
//     are aliased to `vw` and `vh` respectively (matching the default
//     horizontal-tb writing mode). When a future revision threads the
//     element's WritingMode into ResolveLength, the mapping should swap for
//     vertical writing modes.

// L4 LengthUnit values. These are appended to the LengthUnit enum so that
// existing iota-based values remain stable. The Level 4 spec uses
// case-insensitive identifiers; parsing normalizes to lowercase.
const (
	// LhUnit represents the line height of the element ("lh").
	// Terminal context: 1lh = 1 row = current font size.
	LhUnit LengthUnit = iota + 100

	// RlhUnit represents the root element's line height ("rlh").
	// Terminal context: 1rlh = 1 row = root font size.
	RlhUnit

	// IcUnit represents the advance measure of the CJK water ideograph
	// U+6C34 ("ic"). Terminal context: 1ic = 2 character cells (CJK is
	// rendered double-width in monospace terminals).
	IcUnit

	// RicUnit represents the root element's ic value ("ric").
	// Terminal context: 1ric = 2 root character cells.
	RicUnit

	// CapUnit represents the cap height of the font ("cap").
	// Terminal context: 1cap = 1 row (aliases ex; terminals collapse cap
	// height and x-height into the cell grid).
	CapUnit

	// RcapUnit represents the root element's cap height ("rcap").
	// Terminal context: 1rcap = 1 root row.
	RcapUnit

	// RchUnit represents the root element's ch value ("rch").
	// Terminal context: 1rch = width of the ch reference character at the
	// root font size.
	RchUnit

	// RexUnit represents the root element's ex value ("rex").
	// Terminal context: 1rex = 1 root row (aliases ex at the root).
	RexUnit

	// ViUnit represents 1% of the viewport's inline size ("vi").
	// Terminal context: aliased to VwUnit for horizontal-tb writing mode.
	// See package comment for the writing-mode simplification.
	ViUnit

	// VbUnit represents 1% of the viewport's block size ("vb").
	// Terminal context: aliased to VhUnit for horizontal-tb writing mode.
	// See package comment for the writing-mode simplification.
	VbUnit

	// Small viewport variants ("sv*"). Terminals have no UI chrome that
	// can hide or reveal, so these are exact aliases for the base
	// viewport units.

	// SvwUnit aliases VwUnit. Spec: 1% of small viewport width.
	SvwUnit
	// SvhUnit aliases VhUnit. Spec: 1% of small viewport height.
	SvhUnit
	// SviUnit aliases ViUnit. Spec: 1% of small viewport inline size.
	SviUnit
	// SvbUnit aliases VbUnit. Spec: 1% of small viewport block size.
	SvbUnit
	// SvminUnit aliases VminUnit. Spec: small viewport min.
	SvminUnit
	// SvmaxUnit aliases VmaxUnit. Spec: small viewport max.
	SvmaxUnit

	// Large viewport variants ("lv*"). Aliased to the base viewport units
	// in terminal context.

	// LvwUnit aliases VwUnit. Spec: 1% of large viewport width.
	LvwUnit
	// LvhUnit aliases VhUnit. Spec: 1% of large viewport height.
	LvhUnit
	// LviUnit aliases ViUnit. Spec: 1% of large viewport inline size.
	LviUnit
	// LvbUnit aliases VbUnit. Spec: 1% of large viewport block size.
	LvbUnit
	// LvminUnit aliases VminUnit. Spec: large viewport min.
	LvminUnit
	// LvmaxUnit aliases VmaxUnit. Spec: large viewport max.
	LvmaxUnit

	// Dynamic viewport variants ("dv*"). Aliased to the base viewport
	// units in terminal context.

	// DvwUnit aliases VwUnit. Spec: 1% of dynamic viewport width.
	DvwUnit
	// DvhUnit aliases VhUnit. Spec: 1% of dynamic viewport height.
	DvhUnit
	// DviUnit aliases ViUnit. Spec: 1% of dynamic viewport inline size.
	DviUnit
	// DvbUnit aliases VbUnit. Spec: 1% of dynamic viewport block size.
	DvbUnit
	// DvminUnit aliases VminUnit. Spec: dynamic viewport min.
	DvminUnit
	// DvmaxUnit aliases VmaxUnit. Spec: dynamic viewport max.
	DvmaxUnit
)

// stringL4 returns the canonical string for an L4 length unit, or "" if the
// unit is not one of the L4 additions. It is consulted by LengthUnit.String.
func stringL4(u LengthUnit) string {
	switch u {
	case LhUnit:
		return "lh"
	case RlhUnit:
		return "rlh"
	case IcUnit:
		return "ic"
	case RicUnit:
		return "ric"
	case CapUnit:
		return "cap"
	case RcapUnit:
		return "rcap"
	case RchUnit:
		return "rch"
	case RexUnit:
		return "rex"
	case ViUnit:
		return "vi"
	case VbUnit:
		return "vb"
	case SvwUnit:
		return "svw"
	case SvhUnit:
		return "svh"
	case SviUnit:
		return "svi"
	case SvbUnit:
		return "svb"
	case SvminUnit:
		return "svmin"
	case SvmaxUnit:
		return "svmax"
	case LvwUnit:
		return "lvw"
	case LvhUnit:
		return "lvh"
	case LviUnit:
		return "lvi"
	case LvbUnit:
		return "lvb"
	case LvminUnit:
		return "lvmin"
	case LvmaxUnit:
		return "lvmax"
	case DvwUnit:
		return "dvw"
	case DvhUnit:
		return "dvh"
	case DviUnit:
		return "dvi"
	case DvbUnit:
		return "dvb"
	case DvminUnit:
		return "dvmin"
	case DvmaxUnit:
		return "dvmax"
	}
	return ""
}

// resolveL4 resolves an L4 length unit to pixels for the given context. If
// the unit is not an L4 unit it returns (0, false) so that the base
// ResolveLength can fall through to its own switch.
func resolveL4(l Length, ctx *LayoutContext, currentFontSize float64) (float64, bool) {
	switch l.Unit {
	// Font-relative units. In a monospace terminal one row equals one font
	// size unit, so lh / cap / (the implicit) ex all collapse to the
	// element (or root) font size.
	case LhUnit:
		return l.Value * currentFontSize, true
	case RlhUnit:
		return l.Value * ctx.RootFontSize, true
	case CapUnit:
		return l.Value * currentFontSize, true
	case RcapUnit:
		return l.Value * ctx.RootFontSize, true
	case RexUnit:
		return l.Value * ctx.RootFontSize, true

	case IcUnit:
		// CJK ideographs occupy two cells horizontally in a terminal.
		w := measureCharWidth(ctx.ChReferenceChar, currentFontSize, ctx.TextMetrics)
		return l.Value * 2 * w, true
	case RicUnit:
		w := measureCharWidth(ctx.ChReferenceChar, ctx.RootFontSize, ctx.TextMetrics)
		return l.Value * 2 * w, true
	case RchUnit:
		w := measureCharWidth(ctx.ChReferenceChar, ctx.RootFontSize, ctx.TextMetrics)
		return l.Value * w, true

	// Logical viewport units. Without a writing-mode parameter the
	// resolver assumes horizontal-tb, so vi → vw and vb → vh. See the
	// package comment at the top of this file.
	case ViUnit:
		return (l.Value / 100.0) * ctx.ViewportWidth, true
	case VbUnit:
		return (l.Value / 100.0) * ctx.ViewportHeight, true

	// Small / large / dynamic viewport variants. Terminals lack the
	// browser UI chrome distinction these were designed for, so each
	// resolves identically to the corresponding base viewport unit.
	case SvwUnit, LvwUnit, DvwUnit:
		return (l.Value / 100.0) * ctx.ViewportWidth, true
	case SvhUnit, LvhUnit, DvhUnit:
		return (l.Value / 100.0) * ctx.ViewportHeight, true
	case SviUnit, LviUnit, DviUnit:
		return (l.Value / 100.0) * ctx.ViewportWidth, true
	case SvbUnit, LvbUnit, DvbUnit:
		return (l.Value / 100.0) * ctx.ViewportHeight, true
	case SvminUnit, LvminUnit, DvminUnit:
		return (l.Value / 100.0) * math.Min(ctx.ViewportWidth, ctx.ViewportHeight), true
	case SvmaxUnit, LvmaxUnit, DvmaxUnit:
		return (l.Value / 100.0) * math.Max(ctx.ViewportWidth, ctx.ViewportHeight), true
	}
	return 0, false
}

// Constructors. Each L4 unit has a matching Go constructor that mirrors the
// existing Px / Em / Vw helpers for ergonomic call sites.

// Lh creates a Length in lh units (line height, current element).
func Lh(value float64) Length { return Length{Value: value, Unit: LhUnit} }

// Rlh creates a Length in rlh units (line height, root element).
func Rlh(value float64) Length { return Length{Value: value, Unit: RlhUnit} }

// Ic creates a Length in ic units (CJK ideograph advance).
func Ic(value float64) Length { return Length{Value: value, Unit: IcUnit} }

// Ric creates a Length in ric units (ic, resolved against root font size).
func Ric(value float64) Length { return Length{Value: value, Unit: RicUnit} }

// Cap creates a Length in cap units (cap height, current element).
func Cap(value float64) Length { return Length{Value: value, Unit: CapUnit} }

// Rcap creates a Length in rcap units (cap height, root element).
func Rcap(value float64) Length { return Length{Value: value, Unit: RcapUnit} }

// Rch creates a Length in rch units (ch, resolved against root font size).
func Rch(value float64) Length { return Length{Value: value, Unit: RchUnit} }

// Rex creates a Length in rex units (ex, resolved against root font size).
func Rex(value float64) Length { return Length{Value: value, Unit: RexUnit} }

// Vi creates a Length in vi units (1% of viewport inline size).
func Vi(value float64) Length { return Length{Value: value, Unit: ViUnit} }

// Vb creates a Length in vb units (1% of viewport block size).
func Vb(value float64) Length { return Length{Value: value, Unit: VbUnit} }

// Small viewport constructors.

// Svw creates a Length in svw units (small viewport width).
func Svw(value float64) Length { return Length{Value: value, Unit: SvwUnit} }

// Svh creates a Length in svh units (small viewport height).
func Svh(value float64) Length { return Length{Value: value, Unit: SvhUnit} }

// Svi creates a Length in svi units (small viewport inline).
func Svi(value float64) Length { return Length{Value: value, Unit: SviUnit} }

// Svb creates a Length in svb units (small viewport block).
func Svb(value float64) Length { return Length{Value: value, Unit: SvbUnit} }

// Svmin creates a Length in svmin units (small viewport min dimension).
func Svmin(value float64) Length { return Length{Value: value, Unit: SvminUnit} }

// Svmax creates a Length in svmax units (small viewport max dimension).
func Svmax(value float64) Length { return Length{Value: value, Unit: SvmaxUnit} }

// Large viewport constructors.

// Lvw creates a Length in lvw units (large viewport width).
func Lvw(value float64) Length { return Length{Value: value, Unit: LvwUnit} }

// Lvh creates a Length in lvh units (large viewport height).
func Lvh(value float64) Length { return Length{Value: value, Unit: LvhUnit} }

// Lvi creates a Length in lvi units (large viewport inline).
func Lvi(value float64) Length { return Length{Value: value, Unit: LviUnit} }

// Lvb creates a Length in lvb units (large viewport block).
func Lvb(value float64) Length { return Length{Value: value, Unit: LvbUnit} }

// Lvmin creates a Length in lvmin units (large viewport min dimension).
func Lvmin(value float64) Length { return Length{Value: value, Unit: LvminUnit} }

// Lvmax creates a Length in lvmax units (large viewport max dimension).
func Lvmax(value float64) Length { return Length{Value: value, Unit: LvmaxUnit} }

// Dynamic viewport constructors.

// Dvw creates a Length in dvw units (dynamic viewport width).
func Dvw(value float64) Length { return Length{Value: value, Unit: DvwUnit} }

// Dvh creates a Length in dvh units (dynamic viewport height).
func Dvh(value float64) Length { return Length{Value: value, Unit: DvhUnit} }

// Dvi creates a Length in dvi units (dynamic viewport inline).
func Dvi(value float64) Length { return Length{Value: value, Unit: DviUnit} }

// Dvb creates a Length in dvb units (dynamic viewport block).
func Dvb(value float64) Length { return Length{Value: value, Unit: DvbUnit} }

// Dvmin creates a Length in dvmin units (dynamic viewport min dimension).
func Dvmin(value float64) Length { return Length{Value: value, Unit: DvminUnit} }

// Dvmax creates a Length in dvmax units (dynamic viewport max dimension).
func Dvmax(value float64) Length { return Length{Value: value, Unit: DvmaxUnit} }
