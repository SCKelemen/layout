# Fluent API Examples

This directory contains practical examples demonstrating the fluent API for building, querying, and transforming layout trees.

## Examples

### 1. Basic Usage (`basic.go`)

Demonstrates fundamental fluent API concepts:
- Method chaining for building trees
- Immutable modifications (WithPadding, WithMargin, etc.)
- Creating variants without modifying originals
- Chaining multiple operations

**Run:**
```bash
go run basic.go
```

**Key Concepts:**
- `WithX()` methods return new nodes
- Original trees remain unchanged
- Method chaining creates readable, declarative layouts

### 2. Dashboard (`dashboard.go`)

Builds a complete dashboard layout with reusable components:
- Metric cards with title, value, and trend
- Header with title and action button
- Chart sections and activity feeds
- Theme variants using transformations

**Run:**
```bash
go run dashboard.go
```

**Key Concepts:**
- Building reusable component functions
- Composing complex layouts from simple pieces
- Using FindAll to query the tree
- Creating themed variants with Transform

### 3. Querying and Transforming (`querying.go`)

Comprehensive guide to tree queries and transformations:
- Find, FindAll, Any, All predicates
- Fold operations for statistics
- Transform for selective modifications
- Map for applying to all nodes
- Filter and FilterDeep for pruning trees

**Run:**
```bash
go run querying.go
```

**Key Concepts:**
- Predicate-based searching
- Aggregating data with Fold
- Transforming trees immutably
- Filtering while preserving structure

### 4. Context Navigation (`context.go`)

Demonstrates parent navigation using NodeContext:
- Creating contexts for upward traversal
- Finding ancestors and siblings
- Walking up to find containing elements
- Combining context with queries
- Practical use cases for parent navigation

**Run:**
```bash
go run context.go
```

**Key Concepts:**
- `NewContext()` wraps root for parent tracking
- `FindUp()` searches ancestors
- `Siblings()`, `Parent()`, `Ancestors()`
- Context methods return contexts, not nodes
- `Unwrap()` to get underlying node

### 5. Classic vs Fluent Comparison (`comparison.go`)

Side-by-side comparison of classic and fluent APIs:
- Building identical trees both ways
- Mixing both styles seamlessly
- Demonstrating equivalence
- Showing when to use each approach

**Run:**
```bash
go run comparison.go
```

**Key Concepts:**
- Both APIs produce identical layouts
- Fluent methods work on any *Node
- Helper functions (HStack, VStack) work with both
- Choose style based on preference and use case

### 6. Form Builder (`form_builder.go`)

Real-world example building a registration form:
- Reusable form component functions
- FormField, FormRow, FormSection, ButtonGroup
- Creating form variants (compact, wide)
- Conditional fields (view-only mode)
- Simulating validation errors
- Collecting form data with Fold

**Run:**
```bash
go run form_builder.go
```

**Key Concepts:**
- Building component libraries
- Creating variants for different contexts
- Using queries to analyze forms
- Transforming based on state (errors, view-only)
- Collecting and mapping data

## Common Patterns

### Building Trees

```go
// Method chaining
card := (&Node{}).
    WithStyle(Style{Display: DisplayFlex}).
    WithPadding(16).
    AddChildren(header, body, footer)

// Using helpers
stack := HStack(
    Fixed(100, 50),
    Fixed(200, 50),
).WithPadding(10)
```

### Querying

```go
// Find first match
target := root.Find(func(n *Node) bool {
    return n.Text == "Submit"
})

// Find all matches
buttons := root.FindAll(func(n *Node) bool {
    return n.Style.Width == 100
})

// Check existence
hasText := root.Any(func(n *Node) bool {
    return n.Text != ""
})
```

### Transforming

```go
// Selective transformation
doubled := root.Transform(
    func(n *Node) bool { return n.Style.Width > 0 },
    func(n *Node) *Node { return n.WithWidth(n.Style.Width * 2) },
)

// Apply to all
scaled := root.Map(func(n *Node) *Node {
    return n.WithWidth(n.Style.Width * 1.5)
})
```

### Parent Navigation

```go
// Wrap in context
ctx := NewContext(root)

// Find and navigate up
targetCtx := ctx.FindDown(isTarget)
parentCtx := targetCtx.Parent()
ancestors := targetCtx.Ancestors()

// Find containing element
flexCtx := targetCtx.FindUp(isFlex)
```

### Statistics

```go
// Count nodes
count := root.Fold(0, func(acc interface{}, n *Node) interface{} {
    return acc.(int) + 1
}).(int)

// Sum values
total := root.Fold(0.0, func(acc interface{}, n *Node) interface{} {
    return acc.(float64) + n.Style.Width
}).(float64)

// Count by depth
depthMap := root.FoldWithContext(
    make(map[int]int),
    func(acc interface{}, n *Node, depth int) interface{} {
        m := acc.(map[int]int)
        m[depth]++
        return m
    },
).(map[int]int)
```

## Tips

1. **Start with helpers**: Use `HStack()`, `VStack()`, `Fixed()` as starting points
2. **Chain for readability**: Multiple `WithX()` calls create clear intent
3. **Use Find for querying**: Better than manual recursion
4. **Transform for bulk changes**: More efficient than manual tree walking
5. **Context for parent navigation**: When you need upward traversal
6. **Fold for statistics**: Aggregate data across the tree
7. **Original unchanged**: All fluent operations return new nodes

## Next Steps

- Read the [Fluent API Guide](../../docs/fluent-api.md) for complete documentation
- Review [Design Decisions](../../docs/fluent-api-design-decisions.md) for rationale
- Check [Integration Tests](../../fluent_integration_test.go) for equivalence verification
- See main [README](../../README.md) for overview

## Running All Examples

```bash
# Run all examples in sequence
for f in *.go; do
    echo "=== Running $f ==="
    go run "$f"
    echo
done
```
