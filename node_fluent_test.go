package layout

import (
	"testing"
)

// =============================================================================
// Test Helpers
// =============================================================================

// createTestTree creates a multi-level tree for testing:
//
//	root (Display: Block)
//	├── child1 (Display: Flex, Width: 100)
//	│   ├── grandchild1 (Display: Block, Width: 50, Text: "text1")
//	│   └── grandchild2 (Display: Grid, Width: 60)
//	├── child2 (Display: Grid, Width: 200)
//	└── child3 (Display: Flex, Width: 150, Text: "text2")
func createTestTree() *Node {
	return &Node{
		Style: Style{Display: DisplayBlock},
		Children: []*Node{
			{
				Style: Style{Display: DisplayFlex, Width: 100},
				Children: []*Node{
					{Style: Style{Display: DisplayBlock, Width: 50}, Text: "text1"},
					{Style: Style{Display: DisplayGrid, Width: 60}},
				},
			},
			{Style: Style{Display: DisplayGrid, Width: 200}},
			{Style: Style{Display: DisplayFlex, Width: 150}, Text: "text2"},
		},
	}
}

// createDeepTree creates a tree with specified depth (linear chain)
func createDeepTree(depth int) *Node {
	if depth == 0 {
		return &Node{Style: Style{Display: DisplayBlock}}
	}

	return &Node{
		Style:    Style{Display: DisplayBlock},
		Children: []*Node{createDeepTree(depth - 1)},
	}
}

// createLargeTree creates a tree with many nodes (branching factor 3, depth levels)
func createLargeTree(levels int) *Node {
	if levels == 0 {
		return &Node{Style: Style{Display: DisplayBlock}}
	}

	return &Node{
		Style: Style{Display: DisplayBlock},
		Children: []*Node{
			createLargeTree(levels - 1),
			createLargeTree(levels - 1),
			createLargeTree(levels - 1),
		},
	}
}

// =============================================================================
// Navigation Tests
// =============================================================================

func TestDescendants(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		var node *Node
		descendants := node.Descendants()
		if descendants != nil {
			t.Errorf("Expected nil for nil node, got %v", descendants)
		}
	})

	t.Run("no children", func(t *testing.T) {
		node := &Node{Style: Style{Display: DisplayBlock}}
		descendants := node.Descendants()
		if len(descendants) != 0 {
			t.Errorf("Expected 0 descendants, got %d", len(descendants))
		}
	})

	t.Run("single level", func(t *testing.T) {
		root := &Node{
			Children: []*Node{
				{Style: Style{Width: 1}},
				{Style: Style{Width: 2}},
				{Style: Style{Width: 3}},
			},
		}
		descendants := root.Descendants()
		if len(descendants) != 3 {
			t.Errorf("Expected 3 descendants, got %d", len(descendants))
		}
		// Verify order (depth-first)
		if descendants[0].Style.Width != 1 {
			t.Errorf("First descendant should have width 1, got %.2f", descendants[0].Style.Width)
		}
	})

	t.Run("multi-level tree", func(t *testing.T) {
		root := createTestTree()
		descendants := root.Descendants()

		// Should have: child1, grandchild1, grandchild2, child2, child3 = 5 nodes
		if len(descendants) != 5 {
			t.Errorf("Expected 5 descendants, got %d", len(descendants))
		}

		// Verify depth-first order: child1 should come before child2
		if descendants[0].Style.Width != 100 {
			t.Errorf("First descendant (child1) should have width 100, got %.2f", descendants[0].Style.Width)
		}
	})

	t.Run("deep tree", func(t *testing.T) {
		root := createDeepTree(10)
		descendants := root.Descendants()
		if len(descendants) != 10 {
			t.Errorf("Expected 10 descendants in depth-10 tree, got %d", len(descendants))
		}
	})
}

