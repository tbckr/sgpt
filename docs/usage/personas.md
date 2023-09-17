# Personas

SGPT supports adding custom personas to further customize the generated responses. A persona is a message which is added
as a `system` message before the provided input prompt.

The personas are stored in the `~/.config/sgpt/personas/` directory on Linux and MacOS and `%APPDATA%/sgpt/personas/` on
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
