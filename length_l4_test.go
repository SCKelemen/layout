package layout

import (
	"math"
	"strings"
	"testing"
)

// newL4Ctx returns a LayoutContext suitable for L4 unit tests.
//   - 1000x800 viewport keeps percentage math human-readable
//   - 16pt root font size matches existing tests
func newL4Ctx() *LayoutContext {
	return NewLayoutContext(1000, 800, 16)
}

// TestL4UnitConstructors verifies every L4 constructor produces the expected
// (value, unit) pair.
func TestL4UnitConstructors(t *testing.T) {
	cases := []struct {
		name string
		got  Length
		val  float64
		unit LengthUnit
	}{
		{"Lh", Lh(2), 2, LhUnit},
		{"Rlh", Rlh(2), 2, RlhUnit},
		{"Ic", Ic(2), 2, IcUnit},
		{"Ric", Ric(2), 2, RicUnit},
		{"Cap", Cap(2), 2, CapUnit},
		{"Rcap", Rcap(2), 2, RcapUnit},
		{"Rch", Rch(2), 2, RchUnit},
		{"Rex", Rex(2), 2, RexUnit},
		{"Vi", Vi(50), 50, ViUnit},
		{"Vb", Vb(50), 50, VbUnit},
		{"Svw", Svw(50), 50, SvwUnit},
		{"Svh", Svh(50), 50, SvhUnit},
		{"Svi", Svi(50), 50, SviUnit},
		{"Svb", Svb(50), 50, SvbUnit},
		{"Svmin", Svmin(50), 50, SvminUnit},
		{"Svmax", Svmax(50), 50, SvmaxUnit},
		{"Lvw", Lvw(50), 50, LvwUnit},
		{"Lvh", Lvh(50), 50, LvhUnit},
		{"Lvi", Lvi(50), 50, LviUnit},
		{"Lvb", Lvb(50), 50, LvbUnit},
		{"Lvmin", Lvmin(50), 50, LvminUnit},
		{"Lvmax", Lvmax(50), 50, LvmaxUnit},
		{"Dvw", Dvw(50), 50, DvwUnit},
		{"Dvh", Dvh(50), 50, DvhUnit},
		{"Dvi", Dvi(50), 50, DviUnit},
		{"Dvb", Dvb(50), 50, DvbUnit},
		{"Dvmin", Dvmin(50), 50, DvminUnit},
		{"Dvmax", Dvmax(50), 50, DvmaxUnit},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got.Value != tc.val {
				t.Errorf("Value = %v, want %v", tc.got.Value, tc.val)
			}
			if tc.got.Unit != tc.unit {
				t.Errorf("Unit = %v, want %v", tc.got.Unit, tc.unit)
			}
		})
	}
}

// TestL4FontRelativeResolve covers the lh/rlh/cap/rcap/rex resolutions, all
// of which collapse to "one row" in terminal context (current or root font
// size).
func TestL4FontRelativeResolve(t *testing.T) {
	ctx := newL4Ctx()
	const fontSize = 20.0

	cases := []struct {
		name string
		got  float64
		want float64
	}{
		{"2lh = 2 * currentFontSize", ResolveLength(Lh(2), ctx, fontSize), 2 * fontSize},
		{"2rlh = 2 * RootFontSize", ResolveLength(Rlh(2), ctx, fontSize), 2 * ctx.RootFontSize},
		{"2cap = 2 * currentFontSize", ResolveLength(Cap(2), ctx, fontSize), 2 * fontSize},
		{"2rcap = 2 * RootFontSize", ResolveLength(Rcap(2), ctx, fontSize), 2 * ctx.RootFontSize},
		{"2rex = 2 * RootFontSize", ResolveLength(Rex(2), ctx, fontSize), 2 * ctx.RootFontSize},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("got %v, want %v", tc.got, tc.want)
			}
		})
	}
}

