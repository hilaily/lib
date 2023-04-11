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

func Run(format string, a ...any) error {
	return New(format, a...).Run()
}

func MustRun(format string, a ...any) {
	err := Run(format, a...)
	std.CheckErr(err)
}

func MustSHRun(format string, a ...any) {
	c := exec.Command("sh", "-c", fmt.Sprintf(format, a...))
	err := New2(c).Run()
	std.CheckErr(err)
}

func RunCombinedOutput(cmd string, envs []string) (string, error) {
	var b bytes.Buffer
	err := New(cmd).SetEnv(envs...).SetOutput(&b, &b).Run()

	r := b.String()
	if err != nil {
		return "", fmt.Errorf("%s %w", r, err)
	}
	return r, nil
}

func New(format string, a ...any) *CMD {
	c := ToCommand(fmt.Sprintf(format, a...))
	return New2(c)
}

func New2(cmd *exec.Cmd) *CMD {
	c := &CMD{}
	c.cmd = cmd
	c.cmd.Stdin = os.Stdin
	c.cmd.Stdout = os.Stdout
	c.cmd.Stderr = os.Stderr
	return c
}

type CMD struct {
	str string
	cmd *exec.Cmd
}

func (c *CMD) SetEnv(envs ...string) *CMD {
	e := os.Environ()
	if len(envs) > 0 {
		e = append(e, envs...)
	}
	c.cmd.Env = e
	return c
}

func (c *CMD) SetInput(in io.Reader) *CMD {
	c.cmd.Stdin = in
	return c
}

func (c *CMD) SetOutput(output, errput io.Writer) *CMD {
	if output != nil {
		c.cmd.Stdout = output
	}
	if errput != nil {
		c.cmd.Stderr = errput
	}
	return c
}

func (c *CMD) Run() error {
	io.WriteString(c.cmd.Stdout, c.cmd.String()+"\n")
	err := c.cmd.Start()
	if err != nil {
		return err
	}

	return c.cmd.Wait()
}
