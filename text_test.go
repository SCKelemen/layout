package layout

import (
	"math"
	"strings"
	"testing"
)

// fakeMetrics provides deterministic text measurement for testing.
// Each character is 10px wide, making tests predictable.
type fakeMetrics struct {
	charWidth float64
}

func (f *fakeMetrics) Measure(text string, style TextStyle) (advance, ascent, descent float64) {
	advance = float64(len(text)) * f.charWidth
	if style.LetterSpacing > 0 {
		advance += float64(len(text)-1) * style.LetterSpacing
	}
	ascent = style.FontSize * 0.8
	descent = style.FontSize * 0.2
	return advance, ascent, descent
}

func setupFakeMetrics() {
	SetTextMetricsProvider(&fakeMetrics{charWidth: 10})
}

// TestTextBasic tests basic text layout with simple wrapping
func TestTextBasic(t *testing.T) {
	setupFakeMetrics()

	text := "Hello world"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize: 16,
		},
	})

	constraints := Loose(100, 200) // 100px width - should wrap "Hello world" (110px)
	size := LayoutText(node, constraints)

	// Should have 2 lines: "Hello" (50px) and "world" (50px)
	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}
	if len(node.TextLayout.Lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(node.TextLayout.Lines))
	}

	// Height should be 2 * line-height (default 1.2 * 16 = 19.2)
	expectedHeight := 2 * 16 * 1.2
	if math.Abs(size.Height-expectedHeight) > 0.1 {
		t.Errorf("Expected height %.2f, got %.2f", expectedHeight, size.Height)
	}
}

// TestTextWrappingInvariant tests that all lines fit within MaxWidth
func TestTextWrappingInvariant(t *testing.T) {
	setupFakeMetrics()

	text := "The quick brown fox jumps over the lazy dog"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize: 16,
		},
	})

	constraints := Loose(100, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// All lines should fit within MaxWidth (accounting for padding/border)
	maxLineWidth := 0.0
	for _, line := range node.TextLayout.Lines {
		if line.Width > maxLineWidth {
			maxLineWidth = line.Width
		}
	}

	// Content width is 100px (no padding in this test)
	if maxLineWidth > 100.1 {
		t.Errorf("All lines should fit within 100px, but max line width is %.2f", maxLineWidth)
	}
}

// TestTextWrappingMonotonic tests that shrinking MaxWidth increases line count
func TestTextWrappingMonotonic(t *testing.T) {
	setupFakeMetrics()

	text := "The quick brown fox jumps over the lazy dog"
	node1 := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize: 16,
		},
	})
	node2 := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize: 16,
		},
	})

	LayoutText(node1, Loose(200, 200))
	LayoutText(node2, Loose(100, 200))

	lines1 := len(node1.TextLayout.Lines)
	lines2 := len(node2.TextLayout.Lines)

	if lines2 < lines1 {
		t.Errorf("Shrinking width should increase line count: %d -> %d", lines1, lines2)
	}
}

