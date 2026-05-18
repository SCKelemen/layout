package layout

import (
	"fmt"
	"strings"

	"github.com/SCKelemen/units"
)

// ═══════════════════════════════════════════════════════════════
//  CSS Container Queries — Containment Module Level 3
//  Spec: https://www.w3.org/TR/css-contain-3/#container-queries
// ═══════════════════════════════════════════════════════════════
//
// This file implements the layout-side glue for CSS container queries:
//
//   - The `container-type` property (ContainerType + ParseContainerType).
//   - The `container-name` property (ContainerName + ParseContainerName).
//   - The `container` shorthand (ParseContainer).
//   - An ancestor-aware resolver (ResolveLengthInContext) that walks a
//     NodeContext to find the nearest qualifying query container, then
//     delegates pixel resolution to units.Length.Resolve.
//
// The cq* length unit *constants* themselves live in
// github.com/SCKelemen/units and are re-exported through Phase 1's
// aliasing. This file only adds the layout-tree-aware resolution that
// the units package cannot provide.
//
// The `@container` at-rule (size queries, style queries) is out of
// scope; this commit only adds the property model and unit resolution.

// ContainerType is the CSS `container-type` property.
//
// Per spec, three values are defined:
//
//   - normal: not a query container (default).
//   - size: establishes a query container queryable on both axes;
//     forces size containment (the element's intrinsic size becomes
//     independent of its contents).
//   - inline-size: establishes a query container queryable only on the
//     inline axis; forces inline-size containment.
//
// Spec: https://www.w3.org/TR/css-contain-3/#container-type
type ContainerType int

const (
	// ContainerTypeNormal indicates the element does not establish a
	// query container. This is the zero value and the CSS default.
	ContainerTypeNormal ContainerType = iota
	// ContainerTypeSize establishes a query container on both axes.
	ContainerTypeSize
	// ContainerTypeInlineSize establishes a query container on the
	// inline axis only.
	ContainerTypeInlineSize
)

// String returns the CSS keyword for the container-type value.
func (t ContainerType) String() string {
	switch t {
	case ContainerTypeNormal:
		return "normal"
	case ContainerTypeSize:
		return "size"
	case ContainerTypeInlineSize:
		return "inline-size"
	default:
		return "unknown"
	}
}

// ParseContainerType parses a CSS container-type keyword.
//
// Accepted (case-insensitive, leading / trailing whitespace ignored):
//
//   - "" / "normal" → ContainerTypeNormal
//   - "size"        → ContainerTypeSize
//   - "inline-size" → ContainerTypeInlineSize
//
// Any other token is rejected.
func ParseContainerType(s string) (ContainerType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "normal":
		return ContainerTypeNormal, nil
	case "size":
		return ContainerTypeSize, nil
	case "inline-size":
		return ContainerTypeInlineSize, nil
	default:
		return ContainerTypeNormal, fmt.Errorf("layout: invalid container-type %q", s)
	}
}

// ContainerName is the CSS `container-name` property: a (possibly
// empty) list of <custom-ident> tokens.
//
// Spec: https://www.w3.org/TR/css-contain-3/#container-name
type ContainerName []string

// reservedContainerNames are tokens that the CSS spec forbids as
// container names because they collide with CSS-wide keywords or with
// the `container-type` grammar.
//
// Spec: https://www.w3.org/TR/css-contain-3/#container-name — the
// production explicitly excludes the CSS-wide keywords plus "none",
// "and", "not", "or".
var reservedContainerNames = map[string]struct{}{
	"none":     {},
	"and":      {},
	"not":      {},
	"or":       {},
	"initial":  {},
	"inherit":  {},
	"unset":    {},
	"revert":   {},
	"default":  {},
	"normal":   {},
}

// ParseContainerName parses a whitespace-separated list of CSS
// container-name idents. The literal "none" yields an empty list.
//
// Rejects: reserved keywords, empty tokens within a non-empty list,
// and identifiers containing characters outside the CSS ident set
// approximated here as [A-Za-z0-9_-] (no Unicode escapes for now).
func ParseContainerName(s string) (ContainerName, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" || strings.EqualFold(trimmed, "none") {
		return nil, nil
	}
	fields := strings.Fields(trimmed)
	names := make(ContainerName, 0, len(fields))
	for _, f := range fields {
		lower := strings.ToLower(f)
		if _, bad := reservedContainerNames[lower]; bad {
			return nil, fmt.Errorf("layout: %q is reserved and cannot be used as a container-name", f)
		}
		if !isValidContainerIdent(f) {
			return nil, fmt.Errorf("layout: %q is not a valid container-name identifier", f)
		}
		names = append(names, f)
	}
	return names, nil
}

