# SGPT

SGPT (*aka shell-gpt*) is a powerful command-line interface (CLI) tool designed for seamless interaction with OpenAI
models directly from your terminal. Effortlessly run queries, generate shell commands or code, create images from text,
and more, using simple commands. Streamline your workflow and enhance productivity with this powerful and user-friendly
CLI tool.

Developed with the help of [SGPT](https://github.com/tbckr/sgpt).

This is a Go implementation. For the original Python implementation,
visit [shell-gpt](https://github.com/TheR1D/shell_gpt). Please keep this in mind when reporting issues.

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
  - [Other platforms](#other-platforms)
- [Usage Guide](#usage-guide)
  - [Getting started: Obtaining an OpenAI API Key](#getting-started-obtaining-an-openai-api-key)
  - [Querying OpenAI Models](#querying-openai-models)
  - [Chat Capabilities](#chat-capabilities)
  - [Running Queries with Docker](#running-queries-with-docker)
  - [Saving Chat Sessions in Docker](#saving-chat-sessions-in-docker)
  - [Generating and Executing Shell Commands](#generating-and-executing-shell-commands)
  - [Enhancing Your Workflow with Bash Aliases and Functions](#enhancing-your-workflow-with-bash-aliases-and-functions)
  - [Code Generation Capabilities](#code-generation-capabilities)
- [Acknowledgements](#acknowledgements)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Features

- Instant Answers: Obtain quick and accurate responses to simple questions directly in your shell, streamlining your
  workflow.
- Shell Commands Generation: Effortlessly generate and execute shell commands, simplifying complex tasks and enhancing
  productivity.
- Code Production: Generate code snippets in various programming languages, making it easier to learn new languages or
  find solutions to coding problems.
- ChatGPT Integration: Utilize ChatGPT's interactive chat capabilities to refine your prompts and achieve more precise
  results, benefiting from the powerful language model.
- Image Generation with DALLE: Create images from textual prompts using the DALLE API, expanding the range of tasks you
  can accomplish with the tool.
- Bash Functions and Aliases: Seamlessly integrate SGPT responses into custom bash functions and aliases, optimizing
  your workflows and making your daily tasks more efficient.

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
go install github.com/tbckr/sgpt/cmd/sgpt@latest
```

### Docker

To run SGPT with Docker, use the following command to pull the latest image:

```shell
docker pull ghcr.io/tbckr/sgpt:latest
```

Examples on how to use SGPT with Docker can be found [here](#running-queries-with-docker).

### Other platforms

For other platforms, visit the GitHub [release page](https://github.com/tbckr/sgpt/releases) and download the latest
release suitable for your system.

## Usage Guide

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

### Querying OpenAI Models

SGPT allows you to ask simple questions and receive informative answers. For example:

```shell
$ sgpt "mass of sun"
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

You can also pass prompts to SGPT using pipes:

```shell
$ echo -n "mass of sun" | sgpt txt
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

### Chat Capabilities

SGPT provides chat functionality that enables interactive conversations with OpenAI models. You can use the `--chat`
flag with the `txt`, `sh`, and `code` subcommands to initiate and reference chat sessions.

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

To manage active chat sessions, use the `sgpt chat` command. Here are the available options for chat session management:

- `sgpt chat ls`: List all active chat sessions.
- `sgpt chat show <chat session>`: Display the content of a specific chat session.
- `sgpt chat rm <chat session>`: Remove a chat session.
- `sgpt chat rm --all`: Delete all chat sessions.

### Running Queries with Docker

For users who prefer to use Docker, SGPT provides a Docker image:

1. Pull the latest Docker image:

  ```shell
  docker pull ghcr.io/tbckr/sgpt:latest
  ```

1. Run queries using the Docker image:

  ```shell
  $ docker run --rm -e OPENAI_API_KEY=${OPENAI_API_KEY} ghcr.io/tbckr/sgpt:latest txt "mass of sun"
  The mass of the sun is approximately 1.989 x 10^30 kilograms.
  ```

### Saving Chat Sessions in Docker

When using SGPT within a Docker container, you can mount a local folder to the container's `/home/nonroot` path to save
and persist all active chat sessions. This allows you to maintain your chat history and resume previous conversations
across different container instances.

To mount a local folder and save chat sessions, follow these steps:

1. Pull the SGPT Docker image:

  ```shell
  docker pull ghcr.io/tbckr/sgpt:latest
  ```

1. Create a local folder to store your chat sessions, e.g. `sgpt-chat-sessions`:

  ```shell
  mkdir sgpt-chat-sessions
  ```

1. Change the permissions of the folder to the nonroot user of the Docker image:

  ```shell
  sudo chown 65532:65532 sgpt-chat-sessions
  ```

1. Run the Docker container with the local folder mounted to `/home/nonroot`:

  ```shell
  $ docker run --rm -e OPENAI_API_KEY=${OPENAI_API_KEY} -v $(pwd)/sgpt-chat-sessions:/home/nonroot ghcr.io/tbckr/sgpt:latest txt "mass of sun"
  The mass of the sun is approximately 1.99 x 10^30 kilograms.
  $ docker run --rm -e OPENAI_API_KEY=${OPENAI_API_KEY} -v $(pwd)/sgpt-chat-sessions:/home/nonroot ghcr.io/tbckr/sgpt:latest txt "convert to earth masses"
  To convert the mass of the sun to earth masses, we need to divide it by the mass of the Earth:
  1.99 x 10^30 kg / 5.97 x 10^24 kg = 333,000 Earth masses (rounded to the nearest thousand) 
  So the mass of the sun is about 333,000 times greater than the mass of the Earth.
  ```

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
in [.bash_aliases](.bash_aliases).

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

SGPT will return the appropriate Python code to address the FizzBuzz problem

## Acknowledgements

Inspired by [shell-gpt](https://github.com/TheR1D/shell_gpt).
