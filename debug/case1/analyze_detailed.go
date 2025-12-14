//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/SCKelemen/layout"
	"github.com/SCKelemen/layout/serialize"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run analyze_detailed.go <json-file>")
		os.Exit(1)
	}

	jsonFile := os.Args[1]
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	node, err := serialize.FromJSON(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== Detailed Analysis: %s ===\n\n", jsonFile)

	// Re-layout to get current state
	constraints := layout.Loose(1000, layout.Unbounded)
	layout.LayoutGrid(node, constraints)

	// Print all children with their positions
	fmt.Println("All Children:")
	for i, child := range node.Children {
		rowStart := child.Style.GridRowStart
		rowEnd := child.Style.GridRowEnd
		colStart := child.Style.GridColumnStart
		colEnd := child.Style.GridColumnEnd

		if rowStart == 0 && rowEnd == 0 {
			rowStart = -1
			rowEnd = -1
		}
		if colStart == 0 && colEnd == 0 {
			colStart = -1
			colEnd = -1
		}

		fmt.Printf("  Child %d:\n", i)
		fmt.Printf("    Grid: row %d-%d, col %d-%d\n", rowStart, rowEnd, colStart, colEnd)
		fmt.Printf("    MinHeight: %.2f\n", child.Style.MinHeight)
		fmt.Printf("    Rect: (%.2f, %.2f) %.2f x %.2f\n",
			child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
		fmt.Printf("    Top: %.2f, Bottom: %.2f\n", child.Rect.Y, child.Rect.Y+child.Rect.Height)
	}

	// Focus on child 6 and 7
	if len(node.Children) >= 8 {
		child6 := node.Children[6]
		child7 := node.Children[7]

		fmt.Println("\n=== Child 6 and 7 Analysis ===")
		fmt.Printf("Child 6: row %d-%d, y=%.2f, height=%.2f, bottom=%.2f\n",
			child6.Style.GridRowStart, child6.Style.GridRowEnd,
			child6.Rect.Y, child6.Rect.Height, child6.Rect.Y+child6.Rect.Height)
		fmt.Printf("Child 7: row %d-%d, y=%.2f, height=%.2f, bottom=%.2f\n",
			child7.Style.GridRowStart, child7.Style.GridRowEnd,
			child7.Rect.Y, child7.Rect.Height, child7.Rect.Y+child7.Rect.Height)

		gap := child7.Rect.Y - (child6.Rect.Y + child6.Rect.Height)
		fmt.Printf("Gap: %.2f (expected: 8.00)\n", gap)

		// Calculate expected positions
		// Child 6 is in row 2-3 (rowEnd is exclusive, so it's only in row 2)
		// Child 7 is in row 3-6 (starts at row 3)
		// So child 6 should end at the end of row 2, and child 7 should start at the start of row 3
		// Gap should be 8px (rowGap)

		fmt.Printf("\nExpected:\n")
		fmt.Printf("  Child 6 should end at: end of row 2\n")
		fmt.Printf("  Child 7 should start at: start of row 3\n")
		fmt.Printf("  Gap between row 2 and row 3: %.2f\n", node.Style.GridRowGap)
	}
}
