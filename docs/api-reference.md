# API Reference

Complete reference for all exported functions and types.

## Core Types

### Node

Represents a node in the layout tree.

```go
type Node struct {
    Style    Style
    Rect     Rect      // Computed after Layout()
    Children []*Node
}
```

### Style

Contains all layout properties.

```go
type Style struct {
    // Display
    Display Display
    
    // Flexbox
    FlexDirection  FlexDirection
    FlexWrap       FlexWrap
    JustifyContent JustifyContent
    AlignItems     AlignItems
    AlignContent   AlignContent
    FlexGrow       float64
    FlexShrink     float64
    FlexBasis      float64  // -1 means auto
    
    // Grid
    GridTemplateRows    []GridTrack
    GridTemplateColumns []GridTrack
    GridAutoRows        GridTrack
    GridAutoColumns     GridTrack
    GridGap             float64
    GridRowGap          float64
    GridColumnGap       float64
    GridRowStart        int  // -1 means auto
    GridRowEnd          int  // -1 means auto
    GridColumnStart     int  // -1 means auto
    GridColumnEnd       int  // -1 means auto
    
    // Sizing
    Width     float64  // -1 means auto
    Height    float64  // -1 means auto
    MinWidth  float64
    MinHeight float64
    MaxWidth  float64
    MaxHeight float64
    
    // Spacing
    Padding Spacing
    Margin  Spacing  // Supported in Flexbox and Grid layouts
    Border  Spacing
    
    // Box Model
    BoxSizing BoxSizing
    
    // Positioning
    Position Position
    Top      float64  // -1 means auto
    Right    float64  // -1 means auto
    Bottom   float64  // -1 means auto
    Left     float64  // -1 means auto
    ZIndex   int
    
    // Transform
    Transform Transform
}
```

### Constraints

Defines available space for layout.

```go
type Constraints struct {
    MinWidth  float64
    MaxWidth  float64
    MinHeight float64
    MaxHeight float64
}
```

### Rect

Position and size of a laid-out node.

```go
type Rect struct {
    X      float64
    Y      float64
    Width  float64
    Height float64
}
```

### Size

Width and height.

```go
type Size struct {
    Width  float64
    Height float64
}
```

## Layout Functions

### Layout

Main layout function. Routes to appropriate layout algorithm based on display type.

```go
func Layout(root *Node, constraints Constraints) Size
```

Performs normal flow layout only. For positioned elements, use `LayoutWithPositioning`.

### LayoutWithPositioning

Performs layout including positioned elements (absolute, relative, fixed, sticky).

```go
func LayoutWithPositioning(root *Node, constraints Constraints, viewport Rect) Size
```

This performs a two-pass layout:
1. Normal flow layout
2. Positioned elements layout

### LayoutFlexbox

Flexbox layout algorithm.

```go
func LayoutFlexbox(node *Node, constraints Constraints) Size
```

### LayoutGrid

Grid layout algorithm.

```go
func LayoutGrid(node *Node, constraints Constraints) Size
```

### LayoutBlock

Block layout algorithm.

```go
func LayoutBlock(node *Node, constraints Constraints) Size
```

## Constraint Helpers

### Tight

Creates tight constraints (exact size required).

```go
func Tight(width, height float64) Constraints
```

### Loose

Creates loose constraints (maximum size, can be smaller).

```go
func Loose(width, height float64) Constraints
```

### Unconstrained

Creates unconstrained constraints (no size limits).

```go
func Unconstrained() Constraints
```

## Grid Track Helpers

### FixedTrack

Creates a fixed-size grid track.

```go
func FixedTrack(size float64) GridTrack
```

### FractionTrack

Creates a fractional grid track (fr unit).

```go
func FractionTrack(fraction float64) GridTrack
```

### MinMaxTrack

Creates a grid track with min/max constraints.

```go
func MinMaxTrack(min, max float64) GridTrack
```

### AutoTrack

Creates an auto-sized grid track.

```go
func AutoTrack() GridTrack
```

## Spacing Helpers

### Uniform

Creates uniform spacing on all sides.

```go
func Uniform(value float64) Spacing
```

### Horizontal

Creates horizontal spacing (left and right).

```go
func Horizontal(value float64) Spacing
```

### Vertical

Creates vertical spacing (top and bottom).

