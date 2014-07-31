package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"code.google.com/p/gogoprotobuf/proto"

	"github.com/dgryski/httputil"
	cspb "github.com/grobian/carbonserver/carbonserverpb"
	pickle "github.com/kisielk/og-rek"
	"github.com/peterbourgon/g2g"
)

// global debugging level
var Debug int

// configuration values
var Config = struct {
	Backends []string
	MaxProcs int
	Port     int
	Buckets  int

	TimeoutMs               int
	TimeoutMsAfterFirstSeen int

	GraphiteHost string

	mu          sync.RWMutex
	metricPaths map[string][]string

	MaxIdleConnsPerHost int
}{
	MaxProcs: 1,
	Port:     8080,
	Buckets:  10,

	TimeoutMs:               2000,
	TimeoutMsAfterFirstSeen: 500,

	MaxIdleConnsPerHost: 100,

	metricPaths: make(map[string][]string),
}

// grouped expvars for /debug/vars and graphite
var Metrics = struct {
	FindRequests *expvar.Int
	FindErrors   *expvar.Int

	RenderRequests *expvar.Int
	RenderErrors   *expvar.Int

	Timeouts *expvar.Int
}{
	FindRequests: expvar.NewInt("find_requests"),
	FindErrors:   expvar.NewInt("find_errors"),

	RenderRequests: expvar.NewInt("render_requests"),
	RenderErrors:   expvar.NewInt("render_errors"),

	Timeouts: expvar.NewInt("timeouts"),
}

var logger multilog

type serverResponse struct {
	server   string
	response []byte
}

var storageClient = &http.Client{}

func singleGet(uri, server string, ch chan<- serverResponse) {

	u, err := url.Parse(server + uri)
	if err != nil {
		logger.Logln("error parsing uri: ", server+uri, ":", err)
		ch <- serverResponse{server, nil}
		return
	}
	req := http.Request{
		URL:    u,
		Header: make(http.Header),
	}

	resp, err := storageClient.Do(&req)
	if err != nil {
		logger.Logln("singleGet: error querying ", server, "/", uri, ":", err)
		ch <- serverResponse{server, nil}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// carbonsserver replies with Not Found if we request a
		// metric that it doesn't have -- makes sense
		ch <- serverResponse{server, nil}
		return
	}

	if resp.StatusCode != 200 {
		logger.Logln("bad response code ", server, "/", uri, ":", resp.StatusCode)
		ch <- serverResponse{server, nil}
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Logln("error reading body: ", server, "/", uri, ":", err)
		ch <- serverResponse{server, nil}
		return
	}

	ch <- serverResponse{server, body}
}

func multiGetFetch(servers []string, uri string) *cspb.FetchResponse {

	if Debug > 0 {
		logger.Logln("querying servers=", servers, "uri=", uri)
	}

	// buffered channel so the goroutines don't block on send
	ch := make(chan serverResponse, len(servers))
	for _, server := range servers {
		go singleGet(uri, server, ch)
	}

	isFirstResponse := true
	timeout := time.After(time.Duration(Config.TimeoutMs) * time.Millisecond)
	var gaps []int
	var ret *cspb.FetchResponse = nil
	var rservs []string

GATHER:
	for i := 0; i < len(servers); i++ {
		select {
		case r := <-ch:
			if r.response != nil {

				// we are satisfied once we don't have gaps any more
				var d cspb.FetchResponse
				err := proto.Unmarshal(r.response, &d)
				if err != nil {
					logger.Logf("error decoding protobuf response from server:%s: req:%s: err=%s", r.server, uri, err)
					if Debug > 1 {
						logger.Logln("\n" + hex.Dump(r.response))
					}
					Metrics.RenderErrors.Add(1)
					continue
				}

				if isFirstResponse {
					timeout = time.After(time.Duration(Config.TimeoutMsAfterFirstSeen) * time.Millisecond)
					ret = &d
					// record which elements are absent
					for i, v := range d.IsAbsent {
						if v {
							gaps = append(gaps, i)
						}
					}
					isFirstResponse = false
				} else if mergeValues(uri, ret, &d, &gaps) != true {
					Metrics.RenderErrors.Add(1)
					continue
				}

				if len(gaps) == 0 {
					if Debug > 2 {
						rservs = append(rservs, r.server)
						logger.Logln("found all information after querying servers=", rservs, "uri=", uri)
					}
					break GATHER
				}

				rservs = append(rservs, r.server)
			}

		case <-timeout:
			logger.Logln("Timeout waiting for more responses.  uri=", uri, ", servers=", servers, ", answers_from_servers=", rservs)
			Metrics.Timeouts.Add(1)
			break GATHER
		}
	}

	return ret
}

