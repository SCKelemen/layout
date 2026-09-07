// Package serialize converts layout trees to and from JSON (and optionally
// YAML). The wire format is documented in README.md.
//
// Deserialization treats its input as untrusted: numeric values are checked
// for NaN, infinities, and absurd magnitudes, enum strings must be known,
// and the tree depth and per-node child count are capped.
package serialize

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/SCKelemen/layout"
	"github.com/SCKelemen/units"
)

// Input limits enforced by FromJSON and FromYAML (and, for depth, by ToJSON
// and ToYAML so a cyclic tree fails instead of overflowing the stack).
const (
	// MaxTreeDepth is the maximum nesting depth of nodes (root has depth 0).
	MaxTreeDepth = 1024
	// MaxChildren is the maximum number of children a single node may have.
	MaxChildren = 65536
	// MaxNumericValue bounds the magnitude of every finite numeric input.
	// The only exception is the unbounded sentinel: the "unbounded" string,
	// the legacy bare number math.MaxFloat64 (see legacyNumberLength), and
	// math.MaxFloat64 in Rect fields (see checkRect).
	MaxNumericValue = 1e12
	// MaxRepeatCount is the largest integer repetition count accepted for a
	// repeat() track pattern (gridTemplateRowsRepeat / gridTemplateColumnsRepeat).
	MaxRepeatCount = 10000
)

// ErrNullInput is returned by FromJSON and FromYAML when the document is a
// bare null (or, for YAML, empty), which has no layout meaning.
var ErrNullInput = errors.New("serialize: input is null")

// ErrLimitExceeded is wrapped by errors returned when the input exceeds
// MaxTreeDepth or MaxChildren.
var ErrLimitExceeded = errors.New("serialize: input exceeds size limit")

// NodeJSON represents a serializable version of layout.Node
type NodeJSON struct {
	Style    StyleJSON   `json:"style" yaml:"style"`
	Text     string      `json:"text,omitempty" yaml:"text,omitempty"`
	Baseline float64     `json:"baseline,omitempty" yaml:"baseline,omitempty"`
	Children []*NodeJSON `json:"children,omitempty" yaml:"children,omitempty"`
	Rect     *RectJSON   `json:"rect,omitempty" yaml:"rect,omitempty"`
}

// StyleJSON represents a serializable version of layout.Style
type StyleJSON struct {
	Display        string     `json:"display,omitempty" yaml:"display,omitempty"`
	FlexDirection  string     `json:"flexDirection,omitempty" yaml:"flexDirection,omitempty"`
	FlexWrap       string     `json:"flexWrap,omitempty" yaml:"flexWrap,omitempty"`
	JustifyContent string     `json:"justifyContent,omitempty" yaml:"justifyContent,omitempty"`
	AlignItems     string     `json:"alignItems,omitempty" yaml:"alignItems,omitempty"`
	AlignContent   string     `json:"alignContent,omitempty" yaml:"alignContent,omitempty"`
	AlignSelf      string     `json:"alignSelf,omitempty" yaml:"alignSelf,omitempty"`
	JustifyItems   string     `json:"justifyItems,omitempty" yaml:"justifyItems,omitempty"`
	JustifySelf    string     `json:"justifySelf,omitempty" yaml:"justifySelf,omitempty"`
	FlexGrow       float64    `json:"flexGrow,omitempty" yaml:"flexGrow,omitempty"`
	FlexShrink     float64    `json:"flexShrink,omitempty" yaml:"flexShrink,omitempty"`
	FlexBasis      LengthJSON `json:"flexBasis,omitempty" yaml:"flexBasis,omitempty"`
	FlexGap        LengthJSON `json:"flexGap,omitempty" yaml:"flexGap,omitempty"`
	FlexRowGap     LengthJSON `json:"flexRowGap,omitempty" yaml:"flexRowGap,omitempty"`
	FlexColumnGap  LengthJSON `json:"flexColumnGap,omitempty" yaml:"flexColumnGap,omitempty"`
	Order          int        `json:"order,omitempty" yaml:"order,omitempty"`

	// Grid
	GridTemplateRows    []TrackJSON            `json:"gridTemplateRows,omitempty" yaml:"gridTemplateRows,omitempty"`
	GridTemplateColumns []TrackJSON            `json:"gridTemplateColumns,omitempty" yaml:"gridTemplateColumns,omitempty"`
	GridAutoRows        *TrackJSON             `json:"gridAutoRows,omitempty" yaml:"gridAutoRows,omitempty"`
	GridAutoColumns     *TrackJSON             `json:"gridAutoColumns,omitempty" yaml:"gridAutoColumns,omitempty"`
	GridAutoFlow        string                 `json:"gridAutoFlow,omitempty" yaml:"gridAutoFlow,omitempty"`
	GridGap             LengthJSON             `json:"gridGap,omitempty" yaml:"gridGap,omitempty"`
	GridRowGap          LengthJSON             `json:"gridRowGap,omitempty" yaml:"gridRowGap,omitempty"`
	GridColumnGap       LengthJSON             `json:"gridColumnGap,omitempty" yaml:"gridColumnGap,omitempty"`
	GridRowStart        int                    `json:"gridRowStart,omitempty" yaml:"gridRowStart,omitempty"`
	GridRowEnd          int                    `json:"gridRowEnd,omitempty" yaml:"gridRowEnd,omitempty"`
	GridColumnStart     int                    `json:"gridColumnStart,omitempty" yaml:"gridColumnStart,omitempty"`
	GridColumnEnd       int                    `json:"gridColumnEnd,omitempty" yaml:"gridColumnEnd,omitempty"`
	GridTemplateAreas   *GridTemplateAreasJSON `json:"gridTemplateAreas,omitempty" yaml:"gridTemplateAreas,omitempty"`
	GridArea            string                 `json:"gridArea,omitempty" yaml:"gridArea,omitempty"`
	// Repeat patterns appended after the explicit tracks; see RepeatJSON.
	GridTemplateRowsRepeat    []RepeatJSON `json:"gridTemplateRowsRepeat,omitempty" yaml:"gridTemplateRowsRepeat,omitempty"`
	GridTemplateColumnsRepeat []RepeatJSON `json:"gridTemplateColumnsRepeat,omitempty" yaml:"gridTemplateColumnsRepeat,omitempty"`

	// Sizing
	Width            LengthJSON `json:"width,omitempty" yaml:"width,omitempty"`
	Height           LengthJSON `json:"height,omitempty" yaml:"height,omitempty"`
	MinWidth         LengthJSON `json:"minWidth,omitempty" yaml:"minWidth,omitempty"`
	MinHeight        LengthJSON `json:"minHeight,omitempty" yaml:"minHeight,omitempty"`
	MaxWidth         LengthJSON `json:"maxWidth,omitempty" yaml:"maxWidth,omitempty"`
	MaxHeight        LengthJSON `json:"maxHeight,omitempty" yaml:"maxHeight,omitempty"`
	AspectRatio      float64    `json:"aspectRatio,omitempty" yaml:"aspectRatio,omitempty"`
	WidthSizing      string     `json:"widthSizing,omitempty" yaml:"widthSizing,omitempty"`
	HeightSizing     string     `json:"heightSizing,omitempty" yaml:"heightSizing,omitempty"`
	FitContentWidth  LengthJSON `json:"fitContentWidth,omitempty" yaml:"fitContentWidth,omitempty"`
	FitContentHeight LengthJSON `json:"fitContentHeight,omitempty" yaml:"fitContentHeight,omitempty"`

	// Spacing
	Padding *SpacingJSON `json:"padding,omitempty" yaml:"padding,omitempty"`
	Margin  *SpacingJSON `json:"margin,omitempty" yaml:"margin,omitempty"`
	Border  *SpacingJSON `json:"border,omitempty" yaml:"border,omitempty"`

	// Box model
	BoxSizing string `json:"boxSizing,omitempty" yaml:"boxSizing,omitempty"`

	// Positioning
	Position string     `json:"position,omitempty" yaml:"position,omitempty"`
	Top      LengthJSON `json:"top,omitempty" yaml:"top,omitempty"`
	Right    LengthJSON `json:"right,omitempty" yaml:"right,omitempty"`
	Bottom   LengthJSON `json:"bottom,omitempty" yaml:"bottom,omitempty"`
	Left     LengthJSON `json:"left,omitempty" yaml:"left,omitempty"`
	ZIndex   int        `json:"zIndex,omitempty" yaml:"zIndex,omitempty"`

	// Transform
	Transform *TransformJSON `json:"transform,omitempty" yaml:"transform,omitempty"`

	// Writing mode, direction, and container queries
	WritingMode   string   `json:"writingMode,omitempty" yaml:"writingMode,omitempty"`
	Direction     string   `json:"direction,omitempty" yaml:"direction,omitempty"`
	ContainerType string   `json:"containerType,omitempty" yaml:"containerType,omitempty"`
	ContainerName []string `json:"containerName,omitempty" yaml:"containerName,omitempty"`

	// Text
	TextStyle *TextStyleJSON `json:"textStyle,omitempty" yaml:"textStyle,omitempty"`
}

