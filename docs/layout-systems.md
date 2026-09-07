# Layout Systems

`Style.Display` selects the algorithm for a node's children: `DisplayBlock` (default), `DisplayFlex`, `DisplayGrid`, or `DisplayInlineText` for a text leaf. Positioning is a second pass over any tree. All snippets assume `import "github.com/SCKelemen/layout"`.

## Flexbox

**Specification**: [CSS Flexible Box Layout Module Level 1](https://www.w3.org/TR/css-flexbox-1/)

Supported: `FlexDirection` (row, column, and the reverse variants), `FlexWrap` (nowrap, wrap, wrap-reverse), `JustifyContent` (flex-start, flex-end, center, space-between, space-around, space-evenly; `stretch` behaves as flex-start), `AlignItems`/`AlignSelf` (stretch, flex-start, flex-end, center, baseline), `AlignContent` (stretch, flex-start, flex-end, center, space-between, space-around, space-evenly), `FlexGrow`, `FlexShrink`, `FlexBasis`, `FlexGap`/`FlexRowGap`/`FlexColumnGap`, `Order`, margins, and writing modes.

```go
root := &layout.Node{
	Style: layout.Style{
		Display:       layout.DisplayFlex,
		FlexDirection: layout.FlexDirectionColumn,
		AlignItems:    layout.AlignItemsStretch,
		FlexGap:       layout.Px(8),
		Padding:       layout.Uniform(layout.Px(20)),
	},
	Children: []*layout.Node{
		{Style: layout.Style{FlexGrow: 1, FlexBasis: layout.Px(0)}},
		{Style: layout.Style{FlexGrow: 2, FlexBasis: layout.Px(0)}},
	},
}
layout.LayoutSimple(root, layout.Tight(400, 600))
```

Flex properties:

- `FlexGrow` (default 0) and `FlexShrink` (default 1). `FlexShrink: 0` is read as the default 1; there is no way to express `flex-shrink: 0` (see [Limitations](limitations.md)).
- `FlexBasis`: unset means `auto` (use the main size property, else the content size). `Px(0)` is a zero basis, so two items with `FlexGrow: 1, FlexBasis: Px(0)` split a 400px container 200/200 regardless of content.
- `Width`/`Height`: unset is `auto`; `Px(0)` is a definite zero size. `align-items: stretch` only stretches items whose cross size is auto.
- Gaps follow CSS Box Alignment: `FlexRowGap` is the gap between rows and `FlexColumnGap` between columns, so in a `column` container the gap between items is the row gap. A longhand that is set (even `Px(0)`) overrides `FlexGap`; an unset longhand falls back to it.
- Absolutely or fixed positioned children take no flex slot.

## Grid

**Specification**: [CSS Grid Layout Module Level 1](https://www.w3.org/TR/css-grid-1/)

Supported: explicit tracks (`GridTemplateRows`/`GridTemplateColumns`), implicit tracks (`GridAutoRows`/`GridAutoColumns`), `repeat()` via `GridTemplateRowsRepeat`/`GridTemplateColumnsRepeat`, named areas, line-based placement and spanning, auto-placement (`GridAutoFlow` row/column, sparse or dense), gaps, `JustifyItems`/`JustifySelf`, `AlignItems`/`AlignSelf`, `JustifyContent`, `AlignContent`, margins, intrinsic track sizes, and writing modes.

```go
root := &layout.Node{
	Style: layout.Style{
		Display: layout.DisplayGrid,
		GridTemplateRows: []layout.GridTrack{
			layout.FixedTrack(layout.Px(100)),
			layout.FractionTrack(1),
		},
		GridTemplateColumns: []layout.GridTrack{
			layout.FixedTrack(layout.Px(200)), // sidebar
			layout.FractionTrack(1),           // main content
		},
		GridGap: layout.Px(10),
	},
	Children: []*layout.Node{
		{Style: layout.Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 0, GridColumnEnd: 2}}, // spans both columns
		{Style: layout.Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 0, GridColumnEnd: 1}},
		{Style: layout.Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 1, GridColumnEnd: 2}},
	},
}
layout.LayoutSimple(root, layout.Tight(600, 400))
```

Line indices are 0-based; `GridRowEnd`/`GridColumnEnd` are exclusive. An item whose `Start` and `End` are both 0 (the zero value) is auto-placed; `-1` also means auto.

### Tracks

| Helper | CSS | Notes |
|--------|-----|-------|
| `FixedTrack(Px(100))` | `100px` | any `Length` |
| `FractionTrack(1)` | `1fr` | behaves as `minmax(auto, 1fr)` |
| `MinMaxTrack(min, max)` | `minmax(min, max)` | use `PxUnbounded` for an unbounded max |
| `AutoTrack()` | `auto` | sized to the items' contributions |
| `MinContentTrack()`, `MaxContentTrack()`, `FitContentTrack(300)` | `min-content`, `max-content`, `fit-content(300px)` | |
| `RepeatTracks(3, tracks...)` | `repeat(3, ...)` | expands to a `[]GridTrack` at construction time |

A zero-value `GridTrack` (or `GridAutoRows`/`GridAutoColumns`) means `auto`.

### repeat(auto-fill) and repeat(auto-fit)

Auto-repeat depends on the container size, so it is expressed as a `RepeatTrack` in `Style.GridTemplateColumnsRepeat` / `GridTemplateRowsRepeat` and expanded during layout, after the explicit template tracks:

```go
grid := &layout.Node{
	Style: layout.Style{
		Display: layout.DisplayGrid,
		Width:   layout.Px(450),
		GridGap: layout.Px(10),
		GridTemplateColumnsRepeat: []layout.RepeatTrack{
			layout.AutoFillTracks(layout.FixedTrack(layout.Px(100))), // repeat(auto-fill, 100px)
		},
	},
}
for i := 0; i < 5; i++ {
	grid.Children = append(grid.Children, &layout.Node{Style: layout.Style{Height: layout.Px(20)}})
}
layout.LayoutSimple(grid, layout.Loose(800, 600))
// floor((450 + 10) / (100 + 10)) = 4 columns; the fifth item wraps to row 2
```

`AutoFitTracks` is the same but collapses repeated tracks that end up empty, together with their gutters. Patterns must use fixed sizes (no `fr`, `auto`, or intrinsic keywords); one auto-repeat per axis; a `RepeatTrack` with a positive `Count` repeats literally. Both are capped at 10,000 tracks. `RepeatTracks(RepeatCountAutoFill, ...)` returns an empty slice because a plain `[]GridTrack` cannot carry the intent.

### Named areas

```go
areas := layout.NewGridTemplateAreas(2, 3)
_ = areas.DefineArea("header", 0, 1, 0, 3)  // row 0, all columns
_ = areas.DefineArea("sidebar", 1, 2, 0, 1) // row 1, column 0
_ = areas.DefineArea("content", 1, 2, 1, 3) // row 1, columns 1-2

page := layout.GridFractional(2, 3)
page.Style.GridTemplateAreas = areas
page.Children = []*layout.Node{
	layout.PlaceInArea(&layout.Node{}, "header"),
	layout.PlaceInArea(&layout.Node{}, "sidebar"),
	layout.PlaceInArea(&layout.Node{}, "content"),
}
```

`DefineArea` rejects overlapping or out-of-range areas. Area names are resolved during layout without mutating the children, so a tree can be laid out again with a different template. An undefined name falls back to the child's line-based placement.

### Alignment

- `JustifyItems`/`JustifySelf` (inline axis) and `AlignItems`/`AlignSelf` (block axis) default to `stretch`. Stretch fills the area only when the item's size in that axis is auto; an explicit size or an intrinsic sizing mode is kept and the item is aligned at the start.
- `JustifyContent` aligns the columns as a group: start, end, center, space-between, space-around, space-evenly, and `JustifyContentStretch` (grows `auto`-max columns). `AlignContent` does the same for rows.
- Items with `AspectRatio` default to `start` alignment in the block axis.

### Helpers

```go
fixed := layout.Grid(4, 4, 150, 200)   // 4 rows x 4 columns, rows 150px, columns 200px
auto := layout.GridAuto(3, 4)          // auto-sized tracks
fractional := layout.GridFractional(2, 3) // equal fr tracks
fixed.Style.GridGap = layout.Px(10)
```

The `dashboard`, `columns`, `bento`, and `autofill` sections of `examples/grid/main.go` (`go run ./examples/grid bento`) show complete layouts, including a mosaic with items spanning several rows and columns.

## Block

**Specification**: [CSS 2.1 §9–10](https://www.w3.org/TR/CSS21/visuren.html), [CSS Box Sizing Level 3](https://www.w3.org/TR/css-sizing-3/)

Block is the default. Children stack in the block direction; an auto width fills the available width and an auto height wraps the content.

```go
root := &layout.Node{
	Style: layout.Style{Padding: layout.Uniform(layout.Px(10))},
	Children: []*layout.Node{
		{Style: layout.Style{Height: layout.Px(40)}},
		{Style: layout.Style{Height: layout.Px(40), Margin: layout.Spacing{Top: layout.Px(20)}}},
	},
}
layout.LayoutSimple(root, layout.Loose(300, layout.Unbounded))
// root is 300 x 120 (10 + 40 + 20 + 40 + 10)
```

Block behavior:

- Vertical margins collapse between siblings, between a block and its first/last in-flow child (when no padding or border separates them), and through empty blocks; the collapsed margin is the largest positive plus the most negative (CSS 2.1 §8.3.1). Flex and grid containers, text boxes, and absolutely positioned boxes do not collapse through.
- `MinWidth`/`MaxWidth`/`MinHeight`/`MaxHeight` apply max-first, so `min` wins when it exceeds `max`.
- `BoxSizing: BoxSizingBorderBox` makes `Width`/`Height` include padding and border; the content box is floored at 0.
- `AspectRatio` sizes the auto axis from the definite one; with both axes auto the available width is used.
- Absolutely positioned children take no space.

## Text

**Specification**: [CSS Text Module Level 3](https://www.w3.org/TR/css-text-3/), [CSS Writing Modes Level 3](https://www.w3.org/TR/css-writing-modes-3/)

A text node is a leaf with `Display: DisplayInlineText`, `Text` set, and an optional `TextStyle`. `layout.Text(s, style...)` builds one:

```go
para := layout.Text("The quick brown fox jumps over the lazy dog", layout.Style{
	Width: layout.Px(120),
	TextStyle: &layout.TextStyle{
		FontSize:   16,
		LineHeight: 1.5,
		TextAlign:  layout.TextAlignJustify,
		WhiteSpace: layout.WhiteSpaceNormal,
	},
})
size := layout.LayoutSimple(para, layout.Loose(400, 400))
for _, line := range para.TextLayout.Lines {
	_ = line.OffsetY             // block position of the line
	_ = line.Boxes               // words; Boxes[i].SpaceAfter marks an inter-word space
	_ = line.EndsWithForcedBreak // true for the last line and after a preserved newline
}
_ = size
```

Text nodes participate in every container: block stacks them, flex measures them with `LayoutText` and re-wraps them at their flexed size, grid uses their min-content contribution as the track base size. An explicit `Width` is the breaking width; otherwise the available inline size is.

### Writing modes and direction

`Style.WritingMode` selects `horizontal-tb` (default), `vertical-rl`, `vertical-lr`, `sideways-rl`, or `sideways-lr`. In vertical modes the inline axis is the height and lines stack horizontally; block, flex, and grid containers map their logical axes accordingly. `InlineBox.Orientations` reports per rune whether a glyph is upright (CJK) or rotated (Latin) per UAX #50, with sideways modes rotating everything; it is computed from `TextStyle.WritingMode`, so set that field too (in addition to `Style.WritingMode`) when your renderer needs orientations. The engine does not propagate `WritingMode` to children, so set it on each node that needs it.

`Style.Direction` (falling back to `TextStyle.Direction`) selects `ltr` or `rtl`: `TextAlignDefault` resolves to the end edge in RTL, `text-indent` moves to the right edge, and over-constrained absolute boxes drop `Left` instead of `Right`. There is no bidi reordering (UAX #9); runs are laid out in logical order.

See [Text](text.md) for every property and the line-box model.

## Positioned layout

**Specification**: [CSS Positioned Layout Module Level 3](https://www.w3.org/TR/css-position-3/), [CSS 2.1 §10](https://www.w3.org/TR/CSS21/visudet.html)

Positioning is a second pass: `LayoutWithPositioning(root, constraints, viewportRect, ctx)` runs normal flow layout and then places `PositionRelative`, `PositionAbsolute`, `PositionFixed`, and `PositionSticky` boxes. `Layout` alone leaves them at their static position.

```go
root := &layout.Node{
	Style: layout.Style{Position: layout.PositionRelative, Width: layout.Px(400), Height: layout.Px(300)},
	Children: []*layout.Node{
		{
			Style: layout.Style{Margin: layout.Uniform(layout.Px(20))}, // static: not a containing block
			Children: []*layout.Node{
				{Style: layout.Style{
					Position: layout.PositionAbsolute,
					Left:     layout.Px(10),
					Top:      layout.Px(10),
					Width:    layout.Px(50),
					Height:   layout.Px(50),
				}},
			},
		},
	},
}
ctx := layout.NewLayoutContext(800, 600, 16)
layout.LayoutWithPositioning(root, layout.Loose(800, 600), layout.Rect{Width: 800, Height: 600}, ctx)
abs := root.Children[0].Children[0]
// abs.Rect is (-10, -10) in its parent's space: 10px from root's padding box, which starts 20px up and left
```

Rules:

- Offsets (`Top`, `Right`, `Bottom`, `Left`) are `auto` when unset. Any set value, including `Px(0)` and negative lengths, is a real offset. Both offsets of an axis auto keeps the static position.
- The containing block of an absolute box is the padding box of the nearest positioned ancestor (any `Position` other than static); static ancestors are skipped. With no positioned ancestor, the root is the containing block. Fixed boxes use `viewportRect`.
- With `Left`, `Width`, and `Right` all set the box is over-constrained and `Right` is ignored (`Left` in RTL); with `Top`, `Height`, and `Bottom` all set, `Bottom` is ignored. An explicit size on an absolute box is always its used size.
- Relative boxes shift by their offsets and keep their flow space. Sticky boxes behave as relative because there is no scroll offset to stick to.
- `ZIndex` is stored for renderers; layout does not reorder by it.

`ZStack(children...)` marks every child absolute, defaults unset `Left`/`Top` to `Px(0)`, and gives the container the union of its children's pixel extents so it is not 0 x 0.

## Container queries

**Specification**: [CSS Containment Module Level 3](https://www.w3.org/TR/css-contain-3/#container-queries)

`Style.ContainerType` (`ContainerTypeSize` or `ContainerTypeInlineSize`) and `Style.ContainerName` mark query containers; `ParseContainerType`, `ParseContainerName`, and `ParseContainer` parse the CSS keywords. Container-relative lengths (`Cqw`, `Cqh`, `Cqi`, `Cqb`, `Cqmin`, `Cqmax`) need ancestor information, so they are resolved with `ResolveLengthInContext`, not by the layout pass:

```go
card := &layout.Node{Style: layout.Style{ContainerType: layout.ContainerTypeInlineSize, Width: layout.Px(400)}}
child := &layout.Node{Style: layout.Style{Padding: layout.Uniform(layout.Cqw(5))}}
card.Children = []*layout.Node{child}
ctx := layout.NewLayoutContext(800, 600, 16)
layout.Layout(card, layout.Loose(800, 600), ctx) // card.Rect is now 400 wide

nctx := layout.NewContext(card).ChildAt(0)
padding := layout.ResolveLengthInContext(child.Style.Padding.Left, ctx, 16, nctx) // 20 (5% of 400)
_ = padding
```

The resolver walks up from `nctx`: `cqw`/`cqi` accept `size` or `inline-size` containers, `cqh`/`cqb` only `size`, `cqi`/`cqb` follow the container's `WritingMode`, and an axis without a container falls back to the viewport in `ctx`. Containers measure their content box. In the normal layout pass (`Layout`, `LayoutSimple`, `ResolveLength`) a `cq*` length resolves to 0. `@container` rules are out of scope.

## Intrinsic sizing

**Specification**: [CSS Sizing Module Level 3](https://www.w3.org/TR/css-sizing-3/)

`Style.WidthSizing` / `HeightSizing` (`IntrinsicSizeMinContent`, `IntrinsicSizeMaxContent`, `IntrinsicSizeFitContent` with `FitContentWidth`/`FitContentHeight`) size a box from its content instead of from `Width`/`Height`. The helpers `MinContentWidth(node)`, `MaxContentWidth(node)`, `FitContentWidth(node, 500)`, and the height variants set these fields:

```go
label := layout.MaxContentWidth(layout.Text("Do not wrap me"))
```

Block, flex, and grid honor the modes; grid items with a sizing mode keep their measured size instead of stretching. `CalculateIntrinsicWidth(node, constraints, IntrinsicSizeMinContent, ctx)` and `CalculateIntrinsicHeight` expose the measurement directly. Contributions include padding, border, and margins; text children contribute their longest word (min-content) or longest unwrapped line (max-content).

The float sentinels `SizeMinContent`/`SizeMaxContent`/`SizeFitContent` stored in `Width`/`Height` are deprecated but still honored.

## Choosing a layout system

- **Flexbox**: one-dimensional rows or columns with flexible sizing.
- **Grid**: two-dimensional placement, dashboards, card mosaics.
- **Block**: simple vertical stacking and document-like flow with margin collapsing.
- **Text**: any leaf that contains words.
- **Positioned**: overlays and anchored elements on top of any of the above.
