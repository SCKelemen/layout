package layout

// flexboxSetup contains the setup state for flexbox layout
// Algorithm based on CSS Flexible Box Layout Module Level 1: §9.2: Line Length Determination
// Extended with CSS Writing Modes Level 3 support
type flexboxSetup struct {
	// Container dimensions
	horizontalPadding float64
	verticalPadding   float64
	horizontalBorder  float64
	verticalBorder    float64
	contentWidth      float64
	contentHeight     float64

	// Axis determination (considering both flex-direction and writing-mode)
	isRow                bool // True if flex-direction is row/row-reverse
	isMainHorizontal     bool // True if main axis runs horizontally (physical)
	isReverse            bool // True if flex-direction is *-reverse
	mainSize             float64
	crossSize            float64
	hasExplicitMainSize  bool
	hasExplicitCrossSize bool
	writingMode          WritingMode
}

// flexboxDetermineLineLength initializes the flexbox layout state and determines line length.
//
// Algorithm based on CSS Flexible Box Layout Module Level 1:
// - §9.2: Line Length Determination
//
// See: https://www.w3.org/TR/css-flexbox-1/#line-sizing
func flexboxDetermineLineLength(node *Node, constraints Constraints, ctx *LayoutContext) flexboxSetup {
	setup := flexboxSetup{}

	// Get current font size for Length resolution
	fontSize := getCurrentFontSize(node, ctx)

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

	// Account for padding and border (resolve Length to pixels)
	setup.horizontalPadding = ResolveLength(node.Style.Padding.Left, ctx, fontSize) + ResolveLength(node.Style.Padding.Right, ctx, fontSize)
	setup.verticalPadding = ResolveLength(node.Style.Padding.Top, ctx, fontSize) + ResolveLength(node.Style.Padding.Bottom, ctx, fontSize)
	setup.horizontalBorder = ResolveLength(node.Style.Border.Left, ctx, fontSize) + ResolveLength(node.Style.Border.Right, ctx, fontSize)
	setup.verticalBorder = ResolveLength(node.Style.Border.Top, ctx, fontSize) + ResolveLength(node.Style.Border.Bottom, ctx, fontSize)

	// Check for intrinsic sizing (min-content, max-content, fit-content)
	// These override auto sizing and explicit dimensions
	constraintsForIntrinsic := Loose(availableWidth, availableHeight)

	// Handle width intrinsic sizing
	intrinsicWidth := -1.0
	if node.Style.Width.Value == SizeMinContent || node.Style.WidthSizing == IntrinsicSizeMinContent {
		intrinsicWidth = CalculateIntrinsicWidth(node, constraintsForIntrinsic, IntrinsicSizeMinContent, ctx)
	} else if node.Style.Width.Value == SizeMaxContent || node.Style.WidthSizing == IntrinsicSizeMaxContent {
		intrinsicWidth = CalculateIntrinsicWidth(node, constraintsForIntrinsic, IntrinsicSizeMaxContent, ctx)
	} else if node.Style.Width.Value == SizeFitContent || node.Style.WidthSizing == IntrinsicSizeFitContent {
		intrinsicWidth = CalculateIntrinsicWidth(node, constraintsForIntrinsic, IntrinsicSizeFitContent, ctx)
	}

	// Handle height intrinsic sizing
	intrinsicHeight := -1.0
	if node.Style.Height.Value == SizeMinContent || node.Style.HeightSizing == IntrinsicSizeMinContent {
		intrinsicHeight = CalculateIntrinsicHeight(node, constraintsForIntrinsic, IntrinsicSizeMinContent, ctx)
	} else if node.Style.Height.Value == SizeMaxContent || node.Style.HeightSizing == IntrinsicSizeMaxContent {
		intrinsicHeight = CalculateIntrinsicHeight(node, constraintsForIntrinsic, IntrinsicSizeMaxContent, ctx)
	} else if node.Style.Height.Value == SizeFitContent || node.Style.HeightSizing == IntrinsicSizeFitContent {
		intrinsicHeight = CalculateIntrinsicHeight(node, constraintsForIntrinsic, IntrinsicSizeFitContent, ctx)
	}

	// Apply intrinsic width if calculated
	if intrinsicWidth > 0 {
		totalIntrinsicWidth := intrinsicWidth + setup.horizontalPadding + setup.horizontalBorder
		if availableWidth >= Unbounded || availableWidth == 0 {
			availableWidth = totalIntrinsicWidth
		} else if totalIntrinsicWidth <= availableWidth {
			availableWidth = totalIntrinsicWidth
		}
	} else if node.Style.Width.Value > 0 {
		// If container has explicit width/height, use it to constrain available space
		// Similar to grid layout
		// If constraints are zero/unbounded and we have explicit dimensions, use the explicit dimensions
		// Only use explicit width if it's > 0 (not auto/unspecified)
		resolvedWidth := ResolveLength(node.Style.Width, ctx, fontSize)
		specifiedWidthContent := convertToContentSize(resolvedWidth, node.Style.BoxSizing, setup.horizontalPadding+setup.horizontalBorder, setup.verticalPadding+setup.verticalBorder, true)
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
	} else if node.Style.Height.Value > 0 {
		// Only use explicit height if it's > 0 (not auto/unspecified)
		resolvedHeight := ResolveLength(node.Style.Height, ctx, fontSize)
		specifiedHeightContent := convertToContentSize(resolvedHeight, node.Style.BoxSizing, setup.horizontalPadding+setup.horizontalBorder, setup.verticalPadding+setup.verticalBorder, false)
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

	// Determine main and cross axis considering both flex-direction and writing-mode
	// Based on CSS Writing Modes Level 3 and CSS Flexbox Level 1
	//
	// In horizontal writing modes (horizontal-tb):
	//   - row/row-reverse: main axis = horizontal (inline direction)
	//   - column/column-reverse: main axis = vertical (block direction)
	//
	// In vertical writing modes (vertical-lr, vertical-rl, sideways-*):
	//   - row/row-reverse: main axis = vertical (inline direction)
	//   - column/column-reverse: main axis = horizontal (block direction)
	setup.writingMode = node.Style.WritingMode
	setup.isRow = node.Style.FlexDirection == FlexDirectionRow || node.Style.FlexDirection == FlexDirectionRowReverse
	setup.isReverse = node.Style.FlexDirection == FlexDirectionRowReverse || node.Style.FlexDirection == FlexDirectionColumnReverse

	isVerticalWritingMode := setup.writingMode.IsVertical()

	// Determine physical direction of main axis
	if setup.isRow {
		// flex-direction: row/row-reverse
		// Main axis follows inline direction
		setup.isMainHorizontal = !isVerticalWritingMode
	} else {
		// flex-direction: column/column-reverse
		// Main axis follows block direction
		setup.isMainHorizontal = isVerticalWritingMode
	}

	// Set main and cross sizes based on physical axis direction
	if setup.isMainHorizontal {
		setup.mainSize = setup.contentWidth
		setup.crossSize = setup.contentHeight
	} else {
		setup.mainSize = setup.contentHeight
		setup.crossSize = setup.contentWidth
	}

	// Does this container have a definite cross size (via style or constraints)?
	// This determines whether align-items: stretch should use crossSize or content-driven lineCrossSize
	// Check based on physical cross axis direction
	if setup.isMainHorizontal {
		// Main is horizontal, so cross axis is vertical (height)
		if node.Style.Height.Value > 0 || (constraints.MaxHeight > 0 && constraints.MaxHeight < Unbounded) {
			setup.hasExplicitCrossSize = true
		}
	} else {
		// Main is vertical, so cross axis is horizontal (width)
		if node.Style.Width.Value > 0 || (constraints.MaxWidth > 0 && constraints.MaxWidth < Unbounded) {
			setup.hasExplicitCrossSize = true
		}
	}

	// Does this container have a definite main size (via style or constraints)?
	// If not, free-space distribution in the main axis should be skipped (CSS flexbox §9.2 / §9.3).
	// Check based on physical main axis direction
	if setup.isMainHorizontal {
		// Main axis is horizontal (width)
		if node.Style.Width.Value > 0 || (constraints.MaxWidth > 0 && constraints.MaxWidth < Unbounded) {
			setup.hasExplicitMainSize = true
		}
	} else {
		// Main axis is vertical (height)
		if node.Style.Height.Value > 0 || (constraints.MaxHeight > 0 && constraints.MaxHeight < Unbounded) {
			setup.hasExplicitMainSize = true
		}
	}

	return setup
}
