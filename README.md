# chud

## LLM week summaries

The dashboard can show a short, LLM-written summary of each week. It is optional: without `LLM_URL` the card is hidden and the rest of the app works as usual.

Any server with an OpenAI-compatible `/v1/chat/completions` endpoint works, e.g. Ollama or llama.cpp.

| Variable | Example | Notes |
|---|---|---|
| `LLM_URL` | `http://host.docker.internal:11434` | Base URL of the model server; empty disables summaries |
| `LLM_MODEL` | `qwen3:4b` | Model name (required by Ollama) |
| `LLM_TIMEOUT` | `3m` | Time limit for one summary |

When the app runs in Docker and the model on the same machine, `localhost` inside the container is the container itself. Use `host.docker.internal` (already mapped in `docker-compose.yml`) and make the model listen beyond 127.0.0.1, e.g. `OLLAMA_HOST=0.0.0.0 ollama serve`. Keep that port closed to the outside.

Summaries are cached in memory and in the `week_summaries` table: past weeks are kept, the current week is refreshed after 24 hours. Anyone can regenerate a week once every 5 minutes.
