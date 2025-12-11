package serialize

import (
	"encoding/json"

	"github.com/SCKelemen/layout"
)

// NodeJSON represents a serializable version of layout.Node
type NodeJSON struct {
	Style    StyleJSON   `json:"style"`
	Children []*NodeJSON `json:"children,omitempty"`
	Rect     RectJSON    `json:"rect,omitempty"`
}

// StyleJSON represents a serializable version of layout.Style
type StyleJSON struct {
	Display        string  `json:"display,omitempty"`
	FlexDirection  string  `json:"flexDirection,omitempty"`
	FlexWrap       string  `json:"flexWrap,omitempty"`
	JustifyContent string  `json:"justifyContent,omitempty"`
	AlignItems     string  `json:"alignItems,omitempty"`
	AlignContent   string  `json:"alignContent,omitempty"`
	JustifyItems   string  `json:"justifyItems,omitempty"`
	FlexGrow       float64 `json:"flexGrow,omitempty"`
	FlexShrink     float64 `json:"flexShrink,omitempty"`
	FlexBasis      float64 `json:"flexBasis,omitempty"`

	// Grid
	GridTemplateRows    []TrackJSON `json:"gridTemplateRows,omitempty"`
	GridTemplateColumns []TrackJSON `json:"gridTemplateColumns,omitempty"`
	GridAutoRows        TrackJSON   `json:"gridAutoRows,omitempty"`
	GridAutoColumns     TrackJSON   `json:"gridAutoColumns,omitempty"`
	GridGap             float64     `json:"gridGap,omitempty"`
	GridRowGap          float64     `json:"gridRowGap,omitempty"`
	GridColumnGap       float64     `json:"gridColumnGap,omitempty"`
	GridRowStart        int         `json:"gridRowStart,omitempty"`
	GridRowEnd          int         `json:"gridRowEnd,omitempty"`
	GridColumnStart     int         `json:"gridColumnStart,omitempty"`
	GridColumnEnd       int         `json:"gridColumnEnd,omitempty"`

	// Sizing
	Width       float64 `json:"width,omitempty"`
	Height      float64 `json:"height,omitempty"`
	MinWidth    float64 `json:"minWidth,omitempty"`
	MinHeight   float64 `json:"minHeight,omitempty"`
	MaxWidth    float64 `json:"maxWidth,omitempty"`
	MaxHeight   float64 `json:"maxHeight,omitempty"`
	AspectRatio float64 `json:"aspectRatio,omitempty"`

	// Spacing
	Padding SpacingJSON `json:"padding,omitempty"`
	Margin  SpacingJSON `json:"margin,omitempty"`
	Border  SpacingJSON `json:"border,omitempty"`

	// Box model
	BoxSizing string `json:"boxSizing,omitempty"`

	// Positioning
	Position string  `json:"position,omitempty"`
	Top      float64 `json:"top,omitempty"`
	Right    float64 `json:"right,omitempty"`
	Bottom   float64 `json:"bottom,omitempty"`
	Left     float64 `json:"left,omitempty"`
	ZIndex   int     `json:"zIndex,omitempty"`

	// Transform
	Transform TransformJSON `json:"transform,omitempty"`
}

// TrackJSON represents a serializable version of layout.GridTrack
type TrackJSON struct {
	MinSize  float64 `json:"minSize,omitempty"`
	MaxSize  float64 `json:"maxSize,omitempty"`
	Fraction float64 `json:"fraction,omitempty"`
}

// SpacingJSON represents a serializable version of layout.Spacing
type SpacingJSON struct {
	Top    float64 `json:"top,omitempty"`
	Right  float64 `json:"right,omitempty"`
	Bottom float64 `json:"bottom,omitempty"`
	Left   float64 `json:"left,omitempty"`
}

// RectJSON represents a serializable version of layout.Rect
type RectJSON struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// TransformJSON represents a serializable version of layout.Transform
type TransformJSON struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
	C float64 `json:"c"`
	D float64 `json:"d"`
	E float64 `json:"e"`
	F float64 `json:"f"`
}

// ToJSON converts a layout.Node to JSON bytes
func ToJSON(node *layout.Node) ([]byte, error) {
	nodeJSON := nodeToJSON(node)
	return json.MarshalIndent(nodeJSON, "", "  ")
}

