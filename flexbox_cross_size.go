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
	//
	// Use container's cross size when:
	// 1. Container has explicit cross size (from style or constraints), OR
	// 2. All children have zero cross size (meaning they want to stretch)
	//    AND crossSize is definite (not unbounded)
	if alignItems == AlignItemsStretch && isSingleLine {
		shouldUseCrossSize := hasExplicitCrossSize ||
			(lineCrossSize == 0 && crossSize > 0 && crossSize < Unbounded)

		if shouldUseCrossSize && crossSize > 0 && crossSize < Unbounded {
			// Grow to container cross size but do not shrink below content
			// (unless content is 0, in which case we use crossSize)
			if crossSize > lineCrossSize || lineCrossSize == 0 {
				lineCrossSize = crossSize
			}
		}
	}

	return lineCrossSize
}
