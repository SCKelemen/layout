package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Example: Multi-column grid (3 columns, 2 rows)
	// This demonstrates that Grid DOES support multiple columns!

	root := &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayGrid,
			GridTemplateRows: []layout.GridTrack{
				layout.FixedTrack(layout.Px(100)),
				layout.FixedTrack(layout.Px(100)),
			},
			GridTemplateColumns: []layout.GridTrack{
				layout.FixedTrack(layout.Px(150)), // Column 1
				layout.FixedTrack(layout.Px(150)), // Column 2
				layout.FixedTrack(layout.Px(150)), // Column 3
			},
			GridGap: layout.Px(10),
			Padding: layout.Uniform(layout.Px(10)),
		},
		Children: []*layout.Node{
			// Row 1: 3 columns
			{Style: layout.Style{GridRowStart: 0, GridColumnStart: 0}}, // Col 1
			{Style: layout.Style{GridRowStart: 0, GridColumnStart: 1}}, // Col 2
			{Style: layout.Style{GridRowStart: 0, GridColumnStart: 2}}, // Col 3
			// Row 2: 3 columns
			{Style: layout.Style{GridRowStart: 1, GridColumnStart: 0}}, // Col 1
			{Style: layout.Style{GridRowStart: 1, GridColumnStart: 1}}, // Col 2
			{Style: layout.Style{GridRowStart: 1, GridColumnStart: 2}}, // Col 3
		},
	}

	constraints := layout.Loose(500, 250)
	ctx := layout.NewLayoutContext(800, 600, 16)
	size := layout.Layout(root, constraints, ctx)

	fmt.Printf("Multi-column grid (3 columns x 2 rows):\n")
	fmt.Printf("Container size: %.2f x %.2f\n\n", size.Width, size.Height)

	for i, child := range root.Children {
		row := i / 3
		col := i % 3
		fmt.Printf("Row %d, Col %d: (%.2f, %.2f) %.2f x %.2f\n",
			row, col, child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
	}

	fmt.Println("\n✅ Grid supports multiple columns! This is a 3-column grid.")
}
