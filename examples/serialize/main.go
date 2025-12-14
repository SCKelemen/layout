package main

import (
	"fmt"
	"os"

	"github.com/SCKelemen/layout"
	"github.com/SCKelemen/layout/serialize"
)

func main() {
	fmt.Println("=== Layout Tree Serialization Example ===")
	fmt.Println()

	// Create a layout tree
	root := layout.VStack(
		layout.Fixed(100, 50),
		layout.Fixed(100, 50),
	)
	root.Style.Width = layout.Px(200)
	root.Style.Padding = layout.Uniform(layout.Px(10))

	// Perform layout
	constraints := layout.Loose(200, layout.Unbounded)
	ctx := layout.NewLayoutContext(800, 600, 16)
	layout.Layout(root, constraints, ctx)

	// Serialize to JSON
	fmt.Println("1. Serializing to JSON...")
	jsonBytes, err := serialize.ToJSON(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonBytes))
	fmt.Println()

	// Deserialize from JSON
	fmt.Println("2. Deserializing from JSON...")
	deserialized, err := serialize.FromJSON(jsonBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Deserialized root: %.2f x %.2f\n", deserialized.Rect.Width, deserialized.Rect.Height)
	fmt.Printf("Children count: %d\n", len(deserialized.Children))
	fmt.Println()

	// Example with grid
	fmt.Println("3. Grid layout serialization...")
	grid := layout.Grid(2, 2, 100, 100)
	grid.Children[0].Style.GridRowStart = 0
	grid.Children[0].Style.GridRowEnd = 2 // Span 2 rows
	grid.Children[0].Style.GridColumnStart = 0
	grid.Children[0].Style.GridColumnEnd = 1

	constraints2 := layout.Loose(200, 200)
	ctx2 := layout.NewLayoutContext(800, 600, 16)
	layout.Layout(grid, constraints2, ctx2)

	gridJSON, _ := serialize.ToJSON(grid)
	fmt.Println(string(gridJSON))
}
