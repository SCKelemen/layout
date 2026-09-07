package layout

import (
	"testing"
	"time"
)

// TestCalculateAutoRepeatCount tests the calculation of repeat count
func TestCalculateAutoRepeatCount(t *testing.T) {
	// Pattern: [100px]
	repeat := RepeatTrack{
		Count:  RepeatCountAutoFill,
		Tracks: []GridTrack{FixedTrack(Px(100))},
	}

	// Available: 350px, gap: 10px
	// Formula: floor((350 + 10) / (100 + 10)) = floor(360 / 110) = 3
	count := calculateAutoRepeatCount(repeat, 350, 10)
	if count != 3 {
		t.Errorf("Expected 3 repetitions, got %d", count)
	}

	// Available: 250px, gap: 0px
	// Formula: floor((250 + 0) / (100 + 0)) = floor(250 / 100) = 2
	count = calculateAutoRepeatCount(repeat, 250, 0)
	if count != 2 {
		t.Errorf("Expected 2 repetitions, got %d", count)
	}

	// Available: 50px (less than one repetition)
	// Should return minimum of 1
	count = calculateAutoRepeatCount(repeat, 50, 0)
	if count != 1 {
		t.Errorf("Expected minimum 1 repetition, got %d", count)
	}
}

// TestCalculateAutoRepeatCountMultipleTracks tests with multiple tracks in pattern
func TestCalculateAutoRepeatCountMultipleTracks(t *testing.T) {
	// Pattern: [100px, 50px]
	repeat := RepeatTrack{
		Count: RepeatCountAutoFill,
		Tracks: []GridTrack{
			FixedTrack(Px(100)),
			FixedTrack(Px(50)),
		},
	}

	// Available: 500px, gap: 10px
	// Repetition size: 100 + 50 + 10 (gap within) = 160
	// Formula: floor((500 + 10) / (160 + 10)) = floor(510 / 170) = 3
	count := calculateAutoRepeatCount(repeat, 500, 10)
	if count != 3 {
		t.Errorf("Expected 3 repetitions, got %d", count)
	}
}

// TestExpandAutoRepeatTracksBasic tests basic auto-repeat expansion
func TestExpandAutoRepeatTracksBasic(t *testing.T) {
	repeats := []RepeatTrack{
		{
			Count:  RepeatCountAutoFill,
			Tracks: []GridTrack{FixedTrack(Px(100))},
		},
	}

	// Available: 350px, gap: 10px
	// Should create 3 tracks of 100px
	result, autoStart, autoEnd := expandAutoRepeatTracks(repeats, []GridTrack{}, 350, 10)

	if len(result) != 3 {
		t.Errorf("Expected 3 tracks, got %d", len(result))
	}
	if autoStart != 0 || autoEnd != 3 {
		t.Errorf("Expected auto-repeat range [0, 3), got [%d, %d)", autoStart, autoEnd)
	}

	for i, track := range result {
		if track.MinSize.Value != 100 || track.MaxSize.Value != 100 {
			t.Errorf("Track %d: expected 100px, got min=%.0f max=%.0f",
				i, track.MinSize.Value, track.MaxSize.Value)
		}
	}
}

