package layout

import (
	"math"
	"testing"
)

func TestSnapNodes(t *testing.T) {
	// Create nodes with positions that need snapping
	nodes := []*Node{
		{Rect: Rect{X: 12, Y: 23, Width: 50, Height: 50}},
		{Rect: Rect{X: 23, Y: 47, Width: 50, Height: 50}},
		{Rect: Rect{X: 34, Y: 56, Width: 50, Height: 50}},
	}

	SnapNodes(nodes, 10.0)

	// All positions should be multiples of 10
	for i, node := range nodes {
		if math.Mod(node.Rect.X, 10.0) > 0.01 {
			t.Errorf("Node %d: X=%.2f should be a multiple of 10", i, node.Rect.X)
		}
		if math.Mod(node.Rect.Y, 10.0) > 0.01 {
			t.Errorf("Node %d: Y=%.2f should be a multiple of 10", i, node.Rect.Y)
		}
	}

	// Check specific snapped values
	if math.Abs(nodes[0].Rect.X-10.0) > 0.01 {
		t.Errorf("Node 0: expected X=10.0, got %.2f", nodes[0].Rect.X)
	}
	if math.Abs(nodes[0].Rect.Y-20.0) > 0.01 {
		t.Errorf("Node 0: expected Y=20.0, got %.2f", nodes[0].Rect.Y)
	}
	if math.Abs(nodes[1].Rect.X-20.0) > 0.01 {
		t.Errorf("Node 1: expected X=20.0, got %.2f", nodes[1].Rect.X)
	}
	if math.Abs(nodes[1].Rect.Y-50.0) > 0.01 {
		t.Errorf("Node 1: expected Y=50.0, got %.2f", nodes[1].Rect.Y)
	}
}

func TestSnapNodesSmallGrid(t *testing.T) {
	// Test with a smaller grid (5px)
	nodes := []*Node{
		{Rect: Rect{X: 12, Y: 18, Width: 50, Height: 50}},
		{Rect: Rect{X: 23, Y: 47, Width: 50, Height: 50}},
	}

	SnapNodes(nodes, 5.0)

	// All positions should be multiples of 5
	for i, node := range nodes {
		if math.Mod(node.Rect.X, 5.0) > 0.01 {
			t.Errorf("Node %d: X=%.2f should be a multiple of 5", i, node.Rect.X)
		}
		if math.Mod(node.Rect.Y, 5.0) > 0.01 {
			t.Errorf("Node %d: Y=%.2f should be a multiple of 5", i, node.Rect.Y)
		}
	}

	// Check specific snapped values
	if math.Abs(nodes[0].Rect.X-10.0) > 0.01 {
		t.Errorf("Node 0: expected X=10.0, got %.2f", nodes[0].Rect.X)
	}
	if math.Abs(nodes[0].Rect.Y-20.0) > 0.01 {
		t.Errorf("Node 0: expected Y=20.0, got %.2f", nodes[0].Rect.Y)
	}
}

func TestSnapNodesInvalidSize(t *testing.T) {
	// Test with invalid snap sizes (should not modify nodes)
	nodes := []*Node{
		{Rect: Rect{X: 12, Y: 23, Width: 50, Height: 50}},
	}

	originalX := nodes[0].Rect.X
	originalY := nodes[0].Rect.Y

	SnapNodes(nodes, 0.0) // Invalid size
	if math.Abs(nodes[0].Rect.X-originalX) > 0.01 {
		t.Error("Node should not be modified with snapSize=0")
	}
	if math.Abs(nodes[0].Rect.Y-originalY) > 0.01 {
		t.Error("Node should not be modified with snapSize=0")
	}

	SnapNodes(nodes, -5.0) // Negative size
	if math.Abs(nodes[0].Rect.X-originalX) > 0.01 {
		t.Error("Node should not be modified with negative snapSize")
	}
	if math.Abs(nodes[0].Rect.Y-originalY) > 0.01 {
		t.Error("Node should not be modified with negative snapSize")
	}
}

