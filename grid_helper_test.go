package layout

import (
	"math"
	"testing"
)

func TestGridHelperRowPositioning(t *testing.T) {
	// Test that the Grid() helper function creates grids with proper row positioning
	grid := Grid(3, 2, 100, 100) // 3 rows, 2 columns, 100px each

	// Add children without explicit positioning (should auto-place)
	grid.Children = []*Node{
		{Style: Style{Width: 100, Height: 100}},
		{Style: Style{Width: 100, Height: 100}},
		{Style: Style{Width: 100, Height: 100}},
		{Style: Style{Width: 100, Height: 100}},
		{Style: Style{Width: 100, Height: 100}},
		{Style: Style{Width: 100, Height: 100}},
	}

	constraints := Loose(500, 500)
	LayoutGrid(grid, constraints)

	// Verify items are positioned in different rows
	// First row: items 0, 1
	// Second row: items 2, 3
	// Third row: items 4, 5

	// Items in same row should have same Y
	if math.Abs(grid.Children[0].Rect.Y-grid.Children[1].Rect.Y) > 0.1 {
		t.Errorf("Items 0 and 1 should be in same row (same Y), got %.2f and %.2f",
			grid.Children[0].Rect.Y, grid.Children[1].Rect.Y)
	}

	// Items in different rows should have different Y
	if math.Abs(grid.Children[0].Rect.Y-grid.Children[2].Rect.Y) < 1.0 {
		t.Errorf("Items 0 and 2 should be in different rows (different Y), but both are at %.2f",
			grid.Children[0].Rect.Y)
	}

	// Second row should be below first row
	if grid.Children[2].Rect.Y <= grid.Children[0].Rect.Y {
		t.Errorf("Second row should be below first row, but Item 2 Y (%.2f) <= Item 0 Y (%.2f)",
			grid.Children[2].Rect.Y, grid.Children[0].Rect.Y)
	}

	// Third row should be below second row
	if grid.Children[4].Rect.Y <= grid.Children[2].Rect.Y {
		t.Errorf("Third row should be below second row, but Item 4 Y (%.2f) <= Item 2 Y (%.2f)",
			grid.Children[4].Rect.Y, grid.Children[2].Rect.Y)
	}
}

func TestGridHelperWithExplicitPositions(t *testing.T) {
	// Test Grid helper with explicitly positioned items
	grid := Grid(2, 2, 150, 150)

	item1 := &Node{Style: Style{Width: 150, Height: 150}}
	item1.Style.GridRowStart = 0
	item1.Style.GridColumnStart = 0

	item2 := &Node{Style: Style{Width: 150, Height: 150}}
	item2.Style.GridRowStart = 0
	item2.Style.GridColumnStart = 1

	item3 := &Node{Style: Style{Width: 150, Height: 150}}
	item3.Style.GridRowStart = 1
	item3.Style.GridColumnStart = 0

	item4 := &Node{Style: Style{Width: 150, Height: 150}}
	item4.Style.GridRowStart = 1
	item4.Style.GridColumnStart = 1

	grid.Children = []*Node{item1, item2, item3, item4}

	constraints := Loose(500, 500)
	LayoutGrid(grid, constraints)

	// First row items should have same Y
	if math.Abs(grid.Children[0].Rect.Y-grid.Children[1].Rect.Y) > 0.1 {
		t.Errorf("First row items should have same Y, got %.2f and %.2f",
			grid.Children[0].Rect.Y, grid.Children[1].Rect.Y)
	}

	// Second row items should have same Y
	if math.Abs(grid.Children[2].Rect.Y-grid.Children[3].Rect.Y) > 0.1 {
		t.Errorf("Second row items should have same Y, got %.2f and %.2f",
			grid.Children[2].Rect.Y, grid.Children[3].Rect.Y)
	}

	// Second row should be below first row
	expectedY2 := 150.0 // First row height
	actualY2 := grid.Children[2].Rect.Y
	if math.Abs(actualY2-expectedY2) > 0.1 {
		t.Errorf("Second row should be at Y=%.2f, got %.2f", expectedY2, actualY2)
	}
}

func TestGridAutoRowPositioning(t *testing.T) {
	// Test GridAuto helper
	grid := GridAuto(2, 2)

	item1 := &Node{Style: Style{Width: 100, Height: 80}}
	item1.Style.GridRowStart = 0
	item1.Style.GridColumnStart = 0

	item2 := &Node{Style: Style{Width: 100, Height: 120}}
	item2.Style.GridRowStart = 1
	item2.Style.GridColumnStart = 0

	grid.Children = []*Node{item1, item2}

	constraints := Loose(500, 500)
	LayoutGrid(grid, constraints)

	// Second item should be below first
	if grid.Children[1].Rect.Y <= grid.Children[0].Rect.Y {
		t.Errorf("Second row should be below first, but Item 1 Y (%.2f) <= Item 0 Y (%.2f)",
			grid.Children[1].Rect.Y, grid.Children[0].Rect.Y)
	}

	// Second item Y should be at least the height of first item
	minY2 := grid.Children[0].Rect.Height
	if grid.Children[1].Rect.Y < minY2 {
		t.Errorf("Second row should be at least %.2f below first row, but got %.2f",
			minY2, grid.Children[1].Rect.Y)
	}
}

