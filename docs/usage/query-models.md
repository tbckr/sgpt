# Query OpenAI Models

## Ask Questions

SGPT allows you to ask simple questions and receive informative answers. For example:

```shell
$ sgpt "mass of sun"
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

If you want to stream the completion to the command line, you can add the `--stream` flag. This will stream the output
to the command line as it is generated.

You can also pass prompts to SGPT using pipes:

```shell
$ echo -n "mass of sun" | sgpt
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

You can also use the `--clipboard` flag to write the answer to the clipboard:

```shell
$ sgpt --clipboard "mass of sun"
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

## Generate Code

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

The second argument is the code command that identifies the [persona](personas.md) to use.

SGPT will return the appropriate Python code to address the FizzBuzz problem.

## Generate and Execute Shell Commands

SGPT also supports a shell [persona](personas.md) that can generate shell commands based on your input:

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

## Override OpenAI API base URL

You can override the OpenAI base URL by setting the `OPENAI_API_BASE` environment variable:

```shell
export OPENAI_API_BASE=https://api.openai.com/v1
```

### Validation rules

To prevent an attacker-set environment variable from silently redirecting your
API key to a hostile endpoint, the value is validated before use:

- `https://<any-host>` — always allowed.
- `http://localhost`, `http://127.x.x.x`, `http://[::1]` — allowed (loopback).
- `http://10.x.x.x`, `http://172.16-31.x.x`, `http://192.168.x.x`,
  `http://[fd00::/8]` — allowed (RFC1918 + RFC4193 ULA).
- Everything else over plain `http://` is rejected, including link-local
  addresses such as `http://169.254.169.254/` (cloud instance metadata).

### Opt-out for LAN hostnames

Setups that point at a single-label hostname like `http://thinkbox:8080/v1`
can't be classified by IP literal and must opt out of validation explicitly.
Either of the following enables the opt-out (the flag takes precedence over
the config file):

```shell
# 1. CLI flag
sgpt --insecure-api-base "..."

# 2. Config file (~/.config/sgpt/config.yaml)
insecureAPIBase: true
```

There is deliberately no environment-variable form of this opt-out: the
[`OPENAI_API_BASE` SSRF report](https://github.com/tbckr/sgpt/issues/358) that
introduced this validation assumes an attacker who controls the environment,
and an env-var opt-out would undo that protection. The flag and config file
both require either CLI access or write access to your config directory.

When the opt-out is active, a warning is logged once per invocation so you know
validation was skipped.
