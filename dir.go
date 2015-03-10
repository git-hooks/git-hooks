package main

import (
	"encoding/json"
	"github.com/cattail/go-exclude"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

// list directories for project, user and global scopes
func hookDirs() map[string]string {
	dirs := make(map[string]string)

	// project scope
	root, err := getGitRepoRoot()
	if err == nil {
		path := filepath.Join(root, "githooks")
		isExist, _ := exists(path)
		if isExist {
			dirs["project"] = path
		}
	}

	// user scope
	home, err := homedir.Dir()
	if err == nil {
		path := filepath.Join(home, ".githooks")
		isExist, _ := exists(path)
		if isExist {
			dirs["user"] = path
		}
	}

	// global scope
	// NOTE: git-hooks global hook actually configured via git --system
	// configuration file
	global, err := gitExec("config --get --system hooks.global")
	if err == nil {
		path := global
		isExist, _ := exists(path)
		if isExist {
			dirs["global"] = path
		}
	}

	return dirs
}

// list configurations for project, user and global scopes
func hookConfigs() map[string]string {
	configs := make(map[string]string)

	root, err := getGitRepoRoot()
	if err == nil {
		path := filepath.Join(root, "githooks.json")
		isExist, _ := exists(path)
		if isExist {
			configs["project"] = path
		}
	}

	home, err := homedir.Dir()
	if err == nil {
		path := filepath.Join(home, ".githooks.json")
		isExist, _ := exists(path)
		if isExist {
			configs["user"] = path
		}
	}

	global, err := gitExec("config --get --system hooks.globalconfig")
	if err == nil {
		path := global
		isExist, _ := exists(path)
		if isExist {
			configs["global"] = path
		}
	}

	return configs
}

// List available hooks inside directory
// Under trigger directory,
// Treate file as a hook if it's executable,
// Treate directory as a hook if it contain an executable file with the name of `trigger`
// Example:
// githooks
//     ├── _pre-commit
//     │   ├── test
//     │   └── whitespace
//     └── pre-commit
//         ├── dir
//         │   └── pre-commit
//         └── whitespace
func listHooksInDir(scope, dirname string) (hooks map[string][]string, err error) {
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
				// filter files or directories
				file, err := os.Stat(filepath.Join(dirname, dir.Name(), file.Name()))
				if err == nil {
					if file.IsDir() {
						libs, err := ioutil.ReadDir(filepath.Join(dirname, dir.Name(), file.Name()))
						if err == nil {
							for _, lib := range libs {
								libname := lib.Name()
								extension := filepath.Ext(libname)
								if isExecutable(lib) && libname[0:len(libname)-len(extension)] == dir.Name() {
									hooks[dir.Name()] = append(hooks[dir.Name()], filepath.Join(file.Name(), libname))
								}
							}
						}
					} else {
						if isExecutable(file) {
							hooks[dir.Name()] = append(hooks[dir.Name()], file.Name())
						}
					}
				}
			}
		}
	}

	//
	// exclude
	//
	// exclude only works for user and global scope
	if scope == "user" || scope == "global" {
		file, err := ioutil.ReadFile(filepath.Join(dirname, "excludes.json"))
		if err == nil {
			var excludes interface{}
			json.Unmarshal(file, &excludes)

			wrapper := make(map[string]interface{})
			// repoid will be empty string if not in a git repo or don't have any commit yet
			repoid, _ := gitExec(GIT["FirstCommit"])

			if scope == "user" {
				wrapper[repoid] = hooks
				exclude.Exclude(wrapper, excludes)
				if wrapper[repoid] == nil {
					wrapper[repoid] = make(map[string][]string)
				}
				hooks = wrapper[repoid].(map[string][]string)
			} else {
				// global scope exclude
				user, err := user.Current()
				username := ""
				if err == nil {
					username = user.Username
				}
				wrapper[username] = make(map[string]interface{})
				wrapper[username].(map[string]interface{})[repoid] = hooks
				exclude.Exclude(wrapper, excludes)
				if wrapper[username] == nil {
					wrapper[username] = make(map[string][]string)
				}
				if wrapper[repoid] == nil {
					wrapper[repoid] = make(map[string][]string)
				}
				hooks = wrapper[username].(map[string]interface{})[repoid].(map[string][]string)
			}
		}
	}

	return hooks, nil
}

// List available hooks configured by config file
func listHooksInConfig(config string) (hooks map[string]map[string][]string, err error) {
	hooks = make(map[string]map[string][]string)

	file, err := ioutil.ReadFile(config)
	if err != nil {
		return
	}

	json.Unmarshal(file, &hooks)
	return
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
