# Query OpenAI Models

## Ask Questions

SGPT allows you to ask simple questions and receive informative answers. For example:

```shell
$ sgpt "mass of sun"
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

You can also use the `--clipboard` flag to write the answer to the clipboard:

```shell
$ sgpt --clipboard "mass of sun"
The mass of the sun is approximately 1.989 x 10^30 kilograms.
```

In addition, you can pass a file via pipes and add your prompt to ask a question about the file or have it summarized
for you.

```shell
cat yourfile.txt | sgpt "summarize everything above"
```

The prompt passed as an argument is appended to the data that is piped in. Keep in mind: Depending on the model you are
using, there are different limits to the number of tokens you can use in your prompt.

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
