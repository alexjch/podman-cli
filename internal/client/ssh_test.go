package client

import (
	"crypto/rand"
	"crypto/rsa"
	"net"
	"testing"

	"golang.org/x/crypto/ssh"
)

// setupTestSSHServer creates a test SSH server for testing
func setupTestSSHServer(t *testing.T) (net.Listener, *ssh.ServerConfig, string) {
	// Generate a test host key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Configure the SSH server
	config := &ssh.ServerConfig{
		NoClientAuth: true, // Accept all connections for testing
	}
	config.AddHostKey(signer)

	// Listen on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	return listener, config, listener.Addr().String()
}

// startTestSSHServer starts an SSH server that accepts one connection
func startTestSSHServer(t *testing.T, listener net.Listener, config *ssh.ServerConfig) {
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Perform SSH handshake
		_, _, _, err = ssh.NewServerConn(conn, config)
		if err != nil {
			// Connection will fail after handshake since we don't handle requests
			return
		}
	}()
}

func TestNewSSHClient_Success(t *testing.T) {
	listener, serverConfig, addr := setupTestSSHServer(t)
	defer listener.Close()

	startTestSSHServer(t, listener, serverConfig)

	// Create client config
	clientConfig := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpass"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Test connection
	client, err := NewSSHClient(addr, clientConfig)
	if err != nil {
		t.Fatalf("NewSSHClient() unexpected error = %v", err)
	}

	if client == nil {
		t.Fatal("NewSSHClient() returned nil client")
	}

	// Clean up
	client.Close()
}

func TestNewSSHClient_InvalidAddress(t *testing.T) {
	clientConfig := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpass"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Try to connect to an invalid address
	_, err := NewSSHClient("invalid:99999", clientConfig)
	if err == nil {
		t.Error("NewSSHClient() expected error for invalid address, got nil")
	}
}

func TestNewSSHClient_ConnectionRefused(t *testing.T) {
	clientConfig := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpass"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Try to connect to a port that's not listening
	_, err := NewSSHClient("127.0.0.1:19999", clientConfig)
	if err == nil {
		t.Error("NewSSHClient() expected error for connection refused, got nil")
	}
}

func TestNewSSHClient_HostKeyVerification(t *testing.T) {
	listener, serverConfig, addr := setupTestSSHServer(t)
	defer listener.Close()

	startTestSSHServer(t, listener, serverConfig)

	// Generate a different key for verification (will fail)
	wrongKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate wrong key: %v", err)
	}

	wrongSigner, err := ssh.NewSignerFromKey(wrongKey)
	if err != nil {
		t.Fatalf("Failed to create wrong signer: %v", err)
	}

	// Create client config with strict host key checking
	clientConfig := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpass"),
		},
		HostKeyCallback: ssh.FixedHostKey(wrongSigner.PublicKey()),
	}

	// This should fail due to host key mismatch
	_, err = NewSSHClient(addr, clientConfig)
	if err == nil {
		t.Error("NewSSHClient() expected error for host key mismatch, got nil")
	}
}

func TestNewSSHClient_WithPublicKeyAuth(t *testing.T) {
	listener, serverConfig, addr := setupTestSSHServer(t)
	defer listener.Close()

	// Generate client key
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate client key: %v", err)
	}

	clientSigner, err := ssh.NewSignerFromKey(clientKey)
	if err != nil {
		t.Fatalf("Failed to create client signer: %v", err)
	}

	// Update server config to require public key auth
	serverConfig.PublicKeyCallback = func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		// Accept any key for testing
		return nil, nil
	}
	serverConfig.NoClientAuth = false

	startTestSSHServer(t, listener, serverConfig)

	// Create client config with public key auth
	clientConfig := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(clientSigner),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Test connection
	client, err := NewSSHClient(addr, clientConfig)
	if err != nil {
		t.Fatalf("NewSSHClient() unexpected error = %v", err)
	}

	if client == nil {
		t.Fatal("NewSSHClient() returned nil client")
	}

	client.Close()
}

func TestNewSSHClient_EmptyAddress(t *testing.T) {
	clientConfig := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpass"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Try to connect with empty address
	_, err := NewSSHClient("", clientConfig)
	if err == nil {
		t.Error("NewSSHClient() expected error for empty address, got nil")
	}
}

func TestNewSSHClient_IPv6Address(t *testing.T) {
	// Listen on IPv6 localhost
	listener, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Err != nil {
			t.Skip("IPv6 not supported on this system")
		}
		t.Fatalf("Failed to listen on IPv6: %v", err)
	}
	defer listener.Close()

	// Generate host key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	serverConfig := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	serverConfig.AddHostKey(signer)

	startTestSSHServer(t, listener, serverConfig)

	clientConfig := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpass"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Extract IPv6 address
	addr := listener.Addr().String()

	// Test connection
	client, err := NewSSHClient(addr, clientConfig)
	if err != nil {
		t.Fatalf("NewSSHClient() unexpected error for IPv6 = %v", err)
	}

	if client == nil {
		t.Fatal("NewSSHClient() returned nil client for IPv6")
	}

	client.Close()
}

func TestNewSSHClient_MultipleAuthMethods(t *testing.T) {
	listener, serverConfig, addr := setupTestSSHServer(t)
	defer listener.Close()

	// Generate client key
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate client key: %v", err)
	}

	clientSigner, err := ssh.NewSignerFromKey(clientKey)
	if err != nil {
		t.Fatalf("Failed to create client signer: %v", err)
	}

	startTestSSHServer(t, listener, serverConfig)

	// Create client config with multiple auth methods
	clientConfig := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpass"),
			ssh.PublicKeys(clientSigner),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Test connection
	client, err := NewSSHClient(addr, clientConfig)
	if err != nil {
		t.Fatalf("NewSSHClient() unexpected error = %v", err)
	}

	if client == nil {
		t.Fatal("NewSSHClient() returned nil client")
	}

	client.Close()
}

func TestNewSSHClient_CustomPort(t *testing.T) {
	listener, serverConfig, _ := setupTestSSHServer(t)
	defer listener.Close()

	startTestSSHServer(t, listener, serverConfig)

	clientConfig := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpass"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Get the actual address with custom port
	addr := listener.Addr().String()

	// Test connection
	client, err := NewSSHClient(addr, clientConfig)
	if err != nil {
		t.Fatalf("NewSSHClient() unexpected error = %v", err)
	}

	if client == nil {
		t.Fatal("NewSSHClient() returned nil client")
	}

	client.Close()
}
