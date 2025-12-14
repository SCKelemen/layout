package layout

// LayoutPositioned handles positioned elements (absolute, relative, fixed, sticky).
// This should be called after the normal layout flow to position elements.
//
// Algorithm based on CSS Positioned Layout Module Level 3:
// - §2: Positioning Schemes
// - §3: Choosing a positioning scheme (position property)
// - §5: Absolute positioning
// - §6: Fixed positioning
// - §7: Sticky positioning
//
// See: https://www.w3.org/TR/css-position-3/
func LayoutPositioned(node *Node, parentRect Rect, viewportRect Rect, ctx *LayoutContext) {
	if node.Style.Position == PositionStatic {
		// Static positioning is the default, no special handling needed
		return
	}

	// Get current font size for em resolution
	currentFontSize := getCurrentFontSize(node, ctx)

	// Calculate the positioning context
	var positioningContext Rect
	switch node.Style.Position {
	case PositionFixed:
		// Fixed is relative to viewport
		positioningContext = viewportRect
	case PositionAbsolute, PositionRelative, PositionSticky:
		// Absolute/relative/sticky are relative to nearest positioned ancestor
		// For now, we'll use parentRect (in a full implementation, we'd traverse up)
		positioningContext = parentRect
	}

	// Calculate offsets
	// Treat < 0 as auto, >= 0 as explicit value
	// Note: Since the zero value of float64 is 0, we need a heuristic:
	// If a side is 0 and the opposite side is set, treat 0 as auto (unset)
	// This allows users to set only Right/Bottom without explicitly setting Left/Top to -1
	left := node.Style.Left
	right := node.Style.Right
	top := node.Style.Top
	bottom := node.Style.Bottom

	// Resolve Length values to pixels
	leftPx := ResolveLength(left, ctx, currentFontSize)
	rightPx := ResolveLength(right, ctx, currentFontSize)
	topPx := ResolveLength(top, ctx, currentFontSize)
	bottomPx := ResolveLength(bottom, ctx, currentFontSize)

	// Determine if sides are explicitly set
	// If a side is 0 and the opposite is set, treat 0 as auto (unset)
	hasLeft := leftPx >= 0
	hasRight := rightPx >= 0
	if leftPx == 0 && hasRight {
		hasLeft = false
		leftPx = -1
	}
	if rightPx == 0 && hasLeft {
		hasRight = false
		rightPx = -1
	}

	hasTop := topPx >= 0
	hasBottom := bottomPx >= 0
	if topPx == 0 && hasBottom {
		hasTop = false
		topPx = -1
	}
	if bottomPx == 0 && hasTop {
		hasBottom = false
		bottomPx = -1
	}

	// Handle auto values (-1)
	// If both left and right are set, width is constrained
	// If both top and bottom are set, height is constrained
	if node.Style.Position == PositionAbsolute || node.Style.Position == PositionFixed {
		// Ensure absolutely positioned elements have size if specified
		widthPx := ResolveLength(node.Style.Width, ctx, currentFontSize)
		heightPx := ResolveLength(node.Style.Height, ctx, currentFontSize)
		if node.Rect.Width <= 0 && widthPx > 0 {
			node.Rect.Width = widthPx
		}
		if node.Rect.Height <= 0 && heightPx > 0 {
			node.Rect.Height = heightPx
		}

		// For absolute/fixed, calculate position from context
		if !hasLeft && !hasRight {
			// Both auto - position at 0,0 relative to context
			node.Rect.X = positioningContext.X
		} else if hasLeft && hasRight {
			// Both set - constrain width
			availableWidth := positioningContext.Width - leftPx - rightPx
			if availableWidth < 0 {
				availableWidth = 0
			}
			if node.Rect.Width > availableWidth {
				node.Rect.Width = availableWidth
			}
			node.Rect.X = positioningContext.X + leftPx
		} else if hasLeft {
			// Left set
			node.Rect.X = positioningContext.X + leftPx
		} else if hasRight {
			// Right set - position from right edge
			if node.Rect.Width > 0 {
				node.Rect.X = positioningContext.X + positioningContext.Width - node.Rect.Width - rightPx
			} else {
				// Width not set, position at right edge
				node.Rect.X = positioningContext.X + positioningContext.Width - rightPx
			}
		}
	}

	if node.Style.Position == PositionAbsolute || node.Style.Position == PositionFixed {
		// For absolute/fixed, calculate position from context
		if !hasTop && !hasBottom {
			// Both auto - position at 0,0 relative to context
			node.Rect.Y = positioningContext.Y
		} else if hasTop && hasBottom {
			// Both set - constrain height
			availableHeight := positioningContext.Height - topPx - bottomPx
			if availableHeight < 0 {
				availableHeight = 0
			}
			if node.Rect.Height > availableHeight {
				node.Rect.Height = availableHeight
			}
			node.Rect.Y = positioningContext.Y + topPx
		} else if hasTop {
			// Top set
			node.Rect.Y = positioningContext.Y + topPx
		} else if hasBottom {
			// Bottom set - position from bottom edge
			if node.Rect.Height > 0 {
				node.Rect.Y = positioningContext.Y + positioningContext.Height - node.Rect.Height - bottomPx
			} else {
				// Height not set, position at bottom edge
				node.Rect.Y = positioningContext.Y + positioningContext.Height - bottomPx
			}
		}
	}

	// For relative positioning, offset from normal flow position
	// (The normal flow position should already be set, we just offset it)
	// Note: left/right and top/bottom are offsets, not constraints
	if node.Style.Position == PositionRelative {
		// Store the original position before offsetting
		originalX := node.Rect.X
		originalY := node.Rect.Y

		if leftPx >= 0 {
			node.Rect.X = originalX + leftPx
		} else if rightPx >= 0 {
			node.Rect.X = originalX - rightPx
		}

		if topPx >= 0 {
			node.Rect.Y = originalY + topPx
		} else if bottomPx >= 0 {
			node.Rect.Y = originalY - bottomPx
		}
	}

	// For sticky positioning, we'd need scroll information
	// For now, treat it similar to relative
	if node.Style.Position == PositionSticky {
		// Sticky is a hybrid - starts as relative, becomes fixed when scrolled
		// Without scroll context, we'll treat it as relative
		if left.Value >= 0 {
			node.Rect.X += left.Value
		}
		if right.Value >= 0 {
			node.Rect.X -= right.Value
		}
		if top.Value >= 0 {
			node.Rect.Y += top.Value
		}
		if bottom.Value >= 0 {
			node.Rect.Y -= bottom.Value
		}
	}
}

