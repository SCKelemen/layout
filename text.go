package layout

import (
	"strings"
	"unicode"

	"github.com/rivo/uniseg"
)

// TextMetricsProvider abstracts text measurement.
// Users can provide their own implementation (e.g., using x/image/font)
// or use the default approximate metrics.
//
// Based on CSS font metrics: https://www.w3.org/TR/css-fonts-3/#font-metrics
type TextMetricsProvider interface {
	// Measure returns the advance width of 'text' in the given style.
	// Optionally returns ascent/descent for line height calculations.
	Measure(text string, style TextStyle) (advance, ascent, descent float64)
}

// Default approximate metrics (v1 fallback)
type approxMetrics struct{}

func (a *approxMetrics) Measure(text string, style TextStyle) (advance, ascent, descent float64) {
	// Simple approximation: fixed char width
	// Count runes (characters), not bytes, to handle Unicode correctly
	runeCount := len([]rune(text))
	charWidth := style.FontSize * 0.6 // Rough average
	advance = float64(runeCount) * charWidth

	// Add letter spacing (can be positive or negative)
	// Letter spacing applies between characters (not after last one)
	if style.LetterSpacing != -1 && runeCount > 0 {
		advance += float64(runeCount-1) * style.LetterSpacing
	}

	// Line metrics (based on typical font metrics)
	ascent = style.FontSize * 0.8
	descent = style.FontSize * 0.2

	return advance, ascent, descent
}

// Package-level provider
var textMetrics TextMetricsProvider = &approxMetrics{}

// SetTextMetricsProvider allows users to plug in their own measurement.
//
// Thread Safety: This function modifies a package-level variable. For concurrent
// use, set the provider once at initialization time and do not change it during
// layout operations. If you need to change providers concurrently, use external
// synchronization.
func SetTextMetricsProvider(p TextMetricsProvider) {
	if p != nil {
		textMetrics = p
	}
}

// LayoutText lays out text within a node, computing its Size and internal line boxes.
//
// Requirements:
// - The node should have DisplayInlineText set
// - node.Text should be non-empty (empty text produces minimal height)
// - Text nodes should be leaf nodes (children are ignored if present)
//
// Algorithm based on CSS Text Module Level 3:
// - §3: White Space Processing
// - §4: Line Breaking and Word Boundaries
// - §5: Text Transformations and Spacing
// - §7: Text Alignment
//
// Note: This implementation uses simplified algorithms for whitespace collapsing
// and line breaking. See TEXT_LAYOUT_ISSUES.md for details.
func LayoutText(node *Node, constraints Constraints) Size {
	// Validate text node invariants
	if len(node.Children) > 0 {
		// Text nodes should be leaf nodes. Children are ignored during text layout.
		// This is intentional: text layout only processes node.Text, not child nodes.
		// If you need mixed content, use a block container with text and block children.
	}

	if node.Style.TextStyle == nil {
		// Default TextStyle if not set
		node.Style.TextStyle = &TextStyle{
			FontSize:   16,
			TextAlign:  TextAlignDefault,
			LineHeight: 0, // normal
			WhiteSpace: WhiteSpaceNormal,
			Direction:  DirectionLTR,
		}
	}
	style := node.Style.TextStyle

	// 1. Determine available content width from constraints and Style (box sizing)
	availableWidth := constraints.MaxWidth
	horizontalPaddingBorder := node.Style.Padding.Left + node.Style.Padding.Right +
		node.Style.Border.Left + node.Style.Border.Right
	verticalPaddingBorder := node.Style.Padding.Top + node.Style.Padding.Bottom +
		node.Style.Border.Top + node.Style.Border.Bottom
	contentWidth := availableWidth - horizontalPaddingBorder
	if contentWidth < 0 {
		contentWidth = 0
	}

	// 2. Normalize white-space (§3.1)
	processedText := preprocessText(node.Text, style.WhiteSpace)

	// 3. Perform line breaking (§4) with textMetrics.Measure
	lines := breakIntoLines(processedText, contentWidth, *style)

	// 4. Compute per-line positions (x,y) based on text-align (§7.1) and text-indent (§7.2.1)
	lineHeight := resolveLineHeight(style.LineHeight, style.FontSize)
	positionLines(lines, contentWidth, style.TextAlign, style.TextIndent, lineHeight)

	// 5. Compute total height from line count and line-height (§4.4.1)
	// If no lines, use at least one line height for empty text
	numLines := len(lines)
	if numLines == 0 {
		numLines = 1
	}
	contentHeight := float64(numLines) * lineHeight

	// Find max line width (including text-indent for first line)
	maxLineWidth := 0.0
	for i, line := range lines {
		w := line.Width
		// Include text-indent in first line width calculation
		if i == 0 && style.TextIndent != 0 {
			w += style.TextIndent
		}
		if w > maxLineWidth {
			maxLineWidth = w
		}
	}

	// 6. Apply explicit width/height if set, using box-sizing conversion
	hasExplicitWidth := node.Style.Width > 0
	hasExplicitHeight := node.Style.Height > 0

	if hasExplicitWidth {
		// Convert from specified box-sizing to content-box
		contentWidth = convertToContentSize(node.Style.Width, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, true)
	} else {
		// Auto width: use max line width
		contentWidth = maxLineWidth
	}

	if hasExplicitHeight {
		// Convert from specified box-sizing to content-box
		contentHeight = convertToContentSize(node.Style.Height, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, false)
	}

	// Apply min/max constraints (convert to content-box)
	minWidthContent := convertMinMaxToContentSize(node.Style.MinWidth, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, true)
	maxWidthContent := convertMinMaxToContentSize(node.Style.MaxWidth, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, true)
	minHeightContent := convertMinMaxToContentSize(node.Style.MinHeight, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, false)
	maxHeightContent := convertMinMaxToContentSize(node.Style.MaxHeight, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, false)

	// Clamp content dimensions to min/max
	if minWidthContent > 0 && contentWidth < minWidthContent {
		contentWidth = minWidthContent
	}
	if maxWidthContent > 0 && maxWidthContent < Unbounded && contentWidth > maxWidthContent {
		contentWidth = maxWidthContent
	}

	if minHeightContent > 0 && contentHeight < minHeightContent {
		contentHeight = minHeightContent
	}
	if maxHeightContent > 0 && maxHeightContent < Unbounded && contentHeight > maxHeightContent {
		contentHeight = maxHeightContent
	}

	outerWidth := contentWidth + horizontalPaddingBorder
	outerHeight := contentHeight + verticalPaddingBorder

	// Constrain and set Rect
	size := constraints.Constrain(Size{Width: outerWidth, Height: outerHeight})
	node.Rect.Width = size.Width
	node.Rect.Height = size.Height

	// 7. Store line metadata for rendering
	node.TextLayout = &TextLayout{
		Lines:      lines,
		LineHeight: lineHeight,
	}

	return size
}

