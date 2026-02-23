# godiff

[![CI](https://github.com/njchilds90/godiff/actions/workflows/ci.yml/badge.svg)](https://github.com/njchilds90/godiff/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/njchilds90/godiff.svg)](https://pkg.go.dev/github.com/njchilds90/godiff)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Structured, human-readable text and data diffing for Go.**

Inspired by Python's `difflib` and JavaScript's `jsdiff` — but idiomatic Go, zero dependencies, and AI-agent friendly.

## Features

- **Character-level diff** via Myers algorithm
- **Word-level diff** with whitespace preservation
- **Line-level diff** for file comparison
- **Unified diff output** (standard patch format)
- **Patch application** — apply a diff back to a source string
- **JSON structural diff** — deep compare two JSON documents
- **Similarity ratio** — how similar are two strings? (0.0–1.0)
- **LCS** — Longest Common Subsequence of lines
- **Closest match** — find the best candidate from a list
- **context.Context support** throughout
- **Structured, machine-readable errors**
- **Zero external dependencies**

## Install
```bash
go get github.com/njchilds90/godiff
```

## Quick Start
```go
import "github.com/njchilds90/godiff"

// Line diff
patch := godiff.Lines("foo\nbar\nbaz\n", "foo\nqux\nbaz\n")
fmt.Print(godiff.Unified(patch, "old.go", "new.go", 3))

// Word diff
for _, op := range godiff.Words("the cat sat", "the dog sat") {
    fmt.Println(op)
}

// Char diff + apply
patch = godiff.Chars("kitten", "sitting")
result, _ := patch.Apply("kitten")
fmt.Println(result) // "sitting"

// Similarity
fmt.Println(godiff.RatioChars("hello", "helo")) // ~0.88

// JSON structural diff
ops, _ := godiff.JSONStrings(`{"name":"Alice","age":30}`, `{"name":"Bob","age":31}`)
for _, op := range ops {
    fmt.Println(op)
}

// Closest match
best := godiff.ClosestMatch("helo", []string{"hello", "world", "help"})
fmt.Println(best) // "hello"
```

## API

| Function | Description |
|---|---|
| `Chars(a, b string) Patch` | Character-level diff |
| `Words(a, b string) Patch` | Word-level diff |
| `Lines(a, b string) Patch` | Line-level diff |
| `Unified(patch, old, new, ctx) string` | Unified diff format |
| `Ratio(a, b string) float64` | Line-level similarity ratio |
| `RatioChars(a, b string) float64` | Char-level similarity ratio |
| `LCS(a, b string) []string` | Longest common subsequence |
| `ClosestMatch(target, candidates) string` | Best matching candidate |
| `ClosestMatches(target, candidates, n) []string` | Top-n matching candidates |
| `JSON(a, b []byte) (JSONPatch, error)` | JSON structural diff |
| `JSONStrings(a, b string) (JSONPatch, error)` | JSON diff from strings |
| `*Context variants` | All above with `context.Context` |
| `Patch.Apply(src string) (string, error)` | Apply patch to source |
| `Patch.Stats()` | Count equal/insert/delete ops |
| `Patch.HasChanges() bool` | Quick change check |

## AI Agent Usage

`godiff` is designed for use in AI pipelines:

- All functions are **pure and deterministic**
- Errors are **structured and typed**
- No global state
- `context.Context` support throughout
- JSON diff returns machine-readable `[]JSONOp` with typed paths

## License

MIT
