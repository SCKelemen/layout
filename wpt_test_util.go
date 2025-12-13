package layout

import (
	"fmt"
	"math"
	"testing"
)

// TestContext provides ergonomic WPT test utilities with tolerance support
type TestContext struct {
	t                *testing.T
	test             *WPTTest
	browserName      string
	tolerance        Tolerance
	customTolerances map[string]float64
}

// NewTestContext creates a test context for a specific browser
func NewTestContext(t *testing.T, test *WPTTest, browserName string) *TestContext {
	t.Helper()

	browser, ok := test.Results[browserName]
	if !ok {
		t.Fatalf("No results found for browser: %s", browserName)
	}

	return &TestContext{
		t:                t,
		test:             test,
		browserName:      browserName,
		tolerance:        browser.GetTolerance(),
		customTolerances: make(map[string]float64),
	}
}

// WithTolerance sets custom tolerance for a specific property
// Example: ctx.WithTolerance("x", 2.0).WithTolerance("width", 1.5)
func (ctx *TestContext) WithTolerance(property string, value float64) *TestContext {
	ctx.customTolerances[property] = value
	return ctx
}

// WithPositionTolerance sets tolerance for all position properties (x, y)
func (ctx *TestContext) WithPositionTolerance(value float64) *TestContext {
	ctx.customTolerances["x"] = value
	ctx.customTolerances["y"] = value
	return ctx
}

// WithSizeTolerance sets tolerance for all size properties (width, height)
func (ctx *TestContext) WithSizeTolerance(value float64) *TestContext {
	ctx.customTolerances["width"] = value
	ctx.customTolerances["height"] = value
	return ctx
}

// getTolerance returns the tolerance for a given property
func (ctx *TestContext) getTolerance(property string) float64 {
	// Check custom tolerance first
	if tol, ok := ctx.customTolerances[property]; ok {
		return tol
	}

	// Fall back to default tolerances
	switch property {
	case "x", "y":
		return ctx.tolerance.Position
	case "width", "height":
		return ctx.tolerance.Size
	default:
		return ctx.tolerance.Numeric
	}
}

// Element provides fluent assertions for an element
type Element struct {
	ctx      *TestContext
	path     string
	actual   interface{} // Your node type
	expected ElementResult
}

// For finds an element by path and returns an Element for assertions
// Example: ctx.For("root.children[0]", actualNode)
func (ctx *TestContext) For(path string, actual interface{}) *Element {
	ctx.t.Helper()

	// Find expected result for this path
	browser := ctx.test.Results[ctx.browserName]
	var expected *ElementResult
	for i := range browser.Elements {
		if browser.Elements[i].Path == path {
			expected = &browser.Elements[i]
			break
		}
	}

	if expected == nil {
		ctx.t.Fatalf("No expected result found for path: %s", path)
	}

	return &Element{
		ctx:      ctx,
		path:     path,
		actual:   actual,
		expected: *expected,
	}
}

// ExpectX validates the X position
func (el *Element) ExpectX(actual float64) *Element {
	el.ctx.t.Helper()

	expected, ok := el.expected.Expected["x"].(float64)
	if !ok {
		el.ctx.t.Errorf("%s: no expected X value", el.path)
		return el
	}

	tolerance := el.ctx.getTolerance("x")
	diff := math.Abs(actual - expected)

	if diff > tolerance {
		el.ctx.t.Errorf("%s: X mismatch: expected %.2f, got %.2f (diff=%.2f, tolerance=%.2f)",
			el.path, expected, actual, diff, tolerance)
	}

	return el
}

// ExpectY validates the Y position
func (el *Element) ExpectY(actual float64) *Element {
	el.ctx.t.Helper()

	expected, ok := el.expected.Expected["y"].(float64)
	if !ok {
		el.ctx.t.Errorf("%s: no expected Y value", el.path)
		return el
	}

	tolerance := el.ctx.getTolerance("y")
	diff := math.Abs(actual - expected)

	if diff > tolerance {
		el.ctx.t.Errorf("%s: Y mismatch: expected %.2f, got %.2f (diff=%.2f, tolerance=%.2f)",
			el.path, expected, actual, diff, tolerance)
	}

	return el
}

