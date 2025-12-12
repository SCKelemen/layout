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
	FlexGrow       float64
	FlexShrink     float64
	FlexBasis      float64 // or "auto" represented as -1
	FlexGap        float64 // Gap between flex items (0 means no gap)
	FlexRowGap     float64 // Row gap (cross-axis gap, 0 means use FlexGap)
	FlexColumnGap  float64 // Column gap (main-axis gap, 0 means use FlexGap)
	Order          int     // Visual order (default: 0). Items are ordered by ascending order value.

	// Grid properties
	GridTemplateRows    []GridTrack
	GridTemplateColumns []GridTrack
	GridAutoRows        GridTrack
	GridAutoColumns     GridTrack
	GridAutoFlow        GridAutoFlow // Auto-placement algorithm (default: row)
	GridGap             float64
	GridRowGap          float64
	GridColumnGap       float64
	GridRowStart        int          // -1 means auto
	GridRowEnd          int          // -1 means auto
	GridColumnStart     int          // -1 means auto
	GridColumnEnd       int          // -1 means auto
	JustifyItems        JustifyItems // Alignment along inline (row) axis. Default: Stretch
	JustifySelf         JustifyItems // Per-item inline-axis alignment override (0 = use parent's JustifyItems)
	// AlignItems is used for both Flexbox and Grid (block/column axis alignment)
	// For Grid: Default is Stretch, but Start for items with aspect-ratio
	// AlignSelf (defined in Flexbox section) also works for Grid items

	// Sizing
	Width       float64 // -1 means auto
	Height      float64 // -1 means auto
	MinWidth    float64
	MinHeight   float64
	MaxWidth    float64
	MaxHeight   float64
	AspectRatio float64 // Width/Height ratio (0 means not set). Example: 16/9 = 1.777...
	Padding     Spacing
	Margin      Spacing // Margin is supported in Flexbox and Grid layouts
	Border      Spacing

	// Box model
	BoxSizing BoxSizing

	// Positioning
	Position Position
	Top      float64 // -1 means auto
	Right    float64 // -1 means auto
	Bottom   float64 // -1 means auto
	Left     float64 // -1 means auto
	ZIndex   int     // Stacking order

	// Transform (for SVG rendering and visual effects)
	Transform Transform

	// TextStyle contains text-specific properties (nil for non-text nodes).
	// Based on CSS Text Module Level 3: https://www.w3.org/TR/css-text-3/
	TextStyle *TextStyle
}

// Spacing represents spacing on all sides
type Spacing struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

// Uniform creates uniform spacing on all sides
func Uniform(value float64) Spacing {
	return Spacing{
		Top:    value,
		Right:  value,
		Bottom: value,
		Left:   value,
	}
}

// Horizontal creates horizontal spacing
func Horizontal(value float64) Spacing {
	return Spacing{
		Top:    0,
		Right:  value,
		Bottom: 0,
		Left:   value,
	}
}

// Vertical creates vertical spacing
func Vertical(value float64) Spacing {
	return Spacing{
		Top:    value,
		Right:  0,
		Bottom: value,
		Left:   0,
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
	GridAutoFlowRow GridAutoFlow = iota // Default: row-major, sequential
	GridAutoFlowColumn                  // Column-major, sequential
	GridAutoFlowRowDense                // Row-major with dense packing
	GridAutoFlowColumnDense             // Column-major with dense packing
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
	TextJustifyAuto TextJustify = iota // Browser chooses (we use inter-word)
	TextJustifyInterWord               // Expand spaces between words only
	TextJustifyInterCharacter          // Expand spaces between characters
	TextJustifyDistribute              // Like inter-character but optimized for CJK
	TextJustifyNone                    // Disable justification
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
	TextOverflowClip TextOverflow = iota // Clip at content edge (default)
	TextOverflowEllipsis                 // Show ellipsis (...) for overflow
)

// OverflowWrap controls breaking of long words
// CSS Text Module Level 3 §5.3: https://www.w3.org/TR/css-text-3/#overflow-wrap-property
type OverflowWrap int

const (
	OverflowWrapNormal OverflowWrap = iota // Break only at allowed break points
	OverflowWrapBreakWord                  // Break anywhere if word would overflow
	OverflowWrapAnywhere                   // Like break-word but affects intrinsic sizing
)

// WordBreak controls word breaking behavior
// CSS Text Module Level 3 §5.4: https://www.w3.org/TR/css-text-3/#word-break-property
type WordBreak int

const (
	WordBreakNormal WordBreak = iota // Break at word boundaries (default)
	WordBreakBreakAll                // Break between any characters
	WordBreakKeepAll                 // Don't break between CJK characters
)

// Direction controls text direction.
// Based on CSS Writing Modes Level 3: https://www.w3.org/TR/css-writing-modes-3/#propdef-direction
type Direction int

const (
	DirectionLTR Direction = iota // CSS default (zero value)
	// DirectionRTL deferred
)

// FontWeight represents font weight (numeric or named).
type FontWeight int

const (
	FontWeightNormal FontWeight = 400
	FontWeightBold   FontWeight = 700
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

	// Font (for measurement)
	FontSize   float64
	FontFamily string
	FontWeight FontWeight

	// Direction (§2) - LTR only for v1
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
	Boxes           []InlineBox
	Width           float64
	SpaceCount      int     // Number of inter-word spaces (for justify)
	SpaceWidth      float64 // Total width of all spaces (for justify)
	SpaceAdjustment float64 // Extra pixels to add per space (for justify)
	OffsetX         float64 // X offset for text-align
	OffsetY         float64 // Y position (cumulative)
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
}

// GridTrack represents a grid track (row or column)
type GridTrack struct {
	MinSize float64
	MaxSize float64
	// For fr units, we'll use a ratio
	Fraction float64 // 0 means not a fraction
}

// FixedTrack creates a fixed-size track
func FixedTrack(size float64) GridTrack {
	return GridTrack{
		MinSize:  size,
		MaxSize:  size,
		Fraction: 0,
	}
}

// MinMaxTrack creates a minmax track
func MinMaxTrack(min, max float64) GridTrack {
	return GridTrack{
		MinSize:  min,
		MaxSize:  max,
		Fraction: 0,
	}
}

// FractionTrack creates a fractional track (fr unit)
func FractionTrack(fraction float64) GridTrack {
	return GridTrack{
		MinSize:  0,
		MaxSize:  Unbounded,
		Fraction: fraction,
	}
}

// AutoTrack creates an auto-sized track
func AutoTrack() GridTrack {
	return GridTrack{
		MinSize:  0,
		MaxSize:  Unbounded,
		Fraction: 0,
	}
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
func getHorizontalPaddingBorder(padding, border Spacing) float64 {
	return padding.Left + padding.Right + border.Left + border.Right
}

// getVerticalPaddingBorder returns the sum of vertical padding and border
func getVerticalPaddingBorder(padding, border Spacing) float64 {
	return padding.Top + padding.Bottom + border.Top + border.Bottom
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