// TestL4ChAliases verifies that rch resolves to the same value as ch when
// the element font size matches the root font size, and that ic / ric
// resolve to exactly 2x the corresponding ch value.
func TestL4ChAliases(t *testing.T) {
	ctx := newL4Ctx()
	const fontSize = 16.0 // matches ctx.RootFontSize

	gotCh := ResolveLength(Ch(10), ctx, fontSize)
	gotRch := ResolveLength(Rch(10), ctx, fontSize)
	if gotCh != gotRch {
		t.Errorf("10rch (%v) should equal 10ch (%v) when currentFontSize == RootFontSize", gotRch, gotCh)
	}

	gotIc := ResolveLength(Ic(5), ctx, fontSize)
	if want := 2 * ResolveLength(Ch(5), ctx, fontSize); gotIc != want {
		t.Errorf("5ic = %v, want 2 * 5ch = %v", gotIc, want)
	}

	gotRic := ResolveLength(Ric(5), ctx, fontSize)
	if want := 2 * ResolveLength(Rch(5), ctx, fontSize); gotRic != want {
		t.Errorf("5ric = %v, want 2 * 5rch = %v", gotRic, want)
	}
}

// TestL4RchVsCh confirms rch uses the root font size, not the current font
// size: changing currentFontSize must not affect rch.
func TestL4RchVsCh(t *testing.T) {
	ctx := newL4Ctx() // RootFontSize = 16

	at16 := ResolveLength(Rch(10), ctx, 16)
	at32 := ResolveLength(Rch(10), ctx, 32)
	if at16 != at32 {
		t.Errorf("rch should be independent of currentFontSize: 16 -> %v, 32 -> %v", at16, at32)
	}

	// ch, on the other hand, should change.
	chAt16 := ResolveLength(Ch(10), ctx, 16)
	chAt32 := ResolveLength(Ch(10), ctx, 32)
	if chAt16 == chAt32 {
		t.Errorf("ch should depend on currentFontSize, but got identical %v at both font sizes", chAt16)
	}
}

// TestL4LogicalViewport verifies the writing-mode simplification: vi
// resolves like vw, vb resolves like vh. When ResolveLength gains a
// writing-mode parameter the assertions here will need to be parameterized;
// see the comment in length_l4.go.
func TestL4LogicalViewport(t *testing.T) {
	ctx := newL4Ctx() // 1000 x 800

	if got, want := ResolveLength(Vi(50), ctx, 16), ResolveLength(Vw(50), ctx, 16); got != want {
		t.Errorf("50vi = %v, want 50vw = %v", got, want)
	}
	if got, want := ResolveLength(Vb(50), ctx, 16), ResolveLength(Vh(50), ctx, 16); got != want {
		t.Errorf("50vb = %v, want 50vh = %v", got, want)
	}
}

// TestL4ViewportVariants checks the eighteen sv*/lv*/dv* aliases all
// resolve identically to the underlying base viewport unit.
func TestL4ViewportVariants(t *testing.T) {
	ctx := newL4Ctx() // 1000 x 800
	const fs = 16.0

	cases := []struct {
		name   string
		got    float64
		wantTo Length
	}{
		// Small viewport family.
		{"svw == vw", ResolveLength(Svw(40), ctx, fs), Vw(40)},
		{"svh == vh", ResolveLength(Svh(40), ctx, fs), Vh(40)},
		{"svi == vi", ResolveLength(Svi(40), ctx, fs), Vi(40)},
		{"svb == vb", ResolveLength(Svb(40), ctx, fs), Vb(40)},
		{"svmin == vmin", ResolveLength(Svmin(40), ctx, fs), Vmin(40)},
		{"svmax == vmax", ResolveLength(Svmax(40), ctx, fs), Vmax(40)},
		// Large viewport family.
		{"lvw == vw", ResolveLength(Lvw(40), ctx, fs), Vw(40)},
		{"lvh == vh", ResolveLength(Lvh(40), ctx, fs), Vh(40)},
		{"lvi == vi", ResolveLength(Lvi(40), ctx, fs), Vi(40)},
		{"lvb == vb", ResolveLength(Lvb(40), ctx, fs), Vb(40)},
		{"lvmin == vmin", ResolveLength(Lvmin(40), ctx, fs), Vmin(40)},
		{"lvmax == vmax", ResolveLength(Lvmax(40), ctx, fs), Vmax(40)},
		// Dynamic viewport family.
		{"dvw == vw", ResolveLength(Dvw(40), ctx, fs), Vw(40)},
		{"dvh == vh", ResolveLength(Dvh(40), ctx, fs), Vh(40)},
		{"dvi == vi", ResolveLength(Dvi(40), ctx, fs), Vi(40)},
		{"dvb == vb", ResolveLength(Dvb(40), ctx, fs), Vb(40)},
		{"dvmin == vmin", ResolveLength(Dvmin(40), ctx, fs), Vmin(40)},
		{"dvmax == vmax", ResolveLength(Dvmax(40), ctx, fs), Vmax(40)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := ResolveLength(tc.wantTo, ctx, fs)
			if math.Abs(tc.got-want) > 1e-9 {
				t.Errorf("got %v, want %v", tc.got, want)
			}
		})
	}
}

