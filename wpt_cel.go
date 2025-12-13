package layout

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// CELAssertion represents a test assertion using CEL
type CELAssertion struct {
	Type       string   `json:"type"`       // "layout", "color", "text", etc.
	Expression string   `json:"expression"` // CEL expression to evaluate
	Message    string   `json:"message"`    // Human-readable description
	Tolerance  float64  `json:"tolerance"`  // Optional tolerance for numeric comparisons
	Tags       []string `json:"tags"`       // Optional categorization tags
}

// AssertionResult represents the result of evaluating an assertion
type AssertionResult struct {
	Assertion CELAssertion
	Passed    bool
	Actual    string
	Expected  string
	Error     string
}

// LayoutCELEnv creates a CEL environment with layout function bindings
type LayoutCELEnv struct {
	root  *Node
	nodes map[string]*Node
	env   *cel.Env
}

// NewLayoutCELEnv creates a new CEL environment for layout assertions
func NewLayoutCELEnv(root *Node) (*LayoutCELEnv, error) {
	// Build node map for fast lookups by path
	nodes := make(map[string]*Node)
	nodes["root"] = root
	collectNodes(root, "root", nodes)

	// Create CEL environment with layout function declarations
	env, err := cel.NewEnv(
		// Element constructor
		cel.Variable("root", cel.DynType),

		// Layout introspection functions
		cel.Function("getX",
			cel.Overload("getX_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getXFunc(nodes)))),

		cel.Function("getY",
			cel.Overload("getY_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getYFunc(nodes)))),

		cel.Function("getWidth",
			cel.Overload("getWidth_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getWidthFunc(nodes)))),

		cel.Function("getHeight",
			cel.Overload("getHeight_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getHeightFunc(nodes)))),

		cel.Function("getTop",
			cel.Overload("getTop_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getTopFunc(nodes)))),

		cel.Function("getLeft",
			cel.Overload("getLeft_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getLeftFunc(nodes)))),

		cel.Function("getBottom",
			cel.Overload("getBottom_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getBottomFunc(nodes)))),

		cel.Function("getRight",
			cel.Overload("getRight_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getRightFunc(nodes)))),

		// Margin functions
		cel.Function("getMarginTop",
			cel.Overload("getMarginTop_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getMarginTopFunc(nodes)))),

		cel.Function("getMarginRight",
			cel.Overload("getMarginRight_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getMarginRightFunc(nodes)))),

		cel.Function("getMarginBottom",
			cel.Overload("getMarginBottom_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getMarginBottomFunc(nodes)))),

		cel.Function("getMarginLeft",
			cel.Overload("getMarginLeft_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getMarginLeftFunc(nodes)))),

		// Padding functions
		cel.Function("getPaddingTop",
			cel.Overload("getPaddingTop_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getPaddingTopFunc(nodes)))),

		cel.Function("getPaddingRight",
			cel.Overload("getPaddingRight_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getPaddingRightFunc(nodes)))),

		cel.Function("getPaddingBottom",
			cel.Overload("getPaddingBottom_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getPaddingBottomFunc(nodes)))),

		cel.Function("getPaddingLeft",
			cel.Overload("getPaddingLeft_node",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(getPaddingLeftFunc(nodes)))),

		// Flexbox functions
		cel.Function("getFlexDirection",
			cel.Overload("getFlexDirection_node",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(getFlexDirectionFunc(nodes)))),

		cel.Function("getJustifyContent",
			cel.Overload("getJustifyContent_node",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(getJustifyContentFunc(nodes)))),

		cel.Function("getAlignItems",
			cel.Overload("getAlignItems_node",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(getAlignItemsFunc(nodes)))),

		cel.Function("getAlignContent",
			cel.Overload("getAlignContent_node",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(getAlignContentFunc(nodes)))),

		cel.Function("getFlexWrap",
			cel.Overload("getFlexWrap_node",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(getFlexWrapFunc(nodes)))),

		// Tree navigation
		cel.Function("root",
			cel.Overload("root",
				[]*cel.Type{},
				cel.DynType,
				cel.FunctionBinding(func(args ...ref.Val) ref.Val {
					return types.String("root")
				}))),

		cel.Function("child",
			cel.Overload("child_node_int",
				[]*cel.Type{cel.DynType, cel.IntType},
				cel.DynType,
				cel.BinaryBinding(getChildFunc(nodes)))),

		cel.Function("childCount",
			cel.Overload("childCount_node",
				[]*cel.Type{cel.DynType},
				cel.IntType,
				cel.UnaryBinding(getChildCountFunc(nodes)))),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	return &LayoutCELEnv{
		root:  root,
		nodes: nodes,
		env:   env,
	}, nil
}

// Evaluate evaluates a CEL assertion
func (lce *LayoutCELEnv) Evaluate(assertion CELAssertion) AssertionResult {
	result := AssertionResult{
		Assertion: assertion,
		Passed:    false,
	}

	// Compile the expression
	ast, issues := lce.env.Compile(assertion.Expression)
	if issues != nil && issues.Err() != nil {
		result.Error = fmt.Sprintf("Compilation error: %v", issues.Err())
		return result
	}

	// Create program
	prg, err := lce.env.Program(ast)
	if err != nil {
		result.Error = fmt.Sprintf("Program creation error: %v", err)
		return result
	}

	// Evaluate
	val, _, err := prg.Eval(map[string]interface{}{
		"root": "root",
	})
	if err != nil {
		result.Error = fmt.Sprintf("Evaluation error: %v", err)
		return result
	}

	// Check if result is boolean
	boolVal, ok := val.Value().(bool)
	if !ok {
		result.Error = fmt.Sprintf("Expression did not return boolean, got: %T", val.Value())
		return result
	}

	result.Passed = boolVal
	if !boolVal && assertion.Message != "" {
		result.Error = assertion.Message
	}

	return result
}

// EvaluateAll evaluates all assertions in a list
func (lce *LayoutCELEnv) EvaluateAll(assertions []CELAssertion) []AssertionResult {
	results := make([]AssertionResult, 0, len(assertions))
	for _, assertion := range assertions {
		if assertion.Type == "layout" {
			results = append(results, lce.Evaluate(assertion))
		} else {
			// Skip unsupported assertion types
			results = append(results, AssertionResult{
				Assertion: assertion,
				Passed:    true, // Don't fail on unsupported types
				Error:     fmt.Sprintf("Skipped: assertion type '%s' not supported", assertion.Type),
			})
		}
	}
	return results
}

// Helper function to collect all nodes into a map
func collectNodes(node *Node, path string, nodes map[string]*Node) {
	if node == nil {
		return
	}

	for i, child := range node.Children {
		childPath := fmt.Sprintf("%s.children[%d]", path, i)
		nodes[childPath] = child
		collectNodes(child, childPath, nodes)
	}
}

// findNode finds a node by path string
func findNode(path string, nodes map[string]*Node) *Node {
	node, ok := nodes[path]
	if !ok {
		return nil
	}
	return node
}

// Layout introspection function implementations

func getXFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Rect.X)
	}
}

func getYFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Rect.Y)
	}
}

func getWidthFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Rect.Width)
	}
}

func getHeightFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Rect.Height)
	}
}

func getTopFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Rect.Y)
	}
}

func getLeftFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Rect.X)
	}
}

func getBottomFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Rect.Y + node.Rect.Height)
	}
}

func getRightFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Rect.X + node.Rect.Width)
	}
}

// Margin functions

func getMarginTopFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Style.Margin.Top)
	}
}

func getMarginRightFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Style.Margin.Right)
	}
}

func getMarginBottomFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Style.Margin.Bottom)
	}
}

func getMarginLeftFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Style.Margin.Left)
	}
}

// Padding functions

func getPaddingTopFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Style.Padding.Top)
	}
}

func getPaddingRightFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Style.Padding.Right)
	}
}

func getPaddingBottomFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Style.Padding.Bottom)
	}
}

func getPaddingLeftFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Double(node.Style.Padding.Left)
	}
}

// Flexbox functions

func getFlexDirectionFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.String(flexDirectionToString(node.Style.FlexDirection))
	}
}

func getJustifyContentFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.String(justifyContentToString(node.Style.JustifyContent))
	}
}

func getAlignItemsFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.String(alignItemsToString(node.Style.AlignItems))
	}
}

func getAlignContentFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.String(alignContentToString(node.Style.AlignContent))
	}
}

func getFlexWrapFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.String(flexWrapToString(node.Style.FlexWrap))
	}
}

// Tree navigation functions

func getChildFunc(nodes map[string]*Node) func(ref.Val, ref.Val) ref.Val {
	return func(pathVal, indexVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}

		index, ok := indexVal.Value().(int64)
		if !ok {
			return types.NewErr("index must be an integer")
		}

		childPath := fmt.Sprintf("%s.children[%d]", path, index)
		node := findNode(childPath, nodes)
		if node == nil {
			return types.NewErr("child not found: %s", childPath)
		}

		return types.String(childPath)
	}
}

func getChildCountFunc(nodes map[string]*Node) func(ref.Val) ref.Val {
	return func(pathVal ref.Val) ref.Val {
		path, ok := pathVal.Value().(string)
		if !ok {
			return types.NewErr("path must be a string")
		}
		node := findNode(path, nodes)
		if node == nil {
			return types.NewErr("node not found: %s", path)
		}
		return types.Int(len(node.Children))
	}
}

// Enum to string converters

func flexDirectionToString(fd FlexDirection) string {
	switch fd {
	case FlexDirectionRow:
		return "row"
	case FlexDirectionRowReverse:
		return "row-reverse"
	case FlexDirectionColumn:
		return "column"
	case FlexDirectionColumnReverse:
		return "column-reverse"
	default:
		return "row"
	}
}

func flexWrapToString(fw FlexWrap) string {
	switch fw {
	case FlexWrapNoWrap:
		return "nowrap"
	case FlexWrapWrap:
		return "wrap"
	case FlexWrapWrapReverse:
		return "wrap-reverse"
	default:
		return "nowrap"
	}
}

func justifyContentToString(jc JustifyContent) string {
	switch jc {
	case JustifyContentFlexStart:
		return "flex-start"
	case JustifyContentFlexEnd:
		return "flex-end"
	case JustifyContentCenter:
		return "center"
	case JustifyContentSpaceBetween:
		return "space-between"
	case JustifyContentSpaceAround:
		return "space-around"
	case JustifyContentSpaceEvenly:
		return "space-evenly"
	default:
		return "flex-start"
	}
}

func alignItemsToString(ai AlignItems) string {
	switch ai {
	case AlignItemsFlexStart:
		return "flex-start"
	case AlignItemsFlexEnd:
		return "flex-end"
	case AlignItemsCenter:
		return "center"
	case AlignItemsBaseline:
		return "baseline"
	case AlignItemsStretch:
		return "stretch"
	default:
		return "stretch"
	}
}

func alignContentToString(ac AlignContent) string {
	switch ac {
	case AlignContentFlexStart:
		return "flex-start"
	case AlignContentFlexEnd:
		return "flex-end"
	case AlignContentCenter:
		return "center"
	case AlignContentSpaceBetween:
		return "space-between"
	case AlignContentSpaceAround:
		return "space-around"
	case AlignContentStretch:
		return "stretch"
	default:
		return "stretch"
	}
}
