package layout

import "testing"

// TestGridAutoFlowRow tests grid-auto-flow: row (default, row-major)
// Items should fill rows first, then move to next row
func TestGridAutoFlowRow(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(100), FixedTrack(100)},
			GridTemplateRows:    []GridTrack{FixedTrack(50), FixedTrack(50)},
			GridAutoFlow:        GridAutoFlowRow, // Explicit row-major
			Width:               200,
			Height:              100,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}}, // (0,0)
			{Style: Style{Width: 50, Height: 50}}, // (0,1)
			{Style: Style{Width: 50, Height: 50}}, // (1,0)
			{Style: Style{Width: 50, Height: 50}}, // (1,1)
		},
	}

	LayoutGrid(container, Loose(200, 100))

	// Row-major order: items fill row 0, then row 1
	// Item 0: (0,0)
	if container.Children[0].Rect.X != 0 || container.Children[0].Rect.Y != 0 {
		t.Errorf("Item 0 should be at (0,0), got (%v,%v)", container.Children[0].Rect.X, container.Children[0].Rect.Y)
	}
	// Item 1: (100,0) - second column, first row
	if container.Children[1].Rect.X != 100 || container.Children[1].Rect.Y != 0 {
		t.Errorf("Item 1 should be at (100,0), got (%v,%v)", container.Children[1].Rect.X, container.Children[1].Rect.Y)
	}
	// Item 2: (0,50) - first column, second row
	if container.Children[2].Rect.X != 0 || container.Children[2].Rect.Y != 50 {
		t.Errorf("Item 2 should be at (0,50), got (%v,%v)", container.Children[2].Rect.X, container.Children[2].Rect.Y)
	}
	// Item 3: (100,50) - second column, second row
	if container.Children[3].Rect.X != 100 || container.Children[3].Rect.Y != 50 {
		t.Errorf("Item 3 should be at (100,50), got (%v,%v)", container.Children[3].Rect.X, container.Children[3].Rect.Y)
	}
}

// TestGridAutoFlowColumn tests grid-auto-flow: column (column-major)
// Items should fill columns first, then move to next column
func TestGridAutoFlowColumn(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(100), FixedTrack(100)},
			GridTemplateRows:    []GridTrack{FixedTrack(50), FixedTrack(50)},
			GridAutoFlow:        GridAutoFlowColumn, // Column-major
			Width:               200,
			Height:              100,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}}, // (0,0)
			{Style: Style{Width: 50, Height: 50}}, // (1,0)
			{Style: Style{Width: 50, Height: 50}}, // (0,1)
			{Style: Style{Width: 50, Height: 50}}, // (1,1)
		},
	}

	LayoutGrid(container, Loose(200, 100))

	// Column-major order: items fill column 0, then column 1
	// Item 0: (0,0) - first column, first row
	if container.Children[0].Rect.X != 0 || container.Children[0].Rect.Y != 0 {
		t.Errorf("Item 0 should be at (0,0), got (%v,%v)", container.Children[0].Rect.X, container.Children[0].Rect.Y)
	}
	// Item 1: (0,50) - first column, second row
	if container.Children[1].Rect.X != 0 || container.Children[1].Rect.Y != 50 {
		t.Errorf("Item 1 should be at (0,50), got (%v,%v)", container.Children[1].Rect.X, container.Children[1].Rect.Y)
	}
	// Item 2: (100,0) - second column, first row
	if container.Children[2].Rect.X != 100 || container.Children[2].Rect.Y != 0 {
		t.Errorf("Item 2 should be at (100,0), got (%v,%v)", container.Children[2].Rect.X, container.Children[2].Rect.Y)
	}
	// Item 3: (100,50) - second column, second row
	if container.Children[3].Rect.X != 100 || container.Children[3].Rect.Y != 50 {
		t.Errorf("Item 3 should be at (100,50), got (%v,%v)", container.Children[3].Rect.X, container.Children[3].Rect.Y)
	}
}