// preprocessText normalizes text based on white-space property.
// Based on CSS Text Module Level 3 §3.1: https://www.w3.org/TR/css-text-3/#white-space-property
func preprocessText(text string, whiteSpace WhiteSpace) string {
	switch whiteSpace {
	case WhiteSpaceNormal:
		// §3.1: Collapse runs of spaces/tabs to single space
		// Turn \n and \r\n into spaces
		text = strings.ReplaceAll(text, "\r\n", " ")
		text = strings.ReplaceAll(text, "\n", " ")
		text = strings.ReplaceAll(text, "\r", " ")

		// Collapse whitespace, but preserve non-breaking spaces (U+00A0)
		// Non-breaking spaces should not collapse per CSS spec
		text = collapseWhitespace(text)

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

// collapseWhitespace collapses sequences of whitespace to single spaces,
// but preserves non-breaking spaces (U+00A0) as per CSS spec.
// Based on CSS Text Module Level 3 §3.1: https://www.w3.org/TR/css-text-3/#white-space-property
func collapseWhitespace(text string) string {
	if text == "" {
		return text
	}

	var result strings.Builder
	result.Grow(len(text)) // Pre-allocate capacity

	runes := []rune(text)
	inWhitespace := false

	for i, r := range runes {
		isNBSP := r == '\u00A0' // Non-breaking space (U+00A0)
		// Use unicode.IsSpace to check for whitespace, but exclude NBSP
		isWhitespace := unicode.IsSpace(r) && !isNBSP

		if isNBSP {
			// Non-breaking space: preserve as-is, don't collapse
			if inWhitespace {
				// End previous whitespace run with a space
				result.WriteRune(' ')
				inWhitespace = false
			}
			result.WriteRune(r)
		} else if isWhitespace {
			// Regular whitespace: mark that we're in a whitespace run
			if !inWhitespace {
				// Start of whitespace run - we'll collapse to single space later
				inWhitespace = true
			}
		} else {
			// Non-whitespace character
			if inWhitespace {
				// End whitespace run with a single space
				result.WriteRune(' ')
				inWhitespace = false
			}
			result.WriteRune(r)
		}

		// Handle trailing whitespace at end of string
		if i == len(runes)-1 && inWhitespace {
			// Trailing whitespace will be trimmed by TrimSpace in caller
		}
	}

	// If we ended in whitespace, add one space (will be trimmed if trailing)
	if inWhitespace {
		result.WriteRune(' ')
	}

	return result.String()
}

// splitIntoWords splits text into words using Unicode grapheme clusters and line breaking rules.
// Uses uniseg package for proper Unicode grapheme cluster handling (UAX #29),
// preserving non-breaking spaces and handling complex Unicode text correctly.
// Based on CSS Text Module Level 3 §4: https://www.w3.org/TR/css-text-3/#line-breaking
func splitIntoWords(text string) []string {
	if text == "" {
		return []string{}
	}

	var words []string
	var current strings.Builder
	gr := uniseg.NewGraphemes(text)

	for gr.Next() {
		cluster := gr.Str()
		runes := []rune(cluster)

		// Check if this is a non-breaking space or regular whitespace
		isNBSP := len(runes) == 1 && runes[0] == '\u00A0'
		isWhitespace := false
		if len(runes) == 1 {
			r := runes[0]
			// Use unicode.IsSpace for proper Unicode whitespace detection
			isWhitespace = unicode.IsSpace(r) && !isNBSP
		}

		if isWhitespace {
			// Regular whitespace: end current word
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			// Skip the whitespace (it's already collapsed)
		} else {
			// Non-whitespace or NBSP: add to current word
			// Using grapheme clusters ensures emojis and combining characters stay together
			current.WriteString(cluster)
		}
	}

	// Handle final word
	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

// breakIntoLines breaks text into lines based on available width using UAX #14.
// Based on CSS Text Module Level 3 §4: https://www.w3.org/TR/css-text-3/#line-breaking
// Uses Unicode Line Breaking Algorithm (UAX #14) for proper break opportunities.
func breakIntoLines(text string, maxWidth float64, style TextStyle) []TextLine {
	if text == "" {
		return []TextLine{}
	}

	// Treat maxWidth <= 0 as unbounded (no wrapping)
	if maxWidth <= 0 {
		maxWidth = Unbounded
	}

	// For pre mode, split on newlines first
	if style.WhiteSpace == WhiteSpacePre {
		return breakIntoLinesPre(text, maxWidth, style)
	}

	// Use UAX #14 to find line break opportunities
	return breakIntoLinesUAX14(text, maxWidth, style)
}

// breakIntoLinesUAX14 breaks text into lines using UAX #14 line breaking algorithm.
func breakIntoLinesUAX14(text string, maxWidth float64, style TextStyle) []TextLine {
	// Find all line break opportunities using UAX #14
	breakPoints := findLineBreakOpportunities(text)
	if len(breakPoints) < 2 {
		return []TextLine{}
	}

	lines := []TextLine{}
	current := TextLine{Boxes: []InlineBox{}}
	currentWidth := 0.0

	// First line gets text-indent
	firstLineIndent := style.TextIndent
	if firstLineIndent < 0 {
		// Negative indent is allowed
	}

	// Process text segment by segment
	for i := 0; i < len(breakPoints)-1; i++ {
		start := breakPoints[i]
		end := breakPoints[i+1]
		segment := text[start:end]

		// Skip empty segments
		if len(segment) == 0 {
			continue
		}

		// Measure segment
		segmentWidth, ascent, descent := textMetrics.Measure(segment, style)

		// Check if we need to break BEFORE adding this segment
		effectiveLineWidth := currentWidth
		if len(current.Boxes) == 0 && firstLineIndent != 0 {
			effectiveLineWidth += firstLineIndent
		}

		// Add space before segment if not first segment in line (for word spacing)
		if len(current.Boxes) > 0 {
			spaceWidth, _, _ := textMetrics.Measure(" ", style)
			if style.WordSpacing != -1 {
				spaceWidth += style.WordSpacing
			}
			effectiveLineWidth += spaceWidth
		}
		effectiveLineWidth += segmentWidth

		// Break if this segment would exceed maxWidth (and we have content already on this line)
		if maxWidth > 0 && maxWidth < Unbounded && effectiveLineWidth > maxWidth && len(current.Boxes) > 0 && canBreakBefore(style.WhiteSpace) {
			// Break line
			current.Width = currentWidth
			lines = append(lines, current)
			current = TextLine{Boxes: []InlineBox{}}
			currentWidth = 0.0
			firstLineIndent = 0.0 // Only first line gets indent
		}

		// Add space before segment if not first segment in line
		if len(current.Boxes) > 0 {
			spaceWidth, _, _ := textMetrics.Measure(" ", style)
			if style.WordSpacing != -1 {
				spaceWidth += style.WordSpacing
			}
			currentWidth += spaceWidth
		}

		// Add segment to current line
		box := InlineBox{
			Kind:    InlineBoxText,
			Text:    segment,
			Width:   segmentWidth,
			Ascent:  ascent,
			Descent: descent,
		}
		current.Boxes = append(current.Boxes, box)
		currentWidth += segmentWidth
	}

	// Add final line
	if len(current.Boxes) > 0 {
		current.Width = currentWidth
		lines = append(lines, current)
	}

	return lines
}

func canBreakBefore(whiteSpace WhiteSpace) bool {
	if whiteSpace == WhiteSpacePre {
		return false // §3.1: No wrapping in pre
	}
	if whiteSpace == WhiteSpaceNowrap {
		return false // §3.1: No wrapping in nowrap
	}
	return true // §4: Can break before word tokens in normal mode
}

// breakIntoLinesPre breaks text into lines preserving newlines and spaces (pre mode)
func breakIntoLinesPre(text string, maxWidth float64, style TextStyle) []TextLine {
	lines := []TextLine{}

	// Split by newlines
	lineTexts := strings.Split(text, "\n")
	for _, lineText := range lineTexts {
		line := TextLine{Boxes: []InlineBox{}}

		// Measure the entire line text (preserving all spaces)
		// Text-indent affects alignment, not intrinsic width, so handle in positionLines()
		advance, ascent, descent := textMetrics.Measure(lineText, style)
		line.Boxes = append(line.Boxes, InlineBox{
			Kind:    InlineBoxText,
			Text:    lineText,
			Width:   advance,
			Ascent:  ascent,
			Descent: descent,
		})
		line.Width = advance
		lines = append(lines, line)
	}

	return lines
}

// positionLines positions lines based on text-align and text-indent.
// Based on CSS Text Module Level 3 §7.1: https://www.w3.org/TR/css-text-3/#text-align-property
func positionLines(lines []TextLine, contentWidth float64, textAlign TextAlign, textIndent float64, lineHeight float64) {
	// Resolve TextAlignDefault to left (LTR context for v1)
	align := textAlign
	if align == TextAlignDefault {
		align = TextAlignLeft
	}

	currentY := 0.0
	for i := range lines {
		line := &lines[i]
		lineWidth := line.Width
		indent := 0.0

		// First line gets text-indent (can be positive or negative)
		if i == 0 && textIndent != 0 {
			indent = textIndent
		}

		// Calculate X offset based on text-align
		// Per CSS Text Module Level 3 §7.2.1: text-indent is treated as a margin
		// applied to the start edge (left in LTR)
		switch align {
		case TextAlignLeft:
			// Left-aligned: text starts at indent position
			line.OffsetX = indent

		case TextAlignRight:
			// Right-aligned: indent reduces available width, text aligns to (contentWidth - indent)
			// So text ends at (contentWidth - indent), not at contentWidth
			line.OffsetX = contentWidth - lineWidth - indent

		case TextAlignCenter:
			// Center-aligned: indent reduces available width, center within remaining space
			// Available width is (contentWidth - indent), center the line within that
			availableWidth := contentWidth - indent
			line.OffsetX = indent + (availableWidth-lineWidth)/2

		default:
			line.OffsetX = 0.0
		}

		// Set Y position
		line.OffsetY = currentY
		currentY += lineHeight
	}
}

// resolveLineHeight resolves line-height value to absolute pixels.
// Based on CSS Inline Layout Module Level 3 §4.4.1: https://www.w3.org/TR/css-inline-3/#propdef-line-height
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

// Text creates a new text node with the given text and optional style.
// The node will have DisplayInlineText set automatically.
func Text(text string, style ...Style) *Node {
	baseStyle := Style{
		Display: DisplayInlineText,
		Width:   -1, // auto
		Height:  -1, // auto
		TextStyle: &TextStyle{
			FontSize:   16,
			TextAlign:  TextAlignDefault,
			LineHeight: 0, // normal
			WhiteSpace: WhiteSpaceNormal,
			Direction:  DirectionLTR,
		},
	}

	node := &Node{
		Text:  text,
		Style: baseStyle,
	}

	// Merge provided style if any
	if len(style) > 0 {
		node.Style = style[0]
		node.Style.Display = DisplayInlineText
		// Treat 0 as auto (Go zero value issue)
		if node.Style.Width == 0 {
			node.Style.Width = -1
		}
		if node.Style.Height == 0 {
			node.Style.Height = -1
		}
		// Ensure TextStyle is set
		if node.Style.TextStyle == nil {
			node.Style.TextStyle = baseStyle.TextStyle
		}
	}

	return node
}
