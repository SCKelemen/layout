package layout

import "testing"

// Round-1 correctness regressions for positioned layout: the containing block
// of an absolutely positioned box is the padding box of the nearest positioned
// ancestor, not of the direct parent (CSS 2.1 §10.1 item 4,
// https://www.w3.org/TR/CSS21/visudet.html#containing-block-details).

func round1Position(root *Node) {
	ctx := NewLayoutContext(800, 600, 16)
	LayoutWithPositioning(root, Loose(800, 600), Rect{Width: 800, Height: 600}, ctx)
}

// TestPositionedRound1ContainingBlockSkipsStaticParent: root{relative} >
// mid{static, margin 50/40} > abs{left 10, top 10}. abs is positioned against
// root, so in mid's local space it sits at (10-50, 10-40) = (-40, -30).
func TestPositionedRound1ContainingBlockSkipsStaticParent(t *testing.T) {
	abs := &Node{Style: Style{Position: PositionAbsolute, Width: Px(20), Height: Px(20), Left: Px(10), Top: Px(10)}}
	mid := &Node{
		Style:    Style{Display: DisplayBlock, Height: Px(100), Margin: Spacing{Left: Px(50), Top: Px(40)}},
		Children: []*Node{abs},
	}
	root := &Node{
		Style:    Style{Display: DisplayBlock, Position: PositionRelative, Width: Px(400), Height: Px(400)},
		Children: []*Node{mid},
	}
	round1Position(root)

	if mid.Rect.X != 50 || mid.Rect.Y != 40 {
		t.Fatalf("mid should sit at (50,40), got (%.2f,%.2f)", mid.Rect.X, mid.Rect.Y)
	}
	if abs.Rect.X != -40 || abs.Rect.Y != -30 {
		t.Errorf("abs should be at root-relative (10,10) = mid-local (-40,-30), got (%.2f,%.2f)", abs.Rect.X, abs.Rect.Y)
	}

	// Right/bottom offsets are measured from root's padding box too:
	// root-relative X = 400 - 20 - 10 = 370 -> mid-local 320.
	abs.Style.Left, abs.Style.Top = Length{}, Length{}
	abs.Style.Right, abs.Style.Bottom = Px(10), Px(10)
	round1Position(root)
	if abs.Rect.X != 320 || abs.Rect.Y != 330 {
		t.Errorf("abs right:10 bottom:10 against root: expected mid-local (320,330), got (%.2f,%.2f)", abs.Rect.X, abs.Rect.Y)
	}
}

// TestPositionedRound1ContainingBlockThroughSeveralStaticLevels: the
// containing block is threaded through any number of static ancestors, with
// padding, borders, and margins accumulating along the way.
func TestPositionedRound1ContainingBlockThroughSeveralStaticLevels(t *testing.T) {
	abs := &Node{Style: Style{Position: PositionAbsolute, Width: Px(10), Height: Px(10), Left: Px(0), Top: Px(0)}}
	inner := &Node{Style: Style{Display: DisplayBlock, Height: Px(30), Margin: Spacing{Left: Px(7), Top: Px(3)}}, Children: []*Node{abs}}
	outer := &Node{Style: Style{Display: DisplayBlock, Height: Px(80), Padding: Uniform(Px(5))}, Children: []*Node{inner}}
	cb := &Node{
		Style:    Style{Display: DisplayBlock, Position: PositionRelative, Width: Px(300), Height: Px(300), Border: Uniform(Px(2))},
		Children: []*Node{outer},
	}
	root := &Node{Style: Style{Display: DisplayBlock, Width: Px(500), Height: Px(500), Padding: Uniform(Px(20))}, Children: []*Node{cb}}
	round1Position(root)

	// cb's padding box starts at (2,2) in cb-local space. outer is at (2,2)
	// in cb space; inner is at (5+7, 5+3) = (12,8) in outer space. So the
	// padding-box origin in inner-local space is (2-2-12, 2-2-8) = (-12,-8).
	if outer.Rect.X != 2 || outer.Rect.Y != 2 {
		t.Fatalf("outer should sit inside cb's border at (2,2), got (%.2f,%.2f)", outer.Rect.X, outer.Rect.Y)
	}
	if inner.Rect.X != 12 || inner.Rect.Y != 8 {
		t.Fatalf("inner should sit at (12,8) in outer, got (%.2f,%.2f)", inner.Rect.X, inner.Rect.Y)
	}
	if abs.Rect.X != -12 || abs.Rect.Y != -8 {
		t.Errorf("abs left:0 top:0 against cb: expected inner-local (-12,-8), got (%.2f,%.2f)", abs.Rect.X, abs.Rect.Y)
	}
}

// TestPositionedRound1NoPositionedAncestorUsesRoot: with no positioned
// ancestor the root is the initial containing block (today's behavior for
// direct children of the root is preserved for deeper descendants).
func TestPositionedRound1NoPositionedAncestorUsesRoot(t *testing.T) {
	abs := &Node{Style: Style{Position: PositionAbsolute, Width: Px(10), Height: Px(10), Right: Px(0), Bottom: Px(0)}}
	mid := &Node{Style: Style{Display: DisplayBlock, Height: Px(100), Margin: Spacing{Left: Px(30), Top: Px(20)}}, Children: []*Node{abs}}
	root := &Node{Style: Style{Display: DisplayBlock, Width: Px(400), Height: Px(300)}, Children: []*Node{mid}}
	round1Position(root)

	// root-relative (390,290) -> mid-local (360,270).
	if abs.Rect.X != 360 || abs.Rect.Y != 270 {
		t.Errorf("abs right:0 bottom:0 against root: expected mid-local (360,270), got (%.2f,%.2f)", abs.Rect.X, abs.Rect.Y)
	}
}

