package layout

import (
	"testing"
)

// TestFontStyle tests font-style property
func TestFontStyle(t *testing.T) {
	tests := []struct {
		name  string
		style FontStyle
		want  string
	}{
		{"Normal", FontStyleNormal, "normal"},
		{"Italic", FontStyleItalic, "italic"},
		{"Oblique", FontStyleOblique, "oblique"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{
				Text: "Test text",
				Style: Style{
					Display: DisplayInlineText,
					TextStyle: &TextStyle{
						FontSize:  16,
						FontStyle: tt.style,
					},
				},
			}

			ctx := NewLayoutContext(800, 600, 16)
			constraints := Loose(400, 300)
			_ = LayoutText(node, constraints, ctx)

			// Font style is metadata for renderer - layout doesn't change
			// Just verify it's preserved in the node
			if node.Style.TextStyle.FontStyle != tt.style {
				t.Errorf("FontStyle not preserved: got %v, want %v", node.Style.TextStyle.FontStyle, tt.style)
			}
		})
	}
}

// TestTextDecoration tests text-decoration property
func TestTextDecoration(t *testing.T) {
	tests := []struct {
		name       string
		decoration TextDecoration
		hasUnder   bool
		hasOver    bool
		hasThrough bool
	}{
		{"None", TextDecorationNone, false, false, false},
		{"Underline", TextDecorationUnderline, true, false, false},
		{"Overline", TextDecorationOverline, false, true, false},
		{"LineThrough", TextDecorationLineThrough, false, false, true},
		{"UnderlineAndOverline", TextDecorationUnderline | TextDecorationOverline, true, true, false},
		{"All", TextDecorationUnderline | TextDecorationOverline | TextDecorationLineThrough, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.decoration.Has(TextDecorationUnderline) != tt.hasUnder {
				t.Errorf("Has(Underline) = %v, want %v", tt.decoration.Has(TextDecorationUnderline), tt.hasUnder)
			}
			if tt.decoration.Has(TextDecorationOverline) != tt.hasOver {
				t.Errorf("Has(Overline) = %v, want %v", tt.decoration.Has(TextDecorationOverline), tt.hasOver)
			}
			if tt.decoration.Has(TextDecorationLineThrough) != tt.hasThrough {
				t.Errorf("Has(LineThrough) = %v, want %v", tt.decoration.Has(TextDecorationLineThrough), tt.hasThrough)
			}
		})
	}
}

// TestTextDecorationInLayout tests text decoration with layout
func TestTextDecorationInLayout(t *testing.T) {
	node := &Node{
		Text: "Decorated text",
		Style: Style{
			Display: DisplayInlineText,
			TextStyle: &TextStyle{
				FontSize:            16,
				TextDecoration:      TextDecorationUnderline | TextDecorationLineThrough,
				TextDecorationStyle: TextDecorationStyleWavy,
				TextDecorationColor: "red",
			},
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	constraints := Loose(400, 300)
	size := LayoutText(node, constraints, ctx)

	// Layout should succeed
	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("Layout failed: size = %v", size)
	}

	// Decoration properties should be preserved
	if !node.Style.TextStyle.TextDecoration.Has(TextDecorationUnderline) {
		t.Error("Underline decoration not preserved")
	}
	if !node.Style.TextStyle.TextDecoration.Has(TextDecorationLineThrough) {
		t.Error("LineThrough decoration not preserved")
	}
	if node.Style.TextStyle.TextDecorationStyle != TextDecorationStyleWavy {
		t.Errorf("DecorationStyle = %v, want %v", node.Style.TextStyle.TextDecorationStyle, TextDecorationStyleWavy)
	}
	if node.Style.TextStyle.TextDecorationColor != "red" {
		t.Errorf("DecorationColor = %q, want %q", node.Style.TextStyle.TextDecorationColor, "red")
	}
}

// TestTextDecorationStyle tests different decoration styles
func TestTextDecorationStyle(t *testing.T) {
	styles := []TextDecorationStyle{
		TextDecorationStyleSolid,
		TextDecorationStyleDouble,
		TextDecorationStyleDotted,
		TextDecorationStyleDashed,
		TextDecorationStyleWavy,
	}

	for _, style := range styles {
		node := &Node{
			Text: "Test",
			Style: Style{
				Display: DisplayInlineText,
				TextStyle: &TextStyle{
					FontSize:            16,
					TextDecoration:      TextDecorationUnderline,
					TextDecorationStyle: style,
				},
			},
		}

		ctx := NewLayoutContext(800, 600, 16)
		constraints := Loose(400, 300)
		_ = LayoutText(node, constraints, ctx)

		if node.Style.TextStyle.TextDecorationStyle != style {
			t.Errorf("DecorationStyle not preserved: got %v, want %v", node.Style.TextStyle.TextDecorationStyle, style)
		}
	}
}

// TestVerticalAlign tests vertical-align property
func TestVerticalAlign(t *testing.T) {
	alignments := []VerticalAlign{
		VerticalAlignBaseline,
		VerticalAlignSub,
		VerticalAlignSuper,
		VerticalAlignTextTop,
		VerticalAlignTextBottom,
		VerticalAlignMiddle,
		VerticalAlignTop,
		VerticalAlignBottom,
	}

	for _, align := range alignments {
		node := &Node{
			Text: "Test",
			Style: Style{
				Display: DisplayInlineText,
				TextStyle: &TextStyle{
					FontSize:      16,
					VerticalAlign: align,
				},
			},
		}

		ctx := NewLayoutContext(800, 600, 16)
		constraints := Loose(400, 300)
		_ = LayoutText(node, constraints, ctx)

		// Vertical align is metadata for renderer - layout doesn't change
		// Just verify it's preserved
		if node.Style.TextStyle.VerticalAlign != align {
			t.Errorf("VerticalAlign not preserved: got %v, want %v", node.Style.TextStyle.VerticalAlign, align)
		}
	}
}

// TestVerticalAlignWithDifferentBaselines tests vertical-align affecting baseline
func TestVerticalAlignWithDifferentBaselines(t *testing.T) {
	// Create text with subscript and superscript
	tests := []struct {
		name  string
		align VerticalAlign
		text  string
	}{
		{"Baseline", VerticalAlignBaseline, "Normal text"},
		{"Subscript", VerticalAlignSub, "H₂O"},
		{"Superscript", VerticalAlignSuper, "E=mc²"},
		{"Middle", VerticalAlignMiddle, "Middle aligned"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{
				Text: tt.text,
				Style: Style{
					Display: DisplayInlineText,
					TextStyle: &TextStyle{
						FontSize:      16,
						VerticalAlign: tt.align,
					},
				},
			}

			ctx := NewLayoutContext(800, 600, 16)
			constraints := Loose(400, 300)
			size := LayoutText(node, constraints, ctx)

			// Layout should succeed regardless of vertical-align
			if size.Width <= 0 || size.Height <= 0 {
				t.Errorf("Layout failed for %s: size = %v", tt.name, size)
			}

			// Verify vertical-align is preserved
			if node.Style.TextStyle.VerticalAlign != tt.align {
				t.Errorf("VerticalAlign = %v, want %v", node.Style.TextStyle.VerticalAlign, tt.align)
			}
		})
	}
}

