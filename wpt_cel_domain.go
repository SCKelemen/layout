package layout

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// CEL Domain Types - Proper object-oriented API for CEL expressions

// ElementRef represents a reference to a node in the tree
type ElementRef struct {
	path  string
	node  *Node
	nodes map[string]*Node // Shared node map for lookups
}

// Implement ref.Val interface for CEL
func (e *ElementRef) ConvertToNative(typeDesc reflect.Type) (interface{}, error) {
	return e, nil
}

func (e *ElementRef) ConvertToType(typeValue ref.Type) ref.Val {
	return e
}

func (e *ElementRef) Equal(other ref.Val) ref.Val {
	if o, ok := other.(*ElementRef); ok {
		return types.Bool(e.path == o.path)
	}
	return types.Bool(false)
}

func (e *ElementRef) Type() ref.Type {
	return types.NewTypeValue("Element")
}

func (e *ElementRef) Value() interface{} {
	return e
}

// Tree Navigation Methods

func (e *ElementRef) Parent() *ElementRef {
	// Extract parent path
	if e.path == "root" {
		return nil
	}
	// root.children[0].children[1] -> root.children[0]
	lastDot := strings.LastIndex(e.path, ".")
	if lastDot == -1 {
		return nil
	}
	parentPath := e.path[:lastDot]
	// Look up parent node from map
	var parentNode *Node
	if e.nodes != nil {
		parentNode = e.nodes[parentPath]
	}
	return &ElementRef{path: parentPath, node: parentNode, nodes: e.nodes}
}

func (e *ElementRef) Children() []*ElementRef {
	if e.node == nil {
		return nil
	}
	children := make([]*ElementRef, len(e.node.Children))
	for i, child := range e.node.Children {
		childPath := fmt.Sprintf("%s.children[%d]", e.path, i)
		children[i] = &ElementRef{path: childPath, node: child, nodes: e.nodes}
	}
	return children
}

func (e *ElementRef) Child(index int) *ElementRef {
	if e.node == nil || index < 0 || index >= len(e.node.Children) {
		return nil
	}
	childPath := fmt.Sprintf("%s.children[%d]", e.path, index)
	return &ElementRef{path: childPath, node: e.node.Children[index], nodes: e.nodes}
}

func (e *ElementRef) FirstChild() *ElementRef {
	return e.Child(0)
}

func (e *ElementRef) LastChild() *ElementRef {
	if e.node == nil || len(e.node.Children) == 0 {
		return nil
	}
	return e.Child(len(e.node.Children) - 1)
}

func (e *ElementRef) Siblings() []*ElementRef {
	parent := e.Parent()
	if parent == nil {
		return []*ElementRef{}
	}
	return parent.Children()
}

func (e *ElementRef) NextSibling() *ElementRef {
	parent := e.Parent()
	if parent == nil || parent.node == nil {
		return nil
	}

	// Find self in parent's children
	for i := range parent.node.Children {
		childPath := fmt.Sprintf("%s.children[%d]", parent.path, i)
		if childPath == e.path {
			if i+1 < len(parent.node.Children) {
				nextPath := fmt.Sprintf("%s.children[%d]", parent.path, i+1)
				return &ElementRef{path: nextPath, node: parent.node.Children[i+1], nodes: e.nodes}
			}
			return nil
		}
	}
	return nil
}

func (e *ElementRef) PreviousSibling() *ElementRef {
	parent := e.Parent()
	if parent == nil || parent.node == nil {
		return nil
	}

	// Find self in parent's children
	for i := range parent.node.Children {
		childPath := fmt.Sprintf("%s.children[%d]", parent.path, i)
		if childPath == e.path {
			if i > 0 {
				prevPath := fmt.Sprintf("%s.children[%d]", parent.path, i-1)
				return &ElementRef{path: prevPath, node: parent.node.Children[i-1], nodes: e.nodes}
			}
			return nil
		}
	}
	return nil
}

// Tree Query Methods

func (e *ElementRef) Descendants() []*ElementRef {
	if e.node == nil {
		return []*ElementRef{}
	}

	var descendants []*ElementRef
	var walk func(*ElementRef)
	walk = func(elem *ElementRef) {
		for _, child := range elem.Children() {
			descendants = append(descendants, child)
			walk(child)
		}
	}
	walk(e)
	return descendants
}

