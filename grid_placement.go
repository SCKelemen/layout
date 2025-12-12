package layout

// gridPlaceItems performs grid item placement including auto-placement.
//
// Algorithm based on CSS Grid Layout Module Level 1:
// - §12: Grid Item Placement
// - §12.1: Grid Item Placement Algorithm
// - §8.3: Grid Auto-Flow (simplified implementation)
//
// See: https://www.w3.org/TR/css-grid-1/#placement
// See: https://www.w3.org/TR/css-grid-1/#auto-placement-algo
func gridPlaceItems(node *Node, rows *[]GridTrack, columns *[]GridTrack) []*gridItem {
	children := node.Children
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
		needsAutoRow := rowStart < 0 || (rowStart == 0 && rowEnd <= 0)
		needsAutoCol := colStart < 0 || (colStart == 0 && colEnd <= 0)

		if needsAutoRow {
			// Use itemIndex for auto-placement
			rowStart = itemIndex / len(*columns)
			rowEnd = rowStart + 1
		} else {
			// If rowEnd is -1 (explicit auto) or 0 (unset default), set it to rowStart + 1
			if rowEnd <= 0 {
				rowEnd = rowStart + 1
			}
		}

		if needsAutoCol {
			// Use itemIndex for auto-placement
			colStart = itemIndex % len(*columns)
			colEnd = colStart + 1
		} else {
			// If colEnd is -1 (explicit auto) or 0 (unset default), set it to colStart + 1
			if colEnd <= 0 {
				colEnd = colStart + 1
			}
		}

		// Ensure we have enough rows/columns
		if rowEnd > len(*rows) {
			// Extend rows with auto tracks
			for rowEnd > len(*rows) {
				*rows = append(*rows, node.Style.GridAutoRows)
				if (*rows)[len(*rows)-1].MinSize == 0 && (*rows)[len(*rows)-1].MaxSize == Unbounded && (*rows)[len(*rows)-1].Fraction == 0 {
					(*rows)[len(*rows)-1] = AutoTrack()
				}
			}
		}
		if colEnd > len(*columns) {
			// Extend columns with auto tracks
			for colEnd > len(*columns) {
				*columns = append(*columns, node.Style.GridAutoColumns)
				if (*columns)[len(*columns)-1].MinSize == 0 && (*columns)[len(*columns)-1].MaxSize == Unbounded && (*columns)[len(*columns)-1].Fraction == 0 {
					(*columns)[len(*columns)-1] = AutoTrack()
				}
			}
		}

		item.rowStart = rowStart
		item.rowEnd = rowEnd
		item.colStart = colStart
		item.colEnd = colEnd

		gridItems = append(gridItems, item)
		itemIndex++
	}

	return gridItems
}

// gridPlaceDense performs dense auto-placement algorithm.
//
// Algorithm based on CSS Grid Layout Module Level 1:
// - §8.3.2: Dense Packing
//
// The dense packing algorithm tries to fill in holes left by larger items,
// by placing smaller items in earlier grid cells if they fit.
//
// See: https://www.w3.org/TR/css-grid-1/#auto-placement-algo
func gridPlaceDense(items []*gridItem, rows, columns []GridTrack) {
	// Track which cells are occupied
	occupied := make(map[[2]int]bool)

	// Mark cells occupied by items with explicit positions
	for _, item := range items {
		// Only process items that have at least one explicit dimension
		hasExplicitRow := item.node.Style.GridRowStart >= 0
		hasExplicitCol := item.node.Style.GridColumnStart >= 0

		if hasExplicitRow || hasExplicitCol {
			for r := item.rowStart; r < item.rowEnd; r++ {
				for c := item.colStart; c < item.colEnd; c++ {
					occupied[[2]int{r, c}] = true
				}
			}
		}
	}

	// Now place items without explicit positions using dense packing
	for _, item := range items {
		hasExplicitRow := item.node.Style.GridRowStart >= 0
		hasExplicitCol := item.node.Style.GridColumnStart >= 0

		// Only auto-place items without explicit positions
		if !hasExplicitRow && !hasExplicitCol {
			rowSpan := item.rowEnd - item.rowStart
			colSpan := item.colEnd - item.colStart

			// Try to find the first available spot
			placed := false
			for r := 0; r < len(rows) && !placed; r++ {
				for c := 0; c < len(columns) && !placed; c++ {
					// Check if this position can fit the item
					canFit := true
					if r+rowSpan > len(rows) || c+colSpan > len(columns) {
						canFit = false
					} else {
						for dr := 0; dr < rowSpan && canFit; dr++ {
							for dc := 0; dc < colSpan && canFit; dc++ {
								if occupied[[2]int{r + dr, c + dc}] {
									canFit = false
								}
							}
						}
					}

					if canFit {
						// Place the item here
						item.rowStart = r
						item.rowEnd = r + rowSpan
						item.colStart = c
						item.colEnd = c + colSpan

						// Mark these cells as occupied
						for dr := 0; dr < rowSpan; dr++ {
							for dc := 0; dc < colSpan; dc++ {
								occupied[[2]int{r + dr, c + dc}] = true
							}
						}
						placed = true
					}
				}
			}

			// If we couldn't place it, extend the grid and place at the end
			if !placed {
				// This shouldn't normally happen with proper grid sizing
				// But handle it gracefully
				item.rowStart = len(rows)
				item.rowEnd = item.rowStart + rowSpan
			}
		}
	}
}
