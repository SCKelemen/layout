package layout

import (
	"fmt"
	"math"
)

// Size represents a 2D size with width and height.
type Size struct {
	Width  float64
	Height float64
}

// Point represents a 2D point with X and Y coordinates.
type Point struct {
	X float64
	Y float64
}

// Rect represents a rectangle with position (X, Y) and size (Width, Height).
// This is the computed layout result stored in Node.Rect after calling Layout.
type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// Constraints represent the available space for layout.
// MinWidth/MinHeight specify the minimum size, MaxWidth/MaxHeight specify the maximum.
// Use Tight, Loose, or Unconstrained helpers to create constraints.
type Constraints struct {
	MinWidth  float64
	MaxWidth  float64
	MinHeight float64
	MaxHeight float64
}

// Unbounded represents an unbounded constraint
const Unbounded = math.MaxFloat64

// Tight creates tight constraints (min == max)
func Tight(width, height float64) Constraints {
	return Constraints{
		MinWidth:  width,
		MaxWidth:  width,
		MinHeight: height,
		MaxHeight: height,
	}
}

// Loose creates loose constraints (min == 0, max == provided)
func Loose(width, height float64) Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  width,
		MinHeight: 0,
		MaxHeight: height,
	}
}

// Unconstrained creates unconstrained constraints
func Unconstrained() Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  Unbounded,
		MinHeight: 0,
		MaxHeight: Unbounded,
	}
}

// Constrain applies constraints to a size
func (c Constraints) Constrain(size Size) Size {
	return Size{
		Width:  math.Max(c.MinWidth, math.Min(c.MaxWidth, size.Width)),
		Height: math.Max(c.MinHeight, math.Min(c.MaxHeight, size.Height)),
	}
}

// Node represents a layout node in the tree.
// Each node has a Style (layout properties) and Children (child nodes).
// After calling Layout, the Rect field contains the computed position and size.
//
// You can embed Node in your own types to add domain-specific data:
//
//	type Card struct {
//	    layout.Node
//	    Title string
//	}
type Node struct {
	// Style contains all layout properties (display, flex, grid, sizing, etc.)
	Style Style

	// Rect contains the computed layout position and size after calling Layout.
	// Do not modify this directly - it's set by the layout algorithms.
	Rect Rect

	// Children are the child nodes in the layout tree.
	Children []*Node

	// Baseline is the distance from the top of the node to its baseline.
	// Used for baseline alignment in flexbox and grid.
	// A value of 0 means no baseline is set (use default behavior).
	// For text-like elements, this would be the text baseline.
	// For containers, this is typically the baseline of the first child.
	Baseline float64

	// Text contains text content for text leaf nodes (DisplayInlineText).
	// Empty string means this is not a text node.
	Text string

	// TextLayout contains line box information populated by LayoutText.
	// Used by renderers to position text. Nil for non-text nodes.
	TextLayout *TextLayout
}

