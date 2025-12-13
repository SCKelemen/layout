package layout

import (
	"testing"
)

// =============================================================================
// Test Helpers
// =============================================================================

// createContextTestTree creates a tree for context testing:
//
//	root (Block)
//	├── child1 (Flex, Width: 100)
//	│   ├── grandchild1 (Block, Width: 50)
//	│   └── grandchild2 (Grid, Width: 60)
//	├── child2 (Grid, Width: 200)
//	└── child3 (Flex, Width: 150)
func createContextTestTree() *Node {
	return createTestTree() // Reuse from node_fluent_test.go
}

// =============================================================================
// Context Creation Tests
// =============================================================================

func TestNewContext(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		ctx := NewContext(nil)
		if ctx != nil {
			t.Errorf("Expected nil context for nil node")
		}
	})

	t.Run("valid node", func(t *testing.T) {
		root := &Node{Style: Style{Display: DisplayBlock}}
		ctx := NewContext(root)

		if ctx == nil {
			t.Fatal("Expected context, got nil")
		}

		if ctx.Node != root {
			t.Errorf("Context should wrap the root node")
		}

		if ctx.parent != nil {
			t.Errorf("Root context should have no parent")
		}

		if ctx.depth != 0 {
			t.Errorf("Root context should have depth 0, got %d", ctx.depth)
		}
	})
}

// =============================================================================
// Upward Navigation Tests
// =============================================================================

func TestParent(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	t.Run("root has no parent", func(t *testing.T) {
		if rootCtx.Parent() != nil {
			t.Errorf("Root should have no parent")
		}
	})

	t.Run("child has parent", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(0)
		if childCtx == nil {
			t.Fatal("Failed to get child context")
		}

		parentCtx := childCtx.Parent()
		if parentCtx == nil {
			t.Fatal("Child should have parent")
		}

		if parentCtx.Node != root {
			t.Errorf("Parent should be root node")
		}
	})

	t.Run("grandchild has parent", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(0)
		grandchildCtx := childCtx.ChildAt(0)

		if grandchildCtx == nil {
			t.Fatal("Failed to get grandchild context")
		}

		parentCtx := grandchildCtx.Parent()
		if parentCtx == nil {
			t.Fatal("Grandchild should have parent")
		}

		if parentCtx.Node != root.Children[0] {
			t.Errorf("Parent should be child1")
		}
	})
}

func TestAncestors(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	t.Run("root has no ancestors", func(t *testing.T) {
		ancestors := rootCtx.Ancestors()
		if ancestors != nil && len(ancestors) != 0 {
			t.Errorf("Root should have no ancestors")
		}
	})

	t.Run("child has one ancestor", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(0)
		ancestors := childCtx.Ancestors()

		if len(ancestors) != 1 {
			t.Errorf("Child should have 1 ancestor, got %d", len(ancestors))
		}

		if ancestors[0].Node != root {
			t.Errorf("Ancestor should be root")
		}
	})

	t.Run("grandchild has two ancestors", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(0)
		grandchildCtx := childCtx.ChildAt(0)
		ancestors := grandchildCtx.Ancestors()

		if len(ancestors) != 2 {
			t.Errorf("Grandchild should have 2 ancestors, got %d", len(ancestors))
		}

		// Ordered from nearest to furthest
		if ancestors[0].Node != root.Children[0] {
			t.Errorf("First ancestor should be parent")
		}

		if ancestors[1].Node != root {
			t.Errorf("Second ancestor should be root")
		}
	})
}

func TestAncestorsAndSelf(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	t.Run("root includes self", func(t *testing.T) {
		path := rootCtx.AncestorsAndSelf()

		if len(path) != 1 {
			t.Errorf("Root path should have 1 element, got %d", len(path))
		}

		if path[0].Node != root {
			t.Errorf("Path should start with root")
		}
	})

	t.Run("grandchild includes self and ancestors", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(0)
		grandchildCtx := childCtx.ChildAt(0)
		path := grandchildCtx.AncestorsAndSelf()

		if len(path) != 3 {
			t.Errorf("Grandchild path should have 3 elements, got %d", len(path))
		}

		// Ordered from self to root
		if path[0] != grandchildCtx {
			t.Errorf("Path should start with self")
		}

		if path[1].Node != root.Children[0] {
			t.Errorf("Second should be parent")
		}

		if path[2].Node != root {
			t.Errorf("Last should be root")
		}
	})
}

