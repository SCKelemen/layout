# API Reference

A map of the exported API grouped by topic. Signatures and field-level documentation live in the source and on [pkg.go.dev/github.com/SCKelemen/layout](https://pkg.go.dev/github.com/SCKelemen/layout) (`go doc github.com/SCKelemen/layout` locally); this page tells you which identifiers exist and when to reach for them.

## Core types

| Type | Role |
|------|------|
| `Node` | Tree node: `Style`, `Children`, `Text`, `Baseline`; `Rect` and `TextLayout` are outputs of layout |
| `Style` | Every layout property: display, flex, grid, sizing, spacing, positioning, transform, writing mode, container queries, `TextStyle` |
| `TextStyle` | Text properties (font, alignment, spacing, wrapping, decoration metadata, direction) |
| `Spacing` | `Top`/`Right`/`Bottom`/`Left` lengths for padding, margin, border; built with `Uniform`, `Horizontal`, `Vertical` |
| `Constraints` | Available space; built with `Tight`, `Loose`, `Unconstrained`; `Constrain(Size)` clamps |
| `Size`, `Point`, `Rect` | Plain geometry; `Unbounded` (`math.MaxFloat64`) marks an indefinite axis |
| `GridTrack`, `RepeatTrack`, `GridArea`, `GridTemplateAreas` | Grid track definitions and named areas |
| `Transform` | 2D affine matrix; the zero value is the identity |
| `TextLayout`, `TextLine`, `InlineBox` | Line boxes produced by `LayoutText` (see [Text](text.md)) |

## Lengths

`Length` and `LengthUnit` are aliases of `units.Length` / `units.LengthUnit` from `github.com/SCKelemen/units`, so they carry that package's methods (`Add`, `Sub`, `Mul`, `Div`, `IsAbsolute`, `IsFontRelative`, `IsViewportRelative`, `IsContainerRelative`, `String`, ...).

| Constructors | Units |
|--------------|-------|
| `Px`, `Pt`, `Pc`, `In`, `Cm`, `Mm`, `Q` | absolute |
| `Em`, `Rem`, `Ch` | font-relative |
| `Vw`, `Vh`, `Vmin`, `Vmax` | viewport |
| `Cqw`, `Cqh`, `Cqi`, `Cqb`, `Cqmin`, `Cqmax` | container query (resolve with `ResolveLengthInContext`) |
| `UnboundedLength()`, `PxUnbounded` | layout-only "no limit" sentinel (`UnboundedUnit`) |

Semantics: the zero value (`Unit == ""`) is unset and means `auto` (or "fall back to the shorthand" for gap longhands); `Px(0)` is an explicit zero. Negative values are legal for margins and offsets.

The remaining CSS Values Level 4 units (`lh`, `rlh`, `ex`, `cap`, `ic`, `vi`, `vb`, `sv*`/`lv*`/`dv*`, `rch`, `rex`, ...) have no `layout.` constructor but are accepted everywhere a `Length` is: use the `units` constructors (`units.Lh(2)`, `units.Ic(1)`) or a literal `layout.Length{Value: 2, Unit: "lh"}`. In this engine `ex`, `cap`, and `lh` resolve to the current font size, `rlh` to the root font size, and `ic` to twice the `ch` advance. The `LayoutContext` carries only `ViewportWidth`/`ViewportHeight`, so `vi`, `vb`, and the `sv*`/`lv*`/`dv*` variants resolve to 0; use `vw`/`vh`/`vmin`/`vmax`. Unit constants are re-exported as `Pixels`, `EmUnit`, `RemUnit`, `ChUnit`, `VwUnit`, `VhUnit`, `VminUnit`, `VmaxUnit`, `PtUnit`, `PcUnit`, `InUnit`, `CmUnit`, `MmUnit`, `QUnit`.

Resolution:

- `ResolveLength(l, ctx, fontSize) float64`: pixels for any unit except `cq*`, which yield 0 here. Viewport units yield 0 when the context has no viewport size.
- `ResolveLengthInContext(l, ctx, fontSize, nctx) float64`: as above, plus container-query units resolved against the nearest qualifying `Style.ContainerType` ancestor of `nctx` (a `NodeContext`), falling back to the viewport.

## LayoutContext

`NewLayoutContext(viewportWidth, viewportHeight, rootFontSize)` returns a context with the package-level text metrics and `'0'` as the `ch` reference glyph. `WithTextMetrics(provider)` and `WithChReferenceChar(r)` return modified copies. Fields: `ViewportWidth`, `ViewportHeight`, `RootFontSize` (px), `TextMetrics`, `ChReferenceChar`. A nil context is tolerated by the layout functions (root font 16px, no viewport).

## Layout functions

| Function | Use |
|----------|-----|
| `LayoutSimple(root, constraints) Size` | Normal flow with a context derived from the constraints (viewport = max sizes, 16px root font; an unbounded axis gives a 0 viewport dimension) |
| `Layout(root, constraints, ctx) Size` | Normal flow with your context; dispatches on `root.Style.Display`; nil root yields a zero `Size` |
| `LayoutWithPositioning(root, constraints, viewportRect, ctx) Size` | Normal flow followed by the positioning pass for relative/absolute/fixed/sticky boxes |
| `LayoutBlock`, `LayoutFlexbox`, `LayoutGrid`, `LayoutText` | The individual algorithms; `Layout` calls these |
| `LayoutPositioned(node, containingBlock, viewportRect, ctx)` | Position a single node; `LayoutWithPositioning` drives this recursively |
| `CalculateIntrinsicWidth`, `CalculateIntrinsicHeight` | min-/max-/fit-content measurements (`IntrinsicSize` argument) |

## Building trees

Constructors and mutating helpers (they return the node they were given):

- Containers: `HStack`, `VStack`, `ZStack`, `Grid(rows, cols, rowSize, colSize)`, `GridAuto`, `GridFractional`.
- Leaves: `Fixed(width, height)`, `Spacer()`, `Text(text, style...)`.
- Sizing: `Frame(node, w, h)` (skips values `<= 0`), `FrameLength(node, w, h Length)` (exact; `Length{}` resets to auto), `MinWidth`, `MinHeight`, `AspectRatio`, `MinContentWidth`, `MaxContentWidth`, `FitContentWidth(node, max)`, `MinContentHeight`, `MaxContentHeight`, `FitContentHeight(node, max)`.
- Spacing: `Padding`, `PaddingCustom(node, top, right, bottom, left)`, `Margin`.
- Grid: `FixedTrack`, `FractionTrack`, `MinMaxTrack`, `AutoTrack`, `MinContentTrack`, `MaxContentTrack`, `FitContentTrack`, `RepeatTracks`, `AutoFillTracks`, `AutoFitTracks`, `NewGridTemplateAreas` + `DefineArea`, `PlaceInArea`.
- Deprecated: `Background` (no-op), `GetSVGTransform` (use `Transform.ToSVGString`), `CollectNodesForSVG` (use `Node.DescendantsAndSelf`).

## Fluent API

Methods on `*Node`, all nil-safe and copy-on-write (see the [Fluent API guide](fluent-api.md)):

- Navigation: `Descendants`, `DescendantsAndSelf`, `FirstChild`, `LastChild`, `ChildAt`, `ChildCount`.
- Queries: `Find`, `FindAll`, `Where`, `Any`, `All`, `OfDisplayType`.
- Copies: `Clone` (shallow), `CloneDeep`.
- Style: `WithStyle`, `WithPadding`, `WithPaddingCustom`, `WithMargin`, `WithMarginCustom`, `WithWidth`, `WithHeight` (pixels), `WithWidthLength`, `WithHeightLength` (any `Length`), `WithText`, `WithDisplay`, `WithFlexGrow`, `WithFlexShrink`.
- Children: `WithChildren`, `AddChild`, `AddChildren`, `RemoveChildAt`, `ReplaceChildAt`, `InsertChildAt`.
- Transformations: `Transform`, `Map`, `Filter`, `FilterDeep`, `Fold`, `FoldWithContext`, and the generic `FoldNodes[T](n, init, fn)`.
- Parent navigation: `NewContext(root) *NodeContext` with `Parent`, `Ancestors`, `AncestorsAndSelf`, `Root`, `Siblings`, `Children`, `ChildAt`, `Depth`, `IsRoot`, `HasParent`, `HasChildren`, `FindUp`, `FindDown`, `FindDownAll`, `Unwrap`.

## Post-layout helpers

`AlignNodes(nodes, AlignEdge)` (`AlignLeft`, `AlignRight`, `AlignTop`, `AlignBottom`, `AlignCenterX`, `AlignCenterY`), `DistributeNodes(nodes, DistributeHorizontal|DistributeVertical)`, `SnapNodes(nodes, snapSize)`, `SnapToGrid(nodes, snapSize, originX, originY)`. They edit `Rect` after layout; intended for block and absolutely positioned boxes.

## Transforms

`IdentityTransform`, `Translate`, `Scale`, `Rotate` (radians), `RotateDegrees`, `SkewX`, `SkewY`, `Matrix`; methods `Multiply` (right operand applies first), `Apply`, `ApplyToRect`, `IsIdentity`, `ToSVGString` (empty string for the identity). `GetFinalRect(node)` returns the transformed bounding box.

## Text metrics

`TextMetricsProvider` (interface: `Measure(text, style) (advance, ascent, descent)`), `SetTextMetricsProvider` (package-level, safe for concurrent use), `LayoutContext.WithTextMetrics` (per-layout, honored by all text measurement), `TextMetricsAdapter` with `NewTerminalTextMetrics()` and `NewTextMetricsAdapter(text.Config)` wrapping `github.com/SCKelemen/text`.

## Enums

Every CSS-like enum has `String()` returning the CSS keyword and a `Parse<Enum>(string)` function returning an error that lists the valid keywords: `Display`, `FlexDirection`, `FlexWrap`, `JustifyContent`, `AlignItems`, `JustifyItems`, `AlignContent`, `GridAutoFlow`, `BoxSizing`, `Position`, `TextAlign`, `TextAlignLast`, `TextJustify`, `WhiteSpace`, `TextOverflow`, `OverflowWrap`, `WordBreak`, `TextTransform`, `Hyphens`, `HangingPunctuation`, `Direction`, `WritingMode`, `FontStyle`, `TextDecorationStyle`, `VerticalAlign`, `IntrinsicSize`, `InlineBoxKind`, plus `ContainerType`. Matching is exact and lowercase; the css-align-3 aliases `start`/`end` parse for the alignment enums and `JustifyContentStart`/`AlignItemsEnd`-style constants format as `flex-start`/`flex-end`. Out-of-range values format as `Display(7)`.

```go
fmt.Println(layout.JustifyContentSpaceBetween.String()) // space-between
v, err := layout.ParseAlignItems("end")                  // AlignItemsFlexEnd, nil
_, err = layout.ParseDisplay("table")                    // layout: invalid display "table" (valid: block, flex, grid, inline-text, none)
_ = v
_ = err
```

`WritingMode` also has `IsVertical`, `IsHorizontal`, `IsSideways`, `IsRightToLeft`; `TextDecoration` is a bitmask with `Has`. `ContainerName` has `Has(name)`; `ParseContainerName` and `ParseContainer` parse the `container-name` and `container` shorthand.

## UAX #14

`BreakClass` constants (`ClassAL`, `ClassID`, `ClassSP`, ...) and `BreakAction` are exported from the line-breaking implementation used by text layout.

## serialize package

`serialize.ToJSON`, `FromJSON`, `ToYAML`, `FromYAML` (the YAML pair is excluded with `-tags no_yaml`); limits `MaxTreeDepth`, `MaxChildren`, `MaxNumericValue`, `MaxRepeatCount`; errors `ErrLimitExceeded`, `ErrNullInput`. Lengths are written as unit strings (`"10px"`, `"2em"`, `"unbounded"`); bare numbers load as pixels. See [serialize/README.md](../serialize/README.md).
