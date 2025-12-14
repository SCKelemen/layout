package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

// Example: Building a Dashboard with Fluent API
//
// Demonstrates building complex layouts using method chaining
// and reusable component functions.

// CreateMetricCard creates a metric card with title, value, and trend
func CreateMetricCard(title, value string, trend float64, width float64) *layout.Node {
	trendText := fmt.Sprintf("%.1f%%", trend)
	if trend > 0 {
		trendText = "↑ " + trendText
	} else if trend < 0 {
		trendText = "↓ " + trendText
	}

	return (&layout.Node{}).
		WithStyle(layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionColumn,
			Width:         layout.Px(width),
		}).
		WithPadding(16).
		WithMargin(8).
		AddChildren(
			(&layout.Node{}).WithText(title).WithHeight(24),
			(&layout.Node{}).WithText(value).WithHeight(48),
			(&layout.Node{}).WithText(trendText).WithHeight(20),
		)
}

// CreateHeader creates a dashboard header with title and action button
func CreateHeader(title, buttonText string) *layout.Node {
	return layout.HStack(
		(&layout.Node{}).WithText(title).WithFlexGrow(1),
		(&layout.Node{}).WithText(buttonText).WithWidth(100).WithHeight(40),
	).WithPadding(16).WithMargin(8)
}

// CreateSection creates a section with title and content
func CreateSection(title string, content *layout.Node) *layout.Node {
	return layout.VStack(
		(&layout.Node{}).WithText(title).WithHeight(32).WithPadding(8),
		content,
	).WithMargin(8)
}

func main() {
	fmt.Println("=== Dashboard Example with Fluent API ===")

	// Build metrics row
	metricsRow := layout.HStack(
		CreateMetricCard("Revenue", "$125K", 12.5, 200),
		CreateMetricCard("Users", "8,234", -2.1, 200),
		CreateMetricCard("Orders", "342", 8.7, 200),
		CreateMetricCard("Conversion", "3.2%", 0.5, 200),
	).WithStyle(layout.Style{
		JustifyContent: layout.JustifyContentSpaceBetween,
	})

	// Build chart placeholder
	chartSection := CreateSection("Sales Overview",
		(&layout.Node{}).
			WithHeight(300).
			WithStyle(layout.Style{
				Display: layout.DisplayFlex,
			}).
			WithText("[Chart goes here]"),
	)

	// Build activity list
	activities := layout.VStack(
		(&layout.Node{}).WithText("New order #1234").WithHeight(40).WithPadding(8),
		(&layout.Node{}).WithText("User signed up").WithHeight(40).WithPadding(8),
		(&layout.Node{}).WithText("Payment received").WithHeight(40).WithPadding(8),
	)

	activitySection := CreateSection("Recent Activity", activities)

	// Build main dashboard layout
	dashboard := layout.VStack(
		CreateHeader("Dashboard", "Refresh"),
		metricsRow,
		layout.HStack(
			chartSection.WithFlexGrow(2),
			activitySection.WithFlexGrow(1),
		),
	).WithPadding(20)

	// Layout the dashboard
	constraints := layout.Loose(1200, 800)
	ctx := layout.NewLayoutContext(1200, 800, 16)
	layout.Layout(dashboard, constraints, ctx)

	// Print results
	fmt.Printf("Dashboard size: %.0fx%.0f\n",
		dashboard.Rect.Width, dashboard.Rect.Height)

	// Count nodes
	totalNodes := len(dashboard.DescendantsAndSelf())
	fmt.Printf("Total nodes in dashboard: %d\n", totalNodes)

	// Find all metric cards (nodes with 3 children)
	metricCards := dashboard.FindAll(func(n *layout.Node) bool {
		return len(n.Children) == 3 &&
			n.Style.Display == layout.DisplayFlex &&
			n.Style.FlexDirection == layout.FlexDirectionColumn &&
			n.Style.Width.Value == 200 && n.Style.Width.Unit == layout.Pixels
	})
	fmt.Printf("Metric cards found: %d\n", len(metricCards))

	// Show positions of metric cards
	fmt.Println("\nMetric card positions:")
	for i, card := range metricCards {
		if len(card.Children) > 0 {
			title := card.Children[0].Text
			fmt.Printf("  %d. %s at (%.0f, %.0f)\n",
				i+1, title, card.Rect.X, card.Rect.Y)
		}
	}

	// Demonstrate creating themed variant
	fmt.Println("\n=== Creating Dark Theme Variant ===")

	// Apply theme by transforming all containers
	darkTheme := dashboard.Transform(
		func(n *layout.Node) bool {
			return n.Style.Display == layout.DisplayFlex
		},
		func(n *layout.Node) *layout.Node {
			// Increase padding for dark theme
			return n.WithPadding(n.Style.Padding.Top.Value + 4)
		},
	)

	layout.Layout(darkTheme, constraints, ctx)
	fmt.Printf("Dark theme created (same structure, different styling)\n")
	fmt.Printf("Original first metric padding: %.0f\n", metricCards[0].Style.Padding.Top.Value)

	darkMetrics := darkTheme.FindAll(func(n *layout.Node) bool {
		return len(n.Children) == 3 &&
			n.Style.Display == layout.DisplayFlex &&
			n.Style.Width.Value == 200 && n.Style.Width.Unit == layout.Pixels
	})
	if len(darkMetrics) > 0 {
		fmt.Printf("Dark theme first metric padding: %.0f\n", darkMetrics[0].Style.Padding.Top.Value)
	}
}
