package layout

import (
	"testing"
)

// TestCharacterOrientationHorizontal verifies that horizontal modes don't populate orientation data
func TestCharacterOrientationHorizontal(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	node := &Node{
		Text: "Hello world 世界",
		Style: Style{
			Display: DisplayInlineText,
			Width:   Px(0),
			Height:  Px(0),
			TextStyle: &TextStyle{
				FontSize:    16,
				WhiteSpace:  WhiteSpaceNormal,
				WritingMode: WritingModeHorizontalTB,
			},
		},
	}

	constraints := Loose(200, 100)
	LayoutText(node, constraints, ctx)

	if node.TextLayout == nil {
		t.Fatal("TextLayout is nil")
	}

	// In horizontal mode, orientations should be nil (no rotation needed)
	for i, line := range node.TextLayout.Lines {
		for j, box := range line.Boxes {
			if box.Orientations != nil {
				t.Errorf("Line %d, Box %d: Expected nil orientations in horizontal mode, got %v", i, j, box.Orientations)
			}
		}
	}
}

// TestCharacterOrientationVerticalCJK verifies that CJK characters are upright in vertical modes
func TestCharacterOrientationVerticalCJK(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	tests := []struct {
		name        string
		text        string
		writingMode WritingMode
		expectedAll bool // Expected orientation for all characters
	}{
		{
			name:        "CJK in Vertical-LR",
			text:        "世界",
			writingMode: WritingModeVerticalLR,
			expectedAll: true, // CJK should be upright
		},
		{
			name:        "CJK in Vertical-RL",
			text:        "日本語",
			writingMode: WritingModeVerticalRL,
			expectedAll: true, // CJK should be upright
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{
				Text: tt.text,
				Style: Style{
					Display: DisplayInlineText,
					Width:   Px(0),
					Height:  Px(0),
					TextStyle: &TextStyle{
						FontSize:    16,
						WhiteSpace:  WhiteSpaceNormal,
						WritingMode: tt.writingMode,
					},
				},
			}

			constraints := Loose(100, 200)
			LayoutText(node, constraints, ctx)

			if node.TextLayout == nil {
				t.Fatal("TextLayout is nil")
			}

			lines := node.TextLayout.Lines
			if len(lines) == 0 {
				t.Fatal("No lines generated")
			}

			// Check that all CJK characters are upright
			for i, line := range lines {
				for j, box := range line.Boxes {
					if box.Orientations == nil {
						t.Errorf("Line %d, Box %d: Expected orientations in vertical mode, got nil", i, j)
						continue
					}

					textRunes := []rune(box.Text)
					if len(box.Orientations) != len(textRunes) {
						t.Errorf("Line %d, Box %d: Orientations length %d doesn't match text runes %d",
							i, j, len(box.Orientations), len(textRunes))
						continue
					}

					for k, orientation := range box.Orientations {
						if orientation != tt.expectedAll {
							t.Errorf("Line %d, Box %d, Char %d ('%c'): Expected orientation %v, got %v",
								i, j, k, textRunes[k], tt.expectedAll, orientation)
						}
					}
				}
			}
		})
	}
}

// TestCharacterOrientationVerticalLatin verifies that Latin characters are rotated in vertical modes
func TestCharacterOrientationVerticalLatin(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	tests := []struct {
		name        string
		text        string
		writingMode WritingMode
		expectedAll bool // Expected orientation for all characters
	}{
		{
			name:        "Latin in Vertical-LR",
			text:        "Hello",
			writingMode: WritingModeVerticalLR,
			expectedAll: false, // Latin should be rotated
		},
		{
			name:        "Latin in Vertical-RL",
			text:        "World",
			writingMode: WritingModeVerticalRL,
			expectedAll: false, // Latin should be rotated
		},
		{
			name:        "Digits in Vertical-LR",
			text:        "12345",
			writingMode: WritingModeVerticalLR,
			expectedAll: false, // Digits should be rotated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{
				Text: tt.text,
				Style: Style{
					Display: DisplayInlineText,
					Width:   Px(0),
					Height:  Px(0),
					TextStyle: &TextStyle{
						FontSize:    16,
						WhiteSpace:  WhiteSpaceNormal,
						WritingMode: tt.writingMode,
					},
				},
			}

			constraints := Loose(100, 200)
			LayoutText(node, constraints, ctx)

			if node.TextLayout == nil {
				t.Fatal("TextLayout is nil")
			}

			lines := node.TextLayout.Lines
			if len(lines) == 0 {
				t.Fatal("No lines generated")
			}

			// Check that all Latin characters are rotated
			for i, line := range lines {
				for j, box := range line.Boxes {
					if box.Orientations == nil {
						t.Errorf("Line %d, Box %d: Expected orientations in vertical mode, got nil", i, j)
						continue
					}

					textRunes := []rune(box.Text)
					if len(box.Orientations) != len(textRunes) {
						t.Errorf("Line %d, Box %d: Orientations length %d doesn't match text runes %d",
							i, j, len(box.Orientations), len(textRunes))
						continue
					}

					for k, orientation := range box.Orientations {
						if orientation != tt.expectedAll {
							t.Errorf("Line %d, Box %d, Char %d ('%c'): Expected orientation %v, got %v",
								i, j, k, textRunes[k], tt.expectedAll, orientation)
						}
					}
				}
			}
		})
	}
}

