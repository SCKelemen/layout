package layout

import "math"

// Grid auto-repeat algorithms for CSS Grid Layout auto-fill and auto-fit.
//
// Implements dynamic grid track generation based on container size.
//
// Algorithm based on CSS Grid Layout Module Level 1:
// - §7.2.3: Repeat-to-fill (auto-fill and auto-fit)
//
// See: https://www.w3.org/TR/css-grid-1/#auto-repeat

// gridMaxAutoRepeat is the hard cap on the number of repetitions produced by
// auto-fill / auto-fit. It keeps expandAutoRepeatTracks bounded even when the
// arithmetic would otherwise produce an enormous count, and matches
// gridMaxTracks so a repeated pattern can never exceed the implicit grid cap.
const gridMaxAutoRepeat = gridMaxTracks

// autoRepeatTrackSize returns the size a track contributes when counting
// auto-repeat repetitions.
//
// CSS Grid Layout Module Level 1 §7.2.3.2: each track is treated as its max
// track sizing function if that is definite, otherwise as its min track
// sizing function, flooring the max by the min.
//
// See: https://www.w3.org/TR/css-grid-1/#auto-repeat
func autoRepeatTrackSize(track GridTrack) float64 {
	minSize := track.MinSize.Value
	if minSize < 0 || minSize >= Unbounded {
		minSize = 0
	}
	maxSize := track.MaxSize.Value
	maxDefinite := track.MaxSize.Unit != "" && maxSize >= 0 && maxSize < Unbounded
	if !maxDefinite {
		return minSize
	}
	return math.Max(minSize, maxSize)
}

// calculateAutoRepeatCount calculates how many times to repeat a track pattern
// based on available space.
//
// Formula from CSS Grid Layout Module Level 1 §7.2.3.2:
// floor((availableSize + gap) / (repetitionSize + gap))
//
// When the available size is indefinite the pattern repeats exactly once, as
// the spec requires ("Otherwise, the specified track list repeats only once").
// The result is computed in float64, range-checked, and clamped to
// gridMaxAutoRepeat before the conversion to int so a huge or non-finite
// available size can never yield a negative, overflowing, or unbounded count.
//
// Parameters:
//   - repeat: The RepeatTrack pattern to repeat
//   - availableSize: The available space in the grid axis
//   - gap: The gap between tracks
//
// Returns: The number of repetitions (minimum 1 for a non-empty pattern)
//
// See: https://www.w3.org/TR/css-grid-1/#auto-repeat
func calculateAutoRepeatCount(repeat RepeatTrack, availableSize, gap float64) int {
	if len(repeat.Tracks) == 0 {
		return 0
	}

	// Indefinite (or invalid) available size: a single repetition.
	if math.IsNaN(availableSize) || math.IsInf(availableSize, 0) || availableSize >= Unbounded {
		return 1
	}
	if math.IsNaN(gap) || math.IsInf(gap, 0) || gap < 0 {
		gap = 0
	}

	// Calculate the size of one repetition
	repetitionSize := 0.0
	for _, track := range repeat.Tracks {
		repetitionSize += autoRepeatTrackSize(track)
	}

	// Add gaps within the repetition (between tracks in the pattern)
	if len(repeat.Tracks) > 1 {
		repetitionSize += gap * float64(len(repeat.Tracks)-1)
	}

	if repetitionSize <= 0 {
		// Can't calculate with zero or negative size
		return 1
	}

	// Apply the CSS formula: floor((availableSize + gap) / (repetitionSize + gap))
	// The extra gap accounts for the gap after the last repetition
	countF := math.Floor((availableSize + gap) / (repetitionSize + gap))

	// Explicit range check before converting to int.
	if math.IsNaN(countF) || countF < 1 {
		return 1
	}
	if countF > float64(gridMaxAutoRepeat) {
		return gridMaxAutoRepeat
	}
	return int(countF)
}

