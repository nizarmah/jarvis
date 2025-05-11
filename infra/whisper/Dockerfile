# Whisper requires Python 3.8-3.11.
FROM python:3.11-slim

WORKDIR /app

# Install Whisper dependencies.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
    ffmpeg git build-essential

# Install more Whisper dependencies.
RUN pip install --no-cache-dir setuptools-rust

# Install Whisper.
RUN pip install --no-cache-dir openai-whisper

# Keep container alive.
CMD ["tail", "-f", "/dev/null"]
