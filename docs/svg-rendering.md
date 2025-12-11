# SVG Rendering

This guide shows how to render layouts to SVG, useful for generating images for GitHub READMEs, documentation, or other purposes.

## Overview

The layout library provides helpers for SVG rendering:
- Transform support for visual effects
- Helpers to get final positions and transforms
- Functions to collect nodes for rendering

## Basic SVG Rendering

### Simple Example

```go
package main

import (
    "fmt"
    "github.com/SCKelemen/layout"
)

func main() {
    // Create a grid layout
    root := &layout.Node{
        Style: layout.Style{
            Display: layout.DisplayGrid,
            GridTemplateColumns: []layout.GridTrack{
                layout.FixedTrack(150),
                layout.FixedTrack(150),
            },
            GridGap: 10,
        },
        Children: []*layout.Node{
            {Style: layout.Style{Width: 150, Height: 100}},
            {Style: layout.Style{Width: 150, Height: 100}},
            {Style: layout.Style{Width: 150, Height: 100}},
            {Style: layout.Style{Width: 150, Height: 100}},
        },
    }

    // Perform layout
    constraints := layout.Loose(400, 300)
    size := layout.Layout(root, constraints)

    // Generate SVG
    fmt.Println(`<svg width="` + fmt.Sprintf("%.0f", size.Width) + `" height="` + fmt.Sprintf("%.0f", size.Height) + `">`)
    
    for _, child := range root.Children {
        rect := child.Rect
        fmt.Printf(`  <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="#e0e0e0" stroke="#333"/>\n`,
            rect.X, rect.Y, rect.Width, rect.Height)
    }
    
    fmt.Println(`</svg>`)
}
```

## Using Transforms

### Transform Example

```go
// Create a node with a transform
node := &layout.Node{
    Style: layout.Style{
        Width:  100,
        Height: 100,
        Transform: layout.RotateDegrees(45),
    },
}

// Get the SVG transform string
transformStr := layout.GetSVGTransform(node)
// Returns: "rotate(45 50 50)" (rotate around center)

// Get the final rectangle after transform
finalRect := layout.GetFinalRect(node)
```

### Transform Types

```go
import "math"

// Translation
transform := layout.Translate(10, 20)

// Scaling
transform := layout.Scale(1.5, 1.5)

// Rotation (radians)
transform := layout.Rotate(math.Pi / 4)

// Rotation (degrees)
transform := layout.RotateDegrees(45)

// Skew
transform := layout.SkewX(10)
transform := layout.SkewY(10)

// Matrix
transform := layout.Matrix(1, 0, 0, 1, 10, 20)
```

## Complete SVG Example

See `examples/cards/main.go` for a complete example of rendering card layouts to SVG.

### Key Functions

#### GetSVGTransform

Gets the SVG transform string for a node's transform.

```go
func GetSVGTransform(node *Node) string
```

Returns an SVG transform string like `"translate(10, 20) rotate(45)"` or empty string if no transform.

#### GetFinalRect

Gets the final rectangle after applying transforms.

```go
func GetFinalRect(node *Node) Rect
```

This accounts for the transform when calculating the bounding box.

#### CollectNodesForSVG

Collects all nodes in a tree for SVG rendering.

```go
func CollectNodesForSVG(root *Node, nodes *[]*Node)
```

Usage:
```go
var nodes []*Node
CollectNodesForSVG(root, &nodes)
```

Useful for traversing the entire tree to render all nodes.

## SVG Rendering Pattern

### Recommended Pattern

```go
func RenderToSVG(root *layout.Node, constraints layout.Constraints) string {
    // Perform layout
    size := layout.Layout(root, constraints)
    
    // Start SVG
    svg := fmt.Sprintf(`<svg width="%.0f" height="%.0f" xmlns="http://www.w3.org/2000/svg">\n`,
        size.Width, size.Height)
    
    // Collect all nodes
    var nodes []*layout.Node
    layout.CollectNodesForSVG(root, &nodes)
    
    // Render each node
    for _, node := range nodes {
        rect := node.Rect
        transform := layout.GetSVGTransform(node)
        
        // Build attributes
        attrs := fmt.Sprintf(`x="%.2f" y="%.2f" width="%.2f" height="%.2f"`,
            rect.X, rect.Y, rect.Width, rect.Height)
        
        if transform != "" {
            attrs += fmt.Sprintf(` transform="%s"`, transform)
        }
        
        svg += fmt.Sprintf(`  <rect %s fill="#e0e0e0" stroke="#333"/>\n`, attrs)
    }
    
    svg += `</svg>`
    return svg
}
```

## Tips

1. **Use GetFinalRect** when you need the bounding box after transforms
2. **Apply transforms in SVG** rather than manually calculating positions
3. **Use CollectNodesForSVG** to traverse the entire tree
4. **Handle positioned elements** with `LayoutWithPositioning` if needed
5. **Account for padding/borders** when rendering if they're part of your design

## Example: Card Layout to SVG

```go
func CardLayoutToSVG(cards []Card) string {
    // Create grid layout
    root := &layout.Node{
        Style: layout.Style{
            Display: layout.DisplayGrid,
            GridTemplateColumns: []layout.GridTrack{
                layout.FixedTrack(200),
                layout.FixedTrack(200),
                layout.FixedTrack(200),
            },
            GridGap: 20,
            Padding: layout.Uniform(20),
        },
    }
    
    // Add cards
    for _, card := range cards {
        node := &layout.Node{
            Style: layout.Style{
                Width:  200,
                Height: 150,
            },
        }
        root.Children = append(root.Children, node)
    }
    
    // Layout
    constraints := layout.Loose(800, 600)
    size := layout.Layout(root, constraints)
    
    // Render to SVG
    return RenderToSVG(root, constraints)
}
```

This pattern works well for generating images for GitHub READMEs or documentation.

