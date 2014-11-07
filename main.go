/*
Terminology

Example git-hooks directory layout:

	githooks
	├── commit-msg
	│   └── signed-off-by
	└── pre-commit
		└── bsd

trigger: pre-commit
hook: bsd
*/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/mitchellh/go-homedir"
	. "github.com/tj/go-debug"
	"github.com/wsxiaoys/terminal/color"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
)

var TRIGGERS = [...]string{"applypatch-msg", "commit-msg", "post-applypatch", "post-checkout", "post-commit", "post-merge", "post-receive", "pre-applypatch", "pre-auto-gc", "pre-commit", "prepare-commit-msg", "pre-rebase", "pre-receive", "update", "pre-push"}

var CONTRIB_PATH = ".hooks"

var tplPreInstall = `#!/usr/bin/env bash
echo \"git hooks not installed in this repository.  Run 'git hooks --install' to install it or 'git hooks -h' for more information.\"`
var tplPostInstall = `#!/usr/bin/env bash
git-hooks run "$0" "$@"`

var logger = struct {
	Errorln func(...string)
	Warnln  func(...string)
	Infoln  func(...string)
}{
	Errorln: func(msgs ...string) {
		for _, msg := range msgs {
			color.Println("@r" + msg)
		}
		os.Exit(1)
	},
	Warnln: func(msgs ...string) {
		for _, msg := range msgs {
			color.Println("@r" + msg)
		}
	},
	Infoln: func(msgs ...string) {
		for _, msg := range msgs {
			color.Println("@b" + msg)
		}
	},
}

var debug = Debug("main")

func main() {
	app := cli.NewApp()
	app.Name = "git-hooks"
	app.Usage = "tool to manage project, user, and global Git hooks"
	app.Version = "0.3.0"
	app.Action = Bind(List)
	app.Commands = []cli.Command{
		{
			Name:      "install",
			ShortName: "i",
			Usage:     "Replace existing hooks in this repository with a call to git hooks run [hook].  Move old hooks directory to hooks.old",
			Action:    Bind(Install, true),
		},
		{
			Name:   "uninstall",
			Usage:  "Remove existing hooks in this repository and rename hooks.old back to hooks",
			Action: Bind(Uninstall),
		},
		{
			Name:   "install-global",
			Usage:  "Create a template .git directory that that will be used whenever a git repository is created or cloned that will remind the user to install git-hooks",
			Action: Bind(InstallGlobal),
		},
		{
			Name:   "uninstall-global",
			Usage:  "Turn off the global .git directory template that has the reminder",
			Action: Bind(UninstallGlobal),
		},
		{
			Name:  "run",
			Usage: "run <cmd> Run the hooks for <cmd> (such as pre-commit)",
			Action: func(c *cli.Context) {
				Run(c.Args()...)
			},
		},
	}

	app.Run(os.Args)
}

func List() {
	root, err := GetGitRepoRoot()
	if err != nil {
		logger.Infoln("Current directory is not a git repo")
	} else {
		preCommitHook := filepath.Join(root, ".git/hooks/pre-commit")
		hook, err := ioutil.ReadFile(preCommitHook)
		if err == nil && strings.EqualFold(string(hook), tplPostInstall) {
			logger.Infoln("Git hooks ARE installed in this repository.")
		} else {
			logger.Infoln("Git hooks are NOT installed in this repository. (Run 'git hooks install' to install it)")
		}
	}

	for scope, dir := range HookDirs() {
		fmt.Println(scope + " hooks")
		config, err := ListHooksInDir(dir)
		if err == nil {
			for trigger, hooks := range config {
				fmt.Println("  " + trigger)
				for _, hook := range hooks {
					fmt.Println("    - " + hook)
				}
			}
			fmt.Println()
		}
	}

	for scope, configPath := range HookConfigs() {
		fmt.Println(scope + " hooks")
		config, err := ListHooksInConfig(configPath)
		if err == nil {
			for trigger, repo := range config {
				fmt.Println("  " + trigger)
				for repoName, hooks := range repo {
					fmt.Println("  " + repoName)
					for _, hook := range hooks {
						fmt.Println("    - " + hook)
					}
				}
			}
		}
	}
}

func Install(isInstall bool) {
	dirPath, err := GetGitDirPath()
	if err != nil {
		logger.Errorln("Current directory is not a git repo")
	}

	if isInstall {
		isExist, _ := Exists(filepath.Join(dirPath, "hooks.old"))
		if isExist {
			logger.Errorln("@rhooks.old already exists, perhaps you already installed?")
		}
		InstallInto(dirPath, tplPostInstall)
	} else {
		isExist, _ := Exists(filepath.Join(dirPath, "hooks.old"))
		if !isExist {
			logger.Errorln("Error, hooks.old doesn't exists, aborting uninstall to not destroy something")
		}
		os.RemoveAll(filepath.Join(dirPath, "hooks"))
		os.Rename(filepath.Join(dirPath, "hooks.old"), filepath.Join(dirPath, "hooks"))
		logger.Infoln("Restore hooks.old")
	}
}

func Uninstall() {
	Install(false)
}

func InstallGlobal() {
	templatedir := ".git-template-with-git-hooks"
	home, err := homedir.Dir()
	if err == nil {
		templatedir = filepath.Join(home, templatedir)
	}
	isExist, _ := Exists(templatedir)
	if !isExist {
		defaultdir := "/usr/share/git-core/templates"
		isExist, _ = Exists(defaultdir)
		if isExist {
			os.Link(defaultdir, templatedir)
		} else {
			os.Mkdir(filepath.Join(templatedir, "hooks"), 0755)
		}
		InstallInto(templatedir, tplPreInstall)
	}
	GitExec("config --global init.templatedir " + templatedir)
	os.Rename(filepath.Join(templatedir, "hooks.old"), filepath.Join(templatedir, "hooks.original"))
	logger.Infoln("Git global config init.templatedir is now set to " + templatedir)
}

