package config

import (
	"flag"
	"os"
)

type Config struct {
	Mock           bool
	APIKey         string
	BaseURL        string
	Model          string
	EmbeddingModel string
	Output         string
}

func Parse() *Config {
	cfg := &Config{}

	flag.BoolVar(&cfg.Mock, "mock", false, "Use mock mode instead of real LLM API")
	flag.StringVar(&cfg.APIKey, "api-key", os.Getenv("OPENAI_API_KEY"), "OpenAI API key")
	flag.StringVar(&cfg.BaseURL, "base-url", os.Getenv("OPENAI_BASE_URL"), "OpenAI compatible base URL")
	flag.StringVar(&cfg.Model, "model", "gpt-4o", "Model name for LLM scoring")
	flag.StringVar(&cfg.EmbeddingModel, "embedding-model", "text-embedding-3-small", "Model name for embedding")
	flag.StringVar(&cfg.Output, "output", "output", "Output directory")
	flag.Parse()

	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}

	return cfg
}
