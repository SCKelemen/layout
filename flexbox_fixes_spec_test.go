package layout

import (
	"math"
	"testing"
)

// Regression tests for flexbox correctness fixes. Each test references the
// CSS Flexible Box Layout Module Level 1 section it verifies.
// https://www.w3.org/TR/css-flexbox-1/

func flexApprox(t *testing.T, what string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Errorf("%s: got %.3f, want %.3f", what, got, want)
	}
}

func flexLayout(node *Node, c Constraints) Size {
	w, h := c.MaxWidth, c.MaxHeight
	if w >= Unbounded {
		w = 1920
	}
	if h >= Unbounded {
		h = 1080
	}
	return Layout(node, c, NewLayoutContext(w, h, 16))
}

// §9.7 step 4c: shrinking distributes negative free space proportionally to
// the scaled flex shrink factor (flex-shrink * flex base size).
// https://www.w3.org/TR/css-flexbox-1/#resolve-flexible-lengths
func TestFlexboxShrinkUsesScaledShrinkFactor(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(100), Height: Px(20)}},
			{Style: Style{Width: Px(300), Height: Px(20)}},
		},
	}
	flexLayout(root, Tight(200, 100))
	flexApprox(t, "item 0 width", root.Children[0].Rect.Width, 50)
	flexApprox(t, "item 1 width", root.Children[1].Rect.Width, 150)
	flexApprox(t, "item 1 x", root.Children[1].Rect.X, 50)
}

// §9.7 step 4d/4e: an item clamped by max-width is frozen and the remaining
// free space is redistributed to the other items.
// https://www.w3.org/TR/css-flexbox-1/#resolve-flexible-lengths
func TestFlexboxGrowRespectsMaxWidth(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1, MaxWidth: Px(60)}},
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1}},
		},
	}
	flexLayout(root, Tight(300, 100))
	flexApprox(t, "item 0 width", root.Children[0].Rect.Width, 60)
	flexApprox(t, "item 1 width", root.Children[1].Rect.Width, 240)
	flexApprox(t, "item 1 x", root.Children[1].Rect.X, 60)
}

// §9.7 step 4d/4e: an item clamped by min-width is frozen and the other items
// absorb the remaining negative free space.
// https://www.w3.org/TR/css-flexbox-1/#resolve-flexible-lengths
func TestFlexboxShrinkRespectsMinWidth(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(100), Height: Px(20), MinWidth: Px(80)}},
			{Style: Style{Width: Px(100), Height: Px(20)}},
		},
	}
	flexLayout(root, Tight(100, 100))
	flexApprox(t, "item 0 width", root.Children[0].Rect.Width, 80)
	flexApprox(t, "item 1 width", root.Children[1].Rect.Width, 20)
	flexApprox(t, "item 1 x", root.Children[1].Rect.X, 80)
}

// Column direction uses min-height/max-height as the main-axis clamps.
func TestFlexboxColumnShrinkRespectsMinHeight(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex, FlexDirection: FlexDirectionColumn},
		Children: []*Node{
			{Style: Style{Width: Px(20), Height: Px(100), MinHeight: Px(80)}},
			{Style: Style{Width: Px(20), Height: Px(100)}},
		},
	}
	flexLayout(root, Tight(100, 100))
	flexApprox(t, "item 0 height", root.Children[0].Rect.Height, 80)
	flexApprox(t, "item 1 height", root.Children[1].Rect.Height, 20)
	flexApprox(t, "item 1 y", root.Children[1].Rect.Y, 80)
}

// The §9.7 loop freezes violators one round at a time; this needs two rounds.
func TestFlexboxResolveFlexibleLengthsMultipleRounds(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1, MaxWidth: Px(60)}},
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1, MaxWidth: Px(80)}},
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1}},
		},
	}
	flexLayout(root, Tight(300, 100))
	flexApprox(t, "item 0 width", root.Children[0].Rect.Width, 60)
	flexApprox(t, "item 1 width", root.Children[1].Rect.Width, 80)
	flexApprox(t, "item 2 width", root.Children[2].Rect.Width, 160)
}

