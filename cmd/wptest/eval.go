package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/layout"
)

type EvalOptions struct {
	verbose bool
}

// EvalRequest is the JSON input schema for wptest eval
type EvalRequest struct {
	Layout      layout.WPTLayoutTree  `json:"layout"`
	Constraints *ConstraintsSpec      `json:"constraints,omitempty"`
	Assertions  []layout.CELAssertion `json:"assertions"`
	Binding     string                `json:"binding,omitempty"` // "old" or "context", defaults to "old"
}

// ConstraintsSpec defines layout constraints
type ConstraintsSpec struct {
	MinWidth  *float64 `json:"minWidth,omitempty"`
	MaxWidth  *float64 `json:"maxWidth,omitempty"`
	MinHeight *float64 `json:"minHeight,omitempty"`
	MaxHeight *float64 `json:"maxHeight,omitempty"`
}

// EvalResponse is the JSON output schema for wptest eval
type EvalResponse struct {
	Passed  int                      `json:"passed"`
	Failed  int                      `json:"failed"`
	Skipped int                      `json:"skipped"`
	Results []layout.AssertionResult `json:"results"`
	Layout  *LayoutOutput            `json:"layout,omitempty"` // Optional: computed layout
}

// LayoutOutput describes computed layout (optional verbose output)
type LayoutOutput struct {
	X        float64        `json:"x"`
	Y        float64        `json:"y"`
	Width    float64        `json:"width"`
	Height   float64        `json:"height"`
	Children []LayoutOutput `json:"children,omitempty"`
}

func newEvalCommand() *clix.Command {
	opts := &EvalOptions{}

	cmd := clix.NewCommand("eval")
	cmd.Short = "Evaluate layout and assertions from JSON stdin"
	cmd.Long = `Evaluate layout and CEL assertions from JSON input on stdin.

This command provides a language-agnostic interface for testing layouts.
Any language that can spawn a process and write/read JSON can use this.

Input JSON schema (stdin):
  {
    "layout": {
      "type": "container",
      "style": {
        "display": "flex",
        "width": 600,
        ...
      },
      "children": [...]
    },
    "constraints": {
      "maxWidth": 800,
      "maxHeight": 600
    },
    "assertions": [
      {"expression": "getX(root()) == 0", "message": "positioned"}
    ],
    "binding": "old"
  }

Output JSON schema (stdout):
  {
    "passed": 5,
    "failed": 0,
    "skipped": 0,
    "results": [...]
  }

Examples:
  echo '{"layout": {...}, "assertions": [...]}' | wptest eval
  wptest eval < test.json
  cat test.json | wptest eval --verbose`

	cmd.Run = func(ctx *clix.Context) error {
		return evalFromStdin(ctx.Context, opts)
	}

	cmd.Flags.BoolVar(clix.BoolVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "verbose",
			Short: "v",
			Usage: "Include computed layout in output",
		},
		Value: &opts.verbose,
	})

	return cmd
}

func evalFromStdin(ctx context.Context, opts *EvalOptions) error {
	// Read JSON from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	// Parse request
	var req EvalRequest
	if err := json.Unmarshal(input, &req); err != nil {
		return fmt.Errorf("failed to parse JSON input: %w", err)
	}

	// Default binding to "old"
	if req.Binding == "" {
		req.Binding = "old"
	}

	// Validate binding
	if req.Binding != "old" && req.Binding != "context" {
		return fmt.Errorf("invalid binding: %s (must be 'old' or 'context')", req.Binding)
	}

	// Build layout tree from WPT schema
	root := buildLayoutFromWPTTree(&req.Layout)

	// Get constraints
	constraints := getConstraints(req.Constraints)

	// Run layout algorithm
	layout.Layout(root, constraints)

	// Evaluate assertions
	var results []layout.AssertionResult

	switch req.Binding {
	case "old":
		env, err := layout.NewLayoutCELEnv(root)
		if err != nil {
			return fmt.Errorf("failed to create CEL environment: %w", err)
		}
		results = env.EvaluateAll(req.Assertions)

	case "context":
		env, err := layout.NewLayoutCELEnvWithContext(root)
		if err != nil {
			return fmt.Errorf("failed to create CEL environment: %w", err)
		}
		for _, assertion := range req.Assertions {
			results = append(results, env.Evaluate(assertion))
		}

	default:
		return fmt.Errorf("unsupported binding: %s", req.Binding)
	}

	// Count results
	passed := 0
	failed := 0
	skipped := 0

	for _, result := range results {
		if result.Passed {
			passed++
		} else {
			// Check if skipped (unsupported features)
			if strings.Contains(result.Error, "undeclared reference to 'this'") ||
				strings.Contains(result.Error, "undeclared reference to 'parent'") {
				skipped++
			} else {
				failed++
			}
		}
	}

	// Build response
	response := EvalResponse{
		Passed:  passed,
		Failed:  failed,
		Skipped: skipped,
		Results: results,
	}

	// Add layout output if verbose
	if opts.verbose {
		response.Layout = buildLayoutOutput(root)
	}

	// Write JSON to stdout
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}

	// Exit with error if any assertions failed
	if failed > 0 {
		os.Exit(1)
	}

	return nil
}

