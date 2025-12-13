package layout

// WPT Test Schema v1.0.0
// Universal test format for non-browser implementations

import (
	"encoding/json"
	"fmt"
	"os"
)

// WPTTest represents a Web Platform Test in universal JSON format
type WPTTest struct {
	Version     string         `json:"version"`
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Source      WPTSource      `json:"source"`
	Generated   WPTGenerated   `json:"generated"`
	Spec        WPTSpec        `json:"spec"`
	Categories  []string       `json:"categories"`
	Tags        []string       `json:"tags"`
	Properties  []string       `json:"properties"`
	Layout      WPTLayoutTree  `json:"layout"`
	Constraints WPTConstraints `json:"constraints"`
	Results     WPTResults     `json:"results"`
	Notes       []string       `json:"notes,omitempty"`
	KnownIssues []string       `json:"knownIssues,omitempty"`
}

// WPTSource tracks the original test file
type WPTSource struct {
	URL    string  `json:"url"`
	File   string  `json:"file"`
	Commit *string `json:"commit,omitempty"`
}

// WPTGenerated tracks when and how the test was generated
type WPTGenerated struct {
	Timestamp string `json:"timestamp"`
	Tool      string `json:"tool"`
}

// WPTSpec links to the specification
type WPTSpec struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	URL     string `json:"url"`
}

// WPTLayoutTree represents the declarative layout structure
type WPTLayoutTree struct {
	Type      string          `json:"type"` // "container", "block", "text"
	ID        string          `json:"id,omitempty"`
	Style     WPTStyle        `json:"style"`
	Text      string          `json:"text,omitempty"`
	TextStyle *WPTTextStyle   `json:"textStyle,omitempty"`
	Children  []WPTLayoutTree `json:"children,omitempty"`
}

// WPTStyle represents CSS properties relevant to layout
type WPTStyle struct {
	// Display & Positioning
	Display  string `json:"display,omitempty"`
	Position string `json:"position,omitempty"`

	// Flexbox
	FlexDirection  string      `json:"flexDirection,omitempty"`
	FlexWrap       string      `json:"flexWrap,omitempty"`
	JustifyContent string      `json:"justifyContent,omitempty"`
	AlignItems     string      `json:"alignItems,omitempty"`
	AlignContent   string      `json:"alignContent,omitempty"`
	AlignSelf      string      `json:"alignSelf,omitempty"`
	FlexGrow       *float64    `json:"flexGrow,omitempty"`
	FlexShrink     *float64    `json:"flexShrink,omitempty"`
	FlexBasis      interface{} `json:"flexBasis,omitempty"` // number or "auto"

	// Grid
	GridTemplateColumns string   `json:"gridTemplateColumns,omitempty"`
	GridTemplateRows    string   `json:"gridTemplateRows,omitempty"`
	GridGap             *float64 `json:"gridGap,omitempty"`

	// Box Model
	Width     *float64 `json:"width,omitempty"`
	Height    *float64 `json:"height,omitempty"`
	MinWidth  *float64 `json:"minWidth,omitempty"`
	MinHeight *float64 `json:"minHeight,omitempty"`
	MaxWidth  *float64 `json:"maxWidth,omitempty"`
	MaxHeight *float64 `json:"maxHeight,omitempty"`

	// Spacing
	Margin  *WPTSpacing `json:"margin,omitempty"`
	Padding *WPTSpacing `json:"padding,omitempty"`
	Border  *WPTSpacing `json:"border,omitempty"`
}

// WPTSpacing represents margin/padding/border values
type WPTSpacing struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

// WPTTextStyle represents text properties
type WPTTextStyle struct {
	FontSize   *float64 `json:"fontSize,omitempty"`
	FontFamily string   `json:"fontFamily,omitempty"`
	WhiteSpace string   `json:"whiteSpace,omitempty"`
	TextAlign  string   `json:"textAlign,omitempty"`
}

// WPTConstraints represents layout constraints
type WPTConstraints struct {
	Type   string  `json:"type"` // "loose", "tight", "bounded"
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// WPTResults contains expected results from multiple browsers
type WPTResults map[string]WPTBrowserResult

// WPTBrowserResult represents results from a specific browser
type WPTBrowserResult struct {
	Browser   WPTBrowser         `json:"browser"`
	Rendered  WPTRendered        `json:"rendered"`
	Elements  []WPTElementResult `json:"elements"`
	Tolerance *WPTTolerance      `json:"tolerance,omitempty"`
}

// WPTBrowser identifies the browser
type WPTBrowser struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Engine  string `json:"engine"`
}

// WPTRendered tracks rendering metadata
type WPTRendered struct {
	Timestamp string      `json:"timestamp"`
	Viewport  WPTViewport `json:"viewport"`
}

// WPTViewport represents the viewport size
type WPTViewport struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// WPTElementResult represents expected values for an element
type WPTElementResult struct {
	ID         string                 `json:"id,omitempty"`
	Path       string                 `json:"path"`
	Expected   map[string]interface{} `json:"expected"`
	Assertions []CELAssertion         `json:"assertions,omitempty"`
}

