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

// Every repeat() expansion is bounded by gridMaxTracks, the same cap the
// implicit grid uses: the tracks emitted for one axis by the explicit
// template plus all of its repeat() patterns never exceed gridMaxTracks in
// total. The cap applies to tracks, not repetitions, so a pattern with many
// tracks cannot multiply past it (repeat(10000, [500 tracks]) emits 10,000
// tracks, not 5,000,000). See gridAppendRepeat.

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

// gridRepeatMaxCount returns the largest repetition count of a pattern whose
// expansion alone stays within gridMaxTracks (at least 1 for a non-empty
// pattern; the final truncation in gridAppendRepeat handles a pattern that
// is itself larger than the cap).
func gridRepeatMaxCount(pattern []GridTrack) int {
	if len(pattern) == 0 {
		return 0
	}
	n := gridMaxTracks / len(pattern)
	if n < 1 {
		n = 1
	}
	return n
}

// calculateAutoRepeatCount calculates how many times to repeat a track pattern
// in the space left for it.
//
// CSS Grid Layout Module Level 1 §7.2.3.2: the repetition count is the
// largest positive integer that does not cause the grid to overflow its grid
// container, treating each track as its max sizing function if definite and
// its min sizing function otherwise (autoRepeatTrackSize). With one gap
// between consecutive tracks this is
//
//	floor((availableSize + gap) / (repetitionSize + gap))
//
// where availableSize is the space remaining for the repeated block after
// every other track in the axis, their gaps, and the gap between them and
// the repeated block have been subtracted (expandAutoRepeatTracks does that
// subtraction). If even one repetition overflows, the pattern still repeats
// once ("If any number of repetitions would overflow, then 1 repetition").
//
// When the available size is indefinite the pattern repeats exactly once, as
// the spec requires ("Otherwise, the specified track list repeats only once").
// The result is computed in float64, range-checked, and clamped before the
// conversion to int so a huge or non-finite available size can never yield a
// negative, overflowing, or unbounded count; the clamp keeps count ×
// len(Tracks) within gridMaxTracks.
//
// Parameters:
//   - repeat: The RepeatTrack pattern to repeat
//   - availableSize: The space left for the repeated tracks in the grid axis
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
	maxCount := gridRepeatMaxCount(repeat.Tracks)
	if countF > float64(maxCount) {
		return maxCount
	}
	return int(countF)
}

// gridRepeatExtent returns how many tracks count repetitions of pattern emit
// under the gridMaxTracks cap (gridAppendRepeat starting from an empty list)
// and the space those tracks contribute to auto-repeat counting.
func gridRepeatExtent(pattern []GridTrack, count int) (n int, size float64) {
	if len(pattern) == 0 || count <= 0 {
		return 0, 0
	}
	if count > gridMaxTracks || count*len(pattern) > gridMaxTracks {
		n = gridMaxTracks
	} else {
		n = count * len(pattern)
	}
	full, partial := n/len(pattern), n%len(pattern)
	patternSize := 0.0
	for _, track := range pattern {
		patternSize += autoRepeatTrackSize(track)
	}
	size = patternSize * float64(full)
	for _, track := range pattern[:partial] {
		size += autoRepeatTrackSize(track)
	}
	return n, size
}

// gridAppendRepeat appends count repetitions of pattern to tracks, stopping
// as soon as len(tracks) reaches gridMaxTracks. The truncation is at track
// granularity, so the result holds exactly the first gridMaxTracks tracks of
// the full expansion; a pattern that would cross the cap is cut mid-pattern.
func gridAppendRepeat(tracks []GridTrack, pattern []GridTrack, count int) []GridTrack {
	for i := 0; i < count; i++ {
		room := gridMaxTracks - len(tracks)
		if room <= 0 {
			return tracks
		}
		if len(pattern) >= room {
			return append(tracks, pattern[:room]...)
		}
		tracks = append(tracks, pattern...)
	}
	return tracks
}

