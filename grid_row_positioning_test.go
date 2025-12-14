package layout

import (
	"math"
	"testing"
)

func TestGridRowPositioning(t *testing.T) {
	// Test that grid rows are positioned correctly (not all on top of each other)
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(Px(100)),
				FixedTrack(Px(100)),
				FixedTrack(Px(100)),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(Px(100)),
				FixedTrack(Px(100)),
			},
		},
		Children: []*Node{
			// Row 0, Col 0
			{
				Style: Style{
					GridRowStart:    0,
					GridColumnStart: 0,
					Width:           Px(100),
					Height:          Px(100),
				},
			},
			// Row 1, Col 0
			{
				Style: Style{
					GridRowStart:    1,
					GridColumnStart: 0,
					Width:           Px(100),
					Height:          Px(100),
				},
			},
			// Row 2, Col 0
			{
				Style: Style{
					GridRowStart:    2,
					GridColumnStart: 0,
					Width:           Px(100),
					Height:          Px(100),
				},
			},
		},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, constraints, ctx)

	// First item should be at Y=0 (or padding if any)
	item0Y := root.Children[0].Rect.Y
	if item0Y < 0 {
		t.Errorf("Item 0 Y should be >= 0, got %.2f", item0Y)
	}

	// Second item should be at Y=100 (first row height) + gap if any
	item1Y := root.Children[1].Rect.Y
	expectedY1 := 100.0 // First row height
	if math.Abs(item1Y-expectedY1) > 0.1 {
		t.Errorf("Item 1 Y should be approximately %.2f (after first row), got %.2f", expectedY1, item1Y)
	}

	// Third item should be at Y=200 (first two rows) + gaps if any
	item2Y := root.Children[2].Rect.Y
	expectedY2 := 200.0 // First two rows
	if math.Abs(item2Y-expectedY2) > 0.1 {
		t.Errorf("Item 2 Y should be approximately %.2f (after first two rows), got %.2f", expectedY2, item2Y)
	}

	// Verify items are not all at the same Y position
	if math.Abs(item0Y-item1Y) < 1.0 {
		t.Errorf("Item 0 and Item 1 should be at different Y positions, but both are at %.2f", item0Y)
	}
	if math.Abs(item1Y-item2Y) < 1.0 {
		t.Errorf("Item 1 and Item 2 should be at different Y positions, but both are at %.2f", item1Y)
	}
}

func TestGridRowPositioningWithGap(t *testing.T) {
	// Test row positioning with gaps
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(Px(100)),
				FixedTrack(Px(100)),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(Px(100)),
			},
			GridGap: Px(20),
		},
		Children: []*Node{
			{
				Style: Style{
					GridRowStart:    0,
					GridColumnStart: 0,
					Width:           Px(100),
					Height:          Px(100),
				},
			},
			{
				Style: Style{
					GridRowStart:    1,
					GridColumnStart: 0,
					Width:           Px(100),
					Height:          Px(100),
				},
			},
		},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, constraints, ctx)

	// Second item should be at Y=100 (first row) + 20 (gap)
	item1Y := root.Children[1].Rect.Y
	expectedY1 := 120.0 // First row (100) + gap (20)
	if math.Abs(item1Y-expectedY1) > 0.1 {
		t.Errorf("Item 1 Y should be approximately %.2f (first row + gap), got %.2f", expectedY1, item1Y)
	}
}

func TestGridRowPositioningWithAutoRows(t *testing.T) {
	// Test that auto rows get proper heights and positioning
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(Px(100)),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(Px(100)),
			},
			GridAutoRows: AutoTrack(),
		},
		Children: []*Node{
			// Row 0 (fixed)
			{
				Style: Style{
					GridRowStart:    0,
					GridColumnStart: 0,
					Width:           Px(100),
					Height:          Px(100),
				},
			},
			// Row 1 (auto - should size based on content)
			{
				Style: Style{
					GridRowStart:    1,
					GridColumnStart: 0,
					Width:           Px(100),
					Height:          Px(150), // Taller content
				},
			},
		},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, constraints, ctx)

	// First item should be at Y=0
	item0Y := root.Children[0].Rect.Y
	if item0Y < 0 {
		t.Errorf("Item 0 Y should be >= 0, got %.2f", item0Y)
	}

	// Second item should be below first row
	item1Y := root.Children[1].Rect.Y
	expectedY1 := 100.0 // First row height
	if math.Abs(item1Y-expectedY1) > 1.0 {
		t.Errorf("Item 1 Y should be approximately %.2f (after first row), got %.2f", expectedY1, item1Y)
	}

	// Verify they're not stacked
	if math.Abs(item0Y-item1Y) < 1.0 {
		t.Errorf("Items should not be stacked - Item 0 at %.2f, Item 1 at %.2f", item0Y, item1Y)
	}
}

func TestGridRowPositioningWithPadding(t *testing.T) {
	// Test that padding doesn't break row positioning
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(Px(100)),
				FixedTrack(Px(100)),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(Px(100)),
			},
			Padding: Uniform(Px(10)),
		},
		Children: []*Node{
			{
				Style: Style{
					GridRowStart:    0,
					GridColumnStart: 0,
					Width:           Px(100),
					Height:          Px(100),
				},
			},
			{
				Style: Style{
					GridRowStart:    1,
					GridColumnStart: 0,
					Width:           Px(100),
					Height:          Px(100),
				},
			},
		},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, constraints, ctx)

	// Items should still be positioned correctly relative to each other
	item0Y := root.Children[0].Rect.Y
	item1Y := root.Children[1].Rect.Y

	// Second item should be 100px below first (row height)
	expectedDiff := 100.0
	actualDiff := item1Y - item0Y
	if math.Abs(actualDiff-expectedDiff) > 0.1 {
		t.Errorf("Items should be %.2f apart, but are %.2f apart (Item 0: %.2f, Item 1: %.2f)",
			expectedDiff, actualDiff, item0Y, item1Y)
	}
}
