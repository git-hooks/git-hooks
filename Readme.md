# git-hooks
[![Build Status](https://travis-ci.org/git-hooks/git-hooks.svg?branch=master)](https://travis-ci.org/git-hooks/git-hooks)

> Hook manager

Rewritten from [icefox/git-hooks](https://github.com/icefox/git-hooks), with extra features

## Supported Go versions

git-hooks supports the latest two Go version. (Currently 1.12 and 1.13)

## Install

[Download](https://github.com/git-hooks/git-hooks/releases) tarball, extract, place it in your `PATH`, and rename it as `git-hooks`

If you already installed `git-hooks`, update it by `git hooks update`

## Install with `go get`

You need to set `GO111MODULE="on"` if you want to use go get.

You can install with `go get -u github.com/git-hooks/git-hooks` if you already have go toolchain installed.

## Manual install

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

For more info, see [wiki](https://github.com/git-hooks/git-hooks/wiki)
