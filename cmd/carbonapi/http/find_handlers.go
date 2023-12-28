package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/ansel1/merry"
	pbv2 "github.com/go-graphite/protocol/carbonapi_v2_pb"
	pbv3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	pickle "github.com/lomik/og-rek"
	"github.com/lomik/zapwriter"
	"github.com/maruel/natural"
	uuid "github.com/satori/go.uuid"

	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/go-graphite/carbonapi/date"
	"github.com/go-graphite/carbonapi/intervalset"
	utilctx "github.com/go-graphite/carbonapi/util/ctx"
	"github.com/go-graphite/carbonapi/zipper/helper"
)

// Find handler and it's helper functions
type treejson struct {
	AllowChildren int            `json:"allowChildren"`
	Expandable    int            `json:"expandable"`
	Leaf          int            `json:"leaf"`
	ID            string         `json:"id"`
	Text          string         `json:"text"`
	Context       map[string]int `json:"context"` // unused
}

var treejsonContext = make(map[string]int)

func findTreejson(multiGlobs *pbv3.MultiGlobResponse) ([]byte, error) {
	var b bytes.Buffer

	var tree = make([]treejson, 0)

	seen := make(map[string]struct{})

	for _, globs := range multiGlobs.Metrics {
		basepath := globs.Name

		if i := strings.LastIndex(basepath, "."); i != -1 {
			basepath = basepath[:i+1]
		} else {
			basepath = ""
		}

		for _, g := range globs.Matches {
			if strings.HasPrefix(g.Path, "_tag") {
				continue
			}

			name := g.Path

			if i := strings.LastIndex(name, "."); i != -1 {
				name = name[i+1:]
			}

			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}

			t := treejson{
				ID:      basepath + name,
				Context: treejsonContext,
				Text:    name,
			}

			if g.IsLeaf {
				t.Leaf = 1
			} else {
				t.AllowChildren = 1
				t.Expandable = 1
			}

			tree = append(tree, t)
		}
	}

	sort.Slice(tree, func(i, j int) bool {
		if tree[i].Leaf < tree[j].Leaf {
			return true
		}
		if tree[i].Leaf > tree[j].Leaf {
			return false
		}
		return natural.Less(tree[i].Text, tree[j].Text)
	})

	err := json.NewEncoder(&b).Encode(tree)
	return b.Bytes(), err
}

type completer struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	IsLeaf string `json:"is_leaf"`
}

func findCompleter(multiGlobs *pbv3.MultiGlobResponse) ([]byte, error) {
	var b bytes.Buffer

	var complete = make([]completer, 0)

	for _, globs := range multiGlobs.Metrics {
		for _, g := range globs.Matches {
			if strings.HasPrefix(g.Path, "_tag") {
				continue
			}
			path := g.Path
			if !g.IsLeaf && path[len(path)-1:] != "." {
				path = g.Path + "."
			}
			c := completer{
				Path: path,
			}

			if g.IsLeaf {
				c.IsLeaf = "1"
			} else {
				c.IsLeaf = "0"
			}

			i := strings.LastIndex(c.Path, ".")

			if i != -1 {
				c.Name = c.Path[i+1:]
			} else {
				c.Name = g.Path
			}

			complete = append(complete, c)
		}
	}

	err := json.NewEncoder(&b).Encode(struct {
		Metrics []completer `json:"metrics"`
	}{
		Metrics: complete},
	)
	return b.Bytes(), err
}

func findList(multiGlobs *pbv3.MultiGlobResponse) ([]byte, error) {
	var b bytes.Buffer

	for _, globs := range multiGlobs.Metrics {
		for _, g := range globs.Matches {
			if strings.HasPrefix(g.Path, "_tag") {
				continue
			}

			var dot string
			// make sure non-leaves end in one dot
			if !g.IsLeaf && !strings.HasSuffix(g.Path, ".") {
				dot = "."
			}

			fmt.Fprintln(&b, g.Path+dot)
		}
	}

	return b.Bytes(), nil
}

func findHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uid := uuid.NewV4()
	carbonapiUUID := uid.String()
	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), config.Config.ZipperTimeout)
	ctx := utilctx.SetUUID(r.Context(), uid.String())
	username, _, _ := r.BasicAuth()
	requestHeaders := utilctx.GetLogHeaders(ctx)

	format, ok, formatRaw := getFormat(r, treejsonFormat)
	jsonp := r.FormValue("jsonp")

	qtz := r.FormValue("tz")
	from := r.FormValue("from")
	until := r.FormValue("until")
	from64 := date.DateParamToEpoch(from, qtz, timeNow().Add(-time.Hour).Unix(), config.Config.DefaultTimeZone)
	until64 := date.DateParamToEpoch(until, qtz, timeNow().Unix(), config.Config.DefaultTimeZone)

	query := r.Form["query"]
	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)

	accessLogger := zapwriter.Logger("access")
	var accessLogDetails = carbonapipb.AccessLogDetails{
		Handler:        "find",
		Username:       username,
		CarbonapiUUID:  carbonapiUUID,
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

	if !ok || !format.ValidFindFormat() {
		accessLogDetails.HTTPCode = http.StatusBadRequest
		accessLogDetails.Reason = "unsupported format: " + formatRaw
		logAsError = true
		writeErrorResponse(w, int(accessLogDetails.HTTPCode), accessLogDetails.Reason, carbonapiUUID)
		return
	}

	if format == completerFormat {
		var replacer = strings.NewReplacer("/", ".")
		for i := range query {
			query[i] = replacer.Replace(query[i])
			if query[i] == "" || query[i] == "/" || query[i] == "." {
				query[i] = ".*"
			} else {
				query[i] += "*"
			}
		}
	}

	var pv3Request pbv3.MultiGlobRequest

	if format == protoV3Format {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			accessLogDetails.HTTPCode = http.StatusBadRequest
			accessLogDetails.Reason = "failed to parse message body: " + err.Error()
			writeErrorResponse(w, http.StatusBadRequest, accessLogDetails.Reason, carbonapiUUID)
			return
		}

		err = pv3Request.Unmarshal(body)
		if err != nil {
			accessLogDetails.HTTPCode = http.StatusBadRequest
			accessLogDetails.Reason = "failed to parse message body: " + err.Error()
			writeErrorResponse(w, http.StatusBadRequest, accessLogDetails.Reason, carbonapiUUID)
			return
		}
	} else {
		pv3Request.Metrics = query
		pv3Request.StartTime = from64
		pv3Request.StopTime = until64
	}

	if len(pv3Request.Metrics) == 0 {
		accessLogDetails.HTTPCode = http.StatusBadRequest
		accessLogDetails.Reason = "missing parameter `query`"
		logAsError = true
		writeErrorResponse(w, http.StatusBadRequest, accessLogDetails.Reason, carbonapiUUID)
		return
	}

	accessLogDetails.Metrics = pv3Request.Metrics

	multiGlobs, stats, err := config.Config.ZipperInstance.Find(ctx, pv3Request)
	if stats != nil {
		accessLogDetails.ZipperRequests = stats.ZipperRequests
		accessLogDetails.TotalMetricsCount += stats.TotalMetricsCount
	}
	if err != nil {
		returnCode := merry.HTTPCode(helper.HttpErrorByCode(err))
		if returnCode != http.StatusOK || multiGlobs == nil {
			// Allow override status code for 404-not-found replies.
			if returnCode == http.StatusNotFound {
				returnCode = config.Config.NotFoundStatusCode
			}

			if returnCode < 300 {
				multiGlobs = &pbv3.MultiGlobResponse{Metrics: []pbv3.GlobResponse{}}
			} else {
				accessLogDetails.HTTPCode = int32(returnCode)
				accessLogDetails.Reason = err.Error()
				writeErrorResponse(w, returnCode, accessLogDetails.Reason, carbonapiUUID)
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
	switch format {
	case treejsonFormat, jsonFormat:
		b, err2 = findTreejson(multiGlobs)
		err = merry.Wrap(err2)
		format = jsonFormat
	case completerFormat:
		b, err2 = findCompleter(multiGlobs)
		err = merry.Wrap(err2)
		format = jsonFormat
	case rawFormat:
		b, err2 = findList(multiGlobs)
		err = merry.Wrap(err2)
		format = rawFormat
	case protoV2Format:
		r := pbv2.GlobResponse{
			Name:    multiGlobs.Metrics[0].Name,
			Matches: make([]pbv2.GlobMatch, 0, len(multiGlobs.Metrics)),
		}

		for i := range multiGlobs.Metrics {
			for _, m := range multiGlobs.Metrics[i].Matches {
				r.Matches = append(r.Matches, pbv2.GlobMatch{IsLeaf: m.IsLeaf, Path: m.Path})
			}
		}
		b, err2 = r.Marshal()
		err = merry.Wrap(err2)
	case protoV3Format:
		b, err2 = multiGlobs.Marshal()
		err = merry.Wrap(err2)
	case pickleFormat:
		var result []map[string]interface{}
		now := int32(time.Now().Unix() + 60)
		for _, globs := range multiGlobs.Metrics {
			for _, metric := range globs.Matches {
				if strings.HasPrefix(metric.Path, "_tag") {
					continue
				}
				// Tell graphite-web that we have everything
				var mm map[string]interface{}
				if config.Config.GraphiteWeb09Compatibility {
					// graphite-web 0.9.x
					mm = map[string]interface{}{
						// graphite-web 0.9.x
						"metric_path": metric.Path,
						"isLeaf":      metric.IsLeaf,
					}
				} else {
					// graphite-web 1.0
					interval := &intervalset.IntervalSet{Start: 0, End: now}
					mm = map[string]interface{}{
						"is_leaf":   metric.IsLeaf,
						"path":      metric.Path,
						"intervals": interval,
					}
				}
				result = append(result, mm)
			}
		}

		p := bytes.NewBuffer(b)
		pEnc := pickle.NewEncoder(p)
		err = merry.Wrap(pEnc.Encode(result))
		b = p.Bytes()
	}

	if err != nil {
		accessLogDetails.HTTPCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		writeErrorResponse(w, http.StatusInternalServerError, accessLogDetails.Reason, carbonapiUUID)
		return
	}

	writeResponse(w, http.StatusOK, b, format, jsonp, carbonapiUUID)
}