// TestExpandAutoRepeatTracksWithExplicit tests auto-repeat with explicit tracks
func TestExpandAutoRepeatTracksWithExplicit(t *testing.T) {
	explicit := []GridTrack{
		FixedTrack(Px(200)), // Explicit track takes 200px
	}

	repeats := []RepeatTrack{
		{
			Count:  RepeatCountAutoFill,
			Tracks: []GridTrack{FixedTrack(Px(100))},
		},
	}

	// Available: 550px, gap: 10px
	// Explicit uses: 200px plus the 10px gap before the repeated block
	// Remaining: 550 - 200 - 10 = 340px
	// Auto-repeat: floor((340 + 10) / (100 + 10)) = 3
	// Total tracks: 1 explicit + 3 auto = 4 (200 + 3*100 + 3*10 = 530 <= 550)
	result, autoStart, autoEnd := expandAutoRepeatTracks(repeats, explicit, 550, 10)

	if len(result) != 4 {
		t.Errorf("Expected 4 tracks (1 explicit + 3 auto), got %d", len(result))
	}
	if autoStart != 1 || autoEnd != 4 {
		t.Errorf("Expected auto-repeat range [1, 4), got [%d, %d)", autoStart, autoEnd)
	}

	// First track should be explicit (200px)
	if result[0].MinSize.Value != 200 {
		t.Errorf("First track should be 200px, got %.0f", result[0].MinSize.Value)
	}

	// Next 3 should be auto-repeated (100px each)
	for i := 1; i < 4; i++ {
		if result[i].MinSize.Value != 100 {
			t.Errorf("Track %d should be 100px, got %.0f", i, result[i].MinSize.Value)
		}
	}
}

// TestAutoRepeatCountsAgainstOtherTracks checks that the auto-repeat count
// is the largest integer for which ALL tracks in the axis plus their gaps fit
// (css-grid-1 §7.2.3.2), not floor((remaining + gap) / (pattern + gap))
// against the space left by the explicit tracks alone.
//
// https://www.w3.org/TR/css-grid-1/#auto-repeat
func TestAutoRepeatCountsAgainstOtherTracks(t *testing.T) {
	hundred := []GridTrack{FixedTrack(Px(100))}
	autoFill := RepeatTrack{Count: RepeatCountAutoFill, Tracks: hundred}

	// [100px] repeat(auto-fill, 100px), gap 10px, 315px: one repetition is
	// 100 + 10 + 100 = 210, two would be 320 > 315.
	result, _, _ := expandAutoRepeatTracks([]RepeatTrack{autoFill}, hundred, 315, 10)
	if len(result) != 2 {
		t.Errorf("[100px] + auto-fill in 315px with 10px gaps: expected 2 tracks, got %d", len(result))
	}

	// repeat(2, 100px) repeat(auto-fill, 100px) in 400px: the fixed-count
	// pattern uses 200px, leaving room for 2 repetitions, 4 tracks total.
	result, autoStart, autoEnd := expandAutoRepeatTracks([]RepeatTrack{{Count: 2, Tracks: hundred}, autoFill}, nil, 400, 0)
	if len(result) != 4 {
		t.Errorf("repeat(2) + auto-fill in 400px: expected 4 tracks, got %d", len(result))
	}
	if autoStart != 2 || autoEnd != 4 {
		t.Errorf("expected auto-repeat range [2, 4), got [%d, %d)", autoStart, autoEnd)
	}

	// A fixed-count pattern AFTER the auto-repeat counts as well, and with
	// gaps: auto-fill 100px then repeat(2, 100px), gap 10, 430px: the fixed
	// tracks take 200 + 2 gaps = 220, leaving 210 = 1 repetition + gap + 1?
	// No: 100 + 10 + 100 + 10 + 100 = 320 <= 430, 4 tracks = 430 <= 430, so 2
	// repetitions fit exactly; 3 would be 540.
	result, autoStart, autoEnd = expandAutoRepeatTracks([]RepeatTrack{autoFill, {Count: 2, Tracks: hundred}}, nil, 430, 10)
	if len(result) != 4 {
		t.Errorf("auto-fill + repeat(2) in 430px with 10px gaps: expected 4 tracks, got %d", len(result))
	}
	if autoStart != 0 || autoEnd != 2 {
		t.Errorf("expected auto-repeat range [0, 2), got [%d, %d)", autoStart, autoEnd)
	}
	// One pixel less and only one repetition fits.
	result, _, _ = expandAutoRepeatTracks([]RepeatTrack{autoFill, {Count: 2, Tracks: hundred}}, nil, 429, 10)
	if len(result) != 3 {
		t.Errorf("auto-fill + repeat(2) in 429px with 10px gaps: expected 3 tracks, got %d", len(result))
	}

	// Explicit track + fixed-count pattern + auto-fill together: [50px]
	// repeat(1, 50px) repeat(auto-fill, 100px), gap 10, 340px: the four
	// other-track gaps and 100px of other tracks leave 340 - 100 - 20 = 220,
	// floor((220 + 10) / 110) = 2 repetitions: 50+10+50+10+100+10+100 = 330.
	result, _, _ = expandAutoRepeatTracks([]RepeatTrack{{Count: 1, Tracks: []GridTrack{FixedTrack(Px(50))}}, autoFill}, []GridTrack{FixedTrack(Px(50))}, 340, 10)
	if len(result) != 4 {
		t.Errorf("[50px] + repeat(1, 50px) + auto-fill in 340px: expected 4 tracks, got %d", len(result))
	}

	// End to end: the 315px grid lays out two columns, so the third item
	// wraps to the second row (row gap 10, 10px tall items: Y = 20).
	grid := &Node{Style: Style{
		Display:                   DisplayGrid,
		Width:                     Px(315),
		GridGap:                   Px(10),
		GridTemplateColumns:       hundred,
		GridTemplateColumnsRepeat: []RepeatTrack{autoFill},
	}}
	for i := 0; i < 4; i++ {
		grid.Children = append(grid.Children, &Node{Style: Style{Height: Px(10)}})
	}
	LayoutGrid(grid, Loose(315, Unbounded), NewLayoutContext(800, 600, 16))
	if x, y := grid.Children[1].Rect.X, grid.Children[1].Rect.Y; x != 110 || y != 0 {
		t.Errorf("item 1: expected (110, 0), got (%v, %v)", x, y)
	}
	if x, y := grid.Children[2].Rect.X, grid.Children[2].Rect.Y; x != 0 || y != 20 {
		t.Errorf("item 2: expected (0, 20) on the second row, got (%v, %v)", x, y)
	}

	// repeat(2, 100px) repeat(auto-fill, 100px) in 400px: four columns, the
	// fifth item wraps.
	grid = &Node{Style: Style{
		Display:                   DisplayGrid,
		Width:                     Px(400),
		GridTemplateColumnsRepeat: []RepeatTrack{{Count: 2, Tracks: hundred}, autoFill},
	}}
	for i := 0; i < 6; i++ {
		grid.Children = append(grid.Children, &Node{Style: Style{Height: Px(10)}})
	}
	LayoutGrid(grid, Loose(400, Unbounded), NewLayoutContext(800, 600, 16))
	if x, y := grid.Children[3].Rect.X, grid.Children[3].Rect.Y; x != 300 || y != 0 {
		t.Errorf("item 3: expected (300, 0), got (%v, %v)", x, y)
	}
	if x, y := grid.Children[4].Rect.X, grid.Children[4].Rect.Y; x != 0 || y != 10 {
		t.Errorf("item 4: expected (0, 10) on the second row, got (%v, %v)", x, y)
	}
}

