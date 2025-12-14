//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Test aspect ratio in flexbox
	fmt.Println("=== Aspect Ratio in Flexbox ===")

	image := &layout.Node{
		Style: layout.Style{
			Width: Px(800),
			// Height is auto - will be calculated from aspect ratio
		},
	}
	image = layout.AspectRatio(image, 16.0/9.0)

	root := layout.VStack(image)
	root.Style.Width = 1000

	constraints := layout.Loose(1000, layout.Unbounded)
	layout.Layout(root, constraints)

	fmt.Printf("Image Style Width: %.2f\n", image.Style.Width)
	fmt.Printf("Image Rect: %.2f x %.2f\n", image.Rect.Width, image.Rect.Height)
	fmt.Printf("Expected: 800 x 450\n")

	// Test with explicit width in constraints
	fmt.Println("\n=== Direct Block Layout ===")
	image2 := &layout.Node{
		Style: layout.Style{
			Width: Px(800),
		},
	}
	image2 = layout.AspectRatio(image2, 16.0/9.0)

	constraints2 := layout.Loose(1000, layout.Unbounded)
	layout.LayoutBlock(image2, constraints2)

	fmt.Printf("Image2 Rect: %.2f x %.2f\n", image2.Rect.Width, image2.Rect.Height)
}
