# CSS Flexbox and Block Layout Specification Review

## Flexbox Issues Found

### 1. Missing `align-content` Implementation

**Status:** ❌ Not Implemented

**Spec Requirement:**
- CSS Flexbox requires `align-content` property to align multiple flex lines along the cross axis
- Only applies when `flex-wrap: wrap` or `wrap-reverse` (multiple lines)
- Default value is `stretch`
- Values: `flex-start`, `flex-end`, `center`, `space-between`, `space-around`, `stretch`

**Current Behavior:**
- `AlignContent` property exists in `Style` struct and is serialized
- But it's **not used** in `LayoutFlexbox` function
- Lines are simply stacked with `totalCrossSize += lineCrossSize` (line 296)
- No alignment/distribution of lines along cross axis
- Comment on line 212-214 acknowledges this limitation

**Expected Behavior:**
- When multiple lines exist, `align-content` should distribute them along cross axis
- `stretch` (default): Lines stretch to fill available cross space
- `flex-start`: Lines packed at start
- `flex-end`: Lines packed at end
- `center`: Lines centered
- `space-between`: First line at start, last at end, equal space between
- `space-around`: Equal space around each line

**Impact:** Medium - Prevents proper multi-line flexbox layouts

**Code Location:** `flexbox.go` lines 294-304

---

### 2. Flex Direction Reverse Not Implemented

**Status:** ⚠️ Partially Implemented

**Spec Requirement:**
- `flex-direction: row-reverse` should reverse item order along main axis
- `flex-direction: column-reverse` should reverse item order along main axis

**Current Behavior:**
- `FlexDirectionRowReverse` and `FlexDirectionColumnReverse` exist
- But they're treated as aliases for `Row` and `Column` (line 35)
- Items are not reversed in order

**Impact:** Low - Functionality works but order is wrong

---

### 3. Flex Wrap Reverse Not Implemented

**Status:** ⚠️ Partially Implemented

**Spec Requirement:**
- `flex-wrap: wrap-reverse` should reverse the order of flex lines along cross axis

**Current Behavior:**
- `FlexWrapWrapReverse` exists
- But it's treated as `FlexWrapWrap` (line 152)
- Lines are not reversed

**Impact:** Low - Functionality works but line order is wrong

---

### 4. Flex Gap Not Implemented

**Status:** ❌ Not Implemented

**Spec Requirement:**
- CSS Flexbox supports `gap`, `row-gap`, `column-gap` properties
- Adds spacing between flex items (not margins)

**Current Behavior:**
- No gap support in flexbox
- Comment on line 298 mentions "Add gap if specified (simplified, assuming 0 for now)"

**Impact:** Low - Can use margins as workaround

---

## Block Layout Issues Found

### 1. Margin Collapsing Not Implemented

**Status:** ❌ Not Implemented

**Spec Requirement:**
- CSS Block Layout requires margin collapsing between adjacent block-level siblings
- Vertical margins collapse: max(margin1, margin2) instead of margin1 + margin2
- Only vertical margins collapse (not horizontal)
- Margins don't collapse in certain contexts (flexbox, grid, positioned elements)

**Current Behavior:**
- Block layout doesn't handle child margins at all
- Children are positioned with `currentY += childSize.Height` (line 188)
- No margin calculation or collapsing
- Documentation states "Margin support is not yet implemented" (limitations.md line 14)

**Expected Behavior:**
- Calculate child margins
- Collapse adjacent vertical margins
- Position children accounting for collapsed margins
- Don't collapse margins in flexbox/grid contexts

**Impact:** High - Block layout spacing is incorrect

**Code Location:** `block.go` lines 152-192

---

### 2. Child Margins Not Accounted For

**Status:** ❌ Not Implemented

**Spec Requirement:**
- Block layout should account for child margins when positioning
- Margins create space between parent and child, or between siblings

**Current Behavior:**
- Children positioned at `currentY` without margin consideration
- No margin space between children or between parent and first/last child

**Impact:** High - Incorrect spacing in block layouts

---

## Recommendations

### Priority 1: Implement Block Layout Margin Collapsing

This is critical for correct block layout behavior:

1. Calculate child margins
2. Collapse adjacent vertical margins (max of two margins)
3. Position children accounting for margins
4. Handle first/last child margins (don't collapse with parent if parent has padding/border)

**Example Implementation:**
```go
// In LayoutBlock, after measuring child:
childMarginTop := child.Style.Margin.Top
childMarginBottom := child.Style.Margin.Bottom

// Collapse with previous child's bottom margin
if currentY > 0 {
    // Previous child exists - collapse margins
    previousMarginBottom := previousChild.Style.Margin.Bottom
    collapsedMargin := math.Max(previousMarginBottom, childMarginTop)
    currentY += collapsedMargin - previousMarginBottom // Adjust for collapsed margin
} else {
    // First child - use top margin
    currentY += childMarginTop
}

// Position child
child.Rect.Y = node.Style.Padding.Top + node.Style.Border.Top + currentY
currentY += childSize.Height + childMarginBottom
```

---

### Priority 2: Implement Flexbox `align-content`

This enables proper multi-line flexbox layouts:

1. Calculate total cross size of all lines
2. Calculate free cross space (if container cross size > total line cross size)
3. Apply `align-content` to distribute lines:
   - `stretch`: Distribute free space equally to each line
   - `flex-start`: Pack lines at start
   - `flex-end`: Pack lines at end
   - `center`: Center lines
   - `space-between`: First at start, last at end, equal space between
   - `space-around`: Equal space around each line

**Code Location:** `flexbox.go` lines 294-304

---

### Priority 3: Implement Flex Direction/Wrap Reverse

Lower priority but improves spec compliance:

1. For `row-reverse`/`column-reverse`: Reverse item order in each line
2. For `wrap-reverse`: Reverse line order

---

## Other Spec Compliance Notes

### ✅ Correctly Implemented (Flexbox):
- `justify-content` (all values)
- `align-items` (all values)
- `flex-grow` and `flex-shrink`
- `flex-basis`
- Flex wrapping (basic)
- Main axis sizing and distribution
- Cross axis item alignment (single line)

### ✅ Correctly Implemented (Block):
- Basic vertical stacking
- Width/height calculations
- Padding and border handling
- Box-sizing support
- Aspect ratio handling
- Min/max constraints
- Auto width/height

### ⚠️ Partially Implemented:
- Flexbox multi-line: Works but missing `align-content`
- Flexbox reverse directions: Work but don't reverse order
- Block layout: Works but missing margin support

### ❌ Not Implemented (but may be out of scope):
- Flexbox `gap` property
- Block formatting context (BFC) establishment
- Inline layout and text flow
- Baseline alignment
