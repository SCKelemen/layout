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

	// If container has explicit width/height, use it to constrain available space
	// Similar to grid layout
	if node.Style.Width >= 0 {
		specifiedWidthContent := convertToContentSize(node.Style.Width, node.Style.BoxSizing, horizontalPadding+horizontalBorder, verticalPadding+verticalBorder, true)
		totalSpecifiedWidth := specifiedWidthContent + horizontalPadding + horizontalBorder
		if availableWidth >= Unbounded {
			availableWidth = totalSpecifiedWidth
		} else if totalSpecifiedWidth <= availableWidth {
			availableWidth = totalSpecifiedWidth
		}
	}
	if node.Style.Height >= 0 {
		specifiedHeightContent := convertToContentSize(node.Style.Height, node.Style.BoxSizing, horizontalPadding+horizontalBorder, verticalPadding+verticalBorder, false)
		totalSpecifiedHeight := specifiedHeightContent + verticalPadding + verticalBorder
		if availableHeight >= Unbounded {
			availableHeight = totalSpecifiedHeight
		} else if totalSpecifiedHeight <= availableHeight {
			availableHeight = totalSpecifiedHeight
		}
	}

	// Clamp content size to >= 0
	contentWidth := availableWidth - horizontalPadding - horizontalBorder
	if contentWidth < 0 {
		contentWidth = 0
	}
	contentHeight := availableHeight - verticalPadding - verticalBorder
	if contentHeight < 0 {
		contentHeight = 0
	}

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

	// Step 1: Measure all children (filter DisplayNone)
	flexItems := make([]*flexItem, 0, len(children))
	for _, child := range children {
		// Skip display:none children
		if child.Style.Display == DisplayNone {
			continue
		}
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
		} else if child.Style.Display == DisplayInlineText {
			childSize = LayoutText(child, childConstraints)
		} else {
			childSize = LayoutBlock(child, childConstraints)
		}

		if isRow {
			item.mainSize = childSize.Width
			item.crossSize = childSize.Height
			// Use explicit dimensions if measured size is 0 or Unbounded
			// This handles cases where LayoutBlock returns 0 or Unbounded for items with explicit dimensions
			if (item.mainSize == 0 || item.mainSize >= Unbounded) && child.Style.Width >= 0 {
				item.mainSize = child.Style.Width
			}
			if (item.crossSize == 0 || item.crossSize >= Unbounded) && child.Style.Height >= 0 {
				item.crossSize = child.Style.Height
			}
		} else {
			item.mainSize = childSize.Height
			item.crossSize = childSize.Width
			// Use explicit dimensions if measured size is 0 or Unbounded
			if (item.mainSize == 0 || item.mainSize >= Unbounded) && child.Style.Height >= 0 {
				item.mainSize = child.Style.Height
			}
			if (item.crossSize == 0 || item.crossSize >= Unbounded) && child.Style.Width >= 0 {
				item.crossSize = child.Style.Width
			}
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

		// Ensure baseSize is never 0 if we have a measured size or explicit width/height
		if item.baseSize == 0 {
			if measuredMainSize > 0 {
				item.baseSize = measuredMainSize
				item.flexBasis = measuredMainSize
			} else if isRow && child.Style.Width >= 0 {
				// Use explicit width for baseSize
				item.baseSize = child.Style.Width
				item.flexBasis = child.Style.Width
			} else if !isRow && child.Style.Height >= 0 {
				// Use explicit height for baseSize
				item.baseSize = child.Style.Height
				item.flexBasis = child.Style.Height
			}
		}
		flexItems = append(flexItems, item)
	}

	// Step 2: Calculate flex line (for wrapping)
	hasWrap := node.Style.FlexWrap == FlexWrapWrap || node.Style.FlexWrap == FlexWrapWrapReverse
	lines := calculateFlexLines(flexItems, mainSize, hasWrap)
	wrapReverse := node.Style.FlexWrap == FlexWrapWrapReverse

	// Step 3: Layout each line
	// First pass: calculate line cross sizes and total cross size
	lineCrossSizes := make([]float64, len(lines))
	totalCrossSize := 0.0
	maxLineMainSize := 0.0

	// Get gap values
	rowGap := node.Style.FlexRowGap
	if rowGap == 0 {
		rowGap = node.Style.FlexGap
	}
	columnGap := node.Style.FlexColumnGap
	if columnGap == 0 {
		columnGap = node.Style.FlexGap
	}

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
		// For single-line containers, apply stretch if align-items is stretch
		// For multi-line, align-content will handle stretching
		// Zero value is stretch (CSS Flexbox default)
		if node.Style.AlignItems == AlignItemsStretch && len(lines) == 1 {
			lineCrossSize = crossSize
		}

		// Store line cross size for align-content calculation
		lineCrossSizes[lineIdx] = lineCrossSize
		totalCrossSize += lineCrossSize
		if lineIdx < len(lines)-1 {
			totalCrossSize += rowGap
		}
	}

	// Debug: Check if we have multiple lines and what the sizes are
	// For align-content to work, we need multiple lines AND the container cross size must be larger than total

	// Step 4: Apply align-content to distribute lines along cross axis
	// align-content only applies when there are multiple lines
	lineOffsets := make([]float64, len(lines))
	// align-content requires multiple lines AND a bounded cross size (not Unbounded)
	// Check if crossSize is effectively bounded (not the Unbounded constant)
	if len(lines) > 1 && hasWrap {
		alignContent := node.Style.AlignContent
		if alignContent == 0 {
			alignContent = AlignContentStretch // Default
		}

		// Calculate free cross space
		freeCrossSpace := crossSize - totalCrossSize
		if freeCrossSpace < 0 {
			freeCrossSpace = 0
		}

		var startOffset float64
		switch alignContent {
		case AlignContentFlexStart:
			startOffset = 0
		case AlignContentFlexEnd:
			startOffset = freeCrossSpace
		case AlignContentCenter:
			startOffset = freeCrossSpace / 2
		case AlignContentSpaceBetween:
			if len(lines) > 1 {
				startOffset = 0
				spaceBetween := freeCrossSpace / float64(len(lines)-1)
				currentOffset := 0.0
				for i := range lines {
					lineOffsets[i] = currentOffset
					currentOffset += lineCrossSizes[i]
					if i < len(lines)-1 {
						currentOffset += rowGap + spaceBetween
					}
				}
			} else {
				startOffset = 0
			}
		case AlignContentSpaceAround:
			if len(lines) > 0 {
				spaceAround := freeCrossSpace / float64(len(lines))
				currentOffset := spaceAround / 2
				for i := range lines {
					lineOffsets[i] = currentOffset
					currentOffset += lineCrossSizes[i]
					if i < len(lines)-1 {
						currentOffset += rowGap + spaceAround
					}
				}
			}
		case AlignContentStretch:
			// Distribute free space equally to each line
			if freeCrossSpace > 0 && len(lines) > 0 {
				extraPerLine := freeCrossSpace / float64(len(lines))
				for i := range lineCrossSizes {
					lineCrossSizes[i] += extraPerLine
				}
			}
			startOffset = 0
		default:
			startOffset = 0
		}

		// Calculate line offsets for non-space-between/space-around
		if alignContent != AlignContentSpaceBetween && alignContent != AlignContentSpaceAround {
			currentOffset := startOffset
			for i := range lines {
				lineOffsets[i] = currentOffset
				currentOffset += lineCrossSizes[i]
				if i < len(lines)-1 {
					currentOffset += rowGap
				}
			}
		}

		// Update total cross size if stretch was applied
		if alignContent == AlignContentStretch {
			totalCrossSize = crossSize
		}
	} else {
		// Single line - no align-content needed
		for i := range lines {
			lineOffsets[i] = 0
		}
	}

	// Handle flex-wrap: wrap-reverse - reverse line order AFTER calculating align-content
	// For wrap-reverse, lines are visually reversed (last line becomes first visually)
	// We reverse the lines and recalculate offsets so the last line is at the top
	if wrapReverse && len(lines) > 1 {
		// Reverse the order of lines and their corresponding data
		for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
			lines[i], lines[j] = lines[j], lines[i]
			lineCrossSizes[i], lineCrossSizes[j] = lineCrossSizes[j], lineCrossSizes[i]
		}
		// Recalculate offsets for reversed visual order
		// The last line (now first visually) should be at offset 0
		// Recalculate total cross size for reversed order (using reversed lineCrossSizes)
		totalReversedCrossSize := 0.0
		for i := range lineCrossSizes {
			totalReversedCrossSize += lineCrossSizes[i]
			if i < len(lineCrossSizes)-1 {
				totalReversedCrossSize += rowGap
			}
		}
		// Recalculate align-content offsets for reversed order
		freeCrossSpace := crossSize - totalReversedCrossSize
		if freeCrossSpace < 0 {
			freeCrossSpace = 0
		}
		alignContent := node.Style.AlignContent
		if alignContent == 0 {
			alignContent = AlignContentStretch
		}
		var startOffset float64
		switch alignContent {
		case AlignContentFlexStart:
			startOffset = 0
		case AlignContentFlexEnd:
			startOffset = freeCrossSpace
		case AlignContentCenter:
			startOffset = freeCrossSpace / 2
		case AlignContentStretch:
			if freeCrossSpace > 0 {
				extraPerLine := freeCrossSpace / float64(len(lines))
				for i := range lineCrossSizes {
					lineCrossSizes[i] += extraPerLine
				}
			}
			startOffset = 0
		case AlignContentSpaceBetween:
			if len(lines) > 1 {
				spaceBetween := freeCrossSpace / float64(len(lines)-1)
				currentOffset := 0.0
				for i := range lines {
					lineOffsets[i] = currentOffset
					currentOffset += lineCrossSizes[i]
					if i < len(lines)-1 {
						currentOffset += rowGap + spaceBetween
					}
				}
				startOffset = 0 // Set but won't be used
			} else {
				startOffset = 0
			}
		case AlignContentSpaceAround:
			if len(lines) > 0 {
				spaceAround := freeCrossSpace / float64(len(lines))
				currentOffset := spaceAround / 2
				for i := range lines {
					lineOffsets[i] = currentOffset
					currentOffset += lineCrossSizes[i]
					if i < len(lines)-1 {
						currentOffset += rowGap + spaceAround
					}
				}
				startOffset = 0 // Set but won't be used
			} else {
				startOffset = 0
			}
		default:
			startOffset = 0
		}
		// Calculate offsets from startOffset for non-space-between/space-around
		if alignContent != AlignContentSpaceBetween && alignContent != AlignContentSpaceAround {
			currentOffset := startOffset
			for i := range lines {
				lineOffsets[i] = currentOffset
				currentOffset += lineCrossSizes[i]
				if i < len(lines)-1 {
					currentOffset += rowGap
				}
			}
		}
		// Update total cross size for reversed order
		if alignContent == AlignContentStretch {
			totalCrossSize = crossSize
		} else {
			totalCrossSize = totalReversedCrossSize
		}
	}

	// Step 5: Second pass - apply align-content offsets and handle reverse direction
	for lineIdx, line := range lines {
		// Handle flex-direction reverse - reverse items in line
		// For reverse, we reverse the items and then position from the end
		isReverse := node.Style.FlexDirection == FlexDirectionRowReverse || node.Style.FlexDirection == FlexDirectionColumnReverse
		if isReverse {
			// Reverse the order of items in this line
			for i, j := 0, len(line)-1; i < j; i, j = i+1, j-1 {
				line[i], line[j] = line[j], line[i]
			}
		}

		// Get the updated line cross size (may have been stretched by align-content)
		lineCrossSize := lineCrossSizes[lineIdx]
		lineStartCrossOffset := lineOffsets[lineIdx]

		// Note: align-items stretch will be applied per-item below when setting rects
		// This ensures we use the correct crossSize/lineCrossSize values

		// Re-align items in cross axis with updated line cross size
		alignmentCrossSize := crossSize
		if len(lines) == 1 {
			alignmentCrossSize = crossSize
		} else {
			// For multi-line, use line cross size for alignment
			alignmentCrossSize = lineCrossSize
		}

		for _, item := range line {
			crossOffset := 0.0
			// Ensure item.crossSize is set correctly for stretch before calculating offsets
			// This must happen before we use item.crossSize in alignment calculations
			if node.Style.AlignItems == AlignItemsStretch {
				// For single-line, use crossSize; for multi-line, use lineCrossSize
				if len(lines) == 1 {
					item.crossSize = crossSize - item.crossMarginStart - item.crossMarginEnd
				} else {
					item.crossSize = lineCrossSize - item.crossMarginStart - item.crossMarginEnd
				}
				if item.crossSize < 0 {
					item.crossSize = 0
				}
			}

			itemCrossSizeWithMargins := item.crossSize + item.crossMarginStart + item.crossMarginEnd
			switch node.Style.AlignItems {
			case AlignItemsFlexStart:
				crossOffset = item.crossMarginStart
			case AlignItemsFlexEnd:
				crossOffset = alignmentCrossSize - item.crossSize - item.crossMarginEnd
			case AlignItemsCenter:
				crossOffset = (alignmentCrossSize-itemCrossSizeWithMargins)/2 + item.crossMarginStart
			case AlignItemsStretch:
				crossOffset = item.crossMarginStart
			default:
				crossOffset = item.crossMarginStart
			}

			// Set initial rect (will be updated by justify-content or reverse positioning)
			// For reverse, we'll position from the end, so X/Y will be set there
			// Safeguard: if mainSize/crossSize is 0 but explicit dimensions are set, use them
			// This is a last-resort fix for cases where LayoutBlock returned 0
			// We also update item.mainSize/crossSize if they're 0, so justify-content calculations work correctly
			rectWidth := item.mainSize
			rectHeight := item.crossSize

			// Apply align-items stretch if needed (for cross-size)
			// Zero value is stretch (CSS Flexbox default)
			if node.Style.AlignItems == AlignItemsStretch {
				if isRow {
					// For row direction, cross-size is height
					if len(lines) == 1 {
						rectHeight = crossSize - item.crossMarginStart - item.crossMarginEnd
					} else {
						rectHeight = lineCrossSize - item.crossMarginStart - item.crossMarginEnd
					}
					if rectHeight < 0 {
						rectHeight = 0
					}
					item.crossSize = rectHeight
				} else {
					// For column direction, cross-size is width
					if len(lines) == 1 {
						rectWidth = crossSize - item.crossMarginStart - item.crossMarginEnd
					} else {
						rectWidth = lineCrossSize - item.crossMarginStart - item.crossMarginEnd
					}
					if rectWidth < 0 {
						rectWidth = 0
					}
					item.crossSize = rectWidth
				}
			}

			if isRow {
				if rectWidth == 0 && item.node.Style.Width >= 0 {
					rectWidth = item.node.Style.Width
					// Update mainSize so justify-content calculations use correct size
					item.mainSize = item.node.Style.Width
				}
				// For cross-size (height in row), if it's 0 and we have explicit height, use it
				if rectHeight == 0 && item.node.Style.Height >= 0 {
					rectHeight = item.node.Style.Height
					item.crossSize = item.node.Style.Height
				}
			} else {
				if rectHeight == 0 && item.node.Style.Height >= 0 {
					rectHeight = item.node.Style.Height
					item.mainSize = item.node.Style.Height
				}
				if rectWidth == 0 && item.node.Style.Width >= 0 {
					rectWidth = item.node.Style.Width
					item.crossSize = item.node.Style.Width
				}
			}
			if !isReverse {
				if isRow {
					item.node.Rect = Rect{
						X:      node.Style.Padding.Left + node.Style.Border.Left,
						Y:      node.Style.Padding.Top + node.Style.Border.Top + lineStartCrossOffset + crossOffset,
						Width:  rectWidth,
						Height: rectHeight,
					}
				} else {
					item.node.Rect = Rect{
						X:      node.Style.Padding.Left + node.Style.Border.Left + lineStartCrossOffset + crossOffset,
						Y:      node.Style.Padding.Top + node.Style.Border.Top,
						Width:  rectWidth,
						Height: rectHeight,
					}
				}
			} else {
				// For reverse, set cross-axis position now, main-axis will be set below
				if isRow {
					item.node.Rect = Rect{
						X:      0, // Will be set by reverse positioning
						Y:      node.Style.Padding.Top + node.Style.Border.Top + lineStartCrossOffset + crossOffset,
						Width:  rectWidth,
						Height: rectHeight,
					}
				} else {
					item.node.Rect = Rect{
						X:      node.Style.Padding.Left + node.Style.Border.Left + lineStartCrossOffset + crossOffset,
						Y:      0, // Will be set by reverse positioning
						Width:  rectWidth,
						Height: rectHeight,
					}
				}
			}
		}

		// Ensure item.mainSize is set correctly before justify-content calculation
		// This is needed because justifyContentWithGap uses item.mainSize
		for _, item := range line {
			if isRow {
				if item.mainSize == 0 && item.node.Style.Width >= 0 {
					item.mainSize = item.node.Style.Width
				}
			} else {
				if item.mainSize == 0 && item.node.Style.Height >= 0 {
					item.mainSize = item.node.Style.Height
				}
			}
		}

		// Apply justify-content with gap support
		// For reverse direction, position items from the end (right for row, bottom for column)
		if isReverse {
			// Calculate total line size including gaps
			totalLineSize := 0.0
			for i, item := range line {
				totalLineSize += item.mainSize + item.mainMarginStart + item.mainMarginEnd
				if i < len(line)-1 {
					totalLineSize += columnGap
				}
			}
			// Position from the end (right to left for row, bottom to top for column)
			// Items are already reversed in the array, so position them sequentially from right
			// The last item in the reversed array (which is first in original) should be at rightmost
			currentPos := mainSize
			// Work backwards through the reversed array (which is forward through original)
			for i := len(line) - 1; i >= 0; i-- {
				item := line[i]
				currentPos -= item.mainSize + item.mainMarginEnd
				if isRow {
					item.node.Rect.X = node.Style.Padding.Left + node.Style.Border.Left + currentPos
				} else {
					item.node.Rect.Y = node.Style.Padding.Top + node.Style.Border.Top + currentPos
				}
				currentPos -= item.mainMarginStart
				if i > 0 {
					currentPos -= columnGap
				}
			}
		} else {
			justifyContentWithGap(node.Style.JustifyContent, line, 0, mainSize, isRow, columnGap)
		}

		// Calculate this line's main extent (including margins and gaps)
		// Note: item.node.Rect.X/Y are absolute positions including padding/border
		// We need to calculate the extent relative to the content area start
		lineMainSize := 0.0
		contentAreaStart := 0.0
		if isRow {
			contentAreaStart = node.Style.Padding.Left + node.Style.Border.Left
		} else {
			contentAreaStart = node.Style.Padding.Top + node.Style.Border.Top
		}
		for _, item := range line {
			if isRow {
				itemEnd := item.node.Rect.X + item.node.Rect.Width + item.mainMarginEnd
				// Convert to content-area relative
				itemEndRelative := itemEnd - contentAreaStart
				if itemEndRelative > lineMainSize {
					lineMainSize = itemEndRelative
				}
			} else {
				itemEnd := item.node.Rect.Y + item.node.Rect.Height + item.mainMarginEnd
				// Convert to content-area relative
				itemEndRelative := itemEnd - contentAreaStart
				if itemEndRelative > lineMainSize {
					lineMainSize = itemEndRelative
				}
			}
		}

		// Track maximum line main size (for container main dimension)
		if lineMainSize > maxLineMainSize {
			maxLineMainSize = lineMainSize
		}
	}

	// Calculate container size
	// Main dimension = max line main extent (not sum)
	// Cross dimension = sum of line cross sizes
	var containerSize Size
	if isRow {
		containerSize = Size{
			Width:  maxLineMainSize + horizontalPadding + horizontalBorder,
			Height: totalCrossSize + verticalPadding + verticalBorder,
		}
	} else {
		containerSize = Size{
			Width:  totalCrossSize + horizontalPadding + horizontalBorder,
			Height: maxLineMainSize + verticalPadding + verticalBorder,
		}
	}

	// Constrain size and apply to Rect
	// CRITICAL: node.Rect must respect constraints to match the returned Size
	constrainedSize := constraints.Constrain(containerSize)

	// Set container rect
	node.Rect = Rect{
		X:      0,
		Y:      0,
		Width:  constrainedSize.Width,
		Height: constrainedSize.Height,
	}

	return constrainedSize
}

