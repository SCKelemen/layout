// Command grid demonstrates CSS Grid layouts with the layout package.
//
// It has four sections, selectable by name; with no argument every section
// runs in order:
//
//	go run ./examples/grid            # all sections
//	go run ./examples/grid dashboard  # header / sidebar / main / footer template
//	go run ./examples/grid columns    # 3-column, 2-row fixed grid
//	go run ./examples/grid bento      # mosaic of row- and column-spanning items
//	go run ./examples/grid autofill   # repeat(auto-fill, 100px) columns
package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/SCKelemen/layout"
)

var sections = []struct {
	name string
	run  func()
}{
	{"dashboard", dashboard},
	{"columns", columns},
	{"bento", bento},
	{"autofill", autofill},
}

func main() {
	if len(os.Args) > 1 {
		for _, s := range sections {
			if s.name == os.Args[1] {
				s.run()
				return
			}
		}
		fmt.Fprintf(os.Stderr, "unknown section %q; choose one of: dashboard, columns, bento, autofill\n", os.Args[1])
		os.Exit(2)
	}
	for i, s := range sections {
		if i > 0 {
			fmt.Println()
		}
		s.run()
	}
}

func banner(title string) {
	fmt.Println(title)
	fmt.Println(repeatRune('=', len(title)))
}

func repeatRune(r rune, n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = r
	}
	return string(b)
}

// dashboard is the classic page template: a fixed header and footer spanning
// both columns, with a fixed sidebar and a fractional main area between them.
func dashboard() {
	banner("Dashboard template (explicit placement)")

	root := &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayGrid,
			GridTemplateRows: []layout.GridTrack{
				layout.FixedTrack(layout.Px(80)), // Header
				layout.FractionTrack(1),          // Main content area
				layout.FixedTrack(layout.Px(40)), // Footer
			},
			GridTemplateColumns: []layout.GridTrack{
				layout.FixedTrack(layout.Px(200)), // Sidebar
				layout.FractionTrack(1),           // Main content
			},
			GridGap: layout.Px(10),
			Padding: layout.Uniform(layout.Px(10)),
		},
		Children: []*layout.Node{
			// Header spanning full width
			{Style: layout.Style{GridRowStart: 0, GridRowEnd: 1, GridColumnStart: 0, GridColumnEnd: 2}},
			// Sidebar
			{Style: layout.Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 0, GridColumnEnd: 1}},
			// Main content
			{Style: layout.Style{GridRowStart: 1, GridRowEnd: 2, GridColumnStart: 1, GridColumnEnd: 2}},
			// Footer spanning full width
			{Style: layout.Style{GridRowStart: 2, GridRowEnd: 3, GridColumnStart: 0, GridColumnEnd: 2}},
		},
	}

	constraints := layout.Loose(800, 600)
	ctx := layout.NewLayoutContext(800, 600, 16)
	size := layout.Layout(root, constraints, ctx)

	fmt.Printf("Grid container size: %.2f x %.2f\n", size.Width, size.Height)

	areas := []string{"Header", "Sidebar", "Main Content", "Footer"}
	for i, child := range root.Children {
		fmt.Printf("%s: (%.2f, %.2f) %.2f x %.2f\n",
			areas[i], child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
	}
}

// columns is a plain 3-column, 2-row grid of fixed tracks, one item per cell.
func columns() {
	banner("Multi-column grid (3 columns x 2 rows)")

	root := &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayGrid,
			GridTemplateRows: []layout.GridTrack{
				layout.FixedTrack(layout.Px(100)),
				layout.FixedTrack(layout.Px(100)),
			},
			GridTemplateColumns: layout.RepeatTracks(3, layout.FixedTrack(layout.Px(150))),
			GridGap:             layout.Px(10),
			Padding:             layout.Uniform(layout.Px(10)),
		},
	}
	for row := 0; row < 2; row++ {
		for col := 0; col < 3; col++ {
			root.Children = append(root.Children, &layout.Node{
				Style: layout.Style{GridRowStart: row, GridColumnStart: col},
			})
		}
	}

	constraints := layout.Loose(500, 250)
	ctx := layout.NewLayoutContext(800, 600, 16)
	size := layout.Layout(root, constraints, ctx)

	fmt.Printf("Container size: %.2f x %.2f\n\n", size.Width, size.Height)
	for i, child := range root.Children {
		fmt.Printf("Row %d, Col %d: (%.2f, %.2f) %.2f x %.2f\n",
			i/3, i%3, child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
	}
}

