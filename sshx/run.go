package sshx

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hilaily/kit/helper"
	"github.com/hilaily/kit/stringx"
	"github.com/hilaily/lib/cmdx"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// RunResult ...
func (c *Client) RunResult(script string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("create new session error: %w", err)
	}
	defer session.Close()

	buf, err := session.CombinedOutput(script)
	return string(buf), err
}

// RunDirect ...
func (c *Client) RunDirect(script string, stdout, stderr io.Writer) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("create new session error: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	return session.Run(script)
}

// Copy ...
// /tmp/1.txt -> /root/ = /root/1.txt
// /tmp/dir/  -> /root/ = /tmp/1.txt
// /tmp/dir/  -> /root/dir/ = /tmp/dir/1.txx
func (c *Client) Copy(src, dst string, force bool) error {
	client, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("new sftp client fail %w", err)
	}
	defer func() {
		_ = client.Close()
	}()
	return copyAllFiles(client, src, dst, force)
}

func copyAllFiles(c *sftp.Client, src, dst string, force bool) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat src fail, %s, %w", src, err)
	}
	if !srcInfo.IsDir() {
		return copySingleFile(c, src, dst, force)
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, strings.TrimPrefix(path, src))
		if info.IsDir() {
			return c.MkdirAll(dstPath)
		}
		return copySingleFile(c, path, dstPath, force)
	})

}

func copySingleFile(c *sftp.Client, src, dst string, force bool) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if strings.HasSuffix(dst, "/") {
		dst = dst + filepath.Base(src)
	}

	if !force {
		if _, err = c.Stat(dst); err == nil {
			return nil
		}
	}

	srcFile, err := os.Open(filepath.Clean(src))
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := c.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	n, err := dstFile.ReadFrom(srcFile)
	if err != nil {
		return err
	}
	if n != srcInfo.Size() {
		return errors.New("unmatched file size")
	}

	return dstFile.Chmod(srcInfo.Mode())
}

// RawInteract interact with ssh command
func (c *Client) RawInteract() error {
	args := map[string]any{
		"user":    c.user,
		"keyPath": c.keyPath,
		"host":    c.host,
		"port":    c.port,
	}
	content := "{{.user}}@{{.host}} -p {{.port}}"
	if c.keyPath != "" {
		content = "{{.user}}@{{.host}} -p {{.port}} -i {{.keyPath}}"
	}
	if c.jumpClient != nil {
		if c.jumpClient.keyPath == "" {
			return fmt.Errorf("not support jump server without public key")
		}
		content = content + ` -oProxyCommand="ssh {{.proxyUser}}@{{.proxyHost}} -p {{.proxyPort}} -i {{.proxyKeyPath}} -W {{.host}}:{{.port}}"`
		args["proxyUser"] = c.jumpClient.user
		args["proxyKeyPath"] = c.jumpClient.keyPath
		args["proxyHost"] = c.jumpClient.host
		args["proxyPort"] = c.jumpClient.port
	}
	e := exec.Command("bash", "-c", "ssh "+stringx.Format(content, args))
	return cmdx.New2(e).Run()
}

func (c *Client) RawSCP(src, dst string) error {
	args := map[string]any{
		"user":    c.user,
		"keyPath": c.keyPath,
		"host":    c.host,
		"port":    c.port,
		"src":     src,
		"dst":     dst,
	}
	content := `-P {{.port}} {{.src}} {{.user}}@{{.host}}:{{.dst}}`
	if c.keyPath != "" {
		content = `-P {{.port}} -i {{.keyPath}} {{.src}} {{.user}}@{{.host}}:{{.dst}}`
	}
	if c.jumpClient != nil {
		if c.jumpClient.keyPath == "" {
			return fmt.Errorf("not support jump server without public key")
		}
		content = `-oProxyCommand="ssh {{.proxyUser}}@{{.proxyHost}} -p {{.proxyPort}} -i {{.proxyKeyPath}} -W %h:%p" ` + content
		args["proxyUser"] = c.jumpClient.user
		args["proxyKeyPath"] = c.jumpClient.keyPath
		args["proxyHost"] = c.jumpClient.host
		args["proxyPort"] = c.jumpClient.port
	}
	e := exec.Command("bash", "-c", "scp "+stringx.Format(content, args))
	return cmdx.New2(e).Run()
}

// Interact ...
func (c *Client) Interact() error {
	return c.in()
}

func (c *Client) in() error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("create new session error: %w", err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // 禁用回显（0禁用，1启动）
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, //output speed = 14.4kbaud
	}

	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}

	if err = session.RequestPty(termType, 32, 160, modes); err != nil {
		return fmt.Errorf("request pty error: %w", err)
	}

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err = session.Shell(); err != nil {
		return fmt.Errorf("start shell error: %w", err)
	}
	if err = session.Wait(); err != nil {
		return fmt.Errorf("return error: %w", err)
	}
	return nil
}

func copyData(session *ssh.Session) error {
	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := session.StderrPipe()
	helper.CheckErr(err)

	go io.Copy(os.Stderr, stderr)
	go io.Copy(os.Stdout, stdout)
	go func() {
		buf := make([]byte, 128)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				_, _ = fmt.Println(err)
				return
			}
			if n > 0 {
				_, err = stdin.Write(buf[:n])
				if err != nil {
					helper.CheckErr(err)
				}
			}
		}
	}()
	return nil
}
