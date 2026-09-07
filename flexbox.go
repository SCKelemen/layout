package layout

import "math"

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
func LayoutFlexbox(node *Node, constraints Constraints, ctx *LayoutContext) Size {
	if node.Style.Display != DisplayFlex {
		// If not flex, delegate to block layout
		return LayoutBlock(node, constraints, ctx)
	}

	// Get current font size for Length resolution
	fontSize := getCurrentFontSize(node, ctx)

	// §9.2: Line Length Determination - Setup and initial measurement
	setup := flexboxDetermineLineLength(node, constraints, ctx)

	// Handle empty container: it still has its explicit (or available) width
	// and its explicit height, plus padding and border. Never propagate an
	// unbounded content size into the result.
	if len(node.Children) == 0 {
		emptyWidth := 0.0
		if setup.contentWidth < Unbounded {
			emptyWidth = setup.contentWidth
		}
		emptyHeight := 0.0
		if node.Style.Height.Value > 0 && setup.contentHeight < Unbounded {
			emptyHeight = setup.contentHeight
		}
		resultSize := Size{
			Width:  emptyWidth + setup.horizontalPadding + setup.horizontalBorder,
			Height: emptyHeight + setup.verticalPadding + setup.verticalBorder,
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
	flexItems := flexboxMeasureItems(node, setup, ctx)

	// Normalize align-items: zero value is stretch (CSS Flexbox default)
	alignItems := node.Style.AlignItems
	if alignItems == 0 {
		alignItems = AlignItemsStretch
	}

	// Get gap values (resolve Length to pixels)
	rowGap := ResolveLength(node.Style.FlexRowGap, ctx, fontSize)
	if rowGap == 0 {
		rowGap = ResolveLength(node.Style.FlexGap, ctx, fontSize)
	}
	columnGap := ResolveLength(node.Style.FlexColumnGap, ctx, fontSize)
	if columnGap == 0 {
		columnGap = ResolveLength(node.Style.FlexGap, ctx, fontSize)
	}

	// Map row-gap/column-gap onto the flex axes. row-gap separates rows
	// (block axis) and column-gap separates columns (inline axis), so in a row
	// flex container the gap between items is column-gap and the gap between
	// lines is row-gap; in a column flex container it is the other way around.
	// CSS Box Alignment Level 3 §8.3 / CSS Flexbox Level 1 §8.
	// https://www.w3.org/TR/css-align-3/#column-row-gap
	mainGap, crossGap := columnGap, rowGap
	if !setup.isRow {
		mainGap, crossGap = rowGap, columnGap
	}

	// Step 2: Calculate flex lines (for wrapping), including the main-axis gap
	hasWrap := node.Style.FlexWrap == FlexWrapWrap || node.Style.FlexWrap == FlexWrapWrapReverse
	lines := calculateFlexLines(flexItems, setup.mainSize, hasWrap, mainGap)

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
			totalCrossSize += crossGap
		}
	}

	// §10.4: Aligning with align-content - distribute lines along cross axis.
	// align-content is resolved against the cross-start edge; for wrap-reverse
	// the cross-start edge is flipped by flexboxHandleWrapReverse, which mirrors
	// the computed offsets. The value itself must not be swapped here as well,
	// or the two inversions cancel out.
	// https://www.w3.org/TR/css-flexbox-1/#flex-wrap-property
	lineOffsets, totalCrossSize := flexboxAlignWithAlignContent(
		node, lines, lineCrossSizes, setup.crossSize, totalCrossSize, crossGap, setup.hasExplicitCrossSize)

	// §9.2: Line Length Determination - Handle flex-wrap: wrap-reverse
	// For wrap-reverse, we reverse line order and mirror offsets (no need for originalLineCrossSizes)
	lineOffsets, totalCrossSize = flexboxHandleWrapReverse(
		node, lines, lineCrossSizes, lineOffsets, nil, // originalLineCrossSizes no longer needed
		setup.crossSize, totalCrossSize, crossGap, setup.hasExplicitCrossSize)

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
		flexboxAlignmentCrossAxis(node, line, setup, alignItems, lineCrossSize, lineStartCrossOffset, alignmentCrossSize, ctx)

		// §9.5: Main-Axis Alignment - position items along main axis
		lineMainSize := flexboxAlignmentMainAxis(
			node, line, setup, lineCrossSize, lineStartCrossOffset,
			mainGap, setup.mainSize, isReverse, ctx)

		// Re-layout container items whose final size differs from the size they
		// were measured at. This must happen AFTER both cross and main axis
		// alignment, so items have their final Rect. Flexing (grow or shrink) and
		// stretching change an item's size, and its own children must be laid
		// out against that final size rather than the measurement constraints.
		for _, item := range line {
			flexboxRelayoutResizedItem(item, ctx)
		}

		// Track maximum line main size (for container main dimension)
		if lineMainSize > maxLineMainSize {
			maxLineMainSize = lineMainSize
		}
	}

	// Step 7: Calculate container size
	// Main dimension = max line main extent (not sum)
	// Cross dimension = use explicit cross size if available, otherwise sum of line cross sizes
	var containerSize Size
	if setup.isMainHorizontal {
		crossDimension := totalCrossSize
		if setup.hasExplicitCrossSize {
			crossDimension = setup.crossSize
		}
		containerSize = Size{
			Width:  maxLineMainSize + setup.horizontalPadding + setup.horizontalBorder,
			Height: crossDimension + setup.verticalPadding + setup.verticalBorder,
		}
	} else {
		crossDimension := totalCrossSize
		if setup.hasExplicitCrossSize {
			crossDimension = setup.crossSize
		}
		containerSize = Size{
			Width:  crossDimension + setup.horizontalPadding + setup.horizontalBorder,
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

	// hypotheticalMainSize is the flex base size clamped by the item's
	// min/max main size (CSS Flexbox §9.3 step 3).
	hypotheticalMainSize float64
	// minMain/maxMain are the resolved min/max main size; maxMain is
	// Unbounded when unset.
	minMain float64
	maxMain float64
	// frozen marks an item whose main size is final in the §9.7 loop.
	frozen bool
	// hasExplicitCrossSize is true when the item's cross size property is not
	// auto, in which case align-self: stretch does not apply (§9.4 step 11).
	hasExplicitCrossSize bool
	// measuredWidth/measuredHeight are the item's Rect dimensions after the
	// initial measurement pass, used to decide whether to lay it out again.
	measuredWidth  float64
	measuredHeight float64
}

// flexboxRelayoutResizedItem lays out an item again with tight constraints
// when its final flexed/stretched size differs from the size it was measured
// at, so its contents are positioned against the final size.
//
// Container items (flex, grid, block) need this so their own children are laid
// out against the final size. Text items need it so their lines re-wrap: a
// text item's flex base size is its max-content size (CSS Flexbox §9.2 step
// 3E), and once flexing has shrunk or grown it the line boxes must be
// recomputed at the used main size.
// https://www.w3.org/TR/css-flexbox-1/#algo-main-item
//
// Positions set by the parent's alignment are preserved. Childless non-text
// nodes are left alone; unbounded sizes are never used as constraints.
func flexboxRelayoutResizedItem(item *flexItem, ctx *LayoutContext) {
	child := item.node
	switch child.Style.Display {
	case DisplayInlineText:
		// Text leaves have no children but must re-wrap at the final size.
	case DisplayFlex, DisplayGrid, DisplayBlock:
		if len(child.Children) == 0 {
			return
		}
	default:
		return
	}
	finalWidth := child.Rect.Width
	finalHeight := child.Rect.Height
	if finalWidth >= Unbounded || finalHeight >= Unbounded || finalWidth < 0 || finalHeight < 0 {
		return
	}
	const epsilon = 1e-6
	if math.Abs(finalWidth-item.measuredWidth) < epsilon && math.Abs(finalHeight-item.measuredHeight) < epsilon {
		return
	}

	savedX := child.Rect.X
	savedY := child.Rect.Y
	Layout(child, Tight(finalWidth, finalHeight), ctx)
	child.Rect.X = savedX
	child.Rect.Y = savedY
	child.Rect.Width = finalWidth
	child.Rect.Height = finalHeight
	item.measuredWidth = finalWidth
	item.measuredHeight = finalHeight
}

// calculateFlexLines collects items into flex lines.
//
// Implementation of CSS Flexible Box Layout Module Level 1 §9.3 step 5: a
// line is broken before the first item whose outer hypothetical main size,
// plus the main-axis gap that precedes it, would overflow the line. With an
// indefinite container main size every item goes on one line.
// https://www.w3.org/TR/css-flexbox-1/#algo-line-break
func calculateFlexLines(items []*flexItem, containerMainSize float64, wrap bool, gap float64) [][]*flexItem {
	if !wrap || containerMainSize >= Unbounded {
		return [][]*flexItem{items}
	}

	lines := [][]*flexItem{}
	currentLine := []*flexItem{}
	currentLineSize := 0.0

	for _, item := range items {
		// Include margins in item size for wrapping calculation
		itemSize := item.hypotheticalMainSize + item.mainMarginStart + item.mainMarginEnd
		if len(currentLine) > 0 {
			if currentLineSize+gap+itemSize > containerMainSize {
				lines = append(lines, currentLine)
				currentLine = []*flexItem{}
				currentLineSize = 0
			} else {
				currentLineSize += gap
			}
		}
		currentLine = append(currentLine, item)
		currentLineSize += itemSize
	}

	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}

	return lines
}