// LengthJSON is the wire form of layout.Length: a CSS-like string such as
// "10px", "1.5em", or "unbounded".
//
// The empty string is the zero-value layout.Length{} (no unit), which the
// engine treats as "not set" and which differs from "0px". Both
// layout.PxUnbounded and layout.UnboundedLength() are written as
// "unbounded" and read back as layout.PxUnbounded, the form the grid
// engine checks for.
//
// For backward compatibility with files written before lengths carried
// units, a bare JSON or YAML number is accepted and interpreted as pixels.
type LengthJSON string

// UnmarshalJSON accepts either a JSON string or a legacy bare number (px).
func (l *LengthJSON) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "null" {
		*l = ""
		return nil
	}
	if strings.HasPrefix(trimmed, `"`) {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*l = LengthJSON(s)
		return nil
	}
	var v float64
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("length: expected a string like \"10px\" or a number, got %s", trimmed)
	}
	// The previous encoder wrote unbounded lengths (the MaxSize of
	// FractionTrack and AutoTrack) as math.MaxFloat64. Map that sentinel to
	// "unbounded" here rather than formatting it as a 309-digit pixel
	// string that parseLength would reject. See legacyNumberLength.
	if v >= math.MaxFloat64 {
		*l = "unbounded"
		return nil
	}
	// Validation (NaN/Inf/magnitude) happens when the string is parsed.
	*l = LengthJSON(strconv.FormatFloat(v, 'f', -1, 64) + "px")
	return nil
}

// TrackJSON represents a serializable version of layout.GridTrack
type TrackJSON struct {
	MinSize  LengthJSON `json:"minSize,omitempty" yaml:"minSize,omitempty"`
	MaxSize  LengthJSON `json:"maxSize,omitempty" yaml:"maxSize,omitempty"`
	Fraction float64    `json:"fraction,omitempty" yaml:"fraction,omitempty"`
}

// RepeatJSON is the wire form of layout.RepeatTrack, the repeat() notation of
// CSS Grid (https://www.w3.org/TR/css-grid-1/#repeat-notation).
//
// Count is either an integer repetition count in [1, MaxRepeatCount], written
// as a JSON/YAML number, or one of the keywords "auto-fill" / "auto-fit",
// written as a string. Tracks is the pattern to repeat and must hold at least
// one track.
type RepeatJSON struct {
	Count  any         `json:"count" yaml:"count"`
	Tracks []TrackJSON `json:"tracks,omitempty" yaml:"tracks,omitempty"`
}

// Repeat-count keywords.
const (
	repeatAutoFill = "auto-fill"
	repeatAutoFit  = "auto-fit"
)

// SpacingJSON represents a serializable version of layout.Spacing
type SpacingJSON struct {
	Top    LengthJSON `json:"top,omitempty" yaml:"top,omitempty"`
	Right  LengthJSON `json:"right,omitempty" yaml:"right,omitempty"`
	Bottom LengthJSON `json:"bottom,omitempty" yaml:"bottom,omitempty"`
	Left   LengthJSON `json:"left,omitempty" yaml:"left,omitempty"`
}

// RectJSON represents a serializable version of layout.Rect
type RectJSON struct {
	X      float64 `json:"x" yaml:"x"`
	Y      float64 `json:"y" yaml:"y"`
	Width  float64 `json:"width" yaml:"width"`
	Height float64 `json:"height" yaml:"height"`
}

// TransformJSON represents a serializable version of layout.Transform
type TransformJSON struct {
	A float64 `json:"a" yaml:"a"`
	B float64 `json:"b" yaml:"b"`
	C float64 `json:"c" yaml:"c"`
	D float64 `json:"d" yaml:"d"`
	E float64 `json:"e" yaml:"e"`
	F float64 `json:"f" yaml:"f"`
}

// GridTemplateAreasJSON represents a serializable version of layout.GridTemplateAreas
type GridTemplateAreasJSON struct {
	Rows  int            `json:"rows" yaml:"rows"`
	Cols  int            `json:"cols" yaml:"cols"`
	Areas []GridAreaJSON `json:"areas,omitempty" yaml:"areas,omitempty"`
}

// GridAreaJSON represents a serializable version of layout.GridArea
type GridAreaJSON struct {
	Name        string `json:"name" yaml:"name"`
	RowStart    int    `json:"rowStart" yaml:"rowStart"`
	RowEnd      int    `json:"rowEnd" yaml:"rowEnd"`
	ColumnStart int    `json:"columnStart" yaml:"columnStart"`
	ColumnEnd   int    `json:"columnEnd" yaml:"columnEnd"`
}

