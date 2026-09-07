package layout

import (
	"fmt"
	"math"
	"sort"
)

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
//	stack.Style.Padding = layout.Uniform(layout.Px(10))
//
// MDN Guide: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_flexible_box_layout
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
//	stack.Style.Padding = layout.Uniform(layout.Px(10))
//
// MDN Guide: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_flexible_box_layout
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
//
// Because absolutely positioned children take no space in normal flow
// (CSS 2.1 §9.3), a block with only absolute children would otherwise be
// 0×0. ZStack therefore gives the container an intrinsic size: the union of
// its children's explicit pixel sizes, each offset by the child's pixel
// Left/Top. Children whose Width/Height (or offsets) are unset or use a
// non-pixel unit contribute nothing to that union, and a union of zero
// leaves the dimension unset (auto). Override the computed size with Frame
// or FrameLength after construction.
//
// MDN Guide: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_positioned_layout
func ZStack(children ...*Node) *Node {
	var unionWidth, unionHeight float64
	// Make all children absolutely positioned
	for _, child := range children {
		child.Style.Position = PositionAbsolute
		// Default to top-left if not specified.
		// We treat a Length whose Unit is the zero-value (empty string, since
		// LengthUnit is a typed string per units L4) as "unset" and default it
		// to Px(0). Px(0) itself has Unit=="px" and is therefore preserved.
		//
		// Prior to the units migration this check compared Unit == 0 because
		// LengthUnit was an int with Pixels = iota = 0. After the migration the
		// zero value of LengthUnit is "", so we check for that explicitly.
		if child.Style.Left.Value == 0 && child.Style.Left.Unit == "" {
			child.Style.Left = Px(0)
		}
		if child.Style.Top.Value == 0 && child.Style.Top.Unit == "" {
			child.Style.Top = Px(0)
		}

		if w, ok := pixelExtent(child.Style.Left, child.Style.Width); ok && w > unionWidth {
			unionWidth = w
		}
		if h, ok := pixelExtent(child.Style.Top, child.Style.Height); ok && h > unionHeight {
			unionHeight = h
		}
	}
	root := &Node{
		Style: Style{
			Position: PositionRelative, // Container needs to be positioned for absolute children
		},
		Children: children,
	}
	if unionWidth > 0 {
		root.Style.Width = Px(unionWidth)
	}
	if unionHeight > 0 {
		root.Style.Height = Px(unionHeight)
	}
	return root
}

// pixelExtent returns offset + size when both are pixel lengths (an unset
// offset counts as 0) and the size is set. It reports false when the size is
// unset or either value uses a unit that cannot be resolved without a
// LayoutContext.
func pixelExtent(offset, size Length) (float64, bool) {
	if size.Unit != Pixels {
		return 0, false
	}
	switch offset.Unit {
	case "":
		return size.Value, true
	case Pixels:
		return offset.Value + size.Value, true
	}
	return 0, false
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
			Width:  Px(width),
			Height: Px(height),
		},
	}
}

// Padding adds padding to a node
//
// MDN Guide: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_model
func Padding(node *Node, padding float64) *Node {
	node.Style.Padding = Uniform(Px(padding))
	return node
}

// PaddingCustom adds custom padding to a node
//
// MDN Guide: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_model
func PaddingCustom(node *Node, top, right, bottom, left float64) *Node {
	node.Style.Padding = Spacing{
		Top:    Px(top),
		Right:  Px(right),
		Bottom: Px(bottom),
		Left:   Px(left),
	}
	return node
}

