package layout

import "testing"

// TestGridAlignContentStart tests align-content: start for grid rows
// All tracks should start from the beginning with no free space distribution
func TestGridAlignContentStart(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			AlignContent:        AlignContentFlexStart,
			Width:               Px(100),
			Height:              Px(200), // Extra 100px of free space
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 0
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 1
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 200), ctx)

	// First item: should be at Y=0
	if container.Children[0].Rect.Y != 0 {
		t.Errorf("First item should be at Y=0, got %v", container.Children[0].Rect.Y)
	}

	// Second item: should be at Y=50 (directly after first row)
	if container.Children[1].Rect.Y != 50 {
		t.Errorf("Second item should be at Y=50, got %v", container.Children[1].Rect.Y)
	}
}

// TestGridAlignContentEnd tests align-content: end for grid rows
// All tracks should be pushed to the end
func TestGridAlignContentEnd(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			AlignContent:        AlignContentFlexEnd,
			Width:               Px(100),
			Height:              Px(200), // Extra 100px of free space
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 0
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 1
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 200), ctx)

	// First item: should be at Y=100 (200 - 50 - 50)
	if container.Children[0].Rect.Y != 100 {
		t.Errorf("First item should be at Y=100, got %v", container.Children[0].Rect.Y)
	}

	// Second item: should be at Y=150 (100 + 50)
	if container.Children[1].Rect.Y != 150 {
		t.Errorf("Second item should be at Y=150, got %v", container.Children[1].Rect.Y)
	}
}

// TestGridAlignContentCenter tests align-content: center for grid rows
// Tracks should be centered with equal free space at start and end
func TestGridAlignContentCenter(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			AlignContent:        AlignContentCenter,
			Width:               Px(100),
			Height:              Px(200), // Extra 100px of free space
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 0
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 1
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 200), ctx)

	// First item: should be at Y=50 (100px free space / 2 = 50)
	if container.Children[0].Rect.Y != 50 {
		t.Errorf("First item should be at Y=50, got %v", container.Children[0].Rect.Y)
	}

	// Second item: should be at Y=100 (50 + 50)
	if container.Children[1].Rect.Y != 100 {
		t.Errorf("Second item should be at Y=100, got %v", container.Children[1].Rect.Y)
	}
}

// TestGridAlignContentSpaceBetween tests align-content: space-between for grid rows
// Free space should be distributed evenly between tracks
func TestGridAlignContentSpaceBetween(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			AlignContent:        AlignContentSpaceBetween,
			Width:               Px(100),
			Height:              Px(200), // Extra 100px of free space
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 0
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 1
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 200), ctx)

	// First item: should be at Y=0
	if container.Children[0].Rect.Y != 0 {
		t.Errorf("First item should be at Y=0, got %v", container.Children[0].Rect.Y)
	}

	// Second item: should be at Y=150 (0 + 50 + 100px between)
	if container.Children[1].Rect.Y != 150 {
		t.Errorf("Second item should be at Y=150, got %v", container.Children[1].Rect.Y)
	}
}

// TestGridAlignContentSpaceAround tests align-content: space-around for grid rows
// Free space should be distributed around tracks with half at each end
func TestGridAlignContentSpaceAround(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			AlignContent:        AlignContentSpaceAround,
			Width:               Px(100),
			Height:              Px(200), // Extra 100px of free space
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 0
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 1
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 200), ctx)

	// With 100px free space and 2 tracks: 100/2 = 50px per track
	// Half at start (25), full between (50), half at end (25)
	// First item: should be at Y=25 (50/2)
	if container.Children[0].Rect.Y != 25 {
		t.Errorf("First item should be at Y=25, got %v", container.Children[0].Rect.Y)
	}

	// Second item: should be at Y=125 (25 + 50 + 50)
	if container.Children[1].Rect.Y != 125 {
		t.Errorf("Second item should be at Y=125, got %v", container.Children[1].Rect.Y)
	}
}

// TestGridAlignContentStretch tests align-content: stretch for grid rows.
//
// CSS Grid §12.8 "Stretch auto Tracks": only tracks whose max sizing function
// is auto grow to fill free space. Fixed 50px tracks stay 50px, so stretch
// behaves as start here. (This test previously asserted that the fixed rows
// were stretched to 100px each, which the spec does not allow.)
// https://www.w3.org/TR/css-grid-1/#algo-stretch
func TestGridAlignContentStretch(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			AlignContent:        AlignContentStretch,
			Width:               Px(100),
			Height:              Px(200), // Extra 100px of free space
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 0
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 1
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 200), ctx)

	// First item: should be at Y=0
	if container.Children[0].Rect.Y != 0 {
		t.Errorf("First item should be at Y=0, got %v", container.Children[0].Rect.Y)
	}

	// Second item: Y=50. The fixed 50px rows are not auto tracks, so §12.8
	// leaves them alone and the free space stays at the end.
	if container.Children[1].Rect.Y != 50 {
		t.Errorf("Second item should be at Y=50, got %v", container.Children[1].Rect.Y)
	}

	// The container keeps its explicit 200px height even though the tracks
	// only fill 100px.
	if container.Rect.Height != 200 {
		t.Errorf("Container height should be 200, got %v", container.Rect.Height)
	}

	// Items with an explicit height must NOT stretch: per CSS Box Alignment
	// Level 3 §6.2, stretch is a no-op when the axis size is definite, so
	// each item keeps its 50px height.
	// https://www.w3.org/TR/css-align-3/#stretch-alignment
	if container.Children[0].Rect.Height != 50 {
		t.Errorf("First item height should remain 50, got %v", container.Children[0].Rect.Height)
	}
	if container.Children[1].Rect.Height != 50 {
		t.Errorf("Second item height should remain 50, got %v", container.Children[1].Rect.Height)
	}
}

