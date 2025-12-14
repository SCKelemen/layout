package main

import (
	"fmt"
	"log"
	"github.com/SCKelemen/layout"
)

func main() {
	// Test: Does setting Height vs MinHeight make a difference?
	
	columns := 3
	gap := 8.0
	width := 1000.0

	gridColumns := make([]layout.GridTrack, columns)
	for i := 0; i < columns; i++ {
		gridColumns[i] = layout.FractionTrack(1.0)
	}

	gridRows := make([]layout.GridTrack, 4)
	for i := 0; i < 4; i++ {
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

	// Test 1: Using Height (explicit)
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 0, GridRowEnd: 1,
			GridColumnStart: 0, GridColumnEnd: 3,
			Height: Px(60.0), // Explicit height
		},
	})

	// Test 2: Using MinHeight
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 1, GridRowEnd: 2,
			GridColumnStart: 0, GridColumnEnd: 1,
			MinHeight: Px(50.0), // MinHeight
		},
	})

	// Test 3: Using both Height and MinHeight
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 1, GridRowEnd: 2,
			GridColumnStart: 1, GridColumnEnd: 2,
			Height: Px(50.0),    // Explicit height
			MinHeight: Px(40.0), // Also has MinHeight
		},
	})

	// Test 4: Spanning item with Height
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 2, GridRowEnd: 4, // Spans 2 rows
			GridColumnStart: 0, GridColumnEnd: 3,
			Height: Px(200.0), // Explicit height - should be distributed
		},
	})

	constraints := layout.Loose(width, layout.Unbounded)
	layout.Layout(root, constraints)

	fmt.Printf("=== Test: Height vs MinHeight ===\n")
	fmt.Printf("Root: %.2f x %.2f\n", root.Rect.Width, root.Rect.Height)
	fmt.Printf("\nItems:\n")
	for i, child := range root.Children {
		fmt.Printf("  Item %d: y=%.2f, h=%.2f (Height: %.2f, MinHeight: %.2f)\n",
			i, child.Rect.Y, child.Rect.Height,
			child.Style.Height, child.Style.MinHeight)
	}

	// Check if rows are properly spaced
	fmt.Printf("\nRow spacing:\n")
	prevY := -1.0
	for _, child := range root.Children {
		if prevY >= 0 && child.Rect.Y > prevY {
			gap := child.Rect.Y - prevY
			fmt.Printf("  Gap: %.2f\n", gap)
		}
		if child.Rect.Y+child.Rect.Height > prevY {
			prevY = child.Rect.Y + child.Rect.Height
		}
	}

	if root.Rect.Height < 200 {
		log.Printf("ERROR: Root height (%.2f) is too small!\n", root.Rect.Height)
	}
}

