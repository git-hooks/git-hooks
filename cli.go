package main

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/codegangsta/cli"
	"github.com/google/go-github/github"
	"github.com/mitchellh/go-homedir"
	. "github.com/tj/go-debug"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

var debug = Debug("main")

func main() {
	app := cli.NewApp()
	app.Name = NAME
	app.Usage = "tool to manage project, user, and global Git hooks"
	app.Version = VERSION
	app.EnableBashCompletion = true
	app.Action = bind(list)
	app.Commands = []cli.Command{
		{
			Name:      "install",
			ShortName: "i",
			Usage:     "Tell repo to use git-hooks by replace existing hooks with a call to git-hooks. Old hooks will be reserved in hooks.old",
			Action:    bind(install, true),
		},
		{
			Name:   "uninstall",
			Usage:  "Stop using git-hooks and restore old hooks",
			Action: bind(uninstall),
		},
		{
			Name:   "install-global",
			Usage:  "Whenever a git repository is created or cloned user will be remind to install git-hooks",
			Action: bind(installGlobal),
		},
		{
			Name:   "uninstall-global",
			Usage:  "Turn off the global reminder",
			Action: bind(uninstallGlobal),
		},
		{
			Name:   "update",
			Usage:  "Check and update git-hooks",
			Action: bind(update),
		},
		{
			Name:  "run",
			Usage: "Run hooks",
			Action: func(c *cli.Context) {
				run(c.Args()...)
			},
		},
		{
			Name:      "identity",
			ShortName: "id",
			Usage:     "Repo identity",
			Action:    bind(identity),
		},
	}

	app.Run(os.Args)
}

// List directory base hooks and configuration file based hooks
func list() {
	root, err := getGitRepoRoot()
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

	for scope, dir := range hookDirs() {
		logger.Infoln(scope + " hooks")
		config, err := listHooksInDir(scope, dir)
		if err == nil {
			for trigger, hooks := range config {
				logger.Infoln("  " + trigger)
				for _, hook := range hooks {
					logger.Infoln("    - " + hook)
				}
			}
			logger.Infoln()
		}
	}

	logger.Infoln("Community hooks")
	for scope, configPath := range hookConfigs() {
		logger.Infoln(scope + " hooks")
		config, err := listHooksInConfig(configPath)
		if err == nil {
			for trigger, repo := range config {
				logger.Infoln("  " + trigger)
				for repoName, hooks := range repo {
					logger.Infoln("  " + repoName)
					for _, hook := range hooks {
						logger.Infoln("    - " + hook)
					}
				}
			}
		}
	}
}

// Install git-hook into current git repo
func install(isInstall bool) {
	dirPath, err := getGitDirPath()
	if err != nil {
		logger.Errorln("Current directory is not a git repo")
	}

	if isInstall {
		isExist, _ := exists(filepath.Join(dirPath, "hooks.old"))
		if isExist {
			logger.Errorln("@rhooks.old already exists, perhaps you already installed?")
		}
		installInto(dirPath, tplPostInstall)
	} else {
		isExist, _ := exists(filepath.Join(dirPath, "hooks.old"))
		if !isExist {
			logger.Errorln("Error, hooks.old doesn't exists, aborting uninstall to not destroy something")
		}
		os.RemoveAll(filepath.Join(dirPath, "hooks"))
		os.Rename(filepath.Join(dirPath, "hooks.old"), filepath.Join(dirPath, "hooks"))
		logger.Infoln("Restore hooks.old")
	}
}

// Uninstall git-hooks from current git repo
func uninstall() {
	install(false)
}

// Install git-hooks global by setup init.tempdir in ~/.gitconfig
func installGlobal() {
	templatedir := ".git-template-with-git-hooks"
	home, err := homedir.Dir()
	if err == nil {
		templatedir = filepath.Join(home, templatedir)
	}
	isExist, _ := exists(templatedir)
	if !isExist {
		defaultdir := "/usr/share/git-core/templates"
		isExist, _ = exists(defaultdir)
		if isExist {
			os.Link(defaultdir, templatedir)
		} else {
			os.Mkdir(filepath.Join(templatedir, "hooks"), 0755)
		}
		installInto(templatedir, tplPreInstall)
	}
	gitExec("config --global init.templatedir " + templatedir)
	os.Rename(filepath.Join(templatedir, "hooks.old"), filepath.Join(templatedir, "hooks.original"))
	logger.Infoln("Git global config init.templatedir is now set to " + templatedir)
}

// Reset init.tempdir
func uninstallGlobal() {
	gitExec("config --global --unset init.templatedir")
}

