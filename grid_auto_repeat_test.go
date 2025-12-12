package layout

import "testing"

// TestCalculateAutoRepeatCount tests the calculation of repeat count
func TestCalculateAutoRepeatCount(t *testing.T) {
	// Pattern: [100px]
	repeat := RepeatTrack{
		Count:  RepeatCountAutoFill,
		Tracks: []GridTrack{FixedTrack(100)},
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
			FixedTrack(100),
			FixedTrack(50),
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
			Tracks: []GridTrack{FixedTrack(100)},
		},
	}

	// Available: 350px, gap: 10px
	// Should create 3 tracks of 100px
	result := expandAutoRepeatTracks(repeats, []GridTrack{}, 350, 10)

	if len(result) != 3 {
		t.Errorf("Expected 3 tracks, got %d", len(result))
	}

	for i, track := range result {
		if track.MinSize != 100 || track.MaxSize != 100 {
			t.Errorf("Track %d: expected 100px, got min=%.0f max=%.0f",
				i, track.MinSize, track.MaxSize)
		}
	}
}

// TestExpandAutoRepeatTracksWithExplicit tests auto-repeat with explicit tracks
func TestExpandAutoRepeatTracksWithExplicit(t *testing.T) {
	explicit := []GridTrack{
		FixedTrack(200), // Explicit track takes 200px
	}

	repeats := []RepeatTrack{
		{
			Count:  RepeatCountAutoFill,
			Tracks: []GridTrack{FixedTrack(100)},
		},
	}

	// Available: 550px, gap: 10px
	// Explicit uses: 200px
	// Remaining: 550 - 200 = 350px
	// Auto-repeat: floor((350 + 10) / (100 + 10)) = 3
	// Total tracks: 1 explicit + 3 auto = 4
	result := expandAutoRepeatTracks(repeats, explicit, 550, 10)

	if len(result) != 4 {
		t.Errorf("Expected 4 tracks (1 explicit + 3 auto), got %d", len(result))
	}

	// First track should be explicit (200px)
	if result[0].MinSize != 200 {
		t.Errorf("First track should be 200px, got %.0f", result[0].MinSize)
	}

	// Next 3 should be auto-repeated (100px each)
	for i := 1; i < 4; i++ {
		if result[i].MinSize != 100 {
			t.Errorf("Track %d should be 100px, got %.0f", i, result[i].MinSize)
		}
	}
}

// TestCollapseAutoFitEmptyTracks tests collapsing empty tracks for auto-fit
func TestCollapseAutoFitEmptyTracks(t *testing.T) {
	tracks := []GridTrack{
		FixedTrack(100),
		FixedTrack(100),
		FixedTrack(100),
		FixedTrack(100),
	}

	// Items only in tracks 0 and 2
	items := []*gridItem{
		{colStart: 0, colEnd: 1}, // Track 0
		{colStart: 2, colEnd: 3}, // Track 2
		// Tracks 1 and 3 are empty
	}

	result := collapseAutoFitEmptyTracks(tracks, items, true)

	if len(result) != 4 {
		t.Errorf("Expected 4 tracks, got %d", len(result))
	}

	// Track 0: has content, should keep size
	if result[0].MinSize != 100 {
		t.Errorf("Track 0 should be 100px, got %.0f", result[0].MinSize)
	}

	// Track 1: empty, should be collapsed to 0
	if result[1].MinSize != 0 || result[1].MaxSize != 0 {
		t.Errorf("Track 1 should be collapsed to 0, got min=%.0f max=%.0f",
			result[1].MinSize, result[1].MaxSize)
	}

	// Track 2: has content, should keep size
	if result[2].MinSize != 100 {
		t.Errorf("Track 2 should be 100px, got %.0f", result[2].MinSize)
	}

	// Track 3: empty, should be collapsed to 0
	if result[3].MinSize != 0 || result[3].MaxSize != 0 {
		t.Errorf("Track 3 should be collapsed to 0, got min=%.0f max=%.0f",
			result[3].MinSize, result[3].MaxSize)
	}
}

// TestAutoFillTracksHelper tests the AutoFillTracks API helper
func TestAutoFillTracksHelper(t *testing.T) {
	repeat := AutoFillTracks(FixedTrack(100))

	if repeat.Count != RepeatCountAutoFill {
		t.Errorf("AutoFillTracks should set Count to RepeatCountAutoFill")
	}

	if len(repeat.Tracks) != 1 {
		t.Errorf("Expected 1 track, got %d", len(repeat.Tracks))
	}

	if repeat.Tracks[0].MinSize != 100 {
		t.Errorf("Expected track size 100, got %.0f", repeat.Tracks[0].MinSize)
	}
}

