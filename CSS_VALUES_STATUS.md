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

## Not Yet Implemented

### Relative Font Units (CSS Level 4)
These units require additional text metrics:

- **`ex`** - x-height of the font
- **`cap`** - Cap height of the font
- **`ic`** - Width of ideographic character (水)
- **`lh`** - Line height of the element
- **`rlh`** - Line height of the root element

### Logical Viewport Units (CSS Level 4)
These units depend on writing mode:

- **`vi`** - 1% of viewport size in inline axis
- **`vb`** - 1% of viewport size in block axis

### Dynamic Viewport Units (CSS Level 5)
These units account for dynamic browser UI:

- **`dvh`**, **`dvw`**, **`dvmin`**, **`dvmax`** - Dynamic viewport units
- **`svh`**, **`svw`**, **`svmin`**, **`svmax`** - Small viewport units
- **`lvh`**, **`lvw`**, **`lvmin`**, **`lvmax`** - Large viewport units

### Container Query Units (CSS Container Queries)
These require container query context:

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
1. **`ex` unit** - Requires x-height measurement from font metrics
2. **`lh` / `rlh` units** - Requires line height calculation
3. **Logical viewport units** (`vi`, `vb`) - Requires writing mode support

### Priority: Medium
4. **`cap` unit** - Capital letter height measurement
5. **`ic` unit** - Ideographic character width measurement
6. **Dynamic viewport units** - Browser UI-aware measurements

### Priority: Low
7. **Container query units** - Full container query implementation

## References

- [CSS Values and Units Level 4](https://www.w3.org/TR/css-values-4/)
- [CSS Values and Units Level 5 (Draft)](https://drafts.csswg.org/css-values-5/)
- [CSS Containment Level 3](https://drafts.csswg.org/css-contain-3/)
- [MDN: CSS values and units](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_values_and_units)