// Check latest version of git-hooks by github release
// If there are new version of git-hooks, download and replace the current one
func update() {
	logger.Infoln("Current git-hooks version is " + VERSION)
	logger.Infoln("Check latest version...")

	client := github.NewClient(nil)
	releases, _, _ := client.Repositories.ListReleases(
		"git-hooks", "git-hooks", &github.ListOptions{})
	release := releases[0]
	version := *release.TagName
	logger.Infoln("Latest version is " + version)

	// compare version
	current, err := semver.New(VERSION[1:])
	if err != nil {
		logger.Errorln("Semver parse error ", err)
	}
	latest, err := semver.New(version[1:])
	if err != nil {
		logger.Errorln("Semver parse error ", err)
	}
	debug("Current version %s, latest version %s", current, latest)

	if !latest.GT(current) {
		logger.Infoln("Your " + NAME + " is update to date")
		return
	}

	// version compability
	if latest.Major != current.Major {
		logger.Infoln("Current version is ", current)
		logger.Infoln("Latest version is ", latest)
		logger.Infoln("Version incompatible, manually update please")
		return
	}

	logger.Infoln("Download latest version...")
	target := fmt.Sprintf("git-hooks_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	for _, asset := range release.Assets {
		if *asset.Name == target {
			// download
			tmpFileName, err := downloadFromUrl(*asset.BrowserDownloadUrl)
			if err != nil {
				logger.Errorln("Download error", err)
			}
			logger.Infoln("Download complete")

			// uncompress
			tmpFileName, err = extract(tmpFileName)
			if err != nil {
				logger.Errorln("Download error", err)
			}
			logger.Infoln("Extract complete")

			// replace current version
			fileName, err := absExePath(os.Args[0])
			if err != nil {
				logger.Errorln(err)
			}

			debug("Replace %s with temp file %s", fileName, tmpFileName)
			out, err := os.Create(fileName)
			if err != nil {
				logger.Errorln("Create error ", err)
			}
			defer out.Close()

			err = out.Chmod(0755)
			if err != nil {
				logger.Errorln("Create error ", err)
			}

			in, err := os.Open(tmpFileName)
			if err != nil {
				logger.Errorln("Open error ", err)
			}
			defer in.Close()

			_, err = io.Copy(out, in)
			if err != nil {
				logger.Errorln("Copy error ", err)
			}
			logger.Infoln(NAME + " update to " + version)

			break
		}
	}
}

func identity() {
	identity, err := gitExec("rev-list --max-parents=0 HEAD")
	if err != nil {
		logger.Errorln(err)
	}

	logger.Infoln(identity)
}

// Execute project, semi, user and global scope hooks
func run(cmds ...string) {
	t := filepath.Base(cmds[0])
	args := cmds[1:]
	for scope, dir := range hookDirs() {
		config, err := listHooksInDir(scope, dir)
		if err == nil {
			for trigger, hooks := range config {
				// semi scope
				if trigger == t || trigger == ("_"+t) {
					for _, hook := range hooks {
						status, err := runHook(filepath.Join(dir, trigger, hook), args...)
						if err != nil {
							logger.Errorsln(status, err)
						}
					}
				}
			}
		}
	}

	// find contrib directory
	contrib, err := gitExec("config --get hooks.contrib")
	isExist, _ := exists(contrib)
	if err != nil || !isExist {
		// default to use ~/.githooks-contrib
		home, err := homedir.Dir()
		if err != nil {
			// fallback
			home = "~"
		}
		contrib = filepath.Join(home, "."+CONTRIB_DIRNAME)
	} else {
		contrib = filepath.Join(contrib, CONTRIB_DIRNAME)
	}
	for _, configPath := range hookConfigs() {
		config, err := listHooksInConfig(configPath)
		if err == nil {
			for trigger, repo := range config {
				if trigger == t {
					for repoName, hooks := range repo {
						// check if repo exist in local file system
						isExist, _ := exists(filepath.Join(contrib, repoName))
						if !isExist {
							logger.Infoln("Cloning repo " + repoName)
							_, err := gitExec(fmt.Sprintf("clone https://%s %s", repoName, filepath.Join(contrib, repoName)))
							if err != nil {
								fmt.Printf("clone https://%s %s", repoName, filepath.Join(contrib, repoName))
								fmt.Println(err)
								continue
							}
						}
						// execute hook
						for _, hook := range hooks {
							status, err := runHook(filepath.Join(contrib, repoName, hook), args...)
							if err != nil {
								logger.Errorsln(status, err)
							}
						}
					}
				}
			}
		}
	}
}

// Execute specific hook with arguments
// Return error message as out if error occured
func runHook(hook string, args ...string) (status int, err error) {
	debug("Execute contrib hook %s %s", hook, args)

	cmd := exec.Command(hook, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if err != nil {
		if msg, ok := err.(*exec.ExitError); ok {
			status = msg.Sys().(syscall.WaitStatus).ExitStatus()
		}
	} else {
		status = 0
	}
	return status, err
}

func installInto(dir string, template string) {
	// backup
	os.Rename(filepath.Join(dir, "hooks"), filepath.Join(dir, "hooks.old"))
	os.Mkdir(filepath.Join(dir, "hooks"), 0755)
	for _, hook := range TRIGGERS {
		logger.Infoln("Install ", hook)
		f, _ := os.Create(filepath.Join(dir, "hooks", hook))
		f.WriteString(template)
		f.Sync()
		f.Chmod(0755)
	}
}
