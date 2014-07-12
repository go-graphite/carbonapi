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
	UsePB    bool

	TimeoutMs               int
	TimeoutMsAfterFirstSeen int

	GraphiteHost string

	mu          sync.RWMutex
	metricPaths map[string][]string
}{
	MaxProcs: 1,
	Port:     8080,
	Buckets:  10,

	TimeoutMs:               2000,
	TimeoutMsAfterFirstSeen: 500,

	metricPaths: make(map[string][]string),
}

// grouped expvars for /debug/vars and graphite
var Metrics = struct {
	Requests *expvar.Int
	Errors   *expvar.Int
	Timeouts *expvar.Int
}{
	Requests: expvar.NewInt("requests"),
	Errors:   expvar.NewInt("errors"),
	Timeouts: expvar.NewInt("timeouts"),
}

var logger multilog

type serverResponse struct {
	server   string
	response []byte
}

var storageClient = &http.Client{Transport: &http.Transport{ResponseHeaderTimeout: 1 * time.Minute}}

func multiGet(servers []string, uri string) []serverResponse {

	if Debug > 0 {
		logger.Logln("querying servers=", servers, "uri=", uri)
	}

	// buffered channel so the goroutines don't block on send
	ch := make(chan serverResponse, len(servers))

	for _, server := range servers {
		go func(server string, ch chan<- serverResponse) {

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
				logger.Logln("multiGet: error querying ", server, "/", uri, ":", err)
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
		}(server, ch)
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
			logger.Logln("Timeout waiting for more responses: ", uri)
			Metrics.Timeouts.Add(1)
			break GATHER
		}
	}

	return response
}

