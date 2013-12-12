package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	_ "expvar"
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
	"sync"

	pickle "github.com/kisielk/og-rek"
)

var Debug int

var Config = struct {
	Backends []string
	MaxProcs int
	Port     int

	mu          sync.RWMutex
	metricPaths map[string][]string
}{
	MaxProcs:    1,
	Port:        8080,
	metricPaths: make(map[string][]string),
}

var logger multilog

type serverResponse struct {
	server   string
	response []byte
}

func multiGet(servers []string, uri string) []serverResponse {

	ch := make(chan serverResponse)

	if Debug > 0 {
		logger.Logln("querying servers=", servers, "uri=", uri)
	}

	for _, server := range servers {
		go func(server string, ch chan<- serverResponse) {

			u, err := url.Parse(server + uri)
			if err != nil {
				log.Fatal(err)
			}
			req := http.Request{
				URL:    u,
				Header: make(http.Header),
			}

			resp, err := http.DefaultClient.Do(&req)

			if err != nil || resp.StatusCode != 200 {
				logger.Logln("got status code", resp.StatusCode, "while querying", server, "/", uri)
				ch <- serverResponse{server, nil}
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ch <- serverResponse{server, nil}
			}

			ch <- serverResponse{server, body}
		}(server, ch)
	}

	var response []serverResponse

	for i := 0; i < len(servers); i++ {
		r := <-ch
		if r.response != nil {
			response = append(response, r)
		}
	}

	return response
}

