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
// Phase 2: Immutability Tests
// =============================================================================

func TestClone(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		var node *Node
		clone := node.Clone()
		if clone != nil {
			t.Errorf("Expected nil for nil node")
		}
	})

	t.Run("shallow copy", func(t *testing.T) {
		original := &Node{
			Style: Style{Width: 100, Height: 50},
			Text:  "original",
			Children: []*Node{
				{Style: Style{Width: 25}},
			},
		}

		clone := original.Clone()

		// Verify it's a different node
		if clone == original {
			t.Errorf("Clone should be a different instance")
		}

		// Verify fields are copied
		if clone.Style.Width != 100 || clone.Text != "original" {
			t.Errorf("Fields not copied correctly")
		}

		// Verify children slice is shared (shallow copy)
		if len(clone.Children) != 1 {
			t.Errorf("Children should be shared")
		}

		// Modifying clone doesn't affect original
		clone.Style.Width = 200
		if original.Style.Width != 100 {
			t.Errorf("Original was modified by clone change")
		}
	})
}

func TestCloneDeep(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		var node *Node
		clone := node.CloneDeep()
		if clone != nil {
			t.Errorf("Expected nil for nil node")
		}
	})

	t.Run("deep copy", func(t *testing.T) {
		original := createTestTree()

		clone := original.CloneDeep()

		// Verify it's a different node
		if clone == original {
			t.Errorf("CloneDeep should create different instance")
		}

		// Verify children are also cloned
		if len(clone.Children) != len(original.Children) {
			t.Errorf("Children count mismatch")
		}

		if len(clone.Children) > 0 && clone.Children[0] == original.Children[0] {
			t.Errorf("Children should be cloned, not shared")
		}

		// Modifying deep clone doesn't affect original
		if len(clone.Children) > 0 {
			clone.Children[0].Style.Width = 999
			if original.Children[0].Style.Width == 999 {
				t.Errorf("Original child was modified")
			}
		}
	})

	t.Run("deep tree", func(t *testing.T) {
		original := createDeepTree(5)
		clone := original.CloneDeep()

		// Traverse to deepest node
		deepOriginal := original
		deepClone := clone
		for i := 0; i < 5; i++ {
			if len(deepOriginal.Children) == 0 {
				break
			}
			deepOriginal = deepOriginal.Children[0]
			deepClone = deepClone.Children[0]
		}

		// Verify they're different instances
		if deepOriginal == deepClone {
			t.Errorf("Deep nodes should be different instances")
		}
	})
}

func TestWithStyle(t *testing.T) {
	original := &Node{Style: Style{Width: 100, Display: DisplayBlock}}

	newStyle := Style{Width: 200, Display: DisplayFlex}
	modified := original.WithStyle(newStyle)

	if modified == original {
		t.Errorf("WithStyle should return new node")
	}

	if original.Style.Width != 100 || original.Style.Display != DisplayBlock {
		t.Errorf("Original was modified")
	}

	if modified.Style.Width != 200 || modified.Style.Display != DisplayFlex {
		t.Errorf("New style not applied")
	}
}

func TestWithPadding(t *testing.T) {
	original := &Node{Style: Style{Width: 100}}

	padded := original.WithPadding(16)

	if padded == original {
		t.Errorf("WithPadding should return new node")
	}

	if original.Style.Padding.Top != 0 {
		t.Errorf("Original was modified")
	}

	if padded.Style.Padding.Top != 16 || padded.Style.Padding.Right != 16 {
		t.Errorf("Padding not applied correctly")
	}
}

func TestWithPaddingCustom(t *testing.T) {
	original := &Node{Style: Style{Width: 100}}

	padded := original.WithPaddingCustom(10, 20, 30, 40)

	if original.Style.Padding.Top != 0 {
		t.Errorf("Original was modified")
	}

	if padded.Style.Padding.Top != 10 || padded.Style.Padding.Right != 20 ||
		padded.Style.Padding.Bottom != 30 || padded.Style.Padding.Left != 40 {
		t.Errorf("Custom padding not applied correctly")
	}
}

func TestWithMargin(t *testing.T) {
	original := &Node{Style: Style{Width: 100}}

	margined := original.WithMargin(8)

	if original.Style.Margin.Top != 0 {
		t.Errorf("Original was modified")
	}

	if margined.Style.Margin.Top != 8 || margined.Style.Margin.Bottom != 8 {
		t.Errorf("Margin not applied correctly")
	}
}

