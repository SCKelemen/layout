# `test_user` Scripts

This directory contains ad-hoc, manual reproduction scripts used during issue investigation.

- They are **not** part of the main CI test suite.
- They are intended to be run one file at a time, not with `go test ./...`.
- A file ending with `_test.go` was intentionally renamed to avoid confusion with real Go tests.

For automated regression coverage, add or update `*_test.go` files in the main module root.
