# CI Guide

Current CI mainly covers:

- Windows main path
- Linux generation/compilation path

## What CI Does

Windows job:

- Install Go and Zig
- Generate `examples/basic`
- Run `go test ./...`
- Run `go test -bench . ./asmcall`

Linux job:

- Install Go and Zig
- Generate `examples/basic`
- Run `go test ./...`
- Run `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./...`

## Why Linux Runtime Live Testing is Disabled by Default

Current no-`cgo` runtime on Linux is still in a performance-oriented low-level implementation phase.

To make main CI more stable, the current default strategy is:

- Main CI verifies Linux path can generate, compile, and integrate build
- Bottom-level runtime live testing is enabled manually via environment variables

This avoids binding all CI results to assembly call details that are still being refined.

## Future Improvements

- Add manually triggered Linux runtime deep verification workflow
- Add benchmark result archiving
- After action ecosystem stabilizes, migrate to Node 24 supported versions