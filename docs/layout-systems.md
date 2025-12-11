# Layout Systems

This library supports multiple layout systems, each suited for different use cases.

## Flexbox

Flexbox is ideal for one-dimensional layouts (rows or columns) with flexible sizing.

**Specification**: [CSS Flexible Box Layout Module Level 1](https://www.w3.org/TR/css-flexbox-1/)

### Features

- **Flex Direction**: Row, column, and reverse variants
- **Flex Wrap**: Control whether items wrap to new lines
- **Justify Content**: Align items along the main axis (start, end, center, space-between, space-around, space-evenly)
- **Align Items**: Align items along the cross axis (start, end, center, stretch, baseline)
- **Align Content**: Align wrapped lines
- **Flex Properties**: Grow, shrink, and basis for flexible sizing

### Example

```go
root := &layout.Node{
    Style: layout.Style{
        Display:        layout.DisplayFlex,
        FlexDirection:  layout.FlexDirectionColumn,
        JustifyContent: layout.JustifyContentCenter,
        AlignItems:     layout.AlignItemsStretch,
        Padding:        layout.Uniform(20),
    },
    Children: []*layout.Node{
        {
            Style: layout.Style{
                FlexGrow: 1,
                Height:   100,
            },
        },
        {
            Style: layout.Style{
                FlexGrow: 2,
                Height:   100,
            },
        },
    },
}
```

### Flex Properties

- `FlexGrow`: How much the item should grow relative to siblings (default: 0)
- `FlexShrink`: How much the item should shrink (default: 1)
- `FlexBasis`: Initial size before growing/shrinking (default: auto, uses measured size)

## Grid

Grid is perfect for two-dimensional layouts with precise control over rows and columns.

**Specification**: [CSS Grid Layout Module Level 1](https://www.w3.org/TR/css-grid-1/)

### Features

- **Multiple Columns**: Define any number of columns via `GridTemplateColumns`
- **Template Rows/Columns**: Fixed, fractional (fr), or auto-sized tracks
- **Auto Rows/Columns**: Default size for implicit tracks
- **Grid Gaps**: Spacing between grid items
- **Item Positioning**: Place items in specific cells or span multiple cells
- **Bento Box / Mosaic Layouts**: Create Pinterest-style or bento box layouts with items spanning different numbers of rows/columns

### Example

```go
root := &layout.Node{
    Style: layout.Style{
        Display: layout.DisplayGrid,
        GridTemplateRows: []layout.GridTrack{
            layout.FixedTrack(100),
            layout.FractionTrack(1),
        },
        GridTemplateColumns: []layout.GridTrack{
            layout.FixedTrack(200),      // Sidebar
            layout.FractionTrack(1),     // Main content
        },
        GridGap: 10,
    },
    Children: []*layout.Node{
        {
            Style: layout.Style{
                GridRowStart:    0,
                GridColumnStart: 0,
                GridColumnEnd:   2,  // Span both columns
            },
        },
        {
            Style: layout.Style{
                GridRowStart:    1,
                GridColumnStart: 0,
            },
        },
        {
            Style: layout.Style{
                GridRowStart:    1,
                GridColumnStart: 1,
            },
        },
    },
}
```

### Grid Tracks

- `FixedTrack(size)`: Fixed-size track
- `FractionTrack(fraction)`: Fractional unit (fr), shares available space proportionally
- `MinMaxTrack(min, max)`: Track with min/max constraints
- `AutoTrack()`: Auto-sized based on content

### Multi-Column Grids

Grid supports unlimited columns. Just add more tracks:

```go
GridTemplateColumns: []layout.GridTrack{
    layout.FixedTrack(150),  // Column 1
    layout.FixedTrack(150),  // Column 2
    layout.FixedTrack(150),  // Column 3
    // ... as many as you need
}
```

See `examples/multi_column/main.go` for a complete example.

### Grid Helper Functions

For simpler grid creation, use the helper functions:

```go
// Create a grid with fixed track sizes
grid := layout.Grid(4, 4, 150, 200) // 4 rows x 4 columns, rows=150px, cols=200px

// Create a grid with auto-sized tracks
grid := layout.GridAuto(3, 4) // 3 rows x 4 columns, auto-sized

// Create a grid with fractional tracks
grid := layout.GridFractional(2, 3) // 2 rows x 3 columns, equal fractional units
```

This is much simpler than manually creating `GridTemplateRows` and `GridTemplateColumns` arrays!

### Bento Box / Mosaic Layouts

Grid's spanning capabilities make it perfect for creating bento box or mosaic-style layouts where items have different sizes:

```go
root := &layout.Node{
    Style: layout.Style{
        Display: layout.DisplayGrid,
        GridTemplateRows: []layout.GridTrack{
            layout.FixedTrack(150),
            layout.FixedTrack(150),
            layout.FixedTrack(150),
        },
        GridTemplateColumns: []layout.GridTrack{
            layout.FixedTrack(200),
            layout.FixedTrack(200),
            layout.FixedTrack(200),
        },
        GridGap: 10,
    },
    Children: []*layout.Node{
        // Large item spanning 2x2
        {
            Style: layout.Style{
                GridRowStart:    0,
                GridRowEnd:      2,  // Spans 2 rows
                GridColumnStart: 0,
                GridColumnEnd:   2,  // Spans 2 columns
            },
        },
        // Medium item spanning 1x2
        {
            Style: layout.Style{
                GridRowStart:    0,
                GridColumnStart: 2,
                GridColumnEnd:   4,  // Spans 2 columns
            },
        },
        // Small 1x1 items
        {
            Style: layout.Style{
                GridRowStart:    2,
                GridColumnStart: 0,
            },
        },
    },
}
```

See `examples/bento/main.go` for a complete bento box layout example.

## Block Layout

Block layout is a simple vertical stacking layout, used as a fallback for non-flex/grid elements.

### Features

- **Vertical Stacking**: Children are stacked vertically
- **Auto Sizing**: Width/height can be auto (based on content)
- **Padding Support**: Padding is correctly calculated
- **Min/Max Constraints**: Respects min/max width/height

### When to Use

Block layout is primarily used as a fallback when children don't have a specific display type. For most use cases, prefer Flexbox or Grid.

## Positioned Layout

Positioned layout handles absolute, relative, fixed, and sticky positioning.

### Position Types

- **Static**: Default positioning (normal flow)
- **Relative**: Offset from normal flow position
- **Absolute**: Positioned relative to nearest positioned ancestor
- **Fixed**: Positioned relative to viewport
- **Sticky**: Sticks to viewport when scrolling (not fully implemented)

### Example

```go
root := &layout.Node{
    Style: layout.Style{
        Position: layout.PositionRelative,
    },
    Children: []*layout.Node{
        {
            Style: layout.Style{
                Position: layout.PositionAbsolute,
                Left:     50,
                Top:      50,
                Width:    100,
                Height:   100,
            },
        },
    },
}

// Use LayoutWithPositioning for positioned elements
constraints := layout.Loose(800, 600)
layout.LayoutWithPositioning(root, constraints, layout.Rect{Width: 800, Height: 600})
```

### Offsets

- `Top`, `Right`, `Bottom`, `Left`: Offset values (use -1 for auto)
- `ZIndex`: Stacking order (higher values appear on top)

## Choosing a Layout System

- **Flexbox**: One-dimensional layouts (rows or columns), flexible sizing
- **Grid**: Two-dimensional layouts, precise control, card layouts
- **Block**: Simple vertical stacking (fallback)
- **Positioned**: Overlapping elements, absolute positioning

For most use cases, **Grid** or **Flexbox** will be your primary choice.

