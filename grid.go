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
// See: https://www.w3.org/TR/css-grid-1/
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

	// Calculate available space
	// If container has explicit width/height, use that to constrain available space
	// Otherwise, use constraints (similar to block layout)
	availableWidth := constraints.MaxWidth
	availableHeight := constraints.MaxHeight

	// Account for padding and border - resolve Length values
	paddingLeft := ResolveLength(node.Style.Padding.Left, ctx, currentFontSize)
	paddingRight := ResolveLength(node.Style.Padding.Right, ctx, currentFontSize)
	paddingTop := ResolveLength(node.Style.Padding.Top, ctx, currentFontSize)
	paddingBottom := ResolveLength(node.Style.Padding.Bottom, ctx, currentFontSize)
	borderLeft := ResolveLength(node.Style.Border.Left, ctx, currentFontSize)
	borderRight := ResolveLength(node.Style.Border.Right, ctx, currentFontSize)
	borderTop := ResolveLength(node.Style.Border.Top, ctx, currentFontSize)
	borderBottom := ResolveLength(node.Style.Border.Bottom, ctx, currentFontSize)

	horizontalPadding := paddingLeft + paddingRight
	verticalPadding := paddingTop + paddingBottom
	horizontalBorder := borderLeft + borderRight
	verticalBorder := borderTop + borderBottom
	horizontalPaddingBorder := horizontalPadding + horizontalBorder
	verticalPaddingBorder := verticalPadding + verticalBorder

	// If container has explicit width, use it to constrain available width
	// In CSS, an explicit width on a grid container determines the container's size
	// Convert from box-sizing to total size for comparison with constraints
	widthValue := ResolveLength(node.Style.Width, ctx, currentFontSize)
	if widthValue >= 0 {
		// Convert to content size first
		specifiedWidthContent := convertToContentSize(widthValue, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, true)
		// Add padding+border to get total size for comparison
		totalSpecifiedWidth := specifiedWidthContent + horizontalPaddingBorder
		// Use explicit width when set, respecting constraints
		// If constraint is unbounded, always use explicit width
		// Otherwise, use the smaller of explicit width or constraint
		if availableWidth >= Unbounded {
			availableWidth = totalSpecifiedWidth
		} else {
			// Use the smaller value, but if they're equal, prefer explicit width
			if totalSpecifiedWidth <= availableWidth {
				availableWidth = totalSpecifiedWidth
			}
		}
	}

	// If container has explicit height, use it to constrain available height
	heightValue := ResolveLength(node.Style.Height, ctx, currentFontSize)
	if heightValue >= 0 {
		// Convert to content size first
		specifiedHeightContent := convertToContentSize(heightValue, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, false)
		// Add padding+border to get total size for comparison
		totalSpecifiedHeight := specifiedHeightContent + verticalPaddingBorder
		// Use explicit height when set, but don't exceed constraints
		if availableHeight >= Unbounded {
			availableHeight = totalSpecifiedHeight
		} else if totalSpecifiedHeight < availableHeight {
			availableHeight = totalSpecifiedHeight
		}
	}

	// Clamp content size to >= 0
	contentWidth := availableWidth - horizontalPaddingBorder
	if contentWidth < 0 {
		contentWidth = 0
	}
	contentHeight := availableHeight - verticalPaddingBorder
	if contentHeight < 0 {
		contentHeight = 0
	}

	// Get grid template
	rows := node.Style.GridTemplateRows
	columns := node.Style.GridTemplateColumns

	// Use auto tracks if templates not specified
	if len(rows) == 0 {
		rows = []GridTrack{node.Style.GridAutoRows}
		if len(rows) == 0 || (rows[0].MinSize.Value == 0 && rows[0].MaxSize.Value == Unbounded && rows[0].Fraction == 0) {
			rows = []GridTrack{AutoTrack()}
		}
	}
	if len(columns) == 0 {
		columns = []GridTrack{node.Style.GridAutoColumns}
		if len(columns) == 0 || (columns[0].MinSize.Value == 0 && columns[0].MaxSize.Value == Unbounded && columns[0].Fraction == 0) {
			columns = []GridTrack{AutoTrack()}
		}
	}

	// Calculate gap - resolve Length values
	gridGap := ResolveLength(node.Style.GridGap, ctx, currentFontSize)
	rowGap := ResolveLength(node.Style.GridRowGap, ctx, currentFontSize)
	if rowGap == 0 {
		rowGap = gridGap
	}
	columnGap := ResolveLength(node.Style.GridColumnGap, ctx, currentFontSize)
	if columnGap == 0 {
		columnGap = gridGap
	}

	// Step 1: Calculate column sizes
	// CRITICAL: contentWidth must be correct here - it's used to size all columns
	// For row-spanning items with aspect ratio, contentWidth must be correct for proper sizing
	columnSizes := calculateGridTrackSizes(columns, contentWidth, columnGap, len(columns), node, true, ctx, currentFontSize)

	// Step 2: Calculate row sizes (need to measure children first for auto rows)
	// For now, we'll do a two-pass layout
	children := node.Children
	if len(children) == 0 {
		// Empty grid
		totalWidth := sumSizes(columnSizes) + columnGap*float64(len(columnSizes)-1)
		totalHeight := sumSizes(calculateGridTrackSizes(rows, contentHeight, rowGap, len(rows), node, false, ctx, currentFontSize))
		resultSize := Size{
			Width:  totalWidth + horizontalPadding + horizontalBorder,
			Height: totalHeight + verticalPadding + verticalBorder,
		}
		node.Rect = Rect{
			X:      0,
			Y:      0,
			Width:  resultSize.Width,
			Height: resultSize.Height,
		}
		return constraints.Constrain(resultSize)
	}

	// Step 2: Resolve named grid areas to explicit positions
	// This must happen before auto-placement so area-based positions are treated as explicit
	gridResolveAreas(node)

	// Step 3: Place items using grid-auto-flow
	// Use gridPlaceItems from grid_placement.go which handles row/column flow and dense packing
	autoFlow := node.Style.GridAutoFlow
	gridItems := gridPlaceItems(node, &rows, &columns, autoFlow)

	// Recalculate column sizes if columns were extended during placement
	columnSizes = calculateGridTrackSizes(columns, contentWidth, columnGap, len(columns), node, true, ctx, currentFontSize)

	// Step 4: Measure children to determine row sizes
	// Ensure rowSizes and rowHeights are properly sized for all rows
	rowSizes := make([]float64, len(rows))
	rowHeights := make([]float64, len(rows))

	for _, item := range gridItems {
		// Calculate available width for this item
		itemWidth := 0.0
		for col := item.colStart; col < item.colEnd; col++ {
			itemWidth += columnSizes[col]
		}
		if item.colEnd > item.colStart+1 {
			itemWidth += columnGap * float64(item.colEnd-item.colStart-1)
		}

		// Measure child
		childConstraints := Constraints{
			MinWidth:  0,
			MaxWidth:  itemWidth,
			MinHeight: 0,
			MaxHeight: Unbounded,
		}

		var childSize Size
		if item.node.Style.Display == DisplayFlex {
			childSize = LayoutFlexbox(item.node, childConstraints, ctx)
		} else if item.node.Style.Display == DisplayGrid {
			childSize = LayoutGrid(item.node, childConstraints, ctx)
		} else if item.node.Style.Display == DisplayInlineText {
			childSize = LayoutText(item.node, childConstraints, ctx)
		} else {
			childSize = LayoutBlock(item.node, childConstraints, ctx)
		}

		// Store measured size for use in positioning phase
		item.measuredSize = childSize

		// Track required height for each row
		// childSize.Height already respects MinHeight (set in block layout)
		// Note: childSize.Height does NOT include margins - margins are handled separately in positioning
		itemHeight := childSize.Height
		spanRows := item.rowEnd - item.rowStart

		// For spanning items, the item height needs to be distributed across rows
		// The item height is the content height, and the cell height (which includes gaps)
		// is: row0 + gap + row1 + gap + ... + rowN
		// For auto-sized rows, we need to determine row heights such that the sum equals the item height
		// If we assume equal row heights: spanRows * rowHeight + (spanRows-1) * gap = itemHeight
		// So: rowHeight = (itemHeight - (spanRows-1) * gap) / spanRows
		var heightPerRow float64
		if spanRows > 1 {
			// Account for gaps between rows
			totalGaps := rowGap * float64(spanRows-1)
			heightPerRow = (itemHeight - totalGaps) / float64(spanRows)
			// Clamp to >= 0 to prevent negative row heights
			if heightPerRow < 0 {
				heightPerRow = 0
			}
		} else {
			// Single row: item height is the row height
			heightPerRow = itemHeight
		}

		for row := item.rowStart; row < item.rowEnd; row++ {
			if heightPerRow > rowHeights[row] {
				rowHeights[row] = heightPerRow
			}
		}
	}

	// Step 4: Calculate final row sizes
	availableHeightForRows := contentHeight - rowGap*float64(len(rows)-1)
	totalFixedHeight := 0.0
	totalFraction := 0.0

	for i, track := range rows {
		if track.Fraction > 0 {
			totalFraction += track.Fraction
		} else {
			// Resolve track sizes
			minSize := ResolveLength(track.MinSize, ctx, currentFontSize)
			maxSize := ResolveLength(track.MaxSize, ctx, currentFontSize)

			// For fixed tracks (MinSize == MaxSize and both > 0), use the track size directly
			// For auto tracks, use measured height
			var trackHeight float64
			if minSize > 0 && minSize == maxSize {
				// Fixed track - use the fixed size
				trackHeight = minSize
			} else {
				// Auto or minmax track - use measured height or track size
				// The measured height comes from children, which respects MinHeight if set
				trackHeight = math.Max(minSize, rowHeights[i])
				if maxSize < Unbounded {
					trackHeight = math.Min(trackHeight, maxSize)
				}
			}
			rowSizes[i] = trackHeight
			totalFixedHeight += trackHeight
		}
	}

	// Distribute fractional space
	if totalFraction > 0 {
		remainingHeight := availableHeightForRows - totalFixedHeight
		if remainingHeight > 0 {
			for i, track := range rows {
				if track.Fraction > 0 {
					rowSizes[i] = (remainingHeight * track.Fraction) / totalFraction
				}
			}
		}
	} else {
		// All fixed or auto - ensure any unset rows use measured heights or track sizes
		for i := range rows {
			if rowSizes[i] == 0 {
				// Only set if not already set (for auto tracks)
				track := rows[i]
				minSize := ResolveLength(track.MinSize, ctx, currentFontSize)
				maxSize := ResolveLength(track.MaxSize, ctx, currentFontSize)

				if minSize == maxSize && maxSize < Unbounded {
					// Fixed track that wasn't set - use fixed size
					rowSizes[i] = minSize
				} else {
					// Auto track - use measured height or min size
					// The measured height comes from children, which respects MinHeight if set
					rowSizes[i] = math.Max(minSize, rowHeights[i])
					if maxSize < Unbounded {
						rowSizes[i] = math.Min(rowSizes[i], maxSize)
					}
				}
			}
		}
	}

	// Step 4.5: Apply track distribution (justify-content for columns, align-content for rows)
	// This handles free space distribution and track positioning
	//
	// Algorithm based on CSS Grid Layout Module Level 1:
	// - §11.8: Distributing free space
	// - §10.4: Aligning the Grid (align-content, justify-content)
	//
	// See: https://www.w3.org/TR/css-grid-1/#grid-align

	// Apply align-content distribution for rows
	alignContent := node.Style.AlignContent
	distributedRowSizes, totalDistributedRowSize := gridDistributeTrackSpace(rowSizes, contentHeight, rowGap, alignContent)
	rowSizes = distributedRowSizes

	// Calculate track offsets based on alignment
	columnOffsets := make([]float64, len(columnSizes))
	// For now, columns start at 0 (no justify-content distribution yet)
	currentOffset := 0.0
	for i := range columnSizes {
		columnOffsets[i] = currentOffset
		currentOffset += columnSizes[i]
		if i < len(columnSizes)-1 {
			currentOffset += columnGap
		}
	}

	rowOffsets := gridCalculateTrackOffsets(rowSizes, totalDistributedRowSize, contentHeight, rowGap, alignContent)

	// Step 5: Position children
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

		// Calculate grid cell size
		cellWidth := 0.0
		for col := item.colStart; col < item.colEnd; col++ {
			cellWidth += columnSizes[col]
		}
		if item.colEnd > item.colStart+1 {
			cellWidth += columnGap * float64(item.colEnd-item.colStart-1)
		}

		cellHeight := 0.0
		for row := item.rowStart; row < item.rowEnd && row < len(rowSizes); row++ {
			cellHeight += rowSizes[row]
		}
		if item.rowEnd > item.rowStart+1 {
			cellHeight += rowGap * float64(item.rowEnd-item.rowStart-1)
		}

		// Position item within grid cell, accounting for margins
		// In CSS Grid, items stretch to fill their cell by default (align-items: stretch)
		// However, if an item has an aspect ratio, it should maintain that ratio while fitting within the cell
		// Get item's font size for margin resolution
		itemFontSize := 16.0 // Default
		if item.node.Style.TextStyle != nil && item.node.Style.TextStyle.FontSize > 0 {
			itemFontSize = item.node.Style.TextStyle.FontSize
		}
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
				itemWidthValue := ResolveLength(item.node.Style.Width, ctx, itemFontSize)
				if itemWidthValue >= 0 {
					// Item has explicit width - convert to total size accounting for box-sizing
					itemPaddingLeft := ResolveLength(item.node.Style.Padding.Left, ctx, itemFontSize)
					itemPaddingRight := ResolveLength(item.node.Style.Padding.Right, ctx, itemFontSize)
					itemBorderLeft := ResolveLength(item.node.Style.Border.Left, ctx, itemFontSize)
					itemBorderRight := ResolveLength(item.node.Style.Border.Right, ctx, itemFontSize)
					itemPaddingBorder := itemPaddingLeft + itemPaddingRight + itemBorderLeft + itemBorderRight

					if item.node.Style.BoxSizing == BoxSizingBorderBox {
						// Width already includes padding+border, use as-is
						itemWidth = math.Min(itemWidthValue, maxItemWidth)
					} else {
						// Width is content-only, add padding+border
						itemWidth = math.Min(itemWidthValue+itemPaddingBorder, maxItemWidth)
					}
				} else if item.measuredSize.Width > 0 {
					// Use measured size (clamped to cell)
					itemWidth = math.Min(item.measuredSize.Width, maxItemWidth)
				} else {
					// No explicit width and no measured size - use 0 (min content)
					itemWidth = 0
				}
			case JustifyItemsStretch:
				// Stretch to fill cell width
				itemWidth = maxItemWidth
			}

			// Apply align-items (block/column axis)
			switch alignItems {
			case AlignItemsFlexStart, AlignItemsFlexEnd, AlignItemsCenter, AlignItemsBaseline:
				// For non-stretch, always prefer explicit height if set (accounting for box-sizing)
				// Explicit dimensions take precedence over measured size for alignment
				itemHeightValue := ResolveLength(item.node.Style.Height, ctx, itemFontSize)
				if itemHeightValue >= 0 {
					// Item has explicit height - convert to total size accounting for box-sizing
					itemPaddingTop := ResolveLength(item.node.Style.Padding.Top, ctx, itemFontSize)
					itemPaddingBottom := ResolveLength(item.node.Style.Padding.Bottom, ctx, itemFontSize)
					itemBorderTop := ResolveLength(item.node.Style.Border.Top, ctx, itemFontSize)
					itemBorderBottom := ResolveLength(item.node.Style.Border.Bottom, ctx, itemFontSize)
					itemPaddingBorder := itemPaddingTop + itemPaddingBottom + itemBorderTop + itemBorderBottom

					if item.node.Style.BoxSizing == BoxSizingBorderBox {
						// Height already includes padding+border, use as-is
						itemHeight = math.Min(itemHeightValue, maxItemHeight)
					} else {
						// Height is content-only, add padding+border
						itemHeight = math.Min(itemHeightValue+itemPaddingBorder, maxItemHeight)
					}
				} else if item.measuredSize.Height > 0 {
					// Use measured size (clamped to cell)
					itemHeight = math.Min(item.measuredSize.Height, maxItemHeight)
				} else {
					// No explicit height and no measured size - use 0 (min content)
					itemHeight = 0
				}
			case AlignItemsStretch:
				// Stretch to fill cell height
				itemHeight = maxItemHeight
			default:
				// Default to stretch
				itemHeight = maxItemHeight
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
			itemBaseline := item.node.Baseline
			if itemBaseline == 0 {
				// Default: baseline is at the bottom of the item (for boxes without text)
				itemBaseline = itemHeight
			}
			// Position item so its baseline is at a consistent position within the cell
			// For grid, we use the item's own baseline as the reference (align to cell start + baseline)
			// This means items with different baselines will align their baselines together
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
		item.node.Rect = Rect{
			X:      paddingLeft + borderLeft + itemX,
			Y:      paddingTop + borderTop + itemY,
			Width:  itemWidth,
			Height: itemHeight,
		}

		// Note: The margin is already accounted for in maxItemHeight calculation above,
		// so itemHeight is the content height, and the margin positions the item within the cell.
		// The cell boundaries (cellY, cellY + cellHeight) define the grid structure,
		// and margins are purely internal to the cell.

		// Ensure size doesn't go negative
		if item.node.Rect.Width < 0 {
			item.node.Rect.Width = 0
		}
		if item.node.Rect.Height < 0 {
			item.node.Rect.Height = 0
		}
	}

	// Calculate container size
	totalWidth := sumSizes(columnSizes) + columnGap*float64(len(columnSizes)-1)
	totalHeight := sumSizes(rowSizes) + rowGap*float64(len(rowSizes)-1)

	containerSize := Size{
		Width:  totalWidth + horizontalPadding + horizontalBorder,
		Height: totalHeight + verticalPadding + verticalBorder,
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

type gridItem struct {
	node         *Node
	rowStart     int
	rowEnd       int
	colStart     int
	colEnd       int
	measuredSize Size // Store measured size from first pass
}

func calculateGridTrackSizes(tracks []GridTrack, availableSize float64, gap float64, count int, container *Node, isColumn bool, ctx *LayoutContext, currentFontSize float64) []float64 {
	if len(tracks) == 0 {
		return []float64{}
	}

	sizes := make([]float64, len(tracks))
	totalGap := gap * float64(len(tracks)-1)
	// Clamp available space to >= 0
	availableForTracks := availableSize - totalGap
	if availableForTracks < 0 {
		availableForTracks = 0
	}

	// CRITICAL FIX: Handle Unbounded constraints for fractional tracks
	// When availableSize is Unbounded, fractional tracks can't be distributed proportionally
	// Instead, treat them as auto tracks (they'll be sized based on content)
	isUnbounded := availableSize >= Unbounded*0.9 // Use 90% threshold to avoid float precision issues

	// Separate fixed and fractional tracks
	totalFixed := 0.0
	totalFraction := 0.0
	fixedIndices := []int{}
	fractionIndices := []int{}

	for i, track := range tracks {
		// Resolve track sizes
		minSize := ResolveLength(track.MinSize, ctx, currentFontSize)
		maxSize := ResolveLength(track.MaxSize, ctx, currentFontSize)

		// Check for fit-content (Fraction == -1)
		if track.Fraction == -1 {
			// fit-content: clamp max-content to MaxSize
			// CSS Grid Layout §11.5: fit-content(size)
			// See: https://www.w3.org/TR/css-grid-1/#valdef-grid-template-columns-fit-content
			fixedIndices = append(fixedIndices, i)
			maxContent := resolveIntrinsicTrackSize(track, container, i, isColumn, IntrinsicSizeMaxContent, ctx, currentFontSize)
			size := maxSize
			if size >= Unbounded {
				size = maxContent // Use max-content if no limit specified
			} else {
				size = math.Min(maxContent, size) // Clamp to limit
			}
			sizes[i] = size
			totalFixed += size
		} else if track.Fraction > 0 {
			fractionIndices = append(fractionIndices, i)
			totalFraction += track.Fraction
			// For unbounded constraints, fractional tracks will be treated as auto
			// (sized based on content, not distributed proportionally)
			// Don't set sizes[i] here - it will be handled below
		} else {
			fixedIndices = append(fixedIndices, i)

			// Check for intrinsic sizing sentinel values
			// CSS Grid Layout §11.5: Intrinsic Track Sizing
			// See: https://www.w3.org/TR/css-grid-1/#intrinsic-sizes
			if maxSize == SizeMinContent {
				// min-content track: size based on minimum content size
				sizes[i] = resolveIntrinsicTrackSize(track, container, i, isColumn, IntrinsicSizeMinContent, ctx, currentFontSize)
			} else if maxSize == SizeMaxContent {
				// max-content track: size based on maximum content size
				sizes[i] = resolveIntrinsicTrackSize(track, container, i, isColumn, IntrinsicSizeMaxContent, ctx, currentFontSize)
			} else {
				// Normal fixed track
				size := minSize
				if maxSize < Unbounded {
					size = math.Min(size, maxSize)
				}
				sizes[i] = size
			}

			totalFixed += sizes[i]
		}
	}

	// Distribute fractional space (only when not unbounded)
	if totalFraction > 0 && !isUnbounded {
		remainingSize := availableForTracks - totalFixed
		if remainingSize > 0 {
			for _, i := range fractionIndices {
				sizes[i] = (remainingSize * tracks[i].Fraction) / totalFraction
			}
		} else {
			// Not enough space, use min sizes
			for _, i := range fractionIndices {
				sizes[i] = ResolveLength(tracks[i].MinSize, ctx, currentFontSize)
			}
		}
	} else if totalFraction > 0 && isUnbounded {
		// When unbounded, fractional tracks can't be distributed proportionally
		// They should be sized based on content (treated as auto)
		// For now, use MinSize as a fallback (content-based sizing would require
		// measuring children first, which happens later in the grid algorithm)
		for _, i := range fractionIndices {
			sizes[i] = ResolveLength(tracks[i].MinSize, ctx, currentFontSize)
		}
	} else {
		// All fixed, may need to shrink if total exceeds available
		if totalFixed > availableForTracks && availableForTracks > 0 {
			scale := availableForTracks / totalFixed
			for _, i := range fixedIndices {
				sizes[i] *= scale
				// Clamp to >= 0
				if sizes[i] < 0 {
					sizes[i] = 0
				}
			}
		} else if availableForTracks <= 0 {
			// No available space, set all to min size (or 0)
			for _, i := range fixedIndices {
				minSize := ResolveLength(tracks[i].MinSize, ctx, currentFontSize)
				sizes[i] = math.Max(0, minSize)
			}
		}
	}

	return sizes
}

func sumSizes(sizes []float64) float64 {
	sum := 0.0
	for _, s := range sizes {
		sum += s
	}
	return sum
}
