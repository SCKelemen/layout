package layout

// flexboxHandleWrapReverse handles flex-wrap: wrap-reverse by reversing line order and mirroring offsets.
//
// Algorithm based on CSS Flexible Box Layout Module Level 1:
// - §9.2: Line Length Determination (wrap-reverse)
//
// See: https://www.w3.org/TR/css-flexbox-1/#flex-wrap-property
//
// For wrap-reverse, we simply reverse the visual order of lines and mirror their offsets
// in the cross axis. We do NOT re-run align-content, as that was already computed.
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
	// wrap-reverse works with or without explicit cross size
	// If no explicit cross size, use totalCrossSize for mirroring
	if !wrapReverse || len(lines) <= 1 {
		return lineOffsets, totalCrossSize
	}

	// Use explicit cross size if available, otherwise use totalCrossSize
	mirrorCrossSize := crossSize
	if !hasExplicitCrossSize {
		mirrorCrossSize = totalCrossSize
	}

	n := len(lines)

	// For wrap-reverse, we reverse the visual order of lines and mirror their offsets in cross space.
	// Mirroring: for a line at offset `top` with height `height`, the mirrored position is:
	// newTop = crossSize - (top + height)
	//
	// The correct approach per the user's explanation:
	// 1. Mirror each line's offset: newTop = crossSize - (top + height)
	// 2. Reverse the line order
	// 3. The mirrored offsets should align with the reversed line order
	//
	// However, we need to be careful: after mirroring, the last line (originally at the bottom)
	// should be at the top in the reversed visual order.
	//
	// Strategy: Compute mirrored offsets for original positions, then reverse everything together
	for i := 0; i < n; i++ {
		top := lineOffsets[i]
		height := lineCrossSizes[i]
		lineOffsets[i] = mirrorCrossSize - (top + height)
	}

	// Reverse line order (visual order is reversed)
	// This also reverses the already-mirrored offsets to match the reversed line order
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
		lineCrossSizes[i], lineCrossSizes[j] = lineCrossSizes[j], lineCrossSizes[i]
		lineOffsets[i], lineOffsets[j] = lineOffsets[j], lineOffsets[i]
	}

	// totalCrossSize doesn't change - it's just the sum of lineCrossSizes + gaps
	return lineOffsets, totalCrossSize
}

