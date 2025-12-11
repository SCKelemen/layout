package main

import "fmt"

func main() {
	// Manual calculation to verify
	// Row 0: 60px (fixed)
	// Row 1: max(200/2, 50) = 100px (spanning item gives 100px per row, stat cards give 50px)
	// Row 2: max(200/2, 150) = 150px (spanning item gives 100px, line graph gives 150px)
	// Row 3: 200/3 = 66.67px (spanning item)
	// Row 4: 200/3 = 66.67px
	// Row 5: 200/3 = 66.67px
	
	row0 := 60.0
	row1 := 100.0  // max(200/2, 50)
	row2 := 150.0  // max(200/2, 150)
	row3 := 200.0 / 3.0
	row4 := 200.0 / 3.0
	row5 := 200.0 / 3.0
	gap := 8.0
	
	total := row0 + gap + row1 + gap + row2 + gap + row3 + gap + row4 + gap + row5
	fmt.Printf("Expected: %.2f\n", total)
	fmt.Printf("Row breakdown: %.2f + %.2f + %.2f + %.2f + %.2f + %.2f + %.2f + %.2f + %.2f + %.2f + %.2f\n",
		row0, gap, row1, gap, row2, gap, row3, gap, row4, gap, row5)
}
