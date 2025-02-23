package llm

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Manager struct {
	clients map[string]*Client
}

func (m *Manager) GetClient(name string) (*Client, bool) {
	client, ok := m.clients[name]
	return client, ok
}

type LLMConfig struct {
	Name    string `yaml:"name"`
	Model   string `yaml:"model"`
	ApiKey  string `yaml:"apiKey"`
	BaseUrl string `yaml:"baseUrl"`
}

type Conf struct {
	LLM map[string]*LLMConfig `yaml:"llm"`
}

func NewManager(conf string) (*Manager, error) {
	cfg := &Conf{}
	err := yaml.Unmarshal([]byte(conf), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal llm config: %w, conf: %s", err, conf)
	}
	clients := make(map[string]*Client)
	for name, cfg := range cfg.LLM {
		clients[name] = NewClient(WithAPIKey(cfg.ApiKey), WithBaseURL(cfg.BaseUrl), WithModel(cfg.Model))
	}
	return &Manager{
		clients: clients,
	}, nil
}
