package layout

// blockDetermineSize calculates the node's width and height considering aspect ratio.
//
// Algorithm based on CSS Box Sizing Module Level 4:
// - §5: Aspect Ratios
//
// See: https://www.w3.org/TR/css-sizing-4/#aspect-ratio
//
// setup is a pointer because intrinsic sizing may fall back to auto sizing:
// when CalculateIntrinsicWidth/Height cannot produce a size (returns < 0),
// the corresponding isAutoWidth/isAutoHeight flag is set so the caller sizes
// the box from its content instead of using the negative sentinel.
func blockDetermineSize(node *Node, setup *blockSetup, ctx *LayoutContext, currentFontSize float64) (nodeWidth, nodeHeight float64, aspectRatioCalculatedWidth, aspectRatioCalculatedHeight bool) {
	// Check for intrinsic sizing (min-content, max-content, fit-content)
	// These override auto sizing
	constraints := Loose(setup.contentWidth, setup.contentHeight)

	// Handle width intrinsic sizing
	// Check both sentinel values in Width and WidthSizing enum
	widthValue := ResolveLength(node.Style.Width, ctx, currentFontSize)
	intrinsicWidth := -1.0
	hasIntrinsicWidth := true
	if widthValue == SizeMinContent || node.Style.WidthSizing == IntrinsicSizeMinContent {
		intrinsicWidth = CalculateIntrinsicWidth(node, constraints, IntrinsicSizeMinContent, ctx)
	} else if widthValue == SizeMaxContent || node.Style.WidthSizing == IntrinsicSizeMaxContent {
		intrinsicWidth = CalculateIntrinsicWidth(node, constraints, IntrinsicSizeMaxContent, ctx)
	} else if widthValue == SizeFitContent || node.Style.WidthSizing == IntrinsicSizeFitContent {
		intrinsicWidth = CalculateIntrinsicWidth(node, constraints, IntrinsicSizeFitContent, ctx)
	} else {
		hasIntrinsicWidth = false
	}
	if hasIntrinsicWidth && intrinsicWidth >= 0 {
		nodeWidth = intrinsicWidth
		// An intrinsic size is a definite size for the remainder of layout.
		setup.isAutoWidth = false
	} else {
		if hasIntrinsicWidth {
			// Intrinsic sizing requested but not computable: behave as auto.
			// CSS Sizing Level 3 §4.1 (https://www.w3.org/TR/css-sizing-3/#intrinsic-sizes)
			setup.isAutoWidth = true
		}
		// Normal width handling
		nodeWidth = setup.specifiedWidth
		if setup.isAutoWidth {
			nodeWidth = setup.contentWidth // auto
		}
	}

	// Handle height intrinsic sizing
	heightValue := ResolveLength(node.Style.Height, ctx, currentFontSize)
	intrinsicHeight := -1.0
	hasIntrinsicHeight := true
	if heightValue == SizeMinContent || node.Style.HeightSizing == IntrinsicSizeMinContent {
		intrinsicHeight = CalculateIntrinsicHeight(node, constraints, IntrinsicSizeMinContent, ctx)
	} else if heightValue == SizeMaxContent || node.Style.HeightSizing == IntrinsicSizeMaxContent {
		intrinsicHeight = CalculateIntrinsicHeight(node, constraints, IntrinsicSizeMaxContent, ctx)
	} else if heightValue == SizeFitContent || node.Style.HeightSizing == IntrinsicSizeFitContent {
		intrinsicHeight = CalculateIntrinsicHeight(node, constraints, IntrinsicSizeFitContent, ctx)
	} else {
		hasIntrinsicHeight = false
	}
	if hasIntrinsicHeight && intrinsicHeight >= 0 {
		nodeHeight = intrinsicHeight
		setup.isAutoHeight = false
	} else {
		if hasIntrinsicHeight {
			// Intrinsic height is not computable up front (the block axis size
			// of a block container is its content height): fall back to auto
			// so the height is taken from the laid-out children instead of
			// storing the -1 sentinel as a real size.
			// CSS Sizing Level 3 §5.2 (https://www.w3.org/TR/css-sizing-3/#intrinsic-contribution)
			setup.isAutoHeight = true
		}
		// Normal height handling
		nodeHeight = setup.specifiedHeight
		if setup.isAutoHeight {
			// For auto height, don't set to Unbounded initially if aspect ratio will calculate it
			// Aspect ratio calculation happens next and will set height based on width
			if node.Style.AspectRatio > 0 && setup.isAutoWidth && setup.contentWidth > 0 {
				// Will be calculated by aspect ratio below
				nodeHeight = 0
			} else {
				nodeHeight = setup.contentHeight // auto
			}
		}
	}

	// Apply aspect ratio if set (before min/max constraints)
	// Aspect ratio affects sizing when one dimension is auto
	// According to CSS Box Sizing Module Level 4:
	// - When both width and height are auto, aspect ratio uses the available space
	// - Prefer width-based calculation if available width > 0
	// - If available height is bounded and would constrain, use height-based instead
	if node.Style.AspectRatio > 0 {
		if setup.isAutoWidth && setup.isAutoHeight {
			// Both auto: use available space and aspect ratio
			// Prefer width-based calculation (use available width)
			if setup.contentWidth > 0 {
				// Use available width, calculate height from aspect ratio
				nodeHeight = nodeWidth / node.Style.AspectRatio
				aspectRatioCalculatedHeight = true
				aspectRatioCalculatedWidth = true // Width is set from contentWidth (nodeWidth)
				// Constrain to available height if it's bounded
				if setup.contentHeight < Unbounded && nodeHeight > setup.contentHeight {
					nodeHeight = setup.contentHeight
					nodeWidth = nodeHeight * node.Style.AspectRatio
					// Both recalculated
				}
			} else if setup.contentHeight > 0 && setup.contentHeight < Unbounded {
				// Use available height, calculate width from aspect ratio
				// Only if contentWidth is 0 and contentHeight is bounded
				nodeWidth = setup.contentHeight * node.Style.AspectRatio
				nodeHeight = setup.contentHeight
				aspectRatioCalculatedWidth = true
				aspectRatioCalculatedHeight = true
			}
		} else if setup.isAutoWidth {
			// Width is auto, height is set: calculate width from height and aspect ratio
			nodeWidth = nodeHeight * node.Style.AspectRatio
			aspectRatioCalculatedWidth = true
		} else if setup.isAutoHeight {
			// Height is auto, width is set: calculate height from width and aspect ratio
			nodeHeight = nodeWidth / node.Style.AspectRatio
			aspectRatioCalculatedHeight = true
		}
		// If both width and height are explicitly set, aspect ratio is ignored (CSS behavior)
	}

	return nodeWidth, nodeHeight, aspectRatioCalculatedWidth, aspectRatioCalculatedHeight
}

