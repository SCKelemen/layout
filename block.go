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

	contentWidth := availableWidth - horizontalPadding - horizontalBorder
	contentHeight := availableHeight - verticalPadding - verticalBorder

	// Determine node size
	nodeWidth := node.Style.Width
	if nodeWidth < 0 {
		nodeWidth = contentWidth // auto
	}
	nodeHeight := node.Style.Height
	if nodeHeight < 0 {
		nodeHeight = contentHeight // auto
	}

	// Apply min/max constraints
	if node.Style.MinWidth > 0 {
		nodeWidth = max(nodeWidth, node.Style.MinWidth)
	}
	if node.Style.MaxWidth > 0 && node.Style.MaxWidth < Unbounded {
		nodeWidth = min(nodeWidth, node.Style.MaxWidth)
	}
	if node.Style.MinHeight > 0 {
		nodeHeight = max(nodeHeight, node.Style.MinHeight)
	}
	if node.Style.MaxHeight > 0 && node.Style.MaxHeight < Unbounded {
		nodeHeight = min(nodeHeight, node.Style.MaxHeight)
	}

	// Constrain to available space
	nodeWidth = min(nodeWidth, contentWidth)
	nodeHeight = min(nodeHeight, contentHeight)

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
		var childSize Size
		if child.Style.Display == DisplayFlex {
			childSize = LayoutFlexbox(child, childConstraints)
		} else if child.Style.Display == DisplayGrid {
			childSize = LayoutGrid(child, childConstraints)
		} else {
			childSize = LayoutBlock(child, childConstraints)
		}

		child.Rect = Rect{
			X:      0,
			Y:      currentY,
			Width:  childSize.Width,
			Height: childSize.Height,
		}

		currentY += childSize.Height
		if childSize.Width > maxChildWidth {
			maxChildWidth = childSize.Width
		}
	}

	// If height is auto, use children height
	if node.Style.Height < 0 {
		nodeHeight = currentY
		// Ensure MinHeight is still respected even when using children height
		if node.Style.MinHeight > 0 {
			nodeHeight = max(nodeHeight, node.Style.MinHeight)
		}
		// If no children and no MinHeight, height is 0 (which is correct)
		// But this can cause issues in auto-sized grid rows
	}

	// If width is auto, use max child width
	if node.Style.Width < 0 {
		nodeWidth = maxChildWidth
		// Ensure MinWidth is still respected even when using children width
		if node.Style.MinWidth > 0 {
			nodeWidth = max(nodeWidth, node.Style.MinWidth)
		}
	}

	// Calculate final size including padding and border
	finalWidth := nodeWidth + horizontalPadding + horizontalBorder
	finalHeight := nodeHeight + verticalPadding + verticalBorder

	node.Rect = Rect{
		X:      0,
		Y:      0,
		Width:  finalWidth,
		Height: finalHeight,
	}

	return constraints.Constrain(Size{
		Width:  finalWidth,
		Height: finalHeight,
	})
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
