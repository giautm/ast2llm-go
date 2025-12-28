# Copilot Instructions for ast2llm-go

## Repository Overview

**ast2llm-go** is an MCP (Model Context Protocol) server for AST-powered Go code context enhancement. It delivers 3-5x faster context resolution than grep-based approaches.

- **Size**: ~19MB, 20 Go source files
- **Language**: Go 1.22+ required
- **Type**: CLI tool/MCP server (stdio communication)
- **Key Dependencies**: `modelcontextprotocol/go-sdk`, `golang.org/x/tools/go/packages`, `stretchr/testify`

## Build & Development Workflow

### Prerequisites
- Go 1.22+ required
- golangci-lint: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- Ensure `$(go env GOPATH)/bin` is in PATH

### Essential Commands (Always Run in This Order)

**1. Download Dependencies (ALWAYS run first after cloning or pulling)**
```bash
go mod download
```

**2. Run Tests (takes ~2 seconds)**
```bash
go test -v ./...
```

**3. Build the Binary**
```bash
go build -o mcp-server ./cmd/server
```
Or use the Makefile:
```bash
make build
```

**4. Run Linter**

**IMPORTANT**: The Makefile's `make lint` target will fail if golangci-lint is not in your system PATH, even if installed via `go install`. To work around this:

```bash
# Option 1: Add GOPATH/bin to PATH first
export PATH=$PATH:$(go env GOPATH)/bin
make lint

# Option 2: Run golangci-lint directly
$(go env GOPATH)/bin/golangci-lint run ./...
```

**5. Run All Checks (test + lint)**
```bash
# After adding golangci-lint to PATH:
export PATH=$PATH:$(go env GOPATH)/bin
make check
```

### Complete Build & Test Cycle
Full cycle takes ~2 seconds:
```bash
go mod download && go test ./... && go build -o mcp-server ./cmd/server
```

### Known Build Issues & Workarounds

**Issue**: `make lint` fails with "golangci-lint is not installed" even when installed via `go install`
- **Cause**: Makefile uses `command -v` which checks system PATH, but `go install` puts binaries in `$(go env GOPATH)/bin` which may not be in PATH
- **Workaround**: Always run `export PATH=$PATH:$(go env GOPATH)/bin` before `make lint` or `make check`
- **Alternative**: Run golangci-lint directly: `$(go env GOPATH)/bin/golangci-lint run ./...`

## Project Structure

```
ast2llm-go/
├── cmd/
│   ├── server/          # Main MCP server entry point (main.go)
│   └── parser-cli/      # CLI parser tool (no tests)
├── internal/
│   ├── parser/          # Core Go AST parsing logic (uses go/packages)
│   ├── composer/        # Formats parsed AST data for LLM consumption
│   ├── types/           # Type definitions for file/struct/function info
│   ├── tools/           # MCP tool implementations (parse_go tool)
│   └── prompts/         # MCP prompt implementations
├── .github/workflows/
│   ├── ci.yml          # CI pipeline: test, coverage, build, release
│   └── release.yml     # GoReleaser workflow (triggered on version tags)
├── Makefile            # Build automation (build, test, lint, check)
├── .goreleaser.yaml    # Release configuration for multi-platform builds
├── go.mod              # Go module definition
└── README.md           # User documentation
```

### Key Source Files

**Entry Points:**
- `cmd/server/main.go` - MCP server, registers tools and prompts
- `cmd/parser-cli/main.go` - CLI for testing parser

**Core:**
- `internal/parser/project_parser.go` - Parses Go projects, extracts types/functions
- `internal/composer/project_composer.go` - Formats AST for LLM consumption
- `internal/types/types.go` - FileInfo, StructInfo, InterfaceInfo, FunctionInfo
- `internal/tools/tools.go` - `parse_go` MCP tool
- `internal/prompts/prompts.go` - `enhance` MCP prompt

## Testing

