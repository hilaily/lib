package sshx

import (
	"fmt"
	"os"
	"time"

	"github.com/hilaily/kit/pathx"
	"golang.org/x/crypto/ssh"
)

// Client ...
type Client struct {
	host         string
	user         string
	port         int
	pass         string
	keyPath      string
	keyPass      string
	clientConfig *ssh.ClientConfig
	jumpClient   *Client

	client *ssh.Client
}

// New a client
// pass or key choose one
func New(host, pass, keyPath string, ops ...Option) (*Client, error) {
	c := &Client{
		user: "root",
		port: 22,
	}
	c.host = host
	c.pass = pass
	c.keyPath = keyPath
	var err error
	for _, f := range ops {
		err = f(c)
		if err != nil {
			return nil, err
		}

	}
	if c.pass == "" && c.keyPath == "" {
		return nil, fmt.Errorf("password and key both are empty")
	}

	clientConfig, err := c.genConfig()
	if err != nil {
		return nil, err
	}
	c.clientConfig = clientConfig
	client, err := c.newClient()
	if err != nil {
		return nil, err
	}
	c.client = client
	return c, nil
}

// Close ...
func (c *Client) Close() error {
	return c.client.Close()
}

// GetClient get original ssh client
func (c *Client) GetClient() *ssh.Client {
	return c.client
}

func (c *Client) genConfig() (*ssh.ClientConfig, error) {
	if c.clientConfig != nil {
		return c.clientConfig, nil
	}
	conf := &ssh.ClientConfig{
		Timeout:         30 * time.Second,
		User:            c.user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if c.pass != "" {
		conf.Auth = []ssh.AuthMethod{ssh.Password(c.pass)}
		return conf, nil
	}
	keyPath, err := pathx.ExpandHome(c.keyPath)
	if err != nil {
		return nil, fmt.Errorf("parse key path fail %s, %w", keyPath, err)
	}
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read key path fail, %s, %w", keyPath, err)
	}
	signer, err := signerFromPem(key, []byte(c.keyPass))
	if err != nil {
		return nil, fmt.Errorf("parse key fail, %s, %w", keyPath, err)
	}

	conf.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	return conf, nil
}

func (c *Client) newClient() (*ssh.Client, error) {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	if c.jumpClient != nil {
		// connect with the jump host
		// Dial a connection to the service host, from the bastion
		conn, err := c.jumpClient.client.Dial("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("jump server connect to dest server fail, %s, %w", addr, err)
		}
		ncc, chans, reqs, err := ssh.NewClientConn(conn, addr, c.clientConfig)
		if err != nil {
			return nil, fmt.Errorf("new ssh client fail, %s, %w", addr, err)
		}
		sClient := ssh.NewClient(ncc, chans, reqs)
		return sClient, nil
	}
	client, err := ssh.Dial("tcp", addr, c.clientConfig)
	if err != nil {
		return nil, fmt.Errorf("new ssh client fail, %s, %w", addr, err)
	}
	return client, nil
}

func signerFromPem(pemBytes []byte, password []byte) (ssh.Signer, error) {
	var signer ssh.Signer
	var err error

	if len(password) != 0 {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, password)
	} else {
		// generate signer instance from plain key
		signer, err = ssh.ParsePrivateKey(pemBytes)
	}
	if err != nil {
		return nil, fmt.Errorf("parsing private key failed %w", err)
	}
	return signer, nil
}