// TestRepeatExpansionCappedAtMaxTracks checks that repeat() expansion is
// bounded by the total number of tracks (gridMaxTracks), not by the number
// of repetitions: repeat(10000, [500 × 1px]) used to emit 5,000,000 tracks.
// The cap applies to fixed counts, to auto-fill / auto-fit, and to the
// auto-fit markers, and an item placed beyond the cap lands on the last
// track.
//
// https://www.w3.org/TR/css-grid-1/#overlarge-grids
func TestRepeatExpansionCappedAtMaxTracks(t *testing.T) {
	pattern := make([]GridTrack, 500)
	for i := range pattern {
		pattern[i] = FixedTrack(Px(1))
	}

	// Fixed count: 10000 × 500 tracks is truncated to gridMaxTracks.
	start := time.Now()
	tracks, _, _ := expandAutoRepeatTracks([]RepeatTrack{{Count: 10000, Tracks: pattern}}, nil, 400, 0)
	if len(tracks) != gridMaxTracks {
		t.Errorf("repeat(10000, [500 tracks]): expected %d tracks, got %d", gridMaxTracks, len(tracks))
	}

	// Auto-fill with a wide pattern against a huge axis: same cap, and the
	// auto range covers the whole result.
	tracks, autoStart, autoEnd := expandAutoRepeatTracks([]RepeatTrack{{Count: RepeatCountAutoFill, Tracks: pattern}}, nil, 1e9, 0)
	if len(tracks) != gridMaxTracks || autoStart != 0 || autoEnd != gridMaxTracks {
		t.Errorf("auto-fill [500 tracks] in 1e9px: expected %d tracks and range [0, %d), got %d tracks, [%d, %d)", gridMaxTracks, gridMaxTracks, len(tracks), autoStart, autoEnd)
	}
	if got := calculateAutoRepeatCount(RepeatTrack{Count: RepeatCountAutoFill, Tracks: pattern}, 1e300, 0); got != gridMaxTracks/500 {
		t.Errorf("auto-fill count for a 500-track pattern: expected %d, got %d", gridMaxTracks/500, got)
	}

	// Truncation is at track granularity: 9990 explicit tracks + repeat(3,
	// [5 tracks]) stops after 10 of the 15 repeated tracks.
	explicit := make([]GridTrack, 9990)
	for i := range explicit {
		explicit[i] = FixedTrack(Px(1))
	}
	tracks, _, _ = expandAutoRepeatTracks([]RepeatTrack{{Count: 3, Tracks: pattern[:5]}}, explicit, Unbounded, 0)
	if len(tracks) != gridMaxTracks {
		t.Errorf("9990 explicit + repeat(3, [5 tracks]): expected %d tracks, got %d", gridMaxTracks, len(tracks))
	}

	// A pattern that is itself wider than the cap is cut, and later patterns
	// get nothing.
	wide := make([]GridTrack, gridMaxTracks+7)
	for i := range wide {
		wide[i] = FixedTrack(Px(1))
	}
	tracks, _, _ = expandAutoRepeatTracks([]RepeatTrack{{Count: 2, Tracks: wide}, {Count: 3, Tracks: pattern}}, nil, Unbounded, 0)
	if len(tracks) != gridMaxTracks {
		t.Errorf("oversized pattern: expected %d tracks, got %d", gridMaxTracks, len(tracks))
	}

	// The auto-fit markers agree with the truncated expansion.
	template := make([]GridTrack, 9995)
	for i := range template {
		template[i] = FixedTrack(Px(1))
	}
	expanded, autoFit := gridExpandTemplate(template, []RepeatTrack{{Count: RepeatCountAutoFit, Tracks: []GridTrack{FixedTrack(Px(100))}}}, 1e6, 0, NewLayoutContext(800, 600, 16), 16)
	if len(expanded) != gridMaxTracks || len(autoFit) != gridMaxTracks {
		t.Fatalf("auto-fit past the cap: expected %d tracks and markers, got %d and %d", gridMaxTracks, len(expanded), len(autoFit))
	}
	for i := range autoFit {
		if autoFit[i] != (i >= 9995) {
			t.Fatalf("autoFit[%d] = %v, expected %v", i, autoFit[i], i >= 9995)
		}
	}
	if d := time.Since(start); d > 100*time.Millisecond {
		t.Errorf("capped expansions took %v, expected well under 100ms", d)
	}

	// End to end: the whole grid lays out quickly with 10,000 1px columns
	// (gap 0), and an item asking for column 20000 is clamped to the last
	// line, X = 9999, instead of growing the grid past the cap.
	grid := &Node{Style: Style{
		Display:                   DisplayGrid,
		Width:                     Px(10000),
		GridTemplateColumnsRepeat: []RepeatTrack{{Count: 10000, Tracks: pattern}},
	}}
	grid.Children = append(grid.Children, &Node{Style: Style{Height: Px(10), GridColumnStart: 20000, GridColumnEnd: 20001}})
	start = time.Now()
	size := LayoutGrid(grid, Loose(10000, Unbounded), NewLayoutContext(800, 600, 16))
	if d := time.Since(start); d > 100*time.Millisecond {
		t.Errorf("repeat(10000, [500 × 1px]) layout took %v, expected well under 100ms", d)
	}
	if size.Width != 10000 {
		t.Errorf("container width: expected 10000, got %v", size.Width)
	}
	if x := grid.Children[0].Rect.X; x != 9999 {
		t.Errorf("item at column 20000: expected X = 9999 (last track), got %v", x)
	}
	if w := grid.Children[0].Rect.Width; w != 1 {
		t.Errorf("item at column 20000: expected width 1, got %v", w)
	}
	// gridNormalizeSpan agrees: the clamped line is the last of the 10,000.
	if span := gridNormalizeSpan(20000, 20001); span.start != gridMaxTracks-1 || span.span != 1 {
		t.Errorf("gridNormalizeSpan(20000, 20001) = %+v, expected start %d span 1", span, gridMaxTracks-1)
	}
}

