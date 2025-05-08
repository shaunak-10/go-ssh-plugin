package ssh

import (
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client is a wrapper around the SSH client
type Client struct {
	config  *ssh.ClientConfig
	address string
	ip      string
}

// NewClientFromDiscovery creates a new SSH client from a discovery device
func NewClientFromDiscovery(ip string, port int, username, password string, timeout time.Duration) *Client {

	config := &ssh.ClientConfig{

		User: username,

		Auth: []ssh.AuthMethod{

			ssh.Password(password),
		},

		HostKeyCallback: ssh.InsecureIgnoreHostKey(),

		Timeout: timeout,
	}

	addr := fmt.Sprintf("%s:%d", ip, port)

	return &Client{

		config:  config,
		address: addr,
		ip:      ip,
	}
}

// CheckConnection tests if a connection can be established
func (c *Client) CheckConnection() (ok bool, err error) {

	defer func() {

		if r := recover(); r != nil {

			log.Printf("Recovered from panic in CheckConnection (IP: %s): %v", c.ip, r)

			ok = false

			err = fmt.Errorf("panic occurred during SSH connection check")
		}
	}()

	client, err := ssh.Dial("tcp", c.address, c.config)

	if err != nil {

		return false, fmt.Errorf("connection failed to %s: %w", c.ip, err)
	}

	// Make sure we close the connection even if we panic later
	defer func(client *ssh.Client) {

		err := client.Close()

		if err != nil {

			log.Printf("Error closing connection to %s: %v\n", c.address, err)
		}

	}(client)

	return true, nil
}

// ExecuteCommand runs a command on the remote server
func (c *Client) ExecuteCommand(commands []string) (outputStr string, err error) {

	defer func() {

		if r := recover(); r != nil {

			log.Printf("Recovered from panic in ExecuteCommand (IP: %s): %v", c.ip, r)

			outputStr = "N/A"

			err = fmt.Errorf("panic occurred during SSH command execution")
		}
	}()

	// Establish connection
	client, err := ssh.Dial("tcp", c.address, c.config)

	if err != nil {

		return "", fmt.Errorf("failed to dial %s: %w", c.ip, err)
	}

	// Ensure the client is closed when we're done
	defer func(client *ssh.Client) {

		err := client.Close()

		if err != nil {

			log.Printf("Error closing SSH client: %v\n", err)
		}
	}(client)

	// Create a session
	session, err := client.NewSession()

	if err != nil {

		return "", fmt.Errorf("failed to create session on %s: %w", c.ip, err)
	}

	// Ensure the session is closed when we're done
	defer func(session *ssh.Session) {

		err := session.Close()

		if err != nil {

			log.Printf("Error closing SSH session: %v\n", err)
		}
	}(session)

	// Join all commands into one string, separated by semicolons
	cmd := strings.Join(commands, "; ")

	// Execute the commands
	output, err := session.CombinedOutput(cmd)

	if err != nil {

		return "", fmt.Errorf("command execution failed on %s: %w", c.ip, err)
	}

	return string(output), nil
}
