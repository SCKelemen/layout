package layout

// Layout performs layout on a node tree based on display type.
// It routes to the appropriate layout algorithm (Flexbox, Grid, or Block)
// based on the root node's Display property.
//
// The LayoutContext parameter provides information needed to resolve relative
// length units (em, rem, ch, vh, vw) to pixels. Create a context using
// NewLayoutContext with your viewport size and root font size.
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
// - CSS Values and Units Module Level 4: Length units
//
// See:
// - https://www.w3.org/TR/css-display-3/
// - https://www.w3.org/TR/css-flexbox-1/
// - https://www.w3.org/TR/css-grid-1/
// - https://www.w3.org/TR/css-box-3/
// - https://www.w3.org/TR/css-text-3/
// - https://www.w3.org/TR/css-values-4/
func Layout(root *Node, constraints Constraints, ctx *LayoutContext) Size {
	switch root.Style.Display {
	case DisplayFlex:
		return LayoutFlexbox(root, constraints, ctx)
	case DisplayGrid:
		return LayoutGrid(root, constraints, ctx)
	case DisplayInlineText:
		return LayoutText(root, constraints, ctx)
	case DisplayNone:
		return Size{Width: 0, Height: 0}
	default:
		return LayoutBlock(root, constraints, ctx)
	}
}

// LayoutSimple performs layout with a default context.
// This is a convenience wrapper for simple use cases that don't need custom
// viewport or font configuration.
//
// The default context uses:
// - Viewport size from constraints
// - Root font size of 16 points
// - Default text metrics provider
// - Reference character '0' for ch units
//
// For more control over unit resolution, use Layout with a custom LayoutContext.
func LayoutSimple(root *Node, constraints Constraints) Size {
	ctx := NewLayoutContext(
		constraints.MaxWidth,
		constraints.MaxHeight,
		16.0, // default root font size
	)
	return Layout(root, constraints, ctx)
}
