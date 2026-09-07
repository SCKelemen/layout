package layout

// blockLayoutChildren lays out children in block flow direction with margin collapsing.
//
// Algorithm based on CSS Box Model Module Level 3 and CSS Writing Modes Level 3:
// - §8.3.1: Collapsing margins
// - Writing Modes: block direction depends on writing-mode
//
// See: https://www.w3.org/TR/css-box-3/#collapsing-margins
// See: https://www.w3.org/TR/CSS21/box.html#collapsing-margins
// See: https://www.w3.org/TR/css-writing-modes-3/
//
// Block direction (children stacking):
// - Horizontal modes: children stack vertically (Y increases)
// - Vertical modes: children stack horizontally (X increases or decreases)
//
// Margin collapsing rules (CSS 2.1 §8.3.1):
//  1. Adjoining block-axis margins collapse. The collapsed margin is the sum of
//     the largest positive margin and the most negative margin (collapseMargins).
//  2. Parent and first child start margins collapse if no border/padding/content
//     separates them (not implemented: the first child's start margin is kept
//     inside the parent).
//  3. Parent and last child end margins collapse (not implemented, see 2).
//  4. A box with zero block size and no block-axis padding/border collapses its
//     own start and end margins together with the adjoining sibling margins.
//
// Out-of-flow children (position: absolute / fixed) are laid out to determine
// their own size and static position, but they do not take up flow space and
// do not participate in margin collapsing (CSS 2.1 §9.6, css-position-3 §3.4).
//
// In vertical right-to-left writing modes, children are positioned as if the
// block axis ran left-to-right; LayoutBlock mirrors them once the final block
// size (width) is known (see blockMirrorChildrenRTL).
//
// nodeWidth is the physical content width, nodeHeight the physical content
// height of the container; the inline-axis constraint for children is the
// width in horizontal modes and the height in vertical modes.
func blockLayoutChildren(node *Node, setup blockSetup, nodeWidth, nodeHeight float64, ctx *LayoutContext, parentFontSize float64) (currentBlockPos, maxCrossSize float64) {
	children := node.Children
	writingMode := node.Style.WritingMode
	isVertical := writingMode.IsVertical()

	// Set child constraints based on writing mode
	// Horizontal mode: constrain width (inline), unbounded height (block)
	// Vertical mode: unbounded width (block), constrain height (inline)
	var childConstraints Constraints
	if isVertical {
		// In vertical writing modes the inline size is the physical height.
		// CSS Writing Modes Level 3 §7.1: https://www.w3.org/TR/css-writing-modes-3/#logical-to-physical
		inlineSize := nodeHeight
		if inlineSize < 0 {
			inlineSize = setup.contentHeight
		}
		childConstraints = Constraints{
			MinWidth:  0,
			MaxWidth:  Unbounded,
			MinHeight: 0,
			MaxHeight: inlineSize,
		}
	} else {
		childConstraints = Constraints{
			MinWidth:  0,
			MaxWidth:  nodeWidth,
			MinHeight: 0,
			MaxHeight: Unbounded,
		}
	}

	// Resolve parent's padding and border for positioning
	parentPaddingLeft := ResolveLength(node.Style.Padding.Left, ctx, parentFontSize)
	parentPaddingTop := ResolveLength(node.Style.Padding.Top, ctx, parentFontSize)
	parentBorderLeft := ResolveLength(node.Style.Border.Left, ctx, parentFontSize)
	parentBorderTop := ResolveLength(node.Style.Border.Top, ctx, parentFontSize)
	contentOriginX := parentPaddingLeft + parentBorderLeft
	contentOriginY := parentPaddingTop + parentBorderTop

	// flowEnd is the block-axis position of the block-end border edge of the
	// last in-flow child with a non-zero block size (0 before any such child).
	// pendingMargins collects every margin adjoining the gap after flowEnd:
	// the previous child's end margin plus the start (and, for self-collapsing
	// boxes, end) margins of subsequent children. They collapse into one value.
	flowEnd := 0.0
	var pendingMargins []float64
	maxCrossSize = 0.0

	for _, child := range children {
		// Skip display:none children
		if child.Style.Display == DisplayNone {
			continue
		}

		// Get child's font size for margin resolution
		childFontSize := getCurrentFontSize(child, ctx)

		// Resolve child's margins to pixels
		childMarginTop := ResolveLength(child.Style.Margin.Top, ctx, childFontSize)
		childMarginBottom := ResolveLength(child.Style.Margin.Bottom, ctx, childFontSize)
		childMarginLeft := ResolveLength(child.Style.Margin.Left, ctx, childFontSize)
		childMarginRight := ResolveLength(child.Style.Margin.Right, ctx, childFontSize)

		// Map margins to logical directions (start/end in block axis)
		var childMarginBlockStart, childMarginBlockEnd, childMarginInlineStart, childMarginInlineEnd float64
		if isVertical {
			// Vertical mode: block axis is horizontal
			// Direction depends on whether blocks progress left-to-right or right-to-left
			if writingMode.IsRightToLeft() {
				// vertical-rl: blocks progress right-to-left
				childMarginBlockStart = childMarginRight // Start = right for vertical-rl
				childMarginBlockEnd = childMarginLeft    // End = left for vertical-rl
			} else {
				// vertical-lr: blocks progress left-to-right
				childMarginBlockStart = childMarginLeft // Start = left for vertical-lr
				childMarginBlockEnd = childMarginRight  // End = right for vertical-lr
			}
			childMarginInlineStart = childMarginTop  // Inline start = top
			childMarginInlineEnd = childMarginBottom // Inline end = bottom
		} else {
			// Horizontal mode: block axis is vertical
			childMarginBlockStart = childMarginTop   // Start = top
			childMarginBlockEnd = childMarginBottom  // End = bottom
			childMarginInlineStart = childMarginLeft // Inline start = left
			childMarginInlineEnd = childMarginRight  // Inline end = right
		}

		// Layout child
		childSize := layoutBlockChild(child, childConstraints, ctx)

		// Get child block size for positioning
		var childBlockSize float64
		if isVertical {
			childBlockSize = childSize.Width
		} else {
			childBlockSize = childSize.Height
		}

		outOfFlow := isOutOfFlow(child)

		// Block-axis position of the child's block-start border edge.
		// Out-of-flow children get their static position (CSS 2.1 §10.3.7 /
		// §10.6.4: where the box would have been in normal flow) without
		// contributing their margins to the collapse set.
		var childBlockPos float64
		if outOfFlow {
			// Its own start margin adjoins the pending set only for the purpose
			// of computing the static position; pendingMargins is left untouched.
			staticSet := make([]float64, 0, len(pendingMargins)+1)
			staticSet = append(staticSet, pendingMargins...)
			staticSet = append(staticSet, childMarginBlockStart)
			childBlockPos = flowEnd + collapseMargins(staticSet...)
		} else {
			pendingMargins = append(pendingMargins, childMarginBlockStart)
			// CSS 2.1 §8.3.1: a block with zero block size and no block-axis
			// padding/border has adjoining start and end margins, which collapse
			// through it. Only block containers self-collapse; flex/grid/text
			// boxes establish independent formatting contexts.
			selfCollapsing := childBlockSize == 0 && isBlockContainer(child)
			if selfCollapsing {
				pendingMargins = append(pendingMargins, childMarginBlockEnd)
			}
			childBlockPos = flowEnd + collapseMargins(pendingMargins...)
			if !selfCollapsing {
				flowEnd = childBlockPos + childBlockSize
				pendingMargins = append(pendingMargins[:0], childMarginBlockEnd)
			}
		}

		// Position child with padding, border, and margin offset
		// Children are positioned in the content area, which starts after padding + border
		var childX, childY float64
		if isVertical {
			// Vertical mode: block direction is X. Right-to-left modes are
			// positioned left-to-right here and mirrored by blockMirrorChildrenRTL.
			childX = contentOriginX + childBlockPos
			childY = contentOriginY + childMarginInlineStart
		} else {
			// Horizontal mode: block direction is Y
			childX = contentOriginX + childMarginInlineStart
			childY = contentOriginY + childBlockPos
		}

		child.Rect = Rect{
			X:      childX,
			Y:      childY,
			Width:  childSize.Width,
			Height: childSize.Height,
		}

		if outOfFlow {
			// Absolutely positioned boxes take no space in the flow.
			continue
		}

		// Track max cross-axis size (including margins)
		var childCrossSizeWithMargins float64
		if isVertical {
			childCrossSizeWithMargins = childSize.Height + childMarginInlineStart + childMarginInlineEnd
		} else {
			childCrossSizeWithMargins = childSize.Width + childMarginInlineStart + childMarginInlineEnd
		}
		if childCrossSizeWithMargins > maxCrossSize {
			maxCrossSize = childCrossSizeWithMargins
		}
	}

	// The trailing collapsed margin (last child's end margin, plus any
	// self-collapsing boxes after it) still occupies space inside the parent.
	currentBlockPos = flowEnd + collapseMargins(pendingMargins...)

	return currentBlockPos, maxCrossSize
}