// expandAutoRepeatTracks expands auto-fill or auto-fit track patterns
// into concrete track definitions based on available space.
//
// Parameters:
//   - repeats: Array of RepeatTrack patterns (mix of auto-fill/auto-fit and regular)
//   - explicitTracks: Explicitly defined tracks (non-repeating)
//   - availableSize: Available space in the grid axis (>= Unbounded when indefinite)
//   - gap: Gap between tracks
//
// Returns: Expanded array of GridTrack definitions
//
// See: https://www.w3.org/TR/css-grid-1/#auto-repeat
func expandAutoRepeatTracks(repeats []RepeatTrack, explicitTracks []GridTrack, availableSize, gap float64) []GridTrack {
	// Start with explicit tracks
	result := make([]GridTrack, 0, len(explicitTracks)*2)
	result = append(result, explicitTracks...)

	// Calculate space used by explicit tracks
	usedSpace := 0.0
	for _, track := range explicitTracks {
		usedSpace += autoRepeatTrackSize(track)
	}

	// Add gaps for explicit tracks
	if len(explicitTracks) > 1 {
		usedSpace += gap * float64(len(explicitTracks)-1)
	}

	// Remaining space for auto-repeat tracks. An indefinite available size
	// stays indefinite so each auto-repeat pattern repeats exactly once.
	remainingSpace := availableSize
	if availableSize < Unbounded {
		remainingSpace = availableSize - usedSpace
		if remainingSpace < 0 {
			remainingSpace = 0
		}
	}

	// Expand each auto-repeat pattern
	for _, repeat := range repeats {
		if repeat.Count == RepeatCountAutoFill || repeat.Count == RepeatCountAutoFit {
			// Calculate how many repetitions fit (bounded by gridMaxAutoRepeat)
			count := calculateAutoRepeatCount(repeat, remainingSpace, gap)

			// Expand the pattern
			for i := 0; i < count; i++ {
				result = append(result, repeat.Tracks...)
			}

			// Update remaining space
			if remainingSpace < Unbounded {
				for _, track := range repeat.Tracks {
					remainingSpace -= autoRepeatTrackSize(track) * float64(count)
				}
				if count > 0 && len(repeat.Tracks) > 0 {
					remainingSpace -= gap * float64(count*len(repeat.Tracks)-1)
				}
				if remainingSpace < 0 {
					remainingSpace = 0
				}
			}
		} else if repeat.Count > 0 {
			// Regular repeat (not auto-fill/auto-fit), bounded like auto-repeat
			count := repeat.Count
			if count > gridMaxAutoRepeat {
				count = gridMaxAutoRepeat
			}
			for i := 0; i < count; i++ {
				result = append(result, repeat.Tracks...)
			}
		}
	}

	return result
}

// collapseAutoFitEmptyTracks collapses empty tracks to zero size for auto-fit.
// This is the key difference between auto-fill and auto-fit:
// - auto-fill: keeps all tracks, even if empty
// - auto-fit: collapses empty tracks to zero size
//
// Parameters:
//   - tracks: Array of grid tracks
//   - items: Grid items that have been placed
//   - isColumn: true if tracks are columns, false if rows
//
// Returns: Modified array of tracks with empty auto-fit tracks collapsed
func collapseAutoFitEmptyTracks(tracks []GridTrack, items []*gridItem, isColumn bool) []GridTrack {
	if len(tracks) == 0 {
		return tracks
	}

	// Track which tracks have content
	hasContent := make([]bool, len(tracks))

	// Mark tracks that contain items
	for _, item := range items {
		var start, end int
		if isColumn {
			start = item.colStart
			end = item.colEnd
		} else {
			start = item.rowStart
			end = item.rowEnd
		}

		// Clamp to track bounds
		if start < 0 {
			start = 0
		}
		if end > len(tracks) {
			end = len(tracks)
		}

		// Mark all tracks this item spans
		for i := start; i < end; i++ {
			hasContent[i] = true
		}
	}

	// Collapse empty tracks
	result := make([]GridTrack, len(tracks))
	for i, track := range tracks {
		if hasContent[i] {
			// Keep track size
			result[i] = track
		} else {
			// Collapse to zero (empty auto-fit track)
			result[i] = GridTrack{
				MinSize:  Px(0),
				MaxSize:  Px(0),
				Fraction: track.Fraction, // Preserve fraction in case it matters
			}
		}
	}

	return result
}

