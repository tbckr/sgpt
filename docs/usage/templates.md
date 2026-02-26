# Prompt Templating

The `--template` / `-T` flag enables data-driven prompting: pipe YAML or JSON variables via stdin
and inject them into a Go template string.

## Basic Usage

```shell
$ echo "name: Dave\ncountry: France" | sgpt --template "What would {{ .name }} be called in {{ .country }}?"
```

## With a Persona

The first positional argument is treated as a persona name:

```shell
$ echo "lang: Python" | sgpt code --template "Write a hello world program in {{ .lang }}"
```

## JSON Input

Both YAML and JSON are accepted:

```shell
$ echo '{"lang": "Python"}' | sgpt code --template "Write hello world in {{ .lang }}"
```

## Template Syntax

Uses Go's `text/template` — the same engine used in [personas](./personas.md).
Missing variables cause an error; extra variables are silently ignored.

## Constraints

- Requires piped input (`--template` without a pipe returns an error)
- Cannot be combined with `--execute` (stdin is consumed by template variables)
- With one positional arg: treated as persona name (`sgpt code --template "..."`)
- With two positional args: error (template replaces the prompt position)
