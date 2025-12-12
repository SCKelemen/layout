package main

import (
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

// JSON structures matching wpt_renderer.js output

type WPTLayoutData struct {
	TestFile string        `json:"testFile"`
	Viewport ViewportData  `json:"viewport"`
	Elements []ElementData `json:"elements"`
	Metadata MetadataData  `json:"metadata"`
}

type ViewportData struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type ElementData struct {
	Selector string            `json:"selector"`
	TagName  string            `json:"tagName"`
	Rect     RectData          `json:"rect"`
	Computed ComputedStyleData `json:"computed"`
	Children []ChildData       `json:"children,omitempty"`
}

type RectData struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Top    float64 `json:"top"`
	Left   float64 `json:"left"`
	Bottom float64 `json:"bottom"`
	Right  float64 `json:"right"`
}

type ComputedStyleData struct {
	Display        string      `json:"display"`
	Position       string      `json:"position"`
	FlexDirection  string      `json:"flexDirection"`
	FlexWrap       string      `json:"flexWrap"`
	JustifyContent string      `json:"justifyContent"`
	AlignItems     string      `json:"alignItems"`
	AlignContent   string      `json:"alignContent"`
	Width          string      `json:"width"`
	Height         string      `json:"height"`
	MinWidth       string      `json:"minWidth"`
	MinHeight      string      `json:"minHeight"`
	MaxWidth       string      `json:"maxWidth"`
	MaxHeight      string      `json:"maxHeight"`
	Margin         SpacingData `json:"margin"`
	Padding        SpacingData `json:"padding"`
	Border         SpacingData `json:"border"`
}

type SpacingData struct {
	Top    string `json:"top"`
	Right  string `json:"right"`
	Bottom string `json:"bottom"`
	Left   string `json:"left"`
}

type ChildData struct {
	Selector string   `json:"selector"`
	Rect     RectData `json:"rect"`
}

