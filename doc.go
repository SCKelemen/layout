// Package layout provides a pure Go implementation of CSS Grid, Flexbox, Block,
// positioned, and text layout.
//
// This library implements layout algorithms similar to CSS, allowing you to create complex
// layouts programmatically in Go. It's designed to be reusable across different rendering
// backends: terminal UIs (Bubble Tea), web layouts, SVG rendering, PDF generation, etc.
//
// # Layout Systems
//
// The library supports multiple layout systems, selected by Style.Display:
//
//   - Flexbox: Flexible box layout with support for direction, wrap, alignment, gaps,
//     order, and flex properties
//   - Grid: CSS Grid layout with template rows/columns, fractional units (fr),
//     repeat(auto-fill/auto-fit) via Style.GridTemplateColumnsRepeat, named areas,
//     auto-placement, gaps, and item alignment
//   - Block: Vertical flow with margin collapsing, the default for nodes without a
//     Display value
//   - Text: DisplayInlineText leaf nodes created with Text(); LayoutText breaks the
//     text into line boxes (Node.TextLayout) using a pluggable TextMetricsProvider
//   - Positioned: Absolute, relative, fixed, and sticky positioning through
//     LayoutWithPositioning
//
// # Quick Start
//
// Create a simple horizontal stack:
//
//	root := layout.HStack(
//	    layout.Fixed(100, 50),
//	    layout.Spacer(),
//	    layout.Fixed(100, 50),
//	)
//	constraints := layout.Loose(800, 600)
//	size := layout.LayoutSimple(root, constraints)
//
// LayoutSimple derives a LayoutContext from the constraints. Pass an explicit
// context when you need control over the viewport, root font size, or text
// metrics:
//
//	ctx := layout.NewLayoutContext(800, 600, 16)
//	size := layout.Layout(root, constraints, ctx)
//
// # Lengths
//
// Sizes, offsets, gaps, and spacing are Length values (Px, Em, Rem, Vw, ...).
// A Length that was never set (the zero value) means the CSS initial value
// "auto"; Px(0) is an explicit zero. Relative units are resolved through the
// LayoutContext. Container-query units (cqw, cqh, ...) resolve to 0 in the
// normal layout pass; use ResolveLengthInContext with a NodeContext to resolve
// them against a Style.ContainerType ancestor.
//
// # Usage Patterns
//
// The library supports multiple usage patterns:
//
//  1. High-level API: Use HStack, VStack, Spacer, Text for simple layouts
//  2. CSS-like API: Direct Node creation with Style properties
//  3. Fluent API: Immutable With*, Find*, Transform, Map, Filter, FoldNodes methods
//  4. Embedded pattern: Embed Node in your own types
//
// See docs/usage-patterns.md and docs/fluent-api.md for detailed examples of
// each pattern.
//
// # Serialization
//
// The serialize subpackage round-trips trees as JSON or YAML with
// unit-preserving lengths:
//
//	data, err := serialize.ToJSON(root)
//
// # SVG Rendering
//
// Nodes carry an optional 2D Transform and helpers for SVG output:
//
//	node.Style.Transform = layout.RotateDegrees(15)
//	transform := node.Style.Transform.ToSVGString()
//	rect := layout.GetFinalRect(node)
//
// # Examples
//
// See the examples/ directory for complete working examples.
package layout
