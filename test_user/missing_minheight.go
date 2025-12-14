package main

import (
	"fmt"
	"log"

	"github.com/SCKelemen/layout"
)

func main() {
	columns := 3
	gap := 8.0
	width := 1000.0

	gridColumns := make([]layout.GridTrack, columns)
	for i := 0; i < columns; i++ {
		gridColumns[i] = layout.FractionTrack(1.0)
	}

	gridRows := make([]layout.GridTrack, 4)
	for i := 0; i < 4; i++ {
		gridRows[i] = layout.AutoTrack()
	}

	root := &layout.Node{
		Style: layout.Style{
			Display:             layout.DisplayGrid,
			GridTemplateColumns: gridColumns,
			GridTemplateRows:    gridRows,
			GridRowGap:          gap,
			GridColumnGap:       gap,
			Width:               width,
		},
		Children: []*layout.Node{},
	}

	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 0, GridRowEnd: 1,
			GridColumnStart: 0, GridColumnEnd: 3,
			MinHeight: Px(60.0),
		},
	})

	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 1, GridRowEnd: 2,
			GridColumnStart: 0, GridColumnEnd: 1,
			MinHeight: Px(50.0),
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 1, GridRowEnd: 2,
			GridColumnStart: 1, GridColumnEnd: 2,
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 2, GridRowEnd: 3,
			GridColumnStart: 0, GridColumnEnd: 2,
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 2, GridRowEnd: 3,
			GridColumnStart: 2, GridColumnEnd: 3,
		},
	})
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 3, GridRowEnd: 4,
			GridColumnStart: 0, GridColumnEnd: 3,
			MinHeight: Px(50.0),
		},
	})

	constraints := layout.Loose(width, layout.Unbounded)
	layout.Layout(root, constraints)

	fmt.Printf("Root: %.2f x %.2f (expected ~284)\n", root.Rect.Width, root.Rect.Height)
	for i, child := range root.Children {
		fmt.Printf("  Child %d: h=%.2f (MinHeight: %.2f)\n", i, child.Rect.Height, child.Style.MinHeight)
	}

	if root.Rect.Height < 200 {
		log.Printf("ERROR: Root height too small! Demonstrates the bug.\n")
	}
}

