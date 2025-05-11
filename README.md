# jarvis

Jarvis is a voice assistant.

It barely works, but it's still helpful while I'm doing the dishes.

## Capabilities

1. Pause and play YouTube videos.
1. Coming soon: skip YouTube ads.

## Usage

Run Jarvis using the instructions below. Then, open a YouTube video and start doing the dishes.

Now that your hands are busy, just yell "Jarvis, pause the video" and hopefully Jarvis will catch on.

If it doesn't, yell a little longer to get it out of your system, and, then, open an issue.

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
