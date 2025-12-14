package layout

import (
	"testing"
)

func TestGetSVGTransform(t *testing.T) {
	// Test with identity transform
	node := &Node{
		Style: Style{
			Transform: IdentityTransform(),
		},
	}

	result := GetSVGTransform(node)
	if result != "" {
		t.Errorf("Expected empty string for identity transform, got %q", result)
	}

	// Test with actual transform
	node.Style.Transform = Translate(10, 20)
	result = GetSVGTransform(node)
	expected := "matrix(1,0,0,1,10,20)"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGetFinalRect(t *testing.T) {
	// Test without transform
	node := &Node{
		Rect: Rect{Width: 100, Height: 50},
		Style: Style{
			Transform: IdentityTransform(),
		},
	}

	result := GetFinalRect(node)
	if result.X != 10 || result.Y != 20 || result.Width != 100 || result.Height != 50 {
		t.Errorf("Expected rect (10, 20, 100, 50), got (%.2f, %.2f, %.2f, %.2f)",
			result.X, result.Y, result.Width, result.Height)
	}

	// Test with transform
	node.Style.Transform = Translate(5, 10)
	result = GetFinalRect(node)
	// After translation, position should change but size should remain
	if result.X != 15 || result.Y != 30 {
		t.Errorf("Expected translated position (15, 30), got (%.2f, %.2f)", result.X, result.Y)
	}
}

func TestCollectNodesForSVG(t *testing.T) {
	root := &Node{
		Children: []*Node{
			{
				Children: []*Node{
					{},
				},
			},
			{},
		},
	}

	var nodes []*Node
	CollectNodesForSVG(root, &nodes)

	// Should collect root + 2 children + 1 grandchild = 4 nodes
	if len(nodes) != 4 {
		t.Errorf("Expected 4 nodes, got %d", len(nodes))
	}

	// First should be root
	if nodes[0] != root {
		t.Error("First node should be root")
	}
}
