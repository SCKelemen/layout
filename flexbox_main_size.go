package layout

import (
	"math"
)

// flexboxDetermineMainSize determines the main size of items in a line using flex grow/shrink.
//
// Algorithm based on CSS Flexible Box Layout Module Level 1:
// - §9.3: Main Size Determination
//   - §9.3.1: Initial free space calculation
//   - §9.3.2: Flex grow factor distribution
//   - §9.3.3: Flex shrink factor distribution
//
// See: https://www.w3.org/TR/css-flexbox-1/#main-sizing
func flexboxDetermineMainSize(line []*flexItem, mainSize float64, hasExplicitMainSize bool) {
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

	// Distribute free space only if the container has a definite main size
	// If main size is indefinite (auto), don't flex - keep items at base size so container grows to fit content
	if hasExplicitMainSize {
		// Only run the flex grow/shrink algorithm if the container's main size is definite.
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
			// No free space or no flex factors: use base size as-is
			for _, item := range line {
				item.mainSize = item.baseSize
			}
		}
	} else {
		// Indefinite main size (auto-sized flex container): don't flex in the main axis.
		// Keep mainSize = baseSize so the container grows to fit content.
		for _, item := range line {
			item.mainSize = item.baseSize
		}
	}
}

