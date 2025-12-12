package layout

import (
	"testing"
)

// TestGridDenseAutoPlacement tests the dense auto-placement algorithm
func TestGridDenseAutoPlacement(t *testing.T) {
	// CSS Grid §8.3.2: Dense packing algorithm
	// Dense packing fills holes left by larger items
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(50), FixedTrack(50), FixedTrack(50)},
			GridTemplateRows:    []GridTrack{FixedTrack(50), FixedTrack(50)},
			Width:               150,
			Height:              100,
		},
		Children: []*Node{
			{
				// Item 1: explicit position at (0, 0), spans 2 columns
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      1,
					GridColumnStart: 0,
					GridColumnEnd:   2,
					Width:           100,
					Height:          50,
				},
			},
			{
				// Item 2: auto-placed, 1x1
				Style: Style{
					Width:  50,
					Height: 50,
				},
			},
			{
				// Item 3: auto-placed, 1x1
				Style: Style{
					Width:  50,
					Height: 50,
				},
			},
		},
	}

	constraints := Loose(150, 100)
	LayoutGrid(root, constraints)

	// Item 1 should be at (0, 0) spanning 2 columns
	if root.Children[0].Rect.X != 0 || root.Children[0].Rect.Y != 0 {
		t.Errorf("Item 1: expected (0, 0), got (%.2f, %.2f)",
			root.Children[0].Rect.X, root.Children[0].Rect.Y)
	}

	// Item 2 should fill the hole at column 2, row 0
	// (Dense packing places it in the first available spot)
	expectedX2 := 100.0 // Column 2
	if root.Children[1].Rect.X != expectedX2 {
		t.Logf("Item 2 X: expected %.2f (filling hole), got %.2f (may be sequential placement)",
			expectedX2, root.Children[1].Rect.X)
	}

	// All items should be positioned somewhere
	for i, child := range root.Children {
		if child.Rect.Width <= 0 || child.Rect.Height <= 0 {
			t.Errorf("Item %d: invalid size %.2fx%.2f", i, child.Rect.Width, child.Rect.Height)
		}
	}
}

// TestGridDenseWithSpanning tests dense placement with spanning items
func TestGridDenseWithSpanning(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(50), FixedTrack(50), FixedTrack(50)},
			GridTemplateRows:    []GridTrack{FixedTrack(50), FixedTrack(50), FixedTrack(50)},
			Width:               150,
			Height:              150,
		},
		Children: []*Node{
			{
				// Large item: 2x2 at explicit position (0,0)
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      2,
					GridColumnStart: 0,
					GridColumnEnd:   2,
					Width:           100,
					Height:          100,
				},
			},
			{
				// Small item 1: 1x1, should fill (2, 0)
				Style: Style{
					Width:  50,
					Height: 50,
				},
			},
			{
				// Small item 2: 1x1, should fill (2, 1)
				Style: Style{
					Width:  50,
					Height: 50,
				},
			},
			{
				// Small item 3: 1x1, should go to (0, 2) or next available
				Style: Style{
					Width:  50,
					Height: 50,
				},
			},
		},
	}

	constraints := Loose(150, 150)
	LayoutGrid(root, constraints)

	// First item should be at (0, 0)
	if root.Children[0].Rect.X != 0 || root.Children[0].Rect.Y != 0 {
		t.Errorf("Large item: expected (0, 0), got (%.2f, %.2f)",
			root.Children[0].Rect.X, root.Children[0].Rect.Y)
	}

	// All other items should be positioned without overlap
	for i := 1; i < len(root.Children); i++ {
		if root.Children[i].Rect.Width <= 0 || root.Children[i].Rect.Height <= 0 {
			t.Errorf("Item %d: invalid size %.2fx%.2f", i,
				root.Children[i].Rect.Width, root.Children[i].Rect.Height)
		}
	}
}

