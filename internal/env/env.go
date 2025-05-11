// Package env holds relevant env vars in a struct.
package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Env holds relevant env variables.
type Env struct {
	AudioProcessorDebug bool
	CommandDebug        bool
	CombinerDebug       bool
	CombinerOutputDir   string
	ExecutorAddress     string
	ExecutorDebug       bool
	MessageHandlerDebug bool
	OllamaDebug         bool
	OllamaModel         string
	OllamaTimeout       time.Duration
	OllamaURL           string
	RecorderDebug       bool
	RecorderOutputDir   string
	WhisperDebug        bool
	WhisperModel        string
	WhisperLanguage     string
	WhisperOutputDir    string
}

// Init reads env vars.
func Init() (*Env, error) {
	var (
		env Env
		err error
	)

	env.AudioProcessorDebug, err = lookupBool("AUDIO_PROCESSOR_DEBUG")
	if err != nil {
		return nil, err
	}

	env.CommandDebug, err = lookupBool("COMMAND_DEBUG")
	if err != nil {
		return nil, err
	}

	env.CombinerDebug, err = lookupBool("COMBINER_DEBUG")
	if err != nil {
		return nil, err
	}

	env.CombinerOutputDir, err = lookup("COMBINER_OUTPUT_DIR")
	if err != nil {
		return nil, err
	}

	env.ExecutorAddress, err = lookup("EXECUTOR_ADDRESS")
	if err != nil {
		return nil, err
	}

	env.ExecutorDebug, err = lookupBool("EXECUTOR_DEBUG")
	if err != nil {
		return nil, err
	}

	env.MessageHandlerDebug, err = lookupBool("MESSAGE_HANDLER_DEBUG")
	if err != nil {
		return nil, err
	}

	env.OllamaDebug, err = lookupBool("OLLAMA_DEBUG")
	if err != nil {
		return nil, err
	}

	env.OllamaModel, err = lookup("OLLAMA_MODEL")
	if err != nil {
		return nil, err
	}

	env.OllamaTimeout, err = lookupDuration("OLLAMA_TIMEOUT", time.Second)
	if err != nil {
		return nil, err
	}

	env.OllamaURL, err = lookup("OLLAMA_URL")
	if err != nil {
		return nil, err
	}

	env.RecorderDebug, err = lookupBool("RECORDER_DEBUG")
	if err != nil {
		return nil, err
	}

	env.RecorderOutputDir, err = lookup("RECORDER_OUTPUT_DIR")
	if err != nil {
		return nil, err
	}

	env.WhisperDebug, err = lookupBool("WHISPER_DEBUG")
	if err != nil {
		return nil, err
	}

	env.WhisperModel, err = lookup("WHISPER_MODEL")
	if err != nil {
		return nil, err
	}

	env.WhisperLanguage, err = lookup("WHISPER_LANGUAGE")
	if err != nil {
		return nil, err
	}

	env.WhisperOutputDir, err = lookup("WHISPER_OUTPUT_DIR")
	if err != nil {
		return nil, err
	}

	return &env, nil
}

// lookup helps verifying an env var exists.
func lookup(s string) (string, error) {
	value, ok := os.LookupEnv(s)
	if !ok {
		return "", fmt.Errorf("env var %s not set", s)
	}

	return value, nil
}

// lookupBool helps verifying an env var exists and casts its value as bool.
func lookupBool(s string) (bool, error) {
	value, err := lookup(s)
	if err != nil {
		return false, err
	}

	return value == "true", nil
}

// lookupInt helps verifying an env var exists and casts its value as int.
func lookupInt(s string) (int, error) {
	value, err := lookup(s)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(value)
}

// lookupDuration helps verifying an env var exists and casts its value as time.Duration.
func lookupDuration(s string, unitTime time.Duration) (time.Duration, error) {
	value, err := lookupInt(s)
	if err != nil {
		return 0, err
	}

	return time.Duration(value) * unitTime, nil
}
