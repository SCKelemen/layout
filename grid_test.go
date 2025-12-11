package layout

import (
	"math"
	"testing"
)

func TestGridBasic(t *testing.T) {
	// Test basic 2x2 grid
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(100),
				FixedTrack(100),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(100),
				FixedTrack(100),
			},
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 0, GridColumnEnd: 1}},
			{Style: Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 1, GridColumnEnd: 2}},
			{Style: Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 0, GridColumnEnd: 1}},
			{Style: Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 1, GridColumnEnd: 2}},
		},
	}

	constraints := Loose(300, 300)
	size := LayoutGrid(root, constraints)

	// Grid should be 200x200 (2 rows * 100, 2 cols * 100)
	expectedWidth := 200.0
	expectedHeight := 200.0

	if math.Abs(size.Width-expectedWidth) > 1.0 {
		t.Errorf("Expected grid width %.2f, got %.2f", expectedWidth, size.Width)
	}
	if math.Abs(size.Height-expectedHeight) > 1.0 {
		t.Errorf("Expected grid height %.2f, got %.2f", expectedHeight, size.Height)
	}

	// Check first item position
	if root.Children[0].Rect.X != 0 {
		t.Errorf("First item X should be 0, got %.2f", root.Children[0].Rect.X)
	}
	if root.Children[0].Rect.Y != 0 {
		t.Errorf("First item Y should be 0, got %.2f", root.Children[0].Rect.Y)
	}

	// Check second item (should be to the right)
	if math.Abs(root.Children[1].Rect.X-100.0) > 1.0 {
		t.Errorf("Second item X should be 100, got %.2f", root.Children[1].Rect.X)
	}
}

func TestGridFractionalUnits(t *testing.T) {
	// Test fractional units (fr)
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(100),
			},
			GridTemplateColumns: []GridTrack{
				FractionTrack(1),
				FractionTrack(2),
			},
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 0, GridColumnEnd: 1}},
			{Style: Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 1, GridColumnEnd: 2}},
		},
	}

	constraints := Tight(300, 200)
	LayoutGrid(root, constraints)

	// Second column should be twice as wide as first
	col0Width := root.Children[0].Rect.Width
	col1Width := root.Children[1].Rect.Width

	expectedRatio := 2.0
	actualRatio := col1Width / col0Width

	if math.Abs(actualRatio-expectedRatio) > 0.1 {
		t.Errorf("Expected column ratio %.2f, got %.2f (col0=%.2f, col1=%.2f)",
			expectedRatio, actualRatio, col0Width, col1Width)
	}
}

func TestGridGap(t *testing.T) {
	// Test grid gap
	gap := 10.0
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(100),
				FixedTrack(100),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(100),
				FixedTrack(100),
			},
			GridGap: gap,
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 0, GridColumnEnd: 1}},
			{Style: Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 1, GridColumnEnd: 2}},
		},
	}

	constraints := Loose(300, 300)
	size := LayoutGrid(root, constraints)

	// Grid should include gap: 2*100 (columns) + 1*gap = 200 + 10 = 210
	expectedWidth := 200.0 + gap
	if math.Abs(size.Width-expectedWidth) > 1.0 {
		t.Errorf("Expected grid width with gap %.2f, got %.2f", expectedWidth, size.Width)
	}

	// Second item should have gap before it
	expectedX := 100.0 + gap
	if math.Abs(root.Children[1].Rect.X-expectedX) > 1.0 {
		t.Errorf("Second item X should be %.2f (100 + gap), got %.2f", expectedX, root.Children[1].Rect.X)
	}
}

func TestGridSpanning(t *testing.T) {
	// Test grid items spanning multiple cells
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(50),
				FixedTrack(50),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(100),
				FixedTrack(100),
			},
		},
		Children: []*Node{
			// Item spanning full width
			{Style: Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 0, GridColumnEnd: 2}},
			{Style: Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 0, GridColumnEnd: 1}},
			{Style: Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 1, GridColumnEnd: 2}},
		},
	}

	constraints := Loose(300, 200)
	LayoutGrid(root, constraints)

	// First item should span both columns
	expectedWidth := 200.0
	if math.Abs(root.Children[0].Rect.Width-expectedWidth) > 1.0 {
		t.Errorf("First item should span full width %.2f, got %.2f", expectedWidth, root.Children[0].Rect.Width)
	}
}

