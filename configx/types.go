package configx

import "errors"

var (
	ErrPathNotFound = errors.New("path not found")
)

type IConfig interface {
	// IsExist check if the path exists
	IsExist(path string) bool
	// Get get config by path, it is just support first level path now
	Get(path string, ptr any) error
	// Unmarshal whole config to ptr
	Unmarshal(ptr any) error
	// Sub get sub config
	Sub(path string) IUnmarshaler
}

type IUnmarshaler interface {
	Unmarshal(ptr any) error
}

type IParam interface {
	ConfigPath() string
	GetEnv() string
}
