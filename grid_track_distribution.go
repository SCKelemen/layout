package layout

// gridDistributeTrackSpace distributes free space among grid tracks.
//
// Algorithm based on CSS Grid Layout Module Level 1:
// - §11.6: Grid Container Intrinsic Sizes
// - §11.8: Distributing free space
//
// For align-content (distributes space between row tracks along the block axis)
// For justify-content (distributes space between column tracks along the inline axis)
//
// See: https://www.w3.org/TR/css-grid-1/#grid-align
// See: https://www.w3.org/TR/css-grid-1/#grid-justify
func gridDistributeTrackSpace(
	trackSizes []float64,
	availableSpace float64,
	gap float64,
	alignment AlignContent,
) ([]float64, float64) {
	if len(trackSizes) == 0 {
		return trackSizes, 0
	}

	// Calculate total track size including gaps
	totalTrackSize := 0.0
	for _, size := range trackSizes {
		totalTrackSize += size
	}
	if len(trackSizes) > 1 {
		totalTrackSize += gap * float64(len(trackSizes)-1)
	}

	// Calculate free space
	freeSpace := availableSpace - totalTrackSize
	if freeSpace <= 0 {
		// No free space to distribute, return original sizes
		return trackSizes, totalTrackSize
	}

	// Distribute free space based on alignment
	switch alignment {
	case AlignContentFlexStart:
		// All tracks start from the beginning, no distribution needed
		return trackSizes, totalTrackSize

	case AlignContentFlexEnd:
		// All tracks move to the end, free space at start
		// Track positions will be offset by freeSpace
		return trackSizes, totalTrackSize

	case AlignContentCenter:
		// Tracks are centered, free space split evenly at start/end
		// Track positions will be offset by freeSpace/2
		return trackSizes, totalTrackSize

	case AlignContentSpaceBetween:
		// Free space distributed evenly between tracks
		if len(trackSizes) <= 1 {
			// Only one track, behave like flex-start
			return trackSizes, totalTrackSize
		}
		// Free space is distributed between tracks, not added to track sizes
		// This affects track positions, not sizes
		return trackSizes, totalTrackSize

	case AlignContentSpaceAround:
		// Free space distributed around tracks (half at each end)
		// Space before/after each track is equal
		// This affects track positions, not sizes
		return trackSizes, totalTrackSize

	case AlignContentStretch:
		// Distribute free space by increasing track sizes
		// Only for auto tracks (tracks without fixed size)
		// Count auto tracks (tracks that can grow)
		autoTrackCount := 0
		for range trackSizes {
			// Consider tracks that are not at max size as stretchable
			// For simplicity, we stretch all tracks equally
			autoTrackCount++
		}

		if autoTrackCount > 0 {
			spacePerTrack := freeSpace / float64(autoTrackCount)
			newSizes := make([]float64, len(trackSizes))
			newTotalSize := 0.0
			for i, size := range trackSizes {
				newSizes[i] = size + spacePerTrack
				newTotalSize += newSizes[i]
			}
			if len(trackSizes) > 1 {
				newTotalSize += gap * float64(len(trackSizes)-1)
			}
			return newSizes, newTotalSize
		}
		return trackSizes, totalTrackSize

	default:
		// Default to stretch
		return gridDistributeTrackSpace(trackSizes, availableSpace, gap, AlignContentStretch)
	}
}

// gridCalculateTrackOffsets calculates the starting position of each track based on alignment.
//
// This handles justify-content and align-content positioning of tracks within the grid container.
func gridCalculateTrackOffsets(
	trackSizes []float64,
	totalTrackSize float64,
	availableSpace float64,
	gap float64,
	alignment AlignContent,
) []float64 {
	if len(trackSizes) == 0 {
		return []float64{}
	}

	offsets := make([]float64, len(trackSizes))
	freeSpace := availableSpace - totalTrackSize
	if freeSpace < 0 {
		freeSpace = 0
	}

	currentOffset := 0.0

	switch alignment {
	case AlignContentFlexStart:
		// Tracks start from the beginning
		currentOffset = 0

	case AlignContentFlexEnd:
		// Tracks start from the end
		currentOffset = freeSpace

	case AlignContentCenter:
		// Tracks are centered
		currentOffset = freeSpace / 2

	case AlignContentSpaceBetween:
		if len(trackSizes) <= 1 {
			currentOffset = 0
		} else {
			// First track at start, distribute space between tracks
			spaceBetween := freeSpace / float64(len(trackSizes)-1)
			for i := range trackSizes {
				offsets[i] = currentOffset
				currentOffset += trackSizes[i] + gap + spaceBetween
			}
			return offsets
		}

	case AlignContentSpaceAround:
		// Space around each track
		spaceAround := freeSpace / float64(len(trackSizes))
		currentOffset = spaceAround / 2
		for i := range trackSizes {
			offsets[i] = currentOffset
			currentOffset += trackSizes[i] + gap + spaceAround
		}
		return offsets

	case AlignContentStretch:
		// Tracks are stretched (sizes already adjusted), start from beginning
		currentOffset = 0

	default:
		currentOffset = 0
	}

	// For flex-start, flex-end, center, and stretch: calculate offsets sequentially
	for i := range trackSizes {
		offsets[i] = currentOffset
		currentOffset += trackSizes[i]
		if i < len(trackSizes)-1 {
			currentOffset += gap
		}
	}

	return offsets
}
