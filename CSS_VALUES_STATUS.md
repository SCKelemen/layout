# CSS Values and Units Implementation Status

This document tracks the implementation status of length units from the CSS Values and Units Module Level 4 specification.

**Specification Reference**: https://www.w3.org/TR/css-values-4/

## Implemented Length Units

### Absolute Length Units ✅
All absolute length units are fully implemented based on the CSS reference pixel (1in = 96px):

- **`px`** - Pixels (CSS reference pixel)
- **`pt`** - Points (1pt = 1/72 inch = 1.333px)
- **`pc`** - Picas (1pc = 12pt = 16px)
- **`in`** - Inches (1in = 96px)
- **`cm`** - Centimeters (1cm = 96/2.54px ≈ 37.795px)
- **`mm`** - Millimeters (1mm = 96/25.4px ≈ 3.7795px)
- **`Q`** - Quarter-millimeters (1Q = 96/101.6px ≈ 0.945px)

**Example Usage**:
```go
node := &Node{
    Style: Style{
        Width:   Pt(72),  // 1 inch in points
        Height:  Cm(2.54), // 1 inch in centimeters
        Padding: Uniform(Mm(5)), // 5mm padding
    },
}
```

### Relative Font Units ✅ (Partial)
Currently implemented:

- **`em`** - Relative to element's font size
- **`rem`** - Relative to root element's font size
- **`ch`** - Relative to width of '0' character (with configurable reference character)

**Example Usage**:
```go
ctx := NewLayoutContext(1920, 1080, 16) // Root font size: 16pt

node := &Node{
    Style: Style{
        Width:   Em(20),   // 20 × element font size
        Padding: Rem(2),   // 2 × root font size = 32px
        Margin:  Spacing{Left: Ch(4)}, // 4 × '0' character width
        TextStyle: &TextStyle{
            FontSize: 14, // This element's font size
        },
    },
}
```

### Viewport Units ✅
All basic viewport units are implemented:

- **`vh`** - 1% of viewport height
- **`vw`** - 1% of viewport width
- **`vmin`** - 1% of smaller viewport dimension
- **`vmax`** - 1% of larger viewport dimension

**Example Usage**:
```go
node := &Node{
    Style: Style{
        Width:   Vw(100),  // Full viewport width
        Height:  Vh(50),   // Half viewport height
        Padding: Vmin(5),  // 5% of smaller dimension
    },
}
```

### Special Units ✅
- **`UnboundedUnit`** - Represents infinity for maximum sizes

### Level 4 Font-Relative Aliases ✅
The following Level 4 font-relative units are implemented as terminal-context
aliases. See `length_l4.go` for the resolution rules and rationale:

| Unit | Terminal-context meaning |
| --- | --- |
| `lh` | Line height of the element — resolves to one row (current font size) |
| `rlh` | Root line height — resolves to one root row |
| `ic` | CJK ideograph advance — resolves to two character cells |
| `ric` | Root `ic` — two root character cells |
| `cap` | Cap height — one row (collapses with `ex` in monospace terminals) |
| `rcap` | Root cap height — one root row |
| `rch` | Root `ch` — width of the reference character at root font size |
| `rex` | Root `ex` — one root row |

Spec: https://www.w3.org/TR/css-values-4/#font-relative-lengths

### Level 4 Logical Viewport Units ✅
| Unit | Terminal-context meaning |
| --- | --- |
| `vi` | 1% of the viewport inline size; aliased to `vw` for horizontal-tb (simplification) |
| `vb` | 1% of the viewport block size; aliased to `vh` for horizontal-tb (simplification) |

When the resolver gains a writing-mode parameter, `vi`/`vb` should swap for
vertical writing modes. The current behavior matches the default
horizontal-tb mode.

### Level 4 Small / Large / Dynamic Viewport Variants ✅
Terminals have no UI chrome that can hide or reveal, so the
`sv*`, `lv*`, and `dv*` families are exact aliases for their base
viewport counterparts (`vw`, `vh`, `vi`, `vb`, `vmin`, `vmax`):

- Small: `svw`, `svh`, `svi`, `svb`, `svmin`, `svmax`
- Large: `lvw`, `lvh`, `lvi`, `lvb`, `lvmin`, `lvmax`
- Dynamic: `dvw`, `dvh`, `dvi`, `dvb`, `dvmin`, `dvmax`

Spec: https://www.w3.org/TR/css-values-4/#viewport-relative-lengths

### String Parser ✅
`layout.ParseLength("<number><unit>")` parses CSS-style length tokens with
case-insensitive unit matching for every implemented unit (absolute, font
relative, viewport, and all the Level 4 aliases above).

## Not Yet Implemented

### Font-Relative Units Still Pending
- **`ex`** - x-height of the font (not in the L4 alias set; `rex` is implemented but resolves to one root row in terminal context)

### Container Query Units (CSS Container Queries)
These require container query context (tracked separately):

- **`cqw`**, **`cqh`** - Container query width/height
- **`cqi`**, **`cqb`** - Container query inline/block
- **`cqmin`**, **`cqmax`** - Container query min/max

## Implementation Notes

### Resolution Process
All length values are resolved to pixels using `ResolveLength()`:

```go
func ResolveLength(l Length, ctx *LayoutContext, currentFontSize float64) float64
```

- **Absolute units**: Direct conversion using CSS reference pixel ratios
- **Font-relative units**: Multiplication by current or root font size
- **Viewport units**: Percentage calculation of viewport dimensions

### Usage in Layout
Length units are used throughout the layout engine:

- **Sizing**: `Width`, `Height`, `MinWidth`, `MaxWidth`, `MinHeight`, `MaxHeight`
- **Spacing**: `Padding`, `Margin`, `Border`
- **Flexbox**: `FlexBasis`, `FlexGap`, `FlexRowGap`, `FlexColumnGap`
- **Grid**: `GridGap`, `GridRowGap`, `GridColumnGap`
- **Positioning**: `Top`, `Right`, `Bottom`, `Left`
- **Intrinsic**: `FitContentWidth`, `FitContentHeight`

### Test Coverage
Comprehensive test suites cover:

- ✅ Unit constructors (`Px()`, `Pt()`, `Em()`, etc.)
- ✅ Length resolution for all implemented units
- ✅ Unit relationships and conversions
- ✅ Integration tests for all Style fields
- ✅ Mixed unit usage in layouts

## Future Enhancements

### Priority: High
1. **`ex` unit** - Requires x-height measurement from font metrics (the alias
   `rex` is implemented; a per-element `ex` would refine `cap`/`rex` once
   font metrics are tracked).
2. **Writing-mode-aware `vi`/`vb`** - Thread element WritingMode through
   `ResolveLength` so vertical writing modes swap inline/block axes.
3. **`lh`/`rlh` refinement** - Honor an explicit line-height when one is
   set on the element (the current resolution uses font size, which matches
   the terminal cell grid).

### Priority: Low
4. **Container query units** - Full container query implementation

## References

- [CSS Values and Units Level 4](https://www.w3.org/TR/css-values-4/)
- [CSS Values and Units Level 5 (Draft)](https://drafts.csswg.org/css-values-5/)
- [CSS Containment Level 3](https://drafts.csswg.org/css-contain-3/)
- [MDN: CSS values and units](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_values_and_units)
