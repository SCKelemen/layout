package layout

import (
	"testing"

	"github.com/google/cel-go/common/types"
)

// TestDomainCELBasic tests basic property access with domain API
func TestDomainCELBasic(t *testing.T) {
	// Create a simple layout
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

	// Create domain CEL environment
	env, rootRef, err := DomainCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create domain CEL environment: %v", err)
	}

	// Test cases with domain API
	tests := []struct {
		name       string
		expression string
		expected   interface{}
	}{
		// Basic property access
		{"root width", "width(root)", 600.0},
		{"root height", "height(root)", 100.0},
		{"root x", "x(root)", 0.0},
		{"root y", "y(root)", 0.0},

		// Position helpers
		{"root top", "top(root)", 0.0},
		{"root left", "left(root)", 0.0},
		{"root bottom", "bottom(root)", 100.0},
		{"root right", "right(root)", 600.0},

		// Child access
		{"first child x", "x(child(root, 0))", 0.0},
		{"second child x", "x(child(root, 1))", 250.0},
		{"third child x", "x(child(root, 2))", 500.0},

		// Child count
		{"child count", "childCount(root)", int64(3)},

		// Flexbox properties
		{"flex direction", "flexDirection(root)", "row"},
		{"justify content", "justifyContent(root)", "space-between"},
		{"align items", "alignItems(root)", "center"},

		// Vertical centering
		{"child centered", "y(child(root, 0)) == (height(root) - height(child(root, 0))) / 2.0", true},

		// Spacing between children
		{"equal spacing", "x(child(root, 1)) - right(child(root, 0)) == x(child(root, 2)) - right(child(root, 1))", true},

		// First at left edge
		{"first at left", "x(child(root, 0)) == 0.0", true},

		// Last at right edge
		{"last at right", "right(child(root, 2)) == width(root)", true},

		// Using firstChild and lastChild
		{"firstChild x", "x(firstChild(root))", 0.0},
		{"lastChild x", "x(lastChild(root))", 500.0},

		// Exact equality (default)
		{"equal exact", "equal(250.0, 250.0)", true},
		{"equal exact fail", "equal(250.0, 250.1)", false},

		// Tolerance-based equality
		{"equal absolute", "equal(250.0, 250.1, absolute(1.0))", true},
		{"equal absolute fail", "equal(250.0, 251.5, absolute(1.0))", false},

		// Relative tolerance (percentage)
		{"equal relative", "equal(100.0, 101.0, relative(2.0))", true},

		// Between helper
		{"between", "between(250.0, 200.0, 300.0)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compile the expression
			ast, issues := env.Compile(tt.expression)
			if issues != nil && issues.Err() != nil {
				t.Fatalf("Compilation error: %v", issues.Err())
			}

			// Create program
			prg, err := env.Program(ast)
			if err != nil {
				t.Fatalf("Program creation error: %v", err)
			}

			// Evaluate
			result, _, err := prg.Eval(map[string]interface{}{
				"root": rootRef,
			})
			if err != nil {
				t.Fatalf("Evaluation error: %v", err)
			}

			// Check result
			actual := result.Value()
			if actual != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, actual)
			}
		})
	}
}

// TestDomainCELNavigation tests tree navigation with domain API
func TestDomainCELNavigation(t *testing.T) {
	// Create a deeper tree
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
			Width:         400,
			Height:        400,
		},
		Children: []*Node{
			{
				Style: Style{Width: 400, Height: 100},
				Children: []*Node{
					{Style: Style{Width: 100, Height: 100}},
					{Style: Style{Width: 100, Height: 100}},
				},
			},
			{
				Style: Style{Width: 400, Height: 100},
				Children: []*Node{
					{Style: Style{Width: 100, Height: 100}},
					{Style: Style{Width: 100, Height: 100}},
				},
			},
			{Style: Style{Width: 400, Height: 100}},
		},
	}

	// Run layout
	Layout(root, Constraints{
		MinWidth:  0,
		MaxWidth:  800,
		MinHeight: 0,
		MaxHeight: 600,
	})

	// Create domain CEL environment
	env, rootRef, err := DomainCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create domain CEL environment: %v", err)
	}

	// Test navigation
	tests := []struct {
		name       string
		expression string
		expected   interface{}
	}{
		// Child navigation
		{"first level child", "x(child(root, 0))", 0.0},
		{"second level child", "x(child(child(root, 0), 0))", 0.0},

		// firstChild/lastChild
		{"firstChild", "x(firstChild(root))", 0.0},
		{"lastChild", "x(lastChild(root))", 0.0},

		// Sibling navigation
		{"nextSibling", "y(nextSibling(child(root, 0))) > y(child(root, 0))", true},
		{"previousSibling null", "previousSibling(child(root, 0)) == null", true},

		// Children count at different levels
		{"root children", "childCount(root)", int64(3)},
		{"first child children", "childCount(child(root, 0))", int64(2)},
		{"last child children", "childCount(child(root, 2))", int64(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, issues := env.Compile(tt.expression)
			if issues != nil && issues.Err() != nil {
				t.Fatalf("Compilation error: %v", issues.Err())
			}

			prg, err := env.Program(ast)
			if err != nil {
				t.Fatalf("Program creation error: %v", err)
			}

			result, _, err := prg.Eval(map[string]interface{}{
				"root": rootRef,
			})
			if err != nil {
				t.Fatalf("Evaluation error: %v", err)
			}

			actual := result.Value()
			if actual != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, actual)
			}
		})
	}
}

