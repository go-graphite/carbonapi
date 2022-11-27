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
		caCertPool := x509.NewCertPool()

		for _, caCert := range config.TLSClientConfig.CACertFiles {
			cert, err := os.ReadFile(caCert)
			if err != nil {
				logger.Fatal("failed to read CA Cert File",
					zap.Error(err),
				)
			}
			caCertPool.AppendCertsFromPEM(cert)
		}

		certificates := make([]tls.Certificate, 0, len(config.TLSClientConfig.CertificateParis))

		for _, certPair := range config.TLSClientConfig.CertificateParis {
			certificate, err := tls.LoadX509KeyPair(certPair.CertFile, certPair.PrivateKeyFile)
			if err != nil {
				logger.Fatal("failed to load X509 Key Pair",
					zap.Error(err),
				)
			}
			certificates = append(certificates, certificate)
		}
		transport.TLSClientConfig = &tls.Config{
			RootCAs:      caCertPool,
			Certificates: certificates,
			MinVersion:   tls.VersionTLS13,
			MaxVersion:   0,
		}
	}

	httpClient := &http.Client{
		Transport: transport,
	}
	return httpClient
}
