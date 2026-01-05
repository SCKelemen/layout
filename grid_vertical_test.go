package layout

import (
	"testing"
)

// TestGridVerticalLR verifies that grid layout works in vertical-lr writing mode
func TestGridVerticalLR(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a grid container with vertical-lr writing mode
	// 2 rows (which in vertical-lr control horizontal positioning)
	// 2 columns (which in vertical-lr control vertical positioning)
	root := &Node{
		Style: Style{
			Display:            DisplayGrid,
			GridTemplateRows:   []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			Width:              Px(200),
			Height:             Px(200),
			WritingMode:        WritingModeVerticalLR,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:        DisplayBlock,
					GridRowStart:   0,
					GridRowEnd:     1,
					GridColumnStart: 0,
					GridColumnEnd:   1,
				},
			},
			{
				Style: Style{
					Display:        DisplayBlock,
					GridRowStart:   0,
					GridRowEnd:     1,
					GridColumnStart: 1,
					GridColumnEnd:   2,
				},
			},
			{
				Style: Style{
					Display:        DisplayBlock,
					GridRowStart:   1,
					GridRowEnd:     2,
					GridColumnStart: 0,
					GridColumnEnd:   1,
				},
			},
			{
				Style: Style{
					Display:        DisplayBlock,
					GridRowStart:   1,
					GridRowEnd:     2,
					GridColumnStart: 1,
					GridColumnEnd:   2,
				},
			},
		},
	}

	constraints := Tight(200, 200)
	LayoutGrid(root, constraints, ctx)

	// In vertical-lr mode:
	// - Rows control horizontal positioning (X axis)
	// - Columns control vertical positioning (Y axis)
	//
	// Grid should be:
	// Row 0 (X=0-100):   Col 0 (Y=0-50): child0   Col 1 (Y=50-100): child1
	// Row 1 (X=100-200): Col 0 (Y=0-50): child2   Col 1 (Y=50-100): child3

	child0 := root.Children[0] // Row 0, Col 0
	child1 := root.Children[1] // Row 0, Col 1
	child2 := root.Children[2] // Row 1, Col 0
	child3 := root.Children[3] // Row 1, Col 1

	// Check child0 (row 0, col 0): X should be 0, Y should be 0
	if child0.Rect.X != 0 {
		t.Errorf("Child 0 X: expected 0, got %.2f", child0.Rect.X)
	}
	if child0.Rect.Y != 0 {
		t.Errorf("Child 0 Y: expected 0, got %.2f", child0.Rect.Y)
	}

	// Check child1 (row 0, col 1): X should be 0, Y should be 50
	if child1.Rect.X != 0 {
		t.Errorf("Child 1 X: expected 0, got %.2f", child1.Rect.X)
	}
	if child1.Rect.Y != 50 {
		t.Errorf("Child 1 Y: expected 50, got %.2f", child1.Rect.Y)
	}

	// Check child2 (row 1, col 0): X should be 100, Y should be 0
	if child2.Rect.X != 100 {
		t.Errorf("Child 2 X: expected 100, got %.2f", child2.Rect.X)
	}
	if child2.Rect.Y != 0 {
		t.Errorf("Child 2 Y: expected 0, got %.2f", child2.Rect.Y)
	}

	// Check child3 (row 1, col 1): X should be 100, Y should be 50
	if child3.Rect.X != 100 {
		t.Errorf("Child 3 X: expected 100, got %.2f", child3.Rect.X)
	}
	if child3.Rect.Y != 50 {
		t.Errorf("Child 3 Y: expected 50, got %.2f", child3.Rect.Y)
	}

	// Check sizes (should all be 100x50 in vertical mode, which is row x column size)
	expectedWidth := 100.0  // Row size
	expectedHeight := 50.0  // Column size
	if child0.Rect.Width != expectedWidth || child0.Rect.Height != expectedHeight {
		t.Errorf("Child 0 size: expected %.2fx%.2f, got %.2fx%.2f",
			expectedWidth, expectedHeight, child0.Rect.Width, child0.Rect.Height)
	}
}

