package layout

import (
	"math"
	"testing"
)

func TestLayoutFlexbox(t *testing.T) {
	// Test main Layout function routes to flexbox
	root := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionRow,
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
		},
	}

	constraints := Loose(200, 100)
	size := Layout(root, constraints)

	if size.Width < 100 {
		t.Errorf("Layout should return correct size, got width %.2f", size.Width)
	}
}

func TestLayoutGrid(t *testing.T) {
	// Test main Layout function routes to grid
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(100),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(100),
			},
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridColumnStart: 0}},
		},
	}

	constraints := Loose(200, 200)
	size := Layout(root, constraints)

	if size.Width < 100 {
		t.Errorf("Layout should return correct size, got width %.2f", size.Width)
	}
}

func TestLayoutBlock(t *testing.T) {
	// Test main Layout function routes to block (default)
	root := &Node{
		Style: Style{
			Width:  100,
			Height: 100,
		},
		Children: []*Node{},
	}

	constraints := Loose(200, 200)
	size := Layout(root, constraints)

	if math.Abs(size.Width-100.0) > 1.0 {
		t.Errorf("Layout should return correct size, got width %.2f", size.Width)
	}
}

func TestLayoutNone(t *testing.T) {
	// Test display: none
	root := &Node{
		Style: Style{
			Display: DisplayNone,
			Width:   100,
			Height:  100,
		},
		Children: []*Node{},
	}

	constraints := Loose(200, 200)
	size := Layout(root, constraints)

	if size.Width != 0 || size.Height != 0 {
		t.Errorf("Display none should return zero size, got %.2f x %.2f", size.Width, size.Height)
	}
}

func TestConstraintsTight(t *testing.T) {
	// Test tight constraints
	constraints := Tight(200, 100)

	if constraints.MinWidth != 200 || constraints.MaxWidth != 200 {
		t.Errorf("Tight constraints should have min == max")
	}
	if constraints.MinHeight != 100 || constraints.MaxHeight != 100 {
		t.Errorf("Tight constraints should have min == max")
	}
}

func TestConstraintsLoose(t *testing.T) {
	// Test loose constraints
	constraints := Loose(200, 100)

	if constraints.MinWidth != 0 || constraints.MaxWidth != 200 {
		t.Errorf("Loose constraints should have min=0, max=specified")
	}
	if constraints.MinHeight != 0 || constraints.MaxHeight != 100 {
		t.Errorf("Loose constraints should have min=0, max=specified")
	}
}

func TestConstraintsConstrain(t *testing.T) {
	// Test Constrain method
	constraints := Loose(200, 100)

	size := Size{Width: 300, Height: 50}
	constrained := constraints.Constrain(size)

	if constrained.Width != 200 {
		t.Errorf("Width should be constrained to 200, got %.2f", constrained.Width)
	}
	if constrained.Height != 50 {
		t.Errorf("Height should not be constrained (50 < 100), got %.2f", constrained.Height)
	}
}
