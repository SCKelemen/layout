# Text

Text layout breaks a node's `Text` into line boxes according to [CSS Text Module Level 3](https://www.w3.org/TR/css-text-3/), [CSS Writing Modes Level 3](https://www.w3.org/TR/css-writing-modes-3/), and [UAX #14](https://www.unicode.org/reports/tr14/). The engine positions text; it does not draw it. Your renderer reads `node.TextLayout` and `node.Style.TextStyle`.

## Creating text nodes

```go
node := layout.Text("Hello, world", layout.Style{
	Width: layout.Px(200),
	TextStyle: &layout.TextStyle{
		FontSize:   16,
		LineHeight: 1.5,
		TextAlign:  layout.TextAlignCenter,
	},
})
size := layout.LayoutSimple(node, layout.Loose(400, 400))
_ = size
```

`Text` sets `Display: DisplayInlineText`, leaves `Width`/`Height` unset (auto), and supplies a default `TextStyle` (16px, `normal` line height, `white-space: normal`, LTR) when the style has none. A text node is a leaf; children are ignored. `Node.WithText` only replaces the string and does not make a node a text node.

`TextStyle.FontSize == 0` falls back to the root font size (`ctx.RootFontSize`, or 16). Text nodes work as block children, flex items (measured, then re-wrapped at their flexed size), and grid items (their min-content contribution is the track base size).

## Properties

All properties live on `TextStyle` unless noted. Status: **yes** = implemented in layout; **meta** = stored and serialized for renderers but does not change geometry.

### Font

| Property | Field | Status | Notes |
|----------|-------|--------|-------|
| `font-size` | `FontSize` (px) | yes | 0 falls back to the root font size |
| `font-family` | `FontFamily` | meta | passed to `TextMetricsProvider.Measure` |
| `font-weight` | `FontWeight` (`FontWeightNormal`=400, `FontWeightBold`=700, any int) | meta | passed to the provider |
| `font-style` | `FontStyle` (`Normal`, `Italic`, `Oblique`) | meta | |

### Alignment and justification

| Property | Field | Status | Notes |
|----------|-------|--------|-------|
| `text-align` | `TextAlign` (`Default`=start, `Left`, `Right`, `Center`, `Justify`) | yes | `start` resolves to the right edge in RTL |
| `text-align-last` | `TextAlignLast` (`Auto`, `Left`, `Right`, `Center`, `Justify`) | yes | applies to the last line and to lines ending in a forced break |
| `text-justify` | `TextJustify` (`Auto`, `InterWord`, `InterCharacter`, `Distribute`, `None`) | yes | `Auto` = inter-word; adjustments are exposed per line |
| `vertical-align` | `VerticalAlign` (`Baseline`, `Sub`, `Super`, `TextTop`, `TextBottom`, `Middle`, `Top`, `Bottom`) | meta | there is no inline formatting context to align within |

### Spacing

| Property | Field | Status | Notes |
|----------|-------|--------|-------|
| `line-height` | `LineHeight` | yes | `<= 0` normal (1.2 x font size); `0 < x < 10` multiplier; `>= 10` absolute px |
| `letter-spacing` | `LetterSpacing` (px) | yes | `-1` means normal; measured by the provider |
| `word-spacing` | `WordSpacing` (px) | yes | `-1` means normal; added to each inter-word space |
| `text-indent` | `TextIndent` (px) | yes | first line only; a margin on the start edge, may be negative |
| `tab-size` | `TabSize` | yes | `<= 0` means 8; preserved tabs advance to tab stops measured from the line start |

### Wrapping and breaking

