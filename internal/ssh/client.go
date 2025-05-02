// internal/ssh/client.go
package ssh

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client is a wrapper around the SSH client
type Client struct {
	config *ssh.ClientConfig
	addr   string
	ip     string
}

// NewClientFromDiscovery creates a new SSH client from a discovery device
func NewClientFromDiscovery(ip string, port int, username, password string, timeout time.Duration) *Client {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	addr := fmt.Sprintf("%s:%d", ip, port)

	return &Client{
		config: config,
		addr:   addr,
		ip:     ip,
	}
}

// CheckConnection tests if a connection can be established
func (c *Client) CheckConnection() (bool, error) {
	client, err := ssh.Dial("tcp", c.addr, c.config)
	if err != nil {
		return false, fmt.Errorf("connection failed to %s: %w", c.ip, err)
	}

	// Make sure we close the connection even if we panic later
	defer client.Close()

	return true, nil
}

// ExecuteCommand runs a command on the remote server
func (c *Client) ExecuteCommand(commands []string) (string, error) {
	// Establish connection
	client, err := ssh.Dial("tcp", c.addr, c.config)
	if err != nil {
		return "", fmt.Errorf("failed to dial %s: %w", c.ip, err)
	}

	// Ensure the client is closed when we're done
	defer client.Close()

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session on %s: %w", c.ip, err)
	}

	// Ensure the session is closed when we're done
	defer session.Close()

	// Join all commands into one string, separated by semicolons
	cmd := strings.Join(commands, "; ")

	// Execute the commands
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("command execution failed on %s: %w", c.ip, err)
	}

	return string(output), nil
}