func (e *ElementRef) Ancestors() []*ElementRef {
	var ancestors []*ElementRef
	current := e.Parent()
	for current != nil {
		ancestors = append(ancestors, current)
		current = current.Parent()
	}
	return ancestors
}

func (e *ElementRef) Find(predicate func(*ElementRef) bool) *ElementRef {
	if predicate(e) {
		return e
	}
	for _, child := range e.Children() {
		if found := child.Find(predicate); found != nil {
			return found
		}
	}
	return nil
}

func (e *ElementRef) FindAll(predicate func(*ElementRef) bool) []*ElementRef {
	var results []*ElementRef
	if predicate(e) {
		results = append(results, e)
	}
	for _, child := range e.Children() {
		results = append(results, child.FindAll(predicate)...)
	}
	return results
}

// Layout Property Accessors

func (e *ElementRef) X() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Rect.X
}

func (e *ElementRef) Y() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Rect.Y
}

func (e *ElementRef) Width() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Rect.Width
}

func (e *ElementRef) Height() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Rect.Height
}

func (e *ElementRef) Top() float64 {
	return e.Y()
}

func (e *ElementRef) Left() float64 {
	return e.X()
}

func (e *ElementRef) Bottom() float64 {
	return e.Y() + e.Height()
}

func (e *ElementRef) Right() float64 {
	return e.X() + e.Width()
}

// Margin accessors

func (e *ElementRef) MarginTop() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.Margin.Top
}

func (e *ElementRef) MarginRight() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.Margin.Right
}

func (e *ElementRef) MarginBottom() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.Margin.Bottom
}

func (e *ElementRef) MarginLeft() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.Margin.Left
}

// Padding accessors

func (e *ElementRef) PaddingTop() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.Padding.Top
}

func (e *ElementRef) PaddingRight() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.Padding.Right
}

func (e *ElementRef) PaddingBottom() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.Padding.Bottom
}

func (e *ElementRef) PaddingLeft() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.Padding.Left
}

// Border accessors

func (e *ElementRef) BorderTop() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.Border.Top
}

func (e *ElementRef) BorderRight() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.Border.Right
}

func (e *ElementRef) BorderBottom() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.Border.Bottom
}

func (e *ElementRef) BorderLeft() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.Border.Left
}

// Size constraints

func (e *ElementRef) MinWidth() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.MinWidth
}

func (e *ElementRef) MinHeight() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.MinHeight
}

func (e *ElementRef) MaxWidth() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.MaxWidth
}

func (e *ElementRef) MaxHeight() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.MaxHeight
}

func (e *ElementRef) AspectRatio() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.AspectRatio
}

// Box model

func (e *ElementRef) BoxSizing() string {
	if e.node == nil {
		return ""
	}
	return boxSizingToString(e.node.Style.BoxSizing)
}

// Flexbox property accessors

func (e *ElementRef) FlexDirection() string {
	if e.node == nil {
		return ""
	}
	return flexDirectionToString(e.node.Style.FlexDirection)
}

func (e *ElementRef) JustifyContent() string {
	if e.node == nil {
		return ""
	}
	return justifyContentToString(e.node.Style.JustifyContent)
}

func (e *ElementRef) AlignItems() string {
	if e.node == nil {
		return ""
	}
	return alignItemsToString(e.node.Style.AlignItems)
}

func (e *ElementRef) AlignContent() string {
	if e.node == nil {
		return ""
	}
	return alignContentToString(e.node.Style.AlignContent)
}

func (e *ElementRef) FlexWrap() string {
	if e.node == nil {
		return ""
	}
	return flexWrapToString(e.node.Style.FlexWrap)
}

func (e *ElementRef) FlexGrow() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.FlexGrow
}

func (e *ElementRef) FlexShrink() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.FlexShrink
}

func (e *ElementRef) FlexBasis() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.FlexBasis
}

func (e *ElementRef) FlexGap() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.FlexGap
}

func (e *ElementRef) FlexRowGap() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.FlexRowGap
}

func (e *ElementRef) FlexColumnGap() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.FlexColumnGap
}

func (e *ElementRef) Order() int {
	if e.node == nil {
		return 0
	}
	return e.node.Style.Order
}

// Box alignment properties

func (e *ElementRef) AlignSelf() string {
	if e.node == nil {
		return ""
	}
	return alignItemsToString(e.node.Style.AlignSelf)
}