// Style contains CSS-like layout properties
type Style struct {
	// Display mode
	Display Display

	// Flexbox properties
	FlexDirection  FlexDirection
	FlexWrap       FlexWrap
	JustifyContent JustifyContent
	AlignItems     AlignItems
	AlignContent   AlignContent
	AlignSelf      AlignItems // Per-item cross-axis alignment override (0 = use parent's AlignItems)
	FlexGrow       float64    // Flex grow factor (unitless)
	FlexShrink     float64    // Flex shrink factor (unitless)
	FlexBasis      Length     // Initial main size (use Px(0) with WidthSizing/HeightSizing for auto)
	FlexGap        Length     // Gap between flex items (use Px(0) for no gap)
	FlexRowGap     Length     // Row gap (cross-axis gap, use Px(0) to fall back to FlexGap)
	FlexColumnGap  Length     // Column gap (main-axis gap, use Px(0) to fall back to FlexGap)
	Order          int        // Visual order (default: 0). Items are ordered by ascending order value.

	// Grid properties
	GridTemplateRows    []GridTrack
	GridTemplateColumns []GridTrack
	GridAutoRows        GridTrack
	GridAutoColumns     GridTrack
	GridAutoFlow        GridAutoFlow       // Auto-placement algorithm (default: row)
	GridGap             Length             // Gap between grid tracks (use Px(0) for no gap)
	GridRowGap          Length             // Row gap (use Px(0) to fall back to GridGap)
	GridColumnGap       Length             // Column gap (use Px(0) to fall back to GridGap)
	GridRowStart        int                // -1 means auto
	GridRowEnd          int                // -1 means auto
	GridColumnStart     int                // -1 means auto
	GridColumnEnd       int                // -1 means auto
	GridTemplateAreas   *GridTemplateAreas // Named grid areas (nil means not set)
	GridArea            string             // Name of the grid area this item should be placed in (empty means not set)
	JustifyItems        JustifyItems       // Alignment along inline (row) axis. Default: Stretch
	JustifySelf         JustifyItems       // Per-item inline-axis alignment override (0 = use parent's JustifyItems)
	// AlignItems is used for both Flexbox and Grid (block/column axis alignment)
	// For Grid: Default is Stretch, but Start for items with aspect-ratio
	// AlignSelf (defined in Flexbox section) also works for Grid items

	// Sizing
	Width       Length  // Explicit width (use WidthSizing for auto/min-content/max-content/fit-content)
	Height      Length  // Explicit height (use HeightSizing for auto/min-content/max-content/fit-content)
	MinWidth    Length  // Minimum width
	MinHeight   Length  // Minimum height
	MaxWidth    Length  // Maximum width
	MaxHeight   Length  // Maximum height
	AspectRatio float64 // Width/Height ratio (0 means not set). Example: 16/9 = 1.777...

	// Intrinsic sizing (alternative to sentinel values for better API ergonomics)
	// These provide an alternative way to specify intrinsic sizing beyond sentinel values
	WidthSizing      IntrinsicSize // Intrinsic sizing mode for width (0 = none/use Width value)
	HeightSizing     IntrinsicSize // Intrinsic sizing mode for height (0 = none/use Height value)
	FitContentWidth  Length        // Maximum width for fit-content (only used when WidthSizing = IntrinsicSizeFitContent)
	FitContentHeight Length        // Maximum height for fit-content (only used when HeightSizing = IntrinsicSizeFitContent)

	Padding Spacing
	Margin  Spacing // Margin is supported in Flexbox and Grid layouts
	Border  Spacing

	// Box model
	BoxSizing BoxSizing

	// Positioning
	Position Position
	Top      Length // Positioning offset (use Px(0) for zero, check for auto via separate logic)
	Right    Length // Positioning offset
	Bottom   Length // Positioning offset
	Left     Length // Positioning offset
	ZIndex   int    // Stacking order

	// Transform (for SVG rendering and visual effects)
	Transform Transform

	// WritingMode controls the block flow direction for layout containers.
	// Inherited property that applies to all elements (block, flex, grid, text).
	// Based on CSS Writing Modes Level 3: https://www.w3.org/TR/css-writing-modes-3/
	// Default: WritingModeHorizontalTB (zero value)
	WritingMode WritingMode

	// TextStyle contains text-specific properties (nil for non-text nodes).
	// Based on CSS Text Module Level 3: https://www.w3.org/TR/css-text-3/
	// Note: TextStyle.WritingMode is deprecated; use Style.WritingMode instead for inheritance.
	TextStyle *TextStyle
}

// Spacing represents spacing on all sides using Length values
type Spacing struct {
	Top    Length
	Right  Length
	Bottom Length
	Left   Length
}

// Uniform creates uniform spacing on all sides
func Uniform(value Length) Spacing {
	return Spacing{
		Top:    value,
		Right:  value,
		Bottom: value,
		Left:   value,
	}
}

// Horizontal creates horizontal spacing
func Horizontal(value Length) Spacing {
	return Spacing{
		Top:    Px(0),
		Right:  value,
		Bottom: Px(0),
		Left:   value,
	}
}

// Vertical creates vertical spacing
func Vertical(value Length) Spacing {
	return Spacing{
		Top:    value,
		Right:  Px(0),
		Bottom: value,
		Left:   Px(0),
	}
}

// Display mode
type Display int

const (
	DisplayBlock Display = iota
	DisplayFlex
	DisplayGrid
	DisplayInlineText // Text leaf node
	DisplayNone
)

// FlexDirection
type FlexDirection int

const (
	FlexDirectionRow FlexDirection = iota
	FlexDirectionRowReverse
	FlexDirectionColumn
	FlexDirectionColumnReverse
)

// FlexWrap
type FlexWrap int

const (
	FlexWrapNoWrap FlexWrap = iota
	FlexWrapWrap
	FlexWrapWrapReverse
)

// JustifyContent
type JustifyContent int

const (
	JustifyContentFlexStart JustifyContent = iota
	JustifyContentFlexEnd
	JustifyContentCenter
	JustifyContentSpaceBetween
	JustifyContentSpaceAround
	JustifyContentSpaceEvenly
)

// AlignItems
type AlignItems int

const (
	AlignItemsStretch AlignItems = iota // CSS default (zero value) - same for Grid and Flexbox
	AlignItemsFlexStart
	AlignItemsFlexEnd
	AlignItemsCenter
	AlignItemsBaseline
)

// JustifyItems controls alignment along the inline (row) axis in Grid
// Used for justify-items property in CSS Grid
type JustifyItems int

