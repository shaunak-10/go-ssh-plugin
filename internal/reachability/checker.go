package reachability

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"ssh-plugin/internal/security"
	"sync"
	"time"

	"ssh-plugin/internal/models"
	"ssh-plugin/internal/ssh"
)

// CheckAll checks SSH connectivity for all devices and sends results through a channel as they are completed
func CheckAll(devices []models.DiscoveryDevice, timeout time.Duration, concurrency int) {

	// Create a channel to receive results as they're completed
	resultsChan := make(chan models.ReachabilityResult, len(devices))

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
			encryptedOutput, err := security.Encrypt(outputJSON)

			if err != nil {

				log.Printf("Error encrypting output: %v\n", err)

				continue
			}

			fmt.Println(encryptedOutput)

			err = os.Stdout.Sync()

			if err != nil {

				log.Printf("Error syncing output: %v\n", err)
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
						device.DiscoveryID, device.IP, r)

					// Send a failure result for this device
					resultsChan <- models.ReachabilityResult{

						DiscoveryID: device.DiscoveryID,
						Reachable:   false,
						Error:       fmt.Sprintf("Internal error: %v", r),
					}
				}
			}()

			if device.SystemType != "linux" {

				resultsChan <- models.ReachabilityResult{

					DiscoveryID: device.DiscoveryID,
					Reachable:   false,
					Error:       "Unsupported system type",
				}

				return
			}

			// Acquire a token from the semaphore
			semaphore <- struct{}{}

			// Ensure the token is released even in case of panic
			defer func() { <-semaphore }()

			// Create an SSH client and check connection
			client := ssh.NewClientFromDiscovery(device.IP, device.Port, device.Username, device.Password, timeout)

			reachable, err := client.CheckConnection()

			// Create the result object
			result := models.ReachabilityResult{

				DiscoveryID: device.DiscoveryID,

				Reachable: reachable,
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
