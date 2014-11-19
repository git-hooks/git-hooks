package main

import (
	"github.com/stretchr/testify/assert"
	"os/exec"
	"syscall"
	"testing"
)

func TestInstall(t *testing.T) {
	// should exit if not in git repo
	cmd := exec.Command("git", "hooks", "install")
	cmd.Dir = "/"
	err := cmd.Run()
	assert.NotNil(t, err)

	status := 0
	if msg, ok := err.(*exec.ExitError); ok {
		status = msg.Sys().(syscall.WaitStatus).ExitStatus()
	}
	assert.Equal(t, status, 1)
}
