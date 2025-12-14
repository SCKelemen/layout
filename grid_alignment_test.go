package layout

import (
	"math"
	"testing"
)

// TestGridJustifyItems tests justify-items alignment along the inline (row) axis
func TestGridJustifyItems(t *testing.T) {
	tests := []struct {
		name         string
		justifyItems JustifyItems
		itemWidth    float64
		cellWidth    float64
		expectedX    float64
	}{
		{
			name:         "stretch (default)",
			justifyItems: JustifyItemsStretch,
			itemWidth:    200, // Will be stretched to cell width
			cellWidth:    300,
			expectedX:    0, // Start of cell + margin
		},
		{
			name:         "start",
			justifyItems: JustifyItemsStart,
			itemWidth:    100,
			cellWidth:    300,
			expectedX:    0, // Start of cell
		},
		{
			name:         "end",
			justifyItems: JustifyItemsEnd,
			itemWidth:    100,
			cellWidth:    300,
			expectedX:    200, // cellWidth - itemWidth
		},
		{
			name:         "center",
			justifyItems: JustifyItemsCenter,
			itemWidth:    100,
			cellWidth:    300,
			expectedX:    100, // (cellWidth - itemWidth) / 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Node{
				Style: Style{
					Display: DisplayGrid,
					GridTemplateRows: []GridTrack{
						AutoTrack(),
					},
					GridTemplateColumns: []GridTrack{
						FixedTrack(Px(tt.cellWidth)),
					},
					JustifyItems: tt.justifyItems,
					Width:        Px(tt.cellWidth),
				},
				Children: []*Node{
					{
						Style: Style{
							Width: Px(tt.itemWidth),
						},
					},
				},
			}

			constraints := Loose(tt.cellWidth, Unbounded)
			ctx := NewLayoutContext(800, 600, 16)
			LayoutGrid(root, constraints, ctx)

			item := root.Children[0]
			actualX := item.Rect.X - root.Style.Padding.Left.Value - root.Style.Border.Left.Value

			if math.Abs(actualX-tt.expectedX) > 1.0 {
				t.Errorf("Expected X position %.2f, got %.2f", tt.expectedX, actualX)
			}

			// For stretch, item should fill cell width
			if tt.justifyItems == JustifyItemsStretch {
				expectedWidth := tt.cellWidth
				if math.Abs(item.Rect.Width-expectedWidth) > 1.0 {
					t.Errorf("Stretch: expected width %.2f, got %.2f", expectedWidth, item.Rect.Width)
				}
			} else {
				// For non-stretch, item should use its specified width
				if math.Abs(item.Rect.Width-tt.itemWidth) > 1.0 {
					t.Errorf("Non-stretch: expected width %.2f, got %.2f", tt.itemWidth, item.Rect.Width)
				}
			}
		})
	}
}

// TestGridAlignItems tests align-items alignment along the block (column) axis
func TestGridAlignItems(t *testing.T) {
	tests := []struct {
		name       string
		alignItems AlignItems
		itemHeight float64
		cellHeight float64
		expectedY  float64
	}{
		{
			name:       "stretch (default)",
			alignItems: AlignItemsStretch,
			itemHeight: 200, // Will be stretched to cell height
			cellHeight: 300,
			expectedY:  0, // Start of cell + margin
		},
		{
			name:       "start",
			alignItems: AlignItemsFlexStart,
			itemHeight: 100,
			cellHeight: 300,
			expectedY:  0, // Start of cell
		},
		{
			name:       "end",
			alignItems: AlignItemsFlexEnd,
			itemHeight: 100,
			cellHeight: 300,
			expectedY:  200, // cellHeight - itemHeight
		},
		{
			name:       "center",
			alignItems: AlignItemsCenter,
			itemHeight: 100,
			cellHeight: 300,
			expectedY:  100, // (cellHeight - itemHeight) / 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Node{
				Style: Style{
					Display: DisplayGrid,
					GridTemplateRows: []GridTrack{
						FixedTrack(Px(tt.cellHeight)),
					},
					GridTemplateColumns: []GridTrack{
						AutoTrack(),
					},
					AlignItems: tt.alignItems,
					Height:     Px(tt.cellHeight),
				},
				Children: []*Node{
					{
						Style: Style{
							Height: Px(tt.itemHeight),
						},
					},
				},
			}

			constraints := Loose(Unbounded, tt.cellHeight)
			ctx := NewLayoutContext(800, 600, 16)
			LayoutGrid(root, constraints, ctx)

			item := root.Children[0]
			actualY := item.Rect.Y - root.Style.Padding.Top.Value - root.Style.Border.Top.Value

			if math.Abs(actualY-tt.expectedY) > 1.0 {
				t.Errorf("Expected Y position %.2f, got %.2f", tt.expectedY, actualY)
			}

			// For stretch, item should fill cell height
			if tt.alignItems == AlignItemsStretch {
				expectedHeight := tt.cellHeight
				if math.Abs(item.Rect.Height-expectedHeight) > 1.0 {
					t.Errorf("Stretch: expected height %.2f, got %.2f", expectedHeight, item.Rect.Height)
				}
			} else {
				// For non-stretch, item should use its specified height
				if math.Abs(item.Rect.Height-tt.itemHeight) > 1.0 {
					t.Errorf("Non-stretch: expected height %.2f, got %.2f", tt.itemHeight, item.Rect.Height)
				}
			}
		})
	}
}

