//go:build ignore

package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	root := &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayGrid,
			GridTemplateRows: []layout.GridTrack{
				layout.AutoTrack(),
			},
			GridTemplateColumns: []layout.GridTrack{
				layout.FractionTrack(1),
			},
			Width: 1000,
		},
		Children: []*layout.Node{
			{
				Style: layout.Style{
					GridRowStart: 0,
					GridRowEnd:   1,
					MinHeight:    100,
					AspectRatio:  2.0,
				},
			},
		},
	}

	constraints := layout.Loose(1000, layout.Unbounded)
	layout.LayoutGrid(root, constraints)

	item := root.Children[0]
	fmt.Printf("Item: width=%.2f, height=%.2f, ratio=%.2f\n",
		item.Rect.Width, item.Rect.Height, item.Rect.Width/item.Rect.Height)
	fmt.Printf("Row 0 height: %.2f\n", root.Rect.Height)
}