// FromJSON converts JSON bytes to a layout.Node
func FromJSON(data []byte) (*layout.Node, error) {
	var nodeJSON NodeJSON
	if err := json.Unmarshal(data, &nodeJSON); err != nil {
		return nil, err
	}
	return jsonToNode(&nodeJSON), nil
}

// nodeToJSON converts a layout.Node to NodeJSON
func nodeToJSON(node *layout.Node) *NodeJSON {
	if node == nil {
		return nil
	}

	nj := &NodeJSON{
		Style: styleToJSON(&node.Style),
		Rect:  rectToJSON(&node.Rect),
	}

	if len(node.Children) > 0 {
		nj.Children = make([]*NodeJSON, len(node.Children))
		for i, child := range node.Children {
			nj.Children[i] = nodeToJSON(child)
		}
	}

	return nj
}

// jsonToNode converts a NodeJSON to layout.Node
func jsonToNode(nj *NodeJSON) *layout.Node {
	if nj == nil {
		return nil
	}

	node := &layout.Node{
		Style: jsonToStyle(&nj.Style),
		Rect:  jsonToRect(&nj.Rect),
	}

	if len(nj.Children) > 0 {
		node.Children = make([]*layout.Node, len(nj.Children))
		for i, child := range nj.Children {
			node.Children[i] = jsonToNode(child)
		}
	}

	return node
}

// styleToJSON converts layout.Style to StyleJSON
func styleToJSON(s *layout.Style) StyleJSON {
	sj := StyleJSON{
		Width:           s.Width,
		Height:          s.Height,
		MinWidth:        s.MinWidth,
		MinHeight:       s.MinHeight,
		MaxWidth:        s.MaxWidth,
		MaxHeight:       s.MaxHeight,
		AspectRatio:     s.AspectRatio,
		FlexGrow:        s.FlexGrow,
		FlexShrink:      s.FlexShrink,
		FlexBasis:       s.FlexBasis,
		GridGap:         s.GridGap,
		GridRowGap:      s.GridRowGap,
		GridColumnGap:   s.GridColumnGap,
		GridRowStart:    s.GridRowStart,
		GridRowEnd:      s.GridRowEnd,
		GridColumnStart: s.GridColumnStart,
		GridColumnEnd:   s.GridColumnEnd,
		Top:             s.Top,
		Right:           s.Right,
		Bottom:          s.Bottom,
		Left:            s.Left,
		ZIndex:          s.ZIndex,
		Padding:         spacingToJSON(&s.Padding),
		Margin:          spacingToJSON(&s.Margin),
		Border:          spacingToJSON(&s.Border),
		Transform:       transformToJSON(&s.Transform),
	}

	// Convert enums to strings
	if s.Display != 0 {
		sj.Display = displayToString(s.Display)
	}
	if s.FlexDirection != 0 {
		sj.FlexDirection = flexDirectionToString(s.FlexDirection)
	}
	if s.FlexWrap != 0 {
		sj.FlexWrap = flexWrapToString(s.FlexWrap)
	}
	if s.JustifyContent != 0 {
		sj.JustifyContent = justifyContentToString(s.JustifyContent)
	}
	if s.AlignItems != 0 {
		sj.AlignItems = alignItemsToString(s.AlignItems)
	}
	if s.AlignContent != 0 {
		sj.AlignContent = alignContentToString(s.AlignContent)
	}
	if s.JustifyItems != 0 {
		sj.JustifyItems = justifyItemsToString(s.JustifyItems)
	}
	if s.BoxSizing != 0 {
		sj.BoxSizing = boxSizingToString(s.BoxSizing)
	}
	if s.Position != 0 {
		sj.Position = positionToString(s.Position)
	}

	// Convert grid tracks
	if len(s.GridTemplateRows) > 0 {
		sj.GridTemplateRows = make([]TrackJSON, len(s.GridTemplateRows))
		for i := range s.GridTemplateRows {
			sj.GridTemplateRows[i] = trackToJSON(&s.GridTemplateRows[i])
		}
	}
	if len(s.GridTemplateColumns) > 0 {
		sj.GridTemplateColumns = make([]TrackJSON, len(s.GridTemplateColumns))
		for i := range s.GridTemplateColumns {
			sj.GridTemplateColumns[i] = trackToJSON(&s.GridTemplateColumns[i])
		}
	}
	if s.GridAutoRows.MinSize != 0 || s.GridAutoRows.MaxSize != layout.Unbounded || s.GridAutoRows.Fraction != 0 {
		sj.GridAutoRows = trackToJSON(&s.GridAutoRows)
	}
	if s.GridAutoColumns.MinSize != 0 || s.GridAutoColumns.MaxSize != layout.Unbounded || s.GridAutoColumns.Fraction != 0 {
		sj.GridAutoColumns = trackToJSON(&s.GridAutoColumns)
	}

	return sj
}

