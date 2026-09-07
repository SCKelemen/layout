package layout

import (
	"math"
)

// LayoutGrid performs CSS Grid layout on a node.
//
// Algorithm based on CSS Grid Layout Module Level 1:
// - §11: Grid Sizing
//   - §11.1: Track Sizing Algorithm
//   - §11.2: Track Sizing Algorithm for Grid Containers
//   - §11.3: Track Sizing Algorithm for Grid Items
//
// - §12: Grid Item Placement
//   - §12.1: Grid Item Placement Algorithm
//
// - §10: Alignment
//   - §10.1: Aligning with justify-items
//   - §10.2: Aligning with align-items
//
// Writing modes: in a vertical writing mode the grid's rows run along the
// physical X axis and its columns along the physical Y axis, so tracks are
// sized against the matching physical dimension and item positions are
// swapped when written back (CSS Writing Modes Level 3 §7.1).
//
// See: https://www.w3.org/TR/css-grid-1/
// See: https://www.w3.org/TR/css-writing-modes-3/#logical-to-physical
func LayoutGrid(node *Node, constraints Constraints, ctx *LayoutContext) Size {
	if node.Style.Display != DisplayGrid {
		// If not grid, delegate to block layout
		return LayoutBlock(node, constraints, ctx)
	}

	// Get current font size for em unit resolution
	currentFontSize := 16.0 // Default
	if node.Style.TextStyle != nil && node.Style.TextStyle.FontSize > 0 {
		currentFontSize = node.Style.TextStyle.FontSize
	}

	// Account for padding and border - resolve Length values
	paddingLeft := ResolveLength(node.Style.Padding.Left, ctx, currentFontSize)
	paddingRight := ResolveLength(node.Style.Padding.Right, ctx, currentFontSize)
	paddingTop := ResolveLength(node.Style.Padding.Top, ctx, currentFontSize)
	paddingBottom := ResolveLength(node.Style.Padding.Bottom, ctx, currentFontSize)
	borderLeft := ResolveLength(node.Style.Border.Left, ctx, currentFontSize)
	borderRight := ResolveLength(node.Style.Border.Right, ctx, currentFontSize)
	borderTop := ResolveLength(node.Style.Border.Top, ctx, currentFontSize)
	borderBottom := ResolveLength(node.Style.Border.Bottom, ctx, currentFontSize)

	horizontalPaddingBorder := paddingLeft + paddingRight + borderLeft + borderRight
	verticalPaddingBorder := paddingTop + paddingBottom + borderTop + borderBottom

	// Determine the container's content size in each physical axis and
	// whether that size is definite (needed for fr distribution and for
	// align-content free space).
	contentWidth, widthDefinite, widthSpecified := gridResolveContainerAxis(node, true, constraints, horizontalPaddingBorder, verticalPaddingBorder, ctx, currentFontSize)
	contentHeight, heightDefinite, heightSpecified := gridResolveContainerAxis(node, false, constraints, horizontalPaddingBorder, verticalPaddingBorder, ctx, currentFontSize)

	// Determine writing mode for grid positioning
	// Based on CSS Writing Modes Level 3 and CSS Grid Layout Level 1
	writingMode := node.Style.WritingMode
	isVerticalWritingMode := writingMode.IsVertical()

	// Logical axes: the column axis is the one columns are laid out along
	// (physical X in horizontal-tb, physical Y in vertical modes), and the
	// row axis is the other one.
	colAxisSize, colAxisDefinite := contentWidth, widthDefinite
	rowAxisSize, rowAxisDefinite := contentHeight, heightDefinite
	if isVerticalWritingMode {
		colAxisSize, colAxisDefinite = contentHeight, heightDefinite
		rowAxisSize, rowAxisDefinite = contentWidth, widthDefinite
	}

	// Get grid template (copied so implicit tracks never alias the style's
	// backing array). Missing templates get a single implicit track.
	rows := gridTemplateTracks(node.Style.GridTemplateRows, node.Style.GridAutoRows)
	columns := gridTemplateTracks(node.Style.GridTemplateColumns, node.Style.GridAutoColumns)

	// Gaps: row-gap / column-gap each fall back to the gap shorthand only when
	// they were never set (zero-value unit). An explicit Px(0) is a real zero
	// gap that overrides GridGap, matching how the longhands override the
	// shorthand in CSS.
	// See: https://www.w3.org/TR/css-align-3/#gap-shorthand
	rowGap := gridResolveGap(node.Style.GridRowGap, node.Style.GridGap, ctx, currentFontSize)
	columnGap := gridResolveGap(node.Style.GridColumnGap, node.Style.GridGap, ctx, currentFontSize)

	if len(node.Children) == 0 {
		// Empty grid: the tracks alone (including gaps in both axes)
		// determine the size.
		columnSizes := gridSizeTracks(columns, colAxisSize, colAxisDefinite, columnGap, nil, nil, ctx, currentFontSize)
		rowSizes := gridSizeTracks(rows, rowAxisSize, rowAxisDefinite, rowGap, nil, nil, ctx, currentFontSize)
		return gridFinishContainer(node, constraints,
			gridTracksTotal(columnSizes, columnGap), gridTracksTotal(rowSizes, rowGap),
			isVerticalWritingMode,
			widthSpecified, contentWidth, heightSpecified, contentHeight,
			horizontalPaddingBorder, verticalPaddingBorder)
	}

	// Step 1: Resolve named grid areas to explicit positions
	// This must happen before auto-placement so area-based positions are treated as explicit
	gridResolveAreas(node)

	// Step 2: Place items using grid-auto-flow (§8.5). Placement may grow
	// the implicit grid; every item's end line is covered afterwards.
	// Absolutely positioned children are not grid items and are skipped by
	// placement (§9); they are laid out separately below.
	gridItems := gridPlaceItems(node, &rows, &columns, node.Style.GridAutoFlow)

	// Step 3: Size the columns (§12.5 intrinsic contributions, §12.6
	// maximize, §12.7 expand flexible tracks).
	//
	// Intrinsic column tracks (auto, min-content, max-content, fit-content,
	// minmax, and fr whose auto minimum is content-based) need two
	// contributions per item along the column axis: the minimum contribution
	// (min-content, e.g. the longest word of a text item) that sets a track's
	// base size, and the max-content contribution that sets its growth limit.
	// See: https://www.w3.org/TR/css-grid-1/#algo-content
	var colMinContrib, colMaxContrib []float64
	if gridTracksNeedContributions(columns, ctx, currentFontSize) {
		colMinContrib = make([]float64, len(columns))
		colMaxContrib = make([]float64, len(columns))
		for _, item := range gridItems {
			maxSize := gridMeasureItemAxis(item.node, isVerticalWritingMode, ctx)
			minSize := gridMeasureItemMinAxis(item.node, isVerticalWritingMode, maxSize, ctx)
			gridDistributeContribution(colMinContrib, columns, item.colStart, item.colEnd, columnGap, minSize, ctx, currentFontSize)
			gridDistributeContribution(colMaxContrib, columns, item.colStart, item.colEnd, columnGap, maxSize, ctx, currentFontSize)
		}
	}
	columnSizes := gridSizeTracks(columns, colAxisSize, colAxisDefinite, columnGap, colMinContrib, colMaxContrib, ctx, currentFontSize)
	// Note: JustifyContent's zero value is flex-start in this library (there
	// is no stretch keyword), so §12.8 stretch is not applied to columns.

	// Step 4: Measure children against their column-axis size to obtain the
	// row-axis contributions.
	rowContrib := make([]float64, len(rows))
	for _, item := range gridItems {
		itemColSize := gridSpanSize(columnSizes, item.colStart, item.colEnd, columnGap)

		// The column-axis size constrains the physical width in
		// horizontal-tb and the physical height in vertical modes.
		childConstraints := Constraints{MinWidth: 0, MaxWidth: itemColSize, MinHeight: 0, MaxHeight: Unbounded}
		if isVerticalWritingMode {
			childConstraints = Constraints{MinWidth: 0, MaxWidth: Unbounded, MinHeight: 0, MaxHeight: itemColSize}
		}

		childSize := gridLayoutItem(item.node, childConstraints, ctx)

		// Store measured (physical) size for use in positioning phase
		item.measuredSize = childSize

		// Row-axis contribution. childSize does NOT include margins - margins
		// are handled separately in positioning.
		rowSize := childSize.Height
		if isVerticalWritingMode {
			rowSize = childSize.Width
		}
		gridDistributeContribution(rowContrib, rows, item.rowStart, item.rowEnd, rowGap, rowSize, ctx, currentFontSize)
	}

	// Step 5: Size the rows. In the block axis an item's min-content and
	// max-content contributions coincide (both are its size at the resolved
	// column width), so the same array serves as base size and growth limit.
	rowSizes := gridSizeTracks(rows, rowAxisSize, rowAxisDefinite, rowGap, rowContrib, rowContrib, ctx, currentFontSize)

	// Step 6: Apply align-content to the rows (§10.4, §12.8). Free space
	// exists only when the row axis is definite.
	alignContent := node.Style.AlignContent
	rowAlignSpace := Unbounded
	if rowAxisDefinite {
		rowAlignSpace = rowAxisSize
	}
	rowSizes, totalRowSize := gridDistributeTrackSpace(rowSizes, rows, rowAlignSpace, rowGap, alignContent, ctx, currentFontSize)
	rowOffsets := gridCalculateTrackOffsets(rowSizes, totalRowSize, rowAlignSpace, rowGap, alignContent)

	// Extent of the row axis, used to mirror positions in vertical-rl.
	rowAxisExtent := totalRowSize
	if rowAxisDefinite && rowAxisSize > rowAxisExtent {
		rowAxisExtent = rowAxisSize
	}

	// Columns start at 0 (no justify-content distribution).
	columnOffsets := make([]float64, len(columnSizes))
	currentOffset := 0.0
	for i := range columnSizes {
		columnOffsets[i] = currentOffset
		currentOffset += columnSizes[i]
		if i < len(columnSizes)-1 {
			currentOffset += columnGap
		}
	}

	// Step 7: Position children
	for _, item := range gridItems {
		// Calculate grid cell position using track offsets
		cellX := 0.0
		if item.colStart < len(columnOffsets) {
			cellX = columnOffsets[item.colStart]
		}

		cellY := 0.0
		if item.rowStart < len(rowOffsets) {
			cellY = rowOffsets[item.rowStart]
		}

		// Calculate grid cell size (logical: width along columns, height along rows)
		cellWidth := gridSpanSize(columnSizes, item.colStart, item.colEnd, columnGap)
		cellHeight := gridSpanSize(rowSizes, item.rowStart, item.rowEnd, rowGap)

		// Position item within grid cell, accounting for margins
		// In CSS Grid, items stretch to fill their cell by default (align-items: stretch)
		// However, if an item has an aspect ratio, it should maintain that ratio while fitting within the cell
		// Get item's font size for margin resolution
		itemFontSize := getCurrentFontSize(item.node, ctx)
		marginLeft := ResolveLength(item.node.Style.Margin.Left, ctx, itemFontSize)
		marginRight := ResolveLength(item.node.Style.Margin.Right, ctx, itemFontSize)
		marginTop := ResolveLength(item.node.Style.Margin.Top, ctx, itemFontSize)
		marginBottom := ResolveLength(item.node.Style.Margin.Bottom, ctx, itemFontSize)

		maxItemWidth := cellWidth - marginLeft - marginRight
		maxItemHeight := cellHeight - marginTop - marginBottom

		// Clamp to >= 0 to prevent negative sizes
		if maxItemWidth < 0 {
			maxItemWidth = 0
		}
		if maxItemHeight < 0 {
			maxItemHeight = 0
		}

		// Measured sizes in logical terms (column axis, row axis).
		measuredCol, measuredRow := item.measuredSize.Width, item.measuredSize.Height
		if isVerticalWritingMode {
			measuredCol, measuredRow = item.measuredSize.Height, item.measuredSize.Width
		}

		// Explicit sizes in logical terms. In a vertical writing mode the
		// column axis is physical height, so the style's Height applies to
		// it and Width applies to the row axis.
		explicitColAxis := func(limit float64) (float64, bool) {
			if isVerticalWritingMode {
				return gridExplicitHeight(item.node, ctx, itemFontSize, limit)
			}
			return gridExplicitWidth(item.node, ctx, itemFontSize, limit)
		}
		explicitRowAxis := func(limit float64) (float64, bool) {
			if isVerticalWritingMode {
				return gridExplicitWidth(item.node, ctx, itemFontSize, limit)
			}
			return gridExplicitHeight(item.node, ctx, itemFontSize, limit)
		}

		var itemWidth, itemHeight float64

		// If item has aspect ratio, maintain it while fitting within cell
		// In CSS Grid, items with aspect ratio maintain their ratio but fit within the cell
		// For spanning items, we should use the measured size if it's valid and maintains aspect ratio
		if item.node.Style.AspectRatio > 0 {
			// Check if we have a valid measured size that maintains aspect ratio
			measuredRatio := 0.0
			if item.measuredSize.Width > 0 && item.measuredSize.Height > 0 {
				measuredRatio = item.measuredSize.Width / item.measuredSize.Height
			}

			// If measured size maintains aspect ratio, prefer it (especially for spanning items)
			// This ensures consistency between measurement and positioning phases
			// For spanning items, the measured size determines row/column sizes, so we should use it
			// According to CSS spec, items with aspect-ratio maintain their ratio and don't stretch
			// to fill cells (unlike items without aspect-ratio which stretch by default)
			if measuredRatio > 0 && math.Abs(measuredRatio-item.node.Style.AspectRatio) < 0.01 {
				// Use measured size, but ensure it fits within cell
				itemWidth = item.measuredSize.Width
				itemHeight = item.measuredSize.Height

				// Constrain to cell if measured size exceeds cell (shouldn't happen for spanning items)
				// But aspect ratio takes precedence - don't stretch beyond measured size
				if maxItemWidth > 0 && itemWidth > maxItemWidth {
					// Cell is smaller than measured - constrain to cell
					itemWidth = maxItemWidth
					itemHeight = itemWidth / item.node.Style.AspectRatio
				}
				if maxItemHeight > 0 && itemHeight > maxItemHeight {
					// Cell is smaller than measured - constrain to cell
					itemHeight = maxItemHeight
					itemWidth = itemHeight * item.node.Style.AspectRatio
				}
			} else if item.measuredSize.Width > 0 && item.measuredSize.Height > 0 {
				// Measured size exists but doesn't maintain aspect ratio - use it as fallback
				// This can happen if min/max constraints were applied
				itemWidth = item.measuredSize.Width
				itemHeight = item.measuredSize.Height
			} else {
				// Calculate dimensions that maintain aspect ratio and fit within cell
				// Try width-based first (fill cell width)
				if maxItemWidth > 0 {
					itemWidth = maxItemWidth
					itemHeight = itemWidth / item.node.Style.AspectRatio

					// If height exceeds cell, constrain by height instead
					if itemHeight > maxItemHeight && maxItemHeight > 0 {
						itemHeight = maxItemHeight
						itemWidth = itemHeight * item.node.Style.AspectRatio
					}

					// Ensure we don't exceed cell width (might happen if constrained by height)
					if itemWidth > maxItemWidth {
						itemWidth = maxItemWidth
						itemHeight = itemWidth / item.node.Style.AspectRatio
					}
				} else if maxItemHeight > 0 {
					// Cell width is 0, use height-based calculation
					itemHeight = maxItemHeight
					itemWidth = itemHeight * item.node.Style.AspectRatio
				} else {
					// Both are 0, use measured size if available
					if item.measuredSize.Width > 0 && item.measuredSize.Height > 0 {
						itemWidth = item.measuredSize.Width
						itemHeight = item.measuredSize.Height
					}
				}
			}
		} else {
			// No aspect ratio: apply justify-items and align-items alignment
			// Zero value is stretch (CSS Grid default)
			justifyItems := node.Style.JustifyItems
			alignItems := node.Style.AlignItems

			// Override with per-item alignment if set (CSS Grid §10.3)
			if item.node.Style.JustifySelf != 0 {
				justifyItems = item.node.Style.JustifySelf
			}
			if item.node.Style.AlignSelf != 0 {
				alignItems = item.node.Style.AlignSelf
			}

			// Only default to stretch if the value is outside the valid enum range
			if justifyItems > JustifyItemsCenter {
				justifyItems = JustifyItemsStretch
			}
			if alignItems > AlignItemsBaseline {
				alignItems = AlignItemsStretch
			}

			// Apply justify-items (inline/row axis)
			switch justifyItems {
			case JustifyItemsStart, JustifyItemsEnd, JustifyItemsCenter:
				// For non-stretch, always prefer explicit width if set (accounting for box-sizing)
				// Explicit dimensions take precedence over measured size for alignment
				if w, ok := explicitColAxis(maxItemWidth); ok {
					itemWidth = w
				} else if measuredCol > 0 {
					// Use measured size (clamped to cell)
					itemWidth = math.Min(measuredCol, maxItemWidth)
				} else {
					// No explicit width and no measured size - use 0 (min content)
					itemWidth = 0
				}
			case JustifyItemsStretch:
				// CSS Box Alignment Level 3 §6.2: stretch has no effect when the
				// relevant axis size is definite. An item with an explicit width
				// keeps that width and is positioned at the start of its area.
				// https://www.w3.org/TR/css-align-3/#stretch-alignment
				if w, ok := explicitColAxis(maxItemWidth); ok {
					itemWidth = w
				} else {
					// Auto width: stretch to fill cell width
					itemWidth = maxItemWidth
				}
			}

			// Apply align-items (block/column axis)
			switch alignItems {
			case AlignItemsFlexStart, AlignItemsFlexEnd, AlignItemsCenter, AlignItemsBaseline:
				// For non-stretch, always prefer explicit height if set (accounting for box-sizing)
				// Explicit dimensions take precedence over measured size for alignment
				if h, ok := explicitRowAxis(maxItemHeight); ok {
					itemHeight = h
				} else if measuredRow > 0 {
					// Use measured size (clamped to cell)
					itemHeight = math.Min(measuredRow, maxItemHeight)
				} else {
					// No explicit height and no measured size - use 0 (min content)
					itemHeight = 0
				}
			case AlignItemsStretch:
				// CSS Box Alignment Level 3 §6.2: stretch has no effect when the
				// relevant axis size is definite. An item with an explicit height
				// keeps that height and is positioned at the start of its area.
				// https://www.w3.org/TR/css-align-3/#stretch-alignment
				if h, ok := explicitRowAxis(maxItemHeight); ok {
					itemHeight = h
				} else {
					// Auto height: stretch to fill cell height
					itemHeight = maxItemHeight
				}
			default:
				// Default to stretch, but a definite height is preserved
				// per CSS Box Alignment Level 3 §6.2 (see above).
				// https://www.w3.org/TR/css-align-3/#stretch-alignment
				if h, ok := explicitRowAxis(maxItemHeight); ok {
					itemHeight = h
				} else {
					itemHeight = maxItemHeight
				}
			}
		}

		// Calculate item position within cell based on alignment
		var itemX, itemY float64

		// Handle justify-items positioning (inline/row axis)
		// Zero value is stretch (CSS Grid default)
		justifyItems := node.Style.JustifyItems

		// Override with per-item alignment if set (CSS Grid §10.3)
		if item.node.Style.JustifySelf != 0 {
			justifyItems = item.node.Style.JustifySelf
		}

		if justifyItems > JustifyItemsCenter {
			justifyItems = JustifyItemsStretch
		}
		// Items with aspect-ratio default to start alignment per spec
		if item.node.Style.AspectRatio > 0 {
			justifyItems = JustifyItemsStart
		}

		// Calculate total item size including margins for alignment
		totalItemWidth := itemWidth + marginLeft + marginRight
		totalItemHeight := itemHeight + marginTop + marginBottom

		switch justifyItems {
		case JustifyItemsStart:
			itemX = cellX + marginLeft
		case JustifyItemsEnd:
			// Align item+margin box to end, then item starts at margin.Left from that
			itemX = cellX + cellWidth - totalItemWidth + marginLeft
		case JustifyItemsCenter:
			// Center the item+margin box, then item starts at margin.Left from that
			itemX = cellX + (cellWidth-totalItemWidth)/2 + marginLeft
		case JustifyItemsStretch:
			itemX = cellX + marginLeft
		}

		// Handle align-items positioning (block/column axis)
		// Zero value is stretch (CSS default - same for Grid and Flexbox)
		alignItems := node.Style.AlignItems

		// Override with per-item alignment if set (CSS Grid §10.3)
		if item.node.Style.AlignSelf != 0 {
			alignItems = item.node.Style.AlignSelf
		}

		if alignItems > AlignItemsBaseline {
			alignItems = AlignItemsStretch
		}
		// Items with aspect-ratio default to start alignment per spec
		if item.node.Style.AspectRatio > 0 {
			alignItems = AlignItemsFlexStart
		}

		switch alignItems {
		case AlignItemsFlexStart: // Start
			itemY = cellY + marginTop
		case AlignItemsFlexEnd: // End
			// Align item+margin box to end, then item starts at margin.Top from that
			itemY = cellY + cellHeight - totalItemHeight + marginTop
		case AlignItemsCenter:
			// Center the item+margin box, then item starts at margin.Top from that
			itemY = cellY + (cellHeight-totalItemHeight)/2 + marginTop
		case AlignItemsBaseline:
			// For grid baseline alignment, align item's baseline to a reference
			// In CSS Grid, baseline alignment aligns items within their row
			// For simplicity, we align to the first baseline in the cell (top + baseline)
			// NOTE: For proper CSS Grid baseline alignment, we'd need to calculate the max baseline
			// across all items in the same row, similar to flexbox. For now, we use a simpler approach.
			itemY = cellY + marginTop
		case AlignItemsStretch:
			itemY = cellY + marginTop
		default:
			itemY = cellY + marginTop
		}

		// Position item within grid cell, accounting for margins, padding, and border
		// Margins are applied within the cell boundaries, not extending into gaps
		// For spanning items, margins are still contained within the spanned cell area
		// Add padding and border offsets to position items within the container's content area
		//
		// In vertical writing modes, swap X/Y positioning:
		// - Horizontal-TB: columns control X, rows control Y (default behavior)
		// - Vertical-LR/RL: columns control Y, rows control X (swap X and Y)
		//   - Vertical-LR: rows progress left-to-right
		//   - Vertical-RL: rows progress right-to-left
		var finalX, finalY, finalWidth, finalHeight float64
		if isVerticalWritingMode {
			// Vertical mode: rows control X (horizontal), columns control Y (vertical)
			if writingMode.IsRightToLeft() {
				// Vertical-RL: rows progress right-to-left, position from right edge
				finalX = paddingLeft + borderLeft + rowAxisExtent - itemY - itemHeight
			} else {
				// Vertical-LR: rows progress left-to-right
				finalX = paddingLeft + borderLeft + itemY // itemY becomes X
			}
			finalY = paddingTop + borderTop + itemX // itemX becomes Y
			finalWidth = itemHeight                 // height becomes width
			finalHeight = itemWidth                 // width becomes height
		} else {
			// Horizontal mode: columns control X, rows control Y (normal)
			finalX = paddingLeft + borderLeft + itemX
			finalY = paddingTop + borderTop + itemY
			finalWidth = itemWidth
			finalHeight = itemHeight
		}

		item.node.Rect = Rect{
			X:      finalX,
			Y:      finalY,
			Width:  finalWidth,
			Height: finalHeight,
		}

		// Ensure size doesn't go negative
		if item.node.Rect.Width < 0 {
			item.node.Rect.Width = 0
		}
		if item.node.Rect.Height < 0 {
			item.node.Rect.Height = 0
		}
	}

	// Step 8: Absolutely positioned children. They are not grid items (§9):
	// they took part in neither placement nor track sizing above. Each is
	// laid out for its own size against the container's content box and left
	// at its static position, the content-box origin (§9: the static position
	// is determined as if it were the sole grid item in a grid area whose
	// edges coincide with the content edges of the grid container). The
	// positioned pass (LayoutWithPositioning) then applies its offsets
	// relative to the container's padding box.
	// See: https://www.w3.org/TR/css-grid-1/#abspos-items
	// See: https://www.w3.org/TR/css-grid-1/#static-position
	for _, child := range node.Children {
		if child.Style.Display == DisplayNone || !isOutOfFlow(child) {
			continue
		}
		childSize := gridLayoutItem(child, Loose(contentWidth, contentHeight), ctx)
		child.Rect = Rect{
			X:      paddingLeft + borderLeft,
			Y:      paddingTop + borderTop,
			Width:  childSize.Width,
			Height: childSize.Height,
		}
	}

	return gridFinishContainer(node, constraints,
		gridTracksTotal(columnSizes, columnGap), totalRowSize,
		isVerticalWritingMode,
		widthSpecified, contentWidth, heightSpecified, contentHeight,
		horizontalPaddingBorder, verticalPaddingBorder)
}

