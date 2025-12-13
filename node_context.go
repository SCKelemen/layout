package layout

// node_context.go
// Context wrapper for Node that provides parent tracking and upward navigation
// Enables ancestor queries without modifying the Node structure

// NodeContext wraps a Node with parent tracking for upward navigation.
// This provides a way to traverse up the tree without adding parent pointers
// to the Node structure, preserving immutability and avoiding circular references.
//
// Contexts are created on-demand and only allocate memory for accessed paths,
// making them efficient for large trees where only partial navigation is needed.
type NodeContext struct {
	Node   *Node        // The wrapped node
	parent *NodeContext // Parent context (nil for root)
	depth  int          // Distance from root (root = 0)
}

// NewContext creates a new context wrapping the root node.
// This is the entry point for context-based navigation.
//
// Example:
//
//	ctx := layout.NewContext(root)
//	// Now you can navigate up and down the tree
func NewContext(root *Node) *NodeContext {
	if root == nil {
		return nil
	}
	return &NodeContext{
		Node:   root,
		parent: nil,
		depth:  0,
	}
}

// =============================================================================
// Upward Navigation
// =============================================================================

// Parent returns the parent context, or nil if this is the root.
//
// Example:
//
//	parentCtx := ctx.Parent()
//	if parentCtx != nil {
//	    fmt.Printf("Parent display: %v\n", parentCtx.Node.Style.Display)
//	}
func (ctx *NodeContext) Parent() *NodeContext {
	if ctx == nil {
		return nil
	}
	return ctx.parent
}

// Ancestors returns all ancestor contexts from parent to root.
// The slice is ordered from nearest (parent) to furthest (root).
// Returns empty slice if this is the root.
//
// Example:
//
//	ancestors := ctx.Ancestors()
//	for _, ancestor := range ancestors {
//	    fmt.Printf("Ancestor at depth %d\n", ancestor.Depth())
//	}
func (ctx *NodeContext) Ancestors() []*NodeContext {
	if ctx == nil || ctx.parent == nil {
		return nil
	}

	// Count ancestors first
	count := 0
	current := ctx.parent
	for current != nil {
		count++
		current = current.parent
	}

	// Collect ancestors
	result := make([]*NodeContext, count)
	current = ctx.parent
	for i := 0; i < count; i++ {
		result[i] = current
		current = current.parent
	}

	return result
}

// AncestorsAndSelf returns all ancestor contexts including this context.
// The slice is ordered from this context to root.
// Always returns at least one element (this context).
//
// Example:
//
//	path := ctx.AncestorsAndSelf()
//	fmt.Printf("Path length from root: %d\n", len(path))
func (ctx *NodeContext) AncestorsAndSelf() []*NodeContext {
	if ctx == nil {
		return nil
	}

	// Count total nodes in path
	count := 1
	current := ctx.parent
	for current != nil {
		count++
		current = current.parent
	}

	// Collect path
	result := make([]*NodeContext, count)
	result[0] = ctx
	current = ctx.parent
	for i := 1; i < count; i++ {
		result[i] = current
		current = current.parent
	}

	return result
}

// Root returns the root context by walking up to the top of the tree.
//
// Example:
//
//	root := ctx.Root()
//	fmt.Printf("Root node: %v\n", root.Node.Style.Display)
func (ctx *NodeContext) Root() *NodeContext {
	if ctx == nil {
		return nil
	}

	current := ctx
	for current.parent != nil {
		current = current.parent
	}
	return current
}

// Siblings returns all sibling contexts (nodes with the same parent).
// Does not include this context itself.
// Returns empty slice if this is the root or has no siblings.
//
// Example:
//
//	siblings := ctx.Siblings()
//	fmt.Printf("Has %d siblings\n", len(siblings))
func (ctx *NodeContext) Siblings() []*NodeContext {
	if ctx == nil || ctx.parent == nil {
		return nil
	}

	parentNode := ctx.parent.Node
	if parentNode == nil || len(parentNode.Children) <= 1 {
		return nil
	}

	// Find siblings (all children except this one)
	result := make([]*NodeContext, 0, len(parentNode.Children)-1)
	for _, child := range parentNode.Children {
		if child != ctx.Node {
			// Create context for sibling
			siblingCtx := &NodeContext{
				Node:   child,
				parent: ctx.parent,
				depth:  ctx.depth,
			}
			result = append(result, siblingCtx)
		}
	}

	return result
}

// Depth returns the distance from the root (root = 0).
//
// Example:
//
//	depth := ctx.Depth()
//	fmt.Printf("Node is %d levels deep\n", depth)
func (ctx *NodeContext) Depth() int {
	if ctx == nil {
		return -1
	}
	return ctx.depth
}

// =============================================================================
// Downward Navigation (with context)
// =============================================================================