func TestRoot(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	t.Run("root returns self", func(t *testing.T) {
		foundRoot := rootCtx.Root()
		if foundRoot != rootCtx {
			t.Errorf("Root.Root() should return self")
		}
	})

	t.Run("child returns root", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(0)
		foundRoot := childCtx.Root()

		if foundRoot.Node != root {
			t.Errorf("Child.Root() should return root node")
		}
	})

	t.Run("deep node returns root", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(0)
		grandchildCtx := childCtx.ChildAt(0)
		foundRoot := grandchildCtx.Root()

		if foundRoot.Node != root {
			t.Errorf("Grandchild.Root() should return root node")
		}
	})
}

func TestSiblings(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	t.Run("root has no siblings", func(t *testing.T) {
		siblings := rootCtx.Siblings()
		if siblings != nil && len(siblings) != 0 {
			t.Errorf("Root should have no siblings")
		}
	})

	t.Run("child with siblings", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(0)
		siblings := childCtx.Siblings()

		// root has 3 children, so each has 2 siblings
		if len(siblings) != 2 {
			t.Errorf("Child should have 2 siblings, got %d", len(siblings))
		}

		// Verify siblings are not the same as this node
		for _, sibling := range siblings {
			if sibling.Node == childCtx.Node {
				t.Errorf("Sibling should not be self")
			}
		}
	})

	t.Run("only child has no siblings", func(t *testing.T) {
		// Create tree with only child
		onlyChild := &Node{
			Children: []*Node{
				{Style: Style{Width: 100}},
			},
		}
		ctx := NewContext(onlyChild)
		childCtx := ctx.ChildAt(0)

		siblings := childCtx.Siblings()
		if siblings != nil && len(siblings) != 0 {
			t.Errorf("Only child should have no siblings")
		}
	})
}

func TestDepth(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	t.Run("root depth is 0", func(t *testing.T) {
		if rootCtx.Depth() != 0 {
			t.Errorf("Root depth should be 0, got %d", rootCtx.Depth())
		}
	})

	t.Run("child depth is 1", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(0)
		if childCtx.Depth() != 1 {
			t.Errorf("Child depth should be 1, got %d", childCtx.Depth())
		}
	})

	t.Run("grandchild depth is 2", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(0)
		grandchildCtx := childCtx.ChildAt(0)
		if grandchildCtx.Depth() != 2 {
			t.Errorf("Grandchild depth should be 2, got %d", grandchildCtx.Depth())
		}
	})
}

// =============================================================================
// Downward Navigation Tests
// =============================================================================

func TestChildren(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	t.Run("root has children", func(t *testing.T) {
		children := rootCtx.Children()

		if len(children) != 3 {
			t.Errorf("Root should have 3 children, got %d", len(children))
		}

		// Verify each child has correct parent
		for _, childCtx := range children {
			if childCtx.Parent() != rootCtx {
				t.Errorf("Child's parent should be root context")
			}

			if childCtx.Depth() != 1 {
				t.Errorf("Child depth should be 1, got %d", childCtx.Depth())
			}
		}
	})

	t.Run("leaf has no children", func(t *testing.T) {
		leaf := &Node{Style: Style{Width: 100}}
		leafCtx := NewContext(leaf)

		children := leafCtx.Children()
		if children != nil && len(children) != 0 {
			t.Errorf("Leaf should have no children")
		}
	})
}

