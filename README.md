# Layout

A pure Go implementation of CSS Grid and CSS Flexbox layout engines. This library provides a reusable layout system that can be used for terminal UIs (like Bubble Tea), web layouts, or offscreen rendering.

## Features

- **CSS Flexbox Layout** ([Specification](https://www.w3.org/TR/css-flexbox-1/) | [MDN Guide](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_flexible_box_layout)): Complete flexbox implementation with support for:
  - Flex direction (row, column, reverse)
  - Flex wrap
  - Justify content
  - Align items
  - Flex grow/shrink/basis
  - Gap spacing

- **CSS Grid Layout** ([Specification](https://www.w3.org/TR/css-grid-1/) | [MDN Guide](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_grid_layout)): Grid layout implementation with support for:
  - **Multiple columns** via GridTemplateColumns array
  - Grid template rows/columns
  - Auto rows/columns
  - Fractional units (fr)
  - Min/max track sizing
  - Grid gaps
  - Grid item positioning and spanning
  - **Bento box / mosaic layouts** - items spanning multiple rows/columns

- **Block Layout** ([MDN Guide](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_display)): Basic block layout for non-flex/grid elements

- **Aspect Ratio** ([CSS Spec](https://www.w3.org/TR/css-sizing-4/#aspect-ratio) | [MDN Guide](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_sizing)): Maintain consistent width-to-height ratios for responsive elements
  - Helps elements reserve space correctly when one dimension is auto

- **Post-Layout Alignment & Distribution** ([CSS Spec](https://www.w3.org/TR/css-align-3/) | [MDN Guide](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_alignment)): Design-tool-like operations for aligning and distributing nodes after layout:
  - Align to edges (left, right, top, bottom) or centers
  - Distribute with even spacing
  - Snap to grid boundaries (primarily for block/absolute layouts)

- **Box Model** ([MDN Guide](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_model)): Full support for padding, margin, and border spacing

- **Positioned Layout** ([MDN Guide](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_positioned_layout)): Absolute, relative, fixed, and sticky positioning

- **Serialization** (optional `serialize` package): JSON/YAML serialization for debugging and persistence
  - Inspect layout trees
  - Save and load layout configurations
  - Useful for testing and documentation

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

## Fluent API

The library provides a **Roslyn-style fluent API** for working with layout trees. This API offers immutable operations, powerful querying, and elegant tree transformations.

### Two API Styles

You can use either the classic mutable style or the new fluent immutable style:

```go
// Classic style (still supported)
node := &layout.Node{Style: layout.Style{Width: 100}}
layout.Padding(node, 10)

// Fluent style
node := (&layout.Node{}).WithWidth(100).WithPadding(10)
```

### Navigation & Querying

Find and traverse nodes in your layout tree:

```go
root := layout.HStack(
    layout.Fixed(100, 50).WithText("Item 1"),
    layout.Fixed(200, 50).WithText("Item 2"),
    layout.Fixed(150, 50).WithText("Item 3"),
)

// Find all nodes with text
textNodes := root.FindAll(func(n *layout.Node) bool {
    return n.Text != ""
})

// Find first wide node
wide := root.Find(func(n *layout.Node) bool {
    return n.Style.Width > 150
})

// Check if any child is flex
hasFlex := root.Any(func(n *layout.Node) bool {
    return n.Style.Display == layout.DisplayFlex
})

// Get all descendants
allNodes := root.Descendants()

// Filter by display type
grids := root.OfDisplayType(layout.DisplayGrid)
```

### Immutable Modifications

Create modified copies without changing the original:

```go
original := layout.HStack(
    layout.Fixed(100, 50),
    layout.Fixed(200, 50),
)

// Create variants without modifying original
padded := original.WithPadding(16)
withMargin := original.WithMargin(8)
wider := original.WithWidth(500)

// Method chaining
styled := original.
    WithPadding(16).
    WithMargin(8).
    WithDisplay(layout.DisplayFlex).
    AddChild(layout.Fixed(100, 50))

// Original unchanged
fmt.Printf("Original padding: %.0f\n", original.Style.Padding.Top) // 0
fmt.Printf("Variant padding: %.0f\n", padded.Style.Padding.Top)    // 16
```

### Parent Navigation with Context

Walk up the tree to find ancestors:

```go
root := layout.VStack(
    layout.HStack(
        layout.Fixed(100, 50).WithText("Target"),
    ),
)

// Wrap root in context for parent tracking
ctx := layout.NewContext(root)

// Find a node and walk up to find container
targetCtx := ctx.FindDown(func(n *layout.Node) bool {
    return n.Text == "Target"
})

// Find containing flex container
flexCtx := targetCtx.FindUp(func(n *layout.Node) bool {
    return n.Style.Display == layout.DisplayFlex
})

// Get all ancestors
ancestors := targetCtx.Ancestors()
fmt.Printf("Found %d ancestors\n", len(ancestors))

// Get depth in tree
depth := targetCtx.Depth()
fmt.Printf("Node is %d levels deep\n", depth)

// Get siblings
siblings := targetCtx.Siblings()
```

### Transformations

Apply operations across the tree:

```go
root := layout.HStack(
    layout.Fixed(100, 50),
    layout.Fixed(200, 100),
    layout.Fixed(150, 75),
)

// Transform: Selectively modify nodes
doubled := root.Transform(
    func(n *layout.Node) bool {
        return n.Style.Width > 0 && n.Style.Width < 200
    },
    func(n *layout.Node) *layout.Node {
        return n.WithWidth(n.Style.Width * 2)
    },
)

// Map: Apply to all nodes
scaled := root.Map(func(n *layout.Node) *layout.Node {
    return n.
        WithWidth(n.Style.Width * 1.5).
        WithHeight(n.Style.Height * 1.5)
})

// Filter: Keep only matching children (shallow)
wide := root.Filter(func(n *layout.Node) bool {
    return n.Style.Width >= 150
})

// FilterDeep: Recursive filtering
visible := root.FilterDeep(func(n *layout.Node) bool {
    return n.Style.Display != layout.DisplayNone
})

// Fold: Reduce tree to single value
totalWidth := root.Fold(0.0, func(acc interface{}, n *layout.Node) interface{} {
    return acc.(float64) + n.Style.Width
}).(float64)

count := root.Fold(0, func(acc interface{}, n *layout.Node) interface{} {
    return acc.(int) + 1
}).(int)

// FoldWithContext: With depth information
depthMap := root.FoldWithContext(
    make(map[int]int),
    func(acc interface{}, n *layout.Node, depth int) interface{} {
        m := acc.(map[int]int)
        m[depth]++
        return m
    },
).(map[int]int)
```

### Practical Examples

#### Building a Card Layout with Fluent API

```go
func CreateCard(title, body string, width float64) *layout.Node {
    return layout.VStack().
        WithWidth(width).
        WithPadding(16).
        WithMargin(8).
        AddChildren(
            layout.Fixed(0, 32).WithText(title),
            layout.Fixed(0, 0).WithText(body),
        )
}

// Create multiple cards
cards := []*layout.Node{
    CreateCard("Title 1", "Body 1", 200),
    CreateCard("Title 2", "Body 2", 200),
    CreateCard("Title 3", "Body 3", 200),
}

container := layout.HStack().
    WithPadding(20).
    AddChildren(cards...)
```

#### Conditional Styling

```go
func ApplyTheme(root *layout.Node, darkMode bool) *layout.Node {
    padding := 8.0
    margin := 4.0

    if darkMode {
        padding = 12.0
        margin = 6.0
    }

    return root.Map(func(n *layout.Node) *layout.Node {
        return n.WithPadding(padding).WithMargin(margin)
    })
}

lightTheme := ApplyTheme(root, false)
darkTheme := ApplyTheme(root, true)
```

#### Statistics and Analysis

```go
// Count nodes by display type
displayCounts := root.FoldWithContext(
    make(map[layout.Display]int),
    func(acc interface{}, n *layout.Node, depth int) interface{} {
        m := acc.(map[layout.Display]int)
        m[n.Style.Display]++
        return m
    },
).(map[layout.Display]int)

// Find maximum depth
maxDepth := root.FoldWithContext(
    0,
    func(acc interface{}, n *layout.Node, depth int) interface{} {
        current := acc.(int)
        if depth > current {
            return depth
        }
        return current
    },
).(int)

// Sum all padding
totalPadding := root.Fold(0.0, func(acc interface{}, n *layout.Node) interface{} {
    sum := acc.(float64)
    return sum + n.Style.Padding.Top + n.Style.Padding.Right +
           n.Style.Padding.Bottom + n.Style.Padding.Left
}).(float64)
```

#### Tree Manipulation

```go
// Remove all hidden nodes
visible := root.FilterDeep(func(n *layout.Node) bool {
    return n.Style.Display != layout.DisplayNone
})

// Add padding to all containers
padded := root.Transform(
    func(n *layout.Node) bool {
        return len(n.Children) > 0
    },
    func(n *layout.Node) *layout.Node {
        return n.WithPadding(10)
    },
)

// Clone tree and modify
variant := root.CloneDeep().
    WithPadding(20).
    Map(func(n *layout.Node) *layout.Node {
        return n.WithMargin(5)
    })
```

## Documentation

- [Getting Started](docs/getting-started.md) - Installation and quick examples
- [Layout Systems](docs/layout-systems.md) - Flexbox, Grid, Block, and Positioned layouts
- [API Reference](docs/api-reference.md) - Complete API documentation
- [Usage Patterns](docs/usage-patterns.md) - Different ways to use the library
- [Common Gotchas](docs/gotchas.md) - Common pitfalls and how to avoid them ⚠️
- [SVG Rendering](docs/svg-rendering.md) - Rendering layouts to SVG
- [Limitations](docs/limitations.md) - Known limitations and design decisions
- [WPT Sync](docs/wpt-sync.md) - Web Platform Tests integration and tracking

## Learning Resources

This library implements CSS specifications. For deeper understanding of layout concepts, see these MDN guides:

- [CSS Flexible Box Layout](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_flexible_box_layout) - Learn about flexbox
- [CSS Grid Layout](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_grid_layout) - Learn about grid layout
- [CSS Box Alignment](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_alignment) - Alignment in flexbox and grid
- [CSS Box Model](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_model) - Padding, margin, and borders
- [CSS Box Sizing](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_sizing) - Box sizing and aspect ratios
- [CSS Positioned Layout](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_positioned_layout) - Absolute, relative, fixed positioning
- [CSS Display](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_display) - Display modes and formatting contexts

## Use Cases

- **Terminal UIs**: Use with Bubble Tea or other TUI libraries
- **SVG Rendering**: Generate card layouts and graphs for images
- **Web Layouts**: Server-side layout generation
- **PDF Generation**: Layout content for PDFs
- **Game UIs**: Layout game interface elements
- **Offscreen Rendering**: Layout for image generation

## Testing & Quality

- **321/321 tests passing (100%)** 🎉
- Comprehensive CSS spec compliance
- [Spec Compliance Status](SPEC_COMPLIANCE_STATUS.md)
- [Specification Gaps](SPECIFICATION_GAPS.md)
- Weekly [Web Platform Tests sync](.github/workflows/wpt-sync.yml) via CI

Run tests:
```bash
go test -v ./...
```

## License

MIT


## WPT Testing with CEL Assertions

This library integrates with [wpt-test-gen](https://github.com/SCKelemen/wpt-test-gen) for Web Platform Test-style testing using [CEL (Common Expression Language)](https://github.com/google/cel-spec) assertions.

### Using CEL Assertions in Tests

```go
import (
    "testing"
    "github.com/SCKelemen/layout"
    "github.com/SCKelemen/wpt-test-gen/pkg/cel"
)

func TestFlexboxLayout(t *testing.T) {
    // Build your layout
    root := &layout.Node{
        Style: layout.Style{
            Display:        layout.DisplayFlex,
            JustifyContent: layout.JustifyContentSpaceBetween,
            Width:          600,
            Height:         100,
        },
        Children: []*layout.Node{
            {Style: layout.Style{Width: 100, Height: 50}},
            {Style: layout.Style{Width: 100, Height: 50}},
        },
    }

    // Run layout
    layout.Layout(root, layout.Tight(600, 100))

    // Create CEL environment
    env, _ := cel.NewLayoutCELEnv(root)

    // Define assertions using CEL expressions
    assertions := []cel.CELAssertion{
        {
            Expression: "getX(child(root(), 0)) == 0.0",
            Message:    "first-child-at-start",
        },
        {
            Expression: "getRight(child(root(), 1)) == getWidth(root())",
            Message:    "last-child-at-end",
        },
    }

    // Evaluate assertions
    results := env.EvaluateAll(assertions)

    for _, result := range results {
        if !result.Passed {
            t.Errorf("Assertion '%s' failed: %s", result.Assertion.Message, result.Error)
        }
    }
}
```

### Available CEL Functions

- **Node access**: `root()`, `child(node, index)`, `parent(node)`
- **Position**: `getX(node)`, `getY(node)`, `getTop(node)`, `getLeft(node)`
- **Size**: `getWidth(node)`, `getHeight(node)`
- **Edges**: `getRight(node)`, `getBottom(node)`

### Language-Agnostic Testing

For testing from other languages (JavaScript, Python, Rust, etc.), use the `wptest eval` command:

```bash
# Install wptest CLI
go install github.com/SCKelemen/wpt-test-gen/cmd/wptest@latest

# Test via JSON stdin/stdout
echo '{
  "layout": {"display": "flex", "width": 600, ...},
  "assertions": [{"expression": "getX(root()) == 0.0", ...}]
}' | wptest eval
```

See [wpt-test-gen examples](https://github.com/SCKelemen/wpt-test-gen/tree/main/examples/cross-language) for JavaScript, Python, and Rust examples.

### Example Test

See [layout_wpt_example_test.go](layout_wpt_example_test.go) for complete examples of testing flexbox and grid layouts with CEL assertions.

