# Usage Patterns

Different ways to build layout trees. All snippets assume `import "github.com/SCKelemen/layout"`.

## Pattern 1: direct `Node` construction (CSS-like)

The most explicit form; every property is visible.

```go
root := &layout.Node{
	Style: layout.Style{
		Display:       layout.DisplayFlex,
		FlexDirection: layout.FlexDirectionRow,
		Padding:       layout.Uniform(layout.Px(10)),
	},
	Children: []*layout.Node{
		{Style: layout.Style{Width: layout.Px(100), Height: layout.Px(50)}},
	},
}
size := layout.LayoutSimple(root, layout.Loose(800, 600))
_ = size
```

**Use when** you need control over every property.

## Pattern 2: high-level helpers (SwiftUI/Flutter-like)

```go
root := layout.HStack(
	layout.Fixed(100, 50),
	layout.Spacer(),
	layout.Text("Right-aligned label"),
)
layout.Padding(root, 10)
layout.LayoutSimple(root, layout.Loose(800, 600))
```

`HStack`, `VStack`, `ZStack`, `Grid`, `GridAuto`, `GridFractional`, `Fixed`, `Spacer`, `Text`, plus the mutating helpers `Padding`, `PaddingCustom`, `Margin`, `Frame`, `FrameLength`, `MinWidth`, `MinHeight`, `AspectRatio`, `PlaceInArea`, and the intrinsic sizing helpers.

**Use when** you want short, readable code for common layouts.

## Pattern 3: fluent, immutable API

```go
card := layout.VStack().
	WithWidthLength(layout.Px(300)).
	WithPadding(16).
	WithMargin(8).
	AddChildren(
		layout.Text("Title").WithHeight(32),
		layout.Text("Body").WithFlexGrow(1),
	)

compact := card.WithPadding(8) // card is unchanged
```

See the [Fluent API guide](fluent-api.md).

**Use when** you build variants of a tree or query and transform existing trees.

## Pattern 4: embedding `Node` in your types

```go
type Card struct {
	layout.Node
	Title   string
	Content string
}

func NewCard(title, content string) *Card {
	card := &Card{Title: title, Content: content}
	card.Style.Width = layout.Px(200)
	card.Style.Height = layout.Px(150)
	card.Style.Padding = layout.Uniform(layout.Px(10))
	return card
}
```

```go
cards := []*Card{NewCard("Card 1", "Content 1"), NewCard("Card 2", "Content 2")}
nodes := make([]*layout.Node, len(cards))
for i, card := range cards {
	nodes[i] = &card.Node
}
root := layout.HStack(nodes...)
layout.LayoutSimple(root, layout.Loose(800, 600))
// cards[0].Rect is populated because it is the same Node
```

**Use when** layout data and domain data belong together.

## Pattern 5: a builder for your domain

```go
type LayoutBuilder struct {
	node *layout.Node
}

func NewBuilder() *LayoutBuilder {
	return &LayoutBuilder{node: &layout.Node{}}
}

func (b *LayoutBuilder) Flex() *LayoutBuilder {
	b.node.Style.Display = layout.DisplayFlex
	return b
}

func (b *LayoutBuilder) Row() *LayoutBuilder {
	b.node.Style.FlexDirection = layout.FlexDirectionRow
	return b
}

func (b *LayoutBuilder) Padding(p layout.Length) *LayoutBuilder {
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
```

```go
root := NewBuilder().
	Flex().
	Row().
	Padding(layout.Em(1)).
	AddChild(layout.Fixed(100, 50)).
	AddChild(layout.Fixed(100, 50)).
	Build()
```

**Use when** your application has a fixed vocabulary of layouts.

## Pattern 6: functional options

```go
type Option func(*layout.Node)

func WithPadding(p layout.Length) Option {
	return func(n *layout.Node) { n.Style.Padding = layout.Uniform(p) }
}

func WithWidth(w layout.Length) Option {
	return func(n *layout.Node) { n.Style.Width = w }
}

func NewNode(opts ...Option) *layout.Node {
	node := &layout.Node{}
	for _, opt := range opts {
		opt(node)
	}
	return node
}
```

```go
root := NewNode(WithPadding(layout.Px(10)), WithWidth(layout.Px(200)))
```

**Use when** you want optional configuration with defaults.

## Pattern 7: serialized trees

Layouts can be described as JSON or YAML and loaded with the `serialize` package, which is handy for fixtures, design tools, and debugging:

```go
data, err := serialize.ToJSON(root) // after layout, includes every Rect
if err != nil {
	log.Fatal(err)
}
restored, err := serialize.FromJSON(data)
if err != nil {
	log.Fatal(err)
}
_ = restored
```

(`serialize` is `github.com/SCKelemen/layout/serialize`.) See [serialize/README.md](../serialize/README.md).

## Recommendations

1. Simple layouts: helpers (`HStack`, `VStack`, `Text`).
2. Complex layouts: direct `Node` construction.
3. Variants and analysis: the fluent API.
4. Reusable components: embed `Node`.

## Best practices

1. **Lay out once per frame.** `Rect` and `TextLayout` are outputs; changing a style afterwards does not update them until you call layout again.
2. **Do not share a node between two parents.** Layout writes into the node; use `CloneDeep` for copies.
3. **Choose the right entry point.** `LayoutSimple` for plain trees, `Layout` with a `LayoutContext` for viewport/font/text-metrics control, `LayoutWithPositioning` when any node is positioned.
4. **Pick constraints deliberately.** `Tight` for a fixed canvas, `Loose(w, Unbounded)` for content-sized height, `Unconstrained()` for intrinsic measurement.
5. **Prefer `Length` constructors over `Length{}` literals**, and remember that unset means auto while `Px(0)` is zero.
