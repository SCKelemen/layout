package layout

// GetSVGTransform returns the SVG transform attribute string for a node
// This is useful when rendering layouts to SVG
func GetSVGTransform(node *Node) string {
	if node.Style.Transform.IsIdentity() {
		return ""
	}
	return node.Style.Transform.ToSVGString()
}

// GetFinalRect returns the final rectangle position after applying transforms
// This accounts for both positioning and transforms
func GetFinalRect(node *Node) Rect {
	rect := node.Rect
	
	// If there's a transform, apply it to get the bounding box
	if !node.Style.Transform.IsIdentity() {
		// For layout purposes, we might want the original rect
		// But for rendering, we want the transformed bounding box
		return node.Style.Transform.ApplyToRect(rect)
	}
	
	return rect
}

// CollectNodesForSVG collects all nodes in the tree with their final positions
// Useful for iterating over all elements when rendering to SVG
func CollectNodesForSVG(root *Node, nodes *[]*Node) {
	*nodes = append(*nodes, root)
	for _, child := range root.Children {
		CollectNodesForSVG(child, nodes)
	}
}

