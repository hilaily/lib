package configx

import (
	"fmt"
	"os"

	"github.com/hilaily/lib/env"
	"gopkg.in/yaml.v3"
)

type IConfig interface {
	Get(path string, ptr any) error
	Unmarshal(ptr any) error
}

// InitWithCustomConfig 允许用户在保留基础配置的同时添加自定义配置
func New(_env env.IENV) (*config, error) {
	configPath := _env.ConfigPath()
	if configPath == "" {
		configPath = fmt.Sprintf("./conf/config.%s.yaml", _env.GetEnv())
	}

	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file failed, path: %s, err: %w", configPath, err)
	}

	var nodeMap map[string]*yaml.Node
	if err := yaml.Unmarshal(fileContent, &nodeMap); err != nil {
		return nil, fmt.Errorf("unmarshal config file failed, data: %s, err: %w", string(fileContent), err)
	}

	c := &config{
		fileContent: fileContent,
		nodeMap:     nodeMap,
	}
	return c, nil
}

type config struct {
	fileContent []byte
	nodeMap     map[string]*yaml.Node
}

func (c *config) Get(path string, ptr any) error {
	nodes, ok := c.nodeMap[path]
	if !ok {
		return fmt.Errorf("path not found, path: %s", path)
	}
	return nodes.Decode(ptr)
}

func (c *config) Unmarshal(ptr any) error {
	return yaml.Unmarshal(c.fileContent, ptr)
}
