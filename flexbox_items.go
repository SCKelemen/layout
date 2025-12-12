package layout

// flexboxMeasureItems measures all children and creates flex items.
//
// Algorithm based on CSS Flexible Box Layout Module Level 1:
// - §9.2: Line Length Determination (initial measurement phase)
//
// See: https://www.w3.org/TR/css-flexbox-1/#line-sizing
func flexboxMeasureItems(node *Node, setup flexboxSetup) []*flexItem {
	children := node.Children
	flexItems := make([]*flexItem, 0, len(children))

	for _, child := range children {
		// Skip display:none children
		if child.Style.Display == DisplayNone {
			continue
		}
		item := &flexItem{
			node: child,
		}

		// Get child margins
		var childMainMarginStart, childMainMarginEnd, childCrossMarginStart, childCrossMarginEnd float64
		if setup.isRow {
			childMainMarginStart = child.Style.Margin.Left
			childMainMarginEnd = child.Style.Margin.Right
			childCrossMarginStart = child.Style.Margin.Top
			childCrossMarginEnd = child.Style.Margin.Bottom
		} else {
			childMainMarginStart = child.Style.Margin.Top
			childMainMarginEnd = child.Style.Margin.Bottom
			childCrossMarginStart = child.Style.Margin.Left
			childCrossMarginEnd = child.Style.Margin.Right
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
		if !setup.isRow {
			childConstraints.MaxWidth, childConstraints.MaxHeight = childConstraints.MaxHeight, childConstraints.MaxWidth
		}

		// Measure child
		var childSize Size
		if child.Style.Display == DisplayFlex {
			childSize = LayoutFlexbox(child, childConstraints)
		} else if child.Style.Display == DisplayGrid {
			childSize = LayoutGrid(child, childConstraints)
		} else {
			childSize = LayoutBlock(child, childConstraints)
		}

		if setup.isRow {
			item.mainSize = childSize.Width
			item.crossSize = childSize.Height
			// Use explicit dimensions if measured size is 0 or Unbounded
			// This handles cases where LayoutBlock returns 0 or Unbounded for items with explicit dimensions
			if (item.mainSize == 0 || item.mainSize >= Unbounded) && child.Style.Width >= 0 {
				item.mainSize = child.Style.Width
			}
			if (item.crossSize == 0 || item.crossSize >= Unbounded) && child.Style.Height >= 0 {
				item.crossSize = child.Style.Height
			}
		} else {
			item.mainSize = childSize.Height
			item.crossSize = childSize.Width
			// Use explicit dimensions if measured size is 0 or Unbounded
			if (item.mainSize == 0 || item.mainSize >= Unbounded) && child.Style.Height >= 0 {
				item.mainSize = child.Style.Height
			}
			if (item.crossSize == 0 || item.crossSize >= Unbounded) && child.Style.Width >= 0 {
				item.crossSize = child.Style.Width
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
		item.flexBasis = child.Style.FlexBasis
		if item.flexBasis < 0 {
			item.flexBasis = item.mainSize // auto means use main size
		}

		item.baseSize = item.flexBasis

		// Ensure baseSize is never 0 if we have a measured size or explicit width/height
		if item.baseSize == 0 {
			if measuredMainSize > 0 {
				item.baseSize = measuredMainSize
				item.flexBasis = measuredMainSize
			} else if setup.isRow && child.Style.Width >= 0 {
				// Use explicit width for baseSize
				item.baseSize = child.Style.Width
				item.flexBasis = child.Style.Width
			} else if !setup.isRow && child.Style.Height >= 0 {
				// Use explicit height for baseSize
				item.baseSize = child.Style.Height
				item.flexBasis = child.Style.Height
			}
		}
		flexItems = append(flexItems, item)
	}

	return flexItems
}
