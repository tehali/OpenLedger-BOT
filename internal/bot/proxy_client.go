package bot

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// getProxyClient 获取代理客户端
func (o *OpenLedger) getProxyClient(proxyURL string) (*http.Transport, error) {
	// 解析代理URL
	proxyURLParsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
	}

	// 创建transport
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// 根据代理类型设置
	switch strings.ToLower(proxyURLParsed.Scheme) {
	case "http", "https":
		transport.Proxy = http.ProxyURL(proxyURLParsed)
		return transport, nil

	case "socks4", "socks5":
		// 创建SOCKS拨号器
		auth := &proxy.Auth{}
		if proxyURLParsed.User != nil {
			auth.User = proxyURLParsed.User.Username()
			if password, ok := proxyURLParsed.User.Password(); ok {
				auth.Password = password
			}
		} else {
			auth = nil
		}

		// 根据协议类型选择不同的拨号器
		var dialer proxy.Dialer
		var dialErr error

		if strings.ToLower(proxyURLParsed.Scheme) == "socks5" {
			dialer, dialErr = proxy.SOCKS5("tcp", proxyURLParsed.Host, auth, proxy.Direct)
		} else {
			// SOCKS4不直接支持，暂时使用SOCKS5
			dialer, dialErr = proxy.SOCKS5("tcp", proxyURLParsed.Host, auth, proxy.Direct)
		}

		if dialErr != nil {
			return nil, fmt.Errorf("failed to create SOCKS dialer: %w", dialErr)
		}

		// 设置自定义拨号器
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}

		return transport, nil

	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %s", proxyURLParsed.Scheme)
	}
} 