type gridItem struct {
	node         *Node
	rowStart     int
	rowEnd       int
	colStart     int
	colEnd       int
	autoRow      bool // Row position came from auto-placement
	autoCol      bool // Column position came from auto-placement
	measuredSize Size // Store measured (physical) size from first pass
}

// gridResolveContainerAxis determines the grid container's content size in
// one physical axis.
//
// It returns the content size (Unbounded when indefinite), whether the size
// is definite, and whether it came from an explicit Width/Height.
//
// An unset size has the zero-value unit; only a size with an explicit unit
// is definite (the same convention gridExplicitWidth uses for items).
// Otherwise the axis is auto:
//   - width: a block-level grid container fills the available inline size,
//     so a bounded MaxWidth constraint is a definite size (CSS 2.1 §10.3.3);
//   - height: auto is content-sized and therefore indefinite (CSS 2.1
//     §10.6.3) unless the constraints are tight, in which case the parent
//     forces the size.
//
// See: https://www.w3.org/TR/css-grid-1/#algo-terms (free space, definite)
func gridResolveContainerAxis(node *Node, isWidth bool, constraints Constraints, horizontalPaddingBorder, verticalPaddingBorder float64, ctx *LayoutContext, currentFontSize float64) (float64, bool, bool) {
	specified := node.Style.Height
	minConstraint, maxConstraint := constraints.MinHeight, constraints.MaxHeight
	paddingBorder := verticalPaddingBorder
	if isWidth {
		specified = node.Style.Width
		minConstraint, maxConstraint = constraints.MinWidth, constraints.MaxWidth
		paddingBorder = horizontalPaddingBorder
	}

	if specified.Unit != "" {
		value := ResolveLength(specified, ctx, currentFontSize)
		if value >= 0 && value < Unbounded {
			content := convertToContentSize(value, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, isWidth)
			total := content + paddingBorder
			// An explicit size never exceeds a bounded constraint.
			if maxConstraint < Unbounded && total > maxConstraint {
				total = maxConstraint
			}
			return math.Max(0, total-paddingBorder), true, true
		}
	}

	if isWidth {
		if maxConstraint < Unbounded {
			return math.Max(0, maxConstraint-paddingBorder), true, false
		}
		return Unbounded, false, false
	}

	if maxConstraint < Unbounded && minConstraint >= maxConstraint {
		return math.Max(0, maxConstraint-paddingBorder), true, false
	}
	return Unbounded, false, false
}