// TestCombinedTextProperties tests using multiple text properties together
func TestCombinedTextProperties(t *testing.T) {
	node := &Node{
		Text: "Fancy styled text",
		Style: Style{
			Display: DisplayInlineText,
			TextStyle: &TextStyle{
				FontSize:            18,
				FontWeight:          FontWeightBold,
				FontStyle:           FontStyleItalic,
				TextDecoration:      TextDecorationUnderline,
				TextDecorationStyle: TextDecorationStyleDashed,
				TextDecorationColor: "blue",
				VerticalAlign:       VerticalAlignMiddle,
				TextAlign:           TextAlignCenter,
				LineHeight:          1.5,
			},
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	constraints := Loose(400, 300)
	size := LayoutText(node, constraints, ctx)

	// Layout should succeed
	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("Layout failed: size = %v", size)
	}

	// Verify all properties are preserved
	style := node.Style.TextStyle
	if style.FontWeight != FontWeightBold {
		t.Error("FontWeight not preserved")
	}
	if style.FontStyle != FontStyleItalic {
		t.Error("FontStyle not preserved")
	}
	if !style.TextDecoration.Has(TextDecorationUnderline) {
		t.Error("TextDecoration not preserved")
	}
	if style.TextDecorationStyle != TextDecorationStyleDashed {
		t.Error("TextDecorationStyle not preserved")
	}
	if style.TextDecorationColor != "blue" {
		t.Error("TextDecorationColor not preserved")
	}
	if style.VerticalAlign != VerticalAlignMiddle {
		t.Error("VerticalAlign not preserved")
	}
}

// TestDefaultTextProperties tests default values for new properties
func TestDefaultTextProperties(t *testing.T) {
	node := &Node{
		Text: "Default text",
		Style: Style{
			Display:   DisplayInlineText,
			TextStyle: &TextStyle{FontSize: 16},
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	constraints := Loose(400, 300)
	_ = LayoutText(node, constraints, ctx)

	// Check defaults
	style := node.Style.TextStyle
	if style.FontStyle != FontStyleNormal {
		t.Errorf("Default FontStyle = %v, want %v", style.FontStyle, FontStyleNormal)
	}
	if style.TextDecoration != TextDecorationNone {
		t.Errorf("Default TextDecoration = %v, want %v", style.TextDecoration, TextDecorationNone)
	}
	if style.TextDecorationStyle != TextDecorationStyleSolid {
		t.Errorf("Default TextDecorationStyle = %v, want %v", style.TextDecorationStyle, TextDecorationStyleSolid)
	}
	if style.VerticalAlign != VerticalAlignBaseline {
		t.Errorf("Default VerticalAlign = %v, want %v", style.VerticalAlign, VerticalAlignBaseline)
	}
}
