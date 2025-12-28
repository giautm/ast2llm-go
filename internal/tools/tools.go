package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vlad/ast2llm-go/internal/composer"
	"github.com/vlad/ast2llm-go/internal/parser"
)

// ParseGoToolArgs defines the arguments for the parse_go tool
type ParseGoToolArgs struct {
	ProjectPath string `json:"projectPath" jsonschema:"Path to the Go project"`
	FilePath    string `json:"filePath" jsonschema:"Path to the current file"`
}

// ParseGoToolOutput defines the output for the parse_go tool
type ParseGoToolOutput struct {
	Result string `json:"result" jsonschema:"The composed project information"`
}

// ParseGoToolHandler returns a handler for the parse_go tool
func ParseGoToolHandler(p *parser.ProjectParser) func(context.Context, *mcp.CallToolRequest, ParseGoToolArgs) (*mcp.CallToolResult, ParseGoToolOutput, error) {
	return func(ctx context.Context, request *mcp.CallToolRequest, args ParseGoToolArgs) (*mcp.CallToolResult, ParseGoToolOutput, error) {
		projectInfo, err := p.ParseProject(args.ProjectPath)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
			}, ParseGoToolOutput{Result: fmt.Sprintf("failed to parse project: %v", err)}, nil
		}

		fullFilePath := fmt.Sprintf("%s/%s", args.ProjectPath, args.FilePath)
		projectComposer := composer.New(projectInfo)

		info, err := projectComposer.Compose(fullFilePath)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
			}, ParseGoToolOutput{Result: fmt.Sprintf("failed to compose project info: %v", err)}, nil
		}

		return nil, ParseGoToolOutput{Result: info}, nil
	}
}

// RegisterTools registers all tools with the MCP server
func RegisterTools(s *mcp.Server, p *parser.ProjectParser) error {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "parse_go",
		Description: "Parse Go project and return its detailed information",
	}, ParseGoToolHandler(p))
	return nil
}
