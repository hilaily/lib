package configx

import (
	"fmt"

	"github.com/spf13/viper"
)

var (
	_ IConfig = &viperConfig{}
)

type viperConfig struct {
	v *viper.Viper
}

func NewFromViper(v *viper.Viper) *viperConfig {
	return &viperConfig{v: v}
}

func (c *viperConfig) Get(path string, ptr any) error {
	v := c.v.Sub(path)
	if v == nil {
		return ErrPathNotFound
	}

	if err := v.Unmarshal(ptr); err != nil {
		return fmt.Errorf("parse config failed: %w", err)
	}
	return nil
}

func (c *viperConfig) IsExist(path string) bool {
	v := c.v.Sub(path)
	return v != nil
}

func (c *viperConfig) Unmarshal(ptr any) error {
	return c.v.Unmarshal(ptr)
}

func (c *viperConfig) Sub(path string) IUnmarshaler {
	v := c.v.Sub(path)
	return &viperConfig{v: v}
}
