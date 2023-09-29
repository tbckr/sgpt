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

- [Task](https://taskfile.dev/installation)
- [Go 1.21+](https://go.dev/doc/install)
- [Python 3.9+](https://www.python.org/downloads/)
- [Docker](https://docs.docker.com/get-docker/)

Clone `sgpt` anywhere:

```sh
git clone git@github.com:tbckr/sgpt.git
```

`cd` into the directory and install the dependencies:

```sh
task setup
```

This will install the following tools and pre-commit hook:

- [pre-commit](https://pre-commit.com/#install)
- [goreleaser](https://goreleaser.com/install/)
- [golangci-lint](https://golangci-lint.run/usage/install/#local-installation)
- [addlicense](https://github.com/google/addlicense)
- [svu](https://github.com/caarlos0/svu)
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
- [pipenv](https://pypi.org/project/pipenv/)

Check that everything is working by running the `validate:devtools` task:

```sh
task validate:devtools
```

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
