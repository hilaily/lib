package sshx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hilaily/kit/pathx"
)

// CopyID copy public key to remote instance authorized_keys file
func (c *Client) CopyID(pubKey ...string) error {
	pubKeyPath := pathx.MustExpandHome("~/.ssh/id_rsa.pub")
	if len(pubKey) > 0 {
		pubKeyPath = pubKey[0]
	}
	if !pathx.IsExist(pubKeyPath) {
		return fmt.Errorf("copy id fail, key is not exist, key:%s, %w", pubKeyPath, os.ErrNotExist)
	}

	writePublicKeyCmd, err := generateWritePublicKeyCmd(pubKeyPath)
	if err != nil {
		return fmt.Errorf("copy if fail %w", err)
	}
	res, err := c.RunResult(writePublicKeyCmd)
	if err != nil {
		return fmt.Errorf("output:%s, %w", res, err)
	}
	return nil
}

// CopyKey to copy private or public key to remote instance
// keyPath is like ~/.ssh/id_rsa.pub
func (c *Client) CopyKey(keyPath []string) error {
	for _, v := range keyPath {
		err := c.Copy(v, fmt.Sprintf(".ssh/%s", filepath.Base(v)), false)
		if err != nil {
			return fmt.Errorf("copy key failed, key:%s, %w", v, err)
		}
	}
	return nil
}

func generateWritePublicKeyCmd(key string) (string, error) {
	data, err := os.ReadFile(key)
	if err != nil {
		return "", fmt.Errorf("read key fail %w", err)
	}

	keyStr := strings.TrimSuffix(string(data), "\n")
	return fmt.Sprintf("mkdir -p ~/.ssh && touch ~/.ssh/authorized_keys && grep -q \"%s\" ~/.ssh/authorized_keys || echo \"%s\" | tee -a ~/.ssh/authorized_keys", keyStr, keyStr), nil

}
