package layout

// LayoutBlock performs block layout on a node.
//
// Algorithm based on CSS Box Model and Block Layout:
// - CSS Box Model Module Level 3: §4: Box Model
// - CSS Box Model Module Level 3: §8.3.1: Collapsing margins
// - CSS Display Module Level 3: §4: Block-level Boxes
// - CSS Sizing Module Level 3: §5: Intrinsic Size Determination
//
// See:
// - https://www.w3.org/TR/css-box-3/
// - https://www.w3.org/TR/css-display-3/
// - https://www.w3.org/TR/css-sizing-3/
func LayoutBlock(node *Node, constraints Constraints, ctx *LayoutContext) Size {
	// Get current font size for em resolution
	currentFontSize := getCurrentFontSize(node, ctx)

	// §4: Box Model - Setup and determine container dimensions
	setup := blockDetermineContainerSize(node, constraints, ctx, currentFontSize)

	// §5: Aspect Ratios - Determine node size considering aspect ratio
	nodeWidth, nodeHeight, aspectRatioCalculatedWidth, aspectRatioCalculatedHeight := blockDetermineSize(node, setup, ctx, currentFontSize)

	// §5: Intrinsic Size Determination - Apply min/max constraints
	nodeWidth, nodeHeight = blockApplyConstraints(node, setup, nodeWidth, nodeHeight, aspectRatioCalculatedWidth, aspectRatioCalculatedHeight)

	// §8.3.1: Collapsing margins - Layout children with margin collapsing
	currentY, maxChildWidth := blockLayoutChildren(node, setup, nodeWidth, ctx, currentFontSize)

	// If height is auto, use children height (unless aspect ratio already calculated it)
	if setup.isAutoHeight && !aspectRatioCalculatedHeight {
		// Aspect ratio didn't calculate height, so use children height
		nodeHeight = currentY
		// Ensure MinHeight is still respected even when using children height
		if setup.minHeightContent > 0 {
			nodeHeight = max(nodeHeight, setup.minHeightContent)
		}
	} else if setup.isAutoHeight {
		// Aspect ratio calculated height, but ensure MinHeight is still respected
		if setup.minHeightContent > 0 {
			oldHeight := nodeHeight
			nodeHeight = max(nodeHeight, setup.minHeightContent)
			// If MinHeight increased height and aspect ratio is set, recalculate width to maintain ratio
			if node.Style.AspectRatio > 0 && nodeHeight > oldHeight {
				nodeWidth = nodeHeight * node.Style.AspectRatio
			}
		}
	}

	// If width is auto, use max child width (unless aspect ratio already calculated it)
	if setup.isAutoWidth {
		if !aspectRatioCalculatedWidth {
			// Aspect ratio didn't calculate width, so use children width
			// But if there are no children and we have contentWidth, use that
			if maxChildWidth == 0 && setup.contentWidth > 0 {
				nodeWidth = setup.contentWidth
			} else {
				nodeWidth = maxChildWidth
			}
		}
		// Ensure MinWidth is still respected (even if aspect ratio calculated width)
		if setup.minWidthContent > 0 {
			oldWidth := nodeWidth
			nodeWidth = max(nodeWidth, setup.minWidthContent)
			// If MinWidth increased width and aspect ratio is set, recalculate height to maintain ratio
			if node.Style.AspectRatio > 0 && nodeWidth > oldWidth && aspectRatioCalculatedHeight {
				nodeHeight = nodeWidth / node.Style.AspectRatio
			}
		}
	}

	// Calculate final size
	// nodeWidth and nodeHeight are in content-box units
	// For border-box, we need to convert back to total size (add padding + border)
	// For content-box, we add padding and border to get total size
	// Both approaches result in the same total size
	finalWidth := nodeWidth + setup.horizontalPaddingBorder
	finalHeight := nodeHeight + setup.verticalPaddingBorder

	// Constrain size and apply to Rect
	// CRITICAL: node.Rect must respect constraints to match the returned Size
	constrainedSize := constraints.Constrain(Size{
		Width:  finalWidth,
		Height: finalHeight,
	})

	node.Rect = Rect{
		X:      0,
		Y:      0,
		Width:  constrainedSize.Width,
		Height: constrainedSize.Height,
	}

	return constrainedSize
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// getCurrentFontSize returns the current font size for Length resolution.
// Falls back to ctx.RootFontSize if the node's TextStyle is not set.
func getCurrentFontSize(node *Node, ctx *LayoutContext) float64 {
	if node.Style.TextStyle != nil && node.Style.TextStyle.FontSize > 0 {
		return node.Style.TextStyle.FontSize
	}
	return ctx.RootFontSize
}
