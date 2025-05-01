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
	}
}

// CheckConnection tests if a connection can be established
func (c *Client) CheckConnection() bool {
	client, err := ssh.Dial("tcp", c.addr, c.config)
	if err != nil {
		return false
	}
	defer client.Close()
	return true
}

// ExecuteCommand runs a command on the remote server
func (c *Client) ExecuteCommand(commands []string) (string, error) {
	client, err := ssh.Dial("tcp", c.addr, c.config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	// Join all commands into one string, separated by semicolons
	cmd := strings.Join(commands, "; ")

	// Execute the commands
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