// justifyContentWithGap positions the items of a line along the main axis.
//
// Implementation of CSS Flexible Box Layout Module Level 1 §9.5 / §10.2
// (justify-content). space-between, space-around and space-evenly insert extra
// space between consecutive items in addition to the gap; with negative free
// space they fall back to flex-start (space-between) or center (space-around,
// space-evenly).
// https://www.w3.org/TR/css-flexbox-1/#justify-content-property
func justifyContentWithGap(justify JustifyContent, line []*flexItem, startOffset, containerSize float64, isMainHorizontal bool, gap float64, writingMode WritingMode) {
	if len(line) == 0 {
		return
	}

	// Total outer size of the items in the main axis, including gaps.
	// item.mainSize is the resolved flex main size (§9.7).
	totalItemSize := 0.0
	for _, item := range line {
		totalItemSize += item.mainSize + item.mainMarginStart + item.mainMarginEnd
	}
	if len(line) > 1 {
		totalItemSize += gap * float64(len(line)-1)
	}

	// An indefinite container size means there is no free space to distribute.
	if containerSize >= Unbounded {
		containerSize = totalItemSize
	}
	freeSpace := containerSize - totalItemSize

	n := float64(len(line))
	var offset, between float64
	switch justify {
	case JustifyContentFlexStart:
		offset = 0
	case JustifyContentFlexEnd:
		offset = freeSpace
	case JustifyContentCenter:
		offset = freeSpace / 2
	case JustifyContentSpaceBetween:
		if freeSpace > 0 && len(line) > 1 {
			between = freeSpace / (n - 1)
		}
	case JustifyContentSpaceAround:
		if freeSpace > 0 {
			between = freeSpace / n
			offset = between / 2
		} else {
			offset = freeSpace / 2
		}
	case JustifyContentSpaceEvenly:
		if freeSpace > 0 {
			between = freeSpace / (n + 1)
			offset = between
		} else {
			offset = freeSpace / 2
		}
	}

	// Apply offset (accounting for margins, padding, and gap)
	// Note: For main axis horizontal, we modify X. For main axis vertical, we modify Y.
	// The cross-axis position (Y for horizontal main, X for vertical main) is set separately and should not be modified here.
	//
	// For right-to-left writing modes (vertical-rl, sideways-rl), when the main axis is horizontal,
	// items are positioned from right to left instead of left to right.
	isRightToLeft := writingMode.IsRightToLeft() && isMainHorizontal

	currentPos := startOffset + offset
	for i, item := range line {
		if isMainHorizontal {
			// Main axis horizontal: modify X (main axis), preserve Y (cross axis)
			if isRightToLeft {
				// Right-to-left: position from right edge moving leftward
				item.node.Rect.X += startOffset + containerSize - currentPos - item.mainMarginStart - item.mainSize
			} else {
				// Left-to-right: position from left edge moving rightward
				item.node.Rect.X += currentPos + item.mainMarginStart
			}
		} else {
			// Main axis vertical: modify Y (main axis), preserve X (cross axis)
			item.node.Rect.Y += currentPos + item.mainMarginStart
		}
		currentPos += item.mainSize + item.mainMarginStart + item.mainMarginEnd
		if i < len(line)-1 {
			currentPos += gap + between
		}
	}
}

// justifyContent is kept for backward compatibility but now calls justifyContentWithGap with 0 gap
func justifyContent(justify JustifyContent, line []*flexItem, startOffset, containerSize float64, isMainHorizontal bool, writingMode WritingMode) {
	justifyContentWithGap(justify, line, startOffset, containerSize, isMainHorizontal, 0, writingMode)
}
