/*
Unit test unitilies.
*/
package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Create temporary directory
func createDirectory(t *testing.T, dir string, context func(tempdir string)) {
	tempdir, err := ioutil.TempDir(dir, "git-hooks")
	assert.Nil(t, err)

	current, err := os.Getwd()
	assert.Nil(t, err)

	err = os.Chdir(tempdir)
	assert.Nil(t, err)

	context(tempdir)

	err = os.Chdir(current)
	assert.Nil(t, err)

	err = os.RemoveAll(tempdir)
	assert.Nil(t, err)
}

// Create temporary git repo
func createGitRepo(t *testing.T, context func(tempdir string)) {
	createDirectory(t, filepath.Join("fixtures", "repos"), func(tempdir string) {
		cmd := exec.Command("bash", "-c", `
		git init;
		git config user.email "zhongchiyu@gmail.com";
		git config user.name "CatTail";
		`)
		err := cmd.Run()
		assert.Nil(t, err)

		context(tempdir)
	})
}

func createGitRepoFromDir(t *testing.T, dir string, context func(tempdir string)) {
	createDirectory(t, filepath.Join("fixtures", "repos"), func(tempdir string) {
		tempdir = filepath.Join(tempdir, "repo")
		// copy dir content into tempdir
		err := os.Link(dir, tempdir)
		fmt.Println(err)
		assert.Nil(t, err)

		cmd := exec.Command("bash", "-c", `
		git init;
		git config user.email "zhongchiyu@gmail.com";
		git config user.name "CatTail";
		`)
		err = cmd.Run()
		assert.Nil(t, err)

		context(tempdir)
	})
}