func TestSnapToGrid(t *testing.T) {
	// Create nodes with positions relative to an offset grid
	nodes := []*Node{
		{Rect: Rect{X: 12.3, Y: 22.8, Width: 50, Height: 50}}, // 12.3, 22.8 relative to (5, 5) -> (15, 25)
		{Rect: Rect{X: 23.7, Y: 50.2, Width: 50, Height: 50}}, // 23.7, 50.2 relative to (5, 5) -> (25, 55)
	}

	SnapToGrid(nodes, 10.0, 5.0, 5.0)

	// Positions should be snapped relative to origin (5, 5)
	// Node 0: (12.3, 22.8) relative to (5, 5) -> (7.3, 17.8) -> round(0.73, 1.78) = (1, 2) -> (10, 20) -> (15, 25) absolute
	if math.Abs(nodes[0].Rect.X-15.0) > 0.01 {
		t.Errorf("Node 0: expected X=15.0, got %.2f", nodes[0].Rect.X)
	}
	if math.Abs(nodes[0].Rect.Y-25.0) > 0.01 {
		t.Errorf("Node 0: expected Y=25.0, got %.2f", nodes[0].Rect.Y)
	}

	// Node 1: (23.7, 45.2) relative to (5, 5) -> (18.7, 40.2) -> round(1.87, 4.02) = (2, 4) -> (20, 40) -> (25, 45) absolute
	// Actually, let's recalculate: 45.2-5=40.2, round(40.2/10)=4, 4*10+5=45, but test expects 55
	// For Y=55: 55-5=50, round(50/10)=5, so we need Y such that Y-5 rounds to 50
	// So Y should be between 45 and 55, e.g., Y=50.2 gives 45.2 relative, round(4.52)=5, 5*10+5=55
	if math.Abs(nodes[1].Rect.X-25.0) > 0.01 {
		t.Errorf("Node 1: expected X=25.0, got %.2f", nodes[1].Rect.X)
	}
	if math.Abs(nodes[1].Rect.Y-55.0) > 0.01 {
		t.Errorf("Node 1: expected Y=55.0, got %.2f", nodes[1].Rect.Y)
	}
}

func TestSnapToGridZeroOrigin(t *testing.T) {
	// Test with zero origin (should behave like SnapNodes)
	nodes := []*Node{
		{Rect: Rect{X: 12, Y: 23, Width: 50, Height: 50}},
	}

	SnapToGrid(nodes, 10.0, 0.0, 0.0)

	if math.Abs(nodes[0].Rect.X-10.0) > 0.01 {
		t.Errorf("Expected X=10.0, got %.2f", nodes[0].Rect.X)
	}
	if math.Abs(nodes[0].Rect.Y-20.0) > 0.01 {
		t.Errorf("Expected Y=20.0, got %.2f", nodes[0].Rect.Y)
	}
}

func TestSnapToGridInvalidSize(t *testing.T) {
	// Test with invalid snap sizes (should not modify nodes)
	nodes := []*Node{
		{Rect: Rect{X: 12, Y: 23, Width: 50, Height: 50}},
	}

	originalX := nodes[0].Rect.X
	originalY := nodes[0].Rect.Y

	SnapToGrid(nodes, 0.0, 5.0, 5.0) // Invalid size
	if math.Abs(nodes[0].Rect.X-originalX) > 0.01 {
		t.Error("Node should not be modified with snapSize=0")
	}
	if math.Abs(nodes[0].Rect.Y-originalY) > 0.01 {
		t.Error("Node should not be modified with snapSize=0")
	}
}

func TestSnapNodesEmptyList(t *testing.T) {
	// Should not panic with empty list
	nodes := []*Node{}
	SnapNodes(nodes, 10.0)
	// If we get here without panicking, test passes
}