// blockApplyConstraints applies min/max constraints while maintaining aspect ratio if needed.
//
// Algorithm based on CSS Sizing Module Level 3:
// - §5: Intrinsic Size Determination
//
// See: https://www.w3.org/TR/css-sizing-3/#constraints
func blockApplyConstraints(node *Node, setup blockSetup, nodeWidth, nodeHeight float64, aspectRatioCalculatedWidth, aspectRatioCalculatedHeight bool) (float64, float64) {
	// Apply max first, then min, so that min wins when min > max.
	// CSS 2.1 §10.4 / §10.7: https://www.w3.org/TR/CSS21/visudet.html#min-max-widths
	// If aspect ratio calculated dimensions, we need to maintain the ratio when min/max are applied
	if setup.maxWidthContent > 0 && setup.maxWidthContent < Unbounded {
		nodeWidth = min(nodeWidth, setup.maxWidthContent)
		// If aspect ratio calculated width, recalculate height to maintain ratio
		if aspectRatioCalculatedWidth && node.Style.AspectRatio > 0 {
			nodeHeight = nodeWidth / node.Style.AspectRatio
		}
	}
	if setup.minWidthContent > 0 {
		nodeWidth = max(nodeWidth, setup.minWidthContent)
		// If aspect ratio calculated width, recalculate height to maintain ratio
		if aspectRatioCalculatedWidth && node.Style.AspectRatio > 0 {
			nodeHeight = nodeWidth / node.Style.AspectRatio
		}
	}
	if setup.maxHeightContent > 0 && setup.maxHeightContent < Unbounded {
		oldHeight := nodeHeight
		nodeHeight = min(nodeHeight, setup.maxHeightContent)
		// If aspect ratio calculated height and MaxHeight decreased it, recalculate width to maintain ratio
		if aspectRatioCalculatedHeight && node.Style.AspectRatio > 0 && nodeHeight < oldHeight {
			nodeWidth = nodeHeight * node.Style.AspectRatio
		}
	}
	if setup.minHeightContent > 0 {
		oldHeight := nodeHeight
		nodeHeight = max(nodeHeight, setup.minHeightContent)
		// If aspect ratio calculated height and MinHeight changed it, recalculate width to maintain ratio
		if aspectRatioCalculatedHeight && node.Style.AspectRatio > 0 && nodeHeight != oldHeight {
			nodeWidth = nodeHeight * node.Style.AspectRatio
		}
	}

	// Constrain to available space
	// Don't constrain if aspect ratio already calculated dimensions (they're already constrained)
	if !aspectRatioCalculatedWidth {
		nodeWidth = min(nodeWidth, setup.contentWidth)
	}
	if !aspectRatioCalculatedHeight {
		nodeHeight = min(nodeHeight, setup.contentHeight)
	}

	return nodeWidth, nodeHeight
}
