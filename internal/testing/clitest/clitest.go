package clitest

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	mcli "github.com/oxisto/money-gopher/cli"

	"github.com/urfave/cli/v3"
)

type CommandRecorder struct {
	*bytes.Buffer
}

// NewSessionContext creates a new context with a [cli.Session] attached to it.
// The session is connected to the given server.
func NewSessionContext(t *testing.T, srv *httptest.Server) context.Context {
	s := mcli.NewSession(&mcli.SessionOptions{
		BaseURL:    srv.URL,
		HttpClient: srv.Client(),
	})

	return context.WithValue(context.Background(), mcli.SessionKey, s)
}

// NewCommandRecorder creates a new command recorder.
func NewCommandRecorder() *CommandRecorder {
	return &CommandRecorder{Buffer: new(bytes.Buffer)}
}

// Record creates a new command recorder and sets it as the writer for the given
// command.
func Record(cli *cli.Command) *CommandRecorder {
	r := NewCommandRecorder()
	cli.Writer = r
	return r
}

// MockCommand creates a mock command with the given flags and parses the
// supplied arguments.
func MockCommand(t *testing.T, flags []cli.Flag, args ...string) *cli.Command {
	t.Helper()

	// Create a new empty command that we run to parse the flags, but we
	// copy the flags from the real command.
	cmd := &cli.Command{
		Name:  "mock",
		Flags: flags,
	}

	args = append([]string{"mock"}, args...)

	if err := cmd.Run(context.Background(), args); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	return cmd
}
