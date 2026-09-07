# Layout

A pure Go implementation of CSS Flexbox, Grid, Block, positioned, and text layout. The engine computes positions and sizes only; rendering is left to the caller, so the same tree can drive a terminal UI, an SVG, a PDF, or an image.

[![Go Reference](https://pkg.go.dev/badge/github.com/SCKelemen/layout.svg)](https://pkg.go.dev/github.com/SCKelemen/layout)

## Features

- **Flexbox** ([css-flexbox-1](https://www.w3.org/TR/css-flexbox-1/)): direction, wrap, `justify-content`, `align-items`/`align-self`/`align-content`, grow/shrink/basis, gaps, `order`, baseline alignment.
- **Grid** ([css-grid-1](https://www.w3.org/TR/css-grid-1/)): fixed, `fr`, `auto`, `minmax()`, min-/max-/fit-content tracks, `repeat(auto-fill | auto-fit)`, named areas, auto-placement (including dense), spanning, gaps, all alignment properties.
- **Block** ([CSS 2.1 §9–10](https://www.w3.org/TR/CSS21/visuren.html)): vertical flow, margin collapsing, auto sizing, min/max, `box-sizing`, `aspect-ratio`.
- **Text** ([css-text-3](https://www.w3.org/TR/css-text-3/)): UAX #14 line breaking, `white-space`, alignment and justification, `text-indent`, spacing, `text-overflow`, `text-transform`, hanging punctuation, soft hyphens, `direction: rtl`, vertical and sideways writing modes with UAX #50 orientation. Measurement is pluggable (`TextMetricsProvider`).
- **Positioned layout** ([css-position-3](https://www.w3.org/TR/css-position-3/)): relative, absolute, fixed; sticky behaves as relative without a scroll offset.
- **CSS lengths** ([css-values-4](https://www.w3.org/TR/css-values-4/)): every Level 4 length unit via [`github.com/SCKelemen/units`](https://github.com/SCKelemen/units), resolved through a `LayoutContext`; container-query units via `ResolveLengthInContext`.
- **Intrinsic sizing** ([css-sizing-3](https://www.w3.org/TR/css-sizing-3/)): min-content, max-content, fit-content for widths, heights, and grid tracks.
- **Fluent API**: immutable `With*`, `Find*`, `Transform`, `Map`, `Filter`, `FoldNodes`, and `NodeContext` for parent navigation.
- **Serialization** (`serialize` package): JSON and YAML round-tripping of trees with unit-preserving lengths.
- **Post-layout helpers**: `AlignNodes`, `DistributeNodes`, `SnapNodes`, 2D `Transform` for SVG output.

## Installation

```bash
go get github.com/SCKelemen/layout
```

## Quick start

Sizes are `Length` values built with `Px`, `Em`, `Rem`, `Vw`, and friends. A `Length` that is never set means `auto`; `Px(0)` is a real zero. `LayoutSimple` derives a `LayoutContext` (viewport = constraints, 16px root font) for you; use `Layout(root, constraints, ctx)` when you need to control the viewport, root font size, or text metrics.

### Flexbox

```go
package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	root := &layout.Node{
		Style: layout.Style{
			Display:        layout.DisplayFlex,
			JustifyContent: layout.JustifyContentSpaceBetween,
			AlignItems:     layout.AlignItemsCenter,
			Padding:        layout.Uniform(layout.Px(10)),
		},
		Children: []*layout.Node{
			{Style: layout.Style{Width: layout.Px(100), Height: layout.Px(50)}},
			{Style: layout.Style{Width: layout.Px(100), Height: layout.Px(50)}},
		},
	}

	size := layout.LayoutSimple(root, layout.Loose(800, 600))
	fmt.Printf("%.0f x %.0f\n", size.Width, size.Height) // 800 x 600
	fmt.Printf("second child at (%.0f, %.0f)\n",
		root.Children[1].Rect.X, root.Children[1].Rect.Y) // second child at (690, 275)
}
```

### Grid

```go
grid := &layout.Node{
	Style: layout.Style{
		Display: layout.DisplayGrid,
		GridTemplateColumns: []layout.GridTrack{
			layout.FixedTrack(layout.Px(200)), // sidebar
			layout.FractionTrack(1),           // main
		},
		GridTemplateRows: []layout.GridTrack{
			layout.FixedTrack(layout.Px(60)), // header
			layout.FractionTrack(1),
		},
		GridGap: layout.Px(10),
	},
	Children: []*layout.Node{
		{Style: layout.Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 0, GridColumnEnd: 2}}, // header spans both columns
		{Style: layout.Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 0, GridColumnEnd: 1}},
		{Style: layout.Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 1, GridColumnEnd: 2}},
	},
}

layout.LayoutSimple(grid, layout.Tight(600, 400))
// header: 600 x 60 at (0, 0); sidebar: 200 x 330 at (0, 70); main: 390 x 330 at (210, 70)
```

### Text

```go
text := layout.Text("The quick brown fox jumps over the lazy dog", layout.Style{
	Width: layout.Px(120),
	TextStyle: &layout.TextStyle{
		FontSize:   16,
		LineHeight: 1.5,
		TextAlign:  layout.TextAlignCenter,
	},
})

size := layout.LayoutSimple(text, layout.Loose(400, 400))
fmt.Printf("%.0f x %.0f, %d lines\n", size.Width, size.Height, len(text.TextLayout.Lines))
// 120 x 96, 4 lines
for _, line := range text.TextLayout.Lines {
	fmt.Printf("y=%.0f x=%.1f width=%.1f\n", line.OffsetY, line.OffsetX, line.Width)
}
```

`layout.Text` creates a `DisplayInlineText` leaf; `node.TextLayout` holds the line boxes for your renderer. For accurate Unicode widths (CJK, emoji, terminal cells) install `layout.NewTerminalTextMetrics()` with `SetTextMetricsProvider` or `ctx.WithTextMetrics`. See [docs/text.md](docs/text.md).

## Fluent API

Every `*Node` has immutable, chainable methods:

```go
card := layout.VStack().
	WithWidth(300).
	WithPadding(16).
	AddChildren(
		layout.Text("Title"),
		layout.Text("Body"),
	)

wide := card.FindAll(func(n *layout.Node) bool { return n.Style.Width.Value > 200 })
count := layout.FoldNodes(card, 0, func(acc int, _ *layout.Node) int { return acc + 1 })
```

`WithWidth`/`WithHeight` take pixels; `WithWidthLength`/`WithHeightLength` take any `Length` (and `Length{}` resets to auto). See the [Fluent API guide](docs/fluent-api.md).

## Documentation

| Document | Contents |
|----------|----------|
| [Getting started](docs/getting-started.md) | Nodes, styles, lengths, constraints, `LayoutSimple` vs `Layout` |
| [Layout systems](docs/layout-systems.md) | Flexbox, Grid, Block, Text, positioned layout, container queries, intrinsic sizing |
| [Text](docs/text.md) | Text properties, line boxes, writing modes, text metrics |
| [API reference](docs/api-reference.md) | Curated map of the exported API, with a link to pkg.go.dev |
| [Fluent API](docs/fluent-api.md) | Immutable tree building, querying, and transformation |
| [Usage patterns](docs/usage-patterns.md) | Embedding `Node`, builders, functional options |
| [SVG rendering](docs/svg-rendering.md) | Turning a laid-out tree into SVG |
| [Gotchas](docs/gotchas.md) | Behaviors that surprise people, including the v1.4.0 semantics |
| [Spec compliance](docs/spec-compliance.md) | Per-module matrix of what is implemented |
| [Limitations](docs/limitations.md) | The single list of known gaps |
| [WPT testing](docs/wpt-testing.md) | The `wpt/` nested module and CEL assertions |
| [serialize/README.md](serialize/README.md) | JSON/YAML format and limits |
| [CHANGELOG.md](CHANGELOG.md) | Release notes, including behavior changes |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Tests, linters, CI, commit conventions |

Runnable programs live in [`examples/`](examples/) (`go run ./examples/grid`, `go run ./examples/fluent/dashboard`, ...).

## Learning resources

- [MDN: Flexbox](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_flexible_box_layout), [Grid](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_grid_layout), [Box alignment](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_alignment), [Box model](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_box_model), [Positioned layout](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_positioned_layout), [Writing modes](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_writing_modes)

## License

MIT
