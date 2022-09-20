package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"golang.org/x/net/proxy"
)

// WrapSOCKSProxy add a socks proxy for a http client
func WrapSOCKSProxy(client *http.Client, proxyURL string, insecure bool) (*http.Client, error) {
	dialer, err := proxy.SOCKS5("tcp", proxyURL, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("can't connect to the proxy: %s, %w", proxyURL, err)
	}

	dealContext := func(ctx context.Context, network, address string) (net.Conn, error) {
		return dialer.Dial(network, address)
	}

	transport := &http.Transport{
		DialContext: dealContext,
	}
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client.Transport = transport
	return client, nil
}