// jsonToStyle converts StyleJSON to layout.Style
func jsonToStyle(sj *StyleJSON) layout.Style {
	s := layout.Style{
		Width:           sj.Width,
		Height:          sj.Height,
		MinWidth:        sj.MinWidth,
		MinHeight:       sj.MinHeight,
		MaxWidth:        sj.MaxWidth,
		MaxHeight:       sj.MaxHeight,
		AspectRatio:     sj.AspectRatio,
		FlexGrow:        sj.FlexGrow,
		FlexShrink:      sj.FlexShrink,
		FlexBasis:       sj.FlexBasis,
		GridGap:         sj.GridGap,
		GridRowGap:      sj.GridRowGap,
		GridColumnGap:   sj.GridColumnGap,
		GridRowStart:    sj.GridRowStart,
		GridRowEnd:      sj.GridRowEnd,
		GridColumnStart: sj.GridColumnStart,
		GridColumnEnd:   sj.GridColumnEnd,
		Top:             sj.Top,
		Right:           sj.Right,
		Bottom:          sj.Bottom,
		Left:            sj.Left,
		ZIndex:          sj.ZIndex,
		Padding:         jsonToSpacing(&sj.Padding),
		Margin:          jsonToSpacing(&sj.Margin),
		Border:          jsonToSpacing(&sj.Border),
		Transform:       jsonToTransform(&sj.Transform),
	}

	// Convert strings to enums
	if sj.Display != "" {
		s.Display = stringToDisplay(sj.Display)
	}
	if sj.FlexDirection != "" {
		s.FlexDirection = stringToFlexDirection(sj.FlexDirection)
	}
	if sj.FlexWrap != "" {
		s.FlexWrap = stringToFlexWrap(sj.FlexWrap)
	}
	if sj.JustifyContent != "" {
		s.JustifyContent = stringToJustifyContent(sj.JustifyContent)
	}
	if sj.AlignItems != "" {
		s.AlignItems = stringToAlignItems(sj.AlignItems)
	}
	if sj.JustifyItems != "" {
		s.JustifyItems = stringToJustifyItems(sj.JustifyItems)
	}
	if sj.AlignContent != "" {
		s.AlignContent = stringToAlignContent(sj.AlignContent)
	}
	if sj.BoxSizing != "" {
		s.BoxSizing = stringToBoxSizing(sj.BoxSizing)
	}
	if sj.Position != "" {
		s.Position = stringToPosition(sj.Position)
	}

	// Convert grid tracks
	if len(sj.GridTemplateRows) > 0 {
		s.GridTemplateRows = make([]layout.GridTrack, len(sj.GridTemplateRows))
		for i := range sj.GridTemplateRows {
			s.GridTemplateRows[i] = jsonToTrack(&sj.GridTemplateRows[i])
		}
	}
	if len(sj.GridTemplateColumns) > 0 {
		s.GridTemplateColumns = make([]layout.GridTrack, len(sj.GridTemplateColumns))
		for i := range sj.GridTemplateColumns {
			s.GridTemplateColumns[i] = jsonToTrack(&sj.GridTemplateColumns[i])
		}
	}
	if sj.GridAutoRows.MinSize != 0 || sj.GridAutoRows.MaxSize != layout.Unbounded || sj.GridAutoRows.Fraction != 0 {
		s.GridAutoRows = jsonToTrack(&sj.GridAutoRows)
	}
	if sj.GridAutoColumns.MinSize != 0 || sj.GridAutoColumns.MaxSize != layout.Unbounded || sj.GridAutoColumns.Fraction != 0 {
		s.GridAutoColumns = jsonToTrack(&sj.GridAutoColumns)
	}

	return s
}

