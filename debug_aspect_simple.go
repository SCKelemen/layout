//go:build ignore

package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	node := &layout.Node{
		Style: layout.Style{
			MinHeight:  100,
			AspectRatio: 2.0,
		},
	}

	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  1000,
		MinHeight: 0,
		MaxHeight: layout.Unbounded,
	}

	size := layout.LayoutBlock(node, constraints)
	fmt.Printf("Measured size: width=%.2f, height=%.2f\n", size.Width, size.Height)
	fmt.Printf("Node rect: width=%.2f, height=%.2f\n", node.Rect.Width, node.Rect.Height)
	fmt.Printf("Expected from aspect ratio: width=1000, height=%.2f\n", 1000.0/2.0)
}