// gridFinishContainer computes the container's final size, stores its Rect,
// and returns the constrained size.
//
// colTotal and rowTotal are the logical track extents (including gaps); they
// map to physical width/height according to the writing mode. An explicit
// Width/Height wins over the track total in its axis.
func gridFinishContainer(node *Node, constraints Constraints, colTotal, rowTotal float64, isVertical bool, widthSpecified bool, contentWidth float64, heightSpecified bool, contentHeight float64, horizontalPaddingBorder, verticalPaddingBorder float64) Size {
	width, height := colTotal, rowTotal
	if isVertical {
		width, height = rowTotal, colTotal
	}
	if widthSpecified {
		width = contentWidth
	}
	if heightSpecified {
		height = contentHeight
	}

	containerSize := Size{
		Width:  width + horizontalPaddingBorder,
		Height: height + verticalPaddingBorder,
	}

	// Constrain size and apply to Rect
	// CRITICAL: node.Rect must respect constraints to match the returned Size
	constrainedSize := constraints.Constrain(containerSize)
	node.Rect = Rect{
		X:      0,
		Y:      0,
		Width:  constrainedSize.Width,
		Height: constrainedSize.Height,
	}
	return constrainedSize
}

// gridResolveGap resolves one gap longhand (row-gap or column-gap), falling
// back to the gap shorthand when the longhand was never set. Only an unset
// longhand (zero-value unit) falls back; Px(0) is a real zero gap. Negative
// results are clamped to 0 since gaps cannot be negative.
//
// See: https://www.w3.org/TR/css-align-3/#gap-shorthand
func gridResolveGap(longhand, shorthand Length, ctx *LayoutContext, currentFontSize float64) float64 {
	gap := longhand
	if isUnsetLength(gap) {
		gap = shorthand
	}
	value := ResolveLength(gap, ctx, currentFontSize)
	if math.IsNaN(value) || value < 0 || value >= Unbounded {
		return 0
	}
	return value
}

