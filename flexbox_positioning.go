package layout

// flexboxAlignmentMainAxis positions items along the main axis using justify-content.
//
// Algorithm based on CSS Flexible Box Layout Module Level 1:
// - §9.5: Main-Axis Alignment
// - §10.2: Aligning with justify-content
//
// See: https://www.w3.org/TR/css-flexbox-1/#main-alignment
func flexboxAlignmentMainAxis(
	node *Node,
	line []*flexItem,
	setup flexboxSetup,
	lineCrossSize float64,
	lineStartCrossOffset float64,
	columnGap float64,
	mainSize float64,
	isReverse bool,
) float64 {
	// Handle flex-direction reverse - reverse items in line
	// For reverse, we reverse the items and then position from the end
	if isReverse {
		// Reverse the order of items in this line
		for i, j := 0, len(line)-1; i < j; i, j = i+1, j-1 {
			line[i], line[j] = line[j], line[i]
		}
	}

	for _, item := range line {
		// Get rect dimensions - cross-axis was already set by flexboxAlignmentCrossAxis
		// We just need to set/update main-axis dimensions
		var rectWidth, rectHeight float64
		if setup.isRow {
			rectWidth = item.mainSize
			rectHeight = item.node.Rect.Height // Preserve cross-axis value
		} else {
			rectWidth = item.node.Rect.Width // Preserve cross-axis value
			rectHeight = item.mainSize
		}

		// Update main-axis size if needed
		if setup.isRow {
			if rectWidth == 0 && item.node.Style.Width >= 0 {
				rectWidth = item.node.Style.Width
				// Update mainSize so justify-content calculations use correct size
				item.mainSize = item.node.Style.Width
			}
		} else {
			if rectHeight == 0 && item.node.Style.Height >= 0 {
				rectHeight = item.node.Style.Height
				item.mainSize = item.node.Style.Height
			}
		}

		// Update rect with main-axis dimensions (preserving cross-axis from previous step)
		if setup.isRow {
			item.node.Rect.Width = rectWidth
			// Y and Height were already set by flexboxAlignmentCrossAxis
		} else {
			item.node.Rect.Height = rectHeight
			// X and Width were already set by flexboxAlignmentCrossAxis
		}
	}

	// Ensure item.mainSize is set correctly before justify-content calculation
	// This is needed because justifyContentWithGap uses item.mainSize
	for _, item := range line {
		if setup.isRow {
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
	// For reverse direction, we need special handling to ensure gaps are correctly positioned
	// The items array is already reversed, so we position from the start but apply justify-content logic
	if isReverse {
		// For reverse direction, apply justify-content as if it were normal direction
		// but adjust for the reversed semantics
		// FlexStart in reverse means items start from the end (right/bottom)
		// FlexEnd in reverse means items end at the start (left/top)
		reversedJustify := node.Style.JustifyContent
		switch node.Style.JustifyContent {
		case JustifyContentFlexStart:
			reversedJustify = JustifyContentFlexEnd
		case JustifyContentFlexEnd:
			reversedJustify = JustifyContentFlexStart
		}
		// Use normal justify logic with reversed semantics
		justifyContentWithGap(reversedJustify, line, 0, mainSize, setup.isRow, columnGap)
	} else {
		justifyContentWithGap(node.Style.JustifyContent, line, 0, mainSize, setup.isRow, columnGap)
	}

	// Calculate this line's main extent (including margins and gaps)
	// Note: item.node.Rect.X/Y are absolute positions including padding/border
	// We need to calculate the extent relative to the content area start
	lineMainSize := 0.0
	contentAreaStart := 0.0
	if setup.isRow {
		contentAreaStart = node.Style.Padding.Left + node.Style.Border.Left
	} else {
		contentAreaStart = node.Style.Padding.Top + node.Style.Border.Top
	}
	for _, item := range line {
		if setup.isRow {
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

	return lineMainSize
}

// flexboxAlignmentCrossAxis positions items along the cross axis using align-items.
//
// Algorithm based on CSS Flexible Box Layout Module Level 1:
// - §9.6: Cross-Axis Alignment
// - §10.3: Aligning with align-items
// - §10.3.1: Baseline alignment
//
// See: https://www.w3.org/TR/css-flexbox-1/#cross-alignment
// See: https://www.w3.org/TR/css-flexbox-1/#baseline-participation
func flexboxAlignmentCrossAxis(
	node *Node,
	line []*flexItem,
	setup flexboxSetup,
	alignItems AlignItems,
	lineCrossSize float64,
	lineStartCrossOffset float64,
	alignmentCrossSize float64,
) {
	// For baseline alignment, first find the maximum baseline
	// Check if any items use baseline (either via container or align-self)
	var maxBaseline float64 = 0.0
	hasBaseline := false
	for _, item := range line {
		itemAlign := alignItems
		if item.node.Style.AlignSelf != 0 {
			itemAlign = item.node.Style.AlignSelf
		}
		if itemAlign == AlignItemsBaseline {
			hasBaseline = true
			break
		}
	}

	if hasBaseline {
		for _, item := range line {
			itemAlign := alignItems
			if item.node.Style.AlignSelf != 0 {
				itemAlign = item.node.Style.AlignSelf
			}
			if itemAlign == AlignItemsBaseline {
				// Get baseline for this item
				// If node.Baseline is 0 (not set), use the item's cross size as fallback
				baseline := item.node.Baseline
				if baseline == 0 {
					// Default: baseline is at the bottom of the item (for boxes without text)
					baseline = item.crossSize
				}
				// Add top margin to baseline (baseline is relative to content area)
				baselineWithMargin := baseline + item.crossMarginStart
				if baselineWithMargin > maxBaseline {
					maxBaseline = baselineWithMargin
				}
			}
		}
	}

	for _, item := range line {
		// Check for per-item alignment override (CSS Flexbox §8.3)
		itemAlign := alignItems
		if item.node.Style.AlignSelf != 0 {
			itemAlign = item.node.Style.AlignSelf
		}
		// Set initial rect dimensions
		// For row: mainSize=width, crossSize=height
		// For column: mainSize=height, crossSize=width
		var rectWidth, rectHeight float64
		if setup.isRow {
			rectWidth = item.mainSize
			rectHeight = item.crossSize
		} else {
			rectWidth = item.crossSize
			rectHeight = item.mainSize
		}

		// Apply align-self/align-items stretch if needed (for cross-size)
		// Use lineCrossSize consistently - it already accounts for single-line stretch
		if itemAlign == AlignItemsStretch {
			if setup.isRow {
				// For row direction, cross-size is height
				rectHeight = lineCrossSize - item.crossMarginStart - item.crossMarginEnd
				if rectHeight < 0 {
					rectHeight = 0
				}
				item.crossSize = rectHeight
			} else {
				// For column direction, cross-size is width
				rectWidth = lineCrossSize - item.crossMarginStart - item.crossMarginEnd
				if rectWidth < 0 {
					rectWidth = 0
				}
				item.crossSize = rectWidth
			}
		}

		// Calculate cross-axis offset for alignment
		crossOffset := 0.0
		itemCrossSizeWithMargins := item.crossSize + item.crossMarginStart + item.crossMarginEnd
		switch itemAlign {
		case AlignItemsFlexStart:
			crossOffset = item.crossMarginStart
		case AlignItemsFlexEnd:
			crossOffset = alignmentCrossSize - item.crossSize - item.crossMarginEnd
		case AlignItemsCenter:
			crossOffset = (alignmentCrossSize-itemCrossSizeWithMargins)/2 + item.crossMarginStart
		case AlignItemsStretch:
			crossOffset = item.crossMarginStart
		case AlignItemsBaseline:
			// Align item's baseline with the maximum baseline in the line
			itemBaseline := item.node.Baseline
			if itemBaseline == 0 {
				// Default: baseline is at the bottom of the item
				itemBaseline = item.crossSize
			}
			// Offset is the difference between max baseline and this item's baseline
			// Plus the item's top margin (since baseline is relative to content area)
			crossOffset = maxBaseline - itemBaseline
		default:
			crossOffset = item.crossMarginStart
		}

		// Update rect with cross-axis position
		if setup.isRow {
			item.node.Rect.Y = node.Style.Padding.Top + node.Style.Border.Top + lineStartCrossOffset + crossOffset
			item.node.Rect.Height = rectHeight
		} else {
			item.node.Rect.X = node.Style.Padding.Left + node.Style.Border.Left + lineStartCrossOffset + crossOffset
			item.node.Rect.Width = rectWidth
		}
	}
}
