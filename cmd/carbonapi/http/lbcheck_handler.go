package http

import (
	"net/http"
	"time"

	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

func lbcheckHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	accessLogger := zapwriter.Logger("access")

	_, _ = w.Write([]byte("Ok\n"))

	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)

	var accessLogDetails = carbonapipb.AccessLogDetails{
		Handler:  "lbcheck",
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
