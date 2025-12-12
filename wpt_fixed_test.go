package layout

import (
	"math"
	"testing"
)

// Generated from WPT test: text-layout-fixed.html
// These tests are converted from Web Platform Tests

func TestWPT_Text_1(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "This is a very long text that should be truncated with ellipsis when it overflows the container"
	node := Text(text, Style{
		Width:  100.00,
		Height: 20.00,
		TextStyle: &TextStyle{
			FontSize:     12.00,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowEllipsis,
		},
	})

	constraints := Loose(100.00, 20.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("Expected 1 lines, got %d", len(node.TextLayout.Lines))
	}

	if math.Abs(node.Rect.Width-100.00) > 1.0 {
		t.Errorf("Width should be 100.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-20.00) > 1.0 {
		t.Errorf("Height should be 20.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_2(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "This text will be clipped without ellipsis"
	node := Text(text, Style{
		Width:  80.00,
		Height: 20.00,
		TextStyle: &TextStyle{
			FontSize:     12.00,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowClip,
		},
	})

	constraints := Loose(80.00, 20.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("Expected 1 lines, got %d", len(node.TextLayout.Lines))
	}

	if math.Abs(node.Rect.Width-80.00) > 1.0 {
		t.Errorf("Width should be 80.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-20.00) > 1.0 {
		t.Errorf("Height should be 20.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_3(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "Line one\nLine two with    spaces\nLine three"
	node := Text(text, Style{
		Width:  300.00,
		Height: 50.00,
		TextStyle: &TextStyle{
			FontSize:   14.00,
			WhiteSpace: WhiteSpacePreWrap,
		},
	})

	constraints := Loose(300.00, 50.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if len(node.TextLayout.Lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(node.TextLayout.Lines))
	}

	if math.Abs(node.Rect.Width-300.00) > 1.0 {
		t.Errorf("Width should be 300.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-50.00) > 1.0 {
		t.Errorf("Height should be 50.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_4(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "Preserved    spaces    and\nno wrapping at all"
	node := Text(text, Style{
		Width:  400.00,
		Height: 40.00,
		TextStyle: &TextStyle{
			FontSize:   12.00,
			WhiteSpace: WhiteSpacePre,
		},
	})

	constraints := Loose(400.00, 40.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if math.Abs(node.Rect.Width-400.00) > 1.0 {
		t.Errorf("Width should be 400.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-40.00) > 1.0 {
		t.Errorf("Height should be 40.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_5(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "supercalifragilisticexpialidocious"
	node := Text(text, Style{
		Width:  60.00,
		Height: 60.00,
		TextStyle: &TextStyle{
			FontSize:     10.00,
			OverflowWrap: OverflowWrapBreakWord,
		},
	})

	constraints := Loose(60.00, 60.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if math.Abs(node.Rect.Width-60.00) > 1.0 {
		t.Errorf("Width should be 60.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-60.00) > 1.0 {
		t.Errorf("Height should be 60.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_6(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "verylongwordwithnobreaks"
	node := Text(text, Style{
		Width:  70.00,
		Height: 50.00,
		TextStyle: &TextStyle{
			FontSize:  10.00,
			WordBreak: WordBreakBreakAll,
		},
	})

	constraints := Loose(70.00, 50.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if math.Abs(node.Rect.Width-70.00) > 1.0 {
		t.Errorf("Width should be 70.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-50.00) > 1.0 {
		t.Errorf("Height should be 50.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_7(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "Centered text here"
	node := Text(text, Style{
		Width:  200.00,
		Height: 30.00,
		TextStyle: &TextStyle{
			FontSize:  14.00,
			TextAlign: TextAlignCenter,
		},
	})

	constraints := Loose(200.00, 30.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if math.Abs(node.Rect.Width-200.00) > 1.0 {
		t.Errorf("Width should be 200.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-30.00) > 1.0 {
		t.Errorf("Height should be 30.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_8(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "Right-aligned text"
	node := Text(text, Style{
		Width:  180.00,
		Height: 25.00,
		TextStyle: &TextStyle{
			FontSize:  12.00,
			TextAlign: TextAlignRight,
		},
	})

	constraints := Loose(180.00, 25.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if math.Abs(node.Rect.Width-180.00) > 1.0 {
		t.Errorf("Width should be 180.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-25.00) > 1.0 {
		t.Errorf("Height should be 25.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_9(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "This centered text should truncate with ellipsis"
	node := Text(text, Style{
		Width:  120.00,
		Height: 30.00,
		TextStyle: &TextStyle{
			FontSize:     13.00,
			WhiteSpace:   WhiteSpaceNowrap,
			TextOverflow: TextOverflowEllipsis,
			TextAlign:    TextAlignCenter,
		},
	})

	constraints := Loose(120.00, 30.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if len(node.TextLayout.Lines) != 1 {
		t.Errorf("Expected 1 lines, got %d", len(node.TextLayout.Lines))
	}

	if math.Abs(node.Rect.Width-120.00) > 1.0 {
		t.Errorf("Width should be 120.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-30.00) > 1.0 {
		t.Errorf("Height should be 30.00, got %f", node.Rect.Height)
	}
}

func TestWPT_Text_10(t *testing.T) {
	// WPT text test converted to Go
	setupFakeMetrics()

	text := "First line with    extra    spaces\nSecond line preserved"
	node := Text(text, Style{
		Width:  400.00,
		Height: 45.00,
		TextStyle: &TextStyle{
			FontSize:   11.00,
			WhiteSpace: WhiteSpacePreLine,
		},
	})

	constraints := Loose(400.00, 45.00)
	LayoutText(node, constraints)

	if node.TextLayout == nil {
		t.Fatal("TextLayout should be populated")
	}

	if len(node.TextLayout.Lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(node.TextLayout.Lines))
	}

	if math.Abs(node.Rect.Width-400.00) > 1.0 {
		t.Errorf("Width should be 400.00, got %f", node.Rect.Width)
	}
	if math.Abs(node.Rect.Height-45.00) > 1.0 {
		t.Errorf("Height should be 45.00, got %f", node.Rect.Height)
	}
}
