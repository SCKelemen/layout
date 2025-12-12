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

// =============================================================================
// Phase 2: Immutable Modifications
// =============================================================================

// Clone creates a shallow copy of the node.
// The copy has the same Style, Rect, Text, and other fields, but shares the Children slice.
// Use this when you want to modify node properties without affecting the original.
//
// Example:
//
//	copy := node.Clone()
//	copy.Style.Width = 200  // Original unchanged
func (n *Node) Clone() *Node {
	if n == nil {
		return nil
	}
	copy := *n
	return &copy
}

// CloneDeep creates a deep copy of the entire subtree.
// Both the node and all its descendants are recursively copied.
// Use this when you need a completely independent copy of the tree.
//
// Example:
//
//	independentCopy := root.CloneDeep()
//	independentCopy.Children[0].Style.Width = 100  // Original tree unchanged
func (n *Node) CloneDeep() *Node {
	if n == nil {
		return nil
	}

	// Shallow copy first
	copy := *n

	// Deep copy children
	if len(n.Children) > 0 {
		copy.Children = make([]*Node, len(n.Children))
		for i, child := range n.Children {
			copy.Children[i] = child.CloneDeep()
		}
	}

	return &copy
}

// =============================================================================
// Style Modifications - Return new node with modified style
// =============================================================================

// WithStyle returns a new node with the specified style.
// The original node is unchanged.
//
// Example:
//
//	newNode := node.WithStyle(Style{
//	    Display: DisplayFlex,
//	    Width:   200,
//	})
func (n *Node) WithStyle(style Style) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	copy.Style = style
	return copy
}

// WithPadding returns a new node with uniform padding.
// The original node is unchanged.
//
// Example:
//
//	padded := node.WithPadding(16)
func (n *Node) WithPadding(amount float64) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	copy.Style.Padding = Uniform(amount)
	return copy
}

// WithPaddingCustom returns a new node with custom padding for each side.
// The original node is unchanged.
//
// Example:
//
//	padded := node.WithPaddingCustom(10, 20, 10, 20)  // top, right, bottom, left
func (n *Node) WithPaddingCustom(top, right, bottom, left float64) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	copy.Style.Padding = Spacing{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
	return copy
}

// WithMargin returns a new node with uniform margin.
// The original node is unchanged.
//
// Example:
//
//	margined := node.WithMargin(8)
func (n *Node) WithMargin(amount float64) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	copy.Style.Margin = Uniform(amount)
	return copy
}

// WithMarginCustom returns a new node with custom margin for each side.
// The original node is unchanged.
//
// Example:
//
//	margined := node.WithMarginCustom(5, 10, 5, 10)  // top, right, bottom, left
func (n *Node) WithMarginCustom(top, right, bottom, left float64) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	copy.Style.Margin = Spacing{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
	return copy
}

// WithWidth returns a new node with the specified width.
// The original node is unchanged.
//
// Example:
//
//	wider := node.WithWidth(300)
func (n *Node) WithWidth(width float64) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	copy.Style.Width = width
	return copy
}

// WithHeight returns a new node with the specified height.
// The original node is unchanged.
//
// Example:
//
//	taller := node.WithHeight(200)
func (n *Node) WithHeight(height float64) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	copy.Style.Height = height
	return copy
}

// WithText returns a new node with the specified text content.
// The original node is unchanged.
//
// Example:
//
//	textNode := node.WithText("Hello, World!")
func (n *Node) WithText(text string) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	copy.Text = text
	return copy
}

// WithDisplay returns a new node with the specified display mode.
// The original node is unchanged.
//
// Example:
//
//	flexNode := node.WithDisplay(DisplayFlex)
func (n *Node) WithDisplay(display Display) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	copy.Style.Display = display
	return copy
}

// WithFlexGrow returns a new node with the specified flex-grow value.
// The original node is unchanged.
//
// Example:
//
//	growable := node.WithFlexGrow(1)
func (n *Node) WithFlexGrow(grow float64) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	copy.Style.FlexGrow = grow
	return copy
}

