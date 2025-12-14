package layout

import (
	"math"
	"testing"
)

func TestAlignNodesLeft(t *testing.T) {
	// Create nodes with different X positions
	nodes := []*Node{
		{Rect: Rect{Width: 50, Height: 50}},
		{Rect: Rect{Width: 50, Height: 50}},
		{Rect: Rect{Width: 50, Height: 50}},
	}

	AlignNodes(nodes, AlignLeft)

	// All should be aligned to the leftmost (50)
	expectedX := 50.0
	for i, node := range nodes {
		if math.Abs(node.Rect.X-expectedX) > 0.01 {
			t.Errorf("Node %d: expected X=%.2f, got %.2f", i, expectedX, node.Rect.X)
		}
	}
}

func TestAlignNodesRight(t *testing.T) {
	// Create nodes with different X positions
	nodes := []*Node{
		{Rect: Rect{Width: 50, Height: 50}},
		{Rect: Rect{Width: 50, Height: 50}},
		{Rect: Rect{Width: 50, Height: 50}},
	}

	AlignNodes(nodes, AlignRight)

	// All should be aligned to the rightmost (250)
	expectedRight := 250.0
	for i, node := range nodes {
		actualRight := node.Rect.X + node.Rect.Width
		if math.Abs(actualRight-expectedRight) > 0.01 {
			t.Errorf("Node %d: expected right edge=%.2f, got %.2f", i, expectedRight, actualRight)
		}
	}
}

func TestAlignNodesTop(t *testing.T) {
	// Create nodes with different Y positions
	nodes := []*Node{
		{Rect: Rect{Width: 50, Height: 50}},
		{Rect: Rect{Width: 50, Height: 50}},
		{Rect: Rect{Width: 50, Height: 50}},
	}

	AlignNodes(nodes, AlignTop)

	// All should be aligned to the topmost (50)
	expectedY := 50.0
	for i, node := range nodes {
		if math.Abs(node.Rect.Y-expectedY) > 0.01 {
			t.Errorf("Node %d: expected Y=%.2f, got %.2f", i, expectedY, node.Rect.Y)
		}
	}
}

func TestAlignNodesBottom(t *testing.T) {
	// Create nodes with different Y positions
	nodes := []*Node{
		{Rect: Rect{Width: 50, Height: 50}},
		{Rect: Rect{Width: 50, Height: 50}},
		{Rect: Rect{Width: 50, Height: 50}},
	}

	AlignNodes(nodes, AlignBottom)

	// All should be aligned to the bottommost (250)
	expectedBottom := 250.0
	for i, node := range nodes {
		actualBottom := node.Rect.Y + node.Rect.Height
		if math.Abs(actualBottom-expectedBottom) > 0.01 {
			t.Errorf("Node %d: expected bottom edge=%.2f, got %.2f", i, expectedBottom, actualBottom)
		}
	}
}

func TestAlignNodesCenterX(t *testing.T) {
	// Create nodes with different X positions
	nodes := []*Node{
		{Rect: Rect{Width: 50, Height: 50}}, // center at 125
		{Rect: Rect{Width: 50, Height: 50}}, // center at 225
		{Rect: Rect{Width: 50, Height: 50}}, // center at 75
	}

	AlignNodes(nodes, AlignCenterX)

	// All should have the same center X (average of 125, 225, 75 = 141.67)
	expectedCenter := (125.0 + 225.0 + 75.0) / 3.0
	for i, node := range nodes {
		actualCenter := node.Rect.X + node.Rect.Width/2
		if math.Abs(actualCenter-expectedCenter) > 0.01 {
			t.Errorf("Node %d: expected center X=%.2f, got %.2f", i, expectedCenter, actualCenter)
		}
	}
}

func TestAlignNodesCenterY(t *testing.T) {
	// Create nodes with different Y positions
	nodes := []*Node{
		{Rect: Rect{Width: 50, Height: 50}}, // center at 125
		{Rect: Rect{Width: 50, Height: 50}}, // center at 225
		{Rect: Rect{Width: 50, Height: 50}}, // center at 75
	}

	AlignNodes(nodes, AlignCenterY)

	// All should have the same center Y (average of 125, 225, 75 = 141.67)
	expectedCenter := (125.0 + 225.0 + 75.0) / 3.0
	for i, node := range nodes {
		actualCenter := node.Rect.Y + node.Rect.Height/2
		if math.Abs(actualCenter-expectedCenter) > 0.01 {
			t.Errorf("Node %d: expected center Y=%.2f, got %.2f", i, expectedCenter, actualCenter)
		}
	}
}

