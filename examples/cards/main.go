package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Example: Layout cards for a GitHub README graph
	// Cards will be arranged in a grid, some with rotations for visual interest

	root := &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayGrid,
			GridTemplateColumns: []layout.GridTrack{
				layout.FixedTrack(150),
				layout.FixedTrack(150),
				layout.FixedTrack(150),
			},
			GridTemplateRows: []layout.GridTrack{
				layout.FixedTrack(100),
				layout.FixedTrack(100),
			},
			GridGap: 20,
			Padding: layout.Uniform(20),
		},
		Children: []*layout.Node{
			// Card 1: Normal
			{
				Style: layout.Style{
					GridRowStart:    0,
					GridColumnStart: 0,
					Width:           150,
					Height:          100,
				},
			},
			// Card 2: Slightly rotated
			{
				Style: layout.Style{
					GridRowStart:    0,
					GridColumnStart: 1,
					Width:           150,
					Height:          100,
					Transform:       layout.RotateDegrees(5),
				},
			},
			// Card 3: Normal
			{
				Style: layout.Style{
					GridRowStart:    0,
					GridColumnStart: 2,
					Width:           150,
					Height:          100,
				},
			},
			// Card 4: Slightly rotated opposite direction
			{
				Style: layout.Style{
					GridRowStart:    1,
					GridColumnStart: 0,
					Width:           150,
					Height:          100,
					Transform:       layout.RotateDegrees(-3),
				},
			},
			// Card 5: Normal
			{
				Style: layout.Style{
					GridRowStart:    1,
					GridColumnStart: 1,
					Width:           150,
					Height:          100,
				},
			},
			// Card 6: Slightly scaled and rotated
			{
				Style: layout.Style{
					GridRowStart:    1,
					GridColumnStart: 2,
					Width:           150,
					Height:          100,
					Transform:       layout.Scale(1.05, 1.05).Multiply(layout.RotateDegrees(2)),
				},
			},
		},
	}

	// Perform layout
	constraints := layout.Loose(600, 300)
	size := layout.Layout(root, constraints)

	fmt.Printf("Grid layout size: %.2f x %.2f\n\n", size.Width, size.Height)
	fmt.Println("Card positions for SVG rendering:")
	fmt.Println("=====================================")

	// Collect all nodes for SVG rendering
	var nodes []*layout.Node
	layout.CollectNodesForSVG(root, &nodes)

	for i, node := range nodes {
		if i == 0 {
			continue // Skip root
		}
		rect := layout.GetFinalRect(node)
		transform := layout.GetSVGTransform(node)

		fmt.Printf("Card %d:\n", i)
		fmt.Printf("  Position: (%.2f, %.2f)\n", rect.X, rect.Y)
		fmt.Printf("  Size: %.2f x %.2f\n", rect.Width, rect.Height)
		if transform != "" {
			fmt.Printf("  Transform: %s\n", transform)
		}
		fmt.Println()
	}

	// Example SVG output
	fmt.Println("Example SVG (simplified):")
	fmt.Println("=====================================")
	fmt.Printf(`<svg width="%.0f" height="%.0f" xmlns="http://www.w3.org/2000/svg">`+"\n", size.Width, size.Height)
	for i, node := range nodes {
		if i == 0 {
			continue // Skip root
		}
		rect := node.Rect
		transform := layout.GetSVGTransform(node)

		// Example: render as rounded rectangles
		fmt.Printf(`  <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="5" fill="#4CAF50"`,
			rect.X, rect.Y, rect.Width, rect.Height)
		if transform != "" {
			fmt.Printf(` transform="%s"`, transform)
		}
		fmt.Printf(` />` + "\n")
	}
	fmt.Println("</svg>")
}
