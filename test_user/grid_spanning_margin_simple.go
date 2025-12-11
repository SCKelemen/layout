// +build ignore

package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Simple test: spanning item with margin should not affect row gap
	fmt.Println("=== Simple Spanning Margin Test ===")
	
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
			},
			GridRowGap: 10,
		},
		Children: []*layout.Node{
			// Item spanning rows 0-1
			{
				Style: layout.Style{
					GridRowStart:    0,
					GridRowEnd:      2,
					GridColumnStart: 0,
					GridColumnEnd:   1,
					Height:          110, // 2 * 50 + 1 * 10 gap
					Margin:          layout.Uniform(5),
				},
			},
			// Item in row 2
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

	constraints := layout.Loose(100, layout.Unbounded)
	layout.Layout(root, constraints)

	fmt.Printf("Grid height: %.2f\n", root.Rect.Height)
	fmt.Printf("Expected: 3 * 50 + 2 * 10 = 170\n")
	fmt.Println()

	item1 := root.Children[0]
	item2 := root.Children[1]
	
	fmt.Printf("Item 1 (spanning rows 0-1):\n")
	fmt.Printf("  Y: %.2f\n", item1.Rect.Y)
	fmt.Printf("  Height: %.2f\n", item1.Rect.Height)
	fmt.Printf("  Bottom: %.2f\n", item1.Rect.Y + item1.Rect.Height)
	fmt.Printf("  Margin: Top=%.2f, Bottom=%.2f\n", item1.Style.Margin.Top, item1.Style.Margin.Bottom)
	fmt.Printf("  Bottom with margin: %.2f\n", item1.Rect.Y + item1.Rect.Height + item1.Style.Margin.Bottom)
	fmt.Println()
	
	fmt.Printf("Item 2 (row 2):\n")
	fmt.Printf("  Y: %.2f\n", item2.Rect.Y)
	fmt.Printf("  Height: %.2f\n", item2.Rect.Height)
	fmt.Printf("  Top with margin: %.2f\n", item2.Rect.Y - item2.Style.Margin.Top)
	fmt.Println()
	
	// Calculate gap between row 1 and row 2
	// Row 1 ends at: row 0 (50) + gap (10) + row 1 (50) = 110
	// Row 2 starts at: 110 + gap (10) = 120
	// But item 1's bottom with margin is at: 5 + 100 + 5 = 110
	// And item 2's top with margin is at: 125 - 5 = 120
	
	row1End := 50.0 + 10.0 + 50.0 // Row 0 + gap + Row 1
	row2Start := row1End + 10.0   // Row 1 end + gap
	
	fmt.Printf("Row boundaries:\n")
	fmt.Printf("  Row 1 end (cell boundary): %.2f\n", row1End)
	fmt.Printf("  Row 2 start (cell boundary): %.2f\n", row2Start)
	fmt.Printf("  Gap: %.2f\n", row2Start - row1End)
	fmt.Println()
	
	// The gap between item 1's bottom (with margin) and item 2's top (with margin)
	item1BottomWithMargin := item1.Rect.Y + item1.Rect.Height + item1.Style.Margin.Bottom
	item2TopWithMargin := item2.Rect.Y - item2.Style.Margin.Top
	gap := item2TopWithMargin - item1BottomWithMargin
	
	fmt.Printf("Visual gap (item margins):\n")
	fmt.Printf("  Item 1 bottom (with margin): %.2f\n", item1BottomWithMargin)
	fmt.Printf("  Item 2 top (with margin): %.2f\n", item2TopWithMargin)
	fmt.Printf("  Visual gap: %.2f\n", gap)
	fmt.Printf("  Expected gap: %.2f (row gap)\n", 10.0)
	
	if gap > 10.0 {
		fmt.Printf("  ❌ BUG: Visual gap is larger than row gap! Extra: %.2f\n", gap - 10.0)
		fmt.Printf("  This suggests the margin is being duplicated or extending into the gap.\n")
	} else if gap < 10.0 {
		fmt.Printf("  ⚠️  Visual gap is smaller than row gap\n")
	} else {
		fmt.Printf("  ✅ Visual gap matches row gap\n")
	}
}