func findHandler(w http.ResponseWriter, req *http.Request) {

	if Debug > 0 {
		logger.Logln("request: ", req.URL.RequestURI())
	}

	responses := multiGet(Config.Backends, req.URL.RequestURI())

	if responses == nil || len(responses) == 0 {
		logger.Logln("error querying backends for: ", req.URL.RequestURI())
		http.Error(w, "error querying backends", http.StatusInternalServerError)
		return
	}

	// metric -> [server1, ... ]
	paths := make(map[string][]string)

	var metrics []map[interface{}]interface{}
	for _, r := range responses {
		d := pickle.NewDecoder(bytes.NewReader(r.response))
		metric, err := d.Decode()
		if err != nil {
			logger.Logf("error decoding response from server:%s: req:%s: err=%s", r.server, req.URL.RequestURI(), err)
			if Debug > 1 {
				logger.Logln("\n" + hex.Dump(r.response))
			}
			continue
		}

		marray, ok := metric.([]interface{})
		if !ok {
			logger.Logf("bad type for metric:%t from server:%s: req:%s", metric, r.server, req.URL.RequestURI())
			http.Error(w, fmt.Sprintf("bad type for metric: %t", metric), http.StatusInternalServerError)
			return
		}

		for i, m := range marray {
			mm, ok := m.(map[interface{}]interface{})
			if !ok {
				logger.Logf("bad type for metric[%d]:%t from server:%s: req:%s", i, m, r.server, req.URL.RequestURI())
				http.Error(w, fmt.Sprintf("bad type for metric[%d]:%t", i, m), http.StatusInternalServerError)
				return
			}
			name, ok := mm["metric_path"].(string)
			if !ok {
				logger.Logf("bad type for metric_path:%t from server:%s: req:%s", mm["metric_path"], r.server, req.URL.RequestURI())
				http.Error(w, fmt.Sprintf("bad type for metric_path: %t", mm["metric_path"]), http.StatusInternalServerError)
				return
			}
			p, ok := paths[name]
			if !ok {
				// we haven't seen this name yet
				// add the metric to the list of metrics to return
				metrics = append(metrics, mm)
			}
			// add the server to the list of servers that know about this metric
			p = append(p, r.server)
			paths[name] = p
		}
	}

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
	if serverList, ok = Config.metricPaths[target]; !ok {
		serverList = Config.Backends
	}
	Config.mu.RUnlock()

	responses := multiGet(serverList, req.URL.RequestURI())

	if responses == nil || len(responses) == 0 {
		logger.Logln("error querying backends for: ", req.URL.RequestURI())
		http.Error(w, "error querying backends", http.StatusInternalServerError)
		return
	}

	// nothing to merge
	if len(responses) == 1 {
		w.Header().Set("Content-Type", "application/pickle")
		w.Write(responses[0].response)
	}

	// decode everything
	var decoded [][]interface{}
	for _, r := range responses {
		d := pickle.NewDecoder(bytes.NewReader(r.response))
		metric, err := d.Decode()
		if err != nil {
			logger.Logf("error decoding response from server:%s: req:%s: err=%s", r.server, req.URL.RequestURI(), err)
			if Debug > 1 {
				logger.Logln("\n" + hex.Dump(r.response))
			}
			continue
		}

		marray, ok := metric.([]interface{})
		if !ok {
			err := fmt.Sprintf("bad type for metric:%d from server:%s req:%s", metric, r.server, req.URL.RequestURI())
			logger.Logln(err)
			http.Error(w, err, http.StatusInternalServerError)
			return
		}
		if len(marray) == 0 {
			continue
		}
		decoded = append(decoded, marray)
	}

	if Debug > 2 {
		logger.Logf("request: %s: %v", req.URL.RequestURI(), decoded)
	}

	if len(decoded) == 0 {
		logger.Logf("no decoded responses to merge for req:%s", req.URL.RequestURI())
		w.Header().Set("Content-Type", "application/pickle")
		w.Write(responses[0].response)
		return
	}

	if len(decoded) == 1 {
		logger.Logf("only one decoded responses to merge for req:%s", req.URL.RequestURI())
		w.Header().Set("Content-Type", "application/pickle")
		// send back whatever data we have
		e := pickle.NewEncoder(w)
		e.Encode(decoded[0])
		return
	}

	if len(decoded[0]) != 1 {
		err := fmt.Sprintf("bad length for decoded[]:%d from req:%s", len(decoded[0]), req.URL.RequestURI())
		logger.Logln(err)
		http.Error(w, err, http.StatusInternalServerError)
		return
	}

	base, ok := decoded[0][0].(map[interface{}]interface{})
	if !ok {
		err := fmt.Sprintf("bad type for decoded:%t from req:%s", decoded[0][0], req.URL.RequestURI())
		logger.Logln(err)
		http.Error(w, err, http.StatusInternalServerError)
		return
	}

	values, ok := base["values"].([]interface{})
	if !ok {
		err := fmt.Sprintf("bad type for values:%t from req:%s", base["values"], req.URL.RequestURI())
		logger.Logln(err)
		http.Error(w, err, http.StatusInternalServerError)
		return
	}

fixValues:
	for i := 0; i < len(values); i++ {
		if _, ok := values[i].(pickle.None); ok {
			// find one in the other values arrays
			for other := 1; other < len(decoded); other++ {
				m, ok := decoded[other][0].(map[interface{}]interface{})
				if !ok {
					logger.Logln(fmt.Sprintf("bad type for decoded[%d][0]: %t", other, decoded[other][0]))
					break fixValues
				}

				ovalues, ok := m["values"].([]interface{})
				if !ok {
					logger.Logf("bad type for ovalues:%t from req:%s (skipping)", m["values"], req.URL.RequestURI())
					break fixValues
				}

				if len(ovalues) != len(values) {
					logger.Logf("unable to merge ovalues: len(values)=%d but len(ovalues)=%d", req.URL.RequestURI(), len(values), len(ovalues))
					logger.Logf("request: %s: %v", req.URL.RequestURI(), decoded)
					break fixValues
				}

				if _, ok := ovalues[i].(pickle.None); !ok {
					values[i] = ovalues[i]
					break
				}
			}
		}
	}

	// the first response is where we've been filling in our data, so we're ok just to serialize it as our response
	w.Header().Set("Content-Type", "application/pickle")
	e := pickle.NewEncoder(w)
	e.Encode(decoded[0])
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

	http.HandleFunc("/metrics/find/", findHandler)
	http.HandleFunc("/render/", renderHandler)

	portStr := fmt.Sprintf(":%d", Config.Port)
	logger.Logln("listening on", portStr)
	log.Fatal(http.ListenAndServe(portStr, nil))
}

// trivial logging classes

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
