package layout

import "testing"

// Regression tests for container-relative units on the plain ResolveLength
// path and for query containers measuring their content box
// (CSS Containment Level 3 §5, https://www.w3.org/TR/css-contain-3/#container-lengths).

// TestLengthFixContainerUnitsResolveToZeroWithoutContainer: ResolveLength has
// no ancestor information, so cq* must resolve to 0 instead of leaking the raw
// value as pixels.
func TestLengthFixContainerUnitsResolveToZeroWithoutContainer(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)
	for _, l := range []Length{Cqw(50), Cqh(50), Cqi(50), Cqb(50), Cqmin(50), Cqmax(50)} {
		if got := ResolveLength(l, ctx, 16); got != 0 {
			t.Errorf("ResolveLength(%v%s) without a container: expected 0, got %v", l.Value, l.Unit, got)
		}
	}
	// A block using cq* for its width lays out as 0 wide, not 50px.
	node := &Node{Style: Style{Display: DisplayBlock, Width: Cqw(50), Height: Px(10)}}
	size := Layout(node, Loose(500, 500), ctx)
	if size.Width != 0 {
		t.Errorf("block with Width Cqw(50) and no container: expected 0, got %v", size.Width)
	}
}

// TestLengthFixContainerUnitsUseContentBox: container query lengths are a
// percentage of the query container's content box, so padding and border are
// excluded from the container's Rect.
func TestLengthFixContainerUnitsUseContentBox(t *testing.T) {
	rootNode := &Node{Rect: Rect{Width: 1000, Height: 1000}}
	container := &Node{
		Style: Style{
			ContainerType: ContainerTypeSize,
			Padding:       Uniform(Px(50)),
			Border:        Uniform(Px(10)),
		},
		Rect: Rect{Width: 800, Height: 600},
	}
	leafNode := &Node{}
	rootNode.Children = []*Node{container}
	container.Children = []*Node{leafNode}
	leaf := NewContext(rootNode).Children()[0].Children()[0]

	ctx := NewLayoutContext(1920, 1080, 16)
	// Content box: 800 - 2*(50+10) = 680 wide, 600 - 120 = 480 tall.
	if got := ResolveLengthInContext(Cqw(50), ctx, 16, leaf); got != 340 {
		t.Errorf("Cqw(50) of 680px content width: expected 340, got %v", got)
	}
	if got := ResolveLengthInContext(Cqh(50), ctx, 16, leaf); got != 240 {
		t.Errorf("Cqh(50) of 480px content height: expected 240, got %v", got)
	}
}
