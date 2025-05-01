// internal/metrics/collector.go
package metrics

import (
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

// CollectAll collects system metrics from all devices concurrently
func CollectAll(devices []models.ProvisionDevice, timeout time.Duration, concurrency int) []models.MetricsResult {
	results := make([]models.MetricsResult, len(devices))

	// Create a worker pool
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	for i, device := range devices {
		wg.Add(1)

		go func(i int, device models.ProvisionDevice) {
			defer wg.Done()

			// Acquire a token from the semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			client := ssh.NewClientFromDiscovery(device.IP, device.Port, device.Username, device.Password, timeout)

			cpu, memory, disk := collectDeviceMetrics(client)

			results[i] = models.MetricsResult{
				ProvisionID: device.ProvisionID,
				CPU:         cpu,
				Memory:      memory,
				DiskUsage:   disk,
			}
		}(i, device)
	}

	wg.Wait()
	return results
}

// collectDeviceMetrics retrieves metrics from a device over SSH
func collectDeviceMetrics(client *ssh.Client) (cpu, memory, disk string) {
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

	return
}
