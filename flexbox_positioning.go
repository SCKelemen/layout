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
	ctx *LayoutContext,
) float64 {
	// Handle flex-direction reverse - reverse items in line
	// For reverse, we reverse the items and then position from the end
	if isReverse {
		// Reverse the order of items in this line
		for i, j := 0, len(line)-1; i < j; i, j = i+1, j-1 {
			line[i], line[j] = line[j], line[i]
		}
	}

	// Get parent font size for Length resolution
	parentFontSize := getCurrentFontSize(node, ctx)

	// Apply the resolved main size (§9.7) to each item's Rect. The cross-axis
	// position and size were already set by flexboxAlignmentCrossAxis. The
	// resolved size is authoritative: an item legitimately shrunk to zero must
	// not be reset to its explicit width/height, or it would overflow the line.
	for _, item := range line {
		if setup.isMainHorizontal {
			item.node.Rect.Width = item.mainSize
		} else {
			item.node.Rect.Height = item.mainSize
		}
	}

	// Calculate content area start offset (accounting for padding and border)
	contentAreaStart := 0.0
	if setup.isMainHorizontal {
		contentAreaStart = ResolveLength(node.Style.Padding.Left, ctx, parentFontSize) + ResolveLength(node.Style.Border.Left, ctx, parentFontSize)
	} else {
		contentAreaStart = ResolveLength(node.Style.Padding.Top, ctx, parentFontSize) + ResolveLength(node.Style.Border.Top, ctx, parentFontSize)
	}

	// With an indefinite main size the line is exactly as long as its content:
	// there is no free space to distribute and the container must not be sized
	// from an unbounded value (CSS Flexbox §9.2 step 4 / §9.5).
	// https://www.w3.org/TR/css-flexbox-1/#algo-main-container
	if !setup.hasExplicitMainSize || mainSize >= Unbounded {
		mainSize = 0
		for _, item := range line {
			mainSize += item.mainSize + item.mainMarginStart + item.mainMarginEnd
		}
		if len(line) > 1 {
			mainSize += columnGap * float64(len(line)-1)
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
		case JustifyContentFlexStart, JustifyContentStretch:
			// stretch behaves as flex-start in a flex container (§8.2), so it
			// packs items toward the reversed main-start edge as well.
			reversedJustify = JustifyContentFlexEnd
		case JustifyContentFlexEnd:
			reversedJustify = JustifyContentFlexStart
		}
		// Use normal justify logic with reversed semantics
		justifyContentWithGap(reversedJustify, line, contentAreaStart, mainSize, setup.isMainHorizontal, columnGap, setup.writingMode)
	} else {
		justifyContentWithGap(node.Style.JustifyContent, line, contentAreaStart, mainSize, setup.isMainHorizontal, columnGap, setup.writingMode)
	}

	// Calculate this line's main extent (including margins and gaps)
	// Note: item.node.Rect.X/Y are absolute positions including padding/border
	// We need to calculate the extent relative to the content area start
	lineMainSize := 0.0
	for _, item := range line {
		if setup.isMainHorizontal {
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
	ctx *LayoutContext,
) {
	// For baseline alignment, first find the maximum baseline
	// Check if any items use baseline (either via container or align-self)
	var maxBaseline float64 = 0.0
	hasBaseline := false
	for _, item := range line {
		itemAlign := flexboxResolveItemAlign(item, alignItems, setup)
		if itemAlign == AlignItemsBaseline {
			hasBaseline = true
			break
		}
	}

	if hasBaseline {
		for _, item := range line {
			itemAlign := flexboxResolveItemAlign(item, alignItems, setup)
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
		itemAlign := flexboxResolveItemAlign(item, alignItems, setup)
		// Set initial rect dimensions
		// For main axis horizontal: mainSize=width, crossSize=height
		// For main axis vertical: mainSize=height, crossSize=width
		var rectWidth, rectHeight float64
		if setup.isMainHorizontal {
			rectWidth = item.mainSize
			rectHeight = item.crossSize
		} else {
			rectWidth = item.crossSize
			rectHeight = item.mainSize
		}

		// Apply align-self/align-items stretch if needed (for cross-size).
		// Use lineCrossSize consistently - it already accounts for single-line stretch.
		// §9.4 step 11: stretch only applies when the item's cross size is auto;
		// an explicit cross size is kept and the item is aligned like flex-start.
		// https://www.w3.org/TR/css-flexbox-1/#algo-stretch
		if itemAlign == AlignItemsStretch && !item.hasExplicitCrossSize {
			if setup.isMainHorizontal {
				// For main axis horizontal, cross-size is height
				rectHeight = lineCrossSize - item.crossMarginStart - item.crossMarginEnd
				if rectHeight < 0 {
					rectHeight = 0
				}
				item.crossSize = rectHeight
			} else {
				// For main axis vertical, cross-size is width
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

		// Get parent font size for Length resolution
		parentFontSize := getCurrentFontSize(node, ctx)

		// Update rect with cross-axis position
		if setup.isMainHorizontal {
			item.node.Rect.Y = ResolveLength(node.Style.Padding.Top, ctx, parentFontSize) + ResolveLength(node.Style.Border.Top, ctx, parentFontSize) + lineStartCrossOffset + crossOffset
			item.node.Rect.Height = rectHeight
		} else {
			item.node.Rect.X = ResolveLength(node.Style.Padding.Left, ctx, parentFontSize) + ResolveLength(node.Style.Border.Left, ctx, parentFontSize) + lineStartCrossOffset + crossOffset
			item.node.Rect.Width = rectWidth
		}
	}
}

// flexboxResolveItemAlign resolves the effective cross-axis alignment of an
// item: align-self overrides align-items (CSS Flexbox §8.3), and baseline
// alignment in a column flex container (where the baseline runs parallel to the
// main axis) behaves as flex-start (CSS Flexbox §8.3 / CSS Box Alignment §7.2).
// https://www.w3.org/TR/css-flexbox-1/#align-items-property
func flexboxResolveItemAlign(item *flexItem, alignItems AlignItems, setup flexboxSetup) AlignItems {
	itemAlign := alignItems
	if item.node.Style.AlignSelf != 0 {
		itemAlign = item.node.Style.AlignSelf
	}
	if itemAlign == AlignItemsBaseline && !setup.isRow {
		itemAlign = AlignItemsFlexStart
	}
	return itemAlign
}
