package layout

import (
	"math"
	"testing"
)

// TestGridSpanningMargin tests that margins on spanning items don't
// cause extra gap between rows
func TestGridSpanningMargin(t *testing.T) {
	// Create a grid with an item spanning rows, followed by items in subsequent rows
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(50),
				FixedTrack(50),
				FixedTrack(50),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(100),
			},
			GridRowGap: 10,
		},
		Children: []*Node{
			// Item spanning rows 0-1 with margin
			{
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      2,
					GridColumnStart: 0,
					GridColumnEnd:   1,
					Height:          110, // 2 * 50 + 1 * 10 gap
					Margin:          Uniform(5),
				},
			},
			// Item in row 2 with margin
			{
				Style: Style{
					GridRowStart:    2,
					GridRowEnd:      3,
					GridColumnStart: 0,
					GridColumnEnd:   1,
					Height:          50,
					Margin:          Uniform(5),
				},
			},
		},
	}

	constraints := Loose(100, Unbounded)
	Layout(root, constraints)

	item1 := root.Children[0]
	item2 := root.Children[1]

	// Calculate expected positions
	// Row 0: 0-50, Row 1: 60-110, Row 2: 120-170
	// Item 1 spans rows 0-1, so cellY = 0, cellHeight = 50 + 10 + 50 = 110
	// Item 1 Y = cellY + margin.Top = 0 + 5 = 5
	// Item 1 Height = cellHeight - margin.Top - margin.Bottom = 110 - 5 - 5 = 100
	// Item 1 bottom = 5 + 100 = 105
	// Item 1 bottom with margin = 105 + 5 = 110
	
	// Item 2 is in row 2, so cellY = 120 (row 0 + gap + row 1 + gap)
	// Item 2 Y = cellY + margin.Top = 120 + 5 = 125
	// Item 2 top with margin = 125 - 5 = 120
	
	// Gap between item 1 bottom (with margin) and item 2 top (with margin)
	// Should be exactly the row gap (10)
	item1BottomWithMargin := item1.Rect.Y + item1.Rect.Height + item1.Style.Margin.Bottom
	item2TopWithMargin := item2.Rect.Y - item2.Style.Margin.Top
	gap := item2TopWithMargin - item1BottomWithMargin

	expectedGap := 10.0 // GridRowGap

	if math.Abs(gap-expectedGap) > 0.01 {
		t.Errorf("Gap between spanning item and following row is incorrect: expected %.2f, got %.2f", expectedGap, gap)
		if gap > expectedGap {
			t.Errorf("  BUG: Gap is too large by %.2f - margin may be duplicated", gap-expectedGap)
		}
	}
}