// ExpectWidth validates the width
func (el *Element) ExpectWidth(actual float64) *Element {
	el.ctx.t.Helper()

	expected, ok := el.expected.Expected["width"].(float64)
	if !ok {
		el.ctx.t.Errorf("%s: no expected width value", el.path)
		return el
	}

	tolerance := el.ctx.getTolerance("width")
	diff := math.Abs(actual - expected)

	if diff > tolerance {
		el.ctx.t.Errorf("%s: width mismatch: expected %.2f, got %.2f (diff=%.2f, tolerance=%.2f)",
			el.path, expected, actual, diff, tolerance)
	}

	return el
}

// ExpectHeight validates the height
func (el *Element) ExpectHeight(actual float64) *Element {
	el.ctx.t.Helper()

	expected, ok := el.expected.Expected["height"].(float64)
	if !ok {
		el.ctx.t.Errorf("%s: no expected height value", el.path)
		return el
	}

	tolerance := el.ctx.getTolerance("height")
	diff := math.Abs(actual - expected)

	if diff > tolerance {
		el.ctx.t.Errorf("%s: height mismatch: expected %.2f, got %.2f (diff=%.2f, tolerance=%.2f)",
			el.path, expected, actual, diff, tolerance)
	}

	return el
}

// ExpectPosition validates both X and Y
func (el *Element) ExpectPosition(actualX, actualY float64) *Element {
	el.ctx.t.Helper()
	return el.ExpectX(actualX).ExpectY(actualY)
}

// ExpectSize validates both width and height
func (el *Element) ExpectSize(actualWidth, actualHeight float64) *Element {
	el.ctx.t.Helper()
	return el.ExpectWidth(actualWidth).ExpectHeight(actualHeight)
}

// ExpectRect validates X, Y, width, and height
func (el *Element) ExpectRect(x, y, width, height float64) *Element {
	el.ctx.t.Helper()
	return el.ExpectPosition(x, y).ExpectSize(width, height)
}

// NodeGetter is an interface for types that provide position/size
// Implement this interface for your layout node type
type NodeGetter interface {
	GetX() float64
	GetY() float64
	GetWidth() float64
	GetHeight() float64
}

// ExpectNode validates all properties of a node that implements NodeGetter
func (el *Element) ExpectNode(node NodeGetter) *Element {
	el.ctx.t.Helper()
	return el.ExpectRect(node.GetX(), node.GetY(), node.GetWidth(), node.GetHeight())
}

// ExpectProperty validates a custom property with numeric tolerance
func (el *Element) ExpectProperty(name string, actual float64) *Element {
	el.ctx.t.Helper()

	expectedVal, ok := el.expected.Expected[name]
	if !ok {
		el.ctx.t.Errorf("%s: no expected value for property %s", el.path, name)
		return el
	}

	expected, ok := expectedVal.(float64)
	if !ok {
		el.ctx.t.Errorf("%s: expected value for %s is not a number", el.path, name)
		return el
	}

	tolerance := el.ctx.getTolerance(name)
	diff := math.Abs(actual - expected)

	if diff > tolerance {
		el.ctx.t.Errorf("%s: %s mismatch: expected %.2f, got %.2f (diff=%.2f, tolerance=%.2f)",
			el.path, name, expected, actual, diff, tolerance)
	}

	return el
}

// Helper functions for common assertions without creating Element

// ExpectX is a standalone helper for X position
func ExpectX(t *testing.T, path string, expected ElementResult, actual float64, tolerance float64) {
	t.Helper()
	expectedX, _ := expected.Expected["x"].(float64)
	diff := math.Abs(actual - expectedX)
	if diff > tolerance {
		t.Errorf("%s: X mismatch: expected %.2f, got %.2f (diff=%.2f, tolerance=%.2f)",
			path, expectedX, actual, diff, tolerance)
	}
}

