package layout

// LayoutFlexbox performs flexbox layout on a node.
//
// Algorithm based on CSS Flexible Box Layout Module Level 1:
// - §9: Flex Layout Algorithm
//   - §9.2: Line Length Determination
//   - §9.3: Main Size Determination
//   - §9.4: Cross Size Determination
//   - §9.5: Main-Axis Alignment
//   - §9.6: Cross-Axis Alignment
//
// - §10: Alignment
//   - §10.1: Aligning with auto margins
//   - §10.2: Aligning with justify-content
//   - §10.3: Aligning with align-items
//   - §10.4: Aligning with align-content
//
// See: https://www.w3.org/TR/css-flexbox-1/
func LayoutFlexbox(node *Node, constraints Constraints) Size {
	if node.Style.Display != DisplayFlex {
		// If not flex, delegate to block layout
		return LayoutBlock(node, constraints)
	}

	// §9.2: Line Length Determination - Setup and initial measurement
	setup := flexboxDetermineLineLength(node, constraints)

	// Handle empty container
	if len(node.Children) == 0 {
		resultSize := Size{
			Width:  setup.horizontalPadding + setup.horizontalBorder,
			Height: setup.verticalPadding + setup.verticalBorder,
		}
		node.Rect = Rect{
			X:      0,
			Y:      0,
			Width:  resultSize.Width,
			Height: resultSize.Height,
		}
		return constraints.Constrain(resultSize)
	}

	// §9.2: Line Length Determination - Measure items
	flexItems := flexboxMeasureItems(node, setup)

	// Normalize align-items: zero value is stretch (CSS Flexbox default)
	alignItems := node.Style.AlignItems
	if alignItems == 0 {
		alignItems = AlignItemsStretch
	}

	// Step 2: Calculate flex lines (for wrapping)
	hasWrap := node.Style.FlexWrap == FlexWrapWrap || node.Style.FlexWrap == FlexWrapWrapReverse
	lines := calculateFlexLines(flexItems, setup.mainSize, hasWrap)

	// Get gap values
	rowGap := node.Style.FlexRowGap
	if rowGap == 0 {
		rowGap = node.Style.FlexGap
	}
	columnGap := node.Style.FlexColumnGap
	if columnGap == 0 {
		columnGap = node.Style.FlexGap
	}

	// §9.3: Main Size Determination and §9.4: Cross Size Determination
	lineCrossSizes := make([]float64, len(lines))
	totalCrossSize := 0.0

	for lineIdx, line := range lines {
		// §9.3: Main Size Determination - determine main sizes using flex grow/shrink
		flexboxDetermineMainSize(line, setup.mainSize, setup.hasExplicitMainSize)

		// §9.4: Cross Size Determination - determine line cross size
		isSingleLine := len(lines) == 1
		lineCrossSize := flexboxDetermineCrossSize(line, setup.crossSize, alignItems, setup.hasExplicitCrossSize, isSingleLine)

		// Store line cross size for align-content calculation
		lineCrossSizes[lineIdx] = lineCrossSize
		totalCrossSize += lineCrossSize
		if lineIdx < len(lines)-1 {
			totalCrossSize += rowGap
		}
	}

	// §10.4: Aligning with align-content - distribute lines along cross axis
	lineOffsets, totalCrossSize := flexboxAlignWithAlignContent(
		node, lines, lineCrossSizes, setup.crossSize, totalCrossSize, rowGap, setup.hasExplicitCrossSize)

	// §9.2: Line Length Determination - Handle flex-wrap: wrap-reverse
	// For wrap-reverse, we reverse line order and mirror offsets (no need for originalLineCrossSizes)
	lineOffsets, totalCrossSize = flexboxHandleWrapReverse(
		node, lines, lineCrossSizes, lineOffsets, nil, // originalLineCrossSizes no longer needed
		setup.crossSize, totalCrossSize, rowGap, setup.hasExplicitCrossSize)

	// Step 6: Second pass - position items using justify-content and align-items
	maxLineMainSize := 0.0
	isReverse := node.Style.FlexDirection == FlexDirectionRowReverse || node.Style.FlexDirection == FlexDirectionColumnReverse
	for lineIdx, line := range lines {
		// Get the updated line cross size (may have been stretched by align-content)
		lineCrossSize := lineCrossSizes[lineIdx]
		lineStartCrossOffset := lineOffsets[lineIdx]

		// Determine alignment cross size for this line
		// For single-line with explicit cross size, use crossSize for alignment (container's cross size)
		// For multi-line or auto-sized containers, use lineCrossSize (content-driven)
		var alignmentCrossSize float64
		if len(lines) == 1 && setup.hasExplicitCrossSize {
			// Single-line with explicit cross size: align within container's cross size
			alignmentCrossSize = setup.crossSize
		} else {
			// Multi-line or auto-sized: use line's resolved cross size
			alignmentCrossSize = lineCrossSize
		}

		// §9.6: Cross-Axis Alignment - align items along cross axis
		flexboxAlignmentCrossAxis(node, line, setup, alignItems, lineCrossSize, lineStartCrossOffset, alignmentCrossSize)

		// §9.5: Main-Axis Alignment - position items along main axis
		lineMainSize := flexboxAlignmentMainAxis(
			node, line, setup, lineCrossSize, lineStartCrossOffset,
			columnGap, setup.mainSize, isReverse)

		// Track maximum line main size (for container main dimension)
		if lineMainSize > maxLineMainSize {
			maxLineMainSize = lineMainSize
		}
	}

	// Step 7: Calculate container size
	// Main dimension = max line main extent (not sum)
	// Cross dimension = sum of line cross sizes
	var containerSize Size
	if setup.isRow {
		containerSize = Size{
			Width:  maxLineMainSize + setup.horizontalPadding + setup.horizontalBorder,
			Height: totalCrossSize + setup.verticalPadding + setup.verticalBorder,
		}
	} else {
		containerSize = Size{
			Width:  totalCrossSize + setup.horizontalPadding + setup.horizontalBorder,
			Height: maxLineMainSize + setup.verticalPadding + setup.verticalBorder,
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