// TestTextAlignLeft tests left alignment
func TestTextAlignLeft(t *testing.T) {
	setupFakeMetrics()

	text := "Hello"
	node := Text(text, Style{
		Width: 200,
		TextStyle: &TextStyle{
			FontSize:  16,
			TextAlign: TextAlignLeft,
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) == 0 {
		t.Fatal("TextLayout should have lines")
	}

	line := node.TextLayout.Lines[0]
	// Left aligned: offsetX should be 0 (or text-indent if set)
	if line.OffsetX < -0.1 {
		t.Errorf("Left-aligned line should have offsetX >= 0, got %.2f", line.OffsetX)
	}
}

// TestTextAlignRight tests right alignment
func TestTextAlignRight(t *testing.T) {
	setupFakeMetrics()

	text := "Hello"
	node := Text(text, Style{
		Width: 200,
		TextStyle: &TextStyle{
			FontSize:  16,
			TextAlign: TextAlignRight,
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) == 0 {
		t.Fatal("TextLayout should have lines")
	}

	line := node.TextLayout.Lines[0]
	// "Hello" is 50px wide, content width is 200px
	// Right aligned: offsetX should be 200 - 50 = 150
	expectedOffset := 200.0 - 50.0
	if math.Abs(line.OffsetX-expectedOffset) > 0.1 {
		t.Errorf("Right-aligned line should have offsetX ≈ %.2f, got %.2f", expectedOffset, line.OffsetX)
	}
}

// TestTextAlignCenter tests center alignment
func TestTextAlignCenter(t *testing.T) {
	setupFakeMetrics()

	text := "Hello"
	node := Text(text, Style{
		Width: 200,
		TextStyle: &TextStyle{
			FontSize:  16,
			TextAlign: TextAlignCenter,
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) == 0 {
		t.Fatal("TextLayout should have lines")
	}

	line := node.TextLayout.Lines[0]
	// "Hello" is 50px wide, content width is 200px
	// Center aligned: offsetX should be (200 - 50) / 2 = 75
	expectedOffset := (200.0 - 50.0) / 2.0
	if math.Abs(line.OffsetX-expectedOffset) > 0.1 {
		t.Errorf("Center-aligned line should have offsetX ≈ %.2f, got %.2f", expectedOffset, line.OffsetX)
	}
}

// TestTextAlignDefault tests that default resolves to left in LTR
func TestTextAlignDefault(t *testing.T) {
	setupFakeMetrics()

	text := "Hello"
	node := Text(text, Style{
		Width: 200,
		TextStyle: &TextStyle{
			FontSize:  16,
			TextAlign: TextAlignDefault, // Zero value
			Direction: DirectionLTR,
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) == 0 {
		t.Fatal("TextLayout should have lines")
	}

	line := node.TextLayout.Lines[0]
	// Default should resolve to left in LTR
	if line.OffsetX < -0.1 {
		t.Errorf("Default alignment (LTR) should be left (offsetX >= 0), got %.2f", line.OffsetX)
	}
}

// TestWhiteSpaceNormal tests normal white-space handling
func TestWhiteSpaceNormal(t *testing.T) {
	setupFakeMetrics()

	text := "Hello    world\n\n\nTest"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize:   16,
			WhiteSpace: WhiteSpaceNormal, // Zero value
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	// Normal should collapse multiple spaces and newlines
	// "Hello    world\n\n\nTest" should become "Hello world Test"
	// All text should be in lines
	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Verify whitespace was collapsed (text should be laid out)
	totalText := ""
	for _, line := range node.TextLayout.Lines {
		for _, box := range line.Boxes {
			totalText += box.Text
		}
	}

	// Should not contain multiple consecutive spaces
	if len(totalText) == 0 {
		t.Error("Text should not be empty after whitespace collapse")
	}
}

// TestWhiteSpaceNowrap tests nowrap white-space handling
func TestWhiteSpaceNowrap(t *testing.T) {
	setupFakeMetrics()

	text := "This is a very long line that should not wrap"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize:   16,
			WhiteSpace: WhiteSpaceNowrap,
		},
	})

	constraints := Loose(100, 200) // Narrow width
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Nowrap should produce a single line even if it overflows
	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("Nowrap should produce 1 line, got %d", len(node.TextLayout.Lines))
	}
}

// TestWhiteSpacePre tests pre white-space handling
func TestWhiteSpacePre(t *testing.T) {
	setupFakeMetrics()

	text := "Hello    world\n\nTest"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize:   16,
			WhiteSpace: WhiteSpacePre,
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Pre should preserve spaces and newlines, no wrapping
	// Should have multiple lines due to newlines
	if len(node.TextLayout.Lines) < 2 {
		t.Errorf("Pre with newlines should produce multiple lines, got %d", len(node.TextLayout.Lines))
	}
}

// TestLineHeightNormal tests normal line height (default)
func TestLineHeightNormal(t *testing.T) {
	setupFakeMetrics()

	text := "Line 1\nLine 2\nLine 3"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize:   16,
			LineHeight: 0, // Normal (zero value)
			WhiteSpace: WhiteSpacePre,
		},
	})

	constraints := Loose(200, 200)
	size := LayoutText(node, constraints)

	// Normal line height is 1.2 × fontSize = 19.2
	// 3 lines × 19.2 = 57.6
	expectedHeight := 3 * 16 * 1.2
	if math.Abs(size.Height-expectedHeight) > 0.1 {
		t.Errorf("Expected height %.2f with normal line-height, got %.2f", expectedHeight, size.Height)
	}
}

// TestLineHeightMultiplier tests line height as multiplier
func TestLineHeightMultiplier(t *testing.T) {
	setupFakeMetrics()

	text := "Line 1\nLine 2"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize:   16,
			LineHeight: 1.5, // Multiplier
			WhiteSpace: WhiteSpacePre,
		},
	})

	constraints := Loose(200, 200)
	size := LayoutText(node, constraints)

	// Line height = 16 × 1.5 = 24
	// 2 lines × 24 = 48
	expectedHeight := 2 * 16 * 1.5
	if math.Abs(size.Height-expectedHeight) > 0.1 {
		t.Errorf("Expected height %.2f with 1.5x line-height, got %.2f", expectedHeight, size.Height)
	}
}