// gridLayoutItem lays out a grid item with the given constraints using the
// layout algorithm matching its display type.
func gridLayoutItem(child *Node, childConstraints Constraints, ctx *LayoutContext) Size {
	switch child.Style.Display {
	case DisplayFlex:
		return LayoutFlexbox(child, childConstraints, ctx)
	case DisplayGrid:
		return LayoutGrid(child, childConstraints, ctx)
	case DisplayInlineText:
		return LayoutText(child, childConstraints, ctx)
	default:
		return LayoutBlock(child, childConstraints, ctx)
	}
}

// gridMeasureItemAxis returns an item's max-content contribution along the
// column axis (physical width in horizontal-tb, physical height in vertical
// writing modes) by laying it out without constraints.
//
// A block with an auto size fills its (unbounded) available space, which
// shows up as a value of Unbounded magnitude; such values carry no content
// information and contribute nothing.
//
// CSS Grid Layout Module Level 1 §12.5: max-content contribution
// See: https://www.w3.org/TR/css-grid-1/#algo-content
func gridMeasureItemAxis(child *Node, isVertical bool, ctx *LayoutContext) float64 {
	size := gridLayoutItem(child, Constraints{MinWidth: 0, MaxWidth: Unbounded, MinHeight: 0, MaxHeight: Unbounded}, ctx)
	value := size.Width
	if isVertical {
		value = size.Height
	}
	if math.IsNaN(value) || value < 0 || value >= Unbounded/2 {
		return 0
	}
	return value
}

