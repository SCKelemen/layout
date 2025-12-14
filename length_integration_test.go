package layout

import (
	"testing"
)

// TestWidthWithAllUnits tests Width property with all available length units
func TestWidthWithAllUnits(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)

	tests := []struct {
		name     string
		width    Length
		expected float64
	}{
		{"Px", Px(200), 200},
		{"Em", Em(10), 160},     // 10 * 16 = 160
		{"Rem", Rem(10), 160},   // 10 * 16 = 160
		{"Vh", Vh(50), 400},     // 50% of 800 = 400
		{"Vw", Vw(50), 500},     // 50% of 1000 = 500
		{"Vmin", Vmin(50), 400}, // 50% of min(1000, 800) = 50% of 800 = 400
		{"Vmax", Vmax(50), 500}, // 50% of max(1000, 800) = 50% of 1000 = 500
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{
				Style: Style{
					Display: DisplayBlock,
					Width:   tt.width,
					Height:  Px(100),
				},
			}

			constraints := Loose(1000, 800)
			size := Layout(node, constraints, ctx)

			if size.Width != tt.expected {
				t.Errorf("Width with %s: got %.2f, want %.2f", tt.name, size.Width, tt.expected)
			}
		})
	}
}

// TestHeightWithAllUnits tests Height property with all available length units
func TestHeightWithAllUnits(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)

	tests := []struct {
		name     string
		height   Length
		expected float64
	}{
		{"Px", Px(200), 200},
		{"Em", Em(10), 160},     // 10 * 16 = 160
		{"Rem", Rem(10), 160},   // 10 * 16 = 160
		{"Vh", Vh(50), 400},     // 50% of 800 = 400
		{"Vw", Vw(50), 500},     // 50% of 1000 = 500
		{"Vmin", Vmin(50), 400}, // 50% of min(1000, 800) = 50% of 800 = 400
		{"Vmax", Vmax(50), 500}, // 50% of max(1000, 800) = 50% of 1000 = 500
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(100),
					Height:  tt.height,
				},
			}

			constraints := Loose(1000, 800)
			size := Layout(node, constraints, ctx)

			if size.Height != tt.expected {
				t.Errorf("Height with %s: got %.2f, want %.2f", tt.name, size.Height, tt.expected)
			}
		})
	}
}

// TestMinMaxWidthWithUnits tests MinWidth and MaxWidth with various units
func TestMinMaxWidthWithUnits(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)

	t.Run("MinWidth constrains smaller width", func(t *testing.T) {
		node := &Node{
			Style: Style{
				Display:  DisplayBlock,
				Width:    Px(50),
				MinWidth: Em(10), // 10 * 16 = 160
				Height:   Px(100),
			},
		}

		constraints := Loose(1000, 800)
		size := Layout(node, constraints, ctx)

		if size.Width != 160 {
			t.Errorf("MinWidth (Em): got %.2f, want 160", size.Width)
		}
	})

	t.Run("MaxWidth constrains larger width", func(t *testing.T) {
		node := &Node{
			Style: Style{
				Display:  DisplayBlock,
				Width:    Px(500),
				MaxWidth: Vw(20), // 20% of 1000 = 200
				Height:   Px(100),
			},
		}

		constraints := Loose(1000, 800)
		size := Layout(node, constraints, ctx)

		if size.Width != 200 {
			t.Errorf("MaxWidth (Vw): got %.2f, want 200", size.Width)
		}
	})
}

// TestMinMaxHeightWithUnits tests MinHeight and MaxHeight with various units
func TestMinMaxHeightWithUnits(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)

	t.Run("MinHeight constrains smaller height", func(t *testing.T) {
		node := &Node{
			Style: Style{
				Display:   DisplayBlock,
				Width:     Px(100),
				Height:    Px(50),
				MinHeight: Vh(25), // 25% of 800 = 200
			},
		}

		constraints := Loose(1000, 800)
		size := Layout(node, constraints, ctx)

		if size.Height != 200 {
			t.Errorf("MinHeight (Vh): got %.2f, want 200", size.Height)
		}
	})

	t.Run("MaxHeight constrains larger height", func(t *testing.T) {
		node := &Node{
			Style: Style{
				Display:   DisplayBlock,
				Width:     Px(100),
				Height:    Px(500),
				MaxHeight: Rem(10), // 10 * 16 = 160
			},
		}

		constraints := Loose(1000, 800)
		size := Layout(node, constraints, ctx)

		if size.Height != 160 {
			t.Errorf("MaxHeight (Rem): got %.2f, want 160", size.Height)
		}
	})
}

