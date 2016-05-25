# git-hooks
[![Build Status](https://travis-ci.org/git-hooks/git-hooks.svg?branch=master)](https://travis-ci.org/git-hooks/git-hooks)

> Hook manager

Rewritten from [icefox/git-hooks](https://github.com/icefox/git-hooks), with extra features

## Install

[Download](https://github.com/git-hooks/git-hooks/releases) tarball, extract, place it in your `PATH`, and rename it as `git-hooks`

If you already installed `git-hooks`, update it by `git hooks update`


Or manually:

```bash
cd $GOPATH/src
git clone git@github.com:git-hooks/git-hooks.git
cd git-hooks/
# install godep and restore deps
make get
# install binary
go install
```

## Usage

See [Get Started](https://github.com/git-hooks/git-hooks/wiki/Get-Started)

Fow more info, see [wiki](https://github.com/git-hooks/git-hooks/wiki)