// gridMeasureItemMinAxis returns an item's minimum contribution along the
// column axis: the size a track must have so the item does not overflow it.
//
// The min track sizing function of auto and fr tracks is auto, whose
// minimum contribution is the item's min-content size (§6.6 automatic
// minimum size; for text that is its longest unbreakable word, for a block
// or flex container the largest such contribution among its children). An
// item with a definite width has the same min- and max-content contribution,
// so its max-content size (maxContribution) is used directly.
//
// The result never exceeds maxContribution so a track's base size never
// exceeds its growth limit. Only the horizontal-tb writing mode has an
// inline-axis min-content measurement; in vertical writing modes the column
// axis is the physical block axis and the max-content size is used as the
// minimum as well.
//
// CSS Grid Layout Module Level 1 §12.5: "For auto minimums ... set the
// track's base size to the maximum of its items' minimum contributions."
// See: https://www.w3.org/TR/css-grid-1/#algo-single-span-items
// See: https://www.w3.org/TR/css-grid-1/#min-size-auto
// See: https://www.w3.org/TR/css-sizing-3/#min-content
func gridMeasureItemMinAxis(child *Node, isVertical bool, maxContribution float64, ctx *LayoutContext) float64 {
	if isVertical || maxContribution <= 0 {
		return maxContribution
	}
	fontSize := getCurrentFontSize(child, ctx)
	widthPx := ResolveLength(child.Style.Width, ctx, fontSize)
	if !isUnsetLength(child.Style.Width) && widthPx >= 0 && widthPx < Unbounded {
		// Definite width: the minimum contribution is the specified size,
		// which is what the max-content layout pass already produced.
		return maxContribution
	}
	value := CalculateIntrinsicWidth(child, Unconstrained(), IntrinsicSizeMinContent, ctx)
	if math.IsNaN(value) || value < 0 {
		return 0
	}
	if value > maxContribution {
		return maxContribution
	}
	return value
}

