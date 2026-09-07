package layout

import (
	"math"
	"testing"
)

// The intrinsic sizing helpers set the sizing mode and leave Width/Height
// unset (auto) instead of storing the deprecated negative sentinels.
func TestIntrinsicHelpersSetSizingMode(t *testing.T) {
	cases := []struct {
		name   string
		apply  func(*Node) *Node
		width  IntrinsicSize
		height IntrinsicSize
		fitW   Length
		fitH   Length
	}{
		{"MinContentWidth", MinContentWidth, IntrinsicSizeMinContent, IntrinsicSizeNone, Length{}, Length{}},
		{"MaxContentWidth", MaxContentWidth, IntrinsicSizeMaxContent, IntrinsicSizeNone, Length{}, Length{}},
		{"FitContentWidth", func(n *Node) *Node { return FitContentWidth(n, 500) }, IntrinsicSizeFitContent, IntrinsicSizeNone, Px(500), Length{}},
		{"MinContentHeight", MinContentHeight, IntrinsicSizeNone, IntrinsicSizeMinContent, Length{}, Length{}},
		{"MaxContentHeight", MaxContentHeight, IntrinsicSizeNone, IntrinsicSizeMaxContent, Length{}, Length{}},
		{"FitContentHeight", func(n *Node) *Node { return FitContentHeight(n, 300) }, IntrinsicSizeNone, IntrinsicSizeFitContent, Length{}, Px(300)},
	}
	for _, c := range cases {
		n := c.apply(&Node{})
		if n.Style.WidthSizing != c.width || n.Style.HeightSizing != c.height {
			t.Errorf("%s: sizing = (%v, %v), want (%v, %v)", c.name, n.Style.WidthSizing, n.Style.HeightSizing, c.width, c.height)
		}
		if n.Style.Width.Unit != "" || n.Style.Height.Unit != "" {
			t.Errorf("%s: Width/Height should stay unset, got %v / %v", c.name, n.Style.Width, n.Style.Height)
		}
		if n.Style.FitContentWidth != c.fitW || n.Style.FitContentHeight != c.fitH {
			t.Errorf("%s: fit-content limits = %v / %v", c.name, n.Style.FitContentWidth, n.Style.FitContentHeight)
		}
	}

	// The mode is honored by block layout the same way the sentinel was.
	container := &Node{Style: Style{Width: Px(400)}, Children: []*Node{
		MaxContentWidth(&Node{Children: []*Node{Fixed(120, 10), Fixed(80, 10)}}),
	}}
	LayoutSimple(container, Loose(400, 400))
	if got := container.Children[0].Rect.Width; math.Abs(got-120) > 0.01 {
		t.Errorf("max-content block width = %v, want 120", got)
	}
}

func TestFrameSkipsNonPositiveAndFrameLengthDoesNot(t *testing.T) {
	n := Fixed(100, 50)
	Frame(n, 0, -1)
	if n.Style.Width != Px(100) || n.Style.Height != Px(50) {
		t.Errorf("Frame(0, -1) should leave both dimensions: %v x %v", n.Style.Width, n.Style.Height)
	}
	Frame(n, 200, 0)
	if n.Style.Width != Px(200) || n.Style.Height != Px(50) {
		t.Errorf("Frame(200, 0) should set only the width: %v x %v", n.Style.Width, n.Style.Height)
	}

	FrameLength(n, Px(0), Length{})
	if n.Style.Width != Px(0) {
		t.Errorf("FrameLength should set an explicit zero width, got %v", n.Style.Width)
	}
	if n.Style.Height != (Length{}) {
		t.Errorf("FrameLength(Length{}) should reset the height to auto, got %v", n.Style.Height)
	}
	FrameLength(n, Em(10), Vw(50))
	if n.Style.Width != Em(10) || n.Style.Height != Vw(50) {
		t.Errorf("FrameLength should store any unit: %v x %v", n.Style.Width, n.Style.Height)
	}
}