func TestDistributeNodesHorizontal(t *testing.T) {
	// Create nodes at different X positions
	// nodes[0] at X=0 (leftmost, center at 25)
	// nodes[1] at X=200 (rightmost, center at 225)
	// nodes[2] at X=100 (middle, center at 125)
	nodes := []*Node{
		{Rect: Rect{Width: 50, Height: 50}}, // leftmost
		{Rect: Rect{Width: 50, Height: 50}}, // rightmost
		{Rect: Rect{Width: 50, Height: 50}}, // middle
	}

	DistributeNodes(nodes, DistributeHorizontal)

	// After distribution:
	// - Leftmost (nodes[0]) should stay at center 25
	// - Rightmost (nodes[1]) should stay at center 225
	// - Middle (nodes[2]) should be evenly spaced at center 125
	leftmostCenter := nodes[0].Rect.X + nodes[0].Rect.Width/2
	rightmostCenter := nodes[1].Rect.X + nodes[1].Rect.Width/2
	middleCenter := nodes[2].Rect.X + nodes[2].Rect.Width/2

	if math.Abs(leftmostCenter-25.0) > 0.01 {
		t.Errorf("Leftmost node center should be 25, got %.2f", leftmostCenter)
	}
	if math.Abs(rightmostCenter-225.0) > 0.01 {
		t.Errorf("Rightmost node center should be 225, got %.2f", rightmostCenter)
	}
	// Middle should be halfway between leftmost and rightmost
	expectedMiddle := (25.0 + 225.0) / 2.0
	if math.Abs(middleCenter-expectedMiddle) > 0.01 {
		t.Errorf("Middle node center should be %.2f, got %.2f", expectedMiddle, middleCenter)
	}
}

func TestDistributeNodesVertical(t *testing.T) {
	// Create nodes at different Y positions
	// nodes[0] at Y=0 (topmost, center at 25)
	// nodes[1] at Y=200 (bottommost, center at 225)
	// nodes[2] at Y=100 (middle, center at 125)
	nodes := []*Node{
		{Rect: Rect{Width: 50, Height: 50}}, // topmost
		{Rect: Rect{Width: 50, Height: 50}}, // bottommost
		{Rect: Rect{Width: 50, Height: 50}}, // middle
	}

	DistributeNodes(nodes, DistributeVertical)

	// After distribution:
	// - Topmost (nodes[0]) should stay at center 25
	// - Bottommost (nodes[1]) should stay at center 225
	// - Middle (nodes[2]) should be evenly spaced at center 125
	topmostCenter := nodes[0].Rect.Y + nodes[0].Rect.Height/2
	bottommostCenter := nodes[1].Rect.Y + nodes[1].Rect.Height/2
	middleCenter := nodes[2].Rect.Y + nodes[2].Rect.Height/2

	if math.Abs(topmostCenter-25.0) > 0.01 {
		t.Errorf("Topmost node center should be 25, got %.2f", topmostCenter)
	}
	if math.Abs(bottommostCenter-225.0) > 0.01 {
		t.Errorf("Bottommost node center should be 225, got %.2f", bottommostCenter)
	}
	// Middle should be halfway between topmost and bottommost
	expectedMiddle := (25.0 + 225.0) / 2.0
	if math.Abs(middleCenter-expectedMiddle) > 0.01 {
		t.Errorf("Middle node center should be %.2f, got %.2f", expectedMiddle, middleCenter)
	}
}

func TestDistributeNodesLessThanThree(t *testing.T) {
	// Distribution requires at least 3 nodes
	nodes1 := []*Node{
		{Rect: Rect{Width: 50, Height: 50}},
	}
	originalX := nodes1[0].Rect.X
	DistributeNodes(nodes1, DistributeHorizontal)
	if nodes1[0].Rect.X != originalX {
		t.Error("Single node should not be modified")
	}

	nodes2 := []*Node{
		{Rect: Rect{Width: 50, Height: 50}},
		{Rect: Rect{Width: 50, Height: 50}},
	}
	originalX2 := nodes2[0].Rect.X
	DistributeNodes(nodes2, DistributeHorizontal)
	if nodes2[0].Rect.X != originalX2 {
		t.Error("Two nodes should not be modified")
	}
}
