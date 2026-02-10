package client

import (
	"golang.org/x/crypto/ssh"
)

func NewSSHClient(addr string, config *ssh.ClientConfig) (*ssh.Client, error) {

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
