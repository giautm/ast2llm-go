package prompts

import (
	"context"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vlad/ast2llm-go/internal/parser"
)

func TestNewEnhancePrompt(t *testing.T) {
	// Test that the prompt is properly defined with correct name and description
	p := parser.New()
	s := mcp.NewServer(&mcp.Implementation{Name: "Test Server", Version: "1.0.0"}, nil)
	
	err := RegisterPrompts(s, p)
	require.NoError(t, err)
	
	// The new SDK doesn't expose prompts in the same way, so we'll just verify registration succeeds
	// The actual prompt metadata will be tested through execution
}

func TestEnhancePromptHandler(t *testing.T) {
	// Initialize parser and server
	p := parser.New()
	s := mcp.NewServer(&mcp.Implementation{Name: "Test Server", Version: "1.0.0"}, nil)

	// Register the prompt
	err := RegisterPrompts(s, p)
	require.NoError(t, err)

	// Test cases
	tests := []struct {
		name        string
		args        map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid request",
			args: map[string]string{
				"projectPath": "./testdata/validproject",
			},
			wantErr: false,
		},
		{
			name: "missing required args",
			args: map[string]string{
				"focusSymbol": "main",
			},
			wantErr:     true,
			errContains: "projectPath is required",
		},
		{
			name: "with focus symbol",
			args: map[string]string{
				"projectPath": "./testdata/validproject",
				"focusSymbol": "MyStruct",
			},
			wantErr: false,
		},
		{
			name: "with minify",
			args: map[string]string{
				"projectPath": "./testdata/validproject",
				"minify":      "true",
			},
			wantErr: false,
		},
	}

	// Create dummy testdata directory and files
	err = os.MkdirAll("testdata/validproject", 0755)
	require.NoError(t, err)
	err = os.WriteFile("testdata/validproject/main.go", []byte("package main\n\n// MyStruct is a struct\ntype MyStruct struct{}\nfunc main(){}\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile("testdata/validproject/go.mod", []byte("module testproject\ngo 1.21\n"), 0644)
	require.NoError(t, err)

	defer os.RemoveAll("testdata")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &mcp.GetPromptRequest{
				Params: &mcp.GetPromptParams{
					Arguments: tt.args,
				},
			}

			handler := EnhancePromptHandler(p)
			result, err := handler(context.Background(), request)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			// Verify result structure
			assert.NotEmpty(t, result.Description)
			assert.NotEmpty(t, result.Messages)

			// Verify system message
			systemMsg := result.Messages[0]
			assert.Equal(t, mcp.Role("assistant"), systemMsg.Role)
			textContent, ok := systemMsg.Content.(*mcp.TextContent)
			require.True(t, ok)
			assert.Equal(t, "You are a Go code enhancement assistant. Your task is to improve the provided Go project code by adding better documentation, error handling, and following best practices.", textContent.Text)

			// Verify user message with project info
			userMsg := result.Messages[1]
			assert.Equal(t, mcp.Role("user"), userMsg.Role)
			textContent, ok = userMsg.Content.(*mcp.TextContent)
			require.True(t, ok)
			assert.Contains(t, textContent.Text, "project structure and parsed AST information")
			assert.Contains(t, textContent.Text, "MyStruct") // Check for some expected content
			assert.Contains(t, textContent.Text, "main.go")

			if tt.name == "with focus symbol" {
				// Find the message containing the focus symbol
				found := false
				for _, msg := range result.Messages {
					if tc, ok := msg.Content.(*mcp.TextContent); ok {
						if tc.Text == "Please pay special attention to the 'MyStruct' symbol in the code across the project." {
							found = true
							break
						}
					}
				}
				assert.True(t, found, "Expected to find focus symbol message")
			}

			// Check for minify message if applicable
			if tt.args["minify"] == "true" {
				// The minify message is the last one added if minify is true
				lastMsg := result.Messages[len(result.Messages)-1]
				assert.Contains(t, lastMsg.Content.(*mcp.TextContent).Text, "remove all comments and format the code to be more concise.")
			}

		})
	}
}

func TestRegisterPrompts(t *testing.T) {
	// Initialize parser and server
	p := parser.New()
	s := mcp.NewServer(&mcp.Implementation{Name: "Test Server", Version: "1.0.0"}, nil)

	// Register the prompt
	err := RegisterPrompts(s, p)
	require.NoError(t, err)

	// Verify prompt is registered by checking if we can get a handler for it
	handler := EnhancePromptHandler(p)
	require.NotNil(t, handler)

	// Create dummy testdata directory and files for the handler to use
	err = os.MkdirAll("testdata/validproject", 0755)
	require.NoError(t, err)
	err = os.WriteFile("testdata/validproject/main.go", []byte("package main\n\nfunc main(){}\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile("testdata/validproject/go.mod", []byte("module testproject\ngo 1.21\n"), 0644)
	require.NoError(t, err)

	defer os.RemoveAll("testdata")

	// Test the handler with a basic request
	request := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"projectPath": "./testdata/validproject",
			},
		},
	}

	result, err := handler(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Enhance Go project code with better documentation and error handling", result.Description)
}
