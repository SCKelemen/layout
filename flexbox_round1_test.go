package layout

import (
	"math"
	"testing"
)

// Round-1 correctness regression tests for flexbox. Each test references the
// CSS Flexible Box Layout Module Level 1 (or CSS Box Alignment Level 3)
// section it verifies.
// https://www.w3.org/TR/css-flexbox-1/
// https://www.w3.org/TR/css-align-3/

func round1Ctx() *LayoutContext {
	return NewLayoutContext(1920, 1080, 16)
}

func round1Finite(t *testing.T, what string, v float64) {
	t.Helper()
	if math.IsNaN(v) || math.IsInf(v, 0) {
		t.Errorf("%s: got %v, want a finite value", what, v)
	}
}

// §7.2.1: flex-grow only accepts non-negative numbers; an infinite factor is
// invalid and must not poison the §9.7 arithmetic with NaN. It is treated as
// unset (0), so the sibling with flex-grow: 1 takes all the free space.
// https://www.w3.org/TR/css-flexbox-1/#flex-grow-property
func TestFlexboxInfiniteGrowFactorIsIgnored(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: math.Inf(1)}},
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1}},
		},
	}
	Layout(root, Tight(400, 100), round1Ctx())
	for _, c := range root.Children {
		round1Finite(t, "item width", c.Rect.Width)
		round1Finite(t, "item x", c.Rect.X)
	}
	flexApprox(t, "item 0 width", root.Children[0].Rect.Width, 50)
	flexApprox(t, "item 1 width", root.Children[1].Rect.Width, 350)
	flexApprox(t, "item 1 x", root.Children[1].Rect.X, 50)
}

// §7.2.1 / §7.2.2: NaN and negative factors are invalid. A negative or NaN
// flex-grow is ignored (0); a negative flex-shrink falls back to the initial
// value 1, so both items still shrink equally.
// https://www.w3.org/TR/css-flexbox-1/#flex-shrink-property
func TestFlexboxInvalidFactorsFallBackToInitialValues(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: math.NaN()}},
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: -3}},
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1}},
		},
	}
	Layout(root, Tight(400, 100), round1Ctx())
	flexApprox(t, "NaN grow item width", root.Children[0].Rect.Width, 50)
	flexApprox(t, "negative grow item width", root.Children[1].Rect.Width, 50)
	flexApprox(t, "grow 1 item width", root.Children[2].Rect.Width, 300)
	flexApprox(t, "grow 1 item x", root.Children[2].Rect.X, 100)

	shrink := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(100), Height: Px(20), FlexShrink: -1}},
			{Style: Style{Width: Px(100), Height: Px(20), FlexShrink: math.Inf(1)}},
		},
	}
	Layout(shrink, Tight(100, 100), round1Ctx())
	flexApprox(t, "negative shrink item width", shrink.Children[0].Rect.Width, 50)
	flexApprox(t, "infinite shrink item width", shrink.Children[1].Rect.Width, 50)
	flexApprox(t, "infinite shrink item x", shrink.Children[1].Rect.X, 50)
}

// §9.7 step 4: huge but finite factors are valid; their sum must not overflow
// to +Inf (which turned every share into NaN) and free space times a factor
// must not overflow either. Two items with flex-grow: 1e308 share equally, and
// 1e308 against 1 takes essentially all of the free space.
// https://www.w3.org/TR/css-flexbox-1/#resolve-flexible-lengths
func TestFlexboxHugeGrowFactorsDoNotOverflow(t *testing.T) {
	equal := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1e308}},
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1e308}},
		},
	}
	Layout(equal, Tight(400, 100), round1Ctx())
	flexApprox(t, "equal: item 0 width", equal.Children[0].Rect.Width, 200)
	flexApprox(t, "equal: item 1 width", equal.Children[1].Rect.Width, 200)
	flexApprox(t, "equal: item 1 x", equal.Children[1].Rect.X, 200)

	single := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1e308}},
		},
	}
	Layout(single, Tight(400, 100), round1Ctx())
	flexApprox(t, "single: item width", single.Children[0].Rect.Width, 400)

	lopsided := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1e308}},
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1}},
		},
	}
	Layout(lopsided, Tight(400, 100), round1Ctx())
	flexApprox(t, "lopsided: item 0 width", lopsided.Children[0].Rect.Width, 350)
	flexApprox(t, "lopsided: item 1 width", lopsided.Children[1].Rect.Width, 50)
	flexApprox(t, "lopsided: item 1 x", lopsided.Children[1].Rect.X, 350)

	shrink := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(100), Height: Px(20), FlexShrink: 1e308}},
			{Style: Style{Width: Px(100), Height: Px(20), FlexShrink: 1e308}},
		},
	}
	Layout(shrink, Tight(100, 100), round1Ctx())
	flexApprox(t, "shrink: item 0 width", shrink.Children[0].Rect.Width, 50)
	flexApprox(t, "shrink: item 1 width", shrink.Children[1].Rect.Width, 50)
	flexApprox(t, "shrink: item 1 x", shrink.Children[1].Rect.X, 50)
}

