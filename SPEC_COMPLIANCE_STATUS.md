# CSS Specification Compliance Status

**Last reconciled:** 2026-01-19

## Summary

**Overall Test Results:** 853 / 853 passing, 1 skipped, 0 failing.

That is the count produced by `go test ./... -count=1 -v` against this commit,
counted as:

- 528 top-level test functions reporting `PASS` (one additional top-level test
  reports `SKIP`, for 529 top-level tests total).
- 325 subtests reporting `PASS`.
- 1 subtest reporting `SKIP`.

This document supersedes the historical `SPEC_REVIEW.md` and
`FLEXBOX_BLOCK_SPEC_REVIEW.md` reviews, both of which have been retired (they
catalogued features that have since been implemented and tested).

## ✅ Fully Implemented Features

### CSS Grid Layout Module Level 1

#### Core Grid Features
- ✅ **Grid track sizing** — fixed, fractional (fr), auto, minmax()
- ✅ **Grid gaps** — row-gap, column-gap, gap shorthand
- ✅ **Grid spanning** — items can span multiple rows/columns
- ✅ **Auto row/column generation** — implicit grid tracks
- ✅ **Grid template areas** — named grid regions
- ✅ **Grid area placement** — place items in named areas

#### Grid Alignment & Distribution
- ✅ **justify-items** — inline-axis alignment for all items (default: stretch)
- ✅ **align-items** — block-axis alignment for all items (default: stretch)
- ✅ **justify-self** — per-item inline-axis alignment override
- ✅ **align-self** — per-item block-axis alignment override
- ✅ **align-content** — row track distribution (stretch, start, end, center, space-between, space-around)
- ✅ **justify-content** — column track distribution

#### Grid Auto-Placement
- ✅ **grid-auto-flow** — row, column, row dense, column dense
- ✅ **Dense packing algorithm** — fills holes with smaller items
- ✅ **Row-major placement** — default auto-placement
- ✅ **Column-major placement** — alternative auto-placement

#### Grid Advanced Features
- ✅ **repeat()** — track repetition helper
- ✅ **auto-fill** — dynamic track generation (fills container)
- ✅ **auto-fit** — dynamic track generation (collapses empty)

### CSS Flexbox Layout Module Level 1

All four features previously called out as missing in
`FLEXBOX_BLOCK_SPEC_REVIEW.md` are now present in code and covered by tests:

| Feature                       | Implementation site                | Test                                                |
|-------------------------------|------------------------------------|-----------------------------------------------------|
| `align-content`               | `flexbox_align_content.go`         | `flexbox_extended_fixes_test.go` and others         |
| `flex-direction: *-reverse`   | `flexbox_setup.go`, `flexbox.go`   | `flexbox_reverse_test.go`, `flexbox_extended_test.go` |
| `flex-wrap: wrap-reverse`     | `flexbox_wrap_reverse.go`          | `flexbox_reverse_test.go`                            |
| `gap` / `row-gap` / `column-gap` | `flexbox.go`, `flexbox_setup.go` | `flexbox_extended_fixes_test.go::TestFlexboxRowGapAndColumnGap` |

#### Core Flexbox Features
- ✅ **flex-direction** — row, column, row-reverse, column-reverse
- ✅ **flex-wrap** — nowrap, wrap, wrap-reverse
- ✅ **justify-content** — main-axis alignment (start, end, center, space-between, space-around, space-evenly)
- ✅ **align-items** — cross-axis alignment (start, end, center, stretch, baseline)
- ✅ **align-content** — multi-line cross-axis distribution
- ✅ **align-self** — per-item cross-axis alignment override

#### Flexbox Sizing
- ✅ **flex-grow** — growth factor
- ✅ **flex-shrink** — shrink factor
- ✅ **flex-basis** — base size before flex

#### Flexbox Gaps & Order
- ✅ **gap** — gap between flex items
- ✅ **row-gap** — cross-axis gap (wrapping)
- ✅ **column-gap** — main-axis gap
- ✅ **order** — visual reordering of flex items

#### Flexbox Alignment
- ✅ **Baseline alignment** — align items by text baseline

### CSS Box Model Module Level 3

#### Block Layout
- ✅ **Basic vertical stacking** — block-level elements stack vertically
- ✅ **Margin collapsing** — adjacent vertical margins collapse (max, not sum); see `block_children.go` and `block_margin_collapsing_test.go`
- ✅ **Width/height calculations** — auto, explicit, min/max
- ✅ **Padding and border** — box model spacing
- ✅ **box-sizing** — both `content-box` and `border-box` are implemented and tested. See `convertToContentSize` / `convertFromContentSize` / `convertMinMaxToContentSize` in `types.go`, threaded through `block_setup.go`, `flexbox_setup.go`, `grid_setup.go`, `grid.go`, and `text.go`. Test coverage lives in `box_sizing_test.go` (10 tests).

### CSS Sizing Module Level 3

#### Intrinsic Sizing
- ✅ **min-content** — narrowest width without overflow
- ✅ **max-content** — widest natural width (no wrapping)
- ✅ **fit-content** — clamp max-content to specified size
- ✅ **Intrinsic sizing for block layout**
- ✅ **Intrinsic sizing for flexbox**
- ✅ **Intrinsic sizing for grid**

