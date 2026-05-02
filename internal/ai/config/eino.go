package config

import "os"

// AIConfig holds configuration for AI features.
type AIConfig struct {
	// LLM provider settings
	Provider   string // "openai", "doubao", "deepseek"
	APIKey     string
	BaseURL    string
	ModelName  string
	MaxTokens  int
	Temperature float32
}

// DefaultConfig returns AI config from environment variables.
func DefaultConfig() *AIConfig {
	return &AIConfig{
		Provider:    getEnv("AI_PROVIDER", "ollama"),
		APIKey:      getEnv("AI_API_KEY", ""),
		BaseURL:     getEnv("AI_BASE_URL", "http://localhost:11434/v1"),
		ModelName:   getEnv("AI_MODEL", "qwen3.5:9b"),
		MaxTokens:   4096,
		Temperature: 0.7,
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
