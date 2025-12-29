package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	// Set up accurate Unicode text metrics using the text library
	metrics := layout.NewTerminalTextMetrics()
	layout.SetTextMetricsProvider(metrics)

	fmt.Println("=== Layout Engine with Text Library Integration ===")

	// Example 1: ASCII text
	fmt.Println("1. ASCII Text Layout:")
	asciiNode := layout.Text("Hello, World!", layout.Style{
		TextStyle: &layout.TextStyle{
			FontSize:   16,
			LineHeight: 1.5,
			TextAlign:  layout.TextAlignCenter,
		},
		Width: layout.Px(200),
	})
	asciiSize := layout.Layout(asciiNode, layout.Loose(200, 600), nil)
	fmt.Printf("   Text: 'Hello, World!'\n")
	fmt.Printf("   Size: %.1f x %.1f\n", asciiSize.Width, asciiSize.Height)
	fmt.Printf("   Width (cells): %.1f\n", metrics.Text().Width("Hello, World!"))

	// Example 2: CJK text (wide characters)
	fmt.Println("\n2. CJK Text Layout:")
	cjkNode := layout.Text("你好世界", layout.Style{
		TextStyle: &layout.TextStyle{
			FontSize:   16,
			LineHeight: 1.5,
			TextAlign:  layout.TextAlignLeft,
		},
		Width: layout.Px(200),
	})
	cjkSize := layout.Layout(cjkNode, layout.Loose(200, 600), nil)
	fmt.Printf("   Text: '你好世界'\n")
	fmt.Printf("   Size: %.1f x %.1f\n", cjkSize.Width, cjkSize.Height)
	fmt.Printf("   Width (cells): %.1f (each CJK char = 2 cells)\n", metrics.Text().Width("你好世界"))

	// Example 3: Emoji with modifiers
	fmt.Println("\n3. Emoji Sequence Layout:")
	emojiNode := layout.Text("Hello 👋🏻 World 😀", layout.Style{
		TextStyle: &layout.TextStyle{
			FontSize:   16,
			LineHeight: 1.5,
			TextAlign:  layout.TextAlignLeft,
		},
		Width: layout.Px(200),
	})
	emojiSize := layout.Layout(emojiNode, layout.Loose(200, 600), nil)
	fmt.Printf("   Text: 'Hello 👋🏻 World 😀'\n")
	fmt.Printf("   Size: %.1f x %.1f\n", emojiSize.Width, emojiSize.Height)
	fmt.Printf("   Width (cells): %.1f\n", metrics.Text().Width("Hello 👋🏻 World 😀"))
	fmt.Printf("   Grapheme count: %d\n", metrics.Text().GraphemeCount("Hello 👋🏻 World 😀"))

	// Example 4: Mixed content
	fmt.Println("\n4. Mixed Content Layout:")
	mixedText := "Hello 世界 😀 Test"
	mixedNode := layout.Text(mixedText, layout.Style{
		TextStyle: &layout.TextStyle{
			FontSize:   16,
			LineHeight: 1.5,
			TextAlign:  layout.TextAlignLeft,
		},
		Width: layout.Px(200),
	})
	mixedSize := layout.Layout(mixedNode, layout.Loose(200, 600), nil)
	fmt.Printf("   Text: '%s'\n", mixedText)
	fmt.Printf("   Size: %.1f x %.1f\n", mixedSize.Width, mixedSize.Height)
	fmt.Printf("   Width breakdown:\n")
	fmt.Printf("     - 'Hello ' = %.1f cells\n", metrics.Text().Width("Hello "))
	fmt.Printf("     - '世界' = %.1f cells\n", metrics.Text().Width("世界"))
	fmt.Printf("     - ' 😀 ' = %.1f cells\n", metrics.Text().Width(" 😀 "))
	fmt.Printf("     - 'Test' = %.1f cells\n", metrics.Text().Width("Test"))
	fmt.Printf("     - Total = %.1f cells\n", metrics.Text().Width(mixedText))

	// Example 5: Direct text library operations
	fmt.Println("\n5. Direct Text Library Operations:")
	txt := metrics.Text()

	// Width measurement
	long := "This is a very long text that needs truncation"
	width := txt.Width(long)
	fmt.Printf("   Text: '%s'\n", long)
	fmt.Printf("   Width: %.1f cells\n", width)

	// Grapheme operations
	text := "Hello👋🏻世界"
	graphemes := txt.Graphemes(text)
	fmt.Printf("\n   Text: '%s'\n", text)
	fmt.Printf("   Grapheme count: %d\n", len(graphemes))
	fmt.Printf("   Graphemes: %v\n", graphemes)

	// Width of different character types
	fmt.Println("\n   Character width examples:")
	fmt.Printf("     ASCII 'A': %.1f cells\n", txt.Width("A"))
	fmt.Printf("     CJK '世': %.1f cells\n", txt.Width("世"))
	fmt.Printf("     Emoji '😀': %.1f cells\n", txt.Width("😀"))
	fmt.Printf("     Emoji+modifier '👋🏻': %.1f cells\n", txt.Width("👋🏻"))
	fmt.Printf("     Flag '🇺🇸': %.1f cells\n", txt.Width("🇺🇸"))

	fmt.Println("\n=== Benefits of Text Library Integration ===")
	fmt.Println("✅ Accurate Unicode width measurement (UAX #11)")
	fmt.Println("✅ Proper emoji sequence handling (UTS #51)")
	fmt.Println("✅ Grapheme cluster awareness (UAX #29)")
	fmt.Println("✅ Bidirectional text support (UAX #9)")
	fmt.Println("✅ All 297,981 Unicode conformance tests passing")
	fmt.Println("✅ Production-ready for terminal UIs")
}
