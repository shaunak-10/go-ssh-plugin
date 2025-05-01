// cmd/ssh-plugin/main.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"ssh-plugin/internal/config"
	"ssh-plugin/internal/metrics"
	"ssh-plugin/internal/models"
	"ssh-plugin/internal/reachability"
)

func main() {
	// Parse command line arguments
	cfg, err := config.ParseArgs(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	// Read input from stdin
	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Process based on mode
	var result interface{}
	switch cfg.Mode {
	case config.ModeReachability:
		var devices []models.DiscoveryDevice
		if err := json.Unmarshal(inputBytes, &devices); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON for discovery devices: %v\n", err)
			os.Exit(1)
		}
		result = reachability.CheckAll(devices, cfg.Timeout, cfg.Concurrency)

	case config.ModeMetrics:
		var devices []models.ProvisionDevice
		if err := json.Unmarshal(inputBytes, &devices); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON for provision devices: %v\n", err)
			os.Exit(1)
		}
		result = metrics.CollectAll(devices, cfg.Timeout, cfg.Concurrency)

	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", cfg.Mode)
		os.Exit(1)
	}

	// Output the result as JSON
	outputJSON(result)
}

// outputJSON marshals and prints the result to stdout
func outputJSON(data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}
