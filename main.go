package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

type serverResponse struct {
	server   string
	response []byte
}

func multiGet(servers []string, uri string) []serverResponse {

	ch := make(chan serverResponse)

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

			if Debug > 2 {
				d, _ := httputil.DumpRequest(&req, false)
				log.Println(string(d))
			}

			resp, err := http.DefaultClient.Do(&req)

			if Debug > 2 {
				d, _ := httputil.DumpResponse(resp, true)
				log.Println(string(d))
			}

			if err != nil || resp.StatusCode != 200 {
				log.Println("got status code", resp.StatusCode, "while querying", server)
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

	responses := multiGet(Config.Backends, req.URL.RequestURI())

	// metric -> [server1, ... ]
	paths := make(map[string][]string)

	var metrics []map[interface{}]interface{}
	for _, r := range responses {
		d := pickle.NewDecoder(bytes.NewReader(r.response))
		metric, err := d.Decode()
		if err != nil {
			log.Println("error during decode:", err)
			if Debug > 1 {
				log.Println("\n" + hex.Dump(r.response))
			}
			continue
		}

		for _, m := range metric.([]interface{}) {
			mm := m.(map[interface{}]interface{})
			name := mm["metric_path"].(string)
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

	req.ParseForm()
	target := req.FormValue("target")

	var serverList []string
	var ok bool

	Config.mu.RLock()
	// lookup the server list for this metric, or use all the servers if it's unknown
	if serverList, ok = Config.metricPaths[target]; !ok {
		serverList = Config.Backends
	}
	Config.mu.RUnlock()

	responses := multiGet(serverList, req.URL.RequestURI())

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
			log.Println("error during decode:", err)
			continue
		}

		marray := metric.([]interface{})
		decoded = append(decoded, marray)
	}

	if len(decoded) == 1 {
		w.Header().Set("Content-Type", "application/pickle")
		w.Write(responses[0].response)
	}

	// TODO: check len(d) == 1
	base := decoded[0][0].(map[interface{}]interface{})
	values := base["values"].([]interface{})

	for i := 0; i < len(values); i++ {
		if _, ok := values[i].(pickle.None); ok {
			// find one in the other values arrays
		replacenone:
			for other := 1; other < len(decoded); other++ {
				m := decoded[other][0].(map[interface{}]interface{})
				ovalues := m["values"].([]interface{})
				if _, ok := ovalues[i].(pickle.None); !ok {
					values[i] = ovalues[i]
					break replacenone
				}
			}
		}
	}

	// the first response is where we've been filling in our data, so we're ok just to serialize it as our response
	w.Header().Set("Content-Type", "application/pickle")
	e := pickle.NewEncoder(w)
	e.Encode(decoded[0])
}

func main() {

	configFile := flag.String("c", "", "config file (json)")
	port := flag.Int("p", 0, "port to listen on")
	maxprocs := flag.Int("maxprocs", 0, "GOMAXPROCS")
	flag.IntVar(&Debug, "d", 0, "enable debug logging")

	flag.Parse()

	if *configFile == "" {
		log.Fatal("missing config file")
	}

	cfgjs, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal("unable to load config file:", err)
	}

	// strip out the comment header block that begins with '#'
	// as soon as we see a line that starts with something _other_ than '#', we're done
	idx := 0
	for cfgjs[0] == '#' {
		idx = bytes.Index(cfgjs, []byte("\n"))
		if idx == -1 || idx+1 == len(cfgjs) {
			log.Fatal("error removing header comment from ", *configFile)
		}
		cfgjs = cfgjs[idx+1:]
	}

	err = json.Unmarshal(cfgjs, &Config)
	if err != nil {
		log.Fatal("error parsing config file: ", err)
	}

	// command line overrides config file

	if *port != 0 {
		Config.Port = *port
	}

	if *maxprocs != 0 {
		Config.MaxProcs = *maxprocs
	}

	log.Println("setting GOMAXPROCS=", Config.MaxProcs)
	runtime.GOMAXPROCS(Config.MaxProcs)

	http.HandleFunc("/metrics/find/", findHandler)
	http.HandleFunc("/render/", renderHandler)

	portStr := fmt.Sprintf(":%d", Config.Port)
	log.Println("listening on", portStr)
	log.Fatal(http.ListenAndServe(portStr, nil))
}
