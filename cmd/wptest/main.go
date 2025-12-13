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

	if err := app.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newApp() *clix.App {
	app := clix.NewApp("wptest",
		clix.WithVersion(version),
		clix.WithDescription("WPT (Web Platform Test) CLI for layout testing"),
		clix.WithLongHelp(`wptest is a CLI tool for working with Web Platform Tests.

It provides commands for:
  - Running layout tests with CEL assertions
  - Generating test files in different languages
  - Listing available tests
  - Selecting CEL API bindings (old/new/context)
  - Type-directed fuzzing (future)
  - Property-based testing (future)`),
	)

	// Add commands
	app.AddCommand(newRunCommand())
	app.AddCommand(newListCommand())
	app.AddCommand(newGenerateCommand())

	return app
}