// Helper functions for enum conversions
func displayToString(d layout.Display) string {
	switch d {
	case layout.DisplayBlock:
		return "block"
	case layout.DisplayFlex:
		return "flex"
	case layout.DisplayGrid:
		return "grid"
	default:
		return ""
	}
}

func stringToDisplay(s string) layout.Display {
	switch s {
	case "block":
		return layout.DisplayBlock
	case "flex":
		return layout.DisplayFlex
	case "grid":
		return layout.DisplayGrid
	default:
		return 0
	}
}

func flexDirectionToString(fd layout.FlexDirection) string {
	switch fd {
	case layout.FlexDirectionRow:
		return "row"
	case layout.FlexDirectionRowReverse:
		return "row-reverse"
	case layout.FlexDirectionColumn:
		return "column"
	case layout.FlexDirectionColumnReverse:
		return "column-reverse"
	default:
		return ""
	}
}

func stringToFlexDirection(s string) layout.FlexDirection {
	switch s {
	case "row":
		return layout.FlexDirectionRow
	case "row-reverse":
		return layout.FlexDirectionRowReverse
	case "column":
		return layout.FlexDirectionColumn
	case "column-reverse":
		return layout.FlexDirectionColumnReverse
	default:
		return 0
	}
}

func flexWrapToString(fw layout.FlexWrap) string {
	switch fw {
	case layout.FlexWrapNoWrap:
		return "nowrap"
	case layout.FlexWrapWrap:
		return "wrap"
	case layout.FlexWrapWrapReverse:
		return "wrap-reverse"
	default:
		return ""
	}
}

func stringToFlexWrap(s string) layout.FlexWrap {
	switch s {
	case "nowrap":
		return layout.FlexWrapNoWrap
	case "wrap":
		return layout.FlexWrapWrap
	case "wrap-reverse":
		return layout.FlexWrapWrapReverse
	default:
		return 0
	}
}

func justifyContentToString(jc layout.JustifyContent) string {
	switch jc {
	case layout.JustifyContentFlexStart:
		return "flex-start"
	case layout.JustifyContentFlexEnd:
		return "flex-end"
	case layout.JustifyContentCenter:
		return "center"
	case layout.JustifyContentSpaceBetween:
		return "space-between"
	case layout.JustifyContentSpaceAround:
		return "space-around"
	case layout.JustifyContentSpaceEvenly:
		return "space-evenly"
	default:
		return ""
	}
}

func stringToJustifyContent(s string) layout.JustifyContent {
	switch s {
	case "flex-start":
		return layout.JustifyContentFlexStart
	case "flex-end":
		return layout.JustifyContentFlexEnd
	case "center":
		return layout.JustifyContentCenter
	case "space-between":
		return layout.JustifyContentSpaceBetween
	case "space-around":
		return layout.JustifyContentSpaceAround
	case "space-evenly":
		return layout.JustifyContentSpaceEvenly
	default:
		return 0
	}
}

func alignItemsToString(ai layout.AlignItems) string {
	switch ai {
	case layout.AlignItemsFlexStart:
		return "flex-start"
	case layout.AlignItemsFlexEnd:
		return "flex-end"
	case layout.AlignItemsCenter:
		return "center"
	case layout.AlignItemsStretch:
		return "stretch"
	case layout.AlignItemsBaseline:
		return "baseline"
	default:
		return ""
	}
}

func stringToAlignItems(s string) layout.AlignItems {
	switch s {
	case "flex-start":
		return layout.AlignItemsFlexStart
	case "flex-end":
		return layout.AlignItemsFlexEnd
	case "center":
		return layout.AlignItemsCenter
	case "stretch":
		return layout.AlignItemsStretch
	case "baseline":
		return layout.AlignItemsBaseline
	default:
		return 0
	}
}

func justifyItemsToString(ji layout.JustifyItems) string {
	switch ji {
	case layout.JustifyItemsStart:
		return "start"
	case layout.JustifyItemsEnd:
		return "end"
	case layout.JustifyItemsCenter:
		return "center"
	case layout.JustifyItemsStretch:
		return "stretch"
	default:
		return ""
	}
}

