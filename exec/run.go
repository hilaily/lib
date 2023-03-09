package run

import (
	"fmt"
	"os/exec"
	"strings"
)

func Run(cmd string) ([]byte, error) {
	cmds := strings.Fields(cmd)
	c := exec.Command(cmds[0], cmds[1:]...)
	res, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s %w", string(res), err)
	}
	return res, nil
}