type flexItem struct {
	node             *Node
	mainSize         float64
	crossSize        float64
	baseSize         float64
	flexGrow         float64
	flexShrink       float64
	flexBasis        float64
	mainMarginStart  float64
	mainMarginEnd    float64
	crossMarginStart float64
	crossMarginEnd   float64
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

// justifyContentWithGap applies justify-content with gap support
func justifyContentWithGap(justify JustifyContent, line []*flexItem, startOffset, containerSize float64, isRow bool, gap float64) {
	if len(line) == 0 {
		return
	}

	// Calculate total size of items in main axis (including margins)
	// Use item.mainSize instead of item.node.Rect.Width/Height because mainSize is the
	// actual flex-calculated size, while Rect.Width/Height might be set to explicit dimensions
	// If mainSize is 0, fall back to rect width/height as a last resort
	totalItemSize := 0.0
	for _, item := range line {
		itemSize := item.mainSize
		if itemSize == 0 {
			// Fallback to rect size if mainSize is 0
			if isRow {
				itemSize = item.node.Rect.Width
			} else {
				itemSize = item.node.Rect.Height
			}
		}
		totalItemSize += itemSize + item.mainMarginStart + item.mainMarginEnd
	}
	// Add gaps between items
	if len(line) > 1 {
		totalItemSize += gap * float64(len(line)-1)
	}

	// Free space is the container's main size minus total item size (including gaps)
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
			// Distribute free space between items (not including gap)
			spaceBetween := freeSpace / float64(len(line)-1)
			currentPos := startOffset
			for _, item := range line {
				if isRow {
					item.node.Rect.X += currentPos + item.mainMarginStart
					currentPos += item.mainSize + item.mainMarginStart + item.mainMarginEnd + gap + spaceBetween
				} else {
					item.node.Rect.Y += currentPos + item.mainMarginStart
					currentPos += item.mainSize + item.mainMarginStart + item.mainMarginEnd + gap + spaceBetween
				}
			}
			return
		}
		offset = 0
	case JustifyContentSpaceAround:
		if len(line) > 0 {
			// Distribute free space around items
			spaceAround := freeSpace / float64(len(line))
			offset = spaceAround / 2
		}
	case JustifyContentSpaceEvenly:
		if len(line) > 0 {
			// Distribute free space evenly (including edges)
			spaceEvenly := freeSpace / float64(len(line)+1)
			offset = spaceEvenly
		}
	}

	// Apply offset (accounting for margins, padding, and gap)
	// Note: For row direction, we only modify X. For column direction, we only modify Y.
	// The cross-axis position (Y for row, X for column) is set separately and should not be modified here.
	currentPos := startOffset + offset
	for i, item := range line {
		if isRow {
			// Row direction: modify X (main axis), preserve Y (cross axis)
			item.node.Rect.X += currentPos + item.mainMarginStart
			currentPos += item.mainSize + item.mainMarginStart + item.mainMarginEnd
			if i < len(line)-1 {
				currentPos += gap
			}
		} else {
			// Column direction: modify Y (main axis), preserve X (cross axis)
			item.node.Rect.Y += currentPos + item.mainMarginStart
			currentPos += item.mainSize + item.mainMarginStart + item.mainMarginEnd
			if i < len(line)-1 {
				currentPos += gap
			}
		}
	}
}

// justifyContent is kept for backward compatibility but now calls justifyContentWithGap with 0 gap
func justifyContent(justify JustifyContent, line []*flexItem, startOffset, containerSize float64, isRow bool) {
	justifyContentWithGap(justify, line, startOffset, containerSize, isRow, 0)
}
