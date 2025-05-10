# ollama

Docker client for [Ollama](https://github.com/ollama/ollama).

## Usage

### Setup

```bash
# From the repo root directory
docker compose build ollama
docker compose up ollama -d

# Preload the model into the image.
docker compose exec ollama ollama pull tinyllama
```

### Run

```bash
curl -X POST http://localhost:11434/api/generate \
    -H "Content-Type: application/json" \
    -d '{"model": "tinyllama", "stream": false, "prompt": "Reply with one word. Hello."}'
```
