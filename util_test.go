package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
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
	assert.Equal(t, "git-hooks", filepath.Base(root))
}

func TestGetDirPath(t *testing.T) {
	path, err := getGitDirPath()
	assert.Nil(t, err)
	assert.Equal(t, ".git", filepath.Base(path))
}

func TestGitExec(t *testing.T) {
	identity, err := gitExec("rev-list --max-parents=0 HEAD")
	assert.Nil(t, err)
	assert.Equal(t, "553ec650fd4f90003774e2ff00b10bc9aa9ec802", identity)
}

func TestGitExecWithDir(t *testing.T) {
	wd, err := os.Getwd()
	assert.Nil(t, err)

	identity, err := gitExecWithDir(wd, "rev-list --max-parents=0 HEAD")
	assert.Nil(t, err)
	assert.Equal(t, "553ec650fd4f90003774e2ff00b10bc9aa9ec802", identity)
}

func TestBind(t *testing.T) {
	sum := 0
	f := bind(func(a, b int) {
		sum = a + b
	}, 1, 2)
	f(&cli.Context{})
	assert.Equal(t, 3, sum)
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

	fileName, err := downloadFromUrl(ts.URL)
	assert.Nil(t, err)

	file, err := os.Open(fileName)
	assert.Nil(t, err)

	fileinfo, err := file.Stat()
	assert.Nil(t, err)

	b := make([]byte, fileinfo.Size())
	_, err = file.Read(b)
	assert.Nil(t, err)
	assert.Equal(t, string(b), content)
}

func TestExtract(t *testing.T) {
	fileName, err := extract("./fixtures/test.tar.gz")
	assert.Nil(t, err)

	file, err := os.Open(fileName)
	assert.Nil(t, err)

	fileinfo, err := file.Stat()
	assert.Nil(t, err)

	bytes := make([]byte, fileinfo.Size())
	_, err = file.Read(bytes)
	assert.Nil(t, err)
	// vim store the file with extra newline at the EOF
	assert.Equal(t, "test\n", string(bytes))
}

func TestAbsExePath(t *testing.T) {
	path, err := absExePath("ls")
	assert.Nil(t, err)
	assert.Equal(t, "/bin/ls", path)

	// should follow symlic
	temp, err := ioutil.TempDir(os.TempDir(), "git-hooks-test")
	assert.Nil(t, err)
	os.Setenv("PATH", temp+":$PATH")
	err = os.Symlink("/bin/ls", filepath.Join(temp, "ls"))
	assert.Nil(t, err)
	path, err = absExePath("ls")
	assert.Nil(t, err)
	assert.Equal(t, "/bin/ls", path)
}

func TestIsExecutable(t *testing.T) {
	fileinfo, err := os.Stat("/bin/ls")
	assert.Nil(t, err)
	assert.True(t, isExecutable(fileinfo))
}