func TestWithWidth(t *testing.T) {
	original := &Node{Style: Style{Width: 100}}

	wider := original.WithWidth(300)

	if original.Style.Width != 100 {
		t.Errorf("Original was modified")
	}

	if wider.Style.Width != 300 {
		t.Errorf("Width not applied")
	}
}

func TestWithHeight(t *testing.T) {
	original := &Node{Style: Style{Height: 50}}

	taller := original.WithHeight(200)

	if original.Style.Height != 50 {
		t.Errorf("Original was modified")
	}

	if taller.Style.Height != 200 {
		t.Errorf("Height not applied")
	}
}

func TestWithText(t *testing.T) {
	original := &Node{Text: "original"}

	modified := original.WithText("modified")

	if original.Text != "original" {
		t.Errorf("Original was modified")
	}

	if modified.Text != "modified" {
		t.Errorf("Text not applied")
	}
}

func TestWithDisplay(t *testing.T) {
	original := &Node{Style: Style{Display: DisplayBlock}}

	flexNode := original.WithDisplay(DisplayFlex)

	if original.Style.Display != DisplayBlock {
		t.Errorf("Original was modified")
	}

	if flexNode.Style.Display != DisplayFlex {
		t.Errorf("Display not applied")
	}
}

func TestWithFlexGrow(t *testing.T) {
	original := &Node{Style: Style{FlexGrow: 0}}

	growable := original.WithFlexGrow(1)

	if original.Style.FlexGrow != 0 {
		t.Errorf("Original was modified")
	}

	if growable.Style.FlexGrow != 1 {
		t.Errorf("FlexGrow not applied")
	}
}

func TestWithFlexShrink(t *testing.T) {
	original := &Node{Style: Style{FlexShrink: 1}}

	rigid := original.WithFlexShrink(0)

	if original.Style.FlexShrink != 1 {
		t.Errorf("Original was modified")
	}

	if rigid.Style.FlexShrink != 0 {
		t.Errorf("FlexShrink not applied")
	}
}

func TestMethodChaining(t *testing.T) {
	original := &Node{Style: Style{Display: DisplayBlock}}

	// Chain multiple With* methods
	modified := original.
		WithWidth(200).
		WithHeight(100).
		WithPadding(16).
		WithMargin(8).
		WithDisplay(DisplayFlex)

	// Verify original unchanged
	if original.Style.Width != 0 || original.Style.Display != DisplayBlock {
		t.Errorf("Original was modified by chaining")
	}

	// Verify all modifications applied
	if modified.Style.Width != 200 {
		t.Errorf("Width not applied in chain")
	}
	if modified.Style.Height != 100 {
		t.Errorf("Height not applied in chain")
	}
	if modified.Style.Padding.Top != 16 {
		t.Errorf("Padding not applied in chain")
	}
	if modified.Style.Margin.Top != 8 {
		t.Errorf("Margin not applied in chain")
	}
	if modified.Style.Display != DisplayFlex {
		t.Errorf("Display not applied in chain")
	}
}

func TestWithChildren(t *testing.T) {
	child1 := &Node{Style: Style{Width: 100}}
	child2 := &Node{Style: Style{Width: 200}}
	child3 := &Node{Style: Style{Width: 300}}

	original := &Node{
		Children: []*Node{child1, child2},
	}

	modified := original.WithChildren(child2, child3)

	// Verify original unchanged
	if len(original.Children) != 2 {
		t.Errorf("Original children modified")
	}
	if original.Children[0] != child1 {
		t.Errorf("Original first child changed")
	}

	// Verify new children
	if len(modified.Children) != 2 {
		t.Errorf("Modified should have 2 children")
	}
	if modified.Children[0] != child2 || modified.Children[1] != child3 {
		t.Errorf("New children not set correctly")
	}
}

func TestAddChild(t *testing.T) {
	child1 := &Node{Style: Style{Width: 100}}
	child2 := &Node{Style: Style{Width: 200}}

	original := &Node{
		Children: []*Node{child1},
	}

	modified := original.AddChild(child2)

	// Verify original unchanged
	if len(original.Children) != 1 {
		t.Errorf("Original children count changed")
	}

	// Verify child added
	if len(modified.Children) != 2 {
		t.Errorf("Modified should have 2 children")
	}
	if modified.Children[0] != child1 || modified.Children[1] != child2 {
		t.Errorf("Child not added correctly")
	}
}