func TestSnapNodesAlreadyOnGrid(t *testing.T) {
	// Nodes already on grid should stay in place
	nodes := []*Node{
		{Rect: Rect{X: 10, Y: 20, Width: 50, Height: 50}},
		{Rect: Rect{X: 30, Y: 40, Width: 50, Height: 50}},
	}

	SnapNodes(nodes, 10.0)

	if math.Abs(nodes[0].Rect.X-10.0) > 0.01 {
		t.Errorf("Node already on grid should not move: expected X=10.0, got %.2f", nodes[0].Rect.X)
	}
	if math.Abs(nodes[0].Rect.Y-20.0) > 0.01 {
		t.Errorf("Node already on grid should not move: expected Y=20.0, got %.2f", nodes[0].Rect.Y)
	}
	if math.Abs(nodes[1].Rect.X-30.0) > 0.01 {
		t.Errorf("Node already on grid should not move: expected X=30.0, got %.2f", nodes[1].Rect.X)
	}
	if math.Abs(nodes[1].Rect.Y-40.0) > 0.01 {
		t.Errorf("Node already on grid should not move: expected Y=40.0, got %.2f", nodes[1].Rect.Y)
	}
}

func TestSnapNodesBoundaryConditions(t *testing.T) {
	// Test nodes exactly halfway between grid points (should round)
	nodes := []*Node{
		{Rect: Rect{X: 15.0, Y: 0, Width: 50, Height: 50}}, // Exactly halfway
		{Rect: Rect{X: 14.9, Y: 0, Width: 50, Height: 50}}, // Just below/above halfway
		{Rect: Rect{X: 15.1, Y: 0, Width: 50, Height: 50}}, // Just above/below halfway
	}

	SnapNodes(nodes, 10.0)

	// 15.0 should round to 20.0 (round half up)
	if math.Abs(nodes[0].Rect.X-20.0) > 0.01 {
		t.Errorf("15.0 should round to 20.0, got %.2f", nodes[0].Rect.X)
	}
	// 14.9 should round to 10.0
	if math.Abs(nodes[1].Rect.X-10.0) > 0.01 {
		t.Errorf("14.9 should round to 10.0, got %.2f", nodes[1].Rect.X)
	}
	// 15.1 should round to 20.0
	if math.Abs(nodes[2].Rect.X-20.0) > 0.01 {
		t.Errorf("15.1 should round to 20.0, got %.2f", nodes[2].Rect.X)
	}
}

func TestSnapNodesNegativePositions(t *testing.T) {
	// Test snapping negative positions
	nodes := []*Node{
		{Rect: Rect{X: -12.3, Y: -17.8, Width: 50, Height: 50}},
		{Rect: Rect{X: -5.0, Y: -10.0, Width: 50, Height: 50}},
	}

	SnapNodes(nodes, 10.0)

	// -12.3 should snap to -10.0
	if math.Abs(nodes[0].Rect.X-(-10.0)) > 0.01 {
		t.Errorf("Expected X=-10.0, got %.2f", nodes[0].Rect.X)
	}
	// -17.8 should snap to -20.0
	if math.Abs(nodes[0].Rect.Y-(-20.0)) > 0.01 {
		t.Errorf("Expected Y=-20.0, got %.2f", nodes[0].Rect.Y)
	}
	// -5.0 should snap to -10.0 (rounds to nearest multiple: -5/10 = -0.5 -> -1 -> -10)
	if math.Abs(nodes[1].Rect.X-(-10.0)) > 0.01 {
		t.Errorf("Expected X=-10.0, got %.2f", nodes[1].Rect.X)
	}
	// -10.0 is already on grid
	if math.Abs(nodes[1].Rect.Y-(-10.0)) > 0.01 {
		t.Errorf("Expected Y=-10.0, got %.2f", nodes[1].Rect.Y)
	}
}

func TestSnapNodesVerySmallGrid(t *testing.T) {
	// Test with very small grid (0.1px)
	nodes := []*Node{
		{Rect: Rect{X: 12.34, Y: 23.45, Width: 50, Height: 50}},
	}

	SnapNodes(nodes, 0.1)

	// Should snap to 0.1px precision
	if math.Mod(nodes[0].Rect.X, 0.1) > 0.001 {
		t.Errorf("X=%.3f should be a multiple of 0.1", nodes[0].Rect.X)
	}
	if math.Mod(nodes[0].Rect.Y, 0.1) > 0.001 {
		t.Errorf("Y=%.3f should be a multiple of 0.1", nodes[0].Rect.Y)
	}
}

