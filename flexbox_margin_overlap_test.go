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
				Height: Px(20),
				Margin: Uniform(Px(10)),
			},
		},
		{
			Style: Style{
				Height: Px(20),
				Margin: Uniform(Px(10)),
			},
		},
		{
			Style: Style{
				Height: Px(20),
				Margin: Uniform(Px(10)),
			},
		},
		{
			Style: Style{
				Height: Px(20),
				Margin: Uniform(Px(10)),
			},
		},
	}

	root := VStack(items...)
	root.Style.Width = Px(200)

	constraints := Loose(200, Unbounded)
	ctx := NewLayoutContext(1920, 1080, 16)
	Layout(root, constraints, ctx)

	// Check that first item is positioned correctly (at its top margin)
	firstItem := root.Children[0]
	if firstItem.Rect.Y != firstItem.Style.Margin.Top.Value {
		t.Errorf("First item Y position incorrect: expected %.2f (top margin), got %.2f",
			firstItem.Style.Margin.Top.Value, firstItem.Rect.Y)
	}

	// Check spacing between all items
	for i := 1; i < len(root.Children); i++ {
		prev := root.Children[i-1]
		curr := root.Children[i]

		prevBottom := prev.Rect.Y + prev.Rect.Height
		currTop := curr.Rect.Y
		gap := currTop - prevBottom
		expectedGap := prev.Style.Margin.Bottom.Value + curr.Style.Margin.Top.Value

		if math.Abs(gap-expectedGap) > 0.01 {
			t.Errorf("Gap between item %d and %d is incorrect: expected %.2f, got %.2f",
				i, i+1, expectedGap, gap)

			// Check for overlap
			if gap < expectedGap {
				overlap := expectedGap - gap
				t.Errorf("OVERLAP DETECTED: Item %d overlaps item %d by %.2f pixels",
					i+1, i, overlap)
				t.Errorf("  Previous item ends at: %.2f (with margin: %.2f)",
					prevBottom, prevBottom+prev.Style.Margin.Bottom.Value)
				t.Errorf("  Current item starts at: %.2f (should be at: %.2f)",
					currTop, prevBottom+prev.Style.Margin.Bottom.Value+curr.Style.Margin.Top.Value)
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
				MinHeight: Px(20),
				Margin:    Uniform(Px(10)),
			},
		},
		{
			Style: Style{
				MinHeight: Px(20),
				Margin:    Uniform(Px(10)),
			},
		},
		{
			Style: Style{
				MinHeight: Px(20),
				Margin:    Uniform(Px(10)),
			},
		},
	}

	root := VStack(items...)
	root.Style.Width = Px(200)

	constraints := Loose(200, Unbounded)
	ctx := NewLayoutContext(1920, 1080, 16)
	Layout(root, constraints, ctx)

	// Check spacing
	for i := 1; i < len(root.Children); i++ {
		prev := root.Children[i-1]
		curr := root.Children[i]

		prevBottom := prev.Rect.Y + prev.Rect.Height
		currTop := curr.Rect.Y
		gap := currTop - prevBottom
		expectedGap := prev.Style.Margin.Bottom.Value + curr.Style.Margin.Top.Value

		if math.Abs(gap-expectedGap) > 0.01 {
			t.Errorf("Gap between item %d and %d is incorrect: expected %.2f, got %.2f",
				i, i+1, expectedGap, gap)
		}
	}
}