// TestGridAutoFlowRowDense tests grid-auto-flow: row dense
// Dense packing fills holes left by explicitly placed items
func TestGridAutoFlowRowDense(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(50), FixedTrack(50), FixedTrack(50)},
			GridTemplateRows:    []GridTrack{FixedTrack(50), FixedTrack(50)},
			GridAutoFlow:        GridAutoFlowRowDense, // Row-major with dense packing
			Width:               150,
			Height:              100,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}}, // Auto-placed
			{Style: Style{Width: 50, Height: 50}}, // Auto-placed
			{Style: Style{Width: 50, Height: 50}}, // Auto-placed
			{Style: Style{Width: 50, Height: 50}}, // Auto-placed
		},
	}

	LayoutGrid(container, Loose(150, 100))

	// With row dense, items fill row 0 completely, then row 1
	// Item 0: (0,0)
	if container.Children[0].Rect.X != 0 || container.Children[0].Rect.Y != 0 {
		t.Errorf("Item 0 should be at (0,0), got (%v,%v)", container.Children[0].Rect.X, container.Children[0].Rect.Y)
	}
	// Item 1: (50,0)
	if container.Children[1].Rect.X != 50 || container.Children[1].Rect.Y != 0 {
		t.Errorf("Item 1 should be at (50,0), got (%v,%v)", container.Children[1].Rect.X, container.Children[1].Rect.Y)
	}
	// Item 2: (100,0)
	if container.Children[2].Rect.X != 100 || container.Children[2].Rect.Y != 0 {
		t.Errorf("Item 2 should be at (100,0), got (%v,%v)", container.Children[2].Rect.X, container.Children[2].Rect.Y)
	}
	// Item 3: (0,50) - next row
	if container.Children[3].Rect.X != 0 || container.Children[3].Rect.Y != 50 {
		t.Errorf("Item 3 should be at (0,50), got (%v,%v)", container.Children[3].Rect.X, container.Children[3].Rect.Y)
	}
}

// TestGridAutoFlowColumnDense tests grid-auto-flow: column dense
// Dense packing fills holes in column-major order
func TestGridAutoFlowColumnDense(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(50), FixedTrack(50)},
			GridTemplateRows:    []GridTrack{FixedTrack(50), FixedTrack(50), FixedTrack(50)},
			GridAutoFlow:        GridAutoFlowColumnDense, // Column-major with dense packing
			Width:               100,
			Height:              150,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}}, // Auto-placed
			{Style: Style{Width: 50, Height: 50}}, // Auto-placed
			{Style: Style{Width: 50, Height: 50}}, // Auto-placed
			{Style: Style{Width: 50, Height: 50}}, // Auto-placed
		},
	}

	LayoutGrid(container, Loose(100, 150))

	// With column dense, items fill column 0 completely, then column 1
	// Item 0: (0,0)
	if container.Children[0].Rect.X != 0 || container.Children[0].Rect.Y != 0 {
		t.Errorf("Item 0 should be at (0,0), got (%v,%v)", container.Children[0].Rect.X, container.Children[0].Rect.Y)
	}
	// Item 1: (0,50) - same column, next row
	if container.Children[1].Rect.X != 0 || container.Children[1].Rect.Y != 50 {
		t.Errorf("Item 1 should be at (0,50), got (%v,%v)", container.Children[1].Rect.X, container.Children[1].Rect.Y)
	}
	// Item 2: (0,100) - same column, third row
	if container.Children[2].Rect.X != 0 || container.Children[2].Rect.Y != 100 {
		t.Errorf("Item 2 should be at (0,100), got (%v,%v)", container.Children[2].Rect.X, container.Children[2].Rect.Y)
	}
	// Item 3: (50,0) - next column
	if container.Children[3].Rect.X != 50 || container.Children[3].Rect.Y != 0 {
		t.Errorf("Item 3 should be at (50,0), got (%v,%v)", container.Children[3].Rect.X, container.Children[3].Rect.Y)
	}
}

// TestGridAutoFlowDefault tests that default (no GridAutoFlow set) uses row-major
func TestGridAutoFlowDefault(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(100), FixedTrack(100)},
			GridTemplateRows:    []GridTrack{FixedTrack(50)},
			// GridAutoFlow not set - should default to row-major
			Width:  200,
			Height: 50,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}}, // (0,0)
			{Style: Style{Width: 50, Height: 50}}, // (0,1)
		},
	}

	LayoutGrid(container, Loose(200, 50))

	// Default should be row-major: items in same row
	if container.Children[0].Rect.Y != container.Children[1].Rect.Y {
		t.Errorf("Items should be in same row with default auto-flow, got Y: %v and %v",
			container.Children[0].Rect.Y, container.Children[1].Rect.Y)
	}
	if container.Children[1].Rect.X <= container.Children[0].Rect.X {
		t.Errorf("Second item should be to the right of first, got X: %v and %v",
			container.Children[0].Rect.X, container.Children[1].Rect.X)
	}
}

