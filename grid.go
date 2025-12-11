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

	contentWidth := availableWidth - horizontalPadding - horizontalBorder
	contentHeight := availableHeight - verticalPadding - verticalBorder

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

	// Determine grid positions for children
	gridItems := make([]*gridItem, len(children))

	for i, child := range children {
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
			rowStart = i / len(columns)
		}
		if needsAutoCol {
			colStart = i % len(columns)
		}
		// If rowEnd is -1 (explicit auto) or 0 (unset default), set it to rowStart + 1
		// Note: rowEnd=0 is invalid in CSS Grid (would be same as rowStart), so treat as auto
		if rowEnd <= 0 {
			rowEnd = rowStart + 1
		}
		if colEnd <= 0 {
			colEnd = colStart + 1
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

		gridItems[i] = item
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

		// Track required height for each row
		itemHeight := childSize.Height
		spanRows := item.rowEnd - item.rowStart
		heightPerRow := itemHeight / float64(spanRows)

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
		item.node.Rect = Rect{
			X:      cellX + item.node.Style.Margin.Left,
			Y:      cellY + item.node.Style.Margin.Top,
			Width:  cellWidth - item.node.Style.Margin.Left - item.node.Style.Margin.Right,
			Height: cellHeight - item.node.Style.Margin.Top - item.node.Style.Margin.Bottom,
		}
		
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

	node.Rect = Rect{
		X:      0,
		Y:      0,
		Width:  containerSize.Width,
		Height: containerSize.Height,
	}

	return constraints.Constrain(containerSize)
}

type gridItem struct {
	node     *Node
	rowStart int
	rowEnd   int
	colStart int
	colEnd   int
}

func calculateGridTrackSizes(tracks []GridTrack, availableSize float64, gap float64, count int) []float64 {
	if len(tracks) == 0 {
		return []float64{}
	}

	sizes := make([]float64, len(tracks))
	totalGap := gap * float64(len(tracks)-1)
	availableForTracks := availableSize - totalGap

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
		if totalFixed > availableForTracks {
			scale := availableForTracks / totalFixed
			for _, i := range fixedIndices {
				sizes[i] *= scale
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

