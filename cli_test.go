package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/wsxiaoys/terminal/color"
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

func TestIdentity(t *testing.T) {
	cmd := exec.Command("git", "hooks", "id")
	b, err := cmd.Output()
	assert.Nil(t, err)
	assert.Equal(t, string(b), color.Sprint("553ec650fd4f90003774e2ff00b10bc9aa9ec802\n"))
}
