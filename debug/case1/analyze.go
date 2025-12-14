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
		fmt.Println("Usage: go run analyze.go <json-file>")
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

	fmt.Printf("=== Layout Analysis: %s ===\n\n", jsonFile)

	// Print grid structure
	fmt.Println("Grid Structure:")
	fmt.Printf("  Rows: %d (all auto-sized)\n", len(node.Style.GridTemplateRows))
	fmt.Printf("  Columns: %d (fractional: ", len(node.Style.GridTemplateColumns))
	for i, col := range node.Style.GridTemplateColumns {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%.0ffr", col.Fraction)
	}
	fmt.Println(")")
	fmt.Printf("  Gaps: row=%v, col=%v\n", node.Style.GridRowGap, node.Style.GridColumnGap)
	fmt.Printf("  Container: %.2f x %.2f\n\n", node.Rect.Width, node.Rect.Height)

	// Analyze children
	fmt.Println("Children:")
	for i, child := range node.Children {
		rowStart := child.Style.GridRowStart
		rowEnd := child.Style.GridRowEnd
		colStart := child.Style.GridColumnStart
		colEnd := child.Style.GridColumnEnd

		// Handle auto placement
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
		fmt.Printf("    Bottom: %.2f\n", child.Rect.Y+child.Rect.Height)
	}

	// Check for overlaps or gaps
	fmt.Println("\nSpacing Analysis:")
	for i := 0; i < len(node.Children)-1; i++ {
		curr := node.Children[i]
		next := node.Children[i+1]

		// Check if they're in the same column
		currColStart := curr.Style.GridColumnStart
		currColEnd := curr.Style.GridColumnEnd
		nextColStart := next.Style.GridColumnStart
		nextColEnd := next.Style.GridColumnEnd

		if currColStart == 0 {
			currColStart = -1
		}
		if currColEnd == 0 {
			currColEnd = -1
		}
		if nextColStart == 0 {
			nextColStart = -1
		}
		if nextColEnd == 0 {
			nextColEnd = -1
		}

		// Check vertical spacing
		currBottom := curr.Rect.Y + curr.Rect.Height
		nextTop := next.Rect.Y
		gap := nextTop - currBottom

		fmt.Printf("  Between child %d and %d:\n", i, i+1)
		fmt.Printf("    Gap: %.2f (expected: %.2f if same column, or row gap)\n",
			gap, node.Style.GridRowGap)

		// Check if they overlap
		if gap < 0 {
			fmt.Printf("    ⚠️  OVERLAP: %.2f\n", -gap)
		} else if gap > node.Style.GridRowGap+1 {
			fmt.Printf("    ⚠️  LARGE GAP: %.2f (expected ~%.2f)\n", gap, node.Style.GridRowGap)
		}
	}

	// Re-layout to see if results match
	fmt.Println("\nRe-layout Test:")
	constraints := layout.Loose(1000, layout.Unbounded)
	layout.Layout(node, constraints)

	fmt.Printf("  Container after re-layout: %.2f x %.2f\n", node.Rect.Width, node.Rect.Height)
	for i, child := range node.Children {
		fmt.Printf("  Child %d after re-layout: (%.2f, %.2f) %.2f x %.2f\n",
			i, child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
	}
}
