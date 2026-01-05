package layout

// blockLayoutChildren lays out children in block flow direction with margin collapsing.
//
// Algorithm based on CSS Box Model Module Level 3 and CSS Writing Modes Level 3:
// - §8.3.1: Collapsing margins
// - Writing Modes: block direction depends on writing-mode
//
// See: https://www.w3.org/TR/css-box-3/#collapsing-margins
// See: https://www.w3.org/TR/css-writing-modes-3/
//
// Block direction (children stacking):
// - Horizontal modes: children stack vertically (Y increases)
// - Vertical modes: children stack horizontally (X increases or decreases)
//
// Margin collapsing rules (simplified):
// 1. Adjacent block-axis margins collapse (use max, not sum)
// 2. Parent and first child start margins collapse if no border/padding/content separates them
// 3. Parent and last child end margins collapse if no border/padding/size separates them
// 4. Empty block margins collapse with themselves
//
// For this implementation, we focus on the most common case:
// - Adjacent sibling margins collapse (rule 1)
func blockLayoutChildren(node *Node, setup blockSetup, nodeWidth float64, ctx *LayoutContext, parentFontSize float64) (currentBlockPos, maxCrossSize float64) {
	children := node.Children
	writingMode := node.Style.WritingMode
	isVertical := writingMode.IsVertical()

	currentBlockPos = 0.0
	maxCrossSize = 0.0

	// Set child constraints based on writing mode
	// Horizontal mode: constrain width (inline), unbounded height (block)
	// Vertical mode: unbounded width (block), constrain height (inline)
	var childConstraints Constraints
	if isVertical {
		childConstraints = Constraints{
			MinWidth:  0,
			MaxWidth:  Unbounded,
			MinHeight: 0,
			MaxHeight: nodeWidth, // nodeWidth is actually the inline size in vertical mode
		}
	} else {
		childConstraints = Constraints{
			MinWidth:  0,
			MaxWidth:  nodeWidth,
			MinHeight: 0,
			MaxHeight: Unbounded,
		}
	}

	// Track previous child's end margin (bottom for horizontal, right/left for vertical) for collapsing
	var prevEndMargin float64 = 0.0

	for i, child := range children {
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
				childMarginBlockStart = childMarginRight  // Start = right for vertical-rl
				childMarginBlockEnd = childMarginLeft     // End = left for vertical-rl
			} else {
				// vertical-lr: blocks progress left-to-right
				childMarginBlockStart = childMarginLeft   // Start = left for vertical-lr
				childMarginBlockEnd = childMarginRight    // End = right for vertical-lr
			}
			childMarginInlineStart = childMarginTop   // Inline start = top
			childMarginInlineEnd = childMarginBottom  // Inline end = bottom
		} else {
			// Horizontal mode: block axis is vertical
			childMarginBlockStart = childMarginTop     // Start = top
			childMarginBlockEnd = childMarginBottom    // End = bottom
			childMarginInlineStart = childMarginLeft   // Inline start = left
			childMarginInlineEnd = childMarginRight    // Inline end = right
		}

		// Apply margin collapsing with previous sibling
		// Collapsed margin is max of adjacent margins, not sum
		var effectiveStartMargin float64
		if i == 0 {
			// First child: use its start margin as-is (could collapse with parent, but we don't implement that yet)
			effectiveStartMargin = childMarginBlockStart
			currentBlockPos = effectiveStartMargin
		} else {
			// Collapse with previous sibling's end margin
			// The effective space is the max of the two margins
			effectiveStartMargin = max(prevEndMargin, childMarginBlockStart)
			// Since we already added prevEndMargin to currentBlockPos, subtract it and add the collapsed margin
			currentBlockPos = currentBlockPos - prevEndMargin + effectiveStartMargin
		}

		// Layout child
		var childSize Size
		if child.Style.Display == DisplayFlex {
			childSize = LayoutFlexbox(child, childConstraints, ctx)
		} else if child.Style.Display == DisplayGrid {
			childSize = LayoutGrid(child, childConstraints, ctx)
		} else if child.Style.Display == DisplayInlineText {
			childSize = LayoutText(child, childConstraints, ctx)
		} else {
			childSize = LayoutBlock(child, childConstraints, ctx)
		}

		// Resolve parent's padding and border for positioning
		parentPaddingLeft := ResolveLength(node.Style.Padding.Left, ctx, parentFontSize)
		parentPaddingTop := ResolveLength(node.Style.Padding.Top, ctx, parentFontSize)
		parentBorderLeft := ResolveLength(node.Style.Border.Left, ctx, parentFontSize)
		parentBorderTop := ResolveLength(node.Style.Border.Top, ctx, parentFontSize)

		// Get child block size for positioning (needed for right-to-left)
		var childBlockSize float64
		if isVertical {
			childBlockSize = childSize.Width
		} else {
			childBlockSize = childSize.Height
		}

		// Position child with padding, border, and margin offset
		// Children are positioned in the content area, which starts after padding + border
		var childX, childY float64
		if isVertical {
			// Vertical mode: block direction is X
			if writingMode.IsRightToLeft() {
				// vertical-rl: position from right edge, moving leftward
				// We need the content width to calculate position from right
				contentWidth := setup.contentWidth
				childX = parentPaddingLeft + parentBorderLeft + contentWidth - currentBlockPos - childBlockSize
			} else {
				// vertical-lr: position from left edge, moving rightward
				childX = parentPaddingLeft + parentBorderLeft + currentBlockPos
			}
			childY = parentPaddingTop + parentBorderTop + childMarginInlineStart
		} else {
			// Horizontal mode: block direction is Y
			childX = parentPaddingLeft + parentBorderLeft + childMarginInlineStart
			childY = parentPaddingTop + parentBorderTop + currentBlockPos
		}

		child.Rect = Rect{
			X:      childX,
			Y:      childY,
			Width:  childSize.Width,
			Height: childSize.Height,
		}

		// Update currentBlockPos for next child (add child size in block direction and end margin)
		// childBlockSize already calculated above for positioning
		currentBlockPos += childBlockSize + childMarginBlockEnd

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

		// Store end margin for next iteration's collapse calculation
		prevEndMargin = childMarginBlockEnd
	}

	return currentBlockPos, maxCrossSize
}
