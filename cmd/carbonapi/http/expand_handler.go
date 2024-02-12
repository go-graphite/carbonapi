package http

import (
	"encoding/json"
	"html"
	"net/http"
	"sort"
	"time"

	"github.com/ansel1/merry"
	pbv3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/lomik/zapwriter"
	uuid "github.com/satori/go.uuid"

	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/go-graphite/carbonapi/date"
	utilctx "github.com/go-graphite/carbonapi/util/ctx"
)

func expandHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uid := uuid.NewV4()
	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), config.Config.ZipperTimeout)
	ctx := utilctx.SetUUID(r.Context(), uid.String())
	username, _, _ := r.BasicAuth()
	requestHeaders := utilctx.GetLogHeaders(ctx)

	format, ok, formatRaw := getFormat(r, treejsonFormat)
	jsonp := r.FormValue("jsonp")
	groupByExpr := r.FormValue("groupByExpr")
	leavesOnly := r.FormValue("leavesOnly")

	qtz := r.FormValue("tz")
	from := r.FormValue("from")
	until := r.FormValue("until")
	from64 := date.DateParamToEpoch(from, qtz, timeNow().Add(-time.Hour).Unix(), config.Config.DefaultTimeZone)
	until64 := date.DateParamToEpoch(until, qtz, timeNow().Unix(), config.Config.DefaultTimeZone)

	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)

	accessLogger := zapwriter.Logger("access")
	var accessLogDetails = carbonapipb.AccessLogDetails{
		Handler:        "expand",
		Username:       username,
		CarbonapiUUID:  uid.String(),
		URL:            r.URL.RequestURI(),
		PeerIP:         srcIP,
		PeerPort:       srcPort,
		Host:           r.Host,
		Referer:        r.Referer(),
		URI:            r.RequestURI,
		Format:         formatRaw,
		RequestHeaders: requestHeaders,
	}

	logAsError := false
	defer func() {
		deferredAccessLogging(accessLogger, &accessLogDetails, t0, logAsError)
	}()

	err := r.ParseForm()
	if err != nil {
		setError(w, &accessLogDetails, err.Error(), http.StatusBadRequest, uid.String())
		logAsError = true
		return
	}
	query := r.Form["query"]

	if !ok || !format.ValidExpandFormat() {
		http.Error(w, "unsupported format: "+html.EscapeString(formatRaw), http.StatusBadRequest)
		accessLogDetails.HTTPCode = http.StatusBadRequest
		accessLogDetails.Reason = "unsupported format: " + formatRaw
		logAsError = true
		return
	}

	if queryLengthLimitExceeded(query, config.Config.MaxQueryLength) {
		setError(w, &accessLogDetails, "query length limit exceeded", http.StatusBadRequest, uid.String())
		logAsError = true
		return
	}

	var pv3Request pbv3.MultiGlobRequest
	pv3Request.Metrics = query
	pv3Request.StartTime = from64
	pv3Request.StopTime = until64

	multiGlobs, stats, err := config.Config.ZipperInstance.Find(ctx, pv3Request)
	if stats != nil {
		accessLogDetails.ZipperRequests = stats.ZipperRequests
		accessLogDetails.TotalMetricsCount += stats.TotalMetricsCount
	}
	if err != nil {
		returnCode := merry.HTTPCode(err)
		if returnCode != http.StatusOK || multiGlobs == nil {
			// Allow override status code for 404-not-found replies.
			if returnCode == http.StatusNotFound {
				returnCode = config.Config.NotFoundStatusCode
			}

			if returnCode < 300 {
				multiGlobs = &pbv3.MultiGlobResponse{Metrics: []pbv3.GlobResponse{}}
			} else {
				http.Error(w, http.StatusText(returnCode), returnCode)
				accessLogDetails.HTTPCode = int32(returnCode)
				accessLogDetails.Reason = err.Error()
				// We don't want to log this as an error if it's something normal
				// Normal is everything that is >= 500. So if config.Config.NotFoundStatusCode is 500 - this will be
				// logged as error

				if returnCode >= 500 {
					logAsError = true
				}
				return
			}
		}
	}

	var b []byte
	var err2 error
	b, err2 = expandEncoder(multiGlobs, leavesOnly, groupByExpr)
	err = merry.Wrap(err2)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HTTPCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	writeResponse(w, http.StatusOK, b, jsonFormat, jsonp, uid.String())
}

func expandEncoder(multiGlobs *pbv3.MultiGlobResponse, leavesOnly string, groupByExpr string) ([]byte, error) {
	var b []byte
	var err error
	groups := make(map[string][]string)
	seen := make(map[string]bool)
	for _, globs := range multiGlobs.Metrics {
		paths := make([]string, 0, len(globs.Matches))
		for _, g := range globs.Matches {
			if leavesOnly == "1" && !g.IsLeaf {
				continue
			}
			if _, ok := seen[g.Path]; ok {
				continue
			}
			seen[g.Path] = true
			paths = append(paths, g.Path)
		}
		sort.Strings(paths)
		groups[globs.Name] = paths
	}
	if groupByExpr != "1" {
		// results are just []string otherwise
		// so, flatting map
		flatData := make([]string, 0)
		for _, group := range groups {
			flatData = append(flatData, group...)
		}
		// sorting flat list one more to mimic graphite-web
		sort.Strings(flatData)
		data := map[string][]string{
			"results": flatData,
		}
		b, err = json.Marshal(data)
	} else {
		// results are map[string][]string
		data := map[string]map[string][]string{
			"results": groups,
		}
		b, err = json.Marshal(data)
	}
	return b, err
}
