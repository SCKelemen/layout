package layout

// LayoutPositioned handles positioned elements (absolute, relative, fixed, sticky).
// This should be called after the normal layout flow to position elements.
//
// containingBlock is the box that absolute (and, as a simplification, relative
// and sticky) offsets are measured from, expressed in the same coordinate
// space as node.Rect (the parent's local coordinates). For an absolutely
// positioned child this is the parent's padding box: the parent's border-box
// origin offset by its border widths. See containingBlockFor.
//
// viewportRect is the containing block for fixed positioning.
//
// Algorithm based on CSS Positioned Layout Module Level 3:
// - §2: Positioning Schemes
// - §3: Choosing a positioning scheme (position property)
// - §5: Absolute positioning
// - §6: Fixed positioning
// - §7: Sticky positioning
//
// and CSS 2.1 §10.3.7 / §10.6.4 (absolutely positioned, non-replaced elements).
//
// See: https://www.w3.org/TR/css-position-3/
// See: https://www.w3.org/TR/CSS21/visudet.html#abs-non-replaced-width
func LayoutPositioned(node *Node, containingBlock Rect, viewportRect Rect, ctx *LayoutContext) {
	if node.Style.Position == PositionStatic {
		// Static positioning is the default, no special handling needed
		return
	}

	// Get current font size for em resolution
	currentFontSize := getCurrentFontSize(node, ctx)

	// Calculate the positioning context
	var positioningContext Rect
	switch node.Style.Position {
	case PositionFixed:
		// Fixed is relative to viewport
		positioningContext = viewportRect
	case PositionAbsolute, PositionRelative, PositionSticky:
		// Absolute/relative/sticky are relative to nearest positioned ancestor
		// For now, we'll use the parent's padding box (in a full implementation,
		// we'd traverse up to the nearest positioned ancestor)
		positioningContext = containingBlock
	}

	// Resolve offsets to pixels. An offset is auto when it was never set
	// (see offsetIsAuto); a set value, including 0 and negative values, is a
	// real offset.
	left := node.Style.Left
	right := node.Style.Right
	top := node.Style.Top
	bottom := node.Style.Bottom

	hasLeft := !offsetIsAuto(left)
	hasRight := !offsetIsAuto(right)
	hasTop := !offsetIsAuto(top)
	hasBottom := !offsetIsAuto(bottom)

	leftPx := resolveOffset(left, ctx, currentFontSize)
	rightPx := resolveOffset(right, ctx, currentFontSize)
	topPx := resolveOffset(top, ctx, currentFontSize)
	bottomPx := resolveOffset(bottom, ctx, currentFontSize)

	switch node.Style.Position {
	case PositionAbsolute, PositionFixed:
		layoutAbsolute(node, positioningContext, ctx, currentFontSize,
			hasLeft, hasRight, hasTop, hasBottom, leftPx, rightPx, topPx, bottomPx)

	case PositionRelative, PositionSticky:
		// Relative positioning offsets the box from its normal flow position;
		// left/right and top/bottom are offsets, not constraints. When both
		// sides of an axis are set, the start side wins (CSS 2.1 §9.4.3, for
		// direction: ltr).
		// https://www.w3.org/TR/CSS21/visuren.html#relative-positioning
		//
		// Sticky positioning needs scroll information to differ from relative;
		// without a scroll context it is treated as relative (css-position-3 §7).
		if hasLeft {
			node.Rect.X += leftPx
		} else if hasRight {
			node.Rect.X -= rightPx
		}
		if hasTop {
			node.Rect.Y += topPx
		} else if hasBottom {
			node.Rect.Y -= bottomPx
		}
	}
}

// layoutAbsolute positions an absolutely or fixed positioned box inside its
// containing block (CSS 2.1 §10.3.7 and §10.6.4, css-position-3 §5).
//
// When both offsets of an axis are auto the box keeps its static position,
// which normal flow layout already stored in node.Rect (CSS 2.1 §10.3.7 rule
// for "left and right are auto": use the static position).
//
// Containing block dimensions >= Unbounded are indefinite; positions that
// would depend on them (end-edge offsets, both-edges sizing) fall back to the
// start edge so an unbounded value is never written into node.Rect.
func layoutAbsolute(node *Node, cb Rect, ctx *LayoutContext, currentFontSize float64,
	hasLeft, hasRight, hasTop, hasBottom bool, leftPx, rightPx, topPx, bottomPx float64) {
	// Ensure absolutely positioned elements have size if specified
	widthPx := ResolveLength(node.Style.Width, ctx, currentFontSize)
	heightPx := ResolveLength(node.Style.Height, ctx, currentFontSize)
	if node.Rect.Width <= 0 && widthPx > 0 {
		node.Rect.Width = widthPx
	}
	if node.Rect.Height <= 0 && heightPx > 0 {
		node.Rect.Height = heightPx
	}
	widthAuto := isUnsetLength(node.Style.Width) || widthPx < 0
	heightAuto := isUnsetLength(node.Style.Height) || heightPx < 0

	cbWidthDefinite := cb.Width < Unbounded
	cbHeightDefinite := cb.Height < Unbounded

	// Horizontal axis
	switch {
	case !hasLeft && !hasRight:
		// Both auto: static position (already in node.Rect.X from flow layout).
	case hasLeft && hasRight:
		node.Rect.X = cb.X + leftPx
		if cbWidthDefinite {
			// Both set: the box spans the space between the offsets.
			// CSS 2.1 §10.3.7: if width is auto, solve for width. If width is
			// also set (over-constrained), the box is clamped to the available
			// space rather than overflowing the containing block.
			availableWidth := cb.Width - leftPx - rightPx
			if availableWidth < 0 {
				availableWidth = 0
			}
			if widthAuto || node.Rect.Width > availableWidth {
				node.Rect.Width = availableWidth
			}
		}
	case hasLeft:
		node.Rect.X = cb.X + leftPx
	case hasRight:
		if cbWidthDefinite {
			// Right set: position from the right edge of the containing block.
			node.Rect.X = cb.X + cb.Width - node.Rect.Width - rightPx
		} else {
			node.Rect.X = cb.X
		}
	}

	// Vertical axis (CSS 2.1 §10.6.4)
	switch {
	case !hasTop && !hasBottom:
		// Both auto: static position (already in node.Rect.Y from flow layout).
	case hasTop && hasBottom:
		node.Rect.Y = cb.Y + topPx
		if cbHeightDefinite {
			availableHeight := cb.Height - topPx - bottomPx
			if availableHeight < 0 {
				availableHeight = 0
			}
			if heightAuto || node.Rect.Height > availableHeight {
				node.Rect.Height = availableHeight
			}
		}
	case hasTop:
		node.Rect.Y = cb.Y + topPx
	case hasBottom:
		if cbHeightDefinite {
			node.Rect.Y = cb.Y + cb.Height - node.Rect.Height - bottomPx
		} else {
			node.Rect.Y = cb.Y
		}
	}
}