func buildLayoutFromWPTTree(tree *layout.WPTLayoutTree) *layout.Node {
	if tree == nil {
		return nil
	}

	node := &layout.Node{
		Style: layout.Style{},
	}

	// Parse style
	style := &tree.Style

	// Display
	if style.Display != "" {
		switch style.Display {
		case "flex":
			node.Style.Display = layout.DisplayFlex
		case "block":
			node.Style.Display = layout.DisplayBlock
		case "grid":
			node.Style.Display = layout.DisplayGrid
		case "none":
			node.Style.Display = layout.DisplayNone
		default:
			// Default to block
			node.Style.Display = layout.DisplayBlock
		}
	}

	// Box model
	if style.Width != nil {
		node.Style.Width = *style.Width
	}
	if style.Height != nil {
		node.Style.Height = *style.Height
	}
	if style.MinWidth != nil {
		node.Style.MinWidth = *style.MinWidth
	}
	if style.MinHeight != nil {
		node.Style.MinHeight = *style.MinHeight
	}
	if style.MaxWidth != nil {
		node.Style.MaxWidth = *style.MaxWidth
	}
	if style.MaxHeight != nil {
		node.Style.MaxHeight = *style.MaxHeight
	}

	// Margin
	if style.Margin != nil {
		node.Style.Margin = layout.Spacing{
			Top:    style.Margin.Top,
			Right:  style.Margin.Right,
			Bottom: style.Margin.Bottom,
			Left:   style.Margin.Left,
		}
	}

	// Padding
	if style.Padding != nil {
		node.Style.Padding = layout.Spacing{
			Top:    style.Padding.Top,
			Right:  style.Padding.Right,
			Bottom: style.Padding.Bottom,
			Left:   style.Padding.Left,
		}
	}

	// Border
	if style.Border != nil {
		node.Style.Border = layout.Spacing{
			Top:    style.Border.Top,
			Right:  style.Border.Right,
			Bottom: style.Border.Bottom,
			Left:   style.Border.Left,
		}
	}

	// Flexbox
	if style.FlexDirection != "" {
		switch style.FlexDirection {
		case "row":
			node.Style.FlexDirection = layout.FlexDirectionRow
		case "row-reverse":
			node.Style.FlexDirection = layout.FlexDirectionRowReverse
		case "column":
			node.Style.FlexDirection = layout.FlexDirectionColumn
		case "column-reverse":
			node.Style.FlexDirection = layout.FlexDirectionColumnReverse
		}
	}

	if style.JustifyContent != "" {
		switch style.JustifyContent {
		case "flex-start":
			node.Style.JustifyContent = layout.JustifyContentFlexStart
		case "flex-end":
			node.Style.JustifyContent = layout.JustifyContentFlexEnd
		case "center":
			node.Style.JustifyContent = layout.JustifyContentCenter
		case "space-between":
			node.Style.JustifyContent = layout.JustifyContentSpaceBetween
		case "space-around":
			node.Style.JustifyContent = layout.JustifyContentSpaceAround
		case "space-evenly":
			node.Style.JustifyContent = layout.JustifyContentSpaceEvenly
		}
	}

	if style.AlignItems != "" {
		switch style.AlignItems {
		case "flex-start":
			node.Style.AlignItems = layout.AlignItemsFlexStart
		case "flex-end":
			node.Style.AlignItems = layout.AlignItemsFlexEnd
		case "center":
			node.Style.AlignItems = layout.AlignItemsCenter
		case "stretch":
			node.Style.AlignItems = layout.AlignItemsStretch
		case "baseline":
			node.Style.AlignItems = layout.AlignItemsBaseline
		}
	}

	if style.AlignContent != "" {
		switch style.AlignContent {
		case "flex-start":
			node.Style.AlignContent = layout.AlignContentFlexStart
		case "flex-end":
			node.Style.AlignContent = layout.AlignContentFlexEnd
		case "center":
			node.Style.AlignContent = layout.AlignContentCenter
		case "stretch":
			node.Style.AlignContent = layout.AlignContentStretch
		case "space-between":
			node.Style.AlignContent = layout.AlignContentSpaceBetween
		case "space-around":
			node.Style.AlignContent = layout.AlignContentSpaceAround
		}
	}

	if style.FlexWrap != "" {
		switch style.FlexWrap {
		case "nowrap":
			node.Style.FlexWrap = layout.FlexWrapNoWrap
		case "wrap":
			node.Style.FlexWrap = layout.FlexWrapWrap
		case "wrap-reverse":
			node.Style.FlexWrap = layout.FlexWrapWrapReverse
		}
	}

	if style.FlexGrow != nil {
		node.Style.FlexGrow = *style.FlexGrow
	}
	if style.FlexShrink != nil {
		node.Style.FlexShrink = *style.FlexShrink
	}

	// Grid properties (basic support)
	if style.GridTemplateColumns != "" {
		// Parse grid track definitions - simplified, just convert string
		node.Style.GridTemplateColumns = parseGridTracks(style.GridTemplateColumns)
	}
	if style.GridTemplateRows != "" {
		node.Style.GridTemplateRows = parseGridTracks(style.GridTemplateRows)
	}
	if style.GridGap != nil {
		node.Style.GridGap = *style.GridGap
	}

	// Text
	if tree.Text != "" {
		node.Text = tree.Text
	}

	// Recursively build children
	for _, childTree := range tree.Children {
		child := buildLayoutFromWPTTree(&childTree)
		node.Children = append(node.Children, child)
	}

	return node
}