// TestAutoFillTracksHelper tests the AutoFillTracks API helper
func TestAutoFillTracksHelper(t *testing.T) {
	repeat := AutoFillTracks(FixedTrack(Px(100)))

	if repeat.Count != RepeatCountAutoFill {
		t.Errorf("AutoFillTracks should set Count to RepeatCountAutoFill")
	}

	if len(repeat.Tracks) != 1 {
		t.Errorf("Expected 1 track, got %d", len(repeat.Tracks))
	}

	if repeat.Tracks[0].MinSize.Value != 100 {
		t.Errorf("Expected track size 100, got %.0f", repeat.Tracks[0].MinSize.Value)
	}
}

// TestAutoFitTracksHelper tests the AutoFitTracks API helper
func TestAutoFitTracksHelper(t *testing.T) {
	repeat := AutoFitTracks(FixedTrack(Px(100)))

	if repeat.Count != RepeatCountAutoFit {
		t.Errorf("AutoFitTracks should set Count to RepeatCountAutoFit")
	}

	if len(repeat.Tracks) != 1 {
		t.Errorf("Expected 1 track, got %d", len(repeat.Tracks))
	}

	if repeat.Tracks[0].MinSize.Value != 100 {
		t.Errorf("Expected track size 100, got %.0f", repeat.Tracks[0].MinSize.Value)
	}
}

