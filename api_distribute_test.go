package layout

import (
	"math"
	"testing"
)

// TestDistributeNodesNoNodes verifies the function is a no-op when
// called with zero nodes. It must not panic or divide by zero.
func TestDistributeNodesNoNodes(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DistributeNodes panicked with empty slice: %v", r)
		}
	}()

	var nodes []*Node
	DistributeNodes(nodes, DistributeHorizontal)
	DistributeNodes(nodes, DistributeVertical)
}

// TestDistributeNodesSingleNode verifies the function does not mutate
// the position of a single node (avoids division-by-zero producing Inf).
func TestDistributeNodesSingleNode(t *testing.T) {
	node := &Node{Rect: Rect{X: 42, Y: 17, Width: 10, Height: 10}}

	DistributeNodes([]*Node{node}, DistributeHorizontal)
	if node.Rect.X != 42 || node.Rect.Y != 17 {
		t.Fatalf("single-node horizontal distribute mutated position: got X=%v Y=%v want X=42 Y=17",
			node.Rect.X, node.Rect.Y)
	}

	DistributeNodes([]*Node{node}, DistributeVertical)
	if node.Rect.X != 42 || node.Rect.Y != 17 {
		t.Fatalf("single-node vertical distribute mutated position: got X=%v Y=%v want X=42 Y=17",
			node.Rect.X, node.Rect.Y)
	}

	if math.IsInf(node.Rect.X, 0) || math.IsNaN(node.Rect.X) {
		t.Fatalf("single-node distribute produced non-finite X: %v", node.Rect.X)
	}
	if math.IsInf(node.Rect.Y, 0) || math.IsNaN(node.Rect.Y) {
		t.Fatalf("single-node distribute produced non-finite Y: %v", node.Rect.Y)
	}
}

// TestDistributeNodesTwoNodes verifies two nodes are unchanged
// (no middle nodes to redistribute; endpoints stay fixed).
func TestDistributeNodesTwoNodes(t *testing.T) {
	a := &Node{Rect: Rect{X: 0, Y: 0, Width: 10, Height: 10}}
	b := &Node{Rect: Rect{X: 100, Y: 0, Width: 10, Height: 10}}

	DistributeNodes([]*Node{a, b}, DistributeHorizontal)

	if a.Rect.X != 0 {
		t.Errorf("two-node distribute: first node X moved: got %v want 0", a.Rect.X)
	}
	if b.Rect.X != 100 {
		t.Errorf("two-node distribute: last node X moved: got %v want 100", b.Rect.X)
	}
}

// TestDistributeNodesThreeNodesHorizontal verifies three nodes are evenly
// spaced by center along the horizontal axis.
func TestDistributeNodesThreeNodesHorizontal(t *testing.T) {
	// Identical widths so center-based spacing maps to X spacing directly.
	a := &Node{Rect: Rect{X: 0, Y: 0, Width: 10, Height: 10}}
	mid := &Node{Rect: Rect{X: 30, Y: 0, Width: 10, Height: 10}}
	c := &Node{Rect: Rect{X: 100, Y: 0, Width: 10, Height: 10}}

	DistributeNodes([]*Node{a, mid, c}, DistributeHorizontal)

	// Endpoints must not move.
	if a.Rect.X != 0 {
		t.Errorf("first node X changed: got %v want 0", a.Rect.X)
	}
	if c.Rect.X != 100 {
		t.Errorf("last node X changed: got %v want 100", c.Rect.X)
	}

	// First center = 5, last center = 105, spacing = (105-5)/2 = 50.
	// Middle center = 55 -> X = 55 - 5 = 50.
	if math.Abs(mid.Rect.X-50) > 1e-9 {
		t.Errorf("middle node not centered: got X=%v want 50", mid.Rect.X)
	}
}

// TestDistributeNodesThreeNodesVertical verifies the vertical path.
func TestDistributeNodesThreeNodesVertical(t *testing.T) {
	a := &Node{Rect: Rect{X: 0, Y: 0, Width: 10, Height: 10}}
	mid := &Node{Rect: Rect{X: 0, Y: 30, Width: 10, Height: 10}}
	c := &Node{Rect: Rect{X: 0, Y: 100, Width: 10, Height: 10}}

	DistributeNodes([]*Node{a, mid, c}, DistributeVertical)

	if a.Rect.Y != 0 {
		t.Errorf("first node Y changed: got %v want 0", a.Rect.Y)
	}
	if c.Rect.Y != 100 {
		t.Errorf("last node Y changed: got %v want 100", c.Rect.Y)
	}
	if math.Abs(mid.Rect.Y-50) > 1e-9 {
		t.Errorf("middle node not centered: got Y=%v want 50", mid.Rect.Y)
	}
}