// An item legitimately shrunk to zero stays at zero instead of being reset to
// its explicit width and overflowing the line. (FlexShrink's zero value means
// the default factor of 1, so the first item is pinned with min-width.)
func TestFlexboxItemShrunkToZeroStaysZero(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(100), Height: Px(20), MinWidth: Px(100)}},
			{Style: Style{Width: Px(50), Height: Px(20)}},
		},
	}
	flexLayout(root, Tight(100, 100))
	flexApprox(t, "item 0 width", root.Children[0].Rect.Width, 100)
	flexApprox(t, "item 1 width", root.Children[1].Rect.Width, 0)
	flexApprox(t, "item 1 x", root.Children[1].Rect.X, 100)
}

// §9.5 / §10.2: space-around and space-evenly insert space between items, not
// just a leading offset.
// https://www.w3.org/TR/css-flexbox-1/#justify-content-property
func TestFlexboxSpaceAroundDistributesBetweenItems(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex, JustifyContent: JustifyContentSpaceAround},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20)}},
			{Style: Style{Width: Px(50), Height: Px(20)}},
		},
	}
	flexLayout(root, Tight(300, 100))
	flexApprox(t, "item 0 x", root.Children[0].Rect.X, 50)
	flexApprox(t, "item 1 x", root.Children[1].Rect.X, 200)
}

func TestFlexboxSpaceEvenlyDistributesBetweenItems(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex, JustifyContent: JustifyContentSpaceEvenly},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20)}},
			{Style: Style{Width: Px(50), Height: Px(20)}},
		},
	}
	flexLayout(root, Tight(300, 100))
	flexApprox(t, "item 0 x", root.Children[0].Rect.X, 200.0/3)
	flexApprox(t, "item 1 x", root.Children[1].Rect.X, 200.0/3*2+50)
}

// §9.5: with negative free space space-around behaves as center and
// space-between as flex-start.
func TestFlexboxSpaceDistributionNegativeFreeSpaceFallbacks(t *testing.T) {
	mk := func(j JustifyContent) *Node {
		return &Node{
			Style: Style{Display: DisplayFlex, JustifyContent: j},
			Children: []*Node{
				{Style: Style{Width: Px(50), Height: Px(20), MinWidth: Px(50)}},
				{Style: Style{Width: Px(50), Height: Px(20), MinWidth: Px(50)}},
			},
		}
	}
	around := mk(JustifyContentSpaceAround)
	flexLayout(around, Tight(80, 100))
	flexApprox(t, "space-around item 0 x", around.Children[0].Rect.X, -10)
	flexApprox(t, "space-around item 1 x", around.Children[1].Rect.X, 40)

	between := mk(JustifyContentSpaceBetween)
	flexLayout(between, Tight(80, 100))
	flexApprox(t, "space-between item 0 x", between.Children[0].Rect.X, 0)
	flexApprox(t, "space-between item 1 x", between.Children[1].Rect.X, 50)
}

// §10.4: with an indefinite cross size wrapped lines are stacked from the
// cross-start edge, separated by the cross-axis gap.
// https://www.w3.org/TR/css-flexbox-1/#align-content-property
func TestFlexboxWrappedLinesStackWithIndefiniteCrossSize(t *testing.T) {
	mk := func(rowGap Length) *Node {
		return &Node{
			Style: Style{Display: DisplayFlex, FlexWrap: FlexWrapWrap, Width: Px(100), FlexRowGap: rowGap},
			Children: []*Node{
				{Style: Style{Width: Px(60), Height: Px(20)}},
				{Style: Style{Width: Px(60), Height: Px(20)}},
				{Style: Style{Width: Px(60), Height: Px(20)}},
			},
		}
	}
	root := mk(Length{})
	size := flexLayout(root, Unconstrained())
	flexApprox(t, "item 0 y", root.Children[0].Rect.Y, 0)
	flexApprox(t, "item 1 y", root.Children[1].Rect.Y, 20)
	flexApprox(t, "item 2 y", root.Children[2].Rect.Y, 40)
	flexApprox(t, "container height", size.Height, 60)

	withGap := mk(Px(5))
	size = flexLayout(withGap, Unconstrained())
	flexApprox(t, "gap item 1 y", withGap.Children[1].Rect.Y, 25)
	flexApprox(t, "gap item 2 y", withGap.Children[2].Rect.Y, 50)
	flexApprox(t, "gap container height", size.Height, 70)
}

