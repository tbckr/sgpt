# SGPT

A command-line interface (CLI) tool to access the OpenAI models via the command line.

Rewrite of [shell-gpt](https://github.com/TheR1D/shell_gpt) in [go](https://go.dev/).

## Install

Install via go:

```shell
go install github.com/tbckr/sgpt/cmd/sgpt@latest
```

## Usage

Ask simple questions:

```shell
$ sgpt "mass of sun"
The mass of the Sun is approximately 1.989 x 10^30 kilograms, or about 333,000 times the mass of the Earth.
```

Generate shell commands:

```shell
$ sgpt --shell "make all files in current directory read only"
chmod -R 444 *
```

Generate shell command and execute it:

```shell
$ sgpt --shell --execute "make all files in current directory read only"
chmod -R 444 *
Do you want to execute this command? (Y/n) y
```

Generate code:

```shell
$ sgpt --code "Solve classic fizz buzz problem using Python"
Here's the Python code for the classic Fizz Buzz problem:

for i in range(1, 101):
    if i % 3 == 0 and i % 5 == 0:
        print("FizzBuzz")
    elif i % 3 == 0:
        print("Fizz")
    elif i % 5 == 0:
        print("Buzz")
    else:
        print(i)

This code will print the numbers from 1 to 100, replacing multiples of 3 with "Fizz", multiples of 5 with "Buzz", and multiples of both 3 and 5 with "FizzBuzz".
```