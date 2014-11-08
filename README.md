# git-hooks

> Hook manager

Rewritten from [icefox/git-hooks](https://github.com/icefox/git-hooks)

## Install

[Download](https://github.com/git-hooks/git-hooks/releases) binary and place it in your `PATH`

Or manually:

    go get github.com/git-hooks/git-hooks

## Usage

See [Get Started](https://github.com/git-hooks/git-hooks/wiki/Get-Started)

## How it works

When you invoke `git hooks install`, it replace all the hooks under .git/hooks with

    #!/usr/bin/env bash
    git-hooks run "$0" "$@"

Hook execution will be routed with `git-hooks`

Fow more info, see [wiki](https://github.com/git-hooks/git-hooks/wiki)
