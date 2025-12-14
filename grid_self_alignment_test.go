package layout

import "testing"

// TestGridJustifySelfBasic tests basic justify-self functionality
// justify-self should override the container's justify-items for individual items
func TestGridJustifySelfBasic(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100))},
			JustifyItems:        JustifyItemsStart,
			AlignItems:          AlignItemsFlexStart,
			Width:               Px(300),
			Height:              Px(100),
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50), JustifySelf: 0}},                  // Use parent's justify-items (start)
			{Style: Style{Width: Px(50), Height: Px(50), JustifySelf: JustifyItemsEnd}},    // Override: end
			{Style: Style{Width: Px(50), Height: Px(50), JustifySelf: JustifyItemsCenter}}, // Override: center
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(300, 100), ctx)

	// First item: start (X=0)
	if container.Children[0].Rect.X != 0 {
		t.Errorf("Item with default JustifySelf should be at X=0, got %v", container.Children[0].Rect.X)
	}

	// Second item: end (X=150, cell starts at 100, item width 50, so 100+100-50=150)
	if container.Children[1].Rect.X != 150 {
		t.Errorf("Item with JustifySelf=End should be at X=150, got %v", container.Children[1].Rect.X)
	}

	// Third item: center (X=225, cell starts at 200, centered in 100px cell: 200+25=225)
	if container.Children[2].Rect.X != 225 {
		t.Errorf("Item with JustifySelf=Center should be at X=225, got %v", container.Children[2].Rect.X)
	}
}

// TestGridAlignSelfBasic tests basic align-self functionality
func TestGridAlignSelfBasic(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100)), FixedTrack(Px(100))},
			JustifyItems:        JustifyItemsStart,
			AlignItems:          AlignItemsFlexStart,
			Width:               Px(100),
			Height:              Px(300),
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50), AlignSelf: 0}},                 // Use parent's align-items (start)
			{Style: Style{Width: Px(50), Height: Px(50), AlignSelf: AlignItemsFlexEnd}}, // Override: end
			{Style: Style{Width: Px(50), Height: Px(50), AlignSelf: AlignItemsCenter}},  // Override: center
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 300), ctx)

	// First item: start (Y=0)
	if container.Children[0].Rect.Y != 0 {
		t.Errorf("Item with default AlignSelf should be at Y=0, got %v", container.Children[0].Rect.Y)
	}

	// Second item: end (Y=150, cell starts at 100, item height 50, so 100+100-50=150)
	if container.Children[1].Rect.Y != 150 {
		t.Errorf("Item with AlignSelf=FlexEnd should be at Y=150, got %v", container.Children[1].Rect.Y)
	}

	// Third item: center (Y=225, cell starts at 200, centered in 100px cell: 200+25=225)
	if container.Children[2].Rect.Y != 225 {
		t.Errorf("Item with AlignSelf=Center should be at Y=225, got %v", container.Children[2].Rect.Y)
	}
}

// TestGridSelfAlignmentStretch tests that stretch overrides work correctly
func TestGridSelfAlignmentStretch(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100))},
			JustifyItems:        JustifyItemsStart,
			AlignItems:          AlignItemsFlexStart,
			Width:               Px(100),
			Height:              Px(100),
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50), JustifySelf: JustifyItemsStretch, AlignSelf: AlignItemsStretch}},
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 100), ctx)

	// Items with explicit size don't stretch (per CSS spec)
	// Stretch only applies when no explicit size is set
	if container.Children[0].Rect.Width != 50 {
		t.Errorf("Item with JustifySelf=Stretch and explicit width should keep Width=50, got %v", container.Children[0].Rect.Width)
	}
	if container.Children[0].Rect.Height != 50 {
		t.Errorf("Item with AlignSelf=Stretch and explicit height should keep Height=50, got %v", container.Children[0].Rect.Height)
	}
}

// TestGridSelfAlignmentWithMargins tests self-alignment with margins
func TestGridSelfAlignmentWithMargins(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100))},
			JustifyItems:        JustifyItemsStart,
			AlignItems:          AlignItemsFlexStart,
			Width:               Px(100),
			Height:              Px(100),
		},
		Children: []*Node{
			{Style: Style{
				Width:       Px(50),
				Height:      Px(50),
				JustifySelf: JustifyItemsEnd,
				AlignSelf:   AlignItemsFlexEnd,
				Margin:      Uniform(Px(10)),
			}},
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 100), ctx)

	// Cell is 100x100
	// Item is 50x50 with 10px margins
	// JustifySelf=End: total width including margins = 50+10+10=70, so item X = 100-70+10 = 40
	// AlignSelf=End: total height including margins = 50+10+10=70, so item Y = 100-70+10 = 40
	if container.Children[0].Rect.X != 40 {
		t.Errorf("Item with JustifySelf=End and margins should be at X=40, got %v", container.Children[0].Rect.X)
	}
	if container.Children[0].Rect.Y != 40 {
		t.Errorf("Item with AlignSelf=End and margins should be at Y=40, got %v", container.Children[0].Rect.Y)
	}
}