#### Constraints
- ✅ **min-width, min-height** — minimum size constraints
- ✅ **max-width, max-height** — maximum size constraints
- ✅ **aspect-ratio** — maintain width/height ratio

### CSS Text Module Level 3 (v1 MVP)

#### Text Properties
- ✅ **white-space** — normal, nowrap, pre, pre-wrap, pre-line
- ✅ **text-align** — left, right, center, justify
- ✅ **text-align-last** — alignment for last line
- ✅ **text-justify** — justification method (inter-word, none, auto)
- ✅ **line-height** — normal, multiplier, absolute
- ✅ **text-indent** — first line indentation
- ✅ **word-spacing** — extra spacing between words
- ✅ **letter-spacing** — extra spacing between letters
- ✅ **overflow-wrap** — break-word wrapping
- ✅ **word-break** — break-all wrapping
- ✅ **text-overflow** — clip, ellipsis

#### Text Layout Algorithm
- ✅ **Whitespace collapsing** — Unicode-aware
- ✅ **Line breaking** — UAX #14 Unicode line breaking
- ✅ **Grapheme clusters** — UAX #29 handling
- ✅ **CJK text support** — Chinese, Japanese, Korean
- ✅ **Non-breaking spaces** — U+00A0 preservation

### CSS Positioned Layout Module Level 3

- ✅ **position** — static, relative, absolute, fixed, sticky
- ✅ **top, right, bottom, left** — positioning offsets
- ✅ **z-index** — stacking order

### CSS Transforms Module Level 1

- ✅ **transform** — translate, scale, rotate, skew, matrix

## ⚠️ Known Limitations (Acceptable for Current Scope)

### Text Layout
- ⚠️ **Line-height heuristic** — uses heuristic (< 10 = multiplier, >= 10 = absolute)
- ⚠️ **RTL / Vertical writing modes** — deferred (explicitly out of scope for v1)
- ⚠️ **Hyphenation** — deferred (soft hyphens U+00AD are supported)
- ⚠️ **Mixed inline/block content** — deferred (requires inline formatting context)

### Performance
- ⚠️ **TextMetricsProvider concurrency** — global variable, no synchronization (documented)

## 📊 Implementation Coverage by Module

### CSS Grid Layout: 100% (Level 1)
- ✅ Track sizing algorithms (including intrinsic sizing)
- ✅ Auto-placement (all modes)
- ✅ Alignment (all properties)
- ✅ Template areas
- ✅ Auto-fill / auto-fit
- ✅ Dense packing
- ⚠️ Out of scope: subgrid (Level 2 feature)

### CSS Flexbox Layout: 100% (Level 1)
- ✅ All flex properties
- ✅ All alignment properties
- ✅ Gaps
- ✅ Order
- ✅ Baseline alignment
- ✅ Wrap and wrap-reverse

### CSS Box Model: 100%
- ✅ Margin collapsing
- ✅ Padding, border
- ✅ Box-sizing (content-box and border-box)

### CSS Sizing: 100% (Level 3)
- ✅ Intrinsic sizing (min-content, max-content, fit-content)
- ✅ Min/max constraints
- ✅ Aspect ratio
- ⚠️ Out of scope: contain-intrinsic-size (Level 4 feature)

### CSS Text (v1 MVP): 100%
- ✅ All v1 MVP features implemented
- ⚠️ Out of scope: RTL, hyphenation, text decorations

## 🔗 Specification References

### Primary Specifications

1. **CSS Grid Layout Module Level 1** — https://www.w3.org/TR/css-grid-1/
2. **CSS Flexible Box Layout Module Level 1** — https://www.w3.org/TR/css-flexbox-1/
3. **CSS Box Model Module Level 3** — https://www.w3.org/TR/css-box-3/
4. **CSS Sizing Module Level 3** — https://www.w3.org/TR/css-sizing-3/
5. **CSS Text Module Level 3** — https://www.w3.org/TR/css-text-3/
6. **CSS Positioned Layout Module Level 3** — https://www.w3.org/TR/css-position-3/
7. **CSS Transforms Module Level 1** — https://www.w3.org/TR/css-transforms-1/

### Unicode Standards

1. **UAX #14: Unicode Line Breaking Algorithm** — https://www.unicode.org/reports/tr14/
2. **UAX #29: Unicode Text Segmentation** — https://www.unicode.org/reports/tr29/

## 📝 Retired Documentation

The following review documents have been retired because their findings were
out of date (every feature they listed as "not implemented" is now implemented
and tested):

- `SPEC_REVIEW.md` (Grid review) — retired; superseded by this document.
- `FLEXBOX_BLOCK_SPEC_REVIEW.md` (Flexbox / Block review) — retired; superseded
  by this document.

Both files have been removed from the working tree. Git history retains them at
their last commit if you need the original wording.

## 🚀 Next Steps (Optional Future Work)

1. **CSS Grid subgrid** (Level 2 feature)
2. **RTL text direction** (full bidirectional algorithm, UAX #9)
3. **Hyphenation** (language-specific dictionaries)
4. **Text decorations** (underline, overline, line-through)
5. **Inline formatting context** (mixed inline/block content)
6. **contain-intrinsic-size** (CSS Sizing Level 4)
7. **Additional WPT test coverage** for grid and flexbox

---

**Test Coverage:** 853 / 853 passing, 1 skipped.
**Spec Compliance:** within the documented module scope.
