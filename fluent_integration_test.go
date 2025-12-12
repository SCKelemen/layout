package layout

import (
	"testing"
)

// =============================================================================
// Integration Tests: Fluent API vs Classic API
// =============================================================================
//
// These tests verify that the fluent API produces identical layout results
// to the classic API for equivalent tree structures.

// TestFluentVsClassicSimple verifies basic style equivalence
func TestFluentVsClassicSimple(t *testing.T) {
	// Classic API
	classic := &Node{
		Style: Style{
			Display: DisplayBlock,
			Width:   200,
			Height:  100,
			Padding: Uniform(10),
			Margin:  Uniform(8),
		},
	}

	// Fluent API
	fluent := (&Node{}).
		WithDisplay(DisplayBlock).
		WithWidth(200).
		WithHeight(100).
		WithPadding(10).
		WithMargin(8)

	// Layout both
	constraints := Loose(400, 600)
	Layout(classic, constraints)
	Layout(fluent, constraints)

	// Assert equivalent rects
	if classic.Rect != fluent.Rect {
		t.Errorf("Rects should be identical:\nClassic: %+v\nFluent:  %+v",
			classic.Rect, fluent.Rect)
	}

	// Assert equivalent styles
	if classic.Style.Width != fluent.Style.Width {
		t.Errorf("Width mismatch: classic=%.2f, fluent=%.2f",
			classic.Style.Width, fluent.Style.Width)
	}

	if classic.Style.Padding.Top != fluent.Style.Padding.Top {
		t.Errorf("Padding mismatch: classic=%.2f, fluent=%.2f",
			classic.Style.Padding.Top, fluent.Style.Padding.Top)
	}
}

// TestFluentVsClassicWithChildren verifies equivalence with child nodes
func TestFluentVsClassicWithChildren(t *testing.T) {
	// Classic API
	classic := &Node{
		Style: Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentSpaceBetween,
			Width:          400,
			Padding:        Uniform(10),
		},
		Children: []*Node{
			{Style: Style{Width: 100, Height: 50}},
			{Style: Style{Width: 150, Height: 50}},
			{Style: Style{Width: 100, Height: 50}},
		},
	}

	// Fluent API
	fluent := (&Node{}).
		WithStyle(Style{
			Display:        DisplayFlex,
			FlexDirection:  FlexDirectionRow,
			JustifyContent: JustifyContentSpaceBetween,
		}).
		WithWidth(400).
		WithPadding(10).
		AddChildren(
			(&Node{}).WithWidth(100).WithHeight(50),
			(&Node{}).WithWidth(150).WithHeight(50),
			(&Node{}).WithWidth(100).WithHeight(50),
		)

	// Layout both
	constraints := Loose(500, 600)
	Layout(classic, constraints)
	Layout(fluent, constraints)

	// Assert root equivalence
	if classic.Rect != fluent.Rect {
		t.Errorf("Root rects should be identical:\nClassic: %+v\nFluent:  %+v",
			classic.Rect, fluent.Rect)
	}

	// Assert children equivalence
	if len(classic.Children) != len(fluent.Children) {
		t.Fatalf("Child count mismatch: classic=%d, fluent=%d",
			len(classic.Children), len(fluent.Children))
	}

	for i := range classic.Children {
		classicChild := classic.Children[i]
		fluentChild := fluent.Children[i]

		if classicChild.Rect != fluentChild.Rect {
			t.Errorf("Child %d rect mismatch:\nClassic: %+v\nFluent:  %+v",
				i, classicChild.Rect, fluentChild.Rect)
		}
	}
}

// TestFluentVsClassicHelpers verifies equivalence with helper functions
func TestFluentVsClassicHelpers(t *testing.T) {
	// Classic API with helpers
	classic := HStack(
		Fixed(100, 50),
		Fixed(200, 50),
		Fixed(150, 50),
	)
	Padding(classic, 16)
	Margin(classic, 8)

	// Fluent API
	fluent := HStack(
		Fixed(100, 50),
		Fixed(200, 50),
		Fixed(150, 50),
	).WithPadding(16).WithMargin(8)

	// Layout both
	constraints := Loose(600, 400)
	Layout(classic, constraints)
	Layout(fluent, constraints)

	// Assert equivalence
	if classic.Rect != fluent.Rect {
		t.Errorf("Root rects should be identical:\nClassic: %+v\nFluent:  %+v",
			classic.Rect, fluent.Rect)
	}

	// Check padding was applied
	if classic.Style.Padding.Top != 16 || fluent.Style.Padding.Top != 16 {
		t.Errorf("Padding not applied correctly")
	}

	// Check margin was applied
	if classic.Style.Margin.Top != 8 || fluent.Style.Margin.Top != 8 {
		t.Errorf("Margin not applied correctly")
	}
}