// TestGridSelfAlignmentSpanning tests self-alignment with spanning items
func TestGridSelfAlignmentSpanning(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			JustifyItems:        JustifyItemsStart,
			AlignItems:          AlignItemsFlexStart,
			Width:               Px(200),
			Height:              Px(200),
		},
		Children: []*Node{
			{Style: Style{
				Width:           Px(80),
				Height:          Px(80),
				GridColumnStart: 0,
				GridColumnEnd:   2, // Span 2 columns
				GridRowStart:    0,
				GridRowEnd:      2, // Span 2 rows
				JustifySelf:     JustifyItemsCenter,
				AlignSelf:       AlignItemsCenter,
			}},
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(200, 200), ctx)

	// Spanning cell is 200x200 (entire grid)
	// Item is 80x80
	// JustifySelf=Center: X = (200-80)/2 = 60
	// AlignSelf=Center: Y = (200-80)/2 = 60
	if container.Children[0].Rect.X != 60 {
		t.Errorf("Spanning item with JustifySelf=Center should be at X=60, got %v", container.Children[0].Rect.X)
	}
	if container.Children[0].Rect.Y != 60 {
		t.Errorf("Spanning item with AlignSelf=Center should be at Y=60, got %v", container.Children[0].Rect.Y)
	}
}

// TestGridSelfAlignmentWithGaps tests self-alignment with grid gaps
func TestGridSelfAlignmentWithGaps(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100))},
			GridColumnGap:       Px(20),
			JustifyItems:        JustifyItemsStart,
			AlignItems:          AlignItemsFlexStart,
			Width:               Px(220), // 100 + 20 + 100
			Height:              Px(100),
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50), JustifySelf: JustifyItemsCenter}},
			{Style: Style{Width: Px(50), Height: Px(50), JustifySelf: JustifyItemsEnd}},
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(220, 100), ctx)

	// First item: center in first cell (0-100), X = 25
	if container.Children[0].Rect.X != 25 {
		t.Errorf("First item with JustifySelf=Center should be at X=25, got %v", container.Children[0].Rect.X)
	}

	// Second item: end in second cell (120-220), X = 170 (220-50)
	if container.Children[1].Rect.X != 170 {
		t.Errorf("Second item with JustifySelf=End should be at X=170, got %v", container.Children[1].Rect.X)
	}
}

// TestGridSelfAlignmentBaseline tests align-self with baseline
func TestGridSelfAlignmentBaseline(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100))},
			JustifyItems:        JustifyItemsStart,
			AlignItems:          AlignItemsFlexStart,
			Width:               Px(100),
			Height:              Px(100),
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50), AlignSelf: AlignItemsBaseline}},
		},
	}
	container.Children[0].Baseline = 30

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 100), ctx)

	// With baseline alignment, item should be positioned to align its baseline
	// For grid, baseline alignment positions item at start by default
	// (proper baseline alignment across items in same row would require more complex logic)
	if container.Children[0].Rect.Y < 0 || container.Children[0].Rect.Y > 50 {
		t.Errorf("Item with AlignSelf=Baseline should be positioned within cell, got Y=%v", container.Children[0].Rect.Y)
	}
}

// TestGridSelfAlignmentMixed tests mixed container and item alignment
func TestGridSelfAlignmentMixed(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(100)), FixedTrack(Px(100))},
			JustifyItems:        JustifyItemsCenter,
			AlignItems:          AlignItemsCenter,
			Width:               Px(200),
			Height:              Px(200),
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}},                                                             // Use container alignment (center, center)
			{Style: Style{Width: Px(50), Height: Px(50), JustifySelf: JustifyItemsStart}},                             // Override justify only
			{Style: Style{Width: Px(50), Height: Px(50), AlignSelf: AlignItemsFlexStart}},                             // Override align only
			{Style: Style{Width: Px(50), Height: Px(50), JustifySelf: JustifyItemsEnd, AlignSelf: AlignItemsFlexEnd}}, // Override both
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(200, 200), ctx)

	// First item: center/center (25, 25)
	if container.Children[0].Rect.X != 25 || container.Children[0].Rect.Y != 25 {
		t.Errorf("First item should be at (25,25), got (%v,%v)", container.Children[0].Rect.X, container.Children[0].Rect.Y)
	}

	// Second item: start/center (100, 25) - in cell (100,0)
	if container.Children[1].Rect.X != 100 || container.Children[1].Rect.Y != 25 {
		t.Errorf("Second item should be at (100,25), got (%v,%v)", container.Children[1].Rect.X, container.Children[1].Rect.Y)
	}

	// Third item: center/start (25, 100) - in cell (0,100)
	if container.Children[2].Rect.X != 25 || container.Children[2].Rect.Y != 100 {
		t.Errorf("Third item should be at (25,100), got (%v,%v)", container.Children[2].Rect.X, container.Children[2].Rect.Y)
	}

	// Fourth item: end/end (150, 150) - in cell (100,100), aligned to end: 100+100-50=150
	if container.Children[3].Rect.X != 150 || container.Children[3].Rect.Y != 150 {
		t.Errorf("Fourth item should be at (150,150), got (%v,%v)", container.Children[3].Rect.X, container.Children[3].Rect.Y)
	}
}