func multiGetFind(servers []string, uri string) []serverResponse {

	if Debug > 0 {
		logger.Logln("querying servers=", servers, "uri=", uri)
	}

	// buffered channel so the goroutines don't block on send
	ch := make(chan serverResponse, len(servers))
	for _, server := range servers {
		go singleGet(uri, server, ch)
	}

	var response []serverResponse

	isFirstResponse := true
	timeout := time.After(time.Duration(Config.TimeoutMs) * time.Millisecond)

GATHER:
	for i := 0; i < len(servers); i++ {
		select {
		case r := <-ch:
			if r.response != nil {

				response = append(response, r)

				if isFirstResponse {
					timeout = time.After(time.Duration(Config.TimeoutMsAfterFirstSeen) * time.Millisecond)
				}
				isFirstResponse = false
			}

		case <-timeout:
			servs := make([]string, 0)
			for _, r := range response {
				servs = append(servs, r.server)
			}
			logger.Logln("Timeout waiting for more responses.  uri=", uri, ", servers=", servers, ", answers_from_servers=", servs)
			Metrics.Timeouts.Add(1)
			break GATHER
		}
	}

	return response
}

func findHandlerPB(w http.ResponseWriter, req *http.Request, responses []serverResponse) ([]*cspb.GlobMatch, map[string][]string) {

	// metric -> [server1, ... ]
	paths := make(map[string][]string)

	var metrics []*cspb.GlobMatch
	for _, r := range responses {
		var metric cspb.GlobResponse
		err := proto.Unmarshal(r.response, &metric)
		if err != nil {
			logger.Logf("error decoding protobuf response from server:%s: req:%s: err=%s", r.server, req.URL.RequestURI(), err)
			if Debug > 1 {
				logger.Logln("\n" + hex.Dump(r.response))
			}
			Metrics.FindErrors.Add(1)
			continue
		}

		for _, match := range metric.Matches {
			p, ok := paths[*match.Path]
			if !ok {
				// we haven't seen this name yet
				// add the metric to the list of metrics to return
				metrics = append(metrics, match)
			}
			// add the server to the list of servers that know about this metric
			p = append(p, r.server)
			paths[*match.Path] = p
		}
	}

	return metrics, paths
}

func findHandler(w http.ResponseWriter, req *http.Request) {

	if Debug > 0 {
		logger.Logln("request: ", req.URL.RequestURI())
	}

	Metrics.FindRequests.Add(1)

	rewrite, _ := url.ParseRequestURI(req.URL.RequestURI())
	v := rewrite.Query()
	format := req.FormValue("format")
	v.Set("format", "protobuf")
	rewrite.RawQuery = v.Encode()

	responses := multiGetFind(Config.Backends, rewrite.RequestURI())

	if responses == nil || len(responses) == 0 {
		logger.Logln("find: error querying backends for: ", rewrite.RequestURI())
		http.Error(w, "find: error querying backends", http.StatusInternalServerError)
		return
	}

	metrics, paths := findHandlerPB(w, req, responses)

	// update our cache of which servers have which metrics
	Config.mu.Lock()
	for k, v := range paths {
		Config.metricPaths[k] = v
	}
	Config.mu.Unlock()

	switch format {
	case "protobuf":
		w.Header().Set("Content-Type", "application/protobuf")
		var result cspb.GlobResponse
		query := req.FormValue("query")
		result.Name = &query
		result.Matches = metrics
		b, _ := proto.Marshal(&result)
		w.Write(b)
	case "json":
		w.Header().Set("Content-Type", "application/json")
		jEnc := json.NewEncoder(w)
		jEnc.Encode(metrics)
	case "", "pickle":
		w.Header().Set("Content-Type", "application/pickle")

		var result []map[string]interface{}

		for _, metric := range metrics {
			mm := map[string]interface{}{
				"metric_path": *metric.Path,
				"isLeaf":      *metric.IsLeaf,
			}
			result = append(result, mm)
		}

		pEnc := pickle.NewEncoder(w)
		pEnc.Encode(result)
	}
}

