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
//
// Based on CSS specifications:
// - CSS Display Module Level 3: Display types and layout modes
// - CSS Flexbox Layout Module Level 1: Flex layout
// - CSS Grid Layout Module Level 1: Grid layout
// - CSS Box Model Module Level 3: Block layout
// - CSS Text Module Level 3: Text layout
//
// See:
// - https://www.w3.org/TR/css-display-3/
// - https://www.w3.org/TR/css-flexbox-1/
// - https://www.w3.org/TR/css-grid-1/
// - https://www.w3.org/TR/css-box-3/
// - https://www.w3.org/TR/css-text-3/
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
