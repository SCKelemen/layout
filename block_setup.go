package layout

// blockSetup contains the setup state for block layout
// Algorithm based on CSS Box Model Module Level 3: §4: Box Model
type blockSetup struct {
	// Container dimensions
	horizontalPadding       float64
	verticalPadding         float64
	horizontalBorder        float64
	verticalBorder          float64
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

	// Resolve padding and border to pixels
	paddingLeft := ResolveLength(node.Style.Padding.Left, ctx, currentFontSize)
	paddingRight := ResolveLength(node.Style.Padding.Right, ctx, currentFontSize)
	paddingTop := ResolveLength(node.Style.Padding.Top, ctx, currentFontSize)
	paddingBottom := ResolveLength(node.Style.Padding.Bottom, ctx, currentFontSize)
	borderLeft := ResolveLength(node.Style.Border.Left, ctx, currentFontSize)
	borderRight := ResolveLength(node.Style.Border.Right, ctx, currentFontSize)
	borderTop := ResolveLength(node.Style.Border.Top, ctx, currentFontSize)
	borderBottom := ResolveLength(node.Style.Border.Bottom, ctx, currentFontSize)

	// Account for padding and border
	setup.horizontalPadding = paddingLeft + paddingRight
	setup.verticalPadding = paddingTop + paddingBottom
	setup.horizontalBorder = borderLeft + borderRight
	setup.verticalBorder = borderTop + borderBottom
	setup.horizontalPaddingBorder = setup.horizontalPadding + setup.horizontalBorder
	setup.verticalPaddingBorder = setup.verticalPadding + setup.verticalBorder

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

	// Determine if dimensions are auto
	// CRITICAL FIX: Treat 0 as auto when aspect ratio is set (Go zero value issue)
	setup.isAutoWidth = setup.specifiedWidth < 0 || (setup.specifiedWidth == 0 && node.Style.AspectRatio > 0 && setup.specifiedHeight == 0)
	setup.isAutoHeight = setup.specifiedHeight < 0 || (setup.specifiedHeight == 0 && node.Style.AspectRatio > 0 && setup.specifiedWidth == 0)

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
