package layout

import (
	"reflect"
	"testing"
)

// TestParseContainerType covers all three valid keywords plus invalid input.
func TestParseContainerType(t *testing.T) {
	tests := []struct {
		in      string
		want    ContainerType
		wantErr bool
	}{
		{"normal", ContainerTypeNormal, false},
		{"size", ContainerTypeSize, false},
		{"inline-size", ContainerTypeInlineSize, false},
		{"SIZE", ContainerTypeSize, false},
		{"  Inline-Size  ", ContainerTypeInlineSize, false},
		{"", ContainerTypeNormal, false},
		{"bogus", ContainerTypeNormal, true},
		{"block-size", ContainerTypeNormal, true},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseContainerType(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr=%v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseContainerName(t *testing.T) {
	tests := []struct {
		in      string
		want    []string
		wantErr bool
	}{
		{"card", []string{"card"}, false},
		{"card sidebar", []string{"card", "sidebar"}, false},
		{"  card   sidebar  ", []string{"card", "sidebar"}, false},
		{"none", nil, false}, // CSS reset keyword
		{"", nil, false},
		// Reserved names rejected.
		{"normal", nil, true},
		{"inherit", nil, true},
		{"initial", nil, true},
		{"unset", nil, true},
		{"card none", nil, true},
		// Invalid identifiers.
		{"1card", nil, true},
		{"bad name!", nil, true},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseContainerName(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr=%v", err, tc.wantErr)
			}
			if !tc.wantErr && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseContainerShorthand(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		wantNames []string
		wantType  ContainerType
		wantErr   bool
	}{
		{"name only", "foo", []string{"foo"}, ContainerTypeNormal, false},
		{"multi name", "foo bar", []string{"foo", "bar"}, ContainerTypeNormal, false},
		{"name and type", "foo / size", []string{"foo"}, ContainerTypeSize, false},
		{"multi name and type", "card sidebar / inline-size", []string{"card", "sidebar"}, ContainerTypeInlineSize, false},
		{"bare type size", "size", nil, ContainerTypeSize, false},
		{"bare type inline-size", "inline-size", nil, ContainerTypeInlineSize, false},
		{"bare type normal", "normal", nil, ContainerTypeNormal, false},
		{"empty", "", nil, ContainerTypeNormal, false},
		{"name with slash and type case", "Foo / SIZE", []string{"Foo"}, ContainerTypeSize, false},
		{"invalid type", "foo / bogus", nil, ContainerTypeNormal, true},
		// `none` in the shorthand is a valid container-name keyword meaning
		// "no name"; it must NOT be rejected even though direct use as a
		// custom name identifier is reserved.
		{"none clears name", "none / size", nil, ContainerTypeSize, false},
		{"reserved name as name", "inherit / size", nil, ContainerTypeNormal, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			names, ctype, err := ParseContainer(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr=%v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if !reflect.DeepEqual(names, tc.wantNames) {
				t.Errorf("names: got %v, want %v", names, tc.wantNames)
			}
			if ctype != tc.wantType {
				t.Errorf("type: got %v, want %v", ctype, tc.wantType)
			}
		})
	}
}

// TestParseLengthCQUnits covers parsing of the six container query units
// from string, including case-insensitivity and round-tripping.
func TestParseLengthCQUnits(t *testing.T) {
	tests := []struct {
		in       string
		want     Length
		roundStr string
	}{
		{"50cqw", Length{50, CQWUnit}, "cqw"},
		{"100cqh", Length{100, CQHUnit}, "cqh"},
		{"25cqi", Length{25, CQIUnit}, "cqi"},
		{"75cqb", Length{75, CQBUnit}, "cqb"},
		{"10cqmin", Length{10, CQMinUnit}, "cqmin"},
		{"90cqmax", Length{90, CQMaxUnit}, "cqmax"},
		{"50CQW", Length{50, CQWUnit}, "cqw"},
		{"50CqW", Length{50, CQWUnit}, "cqw"},
		{"  1.5cqi  ", Length{1.5, CQIUnit}, "cqi"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseLength(tc.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
			if got.Unit.String() != tc.roundStr {
				t.Errorf("Unit.String() = %q, want %q", got.Unit.String(), tc.roundStr)
			}
		})
	}
}

// TestResolveCQWithContainer covers the simple horizontal-writing-mode case
// where a parent has container-type: size with measured width/height and a
// child queries cqw/cqh.
func TestResolveCQWithContainer(t *testing.T) {
	lctx := NewLayoutContext(2000, 1000, 16)

	parent := &Node{
		Style: Style{ContainerType: ContainerTypeSize},
		Rect:  Rect{Width: 80, Height: 24},
	}
	child := &Node{}
	parent.Children = []*Node{child}

	root := NewContext(parent)
	childCtx := root.ChildAt(0)

	tests := []struct {
		name string
		l    Length
		want float64
	}{
		{"50cqw", Cqw(50), 40}, // 50% of 80
		{"50cqh", Cqh(50), 12}, // 50% of 24
		{"50cqi", Cqi(50), 40}, // inline == width in horizontal-tb
		{"50cqb", Cqb(50), 12}, // block == height in horizontal-tb
		{"50cqmin", Cqmin(50), 12},
		{"50cqmax", Cqmax(50), 40},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveLengthInContext(tc.l, lctx, 16, childCtx)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// TestResolveCQFallbackToViewport covers the "no container ancestor" case:
// every cq* unit must resolve against the viewport.
func TestResolveCQFallbackToViewport(t *testing.T) {
	lctx := NewLayoutContext(1000, 500, 16)
	parent := &Node{} // container-type: normal (default)
	child := &Node{}
	parent.Children = []*Node{child}
	root := NewContext(parent)
	childCtx := root.ChildAt(0)

	tests := []struct {
		name string
		l    Length
		want float64
	}{
		{"50cqw -> 50vw", Cqw(50), 500},
		{"50cqi -> 50vw", Cqi(50), 500},
		{"50cqh -> 50vh", Cqh(50), 250},
		{"50cqb -> 50vh", Cqb(50), 250},
		{"50cqmin -> 50vmin", Cqmin(50), 250},
		{"50cqmax -> 50vmax", Cqmax(50), 500},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveLengthInContext(tc.l, lctx, 16, childCtx)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// TestResolveCQInlineSizeOnly covers a container-type: inline-size parent:
// cqw/cqi use the container, but cqh/cqb fall through (to viewport here,
// since there is no further ancestor that answers the block axis).
func TestResolveCQInlineSizeOnly(t *testing.T) {
	lctx := NewLayoutContext(1000, 800, 16)
	parent := &Node{
		Style: Style{ContainerType: ContainerTypeInlineSize},
		Rect:  Rect{Width: 200, Height: 50},
	}
	child := &Node{}
	parent.Children = []*Node{child}
	root := NewContext(parent)
	childCtx := root.ChildAt(0)

	// cqw uses the container's inline size (200 -> 50%).
	if got := ResolveLengthInContext(Cqw(50), lctx, 16, childCtx); got != 100 {
		t.Errorf("cqw on inline-size container: got %v, want 100", got)
	}
	// cqi same axis.
	if got := ResolveLengthInContext(Cqi(50), lctx, 16, childCtx); got != 100 {
		t.Errorf("cqi on inline-size container: got %v, want 100", got)
	}
	// cqh falls back to viewport height (800 -> 50% = 400) per L4 spec.
	if got := ResolveLengthInContext(Cqh(50), lctx, 16, childCtx); got != 400 {
		t.Errorf("cqh on inline-size container: got %v, want 400 (viewport fallback)", got)
	}
	// cqb same axis.
	if got := ResolveLengthInContext(Cqb(50), lctx, 16, childCtx); got != 400 {
		t.Errorf("cqb on inline-size container: got %v, want 400 (viewport fallback)", got)
	}
}

// TestResolveCQNestedContainersNearestWins ensures the ancestor walk picks
// the nearest container, not the outermost.
func TestResolveCQNestedContainersNearestWins(t *testing.T) {
	lctx := NewLayoutContext(2000, 2000, 16)

	grandparent := &Node{
		Style: Style{ContainerType: ContainerTypeSize},
		Rect:  Rect{Width: 1000, Height: 800},
	}
	parent := &Node{
		Style: Style{ContainerType: ContainerTypeSize},
		Rect:  Rect{Width: 100, Height: 40},
	}
	child := &Node{}
	parent.Children = []*Node{child}
	grandparent.Children = []*Node{parent}

	root := NewContext(grandparent)
	parentCtx := root.ChildAt(0)
	childCtx := parentCtx.ChildAt(0)

	// cqw should resolve against parent (nearest), not grandparent.
	if got := ResolveLengthInContext(Cqw(50), lctx, 16, childCtx); got != 50 {
		t.Errorf("nearest container cqw: got %v, want 50 (50%% of parent.Width=100)", got)
	}
	if got := ResolveLengthInContext(Cqh(50), lctx, 16, childCtx); got != 20 {
		t.Errorf("nearest container cqh: got %v, want 20 (50%% of parent.Height=40)", got)
	}
}

// TestResolveCQInlineSizeChainPromotesBlockToAncestor verifies that an
// `inline-size` ancestor is skipped for block-axis queries and the walk
// continues to a `size` ancestor further up the chain.
func TestResolveCQInlineSizeChainPromotesBlockToAncestor(t *testing.T) {
	lctx := NewLayoutContext(2000, 2000, 16)

	outer := &Node{
		Style: Style{ContainerType: ContainerTypeSize},
		Rect:  Rect{Width: 600, Height: 300},
	}
	mid := &Node{
		Style: Style{ContainerType: ContainerTypeInlineSize},
		Rect:  Rect{Width: 200, Height: 100},
	}
	child := &Node{}
	mid.Children = []*Node{child}
	outer.Children = []*Node{mid}

	root := NewContext(outer)
	midCtx := root.ChildAt(0)
	childCtx := midCtx.ChildAt(0)

	// cqw: nearest inline-size container = mid (200) -> 50% = 100.
	if got := ResolveLengthInContext(Cqw(50), lctx, 16, childCtx); got != 100 {
		t.Errorf("cqw should use nearest mid container: got %v, want 100", got)
	}
	// cqh: mid does not satisfy block axis, walk up to outer (300) -> 50% = 150.
	if got := ResolveLengthInContext(Cqh(50), lctx, 16, childCtx); got != 150 {
		t.Errorf("cqh should skip inline-size mid and use outer: got %v, want 150", got)
	}
}

// TestResolveCQMinMaxAxes verifies cqmin/cqmax combine inline and block
// resolutions correctly.
func TestResolveCQMinMaxAxes(t *testing.T) {
	lctx := NewLayoutContext(2000, 2000, 16)
	parent := &Node{
		Style: Style{ContainerType: ContainerTypeSize},
		Rect:  Rect{Width: 80, Height: 24},
	}
	child := &Node{}
	parent.Children = []*Node{child}
	root := NewContext(parent)
	childCtx := root.ChildAt(0)

	// 50cqmin in 80x24 -> 50% of 24 = 12.
	if got := ResolveLengthInContext(Cqmin(50), lctx, 16, childCtx); got != 12 {
		t.Errorf("cqmin: got %v, want 12", got)
	}
	// 50cqmax in 80x24 -> 50% of 80 = 40.
	if got := ResolveLengthInContext(Cqmax(50), lctx, 16, childCtx); got != 40 {
		t.Errorf("cqmax: got %v, want 40", got)
	}
}

// TestResolveLengthCQFallbackWithoutNodeContext verifies the no-NodeContext
// variant falls back to viewport (documented behavior).
func TestResolveLengthCQFallbackWithoutNodeContext(t *testing.T) {
	lctx := NewLayoutContext(1000, 500, 16)
	if got := ResolveLength(Cqw(25), lctx, 16); got != 250 {
		t.Errorf("ResolveLength cqw with no context: got %v, want 250 (25%% of viewport width)", got)
	}
	if got := ResolveLength(Cqh(25), lctx, 16); got != 125 {
		t.Errorf("ResolveLength cqh with no context: got %v, want 125 (25%% of viewport height)", got)
	}
	if got := ResolveLength(Cqmin(50), lctx, 16); got != 250 {
		t.Errorf("ResolveLength cqmin with no context: got %v, want 250", got)
	}
	if got := ResolveLength(Cqmax(50), lctx, 16); got != 500 {
		t.Errorf("ResolveLength cqmax with no context: got %v, want 500", got)
	}
}

// TestContainerTypeOnStyleStruct verifies the new fields exist and have the
// expected zero values.
func TestContainerTypeOnStyleStruct(t *testing.T) {
	var s Style
	if s.ContainerType != ContainerTypeNormal {
		t.Errorf("default ContainerType: got %v, want ContainerTypeNormal", s.ContainerType)
	}
	if s.ContainerName != nil {
		t.Errorf("default ContainerName: got %v, want nil", s.ContainerName)
	}
	s.ContainerType = ContainerTypeSize
	s.ContainerName = []string{"card"}
	if s.ContainerType.String() != "size" {
		t.Errorf("ContainerType.String(): got %q, want %q", s.ContainerType.String(), "size")
	}
}

// TestCQUnitStringRoundtrip verifies all unit identifiers stringify back to
// their canonical CSS spelling.
func TestCQUnitStringRoundtrip(t *testing.T) {
	cases := []struct {
		u    LengthUnit
		want string
	}{
		{CQWUnit, "cqw"},
		{CQHUnit, "cqh"},
		{CQIUnit, "cqi"},
		{CQBUnit, "cqb"},
		{CQMinUnit, "cqmin"},
		{CQMaxUnit, "cqmax"},
	}
	for _, c := range cases {
		if got := c.u.String(); got != c.want {
			t.Errorf("Unit %v String() = %q, want %q", c.u, got, c.want)
		}
		if !c.u.IsContainerQuery() {
			t.Errorf("Unit %v should be IsContainerQuery()", c.u)
		}
	}
}

// TestParseLengthOtherUnits ensures the parser still handles existing units
// correctly (regression check for the cq* additions).
func TestParseLengthOtherUnits(t *testing.T) {
	cases := []struct {
		in   string
		want Length
	}{
		{"100px", Length{100, Pixels}},
		{"1.5rem", Length{1.5, RemUnit}},
		{"12", Length{12, Pixels}},
		{"50vw", Length{50, VwUnit}},
		{"50VW", Length{50, VwUnit}},
		{"-10px", Length{-10, Pixels}},
		{"1e2px", Length{100, Pixels}},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseLength(tc.in)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// TestParseLengthErrors checks malformed input.
func TestParseLengthErrors(t *testing.T) {
	cases := []string{
		"",
		"  ",
		"px",
		"100xx",
		"abc",
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			if _, err := ParseLength(in); err == nil {
				t.Errorf("expected error for %q", in)
			}
		})
	}
}
