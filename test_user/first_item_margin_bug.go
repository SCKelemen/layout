//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Test: First item margin might not be respected correctly
	fmt.Println("=== First Item Margin Bug Test ===")

	// Create multiple text nodes with margins
	items := []*layout.Node{
		{
			Style: layout.Style{
				Height: Px(20),
				Margin: layout.Uniform(10),
			},
		},
		{
			Style: layout.Style{
				Height: Px(20),
				Margin: layout.Uniform(10),
			},
		},
		{
			Style: layout.Style{
				Height: Px(20),
				Margin: layout.Uniform(10),
			},
		},
		{
			Style: layout.Style{
				Height: Px(20),
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
	fmt.Printf("Expected: 140 (4 * 20 + 8 * 10 = 160, but first/last margins)\n")
	fmt.Println()

	// Check if first item's top margin is respected
	firstItem := root.Children[0]
	fmt.Printf("First item:\n")
	fmt.Printf("  Y position: %.2f\n", firstItem.Rect.Y)
	fmt.Printf("  Top margin: %.2f\n", firstItem.Style.Margin.Top)
	fmt.Printf("  Expected Y: %.2f (should be at top margin)\n", firstItem.Style.Margin.Top)

	if firstItem.Rect.Y != firstItem.Style.Margin.Top {
		fmt.Printf("  ❌ BUG: First item Y position is wrong! Should be %.2f, got %.2f\n",
			firstItem.Style.Margin.Top, firstItem.Rect.Y)
	} else {
		fmt.Printf("  ✅ First item position is correct\n")
	}
	fmt.Println()

	// Check gaps between all items
	for i := 1; i < len(root.Children); i++ {
		prev := root.Children[i-1]
		curr := root.Children[i]

		prevBottom := prev.Rect.Y + prev.Rect.Height
		currTop := curr.Rect.Y
		gap := currTop - prevBottom
		expectedGap := prev.Style.Margin.Bottom + curr.Style.Margin.Top

		fmt.Printf("Gap between item %d and %d:\n", i, i+1)
		fmt.Printf("  Previous bottom: %.2f\n", prevBottom)
		fmt.Printf("  Current top: %.2f\n", currTop)
		fmt.Printf("  Gap: %.2f\n", gap)
		fmt.Printf("  Expected gap: %.2f (prev margin bottom + curr margin top)\n", expectedGap)

		if gap < expectedGap {
			fmt.Printf("  ❌ OVERLAP! Gap is too small (%.2f < %.2f)\n", gap, expectedGap)
			fmt.Printf("  Overlap amount: %.2f\n", expectedGap-gap)
		} else if gap == expectedGap {
			fmt.Printf("  ✅ Gap is correct\n")
		} else {
			fmt.Printf("  ⚠️  Gap is larger than expected\n")
		}
		fmt.Println()
	}

	// Visual representation focusing on first item
	fmt.Println("Visual representation (first item detail):")
	fmt.Printf("0.00 ──────────────────────── (Container top)\n")
	firstItem = root.Children[0]
	fmt.Printf("%.2f ──────────────────────── (First item should start here, margin top)\n", firstItem.Style.Margin.Top)
	fmt.Printf("%.2f ──────────────────────── (First item actual start)\n", firstItem.Rect.Y)
	if firstItem.Rect.Y != firstItem.Style.Margin.Top {
		fmt.Printf("  ❌ MISMATCH: First item is not respecting top margin!\n")
	}
	fmt.Printf("%.2f ──────────────────────── (First item end)\n", firstItem.Rect.Y+firstItem.Rect.Height)
	fmt.Printf("%.2f ──────────────────────── (First item + margin bottom)\n", firstItem.Rect.Y+firstItem.Rect.Height+firstItem.Style.Margin.Bottom)

	if len(root.Children) > 1 {
		secondItem := root.Children[1]
		fmt.Printf("%.2f ──────────────────────── (Second item should start here)\n",
			firstItem.Rect.Y+firstItem.Rect.Height+firstItem.Style.Margin.Bottom+secondItem.Style.Margin.Top)
		fmt.Printf("%.2f ──────────────────────── (Second item actual start)\n", secondItem.Rect.Y)
		if secondItem.Rect.Y < firstItem.Rect.Y+firstItem.Rect.Height+firstItem.Style.Margin.Bottom {
			fmt.Printf("  ❌ OVERLAP: Second item overlaps first item!\n")
		}
	}
}
