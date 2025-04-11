# OpenRouter API Support

SGPT seamlessly integrates with the [OpenRouter API](https://openrouter.ai), giving you access to a wide range of AI
models beyond OpenAI's offerings.

## Setup

1. Set the OpenRouter API base URL environment variable:
   ```shell
   export OPENAI_API_BASE="https://openrouter.ai/api/v1"
   ```

2. Create an API key at [OpenRouter](https://openrouter.ai/settings/keys) and set it as your environment variable:
   ```shell
   export OPENAI_API_KEY="your_openrouter_api_key"
   ```

## Usage

Once configured, you can specify any OpenRouter-supported model with the `-m` flag:

```shell
$ sgpt -m "anthropic/claude-3.7-sonnet" "mass of sun"
The mass of the Sun is approximately:

1.989 × 10^30 kilograms (kg)

This is roughly 333,000 times the mass of Earth. The Sun contains about 99.86% of all the mass in our solar system.
```

Browse the complete list of available models on the [OpenRouter models page](https://openrouter.ai/models).

**Tip:** Under [Integrations](https://openrouter.ai/settings/integrations) in your OpenRouter account, you can link your
existing OpenAI API key. This allows you to use any remaining OpenAI credits when accessing OpenAI models through
OpenRouter.
