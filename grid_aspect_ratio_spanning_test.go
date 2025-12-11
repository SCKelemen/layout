package layout

import (
	"math"
	"testing"
)

// TestGridAspectRatioSpanningColumns tests aspect ratio with items spanning multiple columns
func TestGridAspectRatioSpanningColumns(t *testing.T) {
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				AutoTrack(),
			},
			GridTemplateColumns: []GridTrack{
				FractionTrack(1),
				FractionTrack(1),
			},
			GridColumnGap: 8,
			Width:         1000,
		},
		Children: []*Node{
			// Item spanning 2 columns with aspect ratio 2:1
			{
				Style: Style{
					GridColumnStart: 0,
					GridColumnEnd:   2,
					AspectRatio:     2.0, // width:height = 2:1
				},
			},
		},
	}

	constraints := Loose(1000, Unbounded)
	LayoutGrid(root, constraints)

	item := root.Children[0]

	// Item spans 2 columns, so cell width = (1000 - 8) / 2 * 2 + 8 = 1000
	// With aspect ratio 2:1, height should be 1000 / 2 = 500
	expectedWidth := 1000.0
	expectedHeight := 500.0

	if math.Abs(item.Rect.Width-expectedWidth) > 1.0 {
		t.Errorf("Item width should be %.2f (spans 2 columns): got %.2f", expectedWidth, item.Rect.Width)
	}

	if math.Abs(item.Rect.Height-expectedHeight) > 1.0 {
		t.Errorf("Item height should be %.2f (from aspect ratio): got %.2f", expectedHeight, item.Rect.Height)
	}

	// Verify aspect ratio is maintained
	actualRatio := item.Rect.Width / item.Rect.Height
	if math.Abs(actualRatio-2.0) > 0.01 {
		t.Errorf("Item should maintain aspect ratio 2:1: got %.2f:1", actualRatio)
	}
}

// TestGridAspectRatioSpanningRows tests aspect ratio with items spanning multiple rows
func TestGridAspectRatioSpanningRows(t *testing.T) {
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
			// Item spanning 2 rows with aspect ratio 2:1
			{
				Style: Style{
					GridRowStart: 0,
					GridRowEnd:   2,
					AspectRatio:  2.0, // width:height = 2:1
				},
			},
		},
	}

	constraints := Loose(1000, Unbounded)
	LayoutGrid(root, constraints)

	item := root.Children[0]

	// Item spans 2 rows, cell width = 1000
	// With aspect ratio 2:1, height should be 1000 / 2 = 500
	// But the row height is determined by the item, so row height = (500 - 8) / 2 = 246
	// Actually, the item height determines the row heights, so:
	// row0 + gap + row1 = itemHeight
	// If row0 = row1 = h, then 2h + 8 = 500, so h = 246
	// But the item should maintain aspect ratio, so height = 500
	expectedWidth := 1000.0
	expectedHeight := 500.0

	if math.Abs(item.Rect.Width-expectedWidth) > 1.0 {
		t.Errorf("Item width should be %.2f: got %.2f", expectedWidth, item.Rect.Width)
	}

	if math.Abs(item.Rect.Height-expectedHeight) > 1.0 {
		t.Errorf("Item height should be %.2f (from aspect ratio): got %.2f", expectedHeight, item.Rect.Height)
	}

	// Verify aspect ratio is maintained
	actualRatio := item.Rect.Width / item.Rect.Height
	if math.Abs(actualRatio-2.0) > 0.01 {
		t.Errorf("Item should maintain aspect ratio 2:1: got %.2f:1", actualRatio)
	}
}

// TestGridAspectRatioSpanningBoth tests aspect ratio with items spanning both rows and columns
func TestGridAspectRatioSpanningBoth(t *testing.T) {
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				AutoTrack(),
				AutoTrack(),
			},
			GridTemplateColumns: []GridTrack{
				FractionTrack(1),
				FractionTrack(1),
			},
			GridRowGap:    8,
			GridColumnGap: 8,
			Width:         1000,
		},
		Children: []*Node{
			// Item spanning 2 rows and 2 columns with aspect ratio 2:1
			{
				Style: Style{
					GridRowStart:   0,
					GridRowEnd:     2,
					GridColumnStart: 0,
					GridColumnEnd:   2,
					AspectRatio:    2.0, // width:height = 2:1
				},
			},
		},
	}

	constraints := Loose(1000, Unbounded)
	LayoutGrid(root, constraints)

	item := root.Children[0]

	// Item spans 2 rows and 2 columns
	// Cell width = (1000 - 8) / 2 * 2 + 8 = 1000
	// With aspect ratio 2:1, height should be 1000 / 2 = 500
	expectedWidth := 1000.0
	expectedHeight := 500.0

	if math.Abs(item.Rect.Width-expectedWidth) > 1.0 {
		t.Errorf("Item width should be %.2f (spans 2 columns): got %.2f", expectedWidth, item.Rect.Width)
	}

	if math.Abs(item.Rect.Height-expectedHeight) > 1.0 {
		t.Errorf("Item height should be %.2f (from aspect ratio): got %.2f", expectedHeight, item.Rect.Height)
	}

	// Verify aspect ratio is maintained
	actualRatio := item.Rect.Width / item.Rect.Height
	if math.Abs(actualRatio-2.0) > 0.01 {
		t.Errorf("Item should maintain aspect ratio 2:1: got %.2f:1", actualRatio)
	}
}

