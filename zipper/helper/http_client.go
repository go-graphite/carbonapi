package helper

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"github.com/go-graphite/carbonapi/internal/dns"
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
		cert, err := os.ReadFile(config.TLSClientConfig.CACertFile)
		if err != nil {
			logger.Fatal("failed to read CA Cert File",
				zap.Error(err),
			)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(cert)

		certificate, err := tls.LoadX509KeyPair(config.TLSClientConfig.CertFile, config.TLSClientConfig.PrivateKeyFile)
		if err != nil {
			logger.Fatal("failed to load X509 Key Pair",
				zap.Error(err),
			)
		}
		transport.TLSClientConfig = &tls.Config{
			RootCAs:      caCertPool,
			Certificates: []tls.Certificate{certificate},
			MinVersion:   tls.VersionTLS13,
			MaxVersion:   0,
		}
	}

	httpClient := &http.Client{
		Transport: transport,
	}
	return httpClient
}
