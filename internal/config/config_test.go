package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewSSHConfig_ValidConfig(t *testing.T) {
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

	got, err := NewSSHConfig("myserver")
	if err != nil {
		t.Fatalf("NewSSHConfig() unexpected error = %v", err)
	}

	if got.HostName != "192.168.1.100" {
		t.Errorf("NewSSHConfig() HostName = %q, want %q", got.HostName, "192.168.1.100")
	}

	if got.Port != 2222 {
		t.Errorf("NewSSHConfig() Port = %d, want %d", got.Port, 2222)
	}

	if got.User != "admin" {
		t.Errorf("NewSSHConfig() User = %q, want %q", got.User, "admin")
	}

	wantIdentityFile := filepath.Join(tmpDir, ".ssh", "id_rsa")
	if got.IdentityFile != wantIdentityFile {
		t.Errorf("NewSSHConfig() IdentityFile = %q, want %q", got.IdentityFile, wantIdentityFile)
	}
}

func TestNewSSHConfig_DefaultPort(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configData := `Host webserver
  HostName example.com
  Port 22
`
	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	got, err := NewSSHConfig("webserver")
	if err != nil {
		t.Fatalf("NewSSHConfig() unexpected error = %v", err)
	}

	if got.HostName != "example.com" {
		t.Errorf("NewSSHConfig() HostName = %q, want %q", got.HostName, "example.com")
	}

	if got.Port != 22 {
		t.Errorf("NewSSHConfig() Port = %d, want %d", got.Port, 22)
	}

	currentUser := os.Getenv("USER")
	if got.User != currentUser {
		t.Errorf("NewSSHConfig() User = %q, want %q (default)", got.User, currentUser)
	}

	wantIdentityFile := filepath.Join(tmpDir, ".ssh", "id_ed25519")
	if got.IdentityFile != wantIdentityFile {
		t.Errorf("NewSSHConfig() IdentityFile = %q, want %q (default)", got.IdentityFile, wantIdentityFile)
	}
}

func TestNewSSHConfig_MultipleHosts(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configData := `Host server1
  HostName 10.0.0.1
  Port 22

Host server2
  HostName 10.0.0.2
  Port 3333
`
	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	got, err := NewSSHConfig("server2")
	if err != nil {
		t.Fatalf("NewSSHConfig() unexpected error = %v", err)
	}

	if got.HostName != "10.0.0.2" {
		t.Errorf("NewSSHConfig() HostName = %q, want %q", got.HostName, "10.0.0.2")
	}

	if got.Port != 3333 {
		t.Errorf("NewSSHConfig() Port = %d, want %d", got.Port, 3333)
	}

	currentUser := os.Getenv("USER")
	if got.User != currentUser {
		t.Errorf("NewSSHConfig() User = %q, want %q (default)", got.User, currentUser)
	}

	wantIdentityFile := filepath.Join(tmpDir, ".ssh", "id_ed25519")
	if got.IdentityFile != wantIdentityFile {
		t.Errorf("NewSSHConfig() IdentityFile = %q, want %q (default)", got.IdentityFile, wantIdentityFile)
	}
}

func TestNewSSHConfig_WithComments(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configData := `# Production servers
Host production
  # Main hostname
  HostName prod.example.com
  # Custom SSH port
  Port 8022
`
	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	got, err := NewSSHConfig("production")
	if err != nil {
		t.Fatalf("NewSSHConfig() unexpected error = %v", err)
	}

	if got.HostName != "prod.example.com" {
		t.Errorf("NewSSHConfig() HostName = %q, want %q", got.HostName, "prod.example.com")
	}

	if got.Port != 8022 {
		t.Errorf("NewSSHConfig() Port = %d, want %d", got.Port, 8022)
	}

	currentUser := os.Getenv("USER")
	if got.User != currentUser {
		t.Errorf("NewSSHConfig() User = %q, want %q (default)", got.User, currentUser)
	}

	wantIdentityFile := filepath.Join(tmpDir, ".ssh", "id_ed25519")
	if got.IdentityFile != wantIdentityFile {
		t.Errorf("NewSSHConfig() IdentityFile = %q, want %q (default)", got.IdentityFile, wantIdentityFile)
	}
}

