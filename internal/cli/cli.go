package cli

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/alexjch/podman-cli/internal/client"
	"github.com/alexjch/podman-cli/internal/config"
)

type RemoteCLI struct {
	host     string
	timeout  time.Duration
	insecure bool
}

func NewRemoteCLI() *RemoteCLI {
	return &RemoteCLI{}
}

func (rc *RemoteCLI) Run(args []string) int {

	flag.StringVar(&rc.host, "host", "", "Host to connect")
	flag.DurationVar(&rc.timeout, "timeout", 30*time.Second, "Command execution timeout")
	flag.BoolVar(&rc.insecure, "no-host-validation", false, "Do not verify host")

	flag.Parse()

	cmd := []string{"podman"}
	cmd = append(cmd, flag.Args()...)

	if rc.host == "" {
		log.Println("Argument host is required")
		return 1
	}

	sshConfig, err := config.NewSSHConfig(rc.host)
	if err != nil {
		log.Printf("Failed while reading config file: %s", err.Error())
		return 1
	}

	sshClientConfig, err := sshConfig.SSHClientConfig()
	if err != nil {
		log.Printf("Failed while configuring ssh connection: %s", err.Error())
		return 1
	}

	sshClient, err := client.NewSSHClient(sshConfig.Addr(), sshClientConfig)
	if err != nil {
		fmt.Printf("%+v", sshClientConfig)
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

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	cmdStr := strings.Join(cmd, " ")
	if err := session.Run(cmdStr); err != nil {
		log.Printf("Failed to run: %s", err.Error())
		return 1
	}
	fmt.Println(b.String())

	return 0
}
