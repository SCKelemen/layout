# WPT Test Generator Tools

This directory contains tools for generating Go layout tests from Web Platform Tests (WPT) using headless Chrome as the source of truth.

## Overview

The WPT test generation process has two main components:

1. **wpt_renderer.js** - Renders WPT HTML tests in headless Chrome and extracts layout data
2. **main.go** - Generates Go test code from the extracted browser layout data

## Prerequisites

```bash
# Install Node.js dependencies
npm install
```

This installs Puppeteer, which provides a headless Chrome browser for rendering tests.

## Usage

### Single Test Conversion

```bash
# 1. Render test in Chrome and extract layout data
node wpt_renderer.js path/to/test.html output.json

# 2. Generate Go test from layout data
go run main.go output.json test_output.go

# 3. Run the generated test
go test -v test_output.go
```

### Batch Processing

```bash
# 1. Render multiple tests
node wpt_renderer.js --batch /tmp/wpt/css/css-flexbox /tmp/wpt_json

# 2. Generate single Go file with all tests
go run main.go --batch /tmp/wpt_json wpt_browser_tests.go

# 3. Run all generated tests
go test -v wpt_browser_tests.go
```

## How It Works

### wpt_renderer.js

This Node.js script:

1. Launches headless Chrome via Puppeteer
2. Loads the WPT HTML test file
3. Extracts layout information:
   - Element positions via `getBoundingClientRect()`
   - Computed styles via `getComputedStyle()`
   - Flex/Grid container properties
   - Child element positions
4. Outputs JSON with browser-measured values

**Element Selection Strategy:**

The renderer identifies elements using computed styles (not inline styles):
- All flex/grid containers (`display: flex`, `display: grid`, etc.)
- Elements with IDs
- Elements with `data-expected-*` attributes

This ensures we capture:
- Tests using CSS classes (most WPT tests)
- Tests using inline styles
- Tests with explicit expected values

**Output Format:**

```json
{
  "testFile": "Test Title",
  "viewport": {"width": 800, "height": 600},
  "elements": [
    {
      "selector": "#container",
      "tagName": "div",
      "dataExpected": {"width": 600, "height": 100},
      "rect": {"x": 0, "y": 0, "width": 600, "height": 100},
      "computed": {
        "display": "flex",
        "flexDirection": "row",
        "justifyContent": "space-between",
        "width": "600px",
        "height": "100px",
        "margin": {"top": "8px", ...},
        "padding": {"top": "0px", ...}
      },
      "children": [
        {"selector": "#child1", "rect": {...}},
        {"selector": "#child2", "rect": {...}}
      ]
    }
  ],
  "metadata": {
    "generatedAt": "2025-12-12T...",
    "browser": "Chrome Headless",
    "sourceFile": "test.html"
  }
}
```

### main.go

This Go program:

1. Parses the JSON layout data
2. Generates Go test functions
3. Creates assertions comparing our engine to browser behavior
4. Outputs formatted Go code

**Generated Test Structure:**

```go
func TestWPTBrowser_1(t *testing.T) {
    // Test: flexbox/align-content-wrap-002.html
    // Browser: Chrome Headless

    root := &Node{
        Style: Style{
            Display: DisplayFlex,
            Width:   600,
            Height:  100,
            // ... other properties from computed styles
        },
        Children: []*Node{
            // ... children based on test structure
        },
    }

    Layout(root, Loose(800, 600))

    // Browser expected values (from Chrome)
    if math.Abs(root.Rect.Width-600.0) > 1.0 {
        t.Errorf("Width: expected 600.0 (browser), got %f", root.Rect.Width)
    }
    if math.Abs(root.Rect.Height-100.0) > 1.0 {
        t.Errorf("Height: expected 100.0 (browser), got %f", root.Rect.Height)
    }
    // ... more assertions for children
}
```

## WPT Test Coverage

### Available Tests

- **CSS Flexbox**: ~1,600+ HTML tests
- **CSS Grid**: ~2,400+ HTML tests
- **CSS Sizing**: ~200+ tests
- **CSS Box Model**: ~300+ tests

### Test Types

WPT tests come in several formats:

1. **Reference Tests** (~90%): Visual comparison, no programmatic assertions
2. **Tests with data-expected** (~5%): Include `data-expected-width`, `data-expected-height` attributes
3. **JavaScript Tests** (~5%): Use testharness.js for assertions

Our tools work with **ALL** test types by using the browser as source of truth.

### Conversion Strategy

