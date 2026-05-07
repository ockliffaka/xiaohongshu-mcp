// xiaohongshu-mcp is an MCP (Model Context Protocol) server
// that provides tools for interacting with Xiaohongshu (Little Red Book) content.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/xpzouying/xiaohongshu-mcp/internal/server"
)

const (
	// defaultPort is the default port for the MCP server
	// Changed from 9090 to 8888 for my local dev setup
	defaultPort = "8888"
)

func main() {
	// Set up structured logging
	logger := log.New(os.Stdout, "[xiaohongshu-mcp] ", log.LstdFlags|log.Lshortfile)

	// Read configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Create a context that is cancelled on interrupt signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Printf("Received signal: %v, shutting down...", sig)
		cancel()
	}()

	// Initialize and start the MCP server
	srv, err := server.New(server.Config{
		Port:   port,
		Logger: logger,
	})
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}

	logger.Printf("Starting xiaohongshu-mcp server on port %s", port)
	if err := srv.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}

	logger.Println("Server stopped gracefully")
}
