//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Test case: Text nodes with auto height and margins
	// This might reveal the bug - if height is auto, margins might not be respected correctly

	// Create text nodes WITHOUT explicit height (auto height)
	text1 := &layout.Node{
		Style: layout.Style{
			// Height is auto (-1 or 0)
			MinHeight: Px(20), // Text has some minimum height
			Margin:    layout.Uniform(10),
		},
	}

	text2 := &layout.Node{
		Style: layout.Style{
			MinHeight: Px(20),
			Margin:    layout.Uniform(10),
		},
	}

	text3 := &layout.Node{
		Style: layout.Style{
			MinHeight: Px(20),
			Margin:    layout.Uniform(10),
		},
	}

	// Create VStack
	root := layout.VStack(text1, text2, text3)
	root.Style.Width = 200

	constraints := layout.Loose(200, layout.Unbounded)
	layout.Layout(root, constraints)

	fmt.Println("=== VStack with Auto-Height Text Nodes (Margins) ===")
	fmt.Printf("Root height: %.2f\n", root.Rect.Height)
	fmt.Printf("Expected: ~120 (3 * 20 + 6 * 10 = 120)\n")
	fmt.Println()

	for i, child := range root.Children {
		fmt.Printf("Text %d:\n", i+1)
		fmt.Printf("  Y position: %.2f\n", child.Rect.Y)
		fmt.Printf("  Height: %.2f (MinHeight: %.2f)\n", child.Rect.Height, child.Style.MinHeight)
		fmt.Printf("  Margin: Top=%.2f, Bottom=%.2f\n", child.Style.Margin.Top, child.Style.Margin.Bottom)
		if i > 0 {
			prev := root.Children[i-1]
			gap := child.Rect.Y - (prev.Rect.Y + prev.Rect.Height)
			expectedGap := prev.Style.Margin.Bottom + child.Style.Margin.Top
			fmt.Printf("  Gap from previous: %.2f (expected: %.2f)\n", gap, expectedGap)
			if gap != expectedGap {
				fmt.Printf("  ❌ BUG: Gap is incorrect!\n")
			}
		}
		fmt.Println()
	}

	// Test with mixed explicit and auto heights
	fmt.Println("=== Mixed Explicit and Auto Heights ===")
	mixed1 := &layout.Node{
		Style: layout.Style{
			Height: Px(30), // Explicit
			Margin: layout.Uniform(10),
		},
	}
	mixed2 := &layout.Node{
		Style: layout.Style{
			// Auto height
			MinHeight: Px(20),
			Margin:    layout.Uniform(10),
		},
	}
	mixed3 := &layout.Node{
		Style: layout.Style{
			Height: Px(25), // Explicit
			Margin: layout.Uniform(10),
		},
	}

	root2 := layout.VStack(mixed1, mixed2, mixed3)
	root2.Style.Width = 200
	layout.Layout(root2, constraints)

	fmt.Printf("Root height: %.2f\n", root2.Rect.Height)
	for i, child := range root2.Children {
		fmt.Printf("Item %d: Y=%.2f, H=%.2f\n", i+1, child.Rect.Y, child.Rect.Height)
		if i > 0 {
			prev := root2.Children[i-1]
			gap := child.Rect.Y - (prev.Rect.Y + prev.Rect.Height)
			expectedGap := prev.Style.Margin.Bottom + child.Style.Margin.Top
			fmt.Printf("  Gap: %.2f (expected: %.2f)\n", gap, expectedGap)
			if gap != expectedGap {
				fmt.Printf("  ❌ BUG: Gap is incorrect!\n")
			}
		}
	}
}
