package layout

import (
	"math"
	"testing"
)

// Regression tests for the grid review fixes. Each test names the bug it
// guards against and links the governing CSS Grid Layout Module Level 1
// section. Helpers gridFixLayout / gridFixExpectRect live in
// grid_fixes_test.go.

// gridReviewText is a wrappable sentence: its min-content width (longest
// word) is far smaller than its max-content width (the whole line).
const gridReviewText = "The quick brown fox jumps over the lazy dog while the cat watches from the window sill quietly today"

func gridReviewTextNode() *Node {
	return &Node{Style: Style{Display: DisplayInlineText}, Text: gridReviewText}
}

// Review fix 1: fr and auto tracks used the max-content contribution as the
// track's base size, so a text item forced its column to the full unwrapped
// line width (about 960px in a 400px grid) and pushed its sibling out of the
// container. §12.5: the base size of a track with an auto minimum is the
// items' minimum (min-content) contribution; the max-content contribution is
// only the growth limit. §12.7.1 then shares the definite free space between
// the fr tracks.
// https://www.w3.org/TR/css-grid-1/#algo-single-span-items
// https://www.w3.org/TR/css-grid-1/#algo-find-fr-size
func TestGridReviewFrTracksUseMinimumContribution(t *testing.T) {
	text := gridReviewTextNode()
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			Width:               Px(400),
			GridTemplateColumns: []GridTrack{FractionTrack(1), FractionTrack(1)},
		},
		Children: []*Node{text, Fixed(50, 50)},
	}
	size := gridFixLayout(t, root, Loose(800, Unbounded))

	// The text wraps inside a 200px column: compare against a standalone
	// text layout at that width.
	probeCtx := NewLayoutContext(800, 600, 16)
	wrapped := LayoutText(gridReviewTextNode(), Loose(200, Unbounded), probeCtx)
	single := LayoutText(gridReviewTextNode(), Loose(Unbounded, Unbounded), probeCtx)
	if wrapped.Height <= single.Height {
		// Sanity: the probe must produce several lines at 200px.
		t.Fatalf("probe text did not wrap at 200px: wrapped %v, single line %v", wrapped, single)
	}

	gridFixExpectRect(t, "text item (column 0)", text, 0, 0, 200, wrapped.Height)
	gridFixExpectRect(t, "fixed item (column 1)", root.Children[1], 200, 0, 50, 50)
	if math.Abs(size.Width-400) > 0.01 {
		t.Errorf("container width: expected 400, got %.2f", size.Width)
	}
	// Row height is the text's height at its resolved 200px column width,
	// not at an unconstrained width (§12.5 for the block axis).
	if math.Abs(size.Height-wrapped.Height) > 0.01 {
		t.Errorf("container height: expected %.2f (text wrapped at 200px), got %.2f", wrapped.Height, size.Height)
	}
}

// Review fix 1 (auto tracks): a single auto column in a definite 400px grid
// grows from its min-content base toward its max-content growth limit but
// stops at the available space (§12.6 maximize tracks), instead of being
// sized to the unwrapped text width.
// https://www.w3.org/TR/css-grid-1/#algo-grow-tracks
func TestGridReviewAutoTrackClampsToAvailableSpace(t *testing.T) {
	text := gridReviewTextNode()
	root := &Node{
		Style:    Style{Display: DisplayGrid, Width: Px(400)},
		Children: []*Node{text},
	}
	size := gridFixLayout(t, root, Loose(800, Unbounded))

	probe := gridReviewTextNode()
	wrapped := LayoutText(probe, Loose(400, Unbounded), NewLayoutContext(800, 600, 16))

	gridFixExpectRect(t, "text in auto column", text, 0, 0, 400, wrapped.Height)
	if math.Abs(size.Width-400) > 0.01 || math.Abs(size.Height-wrapped.Height) > 0.01 {
		t.Errorf("container: expected 400x%.2f, got %.2fx%.2f", wrapped.Height, size.Width, size.Height)
	}

	// With less space than the min-content size the track still holds the
	// longest word (base size) and overflows, rather than collapsing.
	minContent := calculateTextMinContentWidth(gridReviewTextNode(), NewLayoutContext(800, 600, 16))
	narrow := &Node{
		Style:    Style{Display: DisplayGrid, Width: Px(20)},
		Children: []*Node{gridReviewTextNode()},
	}
	gridFixLayout(t, narrow, Loose(800, Unbounded))
	if math.Abs(narrow.Children[0].Rect.Width-minContent) > 0.01 {
		t.Errorf("auto column below min-content: expected %.2f (longest word), got %.2f", minContent, narrow.Children[0].Rect.Width)
	}
}

