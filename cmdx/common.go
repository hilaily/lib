package cmdx

import "os/exec"

func CommandIsExist(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
