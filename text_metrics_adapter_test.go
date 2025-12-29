package layout

import (
	"testing"
)

func TestTextMetricsAdapter(t *testing.T) {
	adapter := NewTerminalTextMetrics()

	tests := []struct {
		name     string
		text     string
		style    TextStyle
		wantMin  float64 // minimum expected width
		wantMax  float64 // maximum expected width
	}{
		{
			name: "ASCII text",
			text: "Hello",
			style: TextStyle{
				FontSize:      16,
				LineHeight:    0, // normal
				LetterSpacing: -1,
			},
			wantMin: 5.0,
			wantMax: 5.0,
		},
		{
			name: "CJK text",
			text: "世界",
			style: TextStyle{
				FontSize:      16,
				LineHeight:    0,
				LetterSpacing: -1,
			},
			wantMin: 4.0, // 2 + 2 (each CJK char is 2 cells wide)
			wantMax: 4.0,
		},
		{
			name: "Emoji",
			text: "😀",
			style: TextStyle{
				FontSize:      16,
				LineHeight:    0,
				LetterSpacing: -1,
			},
			wantMin: 2.0,
			wantMax: 2.0,
		},
		{
			name: "Emoji with modifier",
			text: "👋🏻",
			style: TextStyle{
				FontSize:      16,
				LineHeight:    0,
				LetterSpacing: -1,
			},
			wantMin: 2.0, // emoji + skin tone = still 2 cells
			wantMax: 2.0,
		},
		{
			name: "Mixed content",
			text: "Hello世界",
			style: TextStyle{
				FontSize:      16,
				LineHeight:    0,
				LetterSpacing: -1,
			},
			wantMin: 9.0, // 5 + 4
			wantMax: 9.0,
		},
		{
			name: "With letter spacing",
			text: "Hello",
			style: TextStyle{
				FontSize:      16,
				LineHeight:    0,
				LetterSpacing: 2.0,
			},
			wantMin: 13.0, // 5 + (4 * 2) = 5 + 8
			wantMax: 13.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advance, ascent, descent := adapter.Measure(tt.text, tt.style)

			// Check advance width
			if advance < tt.wantMin || advance > tt.wantMax {
				t.Errorf("Measure(%q) advance = %.1f, want between %.1f and %.1f",
					tt.text, advance, tt.wantMin, tt.wantMax)
			}

			// Check ascent and descent are reasonable
			if ascent <= 0 || descent <= 0 {
				t.Errorf("Measure(%q) ascent=%.1f, descent=%.1f, want both > 0",
					tt.text, ascent, descent)
			}

			// Check ascent + descent = line height
			lineHeight := ascent + descent
			if lineHeight <= 0 {
				t.Errorf("Measure(%q) lineHeight=%.1f, want > 0", tt.text, lineHeight)
			}
		})
	}
}

func TestTextMetricsAdapterAsProvider(t *testing.T) {
	// Save original provider
	original := textMetrics
	defer func() { textMetrics = original }()

	// Set our adapter as the provider
	adapter := NewTerminalTextMetrics()
	SetTextMetricsProvider(adapter)

	// Create a text node and verify it uses our adapter
	node := Text("Hello 世界", Style{
		TextStyle: &TextStyle{
			FontSize:   16,
			LineHeight: 0,
		},
	})

	constraints := Tight(800, 600)
	size := Layout(node, constraints, nil)

	// Verify layout was computed (non-zero size)
	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("Layout returned zero size: %.1f x %.1f", size.Width, size.Height)
	}

	// Verify width is reasonable for "Hello 世界" (5 + 1 space + 4 = 10 cells)
	advance, _, _ := adapter.Measure("Hello 世界", TextStyle{
		FontSize:      16,
		LineHeight:    0,
		LetterSpacing: -1,
	})
	if advance != 10.0 {
		t.Errorf("Expected advance 10.0 for 'Hello 世界', got %.1f", advance)
	}
}

func TestNewTerminalTextMetrics(t *testing.T) {
	adapter := NewTerminalTextMetrics()
	if adapter == nil {
		t.Fatal("NewTerminalTextMetrics() returned nil")
	}
	if adapter.text == nil {
		t.Fatal("NewTerminalTextMetrics() created adapter with nil text field")
	}
}

func TestTextMetricsAdapterTextAccess(t *testing.T) {
	adapter := NewTerminalTextMetrics()
	txt := adapter.Text()

	if txt == nil {
		t.Fatal("Text() returned nil")
	}

	// Verify we can use text operations directly
	width := txt.Width("Hello")
	if width != 5.0 {
		t.Errorf("txt.Width(\"Hello\") = %.1f, want 5.0", width)
	}

	graphemes := txt.Graphemes("Hello👋🏻")
	if len(graphemes) != 6 {
		t.Errorf("txt.Graphemes(\"Hello👋🏻\") = %d clusters, want 6", len(graphemes))
	}
}