func findHandlerPB(w http.ResponseWriter, req *http.Request, responses []serverResponse) ([]map[string]interface{}, map[string][]string) {

	// metric -> [server1, ... ]
	paths := make(map[string][]string)

	var metrics []map[string]interface{}
	for _, r := range responses {
		var metric cspb.GlobResponse
		err := proto.Unmarshal(r.response, &metric)
		if err != nil {
			logger.Logf("error decoding protobuf response from server:%s: req:%s: err=%s", r.server, req.URL.RequestURI(), err)
			if Debug > 1 {
				logger.Logln("\n" + hex.Dump(r.response))
			}
			Metrics.Errors.Add(1)
			continue
		}

		for _, match := range metric.Matches {
			p, ok := paths[*match.Path]
			if !ok {
				// we haven't seen this name yet
				// add the metric to the list of metrics to return
				mm := map[string]interface{}{
					"metric_path": *match.Path,
					"isLeaf":      *match.IsLeaf,
				}
				metrics = append(metrics, mm)
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

	Metrics.Requests.Add(1)

	requrl := req.URL
	if Config.UsePB {
		rewrite, _ := url.ParseRequestURI(req.URL.RequestURI())
		v := rewrite.Query()
		v.Set("format", "protobuf")
		rewrite.RawQuery = v.Encode()
		requrl = rewrite
	}

	responses := multiGet(Config.Backends, requrl.RequestURI())

	if responses == nil || len(responses) == 0 {
		logger.Logln("find: error querying backends for: ", requrl.RequestURI())
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

	w.Header().Set("Content-Type", "application/pickle")

	pEnc := pickle.NewEncoder(w)
	pEnc.Encode(metrics)

}

func renderHandler(w http.ResponseWriter, req *http.Request) {

	if Debug > 0 {
		logger.Logln("request: ", req.URL.RequestURI())
	}

	Metrics.Requests.Add(1)

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

	requrl := req.URL
	if Config.UsePB {
		rewrite, _ := url.ParseRequestURI(req.URL.RequestURI())
		v := rewrite.Query()
		v.Set("format", "protobuf")
		rewrite.RawQuery = v.Encode()
		requrl = rewrite
	}

	responses := multiGet(serverList, requrl.RequestURI())

	if responses == nil || len(responses) == 0 {
		logger.Logln("render: error querying backends for:", req.URL.RequestURI(), "backends:", serverList)
		http.Error(w, "render: error querying backends", http.StatusInternalServerError)
		Metrics.Errors.Add(1)
		return
	}

	handleRenderPB(w, req, responses)

}

func returnRender(w http.ResponseWriter, metric cspb.FetchResponse, pvalues []interface{}) {
	// create a pickle response
	presponse := map[string]interface{}{
		"start":  metric.StartTime,
		"step":   metric.StepTime,
		"end":    metric.StopTime,
		"name":   metric.Name,
		"values": pvalues,
	}

	w.Header().Set("Content-Type", "application/pickle")
	e := pickle.NewEncoder(w)
	e.Encode([]interface{}{presponse})
}

func handleRenderPB(w http.ResponseWriter, req *http.Request, responses []serverResponse) {

	var decoded []cspb.FetchResponse
	for _, r := range responses {
		var d cspb.FetchResponse
		err := proto.Unmarshal(r.response, &d)
		if err != nil {
			logger.Logf("error decoding protobuf response from server:%s: req:%s: err=%s", r.server, req.URL.RequestURI(), err)
			if Debug > 1 {
				logger.Logln("\n" + hex.Dump(r.response))
			}
			Metrics.Errors.Add(1)
			continue
		}
		decoded = append(decoded, d)
	}

	if Debug > 2 {
		logger.Logf("request: %s: %v", req.URL.RequestURI(), decoded)
	}

	if len(decoded) == 0 {
		err := fmt.Sprintf("no decoded responses to merge for req:%s", req.URL.RequestURI())
		logger.Logln(err)
		http.Error(w, err, http.StatusInternalServerError)
		Metrics.Errors.Add(1)
		return
	}

	if len(decoded) == 1 {
		if Debug > 0 {
			logger.Logf("only one decoded responses to merge for req:%s", req.URL.RequestURI())
		}
		metric := decoded[0]

		var pvalues []interface{}

		for i, v := range metric.Values {

			if metric.IsAbsent[i] {
				pvalues = append(pvalues, pickle.None{})
			} else {
				pvalues = append(pvalues, v)
			}
		}

		returnRender(w, metric, pvalues)

		return
	}

	metric := decoded[0]

	// the pickle response values
	var pvalues []interface{}

	var responseLengthMismatch bool
	for i, v := range metric.Values {
		if !metric.IsAbsent[i] {
			pvalues = append(pvalues, v)
			continue
		}

		if responseLengthMismatch {
			pvalues = append(pvalues, pickle.None{})
			continue
		}

		// found a missing value, find a replacement
		var foundReplacement bool
		for other := 1; other < len(decoded); other++ {

			m := decoded[other]

			if len(m.Values) != len(metric.Values) {
				logger.Logf("request: %s: unable to merge ovalues: len(values)=%d but len(ovalues)=%d", req.URL.RequestURI(), len(metric.Values), len(m.Values))
				// TODO(dgryski): we should remove
				// decoded[other] from the list of responses to
				// consider but this assumes that decoded[0] is
				// the 'highest resolution' response and thus
				// the one we want to keep, instead of the one
				// we want to discard

				Metrics.Errors.Add(1)
				responseLengthMismatch = true
				break
			}

			// found one
			if !m.IsAbsent[i] {
				pvalues = append(pvalues, m.Values[i])
				foundReplacement = true
				break
			}
		}

		if !foundReplacement {
			pvalues = append(pvalues, pickle.None{})
		}
	}

	returnRender(w, metric, pvalues)
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

		graphite.Register(fmt.Sprintf("carbon.zipper.%s.requests", hostname), Metrics.Requests)
		graphite.Register(fmt.Sprintf("carbon.zipper.%s.errors", hostname), Metrics.Errors)
		graphite.Register(fmt.Sprintf("carbon.zipper.%s.timeouts", hostname), Metrics.Timeouts)

		for i := 0; i <= Config.Buckets; i++ {
			graphite.Register(fmt.Sprintf("carbon.zipper.%s.requests_in_%dms_to_%dms", hostname, i*100, (i+1)*100), bucketEntry(i))
		}
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