// TextStyleJSON represents a serializable version of layout.TextStyle.
// FontWeight is numeric (CSS 1-1000) and TextDecoration is the bitmask
// value; every other enum is a CSS keyword string.
type TextStyleJSON struct {
	TextAlign           string  `json:"textAlign,omitempty" yaml:"textAlign,omitempty"`
	TextAlignLast       string  `json:"textAlignLast,omitempty" yaml:"textAlignLast,omitempty"`
	TextJustify         string  `json:"textJustify,omitempty" yaml:"textJustify,omitempty"`
	LineHeight          float64 `json:"lineHeight,omitempty" yaml:"lineHeight,omitempty"`
	WordSpacing         float64 `json:"wordSpacing,omitempty" yaml:"wordSpacing,omitempty"`
	LetterSpacing       float64 `json:"letterSpacing,omitempty" yaml:"letterSpacing,omitempty"`
	TextIndent          float64 `json:"textIndent,omitempty" yaml:"textIndent,omitempty"`
	WhiteSpace          string  `json:"whiteSpace,omitempty" yaml:"whiteSpace,omitempty"`
	OverflowWrap        string  `json:"overflowWrap,omitempty" yaml:"overflowWrap,omitempty"`
	WordBreak           string  `json:"wordBreak,omitempty" yaml:"wordBreak,omitempty"`
	TextOverflow        string  `json:"textOverflow,omitempty" yaml:"textOverflow,omitempty"`
	TextTransform       string  `json:"textTransform,omitempty" yaml:"textTransform,omitempty"`
	Hyphens             string  `json:"hyphens,omitempty" yaml:"hyphens,omitempty"`
	HangingPunctuation  string  `json:"hangingPunctuation,omitempty" yaml:"hangingPunctuation,omitempty"`
	TabSize             float64 `json:"tabSize,omitempty" yaml:"tabSize,omitempty"`
	FontSize            float64 `json:"fontSize,omitempty" yaml:"fontSize,omitempty"`
	FontFamily          string  `json:"fontFamily,omitempty" yaml:"fontFamily,omitempty"`
	FontWeight          int     `json:"fontWeight,omitempty" yaml:"fontWeight,omitempty"`
	FontStyle           string  `json:"fontStyle,omitempty" yaml:"fontStyle,omitempty"`
	TextDecoration      int     `json:"textDecoration,omitempty" yaml:"textDecoration,omitempty"`
	TextDecorationStyle string  `json:"textDecorationStyle,omitempty" yaml:"textDecorationStyle,omitempty"`
	TextDecorationColor string  `json:"textDecorationColor,omitempty" yaml:"textDecorationColor,omitempty"`
	VerticalAlign       string  `json:"verticalAlign,omitempty" yaml:"verticalAlign,omitempty"`
	WritingMode         string  `json:"writingMode,omitempty" yaml:"writingMode,omitempty"`
	Direction           string  `json:"direction,omitempty" yaml:"direction,omitempty"`
}

// ToJSON converts a layout.Node to JSON bytes
func ToJSON(node *layout.Node) ([]byte, error) {
	nodeJSON, err := nodeToJSON(node)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(nodeJSON, "", "  ")
}

// FromJSON converts JSON bytes to a layout.Node.
// It returns an error for malformed JSON, a bare null document
// (ErrNullInput), non-finite or absurd numbers, unknown enum keywords, or
// trees exceeding MaxTreeDepth/MaxChildren.
func FromJSON(data []byte) (*layout.Node, error) {
	var nodeJSON *NodeJSON
	if err := json.Unmarshal(data, &nodeJSON); err != nil {
		return nil, err
	}
	if nodeJSON == nil {
		return nil, ErrNullInput
	}
	return jsonToNode(nodeJSON)
}

// =============================================================================
// Encoding (layout -> wire)
// =============================================================================

// nodeToJSON converts a layout.Node to NodeJSON
func nodeToJSON(node *layout.Node) (*NodeJSON, error) {
	return encodeNode(node, "root", 0)
}

func encodeNode(node *layout.Node, path string, depth int) (*NodeJSON, error) {
	if node == nil {
		return nil, nil
	}
	if depth > MaxTreeDepth {
		return nil, fmt.Errorf("%w: %s: depth exceeds %d", ErrLimitExceeded, path, MaxTreeDepth)
	}
	if len(node.Children) > MaxChildren {
		return nil, fmt.Errorf("%w: %s: %d children exceeds %d", ErrLimitExceeded, path, len(node.Children), MaxChildren)
	}

	style, err := styleToJSON(&node.Style, path+".style")
	if err != nil {
		return nil, err
	}
	if err := checkFloat(path+".baseline", node.Baseline); err != nil {
		return nil, err
	}

	nj := &NodeJSON{
		Style:    style,
		Text:     node.Text,
		Baseline: node.Baseline,
	}
	if node.Rect != (layout.Rect{}) {
		if err := checkRect(path+".rect", &node.Rect); err != nil {
			return nil, err
		}
		nj.Rect = rectToJSON(&node.Rect)
	}

	if len(node.Children) > 0 {
		nj.Children = make([]*NodeJSON, len(node.Children))
		for i, child := range node.Children {
			nj.Children[i], err = encodeNode(child, fmt.Sprintf("%s.children[%d]", path, i), depth+1)
			if err != nil {
				return nil, err
			}
		}
	}

	return nj, nil
}