// gridIsZeroTrack reports whether a GridTrack is the zero value, which
// stands for the initial value of grid-auto-rows / grid-auto-columns (auto).
// A zero-value Length has no unit, unlike Px(0), so a genuine 0px track is
// not mistaken for the zero value.
//
// CSS Grid Layout Module Level 1 §7.6: grid-auto-rows / grid-auto-columns
// initial value: auto.
// See: https://www.w3.org/TR/css-grid-1/#auto-tracks
func gridIsZeroTrack(track GridTrack) bool {
	return track.Fraction == 0 &&
		track.MinSize.Unit == "" && track.MinSize.Value == 0 &&
		track.MaxSize.Unit == "" && track.MaxSize.Value == 0
}

// gridNormalizeTrack maps the zero-value track to AutoTrack().
func gridNormalizeTrack(track GridTrack) GridTrack {
	if gridIsZeroTrack(track) {
		return AutoTrack()
	}
	return track
}

// gridTemplateTracks returns a normalized copy of a track template, or a
// single implicit track when the template is empty.
func gridTemplateTracks(template []GridTrack, implicit GridTrack) []GridTrack {
	if len(template) == 0 {
		return []GridTrack{gridNormalizeTrack(implicit)}
	}
	tracks := make([]GridTrack, len(template))
	for i, track := range template {
		tracks[i] = gridNormalizeTrack(track)
	}
	return tracks
}

// gridTrackIsIntrinsic reports whether a track's size depends on its content
// and therefore receives item contributions: auto, min-content, max-content,
// fit-content, flexible tracks (whose auto minimum is content-based, §7.2.4),
// and minmax tracks with room to grow. Only a fixed track (min == max) is
// not intrinsic.
//
// See: https://www.w3.org/TR/css-grid-1/#algo-content
func gridTrackIsIntrinsic(track GridTrack, ctx *LayoutContext, currentFontSize float64) bool {
	if track.Fraction != 0 {
		return true
	}
	minSize := ResolveLength(track.MinSize, ctx, currentFontSize)
	maxSize := ResolveLength(track.MaxSize, ctx, currentFontSize)
	if maxSize == SizeMinContent || maxSize == SizeMaxContent || maxSize >= Unbounded {
		return true
	}
	return minSize != maxSize
}

// gridTracksNeedContributions reports whether any track in the list needs
// item contributions to be sized.
func gridTracksNeedContributions(tracks []GridTrack, ctx *LayoutContext, currentFontSize float64) bool {
	for _, track := range tracks {
		if gridTrackIsIntrinsic(track, ctx, currentFontSize) {
			return true
		}
	}
	return false
}

// gridTrackFixedSize returns the size of a non-intrinsic (fixed) track.
func gridTrackFixedSize(track GridTrack, ctx *LayoutContext, currentFontSize float64) float64 {
	size := ResolveLength(track.MinSize, ctx, currentFontSize)
	if size < 0 || size >= Unbounded {
		return 0
	}
	return size
}

// gridDistributeContribution records an item's size contribution in the
// tracks it spans.
//
// A non-spanning item contributes its whole size to its track. For a
// spanning item, the gaps and the fixed tracks it spans are subtracted first
// and only the remaining "extra space" is distributed, equally, to the
// intrinsic tracks it spans (CSS Grid Layout Module Level 1 §12.5 step 3).
// Fixed tracks never grow from spanning items. The equal split is applied as
// a per-track maximum so the result does not depend on item order.
//
// See: https://www.w3.org/TR/css-grid-1/#algo-content
// See: https://www.w3.org/TR/css-grid-1/#extra-space
func gridDistributeContribution(contrib []float64, tracks []GridTrack, start, end int, gap, size float64, ctx *LayoutContext, currentFontSize float64) {
	if start < 0 {
		start = 0
	}
	if end > len(tracks) {
		end = len(tracks)
	}
	if end > len(contrib) {
		end = len(contrib)
	}
	if end <= start || size <= 0 {
		return
	}

	if end-start == 1 {
		if size > contrib[start] {
			contrib[start] = size
		}
		return
	}

	remaining := size - gap*float64(end-start-1)
	intrinsic := make([]int, 0, end-start)
	for i := start; i < end; i++ {
		if gridTrackIsIntrinsic(tracks[i], ctx, currentFontSize) {
			intrinsic = append(intrinsic, i)
		} else {
			remaining -= gridTrackFixedSize(tracks[i], ctx, currentFontSize)
		}
	}
	if len(intrinsic) == 0 || remaining <= 0 {
		return
	}

	perTrack := remaining / float64(len(intrinsic))
	for _, i := range intrinsic {
		if perTrack > contrib[i] {
			contrib[i] = perTrack
		}
	}
}

