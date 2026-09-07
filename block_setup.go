package layout

// blockSetup contains the setup state for block layout
// Algorithm based on CSS Box Model Module Level 3: §4: Box Model
type blockSetup struct {
	// Container dimensions
	horizontalPaddingBorder float64
	verticalPaddingBorder   float64
	contentWidth            float64
	contentHeight           float64

	// Specified dimensions (in content-box units)
	specifiedWidth  float64
	specifiedHeight float64
	isAutoWidth     bool
	isAutoHeight    bool

	// Constraints
	minWidthContent  float64
	maxWidthContent  float64
	minHeightContent float64
	maxHeightContent float64
}

// blockDetermineContainerSize initializes the block layout state and determines container dimensions.
//
// Algorithm based on CSS Box Model Module Level 3:
// - §4: Box Model
// - §4.2: Margins
//
// CSS Display Module Level 3:
// - §4: Block-level Boxes
//
// See: https://www.w3.org/TR/css-box-3/#box-model
// See: https://www.w3.org/TR/css-display-3/#block-level
func blockDetermineContainerSize(node *Node, constraints Constraints, ctx *LayoutContext, currentFontSize float64) blockSetup {
	setup := blockSetup{}

	// Calculate available space
	availableWidth := constraints.MaxWidth
	availableHeight := constraints.MaxHeight

	// Resolve padding and border to pixels. Only the per-axis sums are needed:
	// the content box is the border box minus padding and border on each axis
	// (CSS Box Model Level 3 §4, https://www.w3.org/TR/css-box-3/#box-model).
	setup.horizontalPaddingBorder = getHorizontalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, currentFontSize)
	setup.verticalPaddingBorder = getVerticalPaddingBorder(node.Style.Padding, node.Style.Border, ctx, currentFontSize)

	// Clamp content size to >= 0
	setup.contentWidth = availableWidth - setup.horizontalPaddingBorder
	if setup.contentWidth < 0 {
		setup.contentWidth = 0
	}
	setup.contentHeight = availableHeight - setup.verticalPaddingBorder
	if setup.contentHeight < 0 {
		setup.contentHeight = 0
	}

	// Resolve width/height to pixels
	widthValue := ResolveLength(node.Style.Width, ctx, currentFontSize)
	heightValue := ResolveLength(node.Style.Height, ctx, currentFontSize)

	// Convert width/height from specified box-sizing to content-box for internal calculations
	// According to W3C CSS Box Sizing spec:
	// - content-box: width/height = content size only
	// - border-box: width/height = content + padding + border
	setup.specifiedWidth = convertToContentSize(widthValue, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, true)
	setup.specifiedHeight = convertToContentSize(heightValue, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, false)

	// Determine if dimensions are auto.
	//
	// A dimension is auto when:
	//   - the Length is the Go zero value (Unit == ""), i.e. the caller never
	//     set Width/Height. This matches the flexbox and grid conventions,
	//     where an unspecified size is "auto", not "0px". An explicit Px(0)
	//     remains a real zero size. See CSS 2.1 §10.3.3: the initial value of
	//     width is auto (https://www.w3.org/TR/CSS21/visudet.html#blockwidth).
	//   - the resolved value is negative (the Px(-1) auto sentinel), or
	//   - the value is 0 while an aspect ratio is set and the other dimension
	//     is also 0 (legacy heuristic kept for callers that use Px(0) + AspectRatio).
	setup.isAutoWidth = isUnsetLength(node.Style.Width) || setup.specifiedWidth < 0 ||
		(setup.specifiedWidth == 0 && node.Style.AspectRatio > 0 && setup.specifiedHeight == 0)
	setup.isAutoHeight = isUnsetLength(node.Style.Height) || setup.specifiedHeight < 0 ||
		(setup.specifiedHeight == 0 && node.Style.AspectRatio > 0 && setup.specifiedWidth == 0)

	// Resolve min/max constraints to pixels
	minWidthValue := ResolveLength(node.Style.MinWidth, ctx, currentFontSize)
	maxWidthValue := ResolveLength(node.Style.MaxWidth, ctx, currentFontSize)
	minHeightValue := ResolveLength(node.Style.MinHeight, ctx, currentFontSize)
	maxHeightValue := ResolveLength(node.Style.MaxHeight, ctx, currentFontSize)

	// Apply min/max constraints
	// Min/Max constraints also respect box-sizing (they apply to the same box as width/height)
	setup.minWidthContent = convertMinMaxToContentSize(minWidthValue, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, true)
	setup.maxWidthContent = convertMinMaxToContentSize(maxWidthValue, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, true)
	setup.minHeightContent = convertMinMaxToContentSize(minHeightValue, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, false)
	setup.maxHeightContent = convertMinMaxToContentSize(maxHeightValue, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, false)

	return setup
}

// isUnsetLength reports whether a Length is the Go zero value, meaning the
// caller never assigned it. Block layout treats an unset Width/Height as
// "auto" (the CSS initial value), while an explicit Px(0) stays a real zero.
// Only Unit is inspected: a Length constructed via Px/Em/... always carries
// a unit, so Unit == "" can only come from a zero-value struct.
func isUnsetLength(l Length) bool {
	return l.Unit == ""
}

// clampMinMax applies a max constraint followed by a min constraint, so that
// when min > max the min value wins.
//
// CSS 2.1 §10.4: "If the computed value of min-width is greater than the value
// of max-width, max-width is set to the value of min-width."
// https://www.w3.org/TR/CSS21/visudet.html#min-max-widths
//
// A max of 0 or >= Unbounded means "none"; a min <= 0 is a no-op.
func clampMinMax(value, minValue, maxValue float64) float64 {
	if maxValue > 0 && maxValue < Unbounded && value > maxValue {
		value = maxValue
	}
	if minValue > 0 && value < minValue {
		value = minValue
	}
	return value
}