// styleToJSON converts layout.Style to StyleJSON
func styleToJSON(s *layout.Style, path string) (StyleJSON, error) {
	e := &encoder{path: path}

	sj := StyleJSON{
		Display:        encodeEnum(e, "display", s.Display, layout.ParseDisplay),
		FlexDirection:  encodeEnum(e, "flexDirection", s.FlexDirection, layout.ParseFlexDirection),
		FlexWrap:       encodeEnum(e, "flexWrap", s.FlexWrap, layout.ParseFlexWrap),
		JustifyContent: encodeEnum(e, "justifyContent", s.JustifyContent, layout.ParseJustifyContent),
		AlignItems:     encodeEnum(e, "alignItems", s.AlignItems, layout.ParseAlignItems),
		AlignContent:   encodeEnum(e, "alignContent", s.AlignContent, layout.ParseAlignContent),
		AlignSelf:      encodeEnum(e, "alignSelf", s.AlignSelf, layout.ParseAlignItems),
		JustifyItems:   encodeEnum(e, "justifyItems", s.JustifyItems, layout.ParseJustifyItems),
		JustifySelf:    encodeEnum(e, "justifySelf", s.JustifySelf, layout.ParseJustifyItems),
		FlexGrow:       e.float("flexGrow", s.FlexGrow),
		FlexShrink:     e.float("flexShrink", s.FlexShrink),
		FlexBasis:      e.length("flexBasis", s.FlexBasis),
		FlexGap:        e.length("flexGap", s.FlexGap),
		FlexRowGap:     e.length("flexRowGap", s.FlexRowGap),
		FlexColumnGap:  e.length("flexColumnGap", s.FlexColumnGap),
		Order:          s.Order,

		GridAutoFlow:    encodeEnum(e, "gridAutoFlow", s.GridAutoFlow, layout.ParseGridAutoFlow),
		GridGap:         e.length("gridGap", s.GridGap),
		GridRowGap:      e.length("gridRowGap", s.GridRowGap),
		GridColumnGap:   e.length("gridColumnGap", s.GridColumnGap),
		GridRowStart:    s.GridRowStart,
		GridRowEnd:      s.GridRowEnd,
		GridColumnStart: s.GridColumnStart,
		GridColumnEnd:   s.GridColumnEnd,
		GridArea:        s.GridArea,

		Width:            e.length("width", s.Width),
		Height:           e.length("height", s.Height),
		MinWidth:         e.length("minWidth", s.MinWidth),
		MinHeight:        e.length("minHeight", s.MinHeight),
		MaxWidth:         e.length("maxWidth", s.MaxWidth),
		MaxHeight:        e.length("maxHeight", s.MaxHeight),
		AspectRatio:      e.float("aspectRatio", s.AspectRatio),
		WidthSizing:      encodeEnum(e, "widthSizing", s.WidthSizing, layout.ParseIntrinsicSize),
		HeightSizing:     encodeEnum(e, "heightSizing", s.HeightSizing, layout.ParseIntrinsicSize),
		FitContentWidth:  e.length("fitContentWidth", s.FitContentWidth),
		FitContentHeight: e.length("fitContentHeight", s.FitContentHeight),

		BoxSizing: encodeEnum(e, "boxSizing", s.BoxSizing, layout.ParseBoxSizing),

		Position: encodeEnum(e, "position", s.Position, layout.ParsePosition),
		Top:      e.length("top", s.Top),
		Right:    e.length("right", s.Right),
		Bottom:   e.length("bottom", s.Bottom),
		Left:     e.length("left", s.Left),
		ZIndex:   s.ZIndex,

		WritingMode:   encodeEnum(e, "writingMode", s.WritingMode, layout.ParseWritingMode),
		Direction:     encodeEnum(e, "direction", s.Direction, layout.ParseDirection),
		ContainerType: encodeEnum(e, "containerType", s.ContainerType, layout.ParseContainerType),
	}

	// Struct-typed fields are pointers so that omitempty can drop them.
	if s.Padding != (layout.Spacing{}) {
		sj.Padding = e.spacing("padding", &s.Padding)
	}
	if s.Margin != (layout.Spacing{}) {
		sj.Margin = e.spacing("margin", &s.Margin)
	}
	if s.Border != (layout.Spacing{}) {
		sj.Border = e.spacing("border", &s.Border)
	}
	if !s.Transform.IsIdentity() {
		e.floats("transform", s.Transform.A, s.Transform.B, s.Transform.C, s.Transform.D, s.Transform.E, s.Transform.F)
		sj.Transform = transformToJSON(&s.Transform)
	}
	if s.GridAutoRows != (layout.GridTrack{}) {
		t := e.track("gridAutoRows", &s.GridAutoRows)
		sj.GridAutoRows = &t
	}
	if s.GridAutoColumns != (layout.GridTrack{}) {
		t := e.track("gridAutoColumns", &s.GridAutoColumns)
		sj.GridAutoColumns = &t
	}

	sj.GridTemplateRows = e.tracks("gridTemplateRows", s.GridTemplateRows)
	sj.GridTemplateColumns = e.tracks("gridTemplateColumns", s.GridTemplateColumns)
	sj.GridTemplateRowsRepeat = e.repeats("gridTemplateRowsRepeat", s.GridTemplateRowsRepeat)
	sj.GridTemplateColumnsRepeat = e.repeats("gridTemplateColumnsRepeat", s.GridTemplateColumnsRepeat)
	if s.GridTemplateAreas != nil {
		sj.GridTemplateAreas = gridTemplateAreasToJSON(s.GridTemplateAreas)
	}
	if len(s.ContainerName) > 0 {
		sj.ContainerName = append([]string(nil), s.ContainerName...)
	}
	if s.TextStyle != nil {
		sj.TextStyle = e.textStyle(s.TextStyle)
	}

	return sj, e.err
}

// encoder accumulates the first encoding error so the field-by-field
// struct literal above stays readable.
type encoder struct {
	path string
	err  error
}

func (e *encoder) fail(err error) {
	if e.err == nil {
		e.err = err
	}
}

// cssEnum is satisfied by every layout enum that has a String method and a
// matching Parse function in enums.go.
type cssEnum interface {
	~int
	String() string
}

// encodeEnum formats v with its String method. The zero value is written as
// the empty string so omitempty drops it. A value outside the enum's range
// formats as "Kind(n)", which its Parse function rejects; that round trip is
// how an unknown value is detected and reported.
func encodeEnum[T cssEnum](e *encoder, field string, v T, parse func(string) (T, error)) string {
	if v == 0 {
		return ""
	}
	s := v.String()
	if _, err := parse(s); err != nil {
		e.fail(fmt.Errorf("%s.%s: unknown value %d", e.path, field, int(v)))
	}
	return s
}

func (e *encoder) float(field string, v float64) float64 {
	if err := checkFloat(e.path+"."+field, v); err != nil {
		e.fail(err)
	}
	return v
}

func (e *encoder) floats(field string, vs ...float64) {
	for _, v := range vs {
		e.float(field, v)
	}
}

func (e *encoder) length(field string, l layout.Length) LengthJSON {
	s, err := lengthToJSON(l)
	if err != nil {
		e.fail(fmt.Errorf("%s.%s: %w", e.path, field, err))
	}
	return s
}

func (e *encoder) spacing(field string, s *layout.Spacing) *SpacingJSON {
	return &SpacingJSON{
		Top:    e.length(field+".top", s.Top),
		Right:  e.length(field+".right", s.Right),
		Bottom: e.length(field+".bottom", s.Bottom),
		Left:   e.length(field+".left", s.Left),
	}
}

func (e *encoder) track(field string, t *layout.GridTrack) TrackJSON {
	return TrackJSON{
		MinSize:  e.length(field+".minSize", t.MinSize),
		MaxSize:  e.length(field+".maxSize", t.MaxSize),
		Fraction: e.float(field+".fraction", t.Fraction),
	}
}

// tracks encodes a track list, applying the MaxChildren cap that decoding
// enforces so an encoded tree always reads back.
func (e *encoder) tracks(field string, ts []layout.GridTrack) []TrackJSON {
	if len(ts) == 0 {
		return nil
	}
	if len(ts) > MaxChildren {
		e.fail(fmt.Errorf("%w: %s.%s: %d tracks exceeds %d", ErrLimitExceeded, e.path, field, len(ts), MaxChildren))
		return nil
	}
	out := make([]TrackJSON, len(ts))
	for i := range ts {
		out[i] = e.track(fmt.Sprintf("%s[%d]", field, i), &ts[i])
	}
	return out
}

