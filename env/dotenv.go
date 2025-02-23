package env

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func MustLoadEnv(name string) {
	err := LoadEnv(name)
	if err != nil {
		panic(err)
	}
}

func LoadEnv(name string) error {
	dir, err := os.Getwd() // 获取当前工作目录
	if err != nil {
		return err
	}
	logrus.Debugf("Loading env file from %s", dir)

	for {
		envPath := filepath.Join(dir, name)
		if _, err := os.Stat(envPath); err == nil {
			// 找到 .env 文件，加载并退出
			err := godotenv.Load(envPath)
			if err != nil {
				return err
			}
			return nil
		}

		// 如果没有找到，向上一级目录查找
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// 已经到达根目录，仍未找到 .env 文件
			return fmt.Errorf("No %s file found", name)
		}
		dir = parentDir
	}
}
