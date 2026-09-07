package layout

import (
	"fmt"
	"sort"
	"strings"
)

// This file provides String() and Parse<Enum>() for every exported CSS-like
// enum in types.go. String returns the canonical CSS keyword ("flex-start",
// "space-between", "inline-text", ...) and Parse accepts that keyword, plus
// any documented alias, and returns an error listing the valid keywords for
// anything else.
//
// Matching is exact: keywords are lowercase and neither case-folded nor
// trimmed, so "Flex" is rejected. This keeps the serialize wire format
// strict for untrusted input. (ParseContainerType in container.go predates
// this file and is case-insensitive.)
//
// Aliases: CSS Box Alignment (css-align-3) spells the flexbox keywords
// flex-start/flex-end as start/end. The alias constants in types.go
// (JustifyContentStart, AlignItemsStart, AlignContentStart, ...) share the
// value of their flex-* counterpart, so Parse accepts "start"/"end" for those
// enums and String returns the canonical "flex-start"/"flex-end".
//
// A value outside the enum's range formats as "Display(7)" (the Go type name
// and the integer) so it is recognizable in logs and never parses back.

// enumString returns names[v] or "Kind(v)" when v is out of range.
func enumString[T ~int](kind string, names []string, v T) string {
	if i := int(v); i >= 0 && i < len(names) {
		return names[i]
	}
	return fmt.Sprintf("%s(%d)", kind, int(v))
}

// parseEnum looks s up in names (by index) and then in aliases. The error
// lists the canonical keywords and the aliases; property is the CSS property
// name used in the message.
func parseEnum[T ~int](property string, names []string, aliases map[string]T, s string) (T, error) {
	for i, name := range names {
		if s == name {
			return T(i), nil
		}
	}
	if v, ok := aliases[s]; ok {
		return v, nil
	}
	valid := make([]string, 0, len(names)+len(aliases))
	valid = append(valid, names...)
	for alias := range aliases {
		valid = append(valid, alias)
	}
	// Aliases come from a map; keep the message deterministic.
	sort.Strings(valid[len(names):])
	return 0, fmt.Errorf("layout: invalid %s %q (valid: %s)", property, s, strings.Join(valid, ", "))
}

// -----------------------------------------------------------------------------
// Display
// -----------------------------------------------------------------------------

var displayNames = []string{"block", "flex", "grid", "inline-text", "none"}

// String returns the CSS keyword for the display value. DisplayInlineText is
// "inline-text", which is not a CSS keyword but names the text leaf node.
func (d Display) String() string { return enumString("Display", displayNames, d) }

// ParseDisplay parses a display keyword: block, flex, grid, inline-text, none.
func ParseDisplay(s string) (Display, error) {
	return parseEnum[Display]("display", displayNames, nil, s)
}

// -----------------------------------------------------------------------------
// FlexDirection
// -----------------------------------------------------------------------------

var flexDirectionNames = []string{"row", "row-reverse", "column", "column-reverse"}

// String returns the CSS keyword for the flex-direction value.
func (d FlexDirection) String() string {
	return enumString("FlexDirection", flexDirectionNames, d)
}

// ParseFlexDirection parses a flex-direction keyword.
// Spec: https://www.w3.org/TR/css-flexbox-1/#flex-direction-property
func ParseFlexDirection(s string) (FlexDirection, error) {
	return parseEnum[FlexDirection]("flex-direction", flexDirectionNames, nil, s)
}

// -----------------------------------------------------------------------------
// FlexWrap
// -----------------------------------------------------------------------------

var flexWrapNames = []string{"nowrap", "wrap", "wrap-reverse"}

// String returns the CSS keyword for the flex-wrap value.
func (w FlexWrap) String() string { return enumString("FlexWrap", flexWrapNames, w) }

// ParseFlexWrap parses a flex-wrap keyword.
// Spec: https://www.w3.org/TR/css-flexbox-1/#flex-wrap-property
func ParseFlexWrap(s string) (FlexWrap, error) {
	return parseEnum[FlexWrap]("flex-wrap", flexWrapNames, nil, s)
}

// -----------------------------------------------------------------------------
// JustifyContent
// -----------------------------------------------------------------------------

