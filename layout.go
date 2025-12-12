package layout

// Layout performs layout on a node tree based on display type.
// It routes to the appropriate layout algorithm (Flexbox, Grid, or Block)
// based on the root node's Display property.
//
// Note: This performs normal flow layout only. For positioned elements
// (absolute, relative, fixed, sticky), use LayoutWithPositioning instead.
//
// After calling Layout, each node's Rect field will contain the computed
// position and size.
func Layout(root *Node, constraints Constraints) Size {
	switch root.Style.Display {
	case DisplayFlex:
		return LayoutFlexbox(root, constraints)
	case DisplayGrid:
		return LayoutGrid(root, constraints)
	case DisplayInlineText:
		return LayoutText(root, constraints)
	case DisplayNone:
		return Size{Width: 0, Height: 0}
	default:
		return LayoutBlock(root, constraints)
	}
}
