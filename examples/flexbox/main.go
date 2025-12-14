package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Create a column flex container with growing children
	root := &layout.Node{
		Style: layout.Style{
			Display:        layout.DisplayFlex,
			FlexDirection:  layout.FlexDirectionColumn,
			JustifyContent: layout.JustifyContentFlexStart,
			AlignItems:     layout.AlignItemsStretch,
			Padding:        layout.Uniform(layout.Px(20)),
		},
		Children: []*layout.Node{
			{
				Style: layout.Style{
					FlexGrow:  1,
					MinHeight: layout.Px(50),
				},
			},
			{
				Style: layout.Style{
					FlexGrow:  2,
					MinHeight: layout.Px(50),
				},
			},
			{
				Style: layout.Style{
					FlexGrow:  1,
					MinHeight: layout.Px(50),
				},
			},
		},
	}

	// Perform layout with tight constraints
	constraints := layout.Tight(400, 600)
	ctx := layout.NewLayoutContext(800, 600, 16)
	size := layout.Layout(root, constraints, ctx)

	fmt.Printf("Flex container size: %.2f x %.2f\n", size.Width, size.Height)
	fmt.Printf("Available height for flex items: %.2f\n", size.Height-40) // minus padding

	for i, child := range root.Children {
		fmt.Printf("Flex item %d: (%.2f, %.2f) %.2f x %.2f (flex-grow: %.1f)\n",
			i+1, child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height, child.Style.FlexGrow)
	}
}
