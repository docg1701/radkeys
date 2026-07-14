package ui

import (
	"os/exec"
	"testing"
)

func TestExecActionRunsCommand(t *testing.T) {
	cmd := exec.Command("echo", "radkeys-test")
	if err := cmd.Start(); err != nil {
		t.Fatalf("exec.Start failed: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("exec.Wait failed: %v", err)
	}
}

func TestExecActionNonExistent(t *testing.T) {
	cmd := exec.Command("nonexistent_command_xyzzy")
	if err := cmd.Start(); err == nil {
		// The process might have started but Wait will fail.
		if err := cmd.Wait(); err == nil {
			t.Error("expected error from non-existent command, got nil")
		}
	}
}
