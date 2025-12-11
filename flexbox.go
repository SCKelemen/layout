package layout

import (
	"math"
)

// LayoutFlexbox performs flexbox layout on a node
func LayoutFlexbox(node *Node, constraints Constraints) Size {
	if node.Style.Display != DisplayFlex {
		// If not flex, delegate to block layout
		return LayoutBlock(node, constraints)
	}

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

	// Determine main and cross axis
	isRow := node.Style.FlexDirection == FlexDirectionRow || node.Style.FlexDirection == FlexDirectionRowReverse
	mainSize := contentWidth
	crossSize := contentHeight
	if !isRow {
		mainSize, crossSize = crossSize, mainSize
	}

	// Layout children
	children := node.Children
	if len(children) == 0 {
		// Empty flex container
		resultSize := Size{
			Width:  horizontalPadding + horizontalBorder,
			Height: verticalPadding + verticalBorder,
		}
		node.Rect = Rect{
			X:      0,
			Y:      0,
			Width:  resultSize.Width,
			Height: resultSize.Height,
		}
		return constraints.Constrain(resultSize)
	}

	// Step 1: Measure all children
	flexItems := make([]*flexItem, len(children))
	for i, child := range children {
		item := &flexItem{
			node: child,
		}

		// Get child margins
		var childMainMarginStart, childMainMarginEnd, childCrossMarginStart, childCrossMarginEnd float64
		if isRow {
			childMainMarginStart = child.Style.Margin.Left
			childMainMarginEnd = child.Style.Margin.Right
			childCrossMarginStart = child.Style.Margin.Top
			childCrossMarginEnd = child.Style.Margin.Bottom
		} else {
			childMainMarginStart = child.Style.Margin.Top
			childMainMarginEnd = child.Style.Margin.Bottom
			childCrossMarginStart = child.Style.Margin.Left
			childCrossMarginEnd = child.Style.Margin.Right
		}
		item.mainMarginStart = childMainMarginStart
		item.mainMarginEnd = childMainMarginEnd
		item.crossMarginStart = childCrossMarginStart
		item.crossMarginEnd = childCrossMarginEnd

		// Determine child constraints (account for margins)
		childMainSize := mainSize
		childCrossSize := crossSize
		if node.Style.FlexWrap == FlexWrapNoWrap {
			// In nowrap, children share main axis space
			childMainSize = Unbounded
		}

		childConstraints := Constraints{
			MinWidth:  0,
			MaxWidth:  childMainSize,
			MinHeight: 0,
			MaxHeight: childCrossSize,
		}
		if !isRow {
			childConstraints.MaxWidth, childConstraints.MaxHeight = childConstraints.MaxHeight, childConstraints.MaxWidth
		}

		// Measure child
		var childSize Size
		if child.Style.Display == DisplayFlex {
			childSize = LayoutFlexbox(child, childConstraints)
		} else if child.Style.Display == DisplayGrid {
			childSize = LayoutGrid(child, childConstraints)
		} else {
			childSize = LayoutBlock(child, childConstraints)
		}

		if isRow {
			item.mainSize = childSize.Width
			item.crossSize = childSize.Height
		} else {
			item.mainSize = childSize.Height
			item.crossSize = childSize.Width
		}

		// Store the measured size as a fallback
		measuredMainSize := item.mainSize

		// Get flex properties
		item.flexGrow = child.Style.FlexGrow
		if item.flexGrow == 0 {
			item.flexGrow = 0
		}
		item.flexShrink = child.Style.FlexShrink
		if item.flexShrink == 0 {
			item.flexShrink = 1 // Default shrink factor
		}
		item.flexBasis = child.Style.FlexBasis
		if item.flexBasis < 0 {
			item.flexBasis = item.mainSize // auto means use main size
		}

		item.baseSize = item.flexBasis

		// Ensure baseSize is never 0 if we have a measured size
		if item.baseSize == 0 && measuredMainSize > 0 {
			item.baseSize = measuredMainSize
			item.flexBasis = measuredMainSize
		}
		flexItems[i] = item
	}

	// Step 2: Calculate flex line (for wrapping)
	lines := calculateFlexLines(flexItems, mainSize, node.Style.FlexWrap == FlexWrapWrap || node.Style.FlexWrap == FlexWrapWrapReverse)

	// Step 3: Layout each line
	totalCrossSize := 0.0
	mainOffset := 0.0
	lineStartCrossOffset := 0.0

	for lineIdx, line := range lines {
		// Calculate total flex grow and shrink
		totalFlexGrow := 0.0
		totalFlexShrink := 0.0
		for _, item := range line {
			totalFlexGrow += item.flexGrow
			totalFlexShrink += item.flexShrink
		}

		// Calculate free space (including margins)
		usedMainSize := 0.0
		for _, item := range line {
			usedMainSize += item.baseSize + item.mainMarginStart + item.mainMarginEnd
		}
		freeSpace := mainSize - usedMainSize

		// Distribute free space
		if freeSpace > 0 && totalFlexGrow > 0 {
			// Grow items
			for _, item := range line {
				if item.flexGrow > 0 {
					growAmount := (freeSpace * item.flexGrow) / totalFlexGrow
					item.mainSize = item.baseSize + growAmount
				} else {
					item.mainSize = item.baseSize
				}
			}
		} else if freeSpace < 0 && totalFlexShrink > 0 {
			// Shrink items
			for _, item := range line {
				if item.flexShrink > 0 {
					shrinkAmount := (-freeSpace * item.flexShrink) / totalFlexShrink
					item.mainSize = math.Max(0, item.baseSize-shrinkAmount)
				} else {
					item.mainSize = item.baseSize
				}
			}
		} else {
			// No flex, use base size (which is the measured/intrinsic size)
			for _, item := range line {
				item.mainSize = item.baseSize
			}
		}

		// Calculate cross size for line (including margins)
		lineCrossSize := 0.0
		for _, item := range line {
			itemCrossSizeWithMargins := item.crossSize + item.crossMarginStart + item.crossMarginEnd
			if itemCrossSizeWithMargins > lineCrossSize {
				lineCrossSize = itemCrossSizeWithMargins
			}
		}
		if node.Style.AlignItems == AlignItemsStretch {
			lineCrossSize = crossSize
		}

		// Align items in cross axis
		// For alignment, use the container's cross size, not the line's cross size
		alignmentCrossSize := crossSize
		if node.Style.AlignItems != AlignItemsStretch {
			// For non-stretch, we can use lineCrossSize if it's smaller than container
			// But for proper centering, we should use container size
			alignmentCrossSize = crossSize
		}

		for _, item := range line {
			crossOffset := 0.0
			itemCrossSizeWithMargins := item.crossSize + item.crossMarginStart + item.crossMarginEnd
			switch node.Style.AlignItems {
			case AlignItemsFlexStart:
				crossOffset = item.crossMarginStart
			case AlignItemsFlexEnd:
				crossOffset = alignmentCrossSize - item.crossSize - item.crossMarginEnd
			case AlignItemsCenter:
				crossOffset = (alignmentCrossSize - itemCrossSizeWithMargins) / 2 + item.crossMarginStart
			case AlignItemsStretch:
				// Stretch: item fills cross axis, but margins are still applied
				item.crossSize = lineCrossSize - item.crossMarginStart - item.crossMarginEnd
				if item.crossSize < 0 {
					item.crossSize = 0
				}
				crossOffset = item.crossMarginStart
			default:
				crossOffset = item.crossMarginStart
			}

			// Set cross-axis position and size (main-axis will be set by justifyContent)
			if isRow {
				item.node.Rect = Rect{
					Y:      lineStartCrossOffset + crossOffset,
					Width:  item.mainSize,
					Height: item.crossSize,
				}
			} else {
				item.node.Rect = Rect{
					X:      lineStartCrossOffset + crossOffset,
					Width:  item.crossSize,
					Height: item.mainSize,
				}
			}
		}

		// Justify content (main axis alignment) - this sets main-axis positions
		justifyContent(node.Style.JustifyContent, line, mainOffset, mainSize, isRow)

		// Update offsets (including margins)
		maxMainInLine := 0.0
		for _, item := range line {
			if isRow {
				itemEnd := item.node.Rect.X + item.node.Rect.Width + item.mainMarginEnd
				if itemEnd > mainOffset+maxMainInLine {
					maxMainInLine = itemEnd - mainOffset
				}
			} else {
				itemEnd := item.node.Rect.Y + item.node.Rect.Height + item.mainMarginEnd
				if itemEnd > mainOffset+maxMainInLine {
					maxMainInLine = itemEnd - mainOffset
				}
			}
		}

		if node.Style.FlexWrap == FlexWrapWrap || node.Style.FlexWrap == FlexWrapWrapReverse {
			totalCrossSize += lineCrossSize
			if lineIdx < len(lines)-1 {
				// Add gap if specified (simplified, assuming 0 for now)
			}
			lineStartCrossOffset = totalCrossSize - lineCrossSize
		} else {
			totalCrossSize = lineCrossSize
			lineStartCrossOffset = 0
		}
		mainOffset += maxMainInLine
	}

	// Calculate container size
	var containerSize Size
	if isRow {
		containerSize = Size{
			Width:  mainOffset + horizontalPadding + horizontalBorder,
			Height: totalCrossSize + verticalPadding + verticalBorder,
		}
	} else {
		containerSize = Size{
			Width:  totalCrossSize + horizontalPadding + horizontalBorder,
			Height: mainOffset + verticalPadding + verticalBorder,
		}
	}

	// Set container rect
	node.Rect = Rect{
		X:      0,
		Y:      0,
		Width:  containerSize.Width,
		Height: containerSize.Height,
	}

	return constraints.Constrain(containerSize)
}

