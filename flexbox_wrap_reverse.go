package layout

// flexboxHandleWrapReverse handles flex-wrap: wrap-reverse by reversing line order and recalculating offsets.
//
// Algorithm based on CSS Flexible Box Layout Module Level 1:
// - §9.2: Line Length Determination (wrap-reverse)
//
// See: https://www.w3.org/TR/css-flexbox-1/#flex-wrap-property
func flexboxHandleWrapReverse(
	node *Node,
	lines [][]*flexItem,
	lineCrossSizes []float64,
	lineOffsets []float64,
	originalLineCrossSizes []float64,
	crossSize float64,
	totalCrossSize float64,
	rowGap float64,
	hasExplicitCrossSize bool,
) ([]float64, float64) {
	wrapReverse := node.Style.FlexWrap == FlexWrapWrapReverse
	if !wrapReverse || len(lines) <= 1 || !hasExplicitCrossSize {
		return lineOffsets, totalCrossSize
	}

	// Reverse the order of lines and their corresponding data
	// Note: lineOffsets will be recalculated below, so we don't need to reverse them
	// IMPORTANT: Reset lineCrossSizes to original (unstretched) values BEFORE reversing
	// This ensures we're working with the correct base sizes
	copy(lineCrossSizes, originalLineCrossSizes)

	// Now reverse everything
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
		lineCrossSizes[i], lineCrossSizes[j] = lineCrossSizes[j], lineCrossSizes[i]
		originalLineCrossSizes[i], originalLineCrossSizes[j] = originalLineCrossSizes[j], originalLineCrossSizes[i]
	}

	// Recalculate offsets for reversed visual order
	// The last line (now first visually) should be at offset 0
	// Use original (unstretched) line cross sizes to avoid double-stretching
	// We'll apply stretching again after reversing if needed
	// Note: originalLineCrossSizes and lineCrossSizes are now in reversed order
	totalReversedCrossSize := 0.0
	for i := range originalLineCrossSizes {
		totalReversedCrossSize += originalLineCrossSizes[i]
		if i < len(originalLineCrossSizes)-1 {
			totalReversedCrossSize += rowGap
		}
	}
	// lineCrossSizes is already reset and reversed above, so we don't need to copy again

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

	// Reset lineOffsets before recalculating for reversed order
	for i := range lineOffsets {
		lineOffsets[i] = 0
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

	return lineOffsets, totalCrossSize
}
