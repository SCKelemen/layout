package layout

import (
	"math"
	"testing"
)

// Regression tests for the zero-value Transform being treated as a real
// (degenerate) transform, and for the Multiply composition order.

func TestZeroTransformIsIdentity(t *testing.T) {
	var zero Transform
	if !zero.IsIdentity() {
		t.Fatalf("Transform{} must be treated as identity")
	}
	if !IdentityTransform().IsIdentity() {
		t.Fatalf("IdentityTransform() must be identity")
	}
	if Translate(1, 0).IsIdentity() {
		t.Fatalf("Translate(1,0) must not be identity")
	}
	// A transform that shares zeros with Transform{} but is not all-zero is
	// still a real (degenerate) transform.
	if (Transform{A: 0, D: 1}).IsIdentity() {
		t.Fatalf("Transform{D:1} must not be identity")
	}
}

func TestZeroTransformToSVGString(t *testing.T) {
	if got := (Transform{}).ToSVGString(); got != "" {
		t.Fatalf("Transform{}.ToSVGString() = %q, want empty", got)
	}
	if got := GetSVGTransform(&Node{}); got != "" {
		t.Fatalf("GetSVGTransform(&Node{}) = %q, want empty", got)
	}
	if got := GetSVGTransform(&Node{Style: Style{Transform: Translate(3, 4)}}); got != "matrix(1,0,0,1,3,4)" {
		t.Fatalf("GetSVGTransform(translate) = %q", got)
	}
}

func TestZeroTransformApplyToRect(t *testing.T) {
	rect := Rect{X: 10, Y: 20, Width: 100, Height: 50}
	if got := (Transform{}).ApplyToRect(rect); got != rect {
		t.Fatalf("Transform{}.ApplyToRect = %+v, want %+v", got, rect)
	}
	if got := IdentityTransform().ApplyToRect(rect); got != rect {
		t.Fatalf("identity ApplyToRect = %+v, want %+v", got, rect)
	}
	if got := GetFinalRect(&Node{Rect: rect}); got != rect {
		t.Fatalf("GetFinalRect(untransformed node) = %+v, want %+v", got, rect)
	}
	// A real transform still moves the rect.
	moved := GetFinalRect(&Node{Rect: rect, Style: Style{Transform: Translate(5, 5)}})
	if moved.X != 15 || moved.Y != 25 || moved.Width != 100 || moved.Height != 50 {
		t.Fatalf("GetFinalRect(translated) = %+v", moved)
	}
}

func TestZeroTransformMultiplyIsIdentityOperand(t *testing.T) {
	rot := RotateDegrees(30)
	if got := (Transform{}).Multiply(rot); got != rot {
		t.Fatalf("Transform{}.Multiply(t) = %+v, want %+v", got, rot)
	}
	if got := rot.Multiply(Transform{}); got != rot {
		t.Fatalf("t.Multiply(Transform{}) = %+v, want %+v", got, rot)
	}
	if got := (Transform{}).Multiply(Transform{}); got != (Transform{}) {
		t.Fatalf("Transform{}.Multiply(Transform{}) = %+v, want zero", got)
	}
}

// TestTransformMultiplyOrderMatchesDoc pins the documented composition order:
// t1.Multiply(t2) applies t2 first, then t1.
func TestTransformMultiplyOrderMatchesDoc(t *testing.T) {
	t1 := Translate(10, 0)
	t2 := Scale(2, 2)
	p := Point{X: 1, Y: 1}

	composed := t1.Multiply(t2).Apply(p)
	sequential := t1.Apply(t2.Apply(p)) // t2 first, then t1

	if math.Abs(composed.X-sequential.X) > 1e-9 || math.Abs(composed.Y-sequential.Y) > 1e-9 {
		t.Fatalf("t1.Multiply(t2).Apply(p) = %+v, want t1(t2(p)) = %+v", composed, sequential)
	}
	// Concretely: scale (1,1) -> (2,2), then translate -> (12,2).
	if math.Abs(composed.X-12) > 1e-9 || math.Abs(composed.Y-2) > 1e-9 {
		t.Fatalf("expected (12,2), got %+v", composed)
	}
}
