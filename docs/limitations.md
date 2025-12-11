# Limitations and Design Decisions

This document outlines known limitations and design decisions for the layout library.

## Known Limitations

### Margin Support

**Status**: ✅ **Margin is now fully supported** in Flexbox and Grid layouts!

**Implementation**:
- **Flexbox**: Margins are accounted for in item positioning and spacing
- **Grid**: Margins are applied within grid cells, reducing the available space for items
- **Block**: Margin support is not yet implemented (use padding instead)

**Usage**:
```go
// Add margins to items in HStack/VStack
item := layout.Fixed(100, 50)
item.Style.Margin = layout.Uniform(10) // 10px margin on all sides

// Or use the Margin helper
layout.Margin(item, 10)
```

**Note**: Margins don't collapse in Flexbox or Grid (unlike block layout), which matches CSS behavior.

### Box-Sizing

**Status**: The `BoxSizing` field exists but is not currently used in calculations.

**Current behavior**: All layouts treat sizing as `content-box` (width/height = content size only).

**Impact**: If you need `border-box` behavior (width/height includes padding + border), you'll need to account for this manually.

### Inline Layout

**Status**: Not implemented.

**What's missing**:
- Text flow and line breaking
- Inline elements
- Baseline alignment

**Impact**: This library focuses on block-level layouts. For text layout, you'll need to handle text measurement and line breaking separately.

**Workaround**: Use fixed-size containers for text, or measure text separately and use those measurements for layout.

### Table Layout

**Status**: Not implemented.

**Impact**: For table-like layouts, use Grid instead. Grid provides all the functionality needed for tabular data.

### CSS Multi-column Layout

**Status**: Not implemented.

**Note**: This is **different** from Grid columns! CSS Multi-column Layout is for flowing text into columns (like a newspaper), not for grid-based layouts.

**Impact**: If you need text flowing into multiple columns, you'll need to handle this separately.

**Clarification**: Grid **does** support multiple columns via `GridTemplateColumns`. See [Layout Systems](layout-systems.md) for details.

### Sticky Positioning

**Status**: Partially implemented.

**Current behavior**: Sticky positioning is defined but may not fully match CSS behavior in all cases.

**Impact**: For most use cases, `PositionAbsolute` or `PositionFixed` will work better.

## Design Decisions

### Block Layout is Minimal

Block layout is intentionally kept simple because:
1. It's primarily a fallback for non-flex/grid elements
2. Grid and Flexbox are the primary layout systems
3. Most use cases don't need full CSS block layout features

### High-Level API vs CSS-like API

The library provides both:
- **High-level API**: Easier to use, SwiftUI/Flutter-like
- **CSS-like API**: More control, lower-level

**Decision**: Keep both to serve different needs:
- Simple layouts → High-level API
- Complex layouts → CSS-like API

### No Built-in Text Layout

**Decision**: Focus on layout, not text rendering.

**Rationale**: Text layout is complex and depends on fonts, rendering engines, etc. This library focuses on spatial layout, leaving text to rendering code.

### Transform Support

**Status**: Full transform support is implemented.

**Decision**: Include transforms because they're useful for SVG rendering and visual effects.

## Test Coverage

- **68+ tests** covering all major features
- All tests passing
- Good coverage of edge cases

## Compatibility

### CSS Compatibility

This library is **not** a full CSS implementation. It implements:
- ✅ Core layout algorithms (Flexbox, Grid, Block)
- ✅ Positioning (absolute, relative, fixed)
- ✅ Transforms
- ⚠️ Partial: Margin, box-sizing
- ❌ Not implemented: Inline layout, table layout, multi-column text

### Use Case Compatibility

**Comprehensive for**:
- ✅ Terminal UIs (Bubble Tea)
- ✅ SVG rendering (card layouts, graphs)
- ✅ Web layouts (server-side)
- ✅ PDF generation
- ✅ Game UIs
- ✅ Offscreen rendering
- ✅ Image generation

**Not comprehensive for**:
- ⚠️ Full CSS compatibility
- ⚠️ Text-heavy layouts
- ⚠️ Complex inline formatting

## Recommendations

1. **Use Grid or Flexbox** for most layouts
2. **Use Grid gaps** instead of margins for spacing
3. **Handle text separately** - measure text, then use measurements for layout
4. **Use positioned layout** for overlapping elements
5. **Check examples** in `examples/` directory for patterns

## Future Considerations

Potential future additions (not currently planned):
- Margin support in block layout
- Box-sizing support
- Improved sticky positioning
- Text measurement helpers (separate from layout)

If you need features that aren't implemented, please open an issue or contribute!

