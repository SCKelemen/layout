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

	// Clamp content size to >= 0
	contentWidth := availableWidth - horizontalPadding - horizontalBorder
	if contentWidth < 0 {
		contentWidth = 0
	}
	contentHeight := availableHeight - verticalPadding - verticalBorder
	if contentHeight < 0 {
		contentHeight = 0
	}

	// Determine node size
	nodeWidth := node.Style.Width
	if nodeWidth < 0 {
		nodeWidth = contentWidth // auto
	}
	nodeHeight := node.Style.Height
	if nodeHeight < 0 {
		// For auto height, don't set to Unbounded initially if aspect ratio will calculate it
		// Aspect ratio calculation happens next and will set height based on width
		if node.Style.AspectRatio > 0 && node.Style.Width < 0 && contentWidth > 0 {
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
		if node.Style.Width < 0 && node.Style.Height < 0 {
			// Both auto: use available space and aspect ratio
			// Prefer width-based calculation (use available width)
			if contentWidth > 0 {
				// Use available width, calculate height from aspect ratio
				// nodeWidth is already set to contentWidth from line 28
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
		} else if node.Style.Width < 0 {
			// Width is auto, height is set: calculate width from height and aspect ratio
			nodeWidth = nodeHeight * node.Style.AspectRatio
			aspectRatioCalculatedWidth = true
		} else if node.Style.Height < 0 {
			// Height is auto, width is set: calculate height from width and aspect ratio
			nodeHeight = nodeWidth / node.Style.AspectRatio
			aspectRatioCalculatedHeight = true
		}
		// If both width and height are explicitly set, aspect ratio is ignored (CSS behavior)
	}

	// Apply min/max constraints
	// If aspect ratio calculated dimensions, we need to maintain the ratio when min/max are applied
	if node.Style.MinWidth > 0 {
		nodeWidth = max(nodeWidth, node.Style.MinWidth)
		// If aspect ratio calculated width, recalculate height to maintain ratio
		if aspectRatioCalculatedWidth && node.Style.AspectRatio > 0 {
			nodeHeight = nodeWidth / node.Style.AspectRatio
		}
	}
	if node.Style.MaxWidth > 0 && node.Style.MaxWidth < Unbounded {
		nodeWidth = min(nodeWidth, node.Style.MaxWidth)
		// If aspect ratio calculated width, recalculate height to maintain ratio
		if aspectRatioCalculatedWidth && node.Style.AspectRatio > 0 {
			nodeHeight = nodeWidth / node.Style.AspectRatio
		}
	}
	if node.Style.MinHeight > 0 {
		oldHeight := nodeHeight
		nodeHeight = max(nodeHeight, node.Style.MinHeight)
		// If aspect ratio calculated height and MinHeight increased it, recalculate width to maintain ratio
		if aspectRatioCalculatedHeight && node.Style.AspectRatio > 0 && nodeHeight > oldHeight {
			nodeWidth = nodeHeight * node.Style.AspectRatio
		}
	}
	if node.Style.MaxHeight > 0 && node.Style.MaxHeight < Unbounded {
		oldHeight := nodeHeight
		nodeHeight = min(nodeHeight, node.Style.MaxHeight)
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

		// Position child with padding offset
		child.Rect = Rect{
			X:      node.Style.Padding.Left,
			Y:      node.Style.Padding.Top + currentY,
			Width:  childSize.Width,
			Height: childSize.Height,
		}

		currentY += childSize.Height
		if childSize.Width > maxChildWidth {
			maxChildWidth = childSize.Width
		}
	}

	// If height is auto, use children height (unless aspect ratio already calculated it)
	if node.Style.Height < 0 && !aspectRatioCalculatedHeight {
		// Aspect ratio didn't calculate height, so use children height
		nodeHeight = currentY
		// Ensure MinHeight is still respected even when using children height
		if node.Style.MinHeight > 0 {
			nodeHeight = max(nodeHeight, node.Style.MinHeight)
		}
		// If no children and no MinHeight and no aspect ratio, height is 0 (which is correct)
		// But this can cause issues in auto-sized grid rows
	} else if node.Style.Height < 0 {
		// Aspect ratio calculated height, but ensure MinHeight is still respected
		if node.Style.MinHeight > 0 {
			oldHeight := nodeHeight
			nodeHeight = max(nodeHeight, node.Style.MinHeight)
			// If MinHeight increased height and aspect ratio is set, recalculate width to maintain ratio
			if node.Style.AspectRatio > 0 && nodeHeight > oldHeight {
				nodeWidth = nodeHeight * node.Style.AspectRatio
			}
		}
	}

	// If width is auto, use max child width (unless aspect ratio already calculated it)
	if node.Style.Width < 0 {
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
		if node.Style.MinWidth > 0 {
			oldWidth := nodeWidth
			nodeWidth = max(nodeWidth, node.Style.MinWidth)
			// If MinWidth increased width and aspect ratio is set, recalculate height to maintain ratio
			if node.Style.AspectRatio > 0 && nodeWidth > oldWidth && aspectRatioCalculatedHeight {
				nodeHeight = nodeWidth / node.Style.AspectRatio
			}
		}
	}

	// Calculate final size including padding and border
	finalWidth := nodeWidth + horizontalPadding + horizontalBorder
	finalHeight := nodeHeight + verticalPadding + verticalBorder

	// Constrain size and apply to Rect
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
