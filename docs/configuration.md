# Configuration

SGPT can be configured via a config file in your default config directory. The default config directory
is `~/.config/sgpt/` on Linux, `~/Library/Application Support/sgpt/` on macOS and `%APPDATA%/sgpt/` on Windows. The
config file is named `config.yaml`.

The config file is a YAML file with the following structure:

```yaml
stream: false
maxTokens: 2048
model: "gpt-4"
temperature: "1"
topP: "1"
insecureAPIBase: false
api_key: "sk-..."
base_url: "https://api.openai.com/v1"
```

These options override the default values for the corresponding command line options.

## API key and base URL

`api_key` and `base_url` let you configure OpenAI API access without exporting
`OPENAI_API_KEY` / `OPENAI_API_BASE` into your shell environment. The
environment variables always take precedence when set — the config file is
only consulted as a fallback. `base_url` is validated with the same rules as
`OPENAI_API_BASE`; see [Query Models](usage/query-models.md#validation-rules).

Since `config.yaml` would then contain your API key in plain text, restrict
its permissions, e.g. `chmod 600 ~/.config/sgpt/config.yaml`.

Set `insecureAPIBase: true` to skip validation of the base URL. Required
for local LLM setups that point at a single-label LAN hostname (e.g.
`http://thinkbox:8080/v1`); see [Query Models](usage/query-models.md#opt-out-for-lan-hostnames)
for the validation rules.
