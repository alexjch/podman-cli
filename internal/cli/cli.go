package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/alexjch/podman-cli/internal/client"
	"github.com/alexjch/podman-cli/internal/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type RemoteCLI struct {
	addr         string
	command      string
	clientConfig *ssh.ClientConfig
}

func NewRemoteCLI(args []string) (*RemoteCLI, error) {
	var host string
	var timeout time.Duration
	var insecure bool

	fs := flag.NewFlagSet("remote-cli", flag.ContinueOnError)

	fs.StringVar(&host, "host", "", "Host to connect")
	fs.DurationVar(&timeout, "timeout", 30*time.Second, "Command execution timeout")
	fs.BoolVar(&insecure, "no-host-validation", false, "Do not verify host")

	if err := fs.Parse(args); err != nil {
		log.Printf("Failed to parse arguments: %s", err.Error())
		return nil, err
	}

	cmd := []string{"podman"}

	// Building a remote command by strings.Join(cmd, \" \") is unsafe and can be
	// incorrect when arguments contain spaces/shell metacharacters; it can also
	// enable shell injection depending on how the remote executes the command.
	// before joining, or otherwise avoid shell interpretation.
	cmdStr := strings.Join(append(cmd, fs.Args()...), " ")

	if host == "" {
		fs.PrintDefaults()
		return nil, errors.New("Flag -host is required. Use -host to specify the remote host.")
	}

	sshConfig, err := config.NewSSHConfig(host)
	if err != nil {
		return nil, err
	}

	clientConfig, err := sshConfig.SSHClientConfig(timeout, insecure)
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%d", sshConfig.HostName, sshConfig.Port)

	return &RemoteCLI{
		addr:         addr,
		command:      cmdStr,
		clientConfig: clientConfig,
	}, nil
}

// The argument parsing/validation and remote command construction have multiple branches
// worth testing (e.g., missing --host, flag parse errors, and args that include
// spaces/metacharacters to ensure correct escaping/quoting). Add unit tests around Run
// to lock in expected exit codes and constructed command behavior.
func (rc *RemoteCLI) Run() int {

	sshClient, err := client.NewSSHClient(rc.addr, rc.clientConfig)
	if err != nil {

		log.Printf("Failed while connecting to client: %s", err.Error())
		return 1
	}
	defer sshClient.Close()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := sshClient.NewSession()
	if err != nil {
		log.Printf("Failed to create session: %s", err)
		return 1
	}
	defer session.Close()

	// Set up terminal modes
	fd := int(os.Stdin.Fd())
	isTerminal := term.IsTerminal(fd)

	if isTerminal {
		// Request a pseudo-terminal
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			log.Printf("Failed to set raw mode: %s", err.Error())
			return 1
		}
		defer term.Restore(fd, oldState)

		// Get terminal size
		width, height, err := term.GetSize(fd)
		if err != nil {
			width, height = 80, 24 // Default size
		}

		// Request PTY
		if err := session.RequestPty("xterm-256color", height, width, nil); err != nil {
			log.Printf("Failed to request pty: %s", err.Error())
			return 1
		}
	}

	// Connect stdin, stdout, and stderr
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		log.Printf("Failed to get stdin pipe: %s", err.Error())
		return 1
	}

	// Build and execute command
	if err := session.Start(rc.command); err != nil {
		log.Printf("Failed to start command: %s", err.Error())
		return 1
	}

	// Copy stdin to the remote session
	go func() {
		io.Copy(stdinPipe, os.Stdin)
		stdinPipe.Close()
	}()

	// Wait for the command to complete
	if err := session.Wait(); err != nil {
		if exitErr, ok := err.(*os.SyscallError); ok {
			log.Printf("Command failed: %s", exitErr.Error())
			return 1
		}
		// Non-zero exit codes are also returned as errors
		return 1
	}

	return 0
}
