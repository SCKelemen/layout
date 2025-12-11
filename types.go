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
	FlexGrow       float64
	FlexShrink     float64
	FlexBasis      float64 // or "auto" represented as -1

	// Grid properties
	GridTemplateRows    []GridTrack
	GridTemplateColumns []GridTrack
	GridAutoRows        GridTrack
	GridAutoColumns     GridTrack
	GridGap             float64
	GridRowGap          float64
	GridColumnGap       float64
	GridRowStart        int // -1 means auto
	GridRowEnd          int // -1 means auto
	GridColumnStart     int // -1 means auto
	GridColumnEnd       int // -1 means auto

	// Sizing
	Width     float64 // -1 means auto
	Height    float64 // -1 means auto
	MinWidth  float64
	MinHeight float64
	MaxWidth  float64
	MaxHeight float64
	Padding   Spacing
	Margin    Spacing // Margin is supported in Flexbox and Grid layouts
	Border    Spacing

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
	AlignItemsFlexStart AlignItems = iota
	AlignItemsFlexEnd
	AlignItemsCenter
	AlignItemsStretch
	AlignItemsBaseline
)

// AlignContent
type AlignContent int

const (
	AlignContentFlexStart AlignContent = iota
	AlignContentFlexEnd
	AlignContentCenter
	AlignContentStretch
	AlignContentSpaceBetween
	AlignContentSpaceAround
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