type MetadataData struct {
	GeneratedAt    string `json:"generatedAt"`
	SourceFile     string `json:"sourceFile"`
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browserVersion"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <layout-data.json> [output_test.go]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s --batch <json-dir> <output_test.go>\n", os.Args[0])
		os.Exit(1)
	}

	if os.Args[1] == "--batch" {
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Batch mode requires json-dir and output file\n")
			os.Exit(1)
		}
		if err := generateBatchTests(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		inputFile := os.Args[1]
		outputFile := "wpt_browser_test.go"
		if len(os.Args) > 2 {
			outputFile = os.Args[2]
		}
		if err := generateTestFromJSON(inputFile, outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func generateTestFromJSON(jsonFile, outputFile string) error {
	data, err := loadLayoutData(jsonFile)
	if err != nil {
		return fmt.Errorf("loading JSON: %w", err)
	}

	code := generateGoTest(data, filepath.Base(jsonFile))

	formatted, err := format.Source([]byte(code))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: formatting failed: %v\n", err)
		formatted = []byte(code)
	}

	if err := os.WriteFile(outputFile, formatted, 0644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	fmt.Printf("Generated test from %s -> %s\n", jsonFile, outputFile)
	fmt.Printf("  Test: %s\n", data.TestFile)
	fmt.Printf("  Elements: %d\n", len(data.Elements))
	fmt.Printf("  Browser: %s\n", data.Metadata.Browser)

	return nil
}

func generateBatchTests(jsonDir, outputFile string) error {
	files, err := filepath.Glob(filepath.Join(jsonDir, "*.json"))
	if err != nil {
		return err
	}

	// Filter out summary file
	var jsonFiles []string
	for _, f := range files {
		if !strings.HasSuffix(f, "_summary.json") {
			jsonFiles = append(jsonFiles, f)
		}
	}

	if len(jsonFiles) == 0 {
		return fmt.Errorf("no JSON files found in %s", jsonDir)
	}

	var allTests []string
	testCount := 0

	for _, jsonFile := range jsonFiles {
		data, err := loadLayoutData(jsonFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", jsonFile, err)
			continue
		}

		if len(data.Elements) == 0 {
			continue
		}

		testCode := generateSingleTest(data, testCount+1)
		allTests = append(allTests, testCode)
		testCount++
	}

	code := fmt.Sprintf(`package layout

import (
	"math"
	"testing"
)

// Generated from WPT tests rendered in browser
// Expected values extracted from actual browser layout
// Tests verify our engine matches browser behavior

%s
`, strings.Join(allTests, "\n\n"))

	formatted, err := format.Source([]byte(code))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: formatting failed: %v\n", err)
		formatted = []byte(code)
	}

	if err := os.WriteFile(outputFile, formatted, 0644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	fmt.Printf("Generated %d tests from %d JSON files -> %s\n", testCount, len(jsonFiles), outputFile)
	return nil
}

func loadLayoutData(jsonFile string) (*WPTLayoutData, error) {
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}

	var layoutData WPTLayoutData
	if err := json.Unmarshal(data, &layoutData); err != nil {
		return nil, err
	}

	return &layoutData, nil
}

func generateGoTest(data *WPTLayoutData, sourceFile string) string {
	var sb strings.Builder

	sb.WriteString("package layout\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"math\"\n")
	sb.WriteString("\t\"testing\"\n")
	sb.WriteString(")\n\n")
	sb.WriteString(fmt.Sprintf("// Generated from WPT test: %s\n", data.TestFile))
	sb.WriteString(fmt.Sprintf("// Source: %s\n", sourceFile))
	sb.WriteString(fmt.Sprintf("// Browser: %s\n", data.Metadata.Browser))
	sb.WriteString("// Expected values extracted from actual browser layout\n\n")

	sb.WriteString(generateSingleTest(data, 1))

	return sb.String()
}

func generateSingleTest(data *WPTLayoutData, testNum int) string {
	var sb strings.Builder

	// Find the main container (first flex or grid element)
	var container *ElementData
	for i := range data.Elements {
		el := &data.Elements[i]
		if strings.Contains(el.Computed.Display, "flex") ||
			strings.Contains(el.Computed.Display, "grid") {
			container = el
			break
		}
	}

	if container == nil {
		return ""
	}

	testName := fmt.Sprintf("TestWPTBrowser_%d", testNum)
	sb.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", testName))
	sb.WriteString(fmt.Sprintf("\t// WPT test: %s\n", data.TestFile))
	sb.WriteString(fmt.Sprintf("\t// Browser expected values for %s\n", container.Selector))
	sb.WriteString("\n")

	// Generate container setup
	sb.WriteString("\troot := &Node{\n")
	sb.WriteString("\t\tStyle: Style{\n")
	generateStyleFromComputed(&sb, container.Computed, "\t\t\t")
	sb.WriteString("\t\t},\n")

	// Generate children
	if len(container.Children) > 0 {
		sb.WriteString("\t\tChildren: []*Node{\n")
		for range container.Children {
			sb.WriteString("\t\t\t{Style: Style{}},\n")
		}
		sb.WriteString("\t\t},\n")
	}
	sb.WriteString("\t}\n\n")

	// Layout call
	sb.WriteString(fmt.Sprintf("\tconstraints := Loose(%.2f, %.2f)\n", data.Viewport.Width, data.Viewport.Height))
	layoutType := detectLayoutType(container.Computed.Display)
	sb.WriteString(fmt.Sprintf("\tLayout%s(root, constraints)\n\n", layoutType))

	// Container assertions
	sb.WriteString("\t// Container dimensions (browser expected)\n")
	sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Rect.Width-%.2f) > 1.0 {\n", container.Rect.Width))
	sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Width: expected %.2f (browser), got %%f\", root.Rect.Width)\n", container.Rect.Width))
	sb.WriteString("\t}\n")
	sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Rect.Height-%.2f) > 1.0 {\n", container.Rect.Height))
	sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Height: expected %.2f (browser), got %%f\", root.Rect.Height)\n", container.Rect.Height))
	sb.WriteString("\t}\n\n")

	// Child assertions
	if len(container.Children) > 0 {
		sb.WriteString("\t// Child positions (browser expected)\n")
		for i, child := range container.Children {
			sb.WriteString(fmt.Sprintf("\t// Child %d\n", i))
			sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Children[%d].Rect.X-%.2f) > 1.0 {\n", i, child.Rect.X))
			sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Child %d X: expected %.2f (browser), got %%f\", %d, root.Children[%d].Rect.X)\n", i, child.Rect.X, i, i))
			sb.WriteString("\t}\n")
			sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Children[%d].Rect.Y-%.2f) > 1.0 {\n", i, child.Rect.Y))
			sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Child %d Y: expected %.2f (browser), got %%f\", %d, root.Children[%d].Rect.Y)\n", i, child.Rect.Y, i, i))
			sb.WriteString("\t}\n")
			sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Children[%d].Rect.Width-%.2f) > 1.0 {\n", i, child.Rect.Width))
			sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Child %d Width: expected %.2f (browser), got %%f\", %d, root.Children[%d].Rect.Width)\n", i, child.Rect.Width, i, i))
			sb.WriteString("\t}\n")
			sb.WriteString(fmt.Sprintf("\tif math.Abs(root.Children[%d].Rect.Height-%.2f) > 1.0 {\n", i, child.Rect.Height))
			sb.WriteString(fmt.Sprintf("\t\tt.Errorf(\"Child %d Height: expected %.2f (browser), got %%f\", %d, root.Children[%d].Rect.Height)\n", i, child.Rect.Height, i, i))
			sb.WriteString("\t}\n")
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