// TestValidateAutoRepeatTracks tests validation of auto-repeat patterns:
// only <fixed-size> tracks are allowed (css-grid-1 §7.2.3).
//
// https://www.w3.org/TR/css-grid-1/#typedef-fixed-size
func TestValidateAutoRepeatTracks(t *testing.T) {
	valid := map[string]GridTrack{
		"fixed length":            FixedTrack(Px(100)),
		"fixed 0px":               FixedTrack(Px(0)),
		"vw length":               FixedTrack(Vw(10)),
		"em length":               FixedTrack(Em(5)),
		"minmax(fixed, auto)":     MinMaxTrack(Px(50), PxUnbounded),
		"minmax(0px, fixed)":      MinMaxTrack(Px(0), Px(100)),
		"minmax(unset min, 50px)": {MaxSize: Px(50)},
	}
	for name, track := range valid {
		if !validateAutoRepeatTracks(RepeatTrack{Count: RepeatCountAutoFill, Tracks: []GridTrack{track}}) {
			t.Errorf("%s should be valid for auto-repeat", name)
		}
	}

	invalid := map[string]GridTrack{
		"fr":               FractionTrack(1),
		"min-content":      MinContentTrack(),
		"max-content":      MaxContentTrack(),
		"fit-content":      FitContentTrack(300),
		"auto":             AutoTrack(),
		"zero value":       {},
		"unbounded length": FixedTrack(UnboundedLength()),
	}
	for name, track := range invalid {
		if validateAutoRepeatTracks(RepeatTrack{Count: RepeatCountAutoFill, Tracks: []GridTrack{track}}) {
			t.Errorf("%s should be invalid for auto-repeat", name)
		}
	}

	// One invalid track invalidates the whole pattern.
	mixed := RepeatTrack{Count: RepeatCountAutoFill, Tracks: []GridTrack{FixedTrack(Px(100)), AutoTrack()}}
	if validateAutoRepeatTracks(mixed) {
		t.Error("a pattern containing an auto track should be invalid")
	}

	// End to end: repeat(auto-fill, auto) is ignored, so the two items stack
	// in the single implicit column instead of filling the 500px row.
	grid := &Node{Style: Style{
		Display:                   DisplayGrid,
		Width:                     Px(500),
		GridTemplateColumnsRepeat: []RepeatTrack{{Count: RepeatCountAutoFill, Tracks: []GridTrack{AutoTrack()}}},
	}}
	grid.Children = []*Node{{Style: Style{Height: Px(10)}}, {Style: Style{Height: Px(10)}}}
	LayoutGrid(grid, Loose(500, Unbounded), NewLayoutContext(800, 600, 16))
	if x, y := grid.Children[1].Rect.X, grid.Children[1].Rect.Y; x != 0 || y != 10 {
		t.Errorf("repeat(auto-fill, auto) should be ignored; item 1 at (%v, %v), expected (0, 10)", x, y)
	}
}

