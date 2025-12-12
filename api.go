package layout

import "math"

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

// AlignEdge represents an edge for alignment operations
type AlignEdge int

const (
	AlignLeft AlignEdge = iota
	AlignRight
	AlignTop
	AlignBottom
	AlignCenterX
	AlignCenterY
)

// AlignNodes aligns multiple nodes to a common reference point.
// This is a post-layout operation that modifies the Rect positions of nodes.
// Useful for design-tool-like alignment operations.
//
// Example:
//
//	nodes := []*layout.Node{item1, item2, item3}
//	layout.AlignNodes(nodes, layout.AlignLeft) // Align all to left edge
//	layout.AlignNodes(nodes, layout.AlignCenterY) // Align all to vertical center
//
// Note: This modifies the Rect positions directly. Call Layout() first to compute initial positions.
func AlignNodes(nodes []*Node, edge AlignEdge) {
	if len(nodes) == 0 {
		return
	}

	switch edge {
	case AlignLeft:
		// Find the leftmost X coordinate
		minX := nodes[0].Rect.X
		for _, node := range nodes {
			if node.Rect.X < minX {
				minX = node.Rect.X
			}
		}
		// Align all nodes to this X coordinate
		for _, node := range nodes {
			node.Rect.X = minX
		}

	case AlignRight:
		// Find the rightmost X coordinate
		maxX := nodes[0].Rect.X + nodes[0].Rect.Width
		for _, node := range nodes {
			right := node.Rect.X + node.Rect.Width
			if right > maxX {
				maxX = right
			}
		}
		// Align all nodes to this X coordinate
		for _, node := range nodes {
			node.Rect.X = maxX - node.Rect.Width
		}

	case AlignTop:
		// Find the topmost Y coordinate
		minY := nodes[0].Rect.Y
		for _, node := range nodes {
			if node.Rect.Y < minY {
				minY = node.Rect.Y
			}
		}
		// Align all nodes to this Y coordinate
		for _, node := range nodes {
			node.Rect.Y = minY
		}

	case AlignBottom:
		// Find the bottommost Y coordinate
		maxY := nodes[0].Rect.Y + nodes[0].Rect.Height
		for _, node := range nodes {
			bottom := node.Rect.Y + node.Rect.Height
			if bottom > maxY {
				maxY = bottom
			}
		}
		// Align all nodes to this Y coordinate
		for _, node := range nodes {
			node.Rect.Y = maxY - node.Rect.Height
		}

	case AlignCenterX:
		// Find the center X coordinate (average of all centers)
		centers := make([]float64, len(nodes))
		for i, node := range nodes {
			centers[i] = node.Rect.X + node.Rect.Width/2
		}
		// Use the average center
		sum := 0.0
		for _, c := range centers {
			sum += c
		}
		avgCenter := sum / float64(len(centers))
		// Align all nodes to this center
		for _, node := range nodes {
			node.Rect.X = avgCenter - node.Rect.Width/2
		}

	case AlignCenterY:
		// Find the center Y coordinate (average of all centers)
		centers := make([]float64, len(nodes))
		for i, node := range nodes {
			centers[i] = node.Rect.Y + node.Rect.Height/2
		}
		// Use the average center
		sum := 0.0
		for _, c := range centers {
			sum += c
		}
		avgCenter := sum / float64(len(centers))
		// Align all nodes to this center
		for _, node := range nodes {
			node.Rect.Y = avgCenter - node.Rect.Height/2
		}
	}
}

// DistributeDirection represents the direction for distribution
type DistributeDirection int

const (
	DistributeHorizontal DistributeDirection = iota
	DistributeVertical
)

// DistributeNodes evenly spaces multiple nodes horizontally or vertically.
// This is a post-layout operation that modifies the Rect positions of nodes.
// Useful for design-tool-like distribution operations.
//
// The distribution is based on the centers of the nodes, similar to design tools.
//
// Example:
//
//	nodes := []*layout.Node{item1, item2, item3}
//	layout.DistributeNodes(nodes, layout.DistributeHorizontal) // Evenly space horizontally
//	layout.DistributeNodes(nodes, layout.DistributeVertical)   // Evenly space vertically
//
// Note: This modifies the Rect positions directly. Call Layout() first to compute initial positions.
// The nodes are sorted by their current position before distribution.
func DistributeNodes(nodes []*Node, direction DistributeDirection) {
	if len(nodes) < 3 {
		// Need at least 3 nodes to distribute (first and last stay fixed)
		return
	}

	// Create a slice of indices sorted by position
	indices := make([]int, len(nodes))
	for i := range indices {
		indices[i] = i
	}

	// Sort indices by position
	if direction == DistributeHorizontal {
		// Sort by X coordinate
		for i := 0; i < len(indices)-1; i++ {
			for j := i + 1; j < len(indices); j++ {
				if nodes[indices[i]].Rect.X > nodes[indices[j]].Rect.X {
					indices[i], indices[j] = indices[j], indices[i]
				}
			}
		}
	} else {
		// Sort by Y coordinate
		for i := 0; i < len(indices)-1; i++ {
			for j := i + 1; j < len(indices); j++ {
				if nodes[indices[i]].Rect.Y > nodes[indices[j]].Rect.Y {
					indices[i], indices[j] = indices[j], indices[i]
				}
			}
		}
	}

	// Get the first and last positions (these stay fixed)
	firstIdx := indices[0]
	lastIdx := indices[len(indices)-1]

	if direction == DistributeHorizontal {
		firstPos := nodes[firstIdx].Rect.X + nodes[firstIdx].Rect.Width/2
		lastPos := nodes[lastIdx].Rect.X + nodes[lastIdx].Rect.Width/2
		totalSpace := lastPos - firstPos

		// Calculate spacing between centers
		spacing := totalSpace / float64(len(nodes)-1)

		// Distribute the middle nodes
		for i := 1; i < len(indices)-1; i++ {
			idx := indices[i]
			centerPos := firstPos + spacing*float64(i)
			nodes[idx].Rect.X = centerPos - nodes[idx].Rect.Width/2
		}
	} else {
		firstPos := nodes[firstIdx].Rect.Y + nodes[firstIdx].Rect.Height/2
		lastPos := nodes[lastIdx].Rect.Y + nodes[lastIdx].Rect.Height/2
		totalSpace := lastPos - firstPos

		// Calculate spacing between centers
		spacing := totalSpace / float64(len(nodes)-1)

		// Distribute the middle nodes
		for i := 1; i < len(indices)-1; i++ {
			idx := indices[i]
			centerPos := firstPos + spacing*float64(i)
			nodes[idx].Rect.Y = centerPos - nodes[idx].Rect.Height/2
		}
	}
}