var justifyContentNames = []string{
	"flex-start", "flex-end", "center", "space-between", "space-around", "space-evenly", "stretch",
}

var justifyContentAliases = map[string]JustifyContent{
	"start": JustifyContentStart,
	"end":   JustifyContentEnd,
}

// String returns the CSS keyword for the justify-content value. The
// JustifyContentStart/End aliases format as "flex-start"/"flex-end".
func (j JustifyContent) String() string {
	return enumString("JustifyContent", justifyContentNames, j)
}

// ParseJustifyContent parses a justify-content keyword. "start" and "end" are
// accepted as aliases for flex-start and flex-end.
// Spec: https://www.w3.org/TR/css-align-3/#propdef-justify-content
func ParseJustifyContent(s string) (JustifyContent, error) {
	return parseEnum("justify-content", justifyContentNames, justifyContentAliases, s)
}

// -----------------------------------------------------------------------------
// AlignItems
// -----------------------------------------------------------------------------

var alignItemsNames = []string{"stretch", "flex-start", "flex-end", "center", "baseline"}

var alignItemsAliases = map[string]AlignItems{
	"start": AlignItemsStart,
	"end":   AlignItemsEnd,
}

// String returns the CSS keyword for the align-items (or align-self) value.
// The AlignItemsStart/End aliases format as "flex-start"/"flex-end".
func (a AlignItems) String() string { return enumString("AlignItems", alignItemsNames, a) }

// ParseAlignItems parses an align-items / align-self keyword. "start" and
// "end" are accepted as aliases for flex-start and flex-end.
// Spec: https://www.w3.org/TR/css-align-3/#propdef-align-items
func ParseAlignItems(s string) (AlignItems, error) {
	return parseEnum("align-items", alignItemsNames, alignItemsAliases, s)
}

// -----------------------------------------------------------------------------
// JustifyItems
// -----------------------------------------------------------------------------

var justifyItemsNames = []string{"stretch", "start", "end", "center"}

// String returns the CSS keyword for the justify-items (or justify-self) value.
func (j JustifyItems) String() string {
	return enumString("JustifyItems", justifyItemsNames, j)
}

// ParseJustifyItems parses a justify-items / justify-self keyword.
// Spec: https://www.w3.org/TR/css-align-3/#propdef-justify-items
func ParseJustifyItems(s string) (JustifyItems, error) {
	return parseEnum[JustifyItems]("justify-items", justifyItemsNames, nil, s)
}

// -----------------------------------------------------------------------------
// AlignContent
// -----------------------------------------------------------------------------

var alignContentNames = []string{
	"stretch", "flex-start", "flex-end", "center", "space-between", "space-around", "space-evenly",
}

var alignContentAliases = map[string]AlignContent{
	"start": AlignContentStart,
	"end":   AlignContentEnd,
}

// String returns the CSS keyword for the align-content value. The
// AlignContentStart/End aliases format as "flex-start"/"flex-end".
func (a AlignContent) String() string {
	return enumString("AlignContent", alignContentNames, a)
}

// ParseAlignContent parses an align-content keyword. "start" and "end" are
// accepted as aliases for flex-start and flex-end.
// Spec: https://www.w3.org/TR/css-align-3/#propdef-align-content
func ParseAlignContent(s string) (AlignContent, error) {
	return parseEnum("align-content", alignContentNames, alignContentAliases, s)
}

// -----------------------------------------------------------------------------
// GridAutoFlow
// -----------------------------------------------------------------------------

var gridAutoFlowNames = []string{"row", "column", "row-dense", "column-dense"}

var gridAutoFlowAliases = map[string]GridAutoFlow{
	// CSS writes the dense variants as two keywords: "row dense" / "column dense".
	"row dense":    GridAutoFlowRowDense,
	"column dense": GridAutoFlowColumnDense,
}

// String returns the keyword for the grid-auto-flow value. The dense variants
// are written hyphenated ("row-dense", "column-dense"); Parse also accepts
// the CSS two-keyword spelling.
func (g GridAutoFlow) String() string {
	return enumString("GridAutoFlow", gridAutoFlowNames, g)
}

