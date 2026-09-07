package layout

import "testing"

// Orientations must follow the resolved writing mode, which prefers
// Style.WritingMode; previously only the legacy TextStyle.WritingMode was
// consulted, so a node using Style.WritingMode got vertical geometry but no
// per-glyph orientations.
func TestOrientationsFollowStyleWritingMode(t *testing.T) {
	n := &Node{Style: Style{Display: DisplayInlineText, WritingMode: WritingModeVerticalRL, TextStyle: &TextStyle{FontSize: 16}}, Text: "ab漢"}
	Layout(n, Loose(200, 200), NewLayoutContext(200, 200, 16))
	if n.TextLayout == nil || len(n.TextLayout.Lines) == 0 || len(n.TextLayout.Lines[0].Boxes) == 0 {
		t.Fatal("no text layout produced")
	}
	box := n.TextLayout.Lines[0].Boxes[0]
	if len(box.Orientations) != len([]rune(box.Text)) {
		t.Fatalf("orientations = %v for %q, want one entry per rune", box.Orientations, box.Text)
	}
	// Probe viewport units too (vi/vb should resolve against the viewport).
	ctx := NewLayoutContext(800, 600, 16)
	if got := ResolveLength(Length{Value: 10, Unit: "vi"}, ctx, 16); got != 80 {
		t.Errorf("10vi = %v, want 80", got)
	}
	if got := ResolveLength(Length{Value: 10, Unit: "vb"}, ctx, 16); got != 60 {
		t.Errorf("10vb = %v, want 60", got)
	}
}