// repeats encodes repeat() patterns. See RepeatJSON for the wire form.
func (e *encoder) repeats(field string, rs []layout.RepeatTrack) []RepeatJSON {
	if len(rs) == 0 {
		return nil
	}
	if len(rs) > MaxChildren {
		e.fail(fmt.Errorf("%w: %s.%s: %d repeats exceeds %d", ErrLimitExceeded, e.path, field, len(rs), MaxChildren))
		return nil
	}
	out := make([]RepeatJSON, len(rs))
	for i, r := range rs {
		item := fmt.Sprintf("%s[%d]", field, i)
		var count any
		switch {
		case r.Count == layout.RepeatCountAutoFill:
			count = repeatAutoFill
		case r.Count == layout.RepeatCountAutoFit:
			count = repeatAutoFit
		case r.Count >= 1 && r.Count <= MaxRepeatCount:
			count = r.Count
		default:
			e.fail(fmt.Errorf("%s.%s.count: %d is not in [1, %d], auto-fill, or auto-fit", e.path, item, r.Count, MaxRepeatCount))
		}
		if len(r.Tracks) == 0 {
			e.fail(fmt.Errorf("%s.%s.tracks: repeat() needs at least one track", e.path, item))
		}
		out[i] = RepeatJSON{Count: count, Tracks: e.tracks(item+".tracks", r.Tracks)}
	}
	return out
}

func (e *encoder) textStyle(ts *layout.TextStyle) *TextStyleJSON {
	saved := e.path
	e.path += ".textStyle"
	defer func() { e.path = saved }()

	if ts.FontWeight < 0 || int(ts.FontWeight) > maxFontWeight {
		e.fail(fmt.Errorf("%s.fontWeight: %d out of range [0, %d]", e.path, ts.FontWeight, maxFontWeight))
	}
	if ts.TextDecoration < 0 || int(ts.TextDecoration) > maxTextDecoration {
		e.fail(fmt.Errorf("%s.textDecoration: %d out of range [0, %d]", e.path, ts.TextDecoration, maxTextDecoration))
	}
	return &TextStyleJSON{
		TextAlign:           encodeEnum(e, "textAlign", ts.TextAlign, layout.ParseTextAlign),
		TextAlignLast:       encodeEnum(e, "textAlignLast", ts.TextAlignLast, layout.ParseTextAlignLast),
		TextJustify:         encodeEnum(e, "textJustify", ts.TextJustify, layout.ParseTextJustify),
		LineHeight:          e.float("lineHeight", ts.LineHeight),
		WordSpacing:         e.float("wordSpacing", ts.WordSpacing),
		LetterSpacing:       e.float("letterSpacing", ts.LetterSpacing),
		TextIndent:          e.float("textIndent", ts.TextIndent),
		WhiteSpace:          encodeEnum(e, "whiteSpace", ts.WhiteSpace, layout.ParseWhiteSpace),
		OverflowWrap:        encodeEnum(e, "overflowWrap", ts.OverflowWrap, layout.ParseOverflowWrap),
		WordBreak:           encodeEnum(e, "wordBreak", ts.WordBreak, layout.ParseWordBreak),
		TextOverflow:        encodeEnum(e, "textOverflow", ts.TextOverflow, layout.ParseTextOverflow),
		TextTransform:       encodeEnum(e, "textTransform", ts.TextTransform, layout.ParseTextTransform),
		Hyphens:             encodeEnum(e, "hyphens", ts.Hyphens, layout.ParseHyphens),
		HangingPunctuation:  encodeEnum(e, "hangingPunctuation", ts.HangingPunctuation, layout.ParseHangingPunctuation),
		TabSize:             e.float("tabSize", ts.TabSize),
		FontSize:            e.float("fontSize", ts.FontSize),
		FontFamily:          ts.FontFamily,
		FontWeight:          int(ts.FontWeight),
		FontStyle:           encodeEnum(e, "fontStyle", ts.FontStyle, layout.ParseFontStyle),
		TextDecoration:      int(ts.TextDecoration),
		TextDecorationStyle: encodeEnum(e, "textDecorationStyle", ts.TextDecorationStyle, layout.ParseTextDecorationStyle),
		TextDecorationColor: ts.TextDecorationColor,
		VerticalAlign:       encodeEnum(e, "verticalAlign", ts.VerticalAlign, layout.ParseVerticalAlign),
		WritingMode:         encodeEnum(e, "writingMode", ts.WritingMode, layout.ParseWritingMode),
		Direction:           encodeEnum(e, "direction", ts.Direction, layout.ParseDirection),
	}
}

// lengthToJSON formats a Length as a CSS-like string. See LengthJSON.
func lengthToJSON(l layout.Length) (LengthJSON, error) {
	if l == (layout.Length{}) {
		return "", nil
	}
	if l.Unit == layout.UnboundedUnit || l.Value >= math.MaxFloat64 {
		return "unbounded", nil
	}
	if math.IsNaN(l.Value) || math.IsInf(l.Value, 0) {
		return "", fmt.Errorf("length value %v is not finite", l.Value)
	}
	if math.Abs(l.Value) > MaxNumericValue {
		return "", fmt.Errorf("length value %v exceeds %g", l.Value, MaxNumericValue)
	}
	unit := string(l.Unit)
	if unit == "" {
		// A non-zero value with no unit resolves as pixels; make that explicit.
		unit = string(layout.Pixels)
	}
	// 'f' formatting never produces an exponent, which units.ParseLength rejects.
	return LengthJSON(strconv.FormatFloat(l.Value, 'f', -1, 64) + unit), nil
}

func rectToJSON(r *layout.Rect) *RectJSON {
	return &RectJSON{X: r.X, Y: r.Y, Width: r.Width, Height: r.Height}
}

func transformToJSON(t *layout.Transform) *TransformJSON {
	return &TransformJSON{A: t.A, B: t.B, C: t.C, D: t.D, E: t.E, F: t.F}
}

func gridTemplateAreasToJSON(g *layout.GridTemplateAreas) *GridTemplateAreasJSON {
	gj := &GridTemplateAreasJSON{Rows: g.Rows, Cols: g.Cols}
	if len(g.Areas) > 0 {
		gj.Areas = make([]GridAreaJSON, len(g.Areas))
		for i, a := range g.Areas {
			gj.Areas[i] = GridAreaJSON{
				Name:        a.Name,
				RowStart:    a.RowStart,
				RowEnd:      a.RowEnd,
				ColumnStart: a.ColumnStart,
				ColumnEnd:   a.ColumnEnd,
			}
		}
	}
	return gj
}

// =============================================================================
// Decoding (wire -> layout)
// =============================================================================

// jsonToNode converts a NodeJSON to layout.Node, validating all input.
func jsonToNode(nj *NodeJSON) (*layout.Node, error) {
	return decodeNode(nj, "root", 0)
}

