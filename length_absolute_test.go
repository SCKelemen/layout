package layout

import (
	"math"
	"testing"
)

// TestAbsoluteLengthConstructors tests constructors for absolute length units
func TestAbsoluteLengthConstructors(t *testing.T) {
	tests := []struct {
		name     string
		length   Length
		wantVal  float64
		wantUnit LengthUnit
	}{
		{"Pt", Pt(12), 12, PtUnit},
		{"Pc", Pc(1), 1, PcUnit},
		{"In", In(1), 1, InUnit},
		{"Cm", Cm(2.54), 2.54, CmUnit},
		{"Mm", Mm(25.4), 25.4, MmUnit},
		{"Q", Q(101.6), 101.6, QUnit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.length.Value != tt.wantVal {
				t.Errorf("Value = %v, want %v", tt.length.Value, tt.wantVal)
			}
			if tt.length.Unit != tt.wantUnit {
				t.Errorf("Unit = %v, want %v", tt.length.Unit, tt.wantUnit)
			}
		})
	}
}

// TestResolveLengthPt tests resolving point units (1pt = 96/72 px)
func TestResolveLengthPt(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)
	fontSize := 16.0

	tests := []struct {
		name   string
		length Length
		want   float64
	}{
		{"12pt", Pt(12), 16},        // 12 * 96/72 = 16
		{"72pt", Pt(72), 96},        // 72 * 96/72 = 96 (1 inch)
		{"1pt", Pt(1), 96.0 / 72.0}, // ≈ 1.333
		{"36pt", Pt(36), 48},        // ≈ 0.5 inch
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLength(tt.length, ctx, fontSize)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("ResolveLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestResolveLengthPc tests resolving pica units (1pc = 12pt = 16px)
func TestResolveLengthPc(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)
	fontSize := 16.0

	tests := []struct {
		name   string
		length Length
		want   float64
	}{
		{"1pc", Pc(1), 16},    // 1pc = 16px
		{"6pc", Pc(6), 96},    // 6pc = 96px (1 inch)
		{"0.5pc", Pc(0.5), 8}, // 0.5pc = 8px
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLength(tt.length, ctx, fontSize)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("ResolveLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestResolveLengthIn tests resolving inch units (1in = 96px)
func TestResolveLengthIn(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)
	fontSize := 16.0

	tests := []struct {
		name   string
		length Length
		want   float64
	}{
		{"1in", In(1), 96},       // 1in = 96px (CSS reference pixel)
		{"0.5in", In(0.5), 48},   // 0.5in = 48px
		{"2in", In(2), 192},      // 2in = 192px
		{"0.25in", In(0.25), 24}, // 0.25in = 24px
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLength(tt.length, ctx, fontSize)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("ResolveLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestResolveLengthCm tests resolving centimeter units (1cm = 96/2.54 px)
func TestResolveLengthCm(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)
	fontSize := 16.0

	tests := []struct {
		name       string
		length     Length
		wantApprox float64
		tolerance  float64
	}{
		{"1cm", Cm(1), 37.795, 0.01},     // 1cm ≈ 37.795px
		{"2.54cm", Cm(2.54), 96, 0.01},   // 2.54cm = 1 inch = 96px
		{"10cm", Cm(10), 377.95, 0.01},   // 10cm ≈ 377.95px
		{"0.5cm", Cm(0.5), 18.898, 0.01}, // 0.5cm ≈ 18.898px
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLength(tt.length, ctx, fontSize)
			if math.Abs(got-tt.wantApprox) > tt.tolerance {
				t.Errorf("ResolveLength() = %v, want approximately %v (±%v)", got, tt.wantApprox, tt.tolerance)
			}
		})
	}
}

// TestResolveLengthMm tests resolving millimeter units (1mm = 96/25.4 px)
func TestResolveLengthMm(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)
	fontSize := 16.0

	tests := []struct {
		name       string
		length     Length
		wantApprox float64
		tolerance  float64
	}{
		{"1mm", Mm(1), 3.7795, 0.01},   // 1mm ≈ 3.7795px
		{"10mm", Mm(10), 37.795, 0.01}, // 10mm = 1cm ≈ 37.795px
		{"25.4mm", Mm(25.4), 96, 0.01}, // 25.4mm = 1 inch = 96px
		{"5mm", Mm(5), 18.898, 0.01},   // 5mm ≈ 18.898px
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLength(tt.length, ctx, fontSize)
			if math.Abs(got-tt.wantApprox) > tt.tolerance {
				t.Errorf("ResolveLength() = %v, want approximately %v (±%v)", got, tt.wantApprox, tt.tolerance)
			}
		})
	}
}

// TestResolveLengthQ tests resolving quarter-millimeter units (1Q = 96/101.6 px)
func TestResolveLengthQ(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)
	fontSize := 16.0

	tests := []struct {
		name       string
		length     Length
		wantApprox float64
		tolerance  float64
	}{
		{"1Q", Q(1), 0.945, 0.01},      // 1Q ≈ 0.945px
		{"10Q", Q(10), 9.45, 0.01},     // 10Q ≈ 9.45px
		{"40Q", Q(40), 37.795, 0.01},   // 40Q = 1cm ≈ 37.795px
		{"101.6Q", Q(101.6), 96, 0.01}, // 101.6Q = 1 inch = 96px
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLength(tt.length, ctx, fontSize)
			if math.Abs(got-tt.wantApprox) > tt.tolerance {
				t.Errorf("ResolveLength() = %v, want approximately %v (±%v)", got, tt.wantApprox, tt.tolerance)
			}
		})
	}
}

