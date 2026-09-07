package layout

import "testing"

// CSS 2.1 §10.3.7: when left, width, and right are all set, the offset that is
// ignored depends on the direction of the containing block, not of the box.
// https://www.w3.org/TR/CSS21/visudet.html#abs-non-replaced-width
func TestAbsoluteOverConstrainedUsesContainingBlockDirection(t *testing.T) {
	abs := &Node{Style: Style{Position: PositionAbsolute, Width: Px(100), Height: Px(10), Left: Px(10), Right: Px(10)}}
	root := &Node{
		Style:    Style{Position: PositionRelative, Direction: DirectionRTL, Width: Px(400), Height: Px(100)},
		Children: []*Node{abs},
	}
	ctx := NewLayoutContext(400, 100, 16)
	LayoutWithPositioning(root, Tight(400, 100), Rect{Width: 400, Height: 100}, ctx)
	if abs.Rect.Width != 100 || abs.Rect.X != 290 {
		t.Fatalf("rtl containing block: got X=%v W=%v, want X=290 W=100 (left ignored)", abs.Rect.X, abs.Rect.Width)
	}

	// The box's own direction does not decide: ltr containing block, rtl box.
	abs2 := &Node{Style: Style{Position: PositionAbsolute, Direction: DirectionRTL, Width: Px(100), Height: Px(10), Left: Px(10), Right: Px(10)}}
	root2 := &Node{Style: Style{Position: PositionRelative, Width: Px(400), Height: Px(100)}, Children: []*Node{abs2}}
	LayoutWithPositioning(root2, Tight(400, 100), Rect{Width: 400, Height: 100}, ctx)
	if abs2.Rect.X != 10 {
		t.Fatalf("ltr containing block: got X=%v, want 10 (right ignored)", abs2.Rect.X)
	}
}

// An absolutely positioned box is the containing block of its descendants
// (CSS 2.1 §10.1), so when its used size differs from the size the flow pass
// gave it, the descendants are laid out again against the used size.
func TestAbsoluteChildrenRelaidOutAtUsedSize(t *testing.T) {
	inner := &Node{Style: Style{Display: DisplayBlock}}
	abs := &Node{Style: Style{Position: PositionAbsolute, Left: Px(0), Top: Px(0), Width: Px(300), Height: Px(50)}, Children: []*Node{inner}}
	root := &Node{Style: Style{Position: PositionRelative, Width: Px(200), Height: Px(100)}, Children: []*Node{abs}}
	ctx := NewLayoutContext(400, 400, 16)
	LayoutWithPositioning(root, Tight(200, 100), Rect{Width: 200, Height: 100}, ctx)
	if abs.Rect.Width != 300 || abs.Rect.X != 0 || abs.Rect.Y != 0 {
		t.Fatalf("abs rect = %+v, want 300 wide at (0,0)", abs.Rect)
	}
	if inner.Rect.Width != 300 {
		t.Fatalf("inner width = %v, want 300 (re-laid out against the used size)", inner.Rect.Width)
	}
}

// A unit-less Length literal keeps resolving as pixels; only the true zero
// value is "unset".
func TestResolveLengthUnitlessLiteralIsPixels(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)
	if got := ResolveLength(Length{Value: 100}, ctx, 16); got != 100 {
		t.Fatalf("Length{Value:100} resolved to %v, want 100", got)
	}
	if got := ResolveLength(Length{}, ctx, 16); got != 0 {
		t.Fatalf("Length{} resolved to %v, want 0", got)
	}
}
