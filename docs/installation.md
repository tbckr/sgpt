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
go install github.com/tbckr/sgpt/v2/cmd/sgpt@latest
```

## Docker

To run SGPT with Docker, use the following command to pull the latest image:

```shell
docker pull ghcr.io/tbckr/sgpt:latest
```

## Ansible

To install SGPT with Ansible, you can use the following ansible playbook as your base and adapt accordingly:

```yaml
---
- hosts: all
  tasks:
    - name: Get latest sgpt release
      uri:
        url: "https://api.github.com/repos/tbckr/sgpt/releases/latest"
        return_content: yes
      register: sgpt_release

    - name: Set latest version of sgpt
      set_fact:
        sgpt_latest_version: "{{ sgpt_release.json.tag_name }}"

    - name: Install sgpt for debian based, amd64 systems
      ansible.builtin.apt:
        deb: https://github.com/tbckr/sgpt/releases/download/{{ sgpt_latest_version }}/sgpt_{{ sgpt_latest_version[1:] }}_amd64.deb
        allow_unauthenticated: true
```

The playbook can be run with the following command:

```shell
ansible-playbook -i <inventory> <playbook>.yml
```

The latest version of the playbook can be found [here](https://github.com/tbckr/sgpt/blob/main/playbook.yml).

## Other platforms

For other platforms, visit the GitHub [release page](https://github.com/tbckr/sgpt/releases) and download the latest
release suitable for your system.

## Verifying artifacts

A checksum is created for all artifacts and stored in the `checksums.txt` file. The checksum file is signed
with [cosign](https://github.com/sigstore/cosign).

1. Download the files you want, and the `checksums.txt`, `checksum.txt.pem` and `checksums.txt.sig` files from the
   [releases](https://github.com/tbckr/sgpt/releases) page.

2. Verify the signature of the checksum file:

```shell
cosign verify-blob \
--certificate-identity 'https://github.com/tbckr/sgpt/.github/workflows/release.yml@refs/heads/main' \
--certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
--cert 'checksums.txt.pem' \
--signature 'checksums.txt.sig' \
./checksums.txt
```

1. If the signature is valid, you can then verify the SHA256 sums match with the downloaded binary:

```shell
sha256sum --ignore-missing -c checksums.txt
```

## Verify docker images

The docker images are signed with [cosign](https://github.com/sigstore/cosign).

Verify the signatures of the docker images:

```shell
cosign verify \
--certificate-identity 'https://github.com/tbckr/sgpt/.github/workflows/release.yml@refs/heads/main' \
--certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
ghcr.io/tbckr/sgpt:latest
```
