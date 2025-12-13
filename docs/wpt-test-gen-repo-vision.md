# WPT Test Generator - Standalone Repository Vision

## Repository

**Location**: https://github.com/SCKelemen/wpt-test-gen.git

## Purpose

Create a standalone repository that generates JSON test files from Web Platform Tests (WPT), making them usable by any implementation regardless of language or platform.

## Problem Statement

### Current Situation

Web Platform Tests (WPT) are designed for browsers:
- Require a full browser environment to run
- Tests are HTML/JavaScript that execute in browser context
- Not directly usable by non-browser implementations

### The Gap

Many projects implement web specifications without building full browsers:
- **Layout engines** in Go, Rust, C (like this project)
- **CSS color libraries** in Python, JavaScript, Ruby
- **Worker APIs** in server-side runtimes
- **Grid/Flexbox implementations** for native UI frameworks

These projects need conformance tests but can't easily use WPT.

## Solution

### Universal JSON Test Format

Generate JSON files from WPT tests with:
1. **Multi-browser results**: Chrome, Firefox, Safari expected values
2. **Declarative structure**: Build layouts/tests programmatically from JSON
3. **Proper categorization**: Find tests by spec section, property, category
4. **Spec-aligned properties**: Use exact names from specifications
5. **Implementation-agnostic**: Usable from any language

### Example Test Structure

```json
{
  "version": "1.0.0",
  "id": "css-flexbox-justify-content-space-between-001",
  "title": "Flexbox justify-content: space-between",

  "source": {
    "url": "https://github.com/web-platform-tests/wpt/blob/master/css/css-flexbox/...",
    "file": "justify-content-space-between-001.html",
    "commit": "abc123"
  },

  "spec": {
    "name": "CSS Flexbox Level 1",
    "section": "8.2 Axis Alignment",
    "url": "https://www.w3.org/TR/css-flexbox-1/#justify-content-property"
  },

  "categories": ["layout", "flexbox", "alignment"],
  "properties": ["justify-content"],

  "layout": {
    "type": "container",
    "style": {
      "display": "flex",
      "justify-content": "space-between",
      "width": 600,
      "height": 100
    },
    "children": [...]
  },

  "constraints": {
    "type": "loose",
    "width": 800,
    "height": 600
  },

  "results": {
    "chrome": {
      "browser": { "name": "Chrome", "version": "121.0", "engine": "Blink" },
      "elements": [
        { "id": "item1", "path": "root.children[0]",
          "expected": { "x": 0, "y": 0, "width": 100, "height": 50 } },
        { "id": "item2", "path": "root.children[1]",
          "expected": { "x": 250, "y": 0, "width": 100, "height": 50 } }
      ]
    },
    "firefox": {
      "browser": { "name": "Firefox", "version": "122.0", "engine": "Gecko" },
      "elements": [...]
    }
  }
}
```

## Repository Structure

```
wpt-test-gen/
├── generator/
│   ├── wpt_renderer.js          # Puppeteer-based test renderer
│   ├── package.json
│   └── README.md
├── tests/                        # Generated JSON tests
│   ├── layout/
│   │   ├── flexbox/
│   │   │   ├── alignment/
│   │   │   │   ├── justify-content-001.json
│   │   │   │   ├── justify-content-002.json
│   │   │   │   └── manifest.json
│   │   │   ├── sizing/
│   │   │   └── wrapping/
│   │   └── grid/
│   ├── color/
│   └── workers/
├── schema/
│   ├── v1.0.0/
│   │   ├── schema.json          # JSON Schema definition
│   │   └── README.md            # Schema documentation
│   └── CHANGELOG.md
├── examples/
│   ├── go/                      # Example Go test loader
│   ├── rust/                    # Example Rust test loader
│   ├── python/                  # Example Python test loader
│   └── c/                       # Example C test loader
├── .github/
│   └── workflows/
│       └── generate-tests.yml   # CI workflow
└── README.md
```

## CI Workflow

### Automated Test Generation