```go
func Vertical(value float64) Spacing
```

## High-Level API

### HStack

Creates a horizontal stack (row flexbox container).

```go
func HStack(children ...*Node) *Node
```

### VStack

Creates a vertical stack (column flexbox container).

```go
func VStack(children ...*Node) *Node
```

### ZStack

Creates a stack with overlapping children (absolute positioning).

```go
func ZStack(children ...*Node) *Node
```

Use `LayoutWithPositioning` to properly layout ZStack children.

### Spacer

Creates a flexible spacer that grows to fill available space.

```go
func Spacer() *Node
```

### Fixed

Creates a node with fixed width and height.

```go
func Fixed(width, height float64) *Node
```

### Padding

Adds uniform padding to a node.

```go
func Padding(node *Node, padding float64) *Node
```

### PaddingCustom

Adds custom padding to a node.

```go
func PaddingCustom(node *Node, top, right, bottom, left float64) *Node
```

### Margin

Adds uniform margin to a node. Margins are fully supported in Flexbox and Grid layouts.

```go
func Margin(node *Node, margin float64) *Node
```

Example:
```go
item := layout.Fixed(100, 50)
item = layout.Margin(item, 10) // 10px margin on all sides

// Or use with HStack/VStack
stack := layout.HStack(
    layout.Margin(layout.Fixed(100, 50), 10),
    layout.Margin(layout.Fixed(100, 50), 10),
)
```

**Note**: Margins don't collapse in Flexbox or Grid (CSS-compliant behavior).

### Frame

Sets the width and/or height of a node.

```go
func Frame(node *Node, width, height float64) *Node
```

## Post-Layout Alignment and Distribution

These functions operate on nodes **after** layout has been computed. They modify the `Rect` positions directly, similar to design tools like Figma or Sketch.

**Note**: Call `Layout()` first to compute initial positions, then use these functions to adjust positions.

### AlignNodes

Aligns multiple nodes to a common reference point (edge or center).

```go
func AlignNodes(nodes []*Node, edge AlignEdge)
```

**Alignment Edges**:
- `AlignLeft`: Align all nodes to the leftmost edge
- `AlignRight`: Align all nodes to the rightmost edge
- `AlignTop`: Align all nodes to the topmost edge
- `AlignBottom`: Align all nodes to the bottommost edge
- `AlignCenterX`: Align all nodes to a common horizontal center (average of all centers)
- `AlignCenterY`: Align all nodes to a common vertical center (average of all centers)

**Example**:
```go
// After layout, align all items to the left edge
items := []*layout.Node{item1, item2, item3}
layout.Layout(root, constraints)
layout.AlignNodes(items, layout.AlignLeft)

// Align to vertical center
layout.AlignNodes(items, layout.AlignCenterY)
```

**Use cases**:
- SVG rendering: Aligning multiple elements
- Design tool-like operations
- Post-layout adjustments

### DistributeNodes

Evenly spaces multiple nodes horizontally or vertically, based on their centers.

```go
func DistributeNodes(nodes []*Node, direction DistributeDirection)
```

**Directions**:
- `DistributeHorizontal`: Evenly space nodes horizontally
- `DistributeVertical`: Evenly space nodes vertically

**Behavior**:
- Requires at least 3 nodes
- The leftmost/topmost and rightmost/bottommost nodes stay fixed
- Middle nodes are evenly spaced between them
- Nodes are automatically sorted by position before distribution

**Example**:
```go
// After layout, evenly space items horizontally
items := []*layout.Node{item1, item2, item3, item4}
layout.Layout(root, constraints)
layout.DistributeNodes(items, layout.DistributeHorizontal)

// Evenly space vertically
layout.DistributeNodes(items, layout.DistributeVertical)
```

**Use cases**:
- Creating evenly spaced button groups
- Distributing labels or icons
- Design tool-like distribution operations

