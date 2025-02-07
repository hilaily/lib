package llm

import "os"

func defaultOption() *Option {
	return &Option{
		baseURL:   os.Getenv("LLM_BASE_URL"),
		apiKey:    os.Getenv("LLM_API_KEY"),
		prompt:    "You are a helpful assistant.",
		model:     "gpt-4o",
		maxTokens: 1000,
	}
}

type ClientOption func(*Option)

type Option struct {
	baseURL   string
	apiKey    string
	prompt    string
	model     string
	maxTokens int
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Option) {
		c.baseURL = baseURL
	}
}

func WithAPIKey(apiKey string) ClientOption {
	return func(c *Option) {
		c.apiKey = apiKey
	}
}

func WithPrompt(prompt string) ClientOption {
	return func(c *Option) {
		c.prompt = prompt
	}
}

func WithModel(model string) ClientOption {
	return func(c *Option) {
		c.model = model
	}
}