// ParseGridAutoFlow parses a grid-auto-flow value: row, column, row-dense,
// column-dense, and the CSS two-keyword spellings "row dense" and
// "column dense". A bare "dense" is rejected because the wire format has
// always required the axis to be explicit.
// Spec: https://www.w3.org/TR/css-grid-1/#grid-auto-flow-property
func ParseGridAutoFlow(s string) (GridAutoFlow, error) {
	return parseEnum("grid-auto-flow", gridAutoFlowNames, gridAutoFlowAliases, s)
}

// -----------------------------------------------------------------------------
// BoxSizing
// -----------------------------------------------------------------------------

var boxSizingNames = []string{"content-box", "border-box"}

// String returns the CSS keyword for the box-sizing value.
func (b BoxSizing) String() string { return enumString("BoxSizing", boxSizingNames, b) }

// ParseBoxSizing parses a box-sizing keyword.
// Spec: https://www.w3.org/TR/css-sizing-3/#box-sizing
func ParseBoxSizing(s string) (BoxSizing, error) {
	return parseEnum[BoxSizing]("box-sizing", boxSizingNames, nil, s)
}

// -----------------------------------------------------------------------------
// Position
// -----------------------------------------------------------------------------

var positionNames = []string{"static", "relative", "absolute", "fixed", "sticky"}

// String returns the CSS keyword for the position value.
func (p Position) String() string { return enumString("Position", positionNames, p) }

// ParsePosition parses a position keyword.
// Spec: https://www.w3.org/TR/css-position-3/#position-property
func ParsePosition(s string) (Position, error) {
	return parseEnum[Position]("position", positionNames, nil, s)
}

// -----------------------------------------------------------------------------
// TextAlign
// -----------------------------------------------------------------------------

var textAlignNames = []string{"start", "left", "right", "center", "justify"}

// String returns the CSS keyword for the text-align value. TextAlignDefault
// is "start", the CSS initial value.
func (t TextAlign) String() string { return enumString("TextAlign", textAlignNames, t) }

// ParseTextAlign parses a text-align keyword.
// Spec: https://www.w3.org/TR/css-text-3/#text-align-property
func ParseTextAlign(s string) (TextAlign, error) {
	return parseEnum[TextAlign]("text-align", textAlignNames, nil, s)
}

// -----------------------------------------------------------------------------
// TextAlignLast
// -----------------------------------------------------------------------------

var textAlignLastNames = []string{"auto", "left", "right", "center", "justify"}

// String returns the CSS keyword for the text-align-last value.
func (t TextAlignLast) String() string {
	return enumString("TextAlignLast", textAlignLastNames, t)
}

// ParseTextAlignLast parses a text-align-last keyword.
// Spec: https://www.w3.org/TR/css-text-3/#text-align-last-property
func ParseTextAlignLast(s string) (TextAlignLast, error) {
	return parseEnum[TextAlignLast]("text-align-last", textAlignLastNames, nil, s)
}

// -----------------------------------------------------------------------------
// TextJustify
// -----------------------------------------------------------------------------

var textJustifyNames = []string{"auto", "inter-word", "inter-character", "distribute", "none"}

// String returns the CSS keyword for the text-justify value.
func (t TextJustify) String() string { return enumString("TextJustify", textJustifyNames, t) }

// ParseTextJustify parses a text-justify keyword.
// Spec: https://www.w3.org/TR/css-text-3/#text-justify-property
func ParseTextJustify(s string) (TextJustify, error) {
	return parseEnum[TextJustify]("text-justify", textJustifyNames, nil, s)
}

// -----------------------------------------------------------------------------
// WhiteSpace
// -----------------------------------------------------------------------------

var whiteSpaceNames = []string{"normal", "nowrap", "pre", "pre-wrap", "pre-line"}

// String returns the CSS keyword for the white-space value.
func (w WhiteSpace) String() string { return enumString("WhiteSpace", whiteSpaceNames, w) }

// ParseWhiteSpace parses a white-space keyword.
// Spec: https://www.w3.org/TR/css-text-3/#white-space-property
func ParseWhiteSpace(s string) (WhiteSpace, error) {
	return parseEnum[WhiteSpace]("white-space", whiteSpaceNames, nil, s)
}

// -----------------------------------------------------------------------------
// TextOverflow
// -----------------------------------------------------------------------------

