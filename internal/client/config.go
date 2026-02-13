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

type UserConfig struct {
	user         string
	port         string
	hostName     string
	knownHosts   string
	identityFile string
}

func sshUserFilePath(fileName string) string {
	return filepath.Join(os.Getenv("HOME"), ".ssh", fileName)
}

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

func (uc *UserConfig) Addr() string {
	return fmt.Sprintf("%s:%s", uc.hostName, uc.port)
}

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
