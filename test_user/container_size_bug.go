// +build ignore

package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Test: Container size might not include first item's top margin
	fmt.Println("=== Container Size Bug Test ===")
	
	// Create items with different margins to make the bug more obvious
	items := []*layout.Node{
		{
			Style: layout.Style{
				Height: 20,
				Margin: layout.Spacing{Top: 20, Right: 10, Bottom: 10, Left: 10}, // Large top margin
			},
		},
		{
			Style: layout.Style{
				Height: 20,
				Margin: layout.Uniform(10),
			},
		},
		{
			Style: layout.Style{
				Height: 20,
				Margin: layout.Uniform(10),
			},
		},
	}

	// Create VStack
	root := layout.VStack(items...)
	root.Style.Width = 200

	constraints := layout.Loose(200, layout.Unbounded)
	layout.Layout(root, constraints)

	fmt.Printf("Root height: %.2f\n", root.Rect.Height)
	
	// Calculate expected height
	expectedHeight := 0.0
	expectedHeight += items[0].Style.Margin.Top // First item top margin
	expectedHeight += items[0].Rect.Height      // First item height
	for i := 1; i < len(items); i++ {
		expectedHeight += items[i-1].Style.Margin.Bottom // Previous item bottom margin
		expectedHeight += items[i].Style.Margin.Top      // Current item top margin
		expectedHeight += items[i].Rect.Height           // Current item height
	}
	expectedHeight += items[len(items)-1].Style.Margin.Bottom // Last item bottom margin
	
	fmt.Printf("Expected height: %.2f\n", expectedHeight)
	fmt.Printf("  = First top margin (%.2f) + items + gaps + last bottom margin (%.2f)\n", 
		items[0].Style.Margin.Top, items[len(items)-1].Style.Margin.Bottom)
	
	if root.Rect.Height < expectedHeight {
		fmt.Printf("  ❌ BUG: Container height is too small! Missing %.2f\n", expectedHeight - root.Rect.Height)
	} else if root.Rect.Height > expectedHeight {
		fmt.Printf("  ⚠️  Container height is larger than expected (might be okay)\n")
	} else {
		fmt.Printf("  ✅ Container height is correct\n")
	}
	fmt.Println()

	// Check if second item overlaps first
	firstItem := root.Children[0]
	secondItem := root.Children[1]
	
	firstItemBottomWithMargin := firstItem.Rect.Y + firstItem.Rect.Height + firstItem.Style.Margin.Bottom
	secondItemTopWithMargin := secondItem.Rect.Y - secondItem.Style.Margin.Top
	
	fmt.Printf("Overlap check:\n")
	fmt.Printf("  First item bottom (with margin): %.2f\n", firstItemBottomWithMargin)
	fmt.Printf("  Second item top (with margin): %.2f\n", secondItemTopWithMargin)
	
	if secondItemTopWithMargin < firstItemBottomWithMargin {
		overlap := firstItemBottomWithMargin - secondItemTopWithMargin
		fmt.Printf("  ❌ OVERLAP DETECTED! Overlap amount: %.2f\n", overlap)
		fmt.Printf("  Second item should start at: %.2f\n", firstItemBottomWithMargin + secondItem.Style.Margin.Top)
		fmt.Printf("  Second item actually starts at: %.2f\n", secondItem.Rect.Y)
	} else {
		fmt.Printf("  ✅ No overlap\n")
	}
	
	// Visual representation
	fmt.Println("\nVisual representation:")
	fmt.Printf("0.00 ──────────────────────── (Container top)\n")
	fmt.Printf("%.2f ──────────────────────── (First item top margin)\n", firstItem.Style.Margin.Top)
	fmt.Printf("%.2f ──────────────────────── (First item start)\n", firstItem.Rect.Y)
	fmt.Printf("%.2f ──────────────────────── (First item end)\n", firstItem.Rect.Y + firstItem.Rect.Height)
	fmt.Printf("%.2f ──────────────────────── (First item + bottom margin)\n", firstItemBottomWithMargin)
	fmt.Printf("%.2f ──────────────────────── (Second item should start here)\n", firstItemBottomWithMargin + secondItem.Style.Margin.Top)
	fmt.Printf("%.2f ──────────────────────── (Second item actual start)\n", secondItem.Rect.Y)
	if secondItem.Rect.Y < firstItemBottomWithMargin {
		fmt.Printf("  ❌ OVERLAP!\n")
	}
	fmt.Printf("%.2f ──────────────────────── (Container bottom)\n", root.Rect.Height)
}