// offsetIsAuto reports whether a positioning offset (Style.Top, Right, Bottom,
// Left) has the value "auto".
//
// An offset is auto only when it was never set, i.e. the Length is the Go zero
// value (Unit == ""). Any offset created with a constructor (Px, Em, ...) is a
// real offset: Px(0) means 0px and negative values are legitimate negative
// offsets (CSS 2.1 §9.3.2 allows <length> to be negative).
// https://www.w3.org/TR/CSS21/visuren.html#position-props
//
// The initial value of top/right/bottom/left is auto (css-position-3 §3.1), so
// an unset field behaves as in CSS.
func offsetIsAuto(l Length) bool {
	return l.Unit == ""
}

// resolveOffset resolves a positioning offset to pixels, treating an
// indefinite (unbounded) result as 0 so it never reaches a coordinate.
func resolveOffset(l Length, ctx *LayoutContext, currentFontSize float64) float64 {
	if offsetIsAuto(l) {
		return 0
	}
	v := ResolveLength(l, ctx, currentFontSize)
	if v >= Unbounded || v <= -Unbounded {
		return 0
	}
	return v
}

// containingBlockFor returns the containing block that the absolutely
// positioned children of node are positioned against, in node's local
// coordinate space (the same space node's children Rects are expressed in).
//
// CSS 2.1 §10.1 item 4: for absolutely positioned boxes, the containing block
// is formed by the padding edge of the nearest positioned ancestor, so the
// origin is offset from the border-box origin by the border widths and the
// size excludes the borders.
// https://www.w3.org/TR/CSS21/visudet.html#containing-block-details
//
// Border widths that would make the padding box negative (or an unbounded
// node size) are clamped so the result stays a definite, non-negative box.
func containingBlockFor(node *Node, ctx *LayoutContext) Rect {
	fontSize := getCurrentFontSize(node, ctx)
	borderLeft := ResolveLength(node.Style.Border.Left, ctx, fontSize)
	borderRight := ResolveLength(node.Style.Border.Right, ctx, fontSize)
	borderTop := ResolveLength(node.Style.Border.Top, ctx, fontSize)
	borderBottom := ResolveLength(node.Style.Border.Bottom, ctx, fontSize)

	width := node.Rect.Width
	height := node.Rect.Height
	if width < Unbounded {
		width -= borderLeft + borderRight
		if width < 0 {
			width = 0
		}
	}
	if height < Unbounded {
		height -= borderTop + borderBottom
		if height < 0 {
			height = 0
		}
	}
	return Rect{X: borderLeft, Y: borderTop, Width: width, Height: height}
}

// findPositionedAncestor finds the nearest positioned ancestor
// (position != static) for absolute positioning context
func findPositionedAncestor(node *Node, root *Node) *Node {
	// This is a simplified version - in a full implementation,
	// we'd need to traverse up the tree from node to root
	// For now, we'll return nil and use parent rect
	return nil
}

// LayoutWithPositioning performs layout including positioned elements
// This is a helper that handles the two-pass layout:
// 1. Normal flow layout
// 2. Positioned elements layout
func LayoutWithPositioning(root *Node, constraints Constraints, viewportRect Rect, ctx *LayoutContext) Size {
	// First pass: normal flow layout
	size := Layout(root, constraints, ctx)

	// Second pass: handle positioned elements
	layoutPositionedRecursive(root, viewportRect, ctx)

	return size
}

// layoutPositionedRecursive positions the positioned descendants of node.
//
// Children's Rects are expressed in their parent's local coordinate space
// (flow layout writes them relative to the parent's border-box origin), so
// the containing block handed to LayoutPositioned is also expressed in that
// space: the parent's padding box at (borderLeft, borderTop). Fixed children
// use the viewport rectangle instead.
func layoutPositionedRecursive(node *Node, viewportRect Rect, ctx *LayoutContext) {
	containingBlock := containingBlockFor(node, ctx)

	// Layout positioned children
	for _, child := range node.Children {
		if child.Style.Position != PositionStatic {
			// In a full implementation, we'd find the nearest positioned ancestor
			LayoutPositioned(child, containingBlock, viewportRect, ctx)
		}

		// Recursively handle children
		layoutPositionedRecursive(child, viewportRect, ctx)
	}
}