// parseGridTracks converts a grid track definition string to []GridTrack
// This is a simplified version - full parsing would be more complex
func parseGridTracks(def string) []layout.GridTrack {
	// For now, just handle simple cases like "100px 1fr 2fr"
	// A full implementation would parse repeat(), minmax(), etc.
	parts := strings.Fields(def)
	tracks := make([]layout.GridTrack, 0, len(parts))

	for _, part := range parts {
		track := layout.GridTrack{}

		// Simple parsing - just handle px and fr units
		if strings.HasSuffix(part, "px") {
			size := 0.0
			fmt.Sscanf(part, "%fpx", &size)
			track.MinSize = size
			track.MaxSize = size
		} else if strings.HasSuffix(part, "fr") {
			fr := 0.0
			fmt.Sscanf(part, "%ffr", &fr)
			// For fr units, use ratio in MaxSize
			track.MinSize = 0
			track.MaxSize = fr
		} else if part == "auto" {
			track.MinSize = -1
			track.MaxSize = math.MaxFloat64
		}

		tracks = append(tracks, track)
	}

	return tracks
}

func getConstraints(spec *ConstraintsSpec) layout.Constraints {
	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  math.MaxFloat64,
		MinHeight: 0,
		MaxHeight: math.MaxFloat64,
	}

	if spec != nil {
		if spec.MinWidth != nil {
			constraints.MinWidth = *spec.MinWidth
		}
		if spec.MaxWidth != nil {
			constraints.MaxWidth = *spec.MaxWidth
		}
		if spec.MinHeight != nil {
			constraints.MinHeight = *spec.MinHeight
		}
		if spec.MaxHeight != nil {
			constraints.MaxHeight = *spec.MaxHeight
		}
	}

	return constraints
}

func buildLayoutOutput(node *layout.Node) *LayoutOutput {
	if node == nil {
		return nil
	}

	output := &LayoutOutput{
		X:      node.Rect.X,
		Y:      node.Rect.Y,
		Width:  node.Rect.Width,
		Height: node.Rect.Height,
	}

	for _, child := range node.Children {
		if childOutput := buildLayoutOutput(child); childOutput != nil {
			output.Children = append(output.Children, *childOutput)
		}
	}

	return output
}