// expandAutoRepeatTracks expands the repeat() patterns of one axis into
// concrete track definitions: the explicit tracks first, then every pattern
// in order, fixed counts literally and auto-fill / auto-fit as many times as
// fit the available space.
//
// CSS Grid Layout Module Level 1 §7.2.3.2: the auto-repeat count is the
// largest positive integer such that all tracks plus gaps fit the grid
// container, so it is computed against the space left after ALL other tracks
// in the axis, the explicit tracks and the fixed-count patterns (wherever
// they appear in the list), including the gaps between those tracks and the
// gap between them and the repeated block. With N other tracks of total size
// S, the repeated block gets availableSize - S - N × gap.
//
// The total number of tracks emitted never exceeds gridMaxTracks: once the
// cap is reached the remaining patterns (or the rest of the current one) are
// dropped, so the result is always bounded and deterministic regardless of
// the requested counts (see gridAppendRepeat).
//
// At most one auto-repeat pattern is expected (gridExpandTemplate filters
// the others, §7.2.3.1); if several are passed, each one is counted against
// the space the previous ones left over.
//
// Parameters:
//   - repeats: Array of RepeatTrack patterns (mix of auto-fill/auto-fit and regular)
//   - explicitTracks: Explicitly defined tracks (non-repeating)
//   - availableSize: Available space in the grid axis (>= Unbounded when indefinite)
//   - gap: Gap between tracks
//
// Returns: the expanded tracks and the index range [autoStart, autoEnd) of
// the tracks produced by the first auto-repeat pattern (both 0 when there is
// none).
//
// See: https://www.w3.org/TR/css-grid-1/#auto-repeat
func expandAutoRepeatTracks(repeats []RepeatTrack, explicitTracks []GridTrack, availableSize, gap float64) (tracks []GridTrack, autoStart, autoEnd int) {
	if math.IsNaN(gap) || math.IsInf(gap, 0) || gap < 0 {
		gap = 0
	}

	// Space and track count of everything that is not auto-repeated.
	otherTracks := len(explicitTracks)
	usedSpace := 0.0
	for _, track := range explicitTracks {
		usedSpace += autoRepeatTrackSize(track)
	}
	capacity := len(explicitTracks)
	for _, repeat := range repeats {
		if isAutoRepeatTrack(repeat) {
			capacity += len(repeat.Tracks)
			continue
		}
		n, size := gridRepeatExtent(repeat.Tracks, repeat.Count)
		otherTracks += n
		usedSpace += size
		capacity += n
	}
	if otherTracks > 0 {
		// Gaps between the other tracks (N-1) plus the gap separating them
		// from the repeated block.
		usedSpace += gap * float64(otherTracks)
	}
	if capacity > gridMaxTracks {
		capacity = gridMaxTracks
	}

	result := make([]GridTrack, 0, capacity)
	result = gridAppendRepeat(result, explicitTracks, 1)

	// Remaining space for auto-repeat tracks. An indefinite available size
	// stays indefinite so each auto-repeat pattern repeats exactly once.
	remainingSpace := availableSize
	if availableSize < Unbounded {
		remainingSpace = availableSize - usedSpace
		if remainingSpace < 0 {
			remainingSpace = 0
		}
	}

	seenAuto := false
	for _, repeat := range repeats {
		switch {
		case isAutoRepeatTrack(repeat):
			count := calculateAutoRepeatCount(repeat, remainingSpace, gap)
			start := len(result)
			result = gridAppendRepeat(result, repeat.Tracks, count)
			if !seenAuto {
				seenAuto = true
				autoStart, autoEnd = start, len(result)
			}

			// Update remaining space for any further (invalid but tolerated)
			// auto-repeat pattern.
			if remainingSpace < Unbounded {
				_, size := gridRepeatExtent(repeat.Tracks, count)
				remainingSpace -= size + gap*float64(len(result)-start)
				if remainingSpace < 0 {
					remainingSpace = 0
				}
			}
		case repeat.Count > 0:
			result = gridAppendRepeat(result, repeat.Tracks, repeat.Count)
		}
	}

	return result, autoStart, autoEnd
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
// algorithm later uses. The expansion emits at most gridMaxTracks tracks in
// total (explicit tracks included), whatever the fixed counts or the
// auto-repeat count, so the result is always bounded.
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
			// Fixed count; expandAutoRepeatTracks caps the emitted tracks.
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

	tracks, autoStart, autoEnd := expandAutoRepeatTracks(valid, template, availableSize, gap)
	if autoIndex < 0 || valid[autoIndex].Count != RepeatCountAutoFit || autoEnd <= autoStart {
		return tracks, nil
	}

	// The auto-fit repetitions occupy [autoStart, autoEnd) of the expansion,
	// already truncated to the track cap by expandAutoRepeatTracks.
	autoFit := make([]bool, len(tracks))
	for i := autoStart; i < autoEnd && i < len(tracks); i++ {
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

// gridLengthIsFixed reports whether a track sizing function is a
// <fixed-breadth>: a length or percentage with a unit, finite, non-negative,
// and not one of the intrinsic-size sentinels or the unbounded value.
//
// See: https://www.w3.org/TR/css-grid-1/#typedef-fixed-breadth
func gridLengthIsFixed(l Length) bool {
	if l.Unit == "" || l.Unit == UnboundedUnit {
		return false
	}
	v := l.Value
	return !math.IsNaN(v) && v >= 0 && v < Unbounded
}

// validateAutoRepeatTracks validates that an auto-fill / auto-fit pattern
// only uses fixed-size tracks.
//
// CSS Grid Layout Module Level 1 §7.2.3: auto-repeat takes a
// <fixed-size> list, which is a <fixed-breadth> (length or percentage),
// minmax(<fixed-breadth>, <track-breadth>), or minmax(<inflexible-breadth>,
// <fixed-breadth>). Hence a track is valid when at least one of its sizing
// functions is a fixed breadth. In this engine's GridTrack representation
// that excludes:
//
//   - flexible tracks (Fraction > 0) and fit-content (Fraction == -1);
//   - min-content / max-content (the SizeMinContent / SizeMaxContent
//     sentinels in MaxSize with a 0 minimum);
//   - auto, i.e. AutoTrack() (min 0, max unbounded) and the zero-value
//     GridTrack that gridNormalizeTrack maps to it. minmax(0px, auto) is
//     grammatically a fixed size but is indistinguishable from auto here, so
//     a fixed minimum must be positive to count; a fixed maximum always
//     counts (minmax(auto, 100px), 0px).
//
// Returns: true if valid, false otherwise
//
// See: https://www.w3.org/TR/css-grid-1/#auto-repeat
// See: https://www.w3.org/TR/css-grid-1/#typedef-fixed-size
func validateAutoRepeatTracks(repeat RepeatTrack) bool {
	for _, track := range repeat.Tracks {
		// fr units and fit-content (Fraction == -1) are not fixed sizes.
		if track.Fraction != 0 {
			return false
		}
		if gridLengthIsFixed(track.MaxSize) {
			continue
		}
		if gridLengthIsFixed(track.MinSize) && track.MinSize.Value > 0 {
			continue
		}
		return false
	}

	return true
}
