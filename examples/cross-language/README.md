# Cross-Language Layout Testing

This directory demonstrates how to use the `wptest eval` command from any programming language without language-specific bindings, C-level FFI, or custom code.

## Prerequisites

Build and install `wptest`:

```bash
cd /path/to/layout
go install ./cmd/wptest
```

Or add it to your PATH:

```bash
export PATH="$PATH:/path/to/layout"
```

## The Approach

`wptest eval` uses a simple stdin/stdout JSON protocol:

1. **Input (stdin)**: JSON test specification with layout tree, constraints, and CEL assertions
2. **Output (stdout)**: JSON test results with passed/failed/skipped counts

This means ANY language that can:
- Spawn a subprocess
- Write JSON to stdin
- Read JSON from stdout

...can test layouts with zero dependencies.

## Examples

### JavaScript (Node.js)

```bash
cd javascript
node flexbox_test.js
```

**No dependencies** - uses only Node.js standard library (`child_process` and `assert`).

### Python

```bash
cd python
python3 flexbox_test.py
```

**No dependencies** - uses only Python standard library (`subprocess` and `json`).

### Rust

```bash
cd rust

# Option 1: Using rust-script (quick)
cargo install rust-script
rust-script flexbox_test.rs

# Option 2: Normal compilation
rustc flexbox_test.rs -o flexbox_test
./flexbox_test
```

**Minimal dependencies** - only `serde_json` for JSON handling (standard practice in Rust).

## JSON Protocol

### Input Schema

```json
{
  "layout": {
    "type": "container",
    "style": {
      "display": "flex",
      "width": 600,
      "height": 100,
      ...
    },
    "children": [...]
  },
  "constraints": {
    "maxWidth": 800,
    "maxHeight": 600
  },
  "assertions": [
    {
      "type": "layout",
      "expression": "getX(root()) == 0",
      "message": "positioned"
    }
  ],
  "binding": "old"  // or "context"
}
```

### Output Schema

```json
{
  "passed": 3,
  "failed": 0,
  "skipped": 0,
  "results": [
    {
      "Assertion": {...},
      "Passed": true,
      "Error": ""
    }
  ]
}
```

Add `--verbose` flag to include computed layout in output:

```bash
echo '...' | wptest eval --verbose
```

```json
{
  "passed": 3,
  "failed": 0,
  "skipped": 0,
  "results": [...],
  "layout": {
    "x": 0,
    "y": 0,
    "width": 600,
    "height": 100,
    "children": [...]
  }
}
```

## Exit Codes

- `0`: All assertions passed
- `1`: One or more assertions failed

This makes integration with CI/CD systems trivial.

## Why This Approach?

### Advantages

1. **Zero language-specific bindings** - No CGO, FFI, JNI, or language-specific wrappers
2. **Simple integration** - Spawn process + JSON I/O works everywhere
3. **Language agnostic** - Same JSON format works in Go, Rust, Python, JS, C++, Ruby, etc.
4. **Easy debugging** - Pipe JSON files directly: `cat test.json | wptest eval`
5. **CI/CD friendly** - Standard exit codes, JSON output can be parsed by test reporters

### Performance Considerations

Spawning a process has overhead (~1-5ms), but for most test suites this is negligible:

- **Single test**: ~5ms overhead per test
- **Batched tests**: Run server mode (future): `wptest serve --socket /tmp/wptest.sock`

For high-performance scenarios, we can add server mode later without changing the JSON protocol.

## Adding Your Language

Want to add an example for your favorite language? Just:

1. Create a directory: `examples/cross-language/<language>/`
2. Write a test that spawns `wptest eval` and passes JSON via stdin
3. Parse JSON output and assert on results
4. Send a PR!

The protocol is intentionally simple to make this trivial.

## CEL Bindings

Two CEL API styles are available:

### Old Binding (default)

```javascript
"binding": "old"
```

- `getX(root())`, `getWidth(child(root(), 0))`
- No context awareness
- Works with all assertions

### Context Binding

```javascript
"binding": "context"
```

- Supports `this()` and `parent()` for context-aware assertions
- More concise: `getY(this)` instead of `getY(node)`
- Some assertions may be skipped if they use unsupported features

## Real-World Usage

This approach is production-ready for:

- **Unit tests** in your layout engine implementation
- **Integration tests** for UI frameworks
- **Regression tests** against W3C Web Platform Tests
- **CI/CD pipelines** with JSON output for test reporters
- **Fuzzing** by generating random layout specs
