# whisper

Docker client for [OpenAI's Whisper](https://github.com/openai/whisper).

## Usage

### Setup

```bash
# From the repo root directory
docker compose build whisper
docker compose up whisper -d
```

### Run

```bash
# From the repo root directory
docker compose exec whisper \
    whisper \
    --model tiny.en \
    --language en \
    --output_format txt \
    --output_dir artifacts/whisper \
    artifacts/samples/skip-ad.wav
```
