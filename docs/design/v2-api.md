# RFC: layout v2 API

| | |
|---|---|
| **Status** | Draft |
| **Date** | 2026-09-07 |
| **Supersedes** | none |
| **Module path** | `github.com/SCKelemen/layout/v2` |
| **Baseline** | v1.4.0 plus the compatible "Round 1" changes in progress |

This document proposes the breaking changes for the next major version of the
layout engine. It follows the format of `docs/fluent-api-design-decisions.md`:
for each area, the problem with a concrete v1 reproduction, the options that
actually competed, a recommendation, and the migration a v1 user performs.

## Table of Contents

- [Motivation and Scope](#motivation-and-scope)
- [Guiding Rule: Zero Value Is the CSS Initial Value](#guiding-rule-zero-value-is-the-css-initial-value)
- [Area 1: Zero-Value Violations](#area-1-zero-value-violations)
- [Area 2: One Layout Entry Point](#area-2-one-layout-entry-point)
- [Area 3: TextMetricsProvider](#area-3-textmetricsprovider)
- [Area 4: API Hygiene](#area-4-api-hygiene)
- [Area 5: Display and Sticky Additions](#area-5-display-and-sticky-additions)
- [Area 6: Serialization](#area-6-serialization)
- [Proposed v2 Style](#proposed-v2-style)
- [Migration Table](#migration-table)
- [Migration Tooling and v1 Support](#migration-tooling-and-v1-support)
- [What Stays Compatible](#what-stays-compatible)
- [Open Questions](#open-questions)
- [Decision Summary](#decision-summary)

## Motivation and Scope

An API audit of v1.4.0 found two classes of problems:

1. **Inconsistent zero-value semantics.** v1.4.0 fixed `Width`, `Height`,
   `Top/Right/Bottom/Left`, and `GridAutoRows` so that an unset field means the
   CSS initial value (`auto`) and `Px(0)` means a real zero. Round 1 extends
   the same rule to flex sizes and gap fallbacks without breaking anyone. The
   remaining violations cannot be fixed compatibly because their zero value is
   already load-bearing with the wrong meaning (`FlexShrink: 0` means 1,
   `AlignSelf: 0` means "inherit", `Hyphens: 0` means `none`).
2. **Things that cannot be expressed.** `flex-shrink: 0`, a per-item
   `align-self: stretch` under a non-stretch parent, `letter-spacing: -1px`,
   `grid-row: 1` (0-based row 0 with the default span), a real `line-height: 12`
   multiplier, and container-query units inside a normal `Layout()` call.

Everything that *can* be added compatibly (aliases, `String()`/`Parse`,
`Style.Direction`, `GridTemplate*Repeat`, `TextLine.EndsWithForcedBreak`,
`InlineBox.SpaceAfter`, `FrameLength`, generic `FoldNodes`) is Round 1 and is
out of scope here. v2 is the set of changes that require a new module path.

**Non-goals for v2.0:** a real inline formatting context (mixed inline content
in line boxes), floats, multi-column, `calc()`. The enum slots for inline
display types are reserved (Area 5) so the wire format does not have to change
again when they land.

## Guiding Rule: Zero Value Is the CSS Initial Value

For every field of `Style`, `TextStyle`, `GridTrack`, and `Node`:

- `var s Style` behaves exactly like an element with no declarations.
- A constructed value (`Px(0)`, `Some(0.0)`, `AlignSelfStretch`) is always an
  explicit value, never a request for the default.
- No negative-number sentinels. Any field that needs "unset" distinct from a
  legitimate zero uses either a typed enum with an `Auto`/`Normal` zero member,
  a `Length` (whose zero value already means unset), or `Opt[T]`.

`Opt[T]` is the one new generic type this RFC introduces:

```go
// Opt is an optional value. The zero value is "unset" and layout substitutes
// the CSS initial value for the field it appears in.
type Opt[T any] struct {
	V   T
	Set bool
}

func Some[T any](v T) Opt[T]          { return Opt[T]{V: v, Set: true} }
func (o Opt[T]) Or(def T) T            { if o.Set { return o.V }; return def }
```

It is used for exactly three fields (`Flex.Shrink`, `TextStyle.TabSize`,
`Node.Baseline`), all of which have a non-zero CSS initial value that a
`float64` cannot encode. Everything else gets a typed zero.

## Area 1: Zero-Value Violations

### 1.1 `FlexShrink`

**Problem.** `flexbox_items.go:199-201` rewrites `FlexShrink == 0` to `1`, so
`flex-shrink: 0` (a very common declaration: "do not squash this sidebar") is
inexpressible.

```go
// v1: both items shrink equally; there is no way to pin the first one.
row := HStack(
    &Node{Style: Style{Width: Px(300), FlexShrink: 0}}, // treated as 1
    &Node{Style: Style{Width: Px(300)}},
)
LayoutSimple(row, Tight(400, 100)) // both end at 200px
```

**Options.**

| Option | Zero literal | Cost | Verdict |
|---|---|---|---|
| A. `Shrink float64`, 0 means 0 | `Style{Flex: Flex{Grow: 1}}` silently sets shrink 0, breaking the rule | none | rejected: the *shorthand* `flex: 1` in CSS sets shrink to 1; users writing `Flex{Grow: 1}` would get the opposite |
| B. `Shrink *float64` | `&one` boilerplate, allocation, `Style` loses value semantics for this field | low | rejected: pointer fields in a value type that is copied by fluent methods invite aliasing bugs |
| C. `Shrink float64` + `ShrinkSet bool` | two fields to keep in sync | low | workable, but ad hoc; the same pattern recurs for `TabSize` and `Baseline` |
| D. `Shrink Opt[float64]` | `Some(0.0)` | one small generic type | **recommended**: names the intent, comparable, zero is unset, reusable |

**Recommendation.** Group the three flex item properties into one struct and
use `Opt` for shrink:

```go
type Flex struct {
	Grow   float64      // initial 0
	Shrink Opt[float64] // initial 1; Some(0) pins the item
	Basis  Length       // zero value = auto (Round 1 already made Px(0) a real zero basis)
}
```

Constructors mirror the CSS shorthand:

```go
func FlexOf(grow, shrink float64, basis Length) Flex // flex: <grow> <shrink> <basis>
func FlexGrowOnly(grow float64) Flex                  // flex: <grow> 1 0
```

**Migration.** `Style.FlexGrow` becomes `Style.Flex.Grow`, `Style.FlexBasis`
becomes `Style.Flex.Basis`. `FlexShrink: x` with `x != 0` becomes
`Flex.Shrink: Some(x)`. Code that wrote `FlexShrink: 0` intending the default
deletes the field; code that wrote it intending "do not shrink" (and was
silently ignored) writes `Some(0.0)` and gets the behavior it asked for. The
rewrite tool can move the fields mechanically but flags every `FlexShrink: 0`
for a human decision. `WithFlexShrink(0)` on the fluent API likewise becomes
meaningful.

### 1.2 `AlignSelf` and `JustifySelf`

**Problem.** `AlignSelf` reuses the `AlignItems` enum, so
`AlignSelf == AlignItemsStretch == 0` is read as "use the parent's value"
(`flexbox_items.go:108`, `grid.go:338`). An explicit per-item `stretch` under a
`center` parent is impossible. `JustifySelf` has the same defect via
`JustifyItems`.

```go
// v1: the second child cannot opt back into stretch.
row := &Node{Style: Style{Display: DisplayFlex, AlignItems: AlignItemsCenter}}
row.Children = []*Node{
    &Node{Style: Style{Width: Px(50), Height: Px(20)}},
    &Node{Style: Style{Width: Px(50), AlignSelf: AlignItemsStretch}}, // == 0 == inherit
}
```

**Options.** (A) keep the shared enum and add a `AlignSelfSet bool`; (B)
distinct `AlignSelf`/`JustifySelf` types whose zero member is `Auto`. B is what
CSS Box Alignment Level 3 §6.2 does (`align-self: auto` is the initial value,
distinct from `stretch`) and needs no extra field.

**Recommendation.** Option B:

```go
type AlignSelf int
const (
	AlignSelfAuto AlignSelf = iota // initial: defer to the parent's align-items
	AlignSelfStretch
	AlignSelfStart
	AlignSelfEnd
	AlignSelfCenter
	AlignSelfBaseline
)

type JustifySelf int
const (
	JustifySelfAuto JustifySelf = iota
	JustifySelfStretch
	JustifySelfStart
	JustifySelfEnd
	JustifySelfCenter
)
```

While we are renumbering, the canonical container names become `Start`/`End`
(`AlignItemsStart`, `JustifyContentEnd`); `FlexStart`/`FlexEnd` remain as
aliases with the same values, so v1 code that used them compiles unchanged.
Round 1 already adds the aliases in the other direction.

**Migration.** `AlignSelf: AlignItemsCenter` becomes `AlignSelf:
AlignSelfCenter` (mechanical, one-to-one for every non-zero member).
`AlignSelf: AlignItemsStretch` in v1 meant *auto*; the rewrite tool maps it to
omitting the field, not to `AlignSelfStretch`.

### 1.3 `Hyphens`

**Problem.** `HyphensNone` is 0 but the CSS initial value is `manual`
(css-text-3 §4.3). Because `text.go:759` passes `style.Hyphens` straight into
the UAX #14 breaker and `uax14.go:465` suppresses soft-hyphen breaks only for
`HyphensNone`, a default `TextStyle{}` never breaks at U+00AD. Ironically the
convenience wrapper `findLineBreakOpportunities` (`uax14.go:395`) hard-codes
`HyphensManual`, so the two paths disagree.

**Recommendation.** Reorder the enum: `HyphensManual = 0`, `HyphensNone = 1`,
`HyphensAuto = 2`. This is the whole change; no field or type is added.

**Migration.** Source that uses the named constants is unaffected. Serialized
documents are by keyword, so a v1 file that omitted `hyphens` (meaning `none`
in v1) must be loaded as `none`, not `manual`; see Area 6.

### 1.4 `LineHeight`

**Problem.** `resolveLineHeight` (`text.go:1446`) treats `<= 0` as normal,
`< 10` as a multiplier, and `>= 10` as pixels. `line-height: 12` (a 12x
multiplier, unusual but legal) and `line-height: 8px` are both inexpressible,
and the meaning of a value flips at an arbitrary threshold.

**Options.** (A) two fields, `LineHeight Length` and `LineHeightMultiplier
float64`, with a precedence rule when both are set; (B) a `LineHeight` value
type with a kind; (C) reuse `Length` with a unitless "number" unit. C would
need a unit the `units` package does not define and would make "" (unset) and
"number" both special. A leaves an ambiguous state the type system cannot rule
out.

**Recommendation.** Option B:

```go
type LineHeight struct {
	Kind   LineHeightKind // LineHeightNormal = 0
	Number float64        // LineHeightNumber: multiplier of the font size
	Length Length         // LineHeightLength: absolute or relative length
}

func LineHeightMultiplier(n float64) LineHeight
func LineHeightOf(l Length) LineHeight
```

`TextLayout.LineHeight float64` (the resolved pixel value) is unchanged.

**Migration.** `LineHeight: 1.5` becomes `LineHeight:
LineHeightMultiplier(1.5)`; `LineHeight: 24` becomes `LineHeight:
LineHeightOf(Px(24))`. The rewrite tool applies the v1 heuristic to literal
constants and flags non-constant expressions.

### 1.5 `LetterSpacing`, `WordSpacing`, `TextIndent`, `TabSize`

**Problem.** `LetterSpacing` and `WordSpacing` use `-1` for `normal`
(`text.go:34`, `text.go:799`, `text_metrics_adapter.go:78`), so `-1px` is
inexpressible and `TextStyle{}` (0) is "0px", which happens to coincide with
`normal` only by accident. `TabSize` uses `-1` for the default of 8, so
`TextStyle{}` yields a tab size of 0 rather than 8. `TextIndent` is a bare
`float64` of pixels while every other distance in `Style` is a `Length`.

**Recommendation.**

```go
LetterSpacing Length       // zero value = normal; Px(-1) is a real -1px
WordSpacing   Length       // zero value = normal
TextIndent    Length       // zero value = 0 (CSS initial), em/rem/% resolve normally
TabSize       Opt[float64] // zero value = 8 (css-text-3 §3.1.1)
```

`normal` for letter- and word-spacing resolves to 0 today and continues to, so
the only observable change is that negative values become legal and the `-1`
sentinel disappears. `TabSize` uses `Opt` because its initial value is 8 and a
tab size of 0 is a legal (if odd) explicit value.

**Migration.** `LetterSpacing: -1` deletes the field; `LetterSpacing: x`
becomes `LetterSpacing: Px(x)`. `TabSize: -1` deletes the field; `TabSize: 4`
becomes `TabSize: Some(4.0)`.

### 1.6 `Width`/`Height` intrinsic sentinels

**Problem.** `SizeMinContent = -2`, `SizeMaxContent = -3`, `SizeFitContent =
-4` are stored *inside a Length* (`Width: Px(SizeMinContent)`, `api.go:709`)
and detected by `block_sizing.go:24`. They duplicate `WidthSizing`/
`HeightSizing`, only work for the `px` unit, and serialize as the nonsense
`"-2px"`. A user who writes `Width: Px(-2)` by arithmetic accident gets
min-content sizing.

**Recommendation.** Remove the three constants. `WidthSizing`/`HeightSizing`
(renamed `Sizing`, zero member `SizingAuto`) are the only mechanism, with the
existing `FitContentWidth`/`FitContentHeight` lengths as the fit-content limit.
The helpers `MinContentWidth(node)` and friends stay, but set `WidthSizing`.

```go
type Sizing int
const (
	SizingAuto Sizing = iota // use Width/Height, or auto when that is unset
	SizingMinContent
	SizingMaxContent
	SizingFitContent
)
```

**Migration.** `Width: Px(SizeMinContent)` becomes `WidthSizing:
SizingMinContent`; the helper functions need no source change.

### 1.7 `GridTrack` markers

**Problem.** A track's kind is encoded by punning on numeric values:
`Fraction: -1` for fit-content (`api.go:799`, `grid.go:968`), `MaxSize:
Px(-2)`/`Px(-3)` for min-/max-content (`api.go:770`, `grid.go:979`), and two
different unbounded representations, `PxUnbounded` (`Px(MaxFloat64)`) and
`UnboundedLength()` (unit `"unbounded"`), that `serialize` has to fold into one
(`serialize.go`, `LengthJSON` doc). The zero value `GridTrack{}` is detected by
a special case (`grid.go:754`) and treated as `auto`, but `MinMaxTrack(Px(100),
Px(0))` is a legal literal that means something else.

**Options.**

- A. `GridTrack{Kind TrackKind; Min, Max Length; Fraction float64}`, one kind
  per CSS production (`auto`, fixed, minmax, flex, min-content, max-content,
  fit-content). Simple, but `minmax(100px, max-content)` and
  `minmax(min-content, 1fr)` are not expressible, and v1 (accidentally, via
  `MaxSize: Px(-3)`) *can* express the former today, so A would be a
  regression.
- B. Model the grammar of css-grid-1 §7.2 directly: a track is a pair of
  breadths.

**Recommendation.** Option B.

```go
type BreadthKind int
const (
	BreadthAuto BreadthKind = iota // zero value: auto
	BreadthLength                   // Length
	BreadthMinContent
	BreadthMaxContent
	BreadthFlex                     // Flex (fr); only valid as Max
	BreadthFitContent               // fit-content(Length); only valid as Max
)

type TrackBreadth struct {
	Kind   BreadthKind
	Length Length  // BreadthLength, BreadthFitContent
	Flex   float64 // BreadthFlex
}

type GridTrack struct {
	Min TrackBreadth // inflexible breadth (css-grid-1 §7.2: <inflexible-breadth>)
	Max TrackBreadth
}
```

`GridTrack{}` is `minmax(auto, auto)`, which is exactly what `auto` means, so
the special case in `grid.go:754` disappears. Constructors keep their v1 names
and gain typed arguments:

```go
func FixedTrack(size Length) GridTrack
func MinMaxTrack(min, max TrackBreadth) GridTrack
func FractionTrack(fr float64) GridTrack
func AutoTrack() GridTrack
func MinContentTrack() GridTrack
func MaxContentTrack() GridTrack
func FitContentTrack(limit Length) GridTrack   // was float64

func BreadthOf(l Length) TrackBreadth
func Fr(v float64) TrackBreadth
func MinContent() TrackBreadth
func MaxContent() TrackBreadth
```

`PxUnbounded` is removed. `UnboundedLength()` remains for `Rect`/constraint
plumbing but is no longer needed to describe a track.

**Migration.** Callers of the constructors change only `FitContentTrack(300)`
to `FitContentTrack(Px(300))` and `MinMaxTrack(Px(a), Px(b))` to
`MinMaxTrack(BreadthOf(Px(a)), BreadthOf(Px(b)))`. Struct literals are rewritten
by hand; the rewrite tool reports them.

### 1.8 Grid line placement

**Problem.** `gridNormalizeSpan` (`grid_placement.go:47`) treats `start < 0`
*or* `start == 0 && end <= 0` as auto. "Place at 0-based row 0 with the default
span" therefore requires writing `GridRowEnd: 1` as well; `GridRowStart: 0`
alone is silently auto-placed. Resolving a named `GridArea` also *writes the
resolved lines back into the child's Style* (`grid_placement.go:369-372`), so a
tree laid out once and then re-used with a different template keeps stale
placement.

**Recommendation.**

```go
type GridLineKind int
const (
	GridLineAuto GridLineKind = iota
	GridLineIndex // 0-based line index, matching v1 and GridArea
	GridLineSpan  // span N from the other edge
)

type GridLine struct {
	Kind  GridLineKind
	Value int
}

type GridPlacement struct{ Start, End GridLine }

func LineAt(i int) GridLine
func Span(n int) GridLine
```

`Style.GridRow` and `Style.GridColumn` are `GridPlacement`s. Indices stay
0-based: switching to CSS's 1-based lines in the same release as everything
else is a footgun with no compensating benefit, and `GridArea` and
`GridTemplateAreas.DefineArea` are already 0-based. Area resolution stores the
result in the per-layout item table, never in `Style`.

**Migration.** `GridRowStart: 1, GridRowEnd: 3` becomes `GridRow:
GridPlacement{Start: LineAt(1), End: LineAt(3)}`; `-1` becomes the zero value.
Mechanical.

### 1.9 `Node.Baseline`

**Problem.** `Baseline == 0` means "no baseline; use the cross size"
(`flexbox_positioning.go:217`). A node whose baseline really is at 0 (text
sitting on its top edge, or an icon whose alignment point is its top) cannot say
so.

**Recommendation.** `Baseline Opt[float64]`, consistent with 1.1 and 1.5.
`*float64` was rejected for the same aliasing reason as in 1.1, and a separate
`HasBaseline` bool is the ad hoc form of `Opt`.

## Area 2: One Layout Entry Point

**Problem.** v1 has four entry points with different scopes:

- `Layout(root, constraints, ctx)` performs flow layout only. Positioned
  descendants keep their static position.
- `LayoutWithPositioning(root, constraints, viewportRect, ctx)` runs `Layout`
  and a second recursive pass; the viewport is a separate parameter even
  though `ctx` already has `ViewportWidth`/`ViewportHeight`.
- `LayoutPositioned(node, containingBlock, viewportRect, ctx)` is the per-node
  step, exported, and cannot find the nearest positioned ancestor
  (`findPositionedAncestor` returns `nil`, `positioned.go:241`) because it has
  no ancestor chain.
- `LayoutSimple(root, constraints)` is `Layout` with a default context.

`ZStack` documents that it "requires" `LayoutWithPositioning`. The split also
means container-query units never resolve inside a normal layout: every
algorithm calls `ResolveLength`, which returns 0 for `cq*` because it has no
ancestor information (`length.go:164`), while the ancestor-aware
`ResolveLengthInContext` needs a `NodeContext` nobody builds during layout.

```go
// v1: cqw resolves to 0 in the normal path.
root := &Node{Style: Style{ContainerType: ContainerTypeInlineSize, Width: Px(400)}}
root.Children = []*Node{{Style: Style{Width: Cqw(50), Height: Px(10)}}}
LayoutSimple(root, Loose(800, 600))
// root.Children[0].Rect.Width == 0, expected 200
```

**Options.**

- A. Keep the split, add a `viewport` field to `LayoutContext`, and make
  `LayoutWithPositioning` the documented default. Leaves two ways to lay out a
  tree and does not fix `cq*`.
- B. `Layout` does everything; `LayoutContext` carries the viewport and an
  internal ancestor stack pushed and popped by the algorithms.

**Recommendation.** Option B.

```go
type LayoutContext struct {
	Viewport        Rect               // replaces ViewportWidth/ViewportHeight and the viewportRect parameter
	ScrollOffset    Point              // scroll position of the viewport, for position: sticky (Area 5)
	RootFontSize    float64
	TextMetrics     TextMetricsProvider // Area 3; nil means DefaultTextMetrics()
	ChReferenceChar rune

	ancestors []*Node // maintained by Layout; read by ResolveLength for cq* and by positioning for the containing block
}

func NewLayoutContext(viewport Rect, rootFontSize float64) *LayoutContext

// Layout performs flow layout, then positions absolute, fixed, relative, and
// sticky descendants, and returns the root's border-box size.
func Layout(root *Node, constraints Constraints, ctx *LayoutContext) Size

// LayoutSimple is Layout with NewLayoutContext(Rect{W: constraints.MaxWidth, H: constraints.MaxHeight}, 16).
func LayoutSimple(root *Node, constraints Constraints) Size
```

`LayoutBlock`, `LayoutFlexbox`, `LayoutGrid`, and `LayoutText` remain exported
as subtree algorithms for callers who dispatch themselves, but they never run
the positioning pass; only `Layout` does. `LayoutWithPositioning` and
`LayoutPositioned` are removed. `ResolveLength(l, ctx, fontSize)` consults
`ctx.ancestors` for `cq*`, so `ResolveLengthInContext` is removed as well; a
caller outside of layout who needs `cq*` resolution constructs a
`LayoutContext` with `ctx.WithAncestors(nctx *NodeContext)`.

The ancestor stack also lets positioning find the nearest positioned ancestor
(css-position-3 §3.1) instead of always using the parent's padding box. That is
a behavior change for absolutely positioned grandchildren of a non-positioned
parent and is documented as such.

**Migration.** `LayoutWithPositioning(root, c, vp, ctx)` becomes
`ctx.Viewport = vp; Layout(root, c, ctx)`. `Layout(root, c, ctx)` is unchanged
in signature; trees with positioned nodes now get positioned, which is what
every caller we could find expected. `NewLayoutContext(w, h, fs)` becomes
`NewLayoutContext(Rect{Width: w, Height: h}, fs)`. The rewrite tool handles all
three.

## Area 3: TextMetricsProvider

**Problem.** The interface returns a three-tuple and takes `TextStyle` by
value (a struct of roughly thirty fields copied per call):

```go
Measure(text string, style TextStyle) (advance, ascent, descent float64)
```

There is no way to ask for the font's line metrics without measuring a string,
so `ascent`/`descent` are returned for every run and the adapter fabricates
them as 80%/20% of the line height (`text_metrics_adapter.go:98`). Worse, the
provider is a process-wide global (`text.go:53`, `SetTextMetricsProvider`).
`LayoutContext.TextMetrics` exists but is consulted **only** for `ch` units
(`length.go:224`); line breaking (`text.go:592`), hanging punctuation, and
intrinsic sizing (`intrinsic_sizing.go:324`) call `getTextMetrics()`. So
`ctx.WithTextMetrics(harfbuzz)` changes `ch` resolution and nothing else, and
two goroutines laying out with different fonts race on the global.

**Recommendation.**

```go
type Metrics struct {
	Advance float64 // inline-axis advance of the run
	Ascent  float64 // above the alphabetic baseline; 0 if unknown
	Descent float64 // below the baseline; 0 if unknown
}

type LineMetrics struct {
	Ascent, Descent, LineGap float64 // css-inline-3 §4.4.1 "normal" = Ascent + Descent + LineGap
}

type TextMetricsProvider interface {
	Measure(text string, style *TextStyle) Metrics
	LineMetrics(style *TextStyle) LineMetrics
}

func DefaultTextMetrics() TextMetricsProvider // the v1 approximation; used when ctx.TextMetrics is nil
```

The global and `SetTextMetricsProvider` are removed. Every measurement goes
through `ctx.TextMetrics`. `line-height: normal` resolves to
`LineMetrics(style)` instead of the hard-coded 1.2 multiplier when the provider
can supply it; the default provider returns 1.2 x font size split 80/20 so
behavior for existing users is unchanged. `NewTerminalTextMetrics` and
`NewTextMetricsAdapter` implement the new interface.

Keeping the global "only as the default" was considered: it would preserve
`SetTextMetricsProvider` for programs that call `LayoutSimple` everywhere. It
was rejected because it keeps a mutable global in a library that is otherwise
pure, and `LayoutSimple` callers can switch to `Layout` with a context in one
line.

**Migration.** Implementers change the signature and return a struct; the
rewrite tool cannot do this. Consumers replace
`layout.SetTextMetricsProvider(p)` with `ctx.TextMetrics = p` (or
`ctx.WithTextMetrics(p)`, which is kept).

## Area 4: API Hygiene

### 4.1 Design-tool helpers move to a subpackage

`AlignNodes`, `DistributeNodes`, `SnapNodes`, `SnapToGrid`, `AlignEdge`, and
`DistributeDirection` mutate `Rect`s after layout. They are not CSS and their
presence in the root package suggests that post-layout mutation of `Rect` is a
supported pattern for everything. They move to
`github.com/SCKelemen/layout/v2/arrange` with unchanged signatures.

### 4.2 Removals

| Symbol | Reason | Replacement |
|---|---|---|
| `Background(node)` | no-op placeholder (`api.go:409`) | none |
| `CollectNodesForSVG(root, &nodes)` | out-parameter API duplicating the fluent traversal | `root.DescendantsAndSelf()` |
| `GetSVGTransform(node)` | one-line wrapper | `node.Style.Transform.ToSVGString()` |
| `GetFinalRect(node)` | free function on a node | `node.TransformedRect()` |
| `InlineBox.Kind`, `InlineBox.Node`, `InlineBoxKind` | placeholders for inline nodes that never landed; an `InlineBox` is a text run | re-added with real semantics when the inline formatting context ships |
| `(*Node).Fold`, `FoldWithContext` with `interface{}` | methods cannot be generic | `Fold[T](n, init, fn)` and `FoldDepth[T]` package functions (Round 1 ships them as `FoldNodes`; v2 takes the short name) |
| `SizeMinContent/MaxContent/FitContent`, `PxUnbounded` | Area 1 | `Sizing`, `TrackBreadth` |
| `LayoutWithPositioning`, `LayoutPositioned`, `ResolveLengthInContext`, `SetTextMetricsProvider` | Areas 2 and 3 | `Layout`, `LayoutContext` |

### 4.3 Geometry helpers

`Rect`, `Size`, and `Constraints` gain the methods every renderer re-implements:

```go
func (r Rect) Right() float64
func (r Rect) Bottom() float64
func (r Rect) Contains(p Point) bool
func (r Rect) Inset(top, right, bottom, left float64) Rect
func (r Rect) Union(o Rect) Rect
func (r Rect) IsEmpty() bool

func (s Size) IsEmpty() bool

func (c Constraints) IsTight() bool
func (c Constraints) IsBounded() bool // both maxima < Unbounded
```

These are additive and could ship in v1.x; they are listed here because
`Rect.Right()` collides with nothing today but must be reserved before v2 adds
`Inset` as a struct name (see the `Style` sketch, where `Inset` groups the four
offsets).

### 4.4 Helpers take `Length`

Every helper that sets a `Length`-typed property takes a `Length`. The
`float64`-pixel convenience layer is removed rather than kept in parallel, so
there is exactly one way to write each thing:

```go
func Frame(node *Node, width, height Length) *Node   // was float64 with a "> 0" guard, so Frame(n, 0, 100) could not set width 0
func Fixed(width, height Length) *Node
func Padding(node *Node, p Length) *Node
func Margin(node *Node, m Length) *Node
func MinWidth(node *Node, w Length) *Node
func FitContentWidth(node *Node, limit Length) *Node
func (n *Node) WithWidth(w Length) *Node
func (n *Node) WithHeight(h Length) *Node
func (n *Node) WithPadding(p Length) *Node
func (n *Node) WithMargin(m Length) *Node
```

The `Frame` guard is the concrete bug: an unset dimension is the zero
`Length`, not zero pixels, so the guard is unnecessary and the function can
finally set `Px(0)`. Round 1's `FrameLength` is the compatible bridge and is
dropped in v2.

**Migration.** `WithWidth(300)` becomes `WithWidth(Px(300))`; the rewrite tool
wraps numeric arguments in `Px(...)` for every function in this list.

## Area 5: Display and Sticky Additions

### 5.1 `Display`

v1 has `Block`, `Flex`, `Grid`, `InlineText`, `None`. v2 adds:

```go
DisplayBlock Display = iota
DisplayFlex
DisplayGrid
DisplayInlineText
DisplayNone
DisplayInlineBlock // new
DisplayInlineFlex  // new
DisplayInlineGrid  // new
DisplayFlowRoot    // new
DisplayContents    // new
```

New members are appended so existing integer values are stable. v2.0 semantics,
honestly scoped:

- `InlineBlock`/`InlineFlex`/`InlineGrid`: lay out as the block-level
  counterpart with **shrink-to-fit inline size** (css-sizing-3 §5.1 fit-content)
  when the parent is a block container. Consecutive inline-level siblings are
  *not* yet placed on a shared line box; that requires the inline formatting
  context, which is a non-goal for v2.0. The enum is added now so documents and
  code written against v2.0 do not change when line boxes land in v2.x.
- `FlowRoot`: a block container that establishes a new block formatting
  context: margins do not collapse through it (CSS 2.1 §8.3.1) and its auto
  height includes the margins of its children.
- `Contents`: the node generates no box; its children participate in the
  parent's formatting context as if they were the parent's own children
  (css-display-3 §2.5). The node's `Rect` is the union of its children's.

### 5.2 `position: sticky`

v1 treats sticky as relative (`positioned.go:77`). With `Viewport` and
`ScrollOffset` on the context (Area 2), v2 implements css-position-3 §7 against
the viewport as the sole scroll container: the box's normal-flow position is
shifted so that it stays inside the sticky view rectangle (`Viewport` offset by
`ScrollOffset` and inset by the sticky offsets) and is clamped to its containing
block. Nested scroll containers are out of scope; `overflow` is not a v2.0
property.

Both additions are additive from a source perspective; they are in this RFC
because they change what a `Display` value means when serialized and because
sticky needs the `LayoutContext` shape from Area 2.

## Area 6: Serialization

**Problem.** A v1 document is a bare node object with no format marker
(`serialize.go:40`, `NodeJSON`). The enum tables map 0 to `""` and omit it,
so a v1 file cannot distinguish "not set" from "the zero member" — which was
fine while the two coincided, and is exactly what Area 1 breaks: in v2
`alignSelf` absent means `auto` and `hyphens` absent means `manual`, while a v1
writer omitted them meaning `stretch`/`inherit` and `none`.

**Recommendation.**

1. **Envelope with a version.** v2 writes

   ```json
   {"version": 2, "root": {"style": {...}, "children": [...]}}
   ```

   Option A (a `"format"` field on the root node) was rejected because it puts a
   document-level property on one node and is lost if the root is ever
   re-parented. The decoder accepts three shapes: an object with `version`
   and `root` (v2+), an object with `style` or `children` at the top level (v1,
   `version` implied 1), anything else is an error. Unknown future versions
   are rejected, not guessed.

2. **v1 loading is a translation, not a reinterpretation.** `FromJSON` on a v1
   document applies these rules before decoding as v2:

   | v1 field | v1 meaning | v2 value written |
   |---|---|---|
   | `flexShrink` absent or `0` | 1 | `flex.shrink` absent |
   | `flexShrink: x` | x | `flex.shrink: x` |
   | `alignSelf` absent | inherit | absent (`auto`) |
   | `alignSelf: "center"` etc. | as named | same keyword |
   | `justifySelf` | same as `alignSelf` | |
   | `hyphens` absent | none | `hyphens: "none"` |
   | `textStyle.lineHeight: n` | heuristic | `n <= 0` absent; `n < 10` `{"number": n}`; else `{"length": "<n>px"}` |
   | `letterSpacing`/`wordSpacing: -1` | normal | absent |
   | `letterSpacing: n` | n px | `"<n>px"` |
   | `tabSize: -1` | 8 | absent |
   | `width: "-2px"`, `"-3px"`, `"-4px"` | intrinsic sentinel | `widthSizing` keyword, `width` absent |
   | track `maxSize: "-2px"`/`"-3px"` | min-/max-content | `max: {"kind": "min-content"}` |
   | track `fraction: -1` | fit-content(maxSize) | `max: {"kind": "fit-content", "length": ...}` |
   | track `fraction: f > 0` | f fr | `max: {"kind": "flex", "flex": f}` |
   | `maxSize: "unbounded"` (or legacy bare 1.79e308) | auto | `max` absent |
   | `gridRowStart: -1`, or `0` with `gridRowEnd <= 0` | auto | `gridRow` absent |
   | `gridRowStart: s`, `gridRowEnd: e` | lines | `gridRow: {"start": s, "end": e}` |
   | `baseline` absent or `0` | none | absent |
   | `display` keywords | unchanged | unchanged |

   The legacy bare-number length form (pixels) is still accepted on v1 input
   and is an error on v2 input.

3. **`serialize.Upgrade(v1 []byte) ([]byte, error)`** performs the translation
   and re-encodes as v2 without constructing a `layout.Node`, for users who want
   to migrate files in bulk and diff the result.

4. Keyword names for enums are unchanged. New members (`inline-block`,
   `flow-root`, `contents`, `start`/`end`) are added to the tables.

## Proposed v2 Style

```go
type Style struct {
	// Box generation and positioning scheme
	Display  Display  // DisplayBlock = 0
	Position Position // PositionStatic = 0
	Inset    Inset    // Top, Right, Bottom, Left Length; zero value = auto
	ZIndex   int

	// Sizing (css-sizing-3). Zero-value Length = auto (Width/Height/Min*) or none (Max*).
	Width, Height             Length
	MinWidth, MinHeight       Length
	MaxWidth, MaxHeight       Length
	WidthSizing, HeightSizing Sizing // SizingAuto = 0
	FitContentWidth           Length // limit for SizingFitContent
	FitContentHeight          Length
	AspectRatio               float64 // 0 = auto
	BoxSizing                 BoxSizing

	// Box model
	Padding, Margin, Border Spacing

	// Flex container
	FlexDirection FlexDirection
	FlexWrap      FlexWrap
	// Flex item
	Flex  Flex // Grow, Shrink Opt[float64], Basis
	Order int

	// Gaps, shared by flex and grid (css-align-3 §8). Zero value = 0.
	Gap Gap // Row, Column Length

	// Grid container
	GridTemplateRows      []GridTrack
	GridTemplateColumns   []GridTrack
	GridTemplateRowsRepeat, GridTemplateColumnsRepeat *RepeatTrack // Round 1
	GridAutoRows          GridTrack // zero value = auto
	GridAutoColumns       GridTrack
	GridAutoFlow          GridAutoFlow
	GridTemplateAreas     *GridTemplateAreas
	// Grid item
	GridRow    GridPlacement // Start, End GridLine; zero value = auto / auto
	GridColumn GridPlacement
	GridArea   string

	// Alignment (css-align-3)
	JustifyContent JustifyContent // Start = 0 (FlexStart alias)
	AlignContent   AlignContent   // Stretch = 0
	JustifyItems   JustifyItems   // Stretch = 0
	AlignItems     AlignItems     // Stretch = 0
	JustifySelf    JustifySelf    // Auto = 0
	AlignSelf      AlignSelf      // Auto = 0

	// Writing system and containment
	WritingMode   WritingMode
	Direction     Direction // Round 1; TextStyle.Direction removed in v2
	ContainerType ContainerType
	ContainerName ContainerName

	// Rendering hint carried through layout
	Transform Transform // zero value = identity

	// Text (nil for non-text nodes)
	Text *TextStyle
}

type Inset struct{ Top, Right, Bottom, Left Length }
type Gap struct{ Row, Column Length }

type TextStyle struct {
	TextAlign     TextAlign
	TextAlignLast TextAlignLast
	TextJustify   TextJustify

	LineHeight    LineHeight   // zero value = normal
	LetterSpacing Length       // zero value = normal
	WordSpacing   Length       // zero value = normal
	TextIndent    Length       // zero value = 0
	TabSize       Opt[float64] // zero value = 8

	WhiteSpace   WhiteSpace
	OverflowWrap OverflowWrap
	WordBreak    WordBreak
	TextOverflow TextOverflow

	TextTransform      TextTransform
	Hyphens            Hyphens // HyphensManual = 0
	HangingPunctuation HangingPunctuation

	FontSize   float64 // 0 = inherit from context (16 at the root)
	FontFamily string
	FontWeight FontWeight // 0 = 400
	FontStyle  FontStyle

	TextDecoration      TextDecoration
	TextDecorationStyle TextDecorationStyle
	TextDecorationColor string

	VerticalAlign VerticalAlign
	// WritingMode and Direction removed: both live on Style and inherit.
}

type Node struct {
	Style      Style
	Rect       Rect
	Children   []*Node
	Baseline   Opt[float64]
	Text       string
	TextLayout *TextLayout
}
```

Changes relative to v1 that are purely organizational (`Inset`, `Gap`, `Flex`,
`GridRow`/`GridColumn`, `Style.Text` instead of `Style.TextStyle`) are made
because the module path is changing anyway and this is the only chance to group
fields without breaking anyone twice. `FlexGap`/`GridGap` collapse into one
`Gap` because CSS has one `gap` property; the Round 1 fallback rules become
simply "`Gap.Row` and `Gap.Column`, each zero unless set".

## Migration Table

| v1 | v2 | Mechanical? |
|---|---|---|
| `import "github.com/SCKelemen/layout"` | `import "github.com/SCKelemen/layout/v2"` | yes |
| `Style.FlexGrow` | `Style.Flex.Grow` | yes |
| `Style.FlexShrink: x` (x != 0) | `Style.Flex.Shrink: Some(x)` | yes |
| `Style.FlexShrink: 0` | delete, or `Some(0.0)` if "do not shrink" was intended | **no** (flagged) |
| `Style.FlexBasis` | `Style.Flex.Basis` | yes |
| `Style.FlexGap`, `GridGap` | `Style.Gap{Row: g, Column: g}` | yes |
| `Style.FlexRowGap`, `GridRowGap` | `Style.Gap.Row` | yes |
| `Style.FlexColumnGap`, `GridColumnGap` | `Style.Gap.Column` | yes |
| `Style.AlignSelf: AlignItemsX` (X != Stretch) | `Style.AlignSelf: AlignSelfX` | yes |
| `Style.AlignSelf: AlignItemsStretch` | delete (was auto) | yes |
| `Style.JustifySelf: JustifyItemsX` | `Style.JustifySelf: JustifySelfX` | yes (same Stretch rule) |
| `AlignItemsFlexStart` etc. | unchanged (alias of `AlignItemsStart`) | n/a |
| `Style.Top/Right/Bottom/Left` | `Style.Inset.Top/...` | yes |
| `Style.WidthSizing IntrinsicSize` | `Style.WidthSizing Sizing` | yes (renamed members) |
| `Width: Px(SizeMinContent)` | `WidthSizing: SizingMinContent` | yes |
| `Style.GridRowStart/End` | `Style.GridRow: GridPlacement{Start: LineAt(s), End: LineAt(e)}` | yes; `-1` and `0/0` become zero value |
| `GridTrack{MinSize, MaxSize, Fraction}` literal | `GridTrack{Min, Max TrackBreadth}` | **no** (flagged); constructors are mechanical |
| `FitContentTrack(300)` | `FitContentTrack(Px(300))` | yes |
| `MinMaxTrack(Px(a), Px(b))` | `MinMaxTrack(BreadthOf(Px(a)), BreadthOf(Px(b)))` | yes |
| `PxUnbounded` | `UnboundedLength()` (rects) or drop (tracks) | yes |
| `Style.TextStyle` | `Style.Text` | yes |
| `TextStyle.LineHeight: n` (constant) | `LineHeightMultiplier(n)` if n < 10 else `LineHeightOf(Px(n))` | yes for literals, flagged otherwise |
| `TextStyle.LetterSpacing/WordSpacing: -1` | delete | yes |
| `TextStyle.LetterSpacing: n` | `Px(n)` | yes |
| `TextStyle.TextIndent: n` | `Px(n)` | yes |
| `TextStyle.TabSize: -1` | delete | yes |
| `TextStyle.TabSize: n` | `Some(n)` | yes |
| `TextStyle.WritingMode`, `TextStyle.Direction` | `Style.WritingMode`, `Style.Direction` | yes |
| `HyphensNone`/`Manual`/`Auto` | unchanged names, renumbered | n/a |
| `Node.Baseline: b` | `Node.Baseline: Some(b)` | yes; `0` is flagged |
| `Layout(root, c, ctx)` | unchanged; now also positions | n/a (behavior) |
| `LayoutWithPositioning(root, c, vp, ctx)` | `ctx.Viewport = vp; Layout(root, c, ctx)` | yes |
| `LayoutPositioned(...)` | none (internal) | **no** |
| `LayoutSimple(root, c)` | unchanged | n/a |
| `NewLayoutContext(w, h, fs)` | `NewLayoutContext(Rect{Width: w, Height: h}, fs)` | yes |
| `ctx.ViewportWidth/Height` | `ctx.Viewport.Width/Height` | yes |
| `ResolveLengthInContext(l, ctx, fs, nctx)` | `ResolveLength(l, ctx.WithAncestors(nctx), fs)` | yes |
| `SetTextMetricsProvider(p)` | `ctx.TextMetrics = p` | **no** (needs a ctx in scope; flagged) |
| `Measure(text, style TextStyle) (a, asc, desc)` | `Measure(text, *TextStyle) Metrics` + `LineMetrics` | **no** |
| `AlignNodes`, `DistributeNodes`, `SnapNodes`, `SnapToGrid` | `arrange.` prefix | yes |
| `Background(n)` | delete | yes |
| `CollectNodesForSVG(root, &ns)` | `ns = root.DescendantsAndSelf()` | yes |
| `GetSVGTransform(n)` | `n.Style.Transform.ToSVGString()` | yes |
| `GetFinalRect(n)` | `n.TransformedRect()` | yes |
| `n.Fold(init, fn)` | `Fold(n, init, fn)` with typed accumulator | **no** (closure types change) |
| `Frame(n, w, h)`, `Fixed(w, h)`, `Padding(n, p)`, `Margin(n, m)`, `MinWidth`, `MinHeight`, `FitContentWidth/Height` | same names, `Length` arguments | yes (`Px(...)` wrap) |
| `n.WithWidth(w)`, `WithHeight`, `WithPadding`, `WithMargin`, `With*Custom` | `Length` arguments | yes |
| `InlineBox.Kind`, `InlineBox.Node` | delete | yes |
| `serialize.ToJSON` output | envelope `{"version": 2, "root": ...}` | n/a; v1 files load |

## Migration Tooling and v1 Support

**`go fix` is not available for this.** Modern `go fix` only applies the
toolchain's own rewrites, and `gofmt -r` handles single-expression patterns
without type information (it cannot tell `layout.Style` fields from another
package's, and cannot wrap only *numeric* arguments in `Px`). The plan is:

1. **`cmd/layoutfix`**, an `x/tools/go/analysis` driver shipped in the v2
   module with `SuggestedFix`es for every row marked "yes" above and a
   diagnostic (no fix) for every row marked "no" or "flagged". Run as
   `go run github.com/SCKelemen/layout/v2/cmd/layoutfix@latest -fix ./...`
   after changing the import path. It is type-aware, so it only touches
   identifiers resolved to the `layout` package.
2. **v1.x deprecation shims are limited.** Most v2 changes redefine the meaning
   of existing zero values, which by definition cannot be shimmed in v1 without
   breaking v1. What v1.x *can* do, and Round 1 does: add the aliases and
   `Length`-typed helper variants, ship `FoldNodes`, and mark
   `LayoutWithPositioning`, `LayoutPositioned`, `SetTextMetricsProvider`,
   `Background`, `CollectNodesForSVG`, `GetSVGTransform`, `PxUnbounded`, and
   the `Size*Content` constants `// Deprecated:` with a pointer to this RFC once
   v2.0.0 is tagged. Users who adopt the Round 1 names first have fewer
   `layoutfix` diagnostics to review.
3. **v1 maintenance.** v1 receives bug fixes for twelve months after v2.0.0 and
   security fixes (the `serialize` input-limit code) for as long as the module
   is published. No new features land in v1 after v2.0.0.
4. **`serialize.Upgrade`** (Area 6) migrates stored documents independently of
   source code.

## What Stays Compatible

- `Node`'s shape (`Style`, `Rect`, `Children`, `Text`, `TextLayout`); only
  `Baseline` changes type.
- `Rect`, `Size`, `Point`, `Constraints`, `Tight`, `Loose`, `Unconstrained`,
  `Unbounded`, `Constrain`.
- `Length`/`LengthUnit` as aliases of `units.Length`/`units.LengthUnit`; every
  constructor (`Px`, `Em`, `Rem`, `Ch`, `Vw`, `Cqw`, ...); `UnboundedLength()`.
- `Spacing`, `Uniform`, `Horizontal`, `Vertical`.
- `Display{Block,Flex,Grid,InlineText,None}` names and values; `FlexDirection`,
  `FlexWrap`, `JustifyContent`, `AlignItems`, `AlignContent`, `JustifyItems`,
  `GridAutoFlow`, `BoxSizing`, `Position`, `WritingMode`, `Direction`,
  `ContainerType`, and every `TextStyle` enum except `Hyphens` keep their
  names and values.
- `Transform` and all of its constructors and methods.
- `GridArea`, `GridTemplateAreas`, `NewGridTemplateAreas`, `DefineArea`,
  `PlaceInArea`, `RepeatTracks`, `AutoFillTracks`, `AutoFitTracks`.
- `HStack`, `VStack`, `ZStack`, `Spacer`, `Grid`, `GridAuto`,
  `GridFractional`, `AspectRatio`.
- All fluent traversal and mutation methods except the `Fold` pair and the
  `float64`-taking `With*` methods; nil-safety and copy-on-write semantics as
  documented in `fluent-api-design-decisions.md`.
- `NodeContext` and its methods.
- `serialize` keyword vocabulary; v1 documents load without user action.
- `LayoutContext.RootFontSize`, `ChReferenceChar`, `WithTextMetrics`,
  `WithChReferenceChar`.

## Open Questions

1. **`Opt[T]` versus three ad hoc `Has*` bools.** `Opt` is used in exactly three
   places. If reviewers find a generic type disproportionate for three fields,
   the fallback is `FlexShrinkSet`, `TabSizeSet`, and `HasBaseline` bools; the
   RFC prefers `Opt` for uniformity and because `Some(0.0)` reads as intent.
2. **0-based grid lines.** Keeping 0-based indices is proposed for continuity
   with `GridArea`. The alternative, adopting CSS's 1-based lines with negative
   indices counting from the end, would be the *only* opportunity to do so
   without a v3. Decision requested.
3. **Exported subtree algorithms.** `LayoutBlock`/`LayoutFlexbox`/`LayoutGrid`/
   `LayoutText` are kept exported for dispatch-it-yourself callers. If nobody
   does that, unexporting them removes four entry points that skip positioning.
4. **`TextStyle` inheritance.** `FontSize: 0` currently means 16 at every node.
   With the ancestor stack from Area 2 it could inherit from the nearest
   ancestor with a `Text` style, which is what CSS does. That is a behavior
   change and is proposed as a follow-up, not part of v2.0.
5. **Package name for the design-tool helpers.** `arrange` is proposed;
   `designtool` and `post` were considered.

## Decision Summary

| Area | Decision | Rejected alternatives |
|---|---|---|
| Zero values | zero == CSS initial everywhere; `Opt[T]` for the three non-zero initials | pointer fields, paired bools, negative sentinels |
| `FlexShrink` | `Flex{Grow, Shrink Opt[float64], Basis}` | `Shrink float64` meaning 0 |
| `AlignSelf`/`JustifySelf` | distinct types with `Auto = 0` | shared enum plus `Set` bool |
| `Hyphens` | `Manual = 0` | keep and document |
| `LineHeight` | kind + number + length struct | two fields, unitless `Length` |
| Spacing/indent/tab | `Length`; `TabSize Opt` | `-1` sentinels |
| Intrinsic sizing | `Sizing` enum only | keep `-2/-3/-4` |
| `GridTrack` | `Min, Max TrackBreadth` | `Kind + Min + Max + Fraction` (cannot express intrinsic minmax) |
| Grid placement | `GridPlacement{Start, End GridLine}`, 0-based | 1-based CSS lines (open) |
| `Baseline` | `Opt[float64]` | `*float64`, `HasBaseline` |
| Entry point | `Layout` does flow + positioning; viewport and ancestors in `LayoutContext` | keep two-pass API |
| Text metrics | struct-returning interface, no global | keep global as default |
| Helpers | `arrange` subpackage; `Length` everywhere; generic `Fold` | parallel `float64` layer |
| Display/sticky | append five members; sticky against viewport + `ScrollOffset` | defer until inline layout exists |
| Serialization | `{"version", "root"}` envelope; v1 translated on load; `Upgrade` | version field on the root node |
