package cmdx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
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
	err := New(cmd).Env(envs...).Output(&b, &b).Run()

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
	cmd        *exec.Cmd
	cancelFunc context.CancelFunc
}

func (c *CMD) Env(envs ...string) *CMD {
	e := os.Environ()
	if len(envs) > 0 {
		e = append(e, envs...)
	}
	c.cmd.Env = e
	return c
}

func (c *CMD) Dir(dir string) *CMD {
	c.cmd.Dir = dir
	return c
}

func (c *CMD) Input(in io.Reader) *CMD {
	c.cmd.Stdin = in
	return c
}

func (c *CMD) Output(output, errput io.Writer) *CMD {
	if output != nil {
		c.cmd.Stdout = output
	}
	if errput != nil {
		c.cmd.Stderr = errput
	}
	return c
}

func (c *CMD) Timeout(t time.Duration) *CMD {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	newCMD := exec.CommandContext(ctx, c.cmd.Path, c.cmd.Args...)
	c.cmd = newCMD
	c.cancelFunc = cancel
	return c
}

func (c *CMD) Run() error {
	defer c.finish()
	io.WriteString(c.cmd.Stdout, c.cmd.String()+"\n")
	err := c.cmd.Start()
	if err != nil {
		return err
	}

	return c.cmd.Wait()
}

func (c *CMD) finish() {
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
}
