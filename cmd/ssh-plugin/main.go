// main.go
package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"ssh-plugin/internal/config"
	"ssh-plugin/internal/metrics"
	"ssh-plugin/internal/models"
	"ssh-plugin/internal/reachability"
	"ssh-plugin/internal/security"
)

func main() {

	// Set up panic recovery for the main goroutine
	defer func() {

		if r := recover(); r != nil {

			log.Printf("Recovered from panic in main: %v\n", r)

			os.Exit(1)
		}

	}()

	// Load encryption key
	if err := security.LoadKeyFromFile("/home/shaunak/IdeaProjects/http-server/src/main/java/org/example/plugin_executable/go_config.json"); err != nil {

		log.Printf("Failed to load encryption key: %v\n", err)

		os.Exit(1)
	}

	// Parse configuration
	cfg, err := config.ParseArgs(os.Args)

	if err != nil {

		log.Printf("Error parsing arguments: %v\n", err)

		os.Exit(1)
	}

	// Read input data from stdin
	encryptedInput, err := io.ReadAll(os.Stdin)

	if err != nil {

		log.Printf("Error reading input: %v\n", err)

		os.Exit(1)
	}

	inputData, err := security.Decrypt(string(encryptedInput))

	if err != nil {

		log.Printf("Error decrypting input: %v\n", err)

		os.Exit(1)
	}

	// Process based on mode
	switch cfg.Mode {

	case config.ModeReachability:

		// Parse devices for reachability check
		var devices []models.DiscoveryDevice

		if err := json.Unmarshal(inputData, &devices); err != nil {

			log.Printf("Error parsing input JSON: %v\n", err)

			os.Exit(1)
		}

		// Run reachability checks
		reachability.CheckAll(devices, cfg.Timeout, cfg.Concurrency)

	case config.ModeMetrics:

		// Parse devices for metrics collection
		var devices []models.ProvisionDevice

		if err := json.Unmarshal(inputData, &devices); err != nil {

			log.Printf("Error parsing input JSON: %v\n", err)

			os.Exit(1)
		}

		// Collect metrics
		metrics.CollectAll(devices, cfg.Timeout, cfg.Concurrency)

	default:

		log.Printf("Unknown mode: %s\n", cfg.Mode)

		os.Exit(1)
	}
}
