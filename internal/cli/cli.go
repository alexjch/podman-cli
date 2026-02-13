package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/alexjch/podman-cli/internal/client"
	"github.com/alexjch/podman-cli/internal/commands"
	"golang.org/x/crypto/ssh"
)

type RemoteCLI struct {
	addr            string
	command         commands.Command
	sshClientConfig *ssh.ClientConfig
}

func NewRemoteCLI(args []string) (*RemoteCLI, error) {

	var host string
	var timeout time.Duration
	var insecure bool

	fs := flag.NewFlagSet("remote-cli", flag.ContinueOnError)

	fs.StringVar(&host, "host", "", "Host to connect")
	fs.DurationVar(&timeout, "timeout", 30*time.Second, "SSH connection timeout")
	fs.BoolVar(&insecure, "no-host-validation", false, "Do not verify host")

	if err := fs.Parse(args); err != nil {
		log.Printf("Failed to parse arguments: %v", err)
		return nil, err
	}

	if fs.NArg() < 1 {
		return nil, fmt.Errorf("at least one command must be provided")
	}

	if host == "" {
		fs.PrintDefaults()
		return nil, errors.New("-host is required (use -host to specify the remote host)")
	}

	cmds := fs.Args()
	command := commands.IsCommand(cmds[0])
	if command == nil {
		return nil, fmt.Errorf("invalid command: %s", cmds[0])
	}

	userConfig, err := client.NewUserConfig(host)
	if err != nil {
		return nil, err
	}

	sshClientConfig, err := client.NewSSHClientConfig(timeout, insecure, userConfig)
	if err != nil {
		return nil, err
	}

	cli := &RemoteCLI{
		addr:            userConfig.Addr(),
		command:         *command,
		sshClientConfig: sshClientConfig,
	}

	return cli, nil
}

func (rc *RemoteCLI) Run() int {

	// SSH Connection / Tunel
	sshClient, err := client.NewSSHClient(rc.addr, rc.sshClientConfig)
	if err != nil {
		log.Printf("Failed while connecting to client: %v", err)
		return 1
	}
	defer sshClient.Close()

	// Socket connections / tuneled connection
	remoteSocket := "/run/user/1000/podman/podman.sock"
	conn, err := sshClient.Dial("unix", remoteSocket)
	if err != nil {
		log.Printf("dial remote socket: %v", err)
		return 1
	}
	defer conn.Close()

	// Build the request (use a dummy host; Host header required by http.ReadResponse)
	u := &url.URL{Scheme: "http", Host: "localhost", Path: rc.command.Path}
	req := &http.Request{
		Method: rc.command.Method,
		URL:    u,
		Host:   "localhost",
		Header: make(http.Header),
	}

	// Write request to the connection
	if err := req.Write(conn); err != nil {
		log.Printf("Error with request: %v\n", err)
		return 1
	}

	// Read response
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		log.Printf("Error with response: %s\n", err)
		return 1
	}
	defer resp.Body.Close()

	// Print status and body
	fmt.Println("Status:", resp.Status)
	body := new(strings.Builder)
	_, err = bufio.NewReader(resp.Body).WriteTo(body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read body: %v\n", err)
		return 1
	}
	fmt.Println(body.String())

	// Use HTTP status code to determine exit code: non-2xx => failure
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return 1
	}
	return 0
}
