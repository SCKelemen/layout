package layout

import "testing"

// TestAdapterNonPositiveLineHeightMeansNormal verifies that the text metrics
// adapter treats a zero or negative LineHeight as "normal" (1.2 x font size)
// instead of using it as a multiplier and producing negative ascent/descent.
// https://www.w3.org/TR/css-inline-3/#propdef-line-height
func TestAdapterNonPositiveLineHeightMeansNormal(t *testing.T) {
	adapter := NewTerminalTextMetrics()

	_, normalAscent, normalDescent := adapter.Measure("x", TextStyle{FontSize: 16, LineHeight: 0})
	_, negAscent, negDescent := adapter.Measure("x", TextStyle{FontSize: 16, LineHeight: -5})

	if negAscent <= 0 || negDescent <= 0 {
		t.Fatalf("negative LineHeight produced non-positive metrics: ascent=%.2f descent=%.2f", negAscent, negDescent)
	}
	if negAscent != normalAscent || negDescent != normalDescent {
		t.Errorf("negative LineHeight should match normal: got ascent=%.2f descent=%.2f, want %.2f/%.2f",
			negAscent, negDescent, normalAscent, normalDescent)
	}
	// normal = 16 * 1.2 = 19.2; ascent 80%, descent 20%
	if want := 19.2 * 0.8; normalAscent != want {
		t.Errorf("normal ascent: expected %.2f, got %.2f", want, normalAscent)
	}
}