func renderHandler(w http.ResponseWriter, req *http.Request) {

	if Debug > 0 {
		logger.Logln("request: ", req.URL.RequestURI())
	}

	Metrics.RenderRequests.Add(1)

	req.ParseForm()
	target := req.FormValue("target")

	if target == "" {
		http.Error(w, "empty target", http.StatusBadRequest)
		return
	}

	var serverList []string
	var ok bool

	Config.mu.RLock()
	// lookup the server list for this metric, or use all the servers if it's unknown
	if serverList, ok = Config.metricPaths[target]; !ok || serverList == nil || len(serverList) == 0 {
		serverList = Config.Backends
	}
	Config.mu.RUnlock()

	format := req.FormValue("format")
	rewrite, _ := url.ParseRequestURI(req.URL.RequestURI())
	v := rewrite.Query()
	v.Set("format", "protobuf")
	rewrite.RawQuery = v.Encode()

	response := multiGetFetch(serverList, rewrite.RequestURI())

	if response == nil {
		logger.Logln("render: error querying backends for:", req.URL.RequestURI(), "backends:", serverList)
		http.Error(w, "render: error querying backends", http.StatusInternalServerError)
		Metrics.RenderErrors.Add(1)
		return
	}

	returnRender(w, format, response)
}

func createRenderResponse(metric *cspb.FetchResponse, missing interface{}) map[string]interface{} {
	var pvalues []interface{}
	for i, v := range metric.Values {
		if metric.IsAbsent[i] {
			pvalues = append(pvalues, missing)
		} else {
			pvalues = append(pvalues, v)
		}
	}

	// create the response
	presponse := map[string]interface{}{
		"start":  metric.StartTime,
		"step":   metric.StepTime,
		"end":    metric.StopTime,
		"name":   metric.Name,
		"values": pvalues,
	}

	return presponse
}

func returnRender(w http.ResponseWriter, format string, metric *cspb.FetchResponse) {

	switch format {
	case "protobuf":
		w.Header().Set("Content-Type", "application/protobuf")
		b, _ := proto.Marshal(metric)
		w.Write(b)

	case "json":
		presponse := createRenderResponse(metric, nil)
		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w)
		e.Encode(presponse)

	case "", "pickle":
		presponse := createRenderResponse(metric, pickle.None{})
		w.Header().Set("Content-Type", "application/pickle")
		e := pickle.NewEncoder(w)
		e.Encode([]interface{}{presponse})
	}

}

func mergeValues(uri string, metric *cspb.FetchResponse, fill *cspb.FetchResponse, gaps *[]int) bool {

	if *metric.StartTime != *fill.StartTime {
		logger.Logf("request %s: responses have different startTimes: we have %d, but new response has %d", uri, int(*metric.StartTime), int(*fill.StartTime))
		return false
	}

	if *metric.StepTime != *fill.StepTime {
		logger.Logf("request %s: responses have different stepTimes: we have %d, but new response has %d", uri, int(*metric.StepTime), int(*fill.StepTime))
		return false
	}

	wgaps := *gaps
	for i := 0; i < len(wgaps); i = i + 1 {
		v := wgaps[i]
		if v < len(fill.IsAbsent) && !fill.IsAbsent[v] {
			metric.IsAbsent[v] = false
			metric.Values[v] = fill.Values[v]
			// this gap disappears, shuffle data around to make the
			// slice one less bigger
			newlen := len(wgaps) - 1
			if i < newlen {
				wgaps[i] = wgaps[newlen]
			}
			wgaps = wgaps[:newlen]
		}
	}
	*gaps = wgaps

	return true
}

func stripCommentHeader(cfg []byte) []byte {

	// strip out the comment header block that begins with '#' characters
	// as soon as we see a line that starts with something _other_ than '#', we're done

	idx := 0
	for cfg[0] == '#' {
		idx = bytes.Index(cfg, []byte("\n"))
		if idx == -1 || idx+1 == len(cfg) {
			return nil
		}
		cfg = cfg[idx+1:]
	}

	return cfg
}

