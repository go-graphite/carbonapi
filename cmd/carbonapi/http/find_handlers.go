package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/go-graphite/carbonapi/intervalset"
	utilctx "github.com/go-graphite/carbonapi/util/ctx"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
	pickle "github.com/lomik/og-rek"
	"github.com/lomik/zapwriter"
	"github.com/satori/go.uuid"
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

func findTreejson(multiGlobs *pb.MultiGlobResponse) ([]byte, error) {
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

	err := json.NewEncoder(&b).Encode(tree)
	return b.Bytes(), err
}

type completer struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	IsLeaf string `json:"is_leaf"`
}

func findCompleter(multiGlobs *pb.MultiGlobResponse) ([]byte, error) {
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

func findList(multiGlobs *pb.MultiGlobResponse) ([]byte, error) {
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
	uuid := uuid.NewV4()
	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), config.Config.ZipperTimeout)
	ctx := utilctx.SetUUID(r.Context(), uuid.String())
	username, _, _ := r.BasicAuth()

	format := r.FormValue("format")
	jsonp := r.FormValue("jsonp")

	query := r.Form["query"]
	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)

	accessLogger := zapwriter.Logger("access")
	var accessLogDetails = carbonapipb.AccessLogDetails{
		Handler:        "find",
		Username:       username,
		CarbonapiUUID:  uuid.String(),
		URL:            r.URL.RequestURI(),
		PeerIP:         srcIP,
		PeerPort:       srcPort,
		Host:           r.Host,
		Referer:        r.Referer(),
		URI:            r.RequestURI,
		RequestHeaders: utilctx.GetLogHeaders(ctx),
	}

	logAsError := false
	defer func() {
		deferredAccessLogging(accessLogger, &accessLogDetails, t0, logAsError)
	}()

	if format == "completer" {
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

	if len(query) == 0 {
		http.Error(w, "missing parameter `query`", http.StatusBadRequest)
		accessLogDetails.HTTPCode = http.StatusBadRequest
		accessLogDetails.Reason = "missing parameter `query`"
		logAsError = true
		return
	}

	if format == "" {
		format = treejsonFormat
	}

	multiGlobs, stats, err := config.Config.ZipperInstance.Find(ctx, query)
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
	switch format {
	case treejsonFormat, jsonFormat:
		b, err = findTreejson(multiGlobs)
		format = jsonFormat
	case "completer":
		b, err = findCompleter(multiGlobs)
		format = jsonFormat
	case rawFormat:
		b, err = findList(multiGlobs)
		format = rawFormat
	case protobufFormat, protobuf3Format:
		b, err = multiGlobs.Marshal()
		format = protobufFormat
	case "", pickleFormat:
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
		err = pEnc.Encode(result)
		b = p.Bytes()
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HTTPCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	writeResponse(w, b, format, jsonp)
}