// §7.2.3 / §9.2 step 3: flex-basis: 0 is a real zero flex base size, so two
// items with flex: 1 0 0 share the container equally regardless of content.
// Previously a zero basis fell back to the content size (300/100 here).
// https://www.w3.org/TR/css-flexbox-1/#flex-basis-property
func TestFlexboxFlexBasisZeroIsRealZero(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{
				Style:    Style{FlexGrow: 1, FlexBasis: Px(0), Height: Px(20)},
				Children: []*Node{{Style: Style{Width: Px(200), Height: Px(20)}}},
			},
			{Style: Style{FlexGrow: 1, FlexBasis: Px(0), Height: Px(20)}},
		},
	}
	Layout(root, Tight(400, 100), round1Ctx())
	flexApprox(t, "item 0 width", root.Children[0].Rect.Width, 200)
	flexApprox(t, "item 1 width", root.Children[1].Rect.Width, 200)
	flexApprox(t, "item 1 x", root.Children[1].Rect.X, 200)

	// The same in column direction.
	column := &Node{
		Style: Style{Display: DisplayFlex, FlexDirection: FlexDirectionColumn},
		Children: []*Node{
			{
				Style:    Style{FlexGrow: 1, FlexBasis: Px(0), Width: Px(20)},
				Children: []*Node{{Style: Style{Width: Px(20), Height: Px(200)}}},
			},
			{Style: Style{FlexGrow: 1, FlexBasis: Px(0), Width: Px(20)}},
		},
	}
	Layout(column, Tight(100, 400), round1Ctx())
	flexApprox(t, "column: item 0 height", column.Children[0].Rect.Height, 200)
	flexApprox(t, "column: item 1 height", column.Children[1].Rect.Height, 200)
	flexApprox(t, "column: item 1 y", column.Children[1].Rect.Y, 200)
}

// §9.2 step 3E: an unset flex-basis is auto and resolves to the item's main
// size property when it is definite, so items grow from their widths.
// https://www.w3.org/TR/css-flexbox-1/#algo-main-item
func TestFlexboxUnsetFlexBasisIsAuto(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(100), Height: Px(20), FlexGrow: 1}},
			{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1}},
		},
	}
	Layout(root, Tight(400, 100), round1Ctx())
	flexApprox(t, "item 0 width", root.Children[0].Rect.Width, 225)
	flexApprox(t, "item 1 width", root.Children[1].Rect.Width, 175)
}

// A layout.Text() item leaves Width/Height unset, so its flex base size is its
// max-content size (§9.2 step 3E), not 0.
// https://www.w3.org/TR/css-flexbox-1/#algo-main-item
func TestFlexboxTextItemStaysAuto(t *testing.T) {
	standalone := Text("hello world")
	Layout(standalone, Unconstrained(), round1Ctx())
	if standalone.Rect.Width <= 0 {
		t.Fatalf("standalone text width = %v, want > 0", standalone.Rect.Width)
	}

	root := &Node{
		Style:    Style{Display: DisplayFlex},
		Children: []*Node{Text("hello world"), {Style: Style{Width: Px(30), Height: Px(30)}}},
	}
	Layout(root, Tight(400, 100), round1Ctx())
	flexApprox(t, "text item width", root.Children[0].Rect.Width, standalone.Rect.Width)
	flexApprox(t, "box item x", root.Children[1].Rect.X, standalone.Rect.Width)
}

