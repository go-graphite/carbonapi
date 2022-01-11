package http

import (
	"net/http"
	"strings"
	"time"

	"encoding/json"
	"sort"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/go-graphite/carbonapi/date"
	utilctx "github.com/go-graphite/carbonapi/util/ctx"
	pbv3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/lomik/zapwriter"
	uuid "github.com/satori/go.uuid"
)

func expandHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uid := uuid.NewV4()
	ctx := utilctx.SetUUID(r.Context(), uid.String())
	username, _, _ := r.BasicAuth()
	requestHeaders := utilctx.GetLogHeaders(ctx)

	jsonp := r.FormValue("jsonp")
	groupByExpr := r.FormValue("groupByExpr")
	leavesOnly := r.FormValue("leavesOnly")

	qtz := r.FormValue("tz")
	from := r.FormValue("from")
	until := r.FormValue("until")
	from64 := date.DateParamToEpoch(from, qtz, timeNow().Add(-time.Hour).Unix(), config.Config.DefaultTimeZone)
	until64 := date.DateParamToEpoch(until, qtz, timeNow().Unix(), config.Config.DefaultTimeZone)

	query := r.Form["query"]
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
		RequestHeaders: requestHeaders,
	}

	logAsError := false
	defer func() {
		deferredAccessLogging(accessLogger, &accessLogDetails, t0, logAsError)
	}()

	var pv3Request = pbv3.MultiGlobRequest{
		Metrics:   query,
		StartTime: from64,
		StopTime:  until64,
	}

	if len(pv3Request.Metrics) == 0 {
		http.Error(w, "missing parameter `query`", http.StatusBadRequest)
		accessLogDetails.HTTPCode = http.StatusBadRequest
		accessLogDetails.Reason = "missing parameter `query`"
		logAsError = true
		return
	}

	multiGlobs, stats, err := config.Config.ZipperInstance.Expand(ctx, pv3Request)
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

	groups := make(map[string][]string)
	seen := make(map[string]struct{})
	for _, globs := range multiGlobs.Metrics {
		nodeCount := len(strings.Split(globs.Name, "."))
		names := make([]string, 0, len(globs.Matches))
		for _, g := range globs.Matches {
			if leavesOnly == "1" && !g.IsLeaf {
				continue
			}

			name := g.Path
			nodes := strings.SplitN(name, ".", nodeCount+1)
			if len(nodes) > nodeCount {
				name = strings.Join(nodes[:nodeCount], ".")
			} else {
				name = strings.Join(nodes, ".")
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
			sort.Strings(names)
		}
		groups[globs.Name] = names
	}

	data := map[string]interface{}{
		"results": groups,
	}
	if groupByExpr != "1" {
		flatData := make([]string, 0)
		for _, group := range groups {
			flatData = append(flatData, group...)
		}
		data["results"] = flatData
	}

	b, merr := json.Marshal(data)
	if merr != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HTTPCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	writeResponse(w, http.StatusOK, b, jsonFormat, jsonp, uid.String())
}
