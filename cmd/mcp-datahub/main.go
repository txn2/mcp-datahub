// Package main provides the mcp-datahub CLI entry point.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/internal/server"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	// Create server with default options
	opts := server.DefaultOptions()
	mcpServer, datahubClient, err := server.New(opts)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Clean up client on exit
	if datahubClient != nil {
		defer func() {
			if err := datahubClient.Close(); err != nil {
				log.Printf("Error closing client: %v", err)
			}
		}()

		// Test connection
		if err := datahubClient.Ping(ctx); err != nil {
			log.Printf("Warning: DataHub connection test failed: %v", err)
		} else {
			log.Printf("mcp-datahub %s (commit: %s, built: %s)", version, commit, buildTime)
			log.Printf("Connected to DataHub at %s", datahubClient.Config().URL)
		}
	} else {
		log.Printf("mcp-datahub %s (commit: %s, built: %s)", version, commit, buildTime)
		log.Println("Warning: DataHub not configured - tools will return configuration errors")
	}

	// Run MCP server with stdio transport
	if err := mcpServer.Run(ctx, &mcp.StdioTransport{}); err != nil {
		if ctx.Err() != nil {
			log.Println("Server stopped")
			return nil
		}
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