// TestAutoFitTracksHelper tests the AutoFitTracks API helper
func TestAutoFitTracksHelper(t *testing.T) {
	repeat := AutoFitTracks(FixedTrack(100))

	if repeat.Count != RepeatCountAutoFit {
		t.Errorf("AutoFitTracks should set Count to RepeatCountAutoFit")
	}

	if len(repeat.Tracks) != 1 {
		t.Errorf("Expected 1 track, got %d", len(repeat.Tracks))
	}

	if repeat.Tracks[0].MinSize != 100 {
		t.Errorf("Expected track size 100, got %.0f", repeat.Tracks[0].MinSize)
	}
}

// TestValidateAutoRepeatTracks tests validation of auto-repeat patterns
func TestValidateAutoRepeatTracks(t *testing.T) {
	// Valid: fixed-size tracks
	validRepeat := RepeatTrack{
		Count:  RepeatCountAutoFill,
		Tracks: []GridTrack{FixedTrack(100)},
	}
	if !validateAutoRepeatTracks(validRepeat) {
		t.Error("Fixed-size tracks should be valid for auto-repeat")
	}

	// Invalid: fractional tracks
	invalidFr := RepeatTrack{
		Count:  RepeatCountAutoFill,
		Tracks: []GridTrack{FractionTrack(1)},
	}
	if validateAutoRepeatTracks(invalidFr) {
		t.Error("Fractional tracks should be invalid for auto-repeat")
	}

	// Invalid: min-content tracks
	invalidMinContent := RepeatTrack{
		Count:  RepeatCountAutoFill,
		Tracks: []GridTrack{MinContentTrack()},
	}
	if validateAutoRepeatTracks(invalidMinContent) {
		t.Error("Min-content tracks should be invalid for auto-repeat")
	}

	// Invalid: fit-content tracks
	invalidFitContent := RepeatTrack{
		Count:  RepeatCountAutoFill,
		Tracks: []GridTrack{FitContentTrack(300)},
	}
	if validateAutoRepeatTracks(invalidFitContent) {
		t.Error("Fit-content tracks should be invalid for auto-repeat")
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
		Tracks: []GridTrack{FixedTrack(100)},
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

// TestAutoRepeatMultiplePatterns tests multiple repeat patterns
func TestAutoRepeatMultiplePatterns(t *testing.T) {
	repeats := []RepeatTrack{
		{
			Count:  RepeatCountAutoFill,
			Tracks: []GridTrack{FixedTrack(100)},
		},
		{
			Count:  3, // Regular repeat
			Tracks: []GridTrack{FixedTrack(50)},
		},
	}

	// Available: 500px, gap: 10px
	// Auto-fill should calculate first, then regular repeat adds 3 tracks
	result := expandAutoRepeatTracks(repeats, []GridTrack{}, 500, 10)

	// Should have auto-fill tracks + 3 regular repeat tracks
	// The exact count depends on how much space auto-fill uses
	if len(result) < 3 {
		t.Errorf("Expected at least 3 tracks, got %d", len(result))
	}
}

// TestCollapseAutoFitSpanningItems tests collapsing with spanning items
func TestCollapseAutoFitSpanningItems(t *testing.T) {
	tracks := []GridTrack{
		FixedTrack(100),
		FixedTrack(100),
		FixedTrack(100),
		FixedTrack(100),
	}

	// Item spans tracks 1-3
	items := []*gridItem{
		{colStart: 1, colEnd: 3}, // Spans tracks 1 and 2
	}

	result := collapseAutoFitEmptyTracks(tracks, items, true)

	// Track 0: empty, should be collapsed
	if result[0].MinSize != 0 {
		t.Errorf("Track 0 should be collapsed, got %.0f", result[0].MinSize)
	}

	// Tracks 1-2: have content (spanned by item), should keep size
	if result[1].MinSize != 100 || result[2].MinSize != 100 {
		t.Errorf("Tracks 1-2 should keep size 100")
	}

	// Track 3: empty, should be collapsed
	if result[3].MinSize != 0 {
		t.Errorf("Track 3 should be collapsed, got %.0f", result[3].MinSize)
	}
}
