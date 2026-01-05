package layout

import "sort"

// flexboxMeasureItems measures all children and creates flex items.
//
// Algorithm based on CSS Flexible Box Layout Module Level 1:
// - §9.2: Line Length Determination (initial measurement phase)
// - §5.4.1: Reordering with the order property
//
// See: https://www.w3.org/TR/css-flexbox-1/#line-sizing
// See: https://www.w3.org/TR/css-flexbox-1/#order-property
func flexboxMeasureItems(node *Node, setup flexboxSetup, ctx *LayoutContext) []*flexItem {
	children := node.Children

	// Sort children by order property (CSS Flexbox §5.4.1)
	// Items with the same order value appear in source order
	orderedChildren := make([]*Node, len(children))
	copy(orderedChildren, children)
	sort.SliceStable(orderedChildren, func(i, j int) bool {
		return orderedChildren[i].Style.Order < orderedChildren[j].Style.Order
	})

	flexItems := make([]*flexItem, 0, len(orderedChildren))

	for _, child := range orderedChildren {
		// Skip display:none children
		if child.Style.Display == DisplayNone {
			continue
		}
		item := &flexItem{
			node: child,
		}

		// Get current font size for child's Length resolution
		childFontSize := getCurrentFontSize(child, ctx)

		// Get child margins (resolve Length to pixels)
		var childMainMarginStart, childMainMarginEnd, childCrossMarginStart, childCrossMarginEnd float64
		if setup.isMainHorizontal {
			childMainMarginStart = ResolveLength(child.Style.Margin.Left, ctx, childFontSize)
			childMainMarginEnd = ResolveLength(child.Style.Margin.Right, ctx, childFontSize)
			childCrossMarginStart = ResolveLength(child.Style.Margin.Top, ctx, childFontSize)
			childCrossMarginEnd = ResolveLength(child.Style.Margin.Bottom, ctx, childFontSize)
		} else {
			childMainMarginStart = ResolveLength(child.Style.Margin.Top, ctx, childFontSize)
			childMainMarginEnd = ResolveLength(child.Style.Margin.Bottom, ctx, childFontSize)
			childCrossMarginStart = ResolveLength(child.Style.Margin.Left, ctx, childFontSize)
			childCrossMarginEnd = ResolveLength(child.Style.Margin.Right, ctx, childFontSize)
		}
		item.mainMarginStart = childMainMarginStart
		item.mainMarginEnd = childMainMarginEnd
		item.crossMarginStart = childCrossMarginStart
		item.crossMarginEnd = childCrossMarginEnd

		// Determine child constraints (account for margins)
		childMainSize := setup.mainSize
		childCrossSize := setup.crossSize
		if node.Style.FlexWrap == FlexWrapNoWrap {
			// In nowrap, children share main axis space
			childMainSize = Unbounded
		}

		childConstraints := Constraints{
			MinWidth:  0,
			MaxWidth:  childMainSize,
			MinHeight: 0,
			MaxHeight: childCrossSize,
		}
		if !setup.isMainHorizontal {
			childConstraints.MaxWidth, childConstraints.MaxHeight = childConstraints.MaxHeight, childConstraints.MaxWidth
		}

		// Measure child
		var childSize Size
		if child.Style.Display == DisplayFlex {
			childSize = LayoutFlexbox(child, childConstraints, ctx)
		} else if child.Style.Display == DisplayGrid {
			childSize = LayoutGrid(child, childConstraints, ctx)
		} else {
			childSize = LayoutBlock(child, childConstraints, ctx)
		}

		if setup.isMainHorizontal {
			item.mainSize = childSize.Width
			item.crossSize = childSize.Height
			// Use explicit dimensions if measured size is 0 or Unbounded
			// This handles cases where LayoutBlock returns 0 or Unbounded for items with explicit dimensions
			if (item.mainSize == 0 || item.mainSize >= Unbounded) && child.Style.Width.Value >= 0 {
				item.mainSize = ResolveLength(child.Style.Width, ctx, childFontSize)
			}
			if (item.crossSize == 0 || item.crossSize >= Unbounded) && child.Style.Height.Value >= 0 {
				item.crossSize = ResolveLength(child.Style.Height, ctx, childFontSize)
			}
		} else {
			item.mainSize = childSize.Height
			item.crossSize = childSize.Width
			// Use explicit dimensions if measured size is 0 or Unbounded
			if (item.mainSize == 0 || item.mainSize >= Unbounded) && child.Style.Height.Value >= 0 {
				item.mainSize = ResolveLength(child.Style.Height, ctx, childFontSize)
			}
			if (item.crossSize == 0 || item.crossSize >= Unbounded) && child.Style.Width.Value >= 0 {
				item.crossSize = ResolveLength(child.Style.Width, ctx, childFontSize)
			}
		}

		// Store the measured size as a fallback
		measuredMainSize := item.mainSize

		// Get flex properties
		item.flexGrow = child.Style.FlexGrow
		if item.flexGrow == 0 {
			item.flexGrow = 0
		}
		item.flexShrink = child.Style.FlexShrink
		if item.flexShrink == 0 {
			item.flexShrink = 1 // Default shrink factor
		}
		item.flexBasis = ResolveLength(child.Style.FlexBasis, ctx, childFontSize)
		if item.flexBasis < 0 {
			item.flexBasis = item.mainSize // auto means use main size
		}

		item.baseSize = item.flexBasis

		// Ensure baseSize is never 0 if we have a measured size or explicit width/height
		if item.baseSize == 0 {
			if measuredMainSize > 0 {
				item.baseSize = measuredMainSize
				item.flexBasis = measuredMainSize
			} else if setup.isMainHorizontal && child.Style.Width.Value >= 0 {
				// Use explicit width for baseSize
				resolvedWidth := ResolveLength(child.Style.Width, ctx, childFontSize)
				item.baseSize = resolvedWidth
				item.flexBasis = resolvedWidth
			} else if !setup.isMainHorizontal && child.Style.Height.Value >= 0 {
				// Use explicit height for baseSize
				resolvedHeight := ResolveLength(child.Style.Height, ctx, childFontSize)
				item.baseSize = resolvedHeight
				item.flexBasis = resolvedHeight
			}
		}
		flexItems = append(flexItems, item)
	}

	return flexItems
}
