package layout

import (
	"math"
)

// LayoutGrid performs CSS Grid layout on a node
func LayoutGrid(node *Node, constraints Constraints) Size {
	if node.Style.Display != DisplayGrid {
		// If not grid, delegate to block layout
		return LayoutBlock(node, constraints)
	}

	// Calculate available space
	availableWidth := constraints.MaxWidth
	availableHeight := constraints.MaxHeight

	// Account for padding and border
	horizontalPadding := node.Style.Padding.Left + node.Style.Padding.Right
	verticalPadding := node.Style.Padding.Top + node.Style.Padding.Bottom
	horizontalBorder := node.Style.Border.Left + node.Style.Border.Right
	verticalBorder := node.Style.Border.Top + node.Style.Border.Bottom

	// Clamp content size to >= 0
	contentWidth := availableWidth - horizontalPadding - horizontalBorder
	if contentWidth < 0 {
		contentWidth = 0
	}
	contentHeight := availableHeight - verticalPadding - verticalBorder
	if contentHeight < 0 {
		contentHeight = 0
	}

	// Get grid template
	rows := node.Style.GridTemplateRows
	columns := node.Style.GridTemplateColumns

	// Use auto tracks if templates not specified
	if len(rows) == 0 {
		rows = []GridTrack{node.Style.GridAutoRows}
		if len(rows) == 0 || (rows[0].MinSize == 0 && rows[0].MaxSize == Unbounded && rows[0].Fraction == 0) {
			rows = []GridTrack{AutoTrack()}
		}
	}
	if len(columns) == 0 {
		columns = []GridTrack{node.Style.GridAutoColumns}
		if len(columns) == 0 || (columns[0].MinSize == 0 && columns[0].MaxSize == Unbounded && columns[0].Fraction == 0) {
			columns = []GridTrack{AutoTrack()}
		}
	}

	// Calculate gap
	rowGap := node.Style.GridRowGap
	if rowGap == 0 {
		rowGap = node.Style.GridGap
	}
	columnGap := node.Style.GridColumnGap
	if columnGap == 0 {
		columnGap = node.Style.GridGap
	}

	// Step 1: Calculate column sizes
	columnSizes := calculateGridTrackSizes(columns, contentWidth, columnGap, len(columns))

	// Step 2: Calculate row sizes (need to measure children first for auto rows)
	// For now, we'll do a two-pass layout
	children := node.Children
	if len(children) == 0 {
		// Empty grid
		totalWidth := sumSizes(columnSizes) + columnGap*float64(len(columnSizes)-1)
		totalHeight := sumSizes(calculateGridTrackSizes(rows, contentHeight, rowGap, len(rows)))
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

	// Determine grid positions for children (filter DisplayNone)
	gridItems := make([]*gridItem, 0, len(children))
	itemIndex := 0
	for _, child := range children {
		// Skip display:none children
		if child.Style.Display == DisplayNone {
			continue
		}
		item := &gridItem{
			node: child,
		}

		// Get grid position
		rowStart := child.Style.GridRowStart
		rowEnd := child.Style.GridRowEnd
		colStart := child.Style.GridColumnStart
		colEnd := child.Style.GridColumnEnd

		// Auto placement (simplified - just place sequentially)
		// -1 means explicit auto, 0 means unset (default value) - both should trigger auto-placement
		// We need to distinguish between "explicitly set to 0" and "unset (defaults to 0)"
		// For now, we'll treat 0 as unset if rowEnd is also 0 or -1 (unset)
		// This means if both rowStart and rowEnd are unset, we auto-place
		needsAutoRow := rowStart < 0 || (rowStart == 0 && rowEnd <= 0)
		needsAutoCol := colStart < 0 || (colStart == 0 && colEnd <= 0)

		if needsAutoRow {
			// Use itemIndex (which only counts non-DisplayNone children) for auto-placement
			rowStart = itemIndex / len(columns)
			// Set rowEnd to rowStart + 1 for auto-placed items
			rowEnd = rowStart + 1
		} else {
			// If rowEnd is -1 (explicit auto) or 0 (unset default), set it to rowStart + 1
			// Note: rowEnd=0 is invalid in CSS Grid (would be same as rowStart), so treat as auto
			if rowEnd <= 0 {
				rowEnd = rowStart + 1
			}
		}
		
		if needsAutoCol {
			// Use itemIndex (which only counts non-DisplayNone children) for auto-placement
			colStart = itemIndex % len(columns)
			// Set colEnd to colStart + 1 for auto-placed items
			colEnd = colStart + 1
		} else {
			// If colEnd is -1 (explicit auto) or 0 (unset default), set it to colStart + 1
			if colEnd <= 0 {
				colEnd = colStart + 1
			}
		}

		// Ensure we have enough rows/columns
		if rowEnd > len(rows) {
			// Extend rows with auto tracks
			for rowEnd > len(rows) {
				rows = append(rows, node.Style.GridAutoRows)
				if rows[len(rows)-1].MinSize == 0 && rows[len(rows)-1].MaxSize == Unbounded && rows[len(rows)-1].Fraction == 0 {
					rows[len(rows)-1] = AutoTrack()
				}
			}
		}
		if colEnd > len(columns) {
			// Extend columns with auto tracks
			for colEnd > len(columns) {
				columns = append(columns, node.Style.GridAutoColumns)
				if columns[len(columns)-1].MinSize == 0 && columns[len(columns)-1].MaxSize == Unbounded && columns[len(columns)-1].Fraction == 0 {
					columns[len(columns)-1] = AutoTrack()
				}
			}
			// Recalculate column sizes
			columnSizes = calculateGridTrackSizes(columns, contentWidth, columnGap, len(columns))
		}

		item.rowStart = rowStart
		item.rowEnd = rowEnd
		item.colStart = colStart
		item.colEnd = colEnd

		gridItems = append(gridItems, item)
		itemIndex++ // Increment AFTER using itemIndex for auto-placement
	}

	// Step 3: Measure children to determine row sizes
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
			childSize = LayoutFlexbox(item.node, childConstraints)
		} else if item.node.Style.Display == DisplayGrid {
			childSize = LayoutGrid(item.node, childConstraints)
		} else {
			childSize = LayoutBlock(item.node, childConstraints)
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
			// For fixed tracks (MinSize == MaxSize and both > 0), use the track size directly
			// For auto tracks, use measured height
			var trackHeight float64
			if track.MinSize > 0 && track.MinSize == track.MaxSize {
				// Fixed track - use the fixed size
				trackHeight = track.MinSize
			} else {
				// Auto or minmax track - use measured height or track size
				// The measured height comes from children, which respects MinHeight if set
				trackHeight = math.Max(track.MinSize, rowHeights[i])
				if track.MaxSize < Unbounded {
					trackHeight = math.Min(trackHeight, track.MaxSize)
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
				if track.MinSize == track.MaxSize && track.MaxSize < Unbounded {
					// Fixed track that wasn't set - use fixed size
					rowSizes[i] = track.MinSize
				} else {
					// Auto track - use measured height or min size
					// The measured height comes from children, which respects MinHeight if set
					rowSizes[i] = math.Max(track.MinSize, rowHeights[i])
					if track.MaxSize < Unbounded {
						rowSizes[i] = math.Min(rowSizes[i], track.MaxSize)
					}
				}
			}
		}
	}

	// Step 5: Position children
	for _, item := range gridItems {
		// Calculate grid cell position
		cellX := 0.0
		for col := 0; col < item.colStart; col++ {
			cellX += columnSizes[col]
			if col < len(columnSizes)-1 {
				cellX += columnGap
			}
		}

		cellY := 0.0
		for row := 0; row < item.rowStart && row < len(rowSizes); row++ {
			cellY += rowSizes[row]
			if row < len(rowSizes)-1 {
				cellY += rowGap
			}
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
		maxItemWidth := cellWidth - item.node.Style.Margin.Left - item.node.Style.Margin.Right
		maxItemHeight := cellHeight - item.node.Style.Margin.Top - item.node.Style.Margin.Bottom

		var itemWidth, itemHeight float64

		// If item has aspect ratio, maintain it while fitting within cell
		// In CSS Grid, items with aspect ratio maintain their ratio but fit within the cell
		if item.node.Style.AspectRatio > 0 {
			// Calculate dimensions that maintain aspect ratio and fit within cell
			// Try width-based first (fill cell width)
			itemWidth = maxItemWidth
			itemHeight = itemWidth / item.node.Style.AspectRatio

			// If height exceeds cell, constrain by height instead
			if itemHeight > maxItemHeight {
				itemHeight = maxItemHeight
				itemWidth = itemHeight * item.node.Style.AspectRatio
			}

			// Ensure we don't exceed cell width (might happen if constrained by height)
			if itemWidth > maxItemWidth {
				itemWidth = maxItemWidth
				itemHeight = itemWidth / item.node.Style.AspectRatio
			}
		} else {
			// No aspect ratio: stretch to fill cell (default CSS Grid behavior)
			itemWidth = maxItemWidth
			itemHeight = maxItemHeight
		}

		// Position item within grid cell, accounting for margins and padding
		// Margins are applied within the cell boundaries, not extending into gaps
		// For spanning items, margins are still contained within the spanned cell area
		// Add padding offsets to position items within the container's content area
		item.node.Rect = Rect{
			X:      node.Style.Padding.Left + cellX + item.node.Style.Margin.Left,
			Y:      node.Style.Padding.Top + cellY + item.node.Style.Margin.Top,
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

func calculateGridTrackSizes(tracks []GridTrack, availableSize float64, gap float64, count int) []float64 {
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

	// Separate fixed and fractional tracks
	totalFixed := 0.0
	totalFraction := 0.0
	fixedIndices := []int{}
	fractionIndices := []int{}

	for i, track := range tracks {
		if track.Fraction > 0 {
			fractionIndices = append(fractionIndices, i)
			totalFraction += track.Fraction
		} else {
			fixedIndices = append(fixedIndices, i)
			size := track.MinSize
			if track.MaxSize < Unbounded {
				size = math.Min(size, track.MaxSize)
			}
			sizes[i] = size
			totalFixed += size
		}
	}

	// Distribute fractional space
	if totalFraction > 0 {
		remainingSize := availableForTracks - totalFixed
		if remainingSize > 0 {
			for _, i := range fractionIndices {
				sizes[i] = (remainingSize * tracks[i].Fraction) / totalFraction
			}
		} else {
			// Not enough space, use min sizes
			for _, i := range fractionIndices {
				sizes[i] = tracks[i].MinSize
			}
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
				sizes[i] = math.Max(0, tracks[i].MinSize)
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
