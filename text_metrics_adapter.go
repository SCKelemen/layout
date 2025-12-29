package layout

import (
	"github.com/SCKelemen/text"
)

// TextMetricsAdapter adapts the github.com/SCKelemen/text library
// to implement the layout.TextMetricsProvider interface.
//
// This provides accurate Unicode-aware text measurement using:
// - UAX #29 (grapheme clustering) for proper emoji and combining character support
// - UAX #14 (line breaking) for proper line break opportunities
// - UTS #51 (emoji sequences) for accurate emoji width calculation
//
// Example:
//
//	// Create adapter for terminal rendering
//	adapter := layout.NewTerminalTextMetrics()
//	layout.SetTextMetricsProvider(adapter)
//
//	// Now all text layout uses accurate Unicode measurements
//	node := layout.Text("Hello 世界 😀", layout.TextStyle{
//	    FontSize: 16,
//	})
//	layout.Layout(node, layout.Loose(800, 600))
type TextMetricsAdapter struct {
	text *text.Text
}

// NewTextMetricsAdapter creates a new adapter with the given text configuration.
//
// For terminal rendering, use NewTerminalTextMetrics() instead.
//
// Example:
//
//	// Custom configuration
//	adapter := layout.NewTextMetricsAdapter(text.Config{
//	    MeasureFunc: text.TerminalMeasure,
//	})
func NewTextMetricsAdapter(config text.Config) *TextMetricsAdapter {
	return &TextMetricsAdapter{
		text: text.New(config),
	}
}

// NewTerminalTextMetrics creates a text metrics adapter configured for
// terminal rendering using Unicode East Asian width properties.
//
// This is the recommended default for terminal UIs and applications that
// need accurate character cell width calculations.
//
// Example:
//
//	metrics := layout.NewTerminalTextMetrics()
//	layout.SetTextMetricsProvider(metrics)
func NewTerminalTextMetrics() *TextMetricsAdapter {
	return &TextMetricsAdapter{
		text: text.NewTerminal(),
	}
}

// Measure implements layout.TextMetricsProvider.
//
// Returns:
//   - advance: The display width of the text (in terminal cells for terminal config)
//   - ascent: The distance above the baseline (80% of line height)
//   - descent: The distance below the baseline (20% of line height)
//
// The width calculation uses:
// - Grapheme cluster boundaries (UAX #29) for proper emoji/combining char handling
// - Emoji sequence width (UTS #51) for accurate emoji measurements
// - East Asian width properties for CJK characters
func (a *TextMetricsAdapter) Measure(textContent string, style TextStyle) (advance, ascent, descent float64) {
	// Calculate advance width using the text library
	advance = a.text.Width(textContent)

	// Apply letter spacing if specified
	// Letter spacing applies between characters (not after last one)
	if style.LetterSpacing != -1 {
		graphemeCount := a.text.GraphemeCount(textContent)
		if graphemeCount > 0 {
			advance += float64(graphemeCount-1) * style.LetterSpacing
		}
	}

	// Calculate line height for ascent/descent
	lineHeight := style.LineHeight
	if lineHeight == 0 {
		// Default line height based on font size
		lineHeight = style.FontSize * 1.2
	} else if lineHeight < 10 {
		// Heuristic: < 10 is a multiplier
		lineHeight = style.FontSize * lineHeight
	}
	// else: >= 10 is absolute pixels

	// Standard proportions: 80% ascent, 20% descent
	ascent = lineHeight * 0.8
	descent = lineHeight * 0.2

	return
}

// Text returns the underlying text.Text instance for direct access
// to text operations like Wrap, Truncate, Align, etc.
//
// Example:
//
//	adapter := layout.NewTerminalTextMetrics()
//	txt := adapter.Text()
//
//	// Use text library operations directly
//	wrapped := txt.Wrap("Long text...", text.WrapOptions{
//	    MaxWidth: 40,
//	})
func (a *TextMetricsAdapter) Text() *text.Text {
	return a.text
}