// TestPaddingWithAllUnits tests Padding with various units
func TestPaddingWithAllUnits(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)

	tests := []struct {
		name           string
		padding        Spacing
		expectedWidth  float64
		expectedHeight float64
		contentWidth   float64
		contentHeight  float64
	}{
		{
			name:           "Px padding",
			padding:        Uniform(Px(10)),
			expectedWidth:  120, // 100 + 10*2
			expectedHeight: 120, // 100 + 10*2
			contentWidth:   100,
			contentHeight:  100,
		},
		{
			name:           "Em padding",
			padding:        Uniform(Em(1)), // 1 * 16 = 16
			expectedWidth:  132,            // 100 + 16*2
			expectedHeight: 132,            // 100 + 16*2
			contentWidth:   100,
			contentHeight:  100,
		},
		{
			name:           "Rem padding",
			padding:        Uniform(Rem(2)), // 2 * 16 = 32
			expectedWidth:  164,             // 100 + 32*2
			expectedHeight: 164,             // 100 + 32*2
			contentWidth:   100,
			contentHeight:  100,
		},
		{
			name: "Mixed unit padding",
			padding: Spacing{
				Top:    Vh(2.5),  // 2.5% of 800 = 20
				Right:  Em(1),    // 1 * 16 = 16
				Bottom: Px(10),   // 10
				Left:   Rem(0.5), // 0.5 * 16 = 8
			},
			expectedWidth:  124, // 100 + 16 + 8
			expectedHeight: 130, // 100 + 20 + 10
			contentWidth:   100,
			contentHeight:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(tt.contentWidth),
					Height:  Px(tt.contentHeight),
					Padding: tt.padding,
				},
			}

			constraints := Loose(1000, 800)
			size := Layout(node, constraints, ctx)

			if size.Width != tt.expectedWidth {
				t.Errorf("Width: got %.2f, want %.2f", size.Width, tt.expectedWidth)
			}
			if size.Height != tt.expectedHeight {
				t.Errorf("Height: got %.2f, want %.2f", size.Height, tt.expectedHeight)
			}
		})
	}
}

// TestMarginWithAllUnits tests Margin with various units in flexbox
func TestMarginWithAllUnits(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)

	tests := []struct {
		name           string
		margin         Spacing
		expectedChildX float64
		expectedChildY float64
	}{
		{
			name:           "Px margin",
			margin:         Spacing{Top: Px(20), Left: Px(30)},
			expectedChildX: 30,
			expectedChildY: 20,
		},
		{
			name:           "Em margin",
			margin:         Spacing{Top: Em(1), Left: Em(2)}, // 16, 32
			expectedChildX: 32,
			expectedChildY: 16,
		},
		{
			name:           "Vh/Vw margin",
			margin:         Spacing{Top: Vh(5), Left: Vw(4)}, // 5% of 800 = 40, 4% of 1000 = 40
			expectedChildX: 40,
			expectedChildY: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Node{
				Style: Style{
					Display: DisplayFlex,
					Width:   Px(500),
					Height:  Px(400),
				},
				Children: []*Node{
					{
						Style: Style{
							Width:  Px(100),
							Height: Px(100),
							Margin: tt.margin,
						},
					},
				},
			}

			constraints := Loose(1000, 800)
			Layout(root, constraints, ctx)

			child := root.Children[0]
			if child.Rect.X != tt.expectedChildX {
				t.Errorf("Child X: got %.2f, want %.2f", child.Rect.X, tt.expectedChildX)
			}
			if child.Rect.Y != tt.expectedChildY {
				t.Errorf("Child Y: got %.2f, want %.2f", child.Rect.Y, tt.expectedChildY)
			}
		})
	}
}

// TestBorderWithAllUnits tests Border with various units
func TestBorderWithAllUnits(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)

	tests := []struct {
		name           string
		border         Spacing
		expectedWidth  float64
		expectedHeight float64
		contentWidth   float64
		contentHeight  float64
	}{
		{
			name:           "Px border",
			border:         Uniform(Px(5)),
			expectedWidth:  110, // 100 + 5*2
			expectedHeight: 110, // 100 + 5*2
			contentWidth:   100,
			contentHeight:  100,
		},
		{
			name:           "Em border",
			border:         Uniform(Em(0.5)), // 0.5 * 16 = 8
			expectedWidth:  116,              // 100 + 8*2
			expectedHeight: 116,              // 100 + 8*2
			contentWidth:   100,
			contentHeight:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(tt.contentWidth),
					Height:  Px(tt.contentHeight),
					Border:  tt.border,
				},
			}

			constraints := Loose(1000, 800)
			size := Layout(node, constraints, ctx)

			if size.Width != tt.expectedWidth {
				t.Errorf("Width: got %.2f, want %.2f", size.Width, tt.expectedWidth)
			}
			if size.Height != tt.expectedHeight {
				t.Errorf("Height: got %.2f, want %.2f", size.Height, tt.expectedHeight)
			}
		})
	}
}

