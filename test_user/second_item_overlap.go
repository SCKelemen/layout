// +build ignore

package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Reproduce the specific bug: second item overlaps first, but rest are correct
	fmt.Println("=== Second Item Overlap Bug ===")
	
	items := []*layout.Node{
		{
			Style: layout.Style{
				Height: Px(30),
				Margin: layout.Uniform(15),
			},
		},
		{
			Style: layout.Style{
				Height: Px(25),
				Margin: layout.Uniform(15),
			},
		},
		{
			Style: layout.Style{
				Height: Px(20),
				Margin: layout.Uniform(15),
			},
		},
		{
			Style: layout.Style{
				Height: Px(20),
				Margin: layout.Uniform(15),
			},
		},
	}

	root := layout.VStack(items...)
	root.Style.Width = 200

	constraints := layout.Loose(200, layout.Unbounded)
	layout.Layout(root, constraints)

	fmt.Printf("Root height: %.2f\n", root.Rect.Height)
	fmt.Println()

	// Detailed analysis of each item
	for i, item := range root.Children {
		fmt.Printf("Item %d:\n", i+1)
		fmt.Printf("  Y: %.2f\n", item.Rect.Y)
		fmt.Printf("  Height: %.2f\n", item.Rect.Height)
		fmt.Printf("  Bottom: %.2f\n", item.Rect.Y + item.Rect.Height)
		fmt.Printf("  Margin: Top=%.2f, Bottom=%.2f\n", item.Style.Margin.Top, item.Style.Margin.Bottom)
		fmt.Printf("  Bottom with margin: %.2f\n", item.Rect.Y + item.Rect.Height + item.Style.Margin.Bottom)
		
		if i > 0 {
			prev := root.Children[i-1]
			prevBottomWithMargin := prev.Rect.Y + prev.Rect.Height + prev.Style.Margin.Bottom
			gap := item.Rect.Y - (prev.Rect.Y + prev.Rect.Height)
			expectedGap := prev.Style.Margin.Bottom + item.Style.Margin.Top
			
			fmt.Printf("  Gap from prev bottom: %.2f\n", gap)
			fmt.Printf("  Expected gap: %.2f\n", expectedGap)
			fmt.Printf("  Prev bottom with margin: %.2f\n", prevBottomWithMargin)
			fmt.Printf("  Should start at: %.2f\n", prevBottomWithMargin + item.Style.Margin.Top)
			
			if gap < expectedGap {
				overlap := expectedGap - gap
				fmt.Printf("  ❌ OVERLAP: %.2f pixels\n", overlap)
			} else if gap == expectedGap {
				fmt.Printf("  ✅ Correct spacing\n")
			} else {
				fmt.Printf("  ⚠️  Larger gap than expected\n")
			}
		} else {
			// First item
			if item.Rect.Y != item.Style.Margin.Top {
				fmt.Printf("  ❌ First item Y should be %.2f (top margin), got %.2f\n", 
					item.Style.Margin.Top, item.Rect.Y)
			} else {
				fmt.Printf("  ✅ First item positioned correctly\n")
			}
		}
		fmt.Println()
	}
	
	// Check specifically for second item overlap
	if len(root.Children) >= 2 {
		first := root.Children[0]
		second := root.Children[1]
		
		firstBottomWithMargin := first.Rect.Y + first.Rect.Height + first.Style.Margin.Bottom
		secondTop := second.Rect.Y
		expectedSecondTop := firstBottomWithMargin + second.Style.Margin.Top
		
		fmt.Println("=== Second Item Overlap Analysis ===")
		fmt.Printf("First item bottom (with margin): %.2f\n", firstBottomWithMargin)
		fmt.Printf("Second item should start at: %.2f\n", expectedSecondTop)
		fmt.Printf("Second item actually starts at: %.2f\n", secondTop)
		
		if secondTop < expectedSecondTop {
			overlap := expectedSecondTop - secondTop
			fmt.Printf("❌ OVERLAP DETECTED: %.2f pixels\n", overlap)
			fmt.Printf("   Second item overlaps first by %.2f pixels\n", overlap)
		} else {
			fmt.Printf("✅ No overlap\n")
		}
	}
}