// WithFlexShrink returns a new node with the specified flex-shrink value.
// The original node is unchanged.
//
// Example:
//
//	shrinkable := node.WithFlexShrink(0)
func (n *Node) WithFlexShrink(shrink float64) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	copy.Style.FlexShrink = shrink
	return copy
}

// =============================================================================
// Children Modifications - Return new node with modified children
// =============================================================================

// WithChildren returns a new node with the specified children.
// Uses copy-on-write: creates a new Children slice.
// The original node is unchanged.
//
// Example:
//
//	newParent := parent.WithChildren(child1, child2, child3)
func (n *Node) WithChildren(children ...*Node) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	// Create new slice (copy-on-write)
	copy.Children = make([]*Node, len(children))
	for i, child := range children {
		copy.Children[i] = child
	}
	return copy
}

// AddChild returns a new node with the specified child appended.
// Uses copy-on-write: creates a new Children slice.
// The original node is unchanged.
//
// Example:
//
//	newParent := parent.AddChild(newChild)
func (n *Node) AddChild(child *Node) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	// Create new slice (copy-on-write)
	copy.Children = append([]*Node{}, n.Children...)
	copy.Children = append(copy.Children, child)
	return copy
}

// AddChildren returns a new node with the specified children appended.
// Uses copy-on-write: creates a new Children slice.
// The original node is unchanged.
//
// Example:
//
//	newParent := parent.AddChildren(child1, child2)
func (n *Node) AddChildren(children ...*Node) *Node {
	if n == nil {
		return nil
	}
	copy := n.Clone()
	// Create new slice (copy-on-write)
	copy.Children = append([]*Node{}, n.Children...)
	copy.Children = append(copy.Children, children...)
	return copy
}

// RemoveChildAt returns a new node with the child at the specified index removed.
// Returns the original node if the index is out of bounds.
// Uses copy-on-write: creates a new Children slice.
// The original node is unchanged.
//
// Example:
//
//	newParent := parent.RemoveChildAt(1)  // Remove second child
func (n *Node) RemoveChildAt(index int) *Node {
	if n == nil || index < 0 || index >= len(n.Children) {
		return n
	}
	copy := n.Clone()
	// Create new slice without the removed child
	copy.Children = make([]*Node, 0, len(n.Children)-1)
	copy.Children = append(copy.Children, n.Children[:index]...)
	copy.Children = append(copy.Children, n.Children[index+1:]...)
	return copy
}

// ReplaceChildAt returns a new node with the child at the specified index replaced.
// Returns the original node if the index is out of bounds.
// Uses copy-on-write: creates a new Children slice.
// The original node is unchanged.
//
// Example:
//
//	newParent := parent.ReplaceChildAt(0, newFirstChild)
func (n *Node) ReplaceChildAt(index int, newChild *Node) *Node {
	if n == nil || index < 0 || index >= len(n.Children) {
		return n
	}
	copy := n.Clone()
	// Create new slice (copy-on-write)
	copy.Children = append([]*Node{}, n.Children...)
	copy.Children[index] = newChild
	return copy
}

// InsertChildAt returns a new node with the child inserted at the specified index.
// If index is out of bounds, appends the child.
// Uses copy-on-write: creates a new Children slice.
// The original node is unchanged.
//
// Example:
//
//	newParent := parent.InsertChildAt(1, newChild)  // Insert at position 1
func (n *Node) InsertChildAt(index int, child *Node) *Node {
	if n == nil {
		return nil
	}

	// Clamp index to valid range
	if index < 0 {
		index = 0
	}
	if index > len(n.Children) {
		index = len(n.Children)
	}

	copy := n.Clone()
	// Create new slice with room for inserted child
	copy.Children = make([]*Node, 0, len(n.Children)+1)
	copy.Children = append(copy.Children, n.Children[:index]...)
	copy.Children = append(copy.Children, child)
	copy.Children = append(copy.Children, n.Children[index:]...)
	return copy
}