// TestDomainCELMarginPadding tests margin and padding accessors
func TestDomainCELMarginPadding(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:  200,
			Height: 200,
			Margin: Spacing{
				Top:    10,
				Right:  20,
				Bottom: 30,
				Left:   40,
			},
			Padding: Spacing{
				Top:    5,
				Right:  10,
				Bottom: 15,
				Left:   20,
			},
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 100}},
		},
	}

	Layout(root, Constraints{
		MinWidth:  0,
		MaxWidth:  800,
		MinHeight: 0,
		MaxHeight: 600,
	})

	env, rootRef, err := DomainCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create domain CEL environment: %v", err)
	}

	tests := []struct {
		name       string
		expression string
		expected   float64
	}{
		{"margin top", "marginTop(root)", 10.0},
		{"margin right", "marginRight(root)", 20.0},
		{"margin bottom", "marginBottom(root)", 30.0},
		{"margin left", "marginLeft(root)", 40.0},
		{"padding top", "paddingTop(root)", 5.0},
		{"padding right", "paddingRight(root)", 10.0},
		{"padding bottom", "paddingBottom(root)", 15.0},
		{"padding left", "paddingLeft(root)", 20.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, issues := env.Compile(tt.expression)
			if issues != nil && issues.Err() != nil {
				t.Fatalf("Compilation error: %v", issues.Err())
			}

			prg, err := env.Program(ast)
			if err != nil {
				t.Fatalf("Program creation error: %v", err)
			}

			result, _, err := prg.Eval(map[string]interface{}{
				"root": rootRef,
			})
			if err != nil {
				t.Fatalf("Evaluation error: %v", err)
			}

			if result.Value().(float64) != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result.Value())
			}
		})
	}
}

// TestDomainCELComplexExpressions tests complex CEL expressions with domain API
func TestDomainCELComplexExpressions(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			FlexWrap:       FlexWrapWrap,
			JustifyContent: JustifyContentCenter,
			AlignContent:   AlignContentSpaceAround,
			Width:          400,
			Height:         400,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 100}},
			{Style: Style{Width: 100, Height: 100}},
			{Style: Style{Width: 100, Height: 100}},
		},
	}

	Layout(root, Constraints{
		MinWidth:  0,
		MaxWidth:  800,
		MinHeight: 0,
		MaxHeight: 600,
	})

	env, rootRef, err := DomainCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create domain CEL environment: %v", err)
	}

	tests := []struct {
		name       string
		expression string
		wantBool   bool
		wantError  bool
	}{
		{
			name:       "All children same width",
			expression: "width(child(root, 0)) == width(child(root, 1)) && width(child(root, 1)) == width(child(root, 2))",
			wantBool:   true,
		},
		{
			name:       "All children same height",
			expression: "height(child(root, 0)) == height(child(root, 1)) && height(child(root, 1)) == height(child(root, 2))",
			wantBool:   true,
		},
		{
			name:       "Children within container",
			expression: "right(child(root, 0)) <= width(root) && right(child(root, 1)) <= width(root) && right(child(root, 2)) <= width(root)",
			wantBool:   true,
		},
		{
			name:       "Flexbox properties combined",
			expression: "flexDirection(root) == 'row' && flexWrap(root) == 'wrap' && justifyContent(root) == 'center'",
			wantBool:   true,
		},
		{
			name:       "Between range check",
			expression: "between(x(child(root, 1)), 0.0, 400.0) && between(y(child(root, 1)), 0.0, 400.0)",
			wantBool:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, issues := env.Compile(tt.expression)
			if issues != nil && issues.Err() != nil {
				if !tt.wantError {
					t.Fatalf("Compilation error: %v", issues.Err())
				}
				return
			}

			prg, err := env.Program(ast)
			if err != nil {
				if !tt.wantError {
					t.Fatalf("Program creation error: %v", err)
				}
				return
			}

			result, _, err := prg.Eval(map[string]interface{}{
				"root": rootRef,
			})
			if err != nil {
				if !tt.wantError {
					t.Fatalf("Evaluation error: %v", err)
				}
				return
			}

			if tt.wantError {
				t.Error("Expected error but got none")
				return
			}

			if result.Type() != types.BoolType {
				t.Fatalf("Expected bool result, got %v", result.Type())
			}

			if result.Value().(bool) != tt.wantBool {
				t.Errorf("Expected %v, got %v", tt.wantBool, result.Value())
			}
		})
	}
}