// An item with Width: Px(0) has a definite zero main size (CSS 2.1 §10.2 has
// no zero-means-auto rule), so it takes no space even when it has content.
// https://www.w3.org/TR/css-flexbox-1/#algo-main-item
func TestFlexboxItemWidthZeroIsExplicit(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{
				Style:    Style{Width: Px(0), Height: Px(20)},
				Children: []*Node{{Style: Style{Width: Px(200), Height: Px(20)}}},
			},
			{Style: Style{Width: Px(50), Height: Px(20)}},
		},
	}
	Layout(root, Tight(400, 100), round1Ctx())
	flexApprox(t, "item 0 width", root.Children[0].Rect.Width, 0)
	flexApprox(t, "item 1 x", root.Children[1].Rect.X, 0)
	flexApprox(t, "item 1 width", root.Children[1].Rect.Width, 50)
}

// A flex container with Width: Px(0) has a zero content width (its main size
// in row direction) instead of being treated as auto and sized to its items.
// https://www.w3.org/TR/css-flexbox-1/#algo-main-container
func TestFlexboxContainerWidthZeroIsExplicit(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex, Width: Px(0), Height: Px(50)},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20)}},
		},
	}
	size := Layout(root, Loose(400, 100), round1Ctx())
	flexApprox(t, "container width", size.Width, 0)
	flexApprox(t, "container rect width", root.Rect.Width, 0)
	flexApprox(t, "container height", size.Height, 50)
	// The only item shrinks into the zero main size (flex-shrink: 1, min 0).
	flexApprox(t, "item width", root.Children[0].Rect.Width, 0)

	// Height: Px(0) on an empty container is a real zero too.
	empty := &Node{Style: Style{Display: DisplayFlex, Width: Px(100), Height: Px(0)}}
	emptySize := Layout(empty, Loose(400, 100), round1Ctx())
	flexApprox(t, "empty container height", emptySize.Height, 0)
	flexApprox(t, "empty container width", emptySize.Width, 100)
}

// §9.4 step 11: align-self: stretch applies only to items whose cross size is
// auto. Height: Px(0) is an explicit zero cross size and is kept, while an
// item with an unset height is stretched to the line.
// https://www.w3.org/TR/css-flexbox-1/#algo-stretch
func TestFlexboxStretchKeepsExplicitZeroCrossSize(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(0)}},
			{Style: Style{Width: Px(50), Height: Px(40)}},
			{Style: Style{Width: Px(50)}},
		},
	}
	Layout(root, Tight(400, 100), round1Ctx())
	flexApprox(t, "zero-height item height", root.Children[0].Rect.Height, 0)
	flexApprox(t, "explicit-height item height", root.Children[1].Rect.Height, 40)
	flexApprox(t, "auto-height item height", root.Children[2].Rect.Height, 100)

	column := &Node{
		Style: Style{Display: DisplayFlex, FlexDirection: FlexDirectionColumn},
		Children: []*Node{
			{Style: Style{Height: Px(20), Width: Px(0)}},
			{Style: Style{Height: Px(20)}},
		},
	}
	Layout(column, Tight(300, 200), round1Ctx())
	flexApprox(t, "column: zero-width item width", column.Children[0].Rect.Width, 0)
	flexApprox(t, "column: auto-width item width", column.Children[1].Rect.Width, 300)
}

