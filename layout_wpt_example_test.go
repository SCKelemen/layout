package layout_test

import (
	"testing"

	"github.com/SCKelemen/layout"
	"github.com/SCKelemen/wpt-test-gen/pkg/cel"
)

// Example: Testing a flexbox layout with CEL assertions
// This demonstrates how to use the WPT testing infrastructure from wpt-test-gen
func TestFlexboxWithCELAssertions(t *testing.T) {
	// Build a flexbox layout
	root := &layout.Node{
		Style: layout.Style{
			Display:        layout.DisplayFlex,
			FlexDirection:  layout.FlexDirectionRow,
			JustifyContent: layout.JustifyContentSpaceBetween,
			AlignItems:     layout.AlignItemsCenter,
			Width:          layout.Px(600),
			Height:         layout.Px(100),
		},
		Children: []*layout.Node{
			{Style: layout.Style{Width: layout.Px(100), Height: layout.Px(50)}},
			{Style: layout.Style{Width: layout.Px(100), Height: layout.Px(50)}},
			{Style: layout.Style{Width: layout.Px(100), Height: layout.Px(50)}},
		},
	}

	// Run layout algorithm
	ctx := layout.NewLayoutContext(800, 600, 16)
	layout.Layout(root, layout.Constraints{
		MinWidth:  0,
		MaxWidth:  800,
		MinHeight: 0,
		MaxHeight: 600,
	}, ctx)

	// Create CEL environment for assertions
	env, err := cel.NewLayoutCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create CEL environment: %v", err)
	}

	// Define assertions using CEL expressions
	assertions := []cel.CELAssertion{
		{
			Type:       "layout",
			Expression: "getX(child(root(), 0)) == 0.0",
			Message:    "first-child-at-start",
		},
		{
			Type:       "layout",
			Expression: "getRight(child(root(), 2)) == getWidth(root())",
			Message:    "last-child-at-end",
		},
		{
			Type:       "layout",
			Expression: "getY(child(root(), 0)) == (getHeight(root()) - getHeight(child(root(), 0))) / 2.0",
			Message:    "children-vertically-centered",
		},
		{
			Type:       "layout",
			Expression: "getX(child(root(), 1)) - getRight(child(root(), 0)) == getX(child(root(), 2)) - getRight(child(root(), 1))",
			Message:    "equal-spacing-between-children",
		},
	}

	// Evaluate all assertions
	results := env.EvaluateAll(assertions)

	// Check results
	for _, result := range results {
		if !result.Passed {
			t.Errorf("Assertion failed: %s\nExpression: %s\nError: %s",
				result.Assertion.Message,
				result.Assertion.Expression,
				result.Error)
		}
	}
}

// Example: Testing grid layout with CEL assertions
func TestGridWithCELAssertions(t *testing.T) {
	// Build a simple 2x2 grid
	root := &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayGrid,
			GridTemplateColumns: []layout.GridTrack{
				{MinSize: layout.Px(100), MaxSize: layout.Px(100)},
				{MinSize: layout.Px(100), MaxSize: layout.Px(100)},
			},
			GridTemplateRows: []layout.GridTrack{
				{MinSize: layout.Px(50), MaxSize: layout.Px(50)},
				{MinSize: layout.Px(50), MaxSize: layout.Px(50)},
			},
			GridGap: layout.Px(10),
			Width:   layout.Px(210), // 2*100 + 10
			Height:  layout.Px(110), // 2*50 + 10
		},
		Children: []*layout.Node{
			{Style: layout.Style{}}, // Grid item [0,0]
			{Style: layout.Style{}}, // Grid item [1,0]
			{Style: layout.Style{}}, // Grid item [0,1]
			{Style: layout.Style{}}, // Grid item [1,1]
		},
	}

	// Run layout
	ctx := layout.NewLayoutContext(800, 600, 16)
	layout.Layout(root, layout.Tight(210, 110), ctx)

	// Create CEL environment
	env, err := cel.NewLayoutCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create CEL environment: %v", err)
	}

	// Define grid-specific assertions
	assertions := []cel.CELAssertion{
		{
			Type:       "layout",
			Expression: "getWidth(child(root(), 0)) == 100.0",
			Message:    "grid-cell-width",
		},
		{
			Type:       "layout",
			Expression: "getHeight(child(root(), 0)) == 50.0",
			Message:    "grid-cell-height",
		},
		{
			Type:       "layout",
			Expression: "getX(child(root(), 1)) == 110.0", // 100 + 10 gap
			Message:    "grid-column-gap",
		},
		{
			Type:       "layout",
			Expression: "getY(child(root(), 2)) == 60.0", // 50 + 10 gap
			Message:    "grid-row-gap",
		},
	}

	// Evaluate assertions
	results := env.EvaluateAll(assertions)

	// Check results
	for _, result := range results {
		if !result.Passed {
			t.Errorf("Assertion failed: %s\nExpression: %s\nError: %s",
				result.Assertion.Message,
				result.Assertion.Expression,
				result.Error)
		}
	}
}