func main() {

	configFile := flag.String("c", "", "config file (json)")
	port := flag.Int("p", 0, "port to listen on")
	maxprocs := flag.Int("maxprocs", 0, "GOMAXPROCS")
	flag.IntVar(&Debug, "d", 0, "enable debug logging")
	logStdout := flag.Bool("stdout", false, "write logging output also to stdout (default: only syslog)")

	flag.Parse()

	if *configFile == "" {
		log.Fatal("missing config file")
	}

	cfgjs, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal("unable to load config file:", err)
	}

	cfgjs = stripCommentHeader(cfgjs)

	if cfgjs == nil {
		log.Fatal("error removing header comment from ", *configFile)
	}

	err = json.Unmarshal(cfgjs, &Config)
	if err != nil {
		log.Fatal("error parsing config file: ", err)
	}

	if len(Config.Backends) == 0 {
		log.Fatal("no Backends loaded -- exiting")
	}

	// command line overrides config file

	if *port != 0 {
		Config.Port = *port
	}

	if *maxprocs != 0 {
		Config.MaxProcs = *maxprocs
	}

	// set up our logging
	slog, err := syslog.New(syslog.LOG_DAEMON, "carbonzipper")
	if err != nil {
		log.Fatal("can't obtain a syslog connection", err)
	}
	logger = append(logger, &sysLogger{w: slog})

	if *logStdout {
		logger = append(logger, &stdoutLogger{log.New(os.Stdout, "", log.LstdFlags)})
	}

	logger.Logln("setting GOMAXPROCS=", Config.MaxProcs)
	runtime.GOMAXPROCS(Config.MaxProcs)

	// +1 to track every over the number of buckets we track
	timeBuckets = make([]int64, Config.Buckets+1)

	httputil.PublishTrackedConnections("httptrack")
	expvar.Publish("requestBuckets", expvar.Func(renderTimeBuckets))

	http.HandleFunc("/metrics/find/", httputil.TrackConnections(httputil.TimeHandler(findHandler, bucketRequestTimes)))
	http.HandleFunc("/render/", httputil.TrackConnections(httputil.TimeHandler(renderHandler, bucketRequestTimes)))

	// nothing in the config? check the environment
	if Config.GraphiteHost == "" {
		if host := os.Getenv("GRAPHITEHOST") + ":" + os.Getenv("GRAPHITEPORT"); host != ":" {
			Config.GraphiteHost = host
		}
	}

	// only register g2g if we have a graphite host
	if Config.GraphiteHost != "" {

		logger.Logln("Using graphite host", Config.GraphiteHost)

		// register our metrics with graphite
		graphite, err := g2g.NewGraphite(Config.GraphiteHost, 60*time.Second, 10*time.Second)
		if err != nil {
			log.Fatal("unable to connect to to graphite: ", Config.GraphiteHost, ":", err)
		}

		hostname, _ := os.Hostname()
		hostname = strings.Replace(hostname, ".", "_", -1)

		graphite.Register(fmt.Sprintf("carbon.zipper.%s.find_requests", hostname), Metrics.FindRequests)
		graphite.Register(fmt.Sprintf("carbon.zipper.%s.find_errors", hostname), Metrics.FindErrors)

		graphite.Register(fmt.Sprintf("carbon.zipper.%s.render_requests", hostname), Metrics.RenderRequests)
		graphite.Register(fmt.Sprintf("carbon.zipper.%s.render_errors", hostname), Metrics.RenderErrors)

		graphite.Register(fmt.Sprintf("carbon.zipper.%s.timeouts", hostname), Metrics.Timeouts)

		for i := 0; i <= Config.Buckets; i++ {
			graphite.Register(fmt.Sprintf("carbon.zipper.%s.requests_in_%dms_to_%dms", hostname, i*100, (i+1)*100), bucketEntry(i))
		}
	}

	// configure the storage client
	storageClient.Transport = &http.Transport{
		MaxIdleConnsPerHost: Config.MaxIdleConnsPerHost,
	}

	portStr := fmt.Sprintf(":%d", Config.Port)
	logger.Logln("listening on", portStr)
	log.Fatal(http.ListenAndServe(portStr, nil))
}

var timeBuckets []int64

type bucketEntry int

func (b bucketEntry) String() string {
	return strconv.Itoa(int(atomic.LoadInt64(&timeBuckets[b])))
}

func renderTimeBuckets() interface{} {
	return timeBuckets
}

func bucketRequestTimes(req *http.Request, t time.Duration) {

	ms := t.Nanoseconds() / int64(time.Millisecond)

	bucket := int(ms / 100)

	if bucket < Config.Buckets {
		atomic.AddInt64(&timeBuckets[bucket], 1)
	} else {
		// Too big? Increment overflow bucket and log
		atomic.AddInt64(&timeBuckets[Config.Buckets], 1)
		logger.Logf("Slow Request: %s: %s", t.String(), req.URL.String())
	}
}

// trivial logging classes

// Logger is something that can log
type Logger interface {
	Log(string)
}

type stdoutLogger struct{ logger *log.Logger }

func (l *stdoutLogger) Log(s string) { l.logger.Print(s) }

type sysLogger struct{ w *syslog.Writer }

func (l *sysLogger) Log(s string) { l.w.Info(s) }

type multilog []Logger

func (ml multilog) Logln(a ...interface{}) {
	s := fmt.Sprintln(a...)
	for _, l := range ml {
		l.Log(s)
	}
}

func (ml multilog) Logf(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	for _, l := range ml {
		l.Log(s)
	}
}
