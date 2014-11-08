# git-hooks

> Hook manager

Rewritten from [icefox/git-hooks](https://github.com/icefox/git-hooks)

## Install

[Download](https://github.com/git-hooks/git-hooks/releases) binary and place it in your `PATH`

Or manually:

    go get github.com/git-hooks/git-hooks

## Usage

```
NAME:
   git-hooks - tool to manage project, user, and global Git hooks

USAGE:
   git-hooks [global options] command [command options] [arguments...]

COMMANDS:
   install, i		Replace existing hooks in this repository with a call to git hooks run [hook].  Move old hooks directory to hooks.old
   uninstall		Remove existing hooks in this repository and rename hooks.old back to hooks
   install-global	Create a template .git directory that that will be used whenever a git repository is created or cloned that will remind the user to install git-hooks
   uninstall-global	Turn off the global .git directory template that has the reminder
   run			    run <cmd> Run the hooks for <cmd> (such as pre-commit)
   help, h		    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h		show help
   --version, -v	print the version
```

## How it works

When you invoke `git hooks install`, it replace all the hooks under .git/hooks with

    #!/usr/bin/env bash
    git-hooks run "$0" "$@"

Hook execution will be routed with `git-hooks`

Fow more info, see [wiki](https://github.com/git-hooks/git-hooks/wiki)
