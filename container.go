package layout

// Container query support: properties and resolution helpers for CSS L4
// container query units (cqw, cqh, cqi, cqb, cqmin, cqmax) and their
// supporting style properties (container-type, container-name, container).
//
// Spec references:
//   - CSS Containment Module Level 3 (container queries):
//     https://www.w3.org/TR/css-contain-3/#container-queries
//   - CSS Values and Units Module Level 4 (cq* units):
//     https://www.w3.org/TR/css-values-4/#container-relative-lengths
//
// This file implements the units, the three new style properties, and the
// nearest-container resolution algorithm. It does NOT implement the
// `@container` at-rule (conditional styles): that is a separate feature,
// explicitly out of scope here.

import (
	"errors"
	"fmt"
	"strings"
)

// ContainerType represents the value of the CSS `container-type` property.
// It controls whether an element establishes a query container, and which
// axes are queryable.
//
// See: https://www.w3.org/TR/css-contain-3/#container-type
type ContainerType int

const (
	// ContainerTypeNormal indicates the element does not establish a query
	// container. This is the CSS default and the zero value. Container query
	// units (cq*) resolved inside the subtree skip this ancestor.
	ContainerTypeNormal ContainerType = iota

	// ContainerTypeSize indicates the element establishes a query container
	// for both inline and block axes. All cq* units may resolve against this
	// container's measured size.
	ContainerTypeSize

	// ContainerTypeInlineSize indicates the element establishes a query
	// container for the inline axis only. cqw/cqi may resolve against this
	// container's inline size, while cqh/cqb fall through to the next
	// ancestor (or ultimately the viewport).
	ContainerTypeInlineSize
)

// String returns the canonical CSS keyword for the container type.
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

// ParseContainerType parses the CSS `container-type` value.
// Accepts (case-insensitive): "normal", "size", "inline-size".
// Whitespace is trimmed.
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

// reservedContainerNames is the set of identifiers that may not be used as
// container names. We implement the simplest correct subset of the spec:
// reject the CSS-wide and `container-type`-overlapping keywords. The full
// spec also reserves identifiers used by `@container` query syntax (e.g.
// `and`, `or`, `not`); those are not handled here because we don't
// implement the at-rule.
var reservedContainerNames = map[string]struct{}{
	"none":         {},
	"normal":       {},
	"inherit":      {},
	"initial":      {},
	"unset":        {},
	"revert":       {},
	"revert-layer": {},
	"default":      {},
}

// isValidContainerNameIdent reports whether s is a syntactically valid
// container-name identifier. We require a non-empty identifier composed of
// ASCII letters, digits, '-', or '_', not starting with a digit, and not in
// the reserved set. (CSS allows a broader Unicode ident; we keep the
// implementation pragmatic for this iteration.)
func isValidContainerNameIdent(s string) bool {
	if s == "" {
		return false
	}
	if _, reserved := reservedContainerNames[strings.ToLower(s)]; reserved {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r == '_' || r == '-':
		case r >= '0' && r <= '9':
			if i == 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// ParseContainerName parses the CSS `container-name` value. It accepts one
// or more whitespace-separated identifiers. The keyword "none" (used in CSS
// to clear the property) yields an empty slice.
//
// See: https://www.w3.org/TR/css-contain-3/#container-name
func ParseContainerName(s string) ([]string, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" || strings.EqualFold(trimmed, "none") {
		return nil, nil
	}
	fields := strings.Fields(trimmed)
	names := make([]string, 0, len(fields))
	for _, f := range fields {
		if !isValidContainerNameIdent(f) {
			return nil, fmt.Errorf("layout: invalid container-name identifier %q", f)
		}
		names = append(names, f)
	}
	return names, nil
}

// ParseContainer parses the CSS `container` shorthand:
//
//	container: <container-name>? [ / <container-type> ]?
//
// Accepted forms:
//
//	"foo"            -> name=["foo"],     type=normal
//	"foo bar"        -> name=["foo","bar"], type=normal
//	"foo / size"     -> name=["foo"],     type=size
//	"size"           -> name=nil,         type=size (bare keyword)
//	"inline-size"    -> name=nil,         type=inline-size
//	""               -> name=nil,         type=normal
//
// The "name only" form may not contain a `/`. The shorthand resets both
// properties.
func ParseContainer(s string) (names []string, ctype ContainerType, err error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil, ContainerTypeNormal, nil
	}

	// Split on `/` for the optional type segment.
	parts := strings.SplitN(trimmed, "/", 2)
	left := strings.TrimSpace(parts[0])

	// If a slash is present, left is the name list and right is the type.
	if len(parts) == 2 {
		right := strings.TrimSpace(parts[1])
		ctype, err = ParseContainerType(right)
		if err != nil {
			return nil, ContainerTypeNormal, err
		}
		names, err = ParseContainerName(left)
		if err != nil {
			return nil, ContainerTypeNormal, err
		}
		return names, ctype, nil
	}

	// No slash: a single bare type keyword sets only the type; otherwise
	// the value is treated as a container-name list.
	if ctype, terr := ParseContainerType(left); terr == nil && (left == "size" || left == "inline-size" || left == "normal") {
		return nil, ctype, nil
	}
	names, err = ParseContainerName(left)
	if err != nil {
		return nil, ContainerTypeNormal, err
	}
	return names, ContainerTypeNormal, nil
}

