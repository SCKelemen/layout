# Text Layout Specification

This document specifies the text layout implementation for the Go layout engine, based on the [CSS Text Module Level 3](https://www.w3.org/TR/css-text-3/) and [CSS Text Module Level 4](https://www.w3.org/TR/css-text-4/) specifications.

## 1. Overview

The text layout system implements inline text flow, line breaking, and alignment within block containers, integrating with the existing block, flexbox, and grid layout systems.

**Key Design Principles:**
- **Single Node Type**: Keep `Node` as the only public layout tree type
- **Text as Leaf Boxes**: Text nodes are leaf boxes that compute their own size internally
- **Pluggable Metrics**: Text measurement via interface, no hard font library dependencies
- **MVP Scope**: Focus on essential features for v1, defer complex features

**References:**
- [CSS Text Module Level 3](https://www.w3.org/TR/css-text-3/)
- [CSS Text Module Level 4](https://www.w3.org/TR/css-text-4/)
- [CSS Inline Layout Model](https://www.w3.org/TR/CSS2/visuren.html#inline-formatting)

## 2. Scope: v1 MVP vs Future

### 2.1 v1 MVP (What We'll Implement Now)

**Core Features:**
- Horizontal, LTR only (no bidi, no vertical writing)
- No inline elements, only "block with raw text content"
- `white-space`: `normal`, `nowrap`, and `pre` ([§3.1](https://www.w3.org/TR/css-text-3/#white-space-property))
- `text-align`: `left`, `right`, `center` ([§7.1](https://www.w3.org/TR/css-text-3/#text-align-property))
- `line-height`: number (multiplier) and absolute px ([§4.4.1](https://www.w3.org/TR/css-inline-3/#propdef-line-height))
- `text-indent`: first line indentation ([§7.2.1](https://www.w3.org/TR/css-text-3/#text-indent-property))
- `word-spacing` and `letter-spacing`: basic support ([§5.1](https://www.w3.org/TR/css-text-3/#word-spacing-property), [§5.2](https://www.w3.org/TR/css-text-3/#letter-spacing-property))

**Text Measurement:**
- Pluggable `TextMetricsProvider` interface
- Default approximate metrics (fixed char width, line-height multiplier)
- Production users can plug in `x/image/font` or their own renderer

### 2.2 Deferred for Later

- `justify`, `text-justify`, `text-align-last` ([§7.1](https://www.w3.org/TR/css-text-3/#text-align-property), [§7.3](https://www.w3.org/TR/css-text-3/#text-justify-property), [§7.2.2](https://www.w3.org/TR/css-text-3/#text-align-last-property))
- `pre-wrap`, `pre-line` ([§3.1](https://www.w3.org/TR/css-text-3/#white-space-property))
- Hyphenation ([§4.3](https://www.w3.org/TR/css-text-3/#hyphenation))
- Complex scripts / text shaping
- Mixing inline elements + text in one formatting context
- RTL and vertical writing modes ([§2](https://www.w3.org/TR/css-text-3/#writing-modes))

## 3. Node Structure

### 3.1 Single Node Type

**No separate `TextNode` type.** Instead, extend `Node` with optional text content:

```go
type Node struct {
    Style    Style
    Rect     Rect
    Children []*Node
    
    // Optional text content; empty when not used
    Text string  // Text content for text leaf nodes
    
    // Internal: populated by LayoutText for rendering
    TextLayout *TextLayout
}
```

**Note:** Zero values for enums match CSS defaults. For contextual defaults (like `text-align: start`), use an explicit sentinel value.

### 3.2 Display Property

Add `DisplayInlineText` variant to route text nodes to the text layout algorithm:

```go
type Display int

const (
    DisplayBlock Display = iota
    DisplayFlex
    DisplayGrid
    DisplayInlineText  // Text leaf node
    DisplayNone
)
```

**Layout routing:**
```go
func Layout(root *Node, constraints Constraints) Size {
    switch root.Style.Display {
    case DisplayFlex:
        return LayoutFlexbox(root, constraints)
    case DisplayGrid:
        return LayoutGrid(root, constraints)
    case DisplayInlineText:
        return LayoutText(root, constraints)
    case DisplayNone:
        root.Rect = Rect{}
        return Size{}
    default:
        return LayoutBlock(root, constraints)
    }
}
```

## 4. Text Style Properties

Group text properties into a `TextStyle` struct, referenced from `Style`. All properties are based on the [CSS Text Module Level 3](https://www.w3.org/TR/css-text-3/) specification.

```go
type Style struct {
    // Existing box/flex/grid properties...
    Display  Display
    Position Position
    Padding  Spacing
    Border   Spacing
    // ... all existing fields ...
    
    // For text nodes only (nil for non-text nodes):
    TextStyle *TextStyle
}

type TextStyle struct {
    // Alignment (§7.1)
    TextAlign TextAlign  // left, right, center (justify deferred)
    
    // Spacing (§4.4.1, §5.1, §5.2, §7.2.1)
    LineHeight    float64  // <=0 = normal (1.2), 0<x<10 = multiplier, >=10 = absolute px
    WordSpacing   float64  // -1 = normal, otherwise spacing in px
    LetterSpacing float64  // -1 = normal, otherwise spacing in px
    TextIndent    float64  // First line indent (0 = none)
    
    // Wrapping (§3.1)
    WhiteSpace WhiteSpace  // normal, nowrap, pre
    
    // Font (for measurement)
    FontSize   float64
    FontFamily string
    FontWeight FontWeight  // or numeric
    
    // Direction (§2) - LTR only for v1
    Direction Direction  // ltr, rtl (rtl deferred)
}
```

### 4.1 Text Alignment

**Property:** `text-align`  
**Specification:** [CSS Text Module Level 3 §7.1](https://www.w3.org/TR/css-text-3/#text-align-property)

```go
type TextAlign int

const (
    TextAlignDefault TextAlign = iota  // CSS default: 'start' (contextual - left in LTR)
    TextAlignLeft
    TextAlignRight
    TextAlignCenter
    // TextAlignJustify  // deferred (§7.1.1)
    // TextAlignStart    // deferred (LTR only for v1)
    // TextAlignEnd      // deferred
)
```

**Values:**
- `TextAlignDefault` (zero value): CSS default `start` - resolves to `left` in LTR context ([§7.1](https://www.w3.org/TR/css-text-3/#valdef-text-align-start))
- `left`: Aligns text to the left edge of the line box ([§7.1](https://www.w3.org/TR/css-text-3/#valdef-text-align-left))
- `right`: Aligns text to the right edge of the line box ([§7.1](https://www.w3.org/TR/css-text-3/#valdef-text-align-right))
- `center`: Centers text within the line box ([§7.1](https://www.w3.org/TR/css-text-3/#valdef-text-align-center))
- `justify`: Stretches text to fill the line width (deferred for v1) ([§7.1.1](https://www.w3.org/TR/css-text-3/#justify))

**Default:** `TextAlignDefault` (resolves to `left` in LTR context for v1).

### 4.2 White Space

**Property:** `white-space`  
**Specification:** [CSS Text Module Level 3 §3.1](https://www.w3.org/TR/css-text-3/#white-space-property)

```go
type WhiteSpace int

const (
    WhiteSpaceNormal WhiteSpace = iota  // CSS default (zero value)
    WhiteSpaceNowrap
    WhiteSpacePre
    // WhiteSpacePreWrap  // deferred (§3.1)
    // WhiteSpacePreLine  // deferred (§3.1)
)
```

**Values:**
- `WhiteSpaceNormal` (zero value): CSS default - collapse whitespace sequences, wrap text ([§3.1](https://www.w3.org/TR/css-text-3/#valdef-white-space-normal))
- `nowrap`: Collapse whitespace, prevent wrapping ([§3.1](https://www.w3.org/TR/css-text-3/#valdef-white-space-nowrap))
- `pre`: Preserve whitespace and newlines, prevent wrapping ([§3.1](https://www.w3.org/TR/css-text-3/#valdef-white-space-pre))
- `pre-wrap`: Preserve whitespace, allow wrapping (deferred) ([§3.1](https://www.w3.org/TR/css-text-3/#valdef-white-space-pre-wrap))
- `pre-line`: Collapse whitespace, preserve newlines, allow wrapping (deferred) ([§3.1](https://www.w3.org/TR/css-text-3/#valdef-white-space-pre-line))

**Default:** `WhiteSpaceNormal` (zero value)

### 4.3 Line Height

**Property:** `line-height`  
**Specification:** [CSS Inline Layout Module Level 3 §4.4.1](https://www.w3.org/TR/css-inline-3/#propdef-line-height)

```go
// LineHeight interpretation:
// <= 0: normal (typically 1.2 × fontSize)
// 0 < x < 10: multiplier (x × fontSize)
// >= 10: absolute pixels
```

**Values:**
- `<number>`: Multiplier of font size ([§4.4.1](https://www.w3.org/TR/css-inline-3/#valdef-line-height-number))
- `<length>`: Absolute line height in pixels ([§4.4.1](https://www.w3.org/TR/css-inline-3/#valdef-line-height-length))
- `normal`: Uses default (typically 1.2 × font size) ([§4.4.1](https://www.w3.org/TR/css-inline-3/#valdef-line-height-normal))

**Default:** `normal` (1.2 × font size)

### 4.4 Text Indent

**Property:** `text-indent`  
**Specification:** [CSS Text Module Level 3 §7.2.1](https://www.w3.org/TR/css-text-3/#text-indent-property)

```go
// TextIndent: First line indent in pixels
// 0 = no indent
// Positive = indent right (LTR)
// Negative = outdent (hanging indent)
```

**Values:**
- `<length>`: Indent amount in pixels ([§7.2.1](https://www.w3.org/TR/css-text-3/#valdef-text-indent-length))
- `0`: No indent

**Default:** `0`

**Behavior:** Only affects the first line of the first block-level box ([§7.2.1](https://www.w3.org/TR/css-text-3/#text-indent-property)).

### 4.5 Word Spacing

**Property:** `word-spacing`  
**Specification:** [CSS Text Module Level 3 §5.1](https://www.w3.org/TR/css-text-3/#word-spacing-property)

```go
// WordSpacing: Spacing between words
// -1 = normal (typically 0.25em)
// Otherwise: spacing in pixels
```

**Values:**
- `normal`: Default word spacing (typically 0.25em) ([§5.1](https://www.w3.org/TR/css-text-3/#valdef-word-spacing-normal))
- `<length>`: Additional spacing between words ([§5.1](https://www.w3.org/TR/css-text-3/#valdef-word-spacing-length))

**Default:** `normal`

### 4.6 Letter Spacing

**Property:** `letter-spacing`  
**Specification:** [CSS Text Module Level 3 §5.2](https://www.w3.org/TR/css-text-3/#letter-spacing-property)

```go
// LetterSpacing: Spacing between characters
// -1 = normal (typically 0)
// Otherwise: spacing in pixels
```

**Values:**
- `normal`: Default letter spacing (typically 0) ([§5.2](https://www.w3.org/TR/css-text-3/#valdef-letter-spacing-normal))
- `<length>`: Additional spacing between characters ([§5.2](https://www.w3.org/TR/css-text-3/#valdef-letter-spacing-length))

**Default:** `normal`

### 4.7 Direction

**Property:** `direction`  
**Specification:** [CSS Writing Modes Level 3](https://www.w3.org/TR/css-writing-modes-3/#propdef-direction)

```go
type Direction int

const (
    DirectionLTR Direction = iota  // CSS default (zero value)
    // DirectionRTL  // deferred
)
```

**Values:**
- `DirectionLTR` (zero value): CSS default - left-to-right text direction
- `rtl`: Right-to-left text direction (deferred for v1)

**Default:** `DirectionLTR` (zero value)

## 5. Text Measurement

### 5.1 TextMetricsProvider Interface

Text measurement is abstracted via a pluggable interface, allowing users to provide their own font measurement implementation.

```go
// TextMetricsProvider abstracts text measurement.
// Users can provide their own implementation (e.g., using x/image/font)
// or use the default approximate metrics.
type TextMetricsProvider interface {
    // Measure returns the advance width of 'text' in the given style.
    // Optionally returns ascent/descent for line height calculations.
    // Based on CSS font metrics: https://www.w3.org/TR/css-fonts-3/#font-metrics
    Measure(text string, style TextStyle) (advance, ascent, descent float64)
}

// TextStyle is used for measurement (subset of full TextStyle)
type TextStyle struct {
    FontSize      float64
    FontFamily    string
    FontWeight    FontWeight
    LetterSpacing float64
    WordSpacing   float64
}
```

### 5.2 Default Implementation

```go
// Default approximate metrics (v1 fallback)
type approxMetrics struct{}

func (a *approxMetrics) Measure(text string, style TextStyle) (advance, ascent, descent float64) {
    // Simple approximation: fixed char width
    charWidth := style.FontSize * 0.6  // Rough average
    advance = float64(len(text)) * charWidth
    
    // Add letter spacing
    if style.LetterSpacing > 0 {
        advance += float64(len(text)-1) * style.LetterSpacing
    }
    
    // Line metrics (based on typical font metrics)
    ascent = style.FontSize * 0.8
    descent = style.FontSize * 0.2
    
    return advance, ascent, descent
}

// Package-level provider
var textMetrics TextMetricsProvider = &approxMetrics{}

// SetTextMetricsProvider allows users to plug in their own measurement.
func SetTextMetricsProvider(p TextMetricsProvider) {
    if p != nil {
        textMetrics = p
    }
}
```

**Benefits:**
- Core layout package completely decoupled from specific font libs
- Compatible with "pure Go + /x/ only" ethos
- Production users can use `x/image/font` or their renderer's measurement
- Tests can use deterministic fake metrics

## 6. Text Layout Algorithm

### 6.1 LayoutText Function

**Function:** `LayoutText`  
**Specification:** Based on [CSS Text Module Level 3](https://www.w3.org/TR/css-text-3/) inline formatting model

```go
// LayoutText lays out text within a node, computing its Size and internal line boxes.
// It assumes the node is a text leaf: node.Text is non-empty, no children.
// The node should have DisplayInlineText set.
// 
// Algorithm based on CSS Text Module Level 3:
// - §3: White Space Processing
// - §4: Line Breaking and Word Boundaries
// - §5: Text Transformations and Spacing
// - §7: Text Alignment
func LayoutText(node *Node, constraints Constraints) Size {
    // 1. Determine available content width from constraints and Style (box sizing)
    // 2. Normalize white-space (§3.1)
    // 3. Perform line breaking (§4) with textMetrics.Measure
    // 4. Compute per-line positions (x,y) based on text-align (§7.1) and text-indent (§7.2.1)
    // 5. Compute total height from line count and line-height (§4.4.1)
    // 6. Set node.Rect.Width/Height (including padding/border)
    // 7. Store line metadata for rendering (e.g., node.TextLayout = []TextLine)
    // 8. Return constrained Size
}
```

### 6.2 Available Content Width

Based on CSS Box Model: [CSS Box Model Module Level 3](https://www.w3.org/TR/css-box-3/)

```go
// Similar to LayoutBlock's content width calculation
availableWidth := constraints.MaxWidth
horizontalPaddingBorder := node.Style.Padding.Left + node.Style.Padding.Right + 
                          node.Style.Border.Left + node.Style.Border.Right
contentWidth := availableWidth - horizontalPaddingBorder
if contentWidth < 0 {
    contentWidth = 0
}
```

### 6.3 White Space Processing

**Specification:** [CSS Text Module Level 3 §3.1](https://www.w3.org/TR/css-text-3/#white-space-property)

```go
func preprocessText(text string, whiteSpace WhiteSpace) string {
    switch whiteSpace {
    case WhiteSpaceNormal:
        // §3.1: Collapse runs of spaces/tabs to single space
        // Turn \n and \r\n into spaces
        text = strings.ReplaceAll(text, "\r\n", " ")
        text = strings.ReplaceAll(text, "\n", " ")
        text = strings.ReplaceAll(text, "\r", " ")
        // Collapse multiple spaces
        text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
        return strings.TrimSpace(text)
        
    case WhiteSpaceNowrap:
        // §3.1: Same as normal, but no wrapping (handled in line breaking)
        return preprocessText(text, WhiteSpaceNormal)
        
    case WhiteSpacePre:
        // §3.1: Preserve spaces and newlines, no wrapping
        return text
        
    default:
        return text
    }
}
```

### 6.4 Line Breaking

**Specification:** [CSS Text Module Level 3 §4](https://www.w3.org/TR/css-text-3/#line-breaking)

Word-based line breaking for v1:

```go
func breakIntoLines(text string, maxWidth float64, style TextStyle, metrics TextMetricsProvider) []TextLine {
    tokens := splitIntoTokens(text, style.WhiteSpace)  // words + whitespace
    lines := []TextLine{}
    current := newLine()
    currentWidth := getFirstLineIndent(style)  // §7.2.1
    
    for _, tok := range tokens {
        w := measureToken(tok, style, metrics)
        
        // §4: Check if we need to break at word boundary
        if maxWidth > 0 && currentWidth + w > maxWidth && 
           !current.isEmpty() && canBreakBefore(tok, style.WhiteSpace) {
            lines = append(lines, current)
            current = newLine()
            currentWidth = 0
        }
        
        current.add(tok, w)
        currentWidth += w
    }
    
    if !current.isEmpty() {
        lines = append(lines, current)
    }
    
    return lines
}

func getFirstLineIndent(style TextStyle) float64 {
    // §7.2.1: Text indent applies to first line only
    if style.TextIndent > 0 {
        return style.TextIndent
    }
    return 0
}

func canBreakBefore(tok token, whiteSpace WhiteSpace) bool {
    if whiteSpace == WhiteSpacePre {
        return false  // §3.1: No wrapping in pre
    }
    if whiteSpace == WhiteSpaceNowrap {
        return false  // §3.1: No wrapping in nowrap
    }
    // §4: Can break before word tokens in normal mode
    return tok.isWord()
}
```

### 6.5 Text Alignment

**Specification:** [CSS Text Module Level 3 §7.1](https://www.w3.org/TR/css-text-3/#text-align-property)

```go
func positionLine(line TextLine, contentWidth float64, textAlign TextAlign) {
    lineWidth := line.totalWidth()
    
    switch textAlign {
    case TextAlignLeft, TextAlignStart:
        // §7.1: Align to left edge
        line.offsetX = 0
        
    case TextAlignRight, TextAlignEnd:
        // §7.1: Align to right edge
        line.offsetX = contentWidth - lineWidth
        
    case TextAlignCenter:
        // §7.1: Center within line box
        line.offsetX = (contentWidth - lineWidth) / 2
        
    case TextAlignJustify:
        // §7.1.1: v1: treat as left, implement real justify later
        line.offsetX = 0
    }
}
```

### 6.6 Line Height Resolution

**Specification:** [CSS Inline Layout Module Level 3 §4.4.1](https://www.w3.org/TR/css-inline-3/#propdef-line-height)

```go
func resolveLineHeight(lineHeight float64, fontSize float64) float64 {
    if lineHeight <= 0 {
        // §4.4.1: Normal: typically 1.2 × fontSize
        return fontSize * 1.2
    }
    if lineHeight < 10 {
        // §4.4.1: Treat as multiplier
        return fontSize * lineHeight
    }
    // §4.4.1: Treat as absolute pixels
    return lineHeight
}
```

### 6.7 Final Size Calculation

Based on CSS Box Model: [CSS Box Model Module Level 3](https://www.w3.org/TR/css-box-3/)

```go
func (node *Node) computeTextSize(lines []TextLine, lineHeight float64, constraints Constraints) Size {
    // Content dimensions
    contentWidth := 0.0
    for _, line := range lines {
        if line.width > contentWidth {
            contentWidth = line.width
        }
    }
    
    contentHeight := float64(len(lines)) * lineHeight
    
    // Apply explicit width/height if set
    if node.Style.Width >= 0 {
        contentWidth = node.Style.Width
    }
    if node.Style.Height >= 0 {
        contentHeight = node.Style.Height
    }
    
    // Add padding/border (like block layout)
    horizontalPaddingBorder := node.Style.Padding.Left + node.Style.Padding.Right +
                              node.Style.Border.Left + node.Style.Border.Right
    verticalPaddingBorder := node.Style.Padding.Top + node.Style.Padding.Bottom +
                            node.Style.Border.Top + node.Style.Border.Bottom
    
    outerWidth := contentWidth + horizontalPaddingBorder
    outerHeight := contentHeight + verticalPaddingBorder
    
    // Constrain and set Rect
    size := constraints.Constrain(Size{Width: outerWidth, Height: outerHeight})
    node.Rect.Width = size.Width
    node.Rect.Height = size.Height
    
    // Store line metadata for rendering (optional)
    node.TextLayout = &TextLayout{
        Lines: lines,
        LineHeight: lineHeight,
    }
    
    return size
}
```

## 7. Integration with Layout Systems

### 7.1 Block Layout

**Specification:** [CSS Box Model Module Level 3](https://www.w3.org/TR/css-box-3/)

Block layout treats text nodes as regular children:

```go
// In LayoutBlock:
for _, child := range node.Children {
    // Check if child is a text node
    if child.Style.Display == DisplayInlineText {
        childSize = LayoutText(child, childConstraints)
    } else {
        childSize = Layout(child, childConstraints)
    }
    
    // Position child (same as before)
    child.Rect.X = node.Style.Padding.Left + node.Style.Border.Left
    child.Rect.Y = node.Style.Padding.Top + node.Style.Border.Top + currentY
    currentY += childSize.Height
}
```

**Flow:**
```
Block Container (width: 300px, height: auto)
  └─ Node (DisplayInlineText, Text: "Lorem ipsum...")
      └─ LayoutText calculates:
          - Line 1: "Lorem ipsum dolor" (fits in 300px)
          - Line 2: "sit amet..." (fits in 300px)
          - Total height: 2 * line-height
      └─ Block height becomes: 2 * line-height + padding
```

### 7.2 Flexbox Layout

**Specification:** [CSS Flexible Box Layout Module Level 1](https://www.w3.org/TR/css-flexbox-1/)

Text nodes as flex items use their intrinsic size:

```go
// In LayoutFlexbox:
// Text nodes are treated as inflexible items (for v1)
// Their base size is computed by LayoutText with Unbounded width
// Then flex sizing applies to the text container

if child.Style.Display == DisplayInlineText {
    // For v1: use intrinsic width (Unbounded)
    // Future: could re-run LayoutText with allocated width
    childSize = LayoutText(child, Unconstrained())
} else {
    childSize = Layout(child, childConstraints)
}
```

**Note:** For v1, text nodes in flex are "intrinsic only" - they don't re-wrap based on flex allocation. This can be enhanced later.

### 7.3 Grid Layout

**Specification:** [CSS Grid Layout Module Level 1](https://www.w3.org/TR/css-grid-1/)

Text nodes flow within their grid cell:

```go
// In LayoutGrid:
// Text nodes get cell width as MaxWidth constraint
cellWidth := calculateCellWidth(...)
childConstraints := Constraints{
    MaxWidth: cellWidth,
    MaxHeight: Unbounded,  // or cell height if constrained
}

if child.Style.Display == DisplayInlineText {
    childSize = LayoutText(child, childConstraints)
} else {
    childSize = Layout(child, childConstraints)
}
```

**Key Point:** Grid passes the cell width as `MaxWidth` to `LayoutText`, which constrains line wrapping.

## 8. Internal Types for Inline Formatting

While `Node` is the only public type, `LayoutText` uses internal types for inline formatting, based on the [CSS Inline Layout Model](https://www.w3.org/TR/CSS2/visuren.html#inline-formatting):

```go
// Internal to LayoutText - not exposed
type InlineBoxKind int

const (
    InlineBoxText InlineBoxKind = iota
    // InlineBoxInlineNode  // for future: spans, inline images
)

type InlineBox struct {
    Kind    InlineBoxKind
    Text    string      // for InlineBoxText
    Node    *Node       // for InlineBoxInlineNode (future)
    Width   float64
    Ascent  float64
    Descent float64
}

type TextLine struct {
    Boxes    []InlineBox
    Width    float64
    OffsetX  float64  // X offset for text-align
    OffsetY  float64  // Y position (cumulative)
}

type TextLayout struct {
    Lines      []TextLine
    LineHeight float64
}
```

These are stored on `Node` for rendering:

```go
type Node struct {
    // ... existing fields ...
    TextLayout *TextLayout  // Populated by LayoutText, used by renderer
}
```

## 9. API Design

### 9.1 Creating Text Nodes

```go
// Helper function for convenience
func Text(text string, style Style) *Node {
    return &Node{
        Text: text,
        Style: Style{
            Display: DisplayInlineText,
            TextStyle: &TextStyle{
                FontSize: 16,  // Default
                TextAlign: TextAlignLeft,
                LineHeight: 0,  // Normal
                WhiteSpace: WhiteSpaceNormal,
            },
            // Merge with provided style
            // ... (style merging logic)
        },
    }
}

// Usage
textNode := layout.Text("Hello, world!", layout.Style{
    Width: 200,
    TextStyle: &layout.TextStyle{
        TextAlign: layout.TextAlignCenter,
        LineHeight: 1.5,
    },
})

// Or manually
node := &layout.Node{
    Text: "Hello, world!",
    Style: layout.Style{
        Display: layout.DisplayInlineText,
        TextStyle: &layout.TextStyle{
            FontSize: 16,
            TextAlign: layout.TextAlignLeft,
        },
    },
}
```

## 10. Testing Strategy

### 10.1 Test Invariants

Based on CSS specification requirements:

1. **Wrapping Invariants** ([§4](https://www.w3.org/TR/css-text-3/#line-breaking)):
   - Given `MaxWidth = N`, all lines' `lineWidth` must be `<= N + epsilon`
   - If you shrink `MaxWidth`, the number of lines must be `>=` previous line count

2. **Alignment** ([§7.1](https://www.w3.org/TR/css-text-3/#text-align-property)):
   - For `TextAlignLeft`: first glyph x >= 0, last glyph x + advance <= contentWidth
   - For `TextAlignRight`: last glyph baseline x is near `contentWidth`
   - For `TextAlignCenter`: line's center ≈ `contentWidth/2`

3. **Interaction with Block Height** ([CSS Box Model](https://www.w3.org/TR/css-box-3/)):
   - Block with only text and `height:auto` should have `Rect.Height ≈ lineHeight * numLines + padding+border`
   - Increasing text increases number of lines and height monotonically

4. **White-Space** ([§3.1](https://www.w3.org/TR/css-text-3/#white-space-property)):
   - Consecutive spaces collapse under `normal`
   - Newlines preserved under `pre`

### 10.2 Fake TextMetricsProvider for Tests

```go
type fakeMetrics struct {
    charWidth float64
}

func (f *fakeMetrics) Measure(text string, style TextStyle) (advance, ascent, descent float64) {
    advance = float64(len(text)) * f.charWidth
    if style.LetterSpacing > 0 {
        advance += float64(len(text)-1) * style.LetterSpacing
    }
    ascent = style.FontSize * 0.8
    descent = style.FontSize * 0.2
    return advance, ascent, descent
}

// Use in tests
func TestTextWrapping(t *testing.T) {
    layout.SetTextMetricsProvider(&fakeMetrics{charWidth: 10})
    // ... test with deterministic metrics
}
```

## 11. Summary: Implementation Prompt

> Implement text layout as a leaf-node, LTR-only, word-wrapping engine.
> 
> Add `DisplayInlineText` and route it to `LayoutText`.
> 
> `LayoutText` should:
> - Use a pluggable `TextMetricsProvider` for widths
> - Support `white-space: normal / nowrap / pre` ([§3.1](https://www.w3.org/TR/css-text-3/#white-space-property))
> - Support `text-align: left/right/center` ([§7.1](https://www.w3.org/TR/css-text-3/#text-align-property)), `line-height` ([§4.4.1](https://www.w3.org/TR/css-inline-3/#propdef-line-height)), and `text-indent` ([§7.2.1](https://www.w3.org/TR/css-text-3/#text-indent-property))
> - Break into lines based on `MaxWidth` and return the resulting width/height to the layout engine
> 
> Block, flex, and grid should treat text nodes as regular boxes that just happen to compute their own size based on text.

## 12. References

### CSS Specifications

- [CSS Text Module Level 3](https://www.w3.org/TR/css-text-3/) - Primary specification for text layout
- [CSS Text Module Level 4](https://www.w3.org/TR/css-text-4/) - Latest text layout features
- [CSS Inline Layout Module Level 3](https://www.w3.org/TR/css-inline-3/) - Line height and inline formatting
- [CSS Box Model Module Level 3](https://www.w3.org/TR/css-box-3/) - Box sizing and padding/border
- [CSS Flexible Box Layout Module Level 1](https://www.w3.org/TR/css-flexbox-1/) - Flexbox integration
- [CSS Grid Layout Module Level 1](https://www.w3.org/TR/css-grid-1/) - Grid integration
- [CSS Writing Modes Level 3](https://www.w3.org/TR/css-writing-modes-3/) - Text direction
- [CSS Fonts Module Level 3](https://www.w3.org/TR/css-fonts-3/) - Font metrics

### Specific Sections

- [§3.1 White Space Processing](https://www.w3.org/TR/css-text-3/#white-space-property)
- [§4 Line Breaking](https://www.w3.org/TR/css-text-3/#line-breaking)
- [§5.1 Word Spacing](https://www.w3.org/TR/css-text-3/#word-spacing-property)
- [§5.2 Letter Spacing](https://www.w3.org/TR/css-text-3/#letter-spacing-property)
- [§7.1 Text Alignment](https://www.w3.org/TR/css-text-3/#text-align-property)
- [§7.2.1 Text Indent](https://www.w3.org/TR/css-text-3/#text-indent-property)
- [§4.4.1 Line Height](https://www.w3.org/TR/css-inline-3/#propdef-line-height)