// findPositionedAncestor finds the nearest positioned ancestor
// (position != static) for absolute positioning context
func findPositionedAncestor(node *Node, root *Node) *Node {
	// This is a simplified version - in a full implementation,
	// we'd need to traverse up the tree from node to root
	// For now, we'll return nil and use parent rect
	return nil
}

// LayoutWithPositioning performs layout including positioned elements
// This is a helper that handles the two-pass layout:
// 1. Normal flow layout
// 2. Positioned elements layout
func LayoutWithPositioning(root *Node, constraints Constraints, viewportRect Rect, ctx *LayoutContext) Size {
	// First pass: normal flow layout
	size := Layout(root, constraints, ctx)

	// Second pass: handle positioned elements
	layoutPositionedRecursive(root, root.Rect, viewportRect, ctx)

	return size
}

func layoutPositionedRecursive(node *Node, parentRect Rect, viewportRect Rect, ctx *LayoutContext) {
	// Layout positioned children
	for _, child := range node.Children {
		if child.Style.Position != PositionStatic {
			// Determine positioning context
			var context Rect
			if child.Style.Position == PositionFixed {
				context = viewportRect
			} else {
				// For absolute/relative/sticky, use parent's rect
				// In a full implementation, we'd find the nearest positioned ancestor
				context = node.Rect
			}

			LayoutPositioned(child, context, viewportRect, ctx)
		}

		// Recursively handle children
		layoutPositionedRecursive(child, child.Rect, viewportRect, ctx)
	}
}
