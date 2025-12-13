package layout

import (
	"testing"
)

func TestPathContext(t *testing.T) {
	// Create a simple tree
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
			{
				Style: Style{Width: 100, Height: 50},
				Children: []*Node{
					{Style: Style{Width: 50, Height: 25}},
				},
			},
		},
	}

	ctx := NewPathContext(root)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Root exists", "root", true},
		{"Child 0 exists", "root.children[0]", true},
		{"Child 1 exists", "root.children[1]", true},
		{"Child 2 exists", "root.children[2]", true},
		{"Grandchild exists", "root.children[2].children[0]", true},
		{"Invalid path", "root.children[99]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := ctx.GetNode(tt.path)
			if tt.expected && node == nil {
				t.Errorf("Expected node at path %s, got nil", tt.path)
			}
			if !tt.expected && node != nil {
				t.Errorf("Expected nil at path %s, got node", tt.path)
			}
		})
	}
}

func TestGetParentPath(t *testing.T) {
	ctx := &PathContext{}

	tests := []struct {
		path     string
		expected string
	}{
		{"root", ""},
		{"root.children[0]", "root"},
		{"root.children[1]", "root"},
		{"root.children[0].children[0]", "root.children[0]"},
		{"root.children[2].children[5]", "root.children[2]"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := ctx.GetParentPath(tt.path)
			if result != tt.expected {
				t.Errorf("GetParentPath(%s) = %s, want %s", tt.path, result, tt.expected)
			}
		})
	}
}

func TestCELContextWithThisAndParent(t *testing.T) {
	// Create a layout tree
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentSpaceBetween,
			AlignItems:     AlignItemsCenter,
			Width:          600,
			Height:         100,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
		},
	}

	// Run layout
	Layout(root, Constraints{
		MinWidth:  0,
		MaxWidth:  800,
		MinHeight: 0,
		MaxHeight: 600,
	})

	// Create context-aware environment
	env, err := NewLayoutCELEnvWithContext(root)
	if err != nil {
		t.Fatalf("Failed to create CEL environment: %v", err)
	}

	// Test assertions with this() and parent()
	tests := []struct {
		name       string
		path       string
		expression string
		shouldPass bool
	}{
		{
			name:       "this() refers to current node",
			path:       "root.children[0]",
			expression: "getX(this()) == 0.0",
			shouldPass: true,
		},
		{
			name:       "parent() refers to parent node",
			path:       "root.children[0]",
			expression: "getWidth(parent()) == 600.0",
			shouldPass: true,
		},
		{
			name:       "Complex expression with this and parent",
			path:       "root.children[0]",
			expression: "getY(this()) == (getHeight(parent()) - getHeight(this())) / 2.0",
			shouldPass: true,
		},
		{
			name:       "Margin calculation",
			path:       "root.children[0]",
			expression: "getY(this()) == getMarginTop(parent()) + (getHeight(parent()) - getHeight(this())) / 2.0",
			shouldPass: true,
		},
		{
			name:       "Second child positioning",
			path:       "root.children[1]",
			expression: "getX(this()) == 250.0",
			shouldPass: true,
		},
		{
			name:       "Last child positioning",
			path:       "root.children[2]",
			expression: "getRight(this()) == getWidth(parent())",
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

			result := env.EvaluateAtPath(assertion, tt.path)

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

func TestCELContextRootAssertions(t *testing.T) {
	// Create a layout tree
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentSpaceBetween,
			AlignItems:     AlignItemsCenter,
			Width:          600,
			Height:         100,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
		},
	}

	// Run layout
	Layout(root, Constraints{
		MinWidth:  0,
		MaxWidth:  800,
		MinHeight: 0,
		MaxHeight: 600,
	})

	// Create context-aware environment
	env, err := NewLayoutCELEnvWithContext(root)
	if err != nil {
		t.Fatalf("Failed to create CEL environment: %v", err)
	}

	// Test assertions at root level (without this/parent)
	tests := []struct {
		name       string
		expression string
		shouldPass bool
	}{
		{
			name:       "Root dimensions",
			expression: "getWidth(root()) == 600.0",
			shouldPass: true,
		},
		{
			name:       "Child positioning",
			expression: "getX(child(root(), 0)) == 0.0",
			shouldPass: true,
		},
		{
			name:       "Space between",
			expression: "getX(child(root(), 1)) - getRight(child(root(), 0)) == getX(child(root(), 2)) - getRight(child(root(), 1))",
			shouldPass: true,
		},
		{
			name:       "Last child edge",
			expression: "getRight(child(root(), 2)) == getWidth(root())",
			shouldPass: true,
		},
		{
			name:       "Vertical centering",
			expression: "getY(child(root(), 0)) == (getHeight(root()) - getHeight(child(root(), 0))) / 2.0",
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

			result := env.Evaluate(assertion)

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

func TestCELContextNestedTree(t *testing.T) {
	// Create a nested tree
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
			Width:         600,
			Height:        400,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:       DisplayFlex,
					FlexDirection: FlexDirectionRow,
					Width:         600,
					Height:        200,
				},
				Children: []*Node{
					{Style: Style{Width: 100, Height: 100}},
					{Style: Style{Width: 100, Height: 100}},
				},
			},
			{
				Style: Style{
					Display:       DisplayFlex,
					FlexDirection: FlexDirectionRow,
					Width:         600,
					Height:        200,
				},
				Children: []*Node{
					{Style: Style{Width: 100, Height: 100}},
				},
			},
		},
	}

	// Run layout
	Layout(root, Constraints{
		MinWidth:  0,
		MaxWidth:  800,
		MinHeight: 0,
		MaxHeight: 600,
	})

	// Create context-aware environment
	env, err := NewLayoutCELEnvWithContext(root)
	if err != nil {
		t.Fatalf("Failed to create CEL environment: %v", err)
	}

	// Test nested assertions
	tests := []struct {
		name       string
		path       string
		expression string
		shouldPass bool
	}{
		{
			name:       "Grandchild parent is container",
			path:       "root.children[0].children[0]",
			expression: "getWidth(parent()) == 600.0",
			shouldPass: true,
		},
		{
			name:       "Grandchild grandparent is root",
			path:       "root.children[0].children[0]",
			expression: "getWidth(parent()) == getWidth(root())",
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

			result := env.EvaluateAtPath(assertion, tt.path)

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