// TestCharacterOrientationMixedScript verifies mixed CJK and Latin text
func TestCharacterOrientationMixedScript(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	node := &Node{
		Text: "Hello世界World日本",
		Style: Style{
			Display: DisplayInlineText,
			Width:   Px(0),
			Height:  Px(0),
			TextStyle: &TextStyle{
				FontSize:    16,
				WhiteSpace:  WhiteSpaceNormal,
				WritingMode: WritingModeVerticalRL,
			},
		},
	}

	constraints := Loose(100, 300)
	LayoutText(node, constraints, ctx)

	if node.TextLayout == nil {
		t.Fatal("TextLayout is nil")
	}

	// Expected orientations for "Hello世界World日本"
	// H e l l o = false (Latin, rotated)
	// 世 界 = true (CJK, upright)
	// W o r l d = false (Latin, rotated)
	// 日 本 = true (CJK, upright)
	expectedOrientations := map[rune]bool{
		'H': false, 'e': false, 'l': false, 'o': false,
		'世': true, '界': true,
		'W': false, 'r': false, 'd': false,
		'日': true, '本': true,
	}

	// Verify orientations match expected per character type
	for i, line := range node.TextLayout.Lines {
		for j, box := range line.Boxes {
			if box.Orientations == nil {
				t.Errorf("Line %d, Box %d: Expected orientations in vertical mode, got nil", i, j)
				continue
			}

			textRunes := []rune(box.Text)
			if len(box.Orientations) != len(textRunes) {
				t.Errorf("Line %d, Box %d: Orientations length %d doesn't match text runes %d",
					i, j, len(box.Orientations), len(textRunes))
				continue
			}

			for k, r := range textRunes {
				expected, ok := expectedOrientations[r]
				if !ok {
					// Unknown character, skip
					continue
				}

				if box.Orientations[k] != expected {
					t.Errorf("Line %d, Box %d, Char %d ('%c'): Expected orientation %v, got %v",
						i, j, k, r, expected, box.Orientations[k])
				}
			}
		}
	}
}

// TestCharacterOrientationSideways verifies that sideways modes rotate ALL characters
func TestCharacterOrientationSideways(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	tests := []struct {
		name        string
		text        string
		writingMode WritingMode
	}{
		{
			name:        "Mixed text in Sideways-RL",
			text:        "Hello世界",
			writingMode: WritingModeSidewaysRL,
		},
		{
			name:        "CJK in Sideways-LR",
			text:        "日本語",
			writingMode: WritingModeSidewaysLR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{
				Text: tt.text,
				Style: Style{
					Display: DisplayInlineText,
					Width:   Px(0),
					Height:  Px(0),
					TextStyle: &TextStyle{
						FontSize:    16,
						WhiteSpace:  WhiteSpaceNormal,
						WritingMode: tt.writingMode,
					},
				},
			}

			constraints := Loose(100, 200)
			LayoutText(node, constraints, ctx)

			if node.TextLayout == nil {
				t.Fatal("TextLayout is nil")
			}

			// In sideways modes, ALL characters should be rotated (false)
			for i, line := range node.TextLayout.Lines {
				for j, box := range line.Boxes {
					if box.Orientations == nil {
						t.Errorf("Line %d, Box %d: Expected orientations in sideways mode, got nil", i, j)
						continue
					}

					textRunes := []rune(box.Text)
					if len(box.Orientations) != len(textRunes) {
						t.Errorf("Line %d, Box %d: Orientations length %d doesn't match text runes %d",
							i, j, len(box.Orientations), len(textRunes))
						continue
					}

					for k, orientation := range box.Orientations {
						if orientation != false {
							t.Errorf("Line %d, Box %d, Char %d ('%c'): Expected rotated (false) in sideways mode, got %v",
								i, j, k, textRunes[k], orientation)
						}
					}
				}
			}
		})
	}
}

// TestComputeTextOrientations directly tests the helper function
func TestComputeTextOrientations(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		writingMode WritingMode
		expected    []bool
	}{
		{
			name:        "Horizontal mode returns nil",
			text:        "Hello",
			writingMode: WritingModeHorizontalTB,
			expected:    nil,
		},
		{
			name:        "Empty text returns nil",
			text:        "",
			writingMode: WritingModeVerticalLR,
			expected:    nil,
		},
		{
			name:        "Latin in vertical-rl",
			text:        "Hi",
			writingMode: WritingModeVerticalRL,
			expected:    []bool{false, false}, // Both rotated
		},
		{
			name:        "CJK in vertical-lr",
			text:        "世界",
			writingMode: WritingModeVerticalLR,
			expected:    []bool{true, true}, // Both upright
		},
		{
			name:        "Mixed in vertical-rl",
			text:        "A世",
			writingMode: WritingModeVerticalRL,
			expected:    []bool{false, true}, // Latin rotated, CJK upright
		},
		{
			name:        "Sideways-rl all rotated",
			text:        "A世",
			writingMode: WritingModeSidewaysRL,
			expected:    []bool{false, false}, // All rotated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeTextOrientations(tt.text, tt.writingMode)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("Expected %v, got nil", tt.expected)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Index %d: expected %v, got %v", i, expected, result[i])
				}
			}
		})
	}
}
