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

	// Item 1 has aspect ratio 2:1 and width 1000
	// Expected height: 1000 / 2 = 500
	// But it should be stretched to fill the cell if the cell is larger
	// Row 0 height should be max(item1 height, item2 doesn't affect row 0)
	// So row 0 should be at least 100 (from MinHeight), but aspect ratio gives 500
	
	// Item 1 should maintain aspect ratio: width / height = 2.0
	actualRatio := item1.Rect.Width / item1.Rect.Height
	if math.Abs(actualRatio-2.0) > 0.01 {
		t.Errorf("Item 1 should maintain aspect ratio 2:1: got %.2f:1 (width=%.2f, height=%.2f)",
			actualRatio, item1.Rect.Width, item1.Rect.Height)
	}

	// Item 1 should fill its cell width (1000)
	if math.Abs(item1.Rect.Width-1000.0) > 0.01 {
		t.Errorf("Item 1 should fill cell width: got %.2f, expected 1000", item1.Rect.Width)
	}

	// Item 1 height should be 500 (from aspect ratio)
	if math.Abs(item1.Rect.Height-500.0) > 1.0 {
		t.Errorf("Item 1 height should be 500 (from aspect ratio): got %.2f", item1.Rect.Height)
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

