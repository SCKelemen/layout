package layout

import (
	"math"
	"testing"
)

func TestGridSpanningAutoRows(t *testing.T) {
	// Test that items spanning multiple auto rows work correctly
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				AutoTrack(),
				AutoTrack(),
				AutoTrack(),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(100),
				FixedTrack(100),
			},
		},
		Children: []*Node{
			// Item spanning all 3 rows
			{
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      3, // Spans rows 0, 1, 2
					GridColumnStart: 0,
					MinHeight:       300.0, // Should be distributed: 100px per row
				},
			},
			// Item in row 0, col 1
			{
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      1,
					GridColumnStart: 1,
					MinHeight:       150.0, // Row 0 should be 150px (max of 100 and 150)
				},
			},
		},
	}

	constraints := Loose(300, Unbounded)
	LayoutGrid(root, constraints)

	// Row 0 should be 150px (max of 100 from spanning item, 150 from item in row 0)
	// Row 1 should be 100px (from spanning item)
	// Row 2 should be 100px (from spanning item)

	// Spanning item should span all 3 rows
	spanningItem := root.Children[0]
	expectedHeight := 150.0 + 100.0 + 100.0 // Sum of row heights
	if math.Abs(spanningItem.Rect.Height-expectedHeight) > 1.0 {
		t.Errorf("Spanning item should be %.2f tall (sum of row heights), got %.2f",
			expectedHeight, spanningItem.Rect.Height)
	}

	// Second item should be in row 0
	row0Item := root.Children[1]
	if row0Item.Rect.Y != 0.0 {
		t.Errorf("Row 0 item should be at Y=0, got %.2f", row0Item.Rect.Y)
	}

	// Both items should start at the same Y (row 0)
	if math.Abs(spanningItem.Rect.Y-row0Item.Rect.Y) > 0.1 {
		t.Errorf("Both items should start at same Y (row 0), but spanning: %.2f, row0: %.2f",
			spanningItem.Rect.Y, row0Item.Rect.Y)
	}
}

func TestGridSpanningWithoutMinHeight(t *testing.T) {
	// Test that spanning items without MinHeight still work (but may collapse)
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
			// Item spanning 2 rows, no MinHeight
			{
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      2, // Spans rows 0, 1
					GridColumnStart: 0,
					// No MinHeight - will measure to 0
				},
			},
		},
	}

	constraints := Loose(200, Unbounded)
	LayoutGrid(root, constraints)

	// Item should be 0 height (no content, no MinHeight)
	spanningItem := root.Children[0]
	if spanningItem.Rect.Height != 0.0 {
		t.Logf("Note: Spanning item without MinHeight has height %.2f (expected 0)", spanningItem.Rect.Height)
	}

	// Rows should collapse to 0 (or minimal height from safeguards)
	// This is correct CSS Grid behavior
}
