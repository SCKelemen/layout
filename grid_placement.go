package layout

// gridMaxTracks is the hard upper bound on the number of lines in either
// axis of the implicit grid.
//
// CSS Grid Layout Module Level 1 leaves the implicit grid unbounded, but
// every shipping engine clamps it (Chromium, WebKit, and Gecko all use
// 10,000 lines) so that a hostile or buggy grid-row / grid-column value
// cannot force an unbounded allocation. The same limit is applied here.
//
// See: https://www.w3.org/TR/css-grid-1/#overlarge-grids
const gridMaxTracks = 10000

// gridSpan is a grid item's normalized placement in one axis.
//
// start is the 0-based start line, span is the number of tracks covered
// (always >= 1), and auto reports whether the position in this axis still
// needs to be determined by the auto-placement algorithm.
type gridSpan struct {
	start int
	span  int
	auto  bool
}

// end returns the exclusive end line.
func (s gridSpan) end() int { return s.start + s.span }

// gridNormalizeSpan turns the raw GridRowStart/GridRowEnd (or column)
// style values into a normalized gridSpan.
//
// Sentinels: -1 (or any negative start) means auto; a default-constructed
// item has start 0 and end 0, which is also auto because an explicit row 0
// is always written with a positive end (§8.3: "auto ... indicating
// auto-placement, an automatic span, or a default span of one").
//
// Conflict handling follows CSS Grid Layout Module Level 1 §8.3.1:
//   - an auto start with a definite end resolves to the line before the end;
//   - if end == start, the end line is removed and the span becomes 1;
//   - if end < start, the two lines are swapped.
//
// Every resulting line is clamped to gridMaxTracks.
//
// See: https://www.w3.org/TR/css-grid-1/#grid-placement-errors
// See: https://www.w3.org/TR/css-grid-1/#line-placement
func gridNormalizeSpan(start, end int) gridSpan {
	autoStart := start < 0 || (start == 0 && end <= 0)
	if autoStart {
		if end > 0 {
			// grid-row: auto / N places the item so that it ends at line N
			// with the default span of one.
			s := end - 1
			if s > gridMaxTracks-1 {
				s = gridMaxTracks - 1
			}
			return gridSpan{start: s, span: 1}
		}
		return gridSpan{start: 0, span: 1, auto: true}
	}

	if start > gridMaxTracks-1 {
		start = gridMaxTracks - 1
	}
	if end <= 0 {
		// Unset end: default span of one.
		return gridSpan{start: start, span: 1}
	}
	if end < start {
		start, end = end, start
	}
	if end == start {
		end = start + 1
	}
	if end > gridMaxTracks {
		end = gridMaxTracks
	}
	if end <= start {
		// Only reachable when start was clamped to the last allowed line.
		start = gridMaxTracks - 1
		end = gridMaxTracks
	}
	return gridSpan{start: start, span: end - start}
}

// gridGrowTracks appends implicit tracks until len(*tracks) >= count.
//
// Implicit tracks are sized by the grid-auto-rows / grid-auto-columns track
// (§7.6). A zero-value GridTrack is the property's initial value, auto.
// count is clamped to gridMaxTracks so the allocation is always bounded.
//
// See: https://www.w3.org/TR/css-grid-1/#auto-tracks
func gridGrowTracks(tracks *[]GridTrack, count int, implicit GridTrack) {
	if count > gridMaxTracks {
		count = gridMaxTracks
	}
	if len(*tracks) >= count {
		return
	}
	implicit = gridNormalizeTrack(implicit)
	grown := make([]GridTrack, len(*tracks), count)
	copy(grown, *tracks)
	for len(grown) < count {
		grown = append(grown, implicit)
	}
	*tracks = grown
}

// gridOccupancy records which cells of the grid are covered by an item.
// Cells are keyed by (major, minor) line index.
type gridOccupancy map[[2]int]bool

func (o gridOccupancy) fits(major, minor gridSpan) bool {
	for a := major.start; a < major.end(); a++ {
		for b := minor.start; b < minor.end(); b++ {
			if o[[2]int{a, b}] {
				return false
			}
		}
	}
	return true
}

func (o gridOccupancy) mark(major, minor gridSpan) {
	for a := major.start; a < major.end(); a++ {
		for b := minor.start; b < minor.end(); b++ {
			o[[2]int{a, b}] = true
		}
	}
}

