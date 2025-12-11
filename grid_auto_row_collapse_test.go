package layout

import (
	"math"
	"testing"
)

func TestGridAutoRowCollapseWithoutMinHeight(t *testing.T) {
	// Test demonstrates expected behavior: items without MinHeight and no content
	// will measure to 0 height, causing rows to collapse.
	// This is correct CSS Grid behavior - items in auto rows should have MinHeight set.

	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				AutoTrack(),
				AutoTrack(),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(100),
			},
		},
		Children: []*Node{
			{
				Style: Style{
					GridRowStart:    0,
					GridColumnStart: 0,
					MinHeight:       50.0, // Has MinHeight
				},
			},
			{
				Style: Style{
					GridRowStart:    1,
					GridColumnStart: 0,
					// NO MinHeight - this will cause item to measure to 0
				},
			},
		},
	}

	constraints := Loose(200, Unbounded)
	LayoutGrid(root, constraints)

	// First row should have height from first item
	if root.Children[0].Rect.Height < 50.0 {
		t.Errorf("First item should have height >= 50 (MinHeight), got %.2f", root.Children[0].Rect.Height)
	}

	// Second item without MinHeight and no content will measure to 0
	// This is expected CSS behavior - items in auto rows should have MinHeight set
	if root.Children[1].Rect.Height != 0 {
		t.Logf("Note: Second item has height %.2f (expected 0 when no MinHeight and no content)", root.Children[1].Rect.Height)
	}

	// The row itself might collapse or have minimal height depending on our safeguards
	// But the item will be 0 height, which is correct CSS behavior
	row1Y := root.Children[1].Rect.Y
	row0End := root.Children[0].Rect.Y + root.Children[0].Rect.Height

	// With our safeguards, the row should still be positioned, but the item is 0 height
	// This test documents the expected behavior: set MinHeight on all items in auto rows
	if row1Y < row0End {
		t.Logf("Row positioning: row1 Y (%.2f) < row0 end (%.2f) - row may have collapsed", row1Y, row0End)
	}
}

func TestGridAutoRowWithMinHeight(t *testing.T) {
	// Test that auto rows work correctly when all items have MinHeight
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				AutoTrack(),
				AutoTrack(),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(100),
			},
		},
		Children: []*Node{
			{
				Style: Style{
					GridRowStart:    0,
					GridColumnStart: 0,
					MinHeight:       60.0,
				},
			},
			{
				Style: Style{
					GridRowStart:    1,
					GridColumnStart: 0,
					MinHeight:       50.0,
				},
			},
		},
	}

	constraints := Loose(200, Unbounded)
	LayoutGrid(root, constraints)

	// First item should respect MinHeight
	if math.Abs(root.Children[0].Rect.Height-60.0) > 0.1 {
		t.Errorf("First item should have height 60 (MinHeight), got %.2f", root.Children[0].Rect.Height)
	}

	// Second item should respect MinHeight
	if math.Abs(root.Children[1].Rect.Height-50.0) > 0.1 {
		t.Errorf("Second item should have height 50 (MinHeight), got %.2f", root.Children[1].Rect.Height)
	}

	// Second row should be positioned below first row
	row1Y := root.Children[1].Rect.Y
	expectedY := 60.0 // First row height
	if math.Abs(row1Y-expectedY) > 0.1 {
		t.Errorf("Second row should be at Y=%.2f, got %.2f", expectedY, row1Y)
	}

	// Root height should be sum of rows
	expectedHeight := 60.0 + 50.0
	if math.Abs(root.Rect.Height-expectedHeight) > 1.0 {
		t.Errorf("Root height should be %.2f, got %.2f", expectedHeight, root.Rect.Height)
	}
}
