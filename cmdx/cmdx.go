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

// ToCommand transfer a string to command object
func ToCommand(cmd string) *exec.Cmd {
	cmds := strings.Fields(cmd)
	c := exec.Command(cmds[0], cmds[1:]...)
	return c
}

// Run ...
func Run(format string, a ...any) error {
	return New(format, a...).Run()
}

// MustRun ...
func MustRun(format string, a ...any) {
	err := Run(format, a...)
	std.CheckErr(err)
}

func SHRun(format string, a ...any) error {
	c := exec.Command("sh", "-c", fmt.Sprintf(format, a...))
	err := New2(c).Run()
	return err
}

// MustSHRun ...
func MustSHRun(format string, a ...any) {
	err := SHRun(format, a...)
	std.CheckErr(err)
}

// RunCombinedOutput ...
func RunCombinedOutput(cmd string, envs []string) (string, error) {
	var b bytes.Buffer
	err := New(cmd).Env(envs...).Output(&b, &b).Run()

	r := b.String()
	if err != nil {
		return "", fmt.Errorf("%s %w", r, err)
	}
	return r, nil
}

// New ...
func New(format string, a ...any) *CMD {
	c := ToCommand(fmt.Sprintf(format, a...))
	return New2(c)
}

// New2 ...
func New2(cmd *exec.Cmd) *CMD {
	c := &CMD{}
	c.cmd = cmd
	c.cmd.Stdin = os.Stdin
	c.cmd.Stdout = os.Stdout
	c.cmd.Stderr = os.Stderr
	return c
}

// CMD wrap origin cmd structure
type CMD struct {
	cmd        *exec.Cmd
	cancelFunc context.CancelFunc
}

// Env set environment variables
func (c *CMD) Env(envs ...string) *CMD {
	e := os.Environ()
	if len(envs) > 0 {
		e = append(e, envs...)
	}
	c.cmd.Env = e
	return c
}

// Dir set command workdir
func (c *CMD) Dir(dir string) *CMD {
	c.cmd.Dir = dir
	return c
}

// Input ...
func (c *CMD) Input(in io.Reader) *CMD {
	c.cmd.Stdin = in
	return c
}

// Output ...
func (c *CMD) Output(output, errput io.Writer) *CMD {
	if output != nil {
		c.cmd.Stdout = output
	}
	if errput != nil {
		c.cmd.Stderr = errput
	}
	return c
}

// Timeout ...
func (c *CMD) Timeout(t time.Duration) *CMD {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	newCMD := exec.CommandContext(ctx, c.cmd.Path, c.cmd.Args...)
	c.cmd = newCMD
	c.cancelFunc = cancel
	return c
}

// Run ...
func (c *CMD) Run() error {
	defer c.finish()
	_, err := io.WriteString(c.cmd.Stdout, c.cmd.String()+"\n")
	if err != nil {
		return err
	}
	err = c.cmd.Start()
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
