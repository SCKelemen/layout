package layout

// gridTrackStretches reports whether a track takes part in the
// "Stretch auto Tracks" step: only tracks whose max track sizing function
// is auto grow when align-content / justify-content is normal or stretch.
// Fixed, minmax(fixed, fixed), intrinsic-keyword, and flexible tracks do not.
//
// CSS Grid Layout Module Level 1 §12.8: Stretch auto Tracks
// See: https://www.w3.org/TR/css-grid-1/#algo-stretch
func gridTrackStretches(track GridTrack, ctx *LayoutContext, currentFontSize float64) bool {
	if track.Fraction != 0 {
		return false
	}
	maxSize := ResolveLength(track.MaxSize, ctx, currentFontSize)
	return maxSize >= Unbounded
}

// gridDistributeTrackSpace distributes free space among grid tracks.
//
// Algorithm based on CSS Grid Layout Module Level 1:
// - §12.8: Stretch auto Tracks
// - §10.4: Aligning the Grid (align-content / justify-content)
//
// Only the stretch value changes track sizes: the free space is split
// equally among the tracks whose max sizing function is auto (§12.8). When
// there is no such track the free space is left alone and the tracks behave
// as start-aligned, as the spec requires. All other alignment values only
// affect track positions, which gridCalculateTrackOffsets computes.
//
// Returns the (possibly stretched) track sizes and their total including gaps.
//
// See: https://www.w3.org/TR/css-grid-1/#algo-stretch
// See: https://www.w3.org/TR/css-grid-1/#grid-align
func gridDistributeTrackSpace(
	trackSizes []float64,
	tracks []GridTrack,
	availableSpace float64,
	gap float64,
	alignment AlignContent,
	ctx *LayoutContext,
	currentFontSize float64,
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

	// Indefinite available space means there is no free space to distribute.
	if availableSpace >= Unbounded {
		return trackSizes, totalTrackSize
	}

	// Calculate free space
	freeSpace := availableSpace - totalTrackSize
	if freeSpace <= 0 {
		// No free space to distribute, return original sizes
		return trackSizes, totalTrackSize
	}

	switch alignment {
	case AlignContentFlexStart, AlignContentFlexEnd, AlignContentCenter,
		AlignContentSpaceBetween, AlignContentSpaceAround, AlignContentSpaceEvenly:
		// These only move tracks; sizes are unchanged.
		return trackSizes, totalTrackSize
	}

	// AlignContentStretch (the zero value) and any unknown value: stretch.
	stretchable := make([]int, 0, len(trackSizes))
	for i := range trackSizes {
		if i < len(tracks) && gridTrackStretches(tracks[i], ctx, currentFontSize) {
			stretchable = append(stretchable, i)
		}
	}
	if len(stretchable) == 0 {
		// §12.8: nothing to stretch; behaves as start alignment.
		return trackSizes, totalTrackSize
	}

	spacePerTrack := freeSpace / float64(len(stretchable))
	newSizes := make([]float64, len(trackSizes))
	copy(newSizes, trackSizes)
	for _, i := range stretchable {
		newSizes[i] += spacePerTrack
	}
	return newSizes, availableSpace
}

// gridJustifyToAlignContent maps a justify-content value onto the equivalent
// align-content value so the column axis can reuse gridDistributeTrackSpace
// and gridCalculateTrackOffsets. Both properties take the same content
// distribution keywords (css-align-3 §5.3); flex-start/flex-end are the
// flexbox spellings of start/end. Unknown values behave as start.
//
// Note that JustifyContent's zero value is flex-start in this library, so the
// §12.8 stretch step runs for columns only when JustifyContentStretch is set
// explicitly.
//
// See: https://www.w3.org/TR/css-align-3/#propdef-justify-content
// See: https://www.w3.org/TR/css-grid-1/#grid-align
func gridJustifyToAlignContent(justify JustifyContent) AlignContent {
	switch justify {
	case JustifyContentFlexEnd:
		return AlignContentFlexEnd
	case JustifyContentCenter:
		return AlignContentCenter
	case JustifyContentSpaceBetween:
		return AlignContentSpaceBetween
	case JustifyContentSpaceAround:
		return AlignContentSpaceAround
	case JustifyContentSpaceEvenly:
		return AlignContentSpaceEvenly
	case JustifyContentStretch:
		return AlignContentStretch
	default:
		return AlignContentFlexStart
	}
}

// gridCalculateTrackOffsets calculates the starting position of each track based on alignment.
//
// This handles justify-content and align-content positioning of tracks within the grid container
// (css-align-3 §5.3 content distribution). The column axis maps its justify-content value onto an
// AlignContent value with gridJustifyToAlignContent.
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
	freeSpace := 0.0
	if availableSpace < Unbounded {
		freeSpace = availableSpace - totalTrackSize
	}
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

	case AlignContentSpaceEvenly:
		// Equal space before, between, and after the tracks: n+1 equal
		// portions for n tracks.
		// https://www.w3.org/TR/css-align-3/#valdef-align-content-space-evenly
		spaceEvenly := freeSpace / float64(len(trackSizes)+1)
		currentOffset = spaceEvenly
		for i := range trackSizes {
			offsets[i] = currentOffset
			currentOffset += trackSizes[i] + gap + spaceEvenly
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
