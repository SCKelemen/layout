// Package layout provides a pure Go implementation of CSS Grid, Flexbox, and Block layout engines.
//
// This library implements layout algorithms similar to CSS, allowing you to create complex
// layouts programmatically in Go. It's designed to be reusable across different rendering
// backends: terminal UIs (Bubble Tea), web layouts, SVG rendering, PDF generation, etc.
//
// # Layout Systems
//
// The library supports multiple layout systems:
//
//   - Flexbox: Flexible box layout with support for direction, wrap, alignment, and flex properties
//   - Grid: CSS Grid layout with support for unlimited columns via GridTemplateColumns,
//     template rows, fractional units (fr), gaps, and item positioning
//   - Block: Basic block layout for stacking elements vertically
//   - Positioned: Absolute, relative, fixed, and sticky positioning
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
//	size := layout.Layout(root, constraints)
//
// # Usage Patterns
//
// The library supports multiple usage patterns:
//
//  1. High-level API: Use HStack, VStack, Spacer for simple layouts
//  2. CSS-like API: Direct Node creation with Style properties
//  3. Embedded pattern: Embed Node in your own types
//  4. Builder pattern: Create custom builders for your domain
//
// See USAGE.md for detailed examples of each pattern.
//
// # SVG Rendering
//
// The library includes helpers for SVG rendering:
//
//	transform := layout.GetSVGTransform(node)
//	rect := layout.GetFinalRect(node)
//
// # Transforms
//
// Support for 2D transformations (translate, rotate, scale, skew) for visual effects:
//
//	node.Style.Transform = layout.RotateDegrees(15)
//
// # Examples
//
// See the examples/ directory for complete working examples.
package layout
