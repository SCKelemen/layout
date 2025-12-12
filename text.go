package layout

import (
	"regexp"
	"strings"
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
	charWidth := style.FontSize * 0.6 // Rough average
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

	// 6. Set node.Rect.Width/Height (including padding/border)
	// Apply explicit width/height if set (>= 0 means explicit, < 0 means auto)
	// Note: 0 is a valid explicit value, so we check >= 0
	if node.Style.Width >= 0 {
		contentWidth = node.Style.Width
	}
	// For height, only override if explicitly set (Height > 0, since 0 could be unset)
	// Actually, we should use the same pattern as width: >= 0 means explicit
	// But we need to distinguish between unset (0) and explicit 0
	// For now, only override if > 0 to avoid the zero-value issue
	if node.Style.Height > 0 {
		contentHeight = node.Style.Height
	}

	// Find max line width
	maxLineWidth := 0.0
	for _, line := range lines {
		if line.Width > maxLineWidth {
			maxLineWidth = line.Width
		}
	}
	if node.Style.Width < 0 {
		contentWidth = maxLineWidth
	}

	verticalPaddingBorder := node.Style.Padding.Top + node.Style.Padding.Bottom +
		node.Style.Border.Top + node.Style.Border.Bottom

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

// breakIntoLines breaks text into lines based on available width.
// Based on CSS Text Module Level 3 §4: https://www.w3.org/TR/css-text-3/#line-breaking
func breakIntoLines(text string, maxWidth float64, style TextStyle) []TextLine {
	if text == "" {
		return []TextLine{}
	}

	// For pre mode, split on newlines first
	if style.WhiteSpace == WhiteSpacePre {
		return breakIntoLinesPre(text, maxWidth, style)
	}

	// Split into words (simple word-based breaking for v1)
	words := strings.Fields(text)
	if len(words) == 0 {
		return []TextLine{}
	}

	lines := []TextLine{}
	current := TextLine{Boxes: []InlineBox{}}
	currentWidth := 0.0

	// First line gets text-indent
	firstLineIndent := 0.0
	if style.TextIndent > 0 {
		firstLineIndent = style.TextIndent
	}

	for _, word := range words {
		// Measure word
		wordWidth, _, _ := textMetrics.Measure(word, style)

		// Check if we need to break BEFORE adding this word
		// Account for first line indent when checking width
		effectiveLineWidth := currentWidth
		if len(current.Boxes) == 0 && firstLineIndent > 0 {
			effectiveLineWidth += firstLineIndent
		}
		// Add space before word if not first word in line
		if len(current.Boxes) > 0 {
			spaceWidth, _, _ := textMetrics.Measure(" ", style)
			if style.WordSpacing > 0 {
				spaceWidth += style.WordSpacing
			}
			effectiveLineWidth += spaceWidth
		}
		effectiveLineWidth += wordWidth

		// Break if this word would exceed maxWidth (and we have words already on this line)
		if maxWidth > 0 && effectiveLineWidth > maxWidth && len(current.Boxes) > 0 && canBreakBefore(style.WhiteSpace) {
			// Break line
			current.Width = currentWidth
			lines = append(lines, current)
			current = TextLine{Boxes: []InlineBox{}}
			currentWidth = 0.0
			firstLineIndent = 0.0 // Only first line gets indent
		}

		// Add space before word if not first word in line
		if len(current.Boxes) > 0 {
			spaceWidth, _, _ := textMetrics.Measure(" ", style)
			if style.WordSpacing > 0 {
				spaceWidth += style.WordSpacing
			}
			currentWidth += spaceWidth
		}

		// Add word to current line
		_, ascent, descent := textMetrics.Measure(word, style)
		box := InlineBox{
			Kind:    InlineBoxText,
			Text:    word,
			Width:   wordWidth,
			Ascent:  ascent,
			Descent: descent,
		}
		current.Boxes = append(current.Boxes, box)
		currentWidth += wordWidth
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

// breakIntoLinesPre breaks text into lines preserving newlines (pre mode)
func breakIntoLinesPre(text string, maxWidth float64, style TextStyle) []TextLine {
	lines := []TextLine{}

	// Split by newlines
	lineTexts := strings.Split(text, "\n")
	for lineIdx, lineText := range lineTexts {
		// Each line becomes a TextLine
		line := TextLine{Boxes: []InlineBox{}}
		lineWidth := 0.0

		// First line gets text-indent
		if lineIdx == 0 && style.TextIndent > 0 {
			lineWidth += style.TextIndent
		}

		// Split line into words (preserving spaces would require more complex handling)
		words := strings.Fields(lineText)
		for i, word := range words {
			wordWidth, _, _ := textMetrics.Measure(word, style)

			// Add space before word if not first
			if i > 0 {
				spaceWidth, _, _ := textMetrics.Measure(" ", style)
				if style.WordSpacing > 0 {
					spaceWidth += style.WordSpacing
				}
				lineWidth += spaceWidth
			}

			_, ascent, descent := textMetrics.Measure(word, style)
			box := InlineBox{
				Kind:    InlineBoxText,
				Text:    word,
				Width:   wordWidth,
				Ascent:  ascent,
				Descent: descent,
			}
			line.Boxes = append(line.Boxes, box)
			lineWidth += wordWidth
		}

		line.Width = lineWidth
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

		// First line gets text-indent
		if i == 0 && textIndent > 0 {
			lineWidth += textIndent
		}

		// Calculate X offset based on text-align
		switch align {
		case TextAlignLeft:
			line.OffsetX = 0.0
			if i == 0 && textIndent > 0 {
				line.OffsetX = textIndent
			}

		case TextAlignRight:
			line.OffsetX = contentWidth - lineWidth

		case TextAlignCenter:
			line.OffsetX = (contentWidth - lineWidth) / 2

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
	node := &Node{
		Text: text,
		Style: Style{
			Display: DisplayInlineText,
			TextStyle: &TextStyle{
				FontSize:   16,
				TextAlign:  TextAlignDefault,
				LineHeight: 0, // normal
				WhiteSpace: WhiteSpaceNormal,
				Direction:  DirectionLTR,
			},
		},
	}

	// Merge provided style if any
	if len(style) > 0 {
		node.Style = style[0]
		node.Style.Display = DisplayInlineText
		// Ensure TextStyle is set
		if node.Style.TextStyle == nil {
			node.Style.TextStyle = &TextStyle{
				FontSize:   16,
				TextAlign:  TextAlignDefault,
				LineHeight: 0,
				WhiteSpace: WhiteSpaceNormal,
				Direction:  DirectionLTR,
			}
		}
	}

	return node
}
