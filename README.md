# Layout

A pure Go implementation of CSS Grid and CSS Flexbox layout engines. This library provides a reusable layout system that can be used for terminal UIs (like Bubble Tea), web layouts, or offscreen rendering.

## Features

- **CSS Flexbox Layout** ([Specification](https://www.w3.org/TR/css-flexbox-1/)): Complete flexbox implementation with support for:
  - Flex direction (row, column, reverse)
  - Flex wrap
  - Justify content
  - Align items
  - Flex grow/shrink/basis
  - Gap spacing

- **CSS Grid Layout** ([Specification](https://www.w3.org/TR/css-grid-1/)): Grid layout implementation with support for:
  - **Multiple columns** via GridTemplateColumns array
  - Grid template rows/columns
  - Auto rows/columns
  - Fractional units (fr)
  - Min/max track sizing
  - Grid gaps
  - Grid item positioning and spanning
  - **Bento box / mosaic layouts** - items spanning multiple rows/columns

- **Block Layout**: Basic block layout for non-flex/grid elements

- **Post-Layout Alignment & Distribution**: Design-tool-like operations for aligning and distributing nodes after layout:
  - Align to edges (left, right, top, bottom) or centers
  - Distribute with even spacing
  - Snap to grid boundaries (primarily for block/absolute layouts)
  - Based on [CSS Box Alignment Module Level 3](https://www.w3.org/TR/css-align-3/)

## Installation

```bash
go get github.com/SCKelemen/layout
```

## Usage

### Basic Example

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
    fmt.Printf("Child 1 position: (%.2f, %.2f)\n", root.Children[0].Rect.X, root.Children[0].Rect.Y)
}
```

### Flexbox Example

```go
root := &layout.Node{
    Style: layout.Style{
        Display:        layout.DisplayFlex,
        FlexDirection:  layout.FlexDirectionColumn,
        JustifyContent: layout.JustifyContentCenter,
        AlignItems:     layout.AlignItemsStretch,
        Padding:        layout.Uniform(20),
    },
    Children: []*layout.Node{
        {
            Style: layout.Style{
                FlexGrow: 1,
                Height:   100,
            },
        },
        {
            Style: layout.Style{
                FlexGrow: 2,
                Height:   100,
            },
        },
    },
}

constraints := layout.Tight(400, 600)
layout.Layout(root, constraints)
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
                GridRowEnd:      2,
                GridColumnStart: 0,
                GridColumnEnd:   1,
            },
        },
        {
            Style: layout.Style{
                GridRowStart:    1,
                GridRowEnd:      2,
                GridColumnStart: 1,
                GridColumnEnd:   2,
            },
        },
    },
}

constraints := layout.Loose(600, 400)
layout.Layout(root, constraints)
```

## Documentation

- [Getting Started](docs/getting-started.md) - Installation and quick examples
- [Layout Systems](docs/layout-systems.md) - Flexbox, Grid, Block, and Positioned layouts
- [API Reference](docs/api-reference.md) - Complete API documentation
- [Usage Patterns](docs/usage-patterns.md) - Different ways to use the library
- [Common Gotchas](docs/gotchas.md) - Common pitfalls and how to avoid them ⚠️
- [SVG Rendering](docs/svg-rendering.md) - Rendering layouts to SVG
- [Limitations](docs/limitations.md) - Known limitations and design decisions

## Use Cases

- **Terminal UIs**: Use with Bubble Tea or other TUI libraries
- **SVG Rendering**: Generate card layouts and graphs for images
- **Web Layouts**: Server-side layout generation
- **PDF Generation**: Layout content for PDFs
- **Game UIs**: Layout game interface elements
- **Offscreen Rendering**: Layout for image generation

## License

MIT

