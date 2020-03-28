package http

import (
	"encoding/json"
	"fmt"
	"github.com/ansel1/merry"
	"net/http"
	"time"

	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	utilctx "github.com/go-graphite/carbonapi/util/ctx"

	"github.com/lomik/zapwriter"
	"github.com/satori/go.uuid"
)

func infoHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()
	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), config.Config.ZipperTimeout)
	ctx := utilctx.SetUUID(r.Context(), uuid.String())
	username, _, _ := r.BasicAuth()
	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)
	format, ok, formatRaw := getFormat(r, jsonFormat)

	requestHeaders := utilctx.GetLogHeaders(ctx)

	accessLogger := zapwriter.Logger("access")
	var accessLogDetails = carbonapipb.AccessLogDetails{
		Handler:        "info",
		Username:       username,
		CarbonapiUUID:  uuid.String(),
		URL:            r.URL.RequestURI(),
		PeerIP:         srcIP,
		PeerPort:       srcPort,
		Host:           r.Host,
		Referer:        r.Referer(),
		Format:         formatRaw,
		URI:            r.RequestURI,
		RequestHeaders: requestHeaders,
	}

	logAsError := false
	defer func() {
		deferredAccessLogging(accessLogger, &accessLogDetails, t0, logAsError)
	}()

	if !ok || !format.ValidFindFormat() {
		http.Error(w, "unsupported format: "+formatRaw, http.StatusBadRequest)
		accessLogDetails.HTTPCode = http.StatusBadRequest
		accessLogDetails.Reason = "unsupported format: " + formatRaw
		logAsError = true
		return
	}

	query := r.Form["target"]
	if len(query) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		accessLogDetails.HTTPCode = http.StatusBadRequest
		accessLogDetails.Reason = "no target specified"
		logAsError = true
		return
	}

	data, stats, err := config.Config.ZipperInstance.Info(ctx, query)
	if stats != nil {
		accessLogDetails.ZipperRequests = stats.ZipperRequests
		accessLogDetails.TotalMetricsCount += stats.TotalMetricsCount
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HTTPCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	var b []byte
	var err2 error
	switch format {
	case jsonFormat:
		b, err2 = json.Marshal(data)
	case protoV3Format, protoV2Format:
		err2 = fmt.Errorf("not implemented yet")
	default:
		err2 = fmt.Errorf("unknown format %v", format)
	}
	err = merry.Wrap(err2)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HTTPCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	_, _ = w.Write(b)
	accessLogDetails.Runtime = time.Since(t0).Seconds()
	accessLogDetails.HTTPCode = http.StatusOK
}