func TestDescendantsAndSelf(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		var node *Node
		result := node.DescendantsAndSelf()
		if result != nil {
			t.Errorf("Expected nil for nil node, got %v", result)
		}
	})

	t.Run("includes self", func(t *testing.T) {
		root := createTestTree()
		result := root.DescendantsAndSelf()

		// Should include root + 5 descendants = 6 total
		if len(result) != 6 {
			t.Errorf("Expected 6 nodes (root + 5 descendants), got %d", len(result))
		}

		// First should be root
		if result[0] != root {
			t.Errorf("First node should be root")
		}
	})
}

func TestFirstChild(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		var node *Node
		if node.FirstChild() != nil {
			t.Errorf("Expected nil for nil node")
		}
	})

	t.Run("no children", func(t *testing.T) {
		node := &Node{Style: Style{Display: DisplayBlock}}
		if node.FirstChild() != nil {
			t.Errorf("Expected nil for node with no children")
		}
	})

	t.Run("with children", func(t *testing.T) {
		root := createTestTree()
		first := root.FirstChild()
		if first == nil {
			t.Fatal("Expected first child, got nil")
		}
		if first.Style.Width != 100 {
			t.Errorf("First child should have width 100, got %.2f", first.Style.Width)
		}
	})
}

func TestLastChild(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		var node *Node
		if node.LastChild() != nil {
			t.Errorf("Expected nil for nil node")
		}
	})

	t.Run("no children", func(t *testing.T) {
		node := &Node{Style: Style{Display: DisplayBlock}}
		if node.LastChild() != nil {
			t.Errorf("Expected nil for node with no children")
		}
	})

	t.Run("with children", func(t *testing.T) {
		root := createTestTree()
		last := root.LastChild()
		if last == nil {
			t.Fatal("Expected last child, got nil")
		}
		if last.Style.Width != 150 {
			t.Errorf("Last child should have width 150, got %.2f", last.Style.Width)
		}
	})
}

func TestChildAt(t *testing.T) {
	root := createTestTree()

	t.Run("nil node", func(t *testing.T) {
		var node *Node
		if node.ChildAt(0) != nil {
			t.Errorf("Expected nil for nil node")
		}
	})

	t.Run("negative index", func(t *testing.T) {
		if root.ChildAt(-1) != nil {
			t.Errorf("Expected nil for negative index")
		}
	})

	t.Run("out of bounds", func(t *testing.T) {
		if root.ChildAt(100) != nil {
			t.Errorf("Expected nil for out of bounds index")
		}
	})

	t.Run("valid indices", func(t *testing.T) {
		child0 := root.ChildAt(0)
		if child0 == nil || child0.Style.Width != 100 {
			t.Errorf("Child at index 0 should have width 100")
		}

		child2 := root.ChildAt(2)
		if child2 == nil || child2.Style.Width != 150 {
			t.Errorf("Child at index 2 should have width 150")
		}
	})
}

func TestChildCount(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		var node *Node
		if node.ChildCount() != 0 {
			t.Errorf("Expected 0 for nil node")
		}
	})

	t.Run("no children", func(t *testing.T) {
		node := &Node{Style: Style{Display: DisplayBlock}}
		if node.ChildCount() != 0 {
			t.Errorf("Expected 0 children")
		}
	})

	t.Run("with children", func(t *testing.T) {
		root := createTestTree()
		if root.ChildCount() != 3 {
			t.Errorf("Expected 3 children, got %d", root.ChildCount())
		}
	})
}

// =============================================================================
// Querying Tests
// =============================================================================

