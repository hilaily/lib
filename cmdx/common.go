package cmdx

import (
	"os"
	"os/exec"
)

func CommandIsExist(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func Chdir(dir string) (func() error, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	err = os.Chdir(dir)
	if err != nil {
		return nil, err
	}
	return func() error {
		return os.Chdir(pwd)
	}, nil
}

func MustChdir(dir string) func() {
	f, err := Chdir(dir)
	std.CheckErr(err)
	return func() {
		err := f()
		std.CheckErr(err)
	}
}
