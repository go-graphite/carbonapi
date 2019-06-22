package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/go-graphite/carbonapi/zipper/types"
	"github.com/lomik/zapwriter"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
)

func tagHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()

	// TODO: Migrate to context.WithTimeout
	ctx := r.Context()
	username, _, _ := r.BasicAuth()

	logger := zapwriter.Logger("tag").With(
		zap.String("carbonapi_uuid", uuid.String()),
		zap.String("username", username),
	)

	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)

	accessLogger := zapwriter.Logger("access")
	var accessLogDetails = &carbonapipb.AccessLogDetails{
		Handler:       "tags",
		Username:      username,
		CarbonapiUUID: uuid.String(),
		URL:           r.URL.Path,
		PeerIP:        srcIP,
		PeerPort:      srcPort,
		Host:          r.Host,
		Referer:       r.Referer(),
		URI:           r.RequestURI,
	}

	logAsError := false
	defer func() {
		deferredAccessLogging(accessLogger, accessLogDetails, t0, logAsError)
	}()

	err := r.ParseForm()
	if err != nil {
		logAsError = true
		w.Header().Set("Content-Type", contentTypeJSON)
		w.Write([]byte{'[', ']'})
		return
	}

	prettyStr := r.FormValue("pretty")
	limit := int64(-1)
	limitStr := r.FormValue("limit")
	if limitStr != "" {
		limit, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			logger.Debug("error parsing limit, ignoring",
				zap.String("limit", r.FormValue("limit")),
				zap.Error(err),
			)
			limit = -1
		}
	}

	q := r.URL.Query()
	q.Del("pretty")
	rawQuery := q.Encode()

	// TODO(civil): Implement caching
	var res []string
	if strings.HasSuffix(r.URL.Path, "tags") || strings.HasSuffix(r.URL.Path, "tags/") {
		res, err = config.Config.ZipperInstance.TagNames(ctx, rawQuery, limit)
	} else if strings.HasSuffix(r.URL.Path, "values") || strings.HasSuffix(r.URL.Path, "values/") {
		res, err = config.Config.ZipperInstance.TagValues(ctx, rawQuery, limit)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		accessLogDetails.HTTPCode = http.StatusNotFound
		return
	}

	// TODO(civil): Implement stats
	if err != nil && err != types.ErrNoMetricsFetched {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HTTPCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	var b []byte
	if prettyStr == "1" {
		b, err = json.MarshalIndent(res, "", "\t")
	} else {
		b, err = json.Marshal(res)
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HTTPCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.Write(b)
	accessLogDetails.Runtime = time.Since(t0).Seconds()
	accessLogDetails.HTTPCode = http.StatusOK
}