**All Tests:**
- Render in headless Chrome
- Extract actual browser layout
- Generate Go tests with browser values as expected results
- Tests verify our engine matches browser behavior

**Tests with data-expected attributes:**
- Also capture the data-expected values
- Can compare: our engine ↔ browser ↔ test expectations
- Useful for validating test assumptions

## Example Workflow

### Converting a Specific WPT Test

```bash
# 1. Clone WPT repository
git clone --depth 1 --filter=blob:none --sparse https://github.com/web-platform-tests/wpt.git
cd wpt
git sparse-checkout set css/css-flexbox css/css-grid

# 2. Pick a test to convert
TEST="css/css-flexbox/align-content-wrap-002.html"

# 3. Render in Chrome
node /path/to/wpt_renderer.js "$TEST" output.json

# 4. Generate Go test
go run /path/to/main.go output.json test.go

# 5. Verify it works
go test -v test.go
```

### Batch Converting Tests

```bash
# Sample 10 random flexbox tests
cd wpt/css/css-flexbox
ls *.html | grep -v "\-ref\.html" | shuf | head -10 > /tmp/sample_tests.txt

# Copy to temp directory
mkdir /tmp/wpt_sample
while read test; do
  cp "$test" /tmp/wpt_sample/
done < /tmp/sample_tests.txt

# Render all in Chrome
cd /path/to/tools/wpt-test-generator
node wpt_renderer.js --batch /tmp/wpt_sample /tmp/wpt_json

# Generate Go tests
go run main.go --batch /tmp/wpt_json /tmp/wpt_browser_tests.go

# Run and see results
cd /path/to/layout
cp /tmp/wpt_browser_tests.go .
go test -v wpt_browser_tests.go
```

## Limitations and Considerations

### Current Limitations

1. **External Resources**: Tests that require external CSS/JS files may not work correctly
2. **Fonts**: Font-dependent tests may vary based on system fonts
3. **Dynamic Tests**: Tests with animations or JavaScript interactions capture only initial state
4. **Browser Dependencies**: Some tests target specific browser behaviors

### Best Practices

1. **Focus on Simple Tests**: Start with tests that have minimal dependencies
2. **Verify Test Quality**: Some WPT tests test edge cases or browser bugs
3. **Use Reference Tests**: Tests with `-ref.html` companions are visual comparison tests
4. **Check for data-expected**: Tests with explicit expectations are easier to validate
5. **Sample Before Batch**: Try individual tests before batch processing

### Test Selection Criteria

Good candidates for conversion:
- ✅ Self-contained tests (no external dependencies)
- ✅ Tests with clear layout structure
- ✅ Tests focusing on flex/grid layout properties
- ✅ Tests with data-expected attributes

Less suitable:
- ❌ Tests requiring specific fonts
- ❌ Tests with external stylesheets
- ❌ Animation/transition tests
- ❌ Browser-specific workaround tests

## Troubleshooting

### "Puppeteer not found"

```bash
cd tools/wpt-test-generator
npm install
```

### "No elements found in test"

The test may:
- Use external stylesheets (not supported)
- Have no visible elements
- Use only inline blocks (add to selector)

Solution: Check the HTML structure and ensure elements have IDs or are flex/grid containers.

### "Generated test fails"

Possible causes:
1. Test expects browser-specific behavior we don't implement
2. Test has dependencies we didn't capture
3. Our implementation differs from browser

Steps:
1. Check the generated Go code
2. Compare expected vs actual values
3. Investigate if it's a real bug or test issue

## CI Integration

The `.github/workflows/wpt-sync.yml` workflow:
- Runs weekly to check for new WPT tests
- Samples and renders new tests in Chrome
- Generates browser compatibility reports
- Creates issues for review

See `docs/wpt-sync.md` for complete CI documentation.

## Contributing

To improve the test generator:

1. **Better Selectors**: Enhance element detection in wpt_renderer.js
2. **More Style Properties**: Add additional CSS properties to capture
3. **Test Quality**: Improve test generation logic in main.go
4. **Error Handling**: Add better error messages and recovery

## References

- [Web Platform Tests](https://github.com/web-platform-tests/wpt)
- [Puppeteer Documentation](https://pptr.dev/)
- [CSS Flexbox Spec](https://www.w3.org/TR/css-flexbox-1/)
- [CSS Grid Spec](https://www.w3.org/TR/css-grid-1/)
- [WPT Sync Documentation](../../../docs/wpt-sync.md)
