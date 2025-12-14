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
		fmt.Println("Usage: go run debug_rows.go <json-file>")
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

	// Re-layout
	constraints := layout.Loose(1000, layout.Unbounded)
	layout.LayoutGrid(node, constraints)

	// Manually calculate expected row positions
	fmt.Println("=== Row Position Analysis ===\n")

	// We need to figure out row heights from the children
	// Child 6: row 2-3, height 200, starts at y=329
	// Child 7: row 3-6, starts at y=557.44

	// If child 6 spans rows 2-3, and it has height 200:
	// - It should fill: row2 + gap + row3 = 200
	// - So: row2 + 8 + row3 = 200
	// - If row2 = row3 = h, then: 2h + 8 = 200, so h = 96

	// Child 6 starts at y=329, which should be the start of row 2
	// Let's calculate where row 2 should start:
	// - Row 0: child 0 spans row 0-1, height 105, starts at 0
	//   So row 0 + gap + row 1 = 105
	//   If row 0 = row 1 = h0, then: 2h0 + 8 = 105, so h0 = 48.5
	// - Row 1: child 1 spans row 1-3, height 436.44, starts at 113
	//   So row 1 + gap + row 2 + gap + row 3 = 436.44
	//   But we also have child 2, 3, 4 in row 1-2 with height 208
	//   And child 5 in row 2-3 with height 220.44

	fmt.Printf("Child 6 analysis:\n")
	fmt.Printf("  Position: row 2-3, y=%.2f, height=%.2f\n", node.Children[6].Rect.Y, node.Children[6].Rect.Height)
	fmt.Printf("  If it fills its cell: row2 + gap + row3 = %.2f\n", node.Children[6].Rect.Height)

	fmt.Printf("\nChild 7 analysis:\n")
	fmt.Printf("  Position: row 3-6, y=%.2f, height=%.2f\n", node.Children[7].Rect.Y, node.Children[7].Rect.Height)
	fmt.Printf("  If it fills its cell: row3 + gap + row4 + gap + row5 = %.2f\n", node.Children[7].Rect.Height)

	// Calculate expected positions
	// Row 0 starts at 0
	// Row 1 starts after row 0 + gap
	// Row 2 starts after row 1 + gap
	// Row 3 starts after row 2 + gap

	// From child 6: it's in row 2-3, starts at 329
	// So row 2 starts at 329

	// From child 7: it's in row 3-6, starts at 557.44
	// So row 3 starts at 557.44

	// Gap between row 2 and row 3 should be: 557.44 - (329 + row2_height)
	// But child 6 spans row 2-3, so it should end at: 329 + row2 + gap + row3

	// If child 6 has height 200 and spans row 2-3:
	// 200 = row2 + gap + row3
	// So child 6 should end at: 329 + 200 = 529
	// And row 3 should start at: 529 + gap = 529 + 8 = 537

	// But row 3 actually starts at 557.44
	// Difference: 557.44 - 537 = 20.44

	fmt.Printf("\nExpected vs Actual:\n")
	fmt.Printf("  Child 6 ends at: %.2f\n", node.Children[6].Rect.Y+node.Children[6].Rect.Height)
	fmt.Printf("  Expected row 3 start: %.2f (child 6 end + gap)\n",
		node.Children[6].Rect.Y+node.Children[6].Rect.Height+8)
	fmt.Printf("  Actual row 3 start: %.2f\n", node.Children[7].Rect.Y)
	fmt.Printf("  Difference: %.2f\n",
		node.Children[7].Rect.Y-(node.Children[6].Rect.Y+node.Children[6].Rect.Height+8))
}
