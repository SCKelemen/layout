package layout

import (
	"math"
	"testing"
)

// TestFlexboxSingleChildJustifyContent is a regression test asserting that
// flex containers with exactly one child do not divide by zero when
// computing space-between / space-around / space-evenly free space. The
// underlying spacing calculation divides by len(line)-1, len(line) or
// len(line)+1 respectively; a single-child line must produce finite
// coordinates and a normal (non-Inf, non-NaN) Rect.X / Rect.Width.
func TestFlexboxSingleChildJustifyContent(t *testing.T) {
	cases := []struct {
		name    string
		justify JustifyContent
	}{
		{"SpaceBetween", JustifyContentSpaceBetween},
		{"SpaceAround", JustifyContentSpaceAround},
		{"SpaceEvenly", JustifyContentSpaceEvenly},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			child := Fixed(50, 50)
			container := &Node{
				Style: Style{
					Display:        DisplayFlex,
					FlexDirection:  FlexDirectionRow,
					JustifyContent: tc.justify,
					Width:          Px(400),
					Height:         Px(100),
				},
				Children: []*Node{child},
			}

			LayoutSimple(container, Tight(400, 100))

			if math.IsNaN(child.Rect.X) || math.IsInf(child.Rect.X, 0) {
				t.Fatalf("child.Rect.X is not finite: %v", child.Rect.X)
			}
			if math.IsNaN(child.Rect.Width) || math.IsInf(child.Rect.Width, 0) {
				t.Fatalf("child.Rect.Width is not finite: %v", child.Rect.Width)
			}
			if math.IsNaN(child.Rect.Y) || math.IsInf(child.Rect.Y, 0) {
				t.Fatalf("child.Rect.Y is not finite: %v", child.Rect.Y)
			}
			if math.IsNaN(child.Rect.Height) || math.IsInf(child.Rect.Height, 0) {
				t.Fatalf("child.Rect.Height is not finite: %v", child.Rect.Height)
			}
		})
	}
}
