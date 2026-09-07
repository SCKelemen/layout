package layout

import "testing"

// TestLayoutSimpleUnconstrainedViewportUnits: LayoutSimple derives the
// viewport from the constraints; an unbounded constraint is not a viewport
// size, so vw/vh resolve to 0 instead of ~9e307 (CSS Values L4 §6.2 defines
// viewport units against the actual viewport, which is unknown here).
// https://www.w3.org/TR/css-values-4/#viewport-relative-lengths
func TestLayoutSimpleUnconstrainedViewportUnits(t *testing.T) {
	node := &Node{Style: Style{Width: Vw(50), Height: Vh(50)}}
	size := LayoutSimple(node, Unconstrained())
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("expected 0x0 under Unconstrained(), got %vx%v", size.Width, size.Height)
	}
	if node.Rect.Width >= Unbounded || node.Rect.Height >= Unbounded {
		t.Errorf("unbounded value leaked into Rect: %+v", node.Rect)
	}

	// Only the unbounded axis is affected.
	node = &Node{Style: Style{Width: Vw(50), Height: Vh(50)}}
	size = LayoutSimple(node, Loose(400, Unbounded))
	if size.Width != 200 {
		t.Errorf("expected width 200 (50vw of 400), got %v", size.Width)
	}
	if size.Height != 0 {
		t.Errorf("expected height 0 (50vh of an unknown viewport), got %v", size.Height)
	}

	// Bounded constraints keep working as a viewport.
	node = &Node{Style: Style{Width: Vw(50), Height: Vh(50)}}
	size = LayoutSimple(node, Loose(400, 300))
	if size.Width != 200 || size.Height != 150 {
		t.Errorf("expected 200x150, got %vx%v", size.Width, size.Height)
	}
}

// TestLayoutNilRoot: a nil root has no layout; Layout, LayoutSimple, and
// LayoutWithPositioning return a zero Size instead of panicking.
func TestLayoutNilRoot(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)
	if got := Layout(nil, Loose(400, 300), ctx); got != (Size{}) {
		t.Errorf("Layout(nil) = %+v, want zero Size", got)
	}
	if got := LayoutSimple(nil, Loose(400, 300)); got != (Size{}) {
		t.Errorf("LayoutSimple(nil) = %+v, want zero Size", got)
	}
	if got := LayoutWithPositioning(nil, Loose(400, 300), Rect{Width: 400, Height: 300}, ctx); got != (Size{}) {
		t.Errorf("LayoutWithPositioning(nil) = %+v, want zero Size", got)
	}
}
