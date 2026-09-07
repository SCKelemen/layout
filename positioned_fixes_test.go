package layout

import "testing"

// Regression tests for positioned layout fixes:
//   - unset offsets are auto; a set Px(0) is a real 0 (css-position-3 §3.1)
//   - containing block is the parent's padding box in parent-local coordinates
//     (CSS 2.1 §10.1)
//   - relative positioning honors negative offsets (CSS 2.1 §9.4.3)
//   - sticky uses resolved lengths (em, rem, ...)
//   - unbounded containing blocks never leak into positions

func fixAbsParent(children ...*Node) *Node {
	return &Node{
		Style:    Style{Display: DisplayBlock, Width: Px(200), Height: Px(200)},
		Children: children,
	}
}

func fixLayoutPositioned(root *Node) {
	ctx := NewLayoutContext(800, 600, 16)
	LayoutWithPositioning(root, Loose(800, 600), Rect{Width: 800, Height: 600}, ctx)
}

// TestPositionedFixUnsetOffsetsAreAuto: setting only Left must not pin the box
// to bottom:0; the unset Top/Bottom pair yields the static position.
func TestPositionedFixUnsetOffsetsAreAuto(t *testing.T) {
	root := fixAbsParent(&Node{Style: Style{Position: PositionAbsolute, Width: Px(10), Height: Px(10), Left: Px(5)}})
	fixLayoutPositioned(root)
	child := root.Children[0]
	if child.Rect.X != 5 || child.Rect.Y != 0 {
		t.Errorf("only Left:5 set: expected (5,0), got (%.2f,%.2f)", child.Rect.X, child.Rect.Y)
	}

	root = fixAbsParent(&Node{Style: Style{Position: PositionAbsolute, Width: Px(10), Height: Px(10), Top: Px(5)}})
	fixLayoutPositioned(root)
	child = root.Children[0]
	if child.Rect.X != 0 || child.Rect.Y != 5 {
		t.Errorf("only Top:5 set: expected (0,5), got (%.2f,%.2f)", child.Rect.X, child.Rect.Y)
	}

	// No offsets at all: static position after a 50px sibling.
	root = fixAbsParent(
		&Node{Style: Style{Display: DisplayBlock, Height: Px(50)}},
		&Node{Style: Style{Position: PositionAbsolute, Width: Px(10), Height: Px(10)}},
	)
	fixLayoutPositioned(root)
	child = root.Children[1]
	if child.Rect.X != 0 || child.Rect.Y != 50 {
		t.Errorf("no offsets: expected static position (0,50), got (%.2f,%.2f)", child.Rect.X, child.Rect.Y)
	}
}

// TestPositionedFixExplicitZeroOffsetIsReal: Px(0) is a real offset. Left:0
// and Right:0 together stretch an auto-width box across the containing block
// (CSS 2.1 §10.3.7), and Bottom:0 pins to the bottom edge.
func TestPositionedFixExplicitZeroOffsetIsReal(t *testing.T) {
	root := fixAbsParent(&Node{Style: Style{Position: PositionAbsolute, Height: Px(10), Left: Px(0), Right: Px(0)}})
	fixLayoutPositioned(root)
	child := root.Children[0]
	if child.Rect.X != 0 || child.Rect.Width != 200 {
		t.Errorf("Left:0 Right:0 auto width: expected X=0 W=200, got X=%.2f W=%.2f", child.Rect.X, child.Rect.Width)
	}

	root = fixAbsParent(&Node{Style: Style{Position: PositionAbsolute, Width: Px(10), Height: Px(10), Bottom: Px(0)}})
	fixLayoutPositioned(root)
	child = root.Children[0]
	if child.Rect.Y != 190 {
		t.Errorf("Bottom:0 with 10px height in 200px parent: expected Y=190, got %.2f", child.Rect.Y)
	}

	root = fixAbsParent(&Node{Style: Style{Position: PositionAbsolute, Width: Px(10), Height: Px(10), Top: Px(0), Left: Px(0)}})
	fixLayoutPositioned(root)
	child = root.Children[0]
	if child.Rect.X != 0 || child.Rect.Y != 0 {
		t.Errorf("Top:0 Left:0: expected (0,0), got (%.2f,%.2f)", child.Rect.X, child.Rect.Y)
	}
}