func TestAddChildren(t *testing.T) {
	child1 := &Node{Style: Style{Width: 100}}
	child2 := &Node{Style: Style{Width: 200}}
	child3 := &Node{Style: Style{Width: 300}}

	original := &Node{
		Children: []*Node{child1},
	}

	modified := original.AddChildren(child2, child3)

	// Verify original unchanged
	if len(original.Children) != 1 {
		t.Errorf("Original children count changed")
	}

	// Verify children added
	if len(modified.Children) != 3 {
		t.Errorf("Modified should have 3 children")
	}
	if modified.Children[2] != child3 {
		t.Errorf("Children not added correctly")
	}
}

func TestRemoveChildAt(t *testing.T) {
	child1 := &Node{Style: Style{Width: 100}}
	child2 := &Node{Style: Style{Width: 200}}
	child3 := &Node{Style: Style{Width: 300}}

	original := &Node{
		Children: []*Node{child1, child2, child3},
	}

	// Remove middle child
	modified := original.RemoveChildAt(1)

	// Verify original unchanged
	if len(original.Children) != 3 {
		t.Errorf("Original children count changed")
	}

	// Verify child removed
	if len(modified.Children) != 2 {
		t.Errorf("Modified should have 2 children")
	}
	if modified.Children[0] != child1 || modified.Children[1] != child3 {
		t.Errorf("Wrong child removed")
	}

	// Test out of bounds
	outOfBounds := original.RemoveChildAt(10)
	if outOfBounds != original {
		t.Errorf("Out of bounds should return original")
	}
}

func TestReplaceChildAt(t *testing.T) {
	child1 := &Node{Style: Style{Width: 100}}
	child2 := &Node{Style: Style{Width: 200}}
	newChild := &Node{Style: Style{Width: 999}}

	original := &Node{
		Children: []*Node{child1, child2},
	}

	modified := original.ReplaceChildAt(0, newChild)

	// Verify original unchanged
	if original.Children[0] != child1 {
		t.Errorf("Original child replaced")
	}

	// Verify child replaced
	if modified.Children[0] != newChild {
		t.Errorf("Child not replaced")
	}
	if modified.Children[1] != child2 {
		t.Errorf("Other children affected")
	}
}

func TestInsertChildAt(t *testing.T) {
	child1 := &Node{Style: Style{Width: 100}}
	child2 := &Node{Style: Style{Width: 200}}
	newChild := &Node{Style: Style{Width: 150}}

	original := &Node{
		Children: []*Node{child1, child2},
	}

	// Insert in middle
	modified := original.InsertChildAt(1, newChild)

	// Verify original unchanged
	if len(original.Children) != 2 {
		t.Errorf("Original children count changed")
	}

	// Verify child inserted
	if len(modified.Children) != 3 {
		t.Errorf("Modified should have 3 children")
	}
	if modified.Children[0] != child1 || modified.Children[1] != newChild || modified.Children[2] != child2 {
		t.Errorf("Child not inserted correctly")
	}

	// Insert at beginning
	atStart := original.InsertChildAt(0, newChild)
	if atStart.Children[0] != newChild {
		t.Errorf("Insert at start failed")
	}

	// Insert at end (clamped)
	atEnd := original.InsertChildAt(10, newChild)
	if atEnd.Children[len(atEnd.Children)-1] != newChild {
		t.Errorf("Insert at end failed")
	}
}

func TestCompositionPattern(t *testing.T) {
	// Test that you can build complex trees immutably
	child1 := (&Node{}).WithWidth(100).WithHeight(50)
	child2 := (&Node{}).WithWidth(150).WithHeight(75)

	container := (&Node{}).
		WithDisplay(DisplayFlex).
		WithPadding(16).
		AddChild(child1).
		AddChild(child2)

	// Verify structure
	if container.Style.Display != DisplayFlex {
		t.Errorf("Container display not set")
	}
	if len(container.Children) != 2 {
		t.Errorf("Container should have 2 children")
	}
	if container.Children[0].Style.Width != 100 {
		t.Errorf("Child properties not preserved")
	}
}

func TestSafeComposition(t *testing.T) {
	// Test that creating variants doesn't affect original
	base := (&Node{}).
		WithDisplay(DisplayFlex).
		WithWidth(200)

	variant1 := base.WithPadding(10)
	variant2 := base.WithPadding(20)

	// All should be independent
	if base.Style.Padding.Top != 0 {
		t.Errorf("Base was modified")
	}
	if variant1.Style.Padding.Top != 10 {
		t.Errorf("Variant1 incorrect")
	}
	if variant2.Style.Padding.Top != 20 {
		t.Errorf("Variant2 incorrect")
	}

	// They should all have the same width from base
	if variant1.Style.Width != 200 || variant2.Style.Width != 200 {
		t.Errorf("Base properties not inherited")
	}
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