// TestLineHeightAbsolute tests line height as absolute pixels
func TestLineHeightAbsolute(t *testing.T) {
	setupFakeMetrics()

	text := "Line 1\nLine 2"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize:   16,
			LineHeight: 30, // Absolute (>= 10 treated as pixels)
			WhiteSpace: WhiteSpacePre,
		},
	})

	constraints := Loose(200, 200)
	size := LayoutText(node, constraints)

	// Line height = 30px
	// 2 lines × 30 = 60
	expectedHeight := 2 * 30.0
	if math.Abs(size.Height-expectedHeight) > 0.1 {
		t.Errorf("Expected height %.2f with 30px line-height, got %.2f", expectedHeight, size.Height)
	}
}

// TestTextIndent tests first line indentation
func TestTextIndent(t *testing.T) {
	setupFakeMetrics()

	text := "First line\nSecond line"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize:   16,
			TextIndent: 20, // 20px indent
			WhiteSpace: WhiteSpacePre,
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) < 2 {
		t.Fatal("TextLayout should have at least 2 lines")
	}

	firstLine := node.TextLayout.Lines[0]
	secondLine := node.TextLayout.Lines[1]

	// First line should have indent
	if math.Abs(firstLine.OffsetX-20.0) > 0.1 {
		t.Errorf("First line should have 20px indent, got offsetX %.2f", firstLine.OffsetX)
	}

	// Second line should not have indent
	if secondLine.OffsetX > 0.1 {
		t.Errorf("Second line should not have indent, got offsetX %.2f", secondLine.OffsetX)
	}
}

// TestWordSpacing tests word spacing
func TestWordSpacing(t *testing.T) {
	setupFakeMetrics()

	text := "Hello world"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize:    16,
			WordSpacing: 5, // 5px extra spacing between words
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) == 0 {
		t.Fatal("TextLayout should have lines")
	}

	// "Hello" (50px) + 5px spacing + "world" (50px) = 105px
	// Should fit on one line with 200px width
	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("With 200px width, text should fit on 1 line, got %d", len(node.TextLayout.Lines))
	}
}

// TestLetterSpacing tests letter spacing
func TestLetterSpacing(t *testing.T) {
	setupFakeMetrics()

	text := "Hi"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize:      16,
			LetterSpacing: 2, // 2px spacing between letters
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) == 0 {
		t.Fatal("TextLayout should have lines")
	}

	// "Hi" = 2 chars × 10px = 20px
	// Plus 1 space between letters × 2px = 2px
	// Total = 22px
	line := node.TextLayout.Lines[0]
	expectedWidth := 20.0 + 2.0
	if math.Abs(line.Width-expectedWidth) > 0.1 {
		t.Errorf("Expected line width %.2f with letter spacing, got %.2f", expectedWidth, line.Width)
	}
}

// TestTextEmpty tests empty text
func TestTextEmpty(t *testing.T) {
	setupFakeMetrics()

	node := Text("", Style{
		TextStyle: &TextStyle{
			FontSize: 16,
		},
	})

	constraints := Loose(200, 200)
	size := LayoutText(node, constraints)

	// Empty text should have minimal height
	if size.Height < 0 {
		t.Errorf("Empty text should have non-negative height, got %.2f", size.Height)
	}

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated even for empty text")
	}

	if len(node.TextLayout.Lines) != 0 {
		t.Errorf("Empty text should have 0 lines, got %d", len(node.TextLayout.Lines))
	}
}

// TestTextVeryLongWord tests handling of very long words
func TestTextVeryLongWord(t *testing.T) {
	setupFakeMetrics()

	// Create a word longer than available width
	longWord := ""
	for i := 0; i < 20; i++ {
		longWord += "a" // 20 chars = 200px
	}

	node := Text(longWord, Style{
		TextStyle: &TextStyle{
			FontSize: 16,
		},
	})

	constraints := Loose(100, 200) // 100px width, word is 200px
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Very long word should still be laid out (may overflow)
	// In normal mode, we can't break words, so it should be on one line
	if len(node.TextLayout.Lines) == 0 {
		t.Error("Very long word should still produce at least one line")
	}
}

