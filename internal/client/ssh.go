package client

import (
	"golang.org/x/crypto/ssh"
)

// NewSSHClient dials an SSH server at the given TCP address using the provided
// client configuration and returns an established SSH client.
//
// The addr argument should be in the form "host:port", as expected by
// ssh.Dial (for example, "example.com:22").
//
// The config argument must be a fully initialized *ssh.ClientConfig, including
// authentication methods and any required HostKeyCallback or other options.
func NewSSHClient(addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	return ssh.Dial("tcp", addr, config)
}