### Test Coverage
- All `internal/*` packages have comprehensive test coverage
- `cmd/*` packages have NO test files (CLI entry points only)
- Tests use table-driven approach with subtests
- Tests run in parallel where possible (using `t.Parallel()`)

### Running Tests
```bash
go test -v ./...                    # All tests (verbose)
go test -v ./internal/parser        # Specific package

# Coverage (CI approach, excludes cmd/)
PKGS=$(go list ./... | grep -v '/cmd/')
go test -coverpkg=./... -covermode=atomic -coverprofile=coverage.out $PKGS
```

**Expected**: All tests pass in ~2 seconds. No external test data; uses in-memory fixtures.

## Continuous Integration (CI/CD)

### CI Pipeline (.github/workflows/ci.yml)
Triggers: PRs to main, pushes to main, manual

**Steps**: Checkout → Setup Go 1.22 → Download deps → Test → Coverage → Build multi-platform → UPX compress → Generate checksums → Upload artifacts → Create release (main pushes only, format: `v{YYYYMMDD}-{last4SHA}`)

**Note**: Linting disabled in CI (lines 46-50 commented out)

### Release Pipeline (.github/workflows/release.yml)
Triggers: Version tags (`v*`), manual

Uses GoReleaser: Test → Build multi-platform (`-s -w` ldflags) → Docker (ghcr.io/giautm/ast2llm-go) → Release

### Important CI Notes
- **Go Version**: CI uses Go 1.22 (stable). go.mod specifies 1.24.1 (dev/unstable), but the code is compatible with 1.22+
- Linting not enforced in CI, run locally for quality
- Coverage excludes `/cmd/`
- UPX-compressed artifacts
- Auto releases: `v{YYYYMMDD}-{last4SHA}`, manual: git tags

## Configuration Files

- **go.mod**: Module `github.com/vlad/ast2llm-go` (GitHub repo is `giautm/ast2llm-go` - intentional)
- **Makefile**: `build`, `test`, `lint`, `check`, `help`
- **.goreleaser.yaml**: Multi-platform (Linux/macOS/Windows, amd64/arm64)
- **Linting**: Uses golangci-lint defaults (no custom config)

### Ignored Files (.gitignore)
- Build: `*.exe`, `*.dll`, `*.so`, `dist/`, `build/`, `mcp-server`, `server`
- Tests: `*.test`, `*.out`, `coverage.*`
- Other: `.idea/`, `.vscode/`, `.env`, `go.work*`

## Making Changes

### Before Committing
1. Test: `go test ./...`
2. Build: `go build -o mcp-server ./cmd/server`
3. Lint (optional): `$(go env GOPATH)/bin/golangci-lint run ./...`
4. Verify minimal changes: `git diff`

### Parser/Composer Changes
- Add/update tests in `*_test.go` files
- Follow table-driven test patterns
- Use `testify/assert` and `testify/require`

### MCP Tools/Prompts Changes
- Update tests for registration, validation, execution
- Ensure clear error messages

## Common Pitfalls & Tips

1. **PATH Issue with Go Tools**: Always add `$(go env GOPATH)/bin` to PATH when using `go install` tools like golangci-lint
2. **Module Path Mismatch**: The module path in go.mod is `github.com/vlad/ast2llm-go` but the GitHub repo is `giautm/ast2llm-go` - this is intentional, do not change it
3. **No Testdata**: This project doesn't use external test data files; all tests use in-memory fixtures
4. **Fast Tests**: Full test suite runs in ~2 seconds; if slower, investigate why
5. **Binary Name**: Makefile builds `mcp-server`, but the release binary is named `ast2llm-go`
6. **CI Linting**: Linting is not enforced in CI but should be run locally for code quality

## Trust These Instructions

These instructions have been validated by running all commands and testing the complete development workflow. Trust them as accurate and only search for additional information if:
- You encounter an error not documented here
- You need to understand specific implementation details not covered
- These instructions appear outdated (check git history of this file)

Last validated: 2025-12-28 with Go 1.24.11
