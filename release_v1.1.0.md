# Release v1.1.0

## 🎉 Major Feature: CSS Length Type System

This release introduces a comprehensive CSS length type system, migrating from raw `float64` pixel values to a proper `Length` type that supports multiple CSS units. This is a significant step toward full CSS specification compliance.

### ✨ New Features

#### CSS Length Type System
- **New `Length` type**: Represents CSS `<length>` values with support for multiple units
- **Unit constructors**: `Px()`, `Em()`, `Rem()`, `Percent()`, and more for creating length values
- **Length resolution**: `ResolveLength()` function converts lengths to pixels based on layout context
- **LayoutContext**: New context type that provides viewport dimensions and font size for length resolution

#### Enhanced API
- All `Style` properties (Width, Height, Padding, Margin, Border, etc.) now use `Length` type
- `GridTrack` sizing now uses `Length` for `MinSize` and `MaxSize`
- `Spacing` struct (Padding, Margin, Border) now uses `Length` for all sides
- All layout functions now require `*LayoutContext` parameter for proper length resolution

### 🔧 Breaking Changes

⚠️ **This is a breaking change** - code using raw `float64` values will need to be updated:

**Before:**
```go
node.Style.Width = 100
node.Style.Padding = Uniform(10)
```

**After:**
```go
node.Style.Width = Px(100)
node.Style.Padding = Uniform(Px(10))
```

**Layout function calls now require LayoutContext:**
```go
ctx := NewLayoutContext(800, 600, 16) // viewportWidth, viewportHeight, rootFontSize
Layout(root, constraints, ctx)
```

### 📦 Additional Improvements

- **WPT Test Integration**: Enhanced integration with Web Platform Tests using CEL (Common Expression Language) assertions
- **Fluent API**: Comprehensive fluent API for building layout trees with method chaining
- **Test Coverage**: All tests updated and passing with new Length type system
- **Documentation**: Updated examples and documentation to reflect new API

### 🐛 Bug Fixes

- Fixed test logic after Length type migration
- Corrected align, snap, transform, and SVG tests
- Fixed serialize package to properly handle Length types
- Updated all examples to use new Length API

### 📚 Migration Guide

1. Replace all numeric literals in `Style` properties with `Px()` constructor:
   ```go
   Width: 100 → Width: Px(100)
   ```

2. Add `LayoutContext` to all layout function calls:
   ```go
   ctx := NewLayoutContext(800, 600, 16)
   Layout(root, constraints, ctx)
   ```

3. Access `.Value` when comparing or doing arithmetic with Length types:
   ```go
   if node.Style.Width.Value > 100 { ... }
   ```

4. Use `.Value` when converting Length to float64 for serialization or external APIs

### 🔗 Related Changes

- Updated `wpt-test-gen` integration to use `.Value` for Length accesses
- Serialize package now properly converts between Length and float64
- All examples updated to demonstrate new Length API

### 📝 Full Changelog

See git log for complete list of commits:
```bash
git log v1.0.0..v1.1.0
```

### 🙏 Thanks

Thank you to all contributors and users who provided feedback during this migration!

---

**Upgrade Instructions:**
```bash
go get github.com/SCKelemen/layout@v1.1.0
```