// ContainerAxis identifies an axis for container-query resolution.
type ContainerAxis int

const (
	// ContainerAxisInline is the inline axis (writing-mode dependent).
	// In horizontal writing modes this is the horizontal (width) axis.
	ContainerAxisInline ContainerAxis = iota

	// ContainerAxisBlock is the block axis (writing-mode dependent).
	// In horizontal writing modes this is the vertical (height) axis.
	ContainerAxisBlock
)

// resolvedContainer is the result of walking the ancestor chain for a
// container-query resolution. Found reports whether a matching ancestor was
// located that can answer the requested axis; if false the caller should
// fall back to the viewport.
type resolvedContainer struct {
	Node  *Node
	Type  ContainerType
	Size  float64 // size on the requested axis, in pixels
	Found bool
}

// physicalAxis returns the physical layout axis (true == width/horizontal,
// false == height/vertical) corresponding to a logical container axis for
// the given writing mode.
func physicalAxis(axis ContainerAxis, mode WritingMode) (horizontal bool) {
	inline := axis == ContainerAxisInline
	if mode.IsVertical() {
		// In vertical writing modes the inline axis is vertical.
		inline = !inline
	}
	return inline
}

// rectSizeOnAxis returns the measured rect size of a node on the requested
// (physical) axis.
func rectSizeOnAxis(n *Node, horizontal bool) float64 {
	if n == nil {
		return 0
	}
	if horizontal {
		return n.Rect.Width
	}
	return n.Rect.Height
}

// resolveContainerQuery walks the ancestor chain of nctx and returns the
// nearest query container that can answer the requested axis.
//
// Resolution rules (per CSS Containment L3 / CSS Values L4):
//   - The first ancestor whose computed style has ContainerType != normal is
//     a candidate.
//   - For the inline axis, ContainerType of `size` or `inline-size` both
//     qualify.
//   - For the block axis, only ContainerType `size` qualifies. A candidate
//     of `inline-size` is skipped; the walk continues looking for an
//     ancestor that supports the block axis.
//   - If no matching ancestor exists the caller falls back to the viewport
//     on the corresponding physical axis.
func resolveContainerQuery(nctx *NodeContext, axis ContainerAxis) resolvedContainer {
	if nctx == nil {
		return resolvedContainer{}
	}
	for cur := nctx.Parent(); cur != nil; cur = cur.Parent() {
		node := cur.Node
		if node == nil {
			continue
		}
		ct := node.Style.ContainerType
		if ct == ContainerTypeNormal {
			continue
		}
		// Inline axis is queryable on `size` and `inline-size`.
		// Block axis is queryable only on `size`.
		if axis == ContainerAxisBlock && ct != ContainerTypeSize {
			// Inline-size-only container does not answer block-axis
			// queries: keep walking per the L4 spec.
			continue
		}
		horizontal := physicalAxis(axis, node.Style.WritingMode)
		size := rectSizeOnAxis(node, horizontal)
		return resolvedContainer{Node: node, Type: ct, Size: size, Found: true}
	}
	return resolvedContainer{}
}

// containerQueryViewportSize returns the viewport size on the physical axis
// matching the requested logical axis for the given writing mode. Used as
// the fallback when no query container ancestor exists.
func containerQueryViewportSize(lctx *LayoutContext, axis ContainerAxis, mode WritingMode) float64 {
	if lctx == nil {
		return 0
	}
	if physicalAxis(axis, mode) {
		return lctx.ViewportWidth
	}
	return lctx.ViewportHeight
}

// ErrCQRequiresContext is returned by ResolveLength when a cq* unit is
// resolved without any node context. In that case the caller could not
// provide an ancestor chain, so resolution falls back to the viewport but
// also surfaces this sentinel for callers that want to detect the
// degradation. ResolveLength itself swallows the error and returns the
// viewport-fallback value.
var ErrCQRequiresContext = errors.New("layout: cq* unit resolved without node context; falling back to viewport")
