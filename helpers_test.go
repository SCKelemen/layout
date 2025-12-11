package layout

import (
	"testing"
)

func TestSpacingHelpers(t *testing.T) {
	// Test Uniform spacing
	uniform := Uniform(10)
	if uniform.Top != 10 || uniform.Right != 10 || uniform.Bottom != 10 || uniform.Left != 10 {
		t.Error("Uniform spacing should set all sides to same value")
	}

	// Test Horizontal spacing
	horizontal := Horizontal(20)
	if horizontal.Top != 0 || horizontal.Bottom != 0 {
		t.Error("Horizontal spacing should have 0 top/bottom")
	}
	if horizontal.Left != 20 || horizontal.Right != 20 {
		t.Error("Horizontal spacing should set left/right")
	}

	// Test Vertical spacing
	vertical := Vertical(30)
	if vertical.Left != 0 || vertical.Right != 0 {
		t.Error("Vertical spacing should have 0 left/right")
	}
	if vertical.Top != 30 || vertical.Bottom != 30 {
		t.Error("Vertical spacing should set top/bottom")
	}
}

func TestGridTrackHelpers(t *testing.T) {
	// Test FixedTrack
	fixed := FixedTrack(100)
	if fixed.MinSize != 100 || fixed.MaxSize != 100 || fixed.Fraction != 0 {
		t.Error("FixedTrack should have min=max=size, fraction=0")
	}

	// Test MinMaxTrack
	minmax := MinMaxTrack(50, 150)
	if minmax.MinSize != 50 || minmax.MaxSize != 150 || minmax.Fraction != 0 {
		t.Error("MinMaxTrack should have correct min/max, fraction=0")
	}

	// Test FractionTrack
	fraction := FractionTrack(2)
	if fraction.MinSize != 0 || fraction.MaxSize == 0 || fraction.Fraction != 2 {
		t.Error("FractionTrack should have fraction set, unbounded max")
	}

	// Test AutoTrack
	auto := AutoTrack()
	if auto.MinSize != 0 || auto.MaxSize == 0 || auto.Fraction != 0 {
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

