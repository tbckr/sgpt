# Configuration

SGPT can be configured via a config file in your default config directory. The default config directory
is `~/.config/sgpt/` on Linux and MacOS and `%APPDATA%/sgpt/` on Windows. The config file is named `config.yaml`.

The config file is a YAML file with the following structure:

```yaml
maxTokens: 2048
model: "gpt-4"
temperature: "1"
topP: "1"
```

These options override the default values for the corresponding command line options.