const (
	JustifyItemsStretch JustifyItems = iota // CSS Grid default (zero value)
	JustifyItemsStart
	JustifyItemsEnd
	JustifyItemsCenter
)

// AlignContent
type AlignContent int

const (
	AlignContentStretch AlignContent = iota // Zero value is CSS default
	AlignContentFlexStart
	AlignContentFlexEnd
	AlignContentCenter
	AlignContentSpaceBetween
	AlignContentSpaceAround
)

// GridAutoFlow controls the auto-placement algorithm for grid items
// See: https://www.w3.org/TR/css-grid-1/#grid-auto-flow-property
type GridAutoFlow int

const (
	GridAutoFlowRow         GridAutoFlow = iota // Default: row-major, sequential
	GridAutoFlowColumn                          // Column-major, sequential
	GridAutoFlowRowDense                        // Row-major with dense packing
	GridAutoFlowColumnDense                     // Column-major with dense packing
)

// BoxSizing
type BoxSizing int

const (
	BoxSizingContentBox BoxSizing = iota
	BoxSizingBorderBox
)

// Position
type Position int

const (
	PositionStatic Position = iota
	PositionRelative
	PositionAbsolute
	PositionFixed
	PositionSticky
)

// TextAlign controls horizontal alignment of text within line boxes.
// Based on CSS Text Module Level 3 §7.1: https://www.w3.org/TR/css-text-3/#text-align-property
type TextAlign int

const (
	TextAlignDefault TextAlign = iota // CSS default: 'start' (contextual - left in LTR)
	TextAlignLeft
	TextAlignRight
	TextAlignCenter
	TextAlignJustify // Stretches text to fill the line width (§7.1.1)
)

// TextAlignLast controls alignment of the last line in a block
// CSS Text Module Level 3 §7.2.2: https://www.w3.org/TR/css-text-3/#text-align-last-property
type TextAlignLast int

const (
	TextAlignLastAuto TextAlignLast = iota // Follow text-align (but not justify)
	TextAlignLastLeft
	TextAlignLastRight
	TextAlignLastCenter
	TextAlignLastJustify // Also justify the last line
)

// TextJustify controls the justification algorithm
// CSS Text Module Level 3 §7.3: https://www.w3.org/TR/css-text-3/#text-justify-property
type TextJustify int

const (
	TextJustifyAuto           TextJustify = iota // Browser chooses (we use inter-word)
	TextJustifyInterWord                         // Expand spaces between words only
	TextJustifyInterCharacter                    // Expand spaces between characters
	TextJustifyDistribute                        // Like inter-character but optimized for CJK
	TextJustifyNone                              // Disable justification
)

// WhiteSpace controls how whitespace and line breaks are handled.
// Based on CSS Text Module Level 3 §3.1: https://www.w3.org/TR/css-text-3/#white-space-property
type WhiteSpace int

const (
	WhiteSpaceNormal WhiteSpace = iota // CSS default (zero value)
	WhiteSpaceNowrap
	WhiteSpacePre
	WhiteSpacePreWrap // Preserve whitespace, allow wrapping
	WhiteSpacePreLine // Collapse whitespace, preserve newlines, allow wrapping
)

// TextOverflow controls rendering of overflowing text
// CSS Text Overflow Module Level 3: https://www.w3.org/TR/css-overflow-3/#text-overflow
type TextOverflow int

const (
	TextOverflowClip     TextOverflow = iota // Clip at content edge (default)
	TextOverflowEllipsis                     // Show ellipsis (...) for overflow
)

// OverflowWrap controls breaking of long words
// CSS Text Module Level 3 §5.3: https://www.w3.org/TR/css-text-3/#overflow-wrap-property
type OverflowWrap int

const (
	OverflowWrapNormal    OverflowWrap = iota // Break only at allowed break points
	OverflowWrapBreakWord                     // Break anywhere if word would overflow
	OverflowWrapAnywhere                      // Like break-word but affects intrinsic sizing
)

// WordBreak controls word breaking behavior
// CSS Text Module Level 3 §5.4: https://www.w3.org/TR/css-text-3/#word-break-property
type WordBreak int

const (
	WordBreakNormal   WordBreak = iota // Break at word boundaries (default)
	WordBreakBreakAll                  // Break between any characters
	WordBreakKeepAll                   // Don't break between CJK characters
)

// TextTransform controls text case transformation
// CSS Text Module Level 3 §6: https://www.w3.org/TR/css-text-3/#text-transform-property
type TextTransform int

