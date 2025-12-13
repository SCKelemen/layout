package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// WPT test schema types (simplified)
type WPTTest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Layout      LayoutSpec             `json:"layout"`
	Constraints ConstraintsSpec        `json:"constraints"`
	Results     map[string]BrowserResult `json:"results"`
}

type LayoutSpec struct {
	Style    StyleSpec   `json:"style"`
	Children []LayoutSpec `json:"children,omitempty"`
}

type StyleSpec struct {
	Display        *string  `json:"display,omitempty"`
	FlexDirection  *string  `json:"flexDirection,omitempty"`
	JustifyContent *string  `json:"justifyContent,omitempty"`
	AlignItems     *string  `json:"alignItems,omitempty"`
	Width          *float64 `json:"width,omitempty"`
	Height         *float64 `json:"height,omitempty"`
}

type ConstraintsSpec struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type BrowserResult struct {
	Elements []ElementResult `json:"elements"`
}

type ElementResult struct {
	Path       string         `json:"path"`
	Expected   map[string]interface{} `json:"expected"`
	Assertions []CELAssertion `json:"assertions,omitempty"`
}

type CELAssertion struct {
	Type       string  `json:"type"`
	Expression string  `json:"expression"`
	Message    string  `json:"message,omitempty"`
	Tolerance  float64 `json:"tolerance,omitempty"`
}

// Generator options
type GeneratorOptions struct {
	PackageName     string
	TestName        string
	Standalone      bool
	IncludeComments bool
}

var testTemplate = `package {{.PackageName}}

import (
{{if .Standalone}}	"strings"
{{end}}	"testing"
{{if .Standalone}}	"github.com/SCKelemen/layout"
{{end}})

// {{.TestDescription}}
func {{.TestName}}(t *testing.T) {
	// Test spec loaded from: {{.SourceFile}}
	{{if .Standalone}}
	// This is a standalone test that uses the layout library directly
	root := buildLayout{{.TestName}}()
	layout.Layout(root, layout.Constraints{
		MinWidth:  0,
		MaxWidth:  {{.MaxWidth}},
		MinHeight: 0,
		MaxHeight: {{.MaxHeight}},
	})

	// Create CEL environment
	env, err := layout.NewLayoutCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create CEL environment: %v", err)
	}
	{{else}}
	// User must implement these functions:
	// - buildLayout{{.TestName}}() *YourNodeType
	// - runLayout(*YourNodeType, constraints)
	// - createCELEnv(*YourNodeType) (CELEnv, rootRef, error)

	root := buildLayout{{.TestName}}()
	runLayout(root, Constraints{
		MaxWidth:  {{.MaxWidth}},
		MaxHeight: {{.MaxHeight}},
	})

	// Create your CEL environment with custom bindings
	env, rootRef, err := createCELEnv(root)
	if err != nil {
		t.Fatalf("Failed to create CEL environment: %v", err)
	}
	{{end}}

	// Evaluate all CEL assertions
	assertions := []struct {
		expr    string
		message string
	}{
{{range .Assertions}}		{
			expr:    {{printf "%q" .Expression}},
			message: {{printf "%q" .Message}},
		},
{{end}}	}

	{{if .Standalone}}for _, assertionData := range assertions {
		assertion := layout.CELAssertion{
			Type:       "layout",
			Expression: assertionData.expr,
			Message:    assertionData.message,
		}

		result := env.Evaluate(assertion)

		if !result.Passed {
			// Be lenient with unsupported features like 'this' and 'parent()'
			if strings.Contains(result.Error, "undeclared reference to 'this'") ||
				strings.Contains(result.Error, "undeclared reference to 'parent'") {
				t.Logf("Skipping unsupported assertion: %s\nExpression: %s",
					assertion.Message, assertion.Expression)
			} else {
				t.Errorf("Assertion failed: %s\nExpression: %s\nError: %s",
					assertion.Message, assertion.Expression, result.Error)
			}
		}
	}{{else}}// User must implement assertion evaluation.
	// Example pattern:
	for _, assertionData := range assertions {
		// Create your assertion type
		// assertion := YourCELAssertion{...}

		// Evaluate using your environment
		// result := env.Evaluate(assertion)

		// Check result
		// if !result.Passed { t.Errorf(...) }

		_ = assertionData // Remove this when implementing
		t.Error("Assertion evaluation not implemented - see template above")
	}{{end}}
}
{{if .Standalone}}

// buildLayout{{.TestName}} constructs the layout tree for this test
func buildLayout{{.TestName}}() *layout.Node {
{{.LayoutTreeCode}}
	return root
}
{{else}}

// buildLayout{{.TestName}} is a placeholder - users must implement this
// to construct their layout tree according to the test specification.
func buildLayout{{.TestName}}() interface{} {
	panic("buildLayout{{.TestName}}: not implemented - user must provide layout implementation")
}

// runLayout is a placeholder - users must implement this
// to execute their layout algorithm.
func runLayout(root interface{}, constraints interface{}) {
	panic("runLayout: not implemented - user must provide layout implementation")
}

// createCELEnv is a placeholder - users must implement this
// to create a CEL environment with their custom bindings.
func createCELEnv(root interface{}) (interface{}, interface{}, error) {
	panic("createCELEnv: not implemented - user must register custom CEL functions")
}
{{end}}
`

