# Changelog

## [Unreleased]

### Fixed

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
