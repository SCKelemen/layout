package layout

// LayoutContext carries information needed to resolve relative length units to pixels.
//
// This context is provided by callers and contains:
//   - Viewport dimensions (for vh/vw units)
//   - Root font size (for rem units)
//   - Text metrics provider (for ch units)
//   - Reference character for ch unit (default: '0')
//
// The context is passed through layout algorithms to enable unit resolution
// at any point where sizing or spacing calculations occur.
type LayoutContext struct {
	// ViewportWidth is the width of the viewport in pixels.
	// Used to resolve vw (viewport width) units: 1vw = 1% of ViewportWidth.
	ViewportWidth float64

	// ViewportHeight is the height of the viewport in pixels.
	// Used to resolve vh (viewport height) units: 1vh = 1% of ViewportHeight.
	ViewportHeight float64

	// RootFontSize is the root element's font size in points.
	// Used to resolve rem (root em) units: 1rem = RootFontSize.
	RootFontSize float64

	// TextMetrics is the text measurement provider used to measure character widths.
	// Used to resolve ch units by measuring the reference character.
	// If nil, a monospace approximation is used (60% of font size).
	TextMetrics TextMetricsProvider

	// ChReferenceChar is the reference character for ch unit calculations.
	// Per CSS spec, this is typically '0' (U+0030 DIGIT ZERO).
	// Default: '0'
	ChReferenceChar rune
}

// NewLayoutContext creates a new LayoutContext with the specified parameters
// and sensible defaults.
//
// Parameters:
//   - viewportWidth: Width of the viewport in pixels
//   - viewportHeight: Height of the viewport in pixels
//   - rootFontSize: Root font size in points (typical values: 12-16 points)
//
// Returns a LayoutContext with:
//   - Viewport dimensions set to the provided values
//   - RootFontSize set to the provided value
//   - TextMetrics using the package-level text metrics provider
//   - ChReferenceChar set to '0' (CSS standard)
//
// Example:
//
//	// Create context for a 1920x1080 viewport with 16pt root font
//	ctx := layout.NewLayoutContext(1920, 1080, 16)
//
//	// Use in layout
//	layout.Layout(node, constraints, ctx)
func NewLayoutContext(viewportWidth, viewportHeight, rootFontSize float64) *LayoutContext {
	return &LayoutContext{
		ViewportWidth:   viewportWidth,
		ViewportHeight:  viewportHeight,
		RootFontSize:    rootFontSize,
		TextMetrics:     textMetrics, // Use package-level provider
		ChReferenceChar: '0',         // CSS standard reference character
	}
}

// WithTextMetrics returns a copy of the context with a custom TextMetricsProvider.
// This allows callers to provide their own text measurement implementation
// (e.g., HarfBuzz, FreeType) for accurate ch unit calculations.
//
// Example:
//
//	ctx := layout.NewLayoutContext(1920, 1080, 16)
//	ctx = ctx.WithTextMetrics(myHarfBuzzMetrics)
func (ctx *LayoutContext) WithTextMetrics(metrics TextMetricsProvider) *LayoutContext {
	copy := *ctx
	copy.TextMetrics = metrics
	return &copy
}

// WithChReferenceChar returns a copy of the context with a custom reference character
// for ch unit calculations.
//
// The CSS spec uses '0' (DIGIT ZERO) as the standard reference, but some use cases
// may want to use a different character (e.g., 'M' for em-quad approximation).
//
// Example:
//
//	ctx := layout.NewLayoutContext(1920, 1080, 16)
//	ctx = ctx.WithChReferenceChar('M') // Use 'M' width instead of '0'
func (ctx *LayoutContext) WithChReferenceChar(char rune) *LayoutContext {
	copy := *ctx
	copy.ChReferenceChar = char
	return &copy
}
