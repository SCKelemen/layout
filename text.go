package layout

import (
	"strings"
	"unicode"

	"github.com/SCKelemen/unicode/uax50"
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
func LayoutText(node *Node, constraints Constraints, ctx *LayoutContext) Size {
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

	// Get writing mode from TextStyle (defaults to horizontal-tb)
	writingMode := style.WritingMode

	// Get current font size for em unit resolution
	currentFontSize := 16.0 // Default
	if style.FontSize > 0 {
		currentFontSize = style.FontSize
	}

	// 1. Determine available content size from constraints and Style (box sizing)
	// In vertical modes, we work with inline dimension (height) instead of width
	// TODO(vertical): Full vertical text layout implementation with proper line breaking
	// For now, vertical modes fall back to horizontal logic
	availableWidth := constraints.MaxWidth
	if writingMode.IsVertical() {
		// Vertical mode: inline dimension is height
		availableWidth = constraints.MaxHeight
	}

	// Resolve padding and border Length values to pixels
	paddingLeft := ResolveLength(node.Style.Padding.Left, ctx, currentFontSize)
	paddingRight := ResolveLength(node.Style.Padding.Right, ctx, currentFontSize)
	paddingTop := ResolveLength(node.Style.Padding.Top, ctx, currentFontSize)
	paddingBottom := ResolveLength(node.Style.Padding.Bottom, ctx, currentFontSize)
	borderLeft := ResolveLength(node.Style.Border.Left, ctx, currentFontSize)
	borderRight := ResolveLength(node.Style.Border.Right, ctx, currentFontSize)
	borderTop := ResolveLength(node.Style.Border.Top, ctx, currentFontSize)
	borderBottom := ResolveLength(node.Style.Border.Bottom, ctx, currentFontSize)

	horizontalPaddingBorder := paddingLeft + paddingRight + borderLeft + borderRight
	verticalPaddingBorder := paddingTop + paddingBottom + borderTop + borderBottom
	contentWidth := availableWidth - horizontalPaddingBorder
	if contentWidth < 0 {
		contentWidth = 0
	}

	// 2. Expand tabs based on tab-size (§3.1.1) - BEFORE whitespace processing
	// Only expand tabs for normal and nowrap modes; pre modes preserve tabs
	processedText := node.Text
	if style.WhiteSpace == WhiteSpaceNormal || style.WhiteSpace == WhiteSpaceNowrap {
		processedText = expandTabs(processedText, style.TabSize)
	}

	// 2.5. Normalize white-space (§3.1)
	processedText = preprocessText(processedText, style.WhiteSpace)

	// 2.6. Apply text-transform (§6)
	processedText = applyTextTransform(processedText, style.TextTransform)

	// 3. Perform line breaking (§4) with textMetrics.Measure
	lines := breakIntoLines(processedText, contentWidth, *style)

	// 3.5. Apply text-overflow if needed (ellipsis truncation)
	// CSS Text Overflow Module Level 3: https://www.w3.org/TR/css-overflow-3/#text-overflow
	if style.TextOverflow == TextOverflowEllipsis {
		lines = applyTextOverflow(lines, contentWidth, *style)
	}

	// 4. Compute per-line positions (x,y) based on text-align (§7.1), text-align-last (§7.2.2), text-justify (§7.3), text-indent (§7.2.1), and direction (§2)
	lineHeight := resolveLineHeight(style.LineHeight, style.FontSize)
	positionLines(lines, contentWidth, style.TextAlign, style.TextAlignLast, style.TextJustify, style.TextIndent, style.Direction, lineHeight)

	// 4.5. Apply hanging-punctuation (§9.2)
	applyHangingPunctuation(lines, style.HangingPunctuation, *style)

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
	// Resolve Length values to pixels first
	widthPx := ResolveLength(node.Style.Width, ctx, currentFontSize)
	heightPx := ResolveLength(node.Style.Height, ctx, currentFontSize)
	hasExplicitWidth := widthPx > 0
	hasExplicitHeight := heightPx > 0

	if hasExplicitWidth {
		// Convert from specified box-sizing to content-box
		contentWidth = convertToContentSize(widthPx, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, true)
	} else {
		// Auto width: use max line width
		contentWidth = maxLineWidth
	}

	if hasExplicitHeight {
		// Convert from specified box-sizing to content-box
		contentHeight = convertToContentSize(heightPx, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, false)
	}

	// Apply min/max constraints (convert to content-box)
	// Resolve min/max Length values to pixels
	minWidthPx := ResolveLength(node.Style.MinWidth, ctx, currentFontSize)
	maxWidthPx := ResolveLength(node.Style.MaxWidth, ctx, currentFontSize)
	minHeightPx := ResolveLength(node.Style.MinHeight, ctx, currentFontSize)
	maxHeightPx := ResolveLength(node.Style.MaxHeight, ctx, currentFontSize)

	minWidthContent := convertMinMaxToContentSize(minWidthPx, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, true)
	maxWidthContent := convertMinMaxToContentSize(maxWidthPx, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, true)
	minHeightContent := convertMinMaxToContentSize(minHeightPx, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, false)
	maxHeightContent := convertMinMaxToContentSize(maxHeightPx, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, false)

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

	case WhiteSpacePreWrap:
		// §3.1: Preserve all whitespace, allow wrapping
		// Don't collapse anything, don't convert newlines
		return text

	case WhiteSpacePreLine:
		// §3.1: Preserve newlines, collapse spaces, allow wrapping
		// Normalize line endings but DON'T convert to spaces
		text = strings.ReplaceAll(text, "\r\n", "\n")
		text = strings.ReplaceAll(text, "\r", "\n")

		// Collapse whitespace on each line separately
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimSpace(collapseWhitespace(line))
		}
		return strings.Join(lines, "\n")

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

// applyTextTransform applies text-transform property
// CSS Text Module Level 3 §6: https://www.w3.org/TR/css-text-3/#text-transform-property
func applyTextTransform(text string, transform TextTransform) string {
	switch transform {
	case TextTransformNone:
		return text

	case TextTransformUppercase:
		return strings.ToUpper(text)

	case TextTransformLowercase:
		return strings.ToLower(text)

	case TextTransformCapitalize:
		// Capitalize first letter of each word
		// A "word" is defined as a sequence of non-space characters
		runes := []rune(text)
		var result strings.Builder
		result.Grow(len(runes))

		capitalizeNext := true
		for _, r := range runes {
			if unicode.IsSpace(r) {
				result.WriteRune(r)
				capitalizeNext = true
			} else {
				if capitalizeNext {
					result.WriteRune(unicode.ToUpper(r))
					capitalizeNext = false
				} else {
					result.WriteRune(r)
				}
			}
		}
		return result.String()

	case TextTransformFullWidth:
		// Convert half-width characters to full-width
		// This is primarily for CJK text
		runes := []rune(text)
		var result strings.Builder
		result.Grow(len(runes) * 2)

		for _, r := range runes {
			// ASCII characters (0x21-0x7E) map to full-width (0xFF01-0xFF5E)
			if r >= 0x21 && r <= 0x7E {
				result.WriteRune(r - 0x21 + 0xFF01)
			} else if r == 0x20 { // Space maps to ideographic space
				result.WriteRune(0x3000)
			} else {
				result.WriteRune(r)
			}
		}
		return result.String()

	case TextTransformFullSizeKana:
		// Convert half-width katakana to full-width
		// This is a simplified implementation
		runes := []rune(text)
		var result strings.Builder
		result.Grow(len(runes) * 2)

		for _, r := range runes {
			// Half-width katakana (0xFF65-0xFF9F) map to full-width (0x30A1-0x30FD)
			// This is a simplified mapping; full implementation would need a lookup table
			if r >= 0xFF65 && r <= 0xFF9F {
				// Approximate mapping (not complete)
				result.WriteRune(r - 0xFF65 + 0x30A1)
			} else {
				result.WriteRune(r)
			}
		}
		return result.String()

	default:
		return text
	}
}

// expandTabs replaces tab characters with spaces based on tab-size
// CSS Text Module Level 3 §3.1.1: https://www.w3.org/TR/css-text-3/#tab-size-property
func expandTabs(text string, tabSize float64) string {
	if !strings.Contains(text, "\t") {
		return text
	}

	// Default tab size is 8 spaces
	if tabSize < 0 {
		tabSize = 8
	}

	// Convert tabSize to integer number of spaces
	numSpaces := int(tabSize)
	if numSpaces < 1 {
		numSpaces = 1
	}

	replacement := strings.Repeat(" ", numSpaces)
	return strings.ReplaceAll(text, "\t", replacement)
}

// isOpeningPunctuation checks if a rune is opening punctuation
func isOpeningPunctuation(r rune) bool {
	// Opening brackets, quotes, etc.
	return r == '(' || r == '[' || r == '{' || r == '<' ||
		r == '"' || r == '\'' || r == '\u201C' || r == '\u2018' || // Left double/single quotes
		r == '\u00AB' || r == '\u2039' // Left guillemets
}

// isClosingPunctuation checks if a rune is closing punctuation
func isClosingPunctuation(r rune) bool {
	// Closing brackets, quotes, periods, commas, etc.
	return r == ')' || r == ']' || r == '}' || r == '>' ||
		r == '"' || r == '\'' || r == '\u201D' || r == '\u2019' || // Right double/single quotes
		r == '\u00BB' || r == '\u203A' || // Right guillemets
		r == '.' || r == ',' || r == '!' || r == '?' || r == ';' || r == ':'
}

// applyHangingPunctuation adjusts line boxes for hanging punctuation
// CSS Text Module Level 3 §9.2: https://www.w3.org/TR/css-text-3/#hanging-punctuation-property
func applyHangingPunctuation(lines []TextLine, hanging HangingPunctuation, style TextStyle) {
	if hanging == HangingPunctuationNone {
		return
	}

	for i := range lines {
		line := &lines[i]
		if len(line.Boxes) == 0 {
			continue
		}

		// Handle first punctuation (opening)
		if hanging == HangingPunctuationFirst || hanging == HangingPunctuationAllowEnd {
			firstBox := &line.Boxes[0]
			if len(firstBox.Text) > 0 {
				runes := []rune(firstBox.Text)
				if isOpeningPunctuation(runes[0]) {
					// Measure the punctuation character
					punctWidth, _, _ := textMetrics.Measure(string(runes[0]), style)
					// Hang it by moving line start position
					line.OffsetX -= punctWidth
					line.Width += punctWidth
				}
			}
		}

		// Handle last punctuation (closing)
		if hanging == HangingPunctuationLast || hanging == HangingPunctuationForceEnd || hanging == HangingPunctuationAllowEnd {
			lastBox := &line.Boxes[len(line.Boxes)-1]
			if len(lastBox.Text) > 0 {
				runes := []rune(lastBox.Text)
				if isClosingPunctuation(runes[len(runes)-1]) {
					// Measure the punctuation character
					punctWidth, _, _ := textMetrics.Measure(string(runes[len(runes)-1]), style)
					// Hang it by extending line width beyond container
					line.Width -= punctWidth
				}
			}
		}
	}
}

// splitIntoWords splits text into words by breaking on whitespace.
// Preserves non-breaking spaces (U+00A0) and handles Unicode text correctly.
// Based on CSS Text Module Level 3 §4: https://www.w3.org/TR/css-text-3/#line-breaking
func splitIntoWords(text string) []string {
	if text == "" {
		return []string{}
	}

	var words []string
	var current strings.Builder

	for _, r := range text {
		// Check if this is a non-breaking space or regular whitespace
		isNBSP := r == '\u00A0'
		isWhitespace := unicode.IsSpace(r) && !isNBSP

		if isWhitespace {
			// Regular whitespace: end current word
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			// Skip the whitespace (it's already collapsed)
		} else {
			// Non-whitespace or NBSP: add to current word
			current.WriteRune(r)
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

	// For pre-wrap and pre-line, split on newlines then wrap each segment
	if style.WhiteSpace == WhiteSpacePreWrap || style.WhiteSpace == WhiteSpacePreLine {
		return breakIntoLinesPreWrap(text, maxWidth, style)
	}

	// Use UAX #14 to find line break opportunities
	return breakIntoLinesUAX14(text, maxWidth, style)
}

// breakIntoLinesUAX14 breaks text into lines using UAX #14 line breaking algorithm.
func breakIntoLinesUAX14(text string, maxWidth float64, style TextStyle) []TextLine {
	// Find all line break opportunities using UAX #14, respecting hyphens property
	breakPoints := findLineBreakOpportunitiesWithHyphens(text, style.Hyphens)
	if len(breakPoints) < 2 {
		return []TextLine{}
	}

	lines := []TextLine{}
	current := TextLine{
		Boxes:      []InlineBox{},
		SpaceCount: 0,
		SpaceWidth: 0.0,
	}
	currentWidth := 0.0
	lastWordHadTrailingSpace := false // Track if last word had a trailing space

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

		// Segments may include trailing spaces - check and handle separately
		hasTrailingSpace := len(segment) > 0 && segment[len(segment)-1] == ' '
		wordText := segment
		var spaceWidth float64

		if hasTrailingSpace {
			// Strip trailing space and measure it separately
			wordText = segment[:len(segment)-1]
			spaceWidth, _, _ = textMetrics.Measure(" ", style)
			if style.WordSpacing != -1 {
				spaceWidth += style.WordSpacing
			}
		}

		// Skip if word is empty (segment was just a space)
		if len(wordText) == 0 {
			continue
		}

		// Measure the word (without trailing space)
		wordWidth, ascent, descent := textMetrics.Measure(wordText, style)

		// Check if we need to break BEFORE adding this word
		effectiveLineWidth := currentWidth
		if len(current.Boxes) == 0 && firstLineIndent != 0 {
			effectiveLineWidth += firstLineIndent
		}

		// Add this word's width (space from previous word is already in currentWidth)
		effectiveLineWidth += wordWidth
		if hasTrailingSpace {
			effectiveLineWidth += spaceWidth
		}

		// Break if this word would exceed maxWidth (and we have content already on this line)
		if maxWidth > 0 && maxWidth < Unbounded && effectiveLineWidth > maxWidth && len(current.Boxes) > 0 && canBreakBefore(style.WhiteSpace) {
			// Remove trailing space from line end if last word had one (not used for justification)
			if lastWordHadTrailingSpace && current.SpaceCount > 0 {
				// Get the last space width
				lastSpaceWidth := current.SpaceWidth / float64(current.SpaceCount)
				current.Width = currentWidth - lastSpaceWidth
				current.SpaceCount--
				current.SpaceWidth -= lastSpaceWidth
			} else {
				current.Width = currentWidth
			}

			lines = append(lines, current)
			current = TextLine{
				Boxes:      []InlineBox{},
				SpaceCount: 0,
				SpaceWidth: 0.0,
			}
			currentWidth = 0.0
			lastWordHadTrailingSpace = false
			firstLineIndent = 0.0 // Only first line gets indent
		}

		// Check if word is too long and should be broken (overflow-wrap or word-break)
		// Only break if it's the first word on line and exceeds maxWidth
		if len(current.Boxes) == 0 && maxWidth > 0 && maxWidth < Unbounded && wordWidth > maxWidth {
			if style.OverflowWrap == OverflowWrapBreakWord || style.OverflowWrap == OverflowWrapAnywhere ||
				style.WordBreak == WordBreakBreakAll {
				// Break word into smaller pieces
				pieces := breakWordToFit(wordText, maxWidth, style)
				for j, piece := range pieces {
					if j > 0 {
						// Start new line for subsequent pieces
						current.Width = currentWidth
						lines = append(lines, current)
						current = TextLine{
							Boxes:      []InlineBox{},
							SpaceCount: 0,
							SpaceWidth: 0.0,
						}
						currentWidth = 0.0
						lastWordHadTrailingSpace = false
					}

					pieceWidth, ascent, descent := textMetrics.Measure(piece, style)
					current.Boxes = append(current.Boxes, InlineBox{
						Kind:    InlineBoxText,
						Text:    piece,
						Width:   pieceWidth,
						Ascent:  ascent,
						Descent: descent,
					})
					currentWidth += pieceWidth
				}

				// Handle trailing space if word had one
				if hasTrailingSpace {
					current.SpaceCount++
					current.SpaceWidth += spaceWidth
					currentWidth += spaceWidth
					lastWordHadTrailingSpace = true
				} else {
					lastWordHadTrailingSpace = false
				}

				continue // Skip normal word addition
			}
		}

		// Add the word to current line
		box := InlineBox{
			Kind:    InlineBoxText,
			Text:    wordText,
			Width:   wordWidth,
			Ascent:  ascent,
			Descent: descent,
		}
		current.Boxes = append(current.Boxes, box)
		currentWidth += wordWidth

		// Track space after this word (if it has one)
		if hasTrailingSpace {
			current.SpaceCount++
			current.SpaceWidth += spaceWidth
			currentWidth += spaceWidth
			lastWordHadTrailingSpace = true
		} else {
			lastWordHadTrailingSpace = false
		}
	}

	// Add final line
	if len(current.Boxes) > 0 {
		// Remove trailing space from line end if last word had one (not used for justification)
		if lastWordHadTrailingSpace && current.SpaceCount > 0 {
			// Get the last space width
			lastSpaceWidth := current.SpaceWidth / float64(current.SpaceCount)
			current.Width = currentWidth - lastSpaceWidth
			current.SpaceCount--
			current.SpaceWidth -= lastSpaceWidth
		} else {
			current.Width = currentWidth
		}
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
	// §3.1: pre-wrap, pre-line, and normal modes all allow wrapping
	return true
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

// breakIntoLinesPreWrap handles pre-wrap and pre-line modes
// Split on newlines, then wrap each segment
func breakIntoLinesPreWrap(text string, maxWidth float64, style TextStyle) []TextLine {
	lines := []TextLine{}

	// Split by newlines
	segments := strings.Split(text, "\n")

	for _, segment := range segments {
		if segment == "" {
			// Empty line from consecutive newlines or trailing newline
			lines = append(lines, TextLine{
				Boxes: []InlineBox{},
				Width: 0,
			})
			continue
		}

		// Wrap this segment if it exceeds maxWidth
		// For pre-wrap: preserve spaces within the segment
		// For pre-line: spaces already collapsed in preprocessText
		segmentLines := wrapSegment(segment, maxWidth, style)
		lines = append(lines, segmentLines...)
	}

	return lines
}

// wrapSegment wraps a single segment (between newlines) with preserved spaces
func wrapSegment(segment string, maxWidth float64, style TextStyle) []TextLine {
	// If unlimited width or segment fits, return as single line
	segmentWidth, ascent, descent := textMetrics.Measure(segment, style)

	if maxWidth >= Unbounded || segmentWidth <= maxWidth {
		return []TextLine{{
			Boxes: []InlineBox{{
				Kind:    InlineBoxText,
				Text:    segment,
				Width:   segmentWidth,
				Ascent:  ascent,
				Descent: descent,
			}},
			Width: segmentWidth,
		}}
	}

	// Need to wrap
	// For pre-wrap mode, preserve all spaces including multiple consecutive ones
	if style.WhiteSpace == WhiteSpacePreWrap {
		return wrapSegmentPreserveSpaces(segment, maxWidth, style)
	}

	// For pre-line, use UAX #14 (spaces already collapsed in preprocessText)
	return breakIntoLinesUAX14(segment, maxWidth, style)
}

// wrapSegmentPreserveSpaces wraps text while preserving all spaces (for pre-wrap mode)
func wrapSegmentPreserveSpaces(segment string, maxWidth float64, style TextStyle) []TextLine {
	lines := []TextLine{}
	current := TextLine{Boxes: []InlineBox{}}
	currentWidth := 0.0

	// Build words with preserved spaces by splitting on grapheme boundaries
	// We need to track characters and spaces separately
	runes := []rune(segment)
	wordStart := 0

	for i := 0; i < len(runes); i++ {
		// Find next space or end
		if runes[i] == ' ' || i == len(runes)-1 {
			// Extract word (include trailing char if at end and not space)
			wordEnd := i
			if i == len(runes)-1 && runes[i] != ' ' {
				wordEnd = i + 1
			}

			if wordEnd > wordStart {
				word := string(runes[wordStart:wordEnd])
				wordWidth, ascent, descent := textMetrics.Measure(word, style)

				// Check if adding this word would exceed maxWidth
				if currentWidth > 0 && currentWidth+wordWidth > maxWidth {
					// Start new line
					current.Width = currentWidth
					lines = append(lines, current)
					current = TextLine{Boxes: []InlineBox{}}
					currentWidth = 0.0
				}

				// Add word to current line
				current.Boxes = append(current.Boxes, InlineBox{
					Kind:    InlineBoxText,
					Text:    word,
					Width:   wordWidth,
					Ascent:  ascent,
					Descent: descent,
				})
				currentWidth += wordWidth
			}

			// If current char is a space, add it
			if runes[i] == ' ' {
				spaceWidth, ascent, descent := textMetrics.Measure(" ", style)

				// Check if space fits on current line
				if currentWidth+spaceWidth > maxWidth && currentWidth > 0 {
					// Start new line
					current.Width = currentWidth
					lines = append(lines, current)
					current = TextLine{Boxes: []InlineBox{}}
					currentWidth = 0.0
				}

				// Add space
				current.Boxes = append(current.Boxes, InlineBox{
					Kind:    InlineBoxText,
					Text:    " ",
					Width:   spaceWidth,
					Ascent:  ascent,
					Descent: descent,
				})
				currentWidth += spaceWidth
			}

			wordStart = i + 1
		}
	}

	// Add final line if not empty
	if len(current.Boxes) > 0 {
		current.Width = currentWidth
		lines = append(lines, current)
	}

	return lines
}

// breakWordToFit breaks a word into pieces that fit maxWidth
// Used for overflow-wrap: break-word and word-break: break-all
func breakWordToFit(word string, maxWidth float64, style TextStyle) []string {
	pieces := []string{}
	runes := []rune(word)

	currentPiece := strings.Builder{}
	currentWidth := 0.0

	for _, r := range runes {
		charStr := string(r)
		charWidth, _, _ := textMetrics.Measure(charStr, style)

		if currentWidth+charWidth > maxWidth && currentPiece.Len() > 0 {
			// Finish current piece
			pieces = append(pieces, currentPiece.String())
			currentPiece.Reset()
			currentWidth = 0.0
		}

		currentPiece.WriteRune(r)
		currentWidth += charWidth
	}

	if currentPiece.Len() > 0 {
		pieces = append(pieces, currentPiece.String())
	}

	return pieces
}

// applyTextOverflow applies text-overflow: ellipsis to overflowing lines
// CSS Text Overflow Module Level 3: https://www.w3.org/TR/css-overflow-3/#text-overflow
func applyTextOverflow(lines []TextLine, contentWidth float64, style TextStyle) []TextLine {
	if len(lines) == 0 {
		return lines
	}

	// Measure ellipsis width
	ellipsisText := "..."
	ellipsisWidth, ellipsisAscent, ellipsisDescent := textMetrics.Measure(ellipsisText, style)

	// Process each line that overflows
	for i := range lines {
		line := &lines[i]

		// Check if this line overflows
		if line.Width <= contentWidth {
			continue // No overflow, no truncation needed
		}

		// Line overflows - need to truncate and add ellipsis
		availableWidth := contentWidth - ellipsisWidth
		if availableWidth <= 0 {
			// Not enough space even for ellipsis - just show ellipsis
			line.Boxes = []InlineBox{{
				Kind:    InlineBoxText,
				Text:    ellipsisText,
				Width:   ellipsisWidth,
				Ascent:  ellipsisAscent,
				Descent: ellipsisDescent,
			}}
			line.Width = ellipsisWidth
			line.SpaceCount = 0
			line.SpaceWidth = 0
			line.SpaceAdjustment = 0
			continue
		}

		// Truncate boxes to fit within availableWidth
		truncatedBoxes := []InlineBox{}
		currentWidth := 0.0

		for _, box := range line.Boxes {
			if currentWidth+box.Width <= availableWidth {
				// Box fits completely
				truncatedBoxes = append(truncatedBoxes, box)
				currentWidth += box.Width
			} else {
				// Box would overflow - truncate it
				remainingWidth := availableWidth - currentWidth
				if remainingWidth > 0 {
					// Try to fit part of this box
					truncatedText := truncateTextToWidth(box.Text, remainingWidth, style)
					if truncatedText != "" {
						truncWidth, truncAscent, truncDesc := textMetrics.Measure(truncatedText, style)
						truncatedBoxes = append(truncatedBoxes, InlineBox{
							Kind:    box.Kind,
							Text:    truncatedText,
							Width:   truncWidth,
							Ascent:  truncAscent,
							Descent: truncDesc,
						})
						currentWidth += truncWidth
					}
				}
				break // Stop processing boxes
			}
		}

		// Add ellipsis
		truncatedBoxes = append(truncatedBoxes, InlineBox{
			Kind:    InlineBoxText,
			Text:    ellipsisText,
			Width:   ellipsisWidth,
			Ascent:  ellipsisAscent,
			Descent: ellipsisDescent,
		})

		line.Boxes = truncatedBoxes
		line.Width = currentWidth + ellipsisWidth
		line.SpaceCount = 0 // Reset space tracking for truncated line
		line.SpaceWidth = 0
		line.SpaceAdjustment = 0
	}

	return lines
}

// truncateTextToWidth truncates text to fit within maxWidth
func truncateTextToWidth(text string, maxWidth float64, style TextStyle) string {
	runes := []rune(text)

	// Binary search for the longest prefix that fits
	left, right := 0, len(runes)
	result := ""

	for left <= right {
		mid := (left + right) / 2
		candidate := string(runes[:mid])
		width, _, _ := textMetrics.Measure(candidate, style)

		if width <= maxWidth {
			result = candidate
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return result
}

// resolveTextAlignLast resolves text-align-last auto to actual alignment
// CSS Text Module Level 3 §7.2.2: https://www.w3.org/TR/css-text-3/#text-align-last-property
func resolveTextAlignLast(last TextAlignLast, textAlign TextAlign) TextAlignLast {
	if last != TextAlignLastAuto {
		return last
	}

	// Auto follows text-align, but never justify for last line
	switch textAlign {
	case TextAlignRight:
		return TextAlignLastRight
	case TextAlignCenter:
		return TextAlignLastCenter
	default:
		return TextAlignLastLeft
	}
}

// positionLines positions lines based on text-align, text-align-last, text-justify, and text-indent.
// Based on CSS Text Module Level 3 §7.1, §7.2.2, and §7.3
func positionLines(lines []TextLine, contentWidth float64, textAlign TextAlign, textAlignLast TextAlignLast, textJustify TextJustify, textIndent float64, direction Direction, lineHeight float64) {
	// Resolve TextAlignDefault based on direction
	align := textAlign
	wasDefault := (align == TextAlignDefault)

	if wasDefault {
		if direction == DirectionRTL {
			align = TextAlignRight // RTL defaults to right
		} else {
			align = TextAlignLeft // LTR defaults to left
		}
	}

	// For RTL with explicit left/right (not default), swap alignments
	if direction == DirectionRTL && !wasDefault {
		if align == TextAlignLeft {
			align = TextAlignRight
		} else if align == TextAlignRight {
			align = TextAlignLeft
		}
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

		case TextAlignJustify:
			// Justified: distribute extra space using text-justify algorithm
			// Per CSS Text Module Level 3 §7.1.1, §7.2.2, and §7.3
			isLastLine := (i == len(lines)-1)
			hasMultipleWords := line.SpaceCount > 0

			// Resolve text-justify
			justifyMode := textJustify
			if justifyMode == TextJustifyAuto {
				justifyMode = TextJustifyInterWord
			}

			// Handle text-justify: none
			if justifyMode == TextJustifyNone {
				line.OffsetX = indent
				continue
			}

			if !isLastLine && hasMultipleWords {
				// Middle lines: apply justification algorithm
				availableWidth := contentWidth
				if i == 0 && indent != 0 {
					availableWidth -= indent
				}

				extraSpace := availableWidth - lineWidth
				if extraSpace > 0 {
					switch justifyMode {
					case TextJustifyInterWord:
						// Distribute across word spaces only (current implementation)
						line.SpaceAdjustment = extraSpace / float64(line.SpaceCount)
						line.Width = availableWidth

					case TextJustifyInterCharacter, TextJustifyDistribute:
						// Distribute across both word spaces AND character gaps
						// Calculate total expansion opportunities
						totalChars := 0
						for _, box := range line.Boxes {
							totalChars += len([]rune(box.Text))
						}

						// Character gaps = total characters minus spaces between boxes
						characterGaps := totalChars - len(line.Boxes)
						if characterGaps < 0 {
							characterGaps = 0
						}

						// Total gaps = word spaces + character gaps
						totalGaps := line.SpaceCount + characterGaps

						if totalGaps > 0 {
							// Distribute evenly across all gaps
							gapAdjustment := extraSpace / float64(totalGaps)

							// Store both space and character adjustments
							line.SpaceAdjustment = gapAdjustment
							line.CharacterAdjustment = gapAdjustment
							line.Width = availableWidth
						} else if line.SpaceCount > 0 {
							// Fallback: if no character gaps, use inter-word
							line.SpaceAdjustment = extraSpace / float64(line.SpaceCount)
							line.Width = availableWidth
						}
					}
				}
				line.OffsetX = indent
			} else {
				// Last line or single word: use text-align-last
				lastAlign := resolveTextAlignLast(textAlignLast, align)

				switch lastAlign {
				case TextAlignLastLeft:
					line.OffsetX = indent
				case TextAlignLastRight:
					line.OffsetX = contentWidth - lineWidth
					if i == 0 {
						line.OffsetX -= indent
					}
				case TextAlignLastCenter:
					availableWidth := contentWidth
					if i == 0 {
						availableWidth -= indent
					}
					line.OffsetX = indent + (availableWidth-lineWidth)/2
				case TextAlignLastJustify:
					// Justify even last line
					if hasMultipleWords {
						availableWidth := contentWidth
						if i == 0 && indent != 0 {
							availableWidth -= indent
						}
						extraSpace := availableWidth - lineWidth
						if extraSpace > 0 {
							line.SpaceAdjustment = extraSpace / float64(line.SpaceCount)
							line.Width = availableWidth
						}
					}
					line.OffsetX = indent
				default:
					line.OffsetX = indent
				}
			}

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
		Width:   Px(0), // auto (Px(0) is treated as auto when resolved)
		Height:  Px(0), // auto
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
		// Ensure TextStyle is set
		if node.Style.TextStyle == nil {
			node.Style.TextStyle = baseStyle.TextStyle
		}
	}

	return node
}

// --- Vertical Writing Mode Helpers ---
// These functions abstract the dimension mapping for vertical writing modes.
// In vertical modes, the inline and block dimensions are swapped compared to horizontal modes.

// getInlineSize returns the inline dimension (the direction text flows within a line).
// For horizontal modes: inline = width
// For vertical modes: inline = height
func getInlineSize(width, height float64, wm WritingMode) float64 {
	if wm.IsVertical() {
		return height
	}
	return width
}

// getBlockSize returns the block dimension (the direction lines are stacked).
// For horizontal modes: block = height
// For vertical modes: block = width
func getBlockSize(width, height float64, wm WritingMode) float64 {
	if wm.IsVertical() {
		return width
	}
	return height
}

// getInlineConstraint returns the constraint for the inline dimension.
func getInlineConstraint(constraints Constraints, wm WritingMode) float64 {
	if wm.IsVertical() {
		return constraints.MaxHeight
	}
	return constraints.MaxWidth
}

// makeSize creates a Size with physical width/height from logical inline/block sizes.
func makeSize(inlineSize, blockSize float64, wm WritingMode) Size {
	if wm.IsVertical() {
		// Vertical: inline=height, block=width
		return Size{Width: blockSize, Height: inlineSize}
	}
	// Horizontal: inline=width, block=height
	return Size{Width: inlineSize, Height: blockSize}
}

// getInlinePaddingBorder returns padding+border in the inline dimension.
func getInlinePaddingBorder(paddingLeft, paddingRight, paddingTop, paddingBottom,
	borderLeft, borderRight, borderTop, borderBottom float64, wm WritingMode) float64 {
	if wm.IsVertical() {
		return paddingTop + paddingBottom + borderTop + borderBottom
	}
	return paddingLeft + paddingRight + borderLeft + borderRight
}

// getBlockPaddingBorder returns padding+border in the block dimension.
func getBlockPaddingBorder(paddingLeft, paddingRight, paddingTop, paddingBottom,
	borderLeft, borderRight, borderTop, borderBottom float64, wm WritingMode) float64 {
	if wm.IsVertical() {
		return paddingLeft + paddingRight + borderLeft + borderRight
	}
	return paddingTop + paddingBottom + borderTop + borderBottom
}

// getCharacterOrientation returns whether a character should be displayed upright or rotated
// in vertical text based on UAX #50 (Unicode Vertical Text Layout).
//
// Returns true if the character should be upright (CJK ideographs, kana, etc.),
// false if it should be rotated 90° clockwise (Latin, digits, etc.).
//
// For sideways modes, this function is not used as all characters are rotated.
func getCharacterOrientation(r rune, wm WritingMode) bool {
	// Sideways modes rotate all characters
	if wm.IsSideways() {
		return false
	}
	
	// For vertical-rl and vertical-lr, use UAX #50
	if wm == WritingModeVerticalRL || wm == WritingModeVerticalLR {
		return uax50.IsUpright(r)
	}
	
	// Horizontal modes don't rotate characters
	return true
}