// TestParseLength_L4 verifies the parser accepts every new L4 identifier
// and emits the correct typed Length, both at canonical lowercase casing
// and at mixed/upper casing (CSS unit identifiers are case-insensitive).
func TestParseLength_L4(t *testing.T) {
	cases := []struct {
		in       string
		wantVal  float64
		wantUnit LengthUnit
	}{
		// Font-relative.
		{"10lh", 10, LhUnit},
		{"10rlh", 10, RlhUnit},
		{"10ic", 10, IcUnit},
		{"10ric", 10, RicUnit},
		{"10cap", 10, CapUnit},
		{"10rcap", 10, RcapUnit},
		{"10rch", 10, RchUnit},
		{"10rex", 10, RexUnit},
		// Logical viewport.
		{"10vi", 10, ViUnit},
		{"10vb", 10, VbUnit},
		// Small viewport family.
		{"10svw", 10, SvwUnit},
		{"10svh", 10, SvhUnit},
		{"10svi", 10, SviUnit},
		{"10svb", 10, SvbUnit},
		{"10svmin", 10, SvminUnit},
		{"10svmax", 10, SvmaxUnit},
		// Large viewport family.
		{"10lvw", 10, LvwUnit},
		{"10lvh", 10, LvhUnit},
		{"10lvi", 10, LviUnit},
		{"10lvb", 10, LvbUnit},
		{"10lvmin", 10, LvminUnit},
		{"10lvmax", 10, LvmaxUnit},
		// Dynamic viewport family.
		{"10dvw", 10, DvwUnit},
		{"10dvh", 10, DvhUnit},
		{"10dvi", 10, DviUnit},
		{"10dvb", 10, DvbUnit},
		{"10dvmin", 10, DvminUnit},
		{"10dvmax", 10, DvmaxUnit},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseLength(tc.in)
			if err != nil {
				t.Fatalf("ParseLength(%q) returned error: %v", tc.in, err)
			}
			if got.Value != tc.wantVal {
				t.Errorf("Value = %v, want %v", got.Value, tc.wantVal)
			}
			if got.Unit != tc.wantUnit {
				t.Errorf("Unit = %v, want %v", got.Unit, tc.wantUnit)
			}
		})
	}
}

// TestParseLength_CaseInsensitive ensures CSS unit identifiers are matched
// without regard to case for every L4 identifier.
func TestParseLength_CaseInsensitive(t *testing.T) {
	identifiers := []string{
		"lh", "rlh", "ic", "ric", "cap", "rcap", "rch", "rex",
		"vi", "vb",
		"svw", "svh", "svi", "svb", "svmin", "svmax",
		"lvw", "lvh", "lvi", "lvb", "lvmin", "lvmax",
		"dvw", "dvh", "dvi", "dvb", "dvmin", "dvmax",
	}
	for _, id := range identifiers {
		lower := "10" + id
		upper := "10" + strings.ToUpper(id)
		mixed := "10" + toggleCase(id)

		gotLower, errL := ParseLength(lower)
		if errL != nil {
			t.Fatalf("ParseLength(%q) = error %v", lower, errL)
		}
		for _, alt := range []string{upper, mixed} {
			got, err := ParseLength(alt)
			if err != nil {
				t.Errorf("ParseLength(%q) returned error: %v", alt, err)
				continue
			}
			if got.Unit != gotLower.Unit {
				t.Errorf("ParseLength(%q).Unit = %v, want %v (same as %q)", alt, got.Unit, gotLower.Unit, lower)
			}
			if got.Value != gotLower.Value {
				t.Errorf("ParseLength(%q).Value = %v, want %v", alt, got.Value, gotLower.Value)
			}
		}
	}
}

