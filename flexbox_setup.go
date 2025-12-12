package layout

// flexboxSetup contains the setup state for flexbox layout
// Algorithm based on CSS Flexible Box Layout Module Level 1: §9.2: Line Length Determination
type flexboxSetup struct {
	// Container dimensions
	horizontalPadding float64
	verticalPadding   float64
	horizontalBorder  float64
	verticalBorder    float64
	contentWidth      float64
	contentHeight     float64

	// Axis determination
	isRow              bool
	mainSize           float64
	crossSize          float64
	hasExplicitMainSize   bool
	hasExplicitCrossSize  bool
}

// flexboxDetermineLineLength initializes the flexbox layout state and determines line length.
//
// Algorithm based on CSS Flexible Box Layout Module Level 1:
// - §9.2: Line Length Determination
//
// See: https://www.w3.org/TR/css-flexbox-1/#line-sizing
func flexboxDetermineLineLength(node *Node, constraints Constraints) flexboxSetup {
	setup := flexboxSetup{}

	// Calculate available space
	// For tight constraints, both Min and Max are set to the same value
	// Prefer Max, fall back to Min if Max is 0
	availableWidth := constraints.MaxWidth
	if availableWidth == 0 {
		availableWidth = constraints.MinWidth
	}
	availableHeight := constraints.MaxHeight
	if availableHeight == 0 {
		availableHeight = constraints.MinHeight
	}

	// Account for padding and border
	setup.horizontalPadding = node.Style.Padding.Left + node.Style.Padding.Right
	setup.verticalPadding = node.Style.Padding.Top + node.Style.Padding.Bottom
	setup.horizontalBorder = node.Style.Border.Left + node.Style.Border.Right
	setup.verticalBorder = node.Style.Border.Top + node.Style.Border.Bottom

	// Check for intrinsic sizing (min-content, max-content, fit-content)
	// These override auto sizing and explicit dimensions
	constraintsForIntrinsic := Loose(availableWidth, availableHeight)

	// Handle width intrinsic sizing
	intrinsicWidth := -1.0
	if node.Style.Width == SizeMinContent || node.Style.WidthSizing == IntrinsicSizeMinContent {
		intrinsicWidth = CalculateIntrinsicWidth(node, constraintsForIntrinsic, IntrinsicSizeMinContent)
	} else if node.Style.Width == SizeMaxContent || node.Style.WidthSizing == IntrinsicSizeMaxContent {
		intrinsicWidth = CalculateIntrinsicWidth(node, constraintsForIntrinsic, IntrinsicSizeMaxContent)
	} else if node.Style.Width == SizeFitContent || node.Style.WidthSizing == IntrinsicSizeFitContent {
		intrinsicWidth = CalculateIntrinsicWidth(node, constraintsForIntrinsic, IntrinsicSizeFitContent)
	}

	// Handle height intrinsic sizing
	intrinsicHeight := -1.0
	if node.Style.Height == SizeMinContent || node.Style.HeightSizing == IntrinsicSizeMinContent {
		intrinsicHeight = CalculateIntrinsicHeight(node, constraintsForIntrinsic, IntrinsicSizeMinContent)
	} else if node.Style.Height == SizeMaxContent || node.Style.HeightSizing == IntrinsicSizeMaxContent {
		intrinsicHeight = CalculateIntrinsicHeight(node, constraintsForIntrinsic, IntrinsicSizeMaxContent)
	} else if node.Style.Height == SizeFitContent || node.Style.HeightSizing == IntrinsicSizeFitContent {
		intrinsicHeight = CalculateIntrinsicHeight(node, constraintsForIntrinsic, IntrinsicSizeFitContent)
	}

	// Apply intrinsic width if calculated
	if intrinsicWidth > 0 {
		totalIntrinsicWidth := intrinsicWidth + setup.horizontalPadding + setup.horizontalBorder
		if availableWidth >= Unbounded || availableWidth == 0 {
			availableWidth = totalIntrinsicWidth
		} else if totalIntrinsicWidth <= availableWidth {
			availableWidth = totalIntrinsicWidth
		}
	} else if node.Style.Width > 0 {
		// If container has explicit width/height, use it to constrain available space
		// Similar to grid layout
		// If constraints are zero/unbounded and we have explicit dimensions, use the explicit dimensions
		// Only use explicit width if it's > 0 (not auto/unspecified)
		specifiedWidthContent := convertToContentSize(node.Style.Width, node.Style.BoxSizing, setup.horizontalPadding+setup.horizontalBorder, setup.verticalPadding+setup.verticalBorder, true)
		totalSpecifiedWidth := specifiedWidthContent + setup.horizontalPadding + setup.horizontalBorder
		if availableWidth >= Unbounded || availableWidth == 0 {
			// No meaningful constraint -> use style width
			availableWidth = totalSpecifiedWidth
		} else if totalSpecifiedWidth <= availableWidth {
			availableWidth = totalSpecifiedWidth
		}
	}

	// Apply intrinsic height if calculated
	if intrinsicHeight > 0 {
		totalIntrinsicHeight := intrinsicHeight + setup.verticalPadding + setup.verticalBorder
		if availableHeight >= Unbounded || availableHeight == 0 {
			availableHeight = totalIntrinsicHeight
		} else if totalIntrinsicHeight <= availableHeight {
			availableHeight = totalIntrinsicHeight
		}
	} else if node.Style.Height > 0 {
		// Only use explicit height if it's > 0 (not auto/unspecified)
		specifiedHeightContent := convertToContentSize(node.Style.Height, node.Style.BoxSizing, setup.horizontalPadding+setup.horizontalBorder, setup.verticalPadding+setup.verticalBorder, false)
		totalSpecifiedHeight := specifiedHeightContent + setup.verticalPadding + setup.verticalBorder
		if availableHeight >= Unbounded || availableHeight == 0 {
			// No meaningful constraint -> use style height
			availableHeight = totalSpecifiedHeight
		} else if totalSpecifiedHeight <= availableHeight {
			availableHeight = totalSpecifiedHeight
		}
	}

	// Clamp content size to >= 0
	setup.contentWidth = availableWidth - setup.horizontalPadding - setup.horizontalBorder
	if setup.contentWidth < 0 {
		setup.contentWidth = 0
	}
	setup.contentHeight = availableHeight - setup.verticalPadding - setup.verticalBorder
	if setup.contentHeight < 0 {
		setup.contentHeight = 0
	}

	// Determine main and cross axis
	setup.isRow = node.Style.FlexDirection == FlexDirectionRow || node.Style.FlexDirection == FlexDirectionRowReverse
	setup.mainSize = setup.contentWidth
	setup.crossSize = setup.contentHeight
	if !setup.isRow {
		setup.mainSize, setup.crossSize = setup.crossSize, setup.mainSize
	}

	// Does this container have a definite cross size (via style or constraints)?
	// This determines whether align-items: stretch should use crossSize or content-driven lineCrossSize
	if setup.isRow {
		// Cross axis is height
		if node.Style.Height > 0 || (constraints.MaxHeight > 0 && constraints.MaxHeight < Unbounded) {
			setup.hasExplicitCrossSize = true
		}
	} else {
		// Cross axis is width
		if node.Style.Width > 0 || (constraints.MaxWidth > 0 && constraints.MaxWidth < Unbounded) {
			setup.hasExplicitCrossSize = true
		}
	}

	// Does this container have a definite main size (via style or constraints)?
	// If not, free-space distribution in the main axis should be skipped (CSS flexbox §9.2 / §9.3).
	if setup.isRow {
		// Main axis is width
		if node.Style.Width > 0 || (constraints.MaxWidth > 0 && constraints.MaxWidth < Unbounded) {
			setup.hasExplicitMainSize = true
		}
	} else {
		// Main axis is height
		if node.Style.Height > 0 || (constraints.MaxHeight > 0 && constraints.MaxHeight < Unbounded) {
			setup.hasExplicitMainSize = true
		}
	}

	return setup
}