const (
	TextTransformNone         TextTransform = iota // No transformation (default)
	TextTransformUppercase                         // Convert to uppercase
	TextTransformLowercase                         // Convert to lowercase
	TextTransformCapitalize                        // Capitalize first letter of each word
	TextTransformFullWidth                         // Convert to full-width characters
	TextTransformFullSizeKana                      // Convert kana to full-size
)

// Hyphens controls automatic hyphenation
// CSS Text Module Level 3 §4.3: https://www.w3.org/TR/css-text-3/#hyphenation
type Hyphens int

const (
	HyphensNone   Hyphens = iota // No hyphenation (default)
	HyphensManual                // Only hyphenate at U+00AD soft hyphens
	HyphensAuto                  // Automatic hyphenation with dictionaries
)

// HangingPunctuation controls punctuation placement
// CSS Text Module Level 3 §9.2: https://www.w3.org/TR/css-text-3/#hanging-punctuation-property
type HangingPunctuation int

const (
	HangingPunctuationNone     HangingPunctuation = iota // No hanging (default)
	HangingPunctuationFirst                              // Hang opening punctuation
	HangingPunctuationLast                               // Hang closing punctuation
	HangingPunctuationForceEnd                           // Force hang end punctuation
	HangingPunctuationAllowEnd                           // Allow hang end punctuation
)

// Direction controls text direction.
// Based on CSS Writing Modes Level 3: https://www.w3.org/TR/css-writing-modes-3/#propdef-direction
type Direction int

const (
	DirectionLTR Direction = iota // CSS default (zero value)
	DirectionRTL                  // Right-to-left
)

// WritingMode controls the block flow direction and inline base direction.
// Based on CSS Writing Modes Level 3: https://www.w3.org/TR/css-writing-modes-3/#propdef-writing-mode
//
// The writing-mode property defines whether lines of text are laid out horizontally or vertically,
// and the direction in which blocks progress. It affects the mapping of CSS logical properties
// to physical dimensions.
//
// In horizontal writing modes (horizontal-tb):
//   - Inline dimension = width (text flows left-to-right or right-to-left)
//   - Block dimension = height (blocks flow top-to-bottom)
//   - Lines are stacked vertically
//
// In vertical writing modes (vertical-rl, vertical-lr):
//   - Inline dimension = height (text flows top-to-bottom)
//   - Block dimension = width (blocks flow right-to-left or left-to-right)
//   - Lines are stacked horizontally
//
// Character orientation in vertical modes is determined by UAX #50 (Unicode Vertical Text Layout).
//
// Note: Terminal rendering has limitations for vertical modes (no character rotation),
// but SVG rendering can properly handle vertical text with UAX #50 character orientation.
type WritingMode int

const (
	// WritingModeHorizontalTB is horizontal top-to-bottom writing mode.
	// Text flows horizontally (LTR or RTL based on Direction property),
	// blocks progress from top to bottom.
	// This is the default for Latin, Greek, Cyrillic, Arabic, Hebrew, and most scripts.
	WritingModeHorizontalTB WritingMode = iota // CSS default (zero value)

	// WritingModeVerticalRL is vertical right-to-left writing mode.
	// Text flows vertically top-to-bottom, blocks progress from right to left.
	// This is the traditional mode for Chinese, Japanese, and Korean.
	// Character orientation follows UAX #50:
	//   - CJK ideographs remain upright
	//   - Latin characters are rotated 90° clockwise
	WritingModeVerticalRL

	// WritingModeVerticalLR is vertical left-to-right writing mode.
	// Text flows vertically top-to-bottom, blocks progress from left to right.
	// Used in Mongolian script and some modern CJK layouts.
	// Character orientation follows UAX #50.
	WritingModeVerticalLR

	// WritingModeSidewaysRL is sideways right-to-left writing mode (Level 4).
	// Similar to vertical-rl but all characters (including CJK) are rotated sideways.
	// Text flows left-to-right but rotated 90° clockwise.
	// Less common; used for artistic effects or specific languages.
	WritingModeSidewaysRL

	// WritingModeSidewaysLR is sideways left-to-right writing mode (Level 4).
	// Similar to vertical-lr but all characters are rotated sideways.
	// Text flows left-to-right but rotated 90° counter-clockwise.
	WritingModeSidewaysLR
)

// IsVertical returns true if the writing mode is vertical (vertical-rl, vertical-lr, sideways-rl, or sideways-lr).
func (w WritingMode) IsVertical() bool {
	return w != WritingModeHorizontalTB
}

// IsHorizontal returns true if the writing mode is horizontal-tb.
func (w WritingMode) IsHorizontal() bool {
	return w == WritingModeHorizontalTB
}

// IsSideways returns true if the writing mode is sideways-rl or sideways-lr.
func (w WritingMode) IsSideways() bool {
	return w == WritingModeSidewaysRL || w == WritingModeSidewaysLR
}

