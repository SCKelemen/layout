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
    Margin  Spacing  // Note: Not used in layout calculations
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

### Frame

Sets the width and/or height of a node.

```go
func Frame(node *Node, width, height float64) *Node
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