// TestAutoRepeatCountEdgeCases tests edge cases
func TestAutoRepeatCountEdgeCases(t *testing.T) {
	// Empty pattern
	emptyRepeat := RepeatTrack{
		Count:  RepeatCountAutoFill,
		Tracks: []GridTrack{},
	}
	count := calculateAutoRepeatCount(emptyRepeat, 500, 10)
	if count != 0 {
		t.Errorf("Empty pattern should return 0, got %d", count)
	}

	// Very small available space
	repeat := RepeatTrack{
		Count:  RepeatCountAutoFill,
		Tracks: []GridTrack{FixedTrack(Px(100))},
	}
	count = calculateAutoRepeatCount(repeat, 1, 0)
	if count != 1 {
		t.Errorf("Very small space should return minimum 1, got %d", count)
	}

	// Very large available space
	count = calculateAutoRepeatCount(repeat, 10000, 10)
	if count <= 0 {
		t.Errorf("Large space should return positive count, got %d", count)
	}
}

// TestAutoRepeatMultiplePatterns tests an auto-fill pattern followed by a
// fixed-count pattern: repeat(auto-fill, 100px) repeat(3, 50px) in 500px with
// 10px gaps. The three 50px tracks and their gaps use 150 + 30 = 180, leaving
// 320 for floor((320 + 10) / 110) = 3 repetitions: 6 tracks, 480px in total.
func TestAutoRepeatMultiplePatterns(t *testing.T) {
	repeats := []RepeatTrack{
		{
			Count:  RepeatCountAutoFill,
			Tracks: []GridTrack{FixedTrack(Px(100))},
		},
		{
			Count:  3, // Regular repeat
			Tracks: []GridTrack{FixedTrack(Px(50))},
		},
	}

	result, autoStart, autoEnd := expandAutoRepeatTracks(repeats, []GridTrack{}, 500, 10)

	if len(result) != 6 {
		t.Errorf("Expected 6 tracks (3 auto-fill + 3 fixed), got %d", len(result))
	}
	if autoStart != 0 || autoEnd != 3 {
		t.Errorf("Expected auto-repeat range [0, 3), got [%d, %d)", autoStart, autoEnd)
	}
	for i := 3; i < len(result) && i < 6; i++ {
		if result[i].MinSize.Value != 50 {
			t.Errorf("Track %d should be the 50px fixed pattern, got %.0f", i, result[i].MinSize.Value)
		}
	}
}
