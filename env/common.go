package env

import "os"

var (
	_ IENV = &env{}
)

type IENV interface {
	ConfigPath() string
	GetEnv() string
	IsDev() bool
	IsTest() bool
	IsPre() bool
	IsProd() bool
}

func New() *env {
	return &env{
		configPath: os.Getenv("CONFIG_PATH"),
		env:        os.Getenv("ENV"),
	}
}

type env struct {
	env        string
	configPath string
}

func (e *env) ConfigPath() string {
	return e.configPath
}

func (e *env) GetEnv() string {
	return e.env
}

func (e *env) IsDev() bool {
	return e.env == "dev"
}

func (e *env) IsTest() bool {
	return e.env == "test"
}

func (e *env) IsPre() bool {
	return e.env == "pre"
}

func (e *env) IsProd() bool {
	return e.env == "prod"
}