// isValidContainerIdent returns true if s is a valid simple CSS ident:
// must begin with a letter or '_' (or '-' followed by a letter/'_'),
// remaining chars must be letters, digits, '_' or '-'.
//
// This is intentionally a conservative approximation of the full CSS
// ident grammar — it doesn't admit Unicode escapes or non-ASCII letters.
// Spec (for the full grammar):
// https://www.w3.org/TR/css-syntax-3/#ident-token-diagram
func isValidContainerIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r == '_':
			// allowed anywhere
		case r >= '0' && r <= '9':
			if i == 0 {
				return false
			}
		case r == '-':
			// '-' allowed except as the only character; CSS also
			// reserves identifiers starting with two hyphens for
			// custom-property notation. Accept "-foo" but reject "-".
			if len(s) == 1 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// Has reports whether the container-name list contains the given name
// (case-sensitive — CSS idents are case-sensitive in modern specs).
func (n ContainerName) Has(name string) bool {
	for _, v := range n {
		if v == name {
			return true
		}
	}
	return false
}

// ParseContainer parses the CSS `container` shorthand.
//
// Grammar (per spec):
//
//   container: <'container-name'> [ '/' <'container-type'> ]?
//
// Examples:
//
//   "card"             → name=["card"], type=Normal
//   "card / size"      → name=["card"], type=Size
//   "card / inline-size" → name=["card"], type=InlineSize
//   "size"             → name=[],        type=Size  (no name, type only)
//   "none"             → name=[],        type=Normal
//
// Spec: https://www.w3.org/TR/css-contain-3/#container-shorthand
func ParseContainer(s string) (ContainerName, ContainerType, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil, ContainerTypeNormal, nil
	}

	// Split on '/' to separate name list from type.
	parts := strings.SplitN(trimmed, "/", 2)
	left := strings.TrimSpace(parts[0])

	if len(parts) == 1 {
		// No '/'. Try as container-type first (so `container: size` works);
		// if that fails, treat the left side as a name list.
		if ct, err := ParseContainerType(left); err == nil && ct != ContainerTypeNormal {
			return nil, ct, nil
		}
		name, err := ParseContainerName(left)
		if err != nil {
			return nil, ContainerTypeNormal, err
		}
		return name, ContainerTypeNormal, nil
	}

	name, err := ParseContainerName(left)
	if err != nil {
		return nil, ContainerTypeNormal, err
	}
	ct, err := ParseContainerType(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, ContainerTypeNormal, err
	}
	return name, ct, nil
}

// ═══════════════════════════════════════════════════════════════
//  Container-relative length resolution
// ═══════════════════════════════════════════════════════════════

// ResolveLengthInContext resolves a Length to pixels while honoring
// container query units (cqw, cqh, cqi, cqb, cqmin, cqmax). The
// supplied NodeContext is walked upward to find the nearest qualifying
// query container.
//
// Resolution rules per CSS Containment Module Level 3:
//
//   - cqw / cqi: nearest ancestor with ContainerType ∈ {Size, InlineSize}.
//   - cqh / cqb: nearest ancestor with ContainerType = Size only.
//   - cqmin / cqmax: need both axes; the inline axis follows the cqw rule
//     and the block axis follows the cqh rule. Either axis may fall
//     back to the viewport independently.
//   - When no qualifying ancestor exists, the corresponding axis falls
//     back to the viewport dimensions in ctx.
//
// Writing modes:
//
//   - cqi (inline) and cqb (block) are mapped to the physical axis using
//     the *container's* WritingMode. For horizontal-tb (the default)
//     inline = width and block = height; for any vertical writing mode
//     the mapping is swapped.
//
// For any non-container-relative unit, this function delegates to
// ResolveLength so callers can use it as a drop-in.
//
// Spec: https://www.w3.org/TR/css-contain-3/#container-lengths
func ResolveLengthInContext(l Length, ctx *LayoutContext, currentFontSize float64, nctx *NodeContext) float64 {
	// Layout-specific sentinel: identical handling to ResolveLength.
	if l.Unit == UnboundedUnit {
		return ResolveLength(l, ctx, currentFontSize)
	}
	if !l.IsContainerRelative() {
		return ResolveLength(l, ctx, currentFontSize)
	}

	uctx := buildUnitsContext(ctx, currentFontSize)
	resolveUnit := l.Unit

	// Inline-axis container (accepts size OR inline-size).
	inlineCtr := findQueryContainer(nctx, false)
	// Block-axis container (accepts only size).
	sizeCtr := findQueryContainer(nctx, true)

	switch l.Unit {
	case units.CQW:
		uctx.ContainerWidth = inlineAxisFallback(inlineCtr, ctx, axisWidth)

	case units.CQH:
		uctx.ContainerHeight = blockAxisFallback(sizeCtr, ctx, axisHeight)

	case units.CQI:
		// Translate cqi to cqw/cqh based on the inline container's
		// writing mode. Vertical writing modes swap inline ↔ block
		// physical axes.
		wm := containerWritingMode(inlineCtr)
		if wm.IsVertical() {
			uctx.ContainerHeight = inlineAxisFallback(inlineCtr, ctx, axisHeight)
			resolveUnit = units.CQH
		} else {
			uctx.ContainerWidth = inlineAxisFallback(inlineCtr, ctx, axisWidth)
			resolveUnit = units.CQW
		}

	case units.CQB:
		wm := containerWritingMode(sizeCtr)
		if wm.IsVertical() {
			uctx.ContainerWidth = blockAxisFallback(sizeCtr, ctx, axisWidth)
			resolveUnit = units.CQW
		} else {
			uctx.ContainerHeight = blockAxisFallback(sizeCtr, ctx, axisHeight)
			resolveUnit = units.CQH
		}

	case units.CQMIN, units.CQMAX:
		// Both axes needed. Inline uses the cqw rule, block uses the
		// cqh rule, each independently falling back to the viewport.
		uctx.ContainerWidth = inlineAxisFallback(inlineCtr, ctx, axisWidth)
		uctx.ContainerHeight = blockAxisFallback(sizeCtr, ctx, axisHeight)
	}

	translated := Length{Value: l.Value, Unit: resolveUnit}
	resolved, err := translated.Resolve(uctx)
	if err != nil {
		if l.IsContainerRelative() {
			// No container size available; cq* resolves to 0.
			return 0
		}
		// Non-container fallback (preserves pre-fix behavior for absolute/viewport units).
		return l.Value
	}
	return resolved.Value
}