```yaml
name: Generate WPT JSON Tests

on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly on Sunday
  workflow_dispatch:      # Manual trigger

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout wpt-test-gen
        uses: actions/checkout@v3

      - name: Checkout WPT
        uses: actions/checkout@v3
        with:
          repository: web-platform-tests/wpt
          path: wpt-source

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '20'

      - name: Install dependencies
        run: |
          cd generator
          npm install
          npx puppeteer browsers install chrome firefox

      - name: Generate tests
        run: |
          node generator/wpt_renderer.js \
            --batch wpt-source/css/css-flexbox tests/layout/flexbox \
            --browsers chrome,firefox

          node generator/wpt_renderer.js \
            --batch wpt-source/css/css-grid tests/layout/grid \
            --browsers chrome,firefox

      - name: Create manifests
        run: node generator/create_manifests.js tests/

      - name: Commit and push
        run: |
          git config user.name "WPT Test Generator"
          git config user.email "bot@wpt-test-gen"
          git add tests/
          git commit -m "Update tests from WPT $(date +%Y-%m-%d)"
          git push
```

### Release Process

1. **Weekly updates**: CI runs every Sunday, pulls latest WPT, regenerates tests
2. **Version tags**: Tag releases when schema changes or major WPT updates
3. **Changelog**: Auto-generate changelog showing what tests changed
4. **GitHub Releases**: Publish test bundles as release artifacts

## Consuming Tests

### For Downstream Projects

Projects consume tests in three ways:

#### 1. Git Submodule (Development)

```bash
# Add as submodule
git submodule add https://github.com/SCKelemen/wpt-test-gen.git tests/wpt-gen
git submodule update --init --recursive

# Update to latest
cd tests/wpt-gen
git pull origin main
```

#### 2. Download Release Archive (CI)

```yaml
# In your project's CI
- name: Download WPT JSON Tests
  run: |
    curl -L https://github.com/SCKelemen/wpt-test-gen/archive/v1.0.0.tar.gz \
      | tar xz
```

#### 3. Direct File Access (Single Tests)

```bash
# Download specific test category
curl -O https://raw.githubusercontent.com/SCKelemen/wpt-test-gen/main/tests/layout/flexbox/alignment/justify-content-001.json
```

### Example: Using Tests in Go

```go
package layout_test

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestWPTFlexboxAlignment(t *testing.T) {
    // Walk through all flexbox alignment tests
    testDir := "tests/wpt-gen/tests/layout/flexbox/alignment"

    files, _ := filepath.Glob(filepath.Join(testDir, "*.json"))

    for _, file := range files {
        t.Run(filepath.Base(file), func(t *testing.T) {
            // Load test
            test := loadWPTTest(t, file)

            // Build layout from declarative structure
            root := buildLayout(test.Layout)

            // Run layout algorithm
            layout.LayoutFlexbox(root, test.GetConstraints())

            // Validate against Chrome results
            chrome := test.Results["chrome"]
            for _, expected := range chrome.Elements {
                actual := findNodeByPath(root, expected.Path)
                assertClose(t, actual.Rect.X, expected.Expected["x"], chrome.Tolerance.Position)
                assertClose(t, actual.Rect.Y, expected.Expected["y"], chrome.Tolerance.Position)
            }
        })
    }
}
```

### Example: Using Tests in Python

```python
import json
import glob

def test_wpt_color_conversions():
    """Test color conversions using WPT JSON tests"""
    test_dir = "tests/wpt-gen/tests/color/conversion"

    for test_file in glob.glob(f"{test_dir}/*.json"):
        with open(test_file) as f:
            test = json.load(f)

        # Get input from test
        input_color = Color(test["input"]["colorSpace"],
                           test["input"]["values"])

        # Convert
        result = convert_color(input_color, "hsl")

        # Validate against Chrome
        chrome = test["results"]["chrome"]
        expected = chrome["expected"]
        tolerance = chrome.get("tolerance", {}).get("numeric", 0.01)

        assert_close(result.values, expected["values"], tolerance)
```

## Benefits

### For This Project (Layout Engine)

1. **Continuous conformance**: CI runs latest tests automatically
2. **Multi-browser validation**: Compare against Chrome, Firefox, Safari
3. **Find regressions**: Test failures show when we break something
4. **Track progress**: See which WPT tests pass/fail

### For Other Projects

1. **Ready-made test suite**: No need to write tests from scratch
2. **Language-agnostic**: Works in Go, Rust, C, Python, JavaScript
3. **Spec compliance**: Tests derived from official WPT
4. **Easy discovery**: Browse by category, property, spec section

### For the Ecosystem

1. **Shared conformance**: Multiple implementations can compare results
2. **Identify spec issues**: Cross-browser differences reveal spec ambiguities
3. **Drive improvements**: Test results feed back to specs and browsers
4. **Lower barriers**: Makes it easier to implement web specs outside browsers

## Test Categories

### Layout Tests

