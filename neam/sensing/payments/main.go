package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/neamplatform/sensing/payments"
	"go.uber.org/zap"
)

func main() {
	// Determine config path
	configPath := "config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run the adapter
	if err := payments.Run(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Payment adapter failed: %v\n", err)
		os.Exit(1)
	}

	// Handle shutdown signal
	<-sigChan
	fmt.Println("Shutdown signal received, exiting...")
}

// Import for logging setup
import "go.uber.org/zap"
