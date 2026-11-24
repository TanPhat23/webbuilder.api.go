package services

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"syscall"
	"time"
)

var ErrBlockedWebhookURL = errors.New("webhook URL is not allowed")

func ValidateWebhookURL(rawURL string, allowLocal bool) error {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return fmt.Errorf("%w: URL is required", ErrBlockedWebhookURL)
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("%w: invalid URL", ErrBlockedWebhookURL)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("%w: scheme must be http or https", ErrBlockedWebhookURL)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("%w: host is required", ErrBlockedWebhookURL)
	}
	if allowLocal {
		return nil
	}
	lowerHost := strings.ToLower(host)
	if lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".localhost") || strings.HasSuffix(lowerHost, ".local") {
		return fmt.Errorf("%w: local hosts are blocked", ErrBlockedWebhookURL)
	}
	if addr, err := netip.ParseAddr(host); err == nil {
		if isBlockedAddr(addr) {
			return fmt.Errorf("%w: private or loopback IP is blocked", ErrBlockedWebhookURL)
		}
	}
	return nil
}

func isBlockedAddr(addr netip.Addr) bool {
	a := addr.Unmap()
	return a.IsLoopback() ||
		a.IsPrivate() ||
		a.IsLinkLocalUnicast() ||
		a.IsLinkLocalMulticast() ||
		a.IsMulticast() ||
		a.IsUnspecified()
}

func newWebhookHTTPClient(allowLocal bool, timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}
	if !allowLocal {
		dialer.Control = func(_ string, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			addr, err := netip.ParseAddr(host)
			if err != nil {
				return fmt.Errorf("%w: unparseable resolved IP", ErrBlockedWebhookURL)
			}
			if isBlockedAddr(addr) {
				return fmt.Errorf("%w: resolved IP %s is blocked", ErrBlockedWebhookURL, addr.String())
			}
			return nil
		}
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		},
		MaxIdleConns:        20,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
