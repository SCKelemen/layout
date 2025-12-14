package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Create some items with different positions
	items := []*layout.Node{
		layout.Fixed(80, 40),
		layout.Fixed(80, 40),
		layout.Fixed(80, 40),
		layout.Fixed(80, 40),
	}

	// Create a container and layout the items
	root := layout.HStack(items...)
	root.Style.Padding = layout.Uniform(layout.Px(20))
	root.Style.JustifyContent = layout.JustifyContentFlexStart

	constraints := layout.Loose(400, 200)
	ctx := layout.NewLayoutContext(400, 200, 16)
	layout.Layout(root, constraints, ctx)

	fmt.Println("=== Before Alignment ===")
	for i, item := range items {
		fmt.Printf("Item %d: x=%.2f, y=%.2f\n", i, item.Rect.X, item.Rect.Y)
	}

	// Align all items to the left edge
	layout.AlignNodes(items, layout.AlignLeft)
	fmt.Println("\n=== After AlignLeft ===")
	for i, item := range items {
		fmt.Printf("Item %d: x=%.2f, y=%.2f\n", i, item.Rect.X, item.Rect.Y)
	}

	// Reset and align to vertical center
	layout.Layout(root, constraints, ctx)
	layout.AlignNodes(items, layout.AlignCenterY)
	fmt.Println("\n=== After AlignCenterY ===")
	for i, item := range items {
		fmt.Printf("Item %d: x=%.2f, y=%.2f\n", i, item.Rect.X, item.Rect.Y)
	}

	// Reset and distribute horizontally
	layout.Layout(root, constraints, ctx)
	layout.DistributeNodes(items, layout.DistributeHorizontal)
	fmt.Println("\n=== After DistributeHorizontal ===")
	for i, item := range items {
		fmt.Printf("Item %d: x=%.2f, y=%.2f\n", i, item.Rect.X, item.Rect.Y)
	}
}
