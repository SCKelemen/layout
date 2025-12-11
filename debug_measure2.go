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

	// Simulate what happens in block layout
	availableWidth := constraints.MaxWidth
	availableHeight := constraints.MaxHeight
	
	contentWidth := availableWidth
	contentHeight := availableHeight
	
	nodeWidth := -1.0 // auto
	nodeHeight := -1.0 // auto
	
	if nodeWidth < 0 {
		nodeWidth = contentWidth // 1000
	}
	if nodeHeight < 0 {
		nodeHeight = contentHeight // Unbounded
	}
	
	fmt.Printf("After initial: nodeWidth=%.2f, nodeHeight=%.2f\n", nodeWidth, nodeHeight)
	
	// Apply aspect ratio
	if node.Style.AspectRatio > 0 {
		if node.Style.Width < 0 && node.Style.Height < 0 {
			if contentWidth > 0 {
				nodeHeight = nodeWidth / node.Style.AspectRatio
				fmt.Printf("After aspect ratio: nodeWidth=%.2f, nodeHeight=%.2f\n", nodeWidth, nodeHeight)
			}
		}
	}
	
	// Apply MinHeight
	if node.Style.MinHeight > 0 {
		nodeHeight = max(nodeHeight, node.Style.MinHeight)
		fmt.Printf("After MinHeight: nodeWidth=%.2f, nodeHeight=%.2f\n", nodeWidth, nodeHeight)
	}
	
	// Constrain to available space
	if nodeWidth > contentWidth {
		nodeWidth = contentWidth
	}
	if nodeHeight > contentHeight && contentHeight < layout.Unbounded {
		nodeHeight = contentHeight
	}
	
	fmt.Printf("After constraints: nodeWidth=%.2f, nodeHeight=%.2f\n", nodeWidth, nodeHeight)
	
	// Later: if width is auto, use max child width
	maxChildWidth := 0.0 // no children
	aspectRatioCalculatedWidth := true // from aspect ratio calculation
	
	if node.Style.Width < 0 {
		if !aspectRatioCalculatedWidth {
			nodeWidth = maxChildWidth
		}
	}
	
	fmt.Printf("After children width: nodeWidth=%.2f, nodeHeight=%.2f\n", nodeWidth, nodeHeight)
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

