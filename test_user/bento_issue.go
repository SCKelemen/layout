package main

import (
	"fmt"
	"log"

	"github.com/SCKelemen/layout"
)

func main() {
	// Simulate a bento layout with spanning items
	// This should help identify the issue

	columns := 3
	gap := 8.0
	width := 1000.0

	gridColumns := make([]layout.GridTrack, columns)
	for i := 0; i < columns; i++ {
		gridColumns[i] = layout.FractionTrack(1.0)
	}

	// Auto-sized rows (like in bento layouts)
	gridRows := make([]layout.GridTrack, 6)
	for i := 0; i < 6; i++ {
		gridRows[i] = layout.AutoTrack()
	}

	root := &layout.Node{
		Style: layout.Style{
			Display:             layout.DisplayGrid,
			GridTemplateColumns: gridColumns,
			GridTemplateRows:    gridRows,
			GridRowGap:          gap,
			GridColumnGap:       gap,
			Width:               width,
		},
		Children: []*layout.Node{},
	}

	// Simulate a bento layout with various spanning items
	// Row 0: Full width item
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 0, GridRowEnd: 1,
			GridColumnStart: 0, GridColumnEnd: 3,
			MinHeight: Px(60.0),
		},
	})

	// Row 1: 3 items side by side
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 1, GridRowEnd: 2,
			GridColumnStart: 0, GridColumnEnd: 1,
			MinHeight: Px(50.0),
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 1, GridRowEnd: 2,
			GridColumnStart: 1, GridColumnEnd: 2,
			MinHeight: Px(50.0),
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 1, GridRowEnd: 2,
			GridColumnStart: 2, GridColumnEnd: 3,
			MinHeight: Px(50.0),
		},
	})

	// Row 2-3: Item spanning 2 rows
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 2, GridRowEnd: 4, // Spans rows 2 and 3
			GridColumnStart: 0, GridColumnEnd: 2,
			MinHeight: Px(200.0), // Should be 100px per row
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 2, GridRowEnd: 3,
			GridColumnStart: 2, GridColumnEnd: 3,
			MinHeight: Px(100.0),
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 3, GridRowEnd: 4,
			GridColumnStart: 2, GridColumnEnd: 3,
			MinHeight: Px(100.0),
		},
	})

	// Row 4-5: Another spanning item
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 4, GridRowEnd: 6, // Spans rows 4 and 5
			GridColumnStart: 0, GridColumnEnd: 3,
			MinHeight: Px(120.0), // Should be 60px per row
		},
	})

	constraints := layout.Loose(width, layout.Unbounded)
	layout.Layout(root, constraints)

	fmt.Printf("=== Bento Layout Test ===\n")
	fmt.Printf("Root: %.2f x %.2f\n", root.Rect.Width, root.Rect.Height)
	fmt.Printf("Expected height: ~%.2f (60 + 8 + 50 + 8 + 100 + 8 + 100 + 8 + 60 + 8 + 60 = 470)\n",
		60.0+8.0+50.0+8.0+100.0+8.0+100.0+8.0+60.0+8.0+60.0)
	fmt.Printf("\nItems:\n")
	for i, child := range root.Children {
		spanRows := child.Style.GridRowEnd - child.Style.GridRowStart
		fmt.Printf("  Item %d: y=%.2f, h=%.2f, spans %d rows (MinHeight: %.2f)\n",
			i, child.Rect.Y, child.Rect.Height, spanRows, child.Style.MinHeight)
	}

	// Check for overlaps
	fmt.Printf("\nChecking row positions:\n")
	rowYPositions := make(map[int]float64)
	for i, child := range root.Children {
		row := child.Style.GridRowStart
		if y, exists := rowYPositions[row]; !exists || child.Rect.Y < y {
			rowYPositions[row] = child.Rect.Y
		}
		fmt.Printf("  Row %d: y=%.2f (item %d)\n", row, child.Rect.Y, i)
	}

	// Verify no overlaps
	fmt.Printf("\nChecking for overlaps:\n")
	overlaps := false
	for i := 0; i < len(root.Children); i++ {
		for j := i + 1; j < len(root.Children); j++ {
			c1 := root.Children[i]
			c2 := root.Children[j]
			if c1.Style.GridRowStart != c2.Style.GridRowStart {
				// Different starting rows, check if they overlap
				c1End := c1.Rect.Y + c1.Rect.Height
				c2Start := c2.Rect.Y
				if c1End > c2Start && c1.Style.GridRowEnd <= c2.Style.GridRowStart {
					fmt.Printf("  OVERLAP: Item %d (row %d-%d, y=%.2f, h=%.2f) overlaps with item %d (row %d-%d, y=%.2f, h=%.2f)\n",
						i, c1.Style.GridRowStart, c1.Style.GridRowEnd-1, c1.Rect.Y, c1.Rect.Height,
						j, c2.Style.GridRowStart, c2.Style.GridRowEnd-1, c2.Rect.Y, c2.Rect.Height)
					overlaps = true
				}
			}
		}
	}
	if !overlaps {
		fmt.Printf("  No overlaps detected ✓\n")
	}

	if root.Rect.Height < 400 {
		log.Printf("ERROR: Root height (%.2f) is too small! Expected ~470px\n", root.Rect.Height)
	}
}
