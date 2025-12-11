package layout

// High-level API helpers inspired by SwiftUI and Flutter.
// These provide a simpler, more ergonomic API for common use cases.

// HStack creates a horizontal stack (row flexbox container).
// Children are arranged horizontally from left to right.
//
// Example:
//
//	stack := layout.HStack(
//	    layout.Fixed(100, 50),
//	    layout.Spacer(),
//	    layout.Fixed(100, 50),
//	)
//	stack.Style.Padding = layout.Uniform(10)
func HStack(children ...*Node) *Node {
	return &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
		},
		Children: children,
	}
}

// VStack creates a vertical stack (column flexbox container).
// Children are arranged vertically from top to bottom.
//
// Example:
//
//	stack := layout.VStack(
//	    layout.Fixed(100, 50),
//	    layout.Spacer(),
//	    layout.Fixed(100, 50),
//	)
//	stack.Style.Padding = layout.Uniform(10)
func VStack(children ...*Node) *Node {
	return &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
		},
		Children: children,
	}
}

// ZStack creates a stack with overlapping children (absolute positioning).
// Children are positioned absolutely, allowing them to overlap.
// Use LayoutWithPositioning to properly layout ZStack children.
func ZStack(children ...*Node) *Node {
	// Make all children absolutely positioned
	for _, child := range children {
		child.Style.Position = PositionAbsolute
		// Default to top-left if not specified
		if child.Style.Left < 0 {
			child.Style.Left = 0
		}
		if child.Style.Top < 0 {
			child.Style.Top = 0
		}
	}
	return &Node{
		Style: Style{
			Position: PositionRelative, // Container needs to be positioned for absolute children
		},
		Children: children,
	}
}

// Spacer creates a flexible spacer that grows to fill available space.
// Useful in HStack and VStack to push elements apart.
func Spacer() *Node {
	return &Node{
		Style: Style{
			FlexGrow: 1,
		},
	}
}

// Fixed creates a node with fixed width and height
func Fixed(width, height float64) *Node {
	return &Node{
		Style: Style{
			Width:  width,
			Height: height,
		},
	}
}

// Padding adds padding to a node
func Padding(node *Node, padding float64) *Node {
	node.Style.Padding = Uniform(padding)
	return node
}

// PaddingCustom adds custom padding to a node
func PaddingCustom(node *Node, top, right, bottom, left float64) *Node {
	node.Style.Padding = Spacing{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
	return node
}

// Margin adds margin to a node
func Margin(node *Node, margin float64) *Node {
	node.Style.Margin = Uniform(margin)
	return node
}

// Align centers a node within its parent
func Align(node *Node, horizontal, vertical bool) *Node {
	// This would need to be applied to the parent container
	// For now, we'll set alignment hints
	// In a full implementation, this would modify the parent's style
	return node
}

// Frame sets the width and/or height of a node
func Frame(node *Node, width, height float64) *Node {
	if width > 0 {
		node.Style.Width = width
	}
	if height > 0 {
		node.Style.Height = height
	}
	return node
}

// Background is a placeholder for styling (not layout-related)
// This would be used by rendering code, not layout
func Background(node *Node) *Node {
	return node
}

// MinHeight sets the minimum height of a node.
// This is especially important for items in auto-sized grid rows.
//
// Example:
//
//	item := layout.Fixed(100, 50)
//	item = layout.MinHeight(item, 60) // Ensures item is at least 60px tall
func MinHeight(node *Node, height float64) *Node {
	node.Style.MinHeight = height
	return node
}

// MinWidth sets the minimum width of a node.
//
// Example:
//
//	item := layout.Fixed(100, 50)
//	item = layout.MinWidth(item, 120) // Ensures item is at least 120px wide
func MinWidth(node *Node, width float64) *Node {
	node.Style.MinWidth = width
	return node
}

// Grid creates a grid container with the specified number of rows and columns.
// Each row and column will have the same fixed size.
//
// Example:
//
//	grid := layout.Grid(3, 4, 150, 200) // 3 rows x 4 columns, rows=150px, cols=200px
//	grid.Style.GridGap = 10
func Grid(rows, cols int, rowSize, colSize float64) *Node {
	gridRows := make([]GridTrack, rows)
	for i := range gridRows {
		gridRows[i] = FixedTrack(rowSize)
	}

	gridCols := make([]GridTrack, cols)
	for i := range gridCols {
		gridCols[i] = FixedTrack(colSize)
	}

	return &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    gridRows,
			GridTemplateColumns: gridCols,
		},
	}
}

// GridAuto creates a grid container with auto-sized tracks.
// Useful when you want the grid to size based on content.
//
// Example:
//
//	grid := layout.GridAuto(3, 4) // 3 rows x 4 columns, auto-sized
func GridAuto(rows, cols int) *Node {
	gridRows := make([]GridTrack, rows)
	for i := range gridRows {
		gridRows[i] = AutoTrack()
	}

	gridCols := make([]GridTrack, cols)
	for i := range gridCols {
		gridCols[i] = AutoTrack()
	}

	return &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    gridRows,
			GridTemplateColumns: gridCols,
		},
	}
}

// GridFractional creates a grid container with fractional (fr) tracks.
// All rows/columns will share space equally.
//
// Example:
//
//	grid := layout.GridFractional(3, 4) // 3 rows x 4 columns, all equal fractional units
func GridFractional(rows, cols int) *Node {
	gridRows := make([]GridTrack, rows)
	for i := range gridRows {
		gridRows[i] = FractionTrack(1)
	}

	gridCols := make([]GridTrack, cols)
	for i := range gridCols {
		gridCols[i] = FractionTrack(1)
	}

	return &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    gridRows,
			GridTemplateColumns: gridCols,
		},
	}
}
