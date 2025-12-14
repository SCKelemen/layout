package layout

// blockLayoutChildren lays out children in vertical stack with margin collapsing.
//
// Algorithm based on CSS Box Model Module Level 3:
// - §8.3.1: Collapsing margins
//
// See: https://www.w3.org/TR/css-box-3/#collapsing-margins
//
// Margin collapsing rules (simplified):
// 1. Adjacent vertical margins of block-level boxes collapse (use max, not sum)
// 2. Parent and first child top margins collapse if no border/padding/content separates them
// 3. Parent and last child bottom margins collapse if no border/padding/height separates them
// 4. Empty block margins collapse with themselves
//
// For this implementation, we focus on the most common case:
// - Adjacent sibling margins collapse (rule 1)
func blockLayoutChildren(node *Node, setup blockSetup, nodeWidth float64, ctx *LayoutContext, parentFontSize float64) (currentY, maxChildWidth float64) {
	children := node.Children
	currentY = 0.0
	maxChildWidth = 0.0

	childConstraints := Constraints{
		MinWidth:  0,
		MaxWidth:  nodeWidth,
		MinHeight: 0,
		MaxHeight: Unbounded,
	}

	// Track previous child's bottom margin for collapsing
	var prevBottomMargin float64 = 0.0

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

		// Apply margin collapsing with previous sibling
		// Collapsed margin is max of adjacent margins, not sum
		var effectiveTopMargin float64
		if i == 0 {
			// First child: use its top margin as-is (could collapse with parent, but we don't implement that yet)
			effectiveTopMargin = childMarginTop
			currentY = effectiveTopMargin
		} else {
			// Collapse with previous sibling's bottom margin
			// The effective space is the max of the two margins
			effectiveTopMargin = max(prevBottomMargin, childMarginTop)
			// Since we already added prevBottomMargin to currentY, subtract it and add the collapsed margin
			currentY = currentY - prevBottomMargin + effectiveTopMargin
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

		// Position child with padding, border, and margin offset
		// Children are positioned in the content area, which starts after padding + border
		child.Rect = Rect{
			X:      parentPaddingLeft + parentBorderLeft + childMarginLeft,
			Y:      parentPaddingTop + parentBorderTop + currentY,
			Width:  childSize.Width,
			Height: childSize.Height,
		}

		// Update currentY for next child (add child height and bottom margin)
		currentY += childSize.Height + childMarginBottom

		// Track max child width (including margins)
		childWidthWithMargins := childSize.Width + childMarginLeft + childMarginRight
		if childWidthWithMargins > maxChildWidth {
			maxChildWidth = childWidthWithMargins
		}

		// Store bottom margin for next iteration's collapse calculation
		prevBottomMargin = childMarginBottom
	}

	return currentY, maxChildWidth
}