func UninstallGlobal() {
	GitExec("config --global --unset init.templatedir")
}

func Run(cmds ...string) {
	wd, err := os.Getwd()
	if err == nil {
		t := filepath.Base(cmds[0])
		args := cmds[1:]
		for _, dir := range HookDirs() {
			config, err := ListHooksInDir(dir)
			if err == nil {
				for trigger, hooks := range config {
					if trigger == t || trigger == ("_"+t) {
						for _, hook := range hooks {
							debug("Execute hook %s", hook)
							cmd := exec.Command(filepath.Join(dir, trigger, hook), args...)
							cmd.Dir = wd
							out, err := cmd.Output()
							if err != nil {
								logger.Errorln(string(out), err.Error())
							} else {
								fmt.Print(string(out))
							}
						}
					}
				}
			}
		}

		// find contrib directory
		home, err := homedir.Dir()
		contrib := CONTRIB_PATH
		if err == nil {
			contrib = filepath.Join(home, CONTRIB_PATH)
		}
		for _, configPath := range HookConfigs() {
			config, err := ListHooksInConfig(configPath)
			if err == nil {
				for trigger, repo := range config {
					if trigger == t {
						for repoName, hooks := range repo {
							// check if repo exist in local file system
							isExist, _ := Exists(filepath.Join(contrib, repoName))
							if !isExist {
								logger.Infoln("Cloning repo " + repoName)
								_, err := GitExec(fmt.Sprintf("clone https://%s %s", repoName, filepath.Join(contrib, repoName)))
								if err != nil {
									continue
								}
							}
							// execute hook
							for _, hook := range hooks {
								debug("Execute contrib hook %s", hook)
								cmd := exec.Command(filepath.Join(contrib, repoName, hook, "hook"), args...)
								cmd.Dir = wd
								out, err := cmd.Output()
								if err != nil {
									logger.Errorln(string(out), err.Error())
								} else {
									fmt.Print(string(out))
								}
							}
						}
					}
				}
			}
		}
	}
}

func ListHooksInDir(dirname string) (hooks map[string][]string, err error) {
	hooks = make(map[string][]string)

	dirs, err := ioutil.ReadDir(dirname)
	if err != nil {
		return
	}

	for _, dir := range dirs {
		files, err := ioutil.ReadDir(filepath.Join(dirname, dir.Name()))
		if err == nil {
			hooks[dir.Name()] = make([]string, 0)
			for _, file := range files {
				if file.Name()[0] != '.' {
					hooks[dir.Name()] = append(hooks[dir.Name()], file.Name())
				}
			}
		}
	}
	return hooks, nil
}

func ListHooksInConfig(config string) (hooks map[string]map[string][]string, err error) {
	hooks = make(map[string]map[string][]string)

	file, err := ioutil.ReadFile(config)
	if err != nil {
		return
	}

	json.Unmarshal(file, &hooks)
	return
}

func InstallInto(dir string, template string) {
	// backup
	os.Rename(filepath.Join(dir, "hooks"), filepath.Join(dir, "hooks.old"))
	os.Mkdir(filepath.Join(dir, "hooks"), 0755)
	for _, hook := range TRIGGERS {
		fmt.Println("Install ", hook)
		f, _ := os.Create(filepath.Join(dir, "hooks", hook))
		f.WriteString(template)
		f.Sync()
		f.Chmod(0755)
	}
}

func HookDirs() map[string]string {
	dirs := make(map[string]string)

	root, err := GetGitRepoRoot()
	if err == nil {
		path := filepath.Join(root, "githooks")
		isExist, _ := Exists(path)
		if isExist {
			dirs["project"] = path
		}
	}

	home, err := homedir.Dir()
	if err == nil {
		path := filepath.Join(home, ".githooks")
		isExist, _ := Exists(path)
		if isExist {
			dirs["user"] = path
		}
	}

	global, err := GitExec("config --get --global hooks.global")
	if err == nil {
		path := global
		isExist, _ := Exists(path)
		if isExist {
			dirs["global"] = path
		}
	}

	return dirs
}

func HookConfigs() map[string]string {
	configs := make(map[string]string)

	root, err := GetGitRepoRoot()
	if err == nil {
		path := filepath.Join(root, "githooks.json")
		isExist, _ := Exists(path)
		if isExist {
			configs["project"] = path
		}
	}

	home, err := homedir.Dir()
	if err == nil {
		path := filepath.Join(home, ".githooks.json")
		isExist, _ := Exists(path)
		if isExist {
			configs["user"] = path
		}
	}

	global, err := GitExec("config --get --global hooks.globalconfig")
	if err == nil {
		path := global
		isExist, _ := Exists(path)
		if isExist {
			configs["global"] = path
		}
	}

	return configs
}

func GetGitRepoRoot() (string, error) {
	return GitExec("rev-parse --show-toplevel")
}

func GetGitDirPath() (string, error) {
	return GitExec("rev-parse --git-dir")
}

func GitExec(args ...string) (string, error) {
	args = strings.Split(strings.Join(args, " "), " ")
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = wd

	if out, err := cmd.Output(); err == nil {
		return string(bytes.Trim(out, "\n")), nil
	} else {
		return "", err
	}
}

func Bind(f interface{}, args ...interface{}) func(c *cli.Context) {
	callable := reflect.ValueOf(f)
	arguments := make([]reflect.Value, len(args))
	for i, arg := range args {
		arguments[i] = reflect.ValueOf(arg)
	}
	return func(c *cli.Context) {
		callable.Call(arguments)
	}
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
