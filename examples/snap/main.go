package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	fmt.Println("=== Snapping Example ===")
	fmt.Println()
	fmt.Println("Snapping is primarily for block layouts and absolute positioning.")
	fmt.Println("It's NOT recommended for Flexbox/Grid items as it breaks their layout algorithm.")
	fmt.Println()

	// Example 1: Snapping absolutely positioned elements
	fmt.Println("Example 1: Snapping absolutely positioned elements")
	root := &layout.Node{
		Style: layout.Style{
			Position: layout.PositionRelative,
			Width:    layout.Px(400),
			Height:   layout.Px(300),
		},
		Children: []*layout.Node{
			{
				Style: layout.Style{
					Position: layout.PositionAbsolute,
					Left:     layout.Px(12.3), // Will snap to 10
					Top:      layout.Px(17.8), // Will snap to 20
					Width:    layout.Px(50),
					Height:   layout.Px(50),
				},
			},
			{
				Style: layout.Style{
					Position: layout.PositionAbsolute,
					Left:     layout.Px(23.7), // Will snap to 20
					Top:      layout.Px(45.2), // Will snap to 50
					Width:    layout.Px(50),
					Height:   layout.Px(50),
				},
			},
		},
	}

	constraints := layout.Loose(400, 300)
	viewport := layout.Rect{X: 0, Y: 0, Width: 400, Height: 300}
	ctx := layout.NewLayoutContext(800, 600, 16)
	layout.LayoutWithPositioning(root, constraints, viewport, ctx)

	fmt.Println("Before snapping:")
	for i, child := range root.Children {
		fmt.Printf("  Item %d: x=%.2f, y=%.2f\n", i, child.Rect.X, child.Rect.Y)
	}

	// Snap to 10px grid
	layout.SnapNodes(root.Children, 10.0)

	fmt.Println("\nAfter snapping to 10px grid:")
	for i, child := range root.Children {
		fmt.Printf("  Item %d: x=%.2f, y=%.2f\n", i, child.Rect.X, child.Rect.Y)
	}

	// Example 2: Snapping with offset grid
	fmt.Println("\n\nExample 2: Snapping to offset grid (origin at 5, 5)")
	root2 := &layout.Node{
		Style: layout.Style{
			Position: layout.PositionRelative,
			Width:    layout.Px(400),
			Height:   layout.Px(300),
		},
		Children: []*layout.Node{
			{
				Style: layout.Style{
					Position: layout.PositionAbsolute,
					Left:     layout.Px(17.3), // 12.3 relative to (5, 5) -> snaps to 15
					Top:      layout.Px(22.8), // 17.8 relative to (5, 5) -> snaps to 25
					Width:    layout.Px(50),
					Height:   layout.Px(50),
				},
			},
		},
	}

	layout.LayoutWithPositioning(root2, constraints, viewport, ctx)
	fmt.Println("Before snapping:")
	fmt.Printf("  Item 0: x=%.2f, y=%.2f\n", root2.Children[0].Rect.X, root2.Children[0].Rect.Y)

	layout.SnapToGrid(root2.Children, 10.0, 5.0, 5.0)
	fmt.Println("After snapping to 10px grid with origin (5, 5):")
	fmt.Printf("  Item 0: x=%.2f, y=%.2f\n", root2.Children[0].Rect.X, root2.Children[0].Rect.Y)

	fmt.Println("\n=== Note ===")
	fmt.Println("For Flexbox/Grid layouts, use container alignment properties instead:")
	fmt.Println("  - Flexbox: justify-content, align-items")
	fmt.Println("  - Grid: justify-items, align-items, grid gaps")
}