// layoutBlockChild dispatches a child of a block container to the layout
// algorithm for its display type.
func layoutBlockChild(child *Node, constraints Constraints, ctx *LayoutContext) Size {
	switch child.Style.Display {
	case DisplayFlex:
		return LayoutFlexbox(child, constraints, ctx)
	case DisplayGrid:
		return LayoutGrid(child, constraints, ctx)
	case DisplayInlineText:
		return LayoutText(child, constraints, ctx)
	default:
		return LayoutBlock(child, constraints, ctx)
	}
}

// isOutOfFlow reports whether a box is taken out of the normal flow.
// CSS 2.1 §9.3: absolutely positioned boxes (position: absolute or fixed)
// are removed from the normal flow entirely.
// https://www.w3.org/TR/CSS21/visuren.html#choose-position
func isOutOfFlow(node *Node) bool {
	return node.Style.Position == PositionAbsolute || node.Style.Position == PositionFixed
}

// isBlockContainer reports whether a node is laid out by the block algorithm
// (any display type that is not flex, grid, or text).
func isBlockContainer(node *Node) bool {
	switch node.Style.Display {
	case DisplayFlex, DisplayGrid, DisplayInlineText, DisplayNone:
		return false
	default:
		return true
	}
}

// collapseMargins returns the collapsed value of a set of adjoining margins.
//
// CSS 2.1 §8.3.1: "When two or more margins collapse, the resulting margin
// width is the maximum of the collapsing margins' widths. In the case of
// negative margins, the maximum of the absolute values of the negative
// adjoining margins is deducted from the maximum of the positive adjoining
// margins. If there are no positive margins, the maximum of the absolute
// values of the adjoining margins is deducted from zero."
// https://www.w3.org/TR/CSS21/box.html#collapsing-margins
func collapseMargins(margins ...float64) float64 {
	maxPositive := 0.0
	minNegative := 0.0
	for _, m := range margins {
		if m >= Unbounded || m <= -Unbounded {
			// Indefinite margins never contribute to a position.
			continue
		}
		if m > maxPositive {
			maxPositive = m
		}
		if m < minNegative {
			minNegative = m
		}
	}
	return maxPositive + minNegative
}

// blockMirrorChildrenRTL converts the left-to-right block positions assigned
// by blockLayoutChildren into right-to-left positions for vertical-rl and
// sideways-rl writing modes, using the container's final content width.
//
// CSS Writing Modes Level 3 §7.3: in vertical-rl the block-start edge is the
// right edge, so the first child sits against the right side of the content
// box. https://www.w3.org/TR/css-writing-modes-3/#vertical-layout
func blockMirrorChildrenRTL(node *Node, nodeWidth float64, ctx *LayoutContext, parentFontSize float64) {
	if nodeWidth >= Unbounded || nodeWidth < 0 {
		// Indefinite block size: leave children in their LTR positions rather
		// than propagate an unbounded coordinate.
		return
	}
	contentOriginX := ResolveLength(node.Style.Padding.Left, ctx, parentFontSize) +
		ResolveLength(node.Style.Border.Left, ctx, parentFontSize)
	for _, child := range node.Children {
		if child.Style.Display == DisplayNone {
			continue
		}
		offset := child.Rect.X - contentOriginX
		child.Rect.X = contentOriginX + nodeWidth - offset - child.Rect.Width
	}
}