// FontWeight represents font weight (numeric or named).
type FontWeight int

const (
	FontWeightNormal FontWeight = 400
	FontWeightBold   FontWeight = 700
)

// FontStyle represents the font-style CSS property.
// Based on CSS Fonts Module Level 4: https://www.w3.org/TR/css-fonts-4/#font-style-prop
type FontStyle int

const (
	FontStyleNormal  FontStyle = iota // Normal (upright) font face
	FontStyleItalic                   // Italic font face (cursive)
	FontStyleOblique                  // Oblique font face (slanted)
)

// TextDecoration represents which text decoration lines are present.
// Based on CSS Text Decoration Module Level 3: https://www.w3.org/TR/css-text-decor-3/#text-decoration-line-property
// Multiple decorations can be combined using bitwise OR.
type TextDecoration int

const (
	TextDecorationNone        TextDecoration = 0      // No decoration
	TextDecorationUnderline   TextDecoration = 1 << 0 // Underline
	TextDecorationOverline    TextDecoration = 1 << 1 // Overline (line above)
	TextDecorationLineThrough TextDecoration = 1 << 2 // Line through middle (strikethrough)
)

// Has checks if a specific decoration is present.
func (td TextDecoration) Has(decoration TextDecoration) bool {
	return td&decoration != 0
}

// TextDecorationStyle represents the style of text decoration lines.
// Based on CSS Text Decoration Module Level 3: https://www.w3.org/TR/css-text-decor-3/#text-decoration-style-property
type TextDecorationStyle int

const (
	TextDecorationStyleSolid  TextDecorationStyle = iota // Solid line (default)
	TextDecorationStyleDouble                             // Double line
	TextDecorationStyleDotted                             // Dotted line
	TextDecorationStyleDashed                             // Dashed line
	TextDecorationStyleWavy                               // Wavy line
)

// VerticalAlign represents the vertical-align CSS property for inline elements.
// Based on CSS Inline Layout Module Level 3: https://www.w3.org/TR/css-inline-3/#propdef-vertical-align
type VerticalAlign int

const (
	VerticalAlignBaseline   VerticalAlign = iota // Align baseline with parent baseline (default)
	VerticalAlignSub                             // Lower baseline (subscript)
	VerticalAlignSuper                           // Raise baseline (superscript)
	VerticalAlignTextTop                         // Align top with parent's text top
	VerticalAlignTextBottom                      // Align bottom with parent's text bottom
	VerticalAlignMiddle                          // Align middle with parent's middle
	VerticalAlignTop                             // Align top with line box top
	VerticalAlignBottom                          // Align bottom with line box bottom
)

// TextStyle contains text-specific style properties.
// Based on CSS Text Module Level 3: https://www.w3.org/TR/css-text-3/
type TextStyle struct {
	// Alignment (§7.1, §7.2.2, §7.3)
	TextAlign     TextAlign
	TextAlignLast TextAlignLast // Controls alignment of the last line
	TextJustify   TextJustify   // Controls justification algorithm

	// Spacing (§4.4.1, §5.1, §5.2, §7.2.1)
	// LineHeight: <=0 = normal (1.2×), 0<x<10 = multiplier, >=10 = absolute px
	// Note: This heuristic means line-height: 12 will be 12px regardless of font size
	LineHeight    float64
	WordSpacing   float64 // -1 = normal, otherwise spacing in px (can be negative)
	LetterSpacing float64 // -1 = normal, otherwise spacing in px (can be negative)
	TextIndent    float64 // First line indent in px (0 = none, can be negative for hanging indent)

	// Wrapping (§3.1, §5.3, §5.4)
	WhiteSpace   WhiteSpace
	OverflowWrap OverflowWrap // Controls breaking of long words
	WordBreak    WordBreak    // Controls word breaking behavior
	TextOverflow TextOverflow // Controls rendering of overflowing text

	// Text Transformation (§6)
	TextTransform TextTransform

	// Hyphenation (§4.3)
	Hyphens Hyphens

	// Punctuation (§9.2)
	HangingPunctuation HangingPunctuation

	// Tab Size (§3.1.1) - Number of spaces per tab character
	// -1 = default (8 spaces), otherwise number of spaces
	TabSize float64

	// Font (for measurement)
	FontSize   float64
	FontFamily string
	FontWeight FontWeight
	FontStyle  FontStyle

	// Text Decoration (CSS Text Decoration Module Level 3)
	TextDecoration      TextDecoration      // Which decoration lines to show
	TextDecorationStyle TextDecorationStyle // Style of decoration lines
	TextDecorationColor string              // Color of decoration (CSS color string, "" = currentColor)

	// Vertical Alignment (CSS Inline Layout Module Level 3)
	VerticalAlign VerticalAlign

	// Writing Mode (CSS Writing Modes Level 3 §3.1)
	// Determines whether text flows horizontally or vertically,
	// and the direction in which blocks progress.
	// Default is WritingModeHorizontalTB (zero value).
	WritingMode WritingMode

	// Direction (CSS Writing Modes Level 3 §2.1)
	// Determines inline base direction (LTR or RTL).
	// Works with WritingMode to determine text flow.
	Direction Direction
}

