# git-hooks
[![Build Status](https://travis-ci.org/git-hooks/git-hooks.svg?branch=master)](https://travis-ci.org/git-hooks/git-hooks)

> Hook manager

Rewritten from [icefox/git-hooks](https://github.com/icefox/git-hooks), with extra features

## Install

[Download](https://github.com/git-hooks/git-hooks/releases) tarball, extract it and place it in your `PATH`

If you already installed `git-hooks`, update it by `git hooks update`


Or manually:

    git clone git@github.com:git-hooks/git-hooks.git
    # install godep and restore deps
    make get
    # install binary
    go install

## Usage

See [Get Started](https://github.com/git-hooks/git-hooks/wiki/Get-Started)

Fow more info, see [wiki](https://github.com/git-hooks/git-hooks/wiki)

## Debug

Prefix with `DEBUG=*`, for example

    DEBUG=* git hooks