// gridPlaceItems performs grid item placement including auto-placement.
//
// Algorithm based on CSS Grid Layout Module Level 1 §8.5, Grid Item
// Placement Algorithm:
//
//  0. Generate anonymous grid items (not applicable) and determine the
//     columns in the implicit grid.
//  1. Position anything that is not auto-positioned.
//  2. Process the items locked to a given row (the major axis).
//  3. Position the remaining grid items with the auto-placement cursor,
//     skipping every cell already occupied (including cells covered by
//     spanning items), either sparsely (the cursor only moves forward) or
//     densely (the cursor restarts at the beginning for every item).
//
// grid-auto-flow: column runs the same algorithm with rows and columns
// swapped (§8.3), so the code below works in major/minor coordinates where
// the major axis is the one items flow along.
//
// The implicit grid grows by grid-auto-rows / grid-auto-columns as needed
// (§7.6) and is hard-capped at gridMaxTracks in each axis; every search
// loop below is bounded by the current size of the grid, so malformed input
// can neither loop forever nor allocate without bound.
//
// See: https://www.w3.org/TR/css-grid-1/#auto-placement-algo
// See: https://www.w3.org/TR/css-grid-1/#grid-auto-flow-property
func gridPlaceItems(node *Node, rows *[]GridTrack, columns *[]GridTrack, autoFlow GridAutoFlow) []*gridItem {
	isColumnFlow := autoFlow == GridAutoFlowColumn || autoFlow == GridAutoFlowColumnDense
	isDense := autoFlow == GridAutoFlowRowDense || autoFlow == GridAutoFlowColumnDense

	// Major axis = the axis items flow along (rows for row flow).
	majorTracks, minorTracks := rows, columns
	majorImplicit, minorImplicit := node.Style.GridAutoRows, node.Style.GridAutoColumns
	if isColumnFlow {
		majorTracks, minorTracks = columns, rows
		majorImplicit, minorImplicit = node.Style.GridAutoColumns, node.Style.GridAutoRows
	}

	type placement struct {
		item  *gridItem
		major gridSpan
		minor gridSpan
	}

	// Named areas (§7.3) resolve to definite positions that take precedence
	// over the child's line-based placement, without touching child.Style.
	areas := gridResolveAreas(node)

	placements := make([]placement, 0, len(node.Children))
	for _, child := range node.Children {
		// display:none children generate no box. Absolutely positioned
		// children are not grid items (§9): they take no grid cell and do
		// not advance the auto-placement cursor; LayoutGrid lays them out
		// separately.
		// See: https://www.w3.org/TR/css-grid-1/#abspos-items
		if child.Style.Display == DisplayNone || isOutOfFlow(child) {
			continue
		}
		rowStart, rowEnd := child.Style.GridRowStart, child.Style.GridRowEnd
		colStart, colEnd := child.Style.GridColumnStart, child.Style.GridColumnEnd
		if area, ok := areas[child]; ok {
			rowStart, rowEnd = area.RowStart, area.RowEnd
			colStart, colEnd = area.ColumnStart, area.ColumnEnd
		}
		rowSpan := gridNormalizeSpan(rowStart, rowEnd)
		colSpan := gridNormalizeSpan(colStart, colEnd)
		item := &gridItem{node: child, autoRow: rowSpan.auto, autoCol: colSpan.auto}
		p := placement{item: item, major: rowSpan, minor: colSpan}
		if isColumnFlow {
			p.major, p.minor = colSpan, rowSpan
		}
		placements = append(placements, p)
	}

	occupied := make(gridOccupancy)

	// Step 0: the implicit grid must be wide enough for the largest minor
	// span among items without a definite minor position.
	minorCount := len(*minorTracks)
	for _, p := range placements {
		if p.minor.auto && p.minor.span > minorCount {
			minorCount = p.minor.span
		}
	}
	gridGrowTracks(minorTracks, minorCount, minorImplicit)

	// Step 1: items with a definite position in both axes.
	for i := range placements {
		p := &placements[i]
		if p.major.auto || p.minor.auto {
			continue
		}
		gridGrowTracks(majorTracks, p.major.end(), majorImplicit)
		gridGrowTracks(minorTracks, p.minor.end(), minorImplicit)
		occupied.mark(p.major, p.minor)
	}

	// Step 2: items locked to a major line (definite major, auto minor).
	// In sparse mode the minor start must also be past every item placed in
	// the same major line by this step; dense mode restarts at line 0.
	lineCursor := make(map[int]int)
	for i := range placements {
		p := &placements[i]
		if p.major.auto || !p.minor.auto {
			continue
		}
		startMinor := 0
		if !isDense {
			startMinor = lineCursor[p.major.start]
		}
		// Cells at or beyond len(*minorTracks) are never occupied, so the
		// search always succeeds by minor == len(*minorTracks) at the latest.
		limit := len(*minorTracks)
		minor := startMinor
		for ; minor < limit; minor++ {
			if occupied.fits(p.major, gridSpan{start: minor, span: p.minor.span}) {
				break
			}
		}
		if minor > gridMaxTracks-p.minor.span {
			minor = gridMaxTracks - p.minor.span
		}
		p.minor = gridSpan{start: minor, span: p.minor.span}
		gridGrowTracks(majorTracks, p.major.end(), majorImplicit)
		gridGrowTracks(minorTracks, p.minor.end(), minorImplicit)
		occupied.mark(p.major, p.minor)
		if !isDense {
			lineCursor[p.major.start] = p.minor.end()
		}
	}

	// Step 3: remaining items, driven by the auto-placement cursor.
	cursorMajor, cursorMinor := 0, 0
	for i := range placements {
		p := &placements[i]
		if !p.major.auto {
			continue
		}

		if !p.minor.auto {
			// Definite minor position: set the cursor's minor position to
			// the item's start line. Sparse: if that moves the cursor
			// backwards, advance to the next major line. Dense: restart
			// from the first major line.
			if isDense {
				cursorMajor = 0
			} else if p.minor.start < cursorMinor {
				cursorMajor++
			}
			cursorMinor = p.minor.start
			gridGrowTracks(minorTracks, p.minor.end(), minorImplicit)

			// Advance the major position until the area is free. Every
			// cell at or beyond len(*majorTracks) is free, so the loop
			// terminates within len(*majorTracks)+1 steps.
			limit := len(*majorTracks) + 1
			for step := 0; step < limit; step++ {
				if occupied.fits(gridSpan{start: cursorMajor, span: p.major.span}, p.minor) {
					break
				}
				cursorMajor++
			}
		} else {
			// Both axes auto.
			if isDense {
				cursorMajor, cursorMinor = 0, 0
			}
			minorLimit := len(*minorTracks)
			// Upper bound on the walk: one pass over every existing major
			// line plus one fresh (empty) line, times the minor count.
			maxSteps := (len(*majorTracks) + 2) * (minorLimit + 1)
			for step := 0; step < maxSteps; step++ {
				if cursorMinor+p.minor.span > minorLimit {
					// Overflowed the implicit grid's minor axis: next
					// major line, back to the first minor line.
					cursorMajor++
					cursorMinor = 0
					continue
				}
				if occupied.fits(gridSpan{start: cursorMajor, span: p.major.span}, gridSpan{start: cursorMinor, span: p.minor.span}) {
					break
				}
				cursorMinor++
			}
			p.minor = gridSpan{start: cursorMinor, span: p.minor.span}
		}

		if cursorMajor > gridMaxTracks-p.major.span {
			cursorMajor = gridMaxTracks - p.major.span
		}
		p.major = gridSpan{start: cursorMajor, span: p.major.span}
		gridGrowTracks(majorTracks, p.major.end(), majorImplicit)
		gridGrowTracks(minorTracks, p.minor.end(), minorImplicit)
		occupied.mark(p.major, p.minor)
	}

	// Convert back to row/column coordinates. Every item's end line is
	// covered by the grown track lists, so callers can index track sizes
	// with item positions without further bounds checks.
	items := make([]*gridItem, 0, len(placements))
	for _, p := range placements {
		rowSpan, colSpan := p.major, p.minor
		if isColumnFlow {
			rowSpan, colSpan = p.minor, p.major
		}
		p.item.rowStart = rowSpan.start
		p.item.rowEnd = rowSpan.end()
		p.item.colStart = colSpan.start
		p.item.colEnd = colSpan.end()
		items = append(items, p.item)
	}
	return items
}

