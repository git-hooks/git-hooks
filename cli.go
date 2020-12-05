package main

import (
	"context"
	"fmt"
	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"

	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
)

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
			Usage:     "Install git-hooks in this repo",
			Action:    bind(install, true),
		},
		{
			Name:   "uninstall",
			Usage:  "Restore previous hooks",
			Action: bind(uninstall),
		},
		{
			Name:  "install-global",
			Usage: "Install git-hooks in global. Future initialized repo will install git-hooks by default",
			Action: func(c *cli.Context) {
				home, err := homedir.Dir()
				if err != nil {
					return
				}
				installGlobal(home)
			},
		},
		{
			Name:   "uninstall-global",
			Usage:  "Uninstall global git-hooks",
			Action: bind(uninstallGlobal),
		},
		{
			Name:  "update",
			Usage: "Check and update git-hooks",
			Action: func(c *cli.Context) {
				logger.Infoln("Check latest version...")

				client := github.NewClient(nil)
				releases, _, _ := client.Repositories.ListReleases(
					context.Background(),
					"git-hooks", "git-hooks", &github.ListOptions{})
				update(releases)
			},
		},
		{
			Name:  "run",
			Usage: "Run particular hooks",
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
		logger.Infoln(MESSAGES["NotGitRepo"])
	} else if installed {
		logger.Infoln(MESSAGES["Installed"])
	} else {
		logger.Infoln(MESSAGES["NotInstalled"])
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

	logger.Infoln("Contrib hooks")
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
// If current directory is not a git repo, err will be not `nil`
func isInstalled() (installed bool, err error) {
	installed = false

	root, err := getGitRepoRoot()
	if err != nil {
		return
	}

	preCommitHook := filepath.Join(root, ".git/hooks/pre-commit")
	hook, readErr := ioutil.ReadFile(preCommitHook)
	installed = readErr == nil && strings.EqualFold(string(hook), tplPostInstall)
	return
}

// Install git-hook into current git repo
func install(isInstall bool) {
	dirPath, err := getGitDirPath()
	if err != nil {
		logger.Errorln(MESSAGES["NotGitRepo"])
		return
	}

	if isInstall {
		isExist, _ := exists(filepath.Join(dirPath, "hooks.old"))
		if isExist {
			logger.Errorln(MESSAGES["ExistHooks"])
			return
		}
		installInto(dirPath, tplPostInstall)
	} else {
		isExist, _ := exists(filepath.Join(dirPath, "hooks.old"))
		if !isExist {
			logger.Errorln(MESSAGES["NotExistHooks"])
			return
		}
		os.RemoveAll(filepath.Join(dirPath, "hooks"))
		os.Rename(filepath.Join(dirPath, "hooks.old"), filepath.Join(dirPath, "hooks"))
		logger.Infoln(MESSAGES["Restore"])
	}
}

// Uninstall git-hooks from current git repo
func uninstall() {
	install(false)
}

// Install git-hooks global by setup init.tempdir in ~/.gitconfig
func installGlobal(home string) {
	homeTemplate := DIRS["HomeTemplate"]
	if !filepath.IsAbs(homeTemplate) {
		homeTemplate = filepath.Join(home, homeTemplate)
	}

	isExist, _ := exists(homeTemplate)
	if !isExist {
		os.MkdirAll(filepath.Join(homeTemplate, "hooks"), 0755)
		installInto(homeTemplate, tplPreInstall)
	}

	gitExec(GIT["SetTemplateDir"] + homeTemplate)
	logger.Infoln(MESSAGES["SetTemplateDir"] + homeTemplate)
}

// Reset init.tempdir
func uninstallGlobal() {
	gitExec(GIT["UnsetTemplateDir"])
}

// Check latest version of git-hooks by github release
// If there are new version of git-hooks, download and replace the current one
func update(releases []*github.RepositoryRelease) {
	release := releases[0]
	version := *release.TagName
	logger.Infoln("Current version is " + VERSION + ", latest version is " + version)

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

	// latest version
	if current.GTE(*latest) {
		logger.Infoln(MESSAGES["UpdateToDate"])
		return
	}

	// backward incompatible
	if latest.Major != current.Major {
		logger.Infoln(MESSAGES["Incompatible"])
		return
	}

	// previous version
	target := fmt.Sprintf("git-hooks_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	for _, asset := range release.Assets {
		if *asset.Name != target {
			continue
		}

		// download
		logger.Infoln("Downloading latest version...")
		tmpFileName, err := downloadFromUrl(*asset.BrowserDownloadURL)
		if err != nil {
			logger.Errorln("Fail to download", err)
			return
		}
		logger.Infoln("Download complete")

		// extract
		tmpFileName, err = extract(tmpFileName)
		if err != nil {
			logger.Errorln("Fail to extract tar.gz", err)
			return
		}
		logger.Infoln("Extract complete")

		// install binary in non-test environment
		if !isTestEnv() {
			err = installBinary(tmpFileName)
			if err != nil {
				logger.Errorln("Fail to install binary", err)
				return
			}
		}
		logger.Infoln("Successfully update " + NAME + " to " + version)
		break
	}
}

func identity() {
	identity, err := gitExec(GIT["FirstCommit"])
	if err != nil {
		logger.Errorln(err)
		return
	}

	logger.Infoln(identity)
}

// run(trigger string, args ...string)
// Execute trigger with supplied arguments.
func run(cmds ...string) {
	if len(cmds) == 0 {
		logger.Warnln("Missing trigger")
		return
	}
	trigger := filepath.Base(cmds[0])
	args := cmds[1:]

	runDirHooks(hookDirs(), trigger, args...)
	runConfigHooks(hookConfigs(), getContribDir(), trigger, args...)
}

func runDirHooks(dirs map[string]string, current string, args ...string) {
	for scope, dir := range dirs {
		structure, err := listHooksInDir(scope, dir)
		if err != nil {
			continue
		}

		for trigger, hooks := range structure {
			// semi scope
			if trigger != current && trigger != ("_"+current) {
				continue
			}
			for _, hook := range hooks {
				status, err := runHook(filepath.Join(dir, trigger, hook), args...)
				if err != nil {
					logger.Errorsln(status, err)
					return
				}
			}
		}
	}
}

func runConfigHooks(configs map[string]string, contrib string, current string, args ...string) {
	// wether contrib repo updated
	updated := false

	for _, config := range configs {
		structure, err := listHooksInConfig(config)
		if err != nil {
			continue
		}

		for trigger, repo := range structure {
			if trigger != current {
				continue
			}

			for repoName, hooks := range repo {
				fullGitAddress, strippedGitAddress := findProtocol(repoName)
				// check if repo exist in local file system
				isExist, _ := exists(filepath.Join(contrib, strippedGitAddress))
				if !isExist {
					cmd := fmt.Sprintf("clone %s %s", fullGitAddress, filepath.Join(contrib, strippedGitAddress))
					logger.Infoln(cmd)
					_, err := gitExec(cmd)
					if err != nil {
						logger.Warnln(err)
						continue
					}
				}

				for index := 0; index < len(hooks); index++ {
					hook := hooks[index]

					status, err := runHook(filepath.Join(contrib, strippedGitAddress, hook), args...)
					if err == nil {
						// skip update if everything ok
						continue
					}

					// hook not found
					if status == 126 && !updated {
						// try to update contrib repo
						logger.Infoln("Updating contrib hooks")
						updated = true

						_, err := gitExecWithDir(filepath.Join(contrib, strippedGitAddress), "pull origin master")
						if err == nil {
							// try again
							index--
							continue
						}

						logger.Warnln("Something wrong with contrib hook")
					}
					logger.Errorsln(status, err)
				}
			}
		}
	}
}

// Execute specific hook with arguments
// Return error message as out if error occured
func runHook(hook string, args ...string) (status int, err error) {
	var cmd *exec.Cmd
	// Will run a shell script with sh.exe include in Git for Windows
	if runtime.GOOS == "windows" {
		windowsCmd := "cmd"
		//fmt.Println("windowsCmd is", windowsCmd)
		cmdArgs := []string {"/C","sh.exe", hook}
		//fmt.Println("cmdArgs is", cmdArgs)
		windowsArgs := append(cmdArgs,args...)
		//fmt.Println("windowsArgs is", windowsArgs)
		cmd = exec.Command(windowsCmd, windowsArgs...)
	} else {
		cmd = exec.Command(hook, args...)
	}
	
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
		logger.Infoln("Install " + hook)
		f, _ := os.Create(filepath.Join(dir, "hooks", hook))
		f.WriteString(template)
		f.Sync()
		f.Chmod(0755)
	}
}

func findProtocol(input string) (string, string) {
	// check for ssh
	protocol := regexp.MustCompile("^ssh://([a-zA-Z_-]+@([a-zA-Z0-9.-]+):(.*))")
	match := protocol.MatchString(input)
	if match {
		noProtocolNoUser := protocol.ReplaceAllString(input, "$2/$3")
		noProtocol := protocol.ReplaceAllString(input, "$1")
		return noProtocol, noProtocolNoUser
	}
	// check for http
	protocol = regexp.MustCompile("^http[s]?://(.*)")
	match = protocol.MatchString(input)
	if match {
		noProtocol := protocol.ReplaceAllString(input, "$1")
		return input, noProtocol
	}
	// no protocol
	return fmt.Sprintf("https://%s", input), input
}
