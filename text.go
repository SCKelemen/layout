package layout

import (
	"strings"
	"sync/atomic"
	"unicode"

	"github.com/SCKelemen/unicode/v6/uax50"
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

// Package-level provider.
//
// textMetrics is an atomic.Pointer to a textMetricsHolder. Using
// atomic.Pointer lets us safely swap the provider while concurrent goroutines
// are reading it during layout. All reads must go through getTextMetrics.
//
// atomic.Pointer needs a concrete pointee type, so the
// TextMetricsProvider interface is wrapped in textMetricsHolder.
var textMetrics atomic.Pointer[textMetricsHolder]

type textMetricsHolder struct {
	provider TextMetricsProvider
}

func init() {
	// Install the default approximate metrics so getTextMetrics never
	// returns a nil provider, even before SetTextMetricsProvider is
	// called.
	textMetrics.Store(&textMetricsHolder{provider: &approxMetrics{}})
}

// getTextMetrics returns the currently installed TextMetricsProvider.
// It performs a single atomic load and is safe for concurrent use.
func getTextMetrics() TextMetricsProvider {
	return textMetrics.Load().provider
}

// SetTextMetricsProvider installs a custom text measurement provider.
//
// The provider is consulted by the layout engine whenever it needs to
// measure a run of text (advance width, ascent, descent). Passing a nil
// provider is a no-op; the previously installed provider remains active.
//
// Thread safety: SetTextMetricsProvider is safe to call concurrently with
// other calls to itself and with concurrent layout operations. The
// provider is stored in an atomic.Pointer, so readers always observe a
// fully-initialized provider value.
//
// Even though installation is concurrency-safe, most callers should set
// the provider exactly once at program startup, for example from an
// init() function:
//
//	func init() {
//	    layout.SetTextMetricsProvider(myMetrics)
//	}
//
// This avoids the subtle situation where two concurrent layouts may use
// different providers and produce mismatched measurements.
func SetTextMetricsProvider(provider TextMetricsProvider) {
	if provider == nil {
		return
	}
	textMetrics.Store(&textMetricsHolder{provider: provider})
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

	// Get writing mode - prefer Style.WritingMode (inherited), fall back to TextStyle.WritingMode (legacy)
	writingMode := node.Style.WritingMode
	if writingMode == WritingModeHorizontalTB && style.WritingMode != WritingModeHorizontalTB {
		// Style.WritingMode is default but TextStyle.WritingMode is set, use TextStyle value for backward compat
		writingMode = style.WritingMode
	}

	// Get current font size for em unit resolution
	currentFontSize := 16.0 // Default
	if style.FontSize > 0 {
		currentFontSize = style.FontSize
	}

	// 1. Determine available content size from constraints and Style (box sizing)
	//
	// All sizing below is done in logical (inline/block) dimensions and mapped
	// back to physical width/height at the end. In horizontal writing modes the
	// inline axis is width; in vertical modes it is height.
	// https://www.w3.org/TR/css-writing-modes-3/#logical-to-physical
	isVertical := writingMode.IsVertical()

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
	inlinePaddingBorder := getInlinePaddingBorder(paddingLeft, paddingRight, paddingTop, paddingBottom,
		borderLeft, borderRight, borderTop, borderBottom, writingMode)
	blockPaddingBorder := getBlockPaddingBorder(paddingLeft, paddingRight, paddingTop, paddingBottom,
		borderLeft, borderRight, borderTop, borderBottom, writingMode)

	// Resolve the explicit inline/block sizes and min/max constraints up front.
	// Width/Height are physical; swap them for vertical writing modes.
	widthPx := ResolveLength(node.Style.Width, ctx, currentFontSize)
	heightPx := ResolveLength(node.Style.Height, ctx, currentFontSize)
	minWidthPx := ResolveLength(node.Style.MinWidth, ctx, currentFontSize)
	maxWidthPx := ResolveLength(node.Style.MaxWidth, ctx, currentFontSize)
	minHeightPx := ResolveLength(node.Style.MinHeight, ctx, currentFontSize)
	maxHeightPx := ResolveLength(node.Style.MaxHeight, ctx, currentFontSize)

	inlinePx, blockPx := widthPx, heightPx
	minInlinePx, maxInlinePx := minWidthPx, maxWidthPx
	minBlockPx, maxBlockPx := minHeightPx, maxHeightPx
	if isVertical {
		inlinePx, blockPx = heightPx, widthPx
		minInlinePx, maxInlinePx = minHeightPx, maxHeightPx
		minBlockPx, maxBlockPx = minWidthPx, maxWidthPx
	}
	// convertToContentSize's isWidth flag selects which padding/border sum to
	// subtract; the inline axis is the physical width only in horizontal modes.
	inlineIsWidth := !isVertical
	hasExplicitInline := inlinePx > 0
	hasExplicitBlock := blockPx > 0

	minInlineContent := convertMinMaxToContentSize(minInlinePx, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, inlineIsWidth)
	maxInlineContent := convertMinMaxToContentSize(maxInlinePx, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, inlineIsWidth)
	minBlockContent := convertMinMaxToContentSize(minBlockPx, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, !inlineIsWidth)
	maxBlockContent := convertMinMaxToContentSize(maxBlockPx, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, !inlineIsWidth)

	// The inline size used for line breaking: the explicit inline size when
	// set (converted to content-box), otherwise the available space from the
	// constraints. Either way it is clamped by min/max so lines break at the
	// used size, not at the available size.
	// https://www.w3.org/TR/css-text-3/#line-breaking
	var contentInline float64
	if hasExplicitInline {
		contentInline = convertToContentSize(inlinePx, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, inlineIsWidth)
	} else {
		contentInline = getInlineConstraint(constraints, writingMode) - inlinePaddingBorder
	}
	contentInline = clampContentSize(contentInline, minInlineContent, maxInlineContent)
	if contentInline < 0 {
		contentInline = 0
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

	// 3. Perform line breaking (§4) with getTextMetrics().Measure
	lines, lineMetas := breakIntoLines(processedText, contentInline, *style)

	// 3.5. Apply text-overflow if needed (ellipsis truncation)
	// CSS Text Overflow Module Level 3: https://www.w3.org/TR/css-overflow-3/#text-overflow
	if style.TextOverflow == TextOverflowEllipsis {
		lines = applyTextOverflow(lines, lineMetas, contentInline, *style)
	}

	// 4. Compute the block size from line count and line-height (§4.4.1)
	// If no lines, use at least one line height for empty text
	lineHeight := resolveLineHeight(style.LineHeight, style.FontSize)
	numLines := len(lines)
	if numLines == 0 {
		numLines = 1
	}
	contentBlock := float64(numLines) * lineHeight

	// Find max line extent (including text-indent for first line)
	maxLineInline := 0.0
	for i, line := range lines {
		w := line.Width
		// Include text-indent in first line width calculation
		if i == 0 && style.TextIndent != 0 {
			w += style.TextIndent
		}
		if w > maxLineInline {
			maxLineInline = w
		}
	}

	// 5. Resolve the used inline size: explicit (already clamped) or shrink-to-fit
	if !hasExplicitInline {
		contentInline = clampContentSize(maxLineInline, minInlineContent, maxInlineContent)
	}

	// Resolve the used block size
	if hasExplicitBlock {
		contentBlock = convertToContentSize(blockPx, node.Style.BoxSizing, horizontalPaddingBorder, verticalPaddingBorder, !inlineIsWidth)
	}
	contentBlock = clampContentSize(contentBlock, minBlockContent, maxBlockContent)

	// Constrain and set Rect (mapping logical sizes back to physical)
	outer := makeSize(contentInline+inlinePaddingBorder, contentBlock+blockPaddingBorder, writingMode)
	size := constraints.Constrain(outer)
	node.Rect.Width = size.Width
	node.Rect.Height = size.Height

	// 6. Compute per-line positions based on text-align (§7.1), text-align-last
	// (§7.2.2), text-justify (§7.3), text-indent (§7.2.1), direction (§2), and
	// writing-mode. Lines are positioned within the final content box so that
	// vertical-rl lines start at its right edge.
	// https://www.w3.org/TR/css-writing-modes-3/#block-flow
	//
	// With an explicit inline size, lines align within it. With an auto inline
	// size the box is shrink-to-fit, but alignment still happens within the
	// available inline space (the traditional behavior of this engine, which
	// lets text-align: right/center work inside a wider parent). When the
	// available space is unbounded, fall back to the used size.
	inlineAlignSize := contentInline
	if !hasExplicitInline {
		available := getInlineConstraint(constraints, writingMode) - inlinePaddingBorder
		available = clampContentSize(available, minInlineContent, maxInlineContent)
		if available >= 0 && available < Unbounded {
			inlineAlignSize = available
		}
	}
	finalBlock := getBlockSize(size.Width, size.Height, writingMode) - blockPaddingBorder
	if finalBlock < 0 {
		finalBlock = 0
	}
	positionLines(lines, inlineAlignSize, finalBlock, lineMetas, style.TextAlign, style.TextAlignLast, style.TextJustify, style.TextIndent, style.Direction, lineHeight, writingMode)

	// 6.5. Apply hanging-punctuation (§9.2)
	applyHangingPunctuation(lines, style.HangingPunctuation, *style)

	// 7. Store line metadata for rendering
	node.TextLayout = &TextLayout{
		Lines:      lines,
		LineHeight: lineHeight,
	}

	return size
}

// clampContentSize clamps size to [min, max]. A min of 0 or less is ignored,
// as is a max of 0 or less or an unbounded max.
func clampContentSize(size, min, max float64) float64 {
	if min > 0 && size < min {
		size = min
	}
	if max > 0 && max < Unbounded && size > max {
		size = max
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

		return trimCollapsibleSpace(text)

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
			lines[i] = trimCollapsibleSpace(collapseWhitespace(line))
		}
		return strings.Join(lines, "\n")

	default:
		return text
	}
}

// isCollapsibleSpace reports whether r is white space that the white-space
// property may collapse or trim. U+00A0 NO-BREAK SPACE is not collapsible
// (it is not a "document white space character" in CSS Text 3 §3.1).
// https://www.w3.org/TR/css-text-3/#white-space-processing
func isCollapsibleSpace(r rune) bool {
	return r != '\u00A0' && unicode.IsSpace(r)
}

// trimCollapsibleSpace trims leading and trailing collapsible white space,
// preserving non-breaking spaces. Use it instead of strings.TrimSpace, which
// would strip U+00A0.
func trimCollapsibleSpace(s string) string {
	return strings.TrimFunc(s, isCollapsibleSpace)
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
		isWhitespace := isCollapsibleSpace(r)

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

	// Default tab size is 8 spaces. TextStyle documents -1 as "default", and
	// the zero value (unset) must behave the same way rather than collapsing
	// tabs to a single space.
	if tabSize <= 0 {
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

		// Handle first punctuation (opening). Only "first" hangs opening
		// punctuation; allow-end and force-end apply to the end edge only.
		// https://www.w3.org/TR/css-text-3/#valdef-hanging-punctuation-first
		if hanging == HangingPunctuationFirst {
			firstBox := &line.Boxes[0]
			if len(firstBox.Text) > 0 {
				runes := []rune(firstBox.Text)
				if isOpeningPunctuation(runes[0]) {
					// Measure the punctuation character
					punctWidth, _, _ := getTextMetrics().Measure(string(runes[0]), style)
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
					punctWidth, _, _ := getTextMetrics().Measure(string(runes[len(runes)-1]), style)
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
		// Non-breaking spaces (U+00A0) are not word separators
		if isCollapsibleSpace(r) {
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

// textLineMeta carries per-line information that TextLine cannot hold yet.
// It is kept local to the text layout code; the fields are candidates for
// promotion onto TextLine in types.go.
type textLineMeta struct {
	// EndsWithForcedBreak is true when the line ends because of a preserved
	// newline (a forced line break) rather than a soft wrap.
	// https://www.w3.org/TR/css-text-3/#forced-line-break
	EndsWithForcedBreak bool
	// SpaceAfter[j] reports whether a collapsible inter-word space follows
	// Boxes[j] on the line. len(SpaceAfter) == len(Boxes) when set.
	SpaceAfter []bool
}

// lineEndsWithForcedBreak reports whether line i is terminated by a forced
// break according to metas (nil-safe).
func lineEndsWithForcedBreak(metas []textLineMeta, i int) bool {
	return i >= 0 && i < len(metas) && metas[i].EndsWithForcedBreak
}

// markForcedBreak marks the last line in metas as ending with a forced break.
func markForcedBreak(metas []textLineMeta) {
	if len(metas) > 0 {
		metas[len(metas)-1].EndsWithForcedBreak = true
	}
}

// breakIntoLines breaks text into lines based on available inline size using UAX #14.
// Based on CSS Text Module Level 3 §4: https://www.w3.org/TR/css-text-3/#line-breaking
// Uses Unicode Line Breaking Algorithm (UAX #14) for proper break opportunities.
//
// The maxInlineSize parameter represents:
//   - Horizontal writing modes: max width (text flows left-to-right or right-to-left)
//   - Vertical writing modes: max height (text flows top-to-bottom)
//
// Note: TextLine.Width field represents the inline-size extent:
//   - Horizontal: width in pixels
//   - Vertical: height in pixels (how tall the "line" is when flowing top-to-bottom)
//
// The returned metas slice is parallel to the returned lines.
func breakIntoLines(text string, maxInlineSize float64, style TextStyle) ([]TextLine, []textLineMeta) {
	if text == "" {
		return []TextLine{}, nil
	}

	// Treat maxInlineSize <= 0 as unbounded (no wrapping)
	if maxInlineSize <= 0 {
		maxInlineSize = Unbounded
	}

	// For pre mode, split on newlines first
	if style.WhiteSpace == WhiteSpacePre {
		return breakIntoLinesPre(text, maxInlineSize, style)
	}

	// For pre-wrap and pre-line, split on newlines then wrap each segment
	if style.WhiteSpace == WhiteSpacePreWrap || style.WhiteSpace == WhiteSpacePreLine {
		return breakIntoLinesPreWrap(text, maxInlineSize, style)
	}

	// Use UAX #14 to find line break opportunities
	return breakIntoLinesUAX14(text, maxInlineSize, style)
}

// uax14LineBuilder accumulates boxes for the line currently being built by
// breakIntoLinesUAX14.
type uax14LineBuilder struct {
	line          TextLine
	meta          textLineMeta
	width         float64 // total advance including inter-word spaces
	trailingSpace bool    // the last box is followed by a space
	spaceWidth    float64 // width of that trailing space
}

func newUAX14LineBuilder() uax14LineBuilder {
	return uax14LineBuilder{line: TextLine{Boxes: []InlineBox{}}}
}

// addBox appends a box, optionally followed by an inter-word space.
func (b *uax14LineBuilder) addBox(box InlineBox, hasSpace bool, spaceWidth float64) {
	b.line.Boxes = append(b.line.Boxes, box)
	b.meta.SpaceAfter = append(b.meta.SpaceAfter, hasSpace)
	b.width += box.Width
	if hasSpace {
		b.line.SpaceCount++
		b.line.SpaceWidth += spaceWidth
		b.width += spaceWidth
	}
	b.trailingSpace = hasSpace
	b.spaceWidth = spaceWidth
}

// finish closes the line. A trailing space is removed from the line
// (§4.1.3: trailing collapsible white space is removed at the end of a line)
// and does not count toward justification.
func (b *uax14LineBuilder) finish() (TextLine, textLineMeta) {
	if b.trailingSpace && b.line.SpaceCount > 0 {
		b.line.SpaceCount--
		b.line.SpaceWidth -= b.spaceWidth
		b.width -= b.spaceWidth
		b.meta.SpaceAfter[len(b.meta.SpaceAfter)-1] = false
	}
	b.line.Width = b.width
	return b.line, b.meta
}

// breakIntoLinesUAX14 breaks text into lines using UAX #14 line breaking algorithm.
// maxInlineSize represents the maximum extent in the inline dimension (width for horizontal, height for vertical).
func breakIntoLinesUAX14(text string, maxInlineSize float64, style TextStyle) ([]TextLine, []textLineMeta) {
	// Find all line break opportunities using UAX #14, respecting hyphens property
	breakPoints := findLineBreakOpportunitiesWithHyphens(text, style.Hyphens)
	if len(breakPoints) < 2 {
		return []TextLine{}, nil
	}

	bounded := maxInlineSize > 0 && maxInlineSize < Unbounded
	canWrap := bounded && canBreakBefore(style.WhiteSpace)

	lines := []TextLine{}
	metas := []textLineMeta{}
	current := newUAX14LineBuilder()

	flush := func() {
		line, meta := current.finish()
		lines = append(lines, line)
		metas = append(metas, meta)
		current = newUAX14LineBuilder()
	}

	// Process text segment by segment. The loop is bounded by the number of
	// break points, which is at most the number of runes plus one.
	for i := 0; i < len(breakPoints)-1; i++ {
		start := breakPoints[i]
		end := breakPoints[i+1]
		segment := text[start:end]

		// Skip empty segments
		if len(segment) == 0 {
			continue
		}

		// Segments may include trailing spaces - check and handle separately
		hasTrailingSpace := segment[len(segment)-1] == ' '
		wordText := segment
		var spaceWidth float64

		if hasTrailingSpace {
			// Strip trailing space and measure it separately
			wordText = segment[:len(segment)-1]
			spaceWidth, _, _ = getTextMetrics().Measure(" ", style)
			if style.WordSpacing != -1 {
				spaceWidth += style.WordSpacing
			}
		}

		// Skip if word is empty (segment was just a space)
		if len(wordText) == 0 {
			continue
		}

		// Measure the word (without trailing space)
		wordWidth, ascent, descent := getTextMetrics().Measure(wordText, style)

		// text-indent applies to the first line of the block (§7.2.1),
		// regardless of how many boxes are already on it.
		// https://www.w3.org/TR/css-text-3/#text-indent-property
		isFirstLine := len(lines) == 0
		indent := 0.0
		if isFirstLine {
			indent = style.TextIndent
		}

		// Check whether this word fits. The word's own trailing space is not
		// counted: white space at the end of a line hangs and never causes a
		// wrap (§4.1.3).
		// https://www.w3.org/TR/css-text-3/#white-space-phase-2
		effectiveLineWidth := current.width + wordWidth + indent

		// Break if this word would exceed maxInlineSize (and we have content already on this line)
		if canWrap && effectiveLineWidth > maxInlineSize && len(current.line.Boxes) > 0 {
			flush()
			isFirstLine = false
			indent = 0
		}

		// Check if word is too long and should be broken (overflow-wrap or word-break)
		// Only break if it's the first word on line and exceeds maxInlineSize
		availableForWord := maxInlineSize - indent
		if len(current.line.Boxes) == 0 && bounded && wordWidth > availableForWord {
			if style.OverflowWrap == OverflowWrapBreakWord || style.OverflowWrap == OverflowWrapAnywhere ||
				style.WordBreak == WordBreakBreakAll {
				// Break word into smaller pieces
				pieceMax := availableForWord
				if pieceMax <= 0 {
					pieceMax = maxInlineSize
				}
				pieces := breakWordToFit(wordText, pieceMax, style)
				for j, piece := range pieces {
					if j > 0 {
						// Start new line for subsequent pieces
						flush()
					}

					pieceWidth, ascent, descent := getTextMetrics().Measure(piece, style)
					isLast := j == len(pieces)-1
					current.addBox(newInlineBox(piece, pieceWidth, ascent, descent, style.WritingMode), isLast && hasTrailingSpace, spaceWidth)
				}

				continue // Skip normal word addition
			}
		}

		// Add the word to current line
		current.addBox(newInlineBox(wordText, wordWidth, ascent, descent, style.WritingMode), hasTrailingSpace, spaceWidth)
	}

	// Add final line
	if len(current.line.Boxes) > 0 {
		flush()
	}

	return lines, metas
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

// breakIntoLinesPre breaks text into lines preserving newlines and spaces (pre mode).
// maxInlineSize represents the maximum extent in the inline dimension (width for horizontal, height for vertical).
// Every line except the last ends with a forced break.
func breakIntoLinesPre(text string, maxInlineSize float64, style TextStyle) ([]TextLine, []textLineMeta) {
	// Split by newlines
	lineTexts := strings.Split(text, "\n")
	lines := make([]TextLine, 0, len(lineTexts))
	metas := make([]textLineMeta, 0, len(lineTexts))

	for i, lineText := range lineTexts {
		line := TextLine{Boxes: []InlineBox{}}

		// Measure the entire line text (preserving all spaces)
		// Text-indent affects alignment, not intrinsic width, so handle in positionLines()
		advance, ascent, descent := getTextMetrics().Measure(lineText, style)
		line.Boxes = append(line.Boxes, newInlineBox(lineText, advance, ascent, descent, style.WritingMode))
		line.Width = advance
		lines = append(lines, line)
		metas = append(metas, textLineMeta{EndsWithForcedBreak: i < len(lineTexts)-1})
	}

	return lines, metas
}

// breakIntoLinesPreWrap handles pre-wrap and pre-line modes.
// Split on newlines, then wrap each segment. The last line produced by each
// segment except the final one ends with a forced break.
// maxInlineSize represents the maximum extent in the inline dimension (width for horizontal, height for vertical).
func breakIntoLinesPreWrap(text string, maxInlineSize float64, style TextStyle) ([]TextLine, []textLineMeta) {
	lines := []TextLine{}
	metas := []textLineMeta{}

	// Split by newlines
	segments := strings.Split(text, "\n")

	for si, segment := range segments {
		var segmentLines []TextLine
		var segmentMetas []textLineMeta
		if segment != "" {
			// Wrap this segment if it exceeds maxInlineSize
			// For pre-wrap: preserve spaces within the segment
			// For pre-line: spaces already collapsed in preprocessText
			segmentLines, segmentMetas = wrapSegment(segment, maxInlineSize, style)
		}
		if len(segmentLines) == 0 {
			// Empty line from consecutive newlines, a trailing newline, or a
			// segment that collapsed to nothing
			segmentLines = []TextLine{{Boxes: []InlineBox{}, Width: 0}}
			segmentMetas = []textLineMeta{{}}
		}
		if len(segmentMetas) != len(segmentLines) {
			segmentMetas = make([]textLineMeta, len(segmentLines))
		}
		lines = append(lines, segmentLines...)
		metas = append(metas, segmentMetas...)
		if si < len(segments)-1 {
			markForcedBreak(metas)
		}
	}

	return lines, metas
}

// wrapSegment wraps a single segment (between newlines).
// maxInlineSize represents the maximum extent in the inline dimension (width for horizontal, height for vertical).
//
// Segments always go through the full line builder, even when they fit on one
// line, so that inter-word spaces are tracked (SpaceCount) and the line can be
// justified.
func wrapSegment(segment string, maxInlineSize float64, style TextStyle) ([]TextLine, []textLineMeta) {
	// For pre-wrap mode, preserve all spaces including multiple consecutive ones
	if style.WhiteSpace == WhiteSpacePreWrap {
		return wrapSegmentPreserveSpaces(segment, maxInlineSize, style)
	}

	// For pre-line, use UAX #14 (spaces already collapsed in preprocessText)
	return breakIntoLinesUAX14(segment, maxInlineSize, style)
}

// wrapSegmentPreserveSpaces wraps text while preserving all spaces (for pre-wrap mode).
// maxInlineSize represents the maximum extent in the inline dimension (width for horizontal, height for vertical).
//
// Runs of preserved spaces are kept as their own boxes. Per CSS Text 3 §4.1.3
// a sequence of preserved white space at the end of a line hangs: it never
// causes a wrap and is not counted in the line's width. Before a forced break
// (or the end of the text) the sequence hangs conditionally, i.e. only the
// part that would overflow is excluded.
// https://www.w3.org/TR/css-text-3/#white-space-phase-2
func wrapSegmentPreserveSpaces(segment string, maxInlineSize float64, style TextStyle) ([]TextLine, []textLineMeta) {
	lines := []TextLine{}
	current := TextLine{Boxes: []InlineBox{}}
	currentWidth := 0.0       // width of all boxes on the current line
	trailingSpaceWidth := 0.0 // width of the space boxes at the end of the current line

	runes := []rune(segment)
	n := len(runes)

	// Tokenize into alternating runs of spaces and non-spaces. Each iteration
	// consumes at least one rune, so the loop is bounded by n.
	for i := 0; i < n; {
		isSpace := runes[i] == ' '
		j := i + 1
		for j < n && (runes[j] == ' ') == isSpace {
			j++
		}
		token := string(runes[i:j])
		i = j

		tokenWidth, ascent, descent := getTextMetrics().Measure(token, style)

		if isSpace {
			// Spaces hang at the end of a line: always append to the current line
			current.Boxes = append(current.Boxes, newInlineBox(token, tokenWidth, ascent, descent, style.WritingMode))
			currentWidth += tokenWidth
			trailingSpaceWidth += tokenWidth
			continue
		}

		// A word: wrap first if it does not fit and the line already has content
		if len(current.Boxes) > 0 && currentWidth+tokenWidth > maxInlineSize {
			current.Width = currentWidth - trailingSpaceWidth
			lines = append(lines, current)
			current = TextLine{Boxes: []InlineBox{}}
			currentWidth = 0.0
			trailingSpaceWidth = 0.0
		}

		current.Boxes = append(current.Boxes, newInlineBox(token, tokenWidth, ascent, descent, style.WritingMode))
		currentWidth += tokenWidth
		trailingSpaceWidth = 0.0
	}

	// Add final line if not empty. Trailing spaces hang conditionally here:
	// they count toward the width as long as they fit.
	if len(current.Boxes) > 0 {
		current.Width = currentWidth
		if maxInlineSize < Unbounded && currentWidth > maxInlineSize {
			withoutSpaces := currentWidth - trailingSpaceWidth
			if withoutSpaces > maxInlineSize {
				current.Width = withoutSpaces
			} else {
				current.Width = maxInlineSize
			}
		}
		lines = append(lines, current)
	}

	return lines, make([]textLineMeta, len(lines))
}

// breakWordToFit breaks a word into pieces that fit maxInlineSize.
// Used for overflow-wrap: break-word and word-break: break-all.
// maxInlineSize represents the maximum extent in the inline dimension (width for horizontal, height for vertical).
//
// Each candidate piece is measured as a run so that letter-spacing (and any
// other run-dependent metric) is honored, rather than summing per-character
// advances.
// https://www.w3.org/TR/css-text-3/#overflow-wrap-property
func breakWordToFit(word string, maxInlineSize float64, style TextStyle) []string {
	pieces := []string{}
	runes := []rune(word)
	if len(runes) == 0 {
		return pieces
	}

	current := make([]rune, 0, len(runes))
	// One iteration per rune; every rune ends up in exactly one piece.
	for _, r := range runes {
		current = append(current, r)
		width, _, _ := getTextMetrics().Measure(string(current), style)
		if width > maxInlineSize && len(current) > 1 {
			// Finish the piece without this rune and start a new one with it
			pieces = append(pieces, string(current[:len(current)-1]))
			current = append(current[:0], r)
		}
	}

	if len(current) > 0 {
		pieces = append(pieces, string(current))
	}

	return pieces
}

// applyTextOverflow applies text-overflow: ellipsis to overflowing lines
// CSS Text Overflow Module Level 3: https://www.w3.org/TR/css-overflow-3/#text-overflow
//
// metas (parallel to lines, may be nil) tells which boxes are followed by an
// inter-word space so the space advance is accounted for when fitting boxes.
func applyTextOverflow(lines []TextLine, metas []textLineMeta, contentWidth float64, style TextStyle) []TextLine {
	if len(lines) == 0 {
		return lines
	}

	// Measure ellipsis width
	ellipsisText := "..."
	ellipsisWidth, ellipsisAscent, ellipsisDescent := getTextMetrics().Measure(ellipsisText, style)

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
			line.Boxes = []InlineBox{newInlineBox(ellipsisText, ellipsisWidth, ellipsisAscent, ellipsisDescent, style.WritingMode)}
			line.Width = ellipsisWidth
			line.SpaceCount = 0
			line.SpaceWidth = 0
			line.SpaceAdjustment = 0
			continue
		}

		// Determine which gaps between boxes are inter-word spaces and how
		// wide each one is.
		var spaceAfter []bool
		if i < len(metas) && len(metas[i].SpaceAfter) == len(line.Boxes) {
			spaceAfter = metas[i].SpaceAfter
		}
		gaps := len(line.Boxes) - 1
		perSpace := 0.0
		if line.SpaceCount > 0 {
			perSpace = line.SpaceWidth / float64(line.SpaceCount)
		}
		gapIsSpace := func(j int) bool {
			if j <= 0 || line.SpaceCount == 0 {
				return false
			}
			if spaceAfter != nil {
				return spaceAfter[j-1]
			}
			// Without metadata, assume every gap is a space when the counts agree
			return line.SpaceCount >= gaps
		}

		// Truncate boxes to fit within availableWidth
		truncatedBoxes := []InlineBox{}
		currentWidth := 0.0
		retainedSpaces := 0

		for j, box := range line.Boxes {
			gapWidth := 0.0
			hasGap := gapIsSpace(j)
			if hasGap {
				gapWidth = perSpace
			}

			if currentWidth+gapWidth+box.Width <= availableWidth {
				// Box fits completely
				truncatedBoxes = append(truncatedBoxes, box)
				currentWidth += gapWidth + box.Width
				if hasGap {
					retainedSpaces++
				}
				continue
			}

			// Box would overflow - truncate it
			remainingWidth := availableWidth - currentWidth - gapWidth
			if remainingWidth > 0 {
				// Try to fit part of this box
				truncatedText := truncateTextToWidth(box.Text, remainingWidth, style)
				if truncatedText != "" {
					truncWidth, truncAscent, truncDesc := getTextMetrics().Measure(truncatedText, style)
					truncatedBoxes = append(truncatedBoxes, newInlineBox(truncatedText, truncWidth, truncAscent, truncDesc, style.WritingMode))
					currentWidth += gapWidth + truncWidth
					if hasGap {
						retainedSpaces++
					}
				}
			}
			break // Stop processing boxes
		}

		// Add ellipsis
		truncatedBoxes = append(truncatedBoxes, newInlineBox(ellipsisText, ellipsisWidth, ellipsisAscent, ellipsisDescent, style.WritingMode))

		line.Boxes = truncatedBoxes
		line.Width = currentWidth + ellipsisWidth
		// Keep the retained inter-word gaps so renderers place the words correctly
		line.SpaceCount = retainedSpaces
		line.SpaceWidth = float64(retainedSpaces) * perSpace
		line.SpaceAdjustment = 0
	}

	return lines
}

// truncateTextToWidth truncates text to fit within maxInlineSize.
// maxInlineSize represents the maximum extent in the inline dimension (width for horizontal, height for vertical).
func truncateTextToWidth(text string, maxInlineSize float64, style TextStyle) string {
	runes := []rune(text)

	// Binary search for the longest prefix that fits
	left, right := 0, len(runes)
	result := ""

	for left <= right {
		mid := (left + right) / 2
		candidate := string(runes[:mid])
		width, _, _ := getTextMetrics().Measure(candidate, style)

		if width <= maxInlineSize {
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
//
// contentInlineSize is the size lines are aligned within; contentBlockSize is
// the block extent of the content box, used as the starting edge for writing
// modes whose lines stack right-to-left. metas (parallel to lines, may be nil)
// marks lines that end with a forced break: per §7.1 those lines, like the
// last line of the block, are aligned with text-align-last instead of being
// justified.
// https://www.w3.org/TR/css-text-3/#text-align-property
//
// For vertical writing modes, the logical positioning changes:
//   - Horizontal: lines stack vertically (Y increases), alignment is horizontal (X)
//   - Vertical-LR: lines stack left-to-right (X increases), alignment is vertical (Y)
//   - Vertical-RL: lines stack right-to-left (X decreases), alignment is vertical (Y)
//
// https://www.w3.org/TR/css-writing-modes-3/#block-flow
func positionLines(lines []TextLine, contentInlineSize, contentBlockSize float64, metas []textLineMeta, textAlign TextAlign, textAlignLast TextAlignLast, textJustify TextJustify, textIndent float64, direction Direction, lineHeight float64, writingMode WritingMode) {
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

	// For vertical writing modes, lines stack horizontally
	// For horizontal modes, lines stack vertically (current behavior)
	isVertical := writingMode.IsVertical()
	// Vertical-RL and Sideways-RL both stack right-to-left
	isRightToLeft := (writingMode == WritingModeVerticalRL || writingMode == WritingModeSidewaysRL)

	// Initialize block-axis position
	// Horizontal: block-axis is Y (lines stack downward)
	// Vertical-LR: block-axis is X (lines stack rightward)
	// Vertical-RL/Sideways-RL: block-axis is X (lines stack leftward, starting from the
	// right edge of the content box, i.e. contentBlockSize)
	currentBlockPos := 0.0
	if isVertical && isRightToLeft {
		// Vertical-RL/Sideways-RL: start from the right edge and move leftward
		currentBlockPos = contentBlockSize
	}

	for i := range lines {
		line := &lines[i]
		lineWidth := line.Width
		indent := 0.0

		// First line gets text-indent (can be positive or negative)
		if i == 0 && textIndent != 0 {
			indent = textIndent
		}

		// Calculate inline-axis offset based on text-align
		// Horizontal mode: inline-axis is X
		// Vertical mode: inline-axis is Y
		// Per CSS Text Module Level 3 §7.2.1: text-indent is treated as a margin
		// applied to the start edge (left in LTR, top in vertical)
		var inlineOffset float64

		switch align {
		case TextAlignLeft:
			// Start-aligned: text starts at indent position
			inlineOffset = indent

		case TextAlignRight:
			// End-aligned: indent reduces available size, text aligns to (contentInlineSize - indent)
			// So text ends at (contentInlineSize - indent), not at contentInlineSize
			inlineOffset = contentInlineSize - lineWidth - indent

		case TextAlignCenter:
			// Center-aligned: indent reduces available size, center within remaining space
			// Available size is (contentInlineSize - indent), center the line within that
			availableSize := contentInlineSize - indent
			inlineOffset = indent + (availableSize-lineWidth)/2

		case TextAlignJustify:
			// Justified: distribute extra space using text-justify algorithm
			// Per CSS Text Module Level 3 §7.1.1, §7.2.2, and §7.3
			// §7.1: "the last line before a forced break" is treated like the
			// last line of the block and uses text-align-last.
			isLastLine := i == len(lines)-1 || lineEndsWithForcedBreak(metas, i)
			hasMultipleWords := line.SpaceCount > 0

			// Resolve text-justify
			justifyMode := textJustify
			if justifyMode == TextJustifyAuto {
				justifyMode = TextJustifyInterWord
			}

			// Handle text-justify: none
			if justifyMode == TextJustifyNone {
				inlineOffset = indent
			} else if !isLastLine && hasMultipleWords {
				// Middle lines: apply justification algorithm
				availableSize := contentInlineSize
				if i == 0 && indent != 0 {
					availableSize -= indent
				}

				extraSpace := availableSize - lineWidth
				if extraSpace > 0 {
					switch justifyMode {
					case TextJustifyInterWord:
						// Distribute across word spaces only (current implementation)
						line.SpaceAdjustment = extraSpace / float64(line.SpaceCount)
						line.Width = availableSize

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
							line.Width = availableSize
						} else if line.SpaceCount > 0 {
							// Fallback: if no character gaps, use inter-word
							line.SpaceAdjustment = extraSpace / float64(line.SpaceCount)
							line.Width = availableSize
						}
					}
				}
				inlineOffset = indent
			} else {
				// Last line or single word: use text-align-last
				lastAlign := resolveTextAlignLast(textAlignLast, align)

				switch lastAlign {
				case TextAlignLastLeft:
					inlineOffset = indent
				case TextAlignLastRight:
					inlineOffset = contentInlineSize - lineWidth
					if i == 0 {
						inlineOffset -= indent
					}
				case TextAlignLastCenter:
					availableSize := contentInlineSize
					if i == 0 {
						availableSize -= indent
					}
					inlineOffset = indent + (availableSize-lineWidth)/2
				case TextAlignLastJustify:
					// Justify even last line
					if hasMultipleWords {
						availableSize := contentInlineSize
						if i == 0 && indent != 0 {
							availableSize -= indent
						}
						extraSpace := availableSize - lineWidth
						if extraSpace > 0 {
							line.SpaceAdjustment = extraSpace / float64(line.SpaceCount)
							line.Width = availableSize
						}
					}
					inlineOffset = indent
				default:
					inlineOffset = indent
				}
			}

		default:
			inlineOffset = 0.0
		}

		// Map logical offsets to physical X/Y based on writing mode
		if isVertical {
			// Vertical modes: inline is Y, block is X
			line.OffsetY = inlineOffset
			if isRightToLeft {
				// Vertical-RL/Sideways-RL: lines flow right-to-left (X decreases)
				line.OffsetX = currentBlockPos - lineHeight
			} else {
				// Vertical-LR/Sideways-LR: lines flow left-to-right (X increases)
				line.OffsetX = currentBlockPos
			}
		} else {
			// Horizontal mode: inline is X, block is Y
			line.OffsetX = inlineOffset
			line.OffsetY = currentBlockPos
		}

		// Advance block-axis position
		if isVertical && isRightToLeft {
			// Vertical-RL/Sideways-RL: move leftward
			currentBlockPos -= lineHeight
		} else {
			// Horizontal or Vertical-LR/Sideways-LR: move downward or rightward
			currentBlockPos += lineHeight
		}
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

// computeTextOrientations computes character orientations for the given text string.
// Returns a slice of bools where each element corresponds to a rune in the text:
//   - true: character should be upright (CJK in vertical modes)
//   - false: character should be rotated 90° (Latin in vertical modes)
//
// For horizontal writing modes, returns nil (no orientation data needed).
// For vertical modes, uses UAX #50 to determine orientation per character.
func computeTextOrientations(text string, wm WritingMode) []bool {
	// Horizontal modes don't need orientation data
	if !wm.IsVertical() {
		return nil
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}

	orientations := make([]bool, len(runes))
	for i, r := range runes {
		orientations[i] = getCharacterOrientation(r, wm)
	}

	return orientations
}

// newInlineBox creates an InlineBox with character orientation data populated.
// This helper ensures consistent InlineBox creation throughout the text layout code.
func newInlineBox(text string, width, ascent, descent float64, wm WritingMode) InlineBox {
	return InlineBox{
		Kind:         InlineBoxText,
		Text:         text,
		Width:        width,
		Ascent:       ascent,
		Descent:      descent,
		Orientations: computeTextOrientations(text, wm),
	}
}
