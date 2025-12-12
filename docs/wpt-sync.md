# Web Platform Tests (WPT) Synchronization

This document explains how we track and integrate Web Platform Tests into our test suite.

## Overview

[Web Platform Tests](https://github.com/web-platform-tests/wpt) (WPT) is the official W3C test suite for web platform features. We use WPT as a reference to ensure our CSS layout implementation stays aligned with web standards.

## Current Status

- **Our Test Coverage**: 321/321 tests passing (100%)
- **WPT-Derived Tests**: 14 text layout tests in `wpt_comprehensive_test.go`
- **Test Strategy**: Browser-based correctness testing using headless Chrome as source of truth

## WPT Test Conversion Strategy

### Browser-Based Test Generation

WPT tests are primarily **reference tests** (visual comparison) that don't include programmatic layout assertions. They compare rendered output against reference images.

Our approach uses **headless Chrome as the source of truth**:
1. **Render WPT HTML** in headless Chrome using Puppeteer
2. **Extract actual browser layout data** via `getBoundingClientRect()` and `getComputedStyle()`
3. **Save layout data to JSON** with browser-measured positions and sizes
4. **Generate Go tests** that verify our engine matches browser behavior
5. **Test for correctness**, not just regression

### Tools

#### `tools/wpt_renderer.js` (Primary Tool)

Node.js script using Puppeteer to render WPT tests in headless Chrome and extract browser layout data.

**Features:**
- Renders HTML in actual Chrome browser
- Extracts element positions via `getBoundingClientRect()`
- Captures computed styles via `getComputedStyle()`
- Outputs JSON with browser-measured values
- Supports single file or batch processing

**Requirements:**
```bash
cd tools
npm install  # Installs Puppeteer
```

**Usage:**
```bash
# Single test
node tools/wpt_renderer.js test.html test.json

# Batch processing
node tools/wpt_renderer.js --batch input-dir/ output-dir/
```

**Output JSON Format:**
```json
{
  "testFile": "Test Title",
  "viewport": {"width": 800, "height": 600},
  "elements": [
    {
      "selector": "#container",
      "tagName": "div",
      "rect": {"x": 0, "y": 0, "width": 800, "height": 200},
      "computed": {
        "display": "flex",
        "flexDirection": "row",
        "justifyContent": "center",
        ...
      },
      "children": [
        {"selector": "#child1", "rect": {...}},
        ...
      ]
    }
  ],
  "metadata": {
    "generatedAt": "2025-12-12T...",
    "browser": "Chrome Headless",
    "browserVersion": "Puppeteer Latest"
  }
}
```

#### `tools/wpt_test_generator.go` (Primary Tool)

Go program that reads JSON from wpt_renderer.js and generates Go test code.

**Features:**
- Parses browser layout JSON
- Generates test functions with browser expected values
- Creates assertions comparing our engine to browser behavior
- Supports batch generation from multiple JSON files
- Comments clearly indicate values come from actual browser

**Usage:**
```bash
# Single test
go run tools/wpt_test_generator.go test.json output_test.go

# Batch generation
go run tools/wpt_test_generator.go --batch json-dir/ wpt_browser_tests.go
```

**Generated Test Example:**
```go
func TestWPTBrowser_1(t *testing.T) {
    // WPT test: flexbox/align-content-001.htm
    // Browser expected values for #container

    root := &Node{
        Style: Style{
            Display: DisplayFlex,
            FlexDirection: FlexDirectionRow,
            AlignContent: AlignContentCenter,
        },
        Children: []*Node{
            {Style: Style{}},
            {Style: Style{}},
        },
    }

    constraints := Loose(800.00, 600.00)
    LayoutFlexbox(root, constraints)

    // Container dimensions (browser expected)
    if math.Abs(root.Rect.Width-800.00) > 1.0 {
        t.Errorf("Width: expected 800.00 (browser), got %f", root.Rect.Width)
    }
    // ... more assertions with browser values
}
```

#### `tools/wpt_converter.go` (Legacy)

Original converter for tests with `data-expected-*` attributes. Used for custom test format.

## Automated WPT Sync

### Weekly CI Workflow

`.github/workflows/wpt-sync.yml` runs every Sunday to:

1. **Clone WPT Repository**
   - Fetches latest CSS test files
   - Tracks flexbox, grid, sizing, box model, and alignment tests

2. **Detect Changes**
   - Compares with previous run
   - Identifies new tests
   - Documents removed/renamed tests

3. **Generate Report**
   - Lists new WPT tests
   - Suggests coverage improvements
   - Creates GitHub issue for review

4. **Run Our Tests**
   - Verifies 100% pass rate
   - Ensures no regressions

### What the Workflow Does NOT Do

- ❌ Automatically convert WPT tests to Go
- ❌ Add new tests to our suite
- ❌ Modify existing tests

### What the Workflow DOES Do

- ✅ Tracks WPT test additions/changes
- ✅ Creates informational reports
- ✅ Alerts us to potential coverage gaps
- ✅ Verifies our test suite health

## Manual Test Conversion Process

When new WPT tests are identified:

### 1. Fetch the WPT test
```bash
curl -o test.html "https://raw.githubusercontent.com/web-platform-tests/wpt/master/css/css-flexbox/[test-name].html"
```

### 2. Render in Chrome and extract layout data
```bash
node tools/wpt_renderer.js test.html test.json
```

This creates a JSON file with browser-measured layout values.

### 3. Generate Go test from browser data
```bash
go run tools/wpt_test_generator.go test.json wpt_browser_test.go
```

This creates a Go test file that verifies our engine matches Chrome's layout.

### 4. Run the test to compare
```bash
go test -v -run TestWPTBrowser
```

This will show how closely our engine matches the browser:
- **Pass**: Our engine matches Chrome within 1px tolerance ✅
- **Fail**: Differences identified - review implementation or test validity

### 5. Add passing tests to suite
```bash
# Only add tests that pass (or that reveal bugs we want to track)
git add wpt_browser_test.go
git commit -m "Add WPT browser test: [description]

Converted from WPT test css/css-flexbox/[test-name].html
Uses Chrome headless as source of truth for expected values"
```

### Batch Processing

For multiple tests:

```bash
# Download tests to a directory
mkdir wpt_tests
cd wpt_tests
# ... download multiple test files ...

# Render all tests in Chrome
node ../tools/wpt_renderer.js --batch . ../wpt_json_output/

# Generate single Go test file with all tests
go run tools/wpt_test_generator.go --batch wpt_json_output/ wpt_browser_tests.go

# Run all tests
go test -v -run TestWPTBrowser
```

## Test Coverage Philosophy

### Our Approach

- **100% of our tests pass** (321/321)
- **Comprehensive coverage** of CSS Grid, Flexbox, Box Model, Sizing, and Text
- **Spec-aligned** implementation
- **WPT-aware** (track changes, adapt as needed)

### Why Not All WPT Tests?

1. **Reference Tests**: Most WPT tests are visual, not programmatic
2. **Browser-Specific**: Some tests target browser quirks
3. **Complete Coverage**: Our 321 tests already cover all implemented features
4. **Maintenance**: Manual conversion is more maintainable than automatic

### When to Add WPT Tests

Add WPT-derived tests when:
- New CSS features are implemented
- WPT adds tests for edge cases we haven't covered
- Spec updates require new test scenarios
- Cross-browser compatibility issues are found

## Test Statistics

### Our Test Suite

| Category | Tests | Status |
|----------|-------|--------|
| **Total** | **321** | **100% passing** |
| Flexbox | ~120 | All passing |
| Grid | ~100 | All passing |
| Box Model | ~40 | All passing |
| Sizing | ~30 | All passing |
| Text | ~30 | All passing |

### WPT Repository (as of 2025-12)

| Category | Approximate Count |
|----------|------------------|
| css-flexbox | ~900 tests |
| css-grid | ~1000+ tests |
| css-sizing | ~200 tests |
| css-box | ~300 tests |

**Note**: WPT numbers include reference tests, parsing tests, and browser-specific tests.

## Completed Enhancements

### ✅ Browser-Based Test Generation (Implemented)

- **Headless Chrome rendering** via Puppeteer
- **Automatic layout data extraction** from browser
- **JSON intermediate format** for test data
- **Go test generation** from browser measurements
- **Batch processing** support
- **Correctness testing** against actual browser behavior

## Future Enhancements

### Potential Improvements

1. **Enhanced Style Parser**
   - Parse `<style>` tags and apply to elements
   - Match CSS selectors to DOM nodes
   - Compute cascaded styles for complex tests

2. **CI Integration**
   - Automated weekly WPT test rendering
   - Generate tests from new WPT additions
   - Create PRs with browser-validated tests
   - Automated comparison reports

3. **Multi-Browser Testing**
   - Firefox support via Puppeteer
   - Safari/WebKit testing
   - Cross-browser consistency checks
   - Browser-specific test variants

4. **Advanced Test Generation**
   - Generate tests from CSS spec examples
   - Property combination testing
   - Edge case enumeration
   - Fuzzing-based test discovery

## References

- [Web Platform Tests](https://github.com/web-platform-tests/wpt)
- [CSS Flexbox Spec](https://www.w3.org/TR/css-flexbox-1/)
- [CSS Grid Spec](https://www.w3.org/TR/css-grid-1/)
- [CSS Sizing Spec](https://www.w3.org/TR/css-sizing-3/)
- [Our Spec Compliance Status](../SPEC_COMPLIANCE_STATUS.md)
- [Specification Gaps](../SPECIFICATION_GAPS.md)

## Contributing

To add WPT-derived tests:

1. **Identify relevant WPT test** from WPT repository
2. **Render in Chrome** using `wpt_renderer.js` to extract browser layout
3. **Generate Go test** using `wpt_test_generator.go` from JSON output
4. **Verify test passes** (or documents known gap)
5. **Submit PR** with test description and WPT reference

### Example Workflow

```bash
# Fetch WPT test
curl -o align-content-001.html \
  "https://raw.githubusercontent.com/web-platform-tests/wpt/master/css/css-flexbox/align-content-001.htm"

# Extract browser layout
node tools/wpt_renderer.js align-content-001.html align-content-001.json

# Generate Go test
go run tools/wpt_test_generator.go align-content-001.json wpt_align_content_test.go

# Test it
go test -v -run TestWPTBrowser_1
```

### Example Commit Message

```
Add WPT browser test: flexbox align-content center

Converted from WPT test css/css-flexbox/align-content-001.htm
Uses Chrome headless rendering as source of truth

Tests multi-line flex container with align-content: center.
Expected values extracted from actual browser layout (Chrome).
Test passes with <1px tolerance.
```

---

**Last Updated**: 2025-12-12
**Test Coverage**: 321/321 passing (100%)
**WPT Sync**: Automated weekly via GitHub Actions
