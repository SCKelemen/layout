package layout

import (
	"math"
	"testing"
)

func TestTransformSkewX(t *testing.T) {
	// Test horizontal skew
	transform := SkewX(math.Pi / 4) // 45 degrees
	point := Point{X: 1, Y: 1}
	result := transform.Apply(point)
	
	// SkewX skews horizontally: (1, 1) -> (1 + tan(45°)*1, 1) = (2, 1)
	// The Y coordinate stays the same, X is skewed
	expectedX := 2.0 // 1 + tan(45°)*1 = 1 + 1 = 2
	expectedY := 1.0
	if math.Abs(result.X-expectedX) > 0.001 {
		t.Errorf("Expected X=%.3f, got %.3f", expectedX, result.X)
	}
	if math.Abs(result.Y-expectedY) > 0.001 {
		t.Errorf("Expected Y=%.3f, got %.3f", expectedY, result.Y)
	}
}

func TestTransformSkewY(t *testing.T) {
	// Test vertical skew
	transform := SkewY(math.Pi / 4) // 45 degrees
	point := Point{X: 1, Y: 1}
	result := transform.Apply(point)
	
	// SkewY skews vertically: (1, 1) -> (1, 1 + tan(45°)*1) = (1, 2)
	// The X coordinate stays the same, Y is skewed
	expectedX := 1.0
	expectedY := 2.0 // 1 + tan(45°)*1 = 1 + 1 = 2
	if math.Abs(result.X-expectedX) > 0.001 {
		t.Errorf("Expected X=%.3f, got %.3f", expectedX, result.X)
	}
	if math.Abs(result.Y-expectedY) > 0.001 {
		t.Errorf("Expected Y=%.3f, got %.3f", expectedY, result.Y)
	}
}

func TestTransformMatrix(t *testing.T) {
	// Test custom matrix
	transform := Matrix(2, 0, 0, 3, 10, 20)
	point := Point{X: 5, Y: 5}
	result := transform.Apply(point)
	
	// Scale X by 2, Y by 3, then translate by (10, 20)
	// (5, 5) -> (10, 15) -> (20, 35)
	expectedX := 20.0
	expectedY := 35.0
	
	if math.Abs(result.X-expectedX) > 0.001 {
		t.Errorf("Expected X=%.3f, got %.3f", expectedX, result.X)
	}
	if math.Abs(result.Y-expectedY) > 0.001 {
		t.Errorf("Expected Y=%.3f, got %.3f", expectedY, result.Y)
	}
}

