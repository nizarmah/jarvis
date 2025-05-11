# jarvis

Jarvis is a voice assistant.

It's pathetic, but helpful while I'm doing the dishes.

## Capabilities

1. Pause and play YouTube videos.
1. Coming soon: skip YouTube ads.

## Setup

### Environment

1. Create `.env` file from [`example.env`](./example.env).
   ```bash
   # From the repo root directory
   make env
   ```
1. Modify `.env` with your preferred editor.

### Infrastructure

#### Whisper

Follow the instructions on [infra/whisper](./infra/whisper/README.md).

#### Ollama

Follow the instructions on [infra/ollama](./infra/ollama/README.md).

## Run

#### Executor

Run the executor first.

```bash
# From the repo root directory
make executor
```

#### Listener

Run the listener second.

```bash
# From the repo root directory
make listener
```