type flexItem struct {
	node            *Node
	mainSize        float64
	crossSize       float64
	baseSize        float64
	flexGrow        float64
	flexShrink      float64
	flexBasis       float64
	mainMarginStart float64
	mainMarginEnd   float64
	crossMarginStart float64
	crossMarginEnd  float64
}

func calculateFlexLines(items []*flexItem, containerMainSize float64, wrap bool) [][]*flexItem {
	if !wrap {
		return [][]*flexItem{items}
	}

	lines := [][]*flexItem{}
	currentLine := []*flexItem{}
	currentLineSize := 0.0

	for _, item := range items {
		// Include margins in item size for wrapping calculation
		itemSize := item.baseSize + item.mainMarginStart + item.mainMarginEnd
		if currentLineSize+itemSize > containerMainSize && len(currentLine) > 0 {
			lines = append(lines, currentLine)
			currentLine = []*flexItem{}
			currentLineSize = 0
		}
		currentLine = append(currentLine, item)
		currentLineSize += itemSize
	}

	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}

	return lines
}

func justifyContent(justify JustifyContent, line []*flexItem, startOffset, containerSize float64, isRow bool) {
	if len(line) == 0 {
		return
	}

	// Calculate total size of items in main axis (including margins)
	totalItemSize := 0.0
	for _, item := range line {
		if isRow {
			totalItemSize += item.node.Rect.Width + item.mainMarginStart + item.mainMarginEnd
		} else {
			totalItemSize += item.node.Rect.Height + item.mainMarginStart + item.mainMarginEnd
		}
	}

	freeSpace := containerSize - totalItemSize
	var offset float64

	switch justify {
	case JustifyContentFlexStart:
		offset = 0
	case JustifyContentFlexEnd:
		offset = freeSpace
	case JustifyContentCenter:
		offset = freeSpace / 2
	case JustifyContentSpaceBetween:
		if len(line) > 1 {
			gap := freeSpace / float64(len(line)-1)
			offset = 0
			currentPos := startOffset
			for _, item := range line {
				if isRow {
					item.node.Rect.X = currentPos + item.mainMarginStart
					currentPos += item.node.Rect.Width + item.mainMarginStart + item.mainMarginEnd + gap
				} else {
					item.node.Rect.Y = currentPos + item.mainMarginStart
					currentPos += item.node.Rect.Height + item.mainMarginStart + item.mainMarginEnd + gap
				}
			}
			return
		}
		offset = 0
	case JustifyContentSpaceAround:
		if len(line) > 0 {
			gap := freeSpace / float64(len(line))
			offset = gap / 2
		}
	case JustifyContentSpaceEvenly:
		if len(line) > 0 {
			gap := freeSpace / float64(len(line)+1)
			offset = gap
		}
	}

	// Apply offset (accounting for margins)
	currentPos := startOffset + offset
	for _, item := range line {
		if isRow {
			item.node.Rect.X = currentPos + item.mainMarginStart
			currentPos += item.node.Rect.Width + item.mainMarginStart + item.mainMarginEnd
		} else {
			item.node.Rect.Y = currentPos + item.mainMarginStart
			currentPos += item.node.Rect.Height + item.mainMarginStart + item.mainMarginEnd
		}
	}
}