// TestGridAutoFlowWithExplicitPlacement tests auto-flow with some explicitly placed items
func TestGridAutoFlowWithExplicitPlacement(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(50), FixedTrack(50), FixedTrack(50)},
			GridTemplateRows:    []GridTrack{FixedTrack(50), FixedTrack(50)},
			GridAutoFlow:        GridAutoFlowRow,
			Width:               150,
			Height:              100,
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}}, // Auto-placed
			{Style: Style{ // Explicitly placed at (1,1)
				Width:           50,
				Height:          50,
				GridRowStart:    1,
				GridColumnStart: 1,
			}},
			{Style: Style{Width: 50, Height: 50}}, // Auto-placed
		},
	}

	LayoutGrid(container, Loose(150, 100))

	// Item 0: auto-placed at (0,0)
	if container.Children[0].Rect.X != 0 || container.Children[0].Rect.Y != 0 {
		t.Errorf("Item 0 should be auto-placed at (0,0), got (%v,%v)",
			container.Children[0].Rect.X, container.Children[0].Rect.Y)
	}

	// Item 1: explicitly placed at (1,1)
	if container.Children[1].Rect.X != 50 || container.Children[1].Rect.Y != 50 {
		t.Errorf("Item 1 should be explicitly at (50,50), got (%v,%v)",
			container.Children[1].Rect.X, container.Children[1].Rect.Y)
	}

	// Item 2: auto-placed after item 0, should be at (1,0) or later
	if container.Children[2].Rect.X < 50 {
		t.Errorf("Item 2 should be auto-placed after item 0 (X>=50), got X=%v",
			container.Children[2].Rect.X)
	}
}

// TestGridAutoFlowColumnWithGaps tests column-major flow with gaps
func TestGridAutoFlowColumnWithGaps(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(50), FixedTrack(50)},
			GridTemplateRows:    []GridTrack{FixedTrack(50), FixedTrack(50)},
			GridAutoFlow:        GridAutoFlowColumn,
			GridColumnGap:       10,
			GridRowGap:          10,
			Width:               110, // 50 + 10 + 50
			Height:              110, // 50 + 10 + 50
		},
		Children: []*Node{
			{Style: Style{Width: 50, Height: 50}}, // (0,0)
			{Style: Style{Width: 50, Height: 50}}, // (0,1) with gap
			{Style: Style{Width: 50, Height: 50}}, // (1,0) with gap
			{Style: Style{Width: 50, Height: 50}}, // (1,1) with gaps
		},
	}

	LayoutGrid(container, Loose(110, 110))

	// Item 0: (0,0)
	if container.Children[0].Rect.X != 0 || container.Children[0].Rect.Y != 0 {
		t.Errorf("Item 0 should be at (0,0), got (%v,%v)", container.Children[0].Rect.X, container.Children[0].Rect.Y)
	}

	// Item 1: (0,60) - same column, second row with gap
	if container.Children[1].Rect.X != 0 || container.Children[1].Rect.Y != 60 {
		t.Errorf("Item 1 should be at (0,60), got (%v,%v)", container.Children[1].Rect.X, container.Children[1].Rect.Y)
	}

	// Item 2: (60,0) - second column with gap, first row
	if container.Children[2].Rect.X != 60 || container.Children[2].Rect.Y != 0 {
		t.Errorf("Item 2 should be at (60,0), got (%v,%v)", container.Children[2].Rect.X, container.Children[2].Rect.Y)
	}

	// Item 3: (60,60) - second column, second row with gaps
	if container.Children[3].Rect.X != 60 || container.Children[3].Rect.Y != 60 {
		t.Errorf("Item 3 should be at (60,60), got (%v,%v)", container.Children[3].Rect.X, container.Children[3].Rect.Y)
	}
}
