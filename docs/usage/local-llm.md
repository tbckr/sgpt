# Local LLM Support

SGPT works with any OpenAI-compatible local backend — Ollama, LiteLLM, vLLM, llama.cpp and others — by pointing
`OPENAI_API_BASE` at the backend's URL. No SGPT-side configuration beyond those two environment variables is needed.

## Setup

```shell
export OPENAI_API_KEY="dummy"  # most local backends ignore the key, but the variable must be set
export OPENAI_API_BASE="http://localhost:11434/v1"  # e.g. Ollama

sgpt -m "llama3" "mass of sun"
```

SGPT requires `OPENAI_API_KEY` to be present even when the backend ignores it, so give it any non-empty value.

Pick the model name your backend exposes — for Ollama that is whatever `ollama list` shows.

## Common backends

| Backend | Default base URL |
| --- | --- |
| Ollama | `http://localhost:11434/v1` |
| LiteLLM proxy | `http://localhost:4000/v1` |
| vLLM | `http://localhost:8000/v1` |
| llama.cpp server | `http://localhost:8080/v1` |

Check your backend's own documentation if you changed its port.

## Plain HTTP is allowed for local addresses

`OPENAI_API_BASE` is validated before use, so that an attacker who can set environment variables cannot silently
redirect your API key to a hostile endpoint. Local setups are unaffected: plain `http://` is accepted for loopback and
private network ranges, which covers everything in the table above.

## LAN hostnames need an explicit opt-out

A single-label hostname such as `http://thinkbox:8080/v1` cannot be classified as private without resolving it first,
so it is rejected by default. Opt out explicitly with either:

```shell
# CLI flag
sgpt --insecure-api-base "..."

# or in ~/.config/sgpt/config.yaml
insecureAPIBase: true
```

Using an IP literal instead — `http://192.168.1.50:8080/v1` — avoids the opt-out entirely, since it can be classified
directly.

See [Query Models — Override OpenAI API base URL](query-models.md#override-openai-api-base-url) for the full validation
rules and the reasoning behind them.
