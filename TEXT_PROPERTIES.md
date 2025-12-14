# Text Properties Reference

This document provides examples and usage guide for text layout properties in the layout engine.

## Font Properties

### Font Style

Controls whether text is italic, oblique, or normal (upright).

**CSS Property**: `font-style`
**Specification**: [CSS Fonts Module Level 4 §5.2.2](https://www.w3.org/TR/css-fonts-4/#font-style-prop)
**MDN**: [font-style](https://developer.mozilla.org/en-US/docs/Web/CSS/font-style)

```go
// Normal (upright) text
node.Style.TextStyle.FontStyle = FontStyleNormal

// Italic text (cursive/calligraphic)
node.Style.TextStyle.FontStyle = FontStyleItalic

// Oblique text (slanted)
node.Style.TextStyle.FontStyle = FontStyleOblique
```

**Example**:
```go
italicText := Text("This is italic text", Style{
    TextStyle: &TextStyle{
        FontSize:  16,
        FontStyle: FontStyleItalic,
    },
})
```

### Font Weight

Controls the boldness of text.

**CSS Property**: `font-weight`
**Specification**: [CSS Fonts Module Level 4 §5.2.3](https://www.w3.org/TR/css-fonts-4/#font-weight-prop)
**MDN**: [font-weight](https://developer.mozilla.org/en-US/docs/Web/CSS/font-weight)

```go
// Normal weight
node.Style.TextStyle.FontWeight = FontWeightNormal  // 400

// Bold weight
node.Style.TextStyle.FontWeight = FontWeightBold    // 700

// Custom weight (100-900)
node.Style.TextStyle.FontWeight = FontWeight(600)
```

## Text Decoration

### Text Decoration Line

Controls which decoration lines are shown. Multiple decorations can be combined using bitwise OR.

**CSS Property**: `text-decoration-line`
**Specification**: [CSS Text Decoration Module Level 3 §2.1](https://www.w3.org/TR/css-text-decor-3/#text-decoration-line-property)
**MDN**: [text-decoration-line](https://developer.mozilla.org/en-US/docs/Web/CSS/text-decoration-line)

```go
// No decoration
node.Style.TextStyle.TextDecoration = TextDecorationNone

// Underline
node.Style.TextStyle.TextDecoration = TextDecorationUnderline

// Overline (line above text)
node.Style.TextStyle.TextDecoration = TextDecorationOverline

// Line through (strikethrough)
node.Style.TextStyle.TextDecoration = TextDecorationLineThrough

// Multiple decorations (combine with |)
node.Style.TextStyle.TextDecoration = TextDecorationUnderline | TextDecorationLineThrough
```

**Example**:
```go
underlinedText := Text("Underlined link", Style{
    TextStyle: &TextStyle{
        FontSize:       16,
        TextDecoration: TextDecorationUnderline,
    },
})

// Check if a decoration is present
if node.Style.TextStyle.TextDecoration.Has(TextDecorationUnderline) {
    // Render underline
}
```

### Text Decoration Style

Controls the style of decoration lines.

**CSS Property**: `text-decoration-style`
**Specification**: [CSS Text Decoration Module Level 3 §2.2](https://www.w3.org/TR/css-text-decor-3/#text-decoration-style-property)
**MDN**: [text-decoration-style](https://developer.mozilla.org/en-US/docs/Web/CSS/text-decoration-style)

```go
// Solid line (default)
node.Style.TextStyle.TextDecorationStyle = TextDecorationStyleSolid

// Double line
node.Style.TextStyle.TextDecorationStyle = TextDecorationStyleDouble

// Dotted line
node.Style.TextStyle.TextDecorationStyle = TextDecorationStyleDotted

// Dashed line
node.Style.TextStyle.TextDecorationStyle = TextDecorationStyleDashed

// Wavy line
node.Style.TextStyle.TextDecorationStyle = TextDecorationStyleWavy
```

**Example**:
```go
wavyUnderline := Text("Spelling error", Style{
    TextStyle: &TextStyle{
        FontSize:            14,
        TextDecoration:      TextDecorationUnderline,
        TextDecorationStyle: TextDecorationStyleWavy,
        TextDecorationColor: "red",
    },
})
```

### Text Decoration Color

Controls the color of decoration lines. Can be any CSS color string.

**CSS Property**: `text-decoration-color`
**Specification**: [CSS Text Decoration Module Level 3 §2.3](https://www.w3.org/TR/css-text-decor-3/#text-decoration-color-property)
**MDN**: [text-decoration-color](https://developer.mozilla.org/en-US/docs/Web/CSS/text-decoration-color)

```go
// Use current text color (default)
node.Style.TextStyle.TextDecorationColor = ""

// Specific color
node.Style.TextStyle.TextDecorationColor = "red"
node.Style.TextStyle.TextDecorationColor = "#ff0000"
node.Style.TextStyle.TextDecorationColor = "rgb(255, 0, 0)"
```

## Vertical Alignment

Controls how inline elements are aligned vertically relative to their parent or line box.

**CSS Property**: `vertical-align`
**Specification**: [CSS Inline Layout Module Level 3 §3.2](https://www.w3.org/TR/css-inline-3/#propdef-vertical-align)
**MDN**: [vertical-align](https://developer.mozilla.org/en-US/docs/Web/CSS/vertical-align)

```go
// Align baseline with parent baseline (default)
node.Style.TextStyle.VerticalAlign = VerticalAlignBaseline

// Subscript (lower baseline)
node.Style.TextStyle.VerticalAlign = VerticalAlignSub

// Superscript (raise baseline)
node.Style.TextStyle.VerticalAlign = VerticalAlignSuper

// Align top with parent's text top
node.Style.TextStyle.VerticalAlign = VerticalAlignTextTop

// Align bottom with parent's text bottom
node.Style.TextStyle.VerticalAlign = VerticalAlignTextBottom

// Align middle with parent's middle
node.Style.TextStyle.VerticalAlign = VerticalAlignMiddle

// Align top with line box top
node.Style.TextStyle.VerticalAlign = VerticalAlignTop

// Align bottom with line box bottom
node.Style.TextStyle.VerticalAlign = VerticalAlignBottom
```

**Example**:
```go
// Chemical formula: H₂O
formula := []Node{
    Text("H", Style{TextStyle: &TextStyle{FontSize: 16}}),
    Text("2", Style{TextStyle: &TextStyle{
        FontSize:      12,
        VerticalAlign: VerticalAlignSub,
    }}),
    Text("O", Style{TextStyle: &TextStyle{FontSize: 16}}),
}

// Mathematical expression: E=mc²
equation := []Node{
    Text("E=mc", Style{TextStyle: &TextStyle{FontSize: 16}}),
    Text("2", Style{TextStyle: &TextStyle{
        FontSize:      12,
        VerticalAlign: VerticalAlignSuper,
    }}),
}
```

## Combined Properties

All text properties can be used together:

```go
fancyText := Text("Fancy styled text", Style{
    TextStyle: &TextStyle{
        // Font
        FontSize:   18,
        FontWeight: FontWeightBold,
        FontStyle:  FontStyleItalic,

        // Decoration
        TextDecoration:      TextDecorationUnderline,
        TextDecorationStyle: TextDecorationStyleDashed,
        TextDecorationColor: "blue",

        // Alignment
        TextAlign:     TextAlignCenter,
        VerticalAlign: VerticalAlignMiddle,

        // Spacing
        LineHeight:    1.5,
        LetterSpacing: 2,
        WordSpacing:   4,
    },
})
```

## Rendering Considerations

These properties provide metadata for renderers. The layout engine:

- **Preserves** all property values in the layout tree
- **Does NOT** apply visual effects (that's the renderer's job)
- **May adjust** layout metrics for properties like `vertical-align` (future enhancement)

Renderers should:

1. Read properties from `node.Style.TextStyle`
2. Apply visual effects (italic slant, decoration lines, etc.)
3. Use `VerticalAlign` to adjust baseline positioning
4. Respect `FontStyle` when selecting font faces

## Default Values

All properties have sensible defaults:

| Property | Default Value |
|----------|--------------|
| `FontStyle` | `FontStyleNormal` |
| `FontWeight` | `FontWeightNormal` (400) |
| `TextDecoration` | `TextDecorationNone` |
| `TextDecorationStyle` | `TextDecorationStyleSolid` |
| `TextDecorationColor` | `""` (use currentColor) |
| `VerticalAlign` | `VerticalAlignBaseline` |

## CSS Specification Compliance

These implementations follow the relevant CSS specifications:

- **Font Style**: [CSS Fonts Module Level 4](https://www.w3.org/TR/css-fonts-4/)
- **Text Decoration**: [CSS Text Decoration Module Level 3](https://www.w3.org/TR/css-text-decor-3/)
- **Vertical Align**: [CSS Inline Layout Module Level 3](https://www.w3.org/TR/css-inline-3/)

## Examples

### Link with Hover State

```go
func createLink(text string, isHovered bool) *Node {
    decoration := TextDecorationNone
    if isHovered {
        decoration = TextDecorationUnderline
    }

    return Text(text, Style{
        TextStyle: &TextStyle{
            FontSize:       16,
            TextDecoration: decoration,
        },
    })
}
```

### Emphasized Text

```go
func createEmphasis(text string, level int) *Node {
    var style TextStyle

    switch level {
    case 1: // <em>
        style = TextStyle{
            FontSize:  16,
            FontStyle: FontStyleItalic,
        }
    case 2: // <strong>
        style = TextStyle{
            FontSize:   16,
            FontWeight: FontWeightBold,
        }
    case 3: // <strong><em>
        style = TextStyle{
            FontSize:   16,
            FontWeight: FontWeightBold,
            FontStyle:  FontStyleItalic,
        }
    }

    return Text(text, Style{TextStyle: &style})
}
```

### Deleted and Inserted Text

```go
// Deleted text (strikethrough)
deleted := Text("removed content", Style{
    TextStyle: &TextStyle{
        FontSize:       14,
        TextDecoration: TextDecorationLineThrough,
    },
})

// Inserted text (underline)
inserted := Text("new content", Style{
    TextStyle: &TextStyle{
        FontSize:       14,
        TextDecoration: TextDecorationUnderline,
    },
})
```
