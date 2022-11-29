package helper

import (
	"net/http"

	"github.com/go-graphite/carbonapi/internal/dns"
	"github.com/go-graphite/carbonapi/pkg/tlsconfig"
	"github.com/go-graphite/carbonapi/zipper/types"
	"go.uber.org/zap"
)

func GetHTTPClient(logger *zap.Logger, config types.BackendV2) *http.Client {
	transport := &http.Transport{
		MaxConnsPerHost:     *config.ConcurrencyLimit,
		MaxIdleConnsPerHost: *config.MaxIdleConnsPerHost,
		IdleConnTimeout:     *config.IdleConnectionTimeout,
		ForceAttemptHTTP2:   config.ForceAttemptHTTP2,
		DialContext:         dns.GetDialContextWithTimeout(config.Timeouts.Connect, *config.KeepAliveInterval),
	}

	if config.TLSClientConfig != nil {
		tlsConfig, warns, err := tlsconfig.ParseClientTLSConfig(config.TLSClientConfig)
		if err != nil {
			logger.Fatal("failed to initialize client for group",
				zap.String("group_name", config.GroupName),
				zap.Error(err),
			)
		}
		if len(warns) > 0 {
			logger.Warn("insecure options detected, while parsing HTTP Client TLS Config for backed",
				zap.String("group_name", config.GroupName),
				zap.Strings("warnings", warns),
			)
		}
		transport.TLSClientConfig = tlsConfig
	}

	httpClient := &http.Client{
		Transport: transport,
	}
	return httpClient
}
