package layout

import "math"

// Intrinsic sizing algorithms for CSS Sizing Module Level 3.
//
// Implements min-content, max-content, and fit-content sizing for all layout modes.
//
// Algorithm based on CSS Sizing Module Level 3:
// - §4: Intrinsic Size Determination
// - §5: Extrinsic Size Determination
//
// See: https://www.w3.org/TR/css-sizing-3/#intrinsic-sizes

// CalculateIntrinsicWidth calculates the intrinsic width of a node.
// Returns the width based on the specified sizing type (min-content, max-content, fit-content).
//
// Parameters:
//   - node: The node to calculate intrinsic width for
//   - constraints: The available space constraints
//   - sizingType: The type of intrinsic sizing (min-content, max-content, fit-content)
//   - ctx: Layout context for Length resolution
//
// Returns: The calculated intrinsic width
func CalculateIntrinsicWidth(node *Node, constraints Constraints, sizingType IntrinsicSize, ctx *LayoutContext) float64 {
	switch sizingType {
	case IntrinsicSizeMinContent:
		return calculateMinContentWidth(node, constraints, ctx)
	case IntrinsicSizeMaxContent:
		return calculateMaxContentWidth(node, constraints, ctx)
	case IntrinsicSizeFitContent:
		// fit-content: clamp max-content to FitContentWidth
		maxContent := calculateMaxContentWidth(node, constraints, ctx)
		currentFontSize := getCurrentFontSize(node, ctx)
		fitContentWidth := ResolveLength(node.Style.FitContentWidth, ctx, currentFontSize)
		if fitContentWidth > 0 {
			return math.Min(maxContent, fitContentWidth)
		}
		return maxContent
	default:
		return -1 // Auto
	}
}

// CalculateIntrinsicHeight calculates the intrinsic height of a node.
// Returns the height based on the specified sizing type (min-content, max-content, fit-content).
func CalculateIntrinsicHeight(node *Node, constraints Constraints, sizingType IntrinsicSize, ctx *LayoutContext) float64 {
	switch sizingType {
	case IntrinsicSizeMinContent:
		return calculateMinContentHeight(node, constraints, ctx)
	case IntrinsicSizeMaxContent:
		return calculateMaxContentHeight(node, constraints, ctx)
	case IntrinsicSizeFitContent:
		// fit-content: clamp max-content to FitContentHeight
		maxContent := calculateMaxContentHeight(node, constraints, ctx)
		currentFontSize := getCurrentFontSize(node, ctx)
		fitContentHeight := ResolveLength(node.Style.FitContentHeight, ctx, currentFontSize)
		if fitContentHeight > 0 {
			return math.Min(maxContent, fitContentHeight)
		}
		return maxContent
	default:
		return -1 // Auto
	}
}

// calculateMinContentWidth calculates the min-content width.
// This is the narrowest width the content can take without overflow.
func calculateMinContentWidth(node *Node, constraints Constraints, ctx *LayoutContext) float64 {
	switch node.Style.Display {
	case DisplayFlex:
		return calculateFlexMinContentWidth(node, constraints, ctx)
	case DisplayGrid:
		return calculateGridMinContentWidth(node, constraints, ctx)
	case DisplayBlock:
		return calculateBlockMinContentWidth(node, constraints, ctx)
	default:
		return 0
	}
}

// calculateMaxContentWidth calculates the max-content width.
// This is the widest natural width (no wrapping).
func calculateMaxContentWidth(node *Node, constraints Constraints, ctx *LayoutContext) float64 {
	switch node.Style.Display {
	case DisplayFlex:
		return calculateFlexMaxContentWidth(node, constraints, ctx)
	case DisplayGrid:
		return calculateGridMaxContentWidth(node, constraints, ctx)
	case DisplayBlock:
		return calculateBlockMaxContentWidth(node, constraints, ctx)
	default:
		return 0
	}
}

// calculateMinContentHeight calculates the min-content height.
func calculateMinContentHeight(node *Node, constraints Constraints, ctx *LayoutContext) float64 {
	// For most layouts, min-content height is the same as auto height
	// (height based on content with available width)
	// This is a simplified implementation
	return -1 // Auto (layout will determine from content)
}

// calculateMaxContentHeight calculates the max-content height.
func calculateMaxContentHeight(node *Node, constraints Constraints, ctx *LayoutContext) float64 {
	// For most layouts, max-content height is the same as auto height
	return -1 // Auto (layout will determine from content)
}

