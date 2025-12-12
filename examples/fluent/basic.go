package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

// Example: Basic Fluent API Usage
//
// Demonstrates method chaining and immutable modifications.

func main() {
	fmt.Println("=== Basic Fluent API Example ===\n")

	// Build a simple card using method chaining
	card := (&layout.Node{}).
		WithStyle(layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionColumn,
			Width:         300,
		}).
		WithPadding(20).
		WithMargin(10).
		AddChildren(
			(&layout.Node{}).WithText("Card Title").WithHeight(40),
			(&layout.Node{}).WithText("Card body content").WithFlexGrow(1),
			(&layout.Node{}).WithText("Card footer").WithHeight(30),
		)

	// Layout the card
	layout.Layout(card, layout.Loose(400, 600))

	fmt.Printf("Card rect: %.0fx%.0f at (%.0f, %.0f)\n",
		card.Rect.Width, card.Rect.Height, card.Rect.X, card.Rect.Y)
	fmt.Printf("Card has %d children\n", len(card.Children))

	// Demonstrate immutability - create variants
	fmt.Println("\n=== Creating Variants ===\n")

	// Original card
	fmt.Printf("Original padding: %.0f\n", card.Style.Padding.Top)

	// Create wider variant
	wideCard := card.WithWidth(400)
	fmt.Printf("Wide variant width: %.0f (original: %.0f)\n",
		wideCard.Style.Width, card.Style.Width)

	// Create variant with more padding
	extraPadded := card.WithPadding(30)
	fmt.Printf("Extra padded: %.0f (original: %.0f)\n",
		extraPadded.Style.Padding.Top, card.Style.Padding.Top)

	// Create variant with additional child
	withExtra := card.AddChild((&layout.Node{}).WithText("Extra item"))
	fmt.Printf("With extra child: %d children (original: %d)\n",
		len(withExtra.Children), len(card.Children))

	// Demonstrate chaining multiple operations
	fmt.Println("\n=== Chaining Operations ===\n")

	styledCard := card.
		WithPadding(25).
		WithMargin(15).
		WithWidth(350).
		AddChild((&layout.Node{}).WithText("Another item"))

	fmt.Printf("Styled card: %.0f padding, %.0f margin, %.0f width, %d children\n",
		styledCard.Style.Padding.Top,
		styledCard.Style.Margin.Top,
		styledCard.Style.Width,
		len(styledCard.Children))

	fmt.Printf("Original card unchanged: %.0f padding, %d children\n",
		card.Style.Padding.Top, len(card.Children))
}
