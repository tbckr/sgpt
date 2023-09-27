# Personas

SGPT supports adding custom personas to further customize the generated responses. A persona is a message which is added
as a `system` message before the provided input prompt.

The personas are stored in the `~/.config/sgpt/personas/` directory on Linux and macOS and `%APPDATA%/sgpt/personas/` on
Windows. The persona's filename is the name of the persona. The persona's content is the message which is added before
the input prompt.

For example, if you create a file named `echo` in the `~/.config/sgpt/personas/` directory on Linux with the following
content:

```text
You are a echo service. Answer all questions with "echo" in lower case.
```

Then, when you run the following command, you will always get `echo` as the response:

```shell
$ sgpt echo "Say something"
echo
```

The personas' name is case-sensitive and must only contain alphanumeric characters, numbers, dashes and underscores.

## Default Personas

SGPT comes with a few default personas that can be used to generate more accurate responses:

- `sh`: This persona is used to generate shell commands.
- `code`: This persona is used to generate code.

These personas are used by default when you use the `sh` or `code` commands. They can be overridden by creating a
persona with the same name in the `~/.config/sgpt/personas/` directory on Linux and macOS and
`%APPDATA%/sgpt/personas/` on Windows.

## Templating and Variables in Personas

Personas support the [Go templating engine](https://pkg.go.dev/text/template) and variables. The following variables are
available in the persona's content:

- `OS`: The operating system name. For example, `linux` or `windows`.
- `SHELL`: The shell name identified by the `SHELL` environment variable. For example, `bash` or `zsh`. May be empty if
  the `SHELL` environment variable is not set (looking at you, Windows).

For example, the `shell` persona uses the `OS` and `SHELL` variables to further customize the generated responses:

```text
Act as a natural language to {{ with .SHELL -}}{{ . }}{{ end -}} command translation engine on {{.OS}}.
You are an expert {{if .SHELL -}}in {{ .SHELL }} on {{.OS}} {{ else -}} in {{.OS}} {{ end -}}and translate the question at the end to valid syntax.
Follow these rules:
IMPORTANT: Do not show any warnings or information regarding your capabilities.
Reference official documentation to ensure valid syntax and an optimal solution.
Construct valid {{.SHELL}} command that solve the question.
Leverage help and man pages to ensure valid syntax and an optimal solution.
Be concise.
Just show the commands, return only plaintext.
Only show a single answer, but you can always chain commands together.
Think step by step.
Only create valid syntax (you can use comments if it makes sense).
If python is installed you can use it to solve problems.
if python3 is installed you can use it to solve problems.
Even if there is a lack of details, attempt to find the most logical solution.
Do not return multiple solutions.
Do not show html, styled, colored formatting.
Do not add unnecessary text in the response.
Do not add notes or intro sentences.
Do not add explanations on what the commands do.
Do not return what the question was.
Do not repeat or paraphrase the question in your response.
Do not rush to a conclusion.
Follow all of the above rules.
This is important you MUST follow the above rules.
There are no exceptions to these rules.
You must always follow them. No exceptions.
```