var textOverflowNames = []string{"clip", "ellipsis"}

// String returns the CSS keyword for the text-overflow value.
func (t TextOverflow) String() string { return enumString("TextOverflow", textOverflowNames, t) }

// ParseTextOverflow parses a text-overflow keyword.
// Spec: https://www.w3.org/TR/css-overflow-3/#text-overflow
func ParseTextOverflow(s string) (TextOverflow, error) {
	return parseEnum[TextOverflow]("text-overflow", textOverflowNames, nil, s)
}

// -----------------------------------------------------------------------------
// OverflowWrap
// -----------------------------------------------------------------------------

var overflowWrapNames = []string{"normal", "break-word", "anywhere"}

// String returns the CSS keyword for the overflow-wrap value.
func (o OverflowWrap) String() string { return enumString("OverflowWrap", overflowWrapNames, o) }

// ParseOverflowWrap parses an overflow-wrap keyword.
// Spec: https://www.w3.org/TR/css-text-3/#overflow-wrap-property
func ParseOverflowWrap(s string) (OverflowWrap, error) {
	return parseEnum[OverflowWrap]("overflow-wrap", overflowWrapNames, nil, s)
}

// -----------------------------------------------------------------------------
// WordBreak
// -----------------------------------------------------------------------------

var wordBreakNames = []string{"normal", "break-all", "keep-all"}

// String returns the CSS keyword for the word-break value.
func (w WordBreak) String() string { return enumString("WordBreak", wordBreakNames, w) }

// ParseWordBreak parses a word-break keyword.
// Spec: https://www.w3.org/TR/css-text-3/#word-break-property
func ParseWordBreak(s string) (WordBreak, error) {
	return parseEnum[WordBreak]("word-break", wordBreakNames, nil, s)
}

// -----------------------------------------------------------------------------
// TextTransform
// -----------------------------------------------------------------------------

var textTransformNames = []string{
	"none", "uppercase", "lowercase", "capitalize", "full-width", "full-size-kana",
}

// String returns the CSS keyword for the text-transform value.
func (t TextTransform) String() string {
	return enumString("TextTransform", textTransformNames, t)
}

// ParseTextTransform parses a text-transform keyword.
// Spec: https://www.w3.org/TR/css-text-3/#text-transform-property
func ParseTextTransform(s string) (TextTransform, error) {
	return parseEnum[TextTransform]("text-transform", textTransformNames, nil, s)
}

// -----------------------------------------------------------------------------
// Hyphens
// -----------------------------------------------------------------------------

var hyphensNames = []string{"none", "manual", "auto"}

// String returns the CSS keyword for the hyphens value.
func (h Hyphens) String() string { return enumString("Hyphens", hyphensNames, h) }

// ParseHyphens parses a hyphens keyword.
// Spec: https://www.w3.org/TR/css-text-3/#hyphens-property
func ParseHyphens(s string) (Hyphens, error) {
	return parseEnum[Hyphens]("hyphens", hyphensNames, nil, s)
}

// -----------------------------------------------------------------------------
// HangingPunctuation
// -----------------------------------------------------------------------------

var hangingPunctuationNames = []string{"none", "first", "last", "force-end", "allow-end"}

// String returns the CSS keyword for the hanging-punctuation value.
func (h HangingPunctuation) String() string {
	return enumString("HangingPunctuation", hangingPunctuationNames, h)
}

// ParseHangingPunctuation parses a hanging-punctuation keyword.
// Spec: https://www.w3.org/TR/css-text-3/#hanging-punctuation-property
func ParseHangingPunctuation(s string) (HangingPunctuation, error) {
	return parseEnum[HangingPunctuation]("hanging-punctuation", hangingPunctuationNames, nil, s)
}

// -----------------------------------------------------------------------------
// Direction
// -----------------------------------------------------------------------------

var directionNames = []string{"ltr", "rtl"}

// String returns the CSS keyword for the direction value.
func (d Direction) String() string { return enumString("Direction", directionNames, d) }

// ParseDirection parses a direction keyword.
// Spec: https://www.w3.org/TR/css-writing-modes-3/#propdef-direction
func ParseDirection(s string) (Direction, error) {
	return parseEnum[Direction]("direction", directionNames, nil, s)
}