// TestGridAspectRatioDefaultsToStart tests that items with aspect-ratio
// default to start alignment per CSS spec
func TestGridAspectRatioDefaultsToStart(t *testing.T) {
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(Px(300)),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(Px(300)),
			},
			// Not setting JustifyItems/AlignItems - should default to stretch
			// But aspect-ratio items should default to start
			Width:  Px(300),
			Height: Px(300),
		},
		Children: []*Node{
			{
				Style: Style{
					AspectRatio: 2.0, // width:height = 2:1
					// With cell 300x300 and aspect ratio 2:1, item should be 300x150
				},
			},
		},
	}

	constraints := Loose(300, 300)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, constraints, ctx)

	item := root.Children[0]

	// Item should maintain aspect ratio (not stretch)
	expectedWidth := 300.0
	expectedHeight := 150.0 // 300 / 2

	if math.Abs(item.Rect.Width-expectedWidth) > 1.0 {
		t.Errorf("Expected width %.2f, got %.2f", expectedWidth, item.Rect.Width)
	}
	if math.Abs(item.Rect.Height-expectedHeight) > 1.0 {
		t.Errorf("Expected height %.2f, got %.2f", expectedHeight, item.Rect.Height)
	}

	// Item should be aligned to start (top-left)
	expectedX := 0.0
	expectedY := 0.0
	actualX := item.Rect.X - root.Style.Padding.Left.Value - root.Style.Border.Left.Value
	actualY := item.Rect.Y - root.Style.Padding.Top.Value - root.Style.Border.Top.Value

	if math.Abs(actualX-expectedX) > 1.0 {
		t.Errorf("Expected X position %.2f (start), got %.2f", expectedX, actualX)
	}
	if math.Abs(actualY-expectedY) > 1.0 {
		t.Errorf("Expected Y position %.2f (start), got %.2f", expectedY, actualY)
	}
}

// TestGridAlignmentWithMargins tests that margins are properly accounted for in alignment
func TestGridAlignmentWithMargins(t *testing.T) {
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(Px(300)),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(Px(300)),
			},
			JustifyItems: JustifyItemsCenter,
			AlignItems:   AlignItemsCenter,
			Width:        Px(300),
			Height:       Px(300),
		},
		Children: []*Node{
			{
				Style: Style{
					Width:  Px(100),
					Height: Px(100),
					Margin: Uniform(Px(20)),
				},
			},
		},
	}

	constraints := Loose(300, 300)
	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(root, constraints, ctx)

	item := root.Children[0]

	// Item with 100x100 size and 20px margins should be centered
	// Cell is 300x300, item+margin box is 140x140 (100 + 20*2)
	// The item+margin box is centered at (300-140)/2 = 80 from cell start
	// The item itself starts at 80 + 20 (left margin) = 100 from cell start
	expectedX := 100.0 // 80 (centered box start) + 20 (left margin)
	expectedY := 100.0 // 80 (centered box start) + 20 (top margin)
	actualX := item.Rect.X - root.Style.Padding.Left.Value - root.Style.Border.Left.Value
	actualY := item.Rect.Y - root.Style.Padding.Top.Value - root.Style.Border.Top.Value

	if math.Abs(actualX-expectedX) > 1.0 {
		t.Errorf("Expected X position %.2f, got %.2f", expectedX, actualX)
	}
	if math.Abs(actualY-expectedY) > 1.0 {
		t.Errorf("Expected Y position %.2f, got %.2f", expectedY, actualY)
	}
}