func TestNewSSHConfig_MissingHost(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configData := `Host otherhost
  HostName other.example.com
  Port 22
`
	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	got, err := NewSSHConfig("missinghost")
	if err != nil {
		t.Fatalf("NewSSHConfig() unexpected error = %v", err)
	}

	if got.HostName != "missinghost" {
		t.Errorf("NewSSHConfig() HostName = %q, want %q", got.HostName, "missinghost")
	}

	if got.Port != 22 {
		t.Errorf("NewSSHConfig() Port = %d, want %d", got.Port, 22)
	}

	currentUser := os.Getenv("USER")
	if got.User != currentUser {
		t.Errorf("NewSSHConfig() User = %q, want %q (default)", got.User, currentUser)
	}

	wantIdentityFile := filepath.Join(tmpDir, ".ssh", "id_ed25519")
	if got.IdentityFile != wantIdentityFile {
		t.Errorf("NewSSHConfig() IdentityFile = %q, want %q (default)", got.IdentityFile, wantIdentityFile)
	}
}

func TestNewSSHConfig_InvalidPort(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configData := `Host badport
  HostName example.com
  Port notanumber
`
	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	_, err := NewSSHConfig("badport")
	if err == nil {
		t.Error("NewSSHConfig() expected error for invalid port, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "invalid syntax") {
		t.Errorf("NewSSHConfig() error = %v, want error containing 'invalid syntax'", err)
	}
}

func TestNewSSHConfig_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	got, err := NewSSHConfig("anyhost")
	if err != nil {
		t.Fatalf("NewSSHConfig() unexpected error = %v", err)
	}

	if got.HostName != "anyhost" {
		t.Errorf("NewSSHConfig() HostName = %q, want %q", got.HostName, "anyhost")
	}

	if got.Port != 22 {
		t.Errorf("NewSSHConfig() Port = %d, want %d", got.Port, 22)
	}

	currentUser := os.Getenv("USER")
	if got.User != currentUser {
		t.Errorf("NewSSHConfig() User = %q, want %q (default)", got.User, currentUser)
	}

	wantIdentityFile := filepath.Join(tmpDir, ".ssh", "id_ed25519")
	if got.IdentityFile != wantIdentityFile {
		t.Errorf("NewSSHConfig() IdentityFile = %q, want %q (default)", got.IdentityFile, wantIdentityFile)
	}
}

func TestNewSSHConfig_IPv6Address(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configData := `Host ipv6host
  HostName 2001:db8::1
  Port 2222
`
	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	got, err := NewSSHConfig("ipv6host")
	if err != nil {
		t.Fatalf("NewSSHConfig() unexpected error = %v", err)
	}

	if got.HostName != "2001:db8::1" {
		t.Errorf("NewSSHConfig() HostName = %q, want %q", got.HostName, "2001:db8::1")
	}

	if got.Port != 2222 {
		t.Errorf("NewSSHConfig() Port = %d, want %d", got.Port, 2222)
	}

	currentUser := os.Getenv("USER")
	if got.User != currentUser {
		t.Errorf("NewSSHConfig() User = %q, want %q (default)", got.User, currentUser)
	}

	wantIdentityFile := filepath.Join(tmpDir, ".ssh", "id_ed25519")
	if got.IdentityFile != wantIdentityFile {
		t.Errorf("NewSSHConfig() IdentityFile = %q, want %q (default)", got.IdentityFile, wantIdentityFile)
	}
}

func TestNewSSHConfig_NoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	_, err := NewSSHConfig("anyhost")
	if err == nil {
		t.Error("NewSSHConfig() expected error when config file doesn't exist, got nil")
	}
}

