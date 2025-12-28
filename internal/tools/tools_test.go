package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad/ast2llm-go/internal/parser"
	// Alias ourtypes
)

func TestNewParseGoTool(t *testing.T) {
	// Test that the tool is properly defined with correct name and description
	p := parser.New()
	s := mcp.NewServer(&mcp.Implementation{Name: "Test Server", Version: "1.0.0"}, nil)
	
	err := RegisterTools(s, p)
	require.NoError(t, err)
	
	// The new SDK doesn't expose tools in the same way, so we'll just verify registration succeeds
	// The actual tool metadata will be tested through execution
}

func TestParseGoToolHandler(t *testing.T) {
	p := parser.New()
	handler := ParseGoToolHandler(p)

	// Create a dummy project for testing
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "testproject")
	err := os.MkdirAll(projectPath, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(`package main

import "fmt"

// MyStruct is a simple struct
type MyStruct struct{}

func main(){
	fmt.Println("Hello")
	_ = MyStruct{}
}
`), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte(fmt.Sprintf("module %s\ngo 1.21\n", "example.com/testproject_tools")), 0644)
	require.NoError(t, err)

	// Run go mod tidy to ensure go.mod is valid
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	require.NoError(t, err, "go mod tidy failed in test setup")

	tests := []struct {
		name        string
		args        ParseGoToolArgs
		wantErr     bool
		errContains string
	}{
		{
			name: "valid request",
			args: ParseGoToolArgs{
				ProjectPath: projectPath,
				FilePath:    "main.go",
			},
			wantErr: false,
		},
		{
			name:        "missing filePath",
			args:        ParseGoToolArgs{},
			wantErr:     true,
			errContains: "failed to",
		},
		{
			name: "invalid project path",
			args: ParseGoToolArgs{
				ProjectPath: "/non/existent/path",
				FilePath:    "main.go",
			},
			wantErr:     true,
			errContains: "failed to parse project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &mcp.CallToolRequest{}

			result, output, err := handler(context.Background(), request, tt.args)
			if tt.wantErr {
				require.NoError(t, err) // Handler error is in the result, not returned
				assert.NotNil(t, result)
				assert.True(t, result.IsError)
				assert.Contains(t, output.Result, tt.errContains)
				return
			}

			require.NoError(t, err)
			require.Nil(t, result) // Success returns nil result
			assert.NotEmpty(t, output.Result)

			// Verify the content is a JSON string representing FileInfo map
			composedOutput := output.Result
			assert.Contains(t, composedOutput, "--- File: "+filepath.Join(projectPath, "main.go")+" ---")
			assert.Contains(t, composedOutput, "Package: main")
			assert.Contains(t, composedOutput, "Local Structs:\n  Struct: example.com/testproject_tools.MyStruct")
			assert.NotContains(t, composedOutput, "Used Imported Structs (from this project, if available):\n- fmt")
		})
	}
}

func TestRegisterTools(t *testing.T) {
	p := parser.New()
	s := mcp.NewServer(&mcp.Implementation{Name: "Test Server", Version: "1.0.0"}, nil)

	err := RegisterTools(s, p)
	require.NoError(t, err)

	// Проверяем, что инструмент зарегистрирован
	handler := ParseGoToolHandler(p)
	require.NotNil(t, handler)

	// Create a dummy project for testing the handler
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "testproject_reg")
	err = os.MkdirAll(projectPath, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(projectPath, "main.go"), []byte("package main\nfunc init(){}\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte(fmt.Sprintf("module %s\ngo 1.21\n", "example.com/testproject_reg")), 0644)
	require.NoError(t, err)

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	require.NoError(t, err, "go mod tidy failed in test setup for registration")

	// Тестируем обработчик с базовым запросом
	request := &mcp.CallToolRequest{}
	args := ParseGoToolArgs{
		ProjectPath: projectPath,
		FilePath:    "main.go",
	}

	result, output, err := handler(context.Background(), request, args)
	require.NoError(t, err)
	require.Nil(t, result)
	assert.NotEmpty(t, output.Result)
	composedOutput := output.Result
	assert.Contains(t, composedOutput, "Package: main")
	assert.NotContains(t, composedOutput, "Local Structs:\n  Struct:")
	assert.NotContains(t, composedOutput, "Used Imported Structs (from this project, if available):\n")
}