// TestDomainElementRefMethods tests ElementRef methods directly
func TestDomainElementRefMethods(t *testing.T) {
	root := &Node{
		Style: Style{Width: 300, Height: 300},
		Children: []*Node{
			{
				Style: Style{Width: 100, Height: 100},
				Children: []*Node{
					{Style: Style{Width: 50, Height: 50}},
				},
			},
			{Style: Style{Width: 100, Height: 100}},
			{Style: Style{Width: 100, Height: 100}},
		},
	}

	Layout(root, Constraints{
		MinWidth:  0,
		MaxWidth:  800,
		MinHeight: 0,
		MaxHeight: 600,
	})

	_, rootRef, err := DomainCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create domain CEL environment: %v", err)
	}

	// Test Children()
	children := rootRef.Children()
	if len(children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(children))
	}

	// Test Child()
	firstChild := rootRef.Child(0)
	if firstChild == nil {
		t.Fatal("Expected first child, got nil")
	}
	if firstChild.Width() != 100 {
		t.Errorf("Expected first child width 100, got %f", firstChild.Width())
	}

	// Test FirstChild() and LastChild()
	if rootRef.FirstChild() == nil {
		t.Error("Expected firstChild, got nil")
	}
	if rootRef.LastChild() == nil {
		t.Error("Expected lastChild, got nil")
	}

	// Test Parent()
	parent := firstChild.Parent()
	if parent == nil {
		t.Fatal("Expected parent, got nil")
	}
	if parent.Path() != "root" {
		t.Errorf("Expected parent path 'root', got '%s'", parent.Path())
	}

	// Test NextSibling()
	secondChild := firstChild.NextSibling()
	if secondChild == nil {
		t.Fatal("Expected next sibling, got nil")
	}
	if secondChild.Path() != "root.children[1]" {
		t.Errorf("Expected path 'root.children[1]', got '%s'", secondChild.Path())
	}

	// Test PreviousSibling()
	prevSibling := secondChild.PreviousSibling()
	if prevSibling == nil {
		t.Fatal("Expected previous sibling, got nil")
	}
	if prevSibling.Path() != "root.children[0]" {
		t.Errorf("Expected path 'root.children[0]', got '%s'", prevSibling.Path())
	}

	// Test Descendants()
	descendants := rootRef.Descendants()
	if len(descendants) != 4 { // 3 children + 1 grandchild
		t.Errorf("Expected 4 descendants, got %d", len(descendants))
	}

	// Test Ancestors()
	grandchild := firstChild.Child(0)
	ancestors := grandchild.Ancestors()
	if len(ancestors) != 2 { // parent and root
		t.Errorf("Expected 2 ancestors, got %d", len(ancestors))
	}

	// Test Find()
	found := rootRef.Find(func(e *ElementRef) bool {
		return e.Width() == 50
	})
	if found == nil {
		t.Error("Expected to find element with width 50")
	}

	// Test FindAll()
	foundAll := rootRef.FindAll(func(e *ElementRef) bool {
		return e.Width() == 100
	})
	if len(foundAll) != 3 { // 3 children with width 100
		t.Errorf("Expected 3 elements with width 100, got %d", len(foundAll))
	}

	// Test IsRoot()
	if !rootRef.IsRoot() {
		t.Error("Expected root to be root")
	}
	if firstChild.IsRoot() {
		t.Error("Expected first child not to be root")
	}

	// Test ChildCount()
	if rootRef.ChildCount() != 3 {
		t.Errorf("Expected 3 children, got %d", rootRef.ChildCount())
	}
}

