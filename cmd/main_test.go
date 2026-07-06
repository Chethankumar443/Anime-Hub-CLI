package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCheckDependencies_Missing(t *testing.T) {
	// Mock lookPath to return an error
	oldLookPath := lookPath
	oldFileExists := fileExists
	defer func() {
		lookPath = oldLookPath
		fileExists = oldFileExists
	}()

	lookPath = func(file string) (string, error) {
		return "", errors.New("not found")
	}
	fileExists = func(path string) bool {
		return false
	}

	err := checkDependencies()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errMsg := err.Error()
	// Check if the error message contains the installation commands
	if !strings.Contains(errMsg, "brew install mpv") {
		t.Error("expected error message to contain 'brew install mpv'")
	}
	if !strings.Contains(errMsg, "sudo apt install mpv") {
		t.Error("expected error message to contain 'sudo apt install mpv'")
	}
	if !strings.Contains(errMsg, "winget install mpv") {
		t.Error("expected error message to contain 'winget install mpv'")
	}
	if !strings.Contains(errMsg, "Dependency 'mpv' is missing") {
		t.Error("expected error message to contain missing dependency details")
	}
}

func TestCheckDependencies_Present(t *testing.T) {
	// Mock lookPath to return success
	oldLookPath := lookPath
	defer func() { lookPath = oldLookPath }()

	lookPath = func(file string) (string, error) {
		return "/usr/bin/mpv", nil
	}

	err := checkDependencies()
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestPromptAndInstallMPV_UserAborted(t *testing.T) {
	oldStdIn := stdIn
	defer func() { stdIn = oldStdIn }()

	// Mock stdin returning "n"
	stdIn = strings.NewReader("n\n")

	err := promptAndInstallMPV()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "installation aborted by user") {
		t.Errorf("expected aborted error, got: %v", err)
	}
}

func TestPromptAndInstallMPV_Success(t *testing.T) {
	oldStdIn := stdIn
	oldExecCommand := execCommand
	defer func() {
		stdIn = oldStdIn
		execCommand = oldExecCommand
	}()

	// Mock stdin returning "y"
	stdIn = strings.NewReader("y\n")

	// Mock execCommand to execute a helper process that exits with code 0 (success)
	execCommand = func(name string, arg ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcessInstallSuccess")
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS_INSTALL_SUCCESS=1")
		return cmd
	}

	err := promptAndInstallMPV()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// A helper function that acts as the successful installation command
func TestHelperProcessInstallSuccess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS_INSTALL_SUCCESS") != "1" {
		return
	}
	// Exit successfully
	os.Exit(0)
}
