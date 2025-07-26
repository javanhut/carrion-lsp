package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/javanhut/carrion-lsp/internal/protocol"
	"github.com/javanhut/carrion-lsp/internal/server"
)

const version = "0.1.0"

func main() {
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		showHelp    = flag.Bool("help", false, "Show help information")
		stdio       = flag.Bool("stdio", true, "Use stdio for communication (default)")
		carrionPath = flag.String("carrion-path", "", "Path to Carrion installation directory")
		logFile     = flag.String("log", "", "Log file path (default: stderr)")
		verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Carrion Language Server Protocol implementation\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --stdio                    # Start server with stdio (default)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --carrion-path=/usr/local/carrion  # Specify Carrion installation\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --log=carrion-lsp.log     # Log to file\n", os.Args[0])
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("carrion-lsp version %s\n", version)
		os.Exit(0)
	}

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Set up logging
	var logger *log.Logger
	if *logFile != "" {
		file, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		logger = log.New(file, "[carrion-lsp] ", log.LstdFlags|log.Lshortfile)
	} else {
		if *verbose {
			logger = log.New(os.Stderr, "[carrion-lsp] ", log.LstdFlags|log.Lshortfile)
		} else {
			// Silent logging to avoid interfering with LSP communication
			logger = log.New(os.Stderr, "[carrion-lsp] ", log.LstdFlags)
		}
	}

	// Create server options
	opts := server.ServerOptions{
		CarrionPath: *carrionPath,
		Logger:      logger,
	}

	// Set up transport (currently only stdio is supported)
	if !*stdio {
		fmt.Fprintf(os.Stderr, "Error: Only stdio transport is currently supported\n")
		os.Exit(1)
	}

	// Create server with transport
	transport := protocol.NewStdioTransport(os.Stdin, os.Stdout)
	srv := server.NewServerWithTransport(transport)

	// We need to add a way to apply options to an existing server
	// For now, create a new server with options and set the transport
	srv = server.NewServerWithOptions(opts)
	srv.SetTransport(transport)

	logger.Printf("Starting Carrion LSP server version %s", version)
	if *carrionPath != "" {
		logger.Printf("Using Carrion installation at: %s", *carrionPath)
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Printf("Received signal %v, shutting down gracefully...", sig)
		cancel()
	}()

	// Start the server loop
	if err := runServer(ctx, srv, logger); err != nil {
		logger.Printf("Server error: %v", err)
		os.Exit(1)
	}

	logger.Printf("Server shut down successfully")
}

// runServer runs the main server loop
func runServer(ctx context.Context, srv *server.Server, logger *log.Logger) error {
	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown requested
			logger.Printf("Context cancelled, shutting down...")
			return nil
		default:
			// Process a single request
			if err := srv.ProcessRequest(ctx); err != nil {
				// Check if the server has exited
				if srv.IsExited() {
					logger.Printf("Server exited normally")
					return nil
				}

				// Log the error but continue processing (unless it's a fatal error)
				logger.Printf("Request processing error: %v", err)

				// If it's a transport error (client disconnected), exit gracefully
				if isTransportError(err) {
					logger.Printf("Transport error detected, shutting down")
					return nil
				}

				// For other errors, continue processing
				continue
			}
		}
	}
}

// isTransportError checks if an error is related to transport issues
func isTransportError(err error) bool {
	// Common transport errors that indicate client disconnection
	errStr := err.Error()
	return errStr == "EOF" ||
		errStr == "io: read/write on closed pipe" ||
		errStr == "use of closed network connection" ||
		contains(errStr, "broken pipe") ||
		contains(errStr, "connection reset")
}

// contains checks if a string contains a substring (helper function)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					hasSubstring(s, substr)))
}

// hasSubstring checks if string contains substring anywhere
func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
