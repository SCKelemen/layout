package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/SCKelemen/clix"
)

var version = "1.0.0"

func main() {
	// Setup cancellation
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app := newApp()

	if err := app.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newApp() *clix.App {
	app := clix.NewApp("wptest")
	app.Version = version
	app.Description = "WPT (Web Platform Test) CLI for layout testing"

	app.Root.Long = `wptest is a CLI tool for working with Web Platform Tests.

It provides commands for:
  - Running layout tests with CEL assertions
  - Generating test files in different languages
  - Listing available tests
  - Selecting CEL API bindings (old/new/context)
  - Type-directed fuzzing (future)
  - Property-based testing (future)`

	// Add commands
	app.Root.AddCommand(newRunCommand())
	app.Root.AddCommand(newListCommand())
	app.Root.AddCommand(newGenerateCommand())

	return app
}
