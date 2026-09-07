package layout

import "testing"

// countingMetrics records how many times Measure is called so tests can
// verify that ResolveLength only measures the ch reference glyph when the
// unit needs it.
type countingMetrics struct {
	calls int
}

func (m *countingMetrics) Measure(text string, style TextStyle) (advance, ascent, descent float64) {
	m.calls++
	return float64(len(text)) * style.FontSize * 0.5, style.FontSize * 0.8, style.FontSize * 0.2
}

// TestResolveLengthUnsetFastPath: the zero-value Length resolves to 0 with
// and without a context, exactly as it did through the units error path.
func TestResolveLengthUnsetFastPath(t *testing.T) {
	if got := ResolveLength(Length{}, nil, 16); got != 0 {
		t.Errorf("ResolveLength(Length{}, nil) = %v, want 0", got)
	}
	ctx := NewLayoutContext(800, 600, 16)
	if got := ResolveLength(Length{}, ctx, 16); got != 0 {
		t.Errorf("ResolveLength(Length{}, ctx) = %v, want 0", got)
	}
	// A hand-built unit-less literal keeps the pre-fast-path meaning (the
	// units package resolves an empty unit as pixels).
	if got := ResolveLength(Length{Value: 12}, ctx, 16); got != 12 {
		t.Errorf("ResolveLength(Length{Value: 12}, ctx) = %v, want 12", got)
	}
	// Real units are unaffected.
	if got := ResolveLength(Px(0), ctx, 16); got != 0 {
		t.Errorf("ResolveLength(Px(0)) = %v, want 0", got)
	}
	if got := ResolveLength(Em(2), ctx, 10); got != 20 {
		t.Errorf("ResolveLength(Em(2)) = %v, want 20", got)
	}
}

// TestResolveLengthMeasuresChOnlyWhenNeeded: the text metrics provider is
// consulted for ch/ic (CSS Values L4 §6.1.1) and for nothing else.
func TestResolveLengthMeasuresChOnlyWhenNeeded(t *testing.T) {
	metrics := &countingMetrics{}
	ctx := NewLayoutContext(800, 600, 16).WithTextMetrics(metrics)

	for _, l := range []Length{{}, Px(10), Em(1), Rem(1), Vw(10), Vh(10), Cqw(10)} {
		ResolveLength(l, ctx, 16)
	}
	if metrics.calls != 0 {
		t.Fatalf("expected no Measure calls for non-ch units, got %d", metrics.calls)
	}

	// 1ch = advance of "0" = 16 * 0.5 = 8px with countingMetrics.
	if got := ResolveLength(Ch(2), ctx, 16); got != 16 {
		t.Errorf("ResolveLength(Ch(2)) = %v, want 16", got)
	}
	if metrics.calls != 1 {
		t.Errorf("expected exactly one Measure call for ch, got %d", metrics.calls)
	}
	// ic is modeled as two ch advances.
	if got := ResolveLength(Length{Value: 1, Unit: "ic"}, ctx, 16); got != 16 {
		t.Errorf("ResolveLength(1ic) = %v, want 16", got)
	}
	if metrics.calls != 2 {
		t.Errorf("expected a Measure call for ic, got %d total", metrics.calls)
	}
}

// BenchmarkResolveLengthUnset measures the hot path taken for every unset
// margin/padding/border/offset field during layout.
func BenchmarkResolveLengthUnset(b *testing.B) {
	ctx := NewLayoutContext(800, 600, 16)
	var sink float64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink += ResolveLength(Length{}, ctx, 16)
	}
	_ = sink
}

// BenchmarkResolveLengthPx measures a set absolute length for comparison.
func BenchmarkResolveLengthPx(b *testing.B) {
	ctx := NewLayoutContext(800, 600, 16)
	l := Px(10)
	var sink float64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink += ResolveLength(l, ctx, 16)
	}
	_ = sink
}
