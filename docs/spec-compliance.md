# Spec Compliance

What each CSS module looks like in this engine. **yes** = implemented and covered by tests; **partial** = implemented with a documented restriction; **no** = not implemented (see [Limitations](limitations.md) for the reasons and [docs/design/v2-api.md](design/v2-api.md) for planned fixes). Every entry refers to the exported API in `Style` / `TextStyle`.

## CSS Flexible Box Layout Level 1

| Feature | Status | Notes |
|---------|--------|-------|
| `flex-direction` row, row-reverse, column, column-reverse | yes | |
| `flex-wrap` nowrap, wrap, wrap-reverse | yes | |
| `justify-content` flex-start, flex-end, center, space-between, space-around, space-evenly | yes | `stretch` behaves as flex-start; negative free space falls back per §9.5 |
| `align-items` / `align-self` stretch, flex-start, flex-end, center, baseline | yes | stretch only when the cross size is auto; explicit `align-self: stretch` cannot override a non-stretch parent (zero value = inherit) |
| `align-content` stretch, flex-start, flex-end, center, space-between, space-around, space-evenly | yes | |
| `flex-grow`, `flex-shrink`, `flex-basis` | partial | §9.7 resolution with scaled shrink factors and min/max clamping; `flex-shrink: 0` is not expressible |
| `gap`, `row-gap`, `column-gap` | yes | row/column follow the flex direction (css-align-3 §8.3) |
| `order` | yes | |
| Baseline alignment | yes | column direction behaves as flex-start |
| Indefinite main size, nested containers, out-of-flow children | yes | |
| Intrinsic sizes (§9.9) | partial | simplified: sum of items for rows, max for columns |

## CSS Grid Layout Level 1

| Feature | Status | Notes |
|---------|--------|-------|
| Track sizing: fixed, `fr`, `auto`, `minmax()`, `min-content`, `max-content`, `fit-content()` | yes | `1fr` = `minmax(auto, 1fr)`; base sizes from min-content contributions (§12.5) |
| `repeat(N)`, `repeat(auto-fill)`, `repeat(auto-fit)` | yes | via `GridTemplateColumnsRepeat`/`GridTemplateRowsRepeat`; fixed sizes only in auto-repeat; empty auto-fit tracks collapse with their gutters |
| Implicit tracks (`grid-auto-rows`/`columns`) | yes | zero value = `auto` |
| Line-based placement, spanning, negative/swapped lines | yes | 0-based lines; `end == start` spans 1; `end < start` swaps |
| Named areas (`grid-template-areas`, `grid-area`) | yes | structured `GridTemplateAreas`, not the string syntax |
| Auto-placement sparse/dense, row/column | yes | §8.5 order: definite items, locked items, cursor |
| `gap`, `row-gap`, `column-gap` | yes | unset longhand falls back to the shorthand |
| `justify-items`/`justify-self`, `align-items`/`align-self` | yes | stretch is a no-op on a definite axis (css-align-3 §6.2) |
| `justify-content`, `align-content` (start, end, center, space-*, stretch) | yes | stretch grows `auto`-max tracks |
| Margins, `aspect-ratio`, intrinsic sizing modes on items | yes | |
| Writing modes | yes | logical axes and margins mapped |
| Subgrid (Level 2) | no | |

## Block layout (CSS 2.1 §8-10, css-sizing-3)

| Feature | Status | Notes |
|---------|--------|-------|
| Vertical flow, auto width/height | yes | unset = auto, `Px(0)` = zero |
| Margin collapsing: siblings, parent/first and last child, empty blocks, negative margins | yes | §8.3.1; formatting-context roots do not collapse through |
| `min-*`/`max-*` | yes | max-first, min wins |
| `box-sizing` content-box / border-box | yes | |
| `aspect-ratio` | yes | only a definite size transfers through the ratio |
| Writing modes | yes | |
| `display: contents`, `flow-root`, `inline-block`, `inline-flex`, `inline-grid` | no | |
| Floats, tables, multi-column | no | |

## CSS Positioned Layout Level 3

| Feature | Status | Notes |
|---------|--------|-------|
| `position` static, relative, absolute, fixed | yes | containing block = nearest positioned ancestor's padding box |
| `position: sticky` | partial | behaves as relative; no scroll offset |
| `top`/`right`/`bottom`/`left` | yes | unset = auto, `Px(0)` = 0, negatives allowed; over-constrained boxes per §10.3.7/§10.6.4 |
| `z-index` | meta | stored, not used for ordering |
| Static position for auto offsets | yes | |

## CSS Text Level 3 and Writing Modes Level 3

See [Text](text.md) for the per-property table. Summary:

| Feature | Status |
|---------|--------|
| `white-space` (all five values), `text-align`, `text-align-last`, `text-justify`, `text-indent`, `letter-spacing`, `word-spacing`, `tab-size`, `overflow-wrap`, `word-break`, `text-overflow`, `text-transform`, `hanging-punctuation` | yes |
| `line-height` | partial (`<10` multiplier heuristic) |
| `hyphens` | partial (soft hyphens only) |
| `direction: rtl` | yes (alignment and indent); UAX #9 reordering: no |
| `writing-mode` vertical-rl, vertical-lr, sideways-rl, sideways-lr | yes (not inherited) |
| Text decorations, `vertical-align`, `font-style`, `font-weight` | meta (stored for renderers) |
| Inline formatting context (spans, inline-block, mixed content) | no |
| UAX #14 | yes (LB4-LB31, no tailoring) |

## CSS Values and Units Level 4 / Containment Level 3

| Feature | Status | Notes |
|---------|--------|-------|
| Absolute units, `em`, `rem`, `ch`, `vw`, `vh`, `vmin`, `vmax` | yes | |
| `ex`, `cap`, `lh`, `rlh`, `ic` | partial | approximated from the font size |
| `vi`, `vb`, `sv*`, `lv*`, `dv*` | no | resolve to 0 |
| `cqw`, `cqh`, `cqi`, `cqb`, `cqmin`, `cqmax` | yes | via `ResolveLengthInContext`; 0 in the plain layout pass |
| `container-type`, `container-name`, `container` shorthand | yes | property model and parsers only; `@container` rules: no |
| Percentages | no | there is no `%` unit; use `fr` tracks, `FlexGrow`, or viewport/container units |

## CSS Sizing Level 3

| Feature | Status | Notes |
|---------|--------|-------|
| `min-content`, `max-content`, `fit-content()` on width, height, and grid tracks | yes | `WidthSizing`/`HeightSizing`; deprecated float sentinels still honored |
| Contributions include padding, border, margins, text | yes | |
| `contain-intrinsic-size` (Level 4) | no | |

## CSS Transforms Level 1

| Feature | Status | Notes |
|---------|--------|-------|
| translate, scale, rotate, skew, matrix, composition | yes | layout ignores transforms; `GetFinalRect` and `ToSVGString` expose them to renderers; the zero value is the identity |

## CSS Box Alignment Level 3 (post-layout helpers)

`AlignNodes`, `DistributeNodes`, `SnapNodes`, `SnapToGrid` are design-tool operations on `Rect` after layout, not implementations of the alignment properties (those are handled by flex and grid).