// TextLayout contains line box information for text nodes.
// Populated by LayoutText, used by renderers to position text.
type TextLayout struct {
	Lines      []TextLine
	LineHeight float64
}

// TextLine represents a single line of text with its boxes and positioning.
type TextLine struct {
	Boxes               []InlineBox
	Width               float64
	SpaceCount          int     // Number of inter-word spaces (for justify)
	SpaceWidth          float64 // Total width of all spaces (for justify)
	SpaceAdjustment     float64 // Extra pixels to add per space (for justify)
	CharacterAdjustment float64 // Extra pixels to add between characters (for inter-character justify)
	OffsetX             float64 // X offset for text-align
	OffsetY             float64 // Y position (cumulative)
}

// InlineBoxKind represents the type of inline box.
type InlineBoxKind int

const (
	InlineBoxText InlineBoxKind = iota
	// InlineBoxInlineNode deferred (for future: spans, inline images)
)

// InlineBox represents a single inline box (text run or inline element).
type InlineBox struct {
	Kind    InlineBoxKind
	Text    string // for InlineBoxText
	Node    *Node  // for InlineBoxInlineNode (future)
	Width   float64
	Ascent  float64
	Descent float64

	// Orientations stores character orientation for vertical writing modes.
	// Length matches the number of runes in Text.
	// true = upright (natural vertical orientation, e.g., CJK)
	// false = rotated 90° (e.g., Latin characters in vertical text)
	// Empty slice for horizontal modes (no rotation needed).
	//
	// Based on Unicode UAX #50: Unicode Vertical Text Layout
	// See: https://www.unicode.org/reports/tr50/
	Orientations []bool
}

// IntrinsicSize represents intrinsic sizing keywords from CSS Sizing Module Level 3.
// These control how content-based sizing is calculated.
//
// See: CSS Sizing Module Level 3 §4-5 (Intrinsic Sizes)
// https://www.w3.org/TR/css-sizing-3/#intrinsic-sizes
type IntrinsicSize int

const (
	IntrinsicSizeNone       IntrinsicSize = 0 // Not using intrinsic sizing
	IntrinsicSizeMinContent IntrinsicSize = 1 // min-content: narrowest width without overflow
	IntrinsicSizeMaxContent IntrinsicSize = 2 // max-content: widest natural width (no wrapping)
	IntrinsicSizeFitContent IntrinsicSize = 3 // fit-content: clamp max-content to specified size
)

// Sentinel values for Width/Height to indicate intrinsic sizing.
// These are distinct from -1 (auto) to maintain backward compatibility.
//
// Usage:
//
//	node.Style.Width = SizeMinContent  // Use min-content width
//	node.Style.Width = SizeMaxContent  // Use max-content width
const (
	SizeMinContent = -2.0 // Use min-content intrinsic size
	SizeMaxContent = -3.0 // Use max-content intrinsic size
	SizeFitContent = -4.0 // Use fit-content intrinsic size (requires FitContent* field set)
)

// GridTrack represents a grid track (row or column)
type GridTrack struct {
	MinSize  Length  // Minimum track size
	MaxSize  Length  // Maximum track size (use PxUnbounded or UnboundedLength() for unbounded)
	Fraction float64 // For fr units (0 means not a fraction)
}

// FixedTrack creates a fixed-size track
func FixedTrack(size Length) GridTrack {
	return GridTrack{
		MinSize:  size,
		MaxSize:  size,
		Fraction: 0,
	}
}

// MinMaxTrack creates a minmax track
func MinMaxTrack(min, max Length) GridTrack {
	return GridTrack{
		MinSize:  min,
		MaxSize:  max,
		Fraction: 0,
	}
}

// FractionTrack creates a fractional track (fr unit)
func FractionTrack(fraction float64) GridTrack {
	return GridTrack{
		MinSize:  Px(0),
		MaxSize:  PxUnbounded,
		Fraction: fraction,
	}
}

// AutoTrack creates an auto-sized track
func AutoTrack() GridTrack {
	return GridTrack{
		MinSize:  Px(0),
		MaxSize:  PxUnbounded,
		Fraction: 0,
	}
}

