package main

import (
	"context"
	"fmt"
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

	cmd := clix.NewCommand("run")
	cmd.Short = "Run a WPT test with CEL assertions"
	cmd.Long = `Run a Web Platform Test with CEL assertions.

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
  wptest run test.json --verbose`

	cmd.Run = func(ctx *clix.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("test-file argument required")
		}
		return runTest(ctx.Context, ctx.Args[0], opts)
	}

	cmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "binding",
			Short: "b",
			Usage: "CEL API binding (old|context)",
		},
		Default: "old",
		Value:   &opts.binding,
	})

	cmd.Flags.BoolVar(clix.BoolVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "verbose",
			Short: "v",
			Usage: "Verbose output",
		},
		Value: &opts.verbose,
	})

	return cmd
}

func runTest(ctx context.Context, path string, opts *RunOptions) error {
	// Load test
	test, err := layout.LoadWPTTest(path)
	if err != nil {
		return fmt.Errorf("failed to load test: %w", err)
	}

	fmt.Printf("Running test: %s\n", test.Title)
	fmt.Printf("Description: %s\n", test.Description)
	fmt.Printf("Binding: %s\n\n", opts.binding)

	// Build layout tree from test
	root, err := test.BuildLayout()
	if err != nil {
		return fmt.Errorf("failed to build layout: %w", err)
	}

	// Run layout
	constraints := test.GetConstraints()
	layout.Layout(root, constraints)

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
