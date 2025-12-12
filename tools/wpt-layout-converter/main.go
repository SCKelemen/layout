package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/SCKelemen/layout"
	"golang.org/x/net/html"
)

// WPT test converter that runs layout engine to compute expected values
// This converts WPT reference tests into regression tests for our engine

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <wpt-html-file> [output-go-file]\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile)) + "_test.go"
	if len(os.Args) > 2 {
		outputFile = os.Args[2]
	}

	tests, err := parseAndLayoutWPT(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing WPT file: %v\n", err)
		os.Exit(1)
	}

	if len(tests) == 0 {
		fmt.Fprintf(os.Stderr, "No testable layouts found in %s\n", inputFile)
		os.Exit(1)
	}

	goCode := generateGoTests(tests, filepath.Base(inputFile))

	formatted, err := format.Source([]byte(goCode))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting Go code: %v\n", err)
		fmt.Fprintf(os.Stderr, "Unformatted code:\n%s\n", goCode)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, formatted, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Converted %d test(s) from %s to %s\n", len(tests), inputFile, outputFile)
}

type LayoutTest struct {
	Name      string
	Container *layout.Node
	TestType  string // "flexbox", "grid", "block"
}

func parseAndLayoutWPT(filename string) ([]LayoutTest, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	doc, err := html.Parse(f)
	if err != nil {
		return nil, err
	}

	var tests []LayoutTest
	extractAndLayout(doc, &tests)
	return tests, nil
}

func extractAndLayout(n *html.Node, tests *[]LayoutTest) {
	if n.Type == html.ElementNode && n.Data == "div" {
		node := buildNodeFromHTML(n)
		if node != nil && isTestableLayout(node) {
			// Run layout to compute expected values
			constraints := layout.Loose(1000, 1000) // Default container size
			if node.Style.Width > 0 {
				constraints = layout.Tight(node.Style.Width, node.Style.Height)
			}
			layout.Layout(node, constraints)

			// Create test
			test := LayoutTest{
				Name:      fmt.Sprintf("WPT_%d", len(*tests)+1),
				Container: node,
				TestType:  detectLayoutType(node),
			}
			*tests = append(*tests, test)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractAndLayout(c, tests)
	}
}

func buildNodeFromHTML(n *html.Node) *layout.Node {
	style := getAttribute(n, "style")
	id := getAttribute(n, "id")

	if style == "" && id == "" {
		return nil // Skip divs without styling
	}

	node := &layout.Node{
		Style: parseStyle(style),
	}

	// Build children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "div" {
			child := buildNodeFromHTML(c)
			if child != nil {
				node.Children = append(node.Children, child)
			}
		}
	}

	return node
}

func parseStyle(styleStr string) layout.Style {
	s := layout.Style{}

	// Display
	if display := extractCSS(styleStr, "display"); display != "" {
		switch display {
		case "flex", "inline-flex":
			s.Display = layout.DisplayFlex
		case "grid", "inline-grid":
			s.Display = layout.DisplayGrid
		case "block":
			s.Display = layout.DisplayBlock
		}
	}

	// Flexbox properties
	if flexDir := extractCSS(styleStr, "flex-direction"); flexDir != "" {
		s.FlexDirection = cssToFlexDirection(flexDir)
	}
	if flexWrap := extractCSS(styleStr, "flex-wrap"); flexWrap != "" {
		s.FlexWrap = cssToFlexWrap(flexWrap)
	}
	if justifyContent := extractCSS(styleStr, "justify-content"); justifyContent != "" {
		s.JustifyContent = cssToJustifyContent(justifyContent)
	}
	if alignItems := extractCSS(styleStr, "align-items"); alignItems != "" {
		s.AlignItems = cssToAlignItems(alignItems)
	}
	if alignContent := extractCSS(styleStr, "align-content"); alignContent != "" {
		s.AlignContent = cssToAlignContent(alignContent)
	}

	// Sizing
	s.Width = parseCSSLength(extractCSS(styleStr, "width"))
	s.Height = parseCSSLength(extractCSS(styleStr, "height"))
	s.MinWidth = parseCSSLength(extractCSS(styleStr, "min-width"))
	s.MinHeight = parseCSSLength(extractCSS(styleStr, "min-height"))
	s.MaxWidth = parseCSSLength(extractCSS(styleStr, "max-width"))
	s.MaxHeight = parseCSSLength(extractCSS(styleStr, "max-height"))

	// Flex item properties
	if flexGrow := extractCSS(styleStr, "flex-grow"); flexGrow != "" {
		s.FlexGrow, _ = strconv.ParseFloat(flexGrow, 64)
	}
	if flexShrink := extractCSS(styleStr, "flex-shrink"); flexShrink != "" {
		s.FlexShrink, _ = strconv.ParseFloat(flexShrink, 64)
	}
	if flexBasis := extractCSS(styleStr, "flex-basis"); flexBasis != "" {
		s.FlexBasis = parseCSSLength(flexBasis)
	}

	return s
}

