package tlsconfig

import (
	"crypto/tls"

	"github.com/ansel1/merry/v2"
)

type ClientCertificatePairs struct {
	CertFile       string `mapstructure:"certFile"`
	PrivateKeyFile string `mapstructure:"privateKeyFile"`
}

type TLSConfig struct {
	CACertFiles      []string                 `mapstructure:"caCertFiles"`
	CertificatePairs []ClientCertificatePairs `mapstructure:"certificatePairs"`
	ClientAuth       string                   `mapstructure:"clientAuth"`

	ServerName         string   `mapstructure:"serverName"`
	InsecureSkipVerify bool     `mapstructure:"insecureSkipVerify"`
	MinTLSVersion      string   `mapstructure:"minTLSVersion"`
	MaxTLSVersion      string   `mapstructure:"maxTLSVersion"`
	CipherSuites       []string `mapstructure:"cipherSuites"`
	Curves             []string `mapstructure:"curves"`
}

// https://pkg.go.dev/crypto/tls#ClientAuthType
var supportedClientAuths = map[string]tls.ClientAuthType{
	"NoClientCert":               tls.NoClientCert,
	"RequestClientCert":          tls.RequestClientCert,
	"RequireAnyClientCert":       tls.RequireAnyClientCert,
	"VerifyClientCertIfGiven":    tls.VerifyClientCertIfGiven,
	"RequireAndVerifyClientCert": tls.RequireAndVerifyClientCert,
}

var supportedCurveIDs = map[string]tls.CurveID{
	"CurveP256": tls.CurveP256,
	"CurveP384": tls.CurveP384,
	"CurveP521": tls.CurveP521,
	"X25519":    tls.X25519,
}

var tlsVersionMap = map[string]uint16{
	"VersionTLS10": tls.VersionTLS10,
	"VersionTLS11": tls.VersionTLS11,
	"VersionTLS12": tls.VersionTLS12,
	"VersionTLS13": tls.VersionTLS13,
	"TLS10":        tls.VersionTLS10,
	"TLS11":        tls.VersionTLS11,
	"TLS12":        tls.VersionTLS12,
	"TLS13":        tls.VersionTLS13,
	"TLS1.0":       tls.VersionTLS10,
	"TLS1.1":       tls.VersionTLS11,
	"TLS1.2":       tls.VersionTLS12,
	"TLS1.3":       tls.VersionTLS13,
	"TLS 1.0":      tls.VersionTLS10,
	"TLS 1.1":      tls.VersionTLS11,
	"TLS 1.2":      tls.VersionTLS12,
	"TLS 1.3":      tls.VersionTLS13,
}

func ParseTLSVersion(tlsVersion string) (uint16, error) {
	if tlsVersion == "" {
		return tls.VersionTLS13, nil
	}
	if tv, ok := tlsVersionMap[tlsVersion]; ok {
		return tv, nil
	} else {
		return 0, merry.Errorf("invalid auth type specified: %v", tlsVersion)
	}
}

// ParseCurves returns list of tls.CurveIDs that can be passed to tls.Config or error if they are not supported
// ParseCurves also deduplicate input list
func ParseCurves(curveNames []string) ([]tls.CurveID, error) {
	inputCurveNamesMap := make(map[string]struct{})
	for _, name := range curveNames {
		inputCurveNamesMap[name] = struct{}{}
	}
	res := make([]tls.CurveID, 0, len(inputCurveNamesMap))
	for name := range inputCurveNamesMap {
		if id, ok := supportedCurveIDs[name]; ok {
			res = append(res, id)
		} else {
			return nil, merry.Errorf("invalid curve name specified: %v", name)
		}
	}

	return res, nil
}

func ParseClientAuthType(ClientAuth string) (tls.ClientAuthType, error) {
	if ClientAuth == "" {
		return tls.NoClientCert, nil
	}
	if id, ok := supportedClientAuths[ClientAuth]; ok {
		return id, nil
	} else {
		return tls.NoClientCert, merry.Errorf("invalid auth type specified: %v", ClientAuth)
	}
}

// CipherSuitesToUint16 for a given list of ciphers returns list of corresponding ids, list of insecure ciphers
// if cipher is unknown, it will return an error
func CipherSuitesToUint16(ciphers []string) ([]uint16, []string, error) {
	res := make([]uint16, 0)
	insecureCiphers := make([]string, 0)
	cipherList := tls.CipherSuites()

	cipherNames := make([]string, 0, len(cipherList))
	cipherSuites := make(map[string]uint16)

	for _, cipher := range cipherList {
		cipherSuites[cipher.Name] = cipher.ID
		if cipher.Insecure {
			insecureCiphers = append(insecureCiphers, cipher.Name)
		}
	}

	for _, c := range ciphers {
		if id, ok := cipherSuites[c]; ok {
			res = append(res, id)
		} else {
			return nil, nil, merry.Errorf("unknown cipher specified: %v, supported ciphers: %+v", c, cipherNames)
		}
	}

	return res, insecureCiphers, nil
}
