package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vlad/ast2llm-go/internal/parser"
	"github.com/vlad/ast2llm-go/internal/prompts"
	"github.com/vlad/ast2llm-go/internal/tools"
)

func main() {
	// Initialize components
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "AST2LLM",
		Version: "1.0.0",
	}, nil)
	p := parser.New()

	// Register tools
	if err := tools.RegisterTools(s, p); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	// Register prompts
	if err := prompts.RegisterPrompts(s, p); err != nil {
		log.Fatalf("Failed to register prompts: %v", err)
	}

	// Create a context that will be cancelled on interrupt signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received interrupt signal, shutting down gracefully...")
		cancel()
	}()

	// Start the stdio server
	if err := s.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}
