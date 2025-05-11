# ollama

Instructions for [Ollama](https://github.com/ollama/ollama).

## Usage

### Install

Follow the [official docs](https://github.com/ollama/ollama/tree/main/docs).

### Setup

```bash
# Start ollama
ollama serve

# Pull the model
ollama pull tinyllama
```

### Run

```bash
# Pre-load the model, since ollama lazy-loads it
curl -X POST http://localhost:11434/api/generate \
    -H "Content-Type: application/json" \
    -d '{"model": "tinyllama", "stream": false, "prompt": "Reply with one word. Hello."}'
```
