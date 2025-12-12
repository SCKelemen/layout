package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// WPTTest represents a test case extracted from WPT HTML
type WPTTest struct {
	Name        string
	TestType    string // "flexbox" or "text"
	Container   ContainerStyle
	Children    []ChildStyle
	TextContent TextTestContent
	Expected    ExpectedLayout
	Description string
}

// TextTestContent holds text-specific test data
type TextTestContent struct {
	Text          string
	FontSize      float64
	WhiteSpace    string
	TextOverflow  string
	TextAlign     string
	OverflowWrap  string
	WordBreak     string
	ExpectedText  string // Expected rendered text (after truncation, etc.)
	ExpectedLines int    // Expected number of lines
}

type ContainerStyle struct {
	Display       string
	FlexDirection string
	FlexWrap      string
	AlignContent  string
	JustifyContent string
	AlignItems    string
	Width         float64
	Height        float64
	ExpectedWidth float64
	ExpectedHeight float64
}

type ChildStyle struct {
	Width         float64
	Height        float64
	FlexGrow      float64
	FlexShrink    float64
	FlexBasis     string
	ExpectedWidth float64
	ExpectedHeight float64
	ExpectedX     float64
	ExpectedY     float64
}

type ExpectedLayout struct {
	ContainerWidth  float64
	ContainerHeight float64
	ChildPositions  []ChildPosition
}

type ChildPosition struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <wpt-html-file> [output-go-file]\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := "wpt_converted_test.go"
	if len(os.Args) > 2 {
		outputFile = os.Args[2]
	}

	tests, err := parseWPTFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing WPT file: %v\n", err)
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

	fmt.Printf("Converted %d test(s) to %s\n", len(tests), outputFile)
}

func parseWPTFile(filename string) ([]WPTTest, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	doc, err := html.Parse(f)
	if err != nil {
		return nil, err
	}

	var tests []WPTTest
	extractTests(doc, &tests)
	return tests, nil
}

func extractTests(n *html.Node, tests *[]WPTTest) {
	if n.Type == html.ElementNode {
		// Look for flex containers (divs with display:flex or inline-flex)
		if n.Data == "div" {
			test := extractTestFromDiv(n, tests)
			if test != nil {
				*tests = append(*tests, *test)
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractTests(c, tests)
	}
}

func extractTestFromDiv(n *html.Node, tests *[]WPTTest) *WPTTest {
	// Check if this div has text content or flexbox styles
	style := getAttribute(n, "style")
	class := getAttribute(n, "class")

	// Check if this is a text test (has text content, no flex children)
	textContent := extractTextContent(n)
	hasTextContent := textContent != ""

	// Parse inline styles and classes
	containerStyle := parseContainerStyle(style, class, n)

	// Determine test type
	if hasTextContent && containerStyle.Display != "flex" {
		// Text test
		return extractTextTest(n, style, class, textContent, containerStyle, tests)
	}

	if containerStyle.Display == "" {
		return nil // Not a flex or text container
	}

	// Flexbox test
	// Extract expected dimensions
	expectedWidth := parseDataAttribute(n, "data-expected-width")
	expectedHeight := parseDataAttribute(n, "data-expected-height")
	containerStyle.ExpectedWidth = expectedWidth
	containerStyle.ExpectedHeight = expectedHeight

	// Extract children
	var children []ChildStyle
	childNode := n.FirstChild
	childIdx := 0
	for childNode != nil {
		if childNode.Type == html.ElementNode && childNode.Data == "div" {
			childStyle := extractChildStyle(childNode, childIdx)
			children = append(children, childStyle)
			childIdx++
		}
		childNode = childNode.NextSibling
	}

	if len(children) == 0 {
		return nil
	}

	return &WPTTest{
		Name:      fmt.Sprintf("WPT_%d", len(*tests)+1),
		TestType:  "flexbox",
		Container: containerStyle,
		Children:  children,
		Expected: ExpectedLayout{
			ContainerWidth:  expectedWidth,
			ContainerHeight: expectedHeight,
		},
	}
}

// extractTextContent extracts text content from a node
func extractTextContent(n *html.Node) string {
	var text strings.Builder
	var extract func(*html.Node)
	extract = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			// Only extract direct text, not from nested divs
			if c.Type == html.TextNode {
				text.WriteString(c.Data)
			}
		}
	}
	extract(n)
	return strings.TrimSpace(text.String())
}