// calculateBlockMinContentWidth calculates min-content width for block layout.
// For block layout, this is the maximum of children's min-content widths.
func calculateBlockMinContentWidth(node *Node, constraints Constraints, ctx *LayoutContext) float64 {
	maxChildWidth := 0.0

	for _, child := range node.Children {
		if child.Style.Display == DisplayNone {
			continue
		}

		// Calculate child's min-content width recursively
		childWidth := 0.0
		if child.Style.Width.Value > 0 {
			// Explicit width (assuming pixels, should use ResolveLength with ctx)
			childWidth = child.Style.Width.Value
		} else if child.Style.Width.Value == SizeMinContent || child.Style.WidthSizing == IntrinsicSizeMinContent {
			// Recursive min-content
			childWidth = CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeMinContent, ctx)
		} else if child.Style.Width.Value == SizeMaxContent || child.Style.WidthSizing == IntrinsicSizeMaxContent {
			// Max-content for child
			childWidth = CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeMaxContent, ctx)
		} else {
			// For children with auto width, use max-content as approximation
			childWidth = CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeMaxContent, ctx)
		}

		// Add margins (assuming pixels, should use ResolveLength with ctx)
		childWidth += child.Style.Margin.Left.Value + child.Style.Margin.Right.Value

		if childWidth > maxChildWidth {
			maxChildWidth = childWidth
		}
	}

	// Add padding and border
	currentFontSize := getCurrentFontSize(node, ctx)
	horizontalPaddingBorder := getHorizontalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, currentFontSize)
	return maxChildWidth + horizontalPaddingBorder
}

// calculateBlockMaxContentWidth calculates max-content width for block layout.
// For block layout, this is the maximum of children's max-content widths.
func calculateBlockMaxContentWidth(node *Node, constraints Constraints, ctx *LayoutContext) float64 {
	maxChildWidth := 0.0

	for _, child := range node.Children {
		if child.Style.Display == DisplayNone {
			continue
		}

		// Calculate child's max-content width recursively
		childWidth := 0.0
		if child.Style.Width.Value > 0 {
			// Explicit width (assuming pixels, should use ResolveLength with ctx)
			childWidth = child.Style.Width.Value
		} else {
			// Recursive max-content
			childWidth = CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeMaxContent, ctx)
		}

		// Add margins (assuming pixels, should use ResolveLength with ctx)
		childWidth += child.Style.Margin.Left.Value + child.Style.Margin.Right.Value

		if childWidth > maxChildWidth {
			maxChildWidth = childWidth
		}
	}

	// Add padding and border
	currentFontSize := getCurrentFontSize(node, ctx)
	horizontalPaddingBorder := getHorizontalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, currentFontSize)
	return maxChildWidth + horizontalPaddingBorder
}

// calculateFlexMinContentWidth calculates min-content width for flex layout.
func calculateFlexMinContentWidth(node *Node, constraints Constraints, ctx *LayoutContext) float64 {
	isRow := node.Style.FlexDirection == FlexDirectionRow || node.Style.FlexDirection == FlexDirectionRowReverse

	if isRow {
		// Flex row: sum of children's min-content widths
		totalWidth := 0.0
		for _, child := range node.Children {
			if child.Style.Display == DisplayNone {
				continue
			}
			// If child has explicit width, use it; otherwise calculate intrinsically
			childWidth := 0.0
			if child.Style.Width.Value > 0 {
				childWidth = child.Style.Width.Value
			} else {
				childWidth = CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeMinContent, ctx)
			}
			childWidth += child.Style.Margin.Left.Value + child.Style.Margin.Right.Value
			totalWidth += childWidth
		}

		// Add gaps
		gap := node.Style.FlexGap
		if node.Style.FlexColumnGap.Value > 0 {
			gap = node.Style.FlexColumnGap
		}
		if len(node.Children) > 1 {
			totalWidth += gap.Value * float64(len(node.Children)-1)
		}

		currentFontSize := getCurrentFontSize(node, ctx)
		horizontalPaddingBorder := getHorizontalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, currentFontSize)
		return totalWidth + horizontalPaddingBorder
	} else {
		// Flex column: max of children's min-content widths
		maxWidth := 0.0
		for _, child := range node.Children {
			if child.Style.Display == DisplayNone {
				continue
			}
			// If child has explicit width, use it; otherwise calculate intrinsically
			childWidth := 0.0
			if child.Style.Width.Value > 0 {
				childWidth = child.Style.Width.Value
			} else {
				childWidth = CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeMinContent, ctx)
			}
			childWidth += child.Style.Margin.Left.Value + child.Style.Margin.Right.Value
			if childWidth > maxWidth {
				maxWidth = childWidth
			}
		}

		currentFontSize := getCurrentFontSize(node, ctx)
		horizontalPaddingBorder := getHorizontalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, currentFontSize)
		return maxWidth + horizontalPaddingBorder
	}
}