// RepeatTrack represents a repeating track pattern for grid templates
// Used with auto-fill and auto-fit grid track generation (Feature 4)
type RepeatTrack struct {
	Count  int         // Number of repetitions, or special values (RepeatCountAutoFill, RepeatCountAutoFit)
	Tracks []GridTrack // Track pattern to repeat
}

// RepeatCount constants for auto-fill and auto-fit
const (
	RepeatCountAutoFill = -1 // Auto-fill: create as many tracks as fit
	RepeatCountAutoFit  = -2 // Auto-fit: auto-fill + collapse empty tracks
)

// GridArea represents a named grid region defined by row and column boundaries.
// Used with grid-template-areas for structured grid layout.
//
// See: CSS Grid Layout Module Level 1 §7.3 (Named Areas)
// https://www.w3.org/TR/css-grid-1/#grid-template-areas-property
type GridArea struct {
	Name        string // Name of the area (e.g., "header", "sidebar", "content")
	RowStart    int    // Starting row index (0-based)
	RowEnd      int    // Ending row index (exclusive)
	ColumnStart int    // Starting column index (0-based)
	ColumnEnd   int    // Ending column index (exclusive)
}

// GridTemplateAreas defines a collection of named grid areas.
// This provides a structured, type-safe alternative to CSS string-based template areas.
//
// Example:
//
//	areas := NewGridTemplateAreas(3, 3)
//	areas.DefineArea("header", 0, 1, 0, 3)  // Row 0, columns 0-3 (full width)
//	areas.DefineArea("sidebar", 1, 3, 0, 1) // Rows 1-3, column 0 (left side)
//	areas.DefineArea("content", 1, 3, 1, 3) // Rows 1-3, columns 1-3 (main area)
type GridTemplateAreas struct {
	Areas []GridArea // List of defined named areas
	Rows  int        // Total number of rows in the grid
	Cols  int        // Total number of columns in the grid
}

// Transform represents a 2D transformation matrix
// Used for rotating, scaling, translating, and skewing elements
// Useful for SVG rendering and visual effects
type Transform struct {
	// 2x3 transformation matrix (affine transform)
	// [a c e]   [x]   [a*x + c*y + e]
	// [b d f] * [y] = [b*x + d*y + f]
	// [0 0 1]   [1]   [1]
	A, B, C, D, E, F float64
}

// IdentityTransform returns an identity transformation (no transform)
func IdentityTransform() Transform {
	return Transform{
		A: 1, B: 0,
		C: 0, D: 1,
		E: 0, F: 0,
	}
}

// Translate creates a translation transform
func Translate(x, y float64) Transform {
	return Transform{
		A: 1, B: 0,
		C: 0, D: 1,
		E: x, F: y,
	}
}

// Scale creates a scaling transform
func Scale(sx, sy float64) Transform {
	return Transform{
		A: sx, B: 0,
		C: 0, D: sy,
		E: 0, F: 0,
	}
}

// Rotate creates a rotation transform (angle in radians)
func Rotate(angle float64) Transform {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Transform{
		A: cos, B: sin,
		C: -sin, D: cos,
		E: 0, F: 0,
	}
}

// RotateDegrees creates a rotation transform (angle in degrees)
func RotateDegrees(angle float64) Transform {
	return Rotate(angle * math.Pi / 180.0)
}

// SkewX creates a horizontal skew transform (angle in radians)
func SkewX(angle float64) Transform {
	return Transform{
		A: 1, B: 0,
		C: math.Tan(angle), D: 1,
		E: 0, F: 0,
	}
}

// SkewY creates a vertical skew transform (angle in radians)
func SkewY(angle float64) Transform {
	return Transform{
		A: 1, B: math.Tan(angle),
		C: 0, D: 1,
		E: 0, F: 0,
	}
}

// Matrix creates a transform from a 2x3 matrix
func Matrix(a, b, c, d, e, f float64) Transform {
	return Transform{A: a, B: b, C: c, D: d, E: e, F: f}
}

// Multiply multiplies two transforms (applies t2 after t1)
func (t1 Transform) Multiply(t2 Transform) Transform {
	return Transform{
		A: t1.A*t2.A + t1.C*t2.B,
		B: t1.B*t2.A + t1.D*t2.B,
		C: t1.A*t2.C + t1.C*t2.D,
		D: t1.B*t2.C + t1.D*t2.D,
		E: t1.A*t2.E + t1.C*t2.F + t1.E,
		F: t1.B*t2.E + t1.D*t2.F + t1.F,
	}
}

// Apply applies the transform to a point
func (t Transform) Apply(p Point) Point {
	return Point{
		X: t.A*p.X + t.C*p.Y + t.E,
		Y: t.B*p.X + t.D*p.Y + t.F,
	}
}

