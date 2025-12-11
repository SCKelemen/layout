package layout

// LayoutBlock performs basic block layout on a node
func LayoutBlock(node *Node, constraints Constraints) Size {
	// Calculate available space
	availableWidth := constraints.MaxWidth
	availableHeight := constraints.MaxHeight

	// Account for padding and border
	horizontalPadding := node.Style.Padding.Left + node.Style.Padding.Right
	verticalPadding := node.Style.Padding.Top + node.Style.Padding.Bottom
	horizontalBorder := node.Style.Border.Left + node.Style.Border.Right
	verticalBorder := node.Style.Border.Top + node.Style.Border.Bottom
	horizontalPaddingBorder := horizontalPadding + horizontalBorder
	verticalPaddingBorder := verticalPadding + verticalBorder

	// Clamp content size to >= 0
	contentWidth := availableWidth - horizontalPaddingBorder
	if contentWidth < 0 {
		contentWidth = 0
	}
	contentHeight := availableHeight - verticalPaddingBorder
	if contentHeight < 0 {
		contentHeight = 0
	}
	

	// Convert width/height from specified box-sizing to content-box for internal calculations
	// According to W3C CSS Box Sizing spec:
	// - content-box: width/height = content size only
	// - border-box: width/height = content + padding + border
	specifiedWidth := convertToContentSize(node.Style.Width, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, true)
	specifiedHeight := convertToContentSize(node.Style.Height, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, false)

	// Determine node size (now in content-box units)
	// CRITICAL FIX: Treat 0 as auto when aspect ratio is set (Go zero value issue)
	// In Go, unset float64 fields default to 0, not -1, so we need to treat 0 as auto
	// when aspect ratio is set and both dimensions are 0
	isAutoWidth := specifiedWidth < 0 || (specifiedWidth == 0 && node.Style.AspectRatio > 0 && specifiedHeight == 0)
	isAutoHeight := specifiedHeight < 0 || (specifiedHeight == 0 && node.Style.AspectRatio > 0 && specifiedWidth == 0)
	
	nodeWidth := specifiedWidth
	if isAutoWidth {
		nodeWidth = contentWidth // auto
	}
	nodeHeight := specifiedHeight
	if isAutoHeight {
		// For auto height, don't set to Unbounded initially if aspect ratio will calculate it
		// Aspect ratio calculation happens next and will set height based on width
		if node.Style.AspectRatio > 0 && isAutoWidth && contentWidth > 0 {
			// Will be calculated by aspect ratio below
			nodeHeight = 0
		} else {
			nodeHeight = contentHeight // auto
		}
	}
	
	// Track if aspect ratio calculated dimensions (so we don't overwrite with children later)
	aspectRatioCalculatedWidth := false
	aspectRatioCalculatedHeight := false

	// Apply aspect ratio if set (before min/max constraints)
	// Aspect ratio affects sizing when one dimension is auto
	// According to CSS Box Sizing Module Level 4:
	// - When both width and height are auto, aspect ratio uses the available space
	// - Prefer width-based calculation if available width > 0
	// - If available height is bounded and would constrain, use height-based instead
	if node.Style.AspectRatio > 0 {
		if isAutoWidth && isAutoHeight {
			// Both auto: use available space and aspect ratio
			// Prefer width-based calculation (use available width)
			if contentWidth > 0 {
				// Use available width, calculate height from aspect ratio
				// nodeWidth is already set to contentWidth from line 37
				nodeHeight = nodeWidth / node.Style.AspectRatio
				aspectRatioCalculatedHeight = true
				aspectRatioCalculatedWidth = true // Width is set from contentWidth (nodeWidth)
				// Constrain to available height if it's bounded
				if contentHeight < Unbounded && nodeHeight > contentHeight {
					nodeHeight = contentHeight
					nodeWidth = nodeHeight * node.Style.AspectRatio
					// Both recalculated
				}
			} else if contentHeight > 0 && contentHeight < Unbounded {
				// Use available height, calculate width from aspect ratio
				// Only if contentWidth is 0 and contentHeight is bounded
				nodeWidth = contentHeight * node.Style.AspectRatio
				nodeHeight = contentHeight
				aspectRatioCalculatedWidth = true
				aspectRatioCalculatedHeight = true
			}
		} else if isAutoWidth {
			// Width is auto, height is set: calculate width from height and aspect ratio
			nodeWidth = nodeHeight * node.Style.AspectRatio
			aspectRatioCalculatedWidth = true
		} else if isAutoHeight {
			// Height is auto, width is set: calculate height from width and aspect ratio
			nodeHeight = nodeWidth / node.Style.AspectRatio
			aspectRatioCalculatedHeight = true
		}
		// If both width and height are explicitly set, aspect ratio is ignored (CSS behavior)
	}

	// Apply min/max constraints
	// Min/Max constraints also respect box-sizing (they apply to the same box as width/height)
	minWidthContent := convertMinMaxToContentSize(node.Style.MinWidth, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, true)
	maxWidthContent := convertMinMaxToContentSize(node.Style.MaxWidth, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, true)
	minHeightContent := convertMinMaxToContentSize(node.Style.MinHeight, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, false)
	maxHeightContent := convertMinMaxToContentSize(node.Style.MaxHeight, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, false)

	// If aspect ratio calculated dimensions, we need to maintain the ratio when min/max are applied
	if minWidthContent > 0 {
		nodeWidth = max(nodeWidth, minWidthContent)
		// If aspect ratio calculated width, recalculate height to maintain ratio
		if aspectRatioCalculatedWidth && node.Style.AspectRatio > 0 {
			nodeHeight = nodeWidth / node.Style.AspectRatio
		}
	}
	if maxWidthContent > 0 && maxWidthContent < Unbounded {
		nodeWidth = min(nodeWidth, maxWidthContent)
		// If aspect ratio calculated width, recalculate height to maintain ratio
		if aspectRatioCalculatedWidth && node.Style.AspectRatio > 0 {
			nodeHeight = nodeWidth / node.Style.AspectRatio
		}
	}
	if minHeightContent > 0 {
		oldHeight := nodeHeight
		nodeHeight = max(nodeHeight, minHeightContent)
		// If aspect ratio calculated height and MinHeight changed it, recalculate width to maintain ratio
		// This handles both cases: MinHeight increasing height, and MinHeight constraining a larger height
		if aspectRatioCalculatedHeight && node.Style.AspectRatio > 0 && nodeHeight != oldHeight {
			nodeWidth = nodeHeight * node.Style.AspectRatio
		}
	}
	if maxHeightContent > 0 && maxHeightContent < Unbounded {
		oldHeight := nodeHeight
		nodeHeight = min(nodeHeight, maxHeightContent)
		// If aspect ratio calculated height and MaxHeight decreased it, recalculate width to maintain ratio
		if aspectRatioCalculatedHeight && node.Style.AspectRatio > 0 && nodeHeight < oldHeight {
			nodeWidth = nodeHeight * node.Style.AspectRatio
		}
	}

	// Constrain to available space
	// Don't constrain if aspect ratio already calculated dimensions (they're already constrained)
	if !aspectRatioCalculatedWidth {
		nodeWidth = min(nodeWidth, contentWidth)
	}
	if !aspectRatioCalculatedHeight {
		nodeHeight = min(nodeHeight, contentHeight)
	}

	// Layout children (stack vertically for block layout)
	children := node.Children
	currentY := 0.0
	maxChildWidth := 0.0

	childConstraints := Constraints{
		MinWidth:  0,
		MaxWidth:  nodeWidth,
		MinHeight: 0,
		MaxHeight: Unbounded,
	}

	for _, child := range children {
		// Skip display:none children
		if child.Style.Display == DisplayNone {
			continue
		}

		var childSize Size
		if child.Style.Display == DisplayFlex {
			childSize = LayoutFlexbox(child, childConstraints)
		} else if child.Style.Display == DisplayGrid {
			childSize = LayoutGrid(child, childConstraints)
		} else {
			childSize = LayoutBlock(child, childConstraints)
		}

		// Position child with padding and border offset
		// Children are positioned in the content area, which starts after padding + border
		child.Rect = Rect{
			X:      node.Style.Padding.Left + node.Style.Border.Left,
			Y:      node.Style.Padding.Top + node.Style.Border.Top + currentY,
			Width:  childSize.Width,
			Height: childSize.Height,
		}

		currentY += childSize.Height
		if childSize.Width > maxChildWidth {
			maxChildWidth = childSize.Width
		}
	}

	// If height is auto, use children height (unless aspect ratio already calculated it)
	if isAutoHeight && !aspectRatioCalculatedHeight {
		// Aspect ratio didn't calculate height, so use children height
		nodeHeight = currentY
		// Ensure MinHeight is still respected even when using children height
		if minHeightContent > 0 {
			nodeHeight = max(nodeHeight, minHeightContent)
		}
		// If no children and no MinHeight and no aspect ratio, height is 0 (which is correct)
		// But this can cause issues in auto-sized grid rows
	} else if isAutoHeight {
		// Aspect ratio calculated height, but ensure MinHeight is still respected
		if minHeightContent > 0 {
			oldHeight := nodeHeight
			nodeHeight = max(nodeHeight, minHeightContent)
			// If MinHeight increased height and aspect ratio is set, recalculate width to maintain ratio
			if node.Style.AspectRatio > 0 && nodeHeight > oldHeight {
				nodeWidth = nodeHeight * node.Style.AspectRatio
			}
		}
	}

	// If width is auto, use max child width (unless aspect ratio already calculated it)
	if isAutoWidth {
		if !aspectRatioCalculatedWidth {
			// Aspect ratio didn't calculate width, so use children width
			// But if there are no children and we have contentWidth, use that
			if maxChildWidth == 0 && contentWidth > 0 {
				nodeWidth = contentWidth
			} else {
				nodeWidth = maxChildWidth
			}
		}
		// Ensure MinWidth is still respected (even if aspect ratio calculated width)
		if minWidthContent > 0 {
			oldWidth := nodeWidth
			nodeWidth = max(nodeWidth, minWidthContent)
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
	finalWidth := nodeWidth + horizontalPaddingBorder
	finalHeight := nodeHeight + verticalPaddingBorder
	

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
