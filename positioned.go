package layout

// LayoutPositioned handles positioned elements (absolute, relative, fixed, sticky).
// This should be called after the normal layout flow to position elements.
//
// containingBlock is the box that absolute offsets are measured from,
// expressed in the same coordinate space as node.Rect (the parent's local
// coordinates). Per CSS 2.1 §10.1 it is the padding box of the nearest
// positioned ancestor (or of the root when no ancestor is positioned);
// LayoutWithPositioning computes it via layoutPositionedRecursive. Relative
// and sticky boxes only use their offsets and ignore containingBlock.
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
		// Absolute boxes are positioned against the containing block the
		// caller resolved (nearest positioned ancestor, CSS 2.1 §10.1).
		// Relative and sticky boxes only apply offsets to their flow position
		// and never read positioningContext.
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
// An explicit Width/Height is the box's used size regardless of the flow
// layout pass: an absolutely positioned box is sized against its containing
// block, not against its parent's available space, so a width larger than the
// parent is kept (CSS 2.1 §10.3.7 uses the specified width directly). When
// left, width, and right are all set the box is over-constrained and the end
// offset is ignored: right for direction: ltr, left for direction: rtl
// (§10.3.7); bottom is ignored in the vertical axis (§10.6.4).
//
// Containing block dimensions >= Unbounded are indefinite; positions that
// would depend on them (end-edge offsets, both-edges sizing) fall back to the
// start edge so an unbounded value is never written into node.Rect.
func layoutAbsolute(node *Node, cb Rect, ctx *LayoutContext, currentFontSize float64,
	hasLeft, hasRight, hasTop, hasBottom bool, leftPx, rightPx, topPx, bottomPx float64) {
	explicitWidth, explicitHeight, widthSet, heightSet := absoluteExplicitSize(node, ctx, currentFontSize)
	if widthSet {
		node.Rect.Width = explicitWidth
	}
	if heightSet {
		node.Rect.Height = explicitHeight
	}
	widthAuto := !widthSet
	heightAuto := !heightSet

	cbWidthDefinite := cb.Width < Unbounded
	cbHeightDefinite := cb.Height < Unbounded

	// Horizontal axis (CSS 2.1 §10.3.7)
	switch {
	case !hasLeft && !hasRight:
		// Both auto: static position (already in node.Rect.X from flow layout).
	case hasLeft && hasRight && widthAuto:
		// Both offsets set, width auto: the box spans the space between them.
		node.Rect.X = cb.X + leftPx
		if cbWidthDefinite {
			node.Rect.Width = max(cb.Width-leftPx-rightPx, 0)
		}
	case hasLeft && hasRight && node.Style.Direction == DirectionRTL && cbWidthDefinite:
		// Over-constrained, rtl: ignore left and position from the right edge.
		node.Rect.X = cb.X + cb.Width - node.Rect.Width - rightPx
	case hasLeft:
		// Left only, or over-constrained in ltr (right is ignored).
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
		// Both offsets set: solve for an auto height; with an explicit height
		// the box is over-constrained and bottom is ignored.
		node.Rect.Y = cb.Y + topPx
		if heightAuto && cbHeightDefinite {
			node.Rect.Height = max(cb.Height-topPx-bottomPx, 0)
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

// absoluteExplicitSize returns the border-box width and height that an
// absolutely positioned box takes from its explicit Style.Width/Height, and
// whether each axis is explicit at all.
//
// A dimension is explicit when it is set (Unit != ""), non-negative (negative
// values and the intrinsic-size sentinels are not lengths), and finite. The
// value is converted to a border-box size according to box-sizing (CSS Box
// Sizing Level 3 §3, https://www.w3.org/TR/css-sizing-3/#box-sizing) and
// clamped by min/max in the same box (CSS 2.1 §10.4, §10.7).
func absoluteExplicitSize(node *Node, ctx *LayoutContext, fontSize float64) (width, height float64, widthSet, heightSet bool) {
	hpb := getHorizontalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, fontSize)
	vpb := getVerticalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, fontSize)

	resolveExplicit := func(l Length) (float64, bool) {
		if isUnsetLength(l) {
			return 0, false
		}
		v := ResolveLength(l, ctx, fontSize)
		if v < 0 || v >= Unbounded {
			return 0, false
		}
		return v, true
	}

	if w, ok := resolveExplicit(node.Style.Width); ok {
		content := convertToContentSize(w, node.Style.BoxSizing, hpb, vpb, true)
		minW := convertMinMaxToContentSize(ResolveLength(node.Style.MinWidth, ctx, fontSize), node.Style.BoxSizing, hpb, vpb, true)
		maxW := convertMinMaxToContentSize(ResolveLength(node.Style.MaxWidth, ctx, fontSize), node.Style.BoxSizing, hpb, vpb, true)
		width = clampMinMax(content, minW, maxW) + hpb
		widthSet = true
	}
	if h, ok := resolveExplicit(node.Style.Height); ok {
		content := convertToContentSize(h, node.Style.BoxSizing, hpb, vpb, false)
		minH := convertMinMaxToContentSize(ResolveLength(node.Style.MinHeight, ctx, fontSize), node.Style.BoxSizing, hpb, vpb, false)
		maxH := convertMinMaxToContentSize(ResolveLength(node.Style.MaxHeight, ctx, fontSize), node.Style.BoxSizing, hpb, vpb, false)
		height = clampMinMax(content, minH, maxH) + vpb
		heightSet = true
	}
	return width, height, widthSet, heightSet
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

// LayoutWithPositioning performs layout including positioned elements
// This is a helper that handles the two-pass layout:
// 1. Normal flow layout
// 2. Positioned elements layout
//
// The root acts as the initial containing block: absolutely positioned
// descendants with no positioned ancestor are positioned against the root's
// padding box (CSS 2.1 §10.1 item 1, with the root standing in for the
// viewport-sized initial containing block). Fixed descendants use
// viewportRect.
func LayoutWithPositioning(root *Node, constraints Constraints, viewportRect Rect, ctx *LayoutContext) Size {
	if root == nil {
		return Size{}
	}
	// First pass: normal flow layout
	size := Layout(root, constraints, ctx)

	// Second pass: handle positioned elements
	layoutPositionedRecursive(root, containingBlockFor(root, ctx), viewportRect, ctx)

	return size
}

// layoutPositionedRecursive positions the positioned descendants of node.
//
// containingBlock is the containing block for the absolutely positioned
// children of node: the padding box of the nearest positioned ancestor (or of
// the root), expressed in node's local coordinate space, which is the space
// the children's Rects are written in (CSS 2.1 §10.1 item 4,
// https://www.w3.org/TR/CSS21/visudet.html#containing-block-details).
//
// When descending into a child, the containing block for the grandchildren is
// the child's own padding box if the child is positioned (relative, absolute,
// fixed, or sticky; §10.1 says "positioned", i.e. any value other than
// static), and otherwise the same containing block translated into the
// child's local space (the child's final position is subtracted, after any
// relative offset has been applied). Fixed children use the viewport
// rectangle instead (§10.1 item 3).
//
// Recursion depth is bounded by the tree depth; each node is visited once.
func layoutPositionedRecursive(node *Node, containingBlock Rect, viewportRect Rect, ctx *LayoutContext) {
	for _, child := range node.Children {
		if child == nil {
			continue
		}
		if child.Style.Position != PositionStatic {
			LayoutPositioned(child, containingBlock, viewportRect, ctx)
		}

		var childContainingBlock Rect
		if child.Style.Position != PositionStatic {
			childContainingBlock = containingBlockFor(child, ctx)
		} else {
			childContainingBlock = containingBlock
			childContainingBlock.X -= child.Rect.X
			childContainingBlock.Y -= child.Rect.Y
		}
		layoutPositionedRecursive(child, childContainingBlock, viewportRect, ctx)
	}
}
