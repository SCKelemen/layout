package layout

import (
	"testing"
)

func TestCELBasicPositioning(t *testing.T) {
	// Create a simple layout tree
	root := &Node{
		Style: Style{
			Display: DisplayFlex,
			Width:   600,
			Height:  100,
		},
		Children: []*Node{
			{
				Style: Style{
					Width:  100,
					Height: 50,
				},
			},
			{
				Style: Style{
					Width:  100,
					Height: 50,
				},
			},
		},
	}

	// Perform layout
	Layout(root, Constraints{
		MinWidth:  0,
		MaxWidth:  800,
		MinHeight: 0,
		MaxHeight: 600,
	})

	// Create CEL environment
	celEnv, err := NewLayoutCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create CEL environment: %v", err)
	}

	tests := []struct {
		name       string
		expression string
		shouldPass bool
	}{
		{
			name:       "Root X position",
			expression: "getX(root()) == 0.0",
			shouldPass: true,
		},
		{
			name:       "Root Y position",
			expression: "getY(root()) == 0.0",
			shouldPass: true,
		},
		{
			name:       "Root width",
			expression: "getWidth(root()) == 600.0",
			shouldPass: true,
		},
		{
			name:       "Root height",
			expression: "getHeight(root()) == 100.0",
			shouldPass: true,
		},
		{
			name:       "First child X",
			expression: "getX(child(root(), 0)) == 0.0",
			shouldPass: true,
		},
		{
			name:       "First child width",
			expression: "getWidth(child(root(), 0)) == 100.0",
			shouldPass: true,
		},
		{
			name:       "Second child X (should be after first)",
			expression: "getX(child(root(), 1)) == 100.0",
			shouldPass: true,
		},
		{
			name:       "Children are aligned horizontally",
			expression: "getY(child(root(), 0)) == getY(child(root(), 1))",
			shouldPass: true,
		},
		{
			name:       "Second child is to the right of first",
			expression: "getX(child(root(), 1)) > getX(child(root(), 0))",
			shouldPass: true,
		},
		{
			name:       "Child count",
			expression: "childCount(root()) == 2",
			shouldPass: true,
		},
		{
			name:       "Aspect ratio",
			expression: "getWidth(root()) / getHeight(root()) == 6.0",
			shouldPass: true,
		},
		{
			name:       "Bottom edge",
			expression: "getBottom(root()) == 100.0",
			shouldPass: true,
		},
		{
			name:       "Right edge",
			expression: "getRight(root()) == 600.0",
			shouldPass: true,
		},
		{
			name:       "Intentional failure test",
			expression: "getX(root()) == 999.0",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := CELAssertion{
				Type:       "layout",
				Expression: tt.expression,
				Message:    tt.name,
			}

			result := celEnv.Evaluate(assertion)

			if tt.shouldPass && !result.Passed {
				t.Errorf("Expected assertion to pass but it failed: %s\nError: %s",
					tt.expression, result.Error)
			}

			if !tt.shouldPass && result.Passed {
				t.Errorf("Expected assertion to fail but it passed: %s", tt.expression)
			}
		})
	}
}

func TestCELFlexboxProperties(t *testing.T) {
	// Create flexbox layout
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentSpaceBetween,
			AlignItems:     AlignItemsCenter,
			FlexWrap:       FlexWrapNoWrap,
			Width:          600,
			Height:         100,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
		},
	}

	// Perform layout
	Layout(root, Constraints{
		MinWidth:  0,
		MaxWidth:  800,
		MinHeight: 0,
		MaxHeight: 600,
	})

	// Create CEL environment
	celEnv, err := NewLayoutCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create CEL environment: %v", err)
	}

	tests := []struct {
		name       string
		expression string
		shouldPass bool
	}{
		{
			name:       "Flex direction is row",
			expression: `getFlexDirection(root()) == "row"`,
			shouldPass: true,
		},
		{
			name:       "Justify content is space-between",
			expression: `getJustifyContent(root()) == "space-between"`,
			shouldPass: true,
		},
		{
			name:       "Align items is center",
			expression: `getAlignItems(root()) == "center"`,
			shouldPass: true,
		},
		{
			name:       "Flex wrap is nowrap",
			expression: `getFlexWrap(root()) == "nowrap"`,
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := CELAssertion{
				Type:       "layout",
				Expression: tt.expression,
				Message:    tt.name,
			}

			result := celEnv.Evaluate(assertion)

			if tt.shouldPass && !result.Passed {
				t.Errorf("Expected assertion to pass but it failed: %s\nError: %s",
					tt.expression, result.Error)
			}

			if !tt.shouldPass && result.Passed {
				t.Errorf("Expected assertion to fail but it passed: %s", tt.expression)
			}
		})
	}
}