// TestFlexGapWithUnits tests FlexGap with various units
func TestFlexGapWithUnits(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)

	tests := []struct {
		name               string
		gap                Length
		expectedChild2PosX float64
	}{
		{"Px gap", Px(20), 120},     // 100 + 20
		{"Em gap", Em(1), 116},      // 100 + 16
		{"Vw gap", Vw(2), 120},      // 100 + 20 (2% of 1000)
		{"Rem gap", Rem(1.25), 120}, // 100 + 20 (1.25 * 16)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Node{
				Style: Style{
					Display:       DisplayFlex,
					FlexDirection: FlexDirectionRow,
					FlexGap:       tt.gap,
					Width:         Px(500),
					Height:        Px(200),
				},
				Children: []*Node{
					{Style: Style{Width: Px(100), Height: Px(50)}},
					{Style: Style{Width: Px(100), Height: Px(50)}},
				},
			}

			constraints := Loose(1000, 800)
			Layout(root, constraints, ctx)

			child2 := root.Children[1]
			if child2.Rect.X != tt.expectedChild2PosX {
				t.Errorf("Child 2 X position: got %.2f, want %.2f", child2.Rect.X, tt.expectedChild2PosX)
			}
		})
	}
}

// TestGridGapWithUnits tests GridGap with various units
func TestGridGapWithUnits(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)

	tests := []struct {
		name               string
		gap                Length
		expectedChild2PosX float64
	}{
		{"Px gap", Px(10), 110},  // 100 + 10
		{"Em gap", Em(1), 116},   // 100 + 16
		{"Vh gap", Vh(2.5), 120}, // 100 + 20 (2.5% of 800)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Node{
				Style: Style{
					Display:             DisplayGrid,
					GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
					GridGap:             tt.gap,
					Width:               Px(500),
					Height:              Px(200),
				},
				Children: []*Node{
					{Style: Style{Width: Px(100), Height: Px(50)}},
					{Style: Style{Width: Px(100), Height: Px(50)}},
				},
			}

			constraints := Loose(1000, 800)
			Layout(root, constraints, ctx)

			child2 := root.Children[1]
			if child2.Rect.X != tt.expectedChild2PosX {
				t.Errorf("Child 2 X position: got %.2f, want %.2f", child2.Rect.X, tt.expectedChild2PosX)
			}
		})
	}
}

// TestPositioningWithUnits tests Top, Right, Bottom, Left with various units
func TestPositioningWithUnits(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)

	tests := []struct {
		name      string
		top       Length
		left      Length
		expectedX float64
		expectedY float64
	}{
		{"Px positioning", Px(50), Px(100), 100, 50},
		{"Em positioning", Em(3), Em(5), 80, 48},       // 5*16, 3*16
		{"Vh/Vw positioning", Vh(10), Vw(5), 50, 80},   // 5% of 1000, 10% of 800
		{"Mixed positioning", Rem(2), Vw(10), 100, 32}, // 10% of 1000, 2*16
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Node{
				Style: Style{
					Display: DisplayBlock,
					Width:   Px(500),
					Height:  Px(400),
				},
				Children: []*Node{
					{
						Style: Style{
							Position: PositionAbsolute,
							Top:      tt.top,
							Left:     tt.left,
							Width:    Px(100),
							Height:   Px(100),
						},
					},
				},
			}

			constraints := Loose(1000, 800)
			Layout(root, constraints, ctx)
			LayoutWithPositioning(root, constraints, root.Rect, ctx)

			child := root.Children[0]
			if child.Rect.X != tt.expectedX {
				t.Errorf("Child X: got %.2f, want %.2f", child.Rect.X, tt.expectedX)
			}
			if child.Rect.Y != tt.expectedY {
				t.Errorf("Child Y: got %.2f, want %.2f", child.Rect.Y, tt.expectedY)
			}
		})
	}
}

