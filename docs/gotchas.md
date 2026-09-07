# Gotchas

Behaviors that follow the CSS specifications (or this library's Go representation of them) but regularly surprise people. Items marked **v1.4.0+** changed in that release; see [CHANGELOG.md](../CHANGELOG.md) for the full list of behavior changes.

## Lengths

### Unset means auto, `Px(0)` means zero (v1.4.0+)

A `Length` you never assign has `Unit == ""` and is read as the CSS initial value `auto`. `Px(0)` is a real zero.

```go
autoWidth := &layout.Node{Children: []*layout.Node{{Style: layout.Style{Height: layout.Px(10)}}}}
zeroWidth := &layout.Node{Style: layout.Style{Width: layout.Px(0)}, Children: []*layout.Node{{Style: layout.Style{Height: layout.Px(10)}}}}
layout.LayoutSimple(autoWidth, layout.Loose(300, 300))
layout.LayoutSimple(zeroWidth, layout.Loose(300, 300))
fmt.Println(autoWidth.Rect.Width, zeroWidth.Rect.Width) // 300 0
```

This applies to `Width`, `Height`, `FlexBasis`, `Top`/`Right`/`Bottom`/`Left`, and the gap longhands. Code that used to test `Style.Width.Value == 0` for "not set" must test `Style.Width.Unit == ""` instead; `-1` is not an auto sentinel anywhere except grid line indices.

### `FlexBasis: Px(0)` is a real zero basis (v1.4.0+)

Two items with `FlexGrow: 1, FlexBasis: Px(0)` in a 400px row are 200/200 regardless of their content. Leave `FlexBasis` unset to get `auto` (the item's main size, or its content size).

### `layout.Text()` no longer seeds `Px(0)` (v1.4.0+)

Text nodes leave `Width`/`Height` unset, so they size to their content as block children, flex items, and grid items. Set `Width` when you want a specific breaking width.

### Container-query units resolve to 0 in the layout pass

`Layout`, `LayoutSimple`, and `ResolveLength` have no ancestor information, so `Cqw(50)` is 0 there. Resolve `cq*` values yourself with `ResolveLengthInContext(l, ctx, fontSize, layout.NewContext(root)...)` and store the result as `Px` before layout.

### Viewport units need a viewport

`LayoutSimple` takes the viewport from the constraints. With `Unconstrained()` (or any unbounded axis) that dimension is 0 and `Vw(50)` resolves to 0. Use `Layout` with `NewLayoutContext(w, h, fontSize)` when you mix unbounded constraints with viewport units. `vi`, `vb`, and the `sv*`/`lv*`/`dv*` variants always resolve to 0.

### Other CSS Values Level 4 units

`lh`, `rlh`, `ex`, `cap`, `ic`, and friends have no `layout.` constructor; use `units.Lh(2)` (from `github.com/SCKelemen/units`) or `layout.Length{Value: 2, Unit: "lh"}`. They are approximated from the font size.

## Gaps

### Flex row/column gaps follow the flex direction (v1.4.0+)

`FlexRowGap` is the gap between rows and `FlexColumnGap` the gap between columns, as in CSS. In a `column` container the gap between items is therefore `FlexRowGap`, and `FlexColumnGap` separates wrapped lines.

### A `Px(0)` longhand overrides the shorthand (v1.4.0+)

`FlexColumnGap: Px(0)` cancels `FlexGap: Px(10)` between items in a row; `GridRowGap: Px(0)` cancels `GridGap` between rows. Only an unset longhand falls back to the shorthand.

```go
row := layout.HStack(layout.Fixed(50, 50), layout.Fixed(50, 50))
row.Style.FlexGap = layout.Px(10)
row.Style.FlexColumnGap = layout.Px(0)
layout.LayoutSimple(row, layout.Loose(800, 600))
fmt.Println(row.Children[1].Rect.X) // 50, not 60
```

## Block layout

### Margins collapse, including through parents (v1.4.0+)

Vertical margins between siblings collapse to the larger one (or largest positive plus most negative). A block's top margin collapses with its first in-flow child's top margin when nothing separates them, and the bottom margins collapse when the block's height is auto, so the child ends up at the parent's origin and the margin moves outside the parent:

```go
root := &layout.Node{Children: []*layout.Node{{Children: []*layout.Node{
	{Style: layout.Style{Height: layout.Px(50), Margin: layout.Spacing{Top: layout.Px(20), Bottom: layout.Px(30)}}},
}}}}
layout.LayoutSimple(root, layout.Loose(300, layout.Unbounded))
parent := root.Children[0]
fmt.Println(parent.Rect.Y, parent.Rect.Height, parent.Children[0].Rect.Y, root.Rect.Height) // 20 50 0 100
```

Add padding or a border to the parent to keep the margin inside it. Flex and grid containers, text boxes, and absolutely positioned boxes never collapse through.

### Margins do not collapse in flex or grid

Two adjacent flex items with `Margin: Uniform(Px(10))` are 20px apart.

## Grid

### Empty items in auto rows are 0px tall

An `AutoTrack()` row sizes to its content. An item with no children, no text, and no `Height`/`MinHeight` contributes 0, so the row collapses. Give such items a `MinHeight` (`layout.MinHeight(item, 50)`) or use fixed tracks.

### Explicit placement uses 0-based, end-exclusive lines

`GridRowStart: 1, GridRowEnd: 2` is the second row. Setting only `GridRowStart: 0` (with `GridRowEnd` left at 0) is indistinguishable from an unset item and is auto-placed; set `GridRowEnd: 1` for an explicit first row. `-1` is auto.

### Stretch respects explicit sizes

With the default `stretch` alignment, an item with an explicit `Width`/`Height` (or `WidthSizing`/`HeightSizing`) keeps that size and sits at the start of its area; only auto items fill the cell.

### `repeat(auto-fill)` lives in a different field

`RepeatTracks(layout.RepeatCountAutoFill, ...)` returns an empty slice. Put `AutoFillTracks(...)`/`AutoFitTracks(...)` in `Style.GridTemplateColumnsRepeat` or `GridTemplateRowsRepeat`; the tracks are appended after `GridTemplateColumns`/`Rows` during layout.

## Positioning

### The containing block is the nearest positioned ancestor (v1.4.0+)

An absolute box is placed relative to the padding box of the nearest ancestor whose `Position` is not static, not its direct parent. Static ancestors are skipped; without any positioned ancestor the root is used. `Rect` stays parent-relative, so the coordinates of an absolute child of a static parent can be negative.

### Positioned boxes need `LayoutWithPositioning`

`Layout` leaves relative/absolute/fixed/sticky boxes at their static position. Use `LayoutWithPositioning(root, constraints, viewportRect, ctx)`.

### Offsets: unset is auto, `Px(0)` is 0, negatives are real (v1.4.0+)

`Left: Px(-10)` moves a relative box left. Both offsets of an axis unset keeps the static position. `ZStack` sets `Left`/`Top` to `Px(0)` for you.

### Sticky is relative

There is no scroll offset, so `PositionSticky` behaves as `PositionRelative`.

## Flexbox

### `FlexShrink: 0` is not "don't shrink"

The zero value is read as the default 1. Give an item a `MinWidth`/`MinHeight` to stop it from shrinking.

### `AlignSelf: AlignItemsStretch` does not override the parent

`AlignSelf`'s zero value means "use the parent's `AlignItems`". Only non-stretch overrides are expressible. `JustifySelf` behaves the same way in grid.

### A definite main size is kept under loose constraints (v1.4.0+)

`{Display: DisplayFlex, Width: Px(200)}` is 200px wide under `Loose(...)` or `Unconstrained()`, even when its items overflow; only a container without a set main size is content-sized.

## Text

### `WithText` does not create a text node

`Fixed(100, 50).WithText("x")` sets `Node.Text` but leaves `Display` as block, so no line breaking happens and `TextLayout` stays nil. Use `layout.Text("x", style)`.

### `line-height` below 10 is a multiplier

`LineHeight: 1.5` is 1.5 x font size; `LineHeight: 12` is 12px even with a 24px font.

### `-1` means normal for `LetterSpacing`/`WordSpacing`

The zero value adds 0px, which happens to be the same as normal, but code that inspects the fields (or serialized trees) sees `-1` for "normal".

### `writing-mode` is not inherited

Set `Style.WritingMode` on every node that should lay out vertically. `InlineBox.Orientations` is populated only when `TextStyle.WritingMode` is set as well.

### Text measurement defaults to an approximation

The built-in provider counts every rune as 0.6 x font size. Install `NewTerminalTextMetrics()` (globally with `SetTextMetricsProvider`, or per context with `ctx.WithTextMetrics`) for Unicode-accurate widths. Since round 1 of the correctness sweep, `ctx.TextMetrics` is honored by all text measurement, not only by `ch` units.

### Tabs advance to tab stops from the line start

In `pre`/`pre-wrap`, a tab moves to the next multiple of `TabSize` spaces measured from the start of the line box (not the block), so `"a\tb"` with 6px glyphs and the default tab size is 54px wide, not 18px.

## Transforms and serialization

### `Transform{}` is the identity (v1.4.0+)

Nodes that never set a transform no longer produce `matrix(0,0,0,0,0,0)`; `ToSVGString()` returns `""` for them. As a consequence `Scale(0, 0)` is also treated as the identity.

### Lengths serialize as strings (v1.4.0+)

`serialize` writes `"10px"`, `"2em"`, `"unbounded"`; unset lengths are omitted and `Px(0)` is `"0px"`. Bare numbers are still accepted on input as pixels. `-1px` is a negative length, not auto.

## Fluent API

### `Clone` shares slices and pointers

`Clone`/`WithStyle` copy the `Node` struct but share `Children`, `GridTemplateRows/Columns`, `GridTemplateAreas`, `ContainerName`, `TextStyle`, and `TextLayout`. Use `CloneDeep` for an independent copy.

### `WithWidth` takes pixels

`WithWidth(0)` sets `Px(0)`, a real zero. Use `WithWidthLength(layout.Length{})` to reset to auto and `WithWidthLength(layout.Vw(50))` for other units.

## Debugging checklist

1. Did you call `LayoutSimple`/`Layout` (or `LayoutWithPositioning` for positioned boxes)? `Rect` is zero until you do.
2. Is a dimension unset when you meant `Px(0)`, or `Px(0)` when you meant auto?
3. Do items in auto grid rows have content or a `MinHeight`?
4. Are gap longhands set to `Px(0)` while you rely on the shorthand?
5. For vertical text, is `Style.WritingMode` set on the node itself?
6. For viewport or container units, does the `LayoutContext` carry a viewport, and are `cq*` values pre-resolved?
7. Serialize the tree with `serialize.ToJSON(root)` to see exactly which properties are set.
