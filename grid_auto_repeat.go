package layout

// Grid auto-repeat algorithms for CSS Grid Layout auto-fill and auto-fit.
//
// Implements dynamic grid track generation based on container size.
//
// Algorithm based on CSS Grid Layout Module Level 1:
// - §7.2.3: Repeat-to-fill (auto-fill and auto-fit)
//
// See: https://www.w3.org/TR/css-grid-1/#auto-repeat

// calculateAutoRepeatCount calculates how many times to repeat a track pattern
// based on available space.
//
// Formula from CSS spec:
// floor((availableSize + gap) / (repetitionSize + gap))
//
// Parameters:
//   - repeat: The RepeatTrack pattern to repeat
//   - availableSize: The available space in the grid axis
//   - gap: The gap between tracks
//
// Returns: The number of repetitions (minimum 1)
func calculateAutoRepeatCount(repeat RepeatTrack, availableSize, gap float64) int {
	if len(repeat.Tracks) == 0 {
		return 0
	}

	// Calculate the size of one repetition
	repetitionSize := 0.0
	for _, track := range repeat.Tracks {
		// For auto-repeat, only fixed-size tracks are allowed
		// Use MinSize as the track size (MaxSize should equal MinSize for fixed tracks)
		trackSize := track.MinSize
		if track.MaxSize < trackSize && track.MaxSize > 0 {
			trackSize = track.MaxSize
		}
		repetitionSize += trackSize
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
	count := int((availableSize + gap) / (repetitionSize + gap))

	// Ensure at least 1 repetition
	if count < 1 {
		count = 1
	}

	return count
}

// expandAutoRepeatTracks expands auto-fill or auto-fit track patterns
// into concrete track definitions based on available space.
//
// Parameters:
//   - repeats: Array of RepeatTrack patterns (mix of auto-fill/auto-fit and regular)
//   - explicitTracks: Explicitly defined tracks (non-repeating)
//   - availableSize: Available space in the grid axis
//   - gap: Gap between tracks
//
// Returns: Expanded array of GridTrack definitions
func expandAutoRepeatTracks(repeats []RepeatTrack, explicitTracks []GridTrack, availableSize, gap float64) []GridTrack {
	// Start with explicit tracks
	result := make([]GridTrack, 0, len(explicitTracks)*2)
	result = append(result, explicitTracks...)

	// Calculate space used by explicit tracks
	usedSpace := 0.0
	for _, track := range explicitTracks {
		trackSize := track.MinSize
		if track.MaxSize < trackSize && track.MaxSize > 0 {
			trackSize = track.MaxSize
		}
		usedSpace += trackSize
	}

	// Add gaps for explicit tracks
	if len(explicitTracks) > 1 {
		usedSpace += gap * float64(len(explicitTracks)-1)
	}

	// Remaining space for auto-repeat tracks
	remainingSpace := availableSize - usedSpace
	if remainingSpace < 0 {
		remainingSpace = 0
	}

	// Expand each auto-repeat pattern
	for _, repeat := range repeats {
		if repeat.Count == RepeatCountAutoFill || repeat.Count == RepeatCountAutoFit {
			// Calculate how many repetitions fit
			count := calculateAutoRepeatCount(repeat, remainingSpace, gap)

			// Expand the pattern
			for i := 0; i < count; i++ {
				result = append(result, repeat.Tracks...)
			}

			// Update remaining space
			for _, track := range repeat.Tracks {
				trackSize := track.MinSize
				if track.MaxSize < trackSize && track.MaxSize > 0 {
					trackSize = track.MaxSize
				}
				remainingSpace -= trackSize
			}
			if count > 0 && len(repeat.Tracks) > 0 {
				remainingSpace -= gap * float64(count*len(repeat.Tracks)-1)
			}
		} else if repeat.Count > 0 {
			// Regular repeat (not auto-fill/auto-fit)
			for i := 0; i < repeat.Count; i++ {
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
				MinSize:  0,
				MaxSize:  0,
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
		if track.MaxSize == SizeMinContent || track.MaxSize == SizeMaxContent {
			return false
		}

		// Fit-content is also not allowed in auto-repeat
		if track.Fraction == -1 {
			return false
		}
	}

	return true
}
