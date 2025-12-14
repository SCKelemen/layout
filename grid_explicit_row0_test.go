package layout

import (
	"math"
	"testing"
)

func TestGridExplicitRow0(t *testing.T) {
	// Test that explicitly setting row 0 still works (not treated as auto)
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
		},
		Children: []*Node{
			// Explicitly set to row 0
			{
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      1, // Explicitly set end
					GridColumnStart: 0,
					Width: Px(100),
					Height: Px(100),
				},
			},
			// Explicitly set to row 1
			{
				Style: Style{
					GridRowStart:    1,
					GridRowEnd:      2, // Explicitly set end
					GridColumnStart: 0,
					Width: Px(100),
					Height: Px(100),
				},
			},
		},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, constraints, ctx)

	// Items should be in different rows
	item0Y := root.Children[0].Rect.Y
	item1Y := root.Children[1].Rect.Y

	// Second item should be below first
	if item1Y <= item0Y {
		t.Errorf("Item 1 should be below Item 0, but Item 1 Y (%.2f) <= Item 0 Y (%.2f)",
			item1Y, item0Y)
	}

	// Second item should be at Y=100 (first row height)
	expectedY1 := 100.0
	if math.Abs(item1Y-expectedY1) > 0.1 {
		t.Errorf("Item 1 Y should be approximately %.2f, got %.2f", expectedY1, item1Y)
	}
}

func TestGridAutoPlacementVsExplicit0(t *testing.T) {
	// Test that we can distinguish between auto-placement and explicit row 0
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(Px(100)),
				FixedTrack(Px(100)),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(Px(100)),
				FixedTrack(Px(100)),
			},
		},
		Children: []*Node{
			// Auto-placed (no GridRowStart set, defaults to 0)
			{Style: Style{Width: Px(100), Height: Px(100)}},
			// Auto-placed
			{Style: Style{Width: Px(100), Height: Px(100)}},
			// Auto-placed
			{Style: Style{Width: Px(100), Height: Px(100)}},
			// Auto-placed
			{Style: Style{Width: Px(100), Height: Px(100)}},
		},
	}

	constraints := Loose(500, 500)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, constraints, ctx)

	// Items should be placed in a 2x2 grid
	// Row 0: items 0, 1
	// Row 1: items 2, 3

	// Items in same row should have same Y
	if math.Abs(root.Children[0].Rect.Y-root.Children[1].Rect.Y) > 0.1 {
		t.Errorf("Items 0 and 1 should be in same row, got Y: %.2f and %.2f",
			root.Children[0].Rect.Y, root.Children[1].Rect.Y)
	}

	// Items in different rows should have different Y
	if math.Abs(root.Children[0].Rect.Y-root.Children[2].Rect.Y) < 1.0 {
		t.Errorf("Items 0 and 2 should be in different rows, but both at Y=%.2f",
			root.Children[0].Rect.Y)
	}

	// Second row should be below first row
	if root.Children[2].Rect.Y <= root.Children[0].Rect.Y {
		t.Errorf("Second row should be below first, but Item 2 Y (%.2f) <= Item 0 Y (%.2f)",
			root.Children[2].Rect.Y, root.Children[0].Rect.Y)
	}
}

