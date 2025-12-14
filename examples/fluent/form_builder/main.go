package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

// Example: Building Forms with Fluent API
//
// Demonstrates building reusable form components
// and creating complex forms declaratively.

// FormField creates a labeled input field
func FormField(label, placeholder string, width float64) *layout.Node {
	return layout.VStack(
		(&layout.Node{}).WithText(label).WithHeight(24),
		(&layout.Node{}).
			WithText(placeholder).
			WithWidth(width).
			WithHeight(40).
			WithPadding(8),
	).WithMargin(8)
}

// FormRow creates a horizontal row of form fields
func FormRow(fields ...*layout.Node) *layout.Node {
	return layout.HStack(fields...).
		WithStyle(layout.Style{
			JustifyContent: layout.JustifyContentSpaceBetween,
		}).
		WithMargin(4)
}

// FormSection creates a section with title and fields
func FormSection(title string, fields ...*layout.Node) *layout.Node {
	header := (&layout.Node{}).
		WithText(title).
		WithHeight(32).
		WithPadding(8)

	return layout.VStack(append([]*layout.Node{header}, fields...)...).
		WithMargin(16)
}

// ButtonGroup creates a row of action buttons
func ButtonGroup(buttons ...string) *layout.Node {
	nodes := make([]*layout.Node, len(buttons))
	for i, text := range buttons {
		nodes[i] = (&layout.Node{}).
			WithText(text).
			WithWidth(100).
			WithHeight(40).
			WithMargin(4)
	}

	return layout.HStack(nodes...).
		WithStyle(layout.Style{
			JustifyContent: layout.JustifyContentFlexEnd,
		})
}