// WPTTolerance defines acceptable differences
type WPTTolerance struct {
	Position float64 `json:"position"` // Default: 1.0px
	Size     float64 `json:"size"`     // Default: 1.0px
	Numeric  float64 `json:"numeric"`  // Default: 0.01
}

// LoadWPTTest loads a WPT test from a JSON file
func LoadWPTTest(filename string) (*WPTTest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read test file: %w", err)
	}

	var test WPTTest
	if err := json.Unmarshal(data, &test); err != nil {
		return nil, fmt.Errorf("failed to parse test JSON: %w", err)
	}

	// Validate version
	if test.Version != "1.0.0" {
		return nil, fmt.Errorf("unsupported schema version: %s (expected 1.0.0)", test.Version)
	}

	return &test, nil
}

// BuildLayout constructs a layout Node from the WPT layout tree
func (test *WPTTest) BuildLayout() (*Node, error) {
	return buildNodeFromWPT(&test.Layout)
}

func buildNodeFromWPT(tree *WPTLayoutTree) (*Node, error) {
	node := &Node{
		Style: Style{},
	}

	// Map display types
	switch tree.Style.Display {
	case "flex":
		node.Style.Display = DisplayFlex
	case "grid":
		node.Style.Display = DisplayGrid
	case "block", "inline-block":
		node.Style.Display = DisplayBlock
	default:
		node.Style.Display = DisplayBlock
	}

	// Flexbox properties
	if tree.Style.FlexDirection != "" {
		switch tree.Style.FlexDirection {
		case "row":
			node.Style.FlexDirection = FlexDirectionRow
		case "column":
			node.Style.FlexDirection = FlexDirectionColumn
		case "row-reverse":
			node.Style.FlexDirection = FlexDirectionRowReverse
		case "column-reverse":
			node.Style.FlexDirection = FlexDirectionColumnReverse
		}
	}

	if tree.Style.JustifyContent != "" {
		switch tree.Style.JustifyContent {
		case "flex-start":
			node.Style.JustifyContent = JustifyContentFlexStart
		case "flex-end":
			node.Style.JustifyContent = JustifyContentFlexEnd
		case "center":
			node.Style.JustifyContent = JustifyContentCenter
		case "space-between":
			node.Style.JustifyContent = JustifyContentSpaceBetween
		case "space-around":
			node.Style.JustifyContent = JustifyContentSpaceAround
		case "space-evenly":
			node.Style.JustifyContent = JustifyContentSpaceEvenly
		}
	}

	if tree.Style.AlignItems != "" {
		switch tree.Style.AlignItems {
		case "flex-start":
			node.Style.AlignItems = AlignItemsFlexStart
		case "flex-end":
			node.Style.AlignItems = AlignItemsFlexEnd
		case "center":
			node.Style.AlignItems = AlignItemsCenter
		case "baseline":
			node.Style.AlignItems = AlignItemsBaseline
		case "stretch":
			node.Style.AlignItems = AlignItemsStretch
		}
	}

	if tree.Style.AlignContent != "" {
		switch tree.Style.AlignContent {
		case "flex-start":
			node.Style.AlignContent = AlignContentFlexStart
		case "flex-end":
			node.Style.AlignContent = AlignContentFlexEnd
		case "center":
			node.Style.AlignContent = AlignContentCenter
		case "space-between":
			node.Style.AlignContent = AlignContentSpaceBetween
		case "space-around":
			node.Style.AlignContent = AlignContentSpaceAround
		case "stretch":
			node.Style.AlignContent = AlignContentStretch
		}
	}

	// Box Model
	if tree.Style.Width != nil {
		node.Style.Width = *tree.Style.Width
	}
	if tree.Style.Height != nil {
		node.Style.Height = *tree.Style.Height
	}

	// Spacing
	if tree.Style.Padding != nil {
		node.Style.Padding = Spacing{
			Top:    tree.Style.Padding.Top,
			Right:  tree.Style.Padding.Right,
			Bottom: tree.Style.Padding.Bottom,
			Left:   tree.Style.Padding.Left,
		}
	}

	if tree.Style.Margin != nil {
		node.Style.Margin = Spacing{
			Top:    tree.Style.Margin.Top,
			Right:  tree.Style.Margin.Right,
			Bottom: tree.Style.Margin.Bottom,
			Left:   tree.Style.Margin.Left,
		}
	}

	// Text
	if tree.Text != "" {
		node.Text = tree.Text
	}

	// Children
	for i := range tree.Children {
		child, err := buildNodeFromWPT(&tree.Children[i])
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	return node, nil
}

// GetConstraints returns the layout constraints
func (test *WPTTest) GetConstraints() Constraints {
	switch test.Constraints.Type {
	case "tight":
		return Tight(test.Constraints.Width, test.Constraints.Height)
	case "bounded":
		// Bounded: min=0, max=specified (same as loose for now)
		return Loose(test.Constraints.Width, test.Constraints.Height)
	default:
		return Loose(test.Constraints.Width, test.Constraints.Height)
	}
}

// GetTolerance returns tolerance values for a browser, or defaults
func (result *WPTBrowserResult) GetTolerance() WPTTolerance {
	if result.Tolerance != nil {
		return *result.Tolerance
	}
	return WPTTolerance{
		Position: 1.0,
		Size:     1.0,
		Numeric:  0.01,
	}
}
