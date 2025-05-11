.PHONY: env executor listener infra infra-ollama infra-whisper test-executor

# Run ---

# Start the executor service
executor:
	@echo "Starting executor service..."
	@env $(cat .env | xargs) go run cmd/executor/main.go

# Start the listener service
listener:
	@echo "Starting listener service..."
	@rm -rf artifacts/audio/*
	@env $(cat .env | xargs) go run cmd/listener/main.go

# Setup ---

# Setup --- Environment ---

# Create environment file from example
env:
	@if [ -f .env ]; then \
		echo ".env file already exists"; \
		echo "Please edit the .env file with your preferred editor"; \
		exit 0; \
	fi
	@cp example.env .env
	@echo "Created .env file from example.env"
	@echo "Please edit the env file with your preferred editor"

# Setup --- Infrastructure ---

# Setup all infrastructure
infra:
	@echo "Setting up infrastructure..."
	@make infra-ollama
	@make infra-whisper

# Prepare the ollama infrastructure
infra-ollama:
	@echo "Pulling ollama model..."
	@env $(cat .env | xargs) ollama pull $(OLLAMA_MODEL)
	@echo "Pre-loading ollama model..."
	@env $(cat .env | xargs) ollama run $(OLLAMA_MODEL) "Reply with one word. Hello."

# Prepare the whisper infrastructure
infra-whisper:
	@echo "Building whisper image..."
	@env $(cat .env | xargs) docker compose build whisper
	@echo "Starting whisper container..."
	@env $(cat .env | xargs) docker compose up whisper -d
	@echo "Pre-loading whisper model..."
	@env $(cat .env | xargs) docker compose exec whisper download-model $(WHISPER_MODEL)
	@echo "Pre-loading whisper language..."
	@env $(cat .env | xargs) docker compose exec whisper download-language $(WHISPER_LANGUAGE)
	@echo "Testing whisper container..."
	@env $(cat .env | xargs) docker compose exec whisper \
		whisper \
		--model $(WHISPER_MODEL) \
		--language $(WHISPER_LANGUAGE) \
		artifacts/samples/skip-ad.wav

# Test ---

# Test --- Executor ---

# Test the executor with a command
test-executor:
	@if [ -z "$(event)" ]; then \
		echo "Usage: make test-executor event=pause_video"; \
		exit 1; \
	fi
	@echo "$(event)" | nc $(EXECUTOR_ADDRESS)