// TestPositionedRound1RelativeOffsetAppliedBeforeDescending: a relatively
// positioned static-flow ancestor is the containing block for its absolute
// descendants, and a relatively positioned non-containing ancestor's offset
// is accounted for when translating the containing block into child space.
func TestPositionedRound1RelativeOffsetAppliedBeforeDescending(t *testing.T) {
	abs := &Node{Style: Style{Position: PositionAbsolute, Width: Px(10), Height: Px(10), Left: Px(0), Top: Px(0)}}
	// rel is positioned, so it is abs's containing block; its own offset moves
	// it but abs stays at rel's padding-box origin.
	rel := &Node{Style: Style{Display: DisplayBlock, Position: PositionRelative, Left: Px(15), Top: Px(25), Height: Px(50)}, Children: []*Node{abs}}
	root := &Node{Style: Style{Display: DisplayBlock, Width: Px(400), Height: Px(300)}, Children: []*Node{rel}}
	round1Position(root)
	if rel.Rect.X != 15 || rel.Rect.Y != 25 {
		t.Fatalf("rel should be offset to (15,25), got (%.2f,%.2f)", rel.Rect.X, rel.Rect.Y)
	}
	if abs.Rect.X != 0 || abs.Rect.Y != 0 {
		t.Errorf("abs inside positioned rel: expected (0,0), got (%.2f,%.2f)", abs.Rect.X, abs.Rect.Y)
	}
}

// TestPositionedRound1FixedIgnoresPositionedAncestor: fixed boxes use the
// viewport as containing block regardless of positioned ancestors
// (CSS 2.1 §10.1 item 3).
func TestPositionedRound1FixedIgnoresPositionedAncestor(t *testing.T) {
	fixed := &Node{Style: Style{Position: PositionFixed, Width: Px(10), Height: Px(10), Right: Px(0), Bottom: Px(0)}}
	rel := &Node{Style: Style{Display: DisplayBlock, Position: PositionRelative, Width: Px(100), Height: Px(100)}, Children: []*Node{fixed}}
	root := &Node{Style: Style{Display: DisplayBlock, Width: Px(400), Height: Px(300)}, Children: []*Node{rel}}
	round1Position(root)
	if fixed.Rect.X != 790 || fixed.Rect.Y != 590 {
		t.Errorf("fixed right:0 bottom:0 against 800x600 viewport: expected (790,590), got (%.2f,%.2f)", fixed.Rect.X, fixed.Rect.Y)
	}
}

// TestPositionedRound1ExplicitSizeIsBorderBoxAware: an absolutely positioned
// box's explicit width is its used width, converted per box-sizing and
// clamped by max-width (CSS 2.1 §10.3.7 uses the specified width as-is).
func TestPositionedRound1ExplicitSizeIsBorderBoxAware(t *testing.T) {
	abs := &Node{Style: Style{
		Position: PositionAbsolute, Left: Px(0), Top: Px(0),
		Width: Px(100), Height: Px(40), Padding: Uniform(Px(10)),
	}}
	root := &Node{Style: Style{Display: DisplayBlock, Width: Px(50), Height: Px(50)}, Children: []*Node{abs}}
	round1Position(root)
	// content-box: 100 + 2*10 padding, larger than the 50px parent.
	if abs.Rect.Width != 120 || abs.Rect.Height != 60 {
		t.Errorf("content-box abs: expected 120x60, got %.2fx%.2f", abs.Rect.Width, abs.Rect.Height)
	}

	abs.Style.BoxSizing = BoxSizingBorderBox
	round1Position(root)
	if abs.Rect.Width != 100 || abs.Rect.Height != 40 {
		t.Errorf("border-box abs: expected 100x40, got %.2fx%.2f", abs.Rect.Width, abs.Rect.Height)
	}

	abs.Style.MaxWidth = Px(70)
	round1Position(root)
	if abs.Rect.Width != 70 {
		t.Errorf("max-width 70 on border-box abs: expected 70, got %.2f", abs.Rect.Width)
	}
}

// TestPositionedRound1OverConstrainedVerticalIgnoresBottom: CSS 2.1 §10.6.4,
// top + height + bottom all set: bottom is ignored.
func TestPositionedRound1OverConstrainedVerticalIgnoresBottom(t *testing.T) {
	abs := &Node{Style: Style{Position: PositionAbsolute, Top: Px(10), Bottom: Px(10), Height: Px(500), Width: Px(10)}}
	root := &Node{Style: Style{Display: DisplayBlock, Width: Px(400), Height: Px(300)}, Children: []*Node{abs}}
	round1Position(root)
	if abs.Rect.Y != 10 || abs.Rect.Height != 500 {
		t.Errorf("expected Y=10 H=500 (bottom ignored), got Y=%.2f H=%.2f", abs.Rect.Y, abs.Rect.Height)
	}

	abs.Style.Height = Length{}
	round1Position(root)
	if abs.Rect.Y != 10 || abs.Rect.Height != 280 {
		t.Errorf("auto height: expected Y=10 H=280, got Y=%.2f H=%.2f", abs.Rect.Y, abs.Rect.Height)
	}
}
