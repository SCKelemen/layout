package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Create a flex container with two children
	root := &layout.Node{
		Style: layout.Style{
			Display:        layout.DisplayFlex,
			FlexDirection:  layout.FlexDirectionRow,
			JustifyContent: layout.JustifyContentSpaceBetween,
			AlignItems:     layout.AlignItemsCenter,
			Padding:        layout.Uniform(20),
		},
		Children: []*layout.Node{
			{
				Style: layout.Style{
					Width:  100,
					Height: 50,
				},
			},
			{
				Style: layout.Style{
					Width:  100,
					Height: 50,
				},
			},
		},
	}

	// Perform layout with loose constraints
	constraints := layout.Loose(800, 600)
	size := layout.Layout(root, constraints)

	fmt.Printf("Root container size: %.2f x %.2f\n", size.Width, size.Height)
	fmt.Printf("Root rect: (%.2f, %.2f) %.2f x %.2f\n",
		root.Rect.X, root.Rect.Y, root.Rect.Width, root.Rect.Height)

	for i, child := range root.Children {
		fmt.Printf("Child %d rect: (%.2f, %.2f) %.2f x %.2f\n",
			i+1, child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
	}
}
