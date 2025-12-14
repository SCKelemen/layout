package layout

// flexboxDetermineCrossSize determines the cross size of a line.
//
// Algorithm based on CSS Flexible Box Layout Module Level 1:
// - §9.4: Cross Size Determination
//   - §9.4.1: Single-line flex container cross size
//   - §9.4.2: Multi-line flex container cross size
//
// See: https://www.w3.org/TR/css-flexbox-1/#cross-sizing
func flexboxDetermineCrossSize(line []*flexItem, crossSize float64, alignItems AlignItems, hasExplicitCrossSize bool, isSingleLine bool) float64 {
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
	// Only use crossSize when the cross size is definite (explicit style or constraints)
	// Otherwise, use content-driven lineCrossSize to avoid zeroing out content
	if alignItems == AlignItemsStretch && isSingleLine && hasExplicitCrossSize {
		// Only override line cross size with container cross size if the cross size is definite.
		// If crossSize is smaller than content, we may clamp, but never shrink below content.
		if crossSize > 0 {
			// Grow to container cross size but do not shrink below content
			if crossSize > lineCrossSize {
				lineCrossSize = crossSize
			}
			// If crossSize <= lineCrossSize, leave lineCrossSize as content-driven.
		}
	}

	return lineCrossSize
}