// css-align-3 §8.3: `gap` is a shorthand that sets row-gap and column-gap. An
// explicit longhand, including Px(0), overrides it; only an unset longhand
// falls back to FlexGap. Previously the fallback keyed on the resolved value
// being 0, so `FlexColumnGap: Px(0)` could not cancel `FlexGap: Px(10)`.
// https://www.w3.org/TR/css-align-3/#gap-shorthand
func TestFlexboxExplicitZeroGapOverridesShorthand(t *testing.T) {
	row := &Node{
		Style: Style{Display: DisplayFlex, FlexGap: Px(10), FlexColumnGap: Px(0)},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20)}},
			{Style: Style{Width: Px(50), Height: Px(20)}},
		},
	}
	Layout(row, Tight(400, 100), round1Ctx())
	flexApprox(t, "row: item 1 x with column-gap 0", row.Children[1].Rect.X, 50)

	// Unset longhand still falls back to the shorthand.
	fallback := &Node{
		Style: Style{Display: DisplayFlex, FlexGap: Px(10)},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20)}},
			{Style: Style{Width: Px(50), Height: Px(20)}},
		},
	}
	Layout(fallback, Tight(400, 100), round1Ctx())
	flexApprox(t, "fallback: item 1 x with gap 10", fallback.Children[1].Rect.X, 60)

	// Row gap between wrapped lines: FlexRowGap Px(0) cancels FlexGap.
	// align-content: flex-start keeps the lines packed so only the gap matters.
	wrapped := &Node{
		Style: Style{Display: DisplayFlex, FlexWrap: FlexWrapWrap, AlignContent: AlignContentFlexStart, FlexGap: Px(10), FlexRowGap: Px(0)},
		Children: []*Node{
			{Style: Style{Width: Px(100), Height: Px(20)}},
			{Style: Style{Width: Px(100), Height: Px(20)}},
		},
	}
	Layout(wrapped, Loose(100, 500), round1Ctx())
	flexApprox(t, "wrapped: line 2 y with row-gap 0", wrapped.Children[1].Rect.Y, 20)

	// Column direction: FlexRowGap is the gap between items; Px(0) cancels FlexGap.
	column := &Node{
		Style: Style{Display: DisplayFlex, FlexDirection: FlexDirectionColumn, FlexGap: Px(10), FlexRowGap: Px(0)},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20)}},
			{Style: Style{Width: Px(50), Height: Px(20)}},
		},
	}
	Layout(column, Tight(100, 400), round1Ctx())
	flexApprox(t, "column: item 1 y with row-gap 0", column.Children[1].Rect.Y, 20)
}

// css-align-3 §6.2: align-content: space-evenly distributes the free cross
// space so that the gaps between lines and at both edges are equal. With
// negative free space the fallback is center.
// https://www.w3.org/TR/css-align-3/#valdef-align-content-space-evenly
func TestFlexboxAlignContentSpaceEvenly(t *testing.T) {
	mk := func() *Node {
		return &Node{
			Style: Style{Display: DisplayFlex, FlexWrap: FlexWrapWrap, AlignContent: AlignContentSpaceEvenly},
			Children: []*Node{
				{Style: Style{Width: Px(100), Height: Px(20)}},
				{Style: Style{Width: Px(100), Height: Px(20)}},
				{Style: Style{Width: Px(100), Height: Px(20)}},
			},
		}
	}
	// Three 20px lines in 200px: free 140, split into four 35px gaps.
	root := mk()
	Layout(root, Tight(100, 200), round1Ctx())
	flexApprox(t, "line 0 y", root.Children[0].Rect.Y, 35)
	flexApprox(t, "line 1 y", root.Children[1].Rect.Y, 90)
	flexApprox(t, "line 2 y", root.Children[2].Rect.Y, 145)
	flexApprox(t, "line height kept", root.Children[0].Rect.Height, 20)

	// Cross-axis gap is added to the evenly distributed space.
	gapped := mk()
	gapped.Style.FlexRowGap = Px(10)
	Layout(gapped, Tight(100, 200), round1Ctx())
	// free = 200 - (60 + 20) = 120, four gaps of 30.
	flexApprox(t, "gapped: line 0 y", gapped.Children[0].Rect.Y, 30)
	flexApprox(t, "gapped: line 1 y", gapped.Children[1].Rect.Y, 90)
	flexApprox(t, "gapped: line 2 y", gapped.Children[2].Rect.Y, 150)

	// Negative free space (60px of lines in 40px): center fallback, lines
	// overflow both edges equally.
	overflow := mk()
	Layout(overflow, Tight(100, 40), round1Ctx())
	flexApprox(t, "overflow: line 0 y", overflow.Children[0].Rect.Y, -10)
	flexApprox(t, "overflow: line 1 y", overflow.Children[1].Rect.Y, 10)
	flexApprox(t, "overflow: line 2 y", overflow.Children[2].Rect.Y, 30)

	// wrap-reverse mirrors the distribution across the cross axis.
	reversed := mk()
	reversed.Style.FlexWrap = FlexWrapWrapReverse
	Layout(reversed, Tight(100, 200), round1Ctx())
	flexApprox(t, "wrap-reverse: line 0 y", reversed.Children[0].Rect.Y, 145)
	flexApprox(t, "wrap-reverse: line 1 y", reversed.Children[1].Rect.Y, 90)
	flexApprox(t, "wrap-reverse: line 2 y", reversed.Children[2].Rect.Y, 35)
}