// ZStack children are absolutely positioned and take no flow space, so the
// container gets an intrinsic size from the union of its children.
func TestZStackIntrinsicSize(t *testing.T) {
	z := ZStack(Fixed(100, 50))
	if z.Style.Width != Px(100) || z.Style.Height != Px(50) {
		t.Fatalf("ZStack(Fixed(100,50)) size = %v x %v", z.Style.Width, z.Style.Height)
	}
	size := LayoutWithPositioning(z, Loose(800, 600), Rect{Width: 800, Height: 600}, NewLayoutContext(800, 600, 16))
	if size.Width != 100 || size.Height != 50 {
		t.Errorf("laid out ZStack = %v x %v, want 100 x 50", size.Width, size.Height)
	}

	// Offsets extend the union; the widest/tallest extent wins.
	a := Fixed(100, 50)
	b := Fixed(60, 80)
	b.Style.Left = Px(70)
	b.Style.Top = Px(10)
	z = ZStack(a, b)
	if z.Style.Width != Px(130) || z.Style.Height != Px(90) {
		t.Errorf("union with offsets = %v x %v, want 130 x 90", z.Style.Width, z.Style.Height)
	}

	// Children without explicit pixel sizes contribute nothing; an empty
	// union leaves the size unset (auto).
	z = ZStack(&Node{}, &Node{Style: Style{Width: Em(5), Height: Vw(50)}})
	if z.Style.Width != (Length{}) || z.Style.Height != (Length{}) {
		t.Errorf("non-pixel children should leave the size auto: %v x %v", z.Style.Width, z.Style.Height)
	}
	// A non-pixel offset on a pixel-sized child is skipped for that axis.
	c := Fixed(40, 40)
	c.Style.Left = Em(2)
	z = ZStack(c)
	if z.Style.Width != (Length{}) || z.Style.Height != Px(40) {
		t.Errorf("em offset should skip the width only: %v x %v", z.Style.Width, z.Style.Height)
	}

	// An explicit Frame after construction overrides the computed size.
	z = Frame(ZStack(Fixed(100, 50)), 300, 200)
	if z.Style.Width != Px(300) || z.Style.Height != Px(200) {
		t.Errorf("Frame should override: %v x %v", z.Style.Width, z.Style.Height)
	}
}

func TestDeprecatedSVGHelpersStillWork(t *testing.T) {
	root := &Node{Children: []*Node{{}, {Children: []*Node{{}}}}}
	var nodes []*Node
	CollectNodesForSVG(root, &nodes)
	want := root.DescendantsAndSelf()
	if len(nodes) != len(want) {
		t.Fatalf("CollectNodesForSVG collected %d nodes, want %d", len(nodes), len(want))
	}
	for i := range want {
		if nodes[i] != want[i] {
			t.Errorf("node %d differs from DescendantsAndSelf order", i)
		}
	}
	if GetSVGTransform(&Node{}) != "" {
		t.Errorf("identity transform should be empty")
	}
	n := &Node{Style: Style{Transform: Translate(3, 4)}}
	if GetSVGTransform(n) != n.Style.Transform.ToSVGString() || GetSVGTransform(n) == "" {
		t.Errorf("GetSVGTransform should match ToSVGString")
	}
	if Background(n) != n {
		t.Errorf("Background should return its argument")
	}
}

func TestWithWidthLengthAndWithHeightLength(t *testing.T) {
	orig := Fixed(100, 50)
	w := orig.WithWidthLength(Vw(50))
	h := orig.WithHeightLength(Length{})
	if w.Style.Width != Vw(50) || w.Style.Height != Px(50) {
		t.Errorf("WithWidthLength = %v x %v", w.Style.Width, w.Style.Height)
	}
	if h.Style.Height != (Length{}) || h.Style.Width != Px(100) {
		t.Errorf("WithHeightLength(Length{}) = %v x %v", h.Style.Width, h.Style.Height)
	}
	if orig.Style.Width != Px(100) || orig.Style.Height != Px(50) {
		t.Errorf("original must be unchanged: %v x %v", orig.Style.Width, orig.Style.Height)
	}
	var nilNode *Node
	if nilNode.WithWidthLength(Px(1)) != nil || nilNode.WithHeightLength(Px(1)) != nil {
		t.Errorf("nil receiver should return nil")
	}
}

func TestFoldNodes(t *testing.T) {
	root := VStack(Fixed(10, 1), HStack(Fixed(20, 1), Fixed(30, 1)))

	if got := FoldNodes(root, 0, func(acc int, _ *Node) int { return acc + 1 }); got != 5 {
		t.Errorf("count = %d, want 5", got)
	}
	if got := FoldNodes(root, 0.0, func(acc float64, n *Node) float64 { return acc + n.Style.Width.Value }); got != 60 {
		t.Errorf("width sum = %v, want 60", got)
	}
	// Pre-order: the root is visited first, then children depth-first.
	order := FoldNodes(root, []*Node(nil), func(acc []*Node, n *Node) []*Node { return append(acc, n) })
	want := root.DescendantsAndSelf()
	if len(order) != len(want) {
		t.Fatalf("visited %d nodes, want %d", len(order), len(want))
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("visit order differs at %d", i)
		}
	}
	// Matches the interface{} Fold.
	viaFold := root.Fold(0, func(acc interface{}, _ *Node) interface{} { return acc.(int) + 1 }).(int)
	if viaFold != 5 {
		t.Errorf("Fold count = %d", viaFold)
	}
	// nil node or fn returns init.
	if got := FoldNodes[int](nil, 7, func(acc int, _ *Node) int { return acc + 1 }); got != 7 {
		t.Errorf("nil node: %d", got)
	}
	if got := FoldNodes(root, "x", nil); got != "x" {
		t.Errorf("nil fn: %q", got)
	}
}
