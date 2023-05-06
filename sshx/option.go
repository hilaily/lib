package sshx

import "golang.org/x/crypto/ssh"

type Option = func(*Client) error

// WithUser ...
func WithUser(user string) Option {
	return func(c *Client) error {
		c.user = user
		return nil
	}
}

// WithPort ...
func WithPort(port int) Option {
	return func(c *Client) error {
		c.port = port
		return nil
	}
}

// WithKeyPass ...
func WithKeyPass(keyPass string) Option {
	return func(c *Client) error {
		c.keyPass = keyPass
		return nil
	}
}

// WithClientConfig ...
func WithClientConfig(conf *ssh.ClientConfig) Option {
	return func(c *Client) error {
		c.clientConfig = conf
		return nil
	}

}

// WithJumpProxy ...
func WithJumpProxy(host, pass, keyPath string, ops ...Option) Option {
	return func(c *Client) error {
		j, err := New(host, pass, keyPath, ops...)
		if err != nil {
			return err
		}
		c.jumpClient = j.client
		return nil
	}
}