func TestContextChildAt(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	t.Run("valid index", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(0)

		if childCtx == nil {
			t.Fatal("Expected child context")
		}

		if childCtx.Node != root.Children[0] {
			t.Errorf("ChildAt(0) should return first child")
		}

		if childCtx.Parent().Node != root {
			t.Errorf("Child's parent should be root")
		}
	})

	t.Run("out of bounds", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(100)
		if childCtx != nil {
			t.Errorf("Out of bounds should return nil")
		}
	})

	t.Run("negative index", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(-1)
		if childCtx != nil {
			t.Errorf("Negative index should return nil")
		}
	})
}

// =============================================================================
// Querying Tests
// =============================================================================

func TestFindUp(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	t.Run("find flex container ancestor", func(t *testing.T) {
		// Start from grandchild1 (inside child1 which is flex)
		childCtx := rootCtx.ChildAt(0)       // child1 (Flex)
		grandchildCtx := childCtx.ChildAt(0) // grandchild1 (Block)

		flexCtx := grandchildCtx.FindUp(func(n *Node) bool {
			return n.Style.Display == DisplayFlex
		})

		if flexCtx == nil {
			t.Fatal("Should find flex ancestor")
		}

		if flexCtx.Node != root.Children[0] {
			t.Errorf("Should find child1 as flex ancestor")
		}
	})

	t.Run("no match returns nil", func(t *testing.T) {
		childCtx := rootCtx.ChildAt(0)

		result := childCtx.FindUp(func(n *Node) bool {
			return n.Style.Width == 999 // Doesn't exist
		})

		if result != nil {
			t.Errorf("Should return nil when no match")
		}
	})

	t.Run("root has nothing to search", func(t *testing.T) {
		result := rootCtx.FindUp(func(n *Node) bool {
			return true
		})

		if result != nil {
			t.Errorf("Root should have no ancestors to search")
		}
	})
}

func TestFindDown(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	t.Run("find text node", func(t *testing.T) {
		textCtx := rootCtx.FindDown(func(n *Node) bool {
			return n.Text != ""
		})

		if textCtx == nil {
			t.Fatal("Should find text node")
		}

		if textCtx.Node.Text == "" {
			t.Errorf("Found node should have text")
		}
	})

	t.Run("find grid container", func(t *testing.T) {
		gridCtx := rootCtx.FindDown(func(n *Node) bool {
			return n.Style.Display == DisplayGrid
		})

		if gridCtx == nil {
			t.Fatal("Should find grid node")
		}

		if gridCtx.Node.Style.Display != DisplayGrid {
			t.Errorf("Found node should be grid")
		}

		// Verify context has proper parent chain
		if !gridCtx.HasParent() {
			t.Errorf("Found context should have parent")
		}
	})

	t.Run("no match returns nil", func(t *testing.T) {
		result := rootCtx.FindDown(func(n *Node) bool {
			return n.Style.Width == 999
		})

		if result != nil {
			t.Errorf("Should return nil when no match")
		}
	})
}

func TestFindDownAll(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	t.Run("find all flex containers", func(t *testing.T) {
		flexContexts := rootCtx.FindDownAll(func(n *Node) bool {
			return n.Style.Display == DisplayFlex
		})

		if len(flexContexts) != 2 {
			t.Errorf("Should find 2 flex containers, got %d", len(flexContexts))
		}

		// Verify each has proper context
		for _, ctx := range flexContexts {
			if !ctx.HasParent() {
				t.Errorf("Found context should have parent")
			}

			if ctx.Depth() < 1 {
				t.Errorf("Found context should have depth >= 1")
			}
		}
	})

	t.Run("find all with text", func(t *testing.T) {
		textContexts := rootCtx.FindDownAll(func(n *Node) bool {
			return n.Text != ""
		})

		if len(textContexts) != 2 {
			t.Errorf("Should find 2 text nodes, got %d", len(textContexts))
		}
	})

	t.Run("no matches returns empty", func(t *testing.T) {
		result := rootCtx.FindDownAll(func(n *Node) bool {
			return n.Style.Width == 999
		})

		if len(result) != 0 {
			t.Errorf("Should return empty slice, got %d results", len(result))
		}
	})
}

// =============================================================================
// Utility Tests
// =============================================================================