func TestGridAutoRows(t *testing.T) {
	// Test auto rows
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateColumns: []GridTrack{
				FixedTrack(100),
			},
			GridAutoRows: FixedTrack(50),
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridColumnStart: 0}},
			{Style: Style{GridRowStart: 1, GridColumnStart: 0}},
			{Style: Style{GridRowStart: 2, GridColumnStart: 0}},
		},
	}

	constraints := Loose(200, 300)
	LayoutGrid(root, constraints)

	// All rows should be 50 high
	for i, child := range root.Children {
		if math.Abs(child.Rect.Height-50.0) > 1.0 {
			t.Errorf("Child %d should have height 50, got %.2f", i, child.Rect.Height)
		}
	}

	// Second child should be below first
	if math.Abs(root.Children[1].Rect.Y-50.0) > 1.0 {
		t.Errorf("Second child Y should be 50, got %.2f", root.Children[1].Rect.Y)
	}
}

func TestGridMinMaxTrack(t *testing.T) {
	// Test minmax track sizing
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(100),
			},
			GridTemplateColumns: []GridTrack{
				MinMaxTrack(50, 150),
			},
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridColumnStart: 0}},
		},
	}

	constraints := Tight(200, 200)
	LayoutGrid(root, constraints)

	// Column should be within minmax bounds
	colWidth := root.Children[0].Rect.Width
	if colWidth < 50 || colWidth > 150 {
		t.Errorf("Column width should be between 50 and 150, got %.2f", colWidth)
	}
}

func TestGridEmpty(t *testing.T) {
	// Test empty grid
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
		Children: []*Node{},
	}

	constraints := Loose(200, 200)
	size := LayoutGrid(root, constraints)

	// Empty grid should still have size based on tracks
	expectedWidth := 100.0
	expectedHeight := 100.0

	if math.Abs(size.Width-expectedWidth) > 1.0 {
		t.Errorf("Expected empty grid width %.2f, got %.2f", expectedWidth, size.Width)
	}
	if math.Abs(size.Height-expectedHeight) > 1.0 {
		t.Errorf("Expected empty grid height %.2f, got %.2f", expectedHeight, size.Height)
	}
}

func TestGridPadding(t *testing.T) {
	// Test padding affects grid size
	padding := 20.0
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(100),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(100),
			},
			Padding: Uniform(padding),
		},
		Children: []*Node{
			{Style: Style{GridRowStart: 0, GridColumnStart: 0}},
		},
	}

	constraints := Loose(300, 300)
	size := LayoutGrid(root, constraints)

	// Grid should include padding: 100 (content) + 40 (padding) = 140
	expectedWidth := 100.0 + padding*2
	if math.Abs(size.Width-expectedWidth) > 1.0 {
		t.Errorf("Expected grid width with padding %.2f, got %.2f", expectedWidth, size.Width)
	}
}

func TestGridNested(t *testing.T) {
	// Test nested grids
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
			{
				Style: Style{
					Display: DisplayGrid,
					GridRowStart: 0,
					GridColumnStart: 0,
					GridTemplateRows: []GridTrack{
						FixedTrack(50),
					},
					GridTemplateColumns: []GridTrack{
						FixedTrack(50),
					},
				},
				Children: []*Node{
					{Style: Style{GridRowStart: 0, GridColumnStart: 0}},
				},
			},
		},
	}

	constraints := Loose(200, 200)
	LayoutGrid(root, constraints)

	// Nested grid should be laid out
	if len(root.Children[0].Children) != 1 {
		t.Errorf("Expected 1 child in nested grid, got %d", len(root.Children[0].Children))
	}
}

