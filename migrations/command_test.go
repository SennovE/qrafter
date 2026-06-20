package migrations

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunMigrationsCommandShowsTopLevelHelp(t *testing.T) {
	for _, args := range [][]string{
		nil,
		{"help"},
		{commandHelpLong},
		{"-h"},
	} {
		out, err := captureStdout(t, func() error {
			return RunMigrationsCommand(args)
		})
		if err != nil {
			t.Fatalf("RunMigrationsCommand(%v) error = %v", args, err)
		}
		for _, want := range []string{
			"Usage:",
			"qrafter-migrations <command> [flags]",
			"init",
			"revision",
			"up",
			"down",
		} {
			if !strings.Contains(out, want) {
				t.Fatalf("RunMigrationsCommand(%v) output does not contain %q:\n%s", args, want, out)
			}
		}
	}
}

func TestRunMigrationsCommandShowsSubcommandHelp(t *testing.T) {
	tests := []struct {
		command string
		want    []string
	}{
		{command: commandInit, want: []string{"qrafter-migrations init", "--driver-import", "--dsn"}},
		{command: commandRevision, want: []string{"qrafter-migrations revision", "--comment", "--timeout"}},
		{command: commandUp, want: []string{"qrafter-migrations up", "--to", "--version-table"}},
		{command: commandDown, want: []string{"qrafter-migrations down", "--to", "--version-table"}},
	}

	for _, tt := range tests {
		out, err := captureStdout(t, func() error {
			return RunMigrationsCommand([]string{"help", tt.command})
		})
		if err != nil {
			t.Fatalf("help %s error = %v", tt.command, err)
		}
		for _, want := range tt.want {
			if !strings.Contains(out, want) {
				t.Fatalf("help %s output does not contain %q:\n%s", tt.command, want, out)
			}
		}
	}
}

func TestRunMigrationsCommandFlagHelpIsNotAnError(t *testing.T) {
	for _, args := range [][]string{
		{commandInit, commandHelpLong},
		{commandRevision, commandHelpLong},
		{commandUp, commandHelpLong},
		{commandDown, commandHelpLong},
	} {
		if err := RunMigrationsCommand(args); err != nil {
			t.Fatalf("RunMigrationsCommand(%v) error = %v", args, err)
		}
	}
}

func TestRunMigrationsCommandUnknownHelpTarget(t *testing.T) {
	if err := RunMigrationsCommand([]string{"help", "missing"}); err == nil {
		t.Fatal("expected error for unknown help target")
	}
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = oldStdout
	}()

	fnErr := fn()
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}

	out, readErr := io.ReadAll(reader)
	if err := reader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	if readErr != nil {
		t.Fatalf("read stdout: %v", readErr)
	}
	return string(out), fnErr
}