// TestParseLength_ExistingUnits sanity-checks that the new parser also
// handles the pre-existing CSS units (no regression in coverage).
func TestParseLength_ExistingUnits(t *testing.T) {
	cases := map[string]LengthUnit{
		"10px":   Pixels,
		"10pt":   PtUnit,
		"10pc":   PcUnit,
		"10in":   InUnit,
		"10cm":   CmUnit,
		"10mm":   MmUnit,
		"10Q":    QUnit,
		"2em":    EmUnit,
		"1.5rem": RemUnit,
		"80ch":   ChUnit,
		"50vh":   VhUnit,
		"50vw":   VwUnit,
		"25vmin": VminUnit,
		"75vmax": VmaxUnit,
	}
	for in, wantUnit := range cases {
		t.Run(in, func(t *testing.T) {
			got, err := ParseLength(in)
			if err != nil {
				t.Fatalf("ParseLength(%q) returned error: %v", in, err)
			}
			if got.Unit != wantUnit {
				t.Errorf("Unit = %v, want %v", got.Unit, wantUnit)
			}
		})
	}
}

// TestParseLength_Errors covers the three error paths of the parser.
func TestParseLength_Errors(t *testing.T) {
	cases := []string{
		"",         // empty
		"   ",      // whitespace only
		"px",       // missing number
		"10",       // missing unit
		"10foo",    // unknown unit
		"abcpx",    // non-numeric prefix
		"10.5cqw",  // cq* units belong to another agent / are not parsed here
		"10 px",    // embedded whitespace (not stripped beyond ends)
		"10vminxy", // unknown unit suffix
	}
	for _, s := range cases {
		t.Run(s, func(t *testing.T) {
			if _, err := ParseLength(s); err == nil {
				t.Errorf("ParseLength(%q) returned no error, expected one", s)
			}
		})
	}
}

// TestParseLengthAndResolve_RchAliasesCh demonstrates the documented
// invariant that "10rch" resolves to the same number as "10ch" in this
// library (terminal context).
func TestParseLengthAndResolve_RchAliasesCh(t *testing.T) {
	ctx := newL4Ctx() // RootFontSize matches the current font size below
	const fs = 16.0

	rch, err := ParseLength("10rch")
	if err != nil {
		t.Fatalf("ParseLength: %v", err)
	}
	ch, err := ParseLength("10ch")
	if err != nil {
		t.Fatalf("ParseLength: %v", err)
	}
	if a, b := ResolveLength(rch, ctx, fs), ResolveLength(ch, ctx, fs); a != b {
		t.Errorf("10rch resolved to %v, 10ch resolved to %v; expected equal", a, b)
	}
}

// TestL4UnitString verifies LengthUnit.String returns the canonical CSS
// identifier for each new unit (used by Length.String for debug output).
func TestL4UnitString(t *testing.T) {
	cases := map[LengthUnit]string{
		LhUnit: "lh", RlhUnit: "rlh", IcUnit: "ic", RicUnit: "ric",
		CapUnit: "cap", RcapUnit: "rcap", RchUnit: "rch", RexUnit: "rex",
		ViUnit: "vi", VbUnit: "vb",
		SvwUnit: "svw", SvhUnit: "svh", SviUnit: "svi", SvbUnit: "svb",
		SvminUnit: "svmin", SvmaxUnit: "svmax",
		LvwUnit: "lvw", LvhUnit: "lvh", LviUnit: "lvi", LvbUnit: "lvb",
		LvminUnit: "lvmin", LvmaxUnit: "lvmax",
		DvwUnit: "dvw", DvhUnit: "dvh", DviUnit: "dvi", DvbUnit: "dvb",
		DvminUnit: "dvmin", DvmaxUnit: "dvmax",
	}
	for u, want := range cases {
		if got := u.String(); got != want {
			t.Errorf("LengthUnit(%d).String() = %q, want %q", int(u), got, want)
		}
	}
}

// toggleCase returns s with each letter's case flipped, used to construct
// mixed-case variants for the case-insensitivity test (e.g. "lh" -> "Lh").
func toggleCase(s string) string {
	b := []rune(s)
	for i, r := range b {
		switch {
		case r >= 'a' && r <= 'z':
			if i%2 == 0 {
				b[i] = r - ('a' - 'A')
			}
		case r >= 'A' && r <= 'Z':
			if i%2 == 1 {
				b[i] = r + ('a' - 'A')
			}
		}
	}
	return string(b)
}
