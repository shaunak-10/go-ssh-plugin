package config

import (
	"flag"
	"log"
	"time"
)

// Operation modes
const (
	ModeReachability = "reachability"
	ModeMetrics      = "metrics"
)

// Config holds the application configuration
type Config struct {
	Mode        string
	Timeout     time.Duration
	Concurrency int
}

// ParseArgs parses command line arguments and returns a Config
func ParseArgs(args []string) (*Config, error) {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic while parsing args: %v", r)
		}
	}()

	if len(args) < 2 {
		return &Config{
			Mode:        ModeReachability,
			Timeout:     3 * time.Second,
			Concurrency: 10,
		}, nil
	}

	// Define flags
	flagSet := flag.NewFlagSet(args[0], flag.ExitOnError)
	timeout := flagSet.Int("timeout", 3, "Connection timeout in seconds")
	concurrency := flagSet.Int("concurrency", 10, "Maximum concurrent connections")

	// Parse the flags
	if err := flagSet.Parse(args[2:]); err != nil {
		return nil, err
	}

	return &Config{
		Mode:        args[1],
		Timeout:     time.Duration(*timeout) * time.Second,
		Concurrency: *concurrency,
	}, nil
}
