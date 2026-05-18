# Changelog

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