// TestPositionedFixContainingBlockIsParentPaddingBox: absolute children are
// positioned against the parent's padding box, expressed in the parent's own
// coordinate space (the same space flow layout writes child Rects in), so
// nested containers with padding and borders do not shift the result.
// CSS 2.1 §10.1: https://www.w3.org/TR/CSS21/visudet.html#containing-block-details
func TestPositionedFixContainingBlockIsParentPaddingBox(t *testing.T) {
	abs := &Node{Style: Style{Position: PositionAbsolute, Width: Px(10), Height: Px(10), Top: Px(5), Left: Px(5)}}
	// mid must itself be positioned to be abs's containing block: CSS 2.1
	// §10.1 item 4 picks the nearest ancestor with position other than static.
	// A static mid would make the (unpositioned) root the containing block.
	// https://www.w3.org/TR/CSS21/visudet.html#containing-block-details
	mid := &Node{
		Style:    Style{Display: DisplayBlock, Position: PositionRelative, Width: Px(100), Height: Px(100), Border: Uniform(Px(4))},
		Children: []*Node{abs},
	}
	root := &Node{
		Style: Style{Display: DisplayBlock, Width: Px(300), Height: Px(300), Padding: Uniform(Px(20))},
		Children: []*Node{
			{Style: Style{Display: DisplayBlock, Height: Px(50)}},
			mid,
		},
	}
	fixLayoutPositioned(root)

	if mid.Rect.X != 20 || mid.Rect.Y != 70 {
		t.Fatalf("mid should sit at parent-local (20,70), got (%.2f,%.2f)", mid.Rect.X, mid.Rect.Y)
	}
	if abs.Rect.X != 9 || abs.Rect.Y != 9 {
		t.Errorf("abs child top:5 left:5 inside 4px border: expected (9,9), got (%.2f,%.2f)", abs.Rect.X, abs.Rect.Y)
	}

	// Right/Bottom offsets are measured from the padding box edges. mid is
	// content-box 100px + 2*4px border = 108px, so its padding box is 100px
	// starting at 4: X = 4 + 100 - 10 - 5 = 89.
	if mid.Rect.Width != 108 {
		t.Fatalf("mid border box should be 108px wide, got %.2f", mid.Rect.Width)
	}
	abs.Style.Top, abs.Style.Left = Length{}, Length{}
	abs.Style.Bottom, abs.Style.Right = Px(5), Px(5)
	fixLayoutPositioned(root)
	if abs.Rect.X != 89 || abs.Rect.Y != 89 {
		t.Errorf("abs child bottom:5 right:5 inside 108px box with 4px border: expected (89,89), got (%.2f,%.2f)", abs.Rect.X, abs.Rect.Y)
	}
}

// TestPositionedFixRelativeNegativeOffsets: negative relative offsets move the
// box; they are not "auto". CSS 2.1 §9.4.3 / §9.3.2.
func TestPositionedFixRelativeNegativeOffsets(t *testing.T) {
	root := fixAbsParent(
		&Node{Style: Style{Display: DisplayBlock, Height: Px(50)}},
		&Node{Style: Style{Display: DisplayBlock, Position: PositionRelative, Width: Px(50), Height: Px(50), Left: Px(-10), Top: Px(-5)}},
	)
	fixLayoutPositioned(root)
	child := root.Children[1]
	if child.Rect.X != -10 || child.Rect.Y != 45 {
		t.Errorf("relative Left:-10 Top:-5 after 50px sibling: expected (-10,45), got (%.2f,%.2f)", child.Rect.X, child.Rect.Y)
	}

	// Right/Bottom move the box the other way when Left/Top are unset.
	root = fixAbsParent(
		&Node{Style: Style{Display: DisplayBlock, Position: PositionRelative, Width: Px(50), Height: Px(50), Right: Px(10), Bottom: Px(-5)}},
	)
	fixLayoutPositioned(root)
	child = root.Children[0]
	if child.Rect.X != -10 || child.Rect.Y != 5 {
		t.Errorf("relative Right:10 Bottom:-5: expected (-10,5), got (%.2f,%.2f)", child.Rect.X, child.Rect.Y)
	}
}

// TestPositionedFixStickyUsesResolvedLengths: sticky offsets in em resolve
// against the font size (2em = 32px at 16px), not the raw value.
func TestPositionedFixStickyUsesResolvedLengths(t *testing.T) {
	root := fixAbsParent(
		&Node{Style: Style{Display: DisplayBlock, Position: PositionSticky, Width: Px(50), Height: Px(50), Top: Em(2), Left: Em(1)}},
	)
	fixLayoutPositioned(root)
	child := root.Children[0]
	if child.Rect.X != 16 || child.Rect.Y != 32 {
		t.Errorf("sticky Top:2em Left:1em: expected (16,32), got (%.2f,%.2f)", child.Rect.X, child.Rect.Y)
	}
}

// TestPositionedFixUnboundedContainingBlockNeverLeaks: an indefinite
// containing block dimension must not produce an unbounded coordinate.
func TestPositionedFixUnboundedContainingBlockNeverLeaks(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)
	node := &Node{
		Style: Style{Position: PositionAbsolute, Width: Px(10), Height: Px(10), Right: Px(5), Bottom: Px(5)},
		Rect:  Rect{Width: 10, Height: 10},
	}
	LayoutPositioned(node, Rect{Width: Unbounded, Height: Unbounded}, Rect{Width: 800, Height: 600}, ctx)
	if node.Rect.X >= Unbounded || node.Rect.Y >= Unbounded || node.Rect.X < 0 || node.Rect.Y < 0 {
		t.Fatalf("unbounded containing block leaked into position: (%v,%v)", node.Rect.X, node.Rect.Y)
	}

	node = &Node{
		Style: Style{Position: PositionAbsolute, Height: Px(10), Left: Px(0), Right: Px(0)},
		Rect:  Rect{Width: 50, Height: 10},
	}
	LayoutPositioned(node, Rect{Width: Unbounded, Height: 100}, Rect{Width: 800, Height: 600}, ctx)
	if node.Rect.Width >= Unbounded {
		t.Fatalf("unbounded containing block leaked into width: %v", node.Rect.Width)
	}
}
