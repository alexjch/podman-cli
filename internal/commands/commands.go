package commands

type Command struct {
	Path   string
	Method string
}

var commands = map[string]Command{
	"list_containers": {
		Path:   "/v3.0.0/containers/json",
		Method: "GET",
	},
}

func Commands() map[string]Command {
	copy := make(map[string]Command, len(commands))
	for k, v := range commands {
		copy[k] = v
	}
	return copy
}

func IsCommand(cmd string) *Command {
	command, ok := commands[cmd]
	if !ok {
		return nil
	}
	return &command
}
