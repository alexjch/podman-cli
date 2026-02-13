package cli

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTestSSHConfig(t *testing.T, tmpDir string) string {
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configData := `Host testhost
  HostName test.example.com
  Port 22
  User testuser
  IdentityFile ~/.ssh/id_rsa
`
	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Generate a test RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	keyFile := filepath.Join(sshDir, "id_rsa")
	if err := os.WriteFile(keyFile, privateKeyPEM, 0600); err != nil {
		t.Fatalf("Failed to write test key file: %v", err)
	}

	// Create an empty known_hosts file
	knownHostsFile := filepath.Join(sshDir, "known_hosts")
	if err := os.WriteFile(knownHostsFile, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to write known_hosts file: %v", err)
	}

	return tmpDir
}

func TestNewRemoteCLI_ValidArgs(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestSSHConfig(t, tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	args := []string{"-host", "testhost", "list_containers"}
	cli, err := NewRemoteCLI(args)
	if err != nil {
		t.Fatalf("NewRemoteCLI() unexpected error = %v", err)
	}

	if cli == nil {
		t.Fatal("NewRemoteCLI() returned nil")
	}

	if cli.addr != "test.example.com:22" {
		t.Errorf("NewRemoteCLI() addr = %q, want %q", cli.addr, "test.example.com:22")
	}

	if cli.command.Path != "/v3.0.0/containers/json" {
		t.Errorf("NewRemoteCLI() command.Path = %q, want %q", cli.command.Path, "/v3.0.0/containers/json")
	}

	if cli.command.Method != "GET" {
		t.Errorf("NewRemoteCLI() command.Method = %q, want %q", cli.command.Method, "GET")
	}

	if cli.sshClientConfig == nil {
		t.Error("NewRemoteCLI() sshClientConfig is nil")
	}
}

func TestNewRemoteCLI_MissingHost(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestSSHConfig(t, tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	args := []string{"list_containers"}
	_, err := NewRemoteCLI(args)
	if err == nil {
		t.Error("NewRemoteCLI() expected error for missing host, got nil")
	}

	if !strings.Contains(err.Error(), "host") {
		t.Errorf("NewRemoteCLI() error = %v, want error containing 'host'", err)
	}
}

func TestNewRemoteCLI_MissingCommand(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestSSHConfig(t, tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	args := []string{"-host", "testhost"}
	_, err := NewRemoteCLI(args)
	if err == nil {
		t.Error("NewRemoteCLI() expected error for missing command, got nil")
	}

	if !strings.Contains(err.Error(), "command") {
		t.Errorf("NewRemoteCLI() error = %v, want error containing 'command'", err)
	}
}

func TestNewRemoteCLI_InvalidCommand(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestSSHConfig(t, tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	args := []string{"-host", "testhost", "invalid_command"}
	_, err := NewRemoteCLI(args)
	if err == nil {
		t.Error("NewRemoteCLI() expected error for invalid command, got nil")
	}

	if !strings.Contains(err.Error(), "invalid command") {
		t.Errorf("NewRemoteCLI() error = %v, want error containing 'invalid command'", err)
	}
}

func TestNewRemoteCLI_CustomTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestSSHConfig(t, tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	args := []string{"-host", "testhost", "-timeout", "60s", "list_containers"}
	cli, err := NewRemoteCLI(args)
	if err != nil {
		t.Fatalf("NewRemoteCLI() unexpected error = %v", err)
	}

	if cli.sshClientConfig.Timeout != 60*time.Second {
		t.Errorf("NewRemoteCLI() timeout = %v, want %v", cli.sshClientConfig.Timeout, 60*time.Second)
	}
}

func TestNewRemoteCLI_NoHostValidation(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestSSHConfig(t, tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	args := []string{"-host", "testhost", "-no-host-validation", "list_containers"}
	cli, err := NewRemoteCLI(args)
	if err != nil {
		t.Fatalf("NewRemoteCLI() unexpected error = %v", err)
	}

	if cli == nil {
		t.Fatal("NewRemoteCLI() returned nil")
	}

	// We can't directly test the insecure flag, but we can verify the CLI was created
	if cli.sshClientConfig == nil {
		t.Error("NewRemoteCLI() sshClientConfig is nil")
	}
}

func TestNewRemoteCLI_InvalidFlagFormat(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestSSHConfig(t, tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	args := []string{"-host", "testhost", "-timeout", "invalid", "list_containers"}
	_, err := NewRemoteCLI(args)
	if err == nil {
		t.Error("NewRemoteCLI() expected error for invalid timeout format, got nil")
	}
}

func TestNewRemoteCLI_NoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	args := []string{"-host", "testhost", "list_containers"}
	_, err := NewRemoteCLI(args)
	if err == nil {
		t.Error("NewRemoteCLI() expected error when config file doesn't exist, got nil")
	}
}

func TestRemoteCLI_Struct(t *testing.T) {
	cli := &RemoteCLI{
		addr: "test.example.com:22",
	}

	if cli.addr != "test.example.com:22" {
		t.Errorf("RemoteCLI.addr = %q, want %q", cli.addr, "test.example.com:22")
	}
}
