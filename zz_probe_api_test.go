package layout

import (
	"math"
	"testing"
)

func TestProbeZeroTransform(t *testing.T) {
	n := &Node{Rect: Rect{X: 10, Y: 20, Width: 100, Height: 50}}
	t.Logf("GetSVGTransform(default node) = %q", GetSVGTransform(n))
	t.Logf("GetFinalRect(default node) = %+v", GetFinalRect(n))
	t.Logf("zero Transform IsIdentity = %v", Transform{}.IsIdentity())
}

func TestProbeMultiplyOrder(t *testing.T) {
	// Doc: "applies t2 after t1" => translate(10,0) then scale(2) should map (0,0) -> (20,0)
	c := Translate(10, 0).Multiply(Scale(2, 2))
	p := c.Apply(Point{0, 0})
	t.Logf("Translate(10,0).Multiply(Scale(2,2)).Apply(0,0) = %+v (doc says t2 after t1 => expect X=20)", p)
}

func TestProbeCloneDeepShares(t *testing.T) {
	orig := &Node{Style: Style{
		GridTemplateRows:  []GridTrack{FixedTrack(Px(10))},
		GridTemplateAreas: NewGridTemplateAreas(1, 1),
		TextStyle:         &TextStyle{FontSize: 12},
	}}
	cp := orig.CloneDeep()
	cp.Style.GridTemplateRows[0] = FixedTrack(Px(999))
	cp.Style.TextStyle.FontSize = 99
	cp.Style.GridTemplateAreas.Rows = 42
	t.Logf("after mutating CloneDeep copy: orig rows[0]=%v textFont=%v areasRows=%v",
		orig.Style.GridTemplateRows[0].MinSize.Value, orig.Style.TextStyle.FontSize, orig.Style.GridTemplateAreas.Rows)
}

func TestProbeTransformDiscardsChildEdits(t *testing.T) {
	root := HStack(Fixed(1, 1), Fixed(2, 2), Fixed(3, 3))
	out := root.Transform(
		func(n *Node) bool { return n.Style.Display == DisplayFlex },
		func(n *Node) *Node { return n.RemoveChildAt(0) },
	)
	t.Logf("Transform(RemoveChildAt(0)) child count = %d (expected 2)", len(out.Children))

	out2 := root.Map(func(n *Node) *Node {
		if n.Style.Display == DisplayFlex {
			return n.AddChild(Fixed(4, 4))
		}
		return n
	})
	t.Logf("Map(AddChild) child count = %d (expected 4)", len(out2.Children))

	// nil-returning transform
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Transform with nil-returning fn panicked: %v", r)
		}
	}()
	root.Transform(func(n *Node) bool { return true }, func(n *Node) *Node { return nil })
}

func TestProbeRepeatTracksAutoFill(t *testing.T) {
	t.Logf("RepeatTracks(RepeatCountAutoFill, FixedTrack) len = %d", len(RepeatTracks(RepeatCountAutoFill, FixedTrack(Px(100)))))
}

func TestProbeZeroWidthSemantics(t *testing.T) {
	// Block child with zero-value Width inside a 300px wide parent
	parent := &Node{Style: Style{Display: DisplayBlock, Width: Px(300)}, Children: []*Node{{Style: Style{Height: Px(20)}}}}
	Layout(parent, Loose(300, 300), NewLayoutContext(800, 600, 16))
	t.Logf("block child with zero Width => Rect.Width=%v", parent.Children[0].Rect.Width)

	// README: Fixed(0, 32) inside VStack
	v := VStack(Fixed(0, 32))
	v.Style.Width = Px(200)
	Layout(v, Loose(200, 200), NewLayoutContext(800, 600, 16))
	t.Logf("VStack(Fixed(0,32)) child Rect=%+v", v.Children[0].Rect)
}

func TestProbeSnapHalf(t *testing.T) {
	nodes := []*Node{{Rect: Rect{X: 5, Y: -5}}, {Rect: Rect{X: 15, Y: -15}}, {Rect: Rect{X: -2.5, Y: 2.5}}}
	SnapNodes(nodes, 10)
	for _, n := range nodes {
		t.Logf("snapped %+v", n.Rect)
	}
	_ = math.Pi
}

func TestProbeResolveLength(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)
	for _, l := range []Length{Px(10), Em(2), Rem(2), Vw(10), Vh(10), Cqw(10), {Value: 5, Unit: ""}, UnboundedLength(), Px(-1)} {
		t.Logf("%v%s -> %v", l.Value, l.Unit, ResolveLength(l, ctx, 20))
	}
	t.Logf("nil ctx Em(2) -> %v", ResolveLength(Em(2), nil, 20))
}