// -----------------------------------------------------------------------------
// WritingMode
// -----------------------------------------------------------------------------

var writingModeNames = []string{"horizontal-tb", "vertical-rl", "vertical-lr", "sideways-rl", "sideways-lr"}

// String returns the CSS keyword for the writing-mode value.
func (w WritingMode) String() string { return enumString("WritingMode", writingModeNames, w) }

// ParseWritingMode parses a writing-mode keyword.
// Spec: https://www.w3.org/TR/css-writing-modes-3/#propdef-writing-mode
func ParseWritingMode(s string) (WritingMode, error) {
	return parseEnum[WritingMode]("writing-mode", writingModeNames, nil, s)
}

// -----------------------------------------------------------------------------
// FontStyle
// -----------------------------------------------------------------------------

var fontStyleNames = []string{"normal", "italic", "oblique"}

// String returns the CSS keyword for the font-style value.
func (f FontStyle) String() string { return enumString("FontStyle", fontStyleNames, f) }

// ParseFontStyle parses a font-style keyword.
// Spec: https://www.w3.org/TR/css-fonts-4/#font-style-prop
func ParseFontStyle(s string) (FontStyle, error) {
	return parseEnum[FontStyle]("font-style", fontStyleNames, nil, s)
}

// -----------------------------------------------------------------------------
// TextDecorationStyle
// -----------------------------------------------------------------------------

var textDecorationStyleNames = []string{"solid", "double", "dotted", "dashed", "wavy"}

// String returns the CSS keyword for the text-decoration-style value.
func (t TextDecorationStyle) String() string {
	return enumString("TextDecorationStyle", textDecorationStyleNames, t)
}

// ParseTextDecorationStyle parses a text-decoration-style keyword.
// Spec: https://www.w3.org/TR/css-text-decor-3/#text-decoration-style-property
func ParseTextDecorationStyle(s string) (TextDecorationStyle, error) {
	return parseEnum[TextDecorationStyle]("text-decoration-style", textDecorationStyleNames, nil, s)
}

// -----------------------------------------------------------------------------
// VerticalAlign
// -----------------------------------------------------------------------------

var verticalAlignNames = []string{
	"baseline", "sub", "super", "text-top", "text-bottom", "middle", "top", "bottom",
}

// String returns the CSS keyword for the vertical-align value.
func (v VerticalAlign) String() string {
	return enumString("VerticalAlign", verticalAlignNames, v)
}

// ParseVerticalAlign parses a vertical-align keyword.
// Spec: https://www.w3.org/TR/css-inline-3/#propdef-vertical-align
func ParseVerticalAlign(s string) (VerticalAlign, error) {
	return parseEnum[VerticalAlign]("vertical-align", verticalAlignNames, nil, s)
}

// -----------------------------------------------------------------------------
// IntrinsicSize
// -----------------------------------------------------------------------------

var intrinsicSizeNames = []string{"none", "min-content", "max-content", "fit-content"}

// String returns the CSS keyword for the intrinsic sizing mode.
// IntrinsicSizeNone is "none" (the Width/Height field is used instead).
func (i IntrinsicSize) String() string {
	return enumString("IntrinsicSize", intrinsicSizeNames, i)
}

// ParseIntrinsicSize parses an intrinsic sizing keyword: none, min-content,
// max-content, fit-content.
// Spec: https://www.w3.org/TR/css-sizing-3/#sizing-values
func ParseIntrinsicSize(s string) (IntrinsicSize, error) {
	return parseEnum[IntrinsicSize]("intrinsic-size", intrinsicSizeNames, nil, s)
}

// -----------------------------------------------------------------------------
// InlineBoxKind
// -----------------------------------------------------------------------------

var inlineBoxKindNames = []string{"text"}

// String returns the keyword for the inline box kind ("text").
func (k InlineBoxKind) String() string {
	return enumString("InlineBoxKind", inlineBoxKindNames, k)
}

// ParseInlineBoxKind parses an inline box kind keyword.
func ParseInlineBoxKind(s string) (InlineBoxKind, error) {
	return parseEnum[InlineBoxKind]("inline-box-kind", inlineBoxKindNames, nil, s)
}