func (e *ElementRef) JustifyItems() string {
	if e.node == nil {
		return ""
	}
	return justifyItemsToString(e.node.Style.JustifyItems)
}

func (e *ElementRef) JustifySelf() string {
	if e.node == nil {
		return ""
	}
	return justifyItemsToString(e.node.Style.JustifySelf)
}

// Grid property accessors

func (e *ElementRef) GridAutoFlow() string {
	if e.node == nil {
		return ""
	}
	return gridAutoFlowToString(e.node.Style.GridAutoFlow)
}

func (e *ElementRef) GridGap() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.GridGap
}

func (e *ElementRef) GridRowGap() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.GridRowGap
}

func (e *ElementRef) GridColumnGap() float64 {
	if e.node == nil {
		return 0
	}
	return e.node.Style.GridColumnGap
}

func (e *ElementRef) GridRowStart() int {
	if e.node == nil {
		return -1
	}
	return e.node.Style.GridRowStart
}

func (e *ElementRef) GridRowEnd() int {
	if e.node == nil {
		return -1
	}
	return e.node.Style.GridRowEnd
}

func (e *ElementRef) GridColumnStart() int {
	if e.node == nil {
		return -1
	}
	return e.node.Style.GridColumnStart
}

func (e *ElementRef) GridColumnEnd() int {
	if e.node == nil {
		return -1
	}
	return e.node.Style.GridColumnEnd
}

func (e *ElementRef) GridArea() string {
	if e.node == nil {
		return ""
	}
	return e.node.Style.GridArea
}

// Utility methods

func (e *ElementRef) ChildCount() int {
	if e.node == nil {
		return 0
	}
	return len(e.node.Children)
}

func (e *ElementRef) IsRoot() bool {
	return e.path == "root"
}

func (e *ElementRef) Path() string {
	return e.path
}

// Enum to string converters (additional ones not in wpt_cel.go)

func boxSizingToString(bs BoxSizing) string {
	switch bs {
	case BoxSizingContentBox:
		return "content-box"
	case BoxSizingBorderBox:
		return "border-box"
	default:
		return "content-box"
	}
}

func justifyItemsToString(ji JustifyItems) string {
	switch ji {
	case JustifyItemsStretch:
		return "stretch"
	case JustifyItemsStart:
		return "start"
	case JustifyItemsEnd:
		return "end"
	case JustifyItemsCenter:
		return "center"
	default:
		return "stretch"
	}
}

func gridAutoFlowToString(gaf GridAutoFlow) string {
	switch gaf {
	case GridAutoFlowRow:
		return "row"
	case GridAutoFlowColumn:
		return "column"
	case GridAutoFlowRowDense:
		return "row-dense"
	case GridAutoFlowColumnDense:
		return "column-dense"
	default:
		return "row"
	}
}

// Selector Support

type Selector struct {
	query string
}

func NewSelector(query string) *Selector {
	return &Selector{query: query}
}

// Implement ref.Val interface
func (s *Selector) ConvertToNative(typeDesc reflect.Type) (interface{}, error) {
	return s, nil
}

func (s *Selector) ConvertToType(typeValue ref.Type) ref.Val {
	return s
}

func (s *Selector) Equal(other ref.Val) ref.Val {
	if o, ok := other.(*Selector); ok {
		return types.Bool(s.query == o.query)
	}
	return types.Bool(false)
}

func (s *Selector) Type() ref.Type {
	return types.NewTypeValue("Selector")
}

func (s *Selector) Value() interface{} {
	return s
}

// Tolerance represents comparison tolerance with different modes
type Tolerance struct {
	mode         string  // "exact", "absolute", "relative", "ulp"
	value        float64 // Tolerance value (for symmetric)
	minValue     float64 // Minimum tolerance (for asymmetric)
	maxValue     float64 // Maximum tolerance (for asymmetric)
	isAsymmetric bool    // Whether tolerance is asymmetric
}

// Implement ref.Val interface for Tolerance
func (t *Tolerance) ConvertToNative(typeDesc reflect.Type) (interface{}, error) {
	return t, nil
}

func (t *Tolerance) ConvertToType(typeValue ref.Type) ref.Val {
	return t
}

func (t *Tolerance) Equal(other ref.Val) ref.Val {
	if o, ok := other.(*Tolerance); ok {
		return types.Bool(t.mode == o.mode && t.value == o.value)
	}
	return types.Bool(false)
}