func TestSnapNodesVeryLargeGrid(t *testing.T) {
	// Test with very large grid (100px)
	nodes := []*Node{
		{Rect: Rect{X: 123, Y: 178, Width: 50, Height: 50}},
		{Rect: Rect{X: 45, Y: 67, Width: 50, Height: 50}},
	}

	SnapNodes(nodes, 100.0)

	// 123 should snap to 100
	if math.Abs(nodes[0].Rect.X-100.0) > 0.01 {
		t.Errorf("Expected X=100.0, got %.2f", nodes[0].Rect.X)
	}
	// 178 should snap to 200
	if math.Abs(nodes[0].Rect.Y-200.0) > 0.01 {
		t.Errorf("Expected Y=200.0, got %.2f", nodes[0].Rect.Y)
	}
	// 45 should snap to 0
	if math.Abs(nodes[1].Rect.X-0.0) > 0.01 {
		t.Errorf("Expected X=0.0, got %.2f", nodes[1].Rect.X)
	}
	// 67 should snap to 100
	if math.Abs(nodes[1].Rect.Y-100.0) > 0.01 {
		t.Errorf("Expected Y=100.0, got %.2f", nodes[1].Rect.Y)
	}
}

func TestSnapNodesIdempotency(t *testing.T) {
	// Snapping twice should produce the same result
	nodes := []*Node{
		{Rect: Rect{X: 12, Y: 23, Width: 50, Height: 50}},
		{Rect: Rect{X: 34, Y: 56, Width: 50, Height: 50}},
	}

	SnapNodes(nodes, 10.0)
	firstX := nodes[0].Rect.X
	firstY := nodes[0].Rect.Y

	SnapNodes(nodes, 10.0)
	secondX := nodes[0].Rect.X
	secondY := nodes[0].Rect.Y

	if math.Abs(firstX-secondX) > 0.01 {
		t.Errorf("Snapping twice should be idempotent: first=%.2f, second=%.2f", firstX, secondX)
	}
	if math.Abs(firstY-secondY) > 0.01 {
		t.Errorf("Snapping twice should be idempotent: first=%.2f, second=%.2f", firstY, secondY)
	}
}

func TestSnapNodesMultipleNodes(t *testing.T) {
	// Test with many nodes
	nodes := make([]*Node, 20)
	for i := 0; i < 20; i++ {
		nodes[i] = &Node{
			Rect: Rect{
				X:      float64(i)*7.3 + 1.5,
				Y:      float64(i)*11.7 + 2.3,
				Width:  50,
				Height: 50,
			},
		}
	}

	SnapNodes(nodes, 10.0)

	// All positions should be multiples of 10
	for i, node := range nodes {
		if math.Mod(node.Rect.X, 10.0) > 0.01 {
			t.Errorf("Node %d: X=%.2f should be a multiple of 10", i, node.Rect.X)
		}
		if math.Mod(node.Rect.Y, 10.0) > 0.01 {
			t.Errorf("Node %d: Y=%.2f should be a multiple of 10", i, node.Rect.Y)
		}
	}
}

func TestSnapToGridNegativeOrigin(t *testing.T) {
	// Test with negative origin
	nodes := []*Node{
		{Rect: Rect{X: -7.3, Y: -12.8, Width: 50, Height: 50}},
	}

	SnapToGrid(nodes, 10.0, -5.0, -5.0)

	// Relative to (-5, -5): (-7.3, -12.8) -> (-2.3, -7.8) -> (-0, -10) -> (-5, -15) absolute
	// Actually, let's calculate: -7.3 - (-5) = -2.3, round to 0, add -5 = -5
	if math.Abs(nodes[0].Rect.X-(-5.0)) > 0.01 {
		t.Errorf("Expected X=-5.0, got %.2f", nodes[0].Rect.X)
	}
	// -12.8 - (-5) = -7.8, round to -10, add -5 = -15
	if math.Abs(nodes[0].Rect.Y-(-15.0)) > 0.01 {
		t.Errorf("Expected Y=-15.0, got %.2f", nodes[0].Rect.Y)
	}
}

