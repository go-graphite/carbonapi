package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	merry2 "github.com/ansel1/merry/v2"
)

// ParseServerTLSConfig parses server and client TLSConfig struct and returns &tls.Config, list of warnings or error if
// parsing has failed.
// At this moment warnings are only about insecure ciphers
func ParseServerTLSConfig(serverTLSConfig, clientTLSConfig *TLSConfig) (*tls.Config, []string, error) {
	caCertPool := x509.NewCertPool()

	for _, caCert := range serverTLSConfig.CACertFiles {
		cert, err := os.ReadFile(caCert)
		if err != nil {
			return nil, nil, err
		}
		caCertPool.AppendCertsFromPEM(cert)
	}

	if len(serverTLSConfig.CertificatePairs) == 0 {
		return nil, nil, merry2.Errorf("no server certificate pairs provided")
	}

	certificates := make([]tls.Certificate, 0, len(serverTLSConfig.CertificatePairs))

	for _, certPair := range serverTLSConfig.CertificatePairs {
		certificate, err := tls.LoadX509KeyPair(certPair.CertFile, certPair.PrivateKeyFile)
		if err != nil {
			return nil, nil, err
		}
		certificates = append(certificates, certificate)
	}

	minTLSVersion, err := ParseTLSVersion(serverTLSConfig.MinTLSVersion)
	if err != nil {
		return nil, nil, err
	}
	maxTLSVersion, err := ParseTLSVersion(serverTLSConfig.MaxTLSVersion)
	if err != nil {
		return nil, nil, err
	}

	curves, err := ParseCurves(serverTLSConfig.Curves)
	if err != nil {
		return nil, nil, err
	}

	ciphers, warns, err := CipherSuitesToUint16(serverTLSConfig.CipherSuites)
	if err != nil {
		return nil, nil, err
	}

	clientAuth, err := ParseClientAuthType(serverTLSConfig.ClientAuth)
	if err != nil {
		return nil, nil, err
	}

	if serverTLSConfig.InsecureSkipVerify {
		warns = append(warns, "InsecureSkipVerify is set to true, it's not recommended to use that in production")
	}

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       certificates,
		MinVersion:         minTLSVersion,
		MaxVersion:         maxTLSVersion,
		ServerName:         serverTLSConfig.ServerName,
		InsecureSkipVerify: serverTLSConfig.InsecureSkipVerify,
		CurvePreferences:   curves,
		CipherSuites:       ciphers,
		ClientAuth:         clientAuth,
	}

	if clientAuth == tls.NoClientCert && len(clientTLSConfig.CACertFiles) > 0 {
		warns = append(warns, "NoClientCert checking specified, but client CAs provided")
	}

	if clientAuth != tls.NoClientCert {
		if len(clientTLSConfig.CACertFiles) == 0 {
			return nil, nil, merry2.Errorf("clientAuth set to '%v', but no client CAs provided", serverTLSConfig.ClientAuth)
		}
		clientCACertPool := x509.NewCertPool()

		for _, caCert := range clientTLSConfig.CACertFiles {
			cert, err := os.ReadFile(caCert)
			if err != nil {
				return nil, nil, err
			}
			clientCACertPool.AppendCertsFromPEM(cert)
		}
		tlsConfig.ClientCAs = clientCACertPool
	}

	return tlsConfig, warns, nil
}

// ParseClientTLSConfig parses TLSConfig as it should be used for HTTPS client mTLS and returns &tls.Config, list of
// warnings or error if parsing has failed.
// At this moment warnings are only about insecure ciphers
func ParseClientTLSConfig(serverTLSConfig *TLSConfig) (*tls.Config, []string, error) {
	caCertPool := x509.NewCertPool()

	for _, caCert := range serverTLSConfig.CACertFiles {
		cert, err := os.ReadFile(caCert)
		if err != nil {
			return nil, nil, err
		}
		caCertPool.AppendCertsFromPEM(cert)
	}

	if len(serverTLSConfig.CertificatePairs) == 0 {
		return nil, nil, merry2.Errorf("no server certificate pairs provided")
	}

	certificates := make([]tls.Certificate, 0, len(serverTLSConfig.CertificatePairs))

	for _, certPair := range serverTLSConfig.CertificatePairs {
		certificate, err := tls.LoadX509KeyPair(certPair.CertFile, certPair.PrivateKeyFile)
		if err != nil {
			return nil, nil, err
		}
		certificates = append(certificates, certificate)
	}

	minTLSVersion, err := ParseTLSVersion(serverTLSConfig.MinTLSVersion)
	if err != nil {
		return nil, nil, err
	}
	maxTLSVersion, err := ParseTLSVersion(serverTLSConfig.MaxTLSVersion)
	if err != nil {
		return nil, nil, err
	}

	curves, err := ParseCurves(serverTLSConfig.Curves)
	if err != nil {
		return nil, nil, err
	}

	ciphers, warns, err := CipherSuitesToUint16(serverTLSConfig.CipherSuites)
	if err != nil {
		return nil, nil, err
	}

	if serverTLSConfig.InsecureSkipVerify {
		warns = append(warns, "InsecureSkipVerify is set to true, it's not recommended to use that in production")
	}

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       certificates,
		MinVersion:         minTLSVersion,
		MaxVersion:         maxTLSVersion,
		ServerName:         serverTLSConfig.ServerName,
		InsecureSkipVerify: serverTLSConfig.InsecureSkipVerify,
		CurvePreferences:   curves,
		CipherSuites:       ciphers,
	}

	return tlsConfig, warns, nil
}