// ApplyToRect applies the transform to a rectangle's corners
// Returns the bounding box of the transformed rectangle
func (t Transform) ApplyToRect(r Rect) Rect {
	// Transform all four corners
	corners := []Point{
		{X: r.X, Y: r.Y},
		{X: r.X + r.Width, Y: r.Y},
		{X: r.X + r.Width, Y: r.Y + r.Height},
		{X: r.X, Y: r.Y + r.Height},
	}

	transformed := make([]Point, len(corners))
	for i, corner := range corners {
		transformed[i] = t.Apply(corner)
	}

	// Find bounding box
	minX := transformed[0].X
	maxX := transformed[0].X
	minY := transformed[0].Y
	maxY := transformed[0].Y

	for _, p := range transformed[1:] {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	return Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

// ToSVGString returns the transform as an SVG transform attribute string
func (t Transform) ToSVGString() string {
	if t.IsIdentity() {
		return ""
	}
	return fmt.Sprintf("matrix(%g,%g,%g,%g,%g,%g)", t.A, t.B, t.C, t.D, t.E, t.F)
}

// IsIdentity checks if the transform is an identity (no transformation)
func (t Transform) IsIdentity() bool {
	return t.A == 1 && t.B == 0 && t.C == 0 && t.D == 1 && t.E == 0 && t.F == 0
}

// getHorizontalPaddingBorder returns the sum of horizontal padding and border
func getHorizontalPaddingBorder(padding, border Spacing, ctx *LayoutContext, currentFontSize float64) float64 {
	paddingLeft := ResolveLength(padding.Left, ctx, currentFontSize)
	paddingRight := ResolveLength(padding.Right, ctx, currentFontSize)
	borderLeft := ResolveLength(border.Left, ctx, currentFontSize)
	borderRight := ResolveLength(border.Right, ctx, currentFontSize)
	return paddingLeft + paddingRight + borderLeft + borderRight
}

// getVerticalPaddingBorder returns the sum of vertical padding and border
func getVerticalPaddingBorder(padding, border Spacing, ctx *LayoutContext, currentFontSize float64) float64 {
	paddingTop := ResolveLength(padding.Top, ctx, currentFontSize)
	paddingBottom := ResolveLength(padding.Bottom, ctx, currentFontSize)
	borderTop := ResolveLength(border.Top, ctx, currentFontSize)
	borderBottom := ResolveLength(border.Bottom, ctx, currentFontSize)
	return paddingTop + paddingBottom + borderTop + borderBottom
}

// convertToContentSize converts a width/height from border-box to content-box
// If boxSizing is content-box, returns the value unchanged
// If boxSizing is border-box, subtracts padding and border to get content size
func convertToContentSize(size float64, boxSizing BoxSizing, horizontalPaddingBorder, verticalPaddingBorder float64, isWidth bool) float64 {
	if size < 0 {
		// Auto values are passed through unchanged
		return size
	}
	if boxSizing == BoxSizingBorderBox {
		// border-box: size includes padding + border, so subtract to get content size
		if isWidth {
			return size - horizontalPaddingBorder
		} else {
			return size - verticalPaddingBorder
		}
	}
	// content-box: size is already content size
	return size
}

// convertFromContentSize converts a content size to the appropriate box-sizing format
// If boxSizing is content-box, returns content size unchanged
// If boxSizing is border-box, adds padding and border to get total size
func convertFromContentSize(contentSize float64, boxSizing BoxSizing, horizontalPaddingBorder, verticalPaddingBorder float64, isWidth bool) float64 {
	if contentSize < 0 {
		// Auto values are passed through unchanged
		return contentSize
	}
	if boxSizing == BoxSizingBorderBox {
		// border-box: add padding + border to get total size
		if isWidth {
			return contentSize + horizontalPaddingBorder
		} else {
			return contentSize + verticalPaddingBorder
		}
	}
	// content-box: content size is the total size
	return contentSize
}

// convertMinMaxToContentSize converts min/max constraints from border-box to content-box
// Min/Max constraints in CSS are always interpreted as border-box when box-sizing is border-box
func convertMinMaxToContentSize(size float64, boxSizing BoxSizing, horizontalPaddingBorder, verticalPaddingBorder float64, isWidth bool) float64 {
	if size <= 0 {
		// 0 or negative values are passed through unchanged
		return size
	}
	if boxSizing == BoxSizingBorderBox {
		// border-box: min/max includes padding + border, so subtract to get content size
		if isWidth {
			converted := size - horizontalPaddingBorder
			// Clamp to >= 0 to prevent negative content sizes
			if converted < 0 {
				return 0
			}
			return converted
		} else {
			converted := size - verticalPaddingBorder
			if converted < 0 {
				return 0
			}
			return converted
		}
	}
	// content-box: min/max is already content size
	return size
}
