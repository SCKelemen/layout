# WPT Test Generator

Generate Go test files from WPT (Web Platform Test) JSON specifications with CEL assertions.

## Overview

The WPT Test Generator creates test files in two modes:

1. **Standalone Mode**: Complete, runnable tests using the `layout` library directly
2. **User-Extensible Mode**: Template tests with placeholders for custom implementations

## Installation

```bash
go build -o wpt_test_gen .
```

## Usage

### Basic Command

```bash
go run . -input <test.json> -output <test_gen.go> [options]
```

### Options

- `-input`: Input WPT test JSON file (required)
- `-output`: Output Go test file (default: derived from input filename)
- `-package`: Go package name (default: `layout_test`)
- `-standalone`: Generate standalone test using layout library (default: false)

### Examples

#### Generate User-Extensible Test

```bash
go run . -input test.json -output my_test.go -package mypackage
```

Creates a template test with placeholder functions:

```go
func buildLayoutMyTest() interface{} {
    panic("not implemented - user must provide layout implementation")
}

func runLayout(root interface{}, constraints interface{}) {
    panic("not implemented - user must provide layout implementation")
}

func createCELEnv(root interface{}) (interface{}, interface{}, error) {
    panic("not implemented - user must register custom CEL functions")
}
```

Users implement these functions to integrate with their own layout engine.

#### Generate Standalone Test

```bash
go run . -input test.json -output my_test.go -package mypackage -standalone
```

Creates a complete, runnable test:

```go
func TestMyTest(t *testing.T) {
    root := buildLayoutMyTest()
    layout.Layout(root, layout.Constraints{
        MinWidth:  0,
        MaxWidth:  800,
        MinHeight: 0,
        MaxHeight: 600,
    })

    env, err := layout.NewLayoutCELEnv(root)
    // ... evaluates all assertions
}
```

## Input Format

The generator expects WPT test JSON files with this structure:

```json
{
  "version": "1.0.0",
  "id": "test-id",
  "title": "Test Title",
  "description": "Test description",
  "layout": {
    "style": {
      "display": "flex",
      "justifyContent": "space-between",
      "width": 600,
      "height": 100
    }
  },
  "constraints": {
    "width": 800,
    "height": 600
  },
  "results": {
    "chrome": {
      "elements": [
        {
          "path": "root",
          "assertions": [
            {
              "type": "layout",
              "expression": "getX(child(root(), 0)) == 0.0",
              "message": "First child at left edge"
            }
          ]
        }
      ]
    }
  }
}
```

## Output

The generator creates Go test files with:

1. **Test function**: Named based on input filename (e.g., `TestMyLayoutTest`)
2. **Layout builder**: Function to construct the layout tree
3. **CEL assertions**: All assertions from the JSON, organized by element
4. **Evaluation loop**: Code to compile and evaluate each assertion

### Standalone Mode Output

- Uses `layout.Node` types directly
- Calls `layout.Layout()` for layout computation
- Uses `layout.NewLayoutCELEnv()` for CEL environment
- Includes complete implementation

### User-Extensible Mode Output

- Uses generic `interface{}` types
- Includes placeholder functions with panic statements
- Provides commented examples of what to implement
- Allows integration with any layout engine

## CEL Assertions

The old CEL API provides functions like:

- `root()`: Get the root node
- `child(node, index)`: Get child by index
- `getX(node)`, `getY(node)`: Get position
- `getWidth(node)`, `getHeight(node)`: Get dimensions
- `getRight(node)`, `getBottom(node)`: Get edges
- `getMarginTop(node)`, `getPaddingLeft(node)`: Get spacing

Example assertions:

```
getX(child(root(), 0)) == 0.0
getRight(child(root(), 2)) == getWidth(root())
getX(child(root(), 1)) - getRight(child(root(), 0)) == 250.0
```

## Limitations

The current generator:

- Uses the old CEL API (not the domain-structured API)
- Doesn't support `this` and `parent()` references (marked as unsupported)
- Generates TODO comments for children (manual implementation needed)
- Only generates root-level layout structure

## Future Enhancements

### CLI Tool

Create a comprehensive CLI tool:

```bash
wptest run test.json           # Run a specific test
wptest list                    # List available tests
wptest generate --lang go      # Generate tests in different languages
wptest generate --binding old  # Choose CEL API version
```

### Type-Directed Fuzzing

```bash
wptest fuzz --count 1000       # Generate random layout configurations
wptest fuzz --property "children-fit"  # Test specific properties
```

### Property-Based Testing

```bash
wptest property "children never overflow parent"
wptest property "space-between distributes evenly"
```

### Multi-Language Support

Generate tests in multiple languages:

- **Go**: Current implementation
- **Rust**: Using same JSON format
- **JavaScript**: For browser validation
- **Python**: For reference implementations

## Examples

See generated examples:

- `examples/generated_standalone_test.go` - Complete standalone test
- `tests/wpt/examples/go/generated_user_extensible_test.go` - User-extensible template

## Contributing

To extend the generator:

1. Modify the template in `main.go` (search for `testTemplate`)
2. Update struct types for JSON parsing if needed
3. Add new command-line flags for additional options
4. Rebuild and test with sample JSON files

## License

Same as the layout library.