// TestGridHorizontalTB verifies that horizontal mode still works (baseline behavior)
func TestGridHorizontalTB(t *testing.T) {
	ctx := NewLayoutContext(800, 600, 16)

	// Create a grid container with horizontal-tb writing mode (default)
	// 2 rows (control vertical positioning)
	// 2 columns (control horizontal positioning)
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			Width:               Px(200),
			Height:              Px(200),
			WritingMode:         WritingModeHorizontalTB,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:        DisplayBlock,
					GridRowStart:   0,
					GridRowEnd:     1,
					GridColumnStart: 0,
					GridColumnEnd:   1,
				},
			},
			{
				Style: Style{
					Display:        DisplayBlock,
					GridRowStart:   0,
					GridRowEnd:     1,
					GridColumnStart: 1,
					GridColumnEnd:   2,
				},
			},
			{
				Style: Style{
					Display:        DisplayBlock,
					GridRowStart:   1,
					GridRowEnd:     2,
					GridColumnStart: 0,
					GridColumnEnd:   1,
				},
			},
			{
				Style: Style{
					Display:        DisplayBlock,
					GridRowStart:   1,
					GridRowEnd:     2,
					GridColumnStart: 1,
					GridColumnEnd:   2,
				},
			},
		},
	}

	constraints := Tight(200, 200)
	LayoutGrid(root, constraints, ctx)

	// In horizontal-tb mode:
	// - Rows control vertical positioning (Y axis)
	// - Columns control horizontal positioning (X axis)
	//
	// Grid should be:
	// Row 0 (Y=0-50):  Col 0 (X=0-100): child0   Col 1 (X=100-200): child1
	// Row 1 (Y=50-100): Col 0 (X=0-100): child2   Col 1 (X=100-200): child3

	child0 := root.Children[0] // Row 0, Col 0
	child1 := root.Children[1] // Row 0, Col 1
	child2 := root.Children[2] // Row 1, Col 0
	child3 := root.Children[3] // Row 1, Col 1

	// Check child0 (row 0, col 0): X should be 0, Y should be 0
	if child0.Rect.X != 0 {
		t.Errorf("Child 0 X: expected 0, got %.2f", child0.Rect.X)
	}
	if child0.Rect.Y != 0 {
		t.Errorf("Child 0 Y: expected 0, got %.2f", child0.Rect.Y)
	}

	// Check child1 (row 0, col 1): X should be 100, Y should be 0
	if child1.Rect.X != 100 {
		t.Errorf("Child 1 X: expected 100, got %.2f", child1.Rect.X)
	}
	if child1.Rect.Y != 0 {
		t.Errorf("Child 1 Y: expected 0, got %.2f", child1.Rect.Y)
	}

	// Check child2 (row 1, col 0): X should be 0, Y should be 100 (row 0 has height 100 due to stretch)
	if child2.Rect.X != 0 {
		t.Errorf("Child 2 X: expected 0, got %.2f", child2.Rect.X)
	}
	// Note: With align-content: stretch (default), rows are stretched to fill container height
	// Container height = 200, 2 rows = 100 each, so row 1 starts at Y=100
	if child2.Rect.Y != 100 {
		t.Errorf("Child 2 Y: expected 100, got %.2f", child2.Rect.Y)
	}

	// Check child3 (row 1, col 1): X should be 100, Y should be 100
	if child3.Rect.X != 100 {
		t.Errorf("Child 3 X: expected 100, got %.2f", child3.Rect.X)
	}
	if child3.Rect.Y != 100 {
		t.Errorf("Child 3 Y: expected 100, got %.2f", child3.Rect.Y)
	}

	// Check sizes (rows are stretched to 100 due to align-content: stretch)
	expectedWidth := 100.0
	expectedHeight := 100.0  // Stretched from 50 to 100
	if child0.Rect.Width != expectedWidth || child0.Rect.Height != expectedHeight {
		t.Errorf("Child 0 size: expected %.2fx%.2f, got %.2fx%.2f",
			expectedWidth, expectedHeight, child0.Rect.Width, child0.Rect.Height)
	}
}
