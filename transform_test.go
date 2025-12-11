package layout

import (
	"math"
	"testing"
)

func TestTransformIdentity(t *testing.T) {
	transform := IdentityTransform()
	if !transform.IsIdentity() {
		t.Error("Identity transform should be identity")
	}
	
	point := Point{X: 10, Y: 20}
	result := transform.Apply(point)
	if result.X != point.X || result.Y != point.Y {
		t.Errorf("Identity transform should not change point")
	}
}

func TestTransformTranslate(t *testing.T) {
	transform := Translate(10, 20)
	point := Point{X: 5, Y: 5}
	result := transform.Apply(point)
	
	expectedX := 15.0
	expectedY := 25.0
	
	if math.Abs(result.X-expectedX) > 0.001 {
		t.Errorf("Expected X=%.3f, got %.3f", expectedX, result.X)
	}
	if math.Abs(result.Y-expectedY) > 0.001 {
		t.Errorf("Expected Y=%.3f, got %.3f", expectedY, result.Y)
	}
}

func TestTransformScale(t *testing.T) {
	transform := Scale(2, 3)
	point := Point{X: 5, Y: 5}
	result := transform.Apply(point)
	
	expectedX := 10.0
	expectedY := 15.0
	
	if math.Abs(result.X-expectedX) > 0.001 {
		t.Errorf("Expected X=%.3f, got %.3f", expectedX, result.X)
	}
	if math.Abs(result.Y-expectedY) > 0.001 {
		t.Errorf("Expected Y=%.3f, got %.3f", expectedY, result.Y)
	}
}

func TestTransformRotate(t *testing.T) {
	// Rotate 90 degrees (π/2 radians)
	transform := Rotate(math.Pi / 2)
	point := Point{X: 1, Y: 0}
	result := transform.Apply(point)
	
	// 90 degree rotation: (1, 0) -> (0, 1)
	expectedX := 0.0
	expectedY := 1.0
	
	if math.Abs(result.X-expectedX) > 0.001 {
		t.Errorf("Expected X=%.3f, got %.3f", expectedX, result.X)
	}
	if math.Abs(result.Y-expectedY) > 0.001 {
		t.Errorf("Expected Y=%.3f, got %.3f", expectedY, result.Y)
	}
}

func TestTransformRotateDegrees(t *testing.T) {
	// Rotate 180 degrees
	transform := RotateDegrees(180)
	point := Point{X: 1, Y: 0}
	result := transform.Apply(point)
	
	// 180 degree rotation: (1, 0) -> (-1, 0)
	expectedX := -1.0
	expectedY := 0.0
	
	if math.Abs(result.X-expectedX) > 0.001 {
		t.Errorf("Expected X=%.3f, got %.3f", expectedX, result.X)
	}
	if math.Abs(result.Y-expectedY) > 0.001 {
		t.Errorf("Expected Y=%.3f, got %.3f", expectedY, result.Y)
	}
}

func TestTransformMultiply(t *testing.T) {
	// Translate then scale
	t1 := Translate(10, 20)
	t2 := Scale(2, 2)
	combined := t1.Multiply(t2)
	
	point := Point{X: 5, Y: 5}
	result := combined.Apply(point)
	
	// First scale: (5, 5) -> (10, 10)
	// Then translate: (10, 10) -> (20, 30)
	expectedX := 20.0
	expectedY := 30.0
	
	if math.Abs(result.X-expectedX) > 0.001 {
		t.Errorf("Expected X=%.3f, got %.3f", expectedX, result.X)
	}
	if math.Abs(result.Y-expectedY) > 0.001 {
		t.Errorf("Expected Y=%.3f, got %.3f", expectedY, result.Y)
	}
}

func TestTransformApplyToRect(t *testing.T) {
	// Translate a rectangle
	transform := Translate(10, 20)
	rect := Rect{X: 5, Y: 5, Width: 10, Height: 10}
	result := transform.ApplyToRect(rect)
	
	expectedX := 15.0
	expectedY := 25.0
	
	if math.Abs(result.X-expectedX) > 0.001 {
		t.Errorf("Expected X=%.3f, got %.3f", expectedX, result.X)
	}
	if math.Abs(result.Y-expectedY) > 0.001 {
		t.Errorf("Expected Y=%.3f, got %.3f", expectedY, result.Y)
	}
	// Size should remain the same for translation
	if math.Abs(result.Width-rect.Width) > 0.001 {
		t.Errorf("Width should remain %.3f, got %.3f", rect.Width, result.Width)
	}
	if math.Abs(result.Height-rect.Height) > 0.001 {
		t.Errorf("Height should remain %.3f, got %.3f", rect.Height, result.Height)
	}
}

func TestTransformRotateRect(t *testing.T) {
	// Rotate a rectangle 90 degrees
	transform := RotateDegrees(90)
	rect := Rect{X: 0, Y: 0, Width: 10, Height: 20}
	result := transform.ApplyToRect(rect)
	
	// After 90 degree rotation, width and height swap
	// The bounding box should be 20x10
	if math.Abs(result.Width-20.0) > 0.001 {
		t.Errorf("Expected width 20, got %.3f", result.Width)
	}
	if math.Abs(result.Height-10.0) > 0.001 {
		t.Errorf("Expected height 10, got %.3f", result.Height)
	}
}

func TestTransformToSVGString(t *testing.T) {
	transform := Translate(10, 20)
	svgStr := transform.ToSVGString()
	
	expected := "matrix(1,0,0,1,10,20)"
	if svgStr != expected {
		t.Errorf("Expected SVG string %q, got %q", expected, svgStr)
	}
	
	// Identity should return empty string
	identity := IdentityTransform()
	if identity.ToSVGString() != "" {
		t.Error("Identity transform should return empty SVG string")
	}
}