- **Flexbox**: alignment, wrapping, sizing, ordering, gaps
- **Grid**: template, placement, gaps, auto-flow
- **Block**: margin collapse, floats, clearing
- **Box Model**: padding, borders, width/height

### Color Tests

- **Conversion**: RGB ↔ HSL ↔ Lab ↔ LCH
- **Parsing**: Parse CSS color strings
- **Interpolation**: Color mixing and gradients
- **Gamut mapping**: Out-of-gamut color handling

### Future Categories

- **Worker API**: postMessage, importScripts, shared workers
- **Web Animations**: Timing, easing functions
- **Transforms**: Matrix calculations, interpolation

## Schema Versioning

### Semantic Versioning

- **Major**: Breaking schema changes (1.0.0 → 2.0.0)
- **Minor**: Backward-compatible additions (1.0.0 → 1.1.0)
- **Patch**: Bug fixes, test updates (1.0.0 → 1.0.1)

### Migration Path

When schema changes:
1. Publish new schema version (e.g., v2.0.0)
2. Generate tests in both old and new formats for 6 months
3. Deprecate old format with migration guide
4. Remove old format after deprecation period

## Discovery and Documentation

### Manifest Files

Each test directory has a `manifest.json`:

```json
{
  "category": "layout/flexbox/alignment",
  "description": "Tests for flexbox alignment properties",
  "properties": ["justify-content", "align-items", "align-content"],
  "specSection": "CSS Flexbox Level 1, Section 8",
  "specUrl": "https://www.w3.org/TR/css-flexbox-1/#alignment",
  "testCount": 45,
  "browsers": ["chrome", "firefox", "safari"],
  "updated": "2025-12-13T10:00:00Z"
}
```

### Search API (Future)

```bash
# Find all tests for a specific property
curl "https://wpt-test-gen.github.io/api/search?property=justify-content"

# Find tests by category
curl "https://wpt-test-gen.github.io/api/search?category=flexbox"

# Find tests for a spec section
curl "https://wpt-test-gen.github.io/api/search?spec=css-flexbox-1&section=8.2"
```

## Timeline

### Phase 1: Foundation (Current)
- ✅ JSON schema designed
- ✅ Puppeteer renderer implemented
- ✅ Multi-browser support (Chrome, Firefox)
- ✅ Go test loader
- ✅ Schema documentation

### Phase 2: Repository Setup (Week 1-2)
- [ ] Move generator to wpt-test-gen repo
- [ ] Set up CI workflow
- [ ] Generate initial test sets (Flexbox, Grid)
- [ ] Create manifest files
- [ ] Add example loaders (Go, Rust, Python)

### Phase 3: Automation (Week 3-4)
- [ ] Automated weekly test generation
- [ ] Release process and versioning
- [ ] Changelog generation
- [ ] GitHub releases with test bundles

### Phase 4: Discovery (Month 2)
- [ ] Test browser website
- [ ] Search functionality
- [ ] Usage statistics
- [ ] Community feedback integration

### Phase 5: Expansion (Ongoing)
- [ ] Add more test categories (Color, Workers, etc.)
- [ ] Safari support via Playwright
- [ ] MDN documentation links
- [ ] Spec test generation (beyond WPT)

## Success Metrics

1. **Test coverage**: >80% of relevant WPT tests converted
2. **Multi-browser**: All tests have Chrome + Firefox results
3. **Adoption**: 5+ projects using the tests
4. **Freshness**: Tests updated within 7 days of WPT changes
5. **Quality**: <1% schema validation errors

## Contributing

### Adding New Test Categories

1. Identify WPT test directory (e.g., `css/css-color`)
2. Update CI workflow to process directory
3. Create category manifest
4. Generate initial tests
5. Document expected properties

### Improving the Generator

1. Better CSS property detection
2. Handle more HTML structures
3. Extract additional metrics (e.g., computed styles)
4. Support for visual regression tests

### Writing Example Loaders

Provide reference implementations showing how to:
- Parse the JSON schema
- Build test structures
- Run tests
- Report results

## License

Tests generated from WPT maintain their original licenses. The JSON schema and generator tooling are released under MIT license for maximum reusability.

## Links

- **Repository**: https://github.com/SCKelemen/wpt-test-gen.git
- **Schema Documentation**: [wpt-test-schema.md](./wpt-test-schema.md)
- **Web Platform Tests**: https://github.com/web-platform-tests/wpt
- **CSS Specifications**: https://www.w3.org/Style/CSS/
