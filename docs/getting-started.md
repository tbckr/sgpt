# Getting Started

## Obtaining an OpenAI API Key

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

## Querying OpenAI Models

SGPT allows you to ask simple questions and receive informative answers. For example:

```shell
$ sgpt txt "mass of sun"
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

You can also pass prompts to SGPT using pipes:

```shell
$ echo -n "mass of sun" | sgpt txt
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

## Code Generation Capabilities

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

## Generating and Executing Shell Commands

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

## Chat Capabilities

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