func TestSnapToGridLargeOrigin(t *testing.T) {
	// Test with large origin offset
	nodes := []*Node{
		{Rect: Rect{X: 1007.3, Y: 2017.8, Width: 50, Height: 50}},
	}

	SnapToGrid(nodes, 10.0, 1000.0, 2000.0)

	// Relative to (1000, 2000): (1007.3, 2017.8) -> (7.3, 17.8) -> (10, 20) -> (1010, 2020) absolute
	if math.Abs(nodes[0].Rect.X-1010.0) > 0.01 {
		t.Errorf("Expected X=1010.0, got %.2f", nodes[0].Rect.X)
	}
	if math.Abs(nodes[0].Rect.Y-2020.0) > 0.01 {
		t.Errorf("Expected Y=2020.0, got %.2f", nodes[0].Rect.Y)
	}
}

func TestSnapToGridIdempotency(t *testing.T) {
	// Snapping twice should produce the same result
	nodes := []*Node{
		{Rect: Rect{X: 12.3, Y: 17.8, Width: 50, Height: 50}},
	}

	SnapToGrid(nodes, 10.0, 5.0, 5.0)
	firstX := nodes[0].Rect.X
	firstY := nodes[0].Rect.Y

	SnapToGrid(nodes, 10.0, 5.0, 5.0)
	secondX := nodes[0].Rect.X
	secondY := nodes[0].Rect.Y

	if math.Abs(firstX-secondX) > 0.01 {
		t.Errorf("Snapping twice should be idempotent: first=%.2f, second=%.2f", firstX, secondX)
	}
	if math.Abs(firstY-secondY) > 0.01 {
		t.Errorf("Snapping twice should be idempotent: first=%.2f, second=%.2f", firstY, secondY)
	}
}

func TestSnapToGridEmptyList(t *testing.T) {
	// Should not panic with empty list
	nodes := []*Node{}
	SnapToGrid(nodes, 10.0, 5.0, 5.0)
	// If we get here without panicking, test passes
}

func TestSnapNodesZeroPosition(t *testing.T) {
	// Test nodes at origin
	nodes := []*Node{
		{Rect: Rect{X: 0.0, Y: 0.0, Width: 50, Height: 50}},
		{Rect: Rect{X: 0.1, Y: 0.0, Width: 50, Height: 50}},
	}

	SnapNodes(nodes, 10.0)

	// 0.0 should stay at 0.0
	if math.Abs(nodes[0].Rect.X-0.0) > 0.01 {
		t.Errorf("Expected X=0.0, got %.2f", nodes[0].Rect.X)
	}
	// 0.1 should snap to 0.0
	if math.Abs(nodes[1].Rect.X-0.0) > 0.01 {
		t.Errorf("Expected X=0.0, got %.2f", nodes[1].Rect.X)
	}
}

func TestSnapToGridAlreadyOnGrid(t *testing.T) {
	// Nodes already on grid should stay in place
	nodes := []*Node{
		{Rect: Rect{X: 15.0, Y: 25.0, Width: 50, Height: 50}}, // On grid relative to (5, 5)
	}

	SnapToGrid(nodes, 10.0, 5.0, 5.0)

	if math.Abs(nodes[0].Rect.X-15.0) > 0.01 {
		t.Errorf("Node already on grid should not move: expected X=15.0, got %.2f", nodes[0].Rect.X)
	}
	if math.Abs(nodes[0].Rect.Y-25.0) > 0.01 {
		t.Errorf("Node already on grid should not move: expected Y=25.0, got %.2f", nodes[0].Rect.Y)
	}
}
