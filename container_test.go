package layout

import (
	"strings"
	"testing"
)

// ───────────────────────────────────────────────────────────────
// Parser tests
// ───────────────────────────────────────────────────────────────

func TestParseContainerType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ContainerType
		wantErr bool
	}{
		{"empty", "", ContainerTypeNormal, false},
		{"normal", "normal", ContainerTypeNormal, false},
		{"size", "size", ContainerTypeSize, false},
		{"inline-size", "inline-size", ContainerTypeInlineSize, false},
		{"upper case", "SIZE", ContainerTypeSize, false},
		{"surrounding whitespace", "  inline-size  ", ContainerTypeInlineSize, false},
		{"invalid keyword", "block-size", ContainerTypeNormal, true},
		{"garbage", "foo bar", ContainerTypeNormal, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseContainerType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseContainerType(%q) err=%v, wantErr=%v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseContainerType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}

	// Error wording sanity check.
	if _, err := ParseContainerType("nonsense"); err == nil || !strings.Contains(err.Error(), "invalid container-type") {
		t.Errorf("expected 'invalid container-type' in error, got %v", err)
	}
}

func TestContainerTypeString(t *testing.T) {
	// Round-trip via Parse → String for the three valid values.
	cases := []string{"normal", "size", "inline-size"}
	for _, in := range cases {
		ct, err := ParseContainerType(in)
		if err != nil {
			t.Fatalf("ParseContainerType(%q) unexpected err: %v", in, err)
		}
		if ct.String() != in {
			t.Errorf("round-trip: %q → %v → %q", in, ct, ct.String())
		}
	}
	// Unknown enum integer renders as "unknown".
	if got := ContainerType(99).String(); got != "unknown" {
		t.Errorf("ContainerType(99).String() = %q, want %q", got, "unknown")
	}
}

func TestParseContainerName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ContainerName
		wantErr bool
	}{
		{"empty", "", nil, false},
		{"explicit none", "none", nil, false},
		{"upper-case none", "NONE", nil, false},
		{"single name", "card", ContainerName{"card"}, false},
		{"single name with hyphen", "user-card", ContainerName{"user-card"}, false},
		{"multiple names", "card sidebar header", ContainerName{"card", "sidebar", "header"}, false},
		{"reserved 'normal'", "normal", nil, true},
		{"reserved 'inherit'", "inherit", nil, true},
		{"reserved among valid", "card and sidebar", nil, true},
		{"invalid char", "card!", nil, true},
		{"starts with digit", "1card", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseContainerName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseContainerName(%q) err=%v, wantErr=%v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("ParseContainerName(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseContainerName(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}

	// Has() lookup is case-sensitive.
	names := ContainerName{"card", "sidebar"}
	if !names.Has("card") {
		t.Error("Has(\"card\") = false, want true")
	}
	if names.Has("Card") {
		t.Error("Has(\"Card\") = true, want false (case-sensitive)")
	}
	if names.Has("missing") {
		t.Error("Has(\"missing\") = true, want false")
	}
}

func TestParseContainer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName ContainerName
		wantType ContainerType
		wantErr  bool
	}{
		{"empty", "", nil, ContainerTypeNormal, false},
		{"name only", "card", ContainerName{"card"}, ContainerTypeNormal, false},
		{"type only — size", "size", nil, ContainerTypeSize, false},
		{"type only — inline-size", "inline-size", nil, ContainerTypeInlineSize, false},
		{"name / size", "card / size", ContainerName{"card"}, ContainerTypeSize, false},
		{"name / inline-size", "card / inline-size", ContainerName{"card"}, ContainerTypeInlineSize, false},
		{"none keyword", "none", nil, ContainerTypeNormal, false},
		{"invalid type after slash", "card / weird", nil, ContainerTypeNormal, true},
		{"invalid name before slash", "1bad / size", nil, ContainerTypeNormal, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotType, err := ParseContainer(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseContainer(%q) err=%v, wantErr=%v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotType != tt.wantType {
				t.Errorf("ParseContainer(%q) type = %v, want %v", tt.input, gotType, tt.wantType)
			}
			if len(gotName) != len(tt.wantName) {
				t.Fatalf("ParseContainer(%q) name = %v, want %v", tt.input, gotName, tt.wantName)
			}
			for i := range gotName {
				if gotName[i] != tt.wantName[i] {
					t.Errorf("ParseContainer(%q) name[%d] = %q, want %q", tt.input, i, gotName[i], tt.wantName[i])
				}
			}
		})
	}
}

// ───────────────────────────────────────────────────────────────
// Resolver tests
// ───────────────────────────────────────────────────────────────

