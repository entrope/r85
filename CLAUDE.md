# CLAUDE.md

## Project Overview

r85 is a Go package implementing a base-85 binary-to-text encoding scheme, similar to Ascii85 and ZeroMQ's Z85. It uses a custom alphabet starting at ASCII `(` (decimal 40) with substitutions (`<` -> `}`, `` ` `` -> `~`) to minimize escaping in most languages and HTML.

Module path: `github.com/entrope/r85`

## Build & Test Commands

- Run all tests: `go test ./...`
- Run tests with verbose output: `go test -v ./...`
- Run benchmarks: `go test -bench=. -benchmem ./...`
- Run a specific benchmark: `go test -bench=BenchmarkEncode -benchmem ./...`

## Code Structure

Single-package library with three files:
- `r85.go` — All encoding/decoding logic: `Encode`, `Decode`, `EncodeToString`, `DecodeString`, streaming `NewEncoder`/`NewDecoder`, lookup tables, `CorruptInputError`
- `r85_test.go` — Unit tests
- `r85_bench_test.go` — Benchmarks across various payload sizes and chunking strategies

## Key Design Details

- Encoding: 4 bytes -> 5 chars (big-endian); partial trailing blocks (1-3 bytes) produce 2-4 chars
- Decoding: skips characters outside the r85 alphabet; accepts both canonical (`}`, `~`) and unescaped (`<`, `` ` ``) forms
- Streaming encoder buffers internally (4096-byte output buffer) and only writes partial blocks on `Close()`
- Streaming decoder carries incomplete blocks across `Read` calls
- A single trailing r85 digit is a `CorruptInputError`; value overflows are also reported as `CorruptInputError`
