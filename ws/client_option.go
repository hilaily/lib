package ws

func WithConnectConfig(config *ConnectConfig) ClientOptions {
	return func(c *Client) error {
		c.connectConfig = config
		return nil
	}
}

func WithOnError(onError func(err error)) ClientOptions {
	return func(c *Client) error {
		c.onError = onError
		return nil
	}
}

func WithOnConnect(onConnect func()) ClientOptions {
	return func(c *Client) error {
		c.onConnect = onConnect
		return nil
	}
}

func WithOnDisconnect(onDisconnect func()) ClientOptions {
	return func(c *Client) error {
		c.onDisconnect = onDisconnect
		return nil
	}
}
