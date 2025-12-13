package layout

import (
	"math"
	"testing"
)

// TestFlexboxMarginOverlapBug tests for the reported bug where
// the second item overlaps the first item in a VStack, but
// subsequent items are correctly spaced.
func TestFlexboxMarginOverlapBug(t *testing.T) {
	// Create a VStack with multiple items that have margins
	// This should reproduce the bug where second item overlaps first
	items := []*Node{
		{
			Style: Style{
				Height: 20,
				Margin: Uniform(10),
			},
		},
		{
			Style: Style{
				Height: 20,
				Margin: Uniform(10),
			},
		},
		{
			Style: Style{
				Height: 20,
				Margin: Uniform(10),
			},
		},
		{
			Style: Style{
				Height: 20,
				Margin: Uniform(10),
			},
		},
	}

	root := VStack(items...)
	root.Style.Width = 200

	constraints := Loose(200, Unbounded)
	Layout(root, constraints)

	// Check that first item is positioned correctly (at its top margin)
	firstItem := root.Children[0]
	if firstItem.Rect.Y != firstItem.Style.Margin.Top {
		t.Errorf("First item Y position incorrect: expected %.2f (top margin), got %.2f",
			firstItem.Style.Margin.Top, firstItem.Rect.Y)
	}

	// Check spacing between all items
	for i := 1; i < len(root.Children); i++ {
		prev := root.Children[i-1]
		curr := root.Children[i]

		prevBottom := prev.Rect.Y + prev.Rect.Height
		currTop := curr.Rect.Y
		gap := currTop - prevBottom
		expectedGap := prev.Style.Margin.Bottom + curr.Style.Margin.Top

		if math.Abs(gap-expectedGap) > 0.01 {
			t.Errorf("Gap between item %d and %d is incorrect: expected %.2f, got %.2f",
				i, i+1, expectedGap, gap)

			// Check for overlap
			if gap < expectedGap {
				overlap := expectedGap - gap
				t.Errorf("OVERLAP DETECTED: Item %d overlaps item %d by %.2f pixels",
					i+1, i, overlap)
				t.Errorf("  Previous item ends at: %.2f (with margin: %.2f)",
					prevBottom, prevBottom+prev.Style.Margin.Bottom)
				t.Errorf("  Current item starts at: %.2f (should be at: %.2f)",
					currTop, prevBottom+prev.Style.Margin.Bottom+curr.Style.Margin.Top)
			}
		}
	}
}

// TestFlexboxMarginAutoHeight tests margins with auto-height items
// (like text nodes that might not have explicit height)
func TestFlexboxMarginAutoHeight(t *testing.T) {
	items := []*Node{
		{
			Style: Style{
				// Auto height, but has MinHeight
				MinHeight: 20,
				Margin:    Uniform(10),
			},
		},
		{
			Style: Style{
				MinHeight: 20,
				Margin:    Uniform(10),
			},
		},
		{
			Style: Style{
				MinHeight: 20,
				Margin:    Uniform(10),
			},
		},
	}

	root := VStack(items...)
	root.Style.Width = 200

	constraints := Loose(200, Unbounded)
	Layout(root, constraints)

	// Check spacing
	for i := 1; i < len(root.Children); i++ {
		prev := root.Children[i-1]
		curr := root.Children[i]

		prevBottom := prev.Rect.Y + prev.Rect.Height
		currTop := curr.Rect.Y
		gap := currTop - prevBottom
		expectedGap := prev.Style.Margin.Bottom + curr.Style.Margin.Top

		if math.Abs(gap-expectedGap) > 0.01 {
			t.Errorf("Gap between item %d and %d is incorrect: expected %.2f, got %.2f",
				i, i+1, expectedGap, gap)
		}
	}
}