func TestUnwrap(t *testing.T) {
	root := &Node{Style: Style{Display: DisplayBlock}}
	ctx := NewContext(root)

	unwrapped := ctx.Unwrap()

	if unwrapped != root {
		t.Errorf("Unwrap should return the original node")
	}
}

func TestIsRoot(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	if !rootCtx.IsRoot() {
		t.Errorf("Root context should report IsRoot() = true")
	}

	childCtx := rootCtx.ChildAt(0)
	if childCtx.IsRoot() {
		t.Errorf("Child context should report IsRoot() = false")
	}
}

func TestHasParent(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	if rootCtx.HasParent() {
		t.Errorf("Root should not have parent")
	}

	childCtx := rootCtx.ChildAt(0)
	if !childCtx.HasParent() {
		t.Errorf("Child should have parent")
	}
}

func TestHasChildren(t *testing.T) {
	root := createContextTestTree()
	rootCtx := NewContext(root)

	if !rootCtx.HasChildren() {
		t.Errorf("Root should have children")
	}

	leaf := &Node{Style: Style{Width: 100}}
	leafCtx := NewContext(leaf)

	if leafCtx.HasChildren() {
		t.Errorf("Leaf should not have children")
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestContextNavigationIntegration(t *testing.T) {
	// Build a tree and navigate using context
	root := createContextTestTree()
	rootCtx := NewContext(root)

	// Navigate down to grandchild
	childCtx := rootCtx.ChildAt(0)
	grandchildCtx := childCtx.ChildAt(0)

	// Navigate back up to root
	foundRoot := grandchildCtx.Root()
	if foundRoot.Node != root {
		t.Errorf("Should navigate back to root")
	}

	// Find flex ancestor
	flexAncestor := grandchildCtx.FindUp(func(n *Node) bool {
		return n.Style.Display == DisplayFlex
	})

	if flexAncestor == nil {
		t.Fatal("Should find flex ancestor")
	}

	// Verify it's the direct parent
	if flexAncestor.Node != childCtx.Node {
		t.Errorf("Flex ancestor should be child1")
	}
}

func TestContextWithFluentAPI(t *testing.T) {
	// Test that context works with fluently-built trees
	tree := (&Node{}).
		WithDisplay(DisplayFlex).
		AddChild(
			(&Node{}).
				WithDisplay(DisplayGrid).
				AddChild((&Node{}).WithText("leaf")),
		)

	ctx := NewContext(tree)

	// Find the text leaf
	leafCtx := ctx.FindDown(func(n *Node) bool {
		return n.Text == "leaf"
	})

	if leafCtx == nil {
		t.Fatal("Should find leaf")
	}

	// Navigate up to find grid
	gridCtx := leafCtx.FindUp(func(n *Node) bool {
		return n.Style.Display == DisplayGrid
	})

	if gridCtx == nil {
		t.Fatal("Should find grid ancestor")
	}

	// Navigate up to find flex (root)
	flexCtx := leafCtx.FindUp(func(n *Node) bool {
		return n.Style.Display == DisplayFlex
	})

	if flexCtx == nil {
		t.Fatal("Should find flex ancestor")
	}

	if !flexCtx.IsRoot() {
		t.Errorf("Flex should be root")
	}
}

func TestContextMemoryEfficiency(t *testing.T) {
	// Verify contexts are only created when accessed
	root := createContextTestTree()
	rootCtx := NewContext(root)

	// Getting children creates contexts on-demand
	children := rootCtx.Children()

	if len(children) != 3 {
		t.Errorf("Should have 3 children")
	}

	// Each child should have independent context
	child0Ctx1 := rootCtx.ChildAt(0)
	child0Ctx2 := rootCtx.ChildAt(0)

	// These are different context instances (created on-demand)
	// but they wrap the same node
	if child0Ctx1.Node != child0Ctx2.Node {
		t.Errorf("Should wrap same node")
	}

	// But they're different context instances
	if child0Ctx1 == child0Ctx2 {
		t.Errorf("Contexts created on-demand should be different instances")
	}
}
