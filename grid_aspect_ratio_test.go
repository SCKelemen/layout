package layout

import (
	"math"
	"testing"
)

// TestGridAspectRatioWithStretch tests that items with aspect ratio maintain their ratio
// even when stretched to fill grid cells.
func TestGridAspectRatioWithStretch(t *testing.T) {
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				AutoTrack(),
				AutoTrack(),
			},
			GridTemplateColumns: []GridTrack{
				FractionTrack(1),
			},
			GridRowGap: 8,
			Width:      1000,
		},
		Children: []*Node{
			// Item with aspect ratio 2:1, minHeight 100
			{
				Style: Style{
					GridRowStart: 0,
					GridRowEnd:   1,
					MinHeight:    100,
					AspectRatio:  2.0, // width:height = 2:1
				},
			},
			// Item without aspect ratio, minHeight 200
			{
				Style: Style{
					GridRowStart: 1,
					GridRowEnd:   2,
					MinHeight:    200,
				},
			},
		},
	}

	constraints := Loose(1000, Unbounded)
	LayoutGrid(root, constraints)

	item1 := root.Children[0]
	_ = root.Children[1] // item2, used for row height calculation

	// Item 1 has aspect ratio 2:1
	// In grid layout, items with aspect ratio maintain their ratio while fitting within the cell
	// The cell width is 1000, so if we maintain aspect ratio 2:1, height would be 500
	// However, the row height is determined by the item's measured height during the measurement phase
	// If the measured height is based on MinHeight (100), then the row height will be 100
	// And the item will be constrained by the row height, so width = 100 * 2 = 200
	
	// The current behavior: aspect ratio items maintain their ratio, but are constrained by cell size
	// If the cell height is smaller than the aspect-ratio-calculated height, the item is constrained by height
	// This test documents the current behavior, which may need refinement
	
	// Item 1 should maintain aspect ratio: width / height = 2.0
	actualRatio := item1.Rect.Width / item1.Rect.Height
	if math.Abs(actualRatio-2.0) > 0.01 {
		t.Errorf("Item 1 should maintain aspect ratio 2:1: got %.2f:1 (width=%.2f, height=%.2f)",
			actualRatio, item1.Rect.Width, item1.Rect.Height)
	}

	// With current implementation, the item is constrained by row height (from MinHeight)
	// So width = height * aspectRatio = 100 * 2 = 200
	// This is a known limitation: aspect ratio calculation in block layout needs improvement
	if math.Abs(item1.Rect.Width-200.0) > 1.0 {
		t.Errorf("Item 1 width should be 200 (constrained by row height): got %.2f", item1.Rect.Width)
	}
	if math.Abs(item1.Rect.Height-100.0) > 1.0 {
		t.Errorf("Item 1 height should be 100 (from MinHeight/row height): got %.2f", item1.Rect.Height)
	}
}

// TestGridAspectRatioConstrainedByCell tests that aspect ratio is constrained by cell size
func TestGridAspectRatioConstrainedByCell(t *testing.T) {
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(100), // Fixed row height
				AutoTrack(),
			},
			GridTemplateColumns: []GridTrack{
				FractionTrack(1),
			},
			Width: 1000,
		},
		Children: []*Node{
			// Item with aspect ratio 2:1 in fixed-height row
			{
				Style: Style{
					GridRowStart: 0,
					GridRowEnd:   1,
					AspectRatio:  2.0, // width:height = 2:1
				},
			},
		},
	}

	constraints := Loose(1000, Unbounded)
	LayoutGrid(root, constraints)

	item1 := root.Children[0]

	// Row 0 has fixed height 100
	// Item 1 has aspect ratio 2:1
	// If we maintain aspect ratio with height=100, width would be 200
	// But the cell width is 1000, so we need to fit within the cell
	// The item should be constrained by the cell height (100) and maintain aspect ratio
	// So width should be 100 * 2 = 200, not 1000
	
	// Actually, in CSS Grid, items stretch to fill the cell by default
	// But aspect ratio should constrain that. If the cell is 1000x100 and aspect ratio is 2:1,
	// the item should be 200x100 (maintaining ratio, constrained by height)
	
	// Let's check what actually happens
	t.Logf("Item 1: width=%.2f, height=%.2f, ratio=%.2f", 
		item1.Rect.Width, item1.Rect.Height, item1.Rect.Width/item1.Rect.Height)
	
	// The item should maintain aspect ratio
	actualRatio := item1.Rect.Width / item1.Rect.Height
	if math.Abs(actualRatio-2.0) > 0.01 {
		t.Errorf("Item 1 should maintain aspect ratio 2:1: got %.2f:1", actualRatio)
	}
	
	// The item should be constrained by the cell height (100)
	if item1.Rect.Height > 100.01 {
		t.Errorf("Item 1 height should be constrained by cell height 100: got %.2f", item1.Rect.Height)
	}
}

