package commands

import (
	"testing"
)

func TestIsCommand_ValidCommand(t *testing.T) {
	cmd := IsCommand("list_containers")
	if cmd == nil {
		t.Fatal("IsCommand() returned nil for valid command")
	}

	if cmd.Path != "/v3.0.0/containers/json" {
		t.Errorf("IsCommand() Path = %q, want %q", cmd.Path, "/v3.0.0/containers/json")
	}

	if cmd.Method != "GET" {
		t.Errorf("IsCommand() Method = %q, want %q", cmd.Method, "GET")
	}
}

func TestIsCommand_InvalidCommand(t *testing.T) {
	cmd := IsCommand("nonexistent_command")
	if cmd != nil {
		t.Errorf("IsCommand() returned %v for invalid command, want nil", cmd)
	}
}

func TestIsCommand_EmptyString(t *testing.T) {
	cmd := IsCommand("")
	if cmd != nil {
		t.Errorf("IsCommand() returned %v for empty string, want nil", cmd)
	}
}

func TestCommands_ReturnsCopy(t *testing.T) {
	cmds1 := Commands()
	cmds2 := Commands()

	if len(cmds1) != len(cmds2) {
		t.Errorf("Commands() lengths differ: %d vs %d", len(cmds1), len(cmds2))
	}

	// Verify it's actually a copy by modifying one
	cmds1["test"] = Command{Path: "/test", Method: "POST"}

	if _, exists := cmds2["test"]; exists {
		t.Error("Commands() did not return a copy, modifications affected other calls")
	}
}

func TestCommands_ContainsExpectedCommands(t *testing.T) {
	cmds := Commands()

	expectedCmd := Command{
		Path:   "/v3.0.0/containers/json",
		Method: "GET",
	}

	cmd, exists := cmds["list_containers"]
	if !exists {
		t.Fatal("Commands() missing 'list_containers' command")
	}

	if cmd.Path != expectedCmd.Path {
		t.Errorf("Commands()[list_containers].Path = %q, want %q", cmd.Path, expectedCmd.Path)
	}

	if cmd.Method != expectedCmd.Method {
		t.Errorf("Commands()[list_containers].Method = %q, want %q", cmd.Method, expectedCmd.Method)
	}
}

func TestCommand_StructFields(t *testing.T) {
	cmd := Command{
		Path:   "/test/path",
		Method: "POST",
	}

	if cmd.Path != "/test/path" {
		t.Errorf("Command.Path = %q, want %q", cmd.Path, "/test/path")
	}

	if cmd.Method != "POST" {
		t.Errorf("Command.Method = %q, want %q", cmd.Method, "POST")
	}
}

func TestIsCommand_ReturnsPointer(t *testing.T) {
	cmd1 := IsCommand("list_containers")
	cmd2 := IsCommand("list_containers")

	// Verify we get different pointers each time
	if cmd1 == cmd2 {
		t.Error("IsCommand() returns same pointer, expected different pointers")
	}

	// But they should have the same values
	if cmd1.Path != cmd2.Path || cmd1.Method != cmd2.Method {
		t.Error("IsCommand() returns different values for same command")
	}
}

func TestCommands_Length(t *testing.T) {
	cmds := Commands()
	if len(cmds) == 0 {
		t.Error("Commands() returned empty map")
	}

	// Verify it contains at least the known command
	if _, exists := cmds["list_containers"]; !exists {
		t.Error("Commands() missing list_containers")
	}
}
