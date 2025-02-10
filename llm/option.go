package llm

import (
	"fmt"
	"os"

	"github.com/flosch/pongo2/v6"
)

func defaultOption() *Option {
	return &Option{
		baseURL:   os.Getenv("LLM_BASE_URL"),
		apiKey:    os.Getenv("LLM_API_KEY"),
		prompt:    "You are a helpful assistant.",
		model:     "gpt-4o",
		maxTokens: 1000,
	}
}

type ClientOption func(*Option) error

type Option struct {
	baseURL   string
	apiKey    string
	prompt    string
	model     string
	maxTokens int
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Option) error {
		c.baseURL = baseURL
		return nil
	}
}

func WithAPIKey(apiKey string) ClientOption {
	return func(c *Option) error {
		c.apiKey = apiKey
		return nil
	}
}

func WithPrompt(prompt string) ClientOption {
	return func(c *Option) error {
		c.prompt = prompt
		return nil
	}
}

func WithPromptTpl(tpl string, data map[string]any) ClientOption {
	return func(c *Option) error {
		tmpl, err := pongo2.FromString(tpl)
		if err != nil {
			return fmt.Errorf("failed to parse prompt template: %v, tpl: %s", err, tpl)
		}
		str, err := tmpl.Execute(data)
		if err != nil {
			return fmt.Errorf("failed to execute prompt template: %v, tpl: %s", err, tpl)
		}
		c.prompt = str
		return nil
	}
}

func WithModel(model string) ClientOption {
	return func(c *Option) error {
		c.model = model
		return nil
	}
}