// §8.2: justify-content: stretch behaves as flex-start in a flex container
// because the flexible-length algorithm has already consumed the free space.
// The Start/End aliases are the same values as flex-start/flex-end.
// https://www.w3.org/TR/css-flexbox-1/#justify-content-property
func TestFlexboxJustifyContentStretchBehavesAsFlexStart(t *testing.T) {
	root := &Node{
		Style: Style{Display: DisplayFlex, JustifyContent: JustifyContentStretch},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20)}},
			{Style: Style{Width: Px(50), Height: Px(20)}},
		},
	}
	Layout(root, Tight(400, 100), round1Ctx())
	flexApprox(t, "item 0 x", root.Children[0].Rect.X, 0)
	flexApprox(t, "item 1 x", root.Children[1].Rect.X, 50)
	flexApprox(t, "item 0 width unchanged", root.Children[0].Rect.Width, 50)

	// row-reverse: main-start is the right edge, so stretch packs items there
	// exactly like flex-start does.
	reversed := &Node{
		Style: Style{Display: DisplayFlex, FlexDirection: FlexDirectionRowReverse, JustifyContent: JustifyContentStretch},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20)}},
			{Style: Style{Width: Px(50), Height: Px(20)}},
		},
	}
	Layout(reversed, Tight(400, 100), round1Ctx())
	flexApprox(t, "row-reverse: item 0 x", reversed.Children[0].Rect.X, 350)
	flexApprox(t, "row-reverse: item 1 x", reversed.Children[1].Rect.X, 300)

	if JustifyContentStart != JustifyContentFlexStart || JustifyContentEnd != JustifyContentFlexEnd {
		t.Error("JustifyContentStart/End must alias flex-start/flex-end")
	}
	if AlignContentStart != AlignContentFlexStart || AlignContentEnd != AlignContentFlexEnd {
		t.Error("AlignContentStart/End must alias flex-start/flex-end")
	}
}

// §9.2 step 4: a flex container with a definite main size is that size, in
// every writing mode, even when the constraints are loose or unbounded and the
// content is smaller. Previously the container was content-sized on the main
// axis unless a tight constraint enforced the explicit size, so a vertical-rl
// row container (vertical main axis) with Height: Px(100) ended up 30px tall
// while its block and grid siblings honored the 100px.
// https://www.w3.org/TR/css-flexbox-1/#algo-main-container
func TestFlexboxContainerDefiniteMainSizeUnderLooseConstraints(t *testing.T) {
	cases := []struct {
		name string
		dir  FlexDirection
		wm   WritingMode
	}{
		{"horizontal-tb row", FlexDirectionRow, WritingModeHorizontalTB},
		{"horizontal-tb column", FlexDirectionColumn, WritingModeHorizontalTB},
		{"vertical-rl row", FlexDirectionRow, WritingModeVerticalRL},
		{"vertical-rl column", FlexDirectionColumn, WritingModeVerticalRL},
		{"vertical-lr row", FlexDirectionRow, WritingModeVerticalLR},
	}
	for _, tc := range cases {
		for _, c := range []Constraints{Unconstrained(), Loose(500, 500)} {
			root := &Node{
				Style:    Style{Display: DisplayFlex, FlexDirection: tc.dir, WritingMode: tc.wm, Width: Px(200), Height: Px(100)},
				Children: []*Node{{Style: Style{Width: Px(30), Height: Px(30)}}},
			}
			size := Layout(root, c, round1Ctx())
			flexApprox(t, tc.name+": container width", size.Width, 200)
			flexApprox(t, tc.name+": container height", size.Height, 100)
			flexApprox(t, tc.name+": rect width", root.Rect.Width, 200)
			flexApprox(t, tc.name+": rect height", root.Rect.Height, 100)
			flexApprox(t, tc.name+": item width kept", root.Children[0].Rect.Width, 30)
			flexApprox(t, tc.name+": item height kept", root.Children[0].Rect.Height, 30)
		}
	}

	// A definite main size does not grow to fit overflowing content either.
	// (MinWidth pins the items: a zero FlexShrink is the unset value, 1.)
	overflow := &Node{
		Style: Style{Display: DisplayFlex, Width: Px(100), Height: Px(50)},
		Children: []*Node{
			{Style: Style{Width: Px(150), MinWidth: Px(150), Height: Px(20)}},
			{Style: Style{Width: Px(150), MinWidth: Px(150), Height: Px(20)}},
		},
	}
	size := Layout(overflow, Unconstrained(), round1Ctx())
	flexApprox(t, "overflow: container width", size.Width, 100)
	flexApprox(t, "overflow: item 1 x", overflow.Children[1].Rect.X, 150)

	// Without a set main size the container is still content-sized.
	auto := &Node{
		Style:    Style{Display: DisplayFlex, Height: Px(100)},
		Children: []*Node{{Style: Style{Width: Px(30), Height: Px(30)}}},
	}
	size = Layout(auto, Unconstrained(), round1Ctx())
	flexApprox(t, "auto: container width", size.Width, 30)
	flexApprox(t, "auto: container height", size.Height, 100)
}

