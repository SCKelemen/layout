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
