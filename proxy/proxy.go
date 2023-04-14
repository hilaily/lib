package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/net/http/httpproxy"
	"golang.org/x/net/proxy"
)

// WrapHTTPProxy add a proxy for a http client, support http://x.x.x.x https://x.x.x.x socks5://x.x.x.x
func Wrap(client *http.Client, proxyURL string, insecure bool) (*http.Client, error) {
	if proxyURL == "" {
		return client, nil
	}

	transport := &http.Transport{}
	noProxy := ""
	for _, v := range []string{"no_proxy", "NO_PROXY"} {
		noProxy = os.Getenv(v)
		if noProxy != "" {
			break
		}
	}
	p := &httpproxy.Config{
		HTTPProxy:  proxyURL,
		HTTPSProxy: proxyURL,
		NoProxy:    noProxy,
		CGI:        os.Getenv("REQUEST_METHOD") != "",
	}
	f := p.ProxyFunc()

	transport.Proxy = func(req *http.Request) (*url.URL, error) {
		return f(req.URL)
	}

	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client.Transport = transport
	return client, nil
}

// WrapSOCKSProxy add a socks proxy for a http client
func WrapSOCKSProxy(client *http.Client, proxyURL string, user, pass string, insecure bool) (*http.Client, error) {
	var dialer proxy.Dialer
	var err error
	if user != "" {
		auth := &proxy.Auth{
			User:     user,
			Password: pass,
		}
		dialer, err = proxy.SOCKS5("tcp", proxyURL, auth, proxy.Direct)
	} else {
		dialer, err = proxy.SOCKS5("tcp", proxyURL, nil, proxy.Direct)
	}

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
