//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Debug aspect ratio
	node := &layout.Node{
		Style: layout.Style{
			Width:       800,
			Height:      -1, // auto
			AspectRatio: 16.0 / 9.0,
		},
	}

	constraints := layout.Loose(1000, 1000)
	size := layout.LayoutBlock(node, constraints)

	fmt.Printf("Node Rect: %.2f x %.2f\n", node.Rect.Width, node.Rect.Height)
	fmt.Printf("Returned size: %.2f x %.2f\n", size.Width, size.Height)
	fmt.Printf("Expected height: %.2f\n", 800.0/(16.0/9.0))
}