// TestFluentVsClassicGrid verifies grid layout equivalence
func TestFluentVsClassicGrid(t *testing.T) {
	// Classic API
	classic := &Node{
		Style: Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(100),
				FractionTrack(1),
			},
			GridTemplateColumns: []GridTrack{
				FractionTrack(1),
				FractionTrack(1),
			},
			GridGap: 10,
			Padding: Uniform(20),
		},
		Children: []*Node{
			{
				Style: Style{
					GridRowStart:    0,
					GridRowEnd:      1,
					GridColumnStart: 0,
					GridColumnEnd:   2, // Spans both columns
				},
			},
			{
				Style: Style{
					GridRowStart:    1,
					GridRowEnd:      2,
					GridColumnStart: 0,
					GridColumnEnd:   1,
				},
			},
			{
				Style: Style{
					GridRowStart:    1,
					GridRowEnd:      2,
					GridColumnStart: 1,
					GridColumnEnd:   2,
				},
			},
		},
	}

	// Fluent API
	fluent := (&Node{}).
		WithStyle(Style{
			Display: DisplayGrid,
			GridTemplateRows: []GridTrack{
				FixedTrack(100),
				FractionTrack(1),
			},
			GridTemplateColumns: []GridTrack{
				FractionTrack(1),
				FractionTrack(1),
			},
			GridGap: 10,
		}).
		WithPadding(20).
		AddChildren(
			(&Node{}).WithStyle(Style{
				GridRowStart:    0,
				GridRowEnd:      1,
				GridColumnStart: 0,
				GridColumnEnd:   2,
			}),
			(&Node{}).WithStyle(Style{
				GridRowStart:    1,
				GridRowEnd:      2,
				GridColumnStart: 0,
				GridColumnEnd:   1,
			}),
			(&Node{}).WithStyle(Style{
				GridRowStart:    1,
				GridRowEnd:      2,
				GridColumnStart: 1,
				GridColumnEnd:   2,
			}),
		)

	// Layout both
	constraints := Tight(600, 400)
	Layout(classic, constraints)
	Layout(fluent, constraints)

	// Assert root equivalence
	if classic.Rect != fluent.Rect {
		t.Errorf("Root rects should be identical:\nClassic: %+v\nFluent:  %+v",
			classic.Rect, fluent.Rect)
	}

	// Assert all children equivalence
	for i := range classic.Children {
		classicChild := classic.Children[i]
		fluentChild := fluent.Children[i]

		if classicChild.Rect != fluentChild.Rect {
			t.Errorf("Child %d rect mismatch:\nClassic: %+v\nFluent:  %+v",
				i, classicChild.Rect, fluentChild.Rect)
		}
	}
}

