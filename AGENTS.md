# AGENTS.md

## Purpose
This repository is a small Go CLI named `hlp`.
Make minimal, correct changes that preserve the current package layout and CLI behavior.

## Repo-Specific Rules Inventory
- No `.cursorrules` file exists.
- No `.cursor/rules/` files exist.
- No `.github/copilot-instructions.md` file exists.
- Do not assume additional hidden repository rules.

## Toolchain
- Module: `github.com/yiblet/hlp`
- Declared Go version: `go 1.21`
- Declared toolchain: `go1.24.2`
- Prefer the repository's declared Go toolchain when running commands.

## Layout
- `cmd/hlp/main.go`: CLI process entrypoint and top-level error handling.
- `internal/cli/`: argument structs and subcommand dispatch.
- `internal/config/`: config load/save and OpenAI client construction.
- `internal/llm/`: chat request types and streaming implementation.
- `internal/chatfile/`: chat-file parsing, validation, and tests.
- `internal/term/`: terminal color and optional `bat` output integration.
- `internal/xio/`: small CLI IO helpers such as help rendering.
- `internal/xerr/`: shared error wrappers used at the program boundary.

Keep `cmd/hlp` thin. New business logic should usually go in `internal/...`, not in the entrypoint.

## Build Commands
These commands were verified in this repo.

- Build all packages: `go build ./...`
- Build the CLI package: `go build ./cmd/hlp`
- Build a local binary: `go build -o hlp`
- Build a local binary from the CLI entrypoint: `go build -o hlp ./cmd/hlp`
- Run the CLI locally: `go run ./cmd/hlp`
- Show help: `go run ./cmd/hlp --help`

## Format Commands
There is no custom formatter config. Use standard Go formatting.

- Format all packages: `go fmt ./...`
- Format one package: `go fmt ./internal/chatfile`
- Format the CLI package only: `go fmt ./cmd/hlp`

`go fmt ./...` is the safest default before finishing work.

## Lint / Static Analysis
There is no `golangci-lint`, Makefile, Taskfile, or Justfile in this repo.
Use standard Go tooling.

- Vet all packages: `go vet ./...`
- Vet one package: `go vet ./internal/llm`

For most changes, `go vet ./...` plus `go test ./...` is the minimum verification bar.

## Test Commands
Verified working commands:

- Run all tests: `go test ./...`
- Run one package: `go test ./internal/llm`
- Run one package verbosely: `go test -v ./internal/llm`
- Disable cache: `go test -count=1 ./...`

Run a single top-level test:

```bash
go test ./internal/chatfile -run '^TestParseChatFile$'
```

Run a specific subtest:

```bash
go test ./internal/llm -run '^TestStreamer_ChatStream$/^valid stream$'
```

Useful patterns:
- Single test: `go test ./pkg -run '^TestName$'`
- Single subtest: `go test ./pkg -run '^TestName$/^subtest name$'`

## Recommended Verification Flow
For most code changes, run:

1. `go fmt ./...`
2. `go vet ./...`
3. `go test ./...`

During iteration, package-scoped commands are fine, but finish with `go test ./...` unless the task explicitly does not require full verification.

## Code Style
Follow existing Go idioms already used in this repository instead of introducing a new style.

### Formatting
- Let `go fmt` decide formatting.
- Keep comments short and factual.

### Imports
- Use normal Go grouped imports.
- Standard library first, then a blank line, then third-party or local imports.
- Remove unused imports immediately.

### Types and APIs
- Prefer concrete structs for command state and config.
- Keep implementation details unexported unless another package needs them.
- Export only true cross-package API types, as with `llm.Message`, `llm.Input`, `llm.Streamer`, `config.Config`, and `chatfile.InvalidRoleError`.
- Use interfaces sparingly and only at real abstraction boundaries.

### Naming
- Use short, direct Go names.
- Exported identifiers use `CamelCase`.
- Unexported identifiers use `camelCase`.
- In `internal/cli`, keep command type names consistent with the current exported pattern, such as `AskCmd`, `ChatCmd`, and `ConfigCmd`.
- Avoid new abbreviations unless they are standard or already established here.

### Functions
- Keep functions focused and straightforward.
- Prefer early returns.
- Return `error` from helpers instead of logging inside them.

### Error Handling
- Return errors up the stack instead of swallowing them.
- Wrap with `fmt.Errorf("...: %w", err)` when adding useful context.
- Use `errors.Is` and `errors.As` for classification.
- Log only at the program boundary; `main()` already does this.
- Avoid introducing new `panic` usage for normal runtime failures.

### Context
- Accept and propagate `context.Context` in top-level execution paths and API boundaries.
- Do not replace upstream context with unrelated `context.Background()` inside inner layers.

### I/O and CLI Behavior
- Preserve current `stdin`, `stdout`, and file-based behavior.
- Preserve existing `-` semantics for stdin/stdout flows.
- Keep the `bat`-backed output path isolated in `internal/term` rather than mixing terminal concerns into command or config packages.

### Tests
- Prefer table-driven tests when there are multiple input/output cases.
- Use `t.Run(...)` for named cases.
- Use `t.Parallel()` only when the test is actually concurrency-safe.
- `testify/assert` is already used in `internal/chatfile/parse_test.go`; continuing to use it is acceptable.
- Keep test names stable so single-test invocation remains easy.

### Comments
- Comment exported or non-obvious code.
- Do not add narrative comments for straightforward logic.

## Agent Change Guidelines
- Keep edits minimal and local.
- Do not add new dependencies unless the task clearly needs them.
- Do not add a new lint framework unless requested.
- Preserve the current package split unless the task explicitly calls for another structural change.
- Do not silently rewrite unrelated rough edges while doing an unrelated task.

## Default Agent Workflow
1. Read the relevant package and nearby tests.
2. Make the smallest coherent change.
3. Run `go fmt ./...`.
4. Run targeted tests during iteration.
5. Run `go test ./...` before finishing.
6. Report any verification you could not run.
