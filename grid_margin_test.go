package layout

import (
	"math"
	"testing"
)

func TestGridMargin(t *testing.T) {
	// Test that margins work in grid layout
	root := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(100),
				FixedTrack(100),
			},
			GridTemplateColumns: []GridTrack{
				FixedTrack(100),
				FixedTrack(100),
			},
		},
		Children: []*Node{
			{
				Style: Style{
					GridRowStart:    0,
					GridColumnStart: 0,
					Margin:         Uniform(10),
				},
			},
			{
				Style: Style{
					GridRowStart:    0,
					GridColumnStart: 1,
					Margin:         Uniform(10),
				},
			},
		},
	}

	constraints := Loose(300, 300)
	LayoutGrid(root, constraints)

	// First item should have margin applied
	if math.Abs(root.Children[0].Rect.X-10.0) > 0.1 {
		t.Errorf("First item X should be 10 (margin), got %.2f", root.Children[0].Rect.X)
	}
	if math.Abs(root.Children[0].Rect.Y-10.0) > 0.1 {
		t.Errorf("First item Y should be 10 (margin), got %.2f", root.Children[0].Rect.Y)
	}

	// Item size should account for margins
	// Grid cell is 100x100, margins are 10 on each side, so item should be 80x80
	expectedWidth := 100.0 - 10.0 - 10.0
	expectedHeight := 100.0 - 10.0 - 10.0
	if math.Abs(root.Children[0].Rect.Width-expectedWidth) > 0.1 {
		t.Errorf("First item width should be %.2f (cell - margins), got %.2f", expectedWidth, root.Children[0].Rect.Width)
	}
	if math.Abs(root.Children[0].Rect.Height-expectedHeight) > 0.1 {
		t.Errorf("First item height should be %.2f (cell - margins), got %.2f", expectedHeight, root.Children[0].Rect.Height)
	}
}

