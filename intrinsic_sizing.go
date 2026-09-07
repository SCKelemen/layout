package layout

import (
	"math"
	"strings"
)

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
	case DisplayInlineText:
		return calculateTextMinContentWidth(node, ctx)
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
	case DisplayInlineText:
		return calculateTextMaxContentWidth(node, ctx)
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

// childIntrinsicWidthContribution returns a child's contribution to its
// parent's intrinsic inline size: the child's outer (margin-box) width.
//
// CSS Sizing Level 3 §5.1: the min-/max-content contribution of a box is the
// size of its margin box when sized under the corresponding constraint,
// i.e. content (or the specified width) + padding + border + margins.
// https://www.w3.org/TR/css-sizing-3/#intrinsic-contribution
//
// Lengths are resolved with the child's font size so em/rem/vw values count
// in pixels, and a specified width is interpreted through the child's
// box-sizing (a content-box width has padding and border added on top).
func childIntrinsicWidthContribution(child *Node, sizingType IntrinsicSize, ctx *LayoutContext) float64 {
	fontSize := getCurrentFontSize(child, ctx)
	paddingBorder := getHorizontalPaddingBorder(child.Style.Padding, child.Style.Border, ctx, fontSize)

	widthPx := ResolveLength(child.Style.Width, ctx, fontSize)
	var width float64
	switch {
	case !isUnsetLength(child.Style.Width) && widthPx > 0 && widthPx < Unbounded:
		// Specified width: outer size depends on box-sizing.
		// CSS Box Sizing Level 3 §3.1: https://www.w3.org/TR/css-sizing-3/#box-sizing
		if child.Style.BoxSizing == BoxSizingBorderBox {
			width = widthPx
		} else {
			width = widthPx + paddingBorder
		}
	case widthPx == SizeMinContent || child.Style.WidthSizing == IntrinsicSizeMinContent:
		// The intrinsic width functions already include the child's own padding and border.
		width = CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeMinContent, ctx)
	case widthPx == SizeMaxContent || child.Style.WidthSizing == IntrinsicSizeMaxContent:
		width = CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeMaxContent, ctx)
	case widthPx == SizeFitContent || child.Style.WidthSizing == IntrinsicSizeFitContent:
		width = CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeFitContent, ctx)
	default:
		// Auto width: use the child's intrinsic size under the same constraint.
		width = CalculateIntrinsicWidth(child, Unconstrained(), sizingType, ctx)
	}
	if width < 0 || width >= Unbounded {
		width = 0
	}

	// Margins (may be negative; unbounded values are ignored).
	marginLeft := ResolveLength(child.Style.Margin.Left, ctx, fontSize)
	marginRight := ResolveLength(child.Style.Margin.Right, ctx, fontSize)
	if marginLeft > -Unbounded && marginLeft < Unbounded {
		width += marginLeft
	}
	if marginRight > -Unbounded && marginRight < Unbounded {
		width += marginRight
	}
	return width
}

// countVisibleChildren returns the number of children that participate in
// layout (display: none children generate no box and no gap).
// CSS Box Alignment Level 3 §8: gaps are placed between adjacent items;
// a display:none child is not an item. https://www.w3.org/TR/css-align-3/#gaps
func countVisibleChildren(node *Node) int {
	count := 0
	for _, child := range node.Children {
		if child.Style.Display != DisplayNone {
			count++
		}
	}
	return count
}

// totalGap returns the total space taken by gaps between count items.
func totalGap(gap Length, count int, ctx *LayoutContext, fontSize float64) float64 {
	if count <= 1 {
		return 0
	}
	gapPx := ResolveLength(gap, ctx, fontSize)
	if gapPx <= 0 || gapPx >= Unbounded {
		return 0
	}
	return gapPx * float64(count-1)
}

