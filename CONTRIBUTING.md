# Contributing

<!-- Shamelessly copied from -->
<!-- https://github.com/goreleaser/goreleaser/blob/b7218b0ab03477fa51d4d4a72ccbbb80150dca27/CONTRIBUTING.md -->

By participating in this project, you agree to abide our
[code of conduct](https://github.com/tbckr/sgpt/blob/main/CODE_OF_CONDUCT.md).

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
<!-- param::isNotitle::true:: -->

- [Set up your machine](#set-up-your-machine)
- [Test your change](#test-your-change)
- [Create a commit](#create-a-commit)
- [Submit a pull request](#submit-a-pull-request)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Set up your machine

`sgpt` is written in [Go](https://go.dev/).

Prerequisites:

- [Go 1.23+](https://go.dev/doc/install)
- [Docker](https://docs.docker.com/get-docker/)
- [Nix Package Manager](https://nixos.org/download.html) (install recommended via [Determinate Nix Installer](https://github.com/DeterminateSystems/nix-installer))
- [Direnv](https://direnv.net/docs/installation.html)

Clone `sgpt` anywhere:

```sh
git clone git@github.com:tbckr/sgpt.git
```

`cd` into the directory and allow the `direnv` to load the environment:

```sh
direnv allow
```

This will install all the necessary dependencies for the project via nix flakes.

A good way to make sure everything else is okay is to run the test suite:

```sh
task test
```

## Test your change

You can create a branch for your changes and try to build from the source as you go:

```sh
task build
```

When you are satisfied with the changes, we suggest you run:

```sh
task ci
```

Before you commit the changes, we also suggest you run:

```sh
task fmt
```

## Create a commit

Commit messages should be well formatted, and to make that "standardized", we
are using Conventional Commits.

You can follow the documentation on
[their website](https://www.conventionalcommits.org).

## Submit a pull request

Push your branch to your `sgpt` fork and open a pull request against the main branch.