func isTestableLayout(node *layout.Node) bool {
	// Check if this is a layout container we can test
	return (node.Style.Display == layout.DisplayFlex ||
		node.Style.Display == layout.DisplayGrid) &&
		len(node.Children) > 0
}

func detectLayoutType(node *layout.Node) string {
	switch node.Style.Display {
	case layout.DisplayFlex:
		return "flexbox"
	case layout.DisplayGrid:
		return "grid"
	default:
		return "block"
	}
}

func generateGoTests(tests []LayoutTest, sourceFile string) string {
	var sb strings.Builder

	sb.WriteString("package layout\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"math\"\n")
	sb.WriteString("\t\"testing\"\n")
	sb.WriteString(")\n\n")
	sb.WriteString(fmt.Sprintf("// Generated from WPT test: %s\n", sourceFile))
	sb.WriteString("// Tests converted from Web Platform Tests\n")
	sb.WriteString("// Expected values computed by running our layout engine\n\n")

	for i, test := range tests {
		sb.WriteString(generateSingleTest(test, i+1))
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func generateSingleTest(test LayoutTest, idx int) string {
	var sb strings.Builder

	testName := fmt.Sprintf("TestWPT_%s_%d", strings.Title(test.TestType), idx)
	sb.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", testName))
	sb.WriteString(fmt.Sprintf("\t// WPT %s test\n", test.TestType))

	// Generate container creation code
	sb.WriteString("\troot := &Node{\n")
	sb.WriteString("\t\tStyle: Style{\n")
	generateStyleCode(&sb, test.Container.Style, "\t\t\t")
	sb.WriteString("\t\t},\n")

	// Generate children
	if len(test.Container.Children) > 0 {
		sb.WriteString("\t\tChildren: []*Node{\n")
		for _, child := range test.Container.Children {
			sb.WriteString("\t\t\t{\n")
			sb.WriteString("\t\t\t\tStyle: Style{\n")
			generateStyleCode(&sb, child.Style, "\t\t\t\t\t")
			sb.WriteString("\t\t\t\t},\n")
			sb.WriteString("\t\t\t},\n")
		}
		sb.WriteString("\t\t},\n")
	}
	sb.WriteString("\t}\n\n")

	// Layout call
	width := test.Container.Style.Width
	height := test.Container.Style.Height
	if width == 0 {
		width = test.Container.Rect.Width
	}
	if height == 0 {
		height = test.Container.Rect.Height
	}
	sb.WriteString(fmt.Sprintf("\tconstraints := Tight(%.2f, %.2f)\n", width, height))
	sb.WriteString(fmt.Sprintf("\tLayout%s(root, constraints)\n\n", strings.Title(test.TestType)))

	// Generate assertions
	sb.WriteString("\t// Container size\n")
	sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Rect.Width-%.2f) > 0.5 {\n", test.Container.Rect.Width))
	sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Container width: expected %.2f, got %%f\", root.Rect.Width)\n", test.Container.Rect.Width))
	sb.WriteString("\t}\n")
	sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Rect.Height-%.2f) > 0.5 {\n", test.Container.Rect.Height))
	sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Container height: expected %.2f, got %%f\", root.Rect.Height)\n", test.Container.Rect.Height))
	sb.WriteString("\t}\n\n")

	// Child assertions
	for i, child := range test.Container.Children {
		sb.WriteString(fmt.Sprintf("\t// Child %d position and size\n", i))
		sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Children[%d].Rect.X-%.2f) > 0.5 {\n", i, child.Rect.X))
		sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Child %d X: expected %.2f, got %%f\", %d, root.Children[%d].Rect.X)\n", i, child.Rect.X, i, i))
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Children[%d].Rect.Y-%.2f) > 0.5 {\n", i, child.Rect.Y))
		sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Child %d Y: expected %.2f, got %%f\", %d, root.Children[%d].Rect.Y)\n", i, child.Rect.Y, i, i))
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Children[%d].Rect.Width-%.2f) > 0.5 {\n", i, child.Rect.Width))
		sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Child %d Width: expected %.2f, got %%f\", %d, root.Children[%d].Rect.Width)\n", i, child.Rect.Width, i, i))
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Children[%d].Rect.Height-%.2f) > 0.5 {\n", i, child.Rect.Height))
		sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Child %d Height: expected %.2f, got %%f\", %d, root.Children[%d].Rect.Height)\n", i, child.Rect.Height, i, i))
		sb.WriteString("\t}\n")
	}

	sb.WriteString("}\n")
	return sb.String()
}

