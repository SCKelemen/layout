package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Create a bento box style grid with mosaic layout
	// This demonstrates how items can span different numbers of rows/columns
	// Using the Grid() helper function to simplify grid creation
	root := layout.Grid(4, 4, 150, 200) // 4 rows x 4 columns, rows=150px, cols=200px
	root.Style.GridGap = layout.Px(10)
	root.Style.Padding = layout.Uniform(layout.Px(20))
	root.Children = []*layout.Node{
		// Large featured item - spans 2 rows x 2 columns (top-left)
		{
			Style: layout.Style{
				GridRowStart:    0,
				GridRowEnd:      2, // Spans 2 rows
				GridColumnStart: 0,
				GridColumnEnd:   2,              // Spans 2 columns
				Width:           layout.Px(410), // 2 columns + 1 gap
				Height:          layout.Px(310), // 2 rows + 1 gap
			},
		},
		// Medium item - spans 1 row x 2 columns (top-right)
		{
			Style: layout.Style{
				GridRowStart:    0,
				GridRowEnd:      1,
				GridColumnStart: 2,
				GridColumnEnd:   4, // Spans 2 columns
				Width:           layout.Px(410),
				Height:          layout.Px(150),
			},
		},
		// Small item - 1x1 (top-right, second row)
		{
			Style: layout.Style{
				GridRowStart:    1,
				GridRowEnd:      2,
				GridColumnStart: 2,
				GridColumnEnd:   3,
				Width:           layout.Px(200),
				Height:          layout.Px(150),
			},
		},
		// Small item - 1x1 (top-right, second row, second column)
		{
			Style: layout.Style{
				GridRowStart:    1,
				GridRowEnd:      2,
				GridColumnStart: 3,
				GridColumnEnd:   4,
				Width:           layout.Px(200),
				Height:          layout.Px(150),
			},
		},
		// Medium item - spans 2 rows x 1 column (left side, bottom)
		{
			Style: layout.Style{
				GridRowStart:    2,
				GridRowEnd:      4, // Spans 2 rows
				GridColumnStart: 0,
				GridColumnEnd:   1,
				Width:           layout.Px(200),
				Height:          layout.Px(310), // 2 rows + 1 gap
			},
		},
		// Medium item - spans 1 row x 2 columns (bottom, middle)
		{
			Style: layout.Style{
				GridRowStart:    2,
				GridRowEnd:      3,
				GridColumnStart: 1,
				GridColumnEnd:   3, // Spans 2 columns
				Width:           layout.Px(410),
				Height:          layout.Px(150),
			},
		},
		// Small item - 1x1 (bottom-right)
		{
			Style: layout.Style{
				GridRowStart:    2,
				GridRowEnd:      3,
				GridColumnStart: 3,
				GridColumnEnd:   4,
				Width:           layout.Px(200),
				Height:          layout.Px(150),
			},
		},
		// Medium item - spans 1 row x 3 columns (bottom row)
		{
			Style: layout.Style{
				GridRowStart:    3,
				GridRowEnd:      4,
				GridColumnStart: 1,
				GridColumnEnd:   4,              // Spans 3 columns
				Width:           layout.Px(620), // 3 columns + 2 gaps
				Height:          layout.Px(150),
			},
		},
	}

	// Perform layout
	constraints := layout.Loose(900, 700)
	ctx := layout.NewLayoutContext(800, 600, 16)
	size := layout.Layout(root, constraints, ctx)

	fmt.Printf("Bento Box Grid Layout\n")
	fmt.Printf("====================\n\n")
	fmt.Printf("Container size: %.2f x %.2f\n\n", size.Width, size.Height)

	// Describe each item
	descriptions := []string{
		"Large Featured (2x2)",
		"Medium Horizontal (1x2)",
		"Small (1x1)",
		"Small (1x1)",
		"Medium Vertical (2x1)",
		"Medium Horizontal (1x2)",
		"Small (1x1)",
		"Wide Banner (1x3)",
	}

	for i, child := range root.Children {
		spanRows := (child.Style.GridRowEnd - child.Style.GridRowStart)
		spanCols := (child.Style.GridColumnEnd - child.Style.GridColumnStart)
		fmt.Printf("%s [%dx%d]:\n", descriptions[i], spanRows, spanCols)
		fmt.Printf("  Position: (%.2f, %.2f)\n", child.Rect.X, child.Rect.Y)
		fmt.Printf("  Size: %.2f x %.2f\n", child.Rect.Width, child.Rect.Height)
		fmt.Printf("  Grid: row %d-%d, col %d-%d\n\n",
			child.Style.GridRowStart, child.Style.GridRowEnd-1,
			child.Style.GridColumnStart, child.Style.GridColumnEnd-1)
	}

	fmt.Printf("✅ Bento box layout demonstrates:\n")
	fmt.Printf("   - Items spanning multiple rows\n")
	fmt.Printf("   - Items spanning multiple columns\n")
	fmt.Printf("   - Mixed sizes creating mosaic patterns\n")
	fmt.Printf("   - Flexible grid positioning\n")
}
