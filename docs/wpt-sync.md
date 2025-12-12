# Web Platform Tests (WPT) Synchronization

This document explains how we track and integrate Web Platform Tests into our test suite.

## Overview

[Web Platform Tests](https://github.com/web-platform-tests/wpt) (WPT) is the official W3C test suite for web platform features. We use WPT as a reference to ensure our CSS layout implementation stays aligned with web standards.

## Current Status

- **Our Test Coverage**: 321/321 tests passing (100%)
- **WPT-Derived Tests**: 14 text layout tests in `wpt_comprehensive_test.go`
- **Test Strategy**: Regression testing with computed expected values

## WPT Test Conversion Strategy

### Why Not Direct WPT Test Conversion?

WPT tests are primarily **reference tests** (visual comparison) that don't include programmatic layout assertions. They compare rendered output against reference images.

Our approach:
1. **Parse WPT HTML structure** (layout containers and elements)
2. **Run our layout engine** to compute positions and sizes
3. **Generate Go tests** with computed values as expectations
4. **Use as regression tests** to ensure layout consistency

### Tools

#### `tools/wpt_layout_converter.go`

Enhanced WPT converter that:
- Parses HTML structure from WPT tests
- Builds our `Node` tree from CSS styles
- Runs layout engine to compute expected values
- Generates Go test code

**Usage:**
```bash
go run tools/wpt_layout_converter.go <wpt-test.html> <output_test.go>
```

**Limitations:**
- Currently supports inline styles only
- Needs enhancement for `<style>` tag CSS selectors
- Works best with simple layout tests

#### `tools/wpt_converter.go`

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

1. **Review the test**
   ```bash
   curl -O "https://raw.githubusercontent.com/web-platform-tests/wpt/master/css/css-flexbox/[test-name].html"
   ```

2. **Attempt automatic conversion**
   ```bash
   go run tools/wpt_layout_converter.go [test-name].html [output-test].go
   ```

3. **Manual enhancement** (if needed)
   - Parse `<style>` tags
   - Extract CSS selectors
   - Build equivalent Go test

4. **Add to test suite**
   ```bash
   go test -v ./... # Verify all tests pass
   git add [output-test].go
   git commit -m "Add WPT test: [description]"
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

## Future Enhancements

### Potential Improvements

1. **Enhanced Style Parser**
   - Parse `<style>` tags
   - Match CSS selectors to elements
   - Compute cascaded styles

2. **Automated Conversion**
   - Batch convert WPT tests
   - Generate test coverage reports
   - Identify gaps automatically

3. **CI Integration**
   - Run WPT tests directly (headless browser)
   - Compare visual output
   - Automated regression detection

4. **Test Generation**
   - Generate tests from CSS spec examples
   - Property combination testing
   - Edge case enumeration

## References

- [Web Platform Tests](https://github.com/web-platform-tests/wpt)
- [CSS Flexbox Spec](https://www.w3.org/TR/css-flexbox-1/)
- [CSS Grid Spec](https://www.w3.org/TR/css-grid-1/)
- [CSS Sizing Spec](https://www.w3.org/TR/css-sizing-3/)
- [Our Spec Compliance Status](../SPEC_COMPLIANCE_STATUS.md)
- [Specification Gaps](../SPECIFICATION_GAPS.md)

## Contributing

To add WPT-derived tests:

1. Identify relevant WPT test
2. Convert using `wpt_layout_converter.go` or manually
3. Verify test passes
4. Submit PR with test description and WPT reference

Example commit message:
```
Add WPT test: flexbox align-content center

Converted from WPT test css/css-flexbox/align-content-001.htm

Tests multi-line flex container with align-content: center
```

---

**Last Updated**: 2025-12-12
**Test Coverage**: 321/321 passing (100%)
**WPT Sync**: Automated weekly via GitHub Actions