// gridSizeTracks runs the track sizing algorithm for one axis.
//
// Algorithm based on CSS Grid Layout Module Level 1 §12.3 - §12.7:
//
//   - §12.4 Initialize Track Sizes / §12.5 Resolve Intrinsic Track Sizes:
//     each track's base size comes from its min sizing function and, for an
//     intrinsic minimum, from the items' minimum contributions in minContrib
//     (auto and min-content minimums use the min-content contribution, a
//     max-content minimum uses maxContrib). Each track's growth limit comes
//     from its max sizing function and, for an intrinsic maximum, from the
//     items' max-content contributions in maxContrib (min-content maximums
//     use minContrib; fit-content clamps to its limit). A flexible track's
//     growth limit is its base size (§12.5 final step), so §12.6 never grows
//     it. Both arrays may be nil when no items have been measured.
//   - §12.6 Maximize Tracks: with definite free space, tracks whose growth
//     limit exceeds their base size grow toward it, sharing the free space
//     equally and freezing as they reach their limits; with indefinite
//     space (a max-content constraint) the free space is infinite and every
//     track grows to its growth limit.
//   - §12.7 Expand Flexible Tracks: fr tracks are minmax(auto, Nfr); with
//     definite space the flex fraction is found per §12.7.1, restarting with
//     a track treated as inflexible whenever its content-based base size
//     exceeds its share (bounded by the number of flexible tracks); with
//     indefinite space the flex fraction is the largest of base size / flex
//     factor and max-content contribution / flex factor among the flexible
//     tracks, and no track shrinks below its base size.
//
// §12.8 (Stretch auto Tracks) is applied by gridDistributeTrackSpace.
//
// available is the container's content size in this axis (Unbounded when
// indefinite) and definite says whether free space exists at all.
//
// See: https://www.w3.org/TR/css-grid-1/#algo-track-sizing
// See: https://www.w3.org/TR/css-grid-1/#algo-content
// See: https://www.w3.org/TR/css-grid-1/#algo-grow-tracks
// See: https://www.w3.org/TR/css-grid-1/#algo-flex-tracks
func gridSizeTracks(tracks []GridTrack, available float64, definite bool, gap float64, minContrib, maxContrib []float64, ctx *LayoutContext, currentFontSize float64) []float64 {
	n := len(tracks)
	if n == 0 {
		return []float64{}
	}

	sizes := make([]float64, n)
	limits := make([]float64, n)
	flexIdx := make([]int, 0)

	// contribution returns a finite, non-negative entry of a contribution
	// array, or 0 when the array is shorter than the track list.
	contribution := func(contrib []float64, i int) float64 {
		if i >= len(contrib) {
			return 0
		}
		v := contrib[i]
		if math.IsNaN(v) || v < 0 || v >= Unbounded {
			return 0
		}
		return v
	}

	for i, track := range tracks {
		minC := contribution(minContrib, i)
		maxC := contribution(maxContrib, i)
		if maxC < minC {
			maxC = minC
		}
		rawMin := ResolveLength(track.MinSize, ctx, currentFontSize)
		maxSize := ResolveLength(track.MaxSize, ctx, currentFontSize)

		// Base size from the min sizing function (§12.4, §12.5 step 1).
		var base float64
		switch {
		case rawMin == SizeMaxContent:
			// max-content minimum: the items' max-content contributions.
			base = maxC
		case rawMin < 0 || rawMin >= Unbounded:
			// auto or min-content minimum (an unbounded value also behaves as
			// auto): the items' minimum contributions.
			base = minC
		default:
			// Fixed minimum. Content never shrinks a track below it; an
			// auto-like zero minimum still takes the minimum contribution.
			base = math.Max(rawMin, minC)
		}
		minSize := rawMin
		if minSize < 0 || minSize >= Unbounded {
			minSize = 0
		}

		switch {
		case track.Fraction > 0:
			// §7.2.4: <flex> as a max sizing function implies an auto
			// minimum. Growth limit = base size (§12.5 final step) so the
			// track only grows in §12.7.
			flexIdx = append(flexIdx, i)
			sizes[i] = base
			limits[i] = base
		case track.Fraction == -1:
			// fit-content(limit): auto minimum, growth limit = max-content
			// clamped to the limit (§7.2.2). The auto minimum is limited by
			// the same fixed limit (§6.6 "limited min-content contribution").
			limit := maxC
			if maxSize >= 0 && maxSize < Unbounded {
				limit = math.Min(limit, maxSize)
				base = math.Min(base, math.Max(minSize, maxSize))
			}
			sizes[i] = base
			limits[i] = math.Max(base, limit)
		case maxSize == SizeMinContent:
			// min-content maximum: growth limit = min-content contributions.
			sizes[i] = base
			limits[i] = math.Max(base, minC)
		case maxSize == SizeMaxContent || maxSize >= Unbounded:
			// max-content or auto maximum: growth limit = max-content
			// contributions (§12.5: an auto maximum is treated as
			// max-content here).
			sizes[i] = base
			limits[i] = math.Max(base, maxC)
		case maxSize < minSize:
			// §7.2.1: if max < min, the max is ignored and the track is min.
			sizes[i] = minSize
			limits[i] = minSize
		case minSize == maxSize:
			// Fixed track.
			sizes[i] = minSize
			limits[i] = minSize
		default:
			// minmax(min, fixed): an intrinsic minimum is limited by the
			// fixed maximum (§6.6); the fixed maximum is the growth limit.
			sizes[i] = math.Min(base, maxSize)
			limits[i] = maxSize
		}
	}

	definiteSpace := definite && available < Unbounded
	totalGap := gap * float64(n-1)

	// §12.6 Maximize Tracks.
	if definiteSpace {
		free := available - totalGap - sumSizes(sizes)
		for iter := 0; iter < n && free > 0; iter++ {
			growable := make([]int, 0, n)
			for i := range sizes {
				if limits[i] > sizes[i] {
					growable = append(growable, i)
				}
			}
			if len(growable) == 0 {
				break
			}
			perTrack := free / float64(len(growable))
			for _, i := range growable {
				room := limits[i] - sizes[i]
				add := math.Min(perTrack, room)
				sizes[i] += add
				free -= add
			}
		}
	} else {
		// Indefinite free space (sizing under a max-content constraint):
		// every track grows to its growth limit.
		for i := range sizes {
			if limits[i] > sizes[i] {
				sizes[i] = limits[i]
			}
		}
	}

	// §12.7 Expand Flexible Tracks.
	if len(flexIdx) > 0 {
		if definiteSpace {
			leftover := available - totalGap
			for i := range sizes {
				if tracks[i].Fraction <= 0 {
					leftover -= sizes[i]
				}
			}
			frozen := make([]bool, n)
			// Each iteration either freezes at least one track or finishes,
			// so len(flexIdx)+1 iterations always suffice.
			for iter := 0; iter <= len(flexIdx); iter++ {
				totalFr, frozenSpace := 0.0, 0.0
				for _, i := range flexIdx {
					if frozen[i] {
						frozenSpace += sizes[i]
					} else {
						totalFr += tracks[i].Fraction
					}
				}
				if totalFr <= 0 {
					break
				}
				if totalFr < 1 {
					// §12.7.1: a flex factor sum below 1 is treated as 1.
					totalFr = 1
				}
				space := leftover - frozenSpace
				if space < 0 {
					space = 0
				}
				hypothetical := space / totalFr
				changed := false
				for _, i := range flexIdx {
					if !frozen[i] && sizes[i] > tracks[i].Fraction*hypothetical {
						// Content-based base size wins: treat as inflexible.
						frozen[i] = true
						changed = true
					}
				}
				if !changed {
					for _, i := range flexIdx {
						if !frozen[i] {
							sizes[i] = tracks[i].Fraction * hypothetical
						}
					}
					break
				}
			}
		} else {
			// Indefinite free space (§12.7.1): the flex fraction is the
			// largest of each flexible track's base size divided by its flex
			// factor and each item's max-content contribution divided by the
			// flex factor of the track it sits in. A track never ends up
			// below its base size.
			flexFraction := 0.0
			for _, i := range flexIdx {
				flexFraction = math.Max(flexFraction, sizes[i]/tracks[i].Fraction)
				flexFraction = math.Max(flexFraction, contribution(maxContrib, i)/tracks[i].Fraction)
			}
			for _, i := range flexIdx {
				sizes[i] = math.Max(sizes[i], tracks[i].Fraction*flexFraction)
			}
		}
	}

	return sizes
}

