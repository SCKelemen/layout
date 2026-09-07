# WPT-Style Testing

The `wpt/` directory is a **nested Go module** (`github.com/SCKelemen/layout/wpt`) that exercises the engine with Web Platform Test-style assertions written in [CEL](https://github.com/google/cel-spec), using [wpt-test-gen](https://github.com/SCKelemen/wpt-test-gen).

## Why a separate module

`wpt-test-gen` pulls in cel-go, ANTLR, and protobuf. Keeping it in its own module means consumers of `github.com/SCKelemen/layout` never download those dependencies; the main module's `go.mod` stays at `text`, `unicode/v6`, `units`, and `yaml.v3`. The nested module's `go.mod` has `replace github.com/SCKelemen/layout => ../`, so its tests always run against the working tree.

## Running

```bash
cd wpt && go test ./...
```

CI runs this in the dedicated `WPT Examples (nested module)` job with `go-version-file: wpt/go.mod`. `go test ./...` from the repository root does not descend into `wpt/`.

## Writing a test

```go
func TestFlexboxWithCEL(t *testing.T) {
	root := layout.HStack(layout.Fixed(100, 50), layout.Fixed(100, 50), layout.Fixed(100, 50))
	root.Style.JustifyContent = layout.JustifyContentSpaceBetween
	root.Style.Width = layout.Px(600)
	root.Style.Height = layout.Px(100)
	layout.Layout(root, layout.Tight(600, 100), layout.NewLayoutContext(800, 600, 16))

	env, err := cel.NewLayoutCELEnv(root)
	if err != nil {
		t.Fatal(err)
	}
	results := env.EvaluateAll([]cel.CELAssertion{
		{Type: "layout", Expression: "getX(child(root(), 0)) == 0.0", Message: "first child at start"},
		{Type: "layout", Expression: "getRight(child(root(), 2)) == getWidth(root())", Message: "last child at end"},
	})
	for _, r := range results {
		if !r.Passed {
			t.Errorf("%s: %s", r.Assertion.Message, r.Error)
		}
	}
}
```

(`cel` is `github.com/SCKelemen/wpt-test-gen/pkg/cel`.) Available CEL functions: `root()`, `child(node, i)`, `childCount(node)`, `getX`, `getY`, `getTop`, `getLeft`, `getWidth`, `getHeight`, `getRight`, `getBottom`.

`wpt/flexbox_cel_example_test.go` shows flex and grid examples; `wpt/generated_standalone_test.go` is a test generated from an HTML fixture by `wptest`.

## Language-agnostic evaluation

The `wptest` CLI evaluates the same JSON layout + assertions from any language:

```bash
go install github.com/SCKelemen/wpt-test-gen/cmd/wptest@latest
echo '{"layout": {"display": "flex", "width": 600}, "assertions": [{"expression": "getX(root()) == 0.0"}]}' | wptest eval
```

See the [wpt-test-gen cross-language examples](https://github.com/SCKelemen/wpt-test-gen/tree/main/examples/cross-language).
