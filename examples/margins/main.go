package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	fmt.Println("Margin Support Examples")
	fmt.Println("======================")
	fmt.Println()

	// Example 1: HStack with margins
	fmt.Println("1. HStack with margins:")
	stack1 := layout.HStack(
		layout.Margin(layout.Fixed(100, 50), 10),
		layout.Margin(layout.Fixed(100, 50), 10),
		layout.Margin(layout.Fixed(100, 50), 10),
	)
	stack1.Style.Padding = layout.Uniform(20)

	constraints := layout.Loose(500, 200)
	size1 := layout.Layout(stack1, constraints)

	fmt.Printf("   Container: %.2f x %.2f\n", size1.Width, size1.Height)
	for i, child := range stack1.Children {
		fmt.Printf("   Item %d: (%.2f, %.2f) %.2f x %.2f\n",
			i+1, child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
	}
	fmt.Println()

	// Example 2: VStack with margins
	fmt.Println("2. VStack with margins:")
	stack2 := layout.VStack(
		layout.Margin(layout.Fixed(200, 50), 10),
		layout.Margin(layout.Fixed(200, 50), 10),
		layout.Margin(layout.Fixed(200, 50), 10),
	)
	stack2.Style.Padding = layout.Uniform(20)

	size2 := layout.Layout(stack2, constraints)
	fmt.Printf("   Container: %.2f x %.2f\n", size2.Width, size2.Height)
	for i, child := range stack2.Children {
		fmt.Printf("   Item %d: (%.2f, %.2f) %.2f x %.2f\n",
			i+1, child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
	}
	fmt.Println()

	// Example 3: Grid with margins
	fmt.Println("3. Grid with margins:")
	grid := layout.Grid(2, 2, 150, 150)
	grid.Style.GridGap = 10
	grid.Style.Padding = layout.Uniform(20)

	// Create grid items with explicit positions
	item1 := layout.Margin(layout.Fixed(150, 150), 5)
	item1.Style.GridRowStart = 0
	item1.Style.GridColumnStart = 0

	item2 := layout.Margin(layout.Fixed(150, 150), 5)
	item2.Style.GridRowStart = 0
	item2.Style.GridColumnStart = 1

	item3 := layout.Margin(layout.Fixed(150, 150), 5)
	item3.Style.GridRowStart = 1
	item3.Style.GridColumnStart = 0

	item4 := layout.Margin(layout.Fixed(150, 150), 5)
	item4.Style.GridRowStart = 1
	item4.Style.GridColumnStart = 1

	grid.Children = []*layout.Node{item1, item2, item3, item4}

	size3 := layout.Layout(grid, constraints)
	fmt.Printf("   Container: %.2f x %.2f\n", size3.Width, size3.Height)
	for i, child := range grid.Children {
		fmt.Printf("   Item %d: (%.2f, %.2f) %.2f x %.2f\n",
			i+1, child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
	}
	fmt.Println()

	fmt.Println("✅ Margins are fully supported in Flexbox and Grid!")
	fmt.Println("   - Use margins for spacing between items")
	fmt.Println("   - Margins work with HStack, VStack, and Grid")
	fmt.Println("   - Margins don't collapse (CSS-compliant behavior)")
}