// gridSpanSize returns the size of the cells from start (inclusive) to end
// (exclusive), including the gaps between them.
func gridSpanSize(sizes []float64, start, end int, gap float64) float64 {
	if start < 0 {
		start = 0
	}
	if end > len(sizes) {
		end = len(sizes)
	}
	if end <= start {
		return 0
	}
	total := 0.0
	for i := start; i < end; i++ {
		total += sizes[i]
	}
	return total + gap*float64(end-start-1)
}

// gridTracksTotal returns the total extent of a track list including gaps.
func gridTracksTotal(sizes []float64, gap float64) float64 {
	if len(sizes) == 0 {
		return 0
	}
	return sumSizes(sizes) + gap*float64(len(sizes)-1)
}

// gridExplicitWidth returns the box-sizing-aware used width for a grid item that
// has a definite (non-auto) width, clamped to the available area maxItemWidth.
// The boolean is false when the item's width is auto (ResolveLength < 0), in
// which case callers should fall back to measured/stretch sizing.
//
// This is shared by both the stretch and non-stretch alignment branches so that
// the explicit-size computation stays in one place.
func gridExplicitWidth(n *Node, ctx *LayoutContext, fontSize, maxItemWidth float64) (float64, bool) {
	// An auto/unset width has the zero-value unit; only a width with an
	// explicit unit is a definite size. ResolveLength returns 0 (not < 0) for
	// the zero value, so the unit check is required to tell auto from a
	// genuine 0-length.
	if n.Style.Width.Unit == "" {
		// Auto width.
		return 0, false
	}
	widthValue := ResolveLength(n.Style.Width, ctx, fontSize)
	if widthValue < 0 {
		// Auto width.
		return 0, false
	}
	if n.Style.BoxSizing == BoxSizingBorderBox {
		// Width already includes padding+border, use as-is.
		return math.Min(widthValue, maxItemWidth), true
	}
	// Width is content-only, add padding+border.
	paddingBorder := ResolveLength(n.Style.Padding.Left, ctx, fontSize) +
		ResolveLength(n.Style.Padding.Right, ctx, fontSize) +
		ResolveLength(n.Style.Border.Left, ctx, fontSize) +
		ResolveLength(n.Style.Border.Right, ctx, fontSize)
	return math.Min(widthValue+paddingBorder, maxItemWidth), true
}

// gridExplicitHeight returns the box-sizing-aware used height for a grid item
// that has a definite (non-auto) height, clamped to the available area
// maxItemHeight. The boolean is false when the item's height is auto
// (ResolveLength < 0), in which case callers should fall back to
// measured/stretch sizing.
func gridExplicitHeight(n *Node, ctx *LayoutContext, fontSize, maxItemHeight float64) (float64, bool) {
	// An auto/unset height has the zero-value unit; only a height with an
	// explicit unit is a definite size (ResolveLength returns 0, not < 0, for
	// the zero value, so the unit check is required to tell auto apart).
	if n.Style.Height.Unit == "" {
		// Auto height.
		return 0, false
	}
	heightValue := ResolveLength(n.Style.Height, ctx, fontSize)
	if heightValue < 0 {
		// Auto height.
		return 0, false
	}
	if n.Style.BoxSizing == BoxSizingBorderBox {
		// Height already includes padding+border, use as-is.
		return math.Min(heightValue, maxItemHeight), true
	}
	// Height is content-only, add padding+border.
	paddingBorder := ResolveLength(n.Style.Padding.Top, ctx, fontSize) +
		ResolveLength(n.Style.Padding.Bottom, ctx, fontSize) +
		ResolveLength(n.Style.Border.Top, ctx, fontSize) +
		ResolveLength(n.Style.Border.Bottom, ctx, fontSize)
	return math.Min(heightValue+paddingBorder, maxItemHeight), true
}

func sumSizes(sizes []float64) float64 {
	sum := 0.0
	for _, s := range sizes {
		sum += s
	}
	return sum
}
