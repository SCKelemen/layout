package layout

import (
	"testing"
)

// TestGridRowHeightWithOverlappingSpans tests that when items span overlapping rows,
// the row heights are correctly calculated and items are positioned correctly.
func TestGridRowHeightWithOverlappingSpans(t *testing.T) {
	// This test reproduces the issue from the debug case
	// Child 6: row 2-3, height 200
	// Child 7: row 3-6, height 257.07
	// Row 3 should have a consistent height, and child 7 should start right after child 6 ends + gap

	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				AutoTrack(), // row 0
				AutoTrack(), // row 1
				AutoTrack(), // row 2
				AutoTrack(), // row 3
				AutoTrack(), // row 4
				AutoTrack(), // row 5
			},
			GridTemplateColumns: []GridTrack{
				FractionTrack(1),
			},
			GridRowGap: Px(8),
			Width:      Px(1000),
		},
		Children: []*Node{
			// Child 1: spans rows 1-3, height 416.44 (like in the debug case)
			{
				Style: Style{
					GridRowStart: 1,
					GridRowEnd:   3,
					MinHeight:    Px(416.44),
				},
			},
			// Child 6: spans rows 2-3, height 200
			{
				Style: Style{
					GridRowStart: 2,
					GridRowEnd:   3,
					MinHeight:    Px(200),
				},
			},
			// Child 7: spans rows 3-6, height 257.07
			{
				Style: Style{
					GridRowStart: 3,
					GridRowEnd:   6,
					MinHeight:    Px(257.07),
				},
			},
		},
	}

	constraints := Loose(1000, Unbounded)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, constraints, ctx)

	_ = root.Children[0]       // row 1-3 (child1, used for row height calculation)
	child6 := root.Children[1] // row 2-3
	child7 := root.Children[2] // row 3-6

	// Child 6 should end at: end of row 3
	// Child 7 should start at: start of row 3
	// But wait, child 6 spans row 2-3, so it ends at the end of row 3
	// Child 7 spans row 3-6, so it starts at the start of row 3
	// They should overlap in row 3, or child 7 should start right after child 6 ends

	// Actually, in CSS Grid, items in the same row can overlap
	// But the gap between rows should be consistent

	// The gap between the end of row 2 and start of row 3 should be 8px
	// Child 6 ends at the end of row 3
	// Child 7 starts at the start of row 3
	// So there should be no gap between them (they're both in row 3)

	// But wait, child 6 spans row 2-3, so it fills: row2 + gap + row3
	// Child 7 spans row 3-6, so it starts at: end of row 2 + gap

	// Let's calculate expected positions:
	// Row 1: from child1 = (416.44 - 16) / 3 = 133.48
	// Row 2: max(child1 portion=133.48, child6 portion=(200-8)/2=96) = 133.48
	// Row 3: max(child1 portion=133.48, child6 portion=96, child7 portion=(257.07-16)/3=80.36) = 133.48

	// Child 6 cell: row2 + gap + row3 = 133.48 + 8 + 133.48 = 274.96
	// Child 6 should have height = 274.96 (filling its cell)

	// Child 6 starts at: row0 + gap + row1 = 0 + 8 + 133.48 = 141.48 (if row0=0)
	// Actually, row 0 has no items, so row0 = 0

	// Let me check the actual positions
	child6Bottom := child6.Rect.Y + child6.Rect.Height
	child7Top := child7.Rect.Y

	// Child 6 and child 7 both involve row 3
	// Child 6 ends at the end of row 3
	// Child 7 starts at the start of row 3
	// In CSS Grid, items in the same row can overlap, so child7.top could be < child6.bottom

	// But the gap between row 2 and row 3 should be 8px
	// If child 6 ends at the end of row 3, and child 7 starts at the start of row 3,
	// then child7.top should be < child6.bottom (they overlap in row 3)

	// Actually, I think the issue is different. Let me check if child 6 and child 7
	// are positioned correctly relative to the row boundaries.

	// The key is: child 6 spans row 2-3, so it should end at: start of row 2 + row2 + gap + row3
	// Child 7 spans row 3-6, so it should start at: start of row 3 = start of row 2 + row2 + gap

	// So child 7 should start exactly where row 3 starts
	// And child 6 should end exactly where row 3 ends
	// The gap between them should be 0 (they're both in row 3)

	// But if there's a gap, it means row 3's height is inconsistent

	// Debug output
	t.Logf("Child 6: row 2-3, y=%.2f, height=%.2f, bottom=%.2f",
		child6.Rect.Y, child6.Rect.Height, child6Bottom)
	t.Logf("Child 7: row 3-6, y=%.2f, height=%.2f, top=%.2f",
		child7.Rect.Y, child7.Rect.Height, child7Top)

	// In CSS Grid, items in the same row can overlap
	// Child 6 spans row 2-3, so it ends at the end of row 3
	// Child 7 spans row 3-6, so it starts at the start of row 3
	// They should overlap in row 3, or child 7 should start exactly where row 3 starts
	// and child 6 should end exactly where row 3 ends

	// The gap between the end of row 2 and start of row 3 should be 8px
	// But child 6 ends at the end of row 3, and child 7 starts at the start of row 3
	// So if they don't overlap, there's an issue with row height calculation

	if child7Top < child6Bottom {
		t.Logf("Child 7 starts before child 6 ends (they overlap in row 3): child6.bottom=%.2f, child7.top=%.2f",
			child6Bottom, child7Top)
		// This is actually correct - they overlap in row 3
	} else {
		gap := child7Top - child6Bottom
		t.Logf("Gap between child 6 and child 7: %.2f (expected: 0 or 8)", gap)
		// If there's a gap, it should be 0 (they're both in row 3) or 8 (if there's a row boundary issue)
		// But 12.22 suggests a calculation error
		if gap > 10.0 {
			t.Errorf("Gap between child 6 and child 7 is too large: got %.2f (expected ~0 or ~8)", gap)
		}
	}
}
