# WPT-style example tests

This nested module holds example tests that exercise the layout engine with
CEL assertions from [wpt-test-gen](https://github.com/SCKelemen/wpt-test-gen).
It is a separate Go module so that `wpt-test-gen` and its transitive
dependencies (cel-go, antlr, protobuf) stay out of the main module's
dependency graph; consumers of `github.com/SCKelemen/layout` never download
them.

Run the examples from this directory:

```sh
cd wpt && go test ./...
```

The `replace` directive in `go.mod` points at the parent module, so the tests
always run against the working-tree version of the engine.
