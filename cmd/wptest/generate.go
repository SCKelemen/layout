package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/layout"
)

type GenerateOptions struct {
	output      string
	packageName string
	standalone  bool
	binding     string
	lang        string
}

func newGenerateCommand() *clix.Command {
	opts := &GenerateOptions{}

	cmd := clix.NewCommand("generate")
	cmd.Short = "Generate test files from WPT JSON"
	cmd.Long = `Generate test files in various languages from WPT JSON specifications.

Generation modes:
  - standalone: Complete tests using the layout library
  - user-extensible: Template tests for custom implementations

Binding options:
  - old: Original CEL API (default, compatible with generated tests)
  - context: Path-based API with this() and parent() support

Language options:
  - go: Generate Go test files (default)
  - rust: Generate Rust tests (future)
  - js: Generate JavaScript tests (future)

Examples:
  wptest generate test.json
  wptest generate test.json --output my_test.go
  wptest generate test.json --standalone
  wptest generate test.json --standalone --binding context
  wptest generate test.json --package mypackage`

	cmd.Run = func(ctx *clix.Context) error {
		if len(ctx.Args) < 1 {
			return fmt.Errorf("test-file argument required")
		}
		return generateTest(ctx.Context, ctx.Args[0], opts)
	}

	cmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "output",
			Short: "o",
			Usage: "Output file (default: derived from input)",
		},
		Value: &opts.output,
	})

	cmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "package",
			Short: "p",
			Usage: "Package name",
		},
		Default: "layout_test",
		Value:   &opts.packageName,
	})

	cmd.Flags.BoolVar(clix.BoolVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "standalone",
			Short: "s",
			Usage: "Generate standalone test",
		},
		Value: &opts.standalone,
	})

	cmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "binding",
			Short: "b",
			Usage: "CEL API binding (old|context)",
		},
		Default: "old",
		Value:   &opts.binding,
	})

	cmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "lang",
			Short: "l",
			Usage: "Target language (go|rust|js)",
		},
		Default: "go",
		Value:   &opts.lang,
	})

	return cmd
}

func generateTest(ctx context.Context, input string, opts *GenerateOptions) error {
	// Validate language
	if opts.lang != "go" {
		return fmt.Errorf("language %s not yet supported (only 'go' currently)", opts.lang)
	}

	// Validate binding
	if opts.binding != "old" && opts.binding != "context" {
		return fmt.Errorf("binding %s not supported (use: old|context)", opts.binding)
	}

	// Read test file to validate
	data, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("failed to read test file: %w", err)
	}

	var test layout.WPTTest
	if err := json.Unmarshal(data, &test); err != nil {
		return fmt.Errorf("failed to parse test JSON: %w", err)
	}

	// Determine output file
	outFile := opts.output
	if outFile == "" {
		base := filepath.Base(input)
		name := strings.TrimSuffix(base, filepath.Ext(base))
		outFile = name + "_test.go"
	}

	// Build generator arguments
	args := []string{
		"run",
		"./tools/wpt_test_gen",
		"-input", input,
		"-output", outFile,
		"-package", opts.packageName,
	}

	if opts.standalone {
		args = append(args, "-standalone")
	}

	fmt.Printf("Generating test file...\n")
	fmt.Printf("  Input: %s\n", input)
	fmt.Printf("  Output: %s\n", outFile)
	fmt.Printf("  Package: %s\n", opts.packageName)
	fmt.Printf("  Mode: %s\n", map[bool]string{true: "standalone", false: "user-extensible"}[opts.standalone])
	fmt.Printf("  Binding: %s\n", opts.binding)
	fmt.Println()

	// Run generator
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("generator failed: %w", err)
	}

	// Update binding if context is requested
	if opts.binding == "context" && opts.standalone {
		fmt.Println("Updating CEL binding to context-aware...")
		if err := updateBindingToContext(outFile); err != nil {
			return fmt.Errorf("failed to update binding: %w", err)
		}
	}

	fmt.Printf("\n✓ Generated: %s\n", outFile)
	return nil
}

func updateBindingToContext(filePath string) error {
	// Read generated file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	text := string(content)

	// Replace NewLayoutCELEnv with NewLayoutCELEnvWithContext
	text = strings.ReplaceAll(text,
		"layout.NewLayoutCELEnv(root)",
		"layout.NewLayoutCELEnvWithContext(root)")

	// Update comment
	text = strings.ReplaceAll(text,
		"// Create CEL environment",
		"// Create CEL environment with context (this/parent support)")

	// Write back
	return os.WriteFile(filePath, []byte(text), 0644)
}
