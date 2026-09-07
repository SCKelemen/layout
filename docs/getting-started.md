# Getting Started

`github.com/SCKelemen/layout` computes CSS-style layout for a tree of nodes. You build the tree, call a layout function with the available space, and read back a `Rect` per node. Nothing is drawn; rendering is your code.

## Installation

```bash
go get github.com/SCKelemen/layout
```

## Core concepts

### Nodes and styles

A `Node` has a `Style`, a `Children` slice, and (for text) a `Text` string. After layout, `node.Rect` holds the position relative to the parent and the size.

```go
root := &layout.Node{
	Style: layout.Style{
		Display: layout.DisplayFlex,
		Padding: layout.Uniform(layout.Px(10)),
	},
	Children: []*layout.Node{
		{Style: layout.Style{Width: layout.Px(100), Height: layout.Px(50)}},
		{Style: layout.Style{Width: layout.Px(100), Height: layout.Px(50)}},
	},
}
```

`Style.Display` picks the algorithm: `DisplayBlock` (the zero value), `DisplayFlex`, `DisplayGrid`, `DisplayInlineText` (text leaves), or `DisplayNone`.

### Lengths

Every size, offset, gap, padding, margin, and border is a `Length` (an alias of `units.Length`), created with a constructor:

| Constructor | Unit |
|-------------|------|
| `Px`, `Pt`, `Pc`, `In`, `Cm`, `Mm`, `Q` | absolute |
| `Em`, `Rem`, `Ch` | font-relative |
| `Vw`, `Vh`, `Vmin`, `Vmax` | viewport-relative |
| `Cqw`, `Cqh`, `Cqi`, `Cqb`, `Cqmin`, `Cqmax` | container-relative (see [Layout systems](layout-systems.md#container-queries)) |

Two rules matter everywhere:

- **Unset means `auto`.** A `Length` you never assign (`Unit == ""`) is the CSS initial value: `auto` for `Width`/`Height`/`FlexBasis` and positioning offsets, "fall back to the shorthand" for `GridRowGap`/`FlexColumnGap`, 0 for padding, margin, and border.
- **`Px(0)` is a real zero.** `Width: layout.Px(0)` is a zero-width box, not auto.

Read a resolved value with `.Value`; convert any unit with `layout.ResolveLength(l, ctx, fontSize)`.

### Constraints

Constraints describe the space the root may use:

- `Loose(w, h)`: at most `w` x `h`, may be smaller.
- `Tight(w, h)`: exactly `w` x `h`.
- `Unconstrained()`: no limit (`layout.Unbounded` on both axes).

An unbounded axis is "indefinite" in CSS terms: auto-sized boxes take their content size there.

### Running layout

```go
size := layout.LayoutSimple(root, layout.Loose(800, 600))
fmt.Printf("%.0f x %.0f\n", size.Width, size.Height)
fmt.Printf("first child at (%.0f, %.0f)\n", root.Children[0].Rect.X, root.Children[0].Rect.Y)
```

`LayoutSimple` builds a `LayoutContext` from the constraints (viewport = `MaxWidth` x `MaxHeight`, root font 16px). When the constraint is unbounded the viewport dimension is 0, so `vw`/`vh` resolve to 0 on that axis.

Use `Layout` with your own context to control unit resolution:

```go
ctx := layout.NewLayoutContext(1920, 1080, 16) // viewport width, viewport height, root font size (px)
ctx = ctx.WithTextMetrics(layout.NewTerminalTextMetrics())

node := &layout.Node{Style: layout.Style{
	Width:   layout.Vw(50),  // 960
	Height:  layout.Rem(10), // 160
	Padding: layout.Uniform(layout.Em(1)),
}}
layout.Layout(node, layout.Loose(2000, 2000), ctx)
fmt.Printf("%.0f x %.0f\n", node.Rect.Width, node.Rect.Height) // 992 x 192
```

Positioned elements (`Position: PositionAbsolute`, `PositionFixed`, ...) need the two-pass `LayoutWithPositioning(root, constraints, viewportRect, ctx)`.

### High-level helpers

```go
root := layout.HStack(
	layout.Fixed(100, 50),
	layout.Spacer(),
	layout.Fixed(100, 50),
)
layout.LayoutSimple(root, layout.Loose(800, 600))
// the spacer grows to 600px; the last item starts at x = 700
```

`HStack`/`VStack` are flex containers, `ZStack` overlaps absolutely positioned children, `Fixed` sets pixel sizes, `Spacer` is `FlexGrow: 1`, and `Grid`/`GridAuto`/`GridFractional` build grids. `Padding`, `Margin`, `Frame`, `FrameLength`, `MinWidth`, `MinHeight`, and `AspectRatio` mutate a node and return it.

### Text

```go
title := layout.Text("Hello, world", layout.Style{
	TextStyle: &layout.TextStyle{FontSize: 20, TextAlign: layout.TextAlignCenter},
})
```

`Text` returns a `DisplayInlineText` leaf with unset width and height, so it sizes to its content inside any container. Layout fills `title.TextLayout` with lines and inline boxes. See [Text](text.md).

### Fluent API

The same tree can be built and queried immutably:

```go
card := layout.VStack().
	WithWidthLength(layout.Px(300)).
	WithPadding(16).
	AddChildren(layout.Text("Title"), layout.Text("Body"))

titles := card.FindAll(func(n *layout.Node) bool { return n.Text == "Title" })
```

See the [Fluent API guide](fluent-api.md).

## Next steps

- [Layout systems](layout-systems.md): Flexbox, Grid, Block, Text, positioned layout, container queries, intrinsic sizing
- [API reference](api-reference.md): the exported surface, grouped by topic
- [Gotchas](gotchas.md): unset vs `Px(0)`, gap axes, margin collapsing, and other surprises
- [Limitations](limitations.md): what is not implemented