// isAutoRepeatTrack checks if a RepeatTrack uses auto-fill or auto-fit.
func isAutoRepeatTrack(repeat RepeatTrack) bool {
	return repeat.Count == RepeatCountAutoFill || repeat.Count == RepeatCountAutoFit
}

// gridExpandTemplate builds the explicit track list of one axis from the
// template tracks and its repeat() patterns.
//
// The explicit tracks come first, followed by every valid repeat pattern in
// order (CSS Grid Layout Module Level 1 §7.2.3). Patterns are filtered per
// the grammar before expansion:
//
//   - a pattern with no tracks, a count of 0, or an unknown negative count is
//     ignored;
//   - an auto-fill / auto-fit pattern may only contain fixed sizes
//     (validateAutoRepeatTracks); an invalid pattern is ignored;
//   - only one auto-repeat is allowed per track list (§7.2.3.1); a second
//     one is ignored.
//
// availableSize is the container's content size in this axis, or Unbounded
// when indefinite, in which case every auto-repeat produces one repetition
// (§7.2.3.2). Pattern tracks are resolved to pixels (em/rem/...) before
// counting so the repetition count matches the sizes the track sizing
// algorithm later uses. Fixed repetition counts and auto-repeat counts are
// both capped at gridMaxAutoRepeat, so the result is always bounded.
//
// The second result marks, per resulting track, whether it belongs to an
// auto-fit pattern and is therefore collapsible when empty; it is nil when
// there is no auto-fit pattern.
//
// See: https://www.w3.org/TR/css-grid-1/#repeat-notation
// See: https://www.w3.org/TR/css-grid-1/#auto-repeat
func gridExpandTemplate(template []GridTrack, repeats []RepeatTrack, availableSize, gap float64, ctx *LayoutContext, currentFontSize float64) ([]GridTrack, []bool) {
	if len(repeats) == 0 {
		return template, nil
	}

	valid := make([]RepeatTrack, 0, len(repeats))
	autoIndex := -1 // index in valid of the single auto-repeat, if any
	for _, repeat := range repeats {
		if len(repeat.Tracks) == 0 {
			continue
		}
		switch {
		case isAutoRepeatTrack(repeat):
			if autoIndex >= 0 || !validateAutoRepeatTracks(repeat) {
				continue
			}
			autoIndex = len(valid)
		case repeat.Count > 0:
			// Fixed count; expandAutoRepeatTracks caps it.
		default:
			continue
		}
		resolved := RepeatTrack{Count: repeat.Count, Tracks: make([]GridTrack, len(repeat.Tracks))}
		for i, track := range repeat.Tracks {
			resolved.Tracks[i] = gridResolveTrackLengths(track, ctx, currentFontSize)
		}
		valid = append(valid, resolved)
	}
	if len(valid) == 0 {
		return template, nil
	}

	tracks := expandAutoRepeatTracks(valid, template, availableSize, gap)
	if autoIndex < 0 || valid[autoIndex].Count != RepeatCountAutoFit {
		return tracks, nil
	}

	// Locate the auto-fit repetitions: they start after the explicit tracks
	// and the fixed-count patterns that precede them, and run for however
	// many tracks the fixed-count patterns after them leave over.
	autoStart := len(template)
	autoEnd := len(tracks)
	for i, repeat := range valid {
		if i == autoIndex {
			continue
		}
		count := repeat.Count
		if count > gridMaxAutoRepeat {
			count = gridMaxAutoRepeat
		}
		n := count * len(repeat.Tracks)
		if i < autoIndex {
			autoStart += n
		} else {
			autoEnd -= n
		}
	}
	if autoStart < 0 {
		autoStart = 0
	}
	if autoEnd > len(tracks) {
		autoEnd = len(tracks)
	}
	autoFit := make([]bool, len(tracks))
	for i := autoStart; i < autoEnd; i++ {
		autoFit[i] = true
	}
	return tracks, autoFit
}

