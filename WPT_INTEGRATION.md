# Web Platform Tests (WPT) Integration

## Overview

This document describes our integration with Web Platform Tests (WPT) and test organization strategy.

## WPT Repository Stats

We've analyzed the WPT repository (https://github.com/web-platform-tests/wpt) for CSS layout tests:

### Total Available Tests: 9,312+ HTML files

| Feature | WPT Test Count | Implementation Status |
|---------|----------------|----------------------|
| CSS Text | 2,667 | ✅ Core features implemented |
| CSS Grid | 2,386 | ✅ Core features implemented |
| CSS Flexbox | 1,621 | ✅ 95%+ spec coverage |
| CSS Contain | 805 | ⏳ Not started |
| CSS Sizing | 778 | ⚠️ Partial implementation |
| CSS Position | 417 | ✅ Core features implemented |
| CSS Align | 330 | ✅ Implemented in flex/grid |
| CSS Box | 156 | ✅ Complete |
| CSS Display | 152 | ⚠️ Partial implementation |

## Implementation Status

### ✅ Fully Implemented (Blocking Tests)
- **Flexbox**: All major features, 321 Go tests passing
- **Grid**: Core grid layout, spanning, alignment
- **Text Layout**: Line breaking (UAX #14), white-space, alignment, overflow
- **Position**: Absolute, relative, fixed, sticky
- **Block Layout**: Basic block flow, margin collapsing
- **Box Model**: Padding, borders, margins, box-sizing

### ⚠️ Partially Implemented
- **Sizing**: Basic sizing done, intrinsic sizing (min-content, max-content) partial
- **Display**: Block, flex, grid, inline-text working; flow-root, contents planned

### ⏳ Future Implementation (Non-Blocking Tests)
- **Subgrid**: Grid Level 2 feature
- **Containment**: Size, layout, paint containment
- **Advanced Text**: text-align-last, text-justify inter-character
- **Container Queries**: Size and style queries

## Direct WPT Conversion Challenges

Real WPT tests use sophisticated infrastructure that our simple converter can't handle:

### WPT Test Structure:
```html
<style>
.test {
  display: flex;
  width: 100px;
}
</style>
<div class="test" data-offset-x="0">
  <div>Content</div>
</div>
<script src="/resources/testharness.js"></script>
<script src="/resources/check-layout-th.js"></script>
```

### Our Converter Supports:
```html
<div style="display: flex; width: 100px" 
     data-expected-width="100">
  Content
</div>
```

### Limitations:
- ❌ `<style>` blocks → Only inline styles
- ❌ CSS selectors → No class-based styling
- ❌ JavaScript test harness → Manual Go test assertions
- ❌ Reference files (reftests) → No visual comparison
- ❌ Complex nested structures → Flat structure only
- ❌ `data-offset-x/y` → Different assertion format

## Our Testing Strategy

Instead of automated WPT conversion, we use a hybrid approach:

### 1. Manual Test Adaptation
- Identify high-value WPT tests
- Manually adapt to Go test format
- Link to source WPT test for traceability

Example:
```go
// TestFlexboxAlignContent validates align-content behavior
// Based on WPT: css/css-flexbox/align-content-wrap-001.html
func TestFlexboxAlignContent(t *testing.T) {
    root := &Node{
        Style: Style{
            Display:      DisplayFlex,
            FlexWrap:     FlexWrapWrap,
            AlignContent: AlignContentCenter,
            Height:       100,
        },
        Children: []*Node{
            {Style: Style{Width: 50, Height: 20}},
        },
    }
    
    LayoutFlexbox(root, Loose(200, 100))
    
    // Verify alignment
    expectedY := 40.0  // (100 - 20) / 2
    if math.Abs(root.Children[0].Rect.Y - expectedY) > 0.1 {
        t.Errorf("Expected Y=%.2f, got %.2f", expectedY, root.Children[0].Rect.Y)
    }
}
```

### 2. Comprehensive Go Test Suite
- **321 tests** currently passing (100% success rate)
- Organized by feature (flexbox, grid, text, position, block, box)
- Each test validates specific spec behavior
- References CSS spec sections

### 3. CI Test Groups
- **Blocking tests**: Implemented features must pass
- **Non-blocking tests**: Future features can fail without blocking
- **Coverage tracking**: Measure test coverage
- **Code quality**: Format and linting checks

## CI Configuration

See `.github/workflows/test.yml` for full configuration.

### Test Matrix:
- **Go versions**: 1.21, 1.23, 1.25
- **Blocking groups**: Flexbox, Grid, Text, Position, Block, Box
- **Non-blocking groups**: Subgrid, Advanced Text, Advanced Sizing, Containment, Display Level 3

### Workflow:
1. **Blocking Tests** run in parallel across Go versions and feature groups
2. **Comprehensive Tests** run all tests with coverage
3. **Non-blocking Tests** run future features (failures allowed)
4. **Quality Checks** verify formatting and linting
5. **Summary** reports overall status

## Future WPT Integration

### Enhanced Converter (Future)
To increase automated WPT coverage, the converter would need:

1. **Style Block Parsing**
   - Parse `<style>` tags
   - Extract CSS rules
   - Apply rules to elements

2. **CSS Selector Support**
   - Parse class selectors
   - Parse ID selectors
   - Parse descendant selectors
   - Apply cascade rules

3. **Reference File Comparison**
   - Parse reference HTML files
   - Layout both test and reference
   - Compare resulting layouts
   - Pixel-perfect matching

4. **JavaScript Harness Integration**
   - Understand testharness.js assertions
   - Parse check-layout-th.js expectations
   - Convert to Go assertions

### Incremental WPT Sync
- Periodic checks for new WPT tests
- Semi-automated adaptation workflow
- Coverage tracking dashboard
- Gap analysis reports

## Test Coverage Goals

| Feature | Current Coverage | WPT Available | Goal |
|---------|------------------|---------------|------|
| Flexbox | ~95% | 1,621 | Maintain 95%+ |
| Grid | ~80% | 2,386 | 90% by Q2 2026 |
| Text | ~75% | 2,667 | 85% by Q2 2026 |
| Position | ~70% | 417 | 85% by Q2 2026 |
| Block | ~60% | (mixed) | 75% by Q3 2026 |
| Box | ~90% | 156 | Maintain 90%+ |

## Contributing Tests

### Adding Tests from WPT:

1. **Find relevant WPT test**:
   ```bash
   cd /tmp/wpt/css/css-flexbox
   grep -l "justify-content: space-between" *.html
   ```

2. **Analyze test**:
   - What behavior is being tested?
   - What are the expected results?
   - Can it be adapted to our format?

3. **Create Go test**:
   ```go
   // TestFlexboxJustifyContentSpaceBetween validates space-between distribution
   // Based on WPT: css/css-flexbox/justify-content-001.html
   func TestFlexboxJustifyContentSpaceBetween(t *testing.T) {
       // Implementation
   }
   ```

4. **Add to appropriate file**:
   - `flexbox_test.go` for core flexbox
   - `flexbox_alignment_test.go` for alignment
   - `flexbox_gap_test.go` for gap properties
   - etc.

5. **Reference CSS spec**:
   ```go
   // CSS Flexbox §10.2: Justify Content
   // https://www.w3.org/TR/css-flexbox-1/#justify-content-property
   ```

### Test Quality Checklist:
- [ ] Test name clearly describes behavior
- [ ] Includes WPT reference if adapted
- [ ] Includes CSS spec reference
- [ ] Has clear assertions with helpful error messages
- [ ] Runs in < 100ms
- [ ] Is deterministic (no randomness)
- [ ] Is isolated (no dependencies on other tests)

## Resources

- **WPT Repository**: https://github.com/web-platform-tests/wpt
- **WPT Documentation**: https://web-platform-tests.org/
- **CSS Specs**: https://www.w3.org/Style/CSS/
- **Our Test Organization**: See `TEST_ORGANIZATION.md`
- **CI Configuration**: See `.github/workflows/test.yml`

## Summary

While we can't directly convert all 9,312+ WPT HTML tests, we've created a robust testing strategy that:

1. ✅ Organizes tests by feature
2. ✅ Separates blocking (implemented) from non-blocking (future) tests
3. ✅ Runs tests in CI on every commit
4. ✅ Tracks coverage and quality
5. ✅ Provides clear path for adding WPT-inspired tests
6. ✅ References WPT and CSS specs for traceability

**Current Status**: 321 tests passing, 100% success rate, comprehensive coverage of flexbox, grid, text, and position layout.
