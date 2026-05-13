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
```

These options override the default values for the corresponding command line options.

Set `insecureAPIBase: true` to skip validation of `OPENAI_API_BASE`. Required
for local LLM setups that point at a single-label LAN hostname (e.g.
`http://thinkbox:8080/v1`); see [Query Models](usage/query-models.md#opt-out-for-lan-hostnames)
for the validation rules.