// bento builds a mosaic on a 4x4 grid where items span different numbers of
// rows and columns. Items are auto-sized to their grid area (stretch), so no
// explicit Width/Height is needed.
func bento() {
	banner("Bento box grid (spanning items)")

	// 4 rows x 4 columns, rows=150px, cols=200px
	root := layout.Grid(4, 4, 150, 200)
	root.Style.GridGap = layout.Px(10)
	root.Style.Padding = layout.Uniform(layout.Px(20))

	type item struct {
		desc                   string
		rowStart, rowEnd       int
		columnStart, columnEnd int
	}
	items := []item{
		{"Large Featured (2x2)", 0, 2, 0, 2},
		{"Medium Horizontal (1x2)", 0, 1, 2, 4},
		{"Small (1x1)", 1, 2, 2, 3},
		{"Small (1x1)", 1, 2, 3, 4},
		{"Medium Vertical (2x1)", 2, 4, 0, 1},
		{"Medium Horizontal (1x2)", 2, 3, 1, 3},
		{"Small (1x1)", 2, 3, 3, 4},
		{"Wide Banner (1x3)", 3, 4, 1, 4},
	}
	for _, it := range items {
		root.Children = append(root.Children, &layout.Node{
			Style: layout.Style{
				GridRowStart:    it.rowStart,
				GridRowEnd:      it.rowEnd,
				GridColumnStart: it.columnStart,
				GridColumnEnd:   it.columnEnd,
			},
		})
	}

	constraints := layout.Loose(900, 700)
	ctx := layout.NewLayoutContext(800, 600, 16)
	size := layout.Layout(root, constraints, ctx)

	fmt.Printf("Container size: %.2f x %.2f\n\n", size.Width, size.Height)
	for i, child := range root.Children {
		it := items[i]
		fmt.Printf("%s: rows %d-%d, cols %d-%d -> (%.2f, %.2f) %.2f x %.2f\n",
			it.desc, it.rowStart, it.rowEnd-1, it.columnStart, it.columnEnd-1,
			child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
	}
}

// autofill uses repeat(auto-fill, 100px) for the columns: the grid creates as
// many 100px columns as fit the available width and auto-places the items
// into them. The repeat pattern lives in Style.GridTemplateColumnsRepeat and
// is expanded against the container size during layout.
//
// CSS equivalent:
//
//	display: grid;
//	grid-template-columns: repeat(auto-fill, 100px);
//	grid-auto-rows: 60px;
//	gap: 10px;
func autofill() {
	banner("Auto-fill columns: repeat(auto-fill, 100px)")

	root := &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayGrid,
			GridTemplateColumnsRepeat: []layout.RepeatTrack{
				layout.AutoFillTracks(layout.FixedTrack(layout.Px(100))),
			},
			GridAutoRows: layout.FixedTrack(layout.Px(60)),
			GridGap:      layout.Px(10),
			Width:        layout.Px(450),
		},
	}
	for i := 0; i < 7; i++ {
		root.Children = append(root.Children, &layout.Node{
			Style: layout.Style{GridRowStart: -1, GridRowEnd: -1, GridColumnStart: -1, GridColumnEnd: -1},
		})
	}

	ctx := layout.NewLayoutContext(800, 600, 16)
	size := layout.Layout(root, layout.Loose(450, layout.Unbounded), ctx)

	fmt.Printf("Container size: %.2f x %.2f\n", size.Width, size.Height)
	// Count the distinct column origins the layout produced. With 450px and
	// a 10px gap, four 100px columns fit (4*100 + 3*10 = 430).
	seen := map[float64]bool{}
	for _, child := range root.Children {
		seen[child.Rect.X] = true
	}
	xs := make([]float64, 0, len(seen))
	for x := range seen {
		xs = append(xs, x)
	}
	sort.Float64s(xs)
	fmt.Printf("Generated %d column(s) at x = %v\n", len(xs), xs)
	fmt.Println("(repeat(auto-fill, 100px) in a 450px container with a 10px gap yields 4 columns: 4*100 + 3*10 = 430)")
	fmt.Println()
	for i, child := range root.Children {
		fmt.Printf("Item %d: (%.2f, %.2f) %.2f x %.2f\n",
			i, child.Rect.X, child.Rect.Y, child.Rect.Width, child.Rect.Height)
	}
}
