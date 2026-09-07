# Changelog

## [Unreleased]

### Fixed

- **Block layout: unset `Width`/`Height` is `auto` (behavior change).** A zero-value `Length` (`Unit == ""`, i.e. the field was never assigned) on a block container now means `auto`, matching the flexbox and grid conventions and the CSS initial value (CSS 2.1 §10.3.3). An explicit `Px(0)` is still a real zero size. Previously an unset width was treated as `0px`, so a `Style{}` root collapsed every descendant to zero width.
- **Positioning offsets: unset means `auto`, `Px(0)` means 0.** `Style.Top/Right/Bottom/Left` are `auto` only when never set (`Unit == ""`); any constructed value, including `Px(0)` and negative values, is a real offset (CSS 2.1 §9.3.2, css-position-3 §3.1). This replaces the "0 is auto if the opposite side is set" heuristic that pinned `{Left: Px(5)}` boxes to `bottom: 0`, and lets relative positioning honor negative offsets (`Left: Px(-10)` now moves the box left). Both offsets of an axis `auto` yields the static position.
- **Absolute containing block is the parent's padding box in parent-local coordinates** (CSS 2.1 §10.1): `layoutPositionedRecursive` passes `Rect{X: borderLeft, Y: borderTop, W: parent.W - borders, H: parent.H - borders}` instead of the parent's own (grandparent-relative) `Rect`, so nested containers with padding/borders no longer shift absolute children.
- **Absolutely positioned children take no block-flow space** (CSS 2.1 §9.3): they are still laid out for their own size and static position, but no longer advance the block position, contribute to the container's auto height, or participate in margin collapsing.
- **Sticky positioning uses resolved lengths** (`Top: Em(2)` moves 32px, not 2px).
- **`max-height` applies to auto-height blocks**, and min/max are applied max-first so `min` wins when `min > max` (CSS 2.1 §10.4, §10.7).
- **Border-box sizes smaller than padding + border** yield a 0 content box instead of a negative value that was misread as `auto` (`convertToContentSize`).
- **`HeightSizing` min-/max-/fit-content on block containers** falls back to auto sizing instead of storing the `-1` sentinel as a height (`-1 + padding`).
- **Margin collapsing with negative margins** follows CSS 2.1 §8.3.1: the collapsed margin is the largest positive margin plus the most negative one. **Empty blocks** (zero block size, no block-axis padding/border) now collapse their own start and end margins together with adjoining siblings.
- **Vertical writing modes:** children receive the container's inline size (physical height) as their inline constraint, and vertical-rl/sideways-rl children are positioned against the container's resolved width rather than the available width.
- **Intrinsic sizing contributions** resolve child widths and margins through `ResolveLength` (em/rem/vw now count), add padding and border for content-box children, count flex gaps only between visible (non-`display: none`) items, and measure text children (longest word for min-content, whole run for max-content) instead of contributing 0.
- **`ResolveLength` returns 0 for container-relative units** (`cq*`) since it has no ancestor information, consistent with `ResolveLengthInContext` when no container size is available; previously the raw value leaked through as pixels. Query containers now measure their content box (padding and border excluded) per css-contain-3 §5.
- `getCurrentFontSize` tolerates a nil `LayoutContext` (falls back to 16px), and an auto-width block with unbounded available width and no children sizes to 0 instead of `math.MaxFloat64`.
- **Flexbox: flexible lengths are resolved per CSS Flexbox §9.7.** Shrinking uses scaled shrink factors (`flex-shrink × flex base size`), grow and shrink respect `min-*`/`max-*` on the main axis, and violators are frozen and the free space redistributed (loop bounded by the item count). Items that shrank to 0 are no longer reset to their explicit width. `space-around`/`space-evenly` now put space between items (previously only a leading offset), and fall back to `center` (or `flex-start` for `space-between`) when free space is negative (§9.5).
- **Flexbox wrapping:** lines in a container with an indefinite cross size stack from the cross-start instead of overlapping at offset 0; the main-axis gap counts toward line length when breaking lines (§9.3 step 5); `wrap-reverse` no longer inverts `align-content: flex-start/flex-end` twice (§5.2, §8.4); `align-content: center/flex-end` allow negative offsets when lines overflow.
- **Flexbox gaps (behavior change):** `FlexRowGap` is the gap between rows and `FlexColumnGap` the gap between columns per css-align-3 §8.3, so in a `column` container the gap between items is the row gap and the gap between lines is the column gap. Previously the mapping was fixed to main/cross regardless of direction.
- **Flexbox `align-items: stretch`** only applies to items whose cross size is `auto` (§9.4 step 11); an explicit `Height` (row) or `Width` (column) is kept.
- **Flexbox indefinite main size:** with unbounded constraints the line's content size is used for justification and for the container size, instead of `math.MaxFloat64` leaking into positions (`justify-content: center` produced X ≈ 9e307). Empty flex containers honor their explicit width/height (and fill a bounded available width like any block-level flex container). Nested containers (flex, grid, block) are re-laid out whenever their final size differs from the size they were measured at (previously only growing flex children), and non-stretched items are measured with an indefinite cross size so a nested container no longer stretches to the parent's cross size. Baseline alignment in `column` direction behaves as `flex-start` (§8.3).
- **Grid auto-placement follows CSS Grid §8.5:** definite items are placed first, row/column-locked items go to the first free cell in their line, and remaining items use an occupancy-aware cursor (spans included) for both `row` and `column` flow. Previously placement was index-based and could overlap explicitly placed items. Dense flow no longer panics when the implicit grid must grow past the last row, and uses the same "auto" predicate as sparse flow. `start == end` lines resolve to a span of 1 and `end < start` swaps (§8.3.1). The implicit grid is capped at 10,000 lines.
- **Grid track sizing:** a zero-value `GridAutoRows`/`GridAutoColumns` (and zero-value template entries) means `auto` (§7.6) instead of `minmax(0, 0)`; `auto` tracks are sized to their items' max-content contributions (§11.5, §12.5) instead of 0; `fr` tracks respect their `auto` minimum so `1fr` behaves as `minmax(auto, 1fr)` (§7.2.4, §12.7.1); a spanning item's extra space is distributed only to intrinsic tracks after subtracting gaps and fixed tracks (§12.5.1); `align-content: stretch` grows only `auto`-max tracks and otherwise behaves as `start` (§12.7); `minmax(fixed, fixed)` tracks grow toward their max with definite free space (§12.6); empty grids include row gaps in their height.
- **Grid container sizing:** an unset (`Unit == ""`) container `Width` fills a bounded constraint and an unset `Height` is indefinite unless the constraint is tight, matching block and flex; previously both resolved to `0px`. Columns are sized against the physical height and rows against the width in vertical writing modes. The style's `GridTemplateRows/Columns` slices are no longer mutated by implicit-track growth.
- **Grid auto-repeat:** `calculateAutoRepeatCount` returns 1 for an indefinite (or NaN/Inf) size instead of `MaxInt` (which made `expandAutoRepeatTracks` loop ~9e18 times), uses the max sizing function when it is definite (§7.2.3.2), and caps the repetition count.
- **Text line breaking:** an explicit `Width` (or `Height` in vertical modes) is used as the breaking width, clamped by min/max (previously lines broke at the available width and the box was resized afterward); a trailing space no longer counts toward the fit test (css-text-3 §4.1.3); `text-indent` participates in the fit test for every word on the first line; `pre-wrap` preserved spaces hang at line end instead of producing space-only lines; `text-overflow: ellipsis` accounts for inter-word spaces and keeps `SpaceCount` for retained gaps; lines ending in a forced break are aligned with `text-align-last` when justifying (§7.1), and `pre-line` segments track spaces so they can be justified; `hanging-punctuation: allow-end` no longer hangs opening punctuation (§9.2); U+00A0 is no longer trimmed or split on; `breakWordToFit` measures the accumulated piece so letter-spacing is honored; a `LineHeight <= 0` in the metrics adapter means `normal`; `TabSize <= 0` means 8.
- **Text in vertical writing modes:** the returned `Size` swaps axes consistently with line positioning (`vertical-rl` with 6 lines of 20px yields `Width: 120`), and lines start from the resolved block size rather than the inline size.
- **UAX #14 line breaking:** the pair-table wildcard entries (`ClassXX`) were dead code, so almost every pair fell through to "break allowed". The rule engine now implements LB4–LB31 in order with SP/CM state tracking: no break before closing punctuation, exclamation, or combining marks after ideographs; no break after opening punctuation; `WJ`/`GL`/NBSP glue; U+200B breaks; `HY × NU` keeps `1-2`; Hangul and numeric rules. Hard hyphens are always break opportunities (LB21) regardless of `Hyphens`; only U+00AD is gated. Hiragana/Katakana/Han are `ID` (small kana and `ー` are `NS`); CJK punctuation, fullwidth brackets, and curly quotes are classified per the standard.
- **Transform (behavior change):** the zero-value `Transform{}` is treated as the identity by `IsIdentity`, `ToSVGString`, `ApplyToRect`, and `Multiply`, so nodes that never set a transform no longer render as `matrix(0,0,0,0,0,0)`. Because `Scale(0, 0)` is the same value as `Transform{}`, it is also treated as the identity. The `Multiply` doc now matches its behavior (`t2` is applied first, then `t1`).
- **Fluent API:** `Node.Transform` and `Node.Map` recurse over the node returned by the callback (child-list edits are kept) and drop a node when the callback returns nil instead of panicking. `CloneDeep` deep-copies `GridTemplateRows/Columns`, `GridTemplateAreas`, `ContainerName`, `TextStyle`, and `TextLayout`; `Clone` and `WithStyle` document that they share those.
- **serialize (format change):** `Length` values are written as CSS strings (`"10px"`, `"2em"`, `"unbounded"`) so units survive a round trip; bare JSON/YAML numbers are still accepted as pixels. Added `Node.Text`, `Node.Baseline`, `AlignSelf`, `Order`, `GridAutoFlow`, `GridTemplateAreas`, `GridArea`, `JustifySelf`, `WidthSizing`/`HeightSizing`/`FitContent*`, `WritingMode`, `ContainerType`/`ContainerName`, `TextStyle`, and the `DisplayInlineText`/`DisplayNone` keywords. Struct-valued fields are pointers so empty `padding`/`margin`/`rect` objects are omitted. Decoding rejects NaN/Inf and out-of-range numbers, unknown keywords, and trees deeper than 1024 levels or wider than 65,536 children. YAML keys are now camelCase. `ToJSON`/`ToYAML` now return an error for cyclic trees or trees over the depth/child limits.
- **Docs and examples:** README, `doc.go`, `api.go`, and the serialize/fluent READMEs use the current API (`Px(...)`, `Layout(root, constraints, ctx)` / `LayoutSimple`); `examples/serialize` no longer panics on an empty grid; the auto-fill/auto-fit helper docs compile and state that they are not yet wired into `LayoutGrid`.

