package examples

import (
	"strings"
	"testing"

	"github.com/SCKelemen/layout"
	"github.com/SCKelemen/wpt-test-gen/pkg/cel"
)

// Browser test for CSS Test: Flexbox with CEL assertions
func TestTestCelAssertions(t *testing.T) {
	// Test spec loaded from: /tmp/test-cel-assertions.json

	// This is a standalone test that uses the layout library directly
	root := buildLayoutTestTestCelAssertions()
	ctx := layout.NewLayoutContext(800, 600, 16)
	layout.Layout(root, layout.Constraints{
		MinWidth:  0,
		MaxWidth:  800,
		MinHeight: 0,
		MaxHeight: 600,
	}, ctx)

	// Create CEL environment
	env, err := cel.NewLayoutCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create CEL environment: %v", err)
	}

	// Evaluate all CEL assertions
	assertions := []struct {
		expr    string
		message string
	}{
		{
			expr:    "getX(child(root(), 1)) - getRight(child(root(), 0)) == getX(child(root(), 2)) - getRight(child(root(), 1))",
			message: "spacing",
		},
		{
			expr:    "getX(child(root(), 0)) == 0.0",
			message: "first-child",
		},
		{
			expr:    "getRight(child(root(), 2)) == getWidth(root())",
			message: "last-child",
		},
		{
			expr:    "getY(child(root(), 0)) == (getHeight(root()) - getHeight(child(root(), 0))) / 2.0",
			message: "vertical-center",
		},
		{
			expr:    "getY(this) == getMarginTop(parent()) + (getHeight(parent()) - getHeight(this)) / 2.0",
			message: "aligned",
		},
	}

	for _, assertionData := range assertions {
		assertion := cel.CELAssertion{
			Type:       "layout",
			Expression: assertionData.expr,
			Message:    assertionData.message,
		}

		result := env.Evaluate(assertion)

		if !result.Passed {
			// Be lenient with unsupported features like 'this' and 'parent()'
			if strings.Contains(result.Error, "undeclared reference to 'this'") ||
				strings.Contains(result.Error, "undeclared reference to 'parent'") {
				t.Logf("Skipping unsupported assertion: %s\nExpression: %s",
					assertion.Message, assertion.Expression)
			} else {
				t.Errorf("Assertion failed: %s\nExpression: %s\nError: %s",
					assertion.Message, assertion.Expression, result.Error)
			}
		}
	}
}

// buildLayoutTestTestCelAssertions constructs the layout tree for this test
func buildLayoutTestTestCelAssertions() *layout.Node {
	root := &layout.Node{
		Style: layout.Style{
			Display:        layout.DisplayFlex,
			JustifyContent: layout.JustifyContentSpaceBetween,
			AlignItems:     layout.AlignItemsCenter,
			Width:          layout.Px(600.0),
			Height:         layout.Px(100.0),
		},
		Children: []*layout.Node{
			{
				Style: layout.Style{
					Width:  layout.Px(100.0),
					Height: layout.Px(50.0),
				},
			},
			{
				Style: layout.Style{
					Width:  layout.Px(100.0),
					Height: layout.Px(50.0),
				},
			},
			{
				Style: layout.Style{
					Width:  layout.Px(100.0),
					Height: layout.Px(50.0),
				},
			},
		},
	}

	return root
}