// extractTextTest extracts a text layout test
func extractTextTest(n *html.Node, style, class, textContent string, containerStyle ContainerStyle, tests *[]WPTTest) *WPTTest {
	textTest := TextTestContent{
		Text:     textContent,
		FontSize: parseCSSLength(extractCSSValue(style, "font-size")),
	}

	// Extract text properties
	if ws := extractCSSValue(style, "white-space"); ws != "" {
		textTest.WhiteSpace = ws
	}
	if to := extractCSSValue(style, "text-overflow"); to != "" {
		textTest.TextOverflow = to
	}
	if ta := extractCSSValue(style, "text-align"); ta != "" {
		textTest.TextAlign = ta
	}
	if ow := extractCSSValue(style, "overflow-wrap"); ow != "" {
		textTest.OverflowWrap = ow
	}
	if wb := extractCSSValue(style, "word-break"); wb != "" {
		textTest.WordBreak = wb
	}

	// Extract expected results
	textTest.ExpectedText = getAttribute(n, "data-expected-text")
	if lines := getAttribute(n, "data-expected-lines"); lines != "" {
		if l, err := strconv.Atoi(lines); err == nil {
			textTest.ExpectedLines = l
		}
	}

	expectedWidth := parseDataAttribute(n, "data-expected-width")
	expectedHeight := parseDataAttribute(n, "data-expected-height")
	containerStyle.ExpectedWidth = expectedWidth
	containerStyle.ExpectedHeight = expectedHeight

	return &WPTTest{
		Name:        fmt.Sprintf("WPT_Text_%d", len(*tests)+1),
		TestType:    "text",
		Container:   containerStyle,
		TextContent: textTest,
		Expected: ExpectedLayout{
			ContainerWidth:  expectedWidth,
			ContainerHeight: expectedHeight,
		},
	}
}

func parseContainerStyle(style, class string, n *html.Node) ContainerStyle {
	cs := ContainerStyle{}
	
	// Parse inline style
	if style != "" {
		cs.Display = extractCSSValue(style, "display")
		cs.FlexDirection = extractCSSValue(style, "flex-direction")
		cs.FlexWrap = extractCSSValue(style, "flex-wrap")
		cs.AlignContent = extractCSSValue(style, "align-content")
		cs.JustifyContent = extractCSSValue(style, "justify-content")
		cs.AlignItems = extractCSSValue(style, "align-items")
		cs.Width = parseCSSLength(extractCSSValue(style, "width"))
		cs.Height = parseCSSLength(extractCSSValue(style, "height"))
	}

	// Parse class names (common pattern in WPT tests)
	if class != "" {
		classes := strings.Fields(class)
		for _, c := range classes {
			switch c {
			case "flex", "inline-flexbox":
				if cs.Display == "" {
					cs.Display = "flex"
				}
			case "column":
				if cs.FlexDirection == "" {
					cs.FlexDirection = "column"
				}
			case "wrap":
				if cs.FlexWrap == "" {
					cs.FlexWrap = "wrap"
				}
			case "wrap-reverse":
				if cs.FlexWrap == "" {
					cs.FlexWrap = "wrap-reverse"
				}
			case "align-content-flex-start":
				if cs.AlignContent == "" {
					cs.AlignContent = "flex-start"
				}
			case "align-content-stretch":
				if cs.AlignContent == "" {
					cs.AlignContent = "stretch"
				}
			}
		}
	}

	return cs
}

func extractChildStyle(n *html.Node, idx int) ChildStyle {
	cs := ChildStyle{}
	style := getAttribute(n, "style")
	
	if style != "" {
		cs.Width = parseCSSLength(extractCSSValue(style, "width"))
		cs.Height = parseCSSLength(extractCSSValue(style, "height"))
		cs.FlexGrow = parseCSSLength(extractCSSValue(style, "flex-grow"))
		cs.FlexShrink = parseCSSLength(extractCSSValue(style, "flex-shrink"))
		cs.FlexBasis = extractCSSValue(style, "flex-basis")
	}

	// Extract expected positions from data attributes
	cs.ExpectedX = parseDataAttribute(n, "data-offset-x")
	cs.ExpectedY = parseDataAttribute(n, "data-offset-y")
	cs.ExpectedWidth = parseDataAttribute(n, "data-expected-width")
	cs.ExpectedHeight = parseDataAttribute(n, "data-expected-height")

	return cs
}

