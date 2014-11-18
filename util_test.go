package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetGitRoot(t *testing.T) {
	root, err := getGitRepoRoot()
	assert.Nil(t, err)
	assert.Equal(t, filepath.Base(root), "git-hooks")
}

func TestGetDirPath(t *testing.T) {
	path, err := getGitDirPath()
	assert.Nil(t, err)
	assert.Equal(t, filepath.Base(path), ".git")
}

func TestGitExec(t *testing.T) {
	identity, err := gitExec("rev-list --max-parents=0 HEAD")
	assert.Nil(t, err)
	assert.Equal(t, identity, "553ec650fd4f90003774e2ff00b10bc9aa9ec802")
}

func TestBind(t *testing.T) {
	sum := 0
	f := bind(func(a, b int) {
		sum = a + b
	}, 1, 2)
	f(&cli.Context{})
	assert.Equal(t, sum, 3)
}

func TestExists(t *testing.T) {
	isExist, err := exists("notExistFileName")
	assert.Nil(t, err)
	assert.False(t, isExist)
	// THIS file is exist
	_, filename, _, ok := runtime.Caller(1)
	assert.True(t, ok)
	isExist, err = exists(filename)
	assert.Nil(t, err)
	assert.True(t, isExist)
}

func TestDownloadFromUrl(t *testing.T) {
	content := "Hello, client"
	// start test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, content)
	}))
	defer ts.Close()

	file, err := downloadFromUrl(ts.URL)
	assert.Nil(t, err)

	fileinfo, err := file.Stat()
	assert.Nil(t, err)

	b := make([]byte, fileinfo.Size())
	_, err = file.Read(b)
	assert.Nil(t, err)
	assert.Equal(t, string(b), content)
}

func TestAbsExePath(t *testing.T) {
	path, err := absExePath("ls")
	assert.Nil(t, err)
	assert.Equal(t, path, "/bin/ls")
}

func TestIsExecutable(t *testing.T) {
	fileinfo, err := os.Stat("/bin/ls")
	assert.Nil(t, err)
	assert.True(t, isExecutable(fileinfo))
}
