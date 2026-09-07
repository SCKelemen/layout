package layout

// GetSVGTransform returns the SVG transform attribute string for a node, or
// "" when the node has no transform.
//
// Deprecated: use node.Style.Transform.ToSVGString(), which has the same
// behavior (it also returns "" for the identity). Kept for compatibility.
func GetSVGTransform(node *Node) string {
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

// CollectNodesForSVG appends root and all of its descendants, in depth-first
// order, to *nodes.
//
// Deprecated: use root.DescendantsAndSelf(), which returns the same nodes in
// the same order without the out-parameter. Kept for compatibility.
func CollectNodesForSVG(root *Node, nodes *[]*Node) {
	*nodes = append(*nodes, root.DescendantsAndSelf()...)
}
