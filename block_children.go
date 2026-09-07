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
//  2. The start margin of the container and the start margin of its first
//     in-flow child collapse when the container has no block-start padding or
//     border and collapseThrough is set (the container is itself an in-flow
//     block child, not a formatting-context root). The collapsed-through
//     margins are not applied inside the container; they are returned in
//     through.start so the container's parent collapses them with the
//     container's own start margin.
//  3. The end margin of the last in-flow child collapses with the container's
//     end margin when, in addition, the container has an auto block size, no
//     block-end padding or border, and no min block size. The margins are
//     returned in through.end.
//  4. A box with zero block size and no block-axis padding/border collapses its
//     own start and end margins together with the adjoining sibling margins.
//     When every in-flow child is such a box, all margins adjoin both edges.
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
//
// currentBlockPos is the block-axis extent of the in-flow content, excluding
// any margins that collapsed through the container's edges.
func blockLayoutChildren(node *Node, setup blockSetup, nodeWidth, nodeHeight float64, ctx *LayoutContext, parentFontSize float64, collapseThrough bool) (currentBlockPos, maxCrossSize float64, through collapsedThroughMargins) {
	children := node.Children
	writingMode := node.Style.WritingMode
	isVertical := writingMode.IsVertical()

	// Whether margins may collapse through the container's block-start and
	// block-end edges (rules 2 and 3 above).
	collapseStart, collapseEnd := blockCollapseThroughEdges(node, setup, ctx, parentFontSize, collapseThrough)

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
	// hasFlowContent records whether such a child has been placed yet; until
	// then the pending margins adjoin the container's start edge.
	flowEnd := 0.0
	var pendingMargins []float64
	hasFlowContent := false
	maxCrossSize = 0.0

	for _, child := range children {
		// Skip display:none children
		if child == nil || child.Style.Display == DisplayNone {
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

		// Layout child. childThrough holds the grandchild margins that collapsed
		// through the child's edges; they adjoin the child's own margins.
		childSize, childThrough := layoutBlockChild(child, childConstraints, ctx)

		// Get child block size for positioning
		var childBlockSize float64
		if isVertical {
			childBlockSize = childSize.Width
		} else {
			childBlockSize = childSize.Height
		}

		outOfFlow := isOutOfFlow(child)

		// While no in-flow content has been placed and the container collapses
		// through its start edge, every pending margin belongs outside the
		// container: children are placed at the content origin instead.
		atCollapsedStart := collapseStart && !hasFlowContent

		// Block-axis position of the child's block-start border edge.
		// Out-of-flow children get their static position (CSS 2.1 §10.3.7 /
		// §10.6.4: where the box would have been in normal flow) without
		// contributing their margins to the collapse set.
		var childBlockPos float64
		if outOfFlow {
			if atCollapsedStart {
				childBlockPos = flowEnd
			} else {
				// Its own start margin adjoins the pending set only for the
				// purpose of computing the static position; pendingMargins is
				// left untouched.
				staticSet := make([]float64, 0, len(pendingMargins)+1)
				staticSet = append(staticSet, pendingMargins...)
				staticSet = append(staticSet, childMarginBlockStart)
				childBlockPos = flowEnd + collapseMargins(staticSet...)
			}
		} else {
			pendingMargins = append(pendingMargins, childMarginBlockStart)
			pendingMargins = append(pendingMargins, childThrough.start...)
			// CSS 2.1 §8.3.1: a block with zero block size and no block-axis
			// padding/border has adjoining start and end margins, which collapse
			// through it. Only block containers self-collapse; flex/grid/text
			// boxes establish independent formatting contexts.
			selfCollapsing := childBlockSize == 0 && isBlockContainer(child)
			if selfCollapsing {
				pendingMargins = append(pendingMargins, childMarginBlockEnd)
				pendingMargins = append(pendingMargins, childThrough.end...)
			}
			if atCollapsedStart {
				childBlockPos = flowEnd
			} else {
				childBlockPos = flowEnd + collapseMargins(pendingMargins...)
			}
			if !selfCollapsing {
				if atCollapsedStart {
					// Rule 2: the margins before the first in-flow child belong
					// to the container's parent.
					through.start = append([]float64(nil), pendingMargins...)
				}
				hasFlowContent = true
				flowEnd = childBlockPos + childBlockSize
				pendingMargins = append(pendingMargins[:0], childMarginBlockEnd)
				pendingMargins = append(pendingMargins, childThrough.end...)
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

	// The trailing margins: the last child's end margin plus any
	// self-collapsing boxes after it.
	switch {
	case !hasFlowContent && collapseStart:
		// No in-flow content at all: every margin adjoins the start edge (and
		// the end edge too when that collapses), so the container is itself
		// self-collapsing as far as its parent is concerned.
		through.start = append([]float64(nil), pendingMargins...)
		if collapseEnd {
			through.end = append([]float64(nil), pendingMargins...)
		}
		currentBlockPos = flowEnd
	case collapseEnd:
		// Rule 3: the trailing margins belong to the container's parent.
		through.end = append([]float64(nil), pendingMargins...)
		currentBlockPos = flowEnd
	default:
		// The trailing collapsed margin still occupies space inside the parent.
		currentBlockPos = flowEnd + collapseMargins(pendingMargins...)
	}

	return currentBlockPos, maxCrossSize, through
}

// blockCollapseThroughEdges decides whether margins may collapse through the
// block-start and block-end edges of a block container (CSS 2.1 §8.3.1).
//
// Start: the container is an in-flow block child (collapseThrough) with no
// block-start padding or border. End: additionally the container's block size
// is auto and its min block size is 0, and it has no block-end padding or
// border. The block axis follows the writing mode (css-writing-modes-3 §7.1):
// the start edge is top for horizontal-tb, left for vertical-lr, and right
// for vertical-rl / sideways-rl.
// https://www.w3.org/TR/CSS21/box.html#collapsing-margins
func blockCollapseThroughEdges(node *Node, setup blockSetup, ctx *LayoutContext, fontSize float64, collapseThrough bool) (start, end bool) {
	if !collapseThrough || !isBlockContainer(node) {
		return false, false
	}
	padding := node.Style.Padding
	border := node.Style.Border
	var startEdge, endEdge float64
	var autoBlockSize, zeroMinBlockSize bool
	switch {
	case node.Style.WritingMode.IsVertical() && node.Style.WritingMode.IsRightToLeft():
		startEdge = ResolveLength(padding.Right, ctx, fontSize) + ResolveLength(border.Right, ctx, fontSize)
		endEdge = ResolveLength(padding.Left, ctx, fontSize) + ResolveLength(border.Left, ctx, fontSize)
		autoBlockSize = setup.isAutoWidth
		zeroMinBlockSize = setup.minWidthContent <= 0
	case node.Style.WritingMode.IsVertical():
		startEdge = ResolveLength(padding.Left, ctx, fontSize) + ResolveLength(border.Left, ctx, fontSize)
		endEdge = ResolveLength(padding.Right, ctx, fontSize) + ResolveLength(border.Right, ctx, fontSize)
		autoBlockSize = setup.isAutoWidth
		zeroMinBlockSize = setup.minWidthContent <= 0
	default:
		startEdge = ResolveLength(padding.Top, ctx, fontSize) + ResolveLength(border.Top, ctx, fontSize)
		endEdge = ResolveLength(padding.Bottom, ctx, fontSize) + ResolveLength(border.Bottom, ctx, fontSize)
		autoBlockSize = setup.isAutoHeight
		zeroMinBlockSize = setup.minHeightContent <= 0
	}
	start = startEdge == 0
	end = endEdge == 0 && autoBlockSize && zeroMinBlockSize
	return start, end
}

// layoutBlockChild dispatches a child of a block container to the layout
// algorithm for its display type. Only in-flow block containers may collapse
// margins through their edges; flex, grid, and text boxes establish
// independent formatting contexts, and out-of-flow boxes do not collapse with
// anything (CSS 2.1 §8.3.1, §9.4.1).
func layoutBlockChild(child *Node, constraints Constraints, ctx *LayoutContext) (Size, collapsedThroughMargins) {
	switch child.Style.Display {
	case DisplayFlex:
		return LayoutFlexbox(child, constraints, ctx), collapsedThroughMargins{}
	case DisplayGrid:
		return LayoutGrid(child, constraints, ctx), collapsedThroughMargins{}
	case DisplayInlineText:
		return LayoutText(child, constraints, ctx), collapsedThroughMargins{}
	default:
		return layoutBlockFlow(child, constraints, ctx, !isOutOfFlow(child))
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
