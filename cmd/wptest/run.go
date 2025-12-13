package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/layout"
)

type RunOptions struct {
	binding string
	verbose bool
}

func newRunCommand() *clix.Command {
	opts := &RunOptions{}

	cmd := clix.NewCommand("run",
		clix.WithDescription("Run a WPT test with CEL assertions"),
		clix.WithLongHelp(`Run a Web Platform Test with CEL assertions.

The test JSON file should contain:
  - Layout specification
  - Constraints
  - CEL assertions to evaluate

Binding options:
  - old: Original CEL API (getX, getWidth functions)
  - new: Domain-structured CEL API (x(elem), width(elem))
  - context: Path-based API with this() and parent() support

Example:
  wptest run test.json
  wptest run test.json --binding context
  wptest run test.json --verbose`),
		clix.WithArgs(clix.Args{
			clix.NewArg("test-file", clix.WithArgRequired()),
		}),
		clix.WithHandler(func(ctx context.Context, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("test-file argument required")
			}
			return runTest(ctx, args[0], opts)
		}),
	)

	flags := cmd.Flags()
	flags.StringVar(&opts.binding, "binding", "old", "CEL API binding (old|context)")
	flags.StringVar(&opts.binding, "b", "old", "CEL API binding (old|context)") // Short form
	flags.BoolVar(&opts.verbose, "verbose", false, "Verbose output")
	flags.BoolVar(&opts.verbose, "v", false, "Verbose output") // Short form

	return cmd
}

func runTest(ctx context.Context, path string, opts *RunOptions) error {
	// Read test file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read test file: %w", err)
	}

	var test layout.WPTTest
	if err := json.Unmarshal(data, &test); err != nil {
		return fmt.Errorf("failed to parse test JSON: %w", err)
	}

	fmt.Printf("Running test: %s\n", test.Title)
	fmt.Printf("Description: %s\n", test.Description)
	fmt.Printf("Binding: %s\n\n", opts.binding)

	// Build layout tree from spec
	root := buildLayoutFromSpec(test.Layout)

	// Run layout
	layout.Layout(root, layout.Constraints{
		MinWidth:  0,
		MaxWidth:  test.Constraints.Width,
		MinHeight: 0,
		MaxHeight: test.Constraints.Height,
	})

	if opts.verbose {
		fmt.Printf("Layout computed:\n")
		fmt.Printf("  Root: x=%.1f, y=%.1f, w=%.1f, h=%.1f\n",
			root.Rect.X, root.Rect.Y, root.Rect.Width, root.Rect.Height)
		for i, child := range root.Children {
			fmt.Printf("  Child %d: x=%.1f, y=%.1f, w=%.1f, h=%.1f\n",
				i, child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
		}
		fmt.Println()
	}

	// Collect all assertions
	var allAssertions []layout.CELAssertion
	for _, browserResult := range test.Results {
		for _, elem := range browserResult.Elements {
			for _, assertion := range elem.Assertions {
				if assertion.Message == "" {
					assertion.Message = assertion.Type
				}
				allAssertions = append(allAssertions, assertion)
			}
		}
	}

	// Evaluate based on binding type
	var results []layout.AssertionResult

	switch opts.binding {
	case "old":
		env, err := layout.NewLayoutCELEnv(root)
		if err != nil {
			return fmt.Errorf("failed to create CEL environment: %w", err)
		}
		results = env.EvaluateAll(allAssertions)

	case "context":
		env, err := layout.NewLayoutCELEnvWithContext(root)
		if err != nil {
			return fmt.Errorf("failed to create CEL environment: %w", err)
		}
		for _, assertion := range allAssertions {
			results = append(results, env.Evaluate(assertion))
		}

	default:
		return fmt.Errorf("unsupported binding: %s (use: old|context)", opts.binding)
	}

	// Report results
	passed := 0
	failed := 0
	skipped := 0

	for _, result := range results {
		if result.Passed {
			passed++
			if opts.verbose {
				fmt.Printf("✓ [%s] %s\n", result.Assertion.Type, result.Assertion.Expression)
			}
		} else {
			// Check if skipped (unsupported features)
			if strings.Contains(result.Error, "undeclared reference to 'this'") ||
				strings.Contains(result.Error, "undeclared reference to 'parent'") {
				skipped++
				if opts.verbose {
					fmt.Printf("⊗ [%s] %s\n", result.Assertion.Type, result.Assertion.Expression)
					fmt.Printf("  Note: Uses unsupported features\n")
				}
			} else {
				failed++
				fmt.Printf("✗ [%s] %s\n", result.Assertion.Type, result.Assertion.Expression)
				fmt.Printf("  Error: %s\n", result.Error)
			}
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("Results:\n")
	fmt.Printf("  Total:   %d\n", len(results))
	fmt.Printf("  Passed:  %d\n", passed)
	fmt.Printf("  Failed:  %d\n", failed)
	fmt.Printf("  Skipped: %d\n", skipped)
	fmt.Printf("%s\n", strings.Repeat("=", 60))

	if failed > 0 {
		return fmt.Errorf("test failed with %d assertion failures", failed)
	}

	return nil
}

func buildLayoutFromSpec(spec layout.LayoutSpec) *layout.Node {
	node := &layout.Node{
		Style: layout.Style{},
	}

	// Set style properties
	if spec.Style.Display != nil {
		switch *spec.Style.Display {
		case "flex":
			node.Style.Display = layout.DisplayFlex
		case "block":
			node.Style.Display = layout.DisplayBlock
		case "none":
			node.Style.Display = layout.DisplayNone
		}
	}

	if spec.Style.FlexDirection != nil {
		switch *spec.Style.FlexDirection {
		case "row":
			node.Style.FlexDirection = layout.FlexDirectionRow
		case "column":
			node.Style.FlexDirection = layout.FlexDirectionColumn
		case "row-reverse":
			node.Style.FlexDirection = layout.FlexDirectionRowReverse
		case "column-reverse":
			node.Style.FlexDirection = layout.FlexDirectionColumnReverse
		}
	}

	if spec.Style.JustifyContent != nil {
		switch *spec.Style.JustifyContent {
		case "flex-start", "start":
			node.Style.JustifyContent = layout.JustifyContentFlexStart
		case "flex-end", "end":
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

	if spec.Style.AlignItems != nil {
		switch *spec.Style.AlignItems {
		case "stretch":
			node.Style.AlignItems = layout.AlignItemsStretch
		case "flex-start", "start":
			node.Style.AlignItems = layout.AlignItemsFlexStart
		case "flex-end", "end":
			node.Style.AlignItems = layout.AlignItemsFlexEnd
		case "center":
			node.Style.AlignItems = layout.AlignItemsCenter
		case "baseline":
			node.Style.AlignItems = layout.AlignItemsBaseline
		}
	}

	if spec.Style.Width != nil {
		node.Style.Width = *spec.Style.Width
	}

	if spec.Style.Height != nil {
		node.Style.Height = *spec.Style.Height
	}

	// Build children recursively
	for _, childSpec := range spec.Children {
		node.Children = append(node.Children, buildLayoutFromSpec(childSpec))
	}

	return node
}