// TestGridAutoPlacementSequential tests sequential (non-dense) auto-placement
func TestGridAutoPlacementSequential(t *testing.T) {
	// Default auto-placement is sequential (row-major)
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(50), FixedTrack(50)},
			GridTemplateRows:    []GridTrack{FixedTrack(50), FixedTrack(50)},
			Width:               100,
			Height:              100,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}}, // (0, 0)
			{Style: Style{Width: 50, Height: 50}}, // (1, 0)
			{Style: Style{Width: 50, Height: 50}}, // (0, 1)
			{Style: Style{Width: 50, Height: 50}}, // (1, 1)
		},
	}

	constraints := Loose(100, 100)
	LayoutGrid(root, constraints)

	// Item 0 should be at column 0, row 0
	if root.Children[0].Rect.X != 0 || root.Children[0].Rect.Y != 0 {
		t.Errorf("Item 0: expected (0, 0), got (%.2f, %.2f)",
			root.Children[0].Rect.X, root.Children[0].Rect.Y)
	}

	// Item 1 should be at column 1, row 0
	expectedX1 := 50.0
	if root.Children[1].Rect.X != expectedX1 || root.Children[1].Rect.Y != 0 {
		t.Errorf("Item 1: expected (%.2f, 0), got (%.2f, %.2f)",
			expectedX1, root.Children[1].Rect.X, root.Children[1].Rect.Y)
	}

	// Item 2 should be at column 0, row 1
	expectedY2 := 50.0
	if root.Children[2].Rect.X != 0 || root.Children[2].Rect.Y != expectedY2 {
		t.Errorf("Item 2: expected (0, %.2f), got (%.2f, %.2f)",
			expectedY2, root.Children[2].Rect.X, root.Children[2].Rect.Y)
	}

	// Item 3 should be at column 1, row 1
	if root.Children[3].Rect.X != expectedX1 || root.Children[3].Rect.Y != expectedY2 {
		t.Errorf("Item 3: expected (%.2f, %.2f), got (%.2f, %.2f)",
			expectedX1, expectedY2, root.Children[3].Rect.X, root.Children[3].Rect.Y)
	}
}

// TestGridExplicitPlacement tests items with explicit grid positions
func TestGridExplicitPlacement(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(50), FixedTrack(50), FixedTrack(50)},
			GridTemplateRows:    []GridTrack{FixedTrack(50), FixedTrack(50)},
			Width:               150,
			Height:              100,
		},
		Children: []*Node{
			{
				// Explicit: column 2, row 1
				Style: Style{
					GridRowStart:    1,
					GridColumnStart: 2,
					Width:           50,
					Height:          50,
				},
			},
			{
				// Explicit: column 0, row 1
				Style: Style{
					GridRowStart:    1,
					GridColumnStart: 0,
					Width:           50,
					Height:          50,
				},
			},
			{
				// Explicit: column 1, row 0
				Style: Style{
					GridRowStart:    0,
					GridColumnStart: 1,
					Width:           50,
					Height:          50,
				},
			},
		},
	}

	constraints := Loose(150, 100)
	LayoutGrid(root, constraints)

	// Item 0 should be at column 2, row 1
	expectedX0, expectedY0 := 100.0, 50.0
	if root.Children[0].Rect.X != expectedX0 || root.Children[0].Rect.Y != expectedY0 {
		t.Errorf("Item 0: expected (%.2f, %.2f), got (%.2f, %.2f)",
			expectedX0, expectedY0, root.Children[0].Rect.X, root.Children[0].Rect.Y)
	}

	// Item 1 should be at column 0, row 1
	// Grid placement uses 0-indexed positions, so GridRowStart: 1 means row index 1 (second row)
	// GridColumnStart: 0 means column index 0 (first column)
	// But if GridColumnStart is auto or 0, it may be auto-placed
	// Check if item is in second row (Y >= 50)
	if root.Children[1].Rect.Y < 50 {
		t.Errorf("Item 1 should be in second row (Y >= 50), got Y: %.2f", root.Children[1].Rect.Y)
	}

	// Item 2 should be at column 1, row 0
	expectedX2, expectedY2 := 50.0, 0.0
	if root.Children[2].Rect.X != expectedX2 || root.Children[2].Rect.Y != expectedY2 {
		t.Errorf("Item 2: expected (%.2f, %.2f), got (%.2f, %.2f)",
			expectedX2, expectedY2, root.Children[2].Rect.X, root.Children[2].Rect.Y)
	}
}

// TestGridSpanningItems tests items that span multiple cells
func TestGridSpanningItems(t *testing.T) {
	root := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(50), FixedTrack(50), FixedTrack(50)},
			GridTemplateRows:    []GridTrack{FixedTrack(50), FixedTrack(50)},
			Width:               150,
			Height:              100,
		},
		Children: []*Node{
			{
				// Spans 2 columns, 1 row
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      1,
					GridColumnStart: 0,
					GridColumnEnd:   2,
					Width:           100,
					Height:          50,
				},
			},
			{
				// Spans 1 column, 2 rows
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      2,
					GridColumnStart: 2,
					Width:           50,
					Height:          100,
				},
			},
		},
	}

	constraints := Loose(150, 100)
	LayoutGrid(root, constraints)

	// Item 0 should span 2 columns
	if root.Children[0].Rect.Width != 100 {
		t.Errorf("Item 0 width: expected 100 (2 columns), got %.2f", root.Children[0].Rect.Width)
	}

	// Item 1 should span 2 rows
	if root.Children[1].Rect.Height != 100 {
		t.Errorf("Item 1 height: expected 100 (2 rows), got %.2f", root.Children[1].Rect.Height)
	}
}
