package layout

import (
	"testing"

	"github.com/SCKelemen/units"
)

// TestLengthUnitAliasing is a smoke test for Phase 1 of the units<->layout
// migration. It verifies that the layout-side named unit constants are
// re-exports of the canonical github.com/SCKelemen/units constants and that
// the layout constructors continue to produce Lengths tagged with the
// expected units. It also exercises ResolveLength against the aliased types
// to confirm the resolver still produces identical numeric output after the
// type-system swap.
func TestLengthUnitAliasing(t *testing.T) {
	t.Run("named constants alias to units constants", func(t *testing.T) {
		cases := []struct {
			name   string
			layout LengthUnit
			canon  units.LengthUnit
		}{
			{"Pixels", Pixels, units.PX},
			{"PtUnit", PtUnit, units.PT},
			{"PcUnit", PcUnit, units.PC},
			{"InUnit", InUnit, units.IN},
			{"CmUnit", CmUnit, units.CM},
			{"MmUnit", MmUnit, units.MM},
			{"QUnit", QUnit, units.QQ},
			{"EmUnit", EmUnit, units.EM},
			{"RemUnit", RemUnit, units.REM},
			{"ChUnit", ChUnit, units.CH},
			{"VhUnit", VhUnit, units.VH},
			{"VwUnit", VwUnit, units.VW},
			{"VmaxUnit", VmaxUnit, units.VMAX},
			{"VminUnit", VminUnit, units.VMIN},
		}
		for _, tc := range cases {
			if tc.layout != tc.canon {
				t.Errorf("%s = %q, want %q", tc.name, tc.layout, tc.canon)
			}
		}
	})

	t.Run("constructors produce the right unit", func(t *testing.T) {
		cases := []struct {
			name string
			got  LengthUnit
			want LengthUnit
		}{
			{"Px", Px(10).Unit, Pixels},
			{"Em", Em(1.5).Unit, EmUnit},
			{"Rem", Rem(1).Unit, RemUnit},
			{"Vw", Vw(100).Unit, VwUnit},
		}
		for _, tc := range cases {
			if tc.got != tc.want {
				t.Errorf("%s().Unit = %q, want %q", tc.name, tc.got, tc.want)
			}
		}
	})

	t.Run("ResolveLength behavior unchanged", func(t *testing.T) {
		ctx := NewLayoutContext(1000, 800, 16)
		if got := ResolveLength(Px(50), ctx, 16); got != 50 {
			t.Errorf("ResolveLength(50px) = %v, want 50", got)
		}
		if got := ResolveLength(Em(2), ctx, 16); got != 32 {
			t.Errorf("ResolveLength(2em with 16pt font) = %v, want 32", got)
		}
		if got := ResolveLength(Rem(1), ctx, 16); got != 16 {
			t.Errorf("ResolveLength(1rem with 16pt root) = %v, want 16", got)
		}
		if got := ResolveLength(Vw(50), ctx, 16); got != 500 {
			t.Errorf("ResolveLength(50vw of 1000) = %v, want 500", got)
		}
		if got := ResolveLength(Vh(25), ctx, 16); got != 200 {
			t.Errorf("ResolveLength(25vh of 800) = %v, want 200", got)
		}
	})

	t.Run("layout Length is interoperable with units.Length", func(t *testing.T) {
		// Because Length is a type alias for units.Length, a value produced
		// by units.Px is directly assignable to a layout.Length variable
		// without conversion. This is the whole point of Phase 1.
		var l Length = units.Px(42)
		if l.Value != 42 || l.Unit != Pixels {
			t.Errorf("units.Px(42) assigned to layout.Length = %+v, want Value=42 Unit=px", l)
		}
	})
}
