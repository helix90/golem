# Golem

**Golem** is a pure Go AIML2 interpreter library with a simple, high-level API and full support for AIML2 categories, patterns, and tags. It includes a command-line interface for testing and interactive chat.

## Features

- Full AIML 2.0 support: `*`, `_`, `<srai>`, `<sr>`, `<think>`, `<set>`, `<get>`, `<condition>`, `<random>`, `<topic>`, `<that>`
- Load `.aiml` files or zipped AIML archives
- Pre-load `.set` and `.map` files
- Per-session context tracking with user variables
- CLI chatbot for local interaction
- Debugging and tracing features
- Lightweight and fast — no external dependencies

## Getting Started

### Requirements

- Go 1.18+

### Installation

```bash
go install github.com/yourusername/golem@latest
```

### Basic Usage (Library)

```go
bot, _ := golem.NewBot(golem.Config{Debug: true})
bot.LoadZip("brain.zip")
response, _ := bot.Respond("hello", "user123")
fmt.Println(response)
```

### Using the CLI

```bash
golem -load brain.zip
```

Type messages interactively.

## Project Layout

```
golem/
├── cmd/         # CLI interface
├── engine/      # Interpreter core
├── parser/      # AIML2 XML parsing
├── store/       # User/session state
├── loader/      # File I/O and utilities
├── debug/       # Logging and diagnostics
├── examples/    # Sample bots
├── testdata/    # Conformance tests
```

## Development

```bash
make build
make test
```

## License

GPL v3 or later

---

For full documentation, see [`spec.md`](./spec.md).

