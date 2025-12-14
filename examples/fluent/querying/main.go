package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

// Example: Querying and Transforming Trees with Fluent API
//
// Demonstrates Find, FindAll, Filter, Transform, Map, and Fold operations.

func main() {
	fmt.Println("=== Querying and Transforming Example ===")

	// Build a complex tree
	tree := layout.VStack(
		layout.HStack(
			layout.Fixed(100, 50).WithText("Button A"),
			layout.Fixed(200, 50).WithText("Input Field"),
			layout.Fixed(100, 50).WithText("Button B"),
		).WithPadding(10),
		layout.HStack(
			layout.VStack(
				layout.Fixed(150, 100).WithText("Chart 1"),
				layout.Fixed(150, 100).WithText("Chart 2"),
			),
			layout.VStack(
				layout.Fixed(150, 100).WithText("Table"),
				layout.Fixed(150, 80).WithText("Stats"),
			),
		).WithPadding(10),
		layout.HStack(
			layout.Fixed(100, 40).WithText("Cancel"),
			layout.Fixed(100, 40).WithText("Submit"),
		).WithPadding(10),
	)

	// Layout the tree
	ctx := layout.NewLayoutContext(800, 600, 16)
	layout.Layout(tree, layout.Loose(800, 600), ctx)

	fmt.Println("=== Finding Nodes ===")

	// Find first node with specific text
	submit := tree.Find(func(n *layout.Node) bool {
		return n.Text == "Submit"
	})
	if submit != nil {
		fmt.Printf("Found 'Submit' button at (%.0f, %.0f)\n",
			submit.Rect.X, submit.Rect.Y)
	}

	// Find all nodes with text
	textNodes := tree.FindAll(func(n *layout.Node) bool {
		return n.Text != ""
	})
	fmt.Printf("Nodes with text: %d\n", len(textNodes))

	// Find all wide nodes
	wideNodes := tree.FindAll(func(n *layout.Node) bool {
		return n.Style.Width.Value >= 150
	})
	fmt.Printf("Wide nodes (>= 150px): %d\n", len(wideNodes))

	// Check if any node is tall
	hasTall := tree.Any(func(n *layout.Node) bool {
		return n.Style.Height.Value > 80
	})
	fmt.Printf("Has tall nodes: %v\n", hasTall)

	// Check if all text nodes are under 200px wide
	allNarrow := tree.FindAll(func(n *layout.Node) bool {
		return n.Text != ""
	})
	allTextNarrow := true
	for _, node := range allNarrow {
		if node.Style.Width.Value >= 200 {
			allTextNarrow = false
			break
		}
	}
	fmt.Printf("All text nodes narrow: %v\n", allTextNarrow)

	fmt.Println("\n=== Collecting Statistics ===")

	// Count total nodes
	totalNodes := tree.Fold(0, func(acc interface{}, n *layout.Node) interface{} {
		return acc.(int) + 1
	}).(int)
	fmt.Printf("Total nodes: %d\n", totalNodes)

	// Sum all widths
	totalWidth := tree.Fold(0.0, func(acc interface{}, n *layout.Node) interface{} {
		return acc.(float64) + n.Style.Width.Value
	}).(float64)
	fmt.Printf("Sum of all widths: %.0f\n", totalWidth)

	// Find max height
	maxHeight := tree.Fold(0.0, func(acc interface{}, n *layout.Node) interface{} {
		current := acc.(float64)
		if n.Style.Height.Value > current {
			return n.Style.Height.Value
		}
		return current
	}).(float64)
	fmt.Printf("Maximum height: %.0f\n", maxHeight)

	// Collect all text content
	allText := tree.Fold([]string{}, func(acc interface{}, n *layout.Node) interface{} {
		list := acc.([]string)
		if n.Text != "" {
			list = append(list, n.Text)
		}
		return list
	}).([]string)
	fmt.Printf("All text: %v\n", allText)

	// Count nodes by depth
	depthCounts := tree.FoldWithContext(
		make(map[int]int),
		func(acc interface{}, n *layout.Node, depth int) interface{} {
			m := acc.(map[int]int)
			m[depth]++
			return m
		},
	).(map[int]int)
	fmt.Printf("Nodes per depth: %v\n", depthCounts)

	fmt.Println("\n=== Transforming Trees ===")

	// Double the width of buttons
	widerButtons := tree.Transform(
		func(n *layout.Node) bool {
			return n.Style.Width.Value == 100 && n.Text != ""
		},
		func(n *layout.Node) *layout.Node {
			return n.WithWidth(200)
		},
	)
	layout.Layout(widerButtons, layout.Loose(800, 600), ctx)

	buttonsAfter := widerButtons.FindAll(func(n *layout.Node) bool {
		return n.Text != "" && n.Style.Width.Value == 200
	})
	fmt.Printf("Buttons widened: %d now have width 200\n", len(buttonsAfter))

	// Scale entire tree by 1.5x
	scaled := tree.Map(func(n *layout.Node) *layout.Node {
		return n.
			WithWidth(n.Style.Width.Value * 1.5).
			WithHeight(n.Style.Height.Value * 1.5)
	})
	layout.Layout(scaled, layout.Loose(1200, 900), ctx)
	fmt.Printf("Tree scaled by 1.5x\n")

	// Add padding to all containers
	_ = tree.Transform(
		func(n *layout.Node) bool {
			return len(n.Children) > 0
		},
		func(n *layout.Node) *layout.Node {
			currentPadding := n.Style.Padding.Top.Value
			return n.WithPadding(currentPadding + 5)
		},
	)
	fmt.Printf("Added padding to all containers\n")

	fmt.Println("\n=== Filtering ===")

	// Keep only nodes with text
	textOnly := tree.FilterDeep(func(n *layout.Node) bool {
		return n.Text != ""
	})
	textOnlyCount := len(textOnly.DescendantsAndSelf())
	fmt.Printf("Filtered to text-only nodes: %d nodes\n", textOnlyCount)

	// Keep only wide elements (shallow filter)
	wideOnly := tree.Filter(func(n *layout.Node) bool {
		return len(n.Children) > 0 || n.Style.Width.Value >= 150
	})
	wideCount := len(wideOnly.Children)
	fmt.Printf("Filtered to wide elements: %d immediate children\n", wideCount)

	fmt.Println("\n=== Original Unchanged ===")

	// Verify original tree is unchanged
	originalTextNodes := tree.FindAll(func(n *layout.Node) bool {
		return n.Text != ""
	})
	fmt.Printf("Original still has %d text nodes\n", len(originalTextNodes))

	originalSum := tree.Fold(0.0, func(acc interface{}, n *layout.Node) interface{} {
		return acc.(float64) + n.Style.Width.Value
	}).(float64)
	fmt.Printf("Original sum of widths: %.0f (unchanged)\n", originalSum)
}