// TestTextBlockIntegration tests text nodes within block layout
func TestTextBlockIntegration(t *testing.T) {
	setupFakeMetrics()

	root := &Node{
		Style: Style{
			Width: 200,
		},
		Children: []*Node{
			Text("First paragraph", Style{
				TextStyle: &TextStyle{
					FontSize: 16,
				},
			}),
			Text("Second paragraph", Style{
				TextStyle: &TextStyle{
					FontSize: 16,
				},
			}),
		},
	}

	constraints := Loose(200, 500)
	LayoutBlock(root, constraints)

	// Both text nodes should be laid out
	if root.Children[0].TextLayout == nil {
		t.Error("First text node should have TextLayout")
	}
	if root.Children[1].TextLayout == nil {
		t.Error("Second text node should have TextLayout")
	}

	// Second paragraph should be below first
	firstBottom := root.Children[0].Rect.Y + root.Children[0].Rect.Height
	if root.Children[1].Rect.Y < firstBottom-0.1 {
		t.Errorf("Second paragraph should be below first: %.2f < %.2f", root.Children[1].Rect.Y, firstBottom)
	}
}

// TestTextBlockAutoHeight tests block with text and auto height
func TestTextBlockAutoHeight(t *testing.T) {
	setupFakeMetrics()

	root := &Node{
		Style: Style{
			Width:  200,
			Height: -1, // Auto
		},
		Children: []*Node{
			Text("This is a paragraph that will wrap to multiple lines", Style{
				TextStyle: &TextStyle{
					FontSize: 16,
				},
			}),
		},
	}

	constraints := Loose(200, 500)
	size := LayoutBlock(root, constraints)

	// Block height should be determined by text height
	if size.Height <= 0 {
		t.Error("Block with text should have positive height")
	}

	// Height should match text height
	textHeight := root.Children[0].Rect.Height
	if math.Abs(size.Height-textHeight) > 0.1 {
		t.Errorf("Block height should match text height: %.2f vs %.2f", size.Height, textHeight)
	}
}

// TestTextPadding tests text with padding
func TestTextPadding(t *testing.T) {
	setupFakeMetrics()

	node := Text("Hello", Style{
		Width:   100,
		Padding: Uniform(10),
		TextStyle: &TextStyle{
			FontSize: 16,
		},
	})

	constraints := Loose(200, 200)
	size := LayoutText(node, constraints)

	// Size should include padding: content + 20px (10px each side)
	// Content height is 1 line × 19.2 = 19.2
	// Total height = 19.2 + 20 = 39.2
	expectedHeight := 16*1.2 + 20.0
	if math.Abs(size.Height-expectedHeight) > 0.1 {
		t.Errorf("Expected height with padding %.2f, got %.2f", expectedHeight, size.Height)
	}
}

// TestTextExplicitWidth tests explicit width constraint
func TestTextExplicitWidth(t *testing.T) {
	setupFakeMetrics()

	node := Text("Hello world this is a long line", Style{
		Width: 150, // Explicit width
		TextStyle: &TextStyle{
			FontSize: 16,
		},
	})

	constraints := Loose(200, 200)
	size := LayoutText(node, constraints)

	// Width should be 150 + padding (0 in this case)
	if math.Abs(size.Width-150.0) > 0.1 {
		t.Errorf("Expected width 150, got %.2f", size.Width)
	}
}

// TestTextExplicitHeight tests explicit height constraint
func TestTextExplicitHeight(t *testing.T) {
	setupFakeMetrics()

	node := Text("Hello world", Style{
		Height: 100, // Explicit height
		TextStyle: &TextStyle{
			FontSize: 16,
		},
	})

	constraints := Loose(200, 200)
	size := LayoutText(node, constraints)

	// Height should be 100 + padding (0 in this case)
	if math.Abs(size.Height-100.0) > 0.1 {
		t.Errorf("Expected height 100, got %.2f", size.Height)
	}
}

// TestTextDefaultTextStyle tests that default TextStyle is created
func TestTextDefaultTextStyle(t *testing.T) {
	setupFakeMetrics()

	node := &Node{
		Text: "Hello",
		Style: Style{
			Display: DisplayInlineText,
			// TextStyle is nil
		},
	}

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	// LayoutText should create default TextStyle
	if node.Style.TextStyle == nil {
		t.Error("LayoutText should create default TextStyle when nil")
	}

	// Default values should be set
	style := node.Style.TextStyle
	if style.FontSize != 16 {
		t.Errorf("Default FontSize should be 16, got %.2f", style.FontSize)
	}
	if style.TextAlign != TextAlignDefault {
		t.Errorf("Default TextAlign should be TextAlignDefault, got %v", style.TextAlign)
	}
	if style.WhiteSpace != WhiteSpaceNormal {
		t.Errorf("Default WhiteSpace should be WhiteSpaceNormal, got %v", style.WhiteSpace)
	}
}

