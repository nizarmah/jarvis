-include .env
export

.PHONY: env executor listener infra infra-ollama infra-whisper test-executor

# Run ---

# Start the executor service
executor:
	@echo "Starting executor service..."
	@go run cmd/executor/main.go

# Start the listener service
listener:
	@echo "Starting listener service..."
	@rm -rf artifacts/audio/*
	@go run cmd/listener/main.go

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

# Setup the ollama infrastructure
infra-ollama:
	@echo "Pulling ollama model..."
	@ollama pull $(OLLAMA_MODEL)
	@echo "Pre-loading ollama model..."
	@ollama run $(OLLAMA_MODEL) "Reply with one word. Hello."

# Setup the whisper infrastructure
infra-whisper:
	@echo "Building whisper image..."
	@docker compose build whisper
	@echo "Starting whisper container..."
	@docker compose up whisper -d
	@echo "Pre-loading whisper model..."
	@docker compose exec whisper download-model $(WHISPER_MODEL)
	@echo "Pre-loading whisper language..."
	@docker compose exec whisper download-language $(WHISPER_LANGUAGE)
	@echo "Testing whisper container..."
	@docker compose exec whisper \
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
