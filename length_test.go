package layout

import (
	"math"
	"testing"
)

// TestLengthConstructors tests the helper constructors for Length values
func TestLengthConstructors(t *testing.T) {
	tests := []struct {
		name     string
		length   Length
		wantVal  float64
		wantUnit LengthUnit
	}{
		{"Px", Px(100), 100, Pixels},
		{"Em", Em(2), 2, EmUnit},
		{"Rem", Rem(1.5), 1.5, RemUnit},
		{"Ch", Ch(80), 80, ChUnit},
		{"Vh", Vh(50), 50, VhUnit},
		{"Vw", Vw(100), 100, VwUnit},
		{"Vmax", Vmax(75), 75, VmaxUnit},
		{"Vmin", Vmin(25), 25, VminUnit},
		{"Zero pixels", Px(0), 0, Pixels},
		{"Negative pixels", Px(-10), -10, Pixels},
		{"UnboundedLength", UnboundedLength(), 0, UnboundedUnit},
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

// TestResolveLengthPixels tests resolving pixel units (should pass through unchanged)
func TestResolveLengthPixels(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)
	fontSize := 16.0

	tests := []struct {
		name   string
		length Length
		want   float64
	}{
		{"100px", Px(100), 100},
		{"0px", Px(0), 0},
		{"negative px", Px(-10), -10},
		{"fractional px", Px(10.5), 10.5},
		{"unbounded", Px(math.MaxFloat64), math.MaxFloat64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLength(tt.length, ctx, fontSize)
			if got != tt.want {
				t.Errorf("ResolveLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestResolveLengthEm tests resolving em units (relative to element font size)
func TestResolveLengthEm(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)

	tests := []struct {
		name         string
		length       Length
		fontSize     float64
		wantResolved float64
	}{
		{"2em with 16pt font", Em(2), 16, 32},
		{"1em with 12pt font", Em(1), 12, 12},
		{"0.5em with 20pt font", Em(0.5), 20, 10},
		{"10em with 8pt font", Em(10), 8, 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLength(tt.length, ctx, tt.fontSize)
			if got != tt.wantResolved {
				t.Errorf("ResolveLength() = %v, want %v", got, tt.wantResolved)
			}
		})
	}
}

// TestResolveLengthRem tests resolving rem units (relative to root font size)
func TestResolveLengthRem(t *testing.T) {
	tests := []struct {
		name         string
		length       Length
		rootFontSize float64
		wantResolved float64
	}{
		{"2rem with 16pt root", Rem(2), 16, 32},
		{"1rem with 12pt root", Rem(1), 12, 12},
		{"0.5rem with 20pt root", Rem(0.5), 20, 10},
		{"10rem with 8pt root", Rem(10), 8, 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewLayoutContext(1920, 1080, tt.rootFontSize)
			// Rem should use root font size regardless of current font size
			currentFontSize := 999.0 // Should be ignored
			got := ResolveLength(tt.length, ctx, currentFontSize)
			if got != tt.wantResolved {
				t.Errorf("ResolveLength() = %v, want %v", got, tt.wantResolved)
			}
		})
	}
}

// TestResolveLengthVh tests resolving vh units (viewport height percentage)
func TestResolveLengthVh(t *testing.T) {
	tests := []struct {
		name           string
		length         Length
		viewportHeight float64
		wantResolved   float64
	}{
		{"50vh of 1080px", Vh(50), 1080, 540},
		{"100vh of 1080px", Vh(100), 1080, 1080},
		{"10vh of 1080px", Vh(10), 1080, 108},
		{"1vh of 1000px", Vh(1), 1000, 10},
		{"0vh", Vh(0), 1080, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewLayoutContext(1920, tt.viewportHeight, 16)
			fontSize := 16.0
			got := ResolveLength(tt.length, ctx, fontSize)
			if got != tt.wantResolved {
				t.Errorf("ResolveLength() = %v, want %v", got, tt.wantResolved)
			}
		})
	}
}

// TestResolveLengthVw tests resolving vw units (viewport width percentage)
func TestResolveLengthVw(t *testing.T) {
	tests := []struct {
		name          string
		length        Length
		viewportWidth float64
		wantResolved  float64
	}{
		{"50vw of 1920px", Vw(50), 1920, 960},
		{"100vw of 1920px", Vw(100), 1920, 1920},
		{"10vw of 1920px", Vw(10), 1920, 192},
		{"1vw of 1000px", Vw(1), 1000, 10},
		{"0vw", Vw(0), 1920, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewLayoutContext(tt.viewportWidth, 1080, 16)
			fontSize := 16.0
			got := ResolveLength(tt.length, ctx, fontSize)
			if got != tt.wantResolved {
				t.Errorf("ResolveLength() = %v, want %v", got, tt.wantResolved)
			}
		})
	}
}

// TestResolveLengthVmax tests resolving vmax units (larger viewport dimension)
func TestResolveLengthVmax(t *testing.T) {
	tests := []struct {
		name           string
		length         Length
		viewportWidth  float64
		viewportHeight float64
		wantResolved   float64
	}{
		{"50vmax when width > height", Vmax(50), 1920, 1080, 960}, // 50% of 1920
		{"50vmax when height > width", Vmax(50), 1080, 1920, 960}, // 50% of 1920
		{"100vmax when equal", Vmax(100), 1000, 1000, 1000},       // 100% of 1000
		{"10vmax of 1920x1080", Vmax(10), 1920, 1080, 192},        // 10% of 1920
		{"1vmax of 1000x800", Vmax(1), 1000, 800, 10},             // 1% of 1000
		{"0vmax", Vmax(0), 1920, 1080, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewLayoutContext(tt.viewportWidth, tt.viewportHeight, 16)
			fontSize := 16.0
			got := ResolveLength(tt.length, ctx, fontSize)
			if got != tt.wantResolved {
				t.Errorf("ResolveLength() = %v, want %v", got, tt.wantResolved)
			}
		})
	}
}