// TestFluentVsClassicNested verifies deeply nested trees
func TestFluentVsClassicNested(t *testing.T) {
	// Classic API - nested structure
	classic := &Node{
		Style: Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
			Width:         400,
		},
		Children: []*Node{
			{
				Style: Style{
					Display:       DisplayFlex,
					FlexDirection: FlexDirectionRow,
					Height:        100,
				},
				Children: []*Node{
					{Style: Style{Width: 100}},
					{Style: Style{Width: 200}},
				},
			},
			{
				Style: Style{
					Display:       DisplayFlex,
					FlexDirection: FlexDirectionRow,
					Height:        100,
				},
				Children: []*Node{
					{Style: Style{Width: 150}},
					{Style: Style{Width: 150}},
				},
			},
		},
	}

	// Fluent API - nested structure
	fluent := (&Node{}).
		WithStyle(Style{
			Display:       DisplayFlex,
			FlexDirection: FlexDirectionColumn,
		}).
		WithWidth(400).
		AddChildren(
			(&Node{}).
				WithStyle(Style{
					Display:       DisplayFlex,
					FlexDirection: FlexDirectionRow,
				}).
				WithHeight(100).
				AddChildren(
					(&Node{}).WithWidth(100),
					(&Node{}).WithWidth(200),
				),
			(&Node{}).
				WithStyle(Style{
					Display:       DisplayFlex,
					FlexDirection: FlexDirectionRow,
				}).
				WithHeight(100).
				AddChildren(
					(&Node{}).WithWidth(150),
					(&Node{}).WithWidth(150),
				),
		)

	// Layout both
	constraints := Loose(500, 600)
	Layout(classic, constraints)
	Layout(fluent, constraints)

	// Assert root equivalence
	if classic.Rect != fluent.Rect {
		t.Errorf("Root rects should be identical")
	}

	// Assert all descendants equivalence
	classicDescendants := classic.DescendantsAndSelf()
	fluentDescendants := fluent.DescendantsAndSelf()

	if len(classicDescendants) != len(fluentDescendants) {
		t.Fatalf("Descendant count mismatch: classic=%d, fluent=%d",
			len(classicDescendants), len(fluentDescendants))
	}

	// Check each corresponding node
	for i := range classicDescendants {
		if classicDescendants[i].Rect != fluentDescendants[i].Rect {
			t.Errorf("Descendant %d rect mismatch", i)
		}
	}
}

// TestFluentImmutability verifies that fluent operations don't mutate originals
func TestFluentImmutability(t *testing.T) {
	original := HStack(
		Fixed(100, 50),
		Fixed(200, 50),
	)

	// Store original state
	originalWidth := original.Style.Width
	originalPadding := original.Style.Padding.Top
	originalChildCount := len(original.Children)

	// Apply fluent operations
	modified := original.
		WithWidth(300).
		WithPadding(20).
		AddChild(Fixed(150, 50))

	// Verify original unchanged
	if original.Style.Width != originalWidth {
		t.Errorf("Original width was modified: %.2f -> %.2f",
			originalWidth, original.Style.Width)
	}

	if original.Style.Padding.Top != originalPadding {
		t.Errorf("Original padding was modified: %.2f -> %.2f",
			originalPadding, original.Style.Padding.Top)
	}

	if len(original.Children) != originalChildCount {
		t.Errorf("Original children were modified: %d -> %d",
			originalChildCount, len(original.Children))
	}

	// Verify modified has new values
	if modified.Style.Width != 300 {
		t.Errorf("Modified width incorrect: %.2f", modified.Style.Width)
	}

	if modified.Style.Padding.Top != 20 {
		t.Errorf("Modified padding incorrect: %.2f", modified.Style.Padding.Top)
	}

	if len(modified.Children) != 3 {
		t.Errorf("Modified children count incorrect: %d", len(modified.Children))
	}
}

// TestFluentChainEquivalence verifies that chained operations produce expected results
func TestFluentChainEquivalence(t *testing.T) {
	// Build using chained fluent operations
	chained := (&Node{}).
		WithStyle(Style{Display: DisplayFlex, FlexDirection: FlexDirectionRow}).
		WithWidth(400).
		WithPadding(10).
		WithMargin(5).
		AddChild((&Node{}).WithWidth(100)).
		AddChild((&Node{}).WithWidth(200)).
		AddChild((&Node{}).WithWidth(100))

	// Build step by step
	step1 := (&Node{}).WithStyle(Style{Display: DisplayFlex, FlexDirection: FlexDirectionRow})
	step2 := step1.WithWidth(400)
	step3 := step2.WithPadding(10)
	step4 := step3.WithMargin(5)
	step5 := step4.AddChild((&Node{}).WithWidth(100))
	step6 := step5.AddChild((&Node{}).WithWidth(200))
	stepwise := step6.AddChild((&Node{}).WithWidth(100))

	// Layout both
	constraints := Loose(500, 600)
	Layout(chained, constraints)
	Layout(stepwise, constraints)

	// Assert equivalence
	if chained.Rect != stepwise.Rect {
		t.Errorf("Chained and stepwise should produce identical results")
	}

	if len(chained.Children) != len(stepwise.Children) {
		t.Errorf("Child count mismatch")
	}

	for i := range chained.Children {
		if chained.Children[i].Rect != stepwise.Children[i].Rect {
			t.Errorf("Child %d rect mismatch", i)
		}
	}
}