// §5.2 flex-wrap: wrap-reverse flips the cross-start edge; align-content:
// flex-start then packs lines toward the bottom.
// https://www.w3.org/TR/css-flexbox-1/#flex-wrap-property
func TestFlexboxWrapReverseAlignContentFlexStart(t *testing.T) {
	root := &Node{
		Style: Style{
			Display: DisplayFlex, FlexWrap: FlexWrapWrapReverse,
			AlignContent: AlignContentFlexStart, Width: Px(100), Height: Px(200),
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}},
			{Style: Style{Width: Px(50), Height: Px(50)}},
			{Style: Style{Width: Px(50), Height: Px(50)}},
		},
	}
	flexLayout(root, Loose(100, 200))
	flexApprox(t, "item 0 y", root.Children[0].Rect.Y, 150)
	flexApprox(t, "item 1 y", root.Children[1].Rect.Y, 150)
	flexApprox(t, "item 2 y", root.Children[2].Rect.Y, 100)

	// And flex-end packs lines toward the top.
	root.Style.AlignContent = AlignContentFlexEnd
	flexLayout(root, Loose(100, 200))
	flexApprox(t, "flex-end item 0 y", root.Children[0].Rect.Y, 50)
	flexApprox(t, "flex-end item 2 y", root.Children[2].Rect.Y, 0)
}

// §9.4 step 11: stretch only applies when the item's cross size is auto.
// https://www.w3.org/TR/css-flexbox-1/#algo-stretch
func TestFlexboxStretchDoesNotOverrideExplicitCrossSize(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(30)}},
			{Style: Style{Width: Px(50)}},
		},
	}
	flexLayout(root, Tight(300, 100))
	flexApprox(t, "explicit item height", root.Children[0].Rect.Height, 30)
	flexApprox(t, "explicit item y", root.Children[0].Rect.Y, 0)
	flexApprox(t, "auto item height", root.Children[1].Rect.Height, 100)

	column := &Node{
		Style: Style{Display: DisplayFlex, FlexDirection: FlexDirectionColumn},
		Children: []*Node{
			{Style: Style{Width: Px(30), Height: Px(50)}},
			{Style: Style{Height: Px(50)}},
		},
	}
	flexLayout(column, Tight(100, 300))
	flexApprox(t, "column explicit item width", column.Children[0].Rect.Width, 30)
	flexApprox(t, "column auto item width", column.Children[1].Rect.Width, 100)
}

// §9.3 step 5: the gap counts toward the line length when breaking lines.
// https://www.w3.org/TR/css-flexbox-1/#algo-line-break
func TestFlexboxLineBreakingIncludesGap(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex, FlexWrap: FlexWrapWrap, Width: Px(100), FlexGap: Px(20)},
		Children: []*Node{
			{Style: Style{Width: Px(45), Height: Px(10)}},
			{Style: Style{Width: Px(45), Height: Px(10)}},
		},
	}
	flexLayout(root, Loose(100, Unbounded))
	flexApprox(t, "item 1 x", root.Children[1].Rect.X, 0)
	flexApprox(t, "item 1 y", root.Children[1].Rect.Y, 30)
}

// CSS Box Alignment §8.3: row-gap separates rows, column-gap separates
// columns, regardless of flex-direction. In a column flex container the gap
// between items is therefore row-gap.
// https://www.w3.org/TR/css-align-3/#column-row-gap
func TestFlexboxColumnDirectionGapAxes(t *testing.T) {
	mk := func(rowGap, columnGap Length) *Node {
		return &Node{
			Style: Style{Display: DisplayFlex, FlexDirection: FlexDirectionColumn, FlexRowGap: rowGap, FlexColumnGap: columnGap},
			Children: []*Node{
				{Style: Style{Width: Px(50), Height: Px(50)}},
				{Style: Style{Width: Px(50), Height: Px(50)}},
			},
		}
	}
	rowGap := mk(Px(10), Length{})
	flexLayout(rowGap, Tight(100, 200))
	flexApprox(t, "row-gap item 1 y", rowGap.Children[1].Rect.Y, 60)

	columnGap := mk(Length{}, Px(10))
	flexLayout(columnGap, Tight(100, 200))
	flexApprox(t, "column-gap item 1 y", columnGap.Children[1].Rect.Y, 50)

	// Wrapped column lines are separated by column-gap.
	wrapped := &Node{
		Style: Style{Display: DisplayFlex, FlexDirection: FlexDirectionColumn, FlexWrap: FlexWrapWrap, AlignContent: AlignContentFlexStart, FlexColumnGap: Px(10), Width: Px(200), Height: Px(100)},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(60)}},
			{Style: Style{Width: Px(50), Height: Px(60)}},
		},
	}
	flexLayout(wrapped, Tight(200, 100))
	flexApprox(t, "wrapped column item 1 x", wrapped.Children[1].Rect.X, 60)
	flexApprox(t, "wrapped column item 1 y", wrapped.Children[1].Rect.Y, 0)
}

