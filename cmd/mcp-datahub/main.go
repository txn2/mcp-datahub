// Package main provides the mcp-datahub CLI entry point.
package main

import (
	"context"
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
	// Setup context with cancellation
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

	// Create server with default options (from environment)
	opts := server.DefaultOptions()
	mcpServer, mgr, err := server.New(opts)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer func() {
		if closeErr := mgr.Close(); closeErr != nil {
			log.Printf("Error closing manager: %v", closeErr)
		}
	}()

	// Test connection to default server
	defaultClient, err := mgr.Client("")
	if err != nil {
		log.Printf("Warning: Could not get default client: %v", err)
	} else if err := defaultClient.Ping(ctx); err != nil {
		log.Printf("Warning: Could not ping DataHub server: %v", err)
	}

	// Log startup info
	infos := mgr.ConnectionInfos()
	var defaultURL string
	for _, info := range infos {
		if info.IsDefault {
			defaultURL = info.URL
			break
		}
	}
	log.Printf("mcp-datahub %s (commit: %s, built: %s)", version, commit, buildTime)
	log.Printf("Starting with %d connection(s), default: %s",
		mgr.ConnectionCount(),
		defaultURL,
	)

	// Run server with stdio transport
	if err := mcpServer.Run(ctx, &mcp.StdioTransport{}); err != nil {
		if ctx.Err() != nil {
			// Context canceled, normal shutdown
			log.Println("Server stopped")
			return
		}
		log.Fatalf("Server error: %v", err)
	}
}
