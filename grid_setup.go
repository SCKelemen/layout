package layout

// gridSetup contains the setup state for grid layout
// Algorithm based on CSS Grid Layout Module Level 1: §11: Grid Sizing
// Extended with CSS Writing Modes Level 3 support
type gridSetup struct {
	// Container dimensions
	horizontalPadding       float64
	verticalPadding         float64
	horizontalBorder        float64
	verticalBorder          float64
	horizontalPaddingBorder float64
	verticalPaddingBorder   float64
	contentWidth            float64
	contentHeight           float64

	// Grid tracks
	rows    []GridTrack
	columns []GridTrack

	// Grid gaps
	rowGap    float64
	columnGap float64

	// Writing mode support (considering both grid structure and writing-mode)
	writingMode     WritingMode
	isRowBlockAxis  bool // True if rows run in block direction (horizontal-tb: false, vertical-lr: true)
	blockAxisSize   float64 // Content size in block dimension
	inlineAxisSize  float64 // Content size in inline dimension
}

// gridDetermineContainerSize initializes the grid layout state and determines container dimensions.
//
// Algorithm based on CSS Grid Layout Module Level 1:
// - §11: Grid Sizing
// - §11.1: Track Sizing Algorithm
//
// See: https://www.w3.org/TR/css-grid-1/#track-sizing
func gridDetermineContainerSize(node *Node, constraints Constraints, ctx *LayoutContext, currentFontSize float64) gridSetup {
	setup := gridSetup{}

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

	setup.horizontalPadding = paddingLeft + paddingRight
	setup.verticalPadding = paddingTop + paddingBottom
	setup.horizontalBorder = borderLeft + borderRight
	setup.verticalBorder = borderTop + borderBottom
	setup.horizontalPaddingBorder = setup.horizontalPadding + setup.horizontalBorder
	setup.verticalPaddingBorder = setup.verticalPadding + setup.verticalBorder

	// If container has explicit width/height, use it to constrain available space
	widthValue := ResolveLength(node.Style.Width, ctx, currentFontSize)
	if widthValue >= 0 {
		specifiedWidthContent := convertToContentSize(widthValue, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, true)
		totalSpecifiedWidth := specifiedWidthContent + setup.horizontalPaddingBorder
		if availableWidth >= Unbounded {
			availableWidth = totalSpecifiedWidth
		} else if totalSpecifiedWidth <= availableWidth {
			availableWidth = totalSpecifiedWidth
		}
	}

	heightValue := ResolveLength(node.Style.Height, ctx, currentFontSize)
	if heightValue >= 0 {
		specifiedHeightContent := convertToContentSize(heightValue, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, false)
		totalSpecifiedHeight := specifiedHeightContent + setup.verticalPaddingBorder
		if availableHeight >= Unbounded {
			availableHeight = totalSpecifiedHeight
		} else if totalSpecifiedHeight < availableHeight {
			availableHeight = totalSpecifiedHeight
		}
	}

	// Clamp content size to >= 0
	setup.contentWidth = availableWidth - setup.horizontalPaddingBorder
	if setup.contentWidth < 0 {
		setup.contentWidth = 0
	}
	setup.contentHeight = availableHeight - setup.verticalPaddingBorder
	if setup.contentHeight < 0 {
		setup.contentHeight = 0
	}

	// Get grid template
	setup.rows = node.Style.GridTemplateRows
	setup.columns = node.Style.GridTemplateColumns

	// Use auto tracks if templates not specified
	if len(setup.rows) == 0 {
		setup.rows = []GridTrack{node.Style.GridAutoRows}
		// Check if it's an auto track by resolving the sizes
		if len(setup.rows) > 0 {
			minSize := ResolveLength(setup.rows[0].MinSize, ctx, currentFontSize)
			maxSize := ResolveLength(setup.rows[0].MaxSize, ctx, currentFontSize)
			if minSize == 0 && maxSize >= Unbounded && setup.rows[0].Fraction == 0 {
				setup.rows = []GridTrack{AutoTrack()}
			}
		}
	}
	if len(setup.columns) == 0 {
		setup.columns = []GridTrack{node.Style.GridAutoColumns}
		// Check if it's an auto track by resolving the sizes
		if len(setup.columns) > 0 {
			minSize := ResolveLength(setup.columns[0].MinSize, ctx, currentFontSize)
			maxSize := ResolveLength(setup.columns[0].MaxSize, ctx, currentFontSize)
			if minSize == 0 && maxSize >= Unbounded && setup.columns[0].Fraction == 0 {
				setup.columns = []GridTrack{AutoTrack()}
			}
		}
	}

	// Resolve gaps to pixels
	rowGapResolved := ResolveLength(node.Style.GridRowGap, ctx, currentFontSize)
	if rowGapResolved == 0 {
		setup.rowGap = ResolveLength(node.Style.GridGap, ctx, currentFontSize)
	} else {
		setup.rowGap = rowGapResolved
	}
	columnGapResolved := ResolveLength(node.Style.GridColumnGap, ctx, currentFontSize)
	if columnGapResolved == 0 {
		setup.columnGap = ResolveLength(node.Style.GridGap, ctx, currentFontSize)
	} else {
		setup.columnGap = columnGapResolved
	}

	// Determine writing mode and axis mapping
	// Based on CSS Writing Modes Level 3 and CSS Grid Layout Level 1
	//
	// In horizontal writing modes (horizontal-tb):
	//   - Rows run horizontally, controlling vertical positioning (Y axis)
	//   - Columns run vertically, controlling horizontal positioning (X axis)
	//   - Block axis = vertical, Inline axis = horizontal
	//   - isRowBlockAxis = false (rows are NOT in block direction)
	//
	// In vertical writing modes (vertical-lr, vertical-rl):
	//   - Rows run vertically, controlling horizontal positioning (X axis)
	//   - Columns run horizontally, controlling vertical positioning (Y axis)
	//   - Block axis = horizontal, Inline axis = vertical
	//   - isRowBlockAxis = true (rows ARE in block direction)
	setup.writingMode = node.Style.WritingMode
	isVerticalWritingMode := setup.writingMode.IsVertical()

	// In vertical writing modes, rows and columns swap their physical meaning
	setup.isRowBlockAxis = isVerticalWritingMode

	// Set block and inline axis sizes based on writing mode
	if isVerticalWritingMode {
		// Vertical mode: block = horizontal, inline = vertical
		setup.blockAxisSize = setup.contentWidth
		setup.inlineAxisSize = setup.contentHeight
	} else {
		// Horizontal mode: block = vertical, inline = horizontal
		setup.blockAxisSize = setup.contentHeight
		setup.inlineAxisSize = setup.contentWidth
	}

	return setup
}