// SnapNodes snaps multiple nodes to a grid boundary.
// This is a post-layout operation that modifies the Rect positions of nodes.
// Useful for pixel-perfect alignment and design-tool-like snapping.
//
// **Important**: Snapping is primarily intended for block layouts and absolutely
// positioned elements. Snapping items within Flexbox or Grid containers may break
// the layout algorithm's intended positioning and cause overlaps or misalignment.
//
// Example:
//
//	// For block/absolute layouts
//	nodes := []*layout.Node{item1, item2, item3}
//	layout.Layout(root, constraints)
//	layout.SnapNodes(nodes, 10.0) // Snap to 10px grid
//
// Note: This modifies the Rect positions directly. Call Layout() first to compute initial positions.
func SnapNodes(nodes []*Node, snapSize float64) {
	if snapSize <= 0 {
		return
	}

	for _, node := range nodes {
		// Snap X to nearest grid boundary
		node.Rect.X = math.Round(node.Rect.X/snapSize) * snapSize
		// Snap Y to nearest grid boundary
		node.Rect.Y = math.Round(node.Rect.Y/snapSize) * snapSize
	}
}

// SnapToGrid snaps nodes to a specific grid with an origin point.
// This allows snapping to a subgrid or offset grid.
//
// **Important**: Snapping is primarily intended for block layouts and absolutely
// positioned elements. Snapping items within Flexbox or Grid containers may break
// the layout algorithm's intended positioning and cause overlaps or misalignment.
//
// Example:
//
//	// For block/absolute layouts with offset grid
//	nodes := []*layout.Node{item1, item2, item3}
//	layout.Layout(root, constraints)
//	layout.SnapToGrid(nodes, 10.0, 5.0, 5.0) // 10px grid, offset by (5, 5)
//
// Note: This modifies the Rect positions directly. Call Layout() first to compute initial positions.
func SnapToGrid(nodes []*Node, snapSize, originX, originY float64) {
	if snapSize <= 0 {
		return
	}

	for _, node := range nodes {
		// Calculate position relative to grid origin
		relativeX := node.Rect.X - originX
		relativeY := node.Rect.Y - originY

		// Snap to grid and add origin back
		node.Rect.X = math.Round(relativeX/snapSize)*snapSize + originX
		node.Rect.Y = math.Round(relativeY/snapSize)*snapSize + originY
	}
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

// AspectRatio sets the aspect ratio (width/height) for a node.
// This helps elements reserve space correctly when one dimension is auto.
// Aspect ratio is applied when width or height is auto (not explicitly set).
//
// Common aspect ratios:
//   - 16/9 = 1.777... (widescreen video)
//   - 4/3 = 1.333... (traditional TV)
//   - 1/1 = 1.0 (square)
//   - 3/2 = 1.5 (photo)
//
// Example:
//
//	// Image that maintains 16:9 aspect ratio
//	image := &layout.Node{
//	    Style: layout.Style{
//	        Width: 800, // Width is set
//	        // Height will be calculated: 800 / 1.777... = 450
//	    },
//	}
//	image = layout.AspectRatio(image, 16.0/9.0)
//
//	// Element that fills available width and maintains aspect ratio
//	video := &layout.Node{
//	    Style: layout.Style{
//	        // Both width and height are auto
//	        // Will use available width and calculate height from aspect ratio
//	    },
//	}
//	video = layout.AspectRatio(video, 16.0/9.0)
func AspectRatio(node *Node, ratio float64) *Node {
	node.Style.AspectRatio = ratio
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

// RepeatTracks creates a repeated pattern of grid tracks.
// This is equivalent to the CSS repeat() function.
//
// Example:
//
//	// Creates: [100px, 100px, 100px]
//	columns := layout.RepeatTracks(3, layout.FixedTrack(100))
//
//	// Creates: [100px, 1fr, 100px, 1fr, 100px, 1fr]
//	columns := layout.RepeatTracks(3, layout.FixedTrack(100), layout.FractionTrack(1))
//
// See: CSS Grid Layout Module Level 1 §5.1.2 (repeat() notation)
// https://www.w3.org/TR/css-grid-1/#repeat-notation
func RepeatTracks(count int, tracks ...GridTrack) []GridTrack {
	if count <= 0 {
		return []GridTrack{}
	}
	result := make([]GridTrack, 0, count*len(tracks))
	for i := 0; i < count; i++ {
		result = append(result, tracks...)
	}
	return result
}
