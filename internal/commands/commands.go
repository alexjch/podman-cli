// Package commands defines the available Podman API commands that can be
// executed through the CLI. Each command maps to a specific Podman API endpoint.
package commands

// Command represents a Podman API endpoint with its HTTP method and path.
type Command struct {
	Path   string // API endpoint path (e.g., "/v3.0.0/containers/json")
	Method string // HTTP method (e.g., "GET", "POST")
}

// commands is the internal registry of available commands.
var commands = map[string]Command{
	"list_containers": {
		Path:   "/v3.0.0/containers/json",
		Method: "GET",
	},
}

// Commands returns a copy of all available commands.
// This prevents external modification of the internal command registry.
func Commands() map[string]Command {
	copy := make(map[string]Command, len(commands))
	for k, v := range commands {
		copy[k] = v
	}
	return copy
}

// IsCommand checks if the given command name exists and returns its definition.
// Returns nil if the command is not found.
func IsCommand(cmd string) *Command {
	command, ok := commands[cmd]
	if !ok {
		return nil
	}
	return &command
}
