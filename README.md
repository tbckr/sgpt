# SGPT

A command-line interface (CLI) tool to access the OpenAI models via the command line.

Developed with the help of [sgpt](https://github.com/tbckr/sgpt).

## Install

Install via go:

```shell
go install github.com/tbckr/sgpt/cmd/sgpt@v1.5.0
```

Download latest release from Github [here](https://github.com/tbckr/sgpt/releases).

## Usage

Ask simple questions:

```shell
$ sgpt txt "mass of sun"
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

Pipe prompts to sgpt:

```shell
$ echo -n "mass of sun" | sgpt txt
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

Use the docker image to run queries:

```shell
$ docker pull ghcr.io/tbckr/sgpt:latest
$ docker run --rm -e OPENAI_API_KEY=${OPENAI_API_KEY} ghcr.io/tbckr/sgpt:latest txt "mass of sun"
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

### Shell commands

Generate shell commands:

```shell
$ sgpt sh "make all files in current directory read only"
chmod -R 444 *
```

Generate shell command and execute it:

```shell
$ sgpt sh --execute "make all files in current directory read only"
chmod -R 444 *
Do you want to execute this command? (Y/n) y
```

Furthermore, you can create bash aliases and functions to automate your workflows with the help of sgpt.

In fact, you can tell sgpt to generate your git commit message with the following bash function:

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

For example, the commit for this description and bash function would look like this:

```shell
$ gsum
feat: Add bash function to generate git commit messages

Added `gsum()` function to `.bash_aliases` that generates a commit message using sgpt to summarize git changes. The user is prompted to confirm the commit message before executing `git add . && git commit -m "<commit_message>"`. This function is meant to automate the commit process and increase productivity in daily work.

Additionally, updated the README.md file to include information about the new bash function and added a section to showcase useful bash aliases and functions found in `.bash_aliases`.
Do you want to commit your changes with this commit message? [y/N] y
[main d6db80a] feat: Add bash function to generate git commit messages
 2 files changed, 48 insertions(+)
 create mode 100644 .bash_aliases
```

A collection of useful bash aliases and functions can be found in [.bash_aliases](.bash_aliases).

### Code generation

Generate code:

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

### Text to Image

Create images via the DALLE api:

```shell
$ sgpt image "v for vendetta"
<image url>
```

Create images via the DALLE api and download it into the current working directory:

```shell
$ sgpt image --download "v for vendetta"
1c561592-6d93-438f-9bee-d96c898a31a8.png
```

## Acknowledgements

Inspired by [shell-gpt](https://github.com/TheR1D/shell_gpt).