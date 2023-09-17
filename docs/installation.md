# Installation

## Linux

SGPT has been tested on Ubuntu LTS releases and is expected to be compatible with the following Linux
distributions:

- Debian
- Ubuntu
- Arch Linux
- Fedora

To install, download the latest release from the [release page](https://github.com/tbckr/sgpt/releases) and use the
package manager specific to your distribution.

## macOS

For users with Homebrew as their package manager, run the following command in the terminal:

```shell
brew install tbckr/tap/sgpt
```

## Windows

For users with Scoop as their package manager, execute these commands in PowerShell:

```shell
scoop bucket add tbckr https://github.com/tbckr/scoop-bucket.git
scoop install tbckr/sgpt
```

## Using Go

To install SGPT with Go, based on the git tag, use this command:

```shell
go install github.com/tbckr/sgpt/cmd/sgpt@latest
```

## Docker

To run SGPT with Docker, use the following command to pull the latest image:

```shell
docker pull ghcr.io/tbckr/sgpt:latest
```

## Other platforms

For other platforms, visit the GitHub [release page](https://github.com/tbckr/sgpt/releases) and download the latest
release suitable for your system.
