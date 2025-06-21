# TODO Checklist for Golem v1

## Phase 1: Project Setup
- [ ] Initialize Go module (`go mod init golem`)
- [ ] Create base directory structure
- [ ] Set up `Makefile` and CI config
- [ ] Add basic `README.md`, `LICENSE`, and `spec.md`

## Phase 2: Core Infrastructure
- [ ] Define `Bot` struct and config
- [ ] Implement session store abstraction
- [ ] Add core API: `Respond`, `RespondVerbose`
- [ ] Add user variable storage and retrieval

## Phase 3: Parser
- [ ] Build XML parser for `.aiml` files
- [ ] Support AIML `<category>`, `<pattern>`, `<template>`
- [ ] Parse and attach `<that>`, `<topic>`
- [ ] Parse `<set>`, `<get>`, `<think>`, `<condition>`
- [ ] Parse wildcards `*` and `_`

## Phase 4: Engine
- [ ] Implement pattern matching tree
- [ ] Handle star capturing and binding
- [ ] Implement evaluation of AIML tags
- [ ] Implement recursion with `<srai>` and `<sr>`
- [ ] Add recursion depth limiter

## Phase 5: File Loader
- [ ] Load individual `.aiml` files
- [ ] Load `.zip` archives
- [ ] Load `.set` and `.map` files
- [ ] Support Gob/JSON brain serialization

## Phase 6: CLI Tool
- [ ] Implement REPL loop
- [ ] Add `-load` flag
- [ ] Add `-debug` flag
- [ ] Use Bot API to process input/output

## Phase 7: Debugging and Logging
- [ ] Add debug flag to config
- [ ] Emit matched pattern and trace info
- [ ] Provide internal error logging

## Phase 8: Testing
- [ ] Add unit tests for parser
- [ ] Add tests for engine behavior (wildcards, sets, recursion)
- [ ] Add test AIML files
- [ ] Implement conformance tests

## Phase 9: Documentation
- [ ] Update `README.md` with examples
- [ ] Document all public API functions
- [ ] Add usage documentation for CLI
- [ ] Provide OOB JSON handling example

## Phase 10: Release Prep
- [ ] Verify Go 1.18+ compatibility
- [ ] Finalize conformance suite
- [ ] Create example bot package
- [ ] Tag v1.0.0 release

---
Future items (v2):
- [ ] Hot-reload brains
- [ ] Input normalization pipeline
- [ ] Batch mode in CLI
- [ ] Swap brain segments to/from disk
- [ ] WebAssembly and i18n support
- [ ] Debian/Homebrew packaging

