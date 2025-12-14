package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

// Example: Parent Navigation with NodeContext
//
// Demonstrates using NodeContext for upward tree traversal
// to find ancestors and navigate to parent nodes.

func main() {
	fmt.Println("=== NodeContext Example ===")

	// Build a nested tree structure
	tree := layout.VStack(
		layout.HStack(
			layout.VStack(
				layout.Fixed(80, 30).WithText("Item 1-1"),
				layout.Fixed(80, 30).WithText("Item 1-2"),
				layout.Fixed(80, 30).WithText("Target"), // This is our target
			).WithPadding(5),
			layout.VStack(
				layout.Fixed(80, 30).WithText("Item 2-1"),
				layout.Fixed(80, 30).WithText("Item 2-2"),
			).WithPadding(5),
		).WithPadding(10),
		layout.HStack(
			layout.Fixed(100, 40).WithText("Button 1"),
			layout.Fixed(100, 40).WithText("Button 2"),
		).WithPadding(10),
	).WithPadding(20)

	// Layout the tree
	layoutCtx := layout.NewLayoutContext(600, 400, 16)
	layout.Layout(tree, layout.Loose(600, 400), layoutCtx)

	fmt.Println("=== Creating Context ===")

	// Wrap root in context for parent tracking
	ctx := layout.NewContext(tree)
	fmt.Printf("Context created for root\n")
	fmt.Printf("Root is at depth: %d\n", ctx.Depth())
	fmt.Printf("Root has %d children\n", len(ctx.Children()))

	fmt.Println("\n=== Finding Nodes Downward ===")

	// Find the target node
	targetCtx := ctx.FindDown(func(n *layout.Node) bool {
		return n.Text == "Target"
	})

	if targetCtx == nil {
		fmt.Println("Target not found!")
		return
	}

	fmt.Printf("Found 'Target' at depth %d\n", targetCtx.Depth())
	fmt.Printf("Target rect: (%.0f, %.0f) size %.0fx%.0f\n",
		targetCtx.Node.Rect.X, targetCtx.Node.Rect.Y,
		targetCtx.Node.Rect.Width, targetCtx.Node.Rect.Height)

	fmt.Println("\n=== Navigating to Parent ===")

	// Get immediate parent
	parentCtx := targetCtx.Parent()
	if parentCtx != nil {
		fmt.Printf("Parent is a VStack with %d children\n", len(parentCtx.Node.Children))
		fmt.Printf("Parent has padding: %.0f\n", parentCtx.Node.Style.Padding.Top.Value)
	}

	// Get grandparent
	grandparentCtx := parentCtx.Parent()
	if grandparentCtx != nil {
		fmt.Printf("Grandparent is an HStack with %d children\n", len(grandparentCtx.Node.Children))
		fmt.Printf("Grandparent has padding: %.0f\n", grandparentCtx.Node.Style.Padding.Top)
	}

	fmt.Println("\n=== Getting All Ancestors ===")

	// Get all ancestors
	ancestors := targetCtx.Ancestors()
	fmt.Printf("Target has %d ancestors:\n", len(ancestors))
	for i, ancestor := range ancestors {
		childCount := len(ancestor.Node.Children)
		fmt.Printf("  %d. Depth %d: %d children, padding %.0f\n",
			i+1, ancestor.Depth(), childCount, ancestor.Node.Style.Padding.Top)
	}

	// Get path from target to root
	pathToRoot := targetCtx.AncestorsAndSelf()
	fmt.Printf("\nPath from target to root: %d nodes\n", len(pathToRoot))

	fmt.Println("\n=== Finding Containing Flex Container ===")

	// Find the nearest flex container ancestor
	flexCtx := targetCtx.FindUp(func(n *layout.Node) bool {
		return n.Style.Display == layout.DisplayFlex
	})

	if flexCtx != nil {
		fmt.Printf("Found containing flex container at depth %d\n", flexCtx.Depth())
		fmt.Printf("Flex container has %d children\n", len(flexCtx.Node.Children))
		fmt.Printf("Flex direction: %v\n", flexCtx.Node.Style.FlexDirection)
	}

	fmt.Println("\n=== Getting Siblings ===")

	// Get siblings of target
	siblings := targetCtx.Siblings()
	fmt.Printf("Target has %d siblings:\n", len(siblings))
	for i, sibling := range siblings {
		fmt.Printf("  %d. %s\n", i+1, sibling.Node.Text)
	}

	fmt.Println("\n=== Finding All Flex Containers ===")

	// Find all flex containers in tree
	allFlexCtx := ctx.FindDownAll(func(n *layout.Node) bool {
		return n.Style.Display == layout.DisplayFlex
	})

	fmt.Printf("Found %d flex containers:\n", len(allFlexCtx))
	for i, flexContainer := range allFlexCtx {
		fmt.Printf("  %d. Depth %d, %d children, direction: %v\n",
			i+1,
			flexContainer.Depth(),
			len(flexContainer.Node.Children),
			flexContainer.Node.Style.FlexDirection)
	}

	fmt.Println("\n=== Utility Methods ===")

	// Check various properties
	fmt.Printf("Target is root: %v\n", targetCtx.IsRoot())
	fmt.Printf("Target has parent: %v\n", targetCtx.HasParent())
	fmt.Printf("Target has children: %v\n", targetCtx.HasChildren())

	fmt.Printf("Root is root: %v\n", ctx.IsRoot())
	fmt.Printf("Root has parent: %v\n", ctx.HasParent())
	fmt.Printf("Root has children: %v\n", ctx.HasChildren())

	fmt.Println("\n=== Practical Use Case: Modifying Based on Context ===")

	// Find all leaf nodes (nodes with no children) that are inside VStacks
	leafsInVStack := []*layout.Node{}

	for _, descendantCtx := range ctx.FindDownAll(func(n *layout.Node) bool {
		return len(n.Children) == 0 && n.Text != ""
	}) {
		// Check if any ancestor is a VStack
		if descendantCtx.FindUp(func(n *layout.Node) bool {
			return n.Style.Display == layout.DisplayFlex &&
				n.Style.FlexDirection == layout.FlexDirectionColumn
		}) != nil {
			leafsInVStack = append(leafsInVStack, descendantCtx.Node)
		}
	}

	fmt.Printf("Found %d leaf nodes inside VStacks:\n", len(leafsInVStack))
	for i, leaf := range leafsInVStack {
		fmt.Printf("  %d. %s\n", i+1, leaf.Text)
	}

	fmt.Println("\n=== Combining Context with Transformations ===")

	// Using context to understand structure, then transform
	// Find all nodes that have exactly 2 siblings and modify them

	_ = tree.Transform(
		func(n *layout.Node) bool {
			// Create temporary context to check siblings
			tempCtx := layout.NewContext(tree)
			nodeCtx := tempCtx.FindDown(func(candidate *layout.Node) bool {
				return candidate == n
			})
			if nodeCtx != nil {
				siblings := nodeCtx.Siblings()
				return len(siblings) == 2
			}
			return false
		},
		func(n *layout.Node) *layout.Node {
			// Add some distinction
			return n.WithPadding(n.Style.Padding.Top.Value + 2)
		},
	)

	fmt.Printf("Modified nodes with exactly 2 siblings\n")
	fmt.Printf("Original tree unchanged\n")

	// Unwrap context to get underlying node
	rootNode := ctx.Unwrap()
	fmt.Printf("\nRoot node children: %d\n", len(rootNode.Children))
}
