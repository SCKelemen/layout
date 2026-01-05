package layout

import (
	"testing"
)

// TestVerticalWritingModeLinePositioning verifies that lines are positioned correctly
// for vertical writing modes (vertical-rl and vertical-lr)
func TestVerticalWritingModeLinePositioning(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	tests := []struct {
		name          string
		writingMode   WritingMode
		textAlign     TextAlign
		text          string
		maxInlineSize float64 // height for vertical modes
		lineHeight    float64
		minLines      int     // Minimum expected lines (text should wrap)
		checkLine0X   float64
		checkLine0Y   float64
	}{
		{
			name:          "Horizontal mode - baseline behavior",
			writingMode:   WritingModeHorizontalTB,
			textAlign:     TextAlignLeft,
			text:          "This is a long line that will wrap to multiple lines when constrained",
			maxInlineSize: 50, // Force wrapping
			lineHeight:    20,
			minLines:      2,
			checkLine0X:   0,
			checkLine0Y:   0,
		},
		{
			name:          "Vertical-LR - lines stack left to right",
			writingMode:   WritingModeVerticalLR,
			textAlign:     TextAlignLeft,
			text:          "This is a long line that will wrap to multiple lines when constrained",
			maxInlineSize: 50, // Force wrapping (height in vertical mode)
			lineHeight:    20,
			minLines:      2,
			checkLine0X:   0, // First line at X=0
			checkLine0Y:   0, // Aligned to top
		},
		{
			name:          "Vertical-RL - lines stack right to left",
			writingMode:   WritingModeVerticalRL,
			textAlign:     TextAlignLeft,
			text:          "This is a long line that will wrap to multiple lines when constrained",
			maxInlineSize: 50, // Force wrapping
			lineHeight:    20,
			minLines:      2,
			checkLine0X:   30, // First line at contentInlineSize - lineHeight = 50 - 20 = 30
			checkLine0Y:   0,  // Aligned to top
		},
		{
			name:          "Vertical-LR with center alignment",
			writingMode:   WritingModeVerticalLR,
			textAlign:     TextAlignCenter,
			text:          "Short text versus longer text content here",
			maxInlineSize: 50,
			lineHeight:    20,
			minLines:      2,
			checkLine0X:   0, // First line at X=0
			checkLine0Y:   -1, // Y should be centered (computed based on line width)
		},
		{
			name:          "Vertical-RL with right alignment",
			writingMode:   WritingModeVerticalRL,
			textAlign:     TextAlignRight,
			text:          "This is a long line that will wrap to multiple lines when constrained",
			maxInlineSize: 50,
			lineHeight:    20,
			minLines:      2,
			checkLine0X:   30, // First line at contentInlineSize - lineHeight
			checkLine0Y:   -1, // Y should be at bottom (contentInlineSize - lineWidth)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{
				Text: tt.text,
				Style: Style{
					Display: DisplayInlineText,
					Width:   Px(0), // auto
					Height:  Px(0), // auto
					TextStyle: &TextStyle{
						FontSize:    16,
						TextAlign:   tt.textAlign,
						LineHeight:  tt.lineHeight,
						WhiteSpace:  WhiteSpaceNormal,
						Direction:   DirectionLTR,
						WritingMode: tt.writingMode,
					},
				},
			}

			// Use appropriate constraint based on writing mode
			var constraints Constraints
			if tt.writingMode.IsVertical() {
				// Vertical: maxInlineSize is height
				constraints = Loose(500, tt.maxInlineSize)
			} else {
				// Horizontal: maxInlineSize is width
				constraints = Loose(tt.maxInlineSize, 500)
			}

			LayoutText(node, constraints, ctx)

			if node.TextLayout == nil {
				t.Fatal("TextLayout is nil")
			}

			lines := node.TextLayout.Lines
			if len(lines) < tt.minLines {
				t.Fatalf("Expected at least %d lines, got %d", tt.minLines, len(lines))
			}

			// Check first line position
			line0 := lines[0]
			if tt.checkLine0X >= 0 && line0.OffsetX != tt.checkLine0X {
				t.Errorf("Line 0 OffsetX: expected %.2f, got %.2f", tt.checkLine0X, line0.OffsetX)
			}
			if tt.checkLine0Y >= 0 && line0.OffsetY != tt.checkLine0Y {
				t.Errorf("Line 0 OffsetY: expected %.2f, got %.2f", tt.checkLine0Y, line0.OffsetY)
			}

			// Check second line position and direction if multiple lines
			if len(lines) > 1 {
				line1 := lines[1]

				// Verify lines move in the correct direction
				if tt.writingMode == WritingModeVerticalLR || tt.writingMode == WritingModeSidewaysLR {
					// Lines should move rightward (X increases)
					if line1.OffsetX <= line0.OffsetX {
						t.Errorf("Vertical-LR: Line 1 X (%.2f) should be greater than Line 0 X (%.2f)", line1.OffsetX, line0.OffsetX)
					}
					// Verify spacing is lineHeight
					if line1.OffsetX != line0.OffsetX+tt.lineHeight {
						t.Errorf("Vertical-LR: Expected line spacing of %.2f, got %.2f", tt.lineHeight, line1.OffsetX-line0.OffsetX)
					}
				} else if tt.writingMode == WritingModeVerticalRL || tt.writingMode == WritingModeSidewaysRL {
					// Lines should move leftward (X decreases)
					if line1.OffsetX >= line0.OffsetX {
						t.Errorf("Vertical-RL: Line 1 X (%.2f) should be less than Line 0 X (%.2f)", line1.OffsetX, line0.OffsetX)
					}
					// Verify spacing is lineHeight
					if line0.OffsetX-line1.OffsetX != tt.lineHeight {
						t.Errorf("Vertical-RL: Expected line spacing of %.2f, got %.2f", tt.lineHeight, line0.OffsetX-line1.OffsetX)
					}
				} else if tt.writingMode == WritingModeHorizontalTB {
					// Lines should move downward (Y increases)
					if line1.OffsetY <= line0.OffsetY {
						t.Errorf("Horizontal: Line 1 Y (%.2f) should be greater than Line 0 Y (%.2f)", line1.OffsetY, line0.OffsetY)
					}
					// Verify spacing is lineHeight
					if line1.OffsetY != line0.OffsetY+tt.lineHeight {
						t.Errorf("Horizontal: Expected line spacing of %.2f, got %.2f", tt.lineHeight, line1.OffsetY-line0.OffsetY)
					}
				}
			}
		})
	}
}