func TestNewSSHConfig_CustomUser(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configData := `Host customuser
  HostName user.example.com
  User deploybot
  Port 22
`
	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	got, err := NewSSHConfig("customuser")
	if err != nil {
		t.Fatalf("NewSSHConfig() unexpected error = %v", err)
	}

	if got.User != "deploybot" {
		t.Errorf("NewSSHConfig() User = %q, want %q", got.User, "deploybot")
	}

	wantIdentityFile := filepath.Join(tmpDir, ".ssh", "id_ed25519")
	if got.IdentityFile != wantIdentityFile {
		t.Errorf("NewSSHConfig() IdentityFile = %q, want %q (default)", got.IdentityFile, wantIdentityFile)
	}
}

func TestNewSSHConfig_CustomIdentityFile(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configData := `Host customkey
  HostName key.example.com
  IdentityFile ~/.ssh/custom_key
  Port 22
`
	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	got, err := NewSSHConfig("customkey")
	if err != nil {
		t.Fatalf("NewSSHConfig() unexpected error = %v", err)
	}

	wantIdentityFile := filepath.Join(tmpDir, ".ssh", "custom_key")
	if got.IdentityFile != wantIdentityFile {
		t.Errorf("NewSSHConfig() IdentityFile = %q, want %q", got.IdentityFile, wantIdentityFile)
	}

	currentUser := os.Getenv("USER")
	if got.User != currentUser {
		t.Errorf("NewSSHConfig() User = %q, want %q (default)", got.User, currentUser)
	}
}

func TestNewSSHConfig_AllCustomFields(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	configData := `Host fullyconfig
  HostName full.example.com
  User fulluser
  IdentityFile ~/.ssh/special_key
  Port 8888
`
	configFile := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configFile, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	got, err := NewSSHConfig("fullyconfig")
	if err != nil {
		t.Fatalf("NewSSHConfig() unexpected error = %v", err)
	}

	if got.HostName != "full.example.com" {
		t.Errorf("NewSSHConfig() HostName = %q, want %q", got.HostName, "full.example.com")
	}

	if got.Port != 8888 {
		t.Errorf("NewSSHConfig() Port = %d, want %d", got.Port, 8888)
	}

	if got.User != "fulluser" {
		t.Errorf("NewSSHConfig() User = %q, want %q", got.User, "fulluser")
	}

	wantIdentityFile := filepath.Join(tmpDir, ".ssh", "special_key")
	if got.IdentityFile != wantIdentityFile {
		t.Errorf("NewSSHConfig() IdentityFile = %q, want %q", got.IdentityFile, wantIdentityFile)
	}
}

func TestSSHConfig_Struct(t *testing.T) {
	config := &SSHConfig{
		HostName:     "test.example.com",
		Port:         2222,
		User:         "testuser",
		IdentityFile: "/home/testuser/.ssh/id_rsa",
	}

	if config.HostName != "test.example.com" {
		t.Errorf("SSHConfig.HostName = %q, want %q", config.HostName, "test.example.com")
	}

	if config.Port != 2222 {
		t.Errorf("SSHConfig.Port = %d, want %d", config.Port, 2222)
	}

	if config.User != "testuser" {
		t.Errorf("SSHConfig.User = %q, want %q", config.User, "testuser")
	}

	if config.IdentityFile != "/home/testuser/.ssh/id_rsa" {
		t.Errorf("SSHConfig.IdentityFile = %q, want %q", config.IdentityFile, "/home/testuser/.ssh/id_rsa")
	}
}

func TestSSHConfig_SSHClientConfig(t *testing.T) {
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

	config := &SSHConfig{
		HostName:     "test.example.com",
		Port:         22,
		User:         "testuser",
		IdentityFile: keyFile,
	}

	clientConfig, err := config.SSHClientConfig()
	if err != nil {
		t.Fatalf("SSHClientConfig() unexpected error = %v", err)
	}

	if clientConfig == nil {
		t.Fatal("SSHClientConfig() returned nil")
	}

	if clientConfig.User != config.User {
		t.Errorf("SSHClientConfig().User = %q, want %q", clientConfig.User, "username")
	}

	if len(clientConfig.Auth) == 0 {
		t.Error("SSHClientConfig().Auth is empty, expected at least one auth method")
	}

	if clientConfig.HostKeyCallback == nil {
		t.Error("SSHClientConfig().HostKeyCallback is nil")
	}
}