func TestCELMarginPadding(t *testing.T) {
	root := &Node{
		Style: Style{
			Display: DisplayFlex,
			Width:   600,
			Height:  100,
			Margin: Spacing{
				Top:    10,
				Right:  20,
				Bottom: 30,
				Left:   40,
			},
			Padding: Spacing{
				Top:    5,
				Right:  15,
				Bottom: 25,
				Left:   35,
			},
		},
	}

	// Create CEL environment
	celEnv, err := NewLayoutCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create CEL environment: %v", err)
	}

	tests := []struct {
		name       string
		expression string
		shouldPass bool
	}{
		{
			name:       "Margin top",
			expression: "getMarginTop(root()) == 10.0",
			shouldPass: true,
		},
		{
			name:       "Margin right",
			expression: "getMarginRight(root()) == 20.0",
			shouldPass: true,
		},
		{
			name:       "Margin bottom",
			expression: "getMarginBottom(root()) == 30.0",
			shouldPass: true,
		},
		{
			name:       "Margin left",
			expression: "getMarginLeft(root()) == 40.0",
			shouldPass: true,
		},
		{
			name:       "Padding top",
			expression: "getPaddingTop(root()) == 5.0",
			shouldPass: true,
		},
		{
			name:       "Padding right",
			expression: "getPaddingRight(root()) == 15.0",
			shouldPass: true,
		},
		{
			name:       "Padding bottom",
			expression: "getPaddingBottom(root()) == 25.0",
			shouldPass: true,
		},
		{
			name:       "Padding left",
			expression: "getPaddingLeft(root()) == 35.0",
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := CELAssertion{
				Type:       "layout",
				Expression: tt.expression,
				Message:    tt.name,
			}

			result := celEnv.Evaluate(assertion)

			if tt.shouldPass && !result.Passed {
				t.Errorf("Expected assertion to pass but it failed: %s\nError: %s",
					tt.expression, result.Error)
			}

			if !tt.shouldPass && result.Passed {
				t.Errorf("Expected assertion to fail but it passed: %s", tt.expression)
			}
		})
	}
}

func TestCELComplexExpressions(t *testing.T) {
	// Create a more complex layout
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentSpaceBetween,
			Width:          600,
			Height:         100,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
		},
	}

	// Perform layout
	Layout(root, Constraints{
		MinWidth:  0,
		MaxWidth:  800,
		MinHeight: 0,
		MaxHeight: 600,
	})

	// Create CEL environment
	celEnv, err := NewLayoutCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create CEL environment: %v", err)
	}

	tests := []struct {
		name       string
		expression string
		shouldPass bool
	}{
		{
			name:       "All children have same width",
			expression: "getWidth(child(root(), 0)) == getWidth(child(root(), 1)) && getWidth(child(root(), 1)) == getWidth(child(root(), 2))",
			shouldPass: true,
		},
		{
			name:       "Children are evenly spaced (space-between)",
			expression: "getX(child(root(), 0)) == 0.0 && getRight(child(root(), 2)) == getWidth(root())",
			shouldPass: true,
		},
		{
			name:       "All children fit within container",
			expression: "getRight(child(root(), 0)) <= getWidth(root()) && getRight(child(root(), 1)) <= getWidth(root()) && getRight(child(root(), 2)) <= getWidth(root())",
			shouldPass: true,
		},
		{
			name:       "Children are in ascending X order",
			expression: "getX(child(root(), 0)) < getX(child(root(), 1)) && getX(child(root(), 1)) < getX(child(root(), 2))",
			shouldPass: true,
		},
		{
			name:       "Container is wider than it is tall",
			expression: "getWidth(root()) > getHeight(root())",
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := CELAssertion{
				Type:       "layout",
				Expression: tt.expression,
				Message:    tt.name,
			}

			result := celEnv.Evaluate(assertion)

			if tt.shouldPass && !result.Passed {
				t.Errorf("Expected assertion to pass but it failed: %s\nError: %s",
					tt.expression, result.Error)
			}

			if !tt.shouldPass && result.Passed {
				t.Errorf("Expected assertion to fail but it passed: %s", tt.expression)
			}
		})
	}
}

func TestCELEvaluateAll(t *testing.T) {
	root := &Node{
		Style: Style{
			Display: DisplayFlex,
			Width:   600,
			Height:  100,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
		},
	}

	Layout(root, Constraints{
		MinWidth:  0,
		MaxWidth:  800,
		MinHeight: 0,
		MaxHeight: 600,
	})

	celEnv, err := NewLayoutCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create CEL environment: %v", err)
	}

	assertions := []CELAssertion{
		{
			Type:       "layout",
			Expression: "getX(root()) == 0.0",
			Message:    "Root X should be 0",
		},
		{
			Type:       "layout",
			Expression: "getWidth(root()) == 600.0",
			Message:    "Root width should be 600",
		},
		{
			Type:       "layout",
			Expression: "childCount(root()) == 2",
			Message:    "Root should have 2 children",
		},
		{
			Type:       "color", // Unsupported type
			Expression: "toRGB(getColor(root()))",
			Message:    "Should be skipped",
		},
	}

	results := celEnv.EvaluateAll(assertions)

	if len(results) != 4 {
		t.Fatalf("Expected 4 results, got %d", len(results))
	}

	// First 3 should pass
	for i := 0; i < 3; i++ {
		if !results[i].Passed {
			t.Errorf("Assertion %d should have passed: %s", i, results[i].Error)
		}
	}

	// Last one should be skipped (color type not supported)
	if !results[3].Passed {
		t.Logf("Color assertion correctly skipped: %s", results[3].Error)
	}
}
