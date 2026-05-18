# CSS Container Query Units (L4)

This document describes the layout library's support for the six CSS L4
container query units and the three supporting style properties.

- Specification (container queries):
  [CSS Containment Module Level 3 — Container Queries](https://www.w3.org/TR/css-contain-3/#container-queries)
- Specification (units):
  [CSS Values and Units Module Level 4 — Container-relative lengths](https://www.w3.org/TR/css-values-4/#container-relative-lengths)

Although container queries themselves are formally in CSS Containment Level 3,
the `cq*` length units are defined in CSS Values and Units Level 4 — hence
"L4" in this feature's name.

## Units

| Unit    | Meaning                                                   | Constructor |
| ------- | --------------------------------------------------------- | ----------- |
| `cqw`   | 1% of the query container's inline-size (width in horizontal writing modes) | `layout.Cqw(50)`   |
| `cqh`   | 1% of the query container's block-size  (height in horizontal writing modes) | `layout.Cqh(50)`   |
| `cqi`   | 1% of the query container's inline-size (writing-mode aware) | `layout.Cqi(50)`   |
| `cqb`   | 1% of the query container's block-size  (writing-mode aware) | `layout.Cqb(50)`   |
| `cqmin` | The smaller of `cqi` and `cqb`                            | `layout.Cqmin(50)` |
| `cqmax` | The larger of `cqi` and `cqb`                             | `layout.Cqmax(50)` |

All six units accept negative and fractional values. They can also be parsed
from strings, case-insensitively:

```go
l, _ := layout.ParseLength("50cqw") // Length{Value: 50, Unit: layout.CQWUnit}
```

## Style properties

The three new properties live on `layout.Style`:

```go
type Style struct {
    // ...
    ContainerType ContainerType // normal | size | inline-size
    ContainerName []string      // one or more identifiers
}
```

### `container-type`

Sets whether the element establishes a query container, and which axes are
queryable.

| Value         | Constant                       | Meaning                                              |
| ------------- | ------------------------------ | ---------------------------------------------------- |
| `normal`      | `layout.ContainerTypeNormal`      | Not a query container (default, zero value).         |
| `size`        | `layout.ContainerTypeSize`        | Queryable on both inline and block axes.             |
| `inline-size` | `layout.ContainerTypeInlineSize` | Queryable on the inline axis only.                   |

Parse from a string:

```go
ct, _ := layout.ParseContainerType("inline-size") // layout.ContainerTypeInlineSize
```

### `container-name`

One or more identifiers naming this container. Identifiers may contain
ASCII letters, digits, `-`, and `_`, and must not start with a digit. The
reserved keywords `none`, `normal`, `inherit`, `initial`, `unset`,
`revert`, `revert-layer`, and `default` are rejected (the keyword `none`
in the shorthand below is a valid reset value, not a name).

```go
names, _ := layout.ParseContainerName("card sidebar") // []string{"card", "sidebar"}
```

> The `container-name` field exists so callers can author and round-trip
> a complete container declaration today. Name-based dispatch (`@container
> <name> (...)`) is **not implemented** — see the out-of-scope section
> below.

### `container` (shorthand)

`container: <container-name> [ / <container-type> ]?`

```go
names, ctype, _ := layout.ParseContainer("card / size")
// names == []string{"card"}, ctype == layout.ContainerTypeSize
```

Accepted forms:

| Input               | Names           | Type                            |
| ------------------- | --------------- | ------------------------------- |
| `"foo"`             | `["foo"]`       | `normal`                        |
| `"foo bar"`         | `["foo","bar"]` | `normal`                        |
| `"foo / size"`      | `["foo"]`       | `size`                          |
| `"size"`            | `nil`           | `size`                          |
| `"inline-size"`     | `nil`           | `inline-size`                   |
| `"none / size"`     | `nil`           | `size` (`none` clears names)    |
| `""`                | `nil`           | `normal`                        |

## Resolution algorithm

For every `cq*` unit, the resolver walks the ancestor chain from the
element being resolved looking for the nearest *query container*:

1. Start from the parent of the current `*NodeContext`.
2. Take the first ancestor whose `Style.ContainerType` is not
   `ContainerTypeNormal`.
3. Inline axis (`cqw`, `cqi`):
   - `ContainerTypeSize` and `ContainerTypeInlineSize` both qualify.
   - Resolve against the ancestor's measured inline-size (its `Rect.Width`
     in horizontal writing modes, `Rect.Height` in vertical writing modes).
4. Block axis (`cqh`, `cqb`):
   - Only `ContainerTypeSize` qualifies. An `inline-size`-only ancestor is
     *skipped*; the walk continues further up.
   - Resolve against the ancestor's measured block-size.
5. `cqmin`/`cqmax` evaluate the inline and block axes independently and
   return the min/max of the two pixel results.
6. If the walk reaches the root without a matching ancestor, the unit
   falls back to the viewport on the corresponding physical axis. With no
   container ancestor, `50cqw` is equivalent to `50vw`, `50cqh` to `50vh`,
   `cqmin` to `vmin`, and `cqmax` to `vmax`.

The public entry point for container-aware resolution is:

```go
func ResolveLengthInContext(l Length, ctx *LayoutContext, currentFontSize float64, nctx *NodeContext) float64
```

`ResolveLength` (the existing API, which does not take a `*NodeContext`)
is preserved and unconditionally applies the viewport fallback for `cq*`
units. This keeps the legacy resolver useful for non-tree contexts
(e.g. computing single-value lengths) while delegating tree-aware
resolution to the new variant.

### Resolution timing

The resolver reads a container ancestor's `Rect` directly. Because layout
proceeds parent-first, by the time a child's lengths are resolved the
relevant parent `Rect` has already been measured for the current pass.

There is **no fixed-point iteration**: if a child's `cq*`-derived size
later influences a parent's measured size in a subsequent pass, the new
size takes effect "last-wins" rather than triggering a re-resolution.
This matches the conservative behavior of the underlying layout engine
and avoids the cost of cycle detection.

## Out of scope

The following pieces of the L3/L4 container-queries feature are **not**
implemented here:

- `@container` at-rules (style conditionals based on container size or
  state). These are a substantial separate feature.
- `@container <name> (...)` name-scoped queries.

The `container-name` property is still accepted so that an end-to-end
declaration can be parsed and round-tripped — once the at-rule lands the
data is already in place.

## Example

```go
card := &layout.Node{
    Style: layout.Style{
        Width:         layout.Px(320),
        Height:        layout.Px(200),
        ContainerType: layout.ContainerTypeSize,
        ContainerName: []string{"card"},
    },
}
card.Children = []*layout.Node{
    {Style: layout.Style{Width: layout.Cqw(50), Height: layout.Cqh(50)}},
}
```

After layout, the child measures 160 × 100 pixels (50% of 320 × 200).
