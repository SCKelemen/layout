# CSS Grid Specification Review

## Issues Found

### 1. Missing `justify-items` and `align-items` Support for Grid

**Status:** ❌ Not Implemented

**Spec Requirement:**
- CSS Grid Layout requires `justify-items` (inline/row axis) and `align-items` (block/column axis) properties
- Default value is `stretch` for both
- Items with intrinsic aspect ratio (or `aspect-ratio` property) default to `start` instead of `stretch`

**Current Behavior:**
- Grid items always stretch to fill cells (unless they have `aspect-ratio`)
- No way to center, align to start/end, or use other alignment values
- `AlignItems` exists in `Style` but is only used for Flexbox, not Grid

**Expected Values:**
- `start` - align to start of grid area
- `end` - align to end of grid area  
- `center` - center within grid area
- `stretch` - stretch to fill grid area (default)

**Impact:** Medium - Prevents proper alignment control in Grid layouts

---

### 2. Aspect Ratio Behavior ✅

**Status:** ✅ Correctly Implemented

**Spec Requirement:**
- Items with `aspect-ratio` maintain their ratio
- They do NOT stretch to fill cells (default to `start` alignment)
- Aspect ratio takes precedence over stretching

**Current Behavior:**
- Items with `aspect-ratio` maintain their calculated size
- They don't stretch beyond their aspect-ratio-calculated dimensions
- Matches spec behavior

---

### 3. Grid Item Positioning

**Status:** ✅ Correctly Implemented

**Current Behavior:**
- Items are positioned at `cellX + margin.Left` and `cellY + margin.Top`
- Margins are accounted for in size calculations
- Padding/border offsets are correctly applied

---

## Recommendations

### Priority 1: Add `justify-items` and `align-items` for Grid

This would require:
1. Adding `JustifyItems` property to `Style` (or reusing existing `AlignItems` with context)
2. Implementing alignment logic in `grid.go` positioning code (lines 386-459)
3. Handling default `stretch` behavior
4. Handling `start` default for items with `aspect-ratio`

**Example Implementation:**
```go
// In grid.go, after calculating itemWidth/itemHeight:
var itemX, itemY float64

// Handle justify-items (inline/row axis)
switch node.Style.JustifyItems { // or create new property
case JustifyItemsStart:
    itemX = cellX + item.node.Style.Margin.Left
case JustifyItemsEnd:
    itemX = cellX + cellWidth - itemWidth - item.node.Style.Margin.Right
case JustifyItemsCenter:
    itemX = cellX + (cellWidth - itemWidth) / 2
case JustifyItemsStretch: // default
    itemX = cellX + item.node.Style.Margin.Left
    itemWidth = maxItemWidth // already set above
}

// Handle align-items (block/column axis)
switch node.Style.AlignItems {
case AlignItemsStart:
    itemY = cellY + item.node.Style.Margin.Top
case AlignItemsEnd:
    itemY = cellY + cellHeight - itemHeight - item.node.Style.Margin.Bottom
case AlignItemsCenter:
    itemY = cellY + (cellHeight - itemHeight) / 2
case AlignItemsStretch: // default
    itemY = cellY + item.node.Style.Margin.Top
    itemHeight = maxItemHeight // already set above
}
```

---

## Other Spec Compliance Notes

### ✅ Correctly Implemented:
- Grid track sizing (fractional, auto, minmax)
- Grid gaps (row/column)
- Grid spanning (row/column)
- Auto row/column generation
- Padding/border handling
- Margin handling within cells
- Aspect ratio constraint behavior

### ⚠️ Partially Implemented:
- Alignment: Only `stretch` (default) and implicit `start` for aspect-ratio items
- Missing: `start`, `end`, `center` alignment options

### ❌ Not Implemented (but may be out of scope):
- `justify-content` / `align-content` for grid container (distributing tracks)
- `place-items` shorthand
- `justify-self` / `align-self` per-item overrides
- `grid-auto-flow` (currently always row-major)
- `grid-auto-placement` algorithms
