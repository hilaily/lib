package sshx

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
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