func (t *Tolerance) Type() ref.Type {
	return types.NewTypeValue("Tolerance")
}

func (t *Tolerance) Value() interface{} {
	return t
}

// Tolerance constructors

// ExactTolerance creates a tolerance for exact equality
func ExactTolerance() *Tolerance {
	return &Tolerance{mode: "exact", value: 0, isAsymmetric: false}
}

// AbsoluteTolerance creates a tolerance for absolute difference (symmetric)
func AbsoluteTolerance(value float64) *Tolerance {
	return &Tolerance{mode: "absolute", value: value, isAsymmetric: false}
}

// AsymmetricAbsoluteTolerance creates a tolerance for asymmetric absolute difference
// Example: AsymmetricAbsoluteTolerance(-5.0, 10.0) allows v1 to be in [v2-5.0, v2+10.0]
func AsymmetricAbsoluteTolerance(minDelta, maxDelta float64) *Tolerance {
	return &Tolerance{
		mode:         "absolute",
		minValue:     minDelta,
		maxValue:     maxDelta,
		isAsymmetric: true,
	}
}

// RelativeTolerance creates a tolerance for relative difference (percentage, symmetric)
func RelativeTolerance(percent float64) *Tolerance {
	return &Tolerance{mode: "relative", value: percent, isAsymmetric: false}
}

// AsymmetricRelativeTolerance creates a tolerance for asymmetric relative difference
// Example: AsymmetricRelativeTolerance(-5.0, 10.0) allows v1 to be in [v2*0.95, v2*1.10]
func AsymmetricRelativeTolerance(minPercent, maxPercent float64) *Tolerance {
	return &Tolerance{
		mode:         "relative",
		minValue:     minPercent,
		maxValue:     maxPercent,
		isAsymmetric: true,
	}
}

// ULPTolerance creates a tolerance based on Units in the Last Place
func ULPTolerance(ulps int) *Tolerance {
	return &Tolerance{mode: "ulp", value: float64(ulps), isAsymmetric: false}
}

// Matches checks if two values match within this tolerance
func (t *Tolerance) Matches(v1, v2 float64) bool {
	switch t.mode {
	case "exact":
		return v1 == v2
	case "absolute":
		if t.isAsymmetric {
			// Asymmetric: v1 must be in [v2 + minValue, v2 + maxValue]
			diff := v1 - v2
			return diff >= t.minValue && diff <= t.maxValue
		}
		// Symmetric: |v1 - v2| <= value
		diff := v1 - v2
		if diff < 0 {
			diff = -diff
		}
		return diff <= t.value
	case "relative":
		if v2 == 0 {
			return v1 == 0
		}
		if t.isAsymmetric {
			// Asymmetric: v1 must be in [v2 * (1 + minPercent/100), v2 * (1 + maxPercent/100)]
			ratio := v1 / v2
			minRatio := 1.0 + t.minValue/100.0
			maxRatio := 1.0 + t.maxValue/100.0
			return ratio >= minRatio && ratio <= maxRatio
		}
		// Symmetric: |v1 - v2| / v2 <= percent/100
		diff := v1 - v2
		if diff < 0 {
			diff = -diff
		}
		relativeDiff := diff / v2
		if relativeDiff < 0 {
			relativeDiff = -relativeDiff
		}
		return relativeDiff <= t.value/100.0
	case "ulp":
		// Simplified ULP comparison (always symmetric)
		diff := v1 - v2
		if diff < 0 {
			diff = -diff
		}
		// For simplicity, treat ULP as absolute difference scaled by epsilon
		epsilon := 2.220446049250313e-16 // float64 machine epsilon
		return diff <= t.value*epsilon*((v1+v2)/2.0)
	default:
		return false
	}
}

// Assertion Helpers

type Assert struct{}

// Implement ref.Val interface
func (a *Assert) ConvertToNative(typeDesc reflect.Type) (interface{}, error) {
	return a, nil
}

func (a *Assert) ConvertToType(typeValue ref.Type) ref.Val {
	return a
}

func (a *Assert) Equal(other ref.Val) ref.Val {
	return types.Bool(true)
}

func (a *Assert) Type() ref.Type {
	return types.NewTypeValue("Assert")
}

func (a *Assert) Value() interface{} {
	return a
}

// Generic assertion methods

// Equal checks exact equality (or uses tolerance if provided)
func (a *Assert) Equal2(v1, v2 float64, tol *Tolerance) bool {
	if tol == nil {
		tol = ExactTolerance()
	}
	return tol.Matches(v1, v2)
}