// calculateFlexMaxContentWidth calculates max-content width for flex layout.
func calculateFlexMaxContentWidth(node *Node, constraints Constraints, ctx *LayoutContext) float64 {
	isRow := node.Style.FlexDirection == FlexDirectionRow || node.Style.FlexDirection == FlexDirectionRowReverse

	if isRow {
		// Flex row: sum of children's max-content widths
		totalWidth := 0.0
		for _, child := range node.Children {
			if child.Style.Display == DisplayNone {
				continue
			}
			// If child has explicit width, use it; otherwise calculate intrinsically
			childWidth := 0.0
			if child.Style.Width.Value > 0 {
				childWidth = child.Style.Width.Value
			} else {
				childWidth = CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeMaxContent, ctx)
			}
			childWidth += child.Style.Margin.Left.Value + child.Style.Margin.Right.Value
			totalWidth += childWidth
		}

		// Add gaps
		gap := node.Style.FlexGap
		if node.Style.FlexColumnGap.Value > 0 {
			gap = node.Style.FlexColumnGap
		}
		if len(node.Children) > 1 {
			totalWidth += gap.Value * float64(len(node.Children)-1)
		}

		currentFontSize := getCurrentFontSize(node, ctx)
		horizontalPaddingBorder := getHorizontalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, currentFontSize)
		return totalWidth + horizontalPaddingBorder
	} else {
		// Flex column: max of children's max-content widths
		maxWidth := 0.0
		for _, child := range node.Children {
			if child.Style.Display == DisplayNone {
				continue
			}
			// If child has explicit width, use it; otherwise calculate intrinsically
			childWidth := 0.0
			if child.Style.Width.Value > 0 {
				childWidth = child.Style.Width.Value
			} else {
				childWidth = CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeMaxContent, ctx)
			}
			childWidth += child.Style.Margin.Left.Value + child.Style.Margin.Right.Value
			if childWidth > maxWidth {
				maxWidth = childWidth
			}
		}

		currentFontSize := getCurrentFontSize(node, ctx)
		horizontalPaddingBorder := getHorizontalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, currentFontSize)
		return maxWidth + horizontalPaddingBorder
	}
}

// calculateGridMinContentWidth calculates min-content width for grid layout.
// This is the sum of min-content-sized column tracks.
func calculateGridMinContentWidth(node *Node, constraints Constraints, ctx *LayoutContext) float64 {
	if len(node.Style.GridTemplateColumns) == 0 {
		return 0
	}

	totalWidth := 0.0
	for i, track := range node.Style.GridTemplateColumns {
		trackSize := resolveIntrinsicTrackSize(track, node, i, true, IntrinsicSizeMinContent, ctx, 16.0)
		totalWidth += trackSize
	}

	// Add gaps
	gap := node.Style.GridGap
	if node.Style.GridColumnGap.Value > 0 {
		gap = node.Style.GridColumnGap
	}
	if len(node.Style.GridTemplateColumns) > 1 {
		totalWidth += gap.Value * float64(len(node.Style.GridTemplateColumns)-1)
	}

	currentFontSize := getCurrentFontSize(node, ctx)
	horizontalPaddingBorder := getHorizontalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, currentFontSize)
	return totalWidth + horizontalPaddingBorder
}