**Note**: These functions are based on the [CSS Box Alignment Module Level 3](https://www.w3.org/TR/css-align-3/) specification concepts, but operate as post-layout utilities rather than container-based alignment (which is handled by Flexbox/Grid `justify-content` and `align-items`).

### SnapNodes

Snaps multiple nodes to a grid boundary for pixel-perfect alignment.

```go
func SnapNodes(nodes []*Node, snapSize float64)
```

**Important**: Snapping is primarily intended for **block layouts** and **absolutely positioned elements**. Snapping items within Flexbox or Grid containers may break the layout algorithm's intended positioning and cause overlaps or misalignment.

**Example**:
```go
// For block/absolute layouts - snap to 10px grid
items := []*layout.Node{item1, item2, item3}
layout.Layout(root, constraints)
layout.SnapNodes(items, 10.0) // All positions snap to 10px boundaries
```

**Use cases**:
- Pixel-perfect alignment in block layouts
- Snapping absolutely positioned elements
- Design tool-like grid snapping
- SVG rendering with precise positioning

**When NOT to use**:
- Items in Flexbox containers (use `justify-content` and `align-items` instead)
- Items in Grid containers (use grid alignment properties instead)
- Items that need to maintain responsive spacing

### SnapToGrid

Snaps nodes to a specific grid with an origin point, allowing snapping to a subgrid or offset grid.

```go
func SnapToGrid(nodes []*Node, snapSize, originX, originY float64)
```

**Important**: Snapping is primarily intended for **block layouts** and **absolutely positioned elements**. Snapping items within Flexbox or Grid containers may break the layout algorithm's intended positioning.

**Example**:
```go
// Snap to a 10px grid offset by (5, 5)
items := []*layout.Node{item1, item2, item3}
layout.Layout(root, constraints)
layout.SnapToGrid(items, 10.0, 5.0, 5.0) // Grid at (5, 5) with 10px spacing
```

**Use cases**:
- Snapping to a subgrid within a larger layout
- Offset grid alignment
- Design tool-like subgrid snapping

## Grid Helpers

### Grid

Creates a grid container with the specified number of rows and columns using fixed track sizes.

```go
func Grid(rows, cols int, rowSize, colSize float64) *Node
```

Example:
```go
grid := layout.Grid(4, 4, 150, 200) // 4 rows x 4 columns, rows=150px, cols=200px
grid.Style.GridGap = 10
```

### GridAuto

Creates a grid container with auto-sized tracks.

```go
func GridAuto(rows, cols int) *Node
```

Example:
```go
grid := layout.GridAuto(3, 4) // 3 rows x 4 columns, auto-sized
```

### GridFractional

Creates a grid container with fractional (fr) tracks that share space equally.

```go
func GridFractional(rows, cols int) *Node
```

Example:
```go
grid := layout.GridFractional(2, 3) // 2 rows x 3 columns, all equal fractional units
```

## SVG Helpers

### GetSVGTransform

Gets the SVG transform string for a node's transform.

```go
func GetSVGTransform(node *Node) string
```

### GetFinalRect

Gets the final rectangle after applying transforms.

```go
func GetFinalRect(node *Node) Rect
```

### CollectNodesForSVG

Collects all nodes in a tree for SVG rendering.

```go
func CollectNodesForSVG(root *Node, nodes *[]*Node)
```

Usage:
```go
var nodes []*Node
CollectNodesForSVG(root, &nodes)
```

## Transform Functions

### Translate

Creates a translation transform.

```go
func Translate(x, y float64) Transform
```

### Scale

Creates a scale transform.

```go
func Scale(x, y float64) Transform
```

### Rotate

Creates a rotation transform (radians).

```go
func Rotate(angle float64) Transform
```

### RotateDegrees

Creates a rotation transform (degrees).

```go
func RotateDegrees(angle float64) Transform
```

### SkewX

Creates a horizontal skew transform.

```go
func SkewX(angle float64) Transform
```

### SkewY

Creates a vertical skew transform.

```go
func SkewY(angle float64) Transform
```

### Matrix

Creates a matrix transform.

```go
func Matrix(a, b, c, d, e, f float64) Transform
```

## Enums

### Display

```go
const (
    DisplayBlock Display = iota
    DisplayFlex
    DisplayGrid
    DisplayNone
)
```

### FlexDirection

```go
const (
    FlexDirectionRow FlexDirection = iota
    FlexDirectionRowReverse
    FlexDirectionColumn
    FlexDirectionColumnReverse
)
```

### Position

```go
const (
    PositionStatic Position = iota
    PositionRelative
    PositionAbsolute
    PositionFixed
    PositionSticky
)
```

See godoc for complete enum definitions: `go doc layout`

