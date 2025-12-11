# Common Gotchas and Pitfalls

This document covers common issues and unexpected behaviors when using the layout library. These behaviors match the CSS Grid and Flexbox specifications, but may be surprising if you're not familiar with them.

## Grid Auto Rows and Empty Items

### The Problem

When using auto-sized grid rows (`AutoTrack()`), items without content and without `MinHeight` will measure to **0 height**, causing rows to collapse.

### Example

```go
grid := layout.GridAuto(2, 2) // Auto-sized rows

// ❌ This will collapse to 0 height
item := &layout.Node{
    Style: layout.Style{
        GridRowStart: 0,
        GridColumnStart: 0,
        // No MinHeight, no content = 0 height!
    },
}

// ✅ This works correctly
item := &layout.Node{
    Style: layout.Style{
        GridRowStart: 0,
        GridColumnStart: 0,
        MinHeight: 50.0, // Required for auto rows!
    },
}
```

### Why This Happens

This behavior matches the [CSS Grid specification](https://www.w3.org/TR/css-grid-1/). In CSS Grid:

- Auto-sized rows determine their height based on the **content** of items in that row
- If an item has no content (no children, no text) and no `min-height`, it has **0 intrinsic height**
- Rows with only 0-height items will collapse to 0 height

### Solutions

**Option 1: Set MinHeight on all items** (Recommended)

```go
item := layout.MinHeight(&layout.Node{
    Style: layout.Style{
        GridRowStart: 0,
        GridColumnStart: 0,
    },
}, 50.0)
```

**Option 2: Use fixed-size rows instead of auto**

```go
// Instead of AutoTrack()
gridRows := []layout.GridTrack{
    layout.FixedTrack(100),
    layout.FixedTrack(100),
}
```

**Option 3: Use minmax() tracks** (when implemented)

```go
// This would ensure rows have a minimum height
gridRows := []layout.GridTrack{
    layout.MinMaxTrack(50, layout.Unbounded), // min 50px, max auto
}
```

### Height vs MinHeight

Both `Height` and `MinHeight` work for auto rows, but they behave slightly differently:

- **`Height`**: Explicit height. Item will be exactly this tall (unless constrained).
- **`MinHeight`**: Minimum height. Item will be at least this tall, but can grow if it has content.

For auto rows, **both work**, but `MinHeight` is more flexible:

```go
// Using Height (explicit)
item := &layout.Node{
    Style: layout.Style{
        GridRowStart: 0,
        GridColumnStart: 0,
        Height: 50.0, // Item will be exactly 50px
    },
}

// Using MinHeight (recommended for auto rows)
item := &layout.Node{
    Style: layout.Style{
        GridRowStart: 0,
        GridColumnStart: 0,
        MinHeight: 50.0, // Item will be at least 50px, can grow
    },
}
```

### Best Practice

When using auto-sized grid rows, **always set `MinHeight` (or `Height`) on items** that don't have content:

```go
grid := layout.GridAuto(3, 3)
for i := 0; i < 9; i++ {
    item := &layout.Node{
        Style: layout.Style{
            GridRowStart: i / 3,
            GridColumnStart: i % 3,
            MinHeight: 100.0, // Always set this!
        },
    }
    grid.Children = append(grid.Children, item)
}
```

## Items Spanning Multiple Rows

### How Spanning Works

When an item spans multiple auto rows, its height is **distributed** across those rows:

```go
// Item spanning 3 rows with MinHeight 300
item := &layout.Node{
    Style: layout.Style{
        GridRowStart: 0, GridRowEnd: 3, // Spans rows 0, 1, 2
        GridColumnStart: 0, GridColumnEnd: 1,
        MinHeight: 300.0, // This height is distributed: 100px per row
    },
}
```

### Row Height Calculation

- Each row gets at least `itemHeight / spanRows` height from the spanning item
- If multiple items span the same row, the row uses the **maximum** required height
- **Spanning items always fill their cell** - their final height equals the sum of the row heights they span (plus gaps between rows)
- Single-row items in auto rows use their intrinsic size (content + MinHeight)

### Example

```go
grid := layout.GridAuto(4, 2, 100, 100)

// Item spanning 3 rows
item1 := &layout.Node{
    Style: layout.Style{
        GridRowStart: 0, GridRowEnd: 3, // Spans rows 0-2
        MinHeight: 300.0, // Distributed: 100px per row
    },
}

// Items in individual rows
item2 := &layout.Node{
    Style: layout.Style{
        GridRowStart: 0, GridRowEnd: 1, // Row 0
        MinHeight: 150.0, // Row 0 needs 150px (max of 100 and 150)
    },
}

// Result:
// - Row 0: 150px (max of 100 from item1, 150 from item2)
// - Row 1: 100px (from item1)
// - Row 2: 100px (from item1)
// - item1 final height: 150 + 8 + 100 + 8 + 100 = 366px (with 8px gaps)
//   Note: Spanning items always fill their cell, even in auto rows
```

### Important Notes

- **MinHeight is still required** for spanning items in auto rows
- The spanning item's height may be **larger** than its MinHeight if other items in those rows require more space
- Row gaps are included in the spanning item's total height

## Fixed Rows vs Auto Rows

### Fixed Rows

Items in **fixed-size rows** will stretch to fill the cell height (CSS Grid default `align-items: stretch` behavior):

```go
gridRows := []layout.GridTrack{
    layout.FixedTrack(100), // Fixed 100px height
}
// Items in this row will be 100px tall (minus margins)
// Spanning items will be: (rowHeight * spanRows) + (gap * (spanRows - 1))
```

### Auto Rows

Items in **auto-sized rows** use their intrinsic size (content + MinHeight):

```go
gridRows := []layout.GridTrack{
    layout.AutoTrack(), // Auto-sized based on content
}
// Items in this row will be their measured height (content + MinHeight)
// If no content and no MinHeight, items will be 0 height
// Spanning items distribute their height across the spanned rows
```

## Margin and Padding

### Margins in Flexbox/Grid

Margins are fully supported and work as expected:

```go
item := layout.Margin(layout.Fixed(100, 50), 10)
// Item will have 10px margin on all sides
```

### Margins Don't Collapse

Unlike block layout, margins **don't collapse** in Flexbox and Grid (CSS-compliant behavior):

```go
// These margins won't collapse - you'll get 20px total spacing
item1 := layout.Margin(layout.Fixed(100, 50), 10)
item2 := layout.Margin(layout.Fixed(100, 50), 10)
```

## Grid Auto-Placement

### Default Behavior

If you don't set `GridRowStart` and `GridColumnStart`, items are auto-placed sequentially:

```go
grid := layout.Grid(2, 2, 100, 100)
grid.Children = []*layout.Node{
    {}, // Auto-placed at row 0, col 0
    {}, // Auto-placed at row 0, col 1
    {}, // Auto-placed at row 1, col 0
    {}, // Auto-placed at row 1, col 1
}
```

### Explicit Placement

To place items explicitly, set both `GridRowStart` and `GridRowEnd`:

```go
item := &layout.Node{
    Style: layout.Style{
        GridRowStart: 0,
        GridRowEnd: 1, // Must set both!
        GridColumnStart: 0,
        GridColumnEnd: 1,
    },
}
```

## Constraint Handling

### Unbounded Constraints

Use `layout.Unbounded` for height when you want auto-sized rows:

```go
constraints := layout.Loose(width, layout.Unbounded)
// This allows rows to grow based on content
```

### Tight Constraints

If constraints are too small, items may be clipped:

```go
constraints := layout.Tight(100, 100) // Very small!
// Items may overflow or be clipped
```

## Performance Considerations

### Nested Grids

Deeply nested grids can be expensive. Consider flattening when possible:

```go
// ❌ Deep nesting
grid1 := layout.Grid(2, 2, 100, 100)
grid2 := layout.Grid(2, 2, 50, 50)
grid1.Children = []*layout.Node{grid2}

// ✅ Flatter structure when possible
```

### Large Grids

For very large grids (100+ items), consider pagination or virtualization.

## Debugging Tips

### Items Not Appearing

1. Check if `Layout()` was called
2. Verify constraints are large enough
3. Check if items have `MinHeight` set (for auto rows)
4. Verify `GridRowStart`/`GridColumnStart` are set correctly

### Rows Collapsing

1. Ensure items in auto rows have `MinHeight` set
2. Check that items have content or explicit height
3. Verify row gaps aren't causing issues

### Items Overlapping

1. Check grid positioning (`GridRowStart`, `GridColumnStart`)
2. Verify row/column spans are correct
3. Check margins aren't causing negative sizes

## Debugging Grid Layout Issues

### Checklist for Row Overlapping/Collapsing

If your grid rows are overlapping or collapsing, check:

1. **Did you call `Layout()`?**
   ```go
   // ❌ Missing Layout() call
   root := layout.Grid(2, 2, 100, 100)
   fmt.Println(root.Children[0].Rect.Y) // Will be 0!
   
   // ✅ Correct
   root := layout.Grid(2, 2, 100, 100)
   layout.Layout(root, constraints)
   fmt.Println(root.Children[0].Rect.Y) // Will be correct
   ```

2. **Do all items have `MinHeight` or `Height` set?**
   ```go
   // Check your items
   for i, item := range grid.Children {
       if item.Style.MinHeight == 0 && item.Style.Height <= 0 {
           fmt.Printf("WARNING: Item %d has no MinHeight or Height!\n", i)
       }
   }
   ```

3. **Are constraints large enough?**
   ```go
   // Use Unbounded for auto rows
   constraints := layout.Loose(width, layout.Unbounded)
   ```

4. **Are row/column positions set correctly?**
   ```go
   // Verify GridRowStart/GridColumnStart are set
   for i, item := range grid.Children {
       fmt.Printf("Item %d: row %d-%d, col %d-%d\n",
           i, item.Style.GridRowStart, item.Style.GridRowEnd,
           item.Style.GridColumnStart, item.Style.GridColumnEnd)
   }
   ```

5. **Check the final layout results:**
   ```go
   layout.Layout(root, constraints)
   fmt.Printf("Root: %.2f x %.2f\n", root.Rect.Width, root.Rect.Height)
   for i, child := range root.Children {
       fmt.Printf("Item %d: y=%.2f, h=%.2f\n", i, child.Rect.Y, child.Rect.Height)
   }
   ```

