package main

import (
	"os"
)

var VERSION = "v1.1.2"
var NAME = "git-hooks"
var TRIGGERS = [...]string{"applypatch-msg", "commit-msg", "post-applypatch", "post-checkout", "post-commit", "post-merge", "post-receive", "pre-applypatch", "pre-auto-gc", "pre-commit", "prepare-commit-msg", "pre-rebase", "pre-receive", "update", "pre-push"}

var CONTRIB_DIRNAME = "githooks-contrib"

var tplPreInstall = `#!/usr/bin/env bash
echo \"git hooks not installed in this repository.  Run 'git hooks --install' to install it or 'git hooks -h' for more information.\"`
var tplPostInstall = `#!/usr/bin/env bash
git-hooks run "$0" "$@"`

var ENV = os.Getenv("ENV")

var DIRS = map[string]string{
	"HomeTemplate":   ".git-template-with-git-hooks",
	"GlobalTemplate": "/usr/share/git-core/templates",
}

var GIT = map[string]string{
	"SetTemplateDir":    "config --global init.templatedir ",
	"GetTemplateDir":    "config --global --get init.templatedir",
	"UnsetTemplateDir":  "config --global --unset init.templatedir",
	"RemoveTemplateDir": "config --global --remove init",
	"FirstCommit":       "rev-list --max-parents=0 HEAD",
}

var MESSAGES = map[string]string{
	"NotGitRepo":     "Current directory is not a git repo",
	"Installed":      "Git hooks ARE installed in this repository.",
	"NotInstalled":   "Git hooks are NOT installed in this repository. (Run 'git hooks install' to install it)",
	"ExistHooks":     "hooks.old already exists, perhaps you already installed?",
	"NotExistHooks":  "Error, hooks.old doesn't exists, aborting uninstall to not destroy something",
	"Restore":        "Restore hooks.old",
	"SetTemplateDir": "Git global config init.templatedir is now set to ",
	"UpdateToDate":   "git-hooks is update to date",
	"Incompatible":   "Version backward incompatible, manually update required",
}

func isTestEnv() bool {
	return ENV == "test"
}
