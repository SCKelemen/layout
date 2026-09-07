# Contributing

## Running the tests

```bash
go test ./...                    # main module (race detector: go test -race ./...)
go test -tags no_yaml ./...      # without the YAML dependency (serialize.ToYAML/FromYAML are excluded)
cd wpt && go test ./...          # WPT-style examples; separate module, see docs/wpt-testing.md
```

Test files are named `<feature>_<topic>_test.go` next to the code they cover (`flexbox_wrap_reverse.go`, `grid_auto_repeat_test.go`, `text_vertical_positioning_test.go`). Test functions are `Test<Feature><Behavior>`; cite the spec section being checked in a comment, and assert on `Rect` values with a small epsilon rather than exact float equality.

Every example in `examples/` must run to completion (`go run ./examples/<name>`); CI executes all of them.

## Code quality

CI runs, and pull requests are expected to pass locally:

```bash
gofmt -s -l .        # must print nothing
go vet ./...
go vet -tags no_yaml ./...
go run honnef.co/go/tools/cmd/staticcheck@2026.2.1 ./...
go mod tidy          # go.mod / go.sum must not change
```

## CI jobs

`.github/workflows/test.yml` defines four jobs:

| Job | What it does |
|-----|--------------|
| Tests + Coverage | `go test -race -coverprofile ./...`, then `go test -tags no_yaml ./...`; uploads coverage |
| Code Quality | gofmt, `go vet` (both tag sets), staticcheck, `go mod tidy` clean |
| Examples | `go run` every `examples/*/` and `examples/fluent/*/` program |
| WPT Examples | `go test ./...` inside the `wpt/` nested module |

## Commits and changelog

- Conventional commits: `type(scope): description` in imperative mood (`fix(grid): honor justify-content`, `docs(text): describe SpaceAfter`). Keep the first line short. No trailers.
- Every user-visible change gets an entry in `CHANGELOG.md` under `[Unreleased]`, in the `Added`, `Changed`, or `Fixed` section. Mark entries that alter documented defaults as **behavior change** and cite the spec section that motivates them.
- Documentation and doc comments must describe the code as it is; when a change alters a documented default (for example the unset-means-auto `Length` convention), update `docs/gotchas.md` and `docs/limitations.md` in the same change.
- Dependencies: standard library first, then `golang.org/x/*`, then `github.com/SCKelemen/*`; anything else needs a reason in the PR.
