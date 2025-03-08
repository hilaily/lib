package llm

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type IManager interface {
	GetModel(name string) (*Client, bool)
	GetProvider(name string) (*Client, bool)
}

type manager struct {
	providers map[string]string
	clients   map[string]*Client
}

func (m *manager) GetModel(name string) (*Client, bool) {
	client, ok := m.clients[name]
	return client, ok
}

func (m *manager) GetProvider(name string) (*Client, bool) {
	provider, ok := m.providers[name]
	if !ok {
		return nil, false
	}
	client, ok := m.clients[provider]
	return client, ok
}

type LLMConfig struct {
	Model   string `yaml:"model"`
	ApiKey  string `yaml:"apiKey"`
	BaseUrl string `yaml:"baseUrl"`
}

type Conf struct {
	LLM struct {
		Providers map[string]string     `yaml:"providers"`
		Models    map[string]*LLMConfig `yaml:"models"`
	} `yaml:"llm"`
}

func NewManager(conf string) (*manager, error) {
	confBytes, err := os.ReadFile(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to read llm config file: %w, conf: %s", err, conf)
	}
	return NewManagerFromData(confBytes)
}

func NewManagerFromData(confBytes []byte) (*manager, error) {
	cfg := &Conf{}
	err := yaml.Unmarshal(confBytes, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal llm config: %w, conf: %s", err, confBytes)
	}
	clients := make(map[string]*Client)
	for name, cfg := range cfg.LLM.Models {
		clients[name] = NewClient(WithAPIKey(cfg.ApiKey), WithBaseURL(cfg.BaseUrl), WithModel(cfg.Model))
	}
	if len(clients) == 0 {
		return nil, fmt.Errorf("no llm clients found, config string: %s, parsed config: %#+v", confBytes, cfg)
	}
	return &manager{
		providers: cfg.LLM.Providers,
		clients:   clients,
	}, nil
}
