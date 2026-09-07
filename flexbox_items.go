package layout

import (
	"math"
	"sort"
)

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

	// Content-box origin of the container: the static position of any
	// absolutely positioned child.
	containerFontSize := getCurrentFontSize(node, ctx)
	contentOriginX := ResolveLength(node.Style.Padding.Left, ctx, containerFontSize) +
		ResolveLength(node.Style.Border.Left, ctx, containerFontSize)
	contentOriginY := ResolveLength(node.Style.Padding.Top, ctx, containerFontSize) +
		ResolveLength(node.Style.Border.Top, ctx, containerFontSize)

	for _, child := range orderedChildren {
		// Skip display:none children
		if child.Style.Display == DisplayNone {
			continue
		}

		// CSS Flexbox §4.1: an absolutely positioned child of a flex container
		// does not participate in flex layout and is not a flex item. It is
		// still laid out for its own size, against the container's content
		// box, and left at its static position (the content-box start corner,
		// as if it were the sole flex item) so that the positioned pass
		// (LayoutWithPositioning) can place it.
		// https://www.w3.org/TR/css-flexbox-1/#abs-pos-items
		if isOutOfFlow(child) {
			flexboxLayoutAbsPosChild(child, setup, contentOriginX, contentOriginY, ctx)
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
			// Main axis is horizontal
			// Direction depends on whether progression is left-to-right or right-to-left
			if setup.writingMode.IsRightToLeft() {
				// vertical-rl: main axis progresses right-to-left
				childMainMarginStart = ResolveLength(child.Style.Margin.Right, ctx, childFontSize)
				childMainMarginEnd = ResolveLength(child.Style.Margin.Left, ctx, childFontSize)
			} else {
				// vertical-lr or horizontal-tb: main axis progresses left-to-right
				childMainMarginStart = ResolveLength(child.Style.Margin.Left, ctx, childFontSize)
				childMainMarginEnd = ResolveLength(child.Style.Margin.Right, ctx, childFontSize)
			}
			childCrossMarginStart = ResolveLength(child.Style.Margin.Top, ctx, childFontSize)
			childCrossMarginEnd = ResolveLength(child.Style.Margin.Bottom, ctx, childFontSize)
		} else {
			// Main axis is vertical (always top-to-bottom for now)
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

		// Measurement pass cross constraint. An item is only stretched to the
		// container's cross size when it is a single-line container, the item's
		// resolved alignment is stretch and its cross size is auto (§9.4 step 11).
		// Otherwise its hypothetical cross size is content-based, so a nested
		// container must be measured with an indefinite cross size or it would
		// treat the available height as definite and stretch its own line to it.
		// This only applies when the cross axis is vertical: a horizontal cross
		// axis (column direction) still bounds the width so inline content wraps.
		// https://www.w3.org/TR/css-flexbox-1/#algo-cross-item
		if setup.isMainHorizontal {
			itemAlign := node.Style.AlignItems
			if child.Style.AlignSelf != 0 {
				itemAlign = child.Style.AlignSelf
			}
			isMultiLine := node.Style.FlexWrap != FlexWrapNoWrap
			if itemAlign != AlignItemsStretch || flexIsSetLength(child.Style.Height) || isMultiLine {
				childCrossSize = Unbounded
			}
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

		// Measure child. Dispatch on the child's display type through Layout so
		// that text items are measured with LayoutText (a text flex item's flex
		// base size is its max-content size; CSS Flexbox §9.2 step 3E) instead
		// of the block algorithm, which has no notion of text and reports 0.
		// https://www.w3.org/TR/css-flexbox-1/#algo-main-item
		childSize := Layout(child, childConstraints, ctx)

		// Fall back to the explicit (set, non-negative) dimension when the
		// measured size is 0 or Unbounded, which happens when the child's own
		// layout algorithm cannot size it from the measurement constraints.
		if setup.isMainHorizontal {
			item.mainSize = childSize.Width
			item.crossSize = childSize.Height
			if (item.mainSize == 0 || item.mainSize >= Unbounded) && flexIsSetLength(child.Style.Width) {
				item.mainSize = ResolveLength(child.Style.Width, ctx, childFontSize)
			}
			if (item.crossSize == 0 || item.crossSize >= Unbounded) && flexIsSetLength(child.Style.Height) {
				item.crossSize = ResolveLength(child.Style.Height, ctx, childFontSize)
			}
		} else {
			item.mainSize = childSize.Height
			item.crossSize = childSize.Width
			if (item.mainSize == 0 || item.mainSize >= Unbounded) && flexIsSetLength(child.Style.Height) {
				item.mainSize = ResolveLength(child.Style.Height, ctx, childFontSize)
			}
			if (item.crossSize == 0 || item.crossSize >= Unbounded) && flexIsSetLength(child.Style.Width) {
				item.crossSize = ResolveLength(child.Style.Width, ctx, childFontSize)
			}
		}

		// Remember the size the child was actually laid out at so the parent can
		// re-run layout when flexing changes it (see LayoutFlexbox).
		item.measuredWidth = child.Rect.Width
		item.measuredHeight = child.Rect.Height

		// §9.4 step 11: align-self: stretch only applies when the item's cross
		// size is auto. Remember whether it is explicit; Px(0) is an explicit
		// zero cross size, not auto (see flexIsSetLength).
		// https://www.w3.org/TR/css-flexbox-1/#algo-stretch
		if setup.isMainHorizontal {
			item.hasExplicitCrossSize = flexIsSetLength(child.Style.Height)
		} else {
			item.hasExplicitCrossSize = flexIsSetLength(child.Style.Width)
		}

		// Resolve the item's min/max main size (§9.7 clamps to these).
		// An unset max (zero-value Length) means no maximum; the minimum always
		// wins over the maximum per CSS Sizing.
		// https://www.w3.org/TR/css-flexbox-1/#resolve-flexible-lengths
		minMain, maxMain := child.Style.MinWidth, child.Style.MaxWidth
		if !setup.isMainHorizontal {
			minMain, maxMain = child.Style.MinHeight, child.Style.MaxHeight
		}
		item.minMain = 0
		if minMain.Value > 0 {
			item.minMain = math.Max(0, ResolveLength(minMain, ctx, childFontSize))
		}
		item.maxMain = Unbounded
		if maxMain.Value > 0 {
			item.maxMain = ResolveLength(maxMain, ctx, childFontSize)
			if item.maxMain < item.minMain {
				item.maxMain = item.minMain
			}
		}

		// Flex factors. An invalid (NaN, infinite or negative) factor is treated
		// as unset; for flex-shrink the unset value is the initial value 1
		// (CSS Flexbox §7.2.2).
		item.flexGrow = flexSanitizeFactor(child.Style.FlexGrow)
		item.flexShrink = flexSanitizeFactor(child.Style.FlexShrink)
		if item.flexShrink == 0 {
			item.flexShrink = 1 // Default shrink factor
		}

		// §9.2 step 3: flex base size. A set flex-basis (including Px(0)) is
		// used directly; unset or negative means auto, which resolves to the
		// item's main size property when definite and to its content size
		// otherwise. item.mainSize already holds that value: the child was
		// measured above and an explicit main size property was substituted
		// when the measurement could not size it.
		// https://www.w3.org/TR/css-flexbox-1/#algo-main-item
		if flexIsSetLength(child.Style.FlexBasis) {
			item.flexBasis = ResolveLength(child.Style.FlexBasis, ctx, childFontSize)
		} else {
			item.flexBasis = item.mainSize
		}
		item.baseSize = item.flexBasis
		// §9.3 step 3: hypothetical main size, used for line breaking (§9.3 step 5).
		item.hypotheticalMainSize = clampFlexMainSize(item, item.baseSize)
		flexItems = append(flexItems, item)
	}

	return flexItems
}

// flexboxLayoutAbsPosChild lays out an absolutely positioned child of a flex
// container. Such a child is not a flex item (CSS Flexbox §4.1): it takes no
// main-axis slot and contributes nothing to the container's size. It is laid
// out against the container's content box so it gets its own size, and placed
// at its static position, which for a flex container is the content-box start
// corner (as though it were the sole flex item with the container's default
// alignment). The positioned pass (LayoutWithPositioning) applies any inset
// properties afterwards.
// https://www.w3.org/TR/css-flexbox-1/#abs-pos-items
func flexboxLayoutAbsPosChild(child *Node, setup flexboxSetup, contentOriginX, contentOriginY float64, ctx *LayoutContext) {
	Layout(child, Loose(setup.contentWidth, setup.contentHeight), ctx)
	child.Rect.X = contentOriginX
	child.Rect.Y = contentOriginY
}
