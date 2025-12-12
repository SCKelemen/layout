# Test Organization

This document describes the test organization structure for the layout engine.

## Test Categories

### 1. Blocking Tests (CI Must Pass)

These test suites validate fully implemented CSS specifications. CI builds will fail if any of these tests fail.

#### Implemented Features:
- **Flexbox** (`flexbox_*.go`)
  - Basic flex layout (direction, wrap)
  - Justify-content, align-items, align-content
  - Flex grow/shrink
  - Gaps (row-gap, column-gap, gap)
  - Wrap-reverse
  - Padding and margins
  - Status: **95%+ spec coverage**
  - WPT: 1,621 test files available

- **Grid Layout** (`grid_*.go`)
  - Basic grid structure
  - Grid template rows/columns
  - Grid auto-flow
  - Grid gaps
  - Alignment (align-content, justify-content)
  - Grid spanning
  - Status: **Core features implemented**
  - WPT: 2,386 test files available

- **Text Layout** (`text_*.go`)
  - Basic text wrapping (UAX #14 line breaking)
  - White-space modes (normal, nowrap, pre, pre-wrap, pre-line)
  - Text alignment (left, right, center, justify)
  - Text overflow (clip, ellipsis)
  - Text-indent
  - Word breaking
  - CJK text support
  - Status: **Core features implemented**
  - WPT: 2,667 test files available

- **Position Layout** (`positioned_*.go`)
  - Absolute positioning
  - Relative positioning
  - Fixed positioning
  - Sticky positioning
  - Status: **Core features implemented**
  - WPT: 417 test files available

- **Block Layout** (`block_*.go`)
  - Basic block layout
  - Margin collapsing
  - Width/height calculations
  - Min/max constraints
  - Status: **Basic implementation**

- **Box Model** (`box_*.go`)
  - Padding, borders, margins
  - Box-sizing
  - Status: **Complete for implemented layouts**
  - WPT: 156 test files available

### 2. Non-Blocking Tests (CI Runs, Warnings Only)

These test suites validate partially implemented or future features. CI will run these tests and report results, but won't block builds.

#### Partial/Future Features:
- **Subgrid** (CSS Grid Level 2, not yet implemented)
  - Nested grids with subgrid keyword
  - Status: **Not started**
  - WPT: Subset of 2,386 grid test files

- **Advanced Text Layout** (partial)
  - Text-align-last (planned)
  - Text-justify inter-character (planned)
  - Text-decoration (partial)
  - Text-transform (not started)
  - Status: **Partially planned**
  - WPT: Subset of 2,667 text test files

- **Advanced Sizing** (partial)
  - Intrinsic sizing (min-content, max-content)
  - Aspect-ratio
  - Container queries
  - Status: **Partial**
  - WPT: 778 test files available

- **CSS Containment** (not yet implemented)
  - Size containment
  - Layout containment
  - Paint containment
  - Status: **Not started**
  - WPT: 805 test files available

- **CSS Display Level 3** (partial)
  - Display: flow-root
  - Display: contents
  - Display: inline-flex, inline-grid
  - Status: **Partial**
  - WPT: 152 test files available

## Test File Naming Convention

```
<feature>_<subfeature>_test.go
```

Examples:
- `flexbox_test.go` - Core flexbox tests
- `flexbox_gap_test.go` - Flexbox gap property tests
- `flexbox_wrap_reverse_test.go` - Wrap-reverse tests
- `text_test.go` - Core text layout tests
- `text_overflow_test.go` - Text overflow tests
- `text_whitespace_test.go` - White-space mode tests
- `block_test.go` - Block layout tests
- `grid_test.go` - Grid layout tests (future)
- `position_test.go` - Position layout tests (future)

## Test Groups

Tests can be run by group using build tags or test name patterns:

### Run all blocking tests:
```bash
go test -v
```

### Run specific feature:
```bash
go test -v -run Flexbox
go test -v -run Text
go test -v -run Block
```

### Run specific subfeature:
```bash
go test -v -run FlexboxGap
go test -v -run TextOverflow
go test -v -run FlexboxWrapReverse
```

## WPT (Web Platform Tests) Integration

### Current WPT Test Counts:
- css-text: 2,667 HTML tests
- css-grid: 2,386 HTML tests
- css-flexbox: 1,621 HTML tests
- css-sizing: 778 HTML tests
- css-contain: 805 HTML tests
- css-position: 417 HTML tests
- css-align: 330 HTML tests
- css-box: 156 HTML tests
- css-display: 152 HTML tests

**Total: 9,312+ HTML test files**

### WPT Test Format

Real WPT tests use:
- `<style>` blocks with CSS classes
- Reference files for visual comparison
- JavaScript test harness (testharness.js, check-layout-th.js)
- Complex nested structures
- data-offset-x, data-offset-y attributes for position assertions

### Our Converter Limitations

Our current HTML-to-Go converter supports:
- Inline `style` attributes only
- Simple data attributes (data-expected-width, data-expected-height)
- Direct text content
- Flat structure

### Strategy for WPT Integration

Since direct conversion is not feasible, we use a hybrid approach:

1. **Manual Test Selection**: Identify high-value WPT tests that validate critical behavior
2. **Test Adaptation**: Manually adapt WPT test cases to our Go test format
3. **Incremental Coverage**: Gradually increase coverage of WPT test scenarios
4. **Reference Documentation**: Link each test to corresponding WPT test for traceability

Example:
```go
// TestFlexboxAlignContent validates align-content behavior
// Based on WPT: css/css-flexbox/align-content-wrap-001.html
func TestFlexboxAlignContent(t *testing.T) {
    // Test implementation
}
```

## CI Configuration

### GitHub Actions Workflow

See `.github/workflows/test.yml` for CI configuration.

### Test Stages:

1. **Fast Tests** (~1s)
   - Basic functionality
   - Quick smoke tests

2. **Core Tests** (~5s)
   - All blocking tests
   - Must pass for merge

3. **Extended Tests** (~10s)
   - Non-blocking tests
   - Report warnings on failure

4. **Coverage Report**
   - Generate test coverage
   - Upload to coverage service

## Adding New Tests

### For Implemented Features (Blocking):

1. Add test to appropriate `<feature>_test.go` file
2. Test should validate spec compliance
3. Test name should be descriptive: `TestFeatureSpecificBehavior`
4. Include reference to CSS spec section if applicable

### For Future Features (Non-Blocking):

1. Create new `<feature>_test.go` file with `_future` suffix initially
2. Mark test with comment: `// TODO: Feature not yet implemented`
3. Test can fail without blocking CI
4. When feature is implemented, remove `_future` suffix and move to blocking

## Test Quality Guidelines

### All tests should:
- Be deterministic (no randomness)
- Be isolated (no dependencies between tests)
- Be fast (< 100ms each)
- Have clear assertions with helpful error messages
- Include comments explaining what behavior is being validated
- Reference CSS specs when applicable

### Example:
```go
func TestFlexboxJustifyContentCenter(t *testing.T) {
    // CSS Flexbox §10.2: justify-content: center
    // Centers flex items along main axis
    root := &Node{
        Style: Style{
            Display:        DisplayFlex,
            FlexDirection:  FlexDirectionRow,
            JustifyContent: JustifyContentCenter,
            Width:          200,
        },
        Children: []*Node{
            {Style: Style{Width: 50, Height: 50}},
        },
    }

    LayoutFlexbox(root, Loose(200, 100))

    // Item should be centered: (200 - 50) / 2 = 75
    expectedX := 75.0
    if math.Abs(root.Children[0].Rect.X-expectedX) > 0.1 {
        t.Errorf("Item should be centered at X=%.2f, got %.2f",
            expectedX, root.Children[0].Rect.X)
    }
}
```

## Future Directions

### Enhanced Converter
- Support `<style>` blocks
- Parse CSS selectors
- Handle nested structures
- Support reference file comparison

### Automated WPT Sync
- Periodic sync with upstream WPT
- Automated test adaptation
- Coverage tracking

### Visual Regression Testing
- Screenshot comparison
- Reference image generation
- Pixel-perfect layout validation

### Performance Benchmarks
- Benchmark large layouts
- Track performance regressions
- Optimize critical paths
