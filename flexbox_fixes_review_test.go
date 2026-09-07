package layout

import "testing"

// Regression tests for flexbox review fixes. Each test references the CSS
// Flexible Box Layout Module Level 1 section it verifies.
// https://www.w3.org/TR/css-flexbox-1/

// reviewTextNode builds a raw text flex item with auto width and height. It
// deliberately does not use the Text helper so the test does not depend on how
// that helper seeds Width/Height.
func reviewTextNode(text string) *Node {
	return &Node{
		Style: Style{Display: DisplayInlineText},
		Text:  text,
	}
}

// reviewTextHeight returns the height a text node's line boxes occupy.
func reviewTextHeight(t *testing.T, n *Node) float64 {
	t.Helper()
	if n.TextLayout == nil {
		t.Fatalf("text node has no TextLayout")
	}
	lines := len(n.TextLayout.Lines)
	if lines == 0 {
		lines = 1
	}
	return float64(lines) * n.TextLayout.LineHeight
}

// §9.2 step 3E: a text flex item's flex base size is its max-content size. The
// item must be measured with the text algorithm, not the block algorithm.
// https://www.w3.org/TR/css-flexbox-1/#algo-main-item
func TestFlexboxTextItemMeasuredInRow(t *testing.T) {
	text := reviewTextNode("hello world")
	fixed := Fixed(50, 50)
	root := HStack(text, fixed)
	root.Style.AlignItems = AlignItemsFlexStart
	flexLayout(root, Loose(300, 300))

	if text.Rect.Width <= 0 || text.Rect.Width > 300 {
		t.Fatalf("text width: got %.3f, want in (0, 300]", text.Rect.Width)
	}
	if text.TextLayout == nil || len(text.TextLayout.Lines) != 1 {
		t.Fatalf("text should lay out on one line, got %+v", text.TextLayout)
	}
	flexApprox(t, "text height", text.Rect.Height, reviewTextHeight(t, text))
	flexApprox(t, "text x", text.Rect.X, 0)
	flexApprox(t, "fixed x", fixed.Rect.X, text.Rect.Width)
}

// With the default align-items: stretch the text item still takes its
// max-content main size; only its cross size is stretched (§9.4 step 11).
// https://www.w3.org/TR/css-flexbox-1/#algo-stretch
func TestFlexboxTextItemMeasuredInRowStretch(t *testing.T) {
	text := reviewTextNode("hello world")
	fixed := Fixed(50, 50)
	root := HStack(text, fixed)
	flexLayout(root, Loose(300, 300))

	if text.Rect.Width <= 0 || text.Rect.Width > 300 {
		t.Fatalf("text width: got %.3f, want in (0, 300]", text.Rect.Width)
	}
	flexApprox(t, "fixed x", fixed.Rect.X, text.Rect.Width)
	if text.Rect.Height < 50 {
		t.Errorf("stretched text height: got %.3f, want >= 50", text.Rect.Height)
	}
}

// Column container: the text item's main size is its line height, and the
// following item is placed directly below it.
// https://www.w3.org/TR/css-flexbox-1/#algo-main-item
func TestFlexboxTextItemMeasuredInColumn(t *testing.T) {
	text := reviewTextNode("hello world")
	fixed := Fixed(50, 50)
	root := VStack(text, fixed)
	flexLayout(root, Loose(300, 300))

	if text.Rect.Width <= 0 || text.Rect.Width > 300 {
		t.Fatalf("text width: got %.3f, want in (0, 300]", text.Rect.Width)
	}
	if text.Rect.Height <= 0 {
		t.Fatalf("text height: got %.3f, want > 0", text.Rect.Height)
	}
	flexApprox(t, "text height", text.Rect.Height, reviewTextHeight(t, text))
	flexApprox(t, "text y", text.Rect.Y, 0)
	flexApprox(t, "fixed y", fixed.Rect.Y, text.Rect.Height)
}

// §9.7: a text item that is shrunk below its max-content size must be laid out
// again at its used main size so its lines re-wrap.
// https://www.w3.org/TR/css-flexbox-1/#resolve-flexible-lengths
func TestFlexboxTextItemShrunkRewraps(t *testing.T) {
	text := reviewTextNode("hello world")
	fixed := Fixed(60, 50)
	// Keep the fixed item rigid so all negative free space goes to the text.
	fixed.Style.MinWidth = Px(60)
	root := HStack(text, fixed)
	flexLayout(root, Tight(100, 50))

	flexApprox(t, "text width", text.Rect.Width, 40)
	flexApprox(t, "fixed x", fixed.Rect.X, 40)
	flexApprox(t, "fixed width", fixed.Rect.Width, 60)
	if text.TextLayout == nil {
		t.Fatalf("text node has no TextLayout after re-layout")
	}
	if len(text.TextLayout.Lines) < 2 {
		t.Errorf("text should wrap onto multiple lines at width 40, got %d line(s)", len(text.TextLayout.Lines))
	}
	for i, line := range text.TextLayout.Lines {
		if line.Width > 40.01 && len(line.Boxes) > 1 {
			t.Errorf("line %d width %.3f exceeds the item's used main size 40", i, line.Width)
		}
	}
}

