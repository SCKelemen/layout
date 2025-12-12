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
	// Count runes (characters), not bytes, to handle Unicode correctly
	runeCount := len([]rune(text))
	advance = float64(runeCount) * f.charWidth
	// Letter spacing applies between characters (not after last one)
	if style.LetterSpacing != -1 && runeCount > 0 {
		advance += float64(runeCount-1) * style.LetterSpacing
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

// TestTextAlignJustify tests basic text justification with multiple lines
func TestTextAlignJustify(t *testing.T) {
	setupFakeMetrics()

	// "Hello world foo" with width 120 creates two lines:
	// Line 1: "Hello world" (110px) - should justify to 120px
	// Line 2: "foo" (30px) - last line, not justified
	text := "Hello world foo"
	node := Text(text, Style{
		Width: 120,
		TextStyle: &TextStyle{
			FontSize:  16,
			TextAlign: TextAlignJustify,
		},
	})

	constraints := Loose(120, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) < 2 {
		t.Fatal("TextLayout should have at least 2 lines")
	}

	firstLine := node.TextLayout.Lines[0]
	lastLine := node.TextLayout.Lines[1]

	// First line should be justified to full width
	if math.Abs(firstLine.Width-120.0) > 0.1 {
		t.Errorf("First line should be justified to 120px, got %.2f", firstLine.Width)
	}

	// Should have 1 space to adjust
	if firstLine.SpaceCount != 1 {
		t.Errorf("Expected 1 space in first line, got %d", firstLine.SpaceCount)
	}

	// Space adjustment should be 10px (120 - 110)
	expectedAdjustment := 10.0
	if math.Abs(firstLine.SpaceAdjustment-expectedAdjustment) > 0.1 {
		t.Errorf("Expected space adjustment %.2f, got %.2f",
			expectedAdjustment, firstLine.SpaceAdjustment)
	}

	// Should start at left (like left-align)
	if firstLine.OffsetX != 0.0 {
		t.Errorf("First line should start at 0, got %.2f", firstLine.OffsetX)
	}

	// Last line should NOT be justified
	if math.Abs(lastLine.Width-30.0) > 0.1 {
		t.Errorf("Last line should not be justified (30px), got %.2f", lastLine.Width)
	}
}

// TestJustifyLastLineNotJustified tests that last line is not justified
func TestJustifyLastLineNotJustified(t *testing.T) {
	setupFakeMetrics()

	// Two lines: "Hello world" + "test"
	// Only first line should be justified
	text := "Hello world test"
	node := Text(text, Style{
		Width: 120, // Forces wrap after "world"
		TextStyle: &TextStyle{
			FontSize:  16,
			TextAlign: TextAlignJustify,
		},
	})

	constraints := Loose(120, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) < 2 {
		t.Fatal("TextLayout should have at least 2 lines")
	}

	firstLine := node.TextLayout.Lines[0]
	lastLine := node.TextLayout.Lines[1]

	// First line: justified (full width)
	if math.Abs(firstLine.Width-120.0) > 0.1 {
		t.Errorf("First line should be justified to 120px, got %.2f", firstLine.Width)
	}

	// Last line: NOT justified (natural width)
	// "test" = 40px
	if math.Abs(lastLine.Width-40.0) > 0.1 {
		t.Errorf("Last line should not be justified (40px), got width %.2f", lastLine.Width)
	}

	// Last line should be left-aligned
	if lastLine.OffsetX != 0.0 {
		t.Errorf("Last line should be left-aligned, got offsetX %.2f", lastLine.OffsetX)
	}

	// Last line should have no space adjustment
	if lastLine.SpaceAdjustment != 0.0 {
		t.Errorf("Last line should have no space adjustment, got %.2f", lastLine.SpaceAdjustment)
	}
}

// TestJustifySingleWord tests that single-word lines are not justified
func TestJustifySingleWord(t *testing.T) {
	setupFakeMetrics()

	text := "Hello"
	node := Text(text, Style{
		Width: 200,
		TextStyle: &TextStyle{
			FontSize:  16,
			TextAlign: TextAlignJustify,
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) == 0 {
		t.Fatal("TextLayout should have lines")
	}

	line := node.TextLayout.Lines[0]

	// Single word: NOT justified (no spaces to adjust)
	if line.SpaceCount != 0 {
		t.Errorf("Single word should have 0 spaces, got %d", line.SpaceCount)
	}

	// Should be natural width, not full width
	// "Hello" = 50px
	if math.Abs(line.Width-50.0) > 0.1 {
		t.Errorf("Single word should not be justified (50px), got width %.2f", line.Width)
	}

	// Should have no space adjustment
	if line.SpaceAdjustment != 0.0 {
		t.Errorf("Single word should have no space adjustment, got %.2f", line.SpaceAdjustment)
	}

	// Should be left-aligned
	if line.OffsetX != 0.0 {
		t.Errorf("Single word should be left-aligned, got offsetX %.2f", line.OffsetX)
	}
}

// TestJustifyMultipleSpaces tests even distribution across multiple spaces
func TestJustifyMultipleSpaces(t *testing.T) {
	setupFakeMetrics()

	// Create multi-line text to test multiple spaces on first line
	// "The quick brown fox" with width 160 creates two lines:
	// Line 1: "The quick brown" (150px) - should justify to 160px
	// Line 2: "fox" (30px) - last line, not justified
	text := "The quick brown fox"
	node := Text(text, Style{
		Width: 160,
		TextStyle: &TextStyle{
			FontSize:  16,
			TextAlign: TextAlignJustify,
		},
	})

	constraints := Loose(160, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) < 2 {
		t.Fatal("TextLayout should have at least 2 lines")
	}

	firstLine := node.TextLayout.Lines[0]

	// Should have 2 spaces
	if firstLine.SpaceCount != 2 {
		t.Errorf("Expected 2 spaces, got %d", firstLine.SpaceCount)
	}

	// Extra space: 160 - 150 = 10px, distributed over 2 spaces = 5px each
	expectedAdjustment := 5.0
	if math.Abs(firstLine.SpaceAdjustment-expectedAdjustment) > 0.1 {
		t.Errorf("Expected space adjustment %.2f, got %.2f",
			expectedAdjustment, firstLine.SpaceAdjustment)
	}

	// First line should be full width
	if math.Abs(firstLine.Width-160.0) > 0.1 {
		t.Errorf("First line should be justified to 160px, got %.2f", firstLine.Width)
	}
}

// TestJustifyWithTextIndent tests text-indent interaction with justification
func TestJustifyWithTextIndent(t *testing.T) {
	setupFakeMetrics()

	// Use normal white-space mode (not pre) to test text-indent with justify
	// "First line here\nMore text here" with 150px width creates multiple lines
	// First line gets text-indent of 20px, reducing available width to 130px
	text := "First line here and text"
	node := Text(text, Style{
		Width: 150,
		TextStyle: &TextStyle{
			FontSize:   16,
			TextAlign:  TextAlignJustify,
			TextIndent: 20,
		},
	})

	constraints := Loose(150, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) < 2 {
		t.Fatal("TextLayout should have at least 2 lines")
	}

	firstLine := node.TextLayout.Lines[0]

	// First line: indent affects available width (150 - 20 = 130)
	// Should justify to fill that reduced width
	if math.Abs(firstLine.Width-130.0) > 0.1 {
		t.Errorf("First line should justify to 130px (150 - 20 indent), got %.2f", firstLine.Width)
	}

	// First line should start at indent
	if firstLine.OffsetX != 20.0 {
		t.Errorf("First line should start at indent (20), got offsetX %.2f", firstLine.OffsetX)
	}

	// First line should have justified (spaceAdjustment > 0)
	if firstLine.SpaceAdjustment == 0.0 {
		t.Errorf("First line should have space adjustment, got 0")
	}
}

// TestJustifyWithWordSpacing tests word-spacing interaction with justification
func TestJustifyWithWordSpacing(t *testing.T) {
	setupFakeMetrics()

	// "Hello world test" with 5px word-spacing, width 130
	// Line 1: "Hello world" = 100px (chars) + 15px (space with word-spacing) = 115px
	// Should justify to 130px: extra 15px distributed to the 1 space
	// Line 2: "test" = 40px - not justified
	text := "Hello world test"
	node := Text(text, Style{
		Width: 130,
		TextStyle: &TextStyle{
			FontSize:    16,
			TextAlign:   TextAlignJustify,
			WordSpacing: 5,
		},
	})

	constraints := Loose(130, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) < 2 {
		t.Fatal("TextLayout should have at least 2 lines")
	}

	firstLine := node.TextLayout.Lines[0]

	// Total line width should be 130px
	if math.Abs(firstLine.Width-130.0) > 0.1 {
		t.Errorf("First line should be 130px, got %.2f", firstLine.Width)
	}

	// Should have 1 space
	if firstLine.SpaceCount != 1 {
		t.Errorf("Expected 1 space, got %d", firstLine.SpaceCount)
	}

	// Space adjustment accounts for base+word-spacing already included
	// Original space width: 15px (10 base + 5 word-spacing)
	// Extra space: 130 - 115 = 15px
	expectedAdjustment := 15.0
	if math.Abs(firstLine.SpaceAdjustment-expectedAdjustment) > 0.1 {
		t.Errorf("Expected space adjustment %.2f, got %.2f",
			expectedAdjustment, firstLine.SpaceAdjustment)
	}
}

func TestTextAlignLastLeft(t *testing.T) {
	setupFakeMetrics()

	// Justify with explicit left-alignment for last line
	text := "Hello world test again"
	node := Text(text, Style{
		Width: 120,
		TextStyle: &TextStyle{
			FontSize:      16,
			TextAlign:     TextAlignJustify,
			TextAlignLast: TextAlignLastLeft,
		},
	})

	constraints := Loose(120, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) < 2 {
		t.Fatal("TextLayout should have at least 2 lines")
	}

	// Last line should be left-aligned (not justified)
	lastLine := node.TextLayout.Lines[len(node.TextLayout.Lines)-1]
	if lastLine.OffsetX != 0.0 {
		t.Errorf("Last line should be left-aligned (OffsetX=0), got %.2f", lastLine.OffsetX)
	}

	// Last line should NOT be justified (width less than contentWidth)
	if lastLine.Width >= 120.0 {
		t.Errorf("Last line should not be justified, got width %.2f", lastLine.Width)
	}
}

func TestTextAlignLastCenter(t *testing.T) {
	setupFakeMetrics()

	// Justify with center-alignment for last line
	text := "Hello world test again"
	node := Text(text, Style{
		Width: 120,
		TextStyle: &TextStyle{
			FontSize:      16,
			TextAlign:     TextAlignJustify,
			TextAlignLast: TextAlignLastCenter,
		},
	})

	constraints := Loose(120, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) < 2 {
		t.Fatal("TextLayout should have at least 2 lines")
	}

	// First line should be justified
	firstLine := node.TextLayout.Lines[0]
	if math.Abs(firstLine.Width-120.0) > 0.1 {
		t.Errorf("First line should be justified to 120px, got %.2f", firstLine.Width)
	}

	// Last line should be centered (not left-aligned at 0)
	lastLine := node.TextLayout.Lines[len(node.TextLayout.Lines)-1]
	if lastLine.OffsetX <= 0.1 {
		t.Errorf("Last line should be centered (OffsetX > 0), got %.2f", lastLine.OffsetX)
	}

	// Last line should NOT be justified
	if lastLine.Width >= 120.0 {
		t.Errorf("Last line should not be justified, got width %.2f", lastLine.Width)
	}
}

func TestTextAlignLastJustify(t *testing.T) {
	setupFakeMetrics()

	// Justify including last line
	text := "Hello world test again"
	node := Text(text, Style{
		Width: 120,
		TextStyle: &TextStyle{
			FontSize:      16,
			TextAlign:     TextAlignJustify,
			TextAlignLast: TextAlignLastJustify,
		},
	})

	constraints := Loose(120, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) < 2 {
		t.Fatal("TextLayout should have at least 2 lines")
	}

	// All lines should be justified
	for i, line := range node.TextLayout.Lines {
		// Only check lines with multiple words
		if line.SpaceCount > 0 {
			if math.Abs(line.Width-120.0) > 0.1 {
				t.Errorf("Line %d should be justified to 120px, got %.2f", i, line.Width)
			}
		}
	}

	// Last line should also be justified if it has spaces
	lastLine := node.TextLayout.Lines[len(node.TextLayout.Lines)-1]
	if lastLine.SpaceCount > 0 {
		if math.Abs(lastLine.Width-120.0) > 0.1 {
			t.Errorf("Last line should be justified to 120px, got %.2f", lastLine.Width)
		}
	}
}

func TestTextJustifyInterWord(t *testing.T) {
	setupFakeMetrics()

	// Default justification - expand word spaces only
	text := "Hello world test again"
	node := Text(text, Style{
		Width: 120,
		TextStyle: &TextStyle{
			FontSize:    16,
			TextAlign:   TextAlignJustify,
			TextJustify: TextJustifyInterWord,
		},
	})

	constraints := Loose(120, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) < 2 {
		t.Fatal("TextLayout should have at least 2 lines")
	}

	// First line should be justified
	firstLine := node.TextLayout.Lines[0]
	if firstLine.SpaceCount > 0 {
		if math.Abs(firstLine.Width-120.0) > 0.1 {
			t.Errorf("First line should be justified to 120px, got %.2f", firstLine.Width)
		}
	}
}

func TestTextJustifyNone(t *testing.T) {
	setupFakeMetrics()

	// No justification despite text-align: justify
	text := "Hello world test again"
	node := Text(text, Style{
		Width: 120,
		TextStyle: &TextStyle{
			FontSize:    16,
			TextAlign:   TextAlignJustify,
			TextJustify: TextJustifyNone,
		},
	})

	constraints := Loose(120, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) < 2 {
		t.Fatal("TextLayout should have at least 2 lines")
	}

	// Lines should NOT be justified (left-aligned instead)
	for i, line := range node.TextLayout.Lines {
		// Lines should not fill the width
		if line.SpaceCount > 0 && math.Abs(line.Width-120.0) < 0.1 {
			t.Errorf("Line %d should not be justified with text-justify: none, but got width %.2f", i, line.Width)
		}
	}
}

func TestTextJustifyAuto(t *testing.T) {
	setupFakeMetrics()

	// Auto should resolve to inter-word
	text := "Hello world test again"
	node := Text(text, Style{
		Width: 120,
		TextStyle: &TextStyle{
			FontSize:    16,
			TextAlign:   TextAlignJustify,
			TextJustify: TextJustifyAuto, // Should behave like inter-word
		},
	})

	constraints := Loose(120, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil || len(node.TextLayout.Lines) < 2 {
		t.Fatal("TextLayout should have at least 2 lines")
	}

	// First line should be justified
	firstLine := node.TextLayout.Lines[0]
	if firstLine.SpaceCount > 0 {
		if math.Abs(firstLine.Width-120.0) > 0.1 {
			t.Errorf("First line should be justified with auto (inter-word), got %.2f", firstLine.Width)
		}
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

func TestWhiteSpacePreWrap(t *testing.T) {
	setupFakeMetrics()

	// Test that spaces are preserved and wrapping occurs
	text := "Hello    world test"
	node := Text(text, Style{
		Width: 60, // Narrow width to force wrapping
		TextStyle: &TextStyle{
			FontSize:   16,
			WhiteSpace: WhiteSpacePreWrap,
		},
	})

	constraints := Loose(60, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Pre-wrap should preserve spaces but allow wrapping
	// With narrow width, should wrap into multiple lines
	if len(node.TextLayout.Lines) < 2 {
		t.Errorf("Pre-wrap with narrow width should wrap, got %d lines", len(node.TextLayout.Lines))
	}

	// Check that spaces are preserved in the text
	// Count total spaces across all boxes
	totalSpaces := 0
	for _, line := range node.TextLayout.Lines {
		for _, box := range line.Boxes {
			for _, r := range box.Text {
				if r == ' ' {
					totalSpaces++
				}
			}
		}
	}
	// Original text "Hello    world test" has 5 spaces total (4 between Hello and world, 1 between world and test)
	if totalSpaces < 5 {
		t.Errorf("Pre-wrap should preserve all spaces, expected at least 5 spaces, got %d", totalSpaces)
	}
}

func TestWhiteSpacePreWrapNewlines(t *testing.T) {
	setupFakeMetrics()

	// Test that newlines create explicit breaks
	text := "Line1\nLine2\nLine3"
	node := Text(text, Style{
		Width: 200, // Wide enough to fit each line
		TextStyle: &TextStyle{
			FontSize:   16,
			WhiteSpace: WhiteSpacePreWrap,
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Pre-wrap should create line breaks at newlines
	if len(node.TextLayout.Lines) != 3 {
		t.Errorf("Pre-wrap with 2 newlines should create 3 lines, got %d", len(node.TextLayout.Lines))
	}
}

func TestWhiteSpacePreLine(t *testing.T) {
	setupFakeMetrics()

	// Test that spaces collapse but newlines preserved
	text := "Hello    world\nTest    line"
	node := Text(text, Style{
		Width: 200,
		TextStyle: &TextStyle{
			FontSize:   16,
			WhiteSpace: WhiteSpacePreLine,
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Pre-line should preserve newlines (2 lines)
	if len(node.TextLayout.Lines) != 2 {
		t.Errorf("Pre-line should preserve newline, got %d lines", len(node.TextLayout.Lines))
	}

	// Check that multiple spaces are collapsed
	for _, line := range node.TextLayout.Lines {
		for _, box := range line.Boxes {
			if strings.Contains(box.Text, "    ") {
				t.Error("Pre-line should collapse multiple spaces")
			}
		}
	}
}

func TestWhiteSpacePreLineWrapping(t *testing.T) {
	setupFakeMetrics()

	// Test wrapping with pre-line
	text := "This is a very long line that should wrap\nShort line"
	node := Text(text, Style{
		Width: 60, // Narrow width to force wrapping
		TextStyle: &TextStyle{
			FontSize:   16,
			WhiteSpace: WhiteSpacePreLine,
		},
	})

	constraints := Loose(60, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Pre-line should wrap long lines and preserve newlines
	// First segment wraps into multiple lines, second is one line
	if len(node.TextLayout.Lines) < 3 {
		t.Errorf("Pre-line should wrap long line and preserve newline, got %d lines", len(node.TextLayout.Lines))
	}
}

func TestOverflowWrapBreakWord(t *testing.T) {
	setupFakeMetrics()

	// Very long word should break with overflow-wrap: break-word
	text := "supercalifragilisticexpialidocious"
	node := Text(text, Style{
		Width: 60, // Narrow width, word is ~280px (35 chars * 8px)
		TextStyle: &TextStyle{
			FontSize:     16,
			OverflowWrap: OverflowWrapBreakWord,
		},
	})

	constraints := Loose(60, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Long word should break into multiple lines
	if len(node.TextLayout.Lines) < 5 {
		t.Errorf("Long word with overflow-wrap: break-word should break into multiple lines, got %d", len(node.TextLayout.Lines))
	}

	// Each piece should fit within the width (except possibly last)
	for i, line := range node.TextLayout.Lines {
		if i < len(node.TextLayout.Lines)-1 && line.Width > 60.0 {
			t.Errorf("Line %d should fit within 60px, got %.2f", i, line.Width)
		}
	}
}

func TestWordBreakBreakAll(t *testing.T) {
	setupFakeMetrics()

	// Break between any characters with word-break: break-all
	text := "verylongword"
	node := Text(text, Style{
		Width: 50,
		TextStyle: &TextStyle{
			FontSize:  16,
			WordBreak: WordBreakBreakAll,
		},
	})

	constraints := Loose(50, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Word should break into multiple lines
	if len(node.TextLayout.Lines) < 2 {
		t.Errorf("Long word with word-break: break-all should break, got %d lines", len(node.TextLayout.Lines))
	}
}

func TestOverflowWrapNormal(t *testing.T) {
	setupFakeMetrics()

	// Long word overflows without breaking (normal behavior)
	text := "verylongwordthatdoesnotfit"
	node := Text(text, Style{
		Width: 50,
		TextStyle: &TextStyle{
			FontSize:     16,
			OverflowWrap: OverflowWrapNormal, // Default
		},
	})

	constraints := Loose(50, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Word should remain on single line (overflowing)
	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("Long word with overflow-wrap: normal should stay on one line, got %d lines", len(node.TextLayout.Lines))
	}

	// Line width should exceed container width (overflow)
	if node.TextLayout.Lines[0].Width <= 50.0 {
		t.Errorf("Line should overflow (width > 50), got %.2f", node.TextLayout.Lines[0].Width)
	}
}

func TestTextOverflowClip(t *testing.T) {
	setupFakeMetrics()

	// Default behavior - text overflows without ellipsis
	text := "This is a very long text that will overflow"
	node := Text(text, Style{
		Width: 100,
		TextStyle: &TextStyle{
			FontSize:     16,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowClip, // Default
		},
	})

	constraints := Loose(100, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Should have one line that overflows
	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(node.TextLayout.Lines))
	}

	line := node.TextLayout.Lines[0]
	// Line should overflow (no truncation with clip)
	if line.Width <= 100.0 {
		t.Errorf("Line should overflow container (width > 100), got %.2f", line.Width)
	}

	// Should not contain ellipsis
	for _, box := range line.Boxes {
		if strings.Contains(box.Text, "...") {
			t.Error("Text-overflow: clip should not add ellipsis")
		}
	}
}

func TestTextOverflowEllipsis(t *testing.T) {
	setupFakeMetrics()

	// Text should be truncated with ellipsis
	text := "This is a very long text that will overflow"
	node := Text(text, Style{
		Width: 100,
		TextStyle: &TextStyle{
			FontSize:     16,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowEllipsis,
		},
	})

	constraints := Loose(100, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Should have one line
	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(node.TextLayout.Lines))
	}

	line := node.TextLayout.Lines[0]

	// Line should fit within container (truncated)
	if line.Width > 100.0 {
		t.Errorf("Line should fit within container (width <= 100), got %.2f", line.Width)
	}

	// Should contain ellipsis
	hasEllipsis := false
	for _, box := range line.Boxes {
		if box.Text == "..." {
			hasEllipsis = true
			break
		}
	}
	if !hasEllipsis {
		t.Error("Text-overflow: ellipsis should add '...'")
	}

	// Text should be truncated (not all boxes present)
	allText := ""
	for _, box := range line.Boxes {
		allText += box.Text
	}
	if !strings.HasSuffix(allText, "...") {
		t.Errorf("Text should end with ellipsis, got: %s", allText)
	}
	if allText == text+"..." {
		t.Error("Text should be truncated, not just have ellipsis appended")
	}
}

func TestTextOverflowEllipsisAlignRight(t *testing.T) {
	setupFakeMetrics()

	// Test ellipsis with right alignment - use wider container to see offset
	text := "This is a long text"
	node := Text(text, Style{
		Width: 200, // Wider container so truncated text has room for right-align offset
		TextStyle: &TextStyle{
			FontSize:     16,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowEllipsis,
			TextAlign:    TextAlignRight,
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// With wider container, text won't overflow so no ellipsis needed
	// Let's use a case that definitely overflows
	text2 := "This is a very long text that definitely overflows the container width"
	node2 := Text(text2, Style{
		Width: 150,
		TextStyle: &TextStyle{
			FontSize:     16,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowEllipsis,
			TextAlign:    TextAlignRight,
		},
	})

	constraints2 := Loose(150, 200)
	LayoutText(node2, constraints2)

	line2 := node2.TextLayout.Lines[0]

	// Should contain ellipsis
	hasEllipsis := false
	for _, box := range line2.Boxes {
		if box.Text == "..." {
			hasEllipsis = true
			break
		}
	}
	if !hasEllipsis {
		t.Error("Should have ellipsis even with right alignment")
	}

	// Line should fit within container (truncated)
	if line2.Width > 150.0 {
		t.Errorf("Line should fit within container, got %.2f", line2.Width)
	}

	// Right-aligned truncated text should have non-negative offset
	// (may be 0 if line fills most of the width)
	if line2.OffsetX < 0 {
		t.Errorf("Right-aligned text should have non-negative OffsetX, got %.2f", line2.OffsetX)
	}
}

func TestTextOverflowEllipsisVeryNarrow(t *testing.T) {
	setupFakeMetrics()

	// Container too narrow even for ellipsis
	text := "Hello world"
	node := Text(text, Style{
		Width: 20, // Very narrow - ellipsis is 30px (3 chars * 10px)
		TextStyle: &TextStyle{
			FontSize:     16,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowEllipsis,
		},
	})

	constraints := Loose(20, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	line := node.TextLayout.Lines[0]

	// Should show just ellipsis (or as much as fits)
	// Line should not exceed container width significantly
	if line.Width > 35.0 {
		t.Errorf("Very narrow container should show minimal content, got width %.2f", line.Width)
	}
}

func TestTextOverflowEllipsisNoOverflow(t *testing.T) {
	setupFakeMetrics()

	// Text fits - no ellipsis should be added
	text := "Short"
	node := Text(text, Style{
		Width: 200,
		TextStyle: &TextStyle{
			FontSize:     16,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowEllipsis,
		},
	})

	constraints := Loose(200, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	line := node.TextLayout.Lines[0]

	// Should NOT contain ellipsis (text fits)
	for _, box := range line.Boxes {
		if box.Text == "..." {
			t.Error("Should not add ellipsis when text fits")
		}
	}

	// Text should be complete
	allText := ""
	for _, box := range line.Boxes {
		allText += box.Text
	}
	if allText != text {
		t.Errorf("Text should be complete, expected '%s', got '%s'", text, allText)
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

// TestTextCJKWrapping tests that CJK (ideographic) text can wrap between characters
func TestTextCJKWrapping(t *testing.T) {
	setupFakeMetrics()

	// Chinese text: "你好世界" (Hello World)
	text := "你好世界"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize: 16,
		},
	})

	// Each CJK character is 10px with fake metrics (1 rune = 1 char)
	// Total width: 4 chars × 10px = 40px
	// With maxWidth=25px, should wrap into multiple lines
	constraints := Loose(25, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Should have multiple lines (at least 2)
	lineCount := len(node.TextLayout.Lines)
	if lineCount < 2 {
		t.Errorf("CJK text should wrap into multiple lines, got %d line(s)", lineCount)
	}

	// Each line should respect the width constraint
	for i, line := range node.TextLayout.Lines {
		if line.Width > 25.1 {
			t.Errorf("Line %d exceeds maxWidth: %.2f > 25", i, line.Width)
		}
	}
}

// TestTextMixedCJKEnglish tests wrapping of mixed CJK and English text
func TestTextMixedCJKEnglish(t *testing.T) {
	setupFakeMetrics()

	// Mixed text with English word and CJK characters
	text := "Hello你好World世界"
	node := Text(text, Style{
		TextStyle: &TextStyle{
			FontSize: 16,
		},
	})

	// Should allow breaks:
	// - After "Hello" (space)
	// - Between CJK characters
	// - After "World" (space implied by CJK boundary)
	constraints := Loose(60, 200)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	// Should wrap into multiple lines
	lineCount := len(node.TextLayout.Lines)
	if lineCount < 2 {
		t.Errorf("Mixed CJK/English text should wrap, got %d line(s)", lineCount)
	}
}

// TestTextIndentWithAlignment tests text-indent with different alignment modes
func TestTextIndentWithAlignment(t *testing.T) {
	setupFakeMetrics()

	tests := []struct {
		name       string
		align      TextAlign
		indent     float64
		expectFunc func(t *testing.T, offsetX, lineWidth, contentWidth float64)
	}{
		{
			name:   "left-align-positive-indent",
			align:  TextAlignLeft,
			indent: 20,
			expectFunc: func(t *testing.T, offsetX, lineWidth, contentWidth float64) {
				// Left-aligned with positive indent: line should start at indent position
				if offsetX != 20 {
					t.Errorf("Expected offsetX=20, got %.2f", offsetX)
				}
			},
		},
		{
			name:   "left-align-negative-indent",
			align:  TextAlignLeft,
			indent: -10,
			expectFunc: func(t *testing.T, offsetX, lineWidth, contentWidth float64) {
				// Left-aligned with negative indent: line should start before 0
				if offsetX != -10 {
					t.Errorf("Expected offsetX=-10, got %.2f", offsetX)
				}
			},
		},
		{
			name:   "right-align-positive-indent",
			align:  TextAlignRight,
			indent: 20,
			expectFunc: func(t *testing.T, offsetX, lineWidth, contentWidth float64) {
				// Right-aligned with positive indent: indent reduces available width
				// Content should be right-aligned within (contentWidth - indent)
				// Text should end at (contentWidth - indent) not contentWidth
				// So: offsetX = contentWidth - lineWidth - indent
				// Example: contentWidth=200, lineWidth=50, indent=20
				// Without indent: offsetX = 200 - 50 = 150 (ends at 200)
				// With indent: offsetX = 200 - 50 - 20 = 130 (ends at 180, leaving 20px on right)
				indent := 20.0
				expectedOffset := contentWidth - lineWidth - indent
				if math.Abs(offsetX-expectedOffset) > 0.1 {
					t.Errorf("Expected offsetX=%.2f (text ends at %.2f), got %.2f (text ends at %.2f)",
						expectedOffset, expectedOffset+lineWidth, offsetX, offsetX+lineWidth)
				}
			},
		},
		{
			name:   "center-align-positive-indent",
			align:  TextAlignCenter,
			indent: 20,
			expectFunc: func(t *testing.T, offsetX, lineWidth, contentWidth float64) {
				// Center-aligned with positive indent: indent reduces available width
				// Content should be centered within (contentWidth - indent)
				// So: offsetX = indent + (contentWidth - indent - lineWidth) / 2
				indent := 20.0
				availableWidth := contentWidth - indent
				expectedOffset := indent + (availableWidth-lineWidth)/2
				if math.Abs(offsetX-expectedOffset) > 0.1 {
					t.Errorf("Expected offsetX=%.2f, got %.2f", expectedOffset, offsetX)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := "Hello" // 50px wide with fake metrics
			node := Text(text, Style{
				Width: 200,
				TextStyle: &TextStyle{
					FontSize:   16,
					TextAlign:  tt.align,
					TextIndent: tt.indent,
				},
			})

			constraints := Loose(200, 200)
			LayoutText(node, constraints)

			if node.TextLayout == nil || len(node.TextLayout.Lines) == 0 {
				t.Fatal("TextLayout should have lines")
			}

			line := node.TextLayout.Lines[0]
			tt.expectFunc(t, line.OffsetX, line.Width, 200)
		})
	}
}