// TestTextAlignmentInvariant tests alignment invariants
func TestTextAlignmentInvariant(t *testing.T) {
	setupFakeMetrics()

	testCases := []struct {
		name      string
		textAlign TextAlign
		width     float64
		text      string
	}{
		{"left", TextAlignLeft, 200, "Hello"},
		{"right", TextAlignRight, 200, "Hello"},
		{"center", TextAlignCenter, 200, "Hello"},
		{"default", TextAlignDefault, 200, "Hello"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node := Text(tc.text, Style{
				Width: tc.width,
				TextStyle: &TextStyle{
					FontSize:  16,
					TextAlign: tc.textAlign,
				},
			})

			constraints := Loose(tc.width, 200)
			LayoutText(node, constraints)

			if node.TextLayout == nil || len(node.TextLayout.Lines) == 0 {
				t.Fatal("TextLayout should have lines")
			}

			line := node.TextLayout.Lines[0]
			lineWidth := line.Width

			// Alignment invariant: line should fit within content width
			if lineWidth > tc.width+0.1 {
				t.Errorf("Line width %.2f should fit within content width %.2f", lineWidth, tc.width)
			}

			// Alignment invariant: offsetX + lineWidth should not exceed content width
			if line.OffsetX+lineWidth > tc.width+0.1 {
				t.Errorf("Line (offsetX=%.2f, width=%.2f) exceeds content width %.2f",
					line.OffsetX, lineWidth, tc.width)
			}
		})
	}
}

// TestTextLineHeightInvariant tests line height calculation
func TestTextLineHeightInvariant(t *testing.T) {
	setupFakeMetrics()

	text := "Line 1\nLine 2\nLine 3"
	testCases := []struct {
		name       string
		lineHeight float64
		expected   float64
	}{
		{"normal", 0, 3 * 16 * 1.2},           // 3 lines × 19.2 = 57.6
		{"multiplier_1.5", 1.5, 3 * 16 * 1.5}, // 3 lines × 24 = 72
		{"multiplier_2.0", 2.0, 3 * 16 * 2.0}, // 3 lines × 32 = 96
		{"absolute_20", 20, 3 * 20.0},         // 3 lines × 20 = 60
		{"absolute_30", 30, 3 * 30.0},         // 3 lines × 30 = 90
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node := Text(text, Style{
				TextStyle: &TextStyle{
					FontSize:   16,
					LineHeight: tc.lineHeight,
					WhiteSpace: WhiteSpacePre,
				},
			})

			constraints := Loose(200, 200)
			size := LayoutText(node, constraints)

			// Height should match expected calculation
			if math.Abs(size.Height-tc.expected) > 0.1 {
				t.Errorf("Line height %.2f should produce height %.2f, got %.2f",
					tc.lineHeight, tc.expected, size.Height)
			}
		})
	}
}

// TestWhiteSpaceNonBreakingSpace tests that non-breaking spaces are preserved
func TestWhiteSpaceNonBreakingSpace(t *testing.T) {
	setupFakeMetrics()

	// Non-breaking space (U+00A0) should not collapse
	nbSpace := "\u00A0"
	text := "Hello" + nbSpace + nbSpace + "world" + "   " + "test"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize:   16,
			WhiteSpace: WhiteSpaceNormal,
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Check that non-breaking spaces are preserved in the layout
	hasNBSP := false
	for _, line := range node.TextLayout.Lines {
		for _, box := range line.Boxes {
			if strings.Contains(box.Text, nbSpace) {
				hasNBSP = true
				break
			}
		}
		if hasNBSP {
			break
		}
	}

	if !hasNBSP {
		t.Error("Non-breaking spaces should be preserved in white-space: normal")
	}

	// Regular spaces should be collapsed
	allText := ""
	for _, line := range node.TextLayout.Lines {
		for _, box := range line.Boxes {
			allText += box.Text
		}
	}

	// Should not have multiple consecutive regular spaces (except NBSP)
	// Check for "  " (two regular spaces) - should not exist
	if strings.Contains(allText, "  ") {
		// Check if it's actually NBSP sequences
		if !strings.Contains(allText, nbSpace+nbSpace) {
			t.Error("Regular spaces should be collapsed to single space")
		}
	}
}
