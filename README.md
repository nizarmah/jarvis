# jarvis

Jarvis is a voice assistant.

It's pathetic, but helpful while I'm doing the dishes.

## Capabilities

None.

## Usage

### Setup

1. Start [whisper](infra/whisper/README.md).
1. Start [ollama](infra/ollama/README.md).

### Run

Run the listener and executor in separate terminals.

#### Listener

```bash
# From the repo root directory
go run cmd/listener/main.go
```

#### Executor

```bash
# From the repo root directory
go run cmd/executor/main.go
```

## Development

### Debug

Enable debug in the `main.go` files and run as you normally would.
