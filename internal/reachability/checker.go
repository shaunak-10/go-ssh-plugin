// internal/reachability/checker.go
package reachability

import (
	"sync"
	"time"

	"ssh-plugin/internal/models"
	"ssh-plugin/internal/ssh"
)

// CheckAll checks SSH connectivity for all devices concurrently
func CheckAll(devices []models.DiscoveryDevice, timeout time.Duration, concurrency int) []models.ReachabilityResult {
	results := make([]models.ReachabilityResult, len(devices))

	// Create a worker pool
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	for i, device := range devices {
		wg.Add(1)

		go func(i int, device models.DiscoveryDevice) {
			defer wg.Done()

			// Acquire a token from the semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			client := ssh.NewClientFromDiscovery(device.IP, device.Port, device.Username, device.Password, timeout)
			reachable := client.CheckConnection()

			results[i] = models.ReachabilityResult{
				DiscoveryID: device.DiscoveryID,
				Reachable:   reachable,
			}
		}(i, device)
	}

	wg.Wait()
	return results
}
