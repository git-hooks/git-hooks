package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHookDirs(t *testing.T) {
	createGitRepoFromDir(t, "example", func(tempdir string) {
		fmt.Println(tempdir)
		assert.True(t, true)
	})
}

func TestHookConfigs(t *testing.T) {

}

func TestListHooksInDir(t *testing.T) {

}

func TestListHooksInConfig(t *testing.T) {

}
