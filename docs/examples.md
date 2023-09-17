# Examples

## Enhancing Your Workflow with Bash Aliases and Functions

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
in [.bash_aliases](https://github.com/tbckr/sgpt/blob/main/.bash_aliases).

## Using Personas to Generate More Accurate Responses

SGPT allows you to specify a persona to generate more accurate responses. This is particularly useful when you want to
obtain responses that are more relevant to a specific topic or domain.

A simple example is an echo service. This persona will always return `echo` as the response to any question.

1. Create a `personas` directory in your SGPT config directory. For example, `~/.config/sgpt/personas` on Linux. Then
create a file named `echo` with the following content:

```text
You are a echo service. Answer all questions with "echo" in lower case.
```

1. Now, when you run the following command, you will always get `echo` as the response:

```shell
$ sgpt echo "Say something"
echo
```