// Review fix 1 (non-text): an auto column whose item has a definite width
// still sizes to that width when it fits (min- and max-content contributions
// coincide for a definite size), so existing content-sized layouts are
// unchanged.
// https://www.w3.org/TR/css-grid-1/#algo-content
func TestGridReviewAutoTrackKeepsMaxContentWhenItFits(t *testing.T) {
	root := &Node{
		Style:    Style{Display: DisplayGrid, Width: Px(400)},
		Children: []*Node{Fixed(100, 100)},
	}
	size := gridFixLayout(t, root, Loose(800, Unbounded))
	gridFixExpectRect(t, "fixed item in auto column", root.Children[0], 0, 0, 100, 100)
	if math.Abs(size.Width-400) > 0.01 {
		t.Errorf("container width: expected 400 (explicit), got %.2f", size.Width)
	}

	// auto next to 1fr: the auto column takes its max-content size and the
	// fr column the rest (flexible tracks do not grow in §12.6). The second
	// item has an auto width so it stretches to reveal its track size.
	root = &Node{
		Style: Style{
			Display:             DisplayGrid,
			Width:               Px(400),
			GridTemplateColumns: []GridTrack{AutoTrack(), FractionTrack(1)},
		},
		Children: []*Node{Fixed(100, 50), {Style: Style{Height: Px(50)}}},
	}
	gridFixLayout(t, root, Loose(800, Unbounded))
	gridFixExpectRect(t, "auto column", root.Children[0], 0, 0, 100, 50)
	gridFixExpectRect(t, "fr column", root.Children[1], 100, 0, 300, 50)
}

// Review fix 1 (indefinite space): under a max-content constraint the free
// space is infinite, so an auto track still grows to its max-content growth
// limit and the text does not wrap (§12.6).
// https://www.w3.org/TR/css-grid-1/#algo-grow-tracks
func TestGridReviewAutoTrackIndefiniteSpaceIsMaxContent(t *testing.T) {
	text := gridReviewTextNode()
	root := &Node{
		Style:    Style{Display: DisplayGrid},
		Children: []*Node{text},
	}
	size := gridFixLayout(t, root, Loose(Unbounded, Unbounded))

	maxContent := calculateTextMaxContentWidth(gridReviewTextNode(), NewLayoutContext(800, 600, 16))
	if math.Abs(text.Rect.Width-maxContent) > 0.01 || math.Abs(size.Width-maxContent) > 0.01 {
		t.Errorf("indefinite auto column: expected %.2f (max-content), got item %.2f / container %.2f", maxContent, text.Rect.Width, size.Width)
	}
}

// Review fix 2: absolutely positioned children were placed like grid items,
// occupying a cell and advancing the auto-placement cursor, so the following
// in-flow item wrapped to the next row. §9: an absolutely positioned child
// of a grid container is not a grid item and takes part in neither placement
// nor track sizing.
// https://www.w3.org/TR/css-grid-1/#abspos-items
func TestGridReviewAbsolutelyPositionedChildrenAreNotGridItems(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)
	a := Fixed(50, 50)
	b := &Node{Style: Style{Position: PositionAbsolute, Width: Px(20), Height: Px(20), Left: Px(5), Top: Px(5)}}
	c := Fixed(50, 50)
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
		},
		Children: []*Node{a, b, c},
	}
	size := LayoutWithPositioning(root, Loose(800, Unbounded), Rect{X: 0, Y: 0, Width: 800, Height: 600}, ctx)

	gridFixExpectRect(t, "item A", a, 0, 0, 50, 50)
	// C follows A in row 0, column 1: B did not consume a cell.
	gridFixExpectRect(t, "item C", c, 100, 0, 50, 50)
	// B is positioned by its offsets against the container's padding box.
	gridFixExpectRect(t, "abs-pos B", b, 5, 5, 20, 20)
	// B did not add an implicit row either.
	if math.Abs(size.Height-50) > 0.01 {
		t.Errorf("container height: expected 50 (one row), got %.2f", size.Height)
	}

	// Fixed positioning is out of flow as well.
	b.Style.Position = PositionFixed
	root2 := &Node{Style: root.Style, Children: []*Node{Fixed(50, 50), b, Fixed(50, 50)}}
	gridFixLayout(t, root2, Loose(800, Unbounded))
	gridFixExpectRect(t, "item C after fixed child", root2.Children[2], 100, 0, 50, 50)
}

// Review fix 2 (track sizing and static position): an absolutely positioned
// child does not contribute to intrinsic track sizes, is still laid out for
// its own size, and with auto offsets rests at its static position, the
// container's content-box origin (§9 static position).
// https://www.w3.org/TR/css-grid-1/#static-position
func TestGridReviewAbsolutelyPositionedChildStaticPosition(t *testing.T) {
	abs := &Node{Style: Style{Position: PositionAbsolute, Width: Px(300), Height: Px(30)}}
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			Padding: Uniform(Px(10)),
			Border:  Uniform(Px(2)),
		},
		Children: []*Node{Fixed(50, 50), abs},
	}
	size := gridFixLayout(t, root, Loose(Unbounded, Unbounded))

	// The auto column is sized by the in-flow item only (50), not by the
	// 300px absolutely positioned child.
	if math.Abs(size.Width-(50+24)) > 0.01 || math.Abs(size.Height-(50+24)) > 0.01 {
		t.Errorf("container: expected 74x74 (50 + padding/border), got %.2fx%.2f", size.Width, size.Height)
	}
	gridFixExpectRect(t, "in-flow item", root.Children[0], 12, 12, 50, 50)
	// Laid out for its own size and left at the content-box origin.
	gridFixExpectRect(t, "abs-pos child", abs, 12, 12, 300, 30)
}
