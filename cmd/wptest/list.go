package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/layout"
)

type ListOptions struct {
	pattern string
	verbose bool
}

func newListCommand() *clix.Command {
	opts := &ListOptions{}

	cmd := clix.NewCommand("list",
		clix.WithDescription("List available WPT tests"),
		clix.WithLongHelp(`List Web Platform Tests in a directory.

By default, searches for *.json files in the current directory.
You can specify a different directory or use glob patterns.

Examples:
  wptest list
  wptest list tests/
  wptest list --pattern "*flexbox*"
  wptest list --verbose`),
		clix.WithArgs(clix.Args{
			clix.NewArg("directory", clix.WithArgDefault(".")),
		}),
		clix.WithHandler(func(ctx context.Context, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			return listTests(ctx, dir, opts)
		}),
	)

	flags := cmd.Flags()
	flags.StringVar(&opts.pattern, "pattern", "*", "Glob pattern to match test files")
	flags.StringVar(&opts.pattern, "p", "*", "Glob pattern to match test files")
	flags.BoolVar(&opts.verbose, "verbose", false, "Show detailed test information")
	flags.BoolVar(&opts.verbose, "v", false, "Show detailed test information")

	return cmd
}

func listTests(ctx context.Context, dir string, opts *ListOptions) error {
	// Find JSON files
	pattern := filepath.Join(dir, opts.pattern+".json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob pattern: %w", err)
	}

	if len(matches) == 0 {
		fmt.Printf("No test files found matching: %s\n", pattern)
		return nil
	}

	fmt.Printf("Found %d test(s) in %s:\n\n", len(matches), dir)

	for i, match := range matches {
		// Read test file
		data, err := os.ReadFile(match)
		if err != nil {
			fmt.Printf("%d. %s [Error reading file]\n", i+1, filepath.Base(match))
			continue
		}

		var test layout.WPTTest
		if err := json.Unmarshal(data, &test); err != nil {
			fmt.Printf("%d. %s [Error parsing JSON]\n", i+1, filepath.Base(match))
			continue
		}

		// Display basic info
		fmt.Printf("%d. %s\n", i+1, filepath.Base(match))
		fmt.Printf("   ID: %s\n", test.ID)
		fmt.Printf("   Title: %s\n", test.Title)

		if opts.verbose {
			fmt.Printf("   Description: %s\n", test.Description)

			// Count assertions
			assertionCount := 0
			for _, browserResult := range test.Results {
				for _, elem := range browserResult.Elements {
					assertionCount += len(elem.Assertions)
				}
			}
			fmt.Printf("   Assertions: %d\n", assertionCount)

			// Show categories/tags
			if len(test.Categories) > 0 {
				fmt.Printf("   Categories: %s\n", strings.Join(test.Categories, ", "))
			}
			if len(test.Tags) > 0 {
				fmt.Printf("   Tags: %s\n", strings.Join(test.Tags, ", "))
			}
		}
		fmt.Println()
	}

	return nil
}