- **Grid `stretch` now respects definite item sizes (behavior change).** When `align-items`/`justify-items` (or the `*-self` equivalents) resolve to `stretch`, a grid item with a definite (explicit) `width`/`height` is no longer stretched to fill its track — it keeps its explicit, box-sizing-aware size and is positioned at the start of its area. Stretch continues to size auto items to fill the track. This matches CSS Box Alignment Level 3 §6.2, where `stretch` is a no-op on an axis whose size is definite (https://www.w3.org/TR/css-align-3/#stretch-alignment). Previously `LayoutGrid` overwrote the item size with the track size unconditionally on stretch.

## [v1.3.0] - 2026-05-20

### Changed

- `github.com/SCKelemen/text` bumped from `v1.1.3` to `v1.2.0` (`unicode/v6` migration).
- `github.com/SCKelemen/unicode` replaced with `github.com/SCKelemen/unicode/v6` (`v6.2.0`). Brings v6 performance improvements (ASCII fast paths, memory optimization, rule-based state machines) to layout's text measurement and line breaking.
- `github.com/SCKelemen/units` bumped from `v1.2.0` to `v1.2.1`.

### Note

Pure dependency migration. No source-level API changes.

## [v1.2.1] - 2025-07-11

Patch release fixing one MEDIUM bug in `ResolveLengthInContext` caught by an external bug-hunt sweep.

### Fixed

- `ResolveLengthInContext` now returns `0` for container-relative units (`cqw`, `cqh`, `cqi`, `cqb`, `cqmin`, `cqmax`) when no container size is available, instead of returning the raw `l.Value` as a `float64`. Previously, `Cqw(50)` with a zero-size container returned `50.0` — neither pixels nor a meaningful percentage. Non-container units (absolute, viewport, font-relative) are unaffected.

### Tests

- New `TestResolveLengthInContextCqZeroContainerReturnsZero` guards the fix across `cqw`, `cqh`, and `cqmin`.

## [1.2.0] - 2026-05-18

### Added
- Full CSS Values Level 4 length-unit coverage via integration with `github.com/SCKelemen/units` v1.2.0. All 44 L4 length units (lh, cap, ic, vi, vb, sv*, lv*, dv*, cqw, cqh, cqi, cqb, cqmin, cqmax, plus the existing absolute/em/rem/ch/vh/vw/vmin/vmax set) now resolvable through `ResolveLength`.
- Container query support (CSS Containment Module Level 3):
  - `Style.ContainerType` property (`normal` / `size` / `inline-size`) with `ParseContainerType`.
  - `Style.ContainerName` property with `ParseContainerName`.
  - `ParseContainer` shorthand parser.
  - `ResolveLengthInContext` — ancestor-walking resolver that honors the container's `WritingMode` for `cqi`/`cqb` axis mapping.
- `Cqw`, `Cqh`, `Cqi`, `Cqb`, `Cqmin`, `Cqmax` length constructors.
- `Length` now inherits the full method set from `units.Length`: `Add`, `Sub`, `Mul`, `Div`, `IsAbsolute`, `IsFontRelative`, `IsViewportRelative`, `IsContainerRelative`, `LessThan`, `GreaterThan`, `Raw`.

### Changed
- **`layout.Length` and `layout.LengthUnit` are now type aliases for `units.Length` and `units.LengthUnit`.** Existing named constants (`Pixels`, `EmUnit`, `VwUnit`, etc.) and constructors (`Px`, `Em`, etc.) are unchanged.
- **`LengthUnit`'s underlying type changed from `int` to `string`** to match the CSS spec representation. Code that uses the named constants or constructors is unaffected. Code that casts `LengthUnit` to/from `int` or compares with integer literals will break; switch to the named constants.
- `UnboundedUnit`'s underlying value changed from the integer `14` to the string `"unbounded"`. Code using the constant by name is unaffected.
- `ResolveLength`'s internals now delegate to `units.Length.Resolve`. ~130 lines of duplicated unit math removed. Behavior is unchanged for all previously-supported units.
- `github.com/SCKelemen/units` promoted from indirect to direct dependency at `v1.2.0`.

### Fixed
- Documentation reconciled with actual implementation (`docs/limitations.md`, `docs/CSS_VALUES_STATUS.md`).
- Removed mid-flight debug block and stale `test_user/` and `debug/` directories.

### CI
- `actions/checkout` v4 → v6.
- `actions/setup-go` v5 → v6.
- `codecov/codecov-action` v4 → v6.
