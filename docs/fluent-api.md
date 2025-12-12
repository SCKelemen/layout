# Fluent API Guide

The layout library provides a **Roslyn-style fluent API** for working with layout trees. This API offers immutable operations, powerful querying, and elegant tree transformations.

## Table of Contents

- [Design Philosophy](#design-philosophy)
- [API Overview](#api-overview)
- [Core Concepts](#core-concepts)
- [Navigation & Querying](#navigation--querying)
- [Immutable Modifications](#immutable-modifications)
- [Parent Navigation with Context](#parent-navigation-with-context)
- [Transformations](#transformations)
- [Practical Examples](#practical-examples)
- [Integration with Classic API](#integration-with-classic-api)
- [Performance Considerations](#performance-considerations)
- [Best Practices](#best-practices)

## Design Philosophy

The fluent API is designed with these principles:

1. **Immutability by default** - All fluent methods return new nodes; originals are never modified
2. **Copy-on-write semantics** - Shallow copies share children until modified
3. **Method chaining** - Build complex trees with readable, composable operations
4. **Zero breaking changes** - Complete backward compatibility with existing code
5. **Performance-conscious** - Minimal allocations, early termination where possible

### Implementation Approach

Unlike wrapper-based fluent APIs (e.g., `NodeFluent` struct), this library implements fluent methods **directly on `*Node`**:

```go
// Methods are on *Node itself
func (n *Node) WithPadding(amount float64) *Node
func (n *Node) FindAll(predicate func(*Node) bool) []*Node
func (n *Node) Map(transform func(*Node) *Node) *Node
```

**Advantages:**
- No wrapper/unwrapping overhead
- Natural integration with existing code
- Single source of truth (no separate `NodeFluent` type)
- Works seamlessly with classic API

**Trade-offs:**
- Fields remain exported (users can still mutate directly if needed)
- Requires discipline to use immutable patterns consistently

## API Overview

The fluent API consists of four main categories:

### 1. Navigation & Querying
Find and traverse nodes in your tree:
- `Descendants()`, `DescendantsAndSelf()`
- `FirstChild()`, `LastChild()`, `ChildAt(index)`
- `Find(predicate)`, `FindAll(predicate)`
- `Any(predicate)`, `All(predicate)`
- `Where(predicate)` (alias for FindAll)
- `OfDisplayType(display)`

### 2. Immutable Modifications
Create modified copies without changing originals:
- `Clone()`, `CloneDeep()`
- `WithStyle()`, `WithPadding()`, `WithMargin()`
- `WithWidth()`, `WithHeight()`, `WithDisplay()`
- `WithChildren()`, `AddChild()`, `AddChildren()`
- `RemoveChildAt()`, `ReplaceChildAt()`, `InsertChildAt()`

### 3. Parent Navigation (via Context)
Walk up the tree to find ancestors:
- `NewContext(root)` - wrap node for parent tracking
- `Parent()`, `Ancestors()`, `Root()`
- `Siblings()`, `Depth()`
- `FindUp(predicate)`, `FindDown(predicate)`

### 4. Transformations
Apply operations across the tree:
- `Transform(predicate, transform)` - selective transformation
- `Map(transform)` - apply to all nodes
- `Filter(predicate)` - shallow filtering
- `FilterDeep(predicate)` - recursive filtering
- `Fold(initial, fn)` - reduce to single value
- `FoldWithContext(initial, fn)` - fold with depth info

## Core Concepts

### Immutability and Cloning

All fluent methods follow copy-on-write semantics:

```go
original := &Node{Style: Style{Width: 100}}

// These create new nodes; original is unchanged
modified1 := original.WithPadding(10)
modified2 := original.WithMargin(20)

fmt.Printf("Original width: %.0f\n", original.Style.Width)     // 100
fmt.Printf("Original padding: %.0f\n", original.Style.Padding.Top) // 0
fmt.Printf("Modified padding: %.0f\n", modified1.Style.Padding.Top) // 10
```

**Two types of cloning:**

1. **Shallow Clone** (`Clone()`) - Copies node struct, shares Children slice
   - Fast (O(1))
   - Use for style modifications
   - Children array is shared until modified

2. **Deep Clone** (`CloneDeep()`) - Recursively copies entire subtree
   - Slower (O(n) where n = node count)
   - Use when you need fully independent trees
   - No shared references

```go
// Shallow clone - shares children
copy := original.Clone()
copy.Style.Width = 200
// original.Children and copy.Children point to same array

// Deep clone - independent copy
independent := original.CloneDeep()
independent.Children[0].Style.Width = 300
// original.Children[0] unchanged
```

### Nil Safety

All fluent methods handle nil receivers gracefully:

```go
var node *Node = nil

// These all safely return nil or empty results
descendants := node.Descendants()           // []
found := node.Find(func(n *Node) bool { return true }) // nil
cloned := node.Clone()                      // nil
```

This means you can chain operations without nil checks:

```go
result := root.
    Find(someCondition).
    WithPadding(10).      // If Find returns nil, this returns nil
    AddChild(newChild)     // This also returns nil
```

However, **best practice** is to check for nil when semantically meaningful:

```go
target := root.Find(isHeader)
if target == nil {
    // Handle "not found" case explicitly
    target = createDefaultHeader()
}
result := target.WithPadding(10)
```

### Method Chaining

Fluent methods return `*Node`, enabling chains:

```go
card := (&Node{}).
    WithStyle(Style{Display: DisplayFlex, FlexDirection: FlexDirectionColumn}).
    WithPadding(16).
    WithMargin(8).
    AddChild(header).
    AddChild(body).
    AddChild(footer)
```

**Chains are evaluated left-to-right**, each step operating on the result of the previous step.

## Navigation & Querying

### Basic Traversal

```go
root := HStack(
    Fixed(100, 50).WithText("A"),
    Fixed(200, 50).WithText("B"),
    Fixed(150, 50).WithText("C"),
)

// Get all descendants
allNodes := root.Descendants()
fmt.Printf("Total nodes: %d\n", len(allNodes)) // 3

// Include self
allIncludingSelf := root.DescendantsAndSelf()
fmt.Printf("With root: %d\n", len(allIncludingSelf)) // 4

// Access children
first := root.FirstChild()
last := root.LastChild()
second := root.ChildAt(1)
count := root.ChildCount()
```

### Predicate-Based Queries

```go
// Find first matching node
wide := root.Find(func(n *Node) bool {
    return n.Style.Width > 150
})

// Find all matching nodes
allWide := root.FindAll(func(n *Node) bool {
    return n.Style.Width > 100
})

// Where is an alias for FindAll (LINQ-style)
textNodes := root.Where(func(n *Node) bool {
    return n.Text != ""
})

// Check existence
hasText := root.Any(func(n *Node) bool {
    return n.Text != ""
})

allVisible := root.All(func(n *Node) bool {
    return n.Style.Display != DisplayNone
})

// Filter by display type (convenience)
allFlexboxes := root.OfDisplayType(DisplayFlex)
```

### Query Patterns

**Finding nodes by property:**

```go
// By display type
flexContainers := root.FindAll(func(n *Node) bool {
    return n.Style.Display == DisplayFlex
})

// By size
smallNodes := root.FindAll(func(n *Node) bool {
    return n.Style.Width < 100 && n.Style.Height < 100
})

// By text content
buttons := root.FindAll(func(n *Node) bool {
    return strings.HasPrefix(n.Text, "Button:")
})
```

**Checking conditions:**

```go
// Any node has flex grow
anyGrowing := root.Any(func(n *Node) bool {
    return n.Style.FlexGrow > 0
})

// All nodes are visible
allVisible := root.All(func(n *Node) bool {
    return n.Style.Display != DisplayNone
})
```

## Immutable Modifications

### Style Modifications

All `WithX` methods clone the node, modify the clone, and return it:

```go
node := &Node{Style: Style{Width: 100}}

// Individual style properties
padded := node.WithPadding(10)
margined := node.WithMargin(8)
sized := node.WithWidth(200).WithHeight(150)
flex := node.WithDisplay(DisplayFlex)
growable := node.WithFlexGrow(1)
shrinkable := node.WithFlexShrink(0)

// Custom padding/margin
custom := node.WithPaddingCustom(10, 20, 10, 20)  // top, right, bottom, left
customMargin := node.WithMarginCustom(5, 10, 5, 10)

// Replace entire style
newStyle := node.WithStyle(Style{
    Display: DisplayFlex,
    Width:   300,
    Padding: Uniform(16),
})

// Add text content
withText := node.WithText("Hello, World!")
```

### Children Modifications

Modify the children array immutably:

```go
parent := HStack()

// Replace all children
withChildren := parent.WithChildren(child1, child2, child3)

// Add single child
withOne := parent.AddChild(newChild)

// Add multiple children
withMany := parent.AddChildren(child1, child2, child3)

// Remove child by index
removed := parent.RemoveChildAt(1)

// Replace child at index
replaced := parent.ReplaceChildAt(0, newFirstChild)

// Insert child at index
inserted := parent.InsertChildAt(1, middleChild)
```

**All children modifications use copy-on-write:**

```go
original := HStack(child1, child2)
modified := original.AddChild(child3)

// original.Children still has 2 children
// modified.Children has 3 children
fmt.Printf("Original: %d children\n", len(original.Children))   // 2
fmt.Printf("Modified: %d children\n", len(modified.Children))   // 3
```

### Building Complex Trees

Method chaining makes tree construction elegant:

```go
func CreateCard(title, body string, width float64) *Node {
    return (&Node{}).
        WithStyle(Style{
            Display:       DisplayFlex,
            FlexDirection: FlexDirectionColumn,
            Width:         width,
        }).
        WithPadding(16).
        WithMargin(8).
        AddChildren(
            (&Node{}).WithText(title).WithHeight(32),
            (&Node{}).WithText(body).WithFlexGrow(1),
        )
}

dashboard := (&Node{}).
    WithStyle(Style{
        Display:        DisplayFlex,
        FlexDirection:  FlexDirectionColumn,
    }).
    AddChildren(
        CreateCard("Sales", "Q4 Results", 300),
        CreateCard("Users", "Active: 1.2M", 300),
        CreateCard("Revenue", "$2.5M", 300),
    )
```

## Parent Navigation with Context

The `*Node` API doesn't include parent pointers (to keep nodes simple and avoid circular references). For upward navigation, use `NodeContext`:

### Creating a Context

```go
root := VStack(
    HStack(
        Fixed(100, 50).WithText("Button"),
    ),
)

// Wrap root in context for parent tracking
ctx := NewContext(root)
```

### Navigating Upward

```go
// Find a node and walk up
buttonCtx := ctx.FindDown(func(n *Node) bool {
    return n.Text == "Button"
})

// Get parent
parentCtx := buttonCtx.Parent()
if parentCtx != nil {
    fmt.Printf("Parent display: %v\n", parentCtx.Node.Style.Display)
}

// Get all ancestors (nearest to furthest)
ancestors := buttonCtx.Ancestors()
for i, ancestor := range ancestors {
    fmt.Printf("Ancestor %d at depth %d\n", i, ancestor.Depth())
}

// Get root
rootCtx := buttonCtx.Root()

// Get siblings
siblings := buttonCtx.Siblings()
fmt.Printf("Has %d siblings\n", len(siblings))

// Get depth
depth := buttonCtx.Depth()
fmt.Printf("Node is %d levels deep\n", depth)
```

### Context Queries

```go
// Find ancestor matching predicate
flexCtx := buttonCtx.FindUp(func(n *Node) bool {
    return n.Style.Display == DisplayFlex
})

// Find descendant (like regular Find, but returns context)
headerCtx := ctx.FindDown(func(n *Node) bool {
    return n.Text == "Header"
})

// Find all descendants matching predicate
allFlexCtx := ctx.FindDownAll(func(n *Node) bool {
    return n.Style.Display == DisplayFlex
})
```

### Unwrapping Context

Get the underlying `*Node` from a context:

```go
ctx := NewContext(root)
node := ctx.Unwrap()  // Returns the *Node
```

### Context Utility Methods

```go
ctx := NewContext(root)

// Check if root
if ctx.IsRoot() {
    fmt.Println("This is the root node")
}

// Check if has parent
if ctx.HasParent() {
    parent := ctx.Parent()
}

// Check if has children
if ctx.HasChildren() {
    children := ctx.Children()
}
```

## Transformations

### Transform - Selective Modification

Apply a transformation to nodes matching a predicate:

```go
root := HStack(
    Fixed(100, 50),
    Fixed(200, 100),
    Fixed(150, 75),
)

// Double width of nodes under 200px
doubled := root.Transform(
    func(n *Node) bool {
        return n.Style.Width > 0 && n.Style.Width < 200
    },
    func(n *Node) *Node {
        return n.WithWidth(n.Style.Width * 2)
    },
)
// Result: widths are 200, 200 (unchanged), 300
```

### Map - Apply to All Nodes

Apply a transformation to every node in the tree:

```go
// Scale entire tree by 1.5x
scaled := root.Map(func(n *Node) *Node {
    return n.
        WithWidth(n.Style.Width * 1.5).
        WithHeight(n.Style.Height * 1.5)
})

// Add uniform padding to all nodes
padded := root.Map(func(n *Node) *Node {
    return n.WithPadding(10)
})
```

### Filter - Shallow Filtering

Keep only immediate children matching a predicate:

```go
root := HStack(
    Fixed(100, 50),   // narrow
    Fixed(200, 100),  // wide
    Fixed(150, 75),   // medium
)

// Keep only wide children
wide := root.Filter(func(n *Node) bool {
    return n.Style.Width >= 200
})
// Result: root with 1 child (200px wide node)
```

**Important:** `Filter` keeps the entire subtree of matching children. For recursive filtering, use `FilterDeep`.

### FilterDeep - Recursive Filtering

Remove non-matching nodes at all levels:

```go
tree := VStack(
    HStack(
        Fixed(100, 50),
        Fixed(200, 100),
    ).WithDisplay(DisplayNone),  // hidden container
    HStack(
        Fixed(100, 50),
    ),
)

// Remove all hidden nodes
visible := tree.FilterDeep(func(n *Node) bool {
    return n.Style.Display != DisplayNone
})
// Result: First HStack is removed entirely
```

### Fold - Reduce to Single Value

Accumulate a value across the entire tree:

```go
// Sum all widths
totalWidth := root.Fold(0.0, func(acc interface{}, n *Node) interface{} {
    return acc.(float64) + n.Style.Width
}).(float64)

// Count nodes
nodeCount := root.Fold(0, func(acc interface{}, n *Node) interface{} {
    return acc.(int) + 1
}).(int)

// Find maximum width
maxWidth := root.Fold(0.0, func(acc interface{}, n *Node) interface{} {
    current := acc.(float64)
    if n.Style.Width > current {
        return n.Style.Width
    }
    return current
}).(float64)

// Collect all text content
allText := root.Fold([]string{}, func(acc interface{}, n *Node) interface{} {
    list := acc.([]string)
    if n.Text != "" {
        list = append(list, n.Text)
    }
    return list
}).([]string)
```

### FoldWithContext - Fold with Depth

Like Fold, but provides depth information:

```go
// Count nodes at each depth
depthCounts := root.FoldWithContext(
    make(map[int]int),
    func(acc interface{}, n *Node, depth int) interface{} {
        m := acc.(map[int]int)
        m[depth]++
        return m
    },
).(map[int]int)

// Sum widths by depth
depthWidths := root.FoldWithContext(
    make(map[int]float64),
    func(acc interface{}, n *Node, depth int) interface{} {
        m := acc.(map[int]float64)
        m[depth] += n.Style.Width
        return m
    },
).(map[int]float64)

// Find maximum depth
maxDepth := root.FoldWithContext(
    0,
    func(acc interface{}, n *Node, depth int) interface{} {
        if depth > acc.(int) {
            return depth
        }
        return acc
    },
).(int)
```

## Practical Examples

### Example 1: Building a Dashboard

```go
func CreateMetricCard(title string, value string, trend float64) *Node {
    trendColor := "green"
    if trend < 0 {
        trendColor = "red"
    }

    return (&Node{}).
        WithStyle(Style{
            Display:       DisplayFlex,
            FlexDirection: FlexDirectionColumn,
            Width:         200,
        }).
        WithPadding(16).
        WithMargin(8).
        AddChildren(
            (&Node{}).WithText(title).WithHeight(24),
            (&Node{}).WithText(value).WithHeight(48),
            (&Node{}).WithText(fmt.Sprintf("%.1f%%", trend)).WithHeight(20),
        )
}

dashboard := (&Node{}).
    WithStyle(Style{
        Display:        DisplayFlex,
        FlexDirection:  FlexDirectionRow,
        JustifyContent: JustifyContentSpaceAround,
    }).
    WithPadding(20).
    AddChildren(
        CreateMetricCard("Revenue", "$125K", 12.5),
        CreateMetricCard("Users", "8.2K", -2.1),
        CreateMetricCard("Orders", "342", 8.7),
    )

// Layout the dashboard
Layout(dashboard, Loose(1000, 600))
```

### Example 2: Conditional Styling

```go
func ApplyTheme(tree *Node, darkMode bool) *Node {
    basePadding := 8.0
    baseMargin := 4.0

    if darkMode {
        basePadding = 12.0
        baseMargin = 6.0
    }

    return tree.Map(func(n *Node) *Node {
        return n.
            WithPadding(basePadding).
            WithMargin(baseMargin)
    })
}

original := CreateDashboard()
lightTheme := ApplyTheme(original, false)
darkTheme := ApplyTheme(original, true)

// Original unchanged, two themed variants created
```

### Example 3: Tree Analysis

```go
// Analyze tree structure
func AnalyzeTree(root *Node) {
    // Count nodes by display type
    displayCounts := root.Fold(
        make(map[Display]int),
        func(acc interface{}, n *Node) interface{} {
            m := acc.(map[Display]int)
            m[n.Style.Display]++
            return m
        },
    ).(map[Display]int)

    // Find deepest node
    maxDepth := root.FoldWithContext(
        0,
        func(acc interface{}, n *Node, depth int) interface{} {
            if depth > acc.(int) {
                return depth
            }
            return acc
        },
    ).(int)

    // Sum all padding
    totalPadding := root.Fold(0.0, func(acc interface{}, n *Node) interface{} {
        sum := acc.(float64)
        p := n.Style.Padding
        return sum + p.Top + p.Right + p.Bottom + p.Left
    }).(float64)

    fmt.Printf("Display types: %v\n", displayCounts)
    fmt.Printf("Max depth: %d\n", maxDepth)
    fmt.Printf("Total padding: %.2f\n", totalPadding)
}
```

### Example 4: Tree Manipulation

```go
// Remove all hidden nodes
visible := root.FilterDeep(func(n *Node) bool {
    return n.Style.Display != DisplayNone
})

// Add margin to all containers
withMargins := root.Transform(
    func(n *Node) bool {
        return len(n.Children) > 0
    },
    func(n *Node) *Node {
        return n.WithMargin(8)
    },
)

// Scale flex containers
scaled := root.Transform(
    func(n *Node) bool {
        return n.Style.Display == DisplayFlex
    },
    func(n *Node) *Node {
        return n.
            WithWidth(n.Style.Width * 1.2).
            WithHeight(n.Style.Height * 1.2)
    },
)
```

### Example 5: Finding and Modifying

```go
root := CreateComplexLayout()
ctx := NewContext(root)

// Find header and update
headerCtx := ctx.FindDown(func(n *Node) bool {
    return n.Text == "Header"
})

if headerCtx != nil {
    updatedHeader := headerCtx.Node.
        WithPadding(20).
        WithText("Updated Header")

    // Note: This creates a new node, but doesn't replace it in tree
    // For structural changes, you need to rebuild or use Transform
}

// Better approach: Transform the tree
updated := root.Transform(
    func(n *Node) bool {
        return n.Text == "Header"
    },
    func(n *Node) *Node {
        return n.
            WithPadding(20).
            WithText("Updated Header")
    },
)
```

## Integration with Classic API

The fluent API is fully compatible with the classic API. You can mix and match:

### Classic API

```go
// Classic style (still works)
node := &Node{
    Style: Style{
        Display: DisplayFlex,
        Width:   200,
    },
    Children: []*Node{
        {Style: Style{Width: 100}},
        {Style: Style{Width: 100}},
    },
}

// Classic helper functions
Padding(node, 10)
Margin(node, 8)
```

### Fluent API

```go
// Fluent style
node := (&Node{}).
    WithDisplay(DisplayFlex).
    WithWidth(200).
    WithPadding(10).
    WithMargin(8).
    AddChildren(
        (&Node{}).WithWidth(100),
        (&Node{}).WithWidth(100),
    )
```

### Mixing Both

```go
// Start with classic
classic := HStack(
    Fixed(100, 50),
    Fixed(200, 50),
)

// Apply fluent operations
enhanced := classic.
    WithPadding(16).
    AddChild(Fixed(150, 50))

// Use classic helper
Margin(enhanced, 8)

// Back to fluent
final := enhanced.WithDisplay(DisplayFlex)
```

### Equivalence

These produce identical trees:

```go
// Classic
classic := &Node{
    Style: Style{
        Display: DisplayFlex,
        Width:   200,
        Padding: Uniform(10),
        Margin:  Uniform(8),
    },
    Children: []*Node{
        {Style: Style{Width: 100}},
    },
}

// Fluent
fluent := (&Node{}).
    WithDisplay(DisplayFlex).
    WithWidth(200).
    WithPadding(10).
    WithMargin(8).
    AddChild((&Node{}).WithWidth(100))

// After layout, both trees have identical rects
Layout(classic, Loose(400, 600))
Layout(fluent, Loose(400, 600))
// classic.Children[0].Rect == fluent.Children[0].Rect
```

## Performance Considerations

### Cloning Overhead

Each `WithX` method clones the node:

```go
// This creates 4 intermediate nodes
result := node.
    WithPadding(10).      // clone 1
    WithMargin(8).        // clone 2
    WithWidth(200).       // clone 3
    WithHeight(150)       // clone 4 (final result)
```

For performance-critical code, consider:

1. **Batch modifications:**
   ```go
   // Instead of chaining many WithX calls:
   result := node.WithStyle(Style{
       Padding: Uniform(10),
       Margin:  Uniform(8),
       Width:   200,
       Height:  150,
   })
   ```

2. **Use classic API for bulk operations:**
   ```go
   node := &Node{}
   node.Style.Padding = Uniform(10)
   node.Style.Margin = Uniform(8)
   node.Style.Width = 200
   // Direct mutation is faster for initialization
   ```

### Transformation Chains

Multiple transformations create intermediate trees:

```go
// This walks the tree 3 times and creates 3 copies
result := root.
    Filter(onlyVisible).
    Transform(applyTheme).
    Map(addPadding)
```

**Optimization:** Combine operations when possible:

```go
// Single pass
result := root.Transform(
    func(n *Node) bool {
        return n.Style.Display != DisplayNone  // filter + transform in one
    },
    func(n *Node) *Node {
        return applyTheme(n).WithPadding(10)  // combine transforms
    },
)
```

### Query Performance

- `Find` stops at first match (O(n) worst case, often much better)
- `FindAll` must visit all nodes (O(n) always)
- `Any` stops at first match (O(n) worst case)
- `All` stops at first non-match (O(n) worst case)

Use early-terminating methods when possible.

### Memory Usage

- **Shallow clone** (`Clone`): ~48 bytes (one Node struct)
- **Deep clone** (`CloneDeep`): ~48 bytes × node count
- **Copy-on-write**: Children arrays shared until modified

Large trees with many transformations can allocate significantly. Profile and optimize hot paths if needed.

## Best Practices

### 1. Choose the Right Tool

**Use fluent API when:**
- Building trees declaratively
- Creating variants/alternatives
- Transforming existing trees
- Querying and filtering
- Method chaining improves readability

**Use classic API when:**
- Maximum performance is critical
- Initializing large trees
- Mutating in place is acceptable
- Interfacing with existing code

### 2. Handle Nil Explicitly

Fluent methods are nil-safe, but explicit checks improve clarity:

```go
// Works, but unclear if Find failed
result := root.Find(condition).WithPadding(10)

// Better - explicit intent
target := root.Find(condition)
if target == nil {
    target = createDefault()
}
result := target.WithPadding(10)
```

### 3. Avoid Excessive Cloning

```go
// Bad - creates many intermediate nodes
for i := 0; i < 100; i++ {
    node = node.AddChild(children[i])
}

// Good - build children array first
node = node.AddChildren(children...)
```

### 4. Use Transform Over Manual Recursion

```go
// Bad - manual recursion
func updateAll(n *Node) *Node {
    clone := n.Clone()
    clone.Style.Padding = Uniform(10)
    for i, child := range n.Children {
        clone.Children[i] = updateAll(child)
    }
    return clone
}

// Good - use Map
updated := root.Map(func(n *Node) *Node {
    return n.WithPadding(10)
})
```

### 5. Document Immutability Expectations

When writing functions that take `*Node`:

```go
// Good - documents that input is not modified
// ProcessTree analyzes the tree without modifying it.
// Returns a new tree with transformations applied.
func ProcessTree(root *Node) *Node {
    return root.Transform(...)
}

// If you DO mutate, document it:
// MutateTree modifies the input tree in place.
func MutateTree(root *Node) {
    root.Style.Width = 200
}
```

### 6. Use Context for Parent Queries

Don't try to track parents manually:

```go
// Bad - manual parent tracking
type NodeWithParent struct {
    Node   *Node
    Parent *NodeWithParent
}

// Good - use NodeContext
ctx := NewContext(root)
targetCtx := ctx.FindDown(predicate)
parentCtx := targetCtx.Parent()
```

### 7. Combine Operations for Performance

```go
// Less efficient
visible := root.FilterDeep(isVisible)
themed := visible.Map(applyTheme)
padded := themed.Map(addPadding)

// More efficient
result := root.
    FilterDeep(isVisible).
    Map(func(n *Node) *Node {
        return applyTheme(n).WithPadding(10)
    })
```

### 8. Test Both APIs

When writing tests, verify equivalence:

```go
func TestFluentEquivalence(t *testing.T) {
    // Classic
    classic := &Node{
        Style: Style{Width: 200, Padding: Uniform(10)},
        Children: []*Node{{Style: Style{Width: 100}}},
    }

    // Fluent
    fluent := (&Node{}).
        WithWidth(200).
        WithPadding(10).
        AddChild((&Node{}).WithWidth(100))

    // Layout both
    Layout(classic, Loose(400, 600))
    Layout(fluent, Loose(400, 600))

    // Assert equivalence
    if classic.Rect != fluent.Rect {
        t.Error("Trees should have identical rects")
    }
}
```

## Summary

The fluent API provides a powerful, composable way to work with layout trees:

- **Immutable operations** preserve originals and enable variant creation
- **Method chaining** creates readable, declarative tree construction
- **Rich querying** finds nodes efficiently with predicates
- **Parent navigation** (via Context) enables upward traversal
- **Transformations** apply operations across entire trees
- **Full compatibility** with classic API for gradual adoption

Choose fluent for readability and immutability, classic for performance and simplicity. Mix them as needed for optimal results.
