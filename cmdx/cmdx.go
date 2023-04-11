package cmdx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func ToCommand(cmd string) *exec.Cmd {
	cmds := strings.Fields(cmd)
	c := exec.Command(cmds[0], cmds[1:]...)
	return c
}

func Run(cmd string, envs ...string) error {
	c := ToCommand(cmd)
	return RealRun(c, os.Stdout, os.Stderr, os.Stdin, envs...)
}

func MustRun(cmd string, envs ...string) {
	err := Run(cmd, envs...)
	std.CheckErr(err)
}

func MustSHRun(cmdStr string, envs ...string) {
	c := exec.Command("sh", "-c", cmdStr)
	err := RealRun(c, os.Stdout, os.Stderr, os.Stdin, envs...)
	std.CheckErr(err)
}

func RunCombinedOutput(cmd string, envs ...string) (string, error) {
	c := ToCommand(cmd)
	var b bytes.Buffer
	err := RealRun(c, &b, &b, nil, envs...)

	r := b.String()
	if err != nil {
		return "", fmt.Errorf("%s %w", r, err)
	}
	return r, nil
}

func RealRun(cmd *exec.Cmd, output, errput io.Writer, in io.Reader, envs ...string) error {
	env := os.Environ()
	if len(envs) > 0 {
		env = append(env, envs...)
	}
	cmd.Env = env
	if output != nil {
		cmd.Stdout = output
	}
	if errput != nil {
		cmd.Stderr = errput
	}
	if in != nil {
		cmd.Stdin = in
	}

	io.WriteString(cmd.Stdout, cmd.String()+"\n")
	err := cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
}
