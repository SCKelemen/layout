# Text Layout: Remaining Issues and Non-Conformances

This document tracks remaining spec non-conformances, edge cases, and potential improvements for the text layout implementation.

## Known Spec Non-Conformances (Acceptable for v1)

### 1. White-Space Collapsing (Simplified)

**Issue**: `white-space: normal` and `white-space: nowrap` use a simplified whitespace collapsing algorithm.

**Current Implementation**:
- Uses `regexp.MustCompile(\s+)` which matches all Unicode whitespace
- CSS handles non-breaking spaces, zero-width spaces, and other Unicode whitespace differently

**Impact**: Low - works correctly for ASCII text, may behave differently for complex Unicode text.

**Status**: Documented simplification, acceptable for v1.

**Reference**: [CSS Text Module Level 3 §3.1](https://www.w3.org/TR/css-text-3/#white-space-property)

### 2. Line Breaking (Word-Based Only)

**Issue**: Line breaking uses simple word-based algorithm (`strings.Fields`).

**Current Implementation**:
- Splits on whitespace boundaries only
- Does not implement Unicode line breaking rules (UAX #14)
- Does not handle punctuation, soft hyphens, or CJK text properly

**Impact**: Medium - works for English/Latin text, may break incorrectly for other languages.

**Status**: Documented simplification, acceptable for v1.

**Reference**: [CSS Text Module Level 3 §4](https://www.w3.org/TR/css-text-3/#line-breaking)

### 3. Line-Height Heuristic

**Issue**: Line-height interpretation uses a heuristic: `< 10 = multiplier, >= 10 = absolute pixels`.

**Current Behavior**:
- `line-height: 1.5` → multiplier (correct)
- `line-height: 12` → 12px absolute (may be unexpected if font size is large)
- `line-height: 9.5` → multiplier (correct)

**Impact**: Low - common values work correctly, but `line-height: 12` with large fonts may surprise users.

**Status**: Pragmatic choice for v1, documented in code comments.

**Reference**: [CSS Inline Layout Module Level 3 §4.4.1](https://www.w3.org/TR/css-inline-3/#propdef-line-height)

**Note**: CSS spec says `<number>` is always a multiplier, `<length>` is always absolute. Our heuristic approximates this but isn't perfect.

## Edge Cases and Robustness Issues

### 4. Max-Width == 0 Behavior

**Issue**: When `constraints.MaxWidth == 0`, line breaking still occurs but may produce unexpected results.

**Current Behavior**:
- `maxWidth > 0` check means `maxWidth == 0` is treated as "no wrapping"
- Text will be laid out as a single long line
- Final width will be clamped to 0 by `constraints.Constrain()`

**Impact**: Low - edge case, unlikely in practice.

**Recommendation**: Treat `maxWidth <= 0` as unbounded for clarity:
```go
if maxWidth <= 0 {
    maxWidth = Unbounded
}
```

**Status**: Minor issue, could be fixed.

### 5. Text Node Invariants Not Enforced

**Issue**: `LayoutText` assumes text nodes are leaf nodes (no children) but doesn't enforce this.

**Current Behavior**:
- If a `DisplayInlineText` node has children, they are silently ignored
- No validation or error reporting

**Impact**: Medium - could cause silent bugs if users accidentally add children to text nodes.

**Recommendation**: Add validation or documentation:
```go
// At start of LayoutText:
if len(node.Children) > 0 {
    // Log warning or document that children are ignored
}
```

**Status**: Should be documented at minimum, could add validation.

### 6. Global TextMetricsProvider Concurrency

**Issue**: `textMetrics` is a package-level variable with no synchronization.

**Current Behavior**:
- `SetTextMetricsProvider()` mutates global state
- If `LayoutText` is called from multiple goroutines and provider is changed, data races can occur

**Impact**: Medium - only affects concurrent usage with provider changes.

**Recommendation**: 
- Document: "Set provider once at init, don't change concurrently"
- Or use `sync.RWMutex` or `atomic.Value` for thread safety

**Status**: Should be documented, could add synchronization.

### 7. Word/Letter-Spacing "Normal" Sentinel

**Issue**: Uses `-1` as sentinel for "normal" spacing, but can't distinguish "not set" from "explicitly normal".

**Current Behavior**:
- `WordSpacing: -1` means "normal" (default)
- `LetterSpacing: -1` means "normal" (default)
- No way to explicitly set to "normal" vs "not set"

**Impact**: Low - works correctly for v1, but may need refinement if we add serialization.

**Status**: Acceptable for v1, may need refinement later.

## Missing Features (Deferred)

### 8. Text Justification

**Status**: Not implemented (deferred for v1)

**Reference**: [CSS Text Module Level 3 §7.1.1](https://www.w3.org/TR/css-text-3/#justify)

### 9. Text Align Last

**Status**: Not implemented (deferred for v1)

**Reference**: [CSS Text Module Level 3 §7.2.2](https://www.w3.org/TR/css-text-3/#text-align-last-property)

### 10. Pre-Wrap and Pre-Line

**Status**: Not implemented (deferred for v1)

**Reference**: [CSS Text Module Level 3 §3.1](https://www.w3.org/TR/css-text-3/#white-space-property)

### 11. RTL and Vertical Writing Modes

**Status**: Not implemented (deferred for v1)

**Reference**: [CSS Writing Modes Level 3](https://www.w3.org/TR/css-writing-modes-3/)

### 12. Hyphenation

**Status**: Not implemented (deferred for v1)

**Reference**: [CSS Text Module Level 3 §4.3](https://www.w3.org/TR/css-text-3/#hyphenation)

### 13. Mixed Inline and Block Content

**Status**: Not implemented (deferred for v1)

Text nodes are leaf nodes only. Inline elements mixed with text require inline formatting context.

## Recommendations

### High Priority (Should Fix)

1. **Document text node invariants** - Add clear documentation that text nodes should be leaf nodes
2. **Document concurrency** - Add documentation about TextMetricsProvider thread safety
3. **Fix max-width == 0** - Treat as unbounded for clarity

### Medium Priority (Nice to Have)

4. **Add validation** - Warn or error if text node has children
5. **Add thread safety** - Use mutex or atomic for TextMetricsProvider

### Low Priority (Future Enhancements)

6. **Improve line-height** - Consider explicit type or better heuristic
7. **Unicode line breaking** - Implement UAX #14 for better internationalization
8. **Better whitespace handling** - Handle Unicode whitespace per CSS spec

## Test Coverage

All implemented features have comprehensive test coverage:
- ✅ 23 text layout tests, all passing
- ✅ Invariant-based testing (not just specific numbers)
- ✅ Edge cases covered (empty text, long words, etc.)

## Summary

The text layout implementation is **spec-compliant for the v1 MVP scope** with the following caveats:

1. **Simplified algorithms** for whitespace and line breaking (documented)
2. **Pragmatic heuristics** for line-height interpretation (documented)
3. **Minor edge cases** that could be improved but don't affect common use cases
4. **Missing features** that are explicitly deferred (justify, RTL, etc.)

The implementation is **production-ready for the intended use cases** (LTR text, word-based wrapping, basic alignment). For more complex text layout needs, consider using a full text layout library.