func TestFind(t *testing.T) {
	root := createTestTree()

	t.Run("nil node", func(t *testing.T) {
		var node *Node
		result := node.Find(func(n *Node) bool { return true })
		if result != nil {
			t.Errorf("Expected nil for nil node")
		}
	})

	t.Run("nil predicate", func(t *testing.T) {
		result := root.Find(nil)
		if result != nil {
			t.Errorf("Expected nil for nil predicate")
		}
	})

	t.Run("find by display type", func(t *testing.T) {
		grid := root.Find(func(n *Node) bool {
			return n.Style.Display == DisplayGrid
		})
		if grid == nil {
			t.Fatal("Expected to find grid node")
		}
		// Should find the first grid (grandchild2, width 60)
		if grid.Style.Width != 60 {
			t.Errorf("Expected first grid to have width 60, got %.2f", grid.Style.Width)
		}
	})

	t.Run("find by text", func(t *testing.T) {
		textNode := root.Find(func(n *Node) bool {
			return n.Text != ""
		})
		if textNode == nil {
			t.Fatal("Expected to find text node")
		}
		if textNode.Text != "text1" {
			t.Errorf("Expected first text node to have 'text1', got %s", textNode.Text)
		}
	})

	t.Run("no match", func(t *testing.T) {
		result := root.Find(func(n *Node) bool {
			return n.Style.Width == 999
		})
		if result != nil {
			t.Errorf("Expected nil when no match found")
		}
	})
}

func TestFindAll(t *testing.T) {
	root := createTestTree()

	t.Run("nil node", func(t *testing.T) {
		var node *Node
		result := node.FindAll(func(n *Node) bool { return true })
		if result != nil {
			t.Errorf("Expected nil for nil node")
		}
	})

	t.Run("nil predicate", func(t *testing.T) {
		result := root.FindAll(nil)
		if result != nil {
			t.Errorf("Expected nil for nil predicate")
		}
	})

	t.Run("find all flex containers", func(t *testing.T) {
		flexes := root.FindAll(func(n *Node) bool {
			return n.Style.Display == DisplayFlex
		})
		if len(flexes) != 2 {
			t.Errorf("Expected 2 flex containers, got %d", len(flexes))
		}
	})

	t.Run("find all grids", func(t *testing.T) {
		grids := root.FindAll(func(n *Node) bool {
			return n.Style.Display == DisplayGrid
		})
		if len(grids) != 2 {
			t.Errorf("Expected 2 grid containers, got %d", len(grids))
		}
	})

	t.Run("find all with text", func(t *testing.T) {
		textNodes := root.FindAll(func(n *Node) bool {
			return n.Text != ""
		})
		if len(textNodes) != 2 {
			t.Errorf("Expected 2 text nodes, got %d", len(textNodes))
		}
	})

	t.Run("no matches", func(t *testing.T) {
		result := root.FindAll(func(n *Node) bool {
			return n.Style.Width == 999
		})
		if len(result) != 0 {
			t.Errorf("Expected 0 results, got %d", len(result))
		}
	})

	t.Run("match all", func(t *testing.T) {
		all := root.FindAll(func(n *Node) bool {
			return true
		})
		if len(all) != 5 {
			t.Errorf("Expected 5 nodes (all descendants), got %d", len(all))
		}
	})
}

func TestWhere(t *testing.T) {
	root := createTestTree()

	t.Run("alias for FindAll", func(t *testing.T) {
		findAllResult := root.FindAll(func(n *Node) bool {
			return n.Style.Display == DisplayFlex
		})
		whereResult := root.Where(func(n *Node) bool {
			return n.Style.Display == DisplayFlex
		})

		if len(findAllResult) != len(whereResult) {
			t.Errorf("Where and FindAll should return same results")
		}
	})

	t.Run("wide nodes", func(t *testing.T) {
		wideNodes := root.Where(func(n *Node) bool {
			return n.Style.Width > 100
		})
		if len(wideNodes) != 2 {
			t.Errorf("Expected 2 wide nodes (150, 200), got %d", len(wideNodes))
		}
	})
}