// Between checks if value is within range (inclusive)
func (a *Assert) Between(value, min, max float64) bool {
	return value >= min && value <= max
}

// AllEqual checks if all values in array are equal (with optional tolerance)
func (a *Assert) AllEqual(values []float64, tol *Tolerance) bool {
	if len(values) == 0 {
		return true
	}
	if tol == nil {
		tol = ExactTolerance()
	}
	first := values[0]
	for _, v := range values[1:] {
		if !tol.Matches(v, first) {
			return false
		}
	}
	return true
}

// Ascending checks if values are in ascending order
func (a *Assert) Ascending(values []float64) bool {
	for i := 1; i < len(values); i++ {
		if values[i] <= values[i-1] {
			return false
		}
	}
	return true
}

// Descending checks if values are in descending order
func (a *Assert) Descending(values []float64) bool {
	for i := 1; i < len(values); i++ {
		if values[i] >= values[i-1] {
			return false
		}
	}
	return true
}

// Increasing checks if values are strictly increasing
func (a *Assert) Increasing(values []float64) bool {
	return a.Ascending(values)
}

// Decreasing checks if values are strictly decreasing
func (a *Assert) Decreasing(values []float64) bool {
	return a.Descending(values)
}

// DomainCELEnv creates a CEL environment with domain object support
func DomainCELEnv(root *Node) (*cel.Env, *ElementRef, error) {
	// Build node map for fast lookups
	nodes := make(map[string]*Node)
	nodes["root"] = root
	collectNodes(root, "root", nodes)

	rootRef := &ElementRef{path: "root", node: root, nodes: nodes}

	env, err := cel.NewEnv(
		// Variables
		cel.Variable("root", cel.DynType),
		cel.Variable("assert", cel.DynType),

		// Element methods - would need CEL protobuf bindings for proper method syntax
		// For now, we'll use function style with element as first parameter

		// Navigation functions (taking ElementRef)
		cel.Function("parent",
			cel.Overload("parent_element",
				[]*cel.Type{cel.DynType},
				cel.DynType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						if parent := e.Parent(); parent != nil {
							return parent
						}
						return types.NullValue
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("children",
			cel.Overload("children_element",
				[]*cel.Type{cel.DynType},
				cel.ListType(cel.DynType),
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						children := e.Children()
						vals := make([]ref.Val, len(children))
						for i, child := range children {
							vals[i] = child
						}
						return types.NewDynamicList(types.DefaultTypeAdapter, vals)
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("child",
			cel.Overload("child_element_int",
				[]*cel.Type{cel.DynType, cel.IntType},
				cel.DynType,
				cel.BinaryBinding(func(elem, idx ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						if i, ok := idx.Value().(int64); ok {
							if child := e.Child(int(i)); child != nil {
								return child
							}
							return types.NullValue
						}
					}
					return types.NewErr("invalid arguments")
				}))),

		// Property accessors
		cel.Function("x",
			cel.Overload("x_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.X())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("y",
			cel.Overload("y_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.Y())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("width",
			cel.Overload("width_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.Width())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("height",
			cel.Overload("height_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.Height())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("top",
			cel.Overload("top_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.Top())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("left",
			cel.Overload("left_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.Left())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("bottom",
			cel.Overload("bottom_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.Bottom())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("right",
			cel.Overload("right_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.Right())
					}
					return types.NewErr("expected Element")
				}))),

		// Margin accessors
		cel.Function("marginTop",
			cel.Overload("marginTop_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.MarginTop())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("marginRight",
			cel.Overload("marginRight_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.MarginRight())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("marginBottom",
			cel.Overload("marginBottom_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.MarginBottom())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("marginLeft",
			cel.Overload("marginLeft_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.MarginLeft())
					}
					return types.NewErr("expected Element")
				}))),

		// Padding accessors
		cel.Function("paddingTop",
			cel.Overload("paddingTop_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.PaddingTop())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("paddingRight",
			cel.Overload("paddingRight_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.PaddingRight())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("paddingBottom",
			cel.Overload("paddingBottom_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.PaddingBottom())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("paddingLeft",
			cel.Overload("paddingLeft_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.PaddingLeft())
					}
					return types.NewErr("expected Element")
				}))),

		// Flexbox property accessors
		cel.Function("flexDirection",
			cel.Overload("flexDirection_element",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.String(e.FlexDirection())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("justifyContent",
			cel.Overload("justifyContent_element",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.String(e.JustifyContent())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("alignItems",
			cel.Overload("alignItems_element",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.String(e.AlignItems())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("alignContent",
			cel.Overload("alignContent_element",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.String(e.AlignContent())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("flexWrap",
			cel.Overload("flexWrap_element",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.String(e.FlexWrap())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("flexGrow",
			cel.Overload("flexGrow_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.FlexGrow())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("flexShrink",
			cel.Overload("flexShrink_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.FlexShrink())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("flexBasis",
			cel.Overload("flexBasis_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.FlexBasis())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("flexGap",
			cel.Overload("flexGap_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.FlexGap())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("flexRowGap",
			cel.Overload("flexRowGap_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.FlexRowGap())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("flexColumnGap",
			cel.Overload("flexColumnGap_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.FlexColumnGap())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("order",
			cel.Overload("order_element",
				[]*cel.Type{cel.DynType},
				cel.IntType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Int(e.Order())
					}
					return types.NewErr("expected Element")
				}))),

		// Border accessors
		cel.Function("borderTop",
			cel.Overload("borderTop_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.BorderTop())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("borderRight",
			cel.Overload("borderRight_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.BorderRight())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("borderBottom",
			cel.Overload("borderBottom_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.BorderBottom())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("borderLeft",
			cel.Overload("borderLeft_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.BorderLeft())
					}
					return types.NewErr("expected Element")
				}))),

		// Size constraints
		cel.Function("minWidth",
			cel.Overload("minWidth_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.MinWidth())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("minHeight",
			cel.Overload("minHeight_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.MinHeight())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("maxWidth",
			cel.Overload("maxWidth_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.MaxWidth())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("maxHeight",
			cel.Overload("maxHeight_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.MaxHeight())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("aspectRatio",
			cel.Overload("aspectRatio_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.AspectRatio())
					}
					return types.NewErr("expected Element")
				}))),

		// Box model
		cel.Function("boxSizing",
			cel.Overload("boxSizing_element",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.String(e.BoxSizing())
					}
					return types.NewErr("expected Element")
				}))),

		// Box alignment properties
		cel.Function("alignSelf",
			cel.Overload("alignSelf_element",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.String(e.AlignSelf())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("justifyItems",
			cel.Overload("justifyItems_element",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.String(e.JustifyItems())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("justifySelf",
			cel.Overload("justifySelf_element",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.String(e.JustifySelf())
					}
					return types.NewErr("expected Element")
				}))),

		// Grid properties
		cel.Function("gridAutoFlow",
			cel.Overload("gridAutoFlow_element",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.String(e.GridAutoFlow())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("gridGap",
			cel.Overload("gridGap_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.GridGap())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("gridRowGap",
			cel.Overload("gridRowGap_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.GridRowGap())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("gridColumnGap",
			cel.Overload("gridColumnGap_element",
				[]*cel.Type{cel.DynType},
				cel.DoubleType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Double(e.GridColumnGap())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("gridRowStart",
			cel.Overload("gridRowStart_element",
				[]*cel.Type{cel.DynType},
				cel.IntType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Int(e.GridRowStart())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("gridRowEnd",
			cel.Overload("gridRowEnd_element",
				[]*cel.Type{cel.DynType},
				cel.IntType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Int(e.GridRowEnd())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("gridColumnStart",
			cel.Overload("gridColumnStart_element",
				[]*cel.Type{cel.DynType},
				cel.IntType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Int(e.GridColumnStart())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("gridColumnEnd",
			cel.Overload("gridColumnEnd_element",
				[]*cel.Type{cel.DynType},
				cel.IntType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Int(e.GridColumnEnd())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("gridArea",
			cel.Overload("gridArea_element",
				[]*cel.Type{cel.DynType},
				cel.StringType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.String(e.GridArea())
					}
					return types.NewErr("expected Element")
				}))),

		// Utility functions
		cel.Function("childCount",
			cel.Overload("childCount_element",
				[]*cel.Type{cel.DynType},
				cel.IntType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						return types.Int(e.ChildCount())
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("firstChild",
			cel.Overload("firstChild_element",
				[]*cel.Type{cel.DynType},
				cel.DynType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						if child := e.FirstChild(); child != nil {
							return child
						}
						return types.NullValue
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("lastChild",
			cel.Overload("lastChild_element",
				[]*cel.Type{cel.DynType},
				cel.DynType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						if child := e.LastChild(); child != nil {
							return child
						}
						return types.NullValue
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("nextSibling",
			cel.Overload("nextSibling_element",
				[]*cel.Type{cel.DynType},
				cel.DynType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						if sibling := e.NextSibling(); sibling != nil {
							return sibling
						}
						return types.NullValue
					}
					return types.NewErr("expected Element")
				}))),

		cel.Function("previousSibling",
			cel.Overload("previousSibling_element",
				[]*cel.Type{cel.DynType},
				cel.DynType,
				cel.UnaryBinding(func(elem ref.Val) ref.Val {
					if e, ok := elem.(*ElementRef); ok {
						if sibling := e.PreviousSibling(); sibling != nil {
							return sibling
						}
						return types.NullValue
					}
					return types.NewErr("expected Element")
				}))),

		// Tolerance constructors
		cel.Function("exact",
			cel.Overload("exact_tolerance",
				[]*cel.Type{},
				cel.DynType,
				cel.FunctionBinding(func(args ...ref.Val) ref.Val {
					return ExactTolerance()
				}))),

		cel.Function("absolute",
			cel.Overload("absolute_tolerance",
				[]*cel.Type{cel.DoubleType},
				cel.DynType,
				cel.UnaryBinding(func(value ref.Val) ref.Val {
					return AbsoluteTolerance(value.Value().(float64))
				})),
			cel.Overload("absolute_tolerance_asymmetric",
				[]*cel.Type{cel.DoubleType, cel.DoubleType},
				cel.DynType,
				cel.BinaryBinding(func(minVal, maxVal ref.Val) ref.Val {
					return AsymmetricAbsoluteTolerance(
						minVal.Value().(float64),
						maxVal.Value().(float64),
					)
				}))),

		cel.Function("relative",
			cel.Overload("relative_tolerance",
				[]*cel.Type{cel.DoubleType},
				cel.DynType,
				cel.UnaryBinding(func(percent ref.Val) ref.Val {
					return RelativeTolerance(percent.Value().(float64))
				})),
			cel.Overload("relative_tolerance_asymmetric",
				[]*cel.Type{cel.DoubleType, cel.DoubleType},
				cel.DynType,
				cel.BinaryBinding(func(minPercent, maxPercent ref.Val) ref.Val {
					return AsymmetricRelativeTolerance(
						minPercent.Value().(float64),
						maxPercent.Value().(float64),
					)
				}))),

		cel.Function("ulp",
			cel.Overload("ulp_tolerance",
				[]*cel.Type{cel.IntType},
				cel.DynType,
				cel.UnaryBinding(func(ulps ref.Val) ref.Val {
					return ULPTolerance(int(ulps.Value().(int64)))
				}))),

		// Assertion helpers with tolerance support
		cel.Function("equal",
			cel.Overload("equal_double_double",
				[]*cel.Type{cel.DoubleType, cel.DoubleType},
				cel.BoolType,
				cel.BinaryBinding(func(v1, v2 ref.Val) ref.Val {
					a := &Assert{}
					return types.Bool(a.Equal2(v1.Value().(float64), v2.Value().(float64), nil))
				})),
			cel.Overload("equal_double_double_tolerance",
				[]*cel.Type{cel.DoubleType, cel.DoubleType, cel.DynType},
				cel.BoolType,
				cel.FunctionBinding(func(args ...ref.Val) ref.Val {
					v1 := args[0].Value().(float64)
					v2 := args[1].Value().(float64)
					tol, ok := args[2].(*Tolerance)
					if !ok {
						return types.NewErr("third argument must be a Tolerance")
					}
					a := &Assert{}
					return types.Bool(a.Equal2(v1, v2, tol))
				}))),

		cel.Function("between",
			cel.Overload("between_double_double_double",
				[]*cel.Type{cel.DoubleType, cel.DoubleType, cel.DoubleType},
				cel.BoolType,
				cel.FunctionBinding(func(args ...ref.Val) ref.Val {
					value := args[0].Value().(float64)
					min := args[1].Value().(float64)
					max := args[2].Value().(float64)
					a := &Assert{}
					return types.Bool(a.Between(value, min, max))
				}))),
	)

	if err != nil {
		return nil, nil, err
	}

	return env, rootRef, nil
}
