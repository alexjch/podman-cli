package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	HostName     string
	User         string
	IdentityFile string
	Port         int
}

func (c *SSHConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.HostName, c.Port)
}

func (c *SSHConfig) SSHClientConfig() (*ssh.ClientConfig, error) {

	key, err := os.ReadFile(c.IdentityFile)
	if err != nil {
		return nil, err
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return config, nil
}

func NewSSHConfig(host string) (*SSHConfig, error) {

	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "config"))
	if err != nil {
		return nil, err
	}
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
		idFile = filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519")
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

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	return &SSHConfig{
		HostName:     hostName,
		Port:         portInt,
		User:         user,
		IdentityFile: idFile,
	}, nil
}
