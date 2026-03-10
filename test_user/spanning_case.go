package main

import (
	"fmt"
	"log"

	"github.com/SCKelemen/layout"
)

func main() {
	// Test: Item spanning multiple auto rows
	columns := 3
	gap := 8.0
	width := 1000.0

	gridColumns := make([]layout.GridTrack, columns)
	for i := 0; i < columns; i++ {
		gridColumns[i] = layout.FractionTrack(1.0)
	}

	// Auto-sized rows
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

	// Item spanning rows 0-2 (3 rows) with MinHeight
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 0, GridRowEnd: 3, // Spans 3 rows
			GridColumnStart: 0, GridColumnEnd: 1,
			MinHeight: Px(300.0), // Should be distributed across 3 rows
		},
	})

	// Item in row 0, col 1 with MinHeight
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 0, GridRowEnd: 1,
			GridColumnStart: 1, GridColumnEnd: 2,
			MinHeight: Px(100.0),
		},
	})

	// Item in row 1, col 1 with MinHeight
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 1, GridRowEnd: 2,
			GridColumnStart: 1, GridColumnEnd: 2,
			MinHeight: Px(100.0),
		},
	})

	// Item in row 2, col 1 with MinHeight
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 2, GridRowEnd: 3,
			GridColumnStart: 1, GridColumnEnd: 2,
			MinHeight: Px(100.0),
		},
	})

	// Item in row 3, col 0-2 (full width) with MinHeight
	root.Children = append(root.Children, &layout.Node{
		Style: layout.Style{
			GridRowStart: 3, GridRowEnd: 4,
			GridColumnStart: 0, GridColumnEnd: 3,
			MinHeight: Px(50.0),
		},
	})

	constraints := layout.Loose(width, layout.Unbounded)
	ctx := layout.NewLayoutContext(800, 600, 16)
	layout.Layout(root, constraints, ctx)

	fmt.Printf("=== Test: Item spanning multiple auto rows ===\n")
	fmt.Printf("Root: %.2f x %.2f\n", root.Rect.Width, root.Rect.Height)
	fmt.Printf("\nItems:\n")
	for i, child := range root.Children {
		spanRows := child.Style.GridRowEnd - child.Style.GridRowStart
		fmt.Printf("  Item %d: h=%.2f, y=%.2f, spans %d rows (MinHeight: %.2f)\n",
			i, child.Rect.Height, child.Rect.Y, spanRows, child.Style.MinHeight)
	}

	// The spanning item should be 300px tall
	spanningItem := root.Children[0]
	if spanningItem.Rect.Height < 300.0 {
		log.Printf("ERROR: Spanning item should be at least 300px, got %.2f\n", spanningItem.Rect.Height)
	}

	// The spanning item should start at row 0
	if spanningItem.Rect.Y != 0.0 {
		log.Printf("WARNING: Spanning item Y should be 0, got %.2f\n", spanningItem.Rect.Y)
	}

	// Row 3 should be below the spanning item
	row3Item := root.Children[4]
	spanningEnd := spanningItem.Rect.Y + spanningItem.Rect.Height
	if row3Item.Rect.Y <= spanningEnd {
		log.Printf("WARNING: Row 3 should be below spanning item, but row3 Y (%.2f) <= spanning end (%.2f)\n",
			row3Item.Rect.Y, spanningEnd)
	}
}
