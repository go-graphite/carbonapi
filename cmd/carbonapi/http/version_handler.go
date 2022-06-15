package http

import (
	"net/http"
	"time"

	"github.com/grafana/carbonapi/carbonapipb"
	"github.com/grafana/carbonapi/cmd/carbonapi/config"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

func versionHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	accessLogger := zapwriter.Logger("access")

	if config.Config.GraphiteWeb09Compatibility {
		_, _ = w.Write([]byte("0.9.15\n"))
	} else {
		_, _ = w.Write([]byte("1.1.0\n"))
	}

	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)
	var accessLogDetails = carbonapipb.AccessLogDetails{
		Handler:  "version",
		URL:      r.URL.RequestURI(),
		PeerIP:   srcIP,
		PeerPort: srcPort,
		Host:     r.Host,
		Referer:  r.Referer(),
		Runtime:  time.Since(t0).Seconds(),
		HTTPCode: http.StatusOK,
		URI:      r.RequestURI,
	}
	accessLogger.Info("request served", zap.Any("data", accessLogDetails))
}
