package generator

import "fmt"

// ciWorkflowContent returns a GitHub Actions workflow that mirrors this
// project's own validation loop (build/vet/test/gofmt), pinned to
// goVersion (major.minor, e.g. "1.25").
func ciWorkflowContent(goVersion string) string {
	return fmt.Sprintf(`name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "%s"
      - run: go build ./...
      - run: go vet ./...
      - run: go test ./...
      - name: gofmt check
        run: |
          fmt_out="$(gofmt -l .)"
          if [ -n "$fmt_out" ]; then
            echo "$fmt_out"
            exit 1
          fi
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
`, goVersion)
}

// golangciConfigContent returns a minimal .golangci.yml, written alongside
// ci.yml whenever CI is included — lint pairs naturally with CI, so this
// follows the same "derive from context, no separate toggle" pattern as
// config_stub.go's config-loader detection.
func golangciConfigContent() string {
	return `run:
  timeout: 5m

linters:
  enable:
    - govet
    - staticcheck
    - unused
    - errcheck
    - gofmt
`
}
