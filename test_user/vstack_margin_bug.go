// +build ignore

package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Reproduce the VStack margin bug
	// Text nodes with margins in a VStack should have consistent spacing

	// Create text nodes (auto-sized)
	text1 := &layout.Node{
		Style: layout.Style{
			Height:  20, // Simulate text height
			Margin:  layout.Uniform(10),
		},
	}

	text2 := &layout.Node{
		Style: layout.Style{
			Height:  20,
			Margin:  layout.Uniform(10),
		},
	}

	text3 := &layout.Node{
		Style: layout.Style{
			Height:  20,
			Margin:  layout.Uniform(10),
		},
	}

	// Create VStack
	root := layout.VStack(text1, text2, text3)
	root.Style.Width = 200

	constraints := layout.Loose(200, layout.Unbounded)
	layout.Layout(root, constraints)

	fmt.Println("=== VStack with Text Nodes (Margins) ===")
	fmt.Printf("Root height: %.2f\n", root.Rect.Height)
	fmt.Printf("Expected: ~90 (3 * (20 + 10 + 10) = 120, but margins don't double-count)\n")
	fmt.Printf("Expected: ~90 (3 * 20 + 4 * 10 = 100, but first/last margins might collapse)\n")
	fmt.Printf("Actually: ~90 (3 * 20 + 6 * 10 = 120, all margins counted)\n")
	fmt.Println()

	for i, child := range root.Children {
		fmt.Printf("Text %d:\n", i+1)
		fmt.Printf("  Y position: %.2f\n", child.Rect.Y)
		fmt.Printf("  Height: %.2f\n", child.Rect.Height)
		fmt.Printf("  Margin: Top=%.2f, Bottom=%.2f\n", child.Style.Margin.Top, child.Style.Margin.Bottom)
		if i > 0 {
			prev := root.Children[i-1]
			gap := child.Rect.Y - (prev.Rect.Y + prev.Rect.Height)
			fmt.Printf("  Gap from previous: %.2f (should be ~20 = 10 bottom + 10 top)\n", gap)
		}
		fmt.Println()
	}

	// Compare with fixed spacer approach
	fmt.Println("=== VStack with Fixed Spacers (Working) ===")
	spacer1 := layout.Fixed(0, 10)
	text1Fixed := &layout.Node{
		Style: layout.Style{
			Height: Px(20),
		},
	}
	spacer2 := layout.Fixed(0, 10)
	text2Fixed := &layout.Node{
		Style: layout.Style{
			Height: Px(20),
		},
	}
	spacer3 := layout.Fixed(0, 10)
	text3Fixed := &layout.Node{
		Style: layout.Style{
			Height: Px(20),
		},
	}

	root2 := layout.VStack(spacer1, text1Fixed, spacer2, text2Fixed, spacer3, text3Fixed)
	root2.Style.Width = 200

	layout.Layout(root2, constraints)

	fmt.Printf("Root height: %.2f\n", root2.Rect.Height)
	fmt.Printf("Expected: 90 (3 * 10 + 3 * 20 = 90)\n")
	fmt.Println()

	y := 0.0
	for i, child := range root2.Children {
		fmt.Printf("Item %d (type: %v):\n", i+1, child.Style.Height == 10)
		fmt.Printf("  Y position: %.2f\n", child.Rect.Y)
		fmt.Printf("  Height: %.2f\n", child.Rect.Height)
		if i > 0 {
			gap := child.Rect.Y - y
			fmt.Printf("  Gap from previous end: %.2f\n", gap)
		}
		y = child.Rect.Y + child.Rect.Height
	}
}


