// main.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"ssh-plugin/internal/config"
	"ssh-plugin/internal/crypto"
	"ssh-plugin/internal/metrics"
	"ssh-plugin/internal/models"
	"ssh-plugin/internal/reachability"
)

func main() {

	// Load encryption key
	if err := crypto.LoadKeyFromFile("/home/shaunak/IdeaProjects/http-server/src/main/java/org/example/plugin_executable/go_config.json"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load encryption key: %v\n", err)
		os.Exit(1)
	}
	// Set up panic recovery for the main goroutine
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Recovered from panic in main: %v\n", r)
			os.Exit(1)
		}
	}()

	// Parse configuration
	cfg, err := config.ParseArgs(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	// Read input data from stdin
	encryptedInput, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	inputData, err := crypto.Decrypt(string(encryptedInput))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decrypting input: %v\n", err)
		os.Exit(1)
	}

	// Process based on mode
	switch cfg.Mode {
	case config.ModeReachability:
		// Parse devices for reachability check
		var devices []models.DiscoveryDevice
		if err := json.Unmarshal(inputData, &devices); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing input JSON: %v\n", err)
			os.Exit(1)
		}

		// Run reachability checks
		reachability.CheckAll(devices, cfg.Timeout, cfg.Concurrency)

	case config.ModeMetrics:
		// Parse devices for metrics collection
		var devices []models.ProvisionDevice
		if err := json.Unmarshal(inputData, &devices); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing input JSON: %v\n", err)
			os.Exit(1)
		}

		// Collect metrics
		metrics.CollectAll(devices, cfg.Timeout, cfg.Concurrency)

	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", cfg.Mode)
		os.Exit(1)
	}
}
