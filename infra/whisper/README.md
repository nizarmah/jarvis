# whisper

Docker client for [OpenAI's Whisper](https://github.com/openai/whisper).

## Usage

### Setup

```bash
# From the root directory
docker compose build whisper
docker compose up whisper -d
```

### Run

```bash
# From the root directory
docker compose exec whisper \
    whisper \
    --model tiny.en \
    --output_format txt \
    artifacts/samples/skip-ad.wav
```
