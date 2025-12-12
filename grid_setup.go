package layout

// gridSetup contains the setup state for grid layout
// Algorithm based on CSS Grid Layout Module Level 1: §11: Grid Sizing
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
}

// gridDetermineContainerSize initializes the grid layout state and determines container dimensions.
//
// Algorithm based on CSS Grid Layout Module Level 1:
// - §11: Grid Sizing
// - §11.1: Track Sizing Algorithm
//
// See: https://www.w3.org/TR/css-grid-1/#track-sizing
func gridDetermineContainerSize(node *Node, constraints Constraints) gridSetup {
	setup := gridSetup{}

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

	// If container has explicit width/height, use it to constrain available space
	if node.Style.Width >= 0 {
		specifiedWidthContent := convertToContentSize(node.Style.Width, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, true)
		totalSpecifiedWidth := specifiedWidthContent + setup.horizontalPaddingBorder
		if availableWidth >= Unbounded {
			availableWidth = totalSpecifiedWidth
		} else if totalSpecifiedWidth <= availableWidth {
			availableWidth = totalSpecifiedWidth
		}
	}

	if node.Style.Height >= 0 {
		specifiedHeightContent := convertToContentSize(node.Style.Height, node.Style.BoxSizing, setup.horizontalPaddingBorder, setup.verticalPaddingBorder, false)
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
		if len(setup.rows) == 0 || (setup.rows[0].MinSize == 0 && setup.rows[0].MaxSize == Unbounded && setup.rows[0].Fraction == 0) {
			setup.rows = []GridTrack{AutoTrack()}
		}
	}
	if len(setup.columns) == 0 {
		setup.columns = []GridTrack{node.Style.GridAutoColumns}
		if len(setup.columns) == 0 || (setup.columns[0].MinSize == 0 && setup.columns[0].MaxSize == Unbounded && setup.columns[0].Fraction == 0) {
			setup.columns = []GridTrack{AutoTrack()}
		}
	}

	// Calculate gap
	setup.rowGap = node.Style.GridRowGap
	if setup.rowGap == 0 {
		setup.rowGap = node.Style.GridGap
	}
	setup.columnGap = node.Style.GridColumnGap
	if setup.columnGap == 0 {
		setup.columnGap = node.Style.GridGap
	}

	return setup
}
