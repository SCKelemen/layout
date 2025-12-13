package layout

import (
	"math"
	"path/filepath"
	"testing"
)

// TestWPTSchemaLoader tests loading and running a WPT schema test
func TestWPTSchemaLoader(t *testing.T) {
	// Load the test
	testFile := filepath.Join("tools", "wpt-test-generator", "test-flex-row-schema.json")
	test, err := LoadWPTTest(testFile)
	if err != nil {
		t.Fatalf("Failed to load test: %v", err)
	}

	// Verify metadata
	if test.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", test.Version)
	}
	if test.ID != "test-flex-row" {
		t.Errorf("Expected ID test-flex-row, got %s", test.ID)
	}

	// Verify categories
	expectedCategories := map[string]bool{"flexbox": true, "layout": true}
	for _, cat := range test.Categories {
		if !expectedCategories[cat] {
			t.Errorf("Unexpected category: %s", cat)
		}
	}

	// Build layout from declarative structure
	root, err := test.BuildLayout()
	if err != nil {
		t.Fatalf("Failed to build layout: %v", err)
	}

	// Apply constraints and run layout
	constraints := test.GetConstraints()
	LayoutFlexbox(root, constraints)

	// Get Chrome results (primary browser)
	chromeResult, ok := test.Results["chrome"]
	if !ok {
		t.Fatal("No Chrome results found")
	}

	tolerance := chromeResult.GetTolerance()

	// Validate results
	for _, expectedElement := range chromeResult.Elements {
		var actualNode *Node

		// Find the node by path
		switch expectedElement.Path {
		case "root":
			actualNode = root
		case "root.children[0]":
			if len(root.Children) > 0 {
				actualNode = root.Children[0]
			}
		case "root.children[1]":
			if len(root.Children) > 1 {
				actualNode = root.Children[1]
			}
		case "root.children[2]":
			if len(root.Children) > 2 {
				actualNode = root.Children[2]
			}
		default:
			t.Errorf("Unknown element path: %s", expectedElement.Path)
			continue
		}

		if actualNode == nil {
			t.Errorf("Node not found for path: %s", expectedElement.Path)
			continue
		}

		// Check position and size with tolerance
		if x, ok := expectedElement.Expected["x"].(float64); ok {
			if math.Abs(actualNode.Rect.X-x) > tolerance.Position {
				t.Errorf("%s: X position mismatch: expected %.2f, got %.2f",
					expectedElement.Path, x, actualNode.Rect.X)
			}
		}

		if y, ok := expectedElement.Expected["y"].(float64); ok {
			if math.Abs(actualNode.Rect.Y-y) > tolerance.Position {
				t.Errorf("%s: Y position mismatch: expected %.2f, got %.2f",
					expectedElement.Path, y, actualNode.Rect.Y)
			}
		}

		if width, ok := expectedElement.Expected["width"].(float64); ok {
			if math.Abs(actualNode.Rect.Width-width) > tolerance.Size {
				t.Errorf("%s: Width mismatch: expected %.2f, got %.2f",
					expectedElement.Path, width, actualNode.Rect.Width)
			}
		}

		if height, ok := expectedElement.Expected["height"].(float64); ok {
			if math.Abs(actualNode.Rect.Height-height) > tolerance.Size {
				t.Errorf("%s: Height mismatch: expected %.2f, got %.2f",
					expectedElement.Path, height, actualNode.Rect.Height)
			}
		}
	}
}

// TestWPTSchemaMultiBrowser demonstrates multi-browser result comparison
func TestWPTSchemaMultiBrowser(t *testing.T) {
	testFile := filepath.Join("tools", "wpt-test-generator", "test-flex-row-schema.json")
	test, err := LoadWPTTest(testFile)
	if err != nil {
		t.Fatalf("Failed to load test: %v", err)
	}

	// Check which browsers have results
	browsers := []string{}
	for browser := range test.Results {
		browsers = append(browsers, browser)
	}

	if len(browsers) == 0 {
		t.Fatal("No browser results found")
	}

	t.Logf("Test has results from %d browser(s): %v", len(browsers), browsers)

	// Compare results across browsers (when we have multiple)
	if len(browsers) > 1 {
		// Get results from first two browsers
		browser1 := browsers[0]
		browser2 := browsers[1]

		result1 := test.Results[browser1]
		result2 := test.Results[browser2]

		// Compare element counts
		if len(result1.Elements) != len(result2.Elements) {
			t.Errorf("Browser result mismatch: %s has %d elements, %s has %d elements",
				browser1, len(result1.Elements), browser2, len(result2.Elements))
		}

		// Compare positions (should be very close across browsers)
		for i := 0; i < len(result1.Elements) && i < len(result2.Elements); i++ {
			el1 := result1.Elements[i]
			el2 := result2.Elements[i]

			if el1.Path != el2.Path {
				continue // Different elements
			}

			// Check if positions are within reasonable tolerance (2px)
			crossBrowserTolerance := 2.0

			x1, _ := el1.Expected["x"].(float64)
			x2, _ := el2.Expected["x"].(float64)
			if math.Abs(x1-x2) > crossBrowserTolerance {
				t.Logf("Cross-browser X difference for %s: %s=%.2f, %s=%.2f (diff=%.2f)",
					el1.Path, browser1, x1, browser2, x2, math.Abs(x1-x2))
			}

			y1, _ := el1.Expected["y"].(float64)
			y2, _ := el2.Expected["y"].(float64)
			if math.Abs(y1-y2) > crossBrowserTolerance {
				t.Logf("Cross-browser Y difference for %s: %s=%.2f, %s=%.2f (diff=%.2f)",
					el1.Path, browser1, y1, browser2, y2, math.Abs(y1-y2))
			}
		}
	}
}
