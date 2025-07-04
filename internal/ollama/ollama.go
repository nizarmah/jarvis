// Package ollama provides a client for the Ollama API.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// ClientConfig is the configuration for the Ollama client.
type ClientConfig struct {
	// Debug enables logging while prompting the LLM.
	Debug bool
	// Model is the name of the model to use.
	Model string
	// URL is the URL of the Ollama server.
	URL string
}

// Client is a client for the Ollama API.
type Client struct {
	debug bool
	model string
	url   string
}

// GenerateInput is the input for the generate endpoint.
type generateInput struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// GenerateResult is the result for the generate endpoint.
type generateResult struct {
	Response string `json:"response"`
}

// NewClient creates a new Ollama client with default config.
func NewClient(cfg ClientConfig) *Client {
	return &Client{
		debug: cfg.Debug,
		model: cfg.Model,
		url:   cfg.URL,
	}
}

// Prompt sends a prompt to the LLM and returns the response.
func (c *Client) Prompt(ctx context.Context, prompt string) (string, error) {
	req, err := c.buildGenerateRequest(ctx, generateInput{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	var parsed generateResult
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if c.debug {
		log.Println(fmt.Sprintf(
			"prompt: %s, ollama response: %s",
			strings.ReplaceAll(prompt, "\n", " "),
			strings.ReplaceAll(parsed.Response, "\n", " "),
		))
	}

	result := parsed.Response

	return result, nil
}

// buildGenerateRequest builds a request for the generate endpoint.
func (c *Client) buildGenerateRequest(ctx context.Context, input generateInput) (*http.Request, error) {
	jsonBody, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to encode JSON: %w", err)
	}

	body := bytes.NewReader(jsonBody)
	url := fmt.Sprintf("%s/api/generate", c.url)

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}