| Property | Field | Status | Notes |
|----------|-------|--------|-------|
| `white-space` | `WhiteSpace` (`Normal`, `Nowrap`, `Pre`, `PreWrap`, `PreLine`) | yes | preserved spaces hang at line end in `pre-wrap` |
| `overflow-wrap` | `OverflowWrap` (`Normal`, `BreakWord`, `Anywhere`) | yes | `Anywhere` also affects min-content size |
| `word-break` | `WordBreak` (`Normal`, `BreakAll`, `KeepAll`) | yes | |
| `text-overflow` | `TextOverflow` (`Clip`, `Ellipsis`) | yes | with `nowrap`, the last box becomes `"..."` |
| `hyphens` | `Hyphens` (`None`, `Manual`, `Auto`) | partial | soft hyphens (U+00AD) are invisible and become `-` at a break; `None` disables them; `Auto` behaves as `Manual` (no dictionaries) |
| `line-break` | - | no | |

### Transformation, punctuation, decoration

| Property | Field | Status | Notes |
|----------|-------|--------|-------|
| `text-transform` | `TextTransform` (`None`, `Uppercase`, `Lowercase`, `Capitalize`, `FullWidth`, `FullSizeKana`) | yes | applied before measuring; `InlineBox.Text` holds the transformed text |
| `hanging-punctuation` | `HangingPunctuation` (`None`, `First`, `Last`, `ForceEnd`, `AllowEnd`) | yes | `allow-end` does not hang opening punctuation |
| `text-decoration-line` | `TextDecoration` bitmask (`Underline`, `Overline`, `LineThrough`; `Has`) | meta | |
| `text-decoration-style` | `TextDecorationStyle` (`Solid`, `Double`, `Dotted`, `Dashed`, `Wavy`) | meta | |
| `text-decoration-color` | `TextDecorationColor` (CSS color string, `""` = currentColor) | meta | |
| `text-shadow`, `text-emphasis`, `font-variant`, `font-stretch` | - | no | renderer concerns |

### Direction and writing modes

