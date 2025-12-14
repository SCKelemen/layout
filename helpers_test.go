package layout

import (
	"testing"
)

func TestSpacingHelpers(t *testing.T) {
	// Test Uniform spacing
	uniform := Uniform(Px(10))
	if uniform.Top.Value != 10 || uniform.Right.Value != 10 || uniform.Bottom.Value != 10 || uniform.Left.Value != 10 {
		t.Error("Uniform spacing should set all sides to same value")
	}

	// Test Horizontal spacing
	horizontal := Horizontal(Px(20))
	if horizontal.Top.Value != 0 || horizontal.Bottom.Value != 0 {
		t.Error("Horizontal spacing should have 0 top/bottom")
	}
	if horizontal.Left.Value != 20 || horizontal.Right.Value != 20 {
		t.Error("Horizontal spacing should set left/right")
	}

	// Test Vertical spacing
	vertical := Vertical(Px(30))
	if vertical.Left.Value != 0 || vertical.Right.Value != 0 {
		t.Error("Vertical spacing should have 0 left/right")
	}
	if vertical.Top.Value != 30 || vertical.Bottom.Value != 30 {
		t.Error("Vertical spacing should set top/bottom")
	}
}

func TestGridTrackHelpers(t *testing.T) {
	// Test FixedTrack
	fixed := FixedTrack(Px(100))
	if fixed.MinSize.Value != 100 || fixed.MaxSize.Value != 100 || fixed.Fraction != 0 {
		t.Error("FixedTrack should have min=max=size, fraction=0")
	}

	// Test MinMaxTrack
	minmax := MinMaxTrack(Px(50), Px(150))
	if minmax.MinSize.Value != 50 || minmax.MaxSize.Value != 150 || minmax.Fraction != 0 {
		t.Error("MinMaxTrack should have correct min/max, fraction=0")
	}

	// Test FractionTrack
	fraction := FractionTrack(2)
	if fraction.MinSize.Value != 0 || fraction.MaxSize.Value == 0 || fraction.Fraction != 2 {
		t.Error("FractionTrack should have fraction set, unbounded max")
	}

	// Test AutoTrack
	auto := AutoTrack()
	if auto.MinSize.Value != 0 || auto.MaxSize.Value == 0 || auto.Fraction != 0 {
		t.Error("AutoTrack should have unbounded max, fraction=0")
	}
}

func TestUnconstrained(t *testing.T) {
	constraints := Unconstrained()
	if constraints.MinWidth != 0 || constraints.MaxWidth == 0 {
		t.Error("Unconstrained should have min=0, max=unbounded")
	}
	if constraints.MinHeight != 0 || constraints.MaxHeight == 0 {
		t.Error("Unconstrained should have min=0, max=unbounded")
	}
}
