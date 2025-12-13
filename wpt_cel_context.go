package layout

import (
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// PathContext maintains tree paths for CEL evaluation
type PathContext struct {
	nodes       map[string]*Node
	currentPath string
}

// NewPathContext creates a context from a root node
func NewPathContext(root *Node) *PathContext {
	ctx := &PathContext{
		nodes:       make(map[string]*Node),
		currentPath: "root",
	}
	ctx.buildPathMap(root, "root")
	return ctx
}

// buildPathMap recursively builds the path-to-node mapping
func (pc *PathContext) buildPathMap(node *Node, path string) {
	if node == nil {
		return
	}

	pc.nodes[path] = node

	for i, child := range node.Children {
		childPath := fmt.Sprintf("%s.children[%d]", path, i)
		pc.buildPathMap(child, childPath)
	}
}

// GetNode returns the node at the given path
func (pc *PathContext) GetNode(path string) *Node {
	return pc.nodes[path]
}

// GetParentPath returns the parent path of a given path
func (pc *PathContext) GetParentPath(path string) string {
	// "root.children[0]" -> "root"
	// "root.children[0].children[1]" -> "root.children[0]"
	if path == "root" {
		return ""
	}

	// Find the last .children[N] segment
	lastIndex := strings.LastIndex(path, ".children[")
	if lastIndex == -1 {
		return ""
	}

	return path[:lastIndex]
}

// LayoutCELEnvWithContext wraps LayoutCELEnv with path context
type LayoutCELEnvWithContext struct {
	baseEnv *LayoutCELEnv
	context *PathContext
	env     *cel.Env
}

// NewLayoutCELEnvWithContext creates a CEL environment with path-based context support
func NewLayoutCELEnvWithContext(root *Node) (*LayoutCELEnvWithContext, error) {
	// Create base environment
	baseEnv, err := NewLayoutCELEnv(root)
	if err != nil {
		return nil, err
	}

	// Create path context
	context := NewPathContext(root)

	// Create CEL environment with this(), parent(), and root() support
	env, err := cel.NewEnv(
		cel.Function("root",
			cel.Overload("root",
				[]*cel.Type{},
				cel.StringType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.String("root")
				}))),
		cel.Function("this",
			cel.Overload("this",
				[]*cel.Type{},
				cel.StringType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.String(context.currentPath)
				}))),
		cel.Function("parent",
			cel.Overload("parent",
				[]*cel.Type{},
				cel.StringType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					parentPath := context.GetParentPath(context.currentPath)
					if parentPath == "" {
						return types.NewErr("no parent for root node")
					}
					return types.String(parentPath)
				}))),
		// Re-register all the layout functions from base environment
		cel.Function("child",
			cel.Overload("child_string_int",
				[]*cel.Type{cel.StringType, cel.IntType},
				cel.StringType,
				cel.BinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					path := lhs.(types.String)
					index := rhs.(types.Int)
					return types.String(fmt.Sprintf("%s.children[%d]", path, index))
				}))),
		cel.Function("getX",
			cel.Overload("getX_string",
				[]*cel.Type{cel.StringType},
				cel.DoubleType,
				cel.UnaryBinding(func(val ref.Val) ref.Val {
					path := string(val.(types.String))
					node := context.GetNode(path)
					if node == nil {
						return types.NewErr("node not found: %s", path)
					}
					return types.Double(node.Rect.X)
				}))),
		cel.Function("getY",
			cel.Overload("getY_string",
				[]*cel.Type{cel.StringType},
				cel.DoubleType,
				cel.UnaryBinding(func(val ref.Val) ref.Val {
					path := string(val.(types.String))
					node := context.GetNode(path)
					if node == nil {
						return types.NewErr("node not found: %s", path)
					}
					return types.Double(node.Rect.Y)
				}))),
		cel.Function("getWidth",
			cel.Overload("getWidth_string",
				[]*cel.Type{cel.StringType},
				cel.DoubleType,
				cel.UnaryBinding(func(val ref.Val) ref.Val {
					path := string(val.(types.String))
					node := context.GetNode(path)
					if node == nil {
						return types.NewErr("node not found: %s", path)
					}
					return types.Double(node.Rect.Width)
				}))),
		cel.Function("getHeight",
			cel.Overload("getHeight_string",
				[]*cel.Type{cel.StringType},
				cel.DoubleType,
				cel.UnaryBinding(func(val ref.Val) ref.Val {
					path := string(val.(types.String))
					node := context.GetNode(path)
					if node == nil {
						return types.NewErr("node not found: %s", path)
					}
					return types.Double(node.Rect.Height)
				}))),
		cel.Function("getRight",
			cel.Overload("getRight_string",
				[]*cel.Type{cel.StringType},
				cel.DoubleType,
				cel.UnaryBinding(func(val ref.Val) ref.Val {
					path := string(val.(types.String))
					node := context.GetNode(path)
					if node == nil {
						return types.NewErr("node not found: %s", path)
					}
					return types.Double(node.Rect.X + node.Rect.Width)
				}))),
		cel.Function("getBottom",
			cel.Overload("getBottom_string",
				[]*cel.Type{cel.StringType},
				cel.DoubleType,
				cel.UnaryBinding(func(val ref.Val) ref.Val {
					path := string(val.(types.String))
					node := context.GetNode(path)
					if node == nil {
						return types.NewErr("node not found: %s", path)
					}
					return types.Double(node.Rect.Y + node.Rect.Height)
				}))),
		cel.Function("getMarginTop",
			cel.Overload("getMarginTop_string",
				[]*cel.Type{cel.StringType},
				cel.DoubleType,
				cel.UnaryBinding(func(val ref.Val) ref.Val {
					path := string(val.(types.String))
					node := context.GetNode(path)
					if node == nil {
						return types.NewErr("node not found: %s", path)
					}
					return types.Double(node.Style.Margin.Top)
				}))),
		cel.Function("getMarginRight",
			cel.Overload("getMarginRight_string",
				[]*cel.Type{cel.StringType},
				cel.DoubleType,
				cel.UnaryBinding(func(val ref.Val) ref.Val {
					path := string(val.(types.String))
					node := context.GetNode(path)
					if node == nil {
						return types.NewErr("node not found: %s", path)
					}
					return types.Double(node.Style.Margin.Right)
				}))),
		cel.Function("getMarginBottom",
			cel.Overload("getMarginBottom_string",
				[]*cel.Type{cel.StringType},
				cel.DoubleType,
				cel.UnaryBinding(func(val ref.Val) ref.Val {
					path := string(val.(types.String))
					node := context.GetNode(path)
					if node == nil {
						return types.NewErr("node not found: %s", path)
					}
					return types.Double(node.Style.Margin.Bottom)
				}))),
		cel.Function("getMarginLeft",
			cel.Overload("getMarginLeft_string",
				[]*cel.Type{cel.StringType},
				cel.DoubleType,
				cel.UnaryBinding(func(val ref.Val) ref.Val {
					path := string(val.(types.String))
					node := context.GetNode(path)
					if node == nil {
						return types.NewErr("node not found: %s", path)
					}
					return types.Double(node.Style.Margin.Left)
				}))),
		cel.Function("childCount",
			cel.Overload("childCount_string",
				[]*cel.Type{cel.StringType},
				cel.IntType,
				cel.UnaryBinding(func(val ref.Val) ref.Val {
					path := string(val.(types.String))
					node := context.GetNode(path)
					if node == nil {
						return types.NewErr("node not found: %s", path)
					}
					return types.Int(len(node.Children))
				}))),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	return &LayoutCELEnvWithContext{
		baseEnv: baseEnv,
		context: context,
		env:     env,
	}, nil
}

// EvaluateAtPath evaluates an assertion at a specific path in the tree
func (lce *LayoutCELEnvWithContext) EvaluateAtPath(assertion CELAssertion, path string) AssertionResult {
	result := AssertionResult{
		Assertion: assertion,
		Passed:    false,
	}

	// Set current path for this() evaluation
	lce.context.currentPath = path

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

	// Evaluate (no variables needed, root() is a function)
	val, _, err := prg.Eval(map[string]interface{}{})
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

// Evaluate evaluates an assertion at the root path
func (lce *LayoutCELEnvWithContext) Evaluate(assertion CELAssertion) AssertionResult {
	return lce.EvaluateAtPath(assertion, "root")
}
