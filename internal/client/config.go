// Package client provides SSH client configuration and connection management
// for connecting to remote Podman instances. It handles SSH config file parsing,
// authentication, and host key verification.
package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// UserConfig holds the SSH configuration for connecting to a remote host.
// It stores credentials, connection details, and paths to SSH files.
type UserConfig struct {
	user         string
	port         string
	hostName     string
	knownHosts   string
	identityFile string
}

// sshUserFilePath constructs an absolute path to a file in the user's .ssh directory.
func sshUserFilePath(fileName string) string {
	return filepath.Join(os.Getenv("HOME"), ".ssh", fileName)
}

// NewSSHClientConfig creates an SSH client configuration from user config.
// It reads the identity file, sets up authentication, and configures host key verification.
//
// Parameters:
//   - timeout: SSH connection timeout duration
//   - insecure: if true, skips host key verification (not recommended for production)
//   - userConfig: user configuration containing SSH details
//
// Returns an ssh.ClientConfig ready for establishing connections.
func NewSSHClientConfig(timeout time.Duration, insecure bool, userConfig *UserConfig) (*ssh.ClientConfig, error) {

	var hostKeyCallback ssh.HostKeyCallback

	key, err := os.ReadFile(userConfig.identityFile)
	if err != nil {
		return nil, err
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	if insecure {
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		hostKeyCallback, err = knownhosts.New(userConfig.knownHosts)
		if err != nil {
			return nil, err
		}
	}

	clientConfig := &ssh.ClientConfig{
		User: userConfig.user,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         timeout,
	}

	return clientConfig, nil
}

// Addr returns the SSH server address in "host:port" format.
func (uc *UserConfig) Addr() string {
	return fmt.Sprintf("%s:%s", uc.hostName, uc.port)
}

// NewUserConfig reads SSH configuration from ~/.ssh/config and creates a UserConfig.
// It parses the SSH config file for the specified host and applies defaults for
// missing values (port 22, current user, id_ed25519 key).
//
// The function respects standard SSH config directives including:
//   - HostName: the actual hostname or IP to connect to
//   - Port: SSH port (defaults to 22)
//   - User: username for authentication (defaults to current USER)
//   - IdentityFile: path to private key (defaults to ~/.ssh/id_ed25519)
//
// Returns an error if the config file cannot be read or parsed.
func NewUserConfig(host string) (*UserConfig, error) {

	file, err := os.Open(sshUserFilePath("config"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	conf, err := ssh_config.Decode(file)
	if err != nil {
		return nil, err
	}

	hostName, err := conf.Get(host, "HostName")
	if err != nil {
		return nil, err
	}

	// Default to the provided host if HostName is empty
	if hostName == "" {
		hostName = host
	}

	// User
	user, err := conf.Get(host, "User")
	if err != nil {
		return nil, err
	}

	// Default to current user
	if user == "" {
		user = os.Getenv("USER")
	}

	// Identity file
	idFile, err := conf.Get(host, "IdentityFile")
	if err != nil {
		return nil, err
	}

	if idFile == "" {
		idFile = sshUserFilePath("id_ed25519")
	} else if strings.HasPrefix(idFile, "~/") {
		// Expand tilde to HOME directory
		idFile = filepath.Join(os.Getenv("HOME"), idFile[2:])
	}

	port, err := conf.Get(host, "Port")
	if err != nil {
		return nil, err
	}

	// Default to port 22 if empty
	if port == "" {
		port = "22"
	}

	knownHostsFile := sshUserFilePath("known_hosts")

	userConfig := &UserConfig{
		user:         user,
		port:         port,
		hostName:     hostName,
		knownHosts:   knownHostsFile,
		identityFile: idFile,
	}

	return userConfig, nil
}
