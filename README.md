# podman-cli

A Go CLI tool for executing Podman API commands on a remote host via SSH. This tool connects to remote Podman instances by tunneling HTTP requests through SSH to the Podman Unix socket.

## Overview

`podman-cli` provides a command-line interface to interact with remote Podman instances through their HTTP API over SSH tunnels. Unlike traditional SSH command execution, this tool establishes an SSH connection, tunnels to the remote Podman Unix socket (`/run/user/1000/podman/podman.sock`), and sends HTTP requests directly to the Podman API.

This architecture provides:
- Direct API access without shell interpretation
- Structured JSON responses from Podman
- Better security (no shell injection risks)
- More reliable command execution

## Features

- **SSH Config Integration**: Uses `~/.ssh/config` for host configuration
- **HTTP over SSH Tunneling**: Secure communication with Podman API
- **Private Key Authentication**: Supports id_ed25519 and id_rsa keys
- **Known Hosts Verification**: Validates remote host keys using `~/.ssh/known_hosts`
- **Configurable Timeouts**: Set SSH connection timeouts
- **Secure by Default**: Host key verification enabled by default
- **Comprehensive Test Suite**: 32 tests with high coverage

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
podman-cli --host <hostname> [OPTIONS] <command>
```

### Global Options

- `--host <name>`: SSH host from your config file (required)
- `--timeout <duration>`: SSH connection timeout (default: 30s)
- `--no-host-validation`: Skip SSH known_hosts verification (not recommended)

### Available Commands

Currently supported commands:
- `list_containers`: List all containers (equivalent to `GET /v3.0.0/containers/json`)

### Examples

```bash
# List containers on remote host
podman-cli --host myserver list_containers

# Same command with explicit SSH config host
podman-cli --host production list_containers

# With custom timeout
podman-cli --host myserver --timeout 60s list_containers

# Skip host key verification (not recommended)
podman-cli --host myserver --no-host-validation list_containers
```

### Sample Output

```bash
$ podman-cli --hostuses private key authentication:

1. **Private Keys**: Uses the identity file specified in SSH config
   - Defaults to `~/.ssh/id_ed25519` if not configured
   - Supports RSA and Ed25519 keys
2. **Password Authentication**: Not supported

### Known Hosts Verification

By default, the tool verifies SSH host keys using `~/.ssh/known_hosts`. To skip this verification (not recommended for production):

```bash
podman-cli --host myserver --no-host-validation list_container

### Known Hosts Verification
Architecture

### HTTP over SSH Tunneling

The tool uses a sophisticated tunneling approach:

1. **SSH Connection**: Establishes secure SSH connection to remote host
2. **Unix Socket Tunneling**: Tunnels through SSH to Podman Unix socket
3. **HTTP Communication**: Sends HTTP requests directly to Podman API
4. **Response Handling**: Receives and displays JSON responses

This avoids shell interpretation and provides direct API access.

### SSH Configuration Parser

The tool reads SSH configuration from `~/.ssh/config` using the `github.com/kevinburke/ssh_config` library. It supports standard SSH config directives:
- `HostName`: The actual hostname or IP to connect to
- `User`: Remote username (defaults to `$USER`)
- `Port`: SSH port (defaults to 22)
- `IdentityFile`: Private key path (supports `~` expansion)

### SSH Connection

Uses `golang.org/x/crypto/ssh` to establish secure SSH connections. The implementation:
- Reads and parses private keys (RSA, Ed25519)
- Configures host key verification via known_hosts
- Supports configurable connection timeouts
- Uses secure cryptographic defaults

### Command Registry

Commands are defined in `internal/commands/commands.go` and map to Podman API endpoints:
```go
type Command struct {
    Path   string // API endpoint path
   Requirements

- **Remote Host**: 
  - Podman installed and running
  - Unix socket accessible at `/run/user/1000/podman/podman.sock`
  - SSH server running
- **Local Host**:
  - SSH config entry in `~/.ssh/config`
  - Private key for authentication
  - Go 1.24+ (for building from source)

## Limitations

- Currently only supports `list_containers` command (easily extensible)
- Requires Podman Unix socket to be accessible
- Only supports private key authentication (no password auth)
- Hardcoded socket path (`/run/useargument parsing & execution
│   │   ├── cli.go
│   │   └── cli_test.go
│   ├── client/             # SSH client & configuration
│   │   ├── config.go
│   │   ├── config_test.go
│   │   ├── ssh.go
│   │   └── ssh_test.go
│   └── commands/           # Command registry
│       ├── commands.go
│       └── commands_test.go
├── bin/                    # Compiled binaries
├── .vscode/
│   └── launch.json         # Debug configuration
├── Makefile                # Build & test autom
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
│     tests using Makefile
make test

# Build the project
make build
```

### Test Coverage

Comprehensive test suite with 32 tests:
- `internal/commands`: 100.0% coverage (8 tests)
- `internal/client`: 83.7% coverage (15 tests)
- `internal/cli`: 44.1% coverage (9 tests)
and cryptographic operations
- **github.com/kevinburke/ssh_config**: SSH config file parsing

All dependencies are actively maintained and use secure defaults.

## Design Philosophy

- **SSH Config Integration**: Leverages existing SSH infrastructure
- **API-First**: Direct HTTP communication with Podman API
- **Security First**: Host key verification enabled by default
- **SSH Tunneling**: Secure transport without shell risks
- **Testable Code**: Comprehensive test suite with 83%+ coverage
- **Documented Code**: GoDoc comments on all exported types and functions
- **Type Safety**: Strongly typed command definitions
- **Extensibility**: Easy to add new commands to registry
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
Adding New Commands

To add a new Podman API command:

1. Add the command to `internal/commands/commands.go`:
```go
var commands = map[string]Command{
    "list_containers": {
        Path:   "/v3.0.0/containers/json",
        Method: "GET",
    },
    "inspect_container": {
        Path:   "/v3.0.0/containers/{id}/json",
        Method: "GET",
    },
}
```

2. Add tests in `internal/commands/comman list_containers)
```

### Building

```bash
# Build using Makefile (recommended)
make build

# Build for current platform
go build -o bin/podman-cli ./cmd/podman-cli

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/podman-cli-linux ./cmd/podman-cli

# Build for macOS
GOOS=darwin GOARCH=arm64 go build -o bin/podman-cli-darwin ./cmd/podman-cli
```

### Code Organization
Troubleshooting

### Connection Issues

```bash
# Verify SSH connection works
ssh your-host

# Check if Podman socket exists
ssh your-host ls -la /run/user/1000/podman/podman.sock

# Test with host key verification disabled (debugging only)
podman-cli --host your-host --no-host-validation list_containers
```

### Common Errors

- **"dial remote socket: dial unix..."**: Podman socket not accessible
- **"failed to initialize CLI: open .ssh/config..."**: SSH config file missing
- **"invalid command"**: Command not in registry (only `list_containers` supported)

## See Also

- [Podman API Documentation](https://docs.podman.io/en/latest/_static/api.html)- **cmd/podman-cli**: Entry point, minimal logic
- **internal/cli**: Argument parsing, validation, execution flow
- **internal/client**: SSH configuration and connection handling
- **internal/commands**: Command registry and definitions
- All packages have comprehensive GoDoc documentation
- All packages have test coverage
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