// Margin adds margin to a node
//
// MDN Guide: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_model
func Margin(node *Node, margin float64) *Node {
	node.Style.Margin = Uniform(Px(margin))
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
//
// MDN Guide: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_alignment
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
//
// MDN Guide: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_alignment
func DistributeNodes(nodes []*Node, direction DistributeDirection) {
	// Guard against division-by-zero in the spacing calculation below
	// (spacing = totalSpace / float64(len(nodes)-1) would divide by zero
	// when len(nodes) == 1). With fewer than 2 nodes there is nothing to
	// distribute; with exactly 2 nodes the first and last stay fixed and
	// there are no middle nodes to space out, so this is also a no-op.
	if len(nodes) < 2 {
		return
	}

	// Create a slice of indices sorted by position
	indices := make([]int, len(nodes))
	for i := range indices {
		indices[i] = i
	}

	// Sort indices by position along the chosen axis using sort.Slice
	// (O(n log n)) rather than an in-place nested-loop sort (O(n^2)).
	if direction == DistributeHorizontal {
		sort.Slice(indices, func(i, j int) bool {
			return nodes[indices[i]].Rect.X < nodes[indices[j]].Rect.X
		})
	} else {
		sort.Slice(indices, func(i, j int) bool {
			return nodes[indices[i]].Rect.Y < nodes[indices[j]].Rect.Y
		})
	}

	// With exactly 2 nodes, there are no middle nodes to redistribute;
	// the first and last stay fixed, so we can return early.
	if len(nodes) < 3 {
		return
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
//	layout.Layout(root, constraints, ctx)
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
//	layout.Layout(root, constraints, ctx)
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

// Frame sets the width and/or height of a node in pixels.
//
// A width or height <= 0 is skipped and leaves that dimension untouched, so
// Frame(node, 200, 0) sets only the width. This means Frame cannot express an
// explicit zero size or reset a dimension to auto; use FrameLength for that.
func Frame(node *Node, width, height float64) *Node {
	if width > 0 {
		node.Style.Width = Px(width)
	}
	if height > 0 {
		node.Style.Height = Px(height)
	}
	return node
}

// FrameLength sets the width and height of a node to exactly the given
// lengths, with no special-casing: Px(0) is a real zero size and the zero
// value Length{} resets the dimension to auto (unset). Any unit is accepted.
//
// Example:
//
//	layout.FrameLength(node, layout.Em(10), layout.Length{}) // width 10em, height auto
//	layout.FrameLength(node, layout.Px(0), layout.Px(0))     // explicit 0×0
func FrameLength(node *Node, width, height Length) *Node {
	node.Style.Width = width
	node.Style.Height = height
	return node
}

// Background is a placeholder for styling (not layout-related).
// It returns the node unchanged.
//
// Deprecated: Background has never done anything; the layout engine has no
// notion of fill or color. Store visual properties alongside the tree in
// your own renderer instead. It is kept for source compatibility.
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
	node.Style.MinHeight = Px(height)
	return node
}

// MinWidth sets the minimum width of a node.
//
// Example:
//
//	item := layout.Fixed(100, 50)
//	item = layout.MinWidth(item, 120) // Ensures item is at least 120px wide
func MinWidth(node *Node, width float64) *Node {
	node.Style.MinWidth = Px(width)
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
//	        Width: layout.Px(800), // Width is set
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
//
// MDN Guide: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_sizing
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
//	grid.Style.GridGap = layout.Px(10)
//
// MDN Guide: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_grid_layout
func Grid(rows, cols int, rowSize, colSize float64) *Node {
	gridRows := make([]GridTrack, rows)
	for i := range gridRows {
		gridRows[i] = FixedTrack(Px(rowSize))
	}

	gridCols := make([]GridTrack, cols)
	for i := range gridCols {
		gridCols[i] = FixedTrack(Px(colSize))
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
//
// MDN Guide: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_grid_layout
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
//
// MDN Guide: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_grid_layout
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
// This is equivalent to the CSS repeat() function with an integer count.
//
// Example:
//
//	// Creates: [100px, 100px, 100px]
//	columns := layout.RepeatTracks(3, layout.FixedTrack(layout.Px(100)))
//
//	// Creates: [100px, 1fr, 100px, 1fr, 100px, 1fr]
//	columns := layout.RepeatTracks(3, layout.FixedTrack(layout.Px(100)), layout.FractionTrack(1))
//
// Count handling:
//   - count > 0: the pattern is repeated count times.
//   - count <= 0, INCLUDING RepeatCountAutoFill (-1) and RepeatCountAutoFit
//     (-2): an empty, non-nil slice is returned. RepeatTracks cannot expand
//     auto-repeat because the number of repetitions depends on the container
//     size, and a plain []GridTrack cannot carry the auto-repeat intent.
//     Put a RepeatTrack (from AutoFillTracks/AutoFitTracks, or a literal with
//     an integer Count) in Style.GridTemplateColumnsRepeat /
//     GridTemplateRowsRepeat instead; those are expanded against the
//     available size during grid layout.
//
// See: CSS Grid Layout Module Level 1 §5.1.2 (repeat() notation)
// https://www.w3.org/TR/css-grid-1/#repeat-notation
func RepeatTracks(count int, tracks ...GridTrack) []GridTrack {
	if count <= 0 {
		// Covers RepeatCountAutoFill/RepeatCountAutoFit as well; see doc above.
		return []GridTrack{}
	}
	result := make([]GridTrack, 0, count*len(tracks))
	for i := 0; i < count; i++ {
		result = append(result, tracks...)
	}
	return result
}

// NewGridTemplateAreas creates a new GridTemplateAreas with the specified grid dimensions.
// Named areas can then be defined using DefineArea.
//
// Example:
//
//	areas := layout.NewGridTemplateAreas(3, 3) // 3 rows x 3 columns
//	areas.DefineArea("header", 0, 1, 0, 3)      // Full width header
//	areas.DefineArea("sidebar", 1, 3, 0, 1)     // Left sidebar
//	areas.DefineArea("content", 1, 3, 1, 3)     // Main content area
//
// See: CSS Grid Layout Module Level 1 §7.3 (Named Areas)
// https://www.w3.org/TR/css-grid-1/#grid-template-areas-property
func NewGridTemplateAreas(rows, cols int) *GridTemplateAreas {
	return &GridTemplateAreas{
		Areas: make([]GridArea, 0),
		Rows:  rows,
		Cols:  cols,
	}
}

// DefineArea adds a named area to the grid template.
// Returns an error if the area overlaps with an existing area or is out of bounds.
//
// Parameters:
//   - name: The name of the area (e.g., "header", "sidebar", "content")
//   - rowStart: Starting row index (0-based, inclusive)
//   - rowEnd: Ending row index (0-based, exclusive)
//   - colStart: Starting column index (0-based, inclusive)
//   - colEnd: Ending column index (0-based, exclusive)
//
// Example:
//
//	areas := layout.NewGridTemplateAreas(3, 3)
//	err := areas.DefineArea("header", 0, 1, 0, 3) // Row 0, all columns
//	if err != nil {
//	    // Handle error (overlap or out of bounds)
//	}
func (gta *GridTemplateAreas) DefineArea(name string, rowStart, rowEnd, colStart, colEnd int) error {
	// Validate bounds
	if rowStart < 0 || rowEnd > gta.Rows || rowStart >= rowEnd {
		return fmt.Errorf("invalid row range [%d,%d) for grid with %d rows", rowStart, rowEnd, gta.Rows)
	}
	if colStart < 0 || colEnd > gta.Cols || colStart >= colEnd {
		return fmt.Errorf("invalid column range [%d,%d) for grid with %d columns", colStart, colEnd, gta.Cols)
	}

	newArea := GridArea{
		Name:        name,
		RowStart:    rowStart,
		RowEnd:      rowEnd,
		ColumnStart: colStart,
		ColumnEnd:   colEnd,
	}

	// Check for overlaps with existing areas
	for _, existing := range gta.Areas {
		if areasOverlap(existing, newArea) {
			return fmt.Errorf("area '%s' overlaps with existing area '%s'", name, existing.Name)
		}
	}

	gta.Areas = append(gta.Areas, newArea)
	return nil
}

// areasOverlap checks if two grid areas overlap.
// Two areas overlap if they share any grid cells.
func areasOverlap(a, b GridArea) bool {
	// Check if areas are completely separated in rows or columns
	if a.RowEnd <= b.RowStart || b.RowEnd <= a.RowStart {
		return false // Separated vertically
	}
	if a.ColumnEnd <= b.ColumnStart || b.ColumnEnd <= a.ColumnStart {
		return false // Separated horizontally
	}
	return true // They overlap
}

// PlaceInArea sets the GridArea property of a node, causing it to be placed
// in the named grid area during layout.
//
// Example:
//
//	header := layout.PlaceInArea(&layout.Node{}, "header")
//	sidebar := layout.PlaceInArea(&layout.Node{}, "sidebar")
//	content := layout.PlaceInArea(&layout.Node{}, "content")
func PlaceInArea(node *Node, areaName string) *Node {
	node.Style.GridArea = areaName
	return node
}

// The intrinsic sizing helpers below set Style.WidthSizing / HeightSizing
// and leave Style.Width / Height untouched. Before v1.5.0 they stored the
// deprecated SizeMinContent/SizeMaxContent/SizeFitContent sentinels in
// Width/Height, which the engine still honors for code that sets them
// directly, but which can collide with genuinely negative resolved lengths
// (see the sentinel docs in types.go). A Width/Height already present on the
// node is not cleared; the sizing mode takes precedence in every layout
// algorithm, so reset it with FrameLength if you need it gone.

// MinContentWidth sets a node's width to use min-content intrinsic sizing.
// The node will be as narrow as possible without overflowing content.
//
// Example:
//
//	node := MinContentWidth(&layout.Node{})
//
// See: CSS Sizing Module Level 3 §4.1 (min-content)
func MinContentWidth(node *Node) *Node {
	node.Style.WidthSizing = IntrinsicSizeMinContent
	return node
}

// MaxContentWidth sets a node's width to use max-content intrinsic sizing.
// The node will be as wide as its natural content width (no wrapping).
//
// Example:
//
//	node := MaxContentWidth(&layout.Node{})
//
// See: CSS Sizing Module Level 3 §4.2 (max-content)
func MaxContentWidth(node *Node) *Node {
	node.Style.WidthSizing = IntrinsicSizeMaxContent
	return node
}

// FitContentWidth sets a node's width to use fit-content intrinsic sizing.
// The width will be max-content clamped to the specified maximum size
// (stored in Style.FitContentWidth as pixels).
//
// Example:
//
//	node := FitContentWidth(&layout.Node{}, 500) // max-content, but no wider than 500px
//
// See: CSS Sizing Module Level 3 §4.3 (fit-content)
func FitContentWidth(node *Node, maxSize float64) *Node {
	node.Style.WidthSizing = IntrinsicSizeFitContent
	node.Style.FitContentWidth = Px(maxSize)
	return node
}

// MinContentHeight sets a node's height to use min-content intrinsic sizing.
func MinContentHeight(node *Node) *Node {
	node.Style.HeightSizing = IntrinsicSizeMinContent
	return node
}

// MaxContentHeight sets a node's height to use max-content intrinsic sizing.
func MaxContentHeight(node *Node) *Node {
	node.Style.HeightSizing = IntrinsicSizeMaxContent
	return node
}

// FitContentHeight sets a node's height to use fit-content intrinsic sizing,
// clamped to maxSize pixels (stored in Style.FitContentHeight).
func FitContentHeight(node *Node, maxSize float64) *Node {
	node.Style.HeightSizing = IntrinsicSizeFitContent
	node.Style.FitContentHeight = Px(maxSize)
	return node
}

// MinContentTrack creates a grid track that uses min-content sizing.
// The track will be sized to the minimum content size of items in that track.
//
// Example:
//
//	GridTemplateColumns: []GridTrack{MinContentTrack(), FixedTrack(Px(200))}
//
// See: CSS Grid Layout Module Level 1 §7.2.3 (min-content and max-content Track Sizing Functions)
func MinContentTrack() GridTrack {
	return GridTrack{
		MinSize:  Px(0),
		MaxSize:  Px(SizeMinContent),
		Fraction: 0,
	}
}

// MaxContentTrack creates a grid track that uses max-content sizing.
// The track will be sized to the maximum content size of items in that track.
//
// Example:
//
//	GridTemplateColumns: []GridTrack{MaxContentTrack(), FixedTrack(Px(200))}
func MaxContentTrack() GridTrack {
	return GridTrack{
		MinSize:  Px(0),
		MaxSize:  Px(SizeMaxContent),
		Fraction: 0,
	}
}

// FitContentTrack creates a grid track that uses fit-content sizing.
// The track will be max-content clamped to the specified maximum size.
//
// Example:
//
//	GridTemplateColumns: []GridTrack{FitContentTrack(300), FixedTrack(Px(200))}
func FitContentTrack(maxSize float64) GridTrack {
	return GridTrack{
		MinSize:  Px(0),
		MaxSize:  Px(maxSize),
		Fraction: -1, // Special marker for fit-content
	}
}

// AutoFillTracks creates a RepeatTrack for auto-fill grid track generation.
// Auto-fill creates as many tracks as fit in the available space.
//
// Example:
//
//	// CSS: grid-template-columns: repeat(auto-fill, 100px);
//	grid := &layout.Node{Style: layout.Style{
//	    Display: layout.DisplayGrid,
//	    GridTemplateColumnsRepeat: []layout.RepeatTrack{
//	        layout.AutoFillTracks(layout.FixedTrack(layout.Px(100))),
//	    },
//	}}
//
// The repeat is expanded against the container's available size when the
// grid is laid out; the generated tracks are appended after
// Style.GridTemplateColumns. Auto-repeat patterns must use definite track
// sizes (no fr or auto), per the spec. RepeatTracks(RepeatCountAutoFill, ...)
// returns an empty slice because a plain []GridTrack cannot carry the intent.
//
// See: CSS Grid Layout Module Level 1 §7.2.3 (auto-fill)
// https://www.w3.org/TR/css-grid-1/#auto-repeat
func AutoFillTracks(tracks ...GridTrack) RepeatTrack {
	return RepeatTrack{
		Count:  RepeatCountAutoFill,
		Tracks: tracks,
	}
}

// AutoFitTracks creates a RepeatTrack for auto-fit grid track generation.
// Auto-fit creates as many tracks as fit, then collapses empty tracks to zero.
//
// Example:
//
//	// CSS: grid-template-columns: repeat(auto-fit, 100px);
//	grid := &layout.Node{Style: layout.Style{
//	    Display: layout.DisplayGrid,
//	    GridTemplateColumnsRepeat: []layout.RepeatTrack{
//	        layout.AutoFitTracks(layout.FixedTrack(layout.Px(100))),
//	    },
//	}}
//
// The difference between auto-fill and auto-fit:
//   - auto-fill: keeps all generated tracks, even if empty
//   - auto-fit: collapses empty tracks to zero size
//
// See AutoFillTracks for how the pattern is expanded during layout.
//
// See: CSS Grid Layout Module Level 1 §7.2.3 (auto-fit)
// https://www.w3.org/TR/css-grid-1/#auto-repeat
func AutoFitTracks(tracks ...GridTrack) RepeatTrack {
	return RepeatTrack{
		Count:  RepeatCountAutoFit,
		Tracks: tracks,
	}
}
