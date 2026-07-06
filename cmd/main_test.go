package main

import (
	"errors"
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