// calculateGridMaxContentWidth calculates max-content width for grid layout.
// This is the sum of max-content-sized column tracks.
func calculateGridMaxContentWidth(node *Node, constraints Constraints, ctx *LayoutContext) float64 {
	if len(node.Style.GridTemplateColumns) == 0 {
		return 0
	}

	totalWidth := 0.0
	for i, track := range node.Style.GridTemplateColumns {
		trackSize := resolveIntrinsicTrackSize(track, node, i, true, IntrinsicSizeMaxContent, ctx, 16.0)
		totalWidth += trackSize
	}

	// Add gaps
	gap := node.Style.GridGap
	if node.Style.GridColumnGap.Value > 0 {
		gap = node.Style.GridColumnGap
	}
	if len(node.Style.GridTemplateColumns) > 1 {
		totalWidth += gap.Value * float64(len(node.Style.GridTemplateColumns)-1)
	}

	currentFontSize := getCurrentFontSize(node, ctx)
	horizontalPaddingBorder := getHorizontalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, currentFontSize)
	return totalWidth + horizontalPaddingBorder
}

// resolveIntrinsicTrackSize resolves a grid track's size for intrinsic sizing.
// This handles min-content, max-content, and fit-content tracks.
func resolveIntrinsicTrackSize(track GridTrack, container *Node, trackIndex int, isColumn bool, sizingType IntrinsicSize, ctx *LayoutContext, currentFontSize float64) float64 {
	// Resolve track sizes
	minSize := ResolveLength(track.MinSize, ctx, currentFontSize)
	maxSize := ResolveLength(track.MaxSize, ctx, currentFontSize)

	// Fixed tracks use their fixed size
	if minSize == maxSize {
		return minSize
	}

	// Check if track uses intrinsic sizing sentinel values
	if maxSize == SizeMinContent {
		// min-content track: use minimum size of items in this track
		return calculateTrackMinContent(container, trackIndex, isColumn, ctx)
	}
	if maxSize == SizeMaxContent {
		// max-content track: use maximum size of items in this track
		return calculateTrackMaxContent(container, trackIndex, isColumn, ctx)
	}

	// For auto and fractional tracks, use the sizing type passed in
	if sizingType == IntrinsicSizeMinContent {
		// Use MinSize as approximation
		return minSize
	} else {
		// Use MaxSize or a reasonable default
		if maxSize != Unbounded {
			return maxSize
		}
		// For unbounded tracks, calculate from content
		return calculateTrackMaxContent(container, trackIndex, isColumn, ctx)
	}
}

// calculateTrackMinContent calculates the min-content size for a grid track.
func calculateTrackMinContent(container *Node, trackIndex int, isColumn bool, ctx *LayoutContext) float64 {
	maxSize := 0.0

	// Find all items in this track and get their min-content size
	for _, child := range container.Children {
		if child.Style.Display == DisplayNone {
			continue
		}

		// Check if this child is in this track
		var inTrack bool
		if isColumn {
			colStart := child.Style.GridColumnStart
			if colStart < 0 {
				colStart = 0
			}
			inTrack = colStart == trackIndex
		} else {
			rowStart := child.Style.GridRowStart
			if rowStart < 0 {
				rowStart = 0
			}
			inTrack = rowStart == trackIndex
		}

		if !inTrack {
			continue
		}

		// Calculate child's min-content size
		var childSize float64
		if isColumn {
			childSize = CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeMinContent, ctx)
		} else {
			childSize = CalculateIntrinsicHeight(child, Unconstrained(), IntrinsicSizeMinContent, ctx)
		}

		if childSize > maxSize {
			maxSize = childSize
		}
	}

	return maxSize
}

// calculateTrackMaxContent calculates the max-content size for a grid track.
func calculateTrackMaxContent(container *Node, trackIndex int, isColumn bool, ctx *LayoutContext) float64 {
	maxSize := 0.0

	// Find all items in this track and get their max-content size
	for _, child := range container.Children {
		if child.Style.Display == DisplayNone {
			continue
		}

		// Check if this child is in this track
		var inTrack bool
		if isColumn {
			colStart := child.Style.GridColumnStart
			if colStart < 0 {
				colStart = 0
			}
			inTrack = colStart == trackIndex
		} else {
			rowStart := child.Style.GridRowStart
			if rowStart < 0 {
				rowStart = 0
			}
			inTrack = rowStart == trackIndex
		}

		if !inTrack {
			continue
		}

		// Calculate child's max-content size
		var childSize float64
		if isColumn {
			childSize = CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeMaxContent, ctx)
		} else {
			childSize = CalculateIntrinsicHeight(child, Unconstrained(), IntrinsicSizeMaxContent, ctx)
		}

		if childSize > maxSize {
			maxSize = childSize
		}
	}

	return maxSize
}
