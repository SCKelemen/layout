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
func blockDetermineContainerSize(node *Node, constraints Constraints) blockSetup {
	setup := blockSetup{}

	// Calculate available space
	availableWidth := constraints.MaxWidth
	availableHeight := constraints.MaxHeight

	// Account for padding and border
	setup.horizontalPadding = node.Style.Padding.Left + node.Style.Padding.Right
	setup.verticalPadding = node.Style.Padding.Top + node.Style.Padding.Bottom
	setup.horizontalBorder = node.Style.Border.Left + node.Style.Border.Right
	setup.verticalBorder = node.Style.Border.Top + node.Style.Border.Bottom
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

	// Convert width/height from specified box-sizing to content-box for internal calculations
	// According to W3C CSS Box Sizing spec:
	// - content-box: width/height = content size only
	// - border-box: width/height = content + padding + border
	setup.specifiedWidth = convertToContentSize(node.Style.Width, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, true)
	setup.specifiedHeight = convertToContentSize(node.Style.Height, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, false)

	// Determine if dimensions are auto
	// CRITICAL FIX: Treat 0 as auto when aspect ratio is set (Go zero value issue)
	setup.isAutoWidth = setup.specifiedWidth < 0 || (setup.specifiedWidth == 0 && node.Style.AspectRatio > 0 && setup.specifiedHeight == 0)
	setup.isAutoHeight = setup.specifiedHeight < 0 || (setup.specifiedHeight == 0 && node.Style.AspectRatio > 0 && setup.specifiedWidth == 0)

	// Apply min/max constraints
	// Min/Max constraints also respect box-sizing (they apply to the same box as width/height)
	setup.minWidthContent = convertMinMaxToContentSize(node.Style.MinWidth, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, true)
	setup.maxWidthContent = convertMinMaxToContentSize(node.Style.MaxWidth, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, true)
	setup.minHeightContent = convertMinMaxToContentSize(node.Style.MinHeight, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, false)
	setup.maxHeightContent = convertMinMaxToContentSize(node.Style.MaxHeight, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, false)

	return setup
}