// ExpectY is a standalone helper for Y position
func ExpectY(t *testing.T, path string, expected ElementResult, actual float64, tolerance float64) {
	t.Helper()
	expectedY, _ := expected.Expected["y"].(float64)
	diff := math.Abs(actual - expectedY)
	if diff > tolerance {
		t.Errorf("%s: Y mismatch: expected %.2f, got %.2f (diff=%.2f, tolerance=%.2f)",
			path, expectedY, actual, diff, tolerance)
	}
}

// ExpectWidth is a standalone helper for width
func ExpectWidth(t *testing.T, path string, expected ElementResult, actual float64, tolerance float64) {
	t.Helper()
	expectedWidth, _ := expected.Expected["width"].(float64)
	diff := math.Abs(actual - expectedWidth)
	if diff > tolerance {
		t.Errorf("%s: width mismatch: expected %.2f, got %.2f (diff=%.2f, tolerance=%.2f)",
			path, expectedWidth, actual, diff, tolerance)
	}
}

// ExpectHeight is a standalone helper for height
func ExpectHeight(t *testing.T, path string, expected ElementResult, actual float64, tolerance float64) {
	t.Helper()
	expectedHeight, _ := expected.Expected["height"].(float64)
	diff := math.Abs(actual - expectedHeight)
	if diff > tolerance {
		t.Errorf("%s: height mismatch: expected %.2f, got %.2f (diff=%.2f, tolerance=%.2f)",
			path, expectedHeight, actual, diff, tolerance)
	}
}

// Example usage documentation

/*
Example 1: Fluent API with test context

	func TestFlexboxLayout(t *testing.T) {
		test, _ := LoadWPTTest("tests/flexbox/justify-content-001.json")

		// Create test context for Chrome results
		ctx := NewTestContext(t, test, "chrome")

		// Optional: Customize tolerances
		ctx.WithPositionTolerance(2.0).WithSizeTolerance(1.5)

		// Build and run layout
		root := BuildLayout(&test.Layout)
		Layout(root, test.GetConstraints())

		// Validate root element (fluent API)
		ctx.For("root", root).
			ExpectX(root.Rect.X).
			ExpectY(root.Rect.Y).
			ExpectWidth(root.Rect.Width).
			ExpectHeight(root.Rect.Height)

		// Or use ExpectRect for all at once
		ctx.For("root.children[0]", root.Children[0]).
			ExpectRect(
				root.Children[0].Rect.X,
				root.Children[0].Rect.Y,
				root.Children[0].Rect.Width,
				root.Children[0].Rect.Height,
			)

		// Or if your node implements NodeGetter
		ctx.For("root.children[1]", root.Children[1]).
			ExpectNode(root.Children[1])
	}

Example 2: Per-property custom tolerance

	ctx := NewTestContext(t, test, "chrome")
	ctx.WithTolerance("x", 3.0)  // X can differ by up to 3px
	ctx.WithTolerance("width", 0.5)  // Width must be very precise

	ctx.For("root", root).
		ExpectX(root.Rect.X).  // Uses 3.0 tolerance
		ExpectWidth(root.Rect.Width)  // Uses 0.5 tolerance

Example 3: Standalone helpers (no context needed)

	chrome := test.Results["chrome"]
	tolerance := chrome.GetTolerance()

	for _, expected := range chrome.Elements {
		actual := FindNode(root, expected.Path)
		ExpectX(t, expected.Path, expected, actual.Rect.X, tolerance.Position)
		ExpectY(t, expected.Path, expected, actual.Rect.Y, tolerance.Position)
		ExpectWidth(t, expected.Path, expected, actual.Rect.Width, tolerance.Size)
		ExpectHeight(t, expected.Path, expected, actual.Rect.Height, tolerance.Size)
	}
*/