// TestResolveLengthVmin tests resolving vmin units (smaller viewport dimension)
func TestResolveLengthVmin(t *testing.T) {
	tests := []struct {
		name           string
		length         Length
		viewportWidth  float64
		viewportHeight float64
		wantResolved   float64
	}{
		{"50vmin when width > height", Vmin(50), 1920, 1080, 540}, // 50% of 1080
		{"50vmin when height > width", Vmin(50), 1080, 1920, 540}, // 50% of 1080
		{"100vmin when equal", Vmin(100), 1000, 1000, 1000},       // 100% of 1000
		{"10vmin of 1920x1080", Vmin(10), 1920, 1080, 108},        // 10% of 1080
		{"1vmin of 1000x800", Vmin(1), 1000, 800, 8},              // 1% of 800
		{"0vmin", Vmin(0), 1920, 1080, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewLayoutContext(tt.viewportWidth, tt.viewportHeight, 16)
			fontSize := 16.0
			got := ResolveLength(tt.length, ctx, fontSize)
			if got != tt.wantResolved {
				t.Errorf("ResolveLength() = %v, want %v", got, tt.wantResolved)
			}
		})
	}
}

// TestResolveLengthCh tests resolving ch units (character width)
func TestResolveLengthCh(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)

	tests := []struct {
		name       string
		length     Length
		fontSize   float64
		wantApprox float64 // Approximate expected value (monospace approximation)
		tolerance  float64
	}{
		{"80ch with 16pt font", Ch(80), 16, 80 * 16 * 0.6, 1},
		{"1ch with 12pt font", Ch(1), 12, 12 * 0.6, 1},
		{"40ch with 20pt font", Ch(40), 20, 40 * 20 * 0.6, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLength(tt.length, ctx, tt.fontSize)
			// Check if within tolerance (ch uses text measurement which may vary)
			if math.Abs(got-tt.wantApprox) > tt.tolerance {
				t.Errorf("ResolveLength() = %v, want approximately %v (±%v)", got, tt.wantApprox, tt.tolerance)
			}
		})
	}
}

// TestResolveLengthChWithCustomChar tests ch unit with custom reference character
func TestResolveLengthChWithCustomChar(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)
	ctx = ctx.WithChReferenceChar('M') // Use 'M' instead of '0'
	fontSize := 16.0

	got := ResolveLength(Ch(10), ctx, fontSize)

	// Should use 'M' width instead of '0' width
	// We can't assert exact value without knowing text metrics,
	// but should be > 0 and reasonable
	if got <= 0 {
		t.Errorf("ResolveLength(Ch(10)) with custom char = %v, want > 0", got)
	}
	if got > 1000 {
		t.Errorf("ResolveLength(Ch(10)) with custom char = %v, seems too large", got)
	}
}

