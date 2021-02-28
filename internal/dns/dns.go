package dns

import (
	"context"
	"net"
	"time"

	"github.com/lomik/zapwriter"
	"github.com/rs/dnscache"
	"go.uber.org/zap"
)

var (
	resolver *dnscache.Resolver
)

func GetDialContextWithTimeout(dialTimeout, keepaliveTimeout time.Duration) func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
	logger := zapwriter.Logger("dns")
	dialer := net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: keepaliveTimeout,
	}
	if resolver == nil {
		logger.Debug("no caching dns initialized, will return typical DialContext")
		return (&dialer).DialContext
	}

	logger.Debug("returning caching DialContext")
	return func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		ips, err := resolver.LookupHost(ctx, host)
		if err != nil {
			return nil, err
		}
		for _, ip := range ips {
			conn, err = dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
			if err == nil {
				break
			}
		}
		return
	}
}

func UseDNSCache(dnsRefreshTime time.Duration) {
	logger := zapwriter.Logger("dns")
	resolver = &dnscache.Resolver{}

	// Periodically refresh cache
	go func() {
		ticker := time.NewTicker(dnsRefreshTime)
		defer ticker.Stop()
		for range ticker.C {
			logger.Debug("cache refreshed")
			resolver.Refresh(true)
		}
	}()

	logger.Debug("caching dns resolver initialized",
		zap.Duration("refreshTime", dnsRefreshTime),
	)
}
