package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Create a grid layout with header, sidebar, main, and footer
	root := &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayGrid,
			GridTemplateRows: []layout.GridTrack{
				layout.FixedTrack(layout.Px(80)), // Header
				layout.FractionTrack(1),          // Main content area
				layout.FixedTrack(layout.Px(40)), // Footer
			},
			GridTemplateColumns: []layout.GridTrack{
				layout.FixedTrack(layout.Px(200)), // Sidebar
				layout.FractionTrack(1),           // Main content
			},
			GridGap: layout.Px(10),
			Padding: layout.Uniform(layout.Px(10)),
		},
		Children: []*layout.Node{
			// Header spanning full width
			{
				Style: layout.Style{
					GridRowStart:    0,
					GridRowEnd:      1,
					GridColumnStart: 0,
					GridColumnEnd:   2,
				},
			},
			// Sidebar
			{
				Style: layout.Style{
					GridRowStart:    1,
					GridRowEnd:      2,
					GridColumnStart: 0,
					GridColumnEnd:   1,
				},
			},
			// Main content
			{
				Style: layout.Style{
					GridRowStart:    1,
					GridRowEnd:      2,
					GridColumnStart: 1,
					GridColumnEnd:   2,
				},
			},
			// Footer spanning full width
			{
				Style: layout.Style{
					GridRowStart:    2,
					GridRowEnd:      3,
					GridColumnStart: 0,
					GridColumnEnd:   2,
				},
			},
		},
	}

	// Perform layout
	constraints := layout.Loose(800, 600)
	ctx := layout.NewLayoutContext(800, 600, 16)
	size := layout.Layout(root, constraints, ctx)

	fmt.Printf("Grid container size: %.2f x %.2f\n", size.Width, size.Height)

	areas := []string{"Header", "Sidebar", "Main Content", "Footer"}
	for i, child := range root.Children {
		fmt.Printf("%s: (%.2f, %.2f) %.2f x %.2f\n",
			areas[i], child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
	}
}
