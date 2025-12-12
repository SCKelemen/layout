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
func LayoutPositioned(node *Node, parentRect Rect, viewportRect Rect) {
	if node.Style.Position == PositionStatic {
		// Static positioning is the default, no special handling needed
		return
	}

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

	// Determine if sides are explicitly set
	// If a side is 0 and the opposite is set, treat 0 as auto (unset)
	hasLeft := left >= 0
	hasRight := right >= 0
	if left == 0 && hasRight {
		hasLeft = false
		left = -1
	}
	if right == 0 && hasLeft {
		hasRight = false
		right = -1
	}

	hasTop := top >= 0
	hasBottom := bottom >= 0
	if top == 0 && hasBottom {
		hasTop = false
		top = -1
	}
	if bottom == 0 && hasTop {
		hasBottom = false
		bottom = -1
	}

	// Handle auto values (-1)
	// If both left and right are set, width is constrained
	// If both top and bottom are set, height is constrained
	if node.Style.Position == PositionAbsolute || node.Style.Position == PositionFixed {
		// Ensure absolutely positioned elements have size if specified
		if node.Rect.Width <= 0 && node.Style.Width > 0 {
			node.Rect.Width = node.Style.Width
		}
		if node.Rect.Height <= 0 && node.Style.Height > 0 {
			node.Rect.Height = node.Style.Height
		}

		// For absolute/fixed, calculate position from context
		if !hasLeft && !hasRight {
			// Both auto - position at 0,0 relative to context
			node.Rect.X = positioningContext.X
		} else if hasLeft && hasRight {
			// Both set - constrain width
			availableWidth := positioningContext.Width - left - right
			if availableWidth < 0 {
				availableWidth = 0
			}
			if node.Rect.Width > availableWidth {
				node.Rect.Width = availableWidth
			}
			node.Rect.X = positioningContext.X + left
		} else if hasLeft {
			// Left set
			node.Rect.X = positioningContext.X + left
		} else if hasRight {
			// Right set - position from right edge
			if node.Rect.Width > 0 {
				node.Rect.X = positioningContext.X + positioningContext.Width - node.Rect.Width - right
			} else {
				// Width not set, position at right edge
				node.Rect.X = positioningContext.X + positioningContext.Width - right
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
			availableHeight := positioningContext.Height - top - bottom
			if availableHeight < 0 {
				availableHeight = 0
			}
			if node.Rect.Height > availableHeight {
				node.Rect.Height = availableHeight
			}
			node.Rect.Y = positioningContext.Y + top
		} else if hasTop {
			// Top set
			node.Rect.Y = positioningContext.Y + top
		} else if hasBottom {
			// Bottom set - position from bottom edge
			if node.Rect.Height > 0 {
				node.Rect.Y = positioningContext.Y + positioningContext.Height - node.Rect.Height - bottom
			} else {
				// Height not set, position at bottom edge
				node.Rect.Y = positioningContext.Y + positioningContext.Height - bottom
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

		if left >= 0 {
			node.Rect.X = originalX + left
		} else if right >= 0 {
			node.Rect.X = originalX - right
		}

		if top >= 0 {
			node.Rect.Y = originalY + top
		} else if bottom >= 0 {
			node.Rect.Y = originalY - bottom
		}
	}

	// For sticky positioning, we'd need scroll information
	// For now, treat it similar to relative
	if node.Style.Position == PositionSticky {
		// Sticky is a hybrid - starts as relative, becomes fixed when scrolled
		// Without scroll context, we'll treat it as relative
		if left >= 0 {
			node.Rect.X += left
		}
		if right >= 0 {
			node.Rect.X -= right
		}
		if top >= 0 {
			node.Rect.Y += top
		}
		if bottom >= 0 {
			node.Rect.Y -= bottom
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
func LayoutWithPositioning(root *Node, constraints Constraints, viewportRect Rect) Size {
	// First pass: normal flow layout
	size := Layout(root, constraints)

	// Second pass: handle positioned elements
	layoutPositionedRecursive(root, root.Rect, viewportRect)

	return size
}

func layoutPositionedRecursive(node *Node, parentRect Rect, viewportRect Rect) {
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

			LayoutPositioned(child, context, viewportRect)
		}

		// Recursively handle children
		layoutPositionedRecursive(child, child.Rect, viewportRect)
	}
}
