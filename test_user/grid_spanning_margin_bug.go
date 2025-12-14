// +build ignore

package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Reproduce the bug: margin duplicated when item spans rows
	fmt.Println("=== Grid Spanning Item Margin Bug ===")
	
	// Create a grid with items, one spanning multiple rows
	root := &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayGrid,
			GridTemplateRows: []layout.GridTrack{
				layout.FixedTrack(50),
				layout.FixedTrack(50),
				layout.FixedTrack(50),
			},
			GridTemplateColumns: []layout.GridTrack{
				layout.FixedTrack(100),
				layout.FixedTrack(100),
			},
			GridRowGap: 10,
		},
		Children: []*layout.Node{
			// Item 1: Spans rows 0-1 (2 rows)
			{
				Style: layout.Style{
					GridRowStart:    0,
					GridRowEnd:      2, // Spans 2 rows
					GridColumnStart: 0,
					GridColumnEnd:   1,
					Height:          110, // 2 * 50 + 1 * 10 gap
					Margin:          layout.Uniform(5),
				},
			},
			// Item 2: Row 0, column 1
			{
				Style: layout.Style{
					GridRowStart:    0,
					GridRowEnd:      1,
					GridColumnStart: 1,
					GridColumnEnd:   2,
					Height:          50,
					Margin:          layout.Uniform(5),
				},
			},
			// Item 3: Row 1, column 1
			{
				Style: layout.Style{
					GridRowStart:    1,
					GridRowEnd:      2,
					GridColumnStart: 1,
					GridColumnEnd:   2,
					Height:          50,
					Margin:          layout.Uniform(5),
				},
			},
			// Item 4: Row 2, column 0
			{
				Style: layout.Style{
					GridRowStart:    2,
					GridRowEnd:      3,
					GridColumnStart: 0,
					GridColumnEnd:   1,
					Height:          50,
					Margin:          layout.Uniform(5),
				},
			},
		},
	}

	constraints := layout.Loose(200, layout.Unbounded)
	layout.Layout(root, constraints)

	fmt.Printf("Grid height: %.2f\n", root.Rect.Height)
	fmt.Printf("Expected: 3 * 50 + 2 * 10 = 170\n")
	fmt.Println()

	// Analyze each item
	for i, item := range root.Children {
		fmt.Printf("Item %d (row %d-%d, col %d-%d):\n",
			i+1, item.Style.GridRowStart, item.Style.GridRowEnd-1,
			item.Style.GridColumnStart, item.Style.GridColumnEnd-1)
		fmt.Printf("  Y position: %.2f\n", item.Rect.Y)
		fmt.Printf("  Height: %.2f\n", item.Rect.Height)
		fmt.Printf("  Bottom: %.2f\n", item.Rect.Y+item.Rect.Height)
		fmt.Printf("  Margin: Top=%.2f, Bottom=%.2f\n", item.Style.Margin.Top, item.Style.Margin.Bottom)
		fmt.Println()
	}

	// Check spacing between rows
	fmt.Println("=== Row Spacing Analysis ===")
	
	// Row 0 items
	row0Items := []*layout.Node{}
	row1Items := []*layout.Node{}
	row2Items := []*layout.Node{}
	
	for _, item := range root.Children {
		if item.Style.GridRowStart == 0 {
			row0Items = append(row0Items, item)
		}
		if item.Style.GridRowStart == 1 || (item.Style.GridRowStart == 0 && item.Style.GridRowEnd > 2) {
			row1Items = append(row1Items, item)
		}
		if item.Style.GridRowStart == 2 || (item.Style.GridRowStart <= 1 && item.Style.GridRowEnd > 3) {
			row2Items = append(row2Items, item)
		}
	}
	
	// Check gap between row 0 and row 1
	if len(row0Items) > 0 && len(row1Items) > 0 {
		row0Bottom := 0.0
		for _, item := range row0Items {
			bottom := item.Rect.Y + item.Rect.Height + item.Style.Margin.Bottom
			if bottom > row0Bottom {
				row0Bottom = bottom
			}
		}
		
		row1Top := 1000.0
		for _, item := range row1Items {
			top := item.Rect.Y - item.Style.Margin.Top
			if top < row1Top {
				row1Top = top
			}
		}
		
		gap := row1Top - row0Bottom
		expectedGap := 10.0 // GridRowGap
		fmt.Printf("Gap between row 0 and row 1:\n")
		fmt.Printf("  Row 0 bottom (with margin): %.2f\n", row0Bottom)
		fmt.Printf("  Row 1 top (with margin): %.2f\n", row1Top)
		fmt.Printf("  Gap: %.2f (expected: %.2f)\n", gap, expectedGap)
		if gap != expectedGap {
			fmt.Printf("  ❌ BUG: Gap is incorrect! Difference: %.2f\n", gap-expectedGap)
		}
		fmt.Println()
	}
	
	// Check gap between row 1 and row 2
	if len(row1Items) > 0 && len(row2Items) > 0 {
		row1Bottom := 0.0
		for _, item := range row1Items {
			bottom := item.Rect.Y + item.Rect.Height + item.Style.Margin.Bottom
			if bottom > row1Bottom {
				row1Bottom = bottom
			}
		}
		
		row2Top := 1000.0
		for _, item := range row2Items {
			top := item.Rect.Y - item.Style.Margin.Top
			if top < row2Top {
				row2Top = top
			}
		}
		
		gap := row2Top - row1Bottom
		expectedGap := 10.0 // GridRowGap
		fmt.Printf("Gap between row 1 and row 2:\n")
		fmt.Printf("  Row 1 bottom (with margin): %.2f\n", row1Bottom)
		fmt.Printf("  Row 2 top (with margin): %.2f\n", row2Top)
		fmt.Printf("  Gap: %.2f (expected: %.2f)\n", gap, expectedGap)
		if gap != expectedGap {
			fmt.Printf("  ❌ BUG: Gap is incorrect! Difference: %.2f\n", gap-expectedGap)
		}
	}
}


