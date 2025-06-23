package configx

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	_ IConfig = &config{}
)

// InitWithCustomConfig 允许用户在保留基础配置的同时添加自定义配置
func New(_env IParam) (*config, error) {
	configPath := _env.ConfigPath()
	if configPath == "" {
		configPath = fmt.Sprintf("./conf/config.%s.yaml", _env.GetEnv())
	}
	return NewFromFile(configPath)
}

func NewFromFile(path string) (*config, error) {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file failed, path: %s, err: %w", path, err)
	}

	var node yaml.Node
	if err := yaml.Unmarshal(fileContent, &node); err != nil {
		return nil, fmt.Errorf("unmarshal config file failed, data: %s, err: %w", string(fileContent), err)
	}
	if node.Kind != yaml.DocumentNode {
		return nil, fmt.Errorf("config file is not a document node, data: %s", string(fileContent))
	}
	if len(node.Content) == 0 {
		return nil, fmt.Errorf("config file is empty, data: %s", string(fileContent))
	}
	if node.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("config file is not a mapping node(first level must be a map), data: %s", string(fileContent))
	}

	nodeMap := make(map[string]*yaml.Node)
	for i := 0; i < len(node.Content[0].Content); i += 2 {
		key := node.Content[0].Content[i].Value
		value := node.Content[0].Content[i+1]
		nodeMap[key] = value
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

func (c *config) Sub(path string) IUnmarshaler {
	node := c.nodeMap[path]
	return &unmarshaler{node: node}
}

type unmarshaler struct {
	node *yaml.Node
}

func (u *unmarshaler) Unmarshal(ptr any) error {
	if u.node == nil {
		return ErrPathNotFound
	}
	return u.node.Decode(ptr)
}
