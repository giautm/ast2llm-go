# Copilot Instructions for ast2llm-go

## Repository Overview

**ast2llm-go** is an MCP (Model Context Protocol) server that provides AST-powered context enhancement for LLMs working with Go codebases. It analyzes Go project structure using precise AST analysis and injects relevant code context into prompts, delivering 3-5x faster context resolution than grep-based approaches.

- **Size**: ~19MB repository, 20 Go source files
- **Language**: Go 1.24.1+ (go.mod specifies 1.24.1, CI uses 1.22, README mentions 1.22+)
- **Type**: CLI tool/MCP server application
- **Main Binary**: `ast2llm-go` (MCP server for stdio communication)
- **Key Dependencies**: 
  - `github.com/modelcontextprotocol/go-sdk` - MCP protocol implementation
  - `golang.org/x/tools/go/packages` - Go package parsing
  - `github.com/stretchr/testify` - Testing framework

## Build & Development Workflow

### Prerequisites
- **Go 1.22 or higher** required for building and testing
- **golangci-lint** required for linting (install with: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- Ensure `$(go env GOPATH)/bin` is in your PATH to use installed Go tools

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
The full cycle (download deps, test, build) takes approximately **2 seconds** total:
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
- `cmd/server/main.go` - Main MCP server initialization, registers tools and prompts
- `cmd/parser-cli/main.go` - CLI tool for testing parser functionality

**Core Logic:**
- `internal/parser/project_parser.go` - Parses Go projects using `go/packages` and `go/types`, extracts structs, interfaces, functions, and dependencies
- `internal/composer/project_composer.go` - Composes parsed AST data into LLM-friendly text format
- `internal/types/types.go` - Core type definitions: FileInfo, StructInfo, InterfaceInfo, FunctionInfo, GlobalVarInfo
- `internal/tools/tools.go` - Implements the `parse_go` MCP tool
- `internal/prompts/prompts.go` - Implements the `enhance` MCP prompt

## Testing

### Test Coverage
- All `internal/*` packages have comprehensive test coverage
- `cmd/*` packages have NO test files (CLI entry points only)
- Tests use table-driven approach with subtests
- Tests run in parallel where possible (using `t.Parallel()`)

### Running Tests
```bash
# Run all tests with verbose output
go test -v ./...

# Run tests for specific package
go test -v ./internal/parser

# Run with coverage (CI approach, excludes cmd/)
PKGS=$(go list ./... | grep -v '/cmd/')
go test -coverpkg=./... -covermode=atomic -coverprofile=coverage.out $PKGS
```

**Expected Test Output**: All tests should pass in ~2 seconds. No test data directories exist; tests use in-memory test fixtures.

## Continuous Integration (CI/CD)

### CI Pipeline (.github/workflows/ci.yml)
Runs on: Pull requests to `main`, pushes to `main`, manual dispatch

**Steps:**
1. Checkout with full history (`fetch-depth: 0`)
2. Setup Go 1.22 with caching
3. `go mod download`
4. `go test -v ./...`
5. Generate coverage report (excludes `/cmd/`)
6. Upload coverage to Coveralls
7. **Linting is DISABLED** (commented out in workflow, lines 46-50)
8. Build binaries for multiple platforms (linux/darwin/windows, amd64/arm64)
9. Compress binaries with UPX
10. Generate SHA256 checksums
11. Upload build artifacts
12. Create GitHub release (on push to main only) with date-based versioning: `v{YYYYMMDD}-{last4SHA}`

### Release Pipeline (.github/workflows/release.yml)
Runs on: Push of version tags (`v*`), manual dispatch

**Uses GoReleaser** to:
- Run tests before building (`go test ./...`)
- Build multi-platform binaries with optimizations (`-s -w` ldflags)
- Create Docker images and push to `ghcr.io/giautm/ast2llm-go`
- Generate changelog and GitHub release

### Important CI Notes
- **Go Version**: CI uses Go 1.22, but go.mod specifies 1.24.1 (both work)
- **Linting**: Not enforced in CI (commented out), but can be run locally
- **Coverage**: Only packages outside `/cmd/` are included in coverage
- **Build Artifacts**: Compressed with UPX for smaller binary size
- **Release Versioning**: Automatic releases use `v{YYYYMMDD}-{last4SHA}` format, manual releases use git tags

## Configuration Files

### Build Configuration
- **go.mod**: Module path is `github.com/vlad/ast2llm-go` (note: GitHub repo is `giautm/ast2llm-go`, this is intentional)
- **Makefile**: Defines targets for `build`, `test`, `lint`, `check`, `help`
- **.goreleaser.yaml**: Multi-platform build config (Linux/macOS/Windows, amd64/arm64)

### Linting Configuration
- **No custom config**: Uses golangci-lint defaults (no `.golangci.yml` file exists)
- **Current status**: All code passes golangci-lint with default settings

### Ignored Files (.gitignore)
- Build artifacts: `*.exe`, `*.dll`, `*.so`, `*.dylib`, `dist/`, `build/`
- Test outputs: `*.test`, `*.out`, `coverage.*`, `*.coverprofile`
- Binaries: `mcp-server`, `server`
- IDE files: `.idea/`, `.vscode/`, `.DS_Store`
- Environment: `.env`, `go.work`, `go.work.sum`

## Making Changes

### Before Committing
1. **Always test your changes**: `go test ./...`
2. **Build to ensure no compilation errors**: `go build -o mcp-server ./cmd/server`
3. **Run linter** (if golangci-lint is available): `$(go env GOPATH)/bin/golangci-lint run ./...`
4. **Verify changes are minimal**: Check `git diff` and ensure only necessary files are modified

### Testing Changes to Parser/Composer
When modifying `internal/parser` or `internal/composer`:
- Add or update tests in corresponding `*_test.go` files
- Follow existing table-driven test patterns
- Use `testify/assert` and `testify/require` for assertions
- Test with realistic Go code samples (existing tests provide good examples)

### Testing Changes to MCP Tools/Prompts
When modifying `internal/tools` or `internal/prompts`:
- Update corresponding test files
- Test the full MCP integration (tools registration, argument validation, execution)
- Ensure error messages are clear and helpful

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
