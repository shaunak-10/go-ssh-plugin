package metrics

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"ssh-plugin/internal/crypto"
	"strings"
	"sync"
	"time"

	"ssh-plugin/internal/models"
	"ssh-plugin/internal/ssh"
)

// Commands for collecting metrics
const (
	cpuCommand    = "top -bn1 | grep 'Cpu(s)' | awk '{print $2 + $4 \"%\"}'"
	memoryCommand = "free -h | grep Mem | awk '{print $3}'"
	diskCommand   = "df -h / | tail -1 | awk '{print $3}'"
)

// CollectAll collects system metrics from all devices and sends results through a channel as they are completed
func CollectAll(devices []models.ProvisionDevice, timeout time.Duration, concurrency int) {
	// Create a channel to receive results as they're completed
	resultsChan := make(chan models.MetricsResult, len(devices))

	// Create a WaitGroup to know when all goroutines are done
	var wg sync.WaitGroup

	// Create a WaitGroup for the result processor goroutine
	var outputWg sync.WaitGroup
	outputWg.Add(1)

	// Create a semaphore channel for concurrency control
	semaphore := make(chan struct{}, concurrency)

	// Start a goroutine to handle results as they come in
	go func() {
		// Make sure we signal when done processing all results
		defer outputWg.Done()

		for result := range resultsChan {
			// Convert result to JSON and print to stdout
			outputJSON, err := json.Marshal(result)
			if err != nil {
				log.Printf("Error marshaling result: %v\n", err)
				continue
			}

			// Write the result to stdout immediately and flush
			encryptedOutput, err := crypto.Encrypt(outputJSON)
			if err != nil {
				log.Printf("Error encrypting output: %v\n", err)
				continue
			}
			fmt.Println(encryptedOutput)
			err = os.Stdout.Sync()
			if err != nil {
				log.Printf("Error syncing stdout: %v\n", err)
			} // Force flush stdout
		}
	}()

	// Process each device
	for i := range devices {
		device := devices[i] // Create a copy of device for the goroutine
		wg.Add(1)

		// Process each device in a separate goroutine
		go func() {
			// Use defer to ensure the WaitGroup counter is decremented even if panic occurs
			defer wg.Done()

			// Recover from any panics that might occur during processing
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered from panic processing device %d (%s): %v\n",
						device.ProvisionID, device.IP, r)

					// Send a failure result for this device
					resultsChan <- models.MetricsResult{
						ProvisionID: device.ProvisionID,
						CPU:         "ERROR",
						Memory:      "ERROR",
						DiskUsage:   "ERROR",
						Error:       fmt.Sprintf("Internal error: %v", r),
					}
				}
			}()

			// Acquire a token from the semaphore
			semaphore <- struct{}{}

			// Ensure the token is released even in case of panic
			defer func() { <-semaphore }()

			// Create an SSH client for this device
			client := ssh.NewClientFromDiscovery(device.IP, device.Port, device.Username, device.Password, timeout)

			// Collect metrics
			cpu, memory, disk, err := collectDeviceMetrics(client)

			// Create the result object
			result := models.MetricsResult{
				ProvisionID: device.ProvisionID,
				CPU:         cpu,
				Memory:      memory,
				DiskUsage:   disk,
			}

			// Add error information if there was an error
			if err != nil {
				result.Error = err.Error()
			}

			// Send the result through the channel
			resultsChan <- result
		}()
	}

	// Wait for all device processing goroutines to complete
	wg.Wait()

	// Close the results channel to signal we're done sending results
	close(resultsChan)

	// Wait for all results to be processed and output
	outputWg.Wait()
}

// collectDeviceMetrics retrieves metrics from a device over SSH
func collectDeviceMetrics(client *ssh.Client) (cpu, memory, disk string, err error) {
	// Default values in case of connection failure
	cpu = "N/A"
	memory = "N/A"
	disk = "N/A"

	commands := []string{cpuCommand, memoryCommand, diskCommand}

	output, err := client.ExecuteCommand(commands)
	if err == nil {
		// Split the output by lines, assuming each command output is on a new line
		lines := strings.Split(output, "\n")

		// Ensure we have the expected number of results
		if len(lines) >= 3 {
			cpu = lines[0]    // First line is CPU usage
			memory = lines[1] // Second line is Memory usage
			disk = lines[2]   // Third line is Disk usage
		}
	}

	return cpu, memory, disk, err
}