// TestAssertHelpers tests the Assert type methods
func TestAssertHelpers(t *testing.T) {
	assert := &Assert{}

	// Test Equal2 with exact tolerance (default)
	if !assert.Equal2(1.0, 1.0, nil) {
		t.Error("Expected Equal2(1.0, 1.0, nil) to be true")
	}
	if assert.Equal2(1.0, 1.01, nil) {
		t.Error("Expected Equal2(1.0, 1.01, nil) to be false")
	}

	// Test Equal2 with absolute tolerance
	absTol := AbsoluteTolerance(0.1)
	if !assert.Equal2(1.0, 1.01, absTol) {
		t.Error("Expected Equal2(1.0, 1.01, absolute(0.1)) to be true")
	}
	if assert.Equal2(1.0, 2.0, absTol) {
		t.Error("Expected Equal2(1.0, 2.0, absolute(0.1)) to be false")
	}

	// Test Equal2 with relative tolerance
	relTol := RelativeTolerance(5.0) // 5%
	if !assert.Equal2(100.0, 104.0, relTol) {
		t.Error("Expected Equal2(100.0, 104.0, relative(5.0)) to be true")
	}
	if assert.Equal2(100.0, 110.0, relTol) {
		t.Error("Expected Equal2(100.0, 110.0, relative(5.0)) to be false")
	}

	// Test Between
	if !assert.Between(5.0, 0.0, 10.0) {
		t.Error("Expected Between(5.0, 0.0, 10.0) to be true")
	}
	if assert.Between(15.0, 0.0, 10.0) {
		t.Error("Expected Between(15.0, 0.0, 10.0) to be false")
	}

	// Test AllEqual with exact tolerance
	if !assert.AllEqual([]float64{5.0, 5.0, 5.0}, nil) {
		t.Error("Expected AllEqual([5.0, 5.0, 5.0], nil) to be true")
	}
	if assert.AllEqual([]float64{5.0, 5.0, 6.0}, nil) {
		t.Error("Expected AllEqual([5.0, 5.0, 6.0], nil) to be false")
	}

	// Test AllEqual with absolute tolerance
	if !assert.AllEqual([]float64{5.0, 5.01, 4.99}, AbsoluteTolerance(0.1)) {
		t.Error("Expected AllEqual([5.0, 5.01, 4.99], absolute(0.1)) to be true")
	}
	if assert.AllEqual([]float64{5.0, 6.0, 7.0}, AbsoluteTolerance(0.1)) {
		t.Error("Expected AllEqual([5.0, 6.0, 7.0], absolute(0.1)) to be false")
	}

	// Test Ascending
	if !assert.Ascending([]float64{1.0, 2.0, 3.0}) {
		t.Error("Expected Ascending([1.0, 2.0, 3.0]) to be true")
	}
	if assert.Ascending([]float64{1.0, 3.0, 2.0}) {
		t.Error("Expected Ascending([1.0, 3.0, 2.0]) to be false")
	}

	// Test Descending
	if !assert.Descending([]float64{3.0, 2.0, 1.0}) {
		t.Error("Expected Descending([3.0, 2.0, 1.0]) to be true")
	}
	if assert.Descending([]float64{1.0, 2.0, 3.0}) {
		t.Error("Expected Descending([1.0, 2.0, 3.0]) to be false")
	}
}

// TestToleranceTypes tests the different tolerance modes
func TestToleranceTypes(t *testing.T) {
	// Test Exact tolerance
	exact := ExactTolerance()
	if !exact.Matches(1.0, 1.0) {
		t.Error("Exact tolerance should match identical values")
	}
	if exact.Matches(1.0, 1.0000001) {
		t.Error("Exact tolerance should not match different values")
	}

	// Test Absolute tolerance
	abs := AbsoluteTolerance(0.1)
	if !abs.Matches(1.0, 1.05) {
		t.Error("Absolute tolerance should match within range")
	}
	if abs.Matches(1.0, 1.2) {
		t.Error("Absolute tolerance should not match outside range")
	}

	// Test Relative tolerance
	rel := RelativeTolerance(5.0) // 5%
	if !rel.Matches(100.0, 104.0) {
		t.Error("Relative tolerance should match within 5%")
	}
	if rel.Matches(100.0, 110.0) {
		t.Error("Relative tolerance should not match outside 5%")
	}

	// Test ULP tolerance
	ulp := ULPTolerance(10)
	if !ulp.Matches(1.0, 1.0) {
		t.Error("ULP tolerance should match identical values")
	}
}