// An empty flex container keeps its explicit width and height.
func TestFlexboxEmptyContainerHonorsExplicitSize(t *testing.T) {
	root := &Node{Style: Style{Display: DisplayFlex, Width: Px(200), Height: Px(100)}}
	size := flexLayout(root, Unconstrained())
	flexApprox(t, "width", size.Width, 200)
	flexApprox(t, "height", size.Height, 100)
	flexApprox(t, "rect width", root.Rect.Width, 200)
	flexApprox(t, "rect height", root.Rect.Height, 100)

	// Without any size the container collapses to padding and border only.
	empty := &Node{Style: Style{Display: DisplayFlex, Padding: Uniform(Px(4))}}
	size = flexLayout(empty, Unconstrained())
	flexApprox(t, "padding-only width", size.Width, 8)
	flexApprox(t, "padding-only height", size.Height, 8)
}

// §9.2 step 4: with an indefinite main size the container is as wide as its
// content and justify-content has no free space to distribute; unbounded
// values must never leak into positions or sizes.
// https://www.w3.org/TR/css-flexbox-1/#algo-main-container
func TestFlexboxIndefiniteMainSizeJustify(t *testing.T) {
	root := &Node{
		Style:    Style{Display: DisplayFlex, JustifyContent: JustifyContentCenter},
		Children: []*Node{{Style: Style{Width: Px(50), Height: Px(20)}}},
	}
	size := flexLayout(root, Unconstrained())
	flexApprox(t, "item x", root.Children[0].Rect.X, 0)
	flexApprox(t, "container width", size.Width, 50)

	reverse := &Node{
		Style: Style{Display: DisplayFlex, FlexDirection: FlexDirectionRowReverse},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20)}},
			{Style: Style{Width: Px(50), Height: Px(20)}},
		},
	}
	size = flexLayout(reverse, Unconstrained())
	flexApprox(t, "reverse container width", size.Width, 100)
	flexApprox(t, "reverse item 0 x", reverse.Children[0].Rect.X, 50)
	flexApprox(t, "reverse item 1 x", reverse.Children[1].Rect.X, 0)
	for i, c := range reverse.Children {
		if c.Rect.X < 0 || c.Rect.X+c.Rect.Width > 100.01 {
			t.Errorf("reverse item %d overflows: x=%.2f w=%.2f", i, c.Rect.X, c.Rect.Width)
		}
	}
}

// A nested container that shrinks must lay out its children again against its
// final size instead of keeping the geometry from the measurement pass.
func TestFlexboxShrunkNestedContainerRelayout(t *testing.T) {
	nested := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(60), Height: Px(50)}},
			{Style: Style{Width: Px(60), Height: Px(50)}},
		},
	}
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			nested,
			{Style: Style{Width: Px(100), Height: Px(50)}},
		},
	}
	flexLayout(root, Tight(100, 50))

	// Scaled shrink: bases 120 and 100 share -120 free space.
	flexApprox(t, "nested width", nested.Rect.Width, 120-120*120.0/220)
	half := nested.Rect.Width / 2
	flexApprox(t, "grandchild 0 width", nested.Children[0].Rect.Width, half)
	flexApprox(t, "grandchild 1 width", nested.Children[1].Rect.Width, half)
	flexApprox(t, "grandchild 1 x", nested.Children[1].Rect.X, half)
	if end := nested.Children[1].Rect.X + nested.Children[1].Rect.Width; end > nested.Rect.Width+0.01 {
		t.Errorf("grandchildren overflow nested container: %.2f > %.2f", end, nested.Rect.Width)
	}
}