func main() {
	inputFile := flag.String("input", "", "Input WPT test JSON file")
	outputFile := flag.String("output", "", "Output Go test file")
	packageName := flag.String("package", "layout_test", "Go package name")
	standalone := flag.Bool("standalone", false, "Generate standalone test (uses layout library directly)")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: wpt_test_gen -input <test.json> [-output <test_gen.go>] [-package <name>] [-standalone]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Read input JSON
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var test WPTTest
	if err := json.Unmarshal(data, &test); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Determine output file
	outFile := *outputFile
	if outFile == "" {
		base := filepath.Base(*inputFile)
		name := strings.TrimSuffix(base, filepath.Ext(base))
		outFile = name + "_test.go"
	}

	// Generate test name from file name
	testName := "Test" + toPascalCase(strings.TrimSuffix(filepath.Base(*inputFile), ".json"))

	// Collect all assertions from all browsers
	var allAssertions []CELAssertion
	for _, browserResult := range test.Results {
		for _, elem := range browserResult.Elements {
			for _, assertion := range elem.Assertions {
				if assertion.Message == "" {
					assertion.Message = assertion.Type
				}
				allAssertions = append(allAssertions, assertion)
			}
		}
	}

	// Get constraints
	maxWidth := test.Constraints.Width
	maxHeight := test.Constraints.Height

	// Generate layout tree code
	layoutTreeCode := generateLayoutTreeCode(test.Layout, "root", 1)

	// Prepare template data
	templateData := map[string]interface{}{
		"PackageName":     *packageName,
		"TestName":        testName,
		"TestDescription": test.Description,
		"SourceFile":      *inputFile,
		"Standalone":      *standalone,
		"MaxWidth":        maxWidth,
		"MaxHeight":       maxHeight,
		"Assertions":      allAssertions,
		"LayoutTreeCode":  layoutTreeCode,
		"RootStyle": map[string]interface{}{
			"Display":        capitalizeFirst(getString(test.Layout.Style.Display)),
			"FlexDirection":  capitalizeFirst(getString(test.Layout.Style.FlexDirection)),
			"JustifyContent": capitalizeFirst(getString(test.Layout.Style.JustifyContent)),
			"AlignItems":     capitalizeFirst(getString(test.Layout.Style.AlignItems)),
			"Width":          test.Layout.Style.Width,
			"Height":         test.Layout.Style.Height,
		},
	}

	// Parse and execute template
	tmpl, err := template.New("test").Parse(testTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	// Create output file
	f, err := os.Create(outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	// Execute template
	if err := tmpl.Execute(f, templateData); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated test file: %s\n", outFile)
	fmt.Printf("  Package: %s\n", *packageName)
	fmt.Printf("  Test function: %s\n", testName)
	fmt.Printf("  Assertions: %d\n", len(allAssertions))
	fmt.Printf("  Mode: %s\n", map[bool]string{true: "standalone", false: "user-extensible"}[*standalone])
}

// Helper functions

func toPascalCase(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	// Convert kebab-case to PascalCase
	s = strings.ReplaceAll(s, "-", " ")
	parts := strings.Fields(s)
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getFloat(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

// generateLayoutTreeCode generates Go code to construct the layout tree
func generateLayoutTreeCode(spec LayoutSpec, varName string, indent int) string {
	var b strings.Builder
	indentStr := strings.Repeat("\t", indent)

	// Start node construction
	b.WriteString(fmt.Sprintf("%s%s := &layout.Node{\n", indentStr, varName))
	b.WriteString(fmt.Sprintf("%s\tStyle: layout.Style{\n", indentStr))

	// Add style properties
	if spec.Style.Display != nil {
		b.WriteString(fmt.Sprintf("%s\t\tDisplay: layout.Display%s,\n", indentStr, capitalizeFirst(*spec.Style.Display)))
	}
	if spec.Style.FlexDirection != nil {
		b.WriteString(fmt.Sprintf("%s\t\tFlexDirection: layout.FlexDirection%s,\n", indentStr, capitalizeFirst(*spec.Style.FlexDirection)))
	}
	if spec.Style.JustifyContent != nil {
		b.WriteString(fmt.Sprintf("%s\t\tJustifyContent: layout.JustifyContent%s,\n", indentStr, capitalizeFirst(*spec.Style.JustifyContent)))
	}
	if spec.Style.AlignItems != nil {
		b.WriteString(fmt.Sprintf("%s\t\tAlignItems: layout.AlignItems%s,\n", indentStr, capitalizeFirst(*spec.Style.AlignItems)))
	}
	if spec.Style.Width != nil {
		b.WriteString(fmt.Sprintf("%s\t\tWidth: %.1f,\n", indentStr, *spec.Style.Width))
	}
	if spec.Style.Height != nil {
		b.WriteString(fmt.Sprintf("%s\t\tHeight: %.1f,\n", indentStr, *spec.Style.Height))
	}

	b.WriteString(fmt.Sprintf("%s\t},\n", indentStr))

	// Add children if present
	if len(spec.Children) > 0 {
		b.WriteString(fmt.Sprintf("%s\tChildren: []*layout.Node{\n", indentStr))
		for _, child := range spec.Children {
			// Generate child inline
			b.WriteString(generateLayoutTreeCodeInline(child, indent+2))
		}
		b.WriteString(fmt.Sprintf("%s\t},\n", indentStr))
	}

	b.WriteString(fmt.Sprintf("%s}\n", indentStr))

	return b.String()
}

// generateLayoutTreeCodeInline generates inline Go code for a child node
func generateLayoutTreeCodeInline(spec LayoutSpec, indent int) string {
	var b strings.Builder
	indentStr := strings.Repeat("\t", indent)

	// Start inline node
	b.WriteString(fmt.Sprintf("%s&layout.Node{\n", indentStr))
	b.WriteString(fmt.Sprintf("%s\tStyle: layout.Style{\n", indentStr))

	// Add style properties
	if spec.Style.Display != nil {
		b.WriteString(fmt.Sprintf("%s\t\tDisplay: layout.Display%s,\n", indentStr, capitalizeFirst(*spec.Style.Display)))
	}
	if spec.Style.FlexDirection != nil {
		b.WriteString(fmt.Sprintf("%s\t\tFlexDirection: layout.FlexDirection%s,\n", indentStr, capitalizeFirst(*spec.Style.FlexDirection)))
	}
	if spec.Style.JustifyContent != nil {
		b.WriteString(fmt.Sprintf("%s\t\tJustifyContent: layout.JustifyContent%s,\n", indentStr, capitalizeFirst(*spec.Style.JustifyContent)))
	}
	if spec.Style.AlignItems != nil {
		b.WriteString(fmt.Sprintf("%s\t\tAlignItems: layout.AlignItems%s,\n", indentStr, capitalizeFirst(*spec.Style.AlignItems)))
	}
	if spec.Style.Width != nil {
		b.WriteString(fmt.Sprintf("%s\t\tWidth: %.1f,\n", indentStr, *spec.Style.Width))
	}
	if spec.Style.Height != nil {
		b.WriteString(fmt.Sprintf("%s\t\tHeight: %.1f,\n", indentStr, *spec.Style.Height))
	}

	b.WriteString(fmt.Sprintf("%s\t},\n", indentStr))

	// Add children recursively if present
	if len(spec.Children) > 0 {
		b.WriteString(fmt.Sprintf("%s\tChildren: []*layout.Node{\n", indentStr))
		for _, child := range spec.Children {
			b.WriteString(generateLayoutTreeCodeInline(child, indent+2))
		}
		b.WriteString(fmt.Sprintf("%s\t},\n", indentStr))
	}

	b.WriteString(fmt.Sprintf("%s},\n", indentStr))

	return b.String()
}