func stringToJustifyItems(s string) layout.JustifyItems {
	switch s {
	case "start":
		return layout.JustifyItemsStart
	case "end":
		return layout.JustifyItemsEnd
	case "center":
		return layout.JustifyItemsCenter
	case "stretch":
		return layout.JustifyItemsStretch
	default:
		return 0
	}
}

func alignContentToString(ac layout.AlignContent) string {
	switch ac {
	case layout.AlignContentFlexStart:
		return "flex-start"
	case layout.AlignContentFlexEnd:
		return "flex-end"
	case layout.AlignContentCenter:
		return "center"
	case layout.AlignContentStretch:
		return "stretch"
	case layout.AlignContentSpaceBetween:
		return "space-between"
	case layout.AlignContentSpaceAround:
		return "space-around"
	default:
		return ""
	}
}

func stringToAlignContent(s string) layout.AlignContent {
	switch s {
	case "flex-start":
		return layout.AlignContentFlexStart
	case "flex-end":
		return layout.AlignContentFlexEnd
	case "center":
		return layout.AlignContentCenter
	case "stretch":
		return layout.AlignContentStretch
	case "space-between":
		return layout.AlignContentSpaceBetween
	case "space-around":
		return layout.AlignContentSpaceAround
	default:
		return 0
	}
}

func boxSizingToString(bs layout.BoxSizing) string {
	switch bs {
	case layout.BoxSizingContentBox:
		return "content-box"
	case layout.BoxSizingBorderBox:
		return "border-box"
	default:
		return ""
	}
}

func stringToBoxSizing(s string) layout.BoxSizing {
	switch s {
	case "content-box":
		return layout.BoxSizingContentBox
	case "border-box":
		return layout.BoxSizingBorderBox
	default:
		return 0
	}
}

func positionToString(p layout.Position) string {
	switch p {
	case layout.PositionStatic:
		return "static"
	case layout.PositionRelative:
		return "relative"
	case layout.PositionAbsolute:
		return "absolute"
	case layout.PositionFixed:
		return "fixed"
	case layout.PositionSticky:
		return "sticky"
	default:
		return ""
	}
}

func stringToPosition(s string) layout.Position {
	switch s {
	case "static":
		return layout.PositionStatic
	case "relative":
		return layout.PositionRelative
	case "absolute":
		return layout.PositionAbsolute
	case "fixed":
		return layout.PositionFixed
	case "sticky":
		return layout.PositionSticky
	default:
		return 0
	}
}

func trackToJSON(t *layout.GridTrack) TrackJSON {
	return TrackJSON{
		MinSize:  t.MinSize,
		MaxSize:  t.MaxSize,
		Fraction: t.Fraction,
	}
}

func jsonToTrack(tj *TrackJSON) layout.GridTrack {
	return layout.GridTrack{
		MinSize:  tj.MinSize,
		MaxSize:  tj.MaxSize,
		Fraction: tj.Fraction,
	}
}

func spacingToJSON(s *layout.Spacing) SpacingJSON {
	return SpacingJSON{
		Top:    s.Top,
		Right:  s.Right,
		Bottom: s.Bottom,
		Left:   s.Left,
	}
}

func jsonToSpacing(sj *SpacingJSON) layout.Spacing {
	return layout.Spacing{
		Top:    sj.Top,
		Right:  sj.Right,
		Bottom: sj.Bottom,
		Left:   sj.Left,
	}
}

func rectToJSON(r *layout.Rect) RectJSON {
	return RectJSON{
		X:      r.X,
		Y:      r.Y,
		Width:  r.Width,
		Height: r.Height,
	}
}

func jsonToRect(rj *RectJSON) layout.Rect {
	return layout.Rect{
		X:      rj.X,
		Y:      rj.Y,
		Width:  rj.Width,
		Height: rj.Height,
	}
}

func transformToJSON(t *layout.Transform) TransformJSON {
	return TransformJSON{
		A: t.A,
		B: t.B,
		C: t.C,
		D: t.D,
		E: t.E,
		F: t.F,
	}
}

func jsonToTransform(tj *TransformJSON) layout.Transform {
	return layout.Transform{
		A: tj.A,
		B: tj.B,
		C: tj.C,
		D: tj.D,
		E: tj.E,
		F: tj.F,
	}
}
