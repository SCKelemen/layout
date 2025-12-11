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
layout.Layout(root, constraints)

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

- **Style**: All layout properties (display, flex, grid, sizing, positioning, etc.)
- **Rect**: Computed position and size (after layout)
- **Children**: Recursive child nodes

### Example JSON Output

```json
{
  "style": {
    "display": "flex",
    "flexDirection": "column",
    "width": 200,
    "height": -1
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
        "width": 100,
        "height": 50
      },
      "rect": {
        "x": 0,
        "y": 0,
        "width": 100,
        "height": 50
      }
    }
  ]
}
```

## Notes

- **Enum Values**: Enums are serialized as strings (e.g., `"flex"`, `"grid"`, `"row"`)
- **Auto Values**: `-1` represents "auto" for width/height and positioning properties
- **Zero Values**: Zero values are omitted from JSON output (use `omitempty` tags)
- **Transform**: Transform matrices are serialized with all 6 components (a, b, c, d, e, f)

## YAML Support

YAML support is available as an optional feature. To use it:

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

