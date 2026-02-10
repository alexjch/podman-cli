# podman-cli

A Go CLI tool for executing Podman commands on a remote host via SSH. The same as using "ssh <args> <user>:<host> podman <podman args>"

## Overview

`podman-cli` provides a simple command-line interface to run Podman commands remotely over SSH. It uses your existing SSH configuration (`~/.ssh/config`) to connect to remote hosts and execute Podman commands.

## Features

- **SSH Config Integration**: Uses `~/.ssh/config` for host configuration
- **Multiple Authentication**: Supports SSH agent and private keys (id_ed25519, id_rsa)
- **Known Hosts Verification**: Validates remote host keys using `~/.ssh/known_hosts`
- **Configurable Timeouts**: Set command execution timeouts
- **Secure by Default**: Host key verification enabled by default

## Installation

```bash
go install github.com/alexjch/podman-cli/cmd/podman-cli@latest
```

Or build from source:

```bash
git clone https://github.com/alexjch/podman-cli
cd podman-cli
go build -o bin/podman-cli ./cmd/podman-cli
```

## SSH Configuration

Configure your remote hosts in `~/.ssh/config`:

```
Host myserver
  HostName example.com
  User myuser
  Port 22
  IdentityFile ~/.ssh/id_rsa

Host production
  HostName prod.example.com
  User admin
  Port 2222
  IdentityFile ~/.ssh/id_ed25519
```

## Usage

```bash
podman-cli [OPTIONS] --host <hostname> [PODMAN_COMMAND...]
```

### Global Options

- `--host <name>`: SSH host from your config file (required)
- `--timeout <duration>`: Command execution timeout (default: 30s)
- `--no-host-validation`: Skip SSH known_hosts verification (not recommended)

### Examples

```bash
# List containers on remote host
podman-cli --host myserver ps -a

# Check Podman version
podman-cli --host myserver version

# Run container on remote host
podman-cli --host myserver run -d nginx

# Pull image on remote host
podman-cli --host production pull docker.io/library/alpine

# View system info
podman-cli --host myserver info

# With custom timeout
podman-cli --host myserver --timeout 60s ps -a
```

## Authentication

The SSH connection supports multiple authentication methods:

1. **Private Keys**: Uses the identity file specified in SSH config
   - Defaults to `~/.ssh/id_ed25519` if not configured
   - Falls back to `~/.ssh/id_rsa`
2. **SSH Agent**: Future support planned

### Known Hosts Verification

By default, the tool verifies SSH host keys using `~/.ssh/known_hosts`. To skip this verification (not recommended):

```bash
podman-cli --no-host-validation --host myserver ps
```

## Implementation Details

### SSH Configuration Parser

The tool reads SSH configuration from `~/.ssh/config` using the `github.com/kevinburke/ssh_config` library. It supports standard SSH config directives:
- `HostName`: The actual hostname or IP to connect to
- `User`: Remote username (defaults to `$USER`)
- `Port`: SSH port (defaults to 22)
- `IdentityFile`: Private key path (supports `~` expansion)

### SSH Connection

Uses `golang.org/x/crypto/ssh` to establish secure SSH connections. The implementation:
- Reads and parses private keys (RSA, Ed25519)
- Supports modern cryptographic algorithms (SHA-2 family)
- Disables insecure SHA-1 algorithms by default
- Uses `ssh.SupportedAlgorithms()` for secure defaults

### Command Execution

Once connected, the tool creates an SSH session and executes the Podman command directly on the remote host. Output is captured and displayed locally.

## Supported Platforms

- Linux
- macOS
- Any platform with SSH access and Podman installed on the remote host

## Limitations

- Requires Podman to be installed and accessible on the remote host
- Requires SSH access to the remote host
- Requires SSH config entry for the target host in `~/.ssh/config`
- Currently only supports private key authentication (password auth not supported)
- Command output is displayed after completion (no streaming)

## Project Structure

```
podman-cli/
├── cmd/
│   └── podman-cli/         # Main CLI entry point
│       └── main.go
├── internal/
│   ├── cli/                # CLI command handling
│   │   └── cli.go
│   ├── client/             # SSH client implementation
│   │   └── ssh.go
│   ├── config/             # SSH config parser
│   │   ├── config.go
│   │   └── config_test.go
│   └── uri/                # URI parser (legacy)
│       ├── parser.go
│       └── parser_test.go
├── .vscode/
│   └── launch.json         # Debug configuration
├── go.mod
├── go.sum
└── README.md
```

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/config/...

# Run config tests with coverage
go test ./internal/config/... -cover
```

Current test coverage:
- `internal/config`: 83.3%
- `internal/uri`: Comprehensive test suite

## Dependencies

- **golang.org/x/crypto/ssh**: SSH client functionality and cryptographic operations
- **github.com/kevinburke/ssh_config**: SSH config file parsing

### Development Dependencies

- **crypto/rsa**: RSA key handling (test key generation)
- **crypto/x509**: X.509 encoding for private keys

All cryptographic operations use modern, secure algorithms and disable insecure SHA-1 based methods by default.

## Design Philosophy

- **SSH Config Integration**: Leverages existing SSH configuration instead of custom formats
- **Standard Go Libraries**: Uses Go standard and extended libraries wherever possible
- **Security First**: Host key verification enabled by default, modern cryptography only
- **Simple Architecture**: Direct SSH command execution (no complex REST API tunneling)
- **Testable Code**: Comprehensive test coverage with isolated test environments

## Contributing

Contributions are welcome! Please ensure:
- Code follows Go best practices and conventions
- All tests pass (`go test ./...`)
- New features include tests
- Test coverage is maintained or improved
- Documentation is updated
- Run `go fmt` before committing

## Development

### Debug Configuration

A VS Code launch configuration is included in `.vscode/launch.json`:

```bash
# Debug with specific host
F5 in VS Code (uses --host fedora-console)
```

### Building

```bash
# Build for current platform
go build -o bin/podman-cli ./cmd/podman-cli

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/podman-cli-linux ./cmd/podman-cli

# Build for macOS
GOOS=darwin GOARCH=arm64 go build -o bin/podman-cli-darwin ./cmd/podman-cli
```

## License

[Specify your license here]

## See Also

- [Podman Documentation](https://docs.podman.io/)
- [OpenSSH Config File](https://man.openbsd.org/ssh_config)
- [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh)
- [SSH Config Parser](https://github.com/kevinburke/ssh_config)