// CSS Writing Modes Level 3 §7.1 / CSS Flexbox §2: in vertical-rl a row flex
// container's main axis is the inline axis (physical vertical, top to bottom)
// and a column container's main axis is the block axis (physical horizontal,
// right to left). Positions and sizes must both follow that mapping, and
// align-items: stretch stretches the block-axis (horizontal) size of auto items
// in a row container.
// https://www.w3.org/TR/css-writing-modes-3/#logical-to-physical
// https://www.w3.org/TR/css-flexbox-1/#flex-direction-property
func TestFlexboxVerticalRLAxisMapping(t *testing.T) {
	row := &Node{
		Style: Style{Display: DisplayFlex, FlexDirection: FlexDirectionRow, WritingMode: WritingModeVerticalRL, Width: Px(200), Height: Px(300)},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20)}},
			{Style: Style{Width: Px(50), Height: Px(20)}},
			{Style: Style{Height: Px(20)}}, // auto block size: stretched
		},
	}
	Layout(row, Unconstrained(), round1Ctx())
	flexApprox(t, "row: container width", row.Rect.Width, 200)
	flexApprox(t, "row: container height", row.Rect.Height, 300)
	flexApprox(t, "row: item 0 y", row.Children[0].Rect.Y, 0)
	flexApprox(t, "row: item 1 y", row.Children[1].Rect.Y, 20)
	flexApprox(t, "row: item 2 y", row.Children[2].Rect.Y, 40)
	for i, c := range row.Children {
		flexApprox(t, "row: item x", c.Rect.X, 0)
		flexApprox(t, "row: item height (main size)", c.Rect.Height, 20)
		if i < 2 {
			flexApprox(t, "row: explicit width kept", c.Rect.Width, 50)
		}
	}
	flexApprox(t, "row: auto item stretched across block axis", row.Children[2].Rect.Width, 200)

	column := &Node{
		Style: Style{Display: DisplayFlex, FlexDirection: FlexDirectionColumn, WritingMode: WritingModeVerticalRL, Width: Px(200), Height: Px(300)},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(20)}},
			{Style: Style{Width: Px(50), Height: Px(20)}},
			{Style: Style{Width: Px(50)}}, // auto inline size: stretched
		},
	}
	Layout(column, Unconstrained(), round1Ctx())
	flexApprox(t, "column: container width", column.Rect.Width, 200)
	flexApprox(t, "column: container height", column.Rect.Height, 300)
	flexApprox(t, "column: item 0 x (right edge)", column.Children[0].Rect.X, 150)
	flexApprox(t, "column: item 1 x", column.Children[1].Rect.X, 100)
	flexApprox(t, "column: item 2 x", column.Children[2].Rect.X, 50)
	for _, c := range column.Children {
		flexApprox(t, "column: item y", c.Rect.Y, 0)
		flexApprox(t, "column: item width (main size)", c.Rect.Width, 50)
	}
	flexApprox(t, "column: explicit height kept", column.Children[0].Rect.Height, 20)
	flexApprox(t, "column: auto item stretched across inline axis", column.Children[2].Rect.Height, 300)
}
