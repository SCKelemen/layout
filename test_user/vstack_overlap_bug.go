//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Reproduce the overlap bug: second row overlapping first row
	fmt.Println("=== VStack Overlap Bug Test ===")

	// Create text nodes with margins
	text1 := &layout.Node{
		Style: layout.Style{
			Height: Px(20),
			Margin: layout.Uniform(10),
		},
	}

	text2 := &layout.Node{
		Style: layout.Style{
			Height: Px(20),
			Margin: layout.Uniform(10),
		},
	}

	// Create VStack
	root := layout.VStack(text1, text2)
	root.Style.Width = 200

	constraints := layout.Loose(200, layout.Unbounded)
	layout.Layout(root, constraints)

	fmt.Printf("Root height: %.2f\n", root.Rect.Height)
	fmt.Printf("Expected: 60 (2 * 20 + 4 * 10 = 80, but first/last margins might be handled differently)\n")
	fmt.Println()

	for i, child := range root.Children {
		fmt.Printf("Text %d:\n", i+1)
		fmt.Printf("  Y position: %.2f\n", child.Rect.Y)
		fmt.Printf("  Height: %.2f\n", child.Rect.Height)
		fmt.Printf("  Bottom edge: %.2f\n", child.Rect.Y+child.Rect.Height)
		fmt.Printf("  Margin: Top=%.2f, Bottom=%.2f\n", child.Style.Margin.Top, child.Style.Margin.Bottom)

		if i > 0 {
			prev := root.Children[i-1]
			prevBottom := prev.Rect.Y + prev.Rect.Height
			gap := child.Rect.Y - prevBottom
			expectedGap := prev.Style.Margin.Bottom + child.Style.Margin.Top

			fmt.Printf("  Gap from previous bottom: %.2f\n", gap)
			fmt.Printf("  Expected gap: %.2f (prev margin bottom + this margin top)\n", expectedGap)

			if gap < expectedGap {
				fmt.Printf("  ❌ OVERLAP DETECTED! Gap is too small (%.2f < %.2f)\n", gap, expectedGap)
				fmt.Printf("  Previous item ends at: %.2f\n", prevBottom)
				fmt.Printf("  This item starts at: %.2f\n", child.Rect.Y)
				fmt.Printf("  Overlap amount: %.2f\n", expectedGap-gap)
			} else if gap == expectedGap {
				fmt.Printf("  ✅ Gap is correct\n")
			} else {
				fmt.Printf("  ⚠️  Gap is larger than expected (might be okay)\n")
			}
		}
		fmt.Println()
	}

	// Visual representation
	fmt.Println("Visual representation:")
	fmt.Println("0.00 ────────────────────────")
	for i, child := range root.Children {
		fmt.Printf("%.2f ──────────────────────── (Text %d top, margin top)\n", child.Rect.Y-child.Style.Margin.Top, i+1)
		fmt.Printf("%.2f ──────────────────────── (Text %d start)\n", child.Rect.Y, i+1)
		fmt.Printf("%.2f ──────────────────────── (Text %d end)\n", child.Rect.Y+child.Rect.Height, i+1)
		fmt.Printf("%.2f ──────────────────────── (Text %d bottom, margin bottom)\n", child.Rect.Y+child.Rect.Height+child.Style.Margin.Bottom, i+1)
	}
	fmt.Printf("%.2f ──────────────────────── (Root bottom)\n", root.Rect.Height)
}