func getAttribute(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func extractCSSValue(style, property string) string {
	re := regexp.MustCompile(fmt.Sprintf(`%s\s*:\s*([^;]+)`, regexp.QuoteMeta(property)))
	matches := re.FindStringSubmatch(style)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func parseCSSLength(value string) float64 {
	if value == "" {
		return 0
	}
	// Remove units (px, em, etc.) - simplified, assumes px
	re := regexp.MustCompile(`([\d.]+)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) > 1 {
		if f, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return f
		}
	}
	return 0
}

func parseDataAttribute(n *html.Node, attr string) float64 {
	value := getAttribute(n, attr)
	if value == "" {
		return 0
	}
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}
	return 0
}

func generateGoTests(tests []WPTTest, sourceFile string) string {
	var sb strings.Builder
	
	sb.WriteString("package layout\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"math\"\n")
	sb.WriteString("\t\"testing\"\n")
	sb.WriteString(")\n\n")
	sb.WriteString(fmt.Sprintf("// Generated from WPT test: %s\n", sourceFile))
	sb.WriteString("// These tests are converted from Web Platform Tests\n\n")

	for i, test := range tests {
		sb.WriteString(generateGoTest(test, i))
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func generateGoTest(test WPTTest, idx int) string {
	if test.TestType == "text" {
		return generateTextTest(test, idx)
	}
	return generateFlexboxTest(test, idx)
}

func generateTextTest(test WPTTest, idx int) string {
	var sb strings.Builder

	testName := fmt.Sprintf("TestWPT_Text_%d", idx+1)
	if test.Name != "" {
		testName = test.Name
	}

	sb.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", testName))
	sb.WriteString("\t// WPT text test converted to Go\n")
	sb.WriteString("\tsetupFakeMetrics()\n\n")

	// Build text node
	sb.WriteString(fmt.Sprintf("\ttext := %q\n", test.TextContent.Text))
	sb.WriteString("\tnode := Text(text, Style{\n")
	if test.Container.Width > 0 {
		sb.WriteString(fmt.Sprintf("\t\tWidth: %.2f,\n", test.Container.Width))
	}
	if test.Container.Height > 0 {
		sb.WriteString(fmt.Sprintf("\t\tHeight: %.2f,\n", test.Container.Height))
	}
	sb.WriteString("\t\tTextStyle: &TextStyle{\n")
	if test.TextContent.FontSize > 0 {
		sb.WriteString(fmt.Sprintf("\t\t\tFontSize: %.2f,\n", test.TextContent.FontSize))
	}
	if test.TextContent.WhiteSpace != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tWhiteSpace: %s,\n", cssToGoWhiteSpace(test.TextContent.WhiteSpace)))
	}
	if test.TextContent.TextOverflow != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tTextOverflow: %s,\n", cssToGoTextOverflow(test.TextContent.TextOverflow)))
	}
	if test.TextContent.TextAlign != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tTextAlign: %s,\n", cssToGoTextAlign(test.TextContent.TextAlign)))
	}
	if test.TextContent.OverflowWrap != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tOverflowWrap: %s,\n", cssToGoOverflowWrap(test.TextContent.OverflowWrap)))
	}
	if test.TextContent.WordBreak != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tWordBreak: %s,\n", cssToGoWordBreak(test.TextContent.WordBreak)))
	}
	sb.WriteString("\t\t},\n")
	sb.WriteString("\t})\n\n")

	// Layout
	maxWidth := test.Container.Width
	maxHeight := test.Container.Height
	if maxWidth == 0 {
		maxWidth = 1000
	}
	if maxHeight == 0 {
		maxHeight = 1000
	}
	sb.WriteString(fmt.Sprintf("\tconstraints := Loose(%.2f, %.2f)\n", maxWidth, maxHeight))
	sb.WriteString("\tLayoutText(node, constraints)\n\n")

	// Assertions
	sb.WriteString("\tif node.TextLayout == nil {\n")
	sb.WriteString("\t\tt.Fatal(\"TextLayout should be populated\")\n")
	sb.WriteString("\t}\n\n")

	if test.TextContent.ExpectedLines > 0 {
		sb.WriteString(fmt.Sprintf("\tif len(node.TextLayout.Lines) != %d {\n", test.TextContent.ExpectedLines))
		sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Expected %d lines, got %%d\", len(node.TextLayout.Lines))\n", test.TextContent.ExpectedLines))
		sb.WriteString("\t}\n\n")
	}

	if test.Container.ExpectedWidth > 0 {
		sb.WriteString(fmt.Sprintf("\tif math.Abs(node.Rect.Width-%.2f) > 1.0 {\n", test.Container.ExpectedWidth))
		sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Width should be %.2f, got %%f\", node.Rect.Width)\n", test.Container.ExpectedWidth))
		sb.WriteString("\t}\n")
	}

	if test.Container.ExpectedHeight > 0 {
		sb.WriteString(fmt.Sprintf("\tif math.Abs(node.Rect.Height-%.2f) > 1.0 {\n", test.Container.ExpectedHeight))
		sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Height should be %.2f, got %%f\", node.Rect.Height)\n", test.Container.ExpectedHeight))
		sb.WriteString("\t}\n")
	}

	sb.WriteString("}\n")
	return sb.String()
}

func generateFlexboxTest(test WPTTest, idx int) string {
	var sb strings.Builder

	testName := fmt.Sprintf("TestWPT_%d", idx+1)
	if test.Name != "" {
		testName = test.Name
	}

	sb.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", testName))
	sb.WriteString("\t// WPT test converted to Go\n")

	// Build container style
	sb.WriteString("\troot := &Node{\n")
	sb.WriteString("\t\tStyle: Style{\n")
	sb.WriteString("\t\t\tDisplay: DisplayFlex,\n")
	
	if test.Container.FlexDirection != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tFlexDirection: %s,\n", cssToGoFlexDirection(test.Container.FlexDirection)))
	}
	if test.Container.FlexWrap != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tFlexWrap: %s,\n", cssToGoFlexWrap(test.Container.FlexWrap)))
	}
	if test.Container.AlignContent != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tAlignContent: %s,\n", cssToGoAlignContent(test.Container.AlignContent)))
	}
	if test.Container.JustifyContent != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tJustifyContent: %s,\n", cssToGoJustifyContent(test.Container.JustifyContent)))
	}
	if test.Container.AlignItems != "" {
		sb.WriteString(fmt.Sprintf("\t\t\tAlignItems: %s,\n", cssToGoAlignItems(test.Container.AlignItems)))
	}
	if test.Container.Width > 0 {
		sb.WriteString(fmt.Sprintf("\t\t\tWidth: %.2f,\n", test.Container.Width))
	}
	if test.Container.Height > 0 {
		sb.WriteString(fmt.Sprintf("\t\t\tHeight: %.2f,\n", test.Container.Height))
	}
	
	sb.WriteString("\t\t},\n")
	sb.WriteString("\t\tChildren: []*Node{\n")
	
	// Add children
	for _, child := range test.Children {
		sb.WriteString("\t\t\t{\n")
		sb.WriteString("\t\t\t\tStyle: Style{\n")
		if child.Width > 0 {
			sb.WriteString(fmt.Sprintf("\t\t\t\t\tWidth: %.2f,\n", child.Width))
		}
		if child.Height > 0 {
			sb.WriteString(fmt.Sprintf("\t\t\t\t\tHeight: %.2f,\n", child.Height))
		}
		if child.FlexGrow > 0 {
			sb.WriteString(fmt.Sprintf("\t\t\t\t\tFlexGrow: %.2f,\n", child.FlexGrow))
		}
		if child.FlexShrink > 0 {
			sb.WriteString(fmt.Sprintf("\t\t\t\t\tFlexShrink: %.2f,\n", child.FlexShrink))
		}
		sb.WriteString("\t\t\t\t},\n")
		sb.WriteString("\t\t\t},\n")
	}
	
	sb.WriteString("\t\t},\n")
	sb.WriteString("\t}\n\n")
	
	// Determine constraints
	maxWidth := test.Container.Width
	maxHeight := test.Container.Height
	if maxWidth == 0 {
		maxWidth = 1000 // Default
	}
	if maxHeight == 0 {
		maxHeight = 1000 // Default
	}
	
	sb.WriteString(fmt.Sprintf("\tconstraints := Loose(%.2f, %.2f)\n", maxWidth, maxHeight))
	sb.WriteString("\tLayoutFlexbox(root, constraints)\n\n")
	
	// Add assertions
	if test.Container.ExpectedWidth > 0 {
		sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Rect.Width-%.2f) > 1.0 {\n", test.Container.ExpectedWidth))
		sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Container width should be %.2f, got %%f\", root.Rect.Width)\n", test.Container.ExpectedWidth))
		sb.WriteString("\t}\n")
	}
	if test.Container.ExpectedHeight > 0 {
		sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Rect.Height-%.2f) > 1.0 {\n", test.Container.ExpectedHeight))
		sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Container height should be %.2f, got %%f\", root.Rect.Height)\n", test.Container.ExpectedHeight))
		sb.WriteString("\t}\n")
	}
	
	// Check child positions
	for i, child := range test.Children {
		if child.ExpectedX > 0 || child.ExpectedY > 0 {
			sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Children[%d].Rect.X-%.2f) > 1.0 {\n", i, child.ExpectedX))
			sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Child %%d X should be %.2f, got %%f\", %d, root.Children[%d].Rect.X)\n", child.ExpectedX, i, i))
			sb.WriteString("\t}\n")
			sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Children[%d].Rect.Y-%.2f) > 1.0 {\n", i, child.ExpectedY))
			sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Child %%d Y should be %.2f, got %%f\", %d, root.Children[%d].Rect.Y)\n", child.ExpectedY, i, i))
			sb.WriteString("\t}\n")
		}
		if child.ExpectedWidth > 0 {
			sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Children[%d].Rect.Width-%.2f) > 1.0 {\n", i, child.ExpectedWidth))
			sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Child %%d width should be %.2f, got %%f\", %d, root.Children[%d].Rect.Width)\n", child.ExpectedWidth, i, i))
			sb.WriteString("\t}\n")
		}
		if child.ExpectedHeight > 0 {
			sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Children[%d].Rect.Height-%.2f) > 1.0 {\n", i, child.ExpectedHeight))
			sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Child %%d height should be %.2f, got %%f\", %d, root.Children[%d].Rect.Height)\n", child.ExpectedHeight, i, i))
			sb.WriteString("\t}\n")
		}
	}
	
	sb.WriteString("}\n")
	return sb.String()
}

// CSS to Go enum conversion helpers
func cssToGoFlexDirection(css string) string {
	switch css {
	case "row":
		return "FlexDirectionRow"
	case "column":
		return "FlexDirectionColumn"
	case "row-reverse":
		return "FlexDirectionRowReverse"
	case "column-reverse":
		return "FlexDirectionColumnReverse"
	default:
		return "FlexDirectionRow"
	}
}

func cssToGoFlexWrap(css string) string {
	switch css {
	case "wrap":
		return "FlexWrapWrap"
	case "wrap-reverse":
		return "FlexWrapWrapReverse"
	case "nowrap":
		return "FlexWrapNoWrap"
	default:
		return "FlexWrapNoWrap"
	}
}

func cssToGoAlignContent(css string) string {
	switch css {
	case "flex-start":
		return "AlignContentFlexStart"
	case "flex-end":
		return "AlignContentFlexEnd"
	case "center":
		return "AlignContentCenter"
	case "stretch":
		return "AlignContentStretch"
	case "space-between":
		return "AlignContentSpaceBetween"
	case "space-around":
		return "AlignContentSpaceAround"
	default:
		return "AlignContentStretch"
	}
}

func cssToGoJustifyContent(css string) string {
	switch css {
	case "flex-start":
		return "JustifyContentFlexStart"
	case "flex-end":
		return "JustifyContentFlexEnd"
	case "center":
		return "JustifyContentCenter"
	case "space-between":
		return "JustifyContentSpaceBetween"
	case "space-around":
		return "JustifyContentSpaceAround"
	default:
		return "JustifyContentFlexStart"
	}
}

func cssToGoAlignItems(css string) string {
	switch css {
	case "flex-start":
		return "AlignItemsFlexStart"
	case "flex-end":
		return "AlignItemsFlexEnd"
	case "center":
		return "AlignItemsCenter"
	case "stretch":
		return "AlignItemsStretch"
	default:
		return "AlignItemsStretch"
	}
}

func cssToGoWhiteSpace(css string) string {
	switch css {
	case "normal":
		return "WhiteSpaceNormal"
	case "nowrap":
		return "WhiteSpaceNowrap"
	case "pre":
		return "WhiteSpacePre"
	case "pre-wrap":
		return "WhiteSpacePreWrap"
	case "pre-line":
		return "WhiteSpacePreLine"
	default:
		return "WhiteSpaceNormal"
	}
}

func cssToGoTextOverflow(css string) string {
	switch css {
	case "clip":
		return "TextOverflowClip"
	case "ellipsis":
		return "TextOverflowEllipsis"
	default:
		return "TextOverflowClip"
	}
}

func cssToGoTextAlign(css string) string {
	switch css {
	case "left":
		return "TextAlignLeft"
	case "right":
		return "TextAlignRight"
	case "center":
		return "TextAlignCenter"
	case "justify":
		return "TextAlignJustify"
	default:
		return "TextAlignLeft"
	}
}

func cssToGoOverflowWrap(css string) string {
	switch css {
	case "normal":
		return "OverflowWrapNormal"
	case "break-word":
		return "OverflowWrapBreakWord"
	case "anywhere":
		return "OverflowWrapAnywhere"
	default:
		return "OverflowWrapNormal"
	}
}

func cssToGoWordBreak(css string) string {
	switch css {
	case "normal":
		return "WordBreakNormal"
	case "break-all":
		return "WordBreakBreakAll"
	case "keep-all":
		return "WordBreakKeepAll"
	default:
		return "WordBreakNormal"
	}
}
