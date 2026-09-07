# Limitations

The single list of known gaps and inexpressible values. Anything not listed here and described in [Spec compliance](spec-compliance.md) as implemented is expected to work; please open an issue if it does not. Planned fixes for the API-shape items are collected in [docs/design/v2-api.md](design/v2-api.md).

## Not implemented

| Gap | Detail | Workaround |
|-----|--------|------------|
| Inline formatting context | Text nodes are leaves. There is no `inline-block`, no spans, no inline images, no mixing of text runs with different styles on one line, and `vertical-align` does not move anything. `InlineBoxInlineNode` is reserved but unused. | One text node per run; stack runs with flex or block |
| `display: contents`, `flow-root`, `inline-flex`, `inline-grid`, `inline-block` | `Display` has block, flex, grid, inline-text, none | |
| Tables | | Use grid |
| CSS multi-column (`column-count`, `column-width`) | Flowing text into columns | Split text yourself into grid columns |
| Floats | | |
| Grid subgrid (Level 2) | | Nest grids with matching tracks |
| UAX #9 bidirectional algorithm | `direction: rtl` swaps alignment, `text-indent`, and offset resolution; runs are not reordered and mixed-direction text is laid out in logical order | Reorder runs before layout |
| `hyphens: auto` dictionaries | `Auto` behaves as `Manual`: only U+00AD soft hyphens create hyphenation points | Insert soft hyphens |
| `line-break`, `text-orientation`, `text-combine-upright` | Orientation in vertical modes is decided by UAX #50 only (`InlineBox.Orientations`); it cannot be forced upright or sideways | |
| `unicode-bidi`, `text-shadow`, `text-emphasis`, `font-variant`, `font-stretch` | Not in `TextStyle` | Keep them in your renderer's model |
| `contain-intrinsic-size` (css-sizing-4) | | |
| `@container` rules | Only the property model (`ContainerType`, `ContainerName`) and unit resolution exist | Evaluate queries in your code using `Rect` |
| Percentages | `Length` has no `%` unit | `fr` tracks, `FlexGrow`, `vw`/`vh`, or `cq*` |
| `vi`, `vb`, `sv*`, `lv*`, `dv*` units | Resolve to 0; `LayoutContext` only has one viewport size | `vw`/`vh`/`vmin`/`vmax` |

## Partial behavior

| Area | Behavior |
|------|----------|
| Sticky positioning | Behaves as `relative`. Sticking needs a scroll offset, which the engine does not have. |
| `z-index` | Stored on `Style`; layout does not sort or stack by it. |
| Transforms | Do not affect layout; `GetFinalRect` / `Transform.ApplyToRect` give renderers the transformed box. |
| Text decorations, `font-style`, `font-weight`, `vertical-align` | Metadata for renderers; only `FontSize`, `LetterSpacing`, `WordSpacing`, and `FontFamily`/`FontWeight` (through your `TextMetricsProvider`) influence measurement. |
| `writing-mode` | Not inherited. Set `Style.WritingMode` on every node that needs a vertical mode; `InlineBox.Orientations` is additionally driven by `TextStyle.WritingMode`. |
| Container-query units in layout | `Layout`/`LayoutSimple` resolve `cq*` to 0 (they have no ancestor information). Resolve them yourself with `ResolveLengthInContext` and a `NodeContext`, or pre-resolve them into `Px` before layout. |
| `ex`, `cap`, `lh`, `rlh`, `ic` | Approximated (font size, root font size, 2 x `ch`) because terminal metrics have no x-height or cap-height. |
| Tab stops | Measured from the start of the line box, not from the start of the block, so indented lines see shifted stops. |
| UAX #14 | No tailoring, no emoji-modifier or regional-indicator rules, no dictionary-based breaking (Thai, Lao, Khmer). |
| Grapheme clusters | Only with a `github.com/SCKelemen/text` provider (`NewTerminalTextMetrics`); the default provider counts runes. |
| Flex intrinsic sizes | Simplified §9.9: sum of items for rows, largest item for columns. |

## Inexpressible values (API shape)

These follow from Go zero values standing in for "unset":

| Value | Why | Effect |
|-------|-----|--------|
| `flex-shrink: 0` | `FlexShrink == 0` is read as the initial value 1 | Items always shrink; give them a `MinWidth`/`MinHeight` instead |
| `align-self: stretch` overriding a non-stretch parent | `AlignSelf == 0` (`AlignItemsStretch`) means "use the parent's `AlignItems`" | Only non-stretch overrides are possible; the same applies to `JustifySelf` |
| `letter-spacing` / `word-spacing: normal` vs `0` | `-1` is the "normal" sentinel and `0` adds zero spacing; the two coincide in practice but the sentinel leaks into `TextStyle` values and serialization | Treat `-1` and `0` as equivalent |
| `line-height: 12` as a multiplier | Values `< 10` are multipliers, `>= 10` are pixels | Use `LineHeight: 12 * fontSize` or a multiplier below 10 |
| `tab-size` 0 | `TabSize <= 0` means 8 | |
| `grid-row-start: 0 / auto` vs unset | Line 0 with no end is auto-placed | Set `GridRowEnd: 1` (or `-1` for auto explicitly) |
| `Frame(node, 0, 0)` | `Frame` skips values `<= 0` | `FrameLength(node, Px(0), Px(0))`, and `Length{}` to reset to auto |

## Concurrency

`SetTextMetricsProvider` stores the package-level provider in an atomic pointer and is safe to call concurrently with layout. A single tree, however, must not be laid out by two goroutines at once: layout writes `Rect` and `TextLayout` in place.