func generateStyleFromComputed(sb *strings.Builder, computed ComputedStyleData, indent string) {
	if computed.Display == "flex" || computed.Display == "inline-flex" {
		sb.WriteString(indent + "Display: DisplayFlex,\n")
	} else if computed.Display == "grid" || computed.Display == "inline-grid" {
		sb.WriteString(indent + "Display: DisplayGrid,\n")
	}

	if computed.FlexDirection != "" && computed.FlexDirection != "row" {
		sb.WriteString(indent + fmt.Sprintf("FlexDirection: %s,\n", cssToGoFlexDirection(computed.FlexDirection)))
	}

	if computed.FlexWrap != "" && computed.FlexWrap != "nowrap" {
		sb.WriteString(indent + fmt.Sprintf("FlexWrap: %s,\n", cssToGoFlexWrap(computed.FlexWrap)))
	}

	if computed.JustifyContent != "" && computed.JustifyContent != "normal" && computed.JustifyContent != "flex-start" {
		sb.WriteString(indent + fmt.Sprintf("JustifyContent: %s,\n", cssToGoJustifyContent(computed.JustifyContent)))
	}

	if computed.AlignItems != "" && computed.AlignItems != "normal" && computed.AlignItems != "stretch" {
		sb.WriteString(indent + fmt.Sprintf("AlignItems: %s,\n", cssToGoAlignItems(computed.AlignItems)))
	}

	if computed.AlignContent != "" && computed.AlignContent != "normal" && computed.AlignContent != "stretch" {
		sb.WriteString(indent + fmt.Sprintf("AlignContent: %s,\n", cssToGoAlignContent(computed.AlignContent)))
	}

	// Parse dimensions (browser returns px values like "300px")
	if width := parsePxValue(computed.Width); width > 0 {
		sb.WriteString(indent + fmt.Sprintf("Width: %.2f,\n", width))
	}
	if height := parsePxValue(computed.Height); height > 0 {
		sb.WriteString(indent + fmt.Sprintf("Height: %.2f,\n", height))
	}
}

func parsePxValue(val string) float64 {
	if val == "" || val == "auto" || val == "none" {
		return 0
	}
	var f float64
	fmt.Sscanf(val, "%fpx", &f)
	return f
}

func detectLayoutType(display string) string {
	switch display {
	case "flex", "inline-flex":
		return "Flexbox"
	case "grid", "inline-grid":
		return "Grid"
	default:
		return "Block"
	}
}

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
	}
	return "FlexDirectionRow"
}

func cssToGoFlexWrap(css string) string {
	switch css {
	case "wrap":
		return "FlexWrapWrap"
	case "wrap-reverse":
		return "FlexWrapWrapReverse"
	}
	return "FlexWrapNoWrap"
}

func cssToGoJustifyContent(css string) string {
	switch css {
	case "flex-start", "start":
		return "JustifyContentFlexStart"
	case "flex-end", "end":
		return "JustifyContentFlexEnd"
	case "center":
		return "JustifyContentCenter"
	case "space-between":
		return "JustifyContentSpaceBetween"
	case "space-around":
		return "JustifyContentSpaceAround"
	case "space-evenly":
		return "JustifyContentSpaceEvenly"
	}
	return "JustifyContentFlexStart"
}

func cssToGoAlignItems(css string) string {
	switch css {
	case "flex-start", "start":
		return "AlignItemsFlexStart"
	case "flex-end", "end":
		return "AlignItemsFlexEnd"
	case "center":
		return "AlignItemsCenter"
	case "baseline":
		return "AlignItemsBaseline"
	}
	return "AlignItemsStretch"
}

func cssToGoAlignContent(css string) string {
	switch css {
	case "flex-start", "start":
		return "AlignContentFlexStart"
	case "flex-end", "end":
		return "AlignContentFlexEnd"
	case "center":
		return "AlignContentCenter"
	case "space-between":
		return "AlignContentSpaceBetween"
	case "space-around":
		return "AlignContentSpaceAround"
	}
	return "AlignContentStretch"
}