// §4.1: an absolutely positioned child of a flex container is not a flex item.
// It takes no main-axis slot and does not contribute to the container's size.
// https://www.w3.org/TR/css-flexbox-1/#abs-pos-items
func TestFlexboxAbsPosChildTakesNoMainAxisSlot(t *testing.T) {
	a := Fixed(50, 50)
	b := Fixed(30, 50)
	b.Style.Position = PositionAbsolute
	c := Fixed(50, 50)
	root := HStack(a, b, c)
	size := flexLayout(root, Unconstrained())

	flexApprox(t, "a x", a.Rect.X, 0)
	flexApprox(t, "c x", c.Rect.X, 50)
	flexApprox(t, "container width", size.Width, 100)
	flexApprox(t, "container height", size.Height, 50)

	// The abs-pos child is still laid out for its own size and sits at its
	// static position: the container's content-box origin.
	flexApprox(t, "b width", b.Rect.Width, 30)
	flexApprox(t, "b height", b.Rect.Height, 50)
	flexApprox(t, "b x", b.Rect.X, 0)
	flexApprox(t, "b y", b.Rect.Y, 0)
}

// position: fixed children are likewise out of flow (§4.1).
// https://www.w3.org/TR/css-flexbox-1/#abs-pos-items
func TestFlexboxFixedPosChildTakesNoMainAxisSlot(t *testing.T) {
	a := Fixed(50, 50)
	b := Fixed(30, 50)
	b.Style.Position = PositionFixed
	c := Fixed(50, 50)
	root := VStack(a, b, c)
	size := flexLayout(root, Unconstrained())

	flexApprox(t, "c y", c.Rect.Y, 50)
	flexApprox(t, "container height", size.Height, 100)
	flexApprox(t, "container width", size.Width, 50)
}

// The static position of an abs-pos child is the container's content-box
// start corner, inside padding and border, so the positioned pass can place
// it relative to that origin.
// https://www.w3.org/TR/css-flexbox-1/#abs-pos-items
func TestFlexboxAbsPosChildStaticPositionInsidePadding(t *testing.T) {
	a := Fixed(50, 50)
	b := Fixed(30, 20)
	b.Style.Position = PositionAbsolute
	root := HStack(a, b)
	root.Style.Padding = Uniform(Px(10))
	root.Style.Border = Uniform(Px(2))
	size := flexLayout(root, Unconstrained())

	flexApprox(t, "a x", a.Rect.X, 12)
	flexApprox(t, "b x", b.Rect.X, 12)
	flexApprox(t, "b y", b.Rect.Y, 12)
	flexApprox(t, "container width", size.Width, 50+24)
}

// A container whose only children are absolutely positioned has no flex items
// and therefore no content size of its own.
// https://www.w3.org/TR/css-flexbox-1/#abs-pos-items
func TestFlexboxOnlyAbsPosChildren(t *testing.T) {
	b := Fixed(30, 20)
	b.Style.Position = PositionAbsolute
	root := HStack(b)
	root.Style.Padding = Uniform(Px(5))
	size := flexLayout(root, Unconstrained())

	flexApprox(t, "container width", size.Width, 10)
	flexApprox(t, "container height", size.Height, 10)
	flexApprox(t, "b width", b.Rect.Width, 30)
	flexApprox(t, "b height", b.Rect.Height, 20)
}

// End to end: the positioned pass places the abs-pos child using its insets
// while the in-flow items are unaffected.
// https://www.w3.org/TR/css-flexbox-1/#abs-pos-items
func TestFlexboxAbsPosChildWithPositioningPass(t *testing.T) {
	a := Fixed(50, 50)
	b := Fixed(30, 50)
	b.Style.Position = PositionAbsolute
	b.Style.Left = Px(10)
	b.Style.Top = Px(5)
	c := Fixed(50, 50)
	root := HStack(a, b, c)
	ctx := NewLayoutContext(1920, 1080, 16)
	size := LayoutWithPositioning(root, Tight(200, 100), Rect{Width: 1920, Height: 1080}, ctx)

	flexApprox(t, "container width", size.Width, 200)
	flexApprox(t, "a x", a.Rect.X, 0)
	flexApprox(t, "c x", c.Rect.X, 50)
	flexApprox(t, "b x", b.Rect.X, 10)
	flexApprox(t, "b y", b.Rect.Y, 5)
	flexApprox(t, "b width", b.Rect.Width, 30)
}
