# WPT JSON Test Format Schema

**Version**: 1.0.0
**Purpose**: Universal test format for non-browser implementations of web specifications

## Overview

This schema defines a JSON format for representing Web Platform Tests in a way that can be consumed by any implementation, regardless of language or platform. The format captures browser-rendered expected values while providing a declarative structure that implementations can use without needing a full browser.

## Design Goals

1. **Multi-Browser Support**: Store expected results from multiple browsers (Chrome, Firefox, Safari)
2. **Implementation Agnostic**: Usable from Go, Rust, C, Python, JavaScript, etc.
3. **Categorized & Searchable**: Tests organized by categories, tags, and properties
4. **Focused Testing**: Only include relevant properties (layout tests don't include color)
5. **Stable Results**: Deterministic, reproducible expected values
6. **Declarative Structure**: Tree structure can be constructed programmatically
7. **Version Tracked**: Tests include version and timestamp for updates

## Use Cases

### Primary: Layout Engine Testing
Test CSS layout implementations (Flexbox, Grid, Box Model) without a browser:
```go
// Go implementation
test := LoadTest("flexbox-justify-content-001.json")
root := BuildLayout(test.Layout)
Layout(root, test.Constraints)
VerifyResults(root, test.Results["chrome"])
```

### Color Libraries
Test CSS Color Module implementations:
```python
# Python implementation
test = load_test("color-rgb-to-hsl-001.json")
result = convert_rgb_to_hsl(test.input)
assert_close(result, test.results["chrome"])
```

### Web Workers
Test Worker API implementations without a browser:
```rust
// Rust implementation
let test = load_test("worker-postmessage-001.json");
let worker = Worker::new(&test.script);
assert_eq!(worker.run(), test.expected);
```

## Schema Definition

### Root Structure

```typescript
interface WPTTest {
  // Metadata
  version: string;              // Schema version, e.g., "1.0.0"
  id: string;                   // Unique test ID
  title: string;                // Human-readable test title
  description?: string;         // What this test validates
  source: {
    url: string;                // Original WPT test URL
    file: string;               // Original test filename
    commit?: string;            // WPT commit hash
  };
  generated: {
    timestamp: string;          // ISO 8601 timestamp
    tool: string;               // Generator name/version
  };

  // Classification
  spec: {
    name: string;               // E.g., "CSS Flexbox Level 1"
    section: string;            // E.g., "9.2 Flex Container"
    url: string;                // Spec URL
  };
  categories: string[];         // E.g., ["layout", "flexbox", "alignment"]
  tags: string[];               // E.g., ["justify-content", "space-between"]
  properties: string[];         // CSS properties tested, e.g., ["justify-content"]

  // Test Structure
  layout: LayoutTree;           // Declarative layout structure
  constraints: Constraints;     // Layout constraints

  // Expected Results (Multi-Browser)
  results: {
    [browser: string]: BrowserResult;
  };

  // Optional
  notes?: string[];             // Implementation notes
  knownIssues?: string[];       // Known browser bugs/quirks
}
```

### Layout Tree

```typescript
interface LayoutTree {
  type: "container" | "text" | "block";
  id?: string;                  // Optional identifier

  // Style Properties (Only Relevant Ones)
  style: {
    // Display & Positioning
    display?: "flex" | "grid" | "block" | "inline-block";
    position?: "static" | "relative" | "absolute" | "fixed";

    // Flexbox
    flexDirection?: "row" | "column" | "row-reverse" | "column-reverse";
    flexWrap?: "nowrap" | "wrap" | "wrap-reverse";
    justifyContent?: "flex-start" | "flex-end" | "center" | "space-between" | "space-around" | "space-evenly";
    alignItems?: "flex-start" | "flex-end" | "center" | "baseline" | "stretch";
    alignContent?: "flex-start" | "flex-end" | "center" | "space-between" | "space-around" | "stretch";

    // Grid (when applicable)
    gridTemplateColumns?: string;
    gridTemplateRows?: string;
    gridGap?: number;

    // Box Model
    width?: number | "auto";
    height?: number | "auto";
    minWidth?: number;
    minHeight?: number;
    maxWidth?: number;
    maxHeight?: number;

    // Spacing
    margin?: Spacing;
    padding?: Spacing;
    border?: Spacing;

    // Flex Item Properties
    flexGrow?: number;
    flexShrink?: number;
    flexBasis?: number | "auto";
    alignSelf?: "auto" | "flex-start" | "flex-end" | "center" | "baseline" | "stretch";
  };

  // For text nodes
  text?: string;
  textStyle?: {
    fontSize?: number;
    fontFamily?: string;
    whiteSpace?: "normal" | "nowrap" | "pre" | "pre-wrap" | "pre-line";
    textAlign?: "left" | "right" | "center" | "justify";
  };

  // Children
  children?: LayoutTree[];
}

interface Spacing {
  top: number;
  right: number;
  bottom: number;
  left: number;
}

interface Constraints {
  type: "loose" | "tight" | "bounded";
  width: number;
  height: number;
}
```

### Browser Results

```typescript
interface BrowserResult {
  browser: {
    name: "Chrome" | "Firefox" | "Safari" | "Edge";
    version: string;
    engine: string;             // E.g., "Blink 121.0"
  };
  rendered: {
    timestamp: string;
    viewport: { width: number; height: number };
  };

  // Expected Values
  elements: ElementResult[];

  // Tolerances
  tolerance?: {
    position: number;           // Default: 1.0px
    size: number;               // Default: 1.0px
    numeric: number;            // Default: 0.01
  };
}

interface ElementResult {
  id?: string;                  // Matches layout tree ID
  path: string;                 // JSONPath to element, e.g., "root.children[0]"

  expected: {
    // Position & Size
    x?: number;
    y?: number;
    width?: number;
    height?: number;

    // Computed Values (only if relevant)
    computedDisplay?: string;
    computedPosition?: string;

    // Flex-specific (only for flex items)
    flexBaseSize?: number;
    mainSize?: number;
    crossSize?: number;

    // Text-specific (only for text nodes)
    lineCount?: number;
    glyphCount?: number;

    // Any other measured properties
    [key: string]: any;
  };
}
```

## Example: Flexbox Test

```json
{
  "version": "1.0.0",
  "id": "css-flexbox-justify-content-space-between-001",
  "title": "Flexbox justify-content: space-between",
  "description": "Tests that justify-content: space-between correctly distributes space between flex items",
  "source": {
    "url": "https://github.com/web-platform-tests/wpt/blob/master/css/css-flexbox/...",
    "file": "justify-content-space-between-001.html",
    "commit": "abc123def456"
  },
  "generated": {
    "timestamp": "2025-12-13T10:45:00Z",
    "tool": "wpt-test-generator v1.0.0"
  },

  "spec": {
    "name": "CSS Flexbox Level 1",
    "section": "8.2 Axis Alignment: justify-content",
    "url": "https://www.w3.org/TR/css-flexbox-1/#justify-content-property"
  },
  "categories": ["layout", "flexbox", "alignment"],
  "tags": ["justify-content", "space-between", "main-axis"],
  "properties": ["justify-content", "flex-direction"],

  "layout": {
    "type": "container",
    "id": "root",
    "style": {
      "display": "flex",
      "flexDirection": "row",
      "justifyContent": "space-between",
      "width": 600,
      "height": 100
    },
    "children": [
      {
        "type": "block",
        "id": "item1",
        "style": { "width": 100, "height": 50 }
      },
      {
        "type": "block",
        "id": "item2",
        "style": { "width": 100, "height": 50 }
      },
      {
        "type": "block",
        "id": "item3",
        "style": { "width": 100, "height": 50 }
      }
    ]
  },

  "constraints": {
    "type": "loose",
    "width": 800,
    "height": 600
  },

  "results": {
    "chrome": {
      "browser": {
        "name": "Chrome",
        "version": "121.0.6167.85",
        "engine": "Blink"
      },
      "rendered": {
        "timestamp": "2025-12-13T10:45:00Z",
        "viewport": { "width": 800, "height": 600 }
      },
      "elements": [
        {
          "id": "root",
          "path": "root",
          "expected": {
            "x": 0,
            "y": 0,
            "width": 600,
            "height": 100
          }
        },
        {
          "id": "item1",
          "path": "root.children[0]",
          "expected": {
            "x": 0,
            "y": 0,
            "width": 100,
            "height": 50
          }
        },
        {
          "id": "item2",
          "path": "root.children[1]",
          "expected": {
            "x": 250,
            "y": 0,
            "width": 100,
            "height": 50
          }
        },
        {
          "id": "item3",
          "path": "root.children[2]",
          "expected": {
            "x": 500,
            "y": 0,
            "width": 100,
            "height": 50
          }
        }
      ],
      "tolerance": {
        "position": 1.0,
        "size": 1.0
      }
    },
    "firefox": {
      "browser": {
        "name": "Firefox",
        "version": "122.0",
        "engine": "Gecko"
      },
      "rendered": {
        "timestamp": "2025-12-13T10:46:00Z",
        "viewport": { "width": 800, "height": 600 }
      },
      "elements": [
        {
          "id": "root",
          "path": "root",
          "expected": {
            "x": 0,
            "y": 0,
            "width": 600,
            "height": 100
          }
        },
        {
          "id": "item1",
          "path": "root.children[0]",
          "expected": {
            "x": 0,
            "y": 0,
            "width": 100,
            "height": 50
          }
        },
        {
          "id": "item2",
          "path": "root.children[1]",
          "expected": {
            "x": 250,
            "y": 0,
            "width": 100,
            "height": 50
          }
        },
        {
          "id": "item3",
          "path": "root.children[2]",
          "expected": {
            "x": 500,
            "y": 0,
            "width": 100,
            "height": 50
          }
        }
      ],
      "tolerance": {
        "position": 1.0,
        "size": 1.0
      }
    }
  },

  "notes": [
    "All major browsers agree on positioning",
    "Test validates main-axis distribution"
  ]
}
```

## Example: Color Conversion Test

```json
{
  "version": "1.0.0",
  "id": "css-color-rgb-to-hsl-001",
  "title": "RGB to HSL color conversion",
  "description": "Tests conversion from RGB color space to HSL",

  "spec": {
    "name": "CSS Color Module Level 4",
    "section": "7.2 HSL Colors",
    "url": "https://www.w3.org/TR/css-color-4/#the-hsl-notation"
  },
  "categories": ["color", "conversion"],
  "tags": ["rgb", "hsl", "color-space"],
  "properties": ["color"],

  "input": {
    "colorSpace": "srgb",
    "values": [0.5, 0.25, 0.75]
  },

  "results": {
    "chrome": {
      "browser": {
        "name": "Chrome",
        "version": "121.0.6167.85",
        "engine": "Blink"
      },
      "expected": {
        "colorSpace": "hsl",
        "values": [270, 0.5, 0.5],
        "string": "hsl(270deg 50% 50%)"
      },
      "tolerance": {
        "numeric": 0.01
      }
    }
  }
}
```

## Test Organization

Tests should be organized hierarchically:

```
wpt-json-tests/
├── schema/
│   └── v1.0.0/
│       ├── schema.json          # JSON Schema definition
│       └── README.md             # This file
├── layout/
│   ├── flexbox/
│   │   ├── alignment/
│   │   │   ├── justify-content-001.json
│   │   │   ├── justify-content-002.json
│   │   │   └── align-items-001.json
│   │   ├── sizing/
│   │   └── wrapping/
│   ├── grid/
│   │   ├── template/
│   │   ├── gap/
│   │   └── placement/
│   └── block/
├── color/
│   ├── conversion/
│   ├── parsing/
│   └── interpolation/
└── workers/
    ├── postmessage/
    └── importscripts/
```

## Metadata for Discovery

Each directory should have a `manifest.json`:

```json
{
  "category": "layout/flexbox/alignment",
  "description": "Tests for flexbox alignment properties",
  "properties": ["justify-content", "align-items", "align-content", "align-self"],
  "specSection": "8. Alignment",
  "testCount": 45,
  "browsers": ["chrome", "firefox", "safari"],
  "updated": "2025-12-13T10:00:00Z"
}
```

## Implementation Guidelines

### For Test Consumers (Go, Rust, etc.)

1. **Load JSON**: Parse test file
2. **Build Structure**: Construct layout tree from `test.layout`
3. **Apply Constraints**: Use `test.constraints`
4. **Run Layout**: Execute your layout algorithm
5. **Validate**: Compare results against `test.results["chrome"]` (or preferred browser)
6. **Check Tolerance**: Use provided tolerance values

### For Test Generators (Puppeteer, Playwright, etc.)

1. **Render HTML**: Load WPT test in browser
2. **Extract Styles**: Get computed styles for relevant properties only
3. **Measure Layout**: Capture positions and sizes
4. **Normalize**: Remove browser-specific quirks (like body margins)
5. **Categorize**: Add appropriate metadata
6. **Output**: Write to schema-compliant JSON

## Version History

- **v1.0.0** (2025-12-13): Initial schema definition

## Future Considerations

### Potential Additions

1. **Animations/Transitions**: Time-series data
2. **Paint Properties**: When testing rendering, not just layout
3. **Accessibility**: ARIA properties and tree structure
4. **Scripting**: For Worker/API tests with before/after states
5. **References**: Links to reference images for visual tests
6. **Differences**: Document known browser differences with reasons

### Extensibility

The schema supports custom properties in results:

```json
{
  "results": {
    "chrome": {
      "expected": {
        "x": 100,
        "y": 50,
        "_custom_property": "any value",
        "_implementation_specific": { "debug": "info" }
      }
    }
  }
}
```

Properties prefixed with `_` are implementation-specific and should be ignored by other consumers.

## Contributing

When adding tests:

1. ✅ Include only relevant properties (no color in layout tests)
2. ✅ Add proper categorization (categories, tags, properties)
3. ✅ Include at least 2 browser results for validation
4. ✅ Document any known issues or browser quirks
5. ✅ Use stable, deterministic test cases
6. ✅ Follow the directory organization structure
7. ✅ Update manifest.json in the category directory

## License

Schema and test format are released into the public domain (CC0), enabling maximum reusability across projects and languages.