func decodeNode(nj *NodeJSON, path string, depth int) (*layout.Node, error) {
	if nj == nil {
		return nil, nil
	}
	if depth > MaxTreeDepth {
		return nil, fmt.Errorf("%w: %s: depth exceeds %d", ErrLimitExceeded, path, MaxTreeDepth)
	}
	if len(nj.Children) > MaxChildren {
		return nil, fmt.Errorf("%w: %s: %d children exceeds %d", ErrLimitExceeded, path, len(nj.Children), MaxChildren)
	}

	style, err := jsonToStyle(&nj.Style, path+".style")
	if err != nil {
		return nil, err
	}
	if err := checkFloat(path+".baseline", nj.Baseline); err != nil {
		return nil, err
	}

	node := &layout.Node{
		Style:    style,
		Text:     nj.Text,
		Baseline: nj.Baseline,
	}
	if nj.Rect != nil {
		node.Rect = layout.Rect{X: nj.Rect.X, Y: nj.Rect.Y, Width: nj.Rect.Width, Height: nj.Rect.Height}
		if err := checkRect(path+".rect", &node.Rect); err != nil {
			return nil, err
		}
	}

	if len(nj.Children) > 0 {
		node.Children = make([]*layout.Node, 0, len(nj.Children))
		for i, child := range nj.Children {
			c, err := decodeNode(child, fmt.Sprintf("%s.children[%d]", path, i), depth+1)
			if err != nil {
				return nil, err
			}
			if c == nil {
				// A JSON null in the children array has no layout meaning; skip it.
				continue
			}
			node.Children = append(node.Children, c)
		}
	}

	return node, nil
}

// jsonToStyle converts StyleJSON to layout.Style
func jsonToStyle(sj *StyleJSON, path string) (layout.Style, error) {
	d := &decoder{path: path}

	s := layout.Style{
		Display:        decodeEnum(d, "display", sj.Display, layout.ParseDisplay),
		FlexDirection:  decodeEnum(d, "flexDirection", sj.FlexDirection, layout.ParseFlexDirection),
		FlexWrap:       decodeEnum(d, "flexWrap", sj.FlexWrap, layout.ParseFlexWrap),
		JustifyContent: decodeEnum(d, "justifyContent", sj.JustifyContent, layout.ParseJustifyContent),
		AlignItems:     decodeEnum(d, "alignItems", sj.AlignItems, layout.ParseAlignItems),
		AlignContent:   decodeEnum(d, "alignContent", sj.AlignContent, layout.ParseAlignContent),
		AlignSelf:      decodeEnum(d, "alignSelf", sj.AlignSelf, layout.ParseAlignItems),
		JustifyItems:   decodeEnum(d, "justifyItems", sj.JustifyItems, layout.ParseJustifyItems),
		JustifySelf:    decodeEnum(d, "justifySelf", sj.JustifySelf, layout.ParseJustifyItems),
		FlexGrow:       d.float("flexGrow", sj.FlexGrow),
		FlexShrink:     d.float("flexShrink", sj.FlexShrink),
		FlexBasis:      d.length("flexBasis", sj.FlexBasis),
		FlexGap:        d.length("flexGap", sj.FlexGap),
		FlexRowGap:     d.length("flexRowGap", sj.FlexRowGap),
		FlexColumnGap:  d.length("flexColumnGap", sj.FlexColumnGap),
		Order:          sj.Order,

		GridAutoFlow:    decodeEnum(d, "gridAutoFlow", sj.GridAutoFlow, layout.ParseGridAutoFlow),
		GridGap:         d.length("gridGap", sj.GridGap),
		GridRowGap:      d.length("gridRowGap", sj.GridRowGap),
		GridColumnGap:   d.length("gridColumnGap", sj.GridColumnGap),
		GridRowStart:    sj.GridRowStart,
		GridRowEnd:      sj.GridRowEnd,
		GridColumnStart: sj.GridColumnStart,
		GridColumnEnd:   sj.GridColumnEnd,
		GridArea:        sj.GridArea,

		Width:            d.length("width", sj.Width),
		Height:           d.length("height", sj.Height),
		MinWidth:         d.length("minWidth", sj.MinWidth),
		MinHeight:        d.length("minHeight", sj.MinHeight),
		MaxWidth:         d.length("maxWidth", sj.MaxWidth),
		MaxHeight:        d.length("maxHeight", sj.MaxHeight),
		AspectRatio:      d.float("aspectRatio", sj.AspectRatio),
		WidthSizing:      decodeEnum(d, "widthSizing", sj.WidthSizing, layout.ParseIntrinsicSize),
		HeightSizing:     decodeEnum(d, "heightSizing", sj.HeightSizing, layout.ParseIntrinsicSize),
		FitContentWidth:  d.length("fitContentWidth", sj.FitContentWidth),
		FitContentHeight: d.length("fitContentHeight", sj.FitContentHeight),

		BoxSizing: decodeEnum(d, "boxSizing", sj.BoxSizing, layout.ParseBoxSizing),

		Position: decodeEnum(d, "position", sj.Position, layout.ParsePosition),
		Top:      d.length("top", sj.Top),
		Right:    d.length("right", sj.Right),
		Bottom:   d.length("bottom", sj.Bottom),
		Left:     d.length("left", sj.Left),
		ZIndex:   sj.ZIndex,

		WritingMode:   decodeEnum(d, "writingMode", sj.WritingMode, layout.ParseWritingMode),
		Direction:     decodeEnum(d, "direction", sj.Direction, layout.ParseDirection),
		ContainerType: decodeEnum(d, "containerType", sj.ContainerType, layout.ParseContainerType),
	}

	if sj.Padding != nil {
		s.Padding = d.spacing("padding", sj.Padding)
	}
	if sj.Margin != nil {
		s.Margin = d.spacing("margin", sj.Margin)
	}
	if sj.Border != nil {
		s.Border = d.spacing("border", sj.Border)
	}
	if sj.Transform != nil {
		d.floats("transform", sj.Transform.A, sj.Transform.B, sj.Transform.C, sj.Transform.D, sj.Transform.E, sj.Transform.F)
		s.Transform = layout.Transform{
			A: sj.Transform.A, B: sj.Transform.B,
			C: sj.Transform.C, D: sj.Transform.D,
			E: sj.Transform.E, F: sj.Transform.F,
		}
	}
	if sj.GridAutoRows != nil {
		s.GridAutoRows = d.track("gridAutoRows", sj.GridAutoRows)
	}
	if sj.GridAutoColumns != nil {
		s.GridAutoColumns = d.track("gridAutoColumns", sj.GridAutoColumns)
	}

	s.GridTemplateRows = d.tracks("gridTemplateRows", sj.GridTemplateRows)
	s.GridTemplateColumns = d.tracks("gridTemplateColumns", sj.GridTemplateColumns)
	s.GridTemplateRowsRepeat = d.repeats("gridTemplateRowsRepeat", sj.GridTemplateRowsRepeat)
	s.GridTemplateColumnsRepeat = d.repeats("gridTemplateColumnsRepeat", sj.GridTemplateColumnsRepeat)
	if sj.GridTemplateAreas != nil {
		s.GridTemplateAreas = d.gridTemplateAreas(sj.GridTemplateAreas)
	}
	if len(sj.ContainerName) > 0 {
		s.ContainerName = layout.ContainerName(append([]string(nil), sj.ContainerName...))
	}
	if sj.TextStyle != nil {
		s.TextStyle = d.textStyle(sj.TextStyle)
	}

	return s, d.err
}

