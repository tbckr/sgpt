# SGPT

[![Release](https://img.shields.io/github/release/tbckr/sgpt.svg?style=for-the-badge)](https://github.com/tbckr/sgpt/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![Codecov branch](https://img.shields.io/codecov/c/github/tbckr/sgpt/main.svg?style=for-the-badge)](https://codecov.io/gh/tbckr/sgpt)
[![Go Report Card](https://goreportcard.com/badge/github.com/tbckr/sgpt/v2?style=for-the-badge)](https://goreportcard.com/report/github.com/tbckr/sgpt/v2)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/tbckr/sgpt)
[![Powered By: GoReleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=for-the-badge)](https://github.com/goreleaser)
[![Read the Docs](https://img.shields.io/readthedocs/sgpt?style=for-the-badge)](https://sgpt.readthedocs.io/)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)
[![Protected by Gitleaks](https://img.shields.io/badge/protected%20by-gitleaks-blue?style=for-the-badge)](https://github.com/gitleaks/gitleaks-action)

SGPT (*aka shell-gpt*) is a powerful command-line interface (CLI) tool designed for seamless interaction with OpenAI
models directly from your terminal. Effortlessly run queries, generate shell commands or code, create images from text,
and more, using simple commands. Streamline your workflow and enhance productivity with this powerful and user-friendly
CLI tool.

Developed with the help of [SGPT](https://github.com/tbckr/sgpt).

This is a Go implementation. For the original Python implementation,
visit [shell-gpt](https://github.com/TheR1D/shell_gpt). Please keep this in mind when reporting issues.

> [!NOTE]
> Currently under heavy refactoring for v3, but v2 is still maintained.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
<!-- param::isNotitle::true:: -->

- [Features](#features)
- [Installation](#installation)
  - [Linux](#linux)
  - [macOS](#macos)
  - [Windows](#windows)
  - [Using Go](#using-go)
  - [Docker](#docker)
  - [Ansible](#ansible)
  - [Other platforms](#other-platforms)
- [Usage Guide](#usage-guide)
  - [Getting started: Obtaining an OpenAI API Key](#getting-started-obtaining-an-openai-api-key)
  - [Querying OpenAI Models](#querying-openai-models)
  - [GPT-4o and GPT-4 Vision API Support](#gpt-4o-and-gpt-4-vision-api-support)
  - [o1 API Support](#o1-api-support)
  - [OpenRouter API Support](#openrouter-api-support)
  - [Chat Capabilities](#chat-capabilities)
  - [Generating and Executing Shell Commands](#generating-and-executing-shell-commands)
  - [Interactive Shell Sessions](#interactive-shell-sessions)
  - [Code Generation Capabilities](#code-generation-capabilities)
  - [Enhancing Your Workflow with Bash Aliases and Functions](#enhancing-your-workflow-with-bash-aliases-and-functions)
- [Acknowledgements](#acknowledgements)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Features

- **Instant Answers:** Obtain quick and accurate responses to simple questions directly in your shell, streamlining your
  workflow.
- **GPT-4o Integration:** Access the capabilities of the [GPT-4o API](https://platform.openai.com/docs/models/gpt-4o)
  to generate detailed and informative responses.
- **GPT-4 Vision API:** Leverage the capabilities of
  the [GPT-4 Vision API](https://platform.openai.com/docs/guides/vision) to analyze and generate insights from images.
- **Shell Commands Generation:** Effortlessly generate and execute shell commands, simplifying complex tasks and
  enhancing
  productivity.
- **Code Production:** Generate code snippets in various programming languages, making it easier to learn new languages
  or
  find solutions to coding problems.
- **ChatGPT Integration:** Utilize ChatGPT's interactive chat capabilities to refine your prompts and achieve more
  precise
  results, benefiting from the powerful language model.
- **Bash Functions and Aliases:** Seamlessly integrate SGPT responses into custom bash functions and aliases, optimizing
  your workflows and making your daily tasks more efficient.
- **OpenRouter Support:** Use [OpenRouter](https://openrouter.ai) to access various large language models (LLMs) via a
  single API, providing flexibility and convenience in your interactions with different models.

By offering these versatile features, SGPT serves as a powerful tool to enhance your overall productivity, streamline
your workflow, and simplify complex tasks.

## Installation

### Linux

SGPT has been tested on Ubuntu LTS releases and is expected to be compatible with the following Linux
distributions:

- Debian
- Ubuntu
- Arch Linux
- Fedora

To install, download the latest release from the [release page](https://github.com/tbckr/sgpt/releases) and use the
package manager specific to your distribution.

### macOS

For users with Homebrew as their package manager, run the following command in the terminal:

```shell
brew install tbckr/tap/sgpt
```

### Windows

For users with Scoop as their package manager, execute these commands in PowerShell:

```shell
scoop bucket add tbckr https://github.com/tbckr/scoop-bucket.git
scoop install tbckr/sgpt
```

### Using Go

To install SGPT with Go, based on the git tag, use this command:

```shell
go install github.com/tbckr/sgpt/v2/cmd/sgpt@latest
```

### Docker

To run SGPT with Docker, use the following command to pull the latest image:

```shell
docker pull ghcr.io/tbckr/sgpt:latest
```

Examples on how to use SGPT with Docker can be found [here](https://sgpt.readthedocs.io/en/latest/usage/docker/).

### Ansible

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

### Other platforms

For other platforms, visit the GitHub [release page](https://github.com/tbckr/sgpt/releases) and download the latest
release suitable for your system.

## Usage Guide

See the [documentation](https://sgpt.readthedocs.io/en/stable/) for detailed usage instructions.

### Getting started: Obtaining an OpenAI API Key

To use the OpenAI API, you must first obtain an API key.

1. Visit [https://platform.openai.com/overview](https://platform.openai.com/overview) and sign up for an account.
2. Navigate to [https://platform.openai.com/account/api-keys](https://platform.openai.com/account/api-keys) and generate
   a new API key.
3. On Linux or macOS: Update your `.bashrc` or `.zshrc` file to include the following export statement adding your API
   key as the value:

  ```shell
  export OPENAI_API_KEY="sk-..."
  ```

1. On Windows: [Update your environment variables](https://geekflare.com/system-environment-variables-in-windows/) to
   include the `OPENAI_API_KEY` variable with your API key as the value.

After completing these steps, you'll have an OpenAI API key that can be used to interact with the OpenAI models through
the SGPT tool.

> [!IMPORTANT]
> Your API key is sensitive information. Do not share it with anyone.

### Querying OpenAI Models

SGPT allows you to ask simple questions and receive informative answers. For example:

```shell
$ sgpt "mass of sun"
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

You can also pass prompts to SGPT using pipes:

```shell
$ echo -n "mass of sun" | sgpt
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

You can also add another prompt to the piped data by specifying the `stdin` modifier and then specifying the prompt:

```shell
$ echo "Say: Hello World!" | sgpt stdin 'Replace every "World" word with "ChatGPT"'
Hello ChatGPT!
```

If you want to stream the completion to the command line, you can add the `--stream` flag. This will stream the output
to the command line as it is generated.

### GPT-4o and GPT-4 Vision API Support

SGPT additionally facilitates the utilization of the GPT-4o and GPT-4 Vision API. Include input images using the `-i`
or `--input` flag, supporting both URLs and local images.

```shell
$ sgpt -m "gpt-4o" -i pkg/fs/testdata/marvin.jpg "what can you see on the picture?"
The picture shows a robot with a large, round head and an expressive, downward-slanting triangular eye. The body of the robot is designed with a sleek, somewhat shiny, metallic structure and it is pointing with its right hand. The design appears to be humanoid with distinct arms, legs, and a segmented torso.
$ sgpt -m "gpt-4-vision-preview" -i "https://upload.wikimedia.org/wikipedia/en/c/cb/Marvin_%28HHGG%29.jpg" "what can you see on the picture?"
The image shows a figure resembling a robot with a humanoid form. It has a
```

It is also possible to combine URLs and local images:

```shell
$ sgpt -m "gpt-4o" -i "https://upload.wikimedia.org/wikipedia/en/c/cb/Marvin_%28HHGG%29.jpg" -i pkg/fs/testdata/marvin.jpg "what is the difference between those two pictures"
The two pictures you provided appear to be identical. There are no visible differences between them. Both show the same character in the same pose with the same lighting and background.
```

You can also set the default model to GPT-4o or GPT-4 Vision by setting it in
the [configuration file](https://sgpt.readthedocs.io/en/stable/configuration/).

**Important:** The GPT-4o and GPT-4-vision API integration is currently in beta and may change in the future.

### o1 API Support

If you are already whitelisted for the o1 API, you can use it by specifying the model with the `-m` flag. You must also
provide the `--stream=false` flag to not stream the output as it is not supported by the o1 API (this is only necessary,
if you have provided the stream option via the config file).

Example:

```shell
$ sgpt -m "o1-preview" --stream=false "how many rs are in strawberry?"
There are three "r"s in the word "strawberry".
```

You can also create a bash alias to use the o1 API more easily. For example, add the following line to your `.bashrc`:

```shell
alias sgpt-o1="sgpt -m \"o1-preview\" --stream=false"
```

Then you can use the alias like this:

```shell
$ sgpt-o1 "how many rs are in strawberry?"
There are three "r"s in the word "strawberry".
```

**Important:** The o1 API does not support personas.

### OpenRouter API Support

SGPT seamlessly integrates with the [OpenRouter API](https://openrouter.ai), giving you access to a wide range of AI
models beyond OpenAI's offerings.

1. Set the OpenRouter API base URL environment variable:
   ```shell
   export OPENAI_API_BASE="https://openrouter.ai/api/v1"
   ```

2. Create an API key at [OpenRouter](https://openrouter.ai/settings/keys) and set it as your environment variable:
   ```shell
   export OPENAI_API_KEY="your_openrouter_api_key"
   ```

Once configured, you can specify any OpenRouter-supported model with the `-m` flag:

```shell
$ sgpt -m "anthropic/claude-3.7-sonnet" "mass of sun"
The mass of the Sun is approximately:

1.989 × 10^30 kilograms (kg)

This is roughly 333,000 times the mass of Earth. The Sun contains about 99.86% of all the mass in our solar system.
```

Browse the complete list of available models on the [OpenRouter models page](https://openrouter.ai/models).

> [!TIP]
> Under [Integrations](https://openrouter.ai/settings/integrations) in your OpenRouter account, you can link your
> existing OpenAI API key. This allows you to use any remaining OpenAI credits when accessing OpenAI models through
> OpenRouter.

### Chat Capabilities

SGPT provides chat functionality that enables interactive conversations with OpenAI models. You can use the `--chat`
flag to initiate and reference chat sessions.

The chat capabilities allow you to interact with OpenAI models in a more dynamic and engaging way, making it
easier to obtain relevant responses, code, or shell commands through continuous conversations.

The example below demonstrates how to fine-tune the model's responses for more targeted outcomes.

1. The first command initiates a chat session named `ls-files` and asks the model to "list all files directory":

  ```shell
  $ sgpt sh --chat ls-files "list all files directory"
  ls
  ```

1. The second command continues the conversation within the `ls-files` chat session and requests to "sort by name":

  ```shell
  $ sgpt sh --chat ls-files "sort by name"
  ls | sort
  ```

The model provides the appropriate shell command `ls | sort`, which lists all files in a directory and sorts them by
name.

### Generating and Executing Shell Commands

SGPT can generate shell commands based on your input:

```shell
$ sgpt sh "make all files in current directory read only"
chmod -R 444 *
```

You can also generate a shell command and execute it directly:

```shell
$ sgpt sh --execute "make all files in current directory read only"
chmod -R 444 *
Do you want to execute this command? (Y/n) y
```

The `sh` command is a default persona to generate shell commands. For more information on personas, see
the [docs](https://sgpt.readthedocs.io/en/latest/usage/personas/).

### Interactive Shell Sessions

Currently, SGPT does not support interactive shell sessions. However, `rlwrap` can be used to enable
interactive-like shell sessions ([source](https://github.com/tbckr/sgpt/issues/111#issuecomment-1869814041)):

```text
$ rlwrap bash -c 'echo ▶; while read in; do [[ -n "$in" ]] && echo ■ && sgpt --chat chat_name "$in" && echo ▶; done'
▶
mass of sun
■
The mass of the Sun is approximately 1.989 x 10^30 kilograms, or about 330,000 times the mass of Earth. It contains about 99.86% of the total mass of the Solar System and is by far the most dominant object in it. The Sun's mass is composed mostly of hydrogen (~74%) and helium (~24%), with the remaining 2% consisting of heavier elements.
▶
convert to earth masses
■
To convert the mass of the Sun to Earth masses, you can simply divide the Sun's mass by the mass of the Earth. Given that:


A. The Sun's mass is approximately 1.989 x 10^30 kilograms.

B. The Earth's mass is approximately 5.972 x 10^24 kilograms.

Using these values, you can calculate how many Earth masses the Sun is:

(1.989 x 10^30 kg) / (5.972 x 10^24 kg/Earth) = approximately 333,000 Earth masses

So the Sun is about 333,000 times more massive than the Earth.
▶
```

A script with automated session name generation and notification support could look like this:

```shell
#!/usr/bin/env bash

shopt -s -o errexit
shopt -s -o pipefail
shopt -s -o nounset
shopt -s inherit_errexit

export CHAT="$(date '+%Y%m%d%H%M%S%3N')_$(tr -dc 'A-Za-z' </dev/urandom | head -c 3)"
rlwrap bash -c 'echo ▶; while read in; do [[ -n "$in" ]] && echo ■ && sgpt --chat "$CHAT" "$in" && echo ▶ && notify-send --urgency=low 💬 ; done'
```

Thanks to @ilya-bystrov for coming up with this solution.

### Code Generation Capabilities

SGPT can efficiently generate code based on given instructions. For instance, to solve the classic FizzBuzz problem
using Python, simply provide the prompt as follows:

```shell
$ sgpt code "Solve classic fizz buzz problem using Python"
for i in range(1, 101):
    if i % 3 == 0 and i % 5 == 0:
        print("FizzBuzz")
    elif i % 3 == 0:
        print("Fizz")
    elif i % 5 == 0:
        print("Buzz")
    else:
        print(i)
```

SGPT will return the appropriate Python code to address the FizzBuzz problem.

The `code` command is a default persona to generate code. For more information on personas, see
the [docs](https://sgpt.readthedocs.io/en/latest/usage/personas/).

### Enhancing Your Workflow with Bash Aliases and Functions

SGPT can be further integrated into your workflow by creating bash aliases and functions. This enables you to automate
common tasks and improve efficiency when working with OpenAI models and shell commands.

Indeed, you can configure SGPT to generate your git commit message using the following bash function:

```shell
gsum() {
  commit_message="$(sgpt txt "Generate git commit message, my changes: $(git diff)")"
  printf "%s\n" "$commit_message"
  read -rp "Do you want to commit your changes with this commit message? [y/N] " response
  if [[ $response =~ ^[Yy]$ ]]; then
    git add . && git commit -m "$commit_message"
  else
    echo "Commit cancelled."
  fi
}
```

For instance, the commit message for this description and bash function would appear as follows:

```shell
$ gsum
feat: Add bash function to generate git commit messages

Added `gsum()` function to `.bash_aliases` that generates a commit message using sgpt to summarize git changes.
The user is prompted to confirm the commit message before executing `git add . && git commit -m "<commit_message>"`.
This function is meant to automate the commit process and increase productivity in daily work.

Additionally, updated the README.md file to include information about the new bash function and added a section to
showcase useful bash aliases and functions found in `.bash_aliases`.
Do you want to commit your changes with this commit message? [y/N] y
[main d6db80a] feat: Add bash function to generate git commit messages
 2 files changed, 48 insertions(+)
 create mode 100644 .bash_aliases
```

A compilation of beneficial bash aliases and functions, including an updated gsum function, is available
in [.bashrc](.bashrc).

## Acknowledgements

Inspired by [shell-gpt](https://github.com/TheR1D/shell_gpt).
