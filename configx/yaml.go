package configx

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func YAMLByENV(env string, ptr any) error {
	p := os.Getenv(env)
	f, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("read config fail, path:%s, %w", p, err)
	}
	err = yaml.Unmarshal(f, ptr)
	if err != nil {
		return fmt.Errorf("unmarshal config fail, data:%s, %w", string(f), err)
	}
	return nil
}