// decoder accumulates the first validation error while a Style is rebuilt.
type decoder struct {
	path string
	err  error
}

func (d *decoder) fail(err error) {
	if d.err == nil {
		d.err = err
	}
}

// decodeEnum parses a keyword with the enum's Parse function. The empty
// string (field omitted) is the zero value; every other string must be a
// known keyword or alias.
func decodeEnum[T cssEnum](d *decoder, field string, s string, parse func(string) (T, error)) T {
	if s == "" {
		return 0
	}
	v, err := parse(s)
	if err != nil {
		d.fail(fmt.Errorf("%s.%s: %w", d.path, field, err))
	}
	return v
}

func (d *decoder) float(field string, v float64) float64 {
	if err := checkFloat(d.path+"."+field, v); err != nil {
		d.fail(err)
	}
	return v
}

func (d *decoder) floats(field string, vs ...float64) {
	for _, v := range vs {
		d.float(field, v)
	}
}

func (d *decoder) length(field string, l LengthJSON) layout.Length {
	v, err := parseLength(l)
	if err != nil {
		d.fail(fmt.Errorf("%s.%s: %w", d.path, field, err))
	}
	return v
}

func (d *decoder) spacing(field string, sj *SpacingJSON) layout.Spacing {
	return layout.Spacing{
		Top:    d.length(field+".top", sj.Top),
		Right:  d.length(field+".right", sj.Right),
		Bottom: d.length(field+".bottom", sj.Bottom),
		Left:   d.length(field+".left", sj.Left),
	}
}

func (d *decoder) track(field string, tj *TrackJSON) layout.GridTrack {
	return layout.GridTrack{
		MinSize:  d.length(field+".minSize", tj.MinSize),
		MaxSize:  d.length(field+".maxSize", tj.MaxSize),
		Fraction: d.float(field+".fraction", tj.Fraction),
	}
}

// tracks decodes a track list, capped at MaxChildren entries.
func (d *decoder) tracks(field string, tjs []TrackJSON) []layout.GridTrack {
	if len(tjs) == 0 {
		return nil
	}
	if len(tjs) > MaxChildren {
		d.fail(fmt.Errorf("%w: %s.%s: %d tracks exceeds %d", ErrLimitExceeded, d.path, field, len(tjs), MaxChildren))
		return nil
	}
	out := make([]layout.GridTrack, len(tjs))
	for i := range tjs {
		out[i] = d.track(fmt.Sprintf("%s[%d]", field, i), &tjs[i])
	}
	return out
}

// repeats decodes repeat() patterns. The count must be an integer in
// [1, MaxRepeatCount] or one of the keywords "auto-fill" / "auto-fit", and
// every pattern needs at least one track.
func (d *decoder) repeats(field string, rjs []RepeatJSON) []layout.RepeatTrack {
	if len(rjs) == 0 {
		return nil
	}
	if len(rjs) > MaxChildren {
		d.fail(fmt.Errorf("%w: %s.%s: %d repeats exceeds %d", ErrLimitExceeded, d.path, field, len(rjs), MaxChildren))
		return nil
	}
	out := make([]layout.RepeatTrack, len(rjs))
	for i, rj := range rjs {
		item := fmt.Sprintf("%s[%d]", field, i)
		count, err := parseRepeatCount(rj.Count)
		if err != nil {
			d.fail(fmt.Errorf("%s.%s.count: %w", d.path, item, err))
		}
		if len(rj.Tracks) == 0 {
			d.fail(fmt.Errorf("%s.%s.tracks: repeat() needs at least one track", d.path, item))
		}
		out[i] = layout.RepeatTrack{Count: count, Tracks: d.tracks(item+".tracks", rj.Tracks)}
	}
	return out
}

// parseRepeatCount interprets the untyped count of a RepeatJSON. JSON
// delivers numbers as float64; YAML delivers integers as int (or uint64 when
// they overflow int64) and floats as float64; both deliver keywords as string.
func parseRepeatCount(v any) (int, error) {
	switch c := v.(type) {
	case nil:
		return 0, fmt.Errorf("missing (expected 1..%d, %q, or %q)", MaxRepeatCount, repeatAutoFill, repeatAutoFit)
	case string:
		switch c {
		case repeatAutoFill:
			return layout.RepeatCountAutoFill, nil
		case repeatAutoFit:
			return layout.RepeatCountAutoFit, nil
		}
		return 0, fmt.Errorf("unknown keyword %q (expected 1..%d, %q, or %q)", c, MaxRepeatCount, repeatAutoFill, repeatAutoFit)
	case float64:
		if math.IsNaN(c) || math.IsInf(c, 0) || c != math.Trunc(c) {
			return 0, fmt.Errorf("%v is not an integer", c)
		}
		return checkRepeatCount(c)
	case int:
		return checkRepeatCount(float64(c))
	case int64:
		return checkRepeatCount(float64(c))
	case uint64:
		return checkRepeatCount(float64(c))
	default:
		return 0, fmt.Errorf("unexpected %T (expected a number or a keyword)", v)
	}
}

func checkRepeatCount(c float64) (int, error) {
	if c < 1 || c > MaxRepeatCount {
		return 0, fmt.Errorf("%v is not in [1, %d]", c, MaxRepeatCount)
	}
	return int(c), nil
}

func (d *decoder) gridTemplateAreas(gj *GridTemplateAreasJSON) *layout.GridTemplateAreas {
	field := d.path + ".gridTemplateAreas"
	if gj.Rows < 0 || gj.Cols < 0 || gj.Rows > MaxChildren || gj.Cols > MaxChildren {
		d.fail(fmt.Errorf("%s: rows/cols %d/%d out of range [0, %d]", field, gj.Rows, gj.Cols, MaxChildren))
	}
	if len(gj.Areas) > MaxChildren {
		d.fail(fmt.Errorf("%w: %s: %d areas exceeds %d", ErrLimitExceeded, field, len(gj.Areas), MaxChildren))
		return nil
	}
	g := &layout.GridTemplateAreas{Rows: gj.Rows, Cols: gj.Cols, Areas: make([]layout.GridArea, 0, len(gj.Areas))}
	for i, a := range gj.Areas {
		if a.RowStart < 0 || a.ColumnStart < 0 || a.RowEnd < a.RowStart || a.ColumnEnd < a.ColumnStart {
			d.fail(fmt.Errorf("%s.areas[%d]: invalid bounds rows %d..%d cols %d..%d", field, i, a.RowStart, a.RowEnd, a.ColumnStart, a.ColumnEnd))
		}
		g.Areas = append(g.Areas, layout.GridArea{
			Name:        a.Name,
			RowStart:    a.RowStart,
			RowEnd:      a.RowEnd,
			ColumnStart: a.ColumnStart,
			ColumnEnd:   a.ColumnEnd,
		})
	}
	return g
}

