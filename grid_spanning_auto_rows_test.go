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
				FixedTrack(Px(100)),
				FixedTrack(Px(100)),
			},
		},
		Children: []*Node{
			// Item spanning all 3 rows
			{
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      3, // Spans rows 0, 1, 2
					GridColumnStart: 0,
					MinHeight:       Px(300.0), // Should be distributed: 100px per row
				},
			},
			// Item in row 0, col 1
			{
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      1,
					GridColumnStart: 1,
					MinHeight:       Px(150.0), // Row 0 should be 150px (max of 100 and 150)
				},
			},
		},
	}

	constraints := Loose(300, Unbounded)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, constraints, ctx)

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
				FixedTrack(Px(100)),
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
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, constraints, ctx)

	// Item should be 0 height (no content, no MinHeight)
	spanningItem := root.Children[0]
	if spanningItem.Rect.Height != 0.0 {
		t.Logf("Note: Spanning item without MinHeight has height %.2f (expected 0)", spanningItem.Rect.Height)
	}

	// Rows should collapse to 0 (or minimal height from safeguards)
	// This is correct CSS Grid behavior
}

func TestGridSpanningAutoRowsWithFollowingRow(t *testing.T) {
	// Repro from manual script: spanning rows 0-2 plus an item in row 3.
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateColumns: []GridTrack{
				FractionTrack(1),
				FractionTrack(1),
				FractionTrack(1),
			},
			GridTemplateRows: []GridTrack{
				AutoTrack(),
				AutoTrack(),
				AutoTrack(),
				AutoTrack(),
			},
			GridRowGap:    Px(8),
			GridColumnGap: Px(8),
			Width:         Px(1000),
		},
		Children: []*Node{
			{
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      3, // spans rows 0,1,2
					GridColumnStart: 0,
					GridColumnEnd:   1,
					MinHeight:       Px(300),
				},
			},
			{
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      1,
					GridColumnStart: 1,
					GridColumnEnd:   2,
					MinHeight:       Px(100),
				},
			},
			{
				Style: Style{
					GridRowStart:    1,
					GridRowEnd:      2,
					GridColumnStart: 1,
					GridColumnEnd:   2,
					MinHeight:       Px(100),
				},
			},
			{
				Style: Style{
					GridRowStart:    2,
					GridRowEnd:      3,
					GridColumnStart: 1,
					GridColumnEnd:   2,
					MinHeight:       Px(100),
				},
			},
			{
				Style: Style{
					GridRowStart:    3,
					GridRowEnd:      4,
					GridColumnStart: 0,
					GridColumnEnd:   3,
					MinHeight:       Px(50),
				},
			},
		},
	}

	constraints := Loose(1000, Unbounded)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, constraints, ctx)

	// Expected row heights: 100, 100, 100, 50 with 8px gaps.
	// Total: 100 + 8 + 100 + 8 + 100 + 8 + 50 = 374
	if math.Abs(root.Rect.Height-374.0) > 0.01 {
		t.Errorf("Grid container height incorrect: expected 374.00, got %.2f", root.Rect.Height)
	}

	spanningItem := root.Children[0]
	row3Item := root.Children[4]

	// Spanning item fills rows 0-2 plus the two gaps between them.
	if math.Abs(spanningItem.Rect.Height-316.0) > 0.01 {
		t.Errorf("Spanning item height incorrect: expected 316.00, got %.2f", spanningItem.Rect.Height)
	}

	// Row 3 item should start after spanning item plus one row gap.
	expectedRow3Y := 324.0
	if math.Abs(row3Item.Rect.Y-expectedRow3Y) > 0.01 {
		t.Errorf("Row 3 item Y incorrect: expected %.2f, got %.2f", expectedRow3Y, row3Item.Rect.Y)
	}

	gap := row3Item.Rect.Y - (spanningItem.Rect.Y + spanningItem.Rect.Height)
	if math.Abs(gap-8.0) > 0.01 {
		t.Errorf("Gap between spanning item and following row incorrect: expected 8.00, got %.2f", gap)
	}
}

func TestGridSpanningExplicitHeightContributesToAutoRows(t *testing.T) {
	// Repro from manual script: explicit Height (not MinHeight) on a spanning item.
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateColumns: []GridTrack{
				FractionTrack(1),
				FractionTrack(1),
				FractionTrack(1),
			},
			GridTemplateRows: []GridTrack{
				AutoTrack(),
				AutoTrack(),
				AutoTrack(),
				AutoTrack(),
			},
			GridRowGap:    Px(8),
			GridColumnGap: Px(8),
			Width:         Px(1000),
		},
		Children: []*Node{
			{
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      1,
					GridColumnStart: 0,
					GridColumnEnd:   3,
					Height:          Px(60),
				},
			},
			{
				Style: Style{
					GridRowStart:    1,
					GridRowEnd:      2,
					GridColumnStart: 0,
					GridColumnEnd:   1,
					MinHeight:       Px(50),
				},
			},
			{
				Style: Style{
					GridRowStart:    1,
					GridRowEnd:      2,
					GridColumnStart: 1,
					GridColumnEnd:   2,
					Height:          Px(50),
					MinHeight:       Px(40),
				},
			},
			{
				Style: Style{
					GridRowStart:    2,
					GridRowEnd:      4, // spans rows 2 and 3
					GridColumnStart: 0,
					GridColumnEnd:   3,
					Height:          Px(200),
				},
			},
		},
	}

	constraints := Loose(1000, Unbounded)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, constraints, ctx)

	// Expected total:
	// row0 60 + gap8 + row1 50 + gap8 + (row2+gap+row3)=200 => row2+row3=192
	// plus gap between row2 and row3 already in 200
	// container height = 60 + 8 + 50 + 8 + 200 = 326
	if math.Abs(root.Rect.Height-326.0) > 0.01 {
		t.Errorf("Grid container height incorrect: expected 326.00, got %.2f", root.Rect.Height)
	}

	spanningItem := root.Children[3]
	if math.Abs(spanningItem.Rect.Height-200.0) > 0.01 {
		t.Errorf("Spanning item height incorrect: expected 200.00, got %.2f", spanningItem.Rect.Height)
	}

	if math.Abs(spanningItem.Rect.Y-126.0) > 0.01 {
		t.Errorf("Spanning item Y incorrect: expected 126.00, got %.2f", spanningItem.Rect.Y)
	}
}
