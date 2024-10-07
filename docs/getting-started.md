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

**Note:** Your API key is sensitive information. Do not share it with anyone.

## Querying OpenAI Models

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

## GPT-4 Vision API

SGPT additionally facilitates the utilization of the GPT-4 Vision API. Include input images using the `-i` or `--input`
flag, supporting both URLs and local images.

```shell
$ sgpt -m "gpt-4-vision-preview" -i "https://upload.wikimedia.org/wikipedia/en/c/cb/Marvin_%28HHGG%29.jpg" "what can you see on the picture?"
The image shows a figure resembling a robot with a humanoid form. It has a
$ sgpt -m "gpt-4-vision-preview" -i pkg/fs/testdata/marvin.jpg "what can you see on the picture?"
The image shows a figure resembling a robot with a sleek, metallic surface. It
```

It is also possible to combine URLs and local images:

```shell
$ sgpt -m "gpt-4-vision-preview" -i "https://upload.wikimedia.org/wikipedia/en/c/cb/Marvin_%28HHGG%29.jpg" -i pkg/fs/testdata/marvin.jpg "what is the difference between those two pictures"
The two images provided appear to be identical. Both show the same depiction of a
```

To avoid specifying the `-m "gpt-4-vision-preview"` for each request, you can streamline the process by creating a bash
alias:

```shell
alias vision='sgpt -m "gpt-4-vision-preview"'
```

For more bash examples, see [.bashrc](https://github.com/tbckr/sgpt/blob/main/.bashrc).

**Important:** The GPT-4-vision API integration is currently in beta and may change in the future.

## o1 API Support

If you are already whitelisted for the o1 API, you can use it by specifying the model with the `-m` flag. You must also
provide the `--stream=false` flag to not stream the output as it is not supported by the o1 API (this is only necessary,
if you have provided the stream option via the config file).

Example:

```shell
$ sgpt -m "o1" --stream=false "how many rs are in strawberry?"
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

## Code Generation Capabilities

By adding the `code` command to your prompt, you can generate code based on given instructions by using the
`code` [persona](./usage/personas.md). For instance, to solve the classic FizzBuzz problem using Python, simply provide
the prompt as follows:

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

SGPT also supports a `shell` [persona](./usage/personas.md) that can generate shell commands based on your input:

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
