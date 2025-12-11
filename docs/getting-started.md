# Getting Started

A pure Go implementation of CSS Grid and CSS Flexbox layout engines. This library provides a reusable layout system that can be used for terminal UIs (like Bubble Tea), web layouts, SVG rendering, or offscreen rendering.

**Specifications**:
- [CSS Flexible Box Layout Module Level 1](https://www.w3.org/TR/css-flexbox-1/)
- [CSS Grid Layout Module Level 1](https://www.w3.org/TR/css-grid-1/)

## Installation

```bash
go get github.com/SCKelemen/layout
```

## Quick Start

### Basic Flexbox Example

```go
package main

import (
    "fmt"
    "github.com/SCKelemen/layout"
)

func main() {
    // Create a flex container
    root := &layout.Node{
        Style: layout.Style{
            Display:        layout.DisplayFlex,
            FlexDirection:  layout.FlexDirectionRow,
            JustifyContent: layout.JustifyContentSpaceBetween,
            AlignItems:     layout.AlignItemsCenter,
            Padding:        layout.Uniform(10),
        },
        Children: []*layout.Node{
            {
                Style: layout.Style{
                    Width:  100,
                    Height: 50,
                },
            },
            {
                Style: layout.Style{
                    Width:  100,
                    Height: 50,
                },
            },
        },
    }

    // Perform layout
    constraints := layout.Loose(800, 600)
    size := layout.Layout(root, constraints)

    fmt.Printf("Layout size: %.2f x %.2f\n", size.Width, size.Height)
    fmt.Printf("Child 1 position: (%.2f, %.2f)\n", 
        root.Children[0].Rect.X, root.Children[0].Rect.Y)
}
```

### High-Level API Example

For simpler code, use the high-level API:

```go
root := layout.HStack(
    layout.Fixed(100, 50),
    layout.Spacer(),
    layout.Fixed(100, 50),
)

constraints := layout.Loose(800, 600)
size := layout.Layout(root, constraints)
```

### Grid Example

```go
root := &layout.Node{
    Style: layout.Style{
        Display: layout.DisplayGrid,
        GridTemplateRows: []layout.GridTrack{
            layout.FixedTrack(100),
            layout.FractionTrack(1),
            layout.FixedTrack(50),
        },
        GridTemplateColumns: []layout.GridTrack{
            layout.FractionTrack(1),
            layout.FractionTrack(2),
        },
        GridGap: 10,
        Padding: layout.Uniform(10),
    },
    Children: []*layout.Node{
        {
            Style: layout.Style{
                GridRowStart:    0,
                GridRowEnd:      1,
                GridColumnStart: 0,
                GridColumnEnd:   2,
            },
        },
        {
            Style: layout.Style{
                GridRowStart:    1,
                GridColumnStart: 0,
            },
        },
        {
            Style: layout.Style{
                GridRowStart:    1,
                GridColumnStart: 1,
            },
        },
    },
}

constraints := layout.Loose(600, 400)
layout.Layout(root, constraints)
```

## Core Concepts

### Nodes

A `Node` represents an element in the layout tree. Each node has:
- `Style`: Layout properties (display type, flex/grid properties, sizing, etc.)
- `Rect`: Computed position and size (set after calling `Layout()`)
- `Children`: Child nodes

### Constraints

Constraints define the available space for layout:
- `Loose(width, height)`: Maximum size, can be smaller
- `Tight(width, height)`: Exact size required
- `Unconstrained()`: No size limits

### Layout Process

1. Create your node tree with styles
2. Call `Layout(root, constraints)` to compute positions
3. Access `node.Rect` for each node's position and size

## Next Steps

- [Layout Systems](layout-systems.md) - Learn about Flexbox, Grid, Block, and Positioned layouts
- [API Reference](api-reference.md) - Complete API documentation
- [Usage Patterns](usage-patterns.md) - Different ways to use the library
- [SVG Rendering](svg-rendering.md) - Rendering layouts to SVG
- [Limitations](limitations.md) - Known limitations and design decisions

## Use Cases

- **Terminal UIs**: Use with Bubble Tea or other TUI libraries
- **SVG Rendering**: Generate card layouts and graphs for images
- **Web Layouts**: Server-side layout generation
- **PDF Generation**: Layout content for PDFs
- **Game UIs**: Layout game interface elements
- **Offscreen Rendering**: Layout for image generation

