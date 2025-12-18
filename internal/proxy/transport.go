package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/haltman-io/brave-search/internal/cli"
	"golang.org/x/net/proxy"
)

func BuildTransport(opts cli.Options) (*http.Transport, error) {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   12 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   12 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if opts.NoProxy {
		tr.Proxy = nil
	}

	if opts.InsecureTLS {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 (explicit user flag)
	}

	if strings.TrimSpace(opts.ProxyURL) == "" {
		return tr, nil
	}

	u, err := parseProxyURL(opts.ProxyURL)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		if opts.ProxyAuth != "" && (u.User == nil || u.User.Username() == "") {
			user, pass, ok := strings.Cut(opts.ProxyAuth, ":")
			if ok {
				u.User = url.UserPassword(user, pass)
			} else {
				u.User = url.User(opts.ProxyAuth)
			}
		}
		tr.Proxy = http.ProxyURL(u)
		return tr, nil

	case "socks5", "socks5h":
		host := u.Host
		if host == "" {
			return nil, fmt.Errorf("invalid socks5 proxy url: missing host")
		}

		var auth *proxy.Auth
		if opts.ProxyAuth != "" {
			user, pass, ok := strings.Cut(opts.ProxyAuth, ":")
			if ok {
				auth = &proxy.Auth{User: user, Password: pass}
			} else {
				auth = &proxy.Auth{User: opts.ProxyAuth}
			}
		}

		dialer, err := proxy.SOCKS5("tcp", host, auth, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("failed to create socks5 dialer: %w", err)
		}

		tr.Proxy = nil
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
		return tr, nil

	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %q (use http(s):// or socks5://)", u.Scheme)
	}
}

func parseProxyURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty proxy url")
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy url: %w", err)
	}
	if u.Scheme == "" {
		return nil, fmt.Errorf("invalid proxy url: missing scheme")
	}
	return u, nil
}
