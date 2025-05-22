package configx

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	_               IConfig = &config{}
	ErrPathNotFound         = errors.New("path not found")
)

type IConfig interface {
	// IsExist check if the path exists
	IsExist(path string) bool
	// Get get config by path, it is just support first level path now
	Get(path string, ptr any) error
	// Unmarshal whole config to ptr
	Unmarshal(ptr any) error
}

type IParam interface {
	ConfigPath() string
	GetEnv() string
}

// InitWithCustomConfig 允许用户在保留基础配置的同时添加自定义配置
func New(_env IParam) (*config, error) {
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

// Get
// if path is not found, return false, nil
func (c *config) Get(path string, ptr any) error {
	nodes, ok := c.nodeMap[path]
	if !ok {
		return ErrPathNotFound
	}
	return nodes.Decode(ptr)
}

func (c *config) IsExist(path string) bool {
	_, ok := c.nodeMap[path]
	return ok
}

func (c *config) Unmarshal(ptr any) error {
	return yaml.Unmarshal(c.fileContent, ptr)
}