// containerTree builds a linear ancestor chain:
//
//   root -> middle -> leaf
//
// Each node carries a Rect and (optionally) a ContainerType. Returns
// the leaf NodeContext, which is the deepest descendant — exactly what
// callers of ResolveLengthInContext typically pass in.
func containerTree(root, middle, leaf containerNodeSpec) *NodeContext {
	rootNode := root.build()
	middleNode := middle.build()
	leafNode := leaf.build()
	rootNode.Children = []*Node{middleNode}
	middleNode.Children = []*Node{leafNode}

	rootCtx := NewContext(rootNode)
	// Children() returns the children of rootCtx as NodeContext slice;
	// their parent is wired to rootCtx so Parent()/Ancestors() work.
	middleCtx := rootCtx.Children()[0]
	return middleCtx.Children()[0]
}

type containerNodeSpec struct {
	rect          Rect
	containerType ContainerType
	writingMode   WritingMode
}

func (s containerNodeSpec) build() *Node {
	return &Node{
		Style: Style{
			ContainerType: s.containerType,
			WritingMode:   s.writingMode,
		},
		Rect: s.rect,
	}
}

func TestResolveLengthInContext_Viewport(t *testing.T) {
	// No query container in the ancestor chain → cq* falls back to viewport.
	ctx := NewLayoutContext(1920, 1080, 16)
	leaf := containerTree(
		containerNodeSpec{rect: Rect{Width: 1920, Height: 1080}},
		containerNodeSpec{rect: Rect{Width: 800, Height: 600}},
		containerNodeSpec{rect: Rect{Width: 400, Height: 300}},
	)

	tests := []struct {
		name string
		l    Length
		want float64
	}{
		{"50cqw → 50% of viewport width", Cqw(50), 960},
		{"50cqh → 50% of viewport height", Cqh(50), 540},
		{"50cqi → inline = viewport width (horizontal-tb default)", Cqi(50), 960},
		{"50cqb → block = viewport height (horizontal-tb default)", Cqb(50), 540},
		{"50cqmin → 50% of min(vw,vh)", Cqmin(50), 540},
		{"50cqmax → 50% of max(vw,vh)", Cqmax(50), 960},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLengthInContext(tt.l, ctx, 16, leaf)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveLengthInContext_SizeContainer(t *testing.T) {
	// The middle ancestor is a `size` container of 800x600. cqw/cqh
	// resolve against it.
	ctx := NewLayoutContext(1920, 1080, 16)
	leaf := containerTree(
		containerNodeSpec{rect: Rect{Width: 1920, Height: 1080}},
		containerNodeSpec{rect: Rect{Width: 800, Height: 600}, containerType: ContainerTypeSize},
		containerNodeSpec{rect: Rect{Width: 400, Height: 300}},
	)

	tests := []struct {
		name string
		l    Length
		want float64
	}{
		{"50cqw of 800px container width", Cqw(50), 400},
		{"50cqh of 600px container height", Cqh(50), 300},
		{"50cqi (horizontal-tb container) → cqw", Cqi(50), 400},
		{"50cqb (horizontal-tb container) → cqh", Cqb(50), 300},
		{"50cqmin → 50% of min(800,600)", Cqmin(50), 300},
		{"50cqmax → 50% of max(800,600)", Cqmax(50), 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLengthInContext(tt.l, ctx, 16, leaf)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveLengthInContext_InlineSizeContainer(t *testing.T) {
	// The middle ancestor is an `inline-size` container of 800x600.
	// cqw and cqi use its inline (= width); cqh and cqb fall through
	// to the viewport.
	ctx := NewLayoutContext(1920, 1080, 16)
	leaf := containerTree(
		containerNodeSpec{rect: Rect{Width: 1920, Height: 1080}},
		containerNodeSpec{rect: Rect{Width: 800, Height: 600}, containerType: ContainerTypeInlineSize},
		containerNodeSpec{rect: Rect{Width: 400, Height: 300}},
	)

	tests := []struct {
		name string
		l    Length
		want float64
	}{
		{"50cqw against inline-size container width", Cqw(50), 400},
		{"50cqi against inline-size container inline", Cqi(50), 400},
		{"50cqh falls back to viewport height (no size ancestor)", Cqh(50), 540},
		{"50cqb falls back to viewport block (no size ancestor)", Cqb(50), 540},
		{"50cqmin: inline=container(800), block=viewport(1080) → min=800 → 400", Cqmin(50), 400},
		{"50cqmax: inline=container(800), block=viewport(1080) → max=1080 → 540", Cqmax(50), 540},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLengthInContext(tt.l, ctx, 16, leaf)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveLengthInContext_NestedContainers(t *testing.T) {
	// Inner size container (300x200) wins over outer size container
	// (1000x800). cqw/cqh resolve against the inner.
	ctx := NewLayoutContext(1920, 1080, 16)
	leaf := containerTree(
		containerNodeSpec{rect: Rect{Width: 1000, Height: 800}, containerType: ContainerTypeSize},
		containerNodeSpec{rect: Rect{Width: 300, Height: 200}, containerType: ContainerTypeSize},
		containerNodeSpec{rect: Rect{Width: 100, Height: 50}},
	)

	if got := ResolveLengthInContext(Cqw(50), ctx, 16, leaf); got != 150 {
		t.Errorf("Cqw(50) against inner size container = %v, want 150", got)
	}
	if got := ResolveLengthInContext(Cqh(50), ctx, 16, leaf); got != 100 {
		t.Errorf("Cqh(50) against inner size container = %v, want 100", got)
	}
}

func TestResolveLengthInContext_NonContainerUnit(t *testing.T) {
	// Non-cq* units must delegate to ResolveLength with no surprises.
	ctx := NewLayoutContext(1920, 1080, 16)
	leaf := containerTree(
		containerNodeSpec{rect: Rect{Width: 1920, Height: 1080}, containerType: ContainerTypeSize},
		containerNodeSpec{rect: Rect{Width: 800, Height: 600}, containerType: ContainerTypeSize},
		containerNodeSpec{rect: Rect{Width: 400, Height: 300}},
	)

	tests := []struct {
		name string
		l    Length
		want float64
	}{
		{"px passes through", Px(42), 42},
		{"em uses currentFontSize", Em(2), 32},
		{"rem uses RootFontSize", Rem(2), 32},
		{"vw uses viewport", Vw(50), 960},
		{"unbounded → MaxFloat64", UnboundedLength(), ResolveLength(UnboundedLength(), ctx, 16)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLengthInContext(tt.l, ctx, 16, leaf)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveLengthInContextCqZeroContainerReturnsZero(t *testing.T) {
	// Container with zero width — cqw should resolve to 0, not raw l.Value.
	leaf := containerTree(
		containerNodeSpec{rect: Rect{Width: 0, Height: 0}},
		containerNodeSpec{rect: Rect{Width: 0, Height: 0}, containerType: ContainerTypeSize},
		containerNodeSpec{rect: Rect{Width: 0, Height: 0}},
	)
	ctx := NewLayoutContext(0, 0, 16) // also zero viewport, so no fallback can save it

	got := ResolveLengthInContext(Cqw(50), ctx, 16, leaf)
	if got != 0 {
		t.Fatalf("Cqw(50) with zero-size container expected 0, got %v", got)
	}

	got = ResolveLengthInContext(Cqh(50), ctx, 16, leaf)
	if got != 0 {
		t.Fatalf("Cqh(50) with zero-size container expected 0, got %v", got)
	}

	got = ResolveLengthInContext(Cqmin(50), ctx, 16, leaf)
	if got != 0 {
		t.Fatalf("Cqmin(50) with zero-size container expected 0, got %v", got)
	}
}

func TestResolveLengthInContext_VerticalWritingMode(t *testing.T) {
	// Vertical-RL container: cqi maps to height, cqb maps to width.
	ctx := NewLayoutContext(1920, 1080, 16)
	leaf := containerTree(
		containerNodeSpec{rect: Rect{Width: 1920, Height: 1080}},
		containerNodeSpec{
			rect:          Rect{Width: 800, Height: 600},
			containerType: ContainerTypeSize,
			writingMode:   WritingModeVerticalRL,
		},
		containerNodeSpec{rect: Rect{Width: 400, Height: 300}},
	)

	// cqi against a vertical container resolves against block-physical
	// (= the container's height = 600).
	if got := ResolveLengthInContext(Cqi(50), ctx, 16, leaf); got != 300 {
		t.Errorf("Cqi(50) in vertical-rl container = %v, want 300 (50%% of height 600)", got)
	}
	// cqb against a vertical container resolves against width (= 800).
	if got := ResolveLengthInContext(Cqb(50), ctx, 16, leaf); got != 400 {
		t.Errorf("Cqb(50) in vertical-rl container = %v, want 400 (50%% of width 800)", got)
	}
	// cqw and cqh remain physical regardless of writing mode.
	if got := ResolveLengthInContext(Cqw(50), ctx, 16, leaf); got != 400 {
		t.Errorf("Cqw(50) in vertical-rl container = %v, want 400 (physical width)", got)
	}
	if got := ResolveLengthInContext(Cqh(50), ctx, 16, leaf); got != 300 {
		t.Errorf("Cqh(50) in vertical-rl container = %v, want 300 (physical height)", got)
	}
}
