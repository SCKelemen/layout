//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/SCKelemen/layout"
)

func main() {
	// Minimal test case matching the bento composition structure
	// 5 columns, mix of fixed and auto rows, items spanning multiple rows

	columns := 5
	gap := 8.0
	width := 1000.0

	// Build grid template columns
	gridColumns := make([]layout.GridTrack, columns)
	for i := 0; i < columns; i++ {
		gridColumns[i] = layout.FractionTrack(1.0)
	}

	// Build grid template rows - row 0 is fixed, rest are auto (matching our actual code)
	gridRows := make([]layout.GridTrack, 6) // We need up to row 6
	gridRows[0] = layout.FixedTrack(60.0)   // Row 0: Fixed height (30-day activity)
	for i := 1; i < 6; i++ {
		gridRows[i] = layout.AutoTrack() // Rows 1-5: Auto-sized
	}

	// Create root grid container
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

	// Row 0: Full width item (30-day activity) - FIXED ROW
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart:    0,
			GridRowEnd:      1,
			GridColumnStart: 0,
			GridColumnEnd:   5,
			MinHeight:       60.0,
		},
	})

	// Row 1: Languages (spans rows 1-3), stat cards (row 1 only) - AUTO ROWS
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart:    1,
			GridRowEnd:      3, // Spans 2 rows
			GridColumnStart: 0,
			GridColumnEnd:   1,
			MinHeight:       200.0, // Languages card height
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart:    1,
			GridRowEnd:      2,
			GridColumnStart: 1,
			GridColumnEnd:   2,
			MinHeight:       50.0, // Stat card height
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart:    1,
			GridRowEnd:      2,
			GridColumnStart: 2,
			GridColumnEnd:   3,
			MinHeight:       50.0,
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart:    1,
			GridRowEnd:      2,
			GridColumnStart: 3,
			GridColumnEnd:   4,
			MinHeight:       50.0,
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart:    1,
			GridRowEnd:      2,
			GridColumnStart: 4,
			GridColumnEnd:   5,
			MinHeight:       50.0,
		},
	})

	// Row 2: Line graph (spans cols 1-3), stat cards (cols 3-5) - AUTO ROWS
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart:    2,
			GridRowEnd:      3,
			GridColumnStart: 1,
			GridColumnEnd:   3,
			MinHeight:       150.0, // Line graph height
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart:    2,
			GridRowEnd:      3,
			GridColumnStart: 3,
			GridColumnEnd:   4,
			MinHeight:       50.0,
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart:    2,
			GridRowEnd:      3,
			GridColumnStart: 4,
			GridColumnEnd:   5,
			MinHeight:       50.0,
		},
	})

	// Row 3: Full width weeks heatmap (spans rows 3-6) - AUTO ROWS
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart:    3,
			GridRowEnd:      6, // Spans 3 rows
			GridColumnStart: 0,
			GridColumnEnd:   5,
			MinHeight:       200.0, // Weeks heatmap height
		},
	})

	// Perform layout calculation
	constraints := layout.Loose(width, layout.Unbounded)
	layout.Layout(root, constraints)

	// Print results
	fmt.Printf("=== Minimal Bento Composition Test ===\n")
	fmt.Printf("Configuration: 5 columns, row 0 fixed (60px), rows 1-5 auto\n")
	fmt.Printf("Gap: %.0fpx\n", gap)
	fmt.Printf("\nRoot dimensions: %.2f x %.2f\n", root.Rect.Width, root.Rect.Height)
	fmt.Printf("Expected height: ~%.2f (60 + 8 + 200 + 8 + 150 + 8 + 200 = 634)\n", 60.0+8.0+200.0+8.0+150.0+8.0+200.0)

	fmt.Printf("\nChild positions:\n")
	for i, child := range root.Children {
		fmt.Printf("  Child %d (row %d-%d, col %d-%d): x=%.2f, y=%.2f, w=%.2f, h=%.2f\n",
			i, child.Style.GridRowStart, child.Style.GridRowEnd-1,
			child.Style.GridColumnStart, child.Style.GridColumnEnd-1,
			child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
	}

	// Check row Y positions
	fmt.Printf("\nRow Y positions:\n")
	rowYPositions := make(map[int]float64)
	for _, child := range root.Children {
		row := child.Style.GridRowStart
		if y, exists := rowYPositions[row]; !exists || child.Rect.Y < y {
			rowYPositions[row] = child.Rect.Y
		}
	}
	for row := 0; row < 6; row++ {
		if y, ok := rowYPositions[row]; ok {
			fmt.Printf("  Row %d: y=%.2f\n", row, y)
		}
	}

	// Check for overlaps
	fmt.Printf("\nChecking for overlaps:\n")
	overlaps := false
	for i := 0; i < len(root.Children); i++ {
		for j := i + 1; j < len(root.Children); j++ {
			c1 := root.Children[i]
			c2 := root.Children[j]
			// Check if they overlap vertically (different rows but too close)
			if c1.Style.GridRowStart != c2.Style.GridRowStart {
				y1End := c1.Rect.Y + c1.Rect.Height
				y2Start := c2.Rect.Y
				// If row 1 ends after row 2 starts, and they're in sequential rows, that's an overlap
				if y1End > y2Start && c1.Style.GridRowEnd <= c2.Style.GridRowStart {
					fmt.Printf("  WARNING: Child %d (row %d-%d, y=%.2f-%.2f) overlaps with child %d (row %d-%d, y=%.2f-%.2f)\n",
						i, c1.Style.GridRowStart, c1.Style.GridRowEnd-1, c1.Rect.Y, c1.Rect.Y+c1.Rect.Height,
						j, c2.Style.GridRowStart, c2.Style.GridRowEnd-1, c2.Rect.Y, c2.Rect.Y+c2.Rect.Height)
					overlaps = true
				}
			}
		}
	}
	if !overlaps {
		fmt.Printf("  ✅ No overlaps detected!\n")
	}

	// Final verdict
	if root.Rect.Height < 400 {
		log.Printf("❌ ERROR: Root height (%.2f) is too small! Expected at least 600\n", root.Rect.Height)
		log.Printf("   This indicates the layout library is not correctly calculating row heights.\n")
	} else {
		fmt.Printf("\n✅ Layout looks correct! Total height: %.2f\n", root.Rect.Height)
	}
}

