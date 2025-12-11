# Usage Patterns

This guide shows different ways to use the layout library in your Go code.

## Pattern 1: Direct Node Creation (Low-level, CSS-like)

This is the most flexible approach, similar to CSS:

```go
import "github.com/SCKelemen/layout"

root := &layout.Node{
    Style: layout.Style{
        Display: layout.DisplayFlex,
        FlexDirection: layout.FlexDirectionRow,
        Padding: layout.Uniform(10),
    },
    Children: []*layout.Node{
        {
            Style: layout.Style{
                Width:  100,
                Height: 50,
            },
        },
    },
}

constraints := layout.Loose(800, 600)
size := layout.Layout(root, constraints)
```

**Use when**: You need precise control over all layout properties.

## Pattern 2: High-level API (SwiftUI/Flutter-like)

Use the helper functions for simpler, more ergonomic code:

```go
import "github.com/SCKelemen/layout"

root := layout.HStack(
    layout.Fixed(100, 50),
    layout.Spacer(),
    layout.Fixed(100, 50),
)

constraints := layout.Loose(800, 600)
size := layout.Layout(root, constraints)
```

**Use when**: You want simple, readable code for common layouts.

## Pattern 3: Embedding Node in Your Types

You can embed `layout.Node` in your own types to add domain-specific data:

```go
type Card struct {
    layout.Node
    Title   string
    Content string
}

func NewCard(title, content string) *Card {
    card := &Card{
        Title:   title,
        Content: content,
    }
    card.Style.Width = 200
    card.Style.Height = 150
    card.Style.Padding = layout.Uniform(10)
    return card
}

// Usage
cards := []*Card{
    NewCard("Card 1", "Content 1"),
    NewCard("Card 2", "Content 2"),
}

// Convert to layout nodes
nodes := make([]*layout.Node, len(cards))
for i, card := range cards {
    nodes[i] = &card.Node
}

root := layout.HStack(nodes...)
```

**Use when**: You want to combine layout with domain-specific data.

## Pattern 4: Builder Pattern

Create a builder for your specific use case:

```go
type LayoutBuilder struct {
    node *layout.Node
}

func NewBuilder() *LayoutBuilder {
    return &LayoutBuilder{
        node: &layout.Node{},
    }
}

func (b *LayoutBuilder) Flex() *LayoutBuilder {
    b.node.Style.Display = layout.DisplayFlex
    return b
}

func (b *LayoutBuilder) Row() *LayoutBuilder {
    b.node.Style.FlexDirection = layout.FlexDirectionRow
    return b
}

func (b *LayoutBuilder) Padding(p float64) *LayoutBuilder {
    b.node.Style.Padding = layout.Uniform(p)
    return b
}

func (b *LayoutBuilder) AddChild(child *layout.Node) *LayoutBuilder {
    b.node.Children = append(b.node.Children, child)
    return b
}

func (b *LayoutBuilder) Build() *layout.Node {
    return b.node
}

// Usage
root := NewBuilder().
    Flex().
    Row().
    Padding(10).
    AddChild(layout.Fixed(100, 50)).
    AddChild(layout.Fixed(100, 50)).
    Build()
```

**Use when**: You want a fluent API for building layouts.

## Pattern 5: Functional Options

Use functional options for configuration:

```go
type Option func(*layout.Node)

func WithPadding(p float64) Option {
    return func(n *layout.Node) {
        n.Style.Padding = layout.Uniform(p)
    }
}

func WithWidth(w float64) Option {
    return func(n *layout.Node) {
        n.Style.Width = w
    }
}

func NewNode(opts ...Option) *layout.Node {
    node := &layout.Node{}
    for _, opt := range opts {
        opt(node)
    }
    return node
}

// Usage
root := NewNode(
    WithPadding(10),
    WithWidth(200),
)
```

**Use when**: You want flexible configuration with optional parameters.

## Pattern 6: Domain-Specific Wrappers

Create domain-specific wrappers for your use case:

```go
// For GitHub README card layouts
type CardLayout struct {
    *layout.Node
}

func NewCardLayout() *CardLayout {
    return &CardLayout{
        Node: &layout.Node{
            Style: layout.Style{
                Display: layout.DisplayGrid,
                GridTemplateColumns: []layout.GridTrack{
                    layout.FixedTrack(150),
                    layout.FixedTrack(150),
                },
                GridGap: 20,
            },
        },
    }
}

func (c *CardLayout) AddCard(title string) *CardLayout {
    card := &layout.Node{
        Style: layout.Style{
            Width:  150,
            Height: 100,
        },
    }
    c.Children = append(c.Children, card)
    return c
}
```

**Use when**: You have a specific domain (e.g., card layouts) with common patterns.

## Recommendations

1. **For simple layouts**: Use the high-level API (`HStack`, `VStack`, `Spacer`)
2. **For complex layouts**: Use direct Node creation with CSS-like properties
3. **For reusable components**: Embed Node in your types
4. **For domain-specific needs**: Create custom builders or wrappers

## Best Practices

1. **Don't modify nodes after layout**: Once `Layout()` is called, the `Rect` field contains the computed layout. Modifying styles after layout won't update positions.

2. **Reuse nodes carefully**: If you reuse the same node in multiple places, create new instances or deep copy.

3. **Handle positioned elements**: Use `LayoutWithPositioning()` if you have absolutely/fixed positioned elements.

4. **For SVG rendering**: Use `GetSVGTransform()` and `GetFinalRect()` to get final positions and transforms.

5. **Use appropriate constraints**: 
   - `Loose()` for maximum size
   - `Tight()` for exact size
   - `Unconstrained()` when size doesn't matter