| Property | Field | Status | Notes |
|----------|-------|--------|-------|
| `direction` | `Style.Direction`, falling back to `TextStyle.Direction` (`LTR`, `RTL`) | yes | alignment, `text-indent`, and over-constrained absolute boxes follow it |
| `unicode-bidi` / UAX #9 | - | no | no bidi reordering; runs stay in logical order |
| `writing-mode` | `Style.WritingMode` (preferred), `TextStyle.WritingMode` (legacy) | yes | `horizontal-tb`, `vertical-rl`, `vertical-lr`, `sideways-rl`, `sideways-lr`; not propagated to children |
| `text-orientation` | - | partial | orientation is derived automatically: `InlineBox.Orientations` marks each rune upright (UAX #50) or rotated; sideways modes rotate everything; there is no property to force `upright` or `sideways` |
| `text-combine-upright` | - | no | |

`Orientations` is computed from `TextStyle.WritingMode`. Nodes that set only `Style.WritingMode` get vertical geometry but a nil `Orientations` slice; set both fields when your renderer needs orientations.

### Unicode algorithms

| Algorithm | Status |
|-----------|--------|
| UAX #14 line breaking | Rules LB4-LB31 implemented in `uax14.go` with SP/CM state tracking: mandatory breaks, no break before closing punctuation or after opening punctuation, glue (`WJ`, `GL`, U+00A0), U+200B, numeric rules, Hangul jamo and syllables (LB26/LB27), CJK ideographs as `ID`, small kana as `NS`. No tailoring, and no emoji-modifier or regional-indicator rules. Dictionary-based breaking for Thai, Lao, and Khmer is not implemented. |
| UAX #29 grapheme clusters | Provided by `github.com/SCKelemen/text` when you install `NewTerminalTextMetrics()`; the built-in approximate provider counts runes |
| UAX #50 vertical orientation | `InlineBox.Orientations` (see above) |
| UAX #9 bidi | Not implemented |

## Line boxes

`LayoutText` fills `node.TextLayout`:

```go
type TextLayout struct {
	Lines      []TextLine
	LineHeight float64 // resolved line height in px
}

type TextLine struct {
	Boxes               []InlineBox
	Width               float64 // sum of box widths and inter-word spaces
	SpaceCount          int     // inter-word spaces (for justify)
	SpaceWidth          float64 // total width of those spaces
	SpaceAdjustment     float64 // extra px per space when justified
	CharacterAdjustment float64 // extra px between characters (inter-character justify)
	OffsetX             float64 // inline offset from text-align / text-indent
	OffsetY             float64 // block position of the line
	EndsWithForcedBreak bool    // true for the last line and for lines ending at a preserved newline
}

type InlineBox struct {
	Kind         InlineBoxKind // InlineBoxText; inline nodes are not implemented
	Text         string        // the (transformed) run, including a trailing "-" added at a soft-hyphen break
	Node         *Node         // reserved for inline nodes; always nil today
	Width        float64
	Ascent       float64
	Descent      float64
	Orientations []bool // per rune, vertical modes only: true upright, false rotated
	SpaceAfter   bool   // an inter-word space follows this box on the same line
}
```

In horizontal modes `OffsetX` is the x position of the first box within the content box and `OffsetY` the top of the line; advance through `Boxes`, adding the space width plus `SpaceAdjustment` after every box with `SpaceAfter`. In vertical modes the roles swap: lines stack along the width (from the right in `vertical-rl`) and boxes advance down the height. `Width` on a vertical line is still the inline extent.

A renderer loop:

```go
layout.LayoutSimple(node, layout.Loose(400, 400))
spaceWidth, _, _ := layout.NewTerminalTextMetrics().Measure(" ", *node.Style.TextStyle)
for _, line := range node.TextLayout.Lines {
	x := node.Rect.X + line.OffsetX
	y := node.Rect.Y + line.OffsetY
	for _, box := range line.Boxes {
		drawText(x, y+box.Ascent, box.Text)
		x += box.Width
		if box.SpaceAfter {
			x += spaceWidth + line.SpaceAdjustment
		}
	}
}
```

(`drawText` is your rendering primitive; measure the space with the same provider you laid out with.)

## Text metrics

Measurement is pluggable through `TextMetricsProvider`:

```go
type TextMetricsProvider interface {
	Measure(text string, style TextStyle) (advance, ascent, descent float64)
}
```

The default provider approximates every rune as 0.6 x `FontSize`. `NewTerminalTextMetrics()` wraps `github.com/SCKelemen/text` for Unicode-accurate widths (East Asian width, emoji sequences, grapheme clusters) in terminal cells; `NewTextMetricsAdapter(text.Config{...})` accepts a custom configuration.

Install a provider globally or per layout context:

```go
metrics := layout.NewTerminalTextMetrics()

// Globally (safe to call concurrently; typically once at startup).
layout.SetTextMetricsProvider(metrics)

// Per context. Every text measurement in this layout pass, including
// ch-unit resolution, line breaking, ellipsis, and intrinsic sizing, uses it.
ctx := layout.NewLayoutContext(80, 24, 1).WithTextMetrics(metrics)
node := layout.Text("Hello 世界", layout.Style{TextStyle: &layout.TextStyle{FontSize: 1}})
size := layout.Layout(node, layout.Loose(80, 24), ctx)
fmt.Printf("%.0f cells x %.1f rows\n", size.Width, size.Height) // 10 cells x 1.2 rows
```

With a terminal provider, use `FontSize: 1` so one cell is one unit; the direct text operations (`metrics.Text().Width(...)`, `.Wrap(...)`) are available for rendering. `examples/text_integration` is a complete program.

## Intrinsic sizes

`MinContentWidth`/`MaxContentWidth` (and `Style.WidthSizing`) size a text node to its longest unbreakable segment or its longest unwrapped line, applying the same preprocessing as layout (tabs, white-space, `text-transform`, soft hyphens removed). See [Layout systems](layout-systems.md#intrinsic-sizing).

## Known gaps

Line-height heuristic, `-1` spacing sentinels, no inline formatting context, no bidi, `hyphens: auto` without dictionaries, orientation tied to `TextStyle.WritingMode`: all listed in [Limitations](limitations.md).