// TestLayoutContextDefaults tests NewLayoutContext creates sensible defaults
func TestLayoutContextDefaults(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)

	if ctx.ViewportWidth != 1920 {
		t.Errorf("ViewportWidth = %v, want 1920", ctx.ViewportWidth)
	}
	if ctx.ViewportHeight != 1080 {
		t.Errorf("ViewportHeight = %v, want 1080", ctx.ViewportHeight)
	}
	if ctx.RootFontSize != 16 {
		t.Errorf("RootFontSize = %v, want 16", ctx.RootFontSize)
	}
	if ctx.TextMetrics == nil {
		t.Error("TextMetrics should not be nil (should use default)")
	}
	if ctx.ChReferenceChar != '0' {
		t.Errorf("ChReferenceChar = %q, want '0'", ctx.ChReferenceChar)
	}
}

// TestLengthString tests the String() method
func TestLengthString(t *testing.T) {
	tests := []struct {
		length Length
		want   string
	}{
		{Px(100), "100.00px"},
		{Em(2), "2.00em"},
		{Rem(1.5), "1.50rem"},
		{Ch(80), "80.00ch"},
		{Vh(50), "50.00vh"},
		{Vw(100), "100.00vw"},
		{Vmax(75), "75.00vmax"},
		{Vmin(25), "25.00vmin"},
		{UnboundedLength(), "0.00unbounded"},
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

// TestUnboundedUnit tests the UnboundedUnit resolves to math.MaxFloat64
func TestUnboundedUnit(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)
	fontSize := 16.0

	tests := []struct {
		name   string
		length Length
		want   float64
	}{
		{"UnboundedLength", UnboundedLength(), math.MaxFloat64},
		{"PxUnbounded", PxUnbounded, math.MaxFloat64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveLength(tt.length, ctx, fontSize)
			if got != tt.want {
				t.Errorf("ResolveLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPxUnboundedEquivalence tests that PxUnbounded and Px(math.MaxFloat64) are equivalent
func TestPxUnboundedEquivalence(t *testing.T) {
	ctx := NewLayoutContext(1920, 1080, 16)
	fontSize := 16.0

	pxMax := Px(math.MaxFloat64)

	gotPxMax := ResolveLength(pxMax, ctx, fontSize)
	gotPxUnbounded := ResolveLength(PxUnbounded, ctx, fontSize)

	if gotPxMax != gotPxUnbounded {
		t.Errorf("Px(math.MaxFloat64) = %v, PxUnbounded = %v, want equal", gotPxMax, gotPxUnbounded)
	}
	if gotPxMax != math.MaxFloat64 {
		t.Errorf("Px(math.MaxFloat64) = %v, want %v", gotPxMax, math.MaxFloat64)
	}
}

// TestMixedUnitsInLayout tests that different units can be used together
func TestMixedUnitsInLayout(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)
	fontSize := 16.0

	// Create a spacing with mixed units
	spacing := Spacing{
		Top:    Vh(10),   // 10% of 800 = 80
		Right:  Em(2),    // 2 * 16 = 32
		Bottom: Px(20),   // 20
		Left:   Rem(1.5), // 1.5 * 16 = 24
	}

	top := ResolveLength(spacing.Top, ctx, fontSize)
	right := ResolveLength(spacing.Right, ctx, fontSize)
	bottom := ResolveLength(spacing.Bottom, ctx, fontSize)
	left := ResolveLength(spacing.Left, ctx, fontSize)

	if top != 80 {
		t.Errorf("Top (Vh(10)) = %v, want 80", top)
	}
	if right != 32 {
		t.Errorf("Right (Em(2)) = %v, want 32", right)
	}
	if bottom != 20 {
		t.Errorf("Bottom (Px(20)) = %v, want 20", bottom)
	}
	if left != 24 {
		t.Errorf("Left (Rem(1.5)) = %v, want 24", left)
	}
}
