package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

// Example: Classic vs Fluent API Comparison
//
// Demonstrates that both APIs produce identical results
// and can be mixed together seamlessly.

func main() {
	fmt.Println("=== Classic vs Fluent API Comparison ===")

	// === Example 1: Simple Card ===
	fmt.Println("Example 1: Simple Card")

	// Classic API
	classicCard := &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionColumn,
			Width:         layout.Px(300),
			Padding:       layout.Uniform(layout.Px(16)),
			Margin:        layout.Uniform(layout.Px(8)),
		},
		Children: []*layout.Node{
			{
				Style: layout.Style{Height: layout.Px(40)},
				Text:  "Title",
			},
			{
				Style: layout.Style{FlexGrow: 1},
				Text:  "Body",
			},
			{
				Style: layout.Style{Height: layout.Px(30)},
				Text:  "Footer",
			},
		},
	}

	// Fluent API
	fluentCard := (&layout.Node{}).
		WithStyle(layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionColumn,
		}).
		WithWidth(300).
		WithPadding(16).
		WithMargin(8).
		AddChildren(
			(&layout.Node{}).WithHeight(40).WithText("Title"),
			(&layout.Node{}).WithFlexGrow(1).WithText("Body"),
			(&layout.Node{}).WithHeight(30).WithText("Footer"),
		)

	// Layout both
	constraints := layout.Loose(400, 600)
	ctx := layout.NewLayoutContext(400, 600, 16)
	layout.Layout(classicCard, constraints, ctx)
	layout.Layout(fluentCard, constraints, ctx)

	// Compare results
	fmt.Printf("Classic card rect: %.0fx%.0f at (%.0f, %.0f)\n",
		classicCard.Rect.Width, classicCard.Rect.Height,
		classicCard.Rect.X, classicCard.Rect.Y)

	fmt.Printf("Fluent card rect:  %.0fx%.0f at (%.0f, %.0f)\n",
		fluentCard.Rect.Width, fluentCard.Rect.Height,
		fluentCard.Rect.X, fluentCard.Rect.Y)

	fmt.Printf("Rects identical: %v\n\n", classicCard.Rect == fluentCard.Rect)

	// === Example 2: Using Helper Functions ===
	fmt.Println("Example 2: Using Helper Functions")

	// Classic with helpers
	classicStack := layout.HStack(
		layout.Fixed(100, 50),
		layout.Fixed(200, 50),
		layout.Fixed(150, 50),
	)
	layout.Padding(classicStack, 10)
	layout.Margin(classicStack, 5)

	// Fluent with helpers
	fluentStack := layout.HStack(
		layout.Fixed(100, 50),
		layout.Fixed(200, 50),
		layout.Fixed(150, 50),
	).WithPadding(10).WithMargin(5)

	// Layout both
	layout.Layout(classicStack, constraints, ctx)
	layout.Layout(fluentStack, constraints, ctx)

	fmt.Printf("Classic stack rect: %.0fx%.0f\n",
		classicStack.Rect.Width, classicStack.Rect.Height)
	fmt.Printf("Fluent stack rect:  %.0fx%.0f\n",
		fluentStack.Rect.Width, fluentStack.Rect.Height)
	fmt.Printf("Rects identical: %v\n\n", classicStack.Rect == fluentStack.Rect)

	// === Example 3: Mixing Both Styles ===
	fmt.Println("Example 3: Mixing Both Styles")

	// Start with classic
	mixed := layout.VStack(
		layout.Fixed(300, 50),
		layout.Fixed(300, 50),
	)

	// Add fluent operations
	mixed = mixed.
		WithPadding(12).
		AddChild(layout.Fixed(300, 50))

	// Use classic helper
	layout.Margin(mixed, 6)

	// Back to fluent
	mixed = mixed.WithWidth(350)

	layout.Layout(mixed, constraints, ctx)

	fmt.Printf("Mixed API result: %.0fx%.0f with %d children\n",
		mixed.Rect.Width, mixed.Rect.Height, len(mixed.Children))
	fmt.Printf("Padding: %.0f, Margin: %.0f\n\n",
		mixed.Style.Padding.Top.Value, mixed.Style.Margin.Top.Value)

	// === Example 4: Building the Same Tree Two Ways ===
	fmt.Println("Example 4: Complex Tree Two Ways")

	// Classic approach
	classicTree := &layout.Node{
		Style: layout.Style{
			Display:        layout.DisplayFlex,
			FlexDirection:  layout.FlexDirectionColumn,
			JustifyContent: layout.JustifyContentSpaceBetween,
			Width:          layout.Px(400),
			Height:         layout.Px(300),
		},
		Children: []*layout.Node{
			{
				Style: layout.Style{
					Display:       layout.DisplayFlex,
					FlexDirection: layout.FlexDirectionRow,
					Height:        layout.Px(50),
				},
				Children: []*layout.Node{
					{Style: layout.Style{Width: layout.Px(100)}, Text: "Logo"},
					{Style: layout.Style{FlexGrow: 1}},
					{Style: layout.Style{Width: layout.Px(100)}, Text: "Menu"},
				},
			},
			{
				Style: layout.Style{
					FlexGrow: 1,
				},
				Text: "Content",
			},
			{
				Style: layout.Style{
					Height: layout.Px(50),
				},
				Text: "Footer",
			},
		},
	}

	// Fluent approach
	fluentTree := (&layout.Node{}).
		WithStyle(layout.Style{
			Display:        layout.DisplayFlex,
			FlexDirection:  layout.FlexDirectionColumn,
			JustifyContent: layout.JustifyContentSpaceBetween,
		}).
		WithWidth(400).
		WithHeight(300).
		AddChildren(
			layout.HStack(
				(&layout.Node{}).WithWidth(100).WithText("Logo"),
				(&layout.Node{}).WithFlexGrow(1),
				(&layout.Node{}).WithWidth(100).WithText("Menu"),
			).WithHeight(50),
			(&layout.Node{}).WithFlexGrow(1).WithText("Content"),
			(&layout.Node{}).WithHeight(50).WithText("Footer"),
		)

	// Layout both
	layout.Layout(classicTree, constraints, ctx)
	layout.Layout(fluentTree, constraints, ctx)

	// Compare all descendants
	classicDesc := classicTree.DescendantsAndSelf()
	fluentDesc := fluentTree.DescendantsAndSelf()

	fmt.Printf("Classic tree nodes: %d\n", len(classicDesc))
	fmt.Printf("Fluent tree nodes:  %d\n", len(fluentDesc))

	allMatch := true
	for i := range classicDesc {
		if classicDesc[i].Rect != fluentDesc[i].Rect {
			allMatch = false
			break
		}
	}
	fmt.Printf("All node rects match: %v\n\n", allMatch)

	// === Example 5: Querying Works the Same ===
	fmt.Println("Example 5: Querying Both Trees")

	// Query classic tree
	classicText := classicTree.FindAll(func(n *layout.Node) bool {
		return n.Text != ""
	})

	// Query fluent tree
	fluentText := fluentTree.FindAll(func(n *layout.Node) bool {
		return n.Text != ""
	})

	fmt.Printf("Classic tree text nodes: %d\n", len(classicText))
	fmt.Printf("Fluent tree text nodes:  %d\n", len(fluentText))

	// Both trees can use transformations
	classicScaled := classicTree.Map(func(n *layout.Node) *layout.Node {
		return n.WithWidth(n.Style.Width.Value * 1.2)
	})

	fluentScaled := fluentTree.Map(func(n *layout.Node) *layout.Node {
		return n.WithWidth(n.Style.Width.Value * 1.2)
	})

	layout.Layout(classicScaled, constraints, ctx)
	layout.Layout(fluentScaled, constraints, ctx)

	fmt.Printf("Both trees scaled successfully\n")
	fmt.Printf("Classic scaled width: %.0f\n", classicScaled.Style.Width.Value)
	fmt.Printf("Fluent scaled width:  %.0f\n\n", fluentScaled.Style.Width.Value)

	// === Summary ===
	fmt.Println("=== Summary ===")
	fmt.Println("✓ Classic and fluent APIs produce identical layouts")
	fmt.Println("✓ Both can use helper functions (HStack, VStack, etc.)")
	fmt.Println("✓ Both can be mixed in the same codebase")
	fmt.Println("✓ Fluent methods available on any *Node")
	fmt.Println("✓ Choose style based on preference and use case")
}
