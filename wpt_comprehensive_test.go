package layout

import (
	"math"
	"testing"
)

// Generated from WPT test: text-comprehensive.html
// These tests are converted from Web Platform Tests

func TestWPT_Text_1(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "Normal   white   space    with    extra spaces"
	node := Text(text, Style{
		Width:  500.00,
		Height: 30.00,
		TextStyle: &TextStyle{
			FontSize:   14.00,
			WhiteSpace: WhiteSpaceNormal,
		},
	})

	constraints := Loose(500.00, 30.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if math.Abs(node.Rect.Width-500.00) > 1.0 {
		t.Errorf("Width should be 500.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-30.00) > 1.0 {
		t.Errorf("Height should be 30.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_2(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "Text that would normally wrap but doesn't with nowrap set"
	node := Text(text, Style{
		Width:  300.00,
		Height: 25.00,
		TextStyle: &TextStyle{
			FontSize:   12.00,
			WhiteSpace: WhiteSpaceNowrap,
		},
	})

	constraints := Loose(300.00, 25.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("Expected 1 lines, got %d", len(node.TextLayout.Lines))
	}

	if math.Abs(node.Rect.Width-300.00) > 1.0 {
		t.Errorf("Width should be 300.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-25.00) > 1.0 {
		t.Errorf("Height should be 25.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_3(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "This text will be justified across the full width of the container"
	node := Text(text, Style{
		Width:  250.00,
		Height: 35.00,
		TextStyle: &TextStyle{
			FontSize:  15.00,
			TextAlign: TextAlignJustify,
		},
	})

	constraints := Loose(250.00, 35.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if math.Abs(node.Rect.Width-250.00) > 1.0 {
		t.Errorf("Width should be 250.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-35.00) > 1.0 {
		t.Errorf("Height should be 35.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_4(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "Center aligned text"
	node := Text(text, Style{
		Width:  220.00,
		Height: 28.00,
		TextStyle: &TextStyle{
			FontSize:  13.00,
			TextAlign: TextAlignCenter,
		},
	})

	constraints := Loose(220.00, 28.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if math.Abs(node.Rect.Width-220.00) > 1.0 {
		t.Errorf("Width should be 220.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-28.00) > 1.0 {
		t.Errorf("Height should be 28.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_5(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "A very long text that definitely needs ellipsis"
	node := Text(text, Style{
		Width:  90.00,
		Height: 22.00,
		TextStyle: &TextStyle{
			FontSize:     11.00,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowEllipsis,
		},
	})

	constraints := Loose(90.00, 22.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("Expected 1 lines, got %d", len(node.TextLayout.Lines))
	}

	if math.Abs(node.Rect.Width-90.00) > 1.0 {
		t.Errorf("Width should be 90.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-22.00) > 1.0 {
		t.Errorf("Height should be 22.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_6(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "Text that gets clipped without ellipsis"
	node := Text(text, Style{
		Width:  110.00,
		Height: 24.00,
		TextStyle: &TextStyle{
			FontSize:     12.00,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowClip,
		},
	})

	constraints := Loose(110.00, 24.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("Expected 1 lines, got %d", len(node.TextLayout.Lines))
	}

	if math.Abs(node.Rect.Width-110.00) > 1.0 {
		t.Errorf("Width should be 110.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-24.00) > 1.0 {
		t.Errorf("Height should be 24.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_7(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "ThisIsAVeryLongWordWithoutSpaces"
	node := Text(text, Style{
		Width:  75.00,
		Height: 55.00,
		TextStyle: &TextStyle{
			FontSize:  11.00,
			WordBreak: WordBreakBreakAll,
		},
	})

	constraints := Loose(75.00, 55.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if math.Abs(node.Rect.Width-75.00) > 1.0 {
		t.Errorf("Width should be 75.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-55.00) > 1.0 {
		t.Errorf("Height should be 55.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_8(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "antidisestablishmentarianism"
	node := Text(text, Style{
		Width:  65.00,
		Height: 50.00,
		TextStyle: &TextStyle{
			FontSize:     10.00,
			OverflowWrap: OverflowWrapBreakWord,
		},
	})

	constraints := Loose(65.00, 50.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if math.Abs(node.Rect.Width-65.00) > 1.0 {
		t.Errorf("Width should be 65.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-50.00) > 1.0 {
		t.Errorf("Height should be 50.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_9(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "Right-aligned text with ellipsis if too long"
	node := Text(text, Style{
		Width:  140.00,
		Height: 32.00,
		TextStyle: &TextStyle{
			FontSize:     12.00,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowEllipsis,
			TextAlign:    TextAlignRight,
		},
	})

	constraints := Loose(140.00, 32.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("Expected 1 lines, got %d", len(node.TextLayout.Lines))
	}

	if math.Abs(node.Rect.Width-140.00) > 1.0 {
		t.Errorf("Width should be 140.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-32.00) > 1.0 {
		t.Errorf("Height should be 32.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_10(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "First line    with    collapsed spaces\nSecond line    also collapsed\nThird line too"
	node := Text(text, Style{
		Width:  350.00,
		Height: 60.00,
		TextStyle: &TextStyle{
			FontSize:   13.00,
			WhiteSpace: WhiteSpacePreLine,
		},
	})

	constraints := Loose(350.00, 60.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if len(node.TextLayout.Lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(node.TextLayout.Lines))
	}

	if math.Abs(node.Rect.Width-350.00) > 1.0 {
		t.Errorf("Width should be 350.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-60.00) > 1.0 {
		t.Errorf("Height should be 60.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_11(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "This line has preserved   spaces   and justification\nSecond line also    preserved"
	node := Text(text, Style{
		Width:  600.00,
		Height: 45.00,
		TextStyle: &TextStyle{
			FontSize:   14.00,
			WhiteSpace: WhiteSpacePreWrap,
			TextAlign:  TextAlignJustify,
		},
	})

	constraints := Loose(600.00, 45.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if len(node.TextLayout.Lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(node.TextLayout.Lines))
	}

	if math.Abs(node.Rect.Width-600.00) > 1.0 {
		t.Errorf("Width should be 600.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-45.00) > 1.0 {
		t.Errorf("Height should be 45.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_12(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "VeryTinyContainer"
	node := Text(text, Style{
		Width:  50.00,
		Height: 40.00,
		TextStyle: &TextStyle{
			FontSize:     10.00,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowEllipsis,
		},
	})

	constraints := Loose(50.00, 40.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("Expected 1 lines, got %d", len(node.TextLayout.Lines))
	}

	if math.Abs(node.Rect.Width-50.00) > 1.0 {
		t.Errorf("Width should be 50.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-40.00) > 1.0 {
		t.Errorf("Height should be 40.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_13(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "Short"
	node := Text(text, Style{
		Width:  800.00,
		Height: 25.00,
		TextStyle: &TextStyle{
			FontSize:   14.00,
			WhiteSpace: WhiteSpaceNowrap,
		},
	})

	constraints := Loose(800.00, 25.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("Expected 1 lines, got %d", len(node.TextLayout.Lines))
	}

	if math.Abs(node.Rect.Width-800.00) > 1.0 {
		t.Errorf("Width should be 800.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-25.00) > 1.0 {
		t.Errorf("Height should be 25.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_14(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "CenteredTextThatBreaksAtEveryCharacterBecauseBreakAll"
	node := Text(text, Style{
		Width:  180.00,
		Height: 70.00,
		TextStyle: &TextStyle{
			FontSize:  11.00,
			TextAlign: TextAlignCenter,
			WordBreak: WordBreakBreakAll,
		},
	})

	constraints := Loose(180.00, 70.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if math.Abs(node.Rect.Width-180.00) > 1.0 {
		t.Errorf("Width should be 180.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-70.00) > 1.0 {
		t.Errorf("Height should be 70.00, got %f", node.Rect.Height)
	}
}
