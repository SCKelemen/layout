package layout

// node_fluent.go
// Roslyn-style fluent API for layout tree navigation, querying, and transformations
// Provides immutable, composable methods for working with layout trees

// =============================================================================
// Phase 1: Downward Navigation & Querying
// =============================================================================

// Descendants returns all descendant nodes (children, grandchildren, etc.) in depth-first order.
// Does not include the receiver node itself.
//
// Example:
//
//	descendants := root.Descendants()
//	for _, node := range descendants {
//	    fmt.Printf("Node: %v\n", node.Style.Display)
//	}
func (n *Node) Descendants() []*Node {
	if n == nil {
		return nil
	}

	// Pre-allocate with rough capacity estimate (assume average of 3 children per node, depth 3)
	result := make([]*Node, 0, len(n.Children)*3)

	// Recursive helper function
	var collect func(*Node)
	collect = func(node *Node) {
		for _, child := range node.Children {
			result = append(result, child)
			collect(child)
		}
	}

	collect(n)
	return result
}

// DescendantsAndSelf returns all descendant nodes plus the receiver node itself.
//
// Example:
//
//	allNodes := root.DescendantsAndSelf()
//	fmt.Printf("Total nodes in tree: %d\n", len(allNodes))
func (n *Node) DescendantsAndSelf() []*Node {
	if n == nil {
		return nil
	}

	result := make([]*Node, 0, len(n.Children)*3+1)
	result = append(result, n)

	// Recursive helper function
	var collect func(*Node)
	collect = func(node *Node) {
		for _, child := range node.Children {
			result = append(result, child)
			collect(child)
		}
	}

	collect(n)
	return result
}

// FirstChild returns the first child node, or nil if there are no children.
//
// Example:
//
//	if first := root.FirstChild(); first != nil {
//	    fmt.Printf("First child width: %.2f\n", first.Style.Width)
//	}
func (n *Node) FirstChild() *Node {
	if n == nil || len(n.Children) == 0 {
		return nil
	}
	return n.Children[0]
}

// LastChild returns the last child node, or nil if there are no children.
//
// Example:
//
//	if last := root.LastChild(); last != nil {
//	    fmt.Printf("Last child width: %.2f\n", last.Style.Width)
//	}
func (n *Node) LastChild() *Node {
	if n == nil || len(n.Children) == 0 {
		return nil
	}
	return n.Children[len(n.Children)-1]
}

// ChildAt returns the child at the specified index, or nil if the index is out of bounds.
// Negative indices are not supported and will return nil.
//
// Example:
//
//	if child := root.ChildAt(2); child != nil {
//	    fmt.Printf("Third child: %v\n", child.Style.Display)
//	}
func (n *Node) ChildAt(index int) *Node {
	if n == nil || index < 0 || index >= len(n.Children) {
		return nil
	}
	return n.Children[index]
}

// ChildCount returns the number of direct children of this node.
//
// Example:
//
//	count := root.ChildCount()
//	fmt.Printf("Root has %d children\n", count)
func (n *Node) ChildCount() int {
	if n == nil {
		return 0
	}
	return len(n.Children)
}

// =============================================================================
// Querying Methods
// =============================================================================

// Find returns the first node in the tree (depth-first) that matches the predicate,
// or nil if no match is found. Searches descendants only, not the receiver node itself.
//
// Example:
//
//	flexContainer := root.Find(func(n *Node) bool {
//	    return n.Style.Display == DisplayFlex
//	})
func (n *Node) Find(predicate func(*Node) bool) *Node {
	if n == nil || predicate == nil {
		return nil
	}

	// Depth-first search with early termination
	var search func(*Node) *Node
	search = func(node *Node) *Node {
		for _, child := range node.Children {
			if predicate(child) {
				return child
			}
			// Recursive search in child's subtree
			if found := search(child); found != nil {
				return found
			}
		}
		return nil
	}

	return search(n)
}

// FindAll returns all nodes in the tree (depth-first) that match the predicate.
// Returns an empty slice if no matches are found. Searches descendants only.
//
// Example:
//
//	allFlexBoxes := root.FindAll(func(n *Node) bool {
//	    return n.Style.Display == DisplayFlex
//	})
func (n *Node) FindAll(predicate func(*Node) bool) []*Node {
	if n == nil || predicate == nil {
		return nil
	}

	result := make([]*Node, 0, 10)

	// Recursive collection
	var collect func(*Node)
	collect = func(node *Node) {
		for _, child := range node.Children {
			if predicate(child) {
				result = append(result, child)
			}
			collect(child)
		}
	}

	collect(n)
	return result
}

// Where is an alias for FindAll - returns all nodes matching the predicate.
// Provides LINQ-style naming for developers familiar with that pattern.
//
// Example:
//
//	wideNodes := root.Where(func(n *Node) bool {
//	    return n.Rect.Width > 500
//	})
func (n *Node) Where(predicate func(*Node) bool) []*Node {
	return n.FindAll(predicate)
}

// Any returns true if any descendant node matches the predicate.
// Returns false if no matches are found or if the node is nil.
//
// Example:
//
//	hasText := root.Any(func(n *Node) bool {
//	    return n.Text != ""
//	})
func (n *Node) Any(predicate func(*Node) bool) bool {
	if n == nil || predicate == nil {
		return false
	}

	// Early termination on first match
	var search func(*Node) bool
	search = func(node *Node) bool {
		for _, child := range node.Children {
			if predicate(child) {
				return true
			}
			if search(child) {
				return true
			}
		}
		return false
	}

	return search(n)
}

// All returns true if all descendant nodes match the predicate.
// Returns true for nodes with no children (vacuous truth).
//
// Example:
//
//	allVisible := root.All(func(n *Node) bool {
//	    return n.Style.Display != DisplayNone
//	})
func (n *Node) All(predicate func(*Node) bool) bool {
	if n == nil || predicate == nil {
		return true
	}

	// Early termination on first non-match
	var check func(*Node) bool
	check = func(node *Node) bool {
		for _, child := range node.Children {
			if !predicate(child) {
				return false
			}
			if !check(child) {
				return false
			}
		}
		return true
	}

	return check(n)
}

// OfDisplayType returns all descendants with the specified display type.
// This is a convenience method that's more readable than using FindAll with a predicate.
//
// Example:
//
//	allGrids := root.OfDisplayType(DisplayGrid)
//	allFlexboxes := root.OfDisplayType(DisplayFlex)
func (n *Node) OfDisplayType(display Display) []*Node {
	return n.FindAll(func(node *Node) bool {
		return node.Style.Display == display
	})
}