// §9.4 step 7: a nested container measured under align-items: flex-start
// gets a content-based hypothetical cross size, not the parent's cross size.
// https://www.w3.org/TR/css-flexbox-1/#algo-cross-item
func TestFlexboxNestedContainerNotStretchedDuringMeasurement(t *testing.T) {
	nested := &Node{
		Style:    Style{Display: DisplayFlex},
		Children: []*Node{{Style: Style{Width: Px(40), Height: Px(20)}}},
	}
	root := &Node{
		Style:    Style{Display: DisplayFlex, AlignItems: AlignItemsFlexStart},
		Children: []*Node{nested},
	}
	flexLayout(root, Tight(300, 200))
	flexApprox(t, "nested height", nested.Rect.Height, 20)
	flexApprox(t, "nested width", nested.Rect.Width, 40)

	// With stretch (the default) the nested container does fill the cross axis,
	// while its child keeps its explicit height (§9.4 step 11).
	root.Style.AlignItems = AlignItemsStretch
	flexLayout(root, Tight(300, 200))
	flexApprox(t, "stretched nested height", nested.Rect.Height, 200)
	flexApprox(t, "explicit grandchild height", nested.Children[0].Rect.Height, 20)

	// An auto-height grandchild is stretched through the nested container.
	nested.Children[0].Style.Height = Length{}
	flexLayout(root, Tight(300, 200))
	flexApprox(t, "auto grandchild height", nested.Children[0].Rect.Height, 200)
}

// §10.4: align-content: center with negative free space produces negative
// offsets (lines overflow equally on both sides) instead of clamping to 0.
// https://www.w3.org/TR/css-flexbox-1/#align-content-property
func TestFlexboxAlignContentCenterNegativeFreeSpace(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex, FlexWrap: FlexWrapWrap, AlignContent: AlignContentCenter, Width: Px(100), Height: Px(30)},
		Children: []*Node{
			{Style: Style{Width: Px(60), Height: Px(20)}},
			{Style: Style{Width: Px(60), Height: Px(20)}},
			{Style: Style{Width: Px(60), Height: Px(20)}},
		},
	}
	flexLayout(root, Tight(100, 30))
	flexApprox(t, "item 0 y", root.Children[0].Rect.Y, -15)
	flexApprox(t, "item 1 y", root.Children[1].Rect.Y, 5)
	flexApprox(t, "item 2 y", root.Children[2].Rect.Y, 25)

	root.Style.AlignContent = AlignContentFlexEnd
	flexLayout(root, Tight(100, 30))
	flexApprox(t, "flex-end item 0 y", root.Children[0].Rect.Y, -30)
	flexApprox(t, "flex-end item 2 y", root.Children[2].Rect.Y, 10)
}

// §8.3: baseline alignment in a column flex container behaves as flex-start.
// https://www.w3.org/TR/css-flexbox-1/#align-items-property
func TestFlexboxBaselineInColumnBehavesAsFlexStart(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex, FlexDirection: FlexDirectionColumn, AlignItems: AlignItemsBaseline},
		Children: []*Node{
			{Style: Style{Width: Px(30), Height: Px(50), Margin: Spacing{Left: Px(4)}}, Baseline: 10},
			{Style: Style{Width: Px(40), Height: Px(50)}, Baseline: 40},
		},
	}
	flexLayout(root, Tight(100, 300))
	flexApprox(t, "item 0 x", root.Children[0].Rect.X, 4)
	flexApprox(t, "item 1 x", root.Children[1].Rect.X, 0)
}

// The §9.7 loop is bounded: pathological factors (NaN, infinite) terminate
// and never produce unbounded sizes.
func TestFlexboxResolveFlexibleLengthsTerminatesOnBadInput(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: math.Inf(1)}},
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: math.NaN()}},
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1}},
		},
	}
	flexLayout(root, Tight(300, 100))
	for i, c := range root.Children {
		if math.IsNaN(c.Rect.X) || math.IsInf(c.Rect.X, 0) || c.Rect.Width >= Unbounded {
			t.Errorf("item %d has an unbounded position/size: x=%v w=%v", i, c.Rect.X, c.Rect.Width)
		}
	}
}
