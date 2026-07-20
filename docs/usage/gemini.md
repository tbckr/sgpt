# Google Gemini Support

SGPT works with Google's Gemini models without any Gemini-specific code. Gemini is reachable over an OpenAI-compatible
chat completions API, so pointing `OPENAI_API_BASE` at the right endpoint is all that is required.

There are two ways to get there: directly against Google's endpoint, or through a gateway such as
[OpenRouter](openrouter.md).

## Option 1: Google's OpenAI-compatible endpoint

1. Create an API key in [Google AI Studio](https://aistudio.google.com/apikey).

2. Point SGPT at Google's OpenAI-compatible base URL and set the key:
   ```shell
   export OPENAI_API_BASE="https://generativelanguage.googleapis.com/v1beta/openai"
   export OPENAI_API_KEY="your_gemini_api_key"
   ```

3. Select a Gemini model with the `-m` flag:
   ```shell
   $ sgpt -m "gemini-3.5-flash" "mass of sun"
   The Sun's mass is approximately 1.989 × 10^30 kg — approximately 333,000 times the mass of Earth.
   ```

Note that SGPT reads the key from `OPENAI_API_KEY`, not from `GEMINI_API_KEY`. The variable name refers to the
protocol SGPT speaks, not to the provider behind it.

Browse the available model IDs in the [Gemini API model list](https://ai.google.dev/gemini-api/docs/models).

## Option 2: Through OpenRouter

If you already route through [OpenRouter](openrouter.md), Gemini models are available under the `google/` prefix — no
second API key and no config switch when you move between providers:

```shell
export OPENAI_API_BASE="https://openrouter.ai/api/v1"
export OPENAI_API_KEY="your_openrouter_api_key"

sgpt -m "google/gemini-3.5-flash" "mass of sun"
```

## Which to choose

Google's endpoint is one hop shorter and gives you Google's own free tier. OpenRouter costs a small routing margin but
lets you switch between Gemini, Claude, GPT and others by changing only the `-m` flag, and gives you a single key and
one billing account across all of them.

## Limitations

The OpenAI compatibility layer covers chat completions, which is what SGPT uses — including personas, chat sessions and
streaming. Gemini-specific features that have no OpenAI equivalent, such as search grounding or the Files API, are not
reachable through it. If you need those, use Google's native SDK directly rather than SGPT.