// TestGridAlignContentWithGaps tests align-content with row gaps
func TestGridAlignContentWithGaps(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			GridRowGap:          Px(10),
			AlignContent:        AlignContentCenter,
			Width:               Px(100),
			Height:              Px(200), // 200 - 50 - 10 - 50 = 90px free space
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 0
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 1
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 200), ctx)

	// Total track size: 50 + 10 + 50 = 110
	// Free space: 200 - 110 = 90
	// Center: 90/2 = 45
	// First item: should be at Y=45
	if container.Children[0].Rect.Y != 45 {
		t.Errorf("First item should be at Y=45, got %v", container.Children[0].Rect.Y)
	}

	// Second item: should be at Y=105 (45 + 50 + 10)
	if container.Children[1].Rect.Y != 105 {
		t.Errorf("Second item should be at Y=105, got %v", container.Children[1].Rect.Y)
	}
}

// TestGridAlignContentWithSpanning tests align-content with spanning items
func TestGridAlignContentWithSpanning(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50)), FixedTrack(Px(50))},
			AlignContent:        AlignContentCenter,
			Width:               Px(100),
			Height:              Px(300), // 300 - 150 = 150px free space
		},
		Children: []*Node{
			{Style: Style{
				Width:        Px(50),
				Height:       Px(110), // Spans rows 0-1 (50+50)
				GridRowStart: 0,
				GridRowEnd:   2,
			}},
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 2
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 300), ctx)

	// Fixed tracks remain at 50px each despite spanning item
	// Total: 50 + 50 + 50 = 150px
	// Free space: 300 - 150 = 150
	// Center offset: 150/2 = 75
	// First item (spanning): should be at Y=75
	if container.Children[0].Rect.Y != 75 {
		t.Errorf("First spanning item should be at Y=75, got %v", container.Children[0].Rect.Y)
	}

	// Second item: Y=175 (75 offset + 50 + 50). Fixed tracks never grow from
	// a spanning item (CSS Grid §12.5 only distributes extra space to
	// intrinsic tracks), so the centered 150px block starts at 75 and row 2
	// begins at 175. (Previously asserted 125, an artifact of the spanning
	// item's height being split into the fixed rows.)
	// https://www.w3.org/TR/css-grid-1/#algo-content
	if container.Children[1].Rect.Y != 175 {
		t.Errorf("Second item should be at Y=175, got %v", container.Children[1].Rect.Y)
	}
}

// TestGridAlignContentSingleTrack tests align-content with single track
// Space-between should behave like flex-start for single track
func TestGridAlignContentSingleTrack(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50))},
			AlignContent:        AlignContentSpaceBetween,
			Width:               Px(100),
			Height:              Px(200), // Extra 150px of free space
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}},
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 200), ctx)

	// With single track, space-between behaves like flex-start
	// Item should be at Y=0
	if container.Children[0].Rect.Y != 0 {
		t.Errorf("Item should be at Y=0 (space-between with single track), got %v", container.Children[0].Rect.Y)
	}
}

// TestGridAlignContentNoFreeSpace tests align-content when there's no free space
// All alignment modes should produce the same result
func TestGridAlignContentNoFreeSpace(t *testing.T) {
	container := &Node{
		Style: Style{
			Display:             DisplayGrid,
			GridTemplateColumns: []GridTrack{FixedTrack(Px(100))},
			GridTemplateRows:    []GridTrack{FixedTrack(Px(50)), FixedTrack(Px(50))},
			AlignContent:        AlignContentCenter,
			Width:               Px(100),
			Height:              Px(100), // Exact fit, no free space
		},
		Children: []*Node{
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 0
			{Style: Style{Width: Px(50), Height: Px(50)}}, // Row 1
		},
	}

	ctx := NewLayoutContext(800, 600, 16)
	LayoutGrid(container, Loose(100, 100), ctx)

	// First item: should be at Y=0
	if container.Children[0].Rect.Y != 0 {
		t.Errorf("First item should be at Y=0, got %v", container.Children[0].Rect.Y)
	}

	// Second item: should be at Y=50
	if container.Children[1].Rect.Y != 50 {
		t.Errorf("Second item should be at Y=50, got %v", container.Children[1].Rect.Y)
	}
}