// gridResolveTrackLengths returns a copy of a track whose min and max sizing
// functions are resolved to pixels when they are lengths. Keywords (auto,
// min-content, max-content, unbounded) and flex factors are kept as-is.
func gridResolveTrackLengths(track GridTrack, ctx *LayoutContext, currentFontSize float64) GridTrack {
	resolve := func(l Length) Length {
		if l.Unit == "" || l.Unit == UnboundedUnit {
			return l
		}
		v := ResolveLength(l, ctx, currentFontSize)
		if v == SizeMinContent || v == SizeMaxContent || v == SizeFitContent || v < 0 || v >= Unbounded || math.IsNaN(v) {
			return l
		}
		return Px(v)
	}
	track.MinSize = resolve(track.MinSize)
	track.MaxSize = resolve(track.MaxSize)
	return track
}

// gridCollapseAutoFitTracks removes the empty auto-fit tracks of one axis and
// remaps the items' line numbers accordingly.
//
// CSS Grid Layout Module Level 1 §7.2.3.2: with auto-fit, "any empty
// repeated tracks are collapsed. A collapsed track is treated as having a
// fixed track sizing function of 0px, and the gutters on either side of it
// ... collapse." A 0px track whose gutters collapse contributes nothing to
// the axis, so removing it from the track list is equivalent and keeps the
// uniform gap between the remaining tracks: trailing empty tracks leave no
// trailing gaps, and an empty track between two occupied ones leaves a
// single gap between them. Items never occupy a collapsed track, so each
// item's start and end lines shift down by the number of collapsed tracks
// before them and every span stays at least one track wide.
//
// autoFit marks the tracks produced by an auto-fit pattern (see
// gridExpandTemplate); tracks outside it, including implicit tracks added by
// placement, are never collapsed. Returns the tracks unchanged when nothing
// collapses.
//
// See: https://www.w3.org/TR/css-grid-1/#auto-repeat
// See: https://www.w3.org/TR/css-grid-1/#collapsed-track
func gridCollapseAutoFitTracks(tracks []GridTrack, autoFit []bool, items []*gridItem, isColumn bool) []GridTrack {
	n := len(tracks)
	if n == 0 || len(autoFit) == 0 {
		return tracks
	}

	occupied := make([]bool, n)
	for _, item := range items {
		start, end := item.rowStart, item.rowEnd
		if isColumn {
			start, end = item.colStart, item.colEnd
		}
		if start < 0 {
			start = 0
		}
		if end > n {
			end = n
		}
		for i := start; i < end; i++ {
			occupied[i] = true
		}
	}

	// removedBefore[l] is the number of collapsed tracks before line l.
	removedBefore := make([]int, n+1)
	collapsedCount := 0
	for i := 0; i < n; i++ {
		removedBefore[i] = collapsedCount
		if i < len(autoFit) && autoFit[i] && !occupied[i] {
			collapsedCount++
		}
	}
	removedBefore[n] = collapsedCount
	if collapsedCount == 0 {
		return tracks
	}

	kept := make([]GridTrack, 0, n-collapsedCount)
	for i := 0; i < n; i++ {
		if removedBefore[i+1] == removedBefore[i] {
			kept = append(kept, tracks[i])
		}
	}

	remap := func(line int) int {
		if line < 0 {
			return 0
		}
		if line > n {
			line = n
		}
		return line - removedBefore[line]
	}
	for _, item := range items {
		if isColumn {
			item.colStart = remap(item.colStart)
			item.colEnd = remap(item.colEnd)
		} else {
			item.rowStart = remap(item.rowStart)
			item.rowEnd = remap(item.rowEnd)
		}
	}
	return kept
}

// validateAutoRepeatTracks validates that auto-repeat patterns only use
// fixed-size tracks (no fr units, no intrinsic sizes).
//
// According to CSS spec, auto-repeat can only contain:
// - Fixed lengths (e.g., 100px)
// - Percentage values (treated as fixed when container size is known)
// - minmax() with fixed min/max
//
// Returns: true if valid, false otherwise
func validateAutoRepeatTracks(repeat RepeatTrack) bool {
	for _, track := range repeat.Tracks {
		// Check for fractional units (not allowed in auto-repeat)
		if track.Fraction > 0 {
			return false
		}

		// Check for intrinsic sizing keywords (not allowed in auto-repeat)
		if track.MaxSize.Value == SizeMinContent || track.MaxSize.Value == SizeMaxContent {
			return false
		}

		// Fit-content is also not allowed in auto-repeat
		if track.Fraction == -1 {
			return false
		}
	}

	return true
}