func generateStyleCode(sb *strings.Builder, style layout.Style, indent string) {
	if style.Display != layout.DisplayBlock {
		sb.WriteString(fmt.Sprintf("%sDisplay: %s,\n", indent, displayToString(style.Display)))
	}
	if style.FlexDirection != 0 {
		sb.WriteString(fmt.Sprintf("%sFlexDirection: %s,\n", indent, flexDirectionToString(style.FlexDirection)))
	}
	if style.FlexWrap != 0 {
		sb.WriteString(fmt.Sprintf("%sFlexWrap: %s,\n", indent, flexWrapToString(style.FlexWrap)))
	}
	if style.JustifyContent != 0 {
		sb.WriteString(fmt.Sprintf("%sJustifyContent: %s,\n", indent, justifyContentToString(style.JustifyContent)))
	}
	if style.AlignItems != 0 {
		sb.WriteString(fmt.Sprintf("%sAlignItems: %s,\n", indent, alignItemsToString(style.AlignItems)))
	}
	if style.AlignContent != 0 {
		sb.WriteString(fmt.Sprintf("%sAlignContent: %s,\n", indent, alignContentToString(style.AlignContent)))
	}
	if style.Width > 0 {
		sb.WriteString(fmt.Sprintf("%sWidth: %.2f,\n", indent, style.Width))
	}
	if style.Height > 0 {
		sb.WriteString(fmt.Sprintf("%sHeight: %.2f,\n", indent, style.Height))
	}
	if style.FlexGrow > 0 {
		sb.WriteString(fmt.Sprintf("%sFlexGrow: %.2f,\n", indent, style.FlexGrow))
	}
	if style.FlexShrink > 0 {
		sb.WriteString(fmt.Sprintf("%sFlexShrink: %.2f,\n", indent, style.FlexShrink))
	}
	if style.FlexBasis > 0 {
		sb.WriteString(fmt.Sprintf("%sFlexBasis: %.2f,\n", indent, style.FlexBasis))
	}
}

