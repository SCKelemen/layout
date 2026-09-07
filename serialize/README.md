# Serialize Package

The `serialize` package provides JSON serialization and deserialization for layout trees. This is useful for:

- **Debugging**: Inspect layout trees to understand structure and computed positions
- **Persistence**: Save and load layout configurations
- **Testing**: Compare layout trees across different runs
- **Documentation**: Generate examples and visualizations

## Usage

### Serialize to JSON

```go
import (
    "github.com/SCKelemen/layout"
    "github.com/SCKelemen/layout/serialize"
)

// Create a layout tree
root := layout.VStack(
    layout.Fixed(100, 50),
    layout.Fixed(100, 50),
)

// Perform layout
constraints := layout.Loose(200, layout.Unbounded)
ctx := layout.NewLayoutContext(800, 600, 16)
layout.Layout(root, constraints, ctx)

// Serialize to JSON
jsonBytes, err := serialize.ToJSON(root)
if err != nil {
    log.Fatal(err)
}

// Print formatted JSON
fmt.Println(string(jsonBytes))
```

### Deserialize from JSON

```go
// Deserialize from JSON bytes
deserialized, err := serialize.FromJSON(jsonBytes)
if err != nil {
    log.Fatal(err)
}

// Use the deserialized node
fmt.Printf("Node width: %.2f\n", deserialized.Rect.Width)
```

## JSON Structure

The serialized JSON includes:

- **Style**: All layout properties (display, flex, grid, sizing, positioning,
  writing mode, container queries, text style, etc.)
- **Text** / **Baseline**: Text content and baseline of the node
- **Rect**: Computed position and size (after layout); omitted when zero
- **Children**: Recursive child nodes

`TextLayout` is derived output of the layout pass and is not serialized.

### Example JSON Output

```json
{
  "style": {
    "display": "flex",
    "flexDirection": "column",
    "width": "200px",
    "padding": {
      "top": "10px",
      "right": "10px",
      "bottom": "10px",
      "left": "10px"
    }
  },
  "rect": {
    "x": 0,
    "y": 0,
    "width": 200,
    "height": 100
  },
  "children": [
    {
      "style": {
        "width": "100px",
        "height": "2em"
      },
      "rect": {
        "x": 10,
        "y": 10,
        "width": 100,
        "height": 32
      }
    }
  ]
}
```

## Notes

- **Lengths**: Every `layout.Length` is written as a CSS-like string that
  keeps its unit: `"10px"`, `"2em"`, `"1.5rem"`, `"50vw"`, `"10cqw"`, ...
  The zero-value `layout.Length{}` (no unit, meaning "not set") is omitted,
  while an explicit `Px(0)` is written as `"0px"`. Both
  `layout.PxUnbounded` and `layout.UnboundedLength()` are written as
  `"unbounded"` and read back as `layout.PxUnbounded`.
  For backward compatibility a bare number (`"width": 200`) is still
  accepted on input and interpreted as pixels. The previous encoder wrote
  unbounded lengths (the `maxSize` of `fr` and `auto` tracks) as the bare
  number `1.7976931348623157e+308` (`math.MaxFloat64`); that sentinel is
  read back as unbounded (`layout.PxUnbounded`), exactly like
  `"unbounded"`. The same value is accepted in `rect` fields, where it
  stays `layout.Unbounded`.
- **Enum Values**: Enums are serialized as CSS keywords (e.g. `"flex"`,
  `"grid"`, `"inline-text"`, `"none"`, `"row-reverse"`, `"space-evenly"`).
  Unknown keywords are rejected with an error rather than silently mapped to
  a default. `fontWeight` is numeric (1-1000) and `textDecoration` is the
  bitmask value.
- **Auto Values**: `"-1px"` represents "auto" for width/height and
  positioning properties; `-1` grid line indices mean auto placement.
- **Zero Values**: Default values are omitted from the output, including
  default alignment (`stretch`), identity transforms, and empty
  padding/margin/border.
- **Transform**: Non-identity transform matrices are serialized with all 6
  components (a, b, c, d, e, f).
- **Validation**: `FromJSON`/`FromYAML` treat input as untrusted. They reject
  NaN or infinite numbers, magnitudes above `MaxNumericValue` (1e12, except
  the unbounded sentinel described above), unknown enum keywords,
  fractional or overflowing integers, trees deeper than `MaxTreeDepth`
  (1024), and nodes with more than `MaxChildren` (65536) children. Limit
  violations wrap `ErrLimitExceeded`.
- **Encoding errors**: `ToJSON`/`ToYAML` return an error (they never did
  before) for trees that cannot be represented faithfully: NaN or infinite
  numbers, magnitudes above `MaxNumericValue`, unknown enum values, and
  trees deeper than `MaxTreeDepth`, which is how a cyclic tree fails instead
  of overflowing the stack. A tree produced by `layout.Layout` from valid
  styles never triggers these; in particular `layout.Unbounded` in `rect`
  fields is accepted.

## YAML Support

YAML support is available as an optional feature. YAML output uses the same
camelCase keys and the same length strings as the JSON output. To use it:

1. Install the YAML library:
   ```bash
   go get gopkg.in/yaml.v3
   ```

2. Use the YAML functions:
   ```go
   // Serialize to YAML
   yamlBytes, err := serialize.ToYAML(root)
   
   // Deserialize from YAML
   deserialized, err := serialize.FromYAML(yamlBytes)
   ```

To disable YAML support (e.g., to avoid the dependency), build with:
```bash
go build -tags no_yaml
```

