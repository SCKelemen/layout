//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Debug: why isn't aspect ratio working?
	node := &layout.Node{
		Style: layout.Style{
			Width:       800,
			Height:      -1, // auto
			AspectRatio: 16.0 / 9.0,
		},
	}

	fmt.Printf("Before layout:\n")
	fmt.Printf("  Width: %.2f\n", node.Style.Width)
	fmt.Printf("  Height: %.2f\n", node.Style.Height)
	fmt.Printf("  AspectRatio: %.2f\n", node.Style.AspectRatio)

	constraints := layout.Loose(1000, 1000)
	size := layout.LayoutBlock(node, constraints)

	fmt.Printf("\nAfter layout:\n")
	fmt.Printf("  Rect Width: %.2f\n", node.Rect.Width)
	fmt.Printf("  Rect Height: %.2f\n", node.Rect.Height)
	fmt.Printf("  Returned size: %.2f x %.2f\n", size.Width, size.Height)
	fmt.Printf("  Expected height: %.2f\n", 800.0/(16.0/9.0))
}
