# ollama

We use [ollama](https://github.com/ollama/ollama) to convert audio transcripts to commands, even if the transcript is not perfect.

## Install

Follow the [official docs](https://github.com/ollama/ollama/tree/main/docs).

Don't forget to start ollama, if it's not a system service.

```bash
ollama serve
```

## Environment

If you haven't already, setup the environment file.

## Setup

```bash
# From the repo root directory
make infra-ollama
```