// TestAbsoluteUnitsRelationships tests the mathematical relationships between absolute units
func TestAbsoluteUnitsRelationships(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)
	fontSize := 16.0

	// Test that 72pt = 1in = 96px
	pt72 := ResolveLength(Pt(72), ctx, fontSize)
	in1 := ResolveLength(In(1), ctx, fontSize)
	if math.Abs(pt72-in1) > 0.01 {
		t.Errorf("72pt should equal 1in: 72pt=%v, 1in=%v", pt72, in1)
	}
	if math.Abs(pt72-96) > 0.01 {
		t.Errorf("72pt should equal 96px: got %v", pt72)
	}

	// Test that 6pc = 1in = 96px
	pc6 := ResolveLength(Pc(6), ctx, fontSize)
	if math.Abs(pc6-in1) > 0.01 {
		t.Errorf("6pc should equal 1in: 6pc=%v, 1in=%v", pc6, in1)
	}

	// Test that 2.54cm = 1in = 96px
	cm254 := ResolveLength(Cm(2.54), ctx, fontSize)
	if math.Abs(cm254-in1) > 0.01 {
		t.Errorf("2.54cm should equal 1in: 2.54cm=%v, 1in=%v", cm254, in1)
	}

	// Test that 25.4mm = 1in = 96px
	mm254 := ResolveLength(Mm(25.4), ctx, fontSize)
	if math.Abs(mm254-in1) > 0.01 {
		t.Errorf("25.4mm should equal 1in: 25.4mm=%v, 1in=%v", mm254, in1)
	}

	// Test that 10mm = 1cm
	mm10 := ResolveLength(Mm(10), ctx, fontSize)
	cm1 := ResolveLength(Cm(1), ctx, fontSize)
	if math.Abs(mm10-cm1) > 0.01 {
		t.Errorf("10mm should equal 1cm: 10mm=%v, 1cm=%v", mm10, cm1)
	}

	// Test that 40Q = 1cm
	q40 := ResolveLength(Q(40), ctx, fontSize)
	if math.Abs(q40-cm1) > 0.01 {
		t.Errorf("40Q should equal 1cm: 40Q=%v, 1cm=%v", q40, cm1)
	}

	// Test that 12pt = 1pc
	pt12 := ResolveLength(Pt(12), ctx, fontSize)
	pc1 := ResolveLength(Pc(1), ctx, fontSize)
	if math.Abs(pt12-pc1) > 0.01 {
		t.Errorf("12pt should equal 1pc: 12pt=%v, 1pc=%v", pt12, pc1)
	}
}

// TestAbsoluteUnitsInLayout tests using absolute units in actual layout
func TestAbsoluteUnitsInLayout(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)

	// Test with points (common for font sizes)
	node1 := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   Pt(72), // 1 inch = 96px
			Height:  Pt(36), // 0.5 inch = 48px
		},
	}
	constraints := Loose(1000, 800)
	size1 := Layout(node1, constraints, ctx)
	if math.Abs(size1.Width-96) > 0.01 {
		t.Errorf("Width with Pt(72): got %.2f, want 96", size1.Width)
	}
	if math.Abs(size1.Height-48) > 0.01 {
		t.Errorf("Height with Pt(36): got %.2f, want 48", size1.Height)
	}

	// Test with centimeters
	node2 := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   Cm(2.54), // 1 inch = 96px
			Height:  Cm(1),    // ≈ 37.795px
		},
	}
	size2 := Layout(node2, constraints, ctx)
	if math.Abs(size2.Width-96) > 0.01 {
		t.Errorf("Width with Cm(2.54): got %.2f, want 96", size2.Width)
	}
	if math.Abs(size2.Height-37.795) > 0.01 {
		t.Errorf("Height with Cm(1): got %.2f, want 37.795", size2.Height)
	}

	// Test with inches
	node3 := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   In(1),          // 96px
			Height:  In(0.5),        // 48px
			Padding: Uniform(Pt(6)), // 6pt = 8px
		},
	}
	size3 := Layout(node3, constraints, ctx)
	// Width/Height include padding (8px each side = 16px total)
	if math.Abs(size3.Width-112) > 0.01 { // 96 + 16
		t.Errorf("Width with In(1) + Pt(6) padding: got %.2f, want 112", size3.Width)
	}
	if math.Abs(size3.Height-64) > 0.01 { // 48 + 16
		t.Errorf("Height with In(0.5) + Pt(6) padding: got %.2f, want 64", size3.Height)
	}
}

// TestAbsoluteUnitString tests string representation
func TestAbsoluteUnitString(t *testing.T) {
	tests := []struct {
		length Length
		want   string
	}{
		{Pt(12), "12.00pt"},
		{Pc(1), "1.00pc"},
		{In(1), "1.00in"},
		{Cm(2.54), "2.54cm"},
		{Mm(25.4), "25.40mm"},
		{Q(40), "40.00Q"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.length.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