// TestFluentTransformEquivalence verifies Transform produces correct results
func TestFluentTransformEquivalence(t *testing.T) {
	// Build tree with classic API
	classic := HStack(
		Fixed(100, 50),
		Fixed(200, 50),
		Fixed(150, 50),
	)

	// Transform using fluent API
	transformed := classic.Transform(
		func(n *Node) bool {
			return n.Style.Width > 0 && n.Style.Width < 200
		},
		func(n *Node) *Node {
			return n.WithWidth(n.Style.Width * 2)
		},
	)

	// Build equivalent tree manually
	manual := HStack(
		Fixed(200, 50), // 100 * 2
		Fixed(200, 50), // unchanged
		Fixed(300, 50), // 150 * 2
	)

	// Layout both
	constraints := Loose(800, 600)
	Layout(transformed, constraints)
	Layout(manual, constraints)

	// Assert equivalence
	if len(transformed.Children) != len(manual.Children) {
		t.Fatalf("Child count mismatch")
	}

	for i := range transformed.Children {
		transformedWidth := transformed.Children[i].Style.Width
		manualWidth := manual.Children[i].Style.Width

		if transformedWidth != manualWidth {
			t.Errorf("Child %d width mismatch: transformed=%.2f, manual=%.2f",
				i, transformedWidth, manualWidth)
		}
	}
}

// TestFluentFilterEquivalence verifies Filter produces correct results
func TestFluentFilterEquivalence(t *testing.T) {
	// Build tree with mixed sizes
	original := HStack(
		Fixed(100, 50),
		Fixed(250, 50),
		Fixed(150, 50),
		Fixed(300, 50),
	)

	// Filter using fluent API
	filtered := original.Filter(func(n *Node) bool {
		return n.Style.Width >= 200
	})

	// Build equivalent tree manually
	manual := HStack(
		Fixed(250, 50),
		Fixed(300, 50),
	)

	// Layout both
	constraints := Loose(800, 600)
	Layout(filtered, constraints)
	Layout(manual, constraints)

	// Assert equivalence
	if len(filtered.Children) != len(manual.Children) {
		t.Errorf("Filtered child count should be %d, got %d",
			len(manual.Children), len(filtered.Children))
	}

	if len(filtered.Children) != 2 {
		t.Fatalf("Expected 2 filtered children, got %d", len(filtered.Children))
	}

	// Check widths match
	if filtered.Children[0].Style.Width != 250 {
		t.Errorf("First filtered child width should be 250, got %.2f",
			filtered.Children[0].Style.Width)
	}

	if filtered.Children[1].Style.Width != 300 {
		t.Errorf("Second filtered child width should be 300, got %.2f",
			filtered.Children[1].Style.Width)
	}
}

// TestFluentMapEquivalence verifies Map produces correct results
func TestFluentMapEquivalence(t *testing.T) {
	// Build tree
	original := HStack(
		Fixed(100, 50),
		Fixed(200, 50),
	)

	// Scale using Map
	scaled := original.Map(func(n *Node) *Node {
		return n.
			WithWidth(n.Style.Width * 1.5).
			WithHeight(n.Style.Height * 1.5)
	})

	// Build equivalent tree manually
	manual := HStack(
		Fixed(150, 75), // 100 * 1.5, 50 * 1.5
		Fixed(300, 75), // 200 * 1.5, 50 * 1.5
	)

	// Note: HStack itself also gets scaled
	manual.Style.Width = manual.Style.Width * 1.5
	manual.Style.Height = manual.Style.Height * 1.5

	// Layout both
	constraints := Loose(800, 600)
	Layout(scaled, constraints)
	Layout(manual, constraints)

	// Check children widths
	if len(scaled.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(scaled.Children))
	}

	if scaled.Children[0].Style.Width != 150 {
		t.Errorf("First child width should be 150, got %.2f",
			scaled.Children[0].Style.Width)
	}

	if scaled.Children[1].Style.Width != 300 {
		t.Errorf("Second child width should be 300, got %.2f",
			scaled.Children[1].Style.Width)
	}
}
