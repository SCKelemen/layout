package layout

import (
	"math"
)

// clampFlexMainSize clamps a main size to the item's resolved min/max main
// size constraints. The minimum always wins over the maximum, and sizes are
// never negative.
//
// Implementation of CSS Flexible Box Layout Module Level 1 §9.7 step 4d
// (clamping to min/max main size).
// https://www.w3.org/TR/css-flexbox-1/#resolve-flexible-lengths
func clampFlexMainSize(item *flexItem, size float64) float64 {
	if item.maxMain > 0 && item.maxMain < Unbounded && size > item.maxMain {
		size = item.maxMain
	}
	if size < item.minMain {
		size = item.minMain
	}
	if size < 0 {
		size = 0
	}
	return size
}

// flexLineFreeSpace computes the free space in a line: the container's main
// size minus the outer sizes of the items, where frozen items contribute their
// target main size and unfrozen items contribute their flex base size.
//
// Implementation of CSS Flexible Box Layout Module Level 1 §9.7 step 4a/4b
// ("calculate the remaining free space").
// https://www.w3.org/TR/css-flexbox-1/#resolve-flexible-lengths
func flexLineFreeSpace(line []*flexItem, mainSize float64) float64 {
	used := 0.0
	for _, item := range line {
		if item.frozen {
			used += item.mainSize
		} else {
			used += item.baseSize
		}
		used += item.mainMarginStart + item.mainMarginEnd
	}
	return mainSize - used
}

// flexboxDetermineMainSize resolves the main size of every item in a flex line.
//
// Implementation of CSS Flexible Box Layout Module Level 1:
// - §9.3 step 3: hypothetical main size (flex base size clamped by min/max)
// - §9.7: Resolving Flexible Lengths
//
// The §9.7 loop distributes free space proportionally to the flex grow
// factors (when growing) or the scaled flex shrink factors, i.e.
// flex-shrink * flex base size (when shrinking), clamps every item to its
// min/max main size, freezes the violating items and repeats. The loop is
// bounded by len(line)+1 iterations: each iteration freezes at least one item
// (or every remaining item), so malformed input can never loop forever.
//
// When the container's main size is indefinite (no definite main size or a
// size >= Unbounded) there is no free space to distribute and every item uses
// its hypothetical main size.
//
// See: https://www.w3.org/TR/css-flexbox-1/#resolve-flexible-lengths
func flexboxDetermineMainSize(line []*flexItem, mainSize float64, hasExplicitMainSize bool) {
	if len(line) == 0 {
		return
	}

	// §9.3 step 3: the hypothetical main size is the flex base size clamped
	// by the item's min and max main size properties.
	for _, item := range line {
		item.frozen = false
		item.hypotheticalMainSize = clampFlexMainSize(item, item.baseSize)
		item.mainSize = item.hypotheticalMainSize
	}

	// Indefinite main size: do not flex. Never let a MaxFloat64 container size
	// take part in the arithmetic below.
	if !hasExplicitMainSize || mainSize >= Unbounded || math.IsNaN(mainSize) || math.IsInf(mainSize, 0) {
		return
	}

	// §9.7 step 1: determine the used flex factor. Compare the sum of the outer
	// hypothetical main sizes with the container's inner main size.
	sumHypothetical := 0.0
	for _, item := range line {
		sumHypothetical += item.hypotheticalMainSize + item.mainMarginStart + item.mainMarginEnd
	}
	growing := sumHypothetical < mainSize

	// §9.7 step 2: size inflexible items. Freeze any item with a zero flex
	// factor, and any item whose flex base size is already on the wrong side of
	// its hypothetical main size for the used factor.
	for _, item := range line {
		factor := item.flexShrink
		if growing {
			factor = item.flexGrow
		}
		if factor <= 0 ||
			(growing && item.baseSize > item.hypotheticalMainSize) ||
			(!growing && item.baseSize < item.hypotheticalMainSize) {
			item.mainSize = item.hypotheticalMainSize
			item.frozen = true
		}
	}

	// §9.7 step 3: calculate the initial free space.
	initialFreeSpace := flexLineFreeSpace(line, mainSize)

	// §9.7 step 4: loop until every item is frozen. Hard cap: each iteration
	// freezes at least one item, so len(line)+1 iterations always suffice.
	for iteration := 0; iteration <= len(line); iteration++ {
		// 4a: check for flexible items.
		//
		// Free space is distributed in proportion to the factors, so scaling
		// every unfrozen factor by the same constant does not change the
		// result. When the largest factor is huge (e.g. 1e308) the sums below
		// would overflow to +Inf and every share would become NaN, so the
		// factors are normalized by the largest one first. The §9.7 step 4b
		// "sum of flex factors less than one" rule is unaffected: after
		// normalization the largest factor is exactly 1, so the sum is >= 1
		// whenever the scale is applied.
		unfrozen := 0
		maxFactor := 0.0
		for _, item := range line {
			if item.frozen {
				continue
			}
			unfrozen++
			factor := item.flexShrink
			if growing {
				factor = item.flexGrow
			}
			if factor > maxFactor {
				maxFactor = factor
			}
		}
		if unfrozen == 0 {
			return
		}
		factorScale := 1.0
		if maxFactor > 1e100 {
			factorScale = 1 / maxFactor
		}
		sumFactors := 0.0
		sumScaledShrink := 0.0
		for _, item := range line {
			if item.frozen {
				continue
			}
			if growing {
				sumFactors += item.flexGrow * factorScale
			} else {
				sumFactors += item.flexShrink * factorScale
				sumScaledShrink += item.flexShrink * factorScale * item.baseSize
			}
		}

		// 4b: calculate the remaining free space. If the sum of the unfrozen
		// flex factors is less than one, only that fraction of the initial free
		// space is distributed (if its magnitude is smaller).
		remaining := flexLineFreeSpace(line, mainSize)
		if sumFactors < 1 {
			scaled := initialFreeSpace * sumFactors
			if math.Abs(scaled) < math.Abs(remaining) {
				remaining = scaled
			}
		}

		// 4c: distribute the free space proportional to the flex factors.
		for _, item := range line {
			if item.frozen {
				continue
			}
			// The ratio is formed before multiplying by the free space so that
			// a large factor times a large free space cannot overflow.
			target := item.baseSize
			if remaining > 0 && growing && sumFactors > 0 {
				target += remaining * (item.flexGrow * factorScale / sumFactors)
			} else if remaining < 0 && !growing && sumScaledShrink > 0 {
				scaledShrink := item.flexShrink * factorScale * item.baseSize
				target -= math.Abs(remaining) * (scaledShrink / sumScaledShrink)
			}
			item.mainSize = target
		}

		// 4d: fix min/max violations, remembering each item's adjustment.
		totalViolation := 0.0
		violations := make(map[*flexItem]float64, unfrozen)
		for _, item := range line {
			if item.frozen {
				continue
			}
			clamped := clampFlexMainSize(item, item.mainSize)
			violation := clamped - item.mainSize
			violations[item] = violation
			totalViolation += violation
			item.mainSize = clamped
		}

		// 4e: freeze over-flexed items.
		for _, item := range line {
			if item.frozen {
				continue
			}
			violation := violations[item]
			switch {
			case totalViolation == 0:
				item.frozen = true
			case totalViolation > 0 && violation > 0:
				// Min violation: the item was clamped up.
				item.frozen = true
			case totalViolation < 0 && violation < 0:
				// Max violation: the item was clamped down.
				item.frozen = true
			}
		}
	}

	// Defensive: if the cap was reached, every item keeps its last clamped size.
	for _, item := range line {
		item.frozen = true
	}
}
