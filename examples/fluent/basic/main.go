package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

// Example: Basic Fluent API Usage
//
// Demonstrates method chaining and immutable modifications.

func main() {
	fmt.Println("=== Basic Fluent API Example ===")

	// Build a simple card using method chaining
	card := (&layout.Node{}).
		WithStyle(layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionColumn,
			Width:         layout.Px(300),
		}).
		WithPadding(20).
		WithMargin(10).
		AddChildren(
			(&layout.Node{}).WithText("Card Title").WithHeight(40),
			(&layout.Node{}).WithText("Card body content").WithFlexGrow(1),
			(&layout.Node{}).WithText("Card footer").WithHeight(30),
		)

	// Layout the card
	ctx := layout.NewLayoutContext(400, 600, 16)
	layout.Layout(card, layout.Loose(400, 600), ctx)

	fmt.Printf("Card rect: %.0fx%.0f at (%.0f, %.0f)\n",
		card.Rect.Width, card.Rect.Height, card.Rect.X, card.Rect.Y)
	fmt.Printf("Card has %d children\n", len(card.Children))

	// Demonstrate immutability - create variants
	fmt.Println("\n=== Creating Variants ===")

	// Original card
	fmt.Printf("Original padding: %.0f\n", card.Style.Padding.Top.Value)

	// Create wider variant
	wideCard := card.WithWidth(400)
	fmt.Printf("Wide variant width: %.0f (original: %.0f)\n",
		wideCard.Style.Width.Value, card.Style.Width.Value)

	// Create variant with more padding
	extraPadded := card.WithPadding(30)
	fmt.Printf("Extra padded: %.0f (original: %.0f)\n",
		extraPadded.Style.Padding.Top.Value, card.Style.Padding.Top.Value)

	// Create variant with additional child
	withExtra := card.AddChild((&layout.Node{}).WithText("Extra item"))
	fmt.Printf("With extra child: %d children (original: %d)\n",
		len(withExtra.Children), len(card.Children))

	// Demonstrate chaining multiple operations
	fmt.Println("\n=== Chaining Operations ===")

	styledCard := card.
		WithPadding(25).
		WithMargin(15).
		WithWidth(350).
		AddChild((&layout.Node{}).WithText("Another item"))

	fmt.Printf("Styled card: %.0f padding, %.0f margin, %.0f width, %d children\n",
		styledCard.Style.Padding.Top.Value,
		styledCard.Style.Margin.Top.Value,
		styledCard.Style.Width.Value,
		len(styledCard.Children))

	fmt.Printf("Original card unchanged: %.0f padding, %d children\n",
		card.Style.Padding.Top.Value, len(card.Children))
}
