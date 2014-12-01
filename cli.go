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
	installed, err := isInstalled()
	if err != nil {
		logger.Infoln("Current directory is not a git repo")
	} else if installed {
		logger.Infoln("Git hooks ARE installed in this repository.")
	} else {
		logger.Infoln("Git hooks are NOT installed in this repository. (Run 'git hooks install' to install it)")
	}

	for scope, dir := range hookDirs() {
		logger.Infoln(scope + " hooks")

		config, err := listHooksInDir(scope, dir)
		if err != nil {
			continue
		}

		for trigger, hooks := range config {
			logger.Infoln("  " + trigger)

			for _, hook := range hooks {
				logger.Infoln("    - " + hook)
			}
		}
		logger.Infoln()
	}

	logger.Infoln("Community hooks")
	for scope, configPath := range hookConfigs() {
		logger.Infoln(scope + " hooks")

		config, err := listHooksInConfig(configPath)
		if err != nil {
			continue
		}

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

// If git-hooks installed in the current git repo
func isInstalled() (installed bool, err error) {
	installed = false

	root, err := getGitRepoRoot()
	if err != nil {
		return
	}

	preCommitHook := filepath.Join(root, ".git/hooks/pre-commit")
	hook, err := ioutil.ReadFile(preCommitHook)
	installed = err == nil && strings.EqualFold(string(hook), tplPostInstall)
	return
}

// Install git-hook into current git repo
func install(isInstall bool) {
	dirPath, err := getGitDirPath()
	if err != nil {
		logger.Errorln("Current directory is not a git repo")
		return
	}

	if isInstall {
		isExist, _ := exists(filepath.Join(dirPath, "hooks.old"))
		if isExist {
			logger.Errorln("@rhooks.old already exists, perhaps you already installed?")
			return
		}
		installInto(dirPath, tplPostInstall)
	} else {
		isExist, _ := exists(filepath.Join(dirPath, "hooks.old"))
		if !isExist {
			logger.Errorln("Error, hooks.old doesn't exists, aborting uninstall to not destroy something")
			return
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
		return
	}

	latest, err := semver.New(version[1:])
	if err != nil {
		logger.Errorln("Semver parse error ", err)
		return
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
		if *asset.Name != target {
			continue
		}

		// download
		tmpFileName, err := downloadFromUrl(*asset.BrowserDownloadUrl)
		if err != nil {
			logger.Errorln("Download error", err)
			return
		}
		logger.Infoln("Download complete")

		// uncompress
		tmpFileName, err = extract(tmpFileName)
		if err != nil {
			logger.Errorln("Download error", err)
			return
		}
		logger.Infoln("Extract complete")

		// replace current version
		fileName, err := absExePath(os.Args[0])
		if err != nil {
			logger.Errorln(err)
			return
		}

		debug("Replace %s with temp file %s", fileName, tmpFileName)
		out, err := os.Create(fileName)
		if err != nil {
			logger.Errorln("Create error ", err)
			return
		}
		defer out.Close()

		err = out.Chmod(0755)
		if err != nil {
			logger.Errorln("Create error ", err)
			return
		}

		in, err := os.Open(tmpFileName)
		if err != nil {
			logger.Errorln("Open error ", err)
			return
		}
		defer in.Close()

		_, err = io.Copy(out, in)
		if err != nil {
			logger.Errorln("Copy error ", err)
			return
		}
		logger.Infoln(NAME + " update to " + version)

		break
	}
}

func identity() {
	identity, err := gitExec("rev-list --max-parents=0 HEAD")
	if err != nil {
		logger.Errorln(err)
		return
	}

	logger.Infoln(identity)
}

// run(trigger string, args ...string)
// Execute trigger with supplied arguments.
func run(cmds ...string) {
	trigger := filepath.Base(cmds[0])
	args := cmds[1:]

	runDirHooks(trigger, args...)
	runContribHooks(trigger, args...)
}

func runDirHooks(trigger string, args ...string) {
	for scope, dir := range hookDirs() {
		config, err := listHooksInDir(scope, dir)
		if err != nil {
			continue
		}

		for t, hooks := range config {
			// semi scope
			if t != trigger && t != ("_"+trigger) {
				continue
			}
			for _, hook := range hooks {
				status, err := runHook(filepath.Join(dir, t, hook), args...)
				if err != nil {
					logger.Errorsln(status, err)
					return
				}
			}
		}
	}
}

func runContribHooks(trigger string, args ...string) {
	contrib := getContribDir()

	// wether contrib repo updated
	updated := false

	for _, configPath := range hookConfigs() {
		config, err := listHooksInConfig(configPath)
		if err != nil {
			continue
		}

		for t, repo := range config {
			if t != trigger {
				continue
			}

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
				for index := 0; index < len(hooks); index++ {
					hook := hooks[index]

					status, err := runHook(filepath.Join(contrib, repoName, hook), args...)
					if err == nil {
						// skip update if everything ok
						continue
					}

					if status == 126 && !updated {
						// try to update contrib repo
						logger.Infoln("Update community hooks")
						updated = true

						_, err := gitExecWithDir(filepath.Join(contrib, repoName), fmt.Sprintf("pull origin master"))
						if err == nil {
							index--
							continue
						}

						logger.Warnln("There is something not right with your community hook repo")
					}
					logger.Errorsln(status, err)
				}
			}
		}
	}
}

// Find contrib directory
func getContribDir() (contrib string) {
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
	return
}

// Execute specific hook with arguments
// Return error message as out if error occured
func runHook(hook string, args ...string) (status int, err error) {
	debug("Execute hook %s %s", hook, args)

	cmd := exec.Command(hook, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if waitStatus, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return waitStatus.ExitStatus(), err
			}
		} else if _, ok := err.(*os.PathError); ok {
			// Command can't be execute
			// http://tldp.org/LDP/abs/html/exitcodes.html
			return 126, err
		} else {
			// exit status unknown
			status = 255
		}
	}

	return
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