// TestVerticalWritingModeTextIndent verifies text-indent works in vertical modes
func TestVerticalWritingModeTextIndent(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	node := &Node{
		Text: "This is a long line that will wrap to multiple lines when constrained by height",
		Style: Style{
			Display: DisplayInlineText,
			Width:   Px(0),
			Height:  Px(0),
			TextStyle: &TextStyle{
				FontSize:    16,
				TextAlign:   TextAlignLeft,
				TextIndent:  20,
				LineHeight:  20,
				WhiteSpace:  WhiteSpaceNormal,
				Direction:   DirectionLTR,
				WritingMode: WritingModeVerticalLR,
			},
		},
	}

	constraints := Loose(500, 50) // Constrain height to force wrapping
	LayoutText(node, constraints, ctx)

	if node.TextLayout == nil {
		t.Fatal("TextLayout is nil")
	}

	lines := node.TextLayout.Lines
	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines, got %d", len(lines))
	}

	// In vertical mode, text-indent affects the inline-axis (Y)
	// First line should have indent, second line should not
	line0 := lines[0]
	line1 := lines[1]

	if line0.OffsetY != 20 {
		t.Errorf("Line 0 OffsetY (with indent): expected 20, got %.2f", line0.OffsetY)
	}
	if line1.OffsetY != 0 {
		t.Errorf("Line 1 OffsetY (without indent): expected 0, got %.2f", line1.OffsetY)
	}

	// X positions should progress correctly (left-to-right)
	if line1.OffsetX != line0.OffsetX+20 {
		t.Errorf("Line 1 should be 20px to the right of Line 0, got X0=%.2f, X1=%.2f", line0.OffsetX, line1.OffsetX)
	}
}

// TestVerticalWritingModeSidewaysRL verifies sideways-rl mode
func TestVerticalWritingModeSidewaysRL(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	node := &Node{
		Text: "This is a long line that will wrap to multiple lines when constrained",
		Style: Style{
			Display: DisplayInlineText,
			Width:   Px(0),
			Height:  Px(0),
			TextStyle: &TextStyle{
				FontSize:    16,
				TextAlign:   TextAlignLeft,
				LineHeight:  20,
				WhiteSpace:  WhiteSpaceNormal,
				Direction:   DirectionLTR,
				WritingMode: WritingModeSidewaysRL,
			},
		},
	}

	constraints := Loose(500, 50) // Constrain height to force wrapping
	LayoutText(node, constraints, ctx)

	if node.TextLayout == nil {
		t.Fatal("TextLayout is nil")
	}

	lines := node.TextLayout.Lines
	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines, got %d", len(lines))
	}

	// Sideways-RL should behave like vertical-rl for line positioning
	// Lines stack right-to-left (X decreases)
	line0 := lines[0]
	line1 := lines[1]

	if line1.OffsetX >= line0.OffsetX {
		t.Errorf("Sideways-RL: Line 1 X (%.2f) should be less than Line 0 X (%.2f)", line1.OffsetX, line0.OffsetX)
	}
}