// Children returns all child contexts.
// Each child context maintains a reference to this context as parent.
//
// Example:
//
//	for _, childCtx := range ctx.Children() {
//	    fmt.Printf("Child at depth %d\n", childCtx.Depth())
//	}
func (ctx *NodeContext) Children() []*NodeContext {
	if ctx == nil || ctx.Node == nil {
		return nil
	}

	if len(ctx.Node.Children) == 0 {
		return nil
	}

	result := make([]*NodeContext, len(ctx.Node.Children))
	for i, child := range ctx.Node.Children {
		result[i] = &NodeContext{
			Node:   child,
			parent: ctx,
			depth:  ctx.depth + 1,
		}
	}

	return result
}

// ChildAt returns the child context at the specified index.
// Returns nil if the index is out of bounds.
//
// Example:
//
//	firstChild := ctx.ChildAt(0)
//	if firstChild != nil {
//	    fmt.Printf("First child: %v\n", firstChild.Node.Style.Display)
//	}
func (ctx *NodeContext) ChildAt(index int) *NodeContext {
	if ctx == nil || ctx.Node == nil {
		return nil
	}

	if index < 0 || index >= len(ctx.Node.Children) {
		return nil
	}

	return &NodeContext{
		Node:   ctx.Node.Children[index],
		parent: ctx,
		depth:  ctx.depth + 1,
	}
}

// =============================================================================
// Querying with Context
// =============================================================================

// FindUp searches ancestors for the first node matching the predicate.
// Searches from parent towards root.
// Returns nil if no match is found or if this is the root.
//
// Example:
//
//	// Find the containing flex container
//	flexCtx := ctx.FindUp(func(n *Node) bool {
//	    return n.Style.Display == DisplayFlex
//	})
func (ctx *NodeContext) FindUp(predicate func(*Node) bool) *NodeContext {
	if ctx == nil || predicate == nil || ctx.parent == nil {
		return nil
	}

	current := ctx.parent
	for current != nil {
		if predicate(current.Node) {
			return current
		}
		current = current.parent
	}

	return nil
}

// FindDown searches descendants for the first node matching the predicate.
// Uses depth-first search with early termination.
// Returns nil if no match is found.
//
// Example:
//
//	// Find a text node in descendants
//	textCtx := ctx.FindDown(func(n *Node) bool {
//	    return n.Text != ""
//	})
func (ctx *NodeContext) FindDown(predicate func(*Node) bool) *NodeContext {
	if ctx == nil || predicate == nil || ctx.Node == nil {
		return nil
	}

	// Depth-first search
	var search func(*NodeContext) *NodeContext
	search = func(current *NodeContext) *NodeContext {
		for _, child := range current.Node.Children {
			if predicate(child) {
				return &NodeContext{
					Node:   child,
					parent: current,
					depth:  current.depth + 1,
				}
			}

			// Recursive search in child's subtree
			childCtx := &NodeContext{
				Node:   child,
				parent: current,
				depth:  current.depth + 1,
			}
			if found := search(childCtx); found != nil {
				return found
			}
		}
		return nil
	}

	return search(ctx)
}

// FindDownAll searches descendants for all nodes matching the predicate.
// Returns all matching contexts in depth-first order.
//
// Example:
//
//	// Find all flex containers in descendants
//	flexContexts := ctx.FindDownAll(func(n *Node) bool {
//	    return n.Style.Display == DisplayFlex
//	})
func (ctx *NodeContext) FindDownAll(predicate func(*Node) bool) []*NodeContext {
	if ctx == nil || predicate == nil || ctx.Node == nil {
		return nil
	}

	result := make([]*NodeContext, 0, 10)

	// Depth-first collection
	var collect func(*NodeContext)
	collect = func(current *NodeContext) {
		for _, child := range current.Node.Children {
			childCtx := &NodeContext{
				Node:   child,
				parent: current,
				depth:  current.depth + 1,
			}

			if predicate(child) {
				result = append(result, childCtx)
			}

			collect(childCtx)
		}
	}

	collect(ctx)
	return result
}

// =============================================================================
// Utility
// =============================================================================

// Unwrap returns the underlying Node.
// Useful when you need to pass the node to functions expecting *Node.
//
// Example:
//
//	node := ctx.Unwrap()
//	Layout(node, constraints)
func (ctx *NodeContext) Unwrap() *Node {
	if ctx == nil {
		return nil
	}
	return ctx.Node
}

// IsRoot returns true if this context represents the root node.
//
// Example:
//
//	if ctx.IsRoot() {
//	    fmt.Println("This is the root")
//	}
func (ctx *NodeContext) IsRoot() bool {
	return ctx != nil && ctx.parent == nil
}

// HasParent returns true if this context has a parent.
//
// Example:
//
//	if ctx.HasParent() {
//	    parent := ctx.Parent()
//	}
func (ctx *NodeContext) HasParent() bool {
	return ctx != nil && ctx.parent != nil
}

// HasChildren returns true if the underlying node has children.
//
// Example:
//
//	if ctx.HasChildren() {
//	    children := ctx.Children()
//	}
func (ctx *NodeContext) HasChildren() bool {
	return ctx != nil && ctx.Node != nil && len(ctx.Node.Children) > 0
}
