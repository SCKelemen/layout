# Fluent API Design Decisions

This document addresses design decisions and trade-offs in the fluent API implementation, particularly in response to architectural review feedback.

## Table of Contents

- [Core Architecture](#core-architecture)
- [Addressing Review Feedback](#addressing-review-feedback)
- [Design Trade-offs](#design-trade-offs)
- [Future Considerations](#future-considerations)

## Core Architecture

### Direct Methods on *Node

Unlike wrapper-based fluent APIs (e.g., `type NodeFluent struct { node *Node }`), this implementation adds fluent methods **directly to `*Node`**:

```go
// Methods are on *Node itself
func (n *Node) WithPadding(amount float64) *Node
func (n *Node) FindAll(predicate func(*Node) bool) []*Node
func (n *Node) Transform(predicate, transform func(*Node) *Node) *Node
```

**Rationale:**
- Single source of truth - no wrapper/unwrapping
- Natural integration with existing code
- Works seamlessly with classic API
- No additional allocation overhead
- Simpler mental model for users

**Trade-off:**
- Fields remain exported (users can still mutate directly)
- Requires discipline to maintain immutability pattern

## Addressing Review Feedback

### 1. Exported Fields vs Encapsulation

**Review Concern:** "Exported `NodeFluent.Node` field undermines immutability story"

**Our Implementation:**
We don't use a wrapper type. Methods are directly on `*Node`, and all fields remain exported as they were before the fluent API existed.

**Design Decision:**
- Maintain 100% backward compatibility
- Users who want mutability can still use direct field access
- Users who want immutability use fluent methods
- No breaking changes to existing code

**Mitigation:**
- Comprehensive documentation emphasizes immutability pattern
- All examples use immutable fluent API
- Integration tests verify fluent operations don't mutate
- Clear "Best Practices" section in docs

**Alternative Considered:**
Create a separate `NodeBuilder` type for fluent operations:
```go
type NodeBuilder struct { node *Node }
func (nb NodeBuilder) Build() *Node { return nb.node }
```

**Why Rejected:**
- Breaks the "feel" of the fluent API
- Requires explicit wrapping/unwrapping
- More cognitive overhead for users
- Harder to integrate with classic API
- Our approach is more like jQuery, Roslyn, etc.

### 2. Zero Value Semantics

**Review Concern:** "Zero value `NodeFluent` handling needs explicit design"

**Our Implementation:**
All fluent methods handle nil receivers gracefully:

```go
func (n *Node) WithPadding(amount float64) *Node {
    if n == nil {
        return nil
    }
    copy := n.Clone()
    copy.Style.Padding = Uniform(amount)
    return copy
}
```

**Design Decision:**
Zero/nil nodes return nil (no auto-initialization). This is:
1. **Consistent** with Go idioms
2. **Explicit** - user must create nodes intentionally
3. **Safe** - nil propagates through chains

**Usage Pattern:**
```go
// User must explicitly create node
node := &Node{}  // or HStack(), VStack(), etc.
result := node.WithPadding(10)

// Nil safety
var node *Node = nil
result := node.WithPadding(10)  // returns nil, doesn't panic
```

**Documentation:**
- "Nil Safety" section explicitly covers this behavior
- Examples show proper initialization patterns

**Alternative Considered:**
Auto-initialize on nil:
```go
func (n *Node) ensure() *Node {
    if n == nil {
        return &Node{}
    }
    return n
}
```

**Why Rejected:**
- Surprising behavior (creating nodes from thin air)
- Inconsistent with Go conventions
- Harder to reason about
- Existing code might assume nil means "no node"

### 3. FindFirst Return Value

**Review Concern:** "`FindFirst` should return `(NodeFluent, bool)` to signal absence"

**Our Implementation:**
We have `Find() *Node` which returns nil if not found:

```go
func (n *Node) Find(predicate func(*Node) bool) *Node {
    // returns nil if not found
}
```

**Design Decision:**
Return nil for "not found" is:
1. **Consistent** with Go conventions (map lookups, etc.)
2. **Chainable** - nil propagates safely
3. **Simple** - no tuple unpacking

**Usage Pattern:**
```go
// Check for nil explicitly when semantically meaningful
header := root.Find(isHeader)
if header == nil {
    header = createDefaultHeader()
}

// Or chain knowing nil is safe
result := root.Find(predicate).WithPadding(10)  // nil if not found
```

**Documentation:**
- "Nil Safety" section explains this pattern
- Best practices show explicit nil checks
- Examples demonstrate both patterns

**Alternative Considered:**
```go
func (n *Node) Find(predicate func(*Node) bool) (*Node, bool)
```

**Why Rejected:**
- Breaks fluent chaining (can't call `.WithPadding()` directly)
- Inconsistent with other fluent APIs (jQuery, Roslyn, etc.)
- More verbose for common case
- Can add `MustFind` variant later if needed

### 4. Ergonomic Helpers

**Review Concern:** "Add fluent methods like `FlexRow()`, `FlexColumn()`, `Grid()`"

**Current Implementation:**
We provide high-level constructor helpers:
```go
func HStack(children ...*Node) *Node  // Flex row
func VStack(children ...*Node) *Node  // Flex column
func Grid(rows, cols int) *Node       // Grid layout
```

**Usage:**
```go
card := VStack(
    header,
    body,
    footer,
).WithPadding(16)
```

**Design Decision:**
- Helpers are package-level functions, not methods
- This is consistent with existing API patterns
- Works for both classic and fluent styles

**Additional Helpers Available:**
```go
func Fixed(width, height float64) *Node
func HStack(children ...*Node) *Node
func VStack(children ...*Node) *Node
func Grid(rows, cols int) *Node
```

**Future Consideration:**
Could add method versions if there's demand:
```go
func (n *Node) AsFlexRow() *Node {
    return n.WithStyle(Style{
        Display:       DisplayFlex,
        FlexDirection: FlexDirectionRow,
    })
}
```

**Why Not Yet:**
- Unclear naming (`.AsFlexRow()` vs `.FlexRow()` vs `.MakeFlexRow()`)
- Package-level functions work well
- Can add later without breaking changes

### 5. Integration Tests

**Review Concern:** "Add tests comparing fluent and classic API for equivalence"

**Implementation:**
Created `fluent_integration_test.go` with 10 comprehensive tests:

1. `TestFluentVsClassicSimple` - Basic style equivalence
2. `TestFluentVsClassicWithChildren` - Trees with children
3. `TestFluentVsClassicHelpers` - Using helper functions
4. `TestFluentVsClassicGrid` - Grid layout equivalence
5. `TestFluentVsClassicNested` - Deeply nested structures
6. `TestFluentImmutability` - Verify no mutation
7. `TestFluentChainEquivalence` - Chaining vs stepwise
8. `TestFluentTransformEquivalence` - Transform results
9. `TestFluentFilterEquivalence` - Filter results
10. `TestFluentMapEquivalence` - Map results

**All tests verify:**
- Identical layout results (Rect positions)
- Equivalent tree structures
- Immutability guarantees
- Correct transformation behavior

**Status:** ✅ All tests passing

### 6. Documentation

**Review Concern:** "Extend docs with fluent API examples and ownership rules"

**Implementation:**
Created comprehensive `docs/fluent-api.md` (9,000+ words):

- **Design Philosophy** - Immutability, copy-on-write, method chaining
- **API Overview** - Four main categories of operations
- **Core Concepts** - Cloning, nil safety, method chaining
- **Navigation & Querying** - Complete guide with examples
- **Immutable Modifications** - All With* methods
- **Parent Navigation** - NodeContext wrapper
- **Transformations** - Transform, Map, Filter, Fold
- **Practical Examples** - Real-world use cases
- **Integration** - Mixing fluent and classic APIs
- **Performance** - Allocation overhead, optimization tips
- **Best Practices** - When to use each approach

**Also Updated:**
- README.md with fluent API section
- Examples showing both API styles

## Design Trade-offs

### Immutability vs Performance

**Trade-off:**
Immutable operations clone nodes, which has allocation overhead.

**Mitigation:**
1. **Copy-on-write** - Children arrays shared until modified
2. **Shallow clones** - Most operations only copy Node struct (~48 bytes)
3. **Early termination** - Find, Any, etc. stop early
4. **Documentation** - Performance section with optimization tips

**When to Use Each:**
- **Fluent (immutable):** Building UIs, creating variants, querying
- **Classic (mutable):** Hot paths, large tree initialization, mutations

### Single API vs Dual API

**Trade-off:**
Supporting both classic and fluent APIs means two ways to do everything.

**Benefits:**
1. **Gradual adoption** - Users can migrate incrementally
2. **Choice** - Use best tool for each situation
3. **Compatibility** - No breaking changes
4. **Flexibility** - Mix and match as needed

**Mitigation:**
- Clear documentation on when to use each
- Examples show both approaches
- Integration tests prove equivalence

### Direct Methods vs Wrapper Type

**Trade-off:**
Methods on `*Node` mean we can't fully enforce immutability.

**Why Chosen:**
1. **Simplicity** - No wrapper/unwrapping
2. **Integration** - Works seamlessly with existing code
3. **Familiar** - Similar to jQuery, LINQ, etc.
4. **Performance** - No allocation overhead for wrapper
5. **Compatibility** - Zero breaking changes

**Enforcement:**
- Documentation emphasizes immutability
- Examples all use immutable patterns
- Tests verify immutability
- Community education

## Future Considerations

### Potential Enhancements

#### 1. More Ergonomic Methods

Could add convenience methods:
```go
func (n *Node) WithFlexRow() *Node
func (n *Node) WithFlexColumn() *Node
func (n *Node) WithGrid(rows, cols int) *Node
```

**Decision:** Defer until user demand is clear

#### 2. MustFind Variant

Add panic-on-not-found variant:
```go
func (n *Node) MustFind(predicate func(*Node) bool) *Node {
    result := n.Find(predicate)
    if result == nil {
        panic("MustFind: no node matched predicate")
    }
    return result
}
```

**Decision:** Can add later without breaking changes

#### 3. Optimized Transform Chains

Add "transient" mode for performance-critical paths:
```go
func (n *Node) TransformMut(predicate, transform func(*Node) *Node) *Node {
    // In-place transformation for internal use
}
```

**Decision:** Wait for demonstrated need

#### 4. Generic Fold Return Type

Once Go 1.18+ is baseline, use generics for Fold:
```go
func Fold[T any](n *Node, initial T, fn func(T, *Node) T) T
```

**Decision:** Keep interface{} for now for compatibility

### Compatibility Guarantees

**Stability Promise:**
- All fluent methods follow copy-on-write semantics
- Nil safety is guaranteed
- Integration tests lock in equivalence with classic API
- No breaking changes to existing fluent API

**Additions:**
- New methods can be added without breaking changes
- New helper functions can be added
- New convenience methods can be added

**Changes:**
- Will not change method signatures (would break chains)
- Will not change nil-safety behavior
- Will not change copy-on-write semantics
- Will not change classic API

## Summary

### Core Design Philosophy

1. **Methods on *Node directly** - No wrapper type
2. **Copy-on-write immutability** - All methods clone
3. **Nil-safe operations** - Graceful nil handling
4. **Full backward compatibility** - Zero breaking changes
5. **Dual API support** - Classic and fluent coexist

### Key Decisions

| Concern | Decision | Rationale |
|---------|----------|-----------|
| Wrapper vs Direct | Direct methods on *Node | Simpler, no wrapping overhead |
| Nil handling | Return nil, don't auto-init | Consistent with Go idioms |
| FindFirst return | Return *Node (nil if not found) | Chainable, simple |
| Ergonomics | Package-level helpers | Consistent with existing API |
| Testing | Comprehensive integration tests | Verify equivalence |
| Documentation | 9,000+ word guide | Clear usage patterns |

### Review Feedback Status

✅ Zero value handling - Documented and tested
✅ Integration tests - 10 tests, all passing
✅ Documentation - Comprehensive guide created
✅ Performance - Guidance provided
✅ Best practices - Clear recommendations

### Trade-offs Accepted

- **Exported fields** - For backward compatibility
- **No enforced immutability** - Documentation + convention
- **Dual API complexity** - Flexibility outweighs cost
- **Clone overhead** - Performance acceptable for use case

The fluent API successfully provides immutable, composable tree operations while maintaining 100% backward compatibility and giving users the choice between mutable (classic) and immutable (fluent) styles.