func main() {
	fmt.Println("=== Form Builder Example ===")

	// Build a user registration form
	registrationForm := layout.VStack(
		// Form header
		(&layout.Node{}).
			WithText("User Registration").
			WithHeight(50).
			WithPadding(16),

		// Personal information section
		FormSection("Personal Information",
			FormRow(
				FormField("First Name", "Enter first name", 200),
				FormField("Last Name", "Enter last name", 200),
			),
			FormField("Email", "your@email.com", 420),
			FormRow(
				FormField("Phone", "(555) 123-4567", 200),
				FormField("Birthday", "MM/DD/YYYY", 200),
			),
		),

		// Address section
		FormSection("Address",
			FormField("Street Address", "123 Main St", 420),
			FormRow(
				FormField("City", "City", 200),
				FormField("State", "ST", 80),
				FormField("ZIP", "12345", 120),
			),
		),

		// Account section
		FormSection("Account",
			FormRow(
				FormField("Username", "username", 200),
				FormField("Password", "••••••••", 200),
			),
		),

		// Action buttons
		ButtonGroup("Cancel", "Save Draft", "Submit").
			WithPadding(16),
	).WithWidth(500).WithPadding(20)

	// Layout the form
	ctx := layout.NewLayoutContext(600, 800, 16)
	layout.Layout(registrationForm, layout.Loose(600, 800), ctx)

	fmt.Printf("Form size: %.0fx%.0f\n",
		registrationForm.Rect.Width, registrationForm.Rect.Height)

	// Analyze the form
	fmt.Println("\n=== Form Analysis ===")

	// Count form fields
	fields := registrationForm.FindAll(func(n *layout.Node) bool {
		return n.Style.Height.Value == 40 && n.Style.Width.Value > 0
	})
	fmt.Printf("Total form fields: %d\n", len(fields))

	// Count sections
	sections := registrationForm.FindAll(func(n *layout.Node) bool {
		return len(n.Children) > 0 &&
			n.Style.Margin.Top.Value == 16
	})
	fmt.Printf("Form sections: %d\n", len(sections))

	// Find all labels (height 24 nodes)
	labels := registrationForm.FindAll(func(n *layout.Node) bool {
		return n.Style.Height.Value == 24 && n.Text != ""
	})
	fmt.Printf("Form labels: %d\n", len(labels))
	fmt.Printf("Labels: ")
	for i, label := range labels {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%s", label.Text)
	}
	fmt.Println()

	// Demonstrate form variants
	fmt.Println("\n=== Creating Form Variants ===")

	// Compact variant - reduce all spacing
	compactForm := registrationForm.Transform(
		func(n *layout.Node) bool {
			return n.Style.Margin.Top.Value > 0 || n.Style.Padding.Top.Value > 0
		},
		func(n *layout.Node) *layout.Node {
			return n.
				WithMargin(n.Style.Margin.Top.Value / 2).
				WithPadding(n.Style.Padding.Top.Value / 2)
		},
	)

	ctx2 := layout.NewLayoutContext(800, 600, 16)
	layout.Layout(compactForm, layout.Loose(600, 800), ctx2)
	fmt.Printf("Compact form height: %.0f (original: %.0f)\n",
		compactForm.Rect.Height, registrationForm.Rect.Height)

	// Wide variant - scale all widths
	wideForm := registrationForm.Transform(
		func(n *layout.Node) bool {
			return n.Style.Width.Value > 0
		},
		func(n *layout.Node) *layout.Node {
			return n.WithWidth(n.Style.Width.Value * 1.3)
		},
	)

	ctx3 := layout.NewLayoutContext(800, 600, 16)
	layout.Layout(wideForm, layout.Loose(800, 800), ctx3)
	fmt.Printf("Wide form width: %.0f (original: %.0f)\n",
		wideForm.Rect.Width, registrationForm.Rect.Width)

	// Demonstrate conditional fields
	fmt.Println("\n=== Conditional Form Fields ===")

	// Remove password field for "view only" mode
	viewOnlyForm := registrationForm.Transform(
		func(n *layout.Node) bool {
			// Check if this is a section containing password
			for _, child := range n.Children {
				if child.Text == "Password" {
					return true
				}
			}
			return false
		},
		func(n *layout.Node) *layout.Node {
			// Remove children with "Password" text
			return n.FilterDeep(func(child *layout.Node) bool {
				return child.Text != "Password" && child.Text != "password"
			})
		},
	)

	ctx4 := layout.NewLayoutContext(800, 600, 16)
	layout.Layout(viewOnlyForm, layout.Loose(600, 800), ctx4)

	viewFields := viewOnlyForm.FindAll(func(n *layout.Node) bool {
		return n.Style.Height.Value == 40 && n.Style.Width.Value > 0
	})
	fmt.Printf("View-only form fields: %d (removed password field)\n", len(viewFields))

	// Demonstrate form validation display
	fmt.Println("\n=== Adding Validation Errors ===")

	// Add error indicators to specific fields
	formWithErrors := registrationForm.Transform(
		func(n *layout.Node) bool {
			// Find email and phone fields
			return n.Text == "your@email.com" || n.Text == "(555) 123-4567"
		},
		func(n *layout.Node) *layout.Node {
			// Add error indicator (simulated with extra height)
			return n.WithHeight(n.Style.Height.Value + 20) // Space for error message
		},
	)

	ctx5 := layout.NewLayoutContext(800, 600, 16)
	layout.Layout(formWithErrors, layout.Loose(600, 800), ctx5)
	fmt.Printf("Form with validation errors height: %.0f\n", formWithErrors.Rect.Height)

	// Demonstrate collecting form data (simulated)
	fmt.Println("\n=== Form Data Collection ===")

	type FormData struct {
		Labels []string
		Values []string
	}

	formData := registrationForm.Fold(FormData{}, func(acc interface{}, n *layout.Node) interface{} {
		data := acc.(FormData)

		// Collect labels (height 24)
		if n.Style.Height.Value == 24 && n.Text != "" {
			data.Labels = append(data.Labels, n.Text)
		}

		// Collect field placeholders (height 40)
		if n.Style.Height.Value == 40 && n.Text != "" {
			data.Values = append(data.Values, n.Text)
		}

		return data
	}).(FormData)

	fmt.Printf("Collected data:\n")
	fmt.Printf("  Labels: %d\n", len(formData.Labels))
	fmt.Printf("  Values: %d\n", len(formData.Values))

	// Show field mapping
	fmt.Println("\nField mapping:")
	for i := 0; i < len(formData.Labels) && i < len(formData.Values); i++ {
		fmt.Printf("  %s: %s\n", formData.Labels[i], formData.Values[i])
	}

	// Demonstrate form analytics
	fmt.Println("\n=== Form Analytics ===")

	// Count nodes by depth to understand nesting
	depthCounts := registrationForm.FoldWithContext(
		make(map[int]int),
		func(acc interface{}, n *layout.Node, depth int) interface{} {
			m := acc.(map[int]int)
			m[depth]++
			return m
		},
	).(map[int]int)

	fmt.Printf("Nodes by depth: %v\n", depthCounts)

	// Calculate total padding in form
	totalPadding := registrationForm.Fold(0.0, func(acc interface{}, n *layout.Node) interface{} {
		sum := acc.(float64)
		p := n.Style.Padding
		return sum + p.Top.Value + p.Right.Value + p.Bottom.Value + p.Left.Value
	}).(float64)

	fmt.Printf("Total padding in form: %.0f\n", totalPadding)

	fmt.Println("\n=== Original Form Unchanged ===")
	fmt.Printf("Original still has %d fields\n", len(fields))
	fmt.Printf("Original size: %.0fx%.0f\n",
		registrationForm.Rect.Width, registrationForm.Rect.Height)
}