// Helper functions
func getAttribute(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func extractCSS(style, property string) string {
	re := regexp.MustCompile(fmt.Sprintf(`%s\s*:\s*([^;]+)`, regexp.QuoteMeta(property)))
	matches := re.FindStringSubmatch(style)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func parseCSSLength(value string) float64 {
	if value == "" || value == "auto" {
		return 0
	}
	re := regexp.MustCompile(`([\d.]+)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) > 1 {
		if f, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return f
		}
	}
	return 0
}

// CSS to Go enum converters
func cssToFlexDirection(css string) layout.FlexDirection {
	switch css {
	case "row":
		return layout.FlexDirectionRow
	case "column":
		return layout.FlexDirectionColumn
	case "row-reverse":
		return layout.FlexDirectionRowReverse
	case "column-reverse":
		return layout.FlexDirectionColumnReverse
	}
	return layout.FlexDirectionRow
}

func cssToFlexWrap(css string) layout.FlexWrap {
	switch css {
	case "wrap":
		return layout.FlexWrapWrap
	case "wrap-reverse":
		return layout.FlexWrapWrapReverse
	case "nowrap":
		return layout.FlexWrapNoWrap
	}
	return layout.FlexWrapNoWrap
}

func cssToJustifyContent(css string) layout.JustifyContent {
	switch css {
	case "flex-start", "start":
		return layout.JustifyContentFlexStart
	case "flex-end", "end":
		return layout.JustifyContentFlexEnd
	case "center":
		return layout.JustifyContentCenter
	case "space-between":
		return layout.JustifyContentSpaceBetween
	case "space-around":
		return layout.JustifyContentSpaceAround
	case "space-evenly":
		return layout.JustifyContentSpaceEvenly
	}
	return layout.JustifyContentFlexStart
}

func cssToAlignItems(css string) layout.AlignItems {
	switch css {
	case "flex-start", "start":
		return layout.AlignItemsFlexStart
	case "flex-end", "end":
		return layout.AlignItemsFlexEnd
	case "center":
		return layout.AlignItemsCenter
	case "stretch":
		return layout.AlignItemsStretch
	case "baseline":
		return layout.AlignItemsBaseline
	}
	return layout.AlignItemsStretch
}

func cssToAlignContent(css string) layout.AlignContent {
	switch css {
	case "flex-start", "start":
		return layout.AlignContentFlexStart
	case "flex-end", "end":
		return layout.AlignContentFlexEnd
	case "center":
		return layout.AlignContentCenter
	case "stretch":
		return layout.AlignContentStretch
	case "space-between":
		return layout.AlignContentSpaceBetween
	case "space-around":
		return layout.AlignContentSpaceAround
	}
	return layout.AlignContentStretch
}

// Enum to string converters for code generation
func displayToString(d layout.Display) string {
	switch d {
	case layout.DisplayFlex:
		return "DisplayFlex"
	case layout.DisplayGrid:
		return "DisplayGrid"
	case layout.DisplayBlock:
		return "DisplayBlock"
	}
	return "DisplayBlock"
}

func flexDirectionToString(fd layout.FlexDirection) string {
	switch fd {
	case layout.FlexDirectionRow:
		return "FlexDirectionRow"
	case layout.FlexDirectionColumn:
		return "FlexDirectionColumn"
	case layout.FlexDirectionRowReverse:
		return "FlexDirectionRowReverse"
	case layout.FlexDirectionColumnReverse:
		return "FlexDirectionColumnReverse"
	}
	return "FlexDirectionRow"
}

func flexWrapToString(fw layout.FlexWrap) string {
	switch fw {
	case layout.FlexWrapNoWrap:
		return "FlexWrapNoWrap"
	case layout.FlexWrapWrap:
		return "FlexWrapWrap"
	case layout.FlexWrapWrapReverse:
		return "FlexWrapWrapReverse"
	}
	return "FlexWrapNoWrap"
}

func justifyContentToString(jc layout.JustifyContent) string {
	switch jc {
	case layout.JustifyContentFlexStart:
		return "JustifyContentFlexStart"
	case layout.JustifyContentFlexEnd:
		return "JustifyContentFlexEnd"
	case layout.JustifyContentCenter:
		return "JustifyContentCenter"
	case layout.JustifyContentSpaceBetween:
		return "JustifyContentSpaceBetween"
	case layout.JustifyContentSpaceAround:
		return "JustifyContentSpaceAround"
	case layout.JustifyContentSpaceEvenly:
		return "JustifyContentSpaceEvenly"
	}
	return "JustifyContentFlexStart"
}

func alignItemsToString(ai layout.AlignItems) string {
	switch ai {
	case layout.AlignItemsFlexStart:
		return "AlignItemsFlexStart"
	case layout.AlignItemsFlexEnd:
		return "AlignItemsFlexEnd"
	case layout.AlignItemsCenter:
		return "AlignItemsCenter"
	case layout.AlignItemsStretch:
		return "AlignItemsStretch"
	case layout.AlignItemsBaseline:
		return "AlignItemsBaseline"
	}
	return "AlignItemsStretch"
}

func alignContentToString(ac layout.AlignContent) string {
	switch ac {
	case layout.AlignContentFlexStart:
		return "AlignContentFlexStart"
	case layout.AlignContentFlexEnd:
		return "AlignContentFlexEnd"
	case layout.AlignContentCenter:
		return "AlignContentCenter"
	case layout.AlignContentStretch:
		return "AlignContentStretch"
	case layout.AlignContentSpaceBetween:
		return "AlignContentSpaceBetween"
	case layout.AlignContentSpaceAround:
		return "AlignContentSpaceAround"
	}
	return "AlignContentStretch"
}
