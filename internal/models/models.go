// internal/models/models.go
package models

// DiscoveryDevice represents a device for SSH reachability check
type DiscoveryDevice struct {
	DiscoveryID int    `json:"id"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

// ProvisionDevice represents a device for metrics collection
type ProvisionDevice struct {
	ProvisionID int    `json:"id"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

// ReachabilityResult represents the output for SSH reachability check
type ReachabilityResult struct {
	DiscoveryID int  `json:"id"`
	Reachable   bool `json:"reachable"`
}

// MetricsResult represents the output for metrics polling
type MetricsResult struct {
	ProvisionID int    `json:"id"`
	CPU         string `json:"cpu"`
	Memory      string `json:"memory"`
	DiskUsage   string `json:"disk-usage"`
}