// gridResolveAreas resolves the named grid areas referenced by the container's
// children to line positions.
//
// For each child whose Style.GridArea names an area of the container's
// GridTemplateAreas, the returned map holds that area's row/column lines.
// The child's Style is left untouched: the resolved position lives only in
// the placement pass, so the same tree can be laid out again with a different
// template (or the same child reused elsewhere) and be placed afresh.
//
// A GridArea name that the template does not define is not an error: the
// child is absent from the map and is auto-placed by its GridRow*/GridColumn*
// values as if GridArea were empty. (In CSS a reference to an undefined named
// area resolves to auto placement as well, §8.3 "if there is no line with
// that name ... all implicit grid lines are assumed to have that name".) The
// returned map is nil when no child references a defined area.
//
// Algorithm based on CSS Grid Layout Module Level 1:
// - §7.3: Grid Template Areas
// - §8.3: Line-based Placement (named areas)
//
// See: https://www.w3.org/TR/css-grid-1/#grid-template-areas-property
// See: https://www.w3.org/TR/css-grid-1/#line-placement
func gridResolveAreas(node *Node) map[*Node]GridArea {
	// If no template areas defined, nothing to resolve
	if node.Style.GridTemplateAreas == nil {
		return nil
	}

	// Build lookup map of area names to definitions
	areaMap := make(map[string]GridArea, len(node.Style.GridTemplateAreas.Areas))
	for _, area := range node.Style.GridTemplateAreas.Areas {
		areaMap[area.Name] = area
	}

	var resolved map[*Node]GridArea
	for _, child := range node.Children {
		if child.Style.GridArea == "" {
			continue
		}
		area, found := areaMap[child.Style.GridArea]
		if !found {
			// Undefined area name: the child keeps its line-based placement
			// (usually auto-placement).
			continue
		}
		if resolved == nil {
			resolved = make(map[*Node]GridArea)
		}
		resolved[child] = area
	}
	return resolved
}
