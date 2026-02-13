package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUserConfig_Addr(t *testing.T) {
	tests := []struct {
		name     string
		config   UserConfig
		expected string
	}{
		{
			name: "standard port",
			config: UserConfig{
				hostName: "example.com",
				port:     "22",
			},
			expected: "example.com:22",
		},
		{
			name: "custom port",
			config: UserConfig{
				hostName: "test.example.com",
				port:     "2222",
			},
			expected: "test.example.com:2222",
		},
		{
			name: "IPv4 address",
			config: UserConfig{
				hostName: "192.168.1.100",
				port:     "22",
			},
			expected: "192.168.1.100:22",
		},
		{
			name: "IPv6 address",
			config: UserConfig{
				hostName: "2001:db8::1",
				port:     "22",
			},
			expected: "2001:db8::1:22",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.Addr()
			if got != tt.expected {
				t.Errorf("Addr() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNewUserConfig_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configData := `Host myserver
  HostName 192.168.1.100
  Port 2222
  User admin
  IdentityFile ~/.ssh/id_rsa
`
	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	got, err := NewUserConfig("myserver")
	if err != nil {
		t.Fatalf("NewUserConfig() unexpected error = %v", err)
	}

	if got.hostName != "192.168.1.100" {
		t.Errorf("NewUserConfig() hostName = %q, want %q", got.hostName, "192.168.1.100")
	}

	if got.port != "2222" {
		t.Errorf("NewUserConfig() port = %q, want %q", got.port, "2222")
	}

	if got.user != "admin" {
		t.Errorf("NewUserConfig() user = %q, want %q", got.user, "admin")
	}

	wantIdentityFile := filepath.Join(tmpDir, ".ssh", "id_rsa")
	if got.identityFile != wantIdentityFile {
		t.Errorf("NewUserConfig() identityFile = %q, want %q", got.identityFile, wantIdentityFile)
	}
}

func TestNewUserConfig_DefaultValues(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configData := `Host webserver
  HostName example.com
`
	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	oldUser := os.Getenv("USER")
	os.Setenv("HOME", tmpDir)
	os.Setenv("USER", "testuser")
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("USER", oldUser)
	}()

	got, err := NewUserConfig("webserver")
	if err != nil {
		t.Fatalf("NewUserConfig() unexpected error = %v", err)
	}

	if got.hostName != "example.com" {
		t.Errorf("NewUserConfig() hostName = %q, want %q", got.hostName, "example.com")
	}

	if got.port != "22" {
		t.Errorf("NewUserConfig() port = %q, want %q (default)", got.port, "22")
	}

	if got.user != "testuser" {
		t.Errorf("NewUserConfig() user = %q, want %q (default)", got.user, "testuser")
	}

	wantIdentityFile := filepath.Join(tmpDir, ".ssh", "id_ed25519")
	if got.identityFile != wantIdentityFile {
		t.Errorf("NewUserConfig() identityFile = %q, want %q (default)", got.identityFile, wantIdentityFile)
	}
}

func TestNewUserConfig_MissingHost(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configData := `Host otherhost
  HostName other.example.com
`
	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	got, err := NewUserConfig("missinghost")
	if err != nil {
		t.Fatalf("NewUserConfig() unexpected error = %v", err)
	}

	// Should default to the host name provided
	if got.hostName != "missinghost" {
		t.Errorf("NewUserConfig() hostName = %q, want %q (default)", got.hostName, "missinghost")
	}

	if got.port != "22" {
		t.Errorf("NewUserConfig() port = %q, want %q (default)", got.port, "22")
	}
}

func TestNewUserConfig_NoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	_, err := NewUserConfig("anyhost")
	if err == nil {
		t.Error("NewUserConfig() expected error when config file doesn't exist, got nil")
	}
}

func TestNewSSHClientConfig_Insecure(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	// Generate a test RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Encode private key to PEM format
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	keyFile := filepath.Join(sshDir, "id_rsa")
	if err := os.WriteFile(keyFile, privateKeyPEM, 0600); err != nil {
		t.Fatalf("Failed to write test key file: %v", err)
	}

	userConfig := &UserConfig{
		user:         "testuser",
		port:         "22",
		hostName:     "test.example.com",
		knownHosts:   filepath.Join(sshDir, "known_hosts"),
		identityFile: keyFile,
	}

	timeout := 30 * time.Second
	insecure := true
	clientConfig, err := NewSSHClientConfig(timeout, insecure, userConfig)
	if err != nil {
		t.Fatalf("NewSSHClientConfig() unexpected error = %v", err)
	}

	if clientConfig == nil {
		t.Fatal("NewSSHClientConfig() returned nil")
	}

	if clientConfig.User != userConfig.user {
		t.Errorf("NewSSHClientConfig().User = %q, want %q", clientConfig.User, userConfig.user)
	}

	if len(clientConfig.Auth) == 0 {
		t.Error("NewSSHClientConfig().Auth is empty, expected at least one auth method")
	}

	if clientConfig.HostKeyCallback == nil {
		t.Error("NewSSHClientConfig().HostKeyCallback is nil")
	}

	if clientConfig.Timeout != timeout {
		t.Errorf("NewSSHClientConfig().Timeout = %v, want %v", clientConfig.Timeout, timeout)
	}
}

func TestNewSSHClientConfig_InvalidKeyFile(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	userConfig := &UserConfig{
		user:         "testuser",
		port:         "22",
		hostName:     "test.example.com",
		knownHosts:   filepath.Join(sshDir, "known_hosts"),
		identityFile: filepath.Join(sshDir, "nonexistent_key"),
	}

	timeout := 30 * time.Second
	_, err := NewSSHClientConfig(timeout, true, userConfig)
	if err == nil {
		t.Error("NewSSHClientConfig() expected error for nonexistent key file, got nil")
	}
}

func TestNewSSHClientConfig_InvalidKeyFormat(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	keyFile := filepath.Join(sshDir, "invalid_key")
	if err := os.WriteFile(keyFile, []byte("not a valid key"), 0600); err != nil {
		t.Fatalf("Failed to write invalid key file: %v", err)
	}

	userConfig := &UserConfig{
		user:         "testuser",
		port:         "22",
		hostName:     "test.example.com",
		knownHosts:   filepath.Join(sshDir, "known_hosts"),
		identityFile: keyFile,
	}

	timeout := 30 * time.Second
	_, err := NewSSHClientConfig(timeout, true, userConfig)
	if err == nil {
		t.Error("NewSSHClientConfig() expected error for invalid key format, got nil")
	}
}

func TestSshUserFilePath(t *testing.T) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", "/home/testuser")
	defer os.Setenv("HOME", oldHome)

	tests := []struct {
		name     string
		fileName string
		expected string
	}{
		{
			name:     "config file",
			fileName: "config",
			expected: "/home/testuser/.ssh/config",
		},
		{
			name:     "known_hosts file",
			fileName: "known_hosts",
			expected: "/home/testuser/.ssh/known_hosts",
		},
		{
			name:     "identity file",
			fileName: "id_rsa",
			expected: "/home/testuser/.ssh/id_rsa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sshUserFilePath(tt.fileName)
			if got != tt.expected {
				t.Errorf("sshUserFilePath(%q) = %q, want %q", tt.fileName, got, tt.expected)
			}
		})
	}
}
