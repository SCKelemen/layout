package layout

// flexboxAlignWithAlignContent applies align-content to distribute lines along the cross axis.
//
// Algorithm based on CSS Flexible Box Layout Module Level 1:
// - §10.4: Aligning with align-content
//
// See: https://www.w3.org/TR/css-flexbox-1/#align-content-property
func flexboxAlignWithAlignContent(
	node *Node,
	lines [][]*flexItem,
	lineCrossSizes []float64,
	crossSize float64,
	totalCrossSize float64,
	rowGap float64,
	hasExplicitCrossSize bool,
) ([]float64, float64) {
	lineOffsets := make([]float64, len(lines))

	// align-content only has effect when there are multiple lines, wrapping is enabled,
	// and the container's cross size is definite (not auto/unbounded).
	// This prevents align-content: stretch from zeroing out content when crossSize is 0 or Unbounded
	hasWrap := node.Style.FlexWrap == FlexWrapWrap || node.Style.FlexWrap == FlexWrapWrapReverse
	if len(lines) > 1 && hasWrap && hasExplicitCrossSize && crossSize < Unbounded {
		alignContent := node.Style.AlignContent
		// Zero value is AlignContentStretch (CSS default), no need to check

		// Calculate free cross space. It may be negative: center and flex-end
		// then produce negative offsets (lines overflow both/one edge), while
		// space-between falls back to flex-start, space-around to center, and
		// stretch distributes nothing.
		// https://www.w3.org/TR/css-flexbox-1/#align-content-property
		freeCrossSpace := crossSize - totalCrossSize

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
				if freeCrossSpace < 0 {
					// Negative free space: behave as flex-start.
					spaceBetween = 0
				}
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
				if freeCrossSpace < 0 {
					// Negative free space: behave as center.
					spaceAround = 0
					currentOffset = freeCrossSpace / 2
				}
				for i := range lines {
					lineOffsets[i] = currentOffset
					currentOffset += lineCrossSizes[i]
					if i < len(lines)-1 {
						currentOffset += rowGap + spaceAround
					}
				}
			}
		case AlignContentSpaceEvenly:
			// Lines are evenly distributed so that the spacing between any two
			// adjacent lines, before the first line, and after the last line is
			// the same (CSS Box Alignment Level 3 §6.2). With negative free
			// space the fallback alignment is center (css-align-3 §6.1.3).
			// https://www.w3.org/TR/css-align-3/#valdef-align-content-space-evenly
			// https://www.w3.org/TR/css-align-3/#distribution-values
			if len(lines) > 0 {
				spaceEvenly := freeCrossSpace / float64(len(lines)+1)
				currentOffset := spaceEvenly
				if freeCrossSpace < 0 {
					spaceEvenly = 0
					currentOffset = freeCrossSpace / 2
				}
				for i := range lines {
					lineOffsets[i] = currentOffset
					currentOffset += lineCrossSizes[i]
					if i < len(lines)-1 {
						currentOffset += rowGap + spaceEvenly
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

		// Calculate line offsets for the values that were not distributed above
		if alignContent != AlignContentSpaceBetween && alignContent != AlignContentSpaceAround && alignContent != AlignContentSpaceEvenly {
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
		// Single line, or an indefinite cross size: align-content has no free
		// space to distribute, so lines are packed from the cross-start edge
		// (flex-start behavior), separated by the cross-axis gap.
		// https://www.w3.org/TR/css-flexbox-1/#align-content-property
		currentOffset := 0.0
		for i := range lines {
			lineOffsets[i] = currentOffset
			currentOffset += lineCrossSizes[i]
			if i < len(lines)-1 {
				currentOffset += rowGap
			}
		}
	}

	return lineOffsets, totalCrossSize
}