func TestAny(t *testing.T) {
	root := createTestTree()

	t.Run("nil node", func(t *testing.T) {
		var node *Node
		if node.Any(func(n *Node) bool { return true }) {
			t.Errorf("Expected false for nil node")
		}
	})

	t.Run("nil predicate", func(t *testing.T) {
		if root.Any(nil) {
			t.Errorf("Expected false for nil predicate")
		}
	})

	t.Run("has text nodes", func(t *testing.T) {
		hasText := root.Any(func(n *Node) bool {
			return n.Text != ""
		})
		if !hasText {
			t.Errorf("Expected to find text nodes")
		}
	})

	t.Run("has flex containers", func(t *testing.T) {
		hasFlex := root.Any(func(n *Node) bool {
			return n.Style.Display == DisplayFlex
		})
		if !hasFlex {
			t.Errorf("Expected to find flex containers")
		}
	})

	t.Run("no matches", func(t *testing.T) {
		result := root.Any(func(n *Node) bool {
			return n.Style.Width == 999
		})
		if result {
			t.Errorf("Expected false when no matches found")
		}
	})

	t.Run("empty tree", func(t *testing.T) {
		emptyNode := &Node{}
		result := emptyNode.Any(func(n *Node) bool {
			return true
		})
		if result {
			t.Errorf("Expected false for empty tree")
		}
	})
}

func TestAll(t *testing.T) {
	root := createTestTree()

	t.Run("nil node", func(t *testing.T) {
		var node *Node
		if !node.All(func(n *Node) bool { return false }) {
			t.Errorf("Expected true for nil node (vacuous truth)")
		}
	})

	t.Run("nil predicate", func(t *testing.T) {
		if !root.All(nil) {
			t.Errorf("Expected true for nil predicate")
		}
	})

	t.Run("all have width less than 300", func(t *testing.T) {
		allNarrow := root.All(func(n *Node) bool {
			return n.Style.Width < 300
		})
		if !allNarrow {
			t.Errorf("Expected all nodes to have width < 300")
		}
	})

	t.Run("not all are flex", func(t *testing.T) {
		allFlex := root.All(func(n *Node) bool {
			return n.Style.Display == DisplayFlex
		})
		if allFlex {
			t.Errorf("Expected not all nodes to be flex")
		}
	})

	t.Run("empty tree is true", func(t *testing.T) {
		emptyNode := &Node{}
		result := emptyNode.All(func(n *Node) bool {
			return false // Even with false predicate, should be true (vacuous)
		})
		if !result {
			t.Errorf("Expected true for empty tree (vacuous truth)")
		}
	})
}

func TestOfDisplayType(t *testing.T) {
	root := createTestTree()

	t.Run("find flex containers", func(t *testing.T) {
		flexes := root.OfDisplayType(DisplayFlex)
		if len(flexes) != 2 {
			t.Errorf("Expected 2 flex containers, got %d", len(flexes))
		}
	})

	t.Run("find grid containers", func(t *testing.T) {
		grids := root.OfDisplayType(DisplayGrid)
		if len(grids) != 2 {
			t.Errorf("Expected 2 grid containers, got %d", len(grids))
		}
	})

	t.Run("find block containers", func(t *testing.T) {
		blocks := root.OfDisplayType(DisplayBlock)
		if len(blocks) != 1 {
			t.Errorf("Expected 1 block container, got %d", len(blocks))
		}
	})

	t.Run("no matches", func(t *testing.T) {
		result := root.OfDisplayType(DisplayInlineText)
		if len(result) != 0 {
			t.Errorf("Expected 0 inline text containers, got %d", len(result))
		}
	})
}

// =============================================================================
// Performance Tests
// =============================================================================

func TestPerformanceDescendants(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create large tree: 3 levels with branching factor 3 = 40 nodes
	root := createLargeTree(3)
	descendants := root.Descendants()

	// 3^1 + 3^2 + 3^3 = 3 + 9 + 27 = 39 descendants
	expectedCount := 39
	if len(descendants) != expectedCount {
		t.Errorf("Expected %d descendants, got %d", expectedCount, len(descendants))
	}
}

func BenchmarkDescendants(b *testing.B) {
	root := createLargeTree(4) // 3^1 + 3^2 + 3^3 + 3^4 = 120 nodes

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.Descendants()
	}
}

func BenchmarkFind(b *testing.B) {
	root := createLargeTree(4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.Find(func(n *Node) bool {
			return n.Style.Display == DisplayGrid
		})
	}
}

func BenchmarkFindAll(b *testing.B) {
	root := createLargeTree(4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.FindAll(func(n *Node) bool {
			return n.Style.Display == DisplayBlock
		})
	}
}
