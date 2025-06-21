# Golem Specification - AIML2 Interpreter Library for Go

## Overview
Golem is a pure Go library for interpreting AIML2-compliant chatbot scripts. It is designed for use in server-based chatbot systems and a simple CLI for interactive use. This document specifies all necessary architectural, functional, and implementation details for version 1.

## Goals for Version 1
- Full support for AIML2 tags and behavior
- Load `.aiml` files and `.zip` archives
- In-memory loading of sets and maps
- Tree-based pattern matching
- Simple Go API for response handling
- Debug and tracing support
- CLI for interactive local chat
- Bundled test suite and conformance tests

## License
- GPL v3 (or later)

## Go Environment
- Minimum required Go version: 1.18
- Pure Go only (no cgo)

## Architecture
### Module Layout
```
golem/
├── cmd/             # CLI entry point
├── engine/          # Pattern match, tag execution
├── parser/          # AIML2 XML parser
├── store/           # Session/context state
├── loader/          # File loading (.aiml, .zip, .set, .map)
├── debug/           # Logging, tracing
├── internal/        # Helpers/utilities
├── examples/        # Sample AIML and sessions
├── testdata/        # Conformance test sets
└── go.mod
```

## Functional Requirements
### Pattern Matching
- Wildcards: `*`, `_`
- Recursive matching (`<srai>`, `<sr>`, `<that>`, `<topic>`)
- Per-session context
- Recursion depth limit (e.g. 10)

### Template Evaluation
- Required tags: `<template>`, `<srai>`, `<sr>`, `<set>`, `<get>`, `<think>`, `<condition>`, `<random>`
- Nested tag handling
- Focus on correct implementation of `<think>` and `<condition>`

### OOB Handling
- `<oob>` is passed through unchanged
- One documented example for JSON-based `<oob>` handling

### Input Expectations
- Pre-normalized (lowercase, stripped punctuation)
- No internal normalization in v1

### Data Loading
- Support `.aiml` files and `.zip` archives
- `.set` and `.map` loaded into memory
- Lenient parsing with informative warnings

### Session Management
- Each session keyed by session ID
- Stores: topic, that, variables
- Injectable or retrievable maps

### Serialization
- Support saving/loading parsed brain as Gob or JSON
- Faster startup and development convenience

## Public API
```go
type Bot struct {
    // internal state
}

func NewBot(config Config) (*Bot, error)
func (b *Bot) LoadAIML(path string) error
func (b *Bot) LoadZip(path string) error
func (b *Bot) LoadSets(path string) error
func (b *Bot) LoadMaps(path string) error
func (b *Bot) Respond(input string, sessionID string) (string, error)
func (b *Bot) RespondVerbose(input string, sessionID string) (Response, error)

type Response struct {
    Text       string
    Matched    string
    Wildcards  []string
    DebugTrace []string
}
```

## CLI Tool
- Live in `cmd/`
- Interactive chat only
- CLI options: `-load`, `-debug`

## Debugging
- Debug mode via config or CLI
- Matched category, wildcards, evaluation trace

## Error Handling
- Graceful fallback on malformed AIML
- Recursion overflow returns safe message
- Non-fatal parse errors logged with context

## Testing Plan
### Unit Tests
- Parser correctness
- Engine logic (wildcards, star, set/get)

### Integration Tests
- Input/output from sample AIML brains

### Conformance Tests
- Located in `testdata/`
- Automatic validation with `go test`

## Developer Tools
- `go generate` for brain snapshots
- `Makefile` for build/test
- CI (GitHub Actions)
- Docs: `README.md`, `CONTRIBUTING.md`, `docs/spec.md`

## Deferred to Future Versions
- Hot-reload
- Normalization pipeline
- Batch mode CLI
- Swappable brain paging
- Custom OOB handlers
- i18n support
- WebAssembly targets
- Package formats (Debian, Homebrew)

## Target Audience
- Developers embedding chatbots in Go-based systems
- Researchers and hobbyists exploring AIML2 bots

---
End of Specification