// calculateBlockMinContentWidth calculates min-content width for block layout.
// For block layout, this is the maximum of children's min-content contributions.
// CSS Sizing Level 3 §5.1: https://www.w3.org/TR/css-sizing-3/#intrinsic-contribution
func calculateBlockMinContentWidth(node *Node, constraints Constraints, ctx *LayoutContext) float64 {
	maxChildWidth := 0.0

	for _, child := range node.Children {
		if child.Style.Display == DisplayNone {
			continue
		}
		childWidth := childIntrinsicWidthContribution(child, IntrinsicSizeMinContent, ctx)
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
// For block layout, this is the maximum of children's max-content contributions.
func calculateBlockMaxContentWidth(node *Node, constraints Constraints, ctx *LayoutContext) float64 {
	maxChildWidth := 0.0

	for _, child := range node.Children {
		if child.Style.Display == DisplayNone {
			continue
		}
		childWidth := childIntrinsicWidthContribution(child, IntrinsicSizeMaxContent, ctx)
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
	return calculateFlexIntrinsicWidth(node, IntrinsicSizeMinContent, ctx)
}

// calculateFlexMaxContentWidth calculates max-content width for flex layout.
func calculateFlexMaxContentWidth(node *Node, constraints Constraints, ctx *LayoutContext) float64 {
	return calculateFlexIntrinsicWidth(node, IntrinsicSizeMaxContent, ctx)
}

// calculateFlexIntrinsicWidth computes the min- or max-content width of a flex
// container.
//
// CSS Flexbox Level 1 §9.9.1 (simplified): for a row container the intrinsic
// main size is the sum of the items' contributions plus gaps; for a column
// container the intrinsic cross size is the largest item contribution.
// https://www.w3.org/TR/css-flexbox-1/#intrinsic-sizes
func calculateFlexIntrinsicWidth(node *Node, sizingType IntrinsicSize, ctx *LayoutContext) float64 {
	isRow := node.Style.FlexDirection == FlexDirectionRow || node.Style.FlexDirection == FlexDirectionRowReverse
	currentFontSize := getCurrentFontSize(node, ctx)

	total := 0.0
	for _, child := range node.Children {
		if child.Style.Display == DisplayNone {
			continue
		}
		childWidth := childIntrinsicWidthContribution(child, sizingType, ctx)
		if isRow {
			total += childWidth
		} else if childWidth > total {
			total = childWidth
		}
	}

	if isRow {
		// Gaps sit between adjacent visible items only.
		gap := node.Style.FlexGap
		if node.Style.FlexColumnGap.Value > 0 {
			gap = node.Style.FlexColumnGap
		}
		total += totalGap(gap, countVisibleChildren(node), ctx, currentFontSize)
	}

	horizontalPaddingBorder := getHorizontalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, currentFontSize)
	return total + horizontalPaddingBorder
}

// textIntrinsicStyle returns the TextStyle used to measure a text node,
// defaulting to the same style LayoutText applies when none is set.
func textIntrinsicStyle(node *Node) TextStyle {
	if node.Style.TextStyle != nil {
		return *node.Style.TextStyle
	}
	return TextStyle{
		FontSize:   16,
		TextAlign:  TextAlignDefault,
		LineHeight: 0,
		WhiteSpace: WhiteSpaceNormal,
		Direction:  DirectionLTR,
	}
}

// textIntrinsicContent applies the same text preprocessing as LayoutText
// (tab expansion, white-space normalization, text-transform) so intrinsic
// measurements match the laid-out text.
func textIntrinsicContent(node *Node, style TextStyle) string {
	text := node.Text
	if style.WhiteSpace == WhiteSpaceNormal || style.WhiteSpace == WhiteSpaceNowrap {
		text = expandTabs(text, style.TabSize)
	}
	text = preprocessText(text, style.WhiteSpace)
	return applyTextTransform(text, style.TextTransform)
}

// calculateTextMaxContentWidth returns the max-content width of a text node:
// the width of its longest line when no soft wrap opportunities are taken.
//
// CSS Sizing Level 3 §4.1 / CSS Text Level 3 §4: the max-content inline size
// of text is the width of the text laid out without any soft wrapping;
// forced line breaks (preserved newlines) still split lines.
// https://www.w3.org/TR/css-sizing-3/#max-content
func calculateTextMaxContentWidth(node *Node, ctx *LayoutContext) float64 {
	style := textIntrinsicStyle(node)
	text := textIntrinsicContent(node, style)
	metrics := getTextMetrics()

	maxLine := 0.0
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			continue
		}
		width, _, _ := metrics.Measure(line, style)
		if width > maxLine {
			maxLine = width
		}
	}

	fontSize := getCurrentFontSize(node, ctx)
	return maxLine + getHorizontalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, fontSize)
}

// calculateTextMinContentWidth returns the min-content width of a text node:
// the width of its widest unbreakable segment (word).
//
// CSS Sizing Level 3 §4.1: the min-content inline size of text is the size
// it would have if every soft wrap opportunity were taken. With
// white-space: nowrap or pre there are no soft wrap opportunities, so the
// min-content size equals the max-content size.
// https://www.w3.org/TR/css-sizing-3/#min-content
func calculateTextMinContentWidth(node *Node, ctx *LayoutContext) float64 {
	style := textIntrinsicStyle(node)
	if style.WhiteSpace == WhiteSpaceNowrap || style.WhiteSpace == WhiteSpacePre {
		return calculateTextMaxContentWidth(node, ctx)
	}
	text := textIntrinsicContent(node, style)
	metrics := getTextMetrics()

	maxWord := 0.0
	for _, word := range splitIntoWords(text) {
		width, _, _ := metrics.Measure(word, style)
		if width > maxWord {
			maxWord = width
		}
	}

	fontSize := getCurrentFontSize(node, ctx)
	return maxWord + getHorizontalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, fontSize)
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
	currentFontSize := getCurrentFontSize(node, ctx)
	gap := node.Style.GridGap
	if node.Style.GridColumnGap.Value > 0 {
		gap = node.Style.GridColumnGap
	}
	totalWidth += totalGap(gap, len(node.Style.GridTemplateColumns), ctx, currentFontSize)

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
	currentFontSize := getCurrentFontSize(node, ctx)
	gap := node.Style.GridGap
	if node.Style.GridColumnGap.Value > 0 {
		gap = node.Style.GridColumnGap
	}
	totalWidth += totalGap(gap, len(node.Style.GridTemplateColumns), ctx, currentFontSize)

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