// TestFlexBasisWithUnits tests FlexBasis with various units
func TestFlexBasisWithUnits(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)

	tests := []struct {
		name          string
		flexBasis     Length
		expectedWidth float64
	}{
		{"Px basis", Px(150), 150},
		{"Em basis", Em(10), 160},     // 10 * 16
		{"Vw basis", Vw(20), 200},     // 20% of 1000
		{"Rem basis", Rem(12.5), 200}, // 12.5 * 16
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Node{
				Style: Style{
					Display:       DisplayFlex,
					FlexDirection: FlexDirectionRow,
					Width:         Px(800),
					Height:        Px(200),
				},
				Children: []*Node{
					{
						Style: Style{
							FlexBasis: tt.flexBasis,
							Height:    Px(50),
						},
					},
				},
			}

			constraints := Loose(1000, 800)
			Layout(root, constraints, ctx)

			child := root.Children[0]
			if child.Rect.Width != tt.expectedWidth {
				t.Errorf("Child width: got %.2f, want %.2f", child.Rect.Width, tt.expectedWidth)
			}
		})
	}
}

// TestFitContentWithUnits tests FitContentWidth/Height with various units
func TestFitContentWithUnits(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)

	tests := []struct {
		name            string
		fitContentWidth Length
		childWidth      float64
		expectedWidth   float64
	}{
		{"Px fit-content", Px(150), 200, 150}, // Clamped to 150
		{"Em fit-content", Em(10), 200, 160},  // Clamped to 160 (10*16)
		{"Vw fit-content", Vw(30), 400, 300},  // Clamped to 300 (30% of 1000)
		{"No clamp", Px(300), 200, 200},       // Not clamped
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &Node{
				Style: Style{
					Display:         DisplayBlock,
					WidthSizing:     IntrinsicSizeFitContent,
					FitContentWidth: tt.fitContentWidth,
					Height:          Px(100),
				},
				Children: []*Node{
					{
						Style: Style{
							Width:  Px(tt.childWidth),
							Height: Px(50),
						},
					},
				},
			}

			constraints := Loose(1000, 800)
			size := Layout(root, constraints, ctx)

			if size.Width != tt.expectedWidth {
				t.Errorf("Width: got %.2f, want %.2f", size.Width, tt.expectedWidth)
			}
		})
	}
}

// TestDifferentFontSizesForEm tests that Em units resolve correctly with different element font sizes
func TestDifferentFontSizesForEm(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16)

	root := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   Px(500),
			Height:  Px(500),
			TextStyle: &TextStyle{
				FontSize: 20, // Root element has 20pt font
			},
		},
		Children: []*Node{
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Em(5), // Should be 5 * 20 = 100
					Height:  Px(50),
					TextStyle: &TextStyle{
						FontSize: 20,
					},
				},
			},
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Em(5), // Should be 5 * 12 = 60
					Height:  Px(50),
					TextStyle: &TextStyle{
						FontSize: 12,
					},
				},
			},
		},
	}

	constraints := Loose(1000, 800)
	Layout(root, constraints, ctx)

	child1 := root.Children[0]
	child2 := root.Children[1]

	if child1.Rect.Width != 100 {
		t.Errorf("Child 1 width (Em with 20pt font): got %.2f, want 100", child1.Rect.Width)
	}
	if child2.Rect.Width != 60 {
		t.Errorf("Child 2 width (Em with 12pt font): got %.2f, want 60", child2.Rect.Width)
	}
}

// TestRemAlwaysUsesRootFontSize tests that Rem always uses root font size regardless of element font
func TestRemAlwaysUsesRootFontSize(t *testing.T) {
	ctx := NewLayoutContext(1000, 800, 16) // Root font is 16

	root := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   Px(500),
			Height:  Px(500),
			TextStyle: &TextStyle{
				FontSize: 24, // Element has different font
			},
		},
		Children: []*Node{
			{
				Style: Style{
					Display: DisplayBlock,
					Width:   Rem(5), // Should ALWAYS be 5 * 16 = 80 (uses root font)
					Height:  Px(50),
					TextStyle: &TextStyle{
						FontSize: 32, // Even with large font, Rem uses root
					},
				},
			},
		},
	}

	constraints := Loose(1000, 800)
	Layout(root, constraints, ctx)

	child := root.Children[0]
	if child.Rect.Width != 80 {
		t.Errorf("Child width (Rem should use root font): got %.2f, want 80", child.Rect.Width)
	}
}
