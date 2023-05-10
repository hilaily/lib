package sshx

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hilaily/kit/helper"
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

/*
// RawInteract interact with ssh command
func (c *Client) RawInteract() error {
	args := c.sshString()
	if c.pass != "" {
		return cmdx.Run("sshpass -p %s %s", c.pass, args)
	}
	return cmdx.Run(args)
}
*/

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

	/*
		err = copyData(session)
		if err != nil {
			return err
		}
	*/

	if err = session.Shell(); err != nil {
		return fmt.Errorf("start shell error: %w", err)
	}
	if err = session.Wait(); err != nil {
		return fmt.Errorf("return error: %w", err)
	}
	return nil
}

/*
func (c *Client) sshString() string {
	var proxy string
	if c.jumpClient != nil {
		proxy = c.jumpClient.sshString()
		proxy = strings.ReplaceAll(proxy, "ssh ", "")
	}
	buf := &bytes.Buffer{}
	_, _ = buf.WriteString("ssh ")
	if proxy != "" {
		_, _ = buf.WriteString(fmt.Sprintf("-oProxyCommand='sshpass -p gatewaypassword ssh -W %h%p uname@gatwayserver'" + proxy + " ")
	}
	_, _ = buf.WriteString(fmt.Sprintf("%s@%s:%d ", c.user, c.host, c.port))
	if c.keyPath != "" {
		_, _ = buf.WriteString("-i " + c.keyPath + " ")
	}
	return buf.String()
}
*/

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
