# Chat Capabilities

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

## Interactive Shell Sessions

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

Thanks to [@ilya-bystrov](https://github.com/ilya-bystrov) for coming up with this solution.