func (d *decoder) textStyle(tj *TextStyleJSON) *layout.TextStyle {
	saved := d.path
	d.path += ".textStyle"
	defer func() { d.path = saved }()

	if tj.FontWeight < 0 || tj.FontWeight > maxFontWeight {
		d.fail(fmt.Errorf("%s.fontWeight: %d out of range [0, %d]", d.path, tj.FontWeight, maxFontWeight))
	}
	if tj.TextDecoration < 0 || tj.TextDecoration > maxTextDecoration {
		d.fail(fmt.Errorf("%s.textDecoration: %d out of range [0, %d]", d.path, tj.TextDecoration, maxTextDecoration))
	}
	return &layout.TextStyle{
		TextAlign:           decodeEnum(d, "textAlign", tj.TextAlign, layout.ParseTextAlign),
		TextAlignLast:       decodeEnum(d, "textAlignLast", tj.TextAlignLast, layout.ParseTextAlignLast),
		TextJustify:         decodeEnum(d, "textJustify", tj.TextJustify, layout.ParseTextJustify),
		LineHeight:          d.float("lineHeight", tj.LineHeight),
		WordSpacing:         d.float("wordSpacing", tj.WordSpacing),
		LetterSpacing:       d.float("letterSpacing", tj.LetterSpacing),
		TextIndent:          d.float("textIndent", tj.TextIndent),
		WhiteSpace:          decodeEnum(d, "whiteSpace", tj.WhiteSpace, layout.ParseWhiteSpace),
		OverflowWrap:        decodeEnum(d, "overflowWrap", tj.OverflowWrap, layout.ParseOverflowWrap),
		WordBreak:           decodeEnum(d, "wordBreak", tj.WordBreak, layout.ParseWordBreak),
		TextOverflow:        decodeEnum(d, "textOverflow", tj.TextOverflow, layout.ParseTextOverflow),
		TextTransform:       decodeEnum(d, "textTransform", tj.TextTransform, layout.ParseTextTransform),
		Hyphens:             decodeEnum(d, "hyphens", tj.Hyphens, layout.ParseHyphens),
		HangingPunctuation:  decodeEnum(d, "hangingPunctuation", tj.HangingPunctuation, layout.ParseHangingPunctuation),
		TabSize:             d.float("tabSize", tj.TabSize),
		FontSize:            d.float("fontSize", tj.FontSize),
		FontFamily:          tj.FontFamily,
		FontWeight:          layout.FontWeight(tj.FontWeight),
		FontStyle:           decodeEnum(d, "fontStyle", tj.FontStyle, layout.ParseFontStyle),
		TextDecoration:      layout.TextDecoration(tj.TextDecoration),
		TextDecorationStyle: decodeEnum(d, "textDecorationStyle", tj.TextDecorationStyle, layout.ParseTextDecorationStyle),
		TextDecorationColor: tj.TextDecorationColor,
		VerticalAlign:       decodeEnum(d, "verticalAlign", tj.VerticalAlign, layout.ParseVerticalAlign),
		WritingMode:         decodeEnum(d, "writingMode", tj.WritingMode, layout.ParseWritingMode),
		Direction:           decodeEnum(d, "direction", tj.Direction, layout.ParseDirection),
	}
}

// parseLength parses the wire form of a Length. See LengthJSON.
func parseLength(l LengthJSON) (layout.Length, error) {
	s := strings.TrimSpace(string(l))
	if s == "" {
		return layout.Length{}, nil
	}
	if strings.EqualFold(s, "unbounded") {
		return layout.PxUnbounded, nil
	}
	// A bare number is the legacy pre-unit form. YAML delivers it as the raw
	// scalar text, which may carry an exponent (1.7976931348623157e+308)
	// that units.ParseLength would reject.
	if v, ok := parseBareNumber(s); ok {
		return legacyNumberLength(v)
	}
	// units.ParseLength rejects NaN, infinities, exponents, and unknown units.
	parsed, err := units.ParseLength(s)
	if err != nil {
		return layout.Length{}, err
	}
	if math.Abs(parsed.Value) > MaxNumericValue {
		return layout.Length{}, fmt.Errorf("length %q exceeds %g", s, MaxNumericValue)
	}
	return parsed, nil
}

// parseBareNumber reports whether s is a plain decimal number, optionally
// with an exponent, and returns its value. A literal that overflows float64
// comes back as +/-Inf. Words such as "inf" and "nan", hex floats, and
// anything carrying a unit are not bare numbers.
func parseBareNumber(s string) (float64, bool) {
	for _, r := range s {
		if (r < '0' || r > '9') && !strings.ContainsRune("+-.eE", r) {
			return 0, false
		}
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return 0, false
	}
	return v, true
}

// legacyNumberLength interprets a bare number from a file written before
// lengths carried units: the value is in pixels. The previous encoder stored
// unbounded lengths (the MaxSize of FractionTrack and AutoTrack) as
// math.MaxFloat64, so that value, and anything larger such as +Inf, is the
// unbounded sentinel and decodes to layout.PxUnbounded exactly like the
// "unbounded" string. NaN and every other magnitude above MaxNumericValue
// are rejected.
func legacyNumberLength(v float64) (layout.Length, error) {
	switch {
	case math.IsNaN(v):
		return layout.Length{}, errors.New("length is not a number")
	case v >= math.MaxFloat64:
		return layout.PxUnbounded, nil
	case math.Abs(v) > MaxNumericValue:
		return layout.Length{}, fmt.Errorf("length %v exceeds %g", v, MaxNumericValue)
	}
	return layout.Px(v), nil
}

// checkFloat rejects NaN, infinities, and magnitudes above MaxNumericValue.
func checkFloat(path string, v float64) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("%s: value %v is not finite", path, v)
	}
	if math.Abs(v) > MaxNumericValue {
		return fmt.Errorf("%s: value %v exceeds %g", path, v, MaxNumericValue)
	}
	return nil
}

// checkRect validates the four Rect fields. Unlike checkFloat it accepts
// exactly math.MaxFloat64 (layout.Unbounded): the layout pass produces it
// for boxes measured under unbounded constraints and the previous encoder
// wrote it verbatim, so ToJSON must not fail on a tree Layout produced and
// FromJSON must read such files back.
func checkRect(path string, r *layout.Rect) error {
	for _, f := range []struct {
		name string
		v    float64
	}{{"x", r.X}, {"y", r.Y}, {"width", r.Width}, {"height", r.Height}} {
		if f.v == math.MaxFloat64 {
			continue
		}
		if err := checkFloat(path+"."+f.name, f.v); err != nil {
			return err
		}
	}
	return nil
}

// =============================================================================
// Numeric bounds
// =============================================================================

// Enum keywords come from the String/Parse pairs in the layout package
// (enums.go); see encodeEnum/decodeEnum. Bounds for the two TextStyle enums
// that are serialized numerically:
const (
	maxFontWeight     = 1000 // CSS Fonts Level 4 font-weight range is [1, 1000]
	maxTextDecoration = int(layout.TextDecorationUnderline | layout.TextDecorationOverline | layout.TextDecorationLineThrough)
)