// findQueryContainer walks the ancestor chain of nctx looking for the
// nearest node whose Style.ContainerType qualifies for the requested
// axis.
//
//   - sizeOnly=true: only ContainerTypeSize qualifies. ContainerTypeInlineSize
//     ancestors are skipped (cqh / cqb / cqmin block-axis / cqmax block-axis).
//   - sizeOnly=false: ContainerTypeSize *or* ContainerTypeInlineSize qualifies
//     (cqw / cqi / cqmin inline-axis / cqmax inline-axis).
//
// Returns nil if no qualifying ancestor exists; the caller is expected
// to fall back to the viewport in that case.
func findQueryContainer(nctx *NodeContext, sizeOnly bool) *Node {
	if nctx == nil {
		return nil
	}
	for cur := nctx.Parent(); cur != nil; cur = cur.Parent() {
		node := cur.Unwrap()
		if node == nil {
			continue
		}
		switch node.Style.ContainerType {
		case ContainerTypeSize:
			return node
		case ContainerTypeInlineSize:
			if !sizeOnly {
				return node
			}
		}
	}
	return nil
}

// containerWritingMode returns the container's WritingMode, defaulting
// to WritingModeHorizontalTB (the CSS default and zero value) when no
// container is supplied. Used to map logical cqi / cqb units to a
// physical axis.
func containerWritingMode(c *Node) WritingMode {
	if c == nil {
		return WritingModeHorizontalTB
	}
	return c.Style.WritingMode
}

// physicalAxis identifies one of the two physical layout axes.
type physicalAxis int

const (
	axisWidth physicalAxis = iota
	axisHeight
)

// inlineAxisFallback returns the physical pixel value to plug into
// units.Context for the inline-axis container dimension. Uses the
// container's Rect if available; otherwise falls back to the viewport
// dimension on the requested physical axis.
func inlineAxisFallback(c *Node, ctx *LayoutContext, axis physicalAxis) float64 {
	if c != nil {
		return rectAxis(c.Rect, axis)
	}
	return viewportAxis(ctx, axis)
}

// blockAxisFallback is the block-axis counterpart of inlineAxisFallback.
// Kept separate for readability at the call sites; the two helpers have
// the same body but they document very different intents (one is
// triggered by inline-axis cq* units, the other by block-axis ones).
func blockAxisFallback(c *Node, ctx *LayoutContext, axis physicalAxis) float64 {
	if c != nil {
		return rectAxis(c.Rect, axis)
	}
	return viewportAxis(ctx, axis)
}

func rectAxis(r Rect, axis physicalAxis) float64 {
	if axis == axisWidth {
		return r.Width
	}
	return r.Height
}

func viewportAxis(ctx *LayoutContext, axis physicalAxis) float64 {
	if ctx == nil {
		return 0
	}
	if axis == axisWidth {
		return ctx.ViewportWidth
	}
	return ctx.ViewportHeight
}
