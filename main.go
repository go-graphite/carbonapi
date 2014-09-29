package main

import (
	"bytes"
	"errors"
	"expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/gogoprotobuf/proto"
	pb "github.com/dgryski/carbonzipper/carbonzipperpb"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/dgryski/httputil"
	pickle "github.com/kisielk/og-rek"
	"github.com/peterbourgon/g2g"
)

type zipper struct {
	z      string
	client *http.Client
}

var Zipper zipper

var Metrics = struct {
	Requests         *expvar.Int
	RequestCacheHits *expvar.Int

	FindRequests  *expvar.Int
	FindCacheHits *expvar.Int

	RenderRequests *expvar.Int

	MemcacheTimeouts *expvar.Int

	CacheSize  expvar.Func
	CacheItems expvar.Func
}{
	Requests:         expvar.NewInt("requests"),
	RequestCacheHits: expvar.NewInt("request_cache_hits"),

	FindRequests:  expvar.NewInt("find_requests"),
	FindCacheHits: expvar.NewInt("find_cache_hits"),

	RenderRequests: expvar.NewInt("render_requests"),

	MemcacheTimeouts: expvar.NewInt("memcache_timeouts"),
}

var BuildVersion string = "(development build)"

var queryCache bytesCache
var findCache bytesCache

var timeFormats = []string{"15:04 20060102", "20060102", "01/02/06"}

var defaultTimeZone = time.Local

// dateParamToEpoch turns a passed string parameter into an epoch we can send to the Zipper
func dateParamToEpoch(s string, d int64) int32 {

	if s == "" {
		// return the default if nothing was passed
		return int32(d)
	}

	// relative timestamp
	if s[0] == '-' {

		offset, err := intervalString(s, -1)
		if err != nil {
			return int32(d)
		}

		return int32(timeNow().Add(time.Duration(offset) * time.Second).Unix())
	}

	if s == "now" {
		return int32(timeNow().Unix())
	}

	sint, err := strconv.Atoi(s)
	if err == nil && len(s) > 8 {
		return int32(sint) // We got a timestamp so returning it
	}

	if strings.Contains(s, "_") {
		s = strings.Replace(s, "_", " ", 1) // Go can't parse _ in date strings
	}

	for _, format := range timeFormats {
		t, err := time.ParseInLocation(format, s, defaultTimeZone)
		if err == nil {
			return int32(t.Unix())
		}
	}
	return int32(d)
}

var errUnknownTimeUnits = errors.New("unknown time units")

func intervalString(s string, defaultSign int) (int32, error) {

	sign := defaultSign

	switch s[0] {
	case '-':
		sign = -1
		s = s[1:]
	case '+':
		sign = 1
		s = s[1:]
	}

	var totalInterval int32
	for len(s) > 0 {
		var j int
		for j < len(s) && '0' <= s[j] && s[j] <= '9' {
			j++
		}
		var offsetStr string
		offsetStr, s = s[:j], s[j:]

		j = 0
		for j < len(s) && (s[j] < '0' || '9' < s[j]) {
			j++
		}
		var unitStr string
		unitStr, s = s[:j], s[j:]

		var units int
		switch unitStr {
		case "s", "sec", "secs", "second", "seconds":
			units = 1
		case "min", "minute", "minutes":
			units = 60
		case "h", "hour", "hours":
			units = 60 * 60
		case "d", "day", "days":
			units = 24 * 60 * 60
		case "w", "week", "weeks":
			units = 7 * 24 * 60 * 60
		case "mon", "month", "months":
			units = 30 * 24 * 60 * 60
		case "y", "year", "years":
			units = 365 * 24 * 60 * 60
		default:
			return 0, errUnknownTimeUnits
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return 0, err
		}
		totalInterval += int32(sign * offset * units)
	}

	return totalInterval, nil
}

func (z zipper) Find(metric string) (pb.GlobResponse, error) {

	u, _ := url.Parse(string(z.z) + "/metrics/find/")

	u.RawQuery = url.Values{
		"query":  []string{metric},
		"format": []string{"protobuf"},
	}.Encode()

	var pbresp pb.GlobResponse

	err := z.get("Find", u, &pbresp)

	return pbresp, err
}

func (z zipper) Render(metric string, from, until int32) (pb.FetchResponse, error) {

	u, _ := url.Parse(string(z.z) + "/render/")

	u.RawQuery = url.Values{
		"target": []string{metric},
		"format": []string{"protobuf"},
		"from":   []string{strconv.Itoa(int(from))},
		"until":  []string{strconv.Itoa(int(until))},
	}.Encode()

	var pbresp pb.FetchResponse

	err := z.get("Render", u, &pbresp)

	return pbresp, err
}

func (z zipper) get(who string, u *url.URL, msg proto.Message) error {
	resp, err := z.client.Get(u.String())
	if err != nil {
		return fmt.Errorf("http.Get: %+v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll: %+v", err)
	}

	err = proto.Unmarshal(body, msg)
	if err != nil {
		return fmt.Errorf("proto.Unmarshal: %+v", err)
	}

	return nil
}

type limiter chan struct{}

func (l limiter) enter() { l <- struct{}{} }
func (l limiter) leave() { <-l }

var Limiter limiter

// for testing
var timeNow = time.Now

func marshalJSON(results []*pb.FetchResponse, jsonp string) []byte {

	var b []byte

	if jsonp != "" {
		b = append(b, []byte(jsonp)...)
		b = append(b, '(')
	}

	b = append(b, '[')

	var topComma bool
	for _, r := range results {
		if r == nil {
			continue
		}

		if topComma {
			b = append(b, ',')
		}
		topComma = true

		b = append(b, `{"target":`...)
		b = strconv.AppendQuoteToASCII(b, r.GetName())
		b = append(b, `,"datapoints":[`...)

		var innerComma bool
		t := *r.StartTime
		for i, v := range r.Values {
			if innerComma {
				b = append(b, ',')
			}
			innerComma = true

			b = append(b, '[')

			if r.IsAbsent[i] {
				b = append(b, "null"...)
			} else {
				b = strconv.AppendFloat(b, v, 'f', -1, 64)
			}

			b = append(b, ',')

			b = strconv.AppendInt(b, int64(t), 10)

			b = append(b, ']')

			t += *r.StepTime
		}

		b = append(b, `]}`...)
	}

	b = append(b, ']')

	if jsonp != "" {
		b = append(b, ')')
	}

	return b
}

func marshalRaw(results []*pb.FetchResponse) []byte {

	var b []byte

	for _, r := range results {

		b = append(b, r.GetName()...)

		b = append(b, ',')
		b = strconv.AppendInt(b, int64(r.GetStartTime()), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(r.GetStopTime()), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(r.GetStepTime()), 10)
		b = append(b, '|')

		var comma bool
		for i, v := range r.Values {
			if comma {
				b = append(b, ',')
			}
			comma = true
			if r.IsAbsent[i] {
				b = append(b, "None"...)
			} else {
				b = strconv.AppendFloat(b, v, 'f', -1, 64)
			}
		}

		b = append(b, '\n')
	}
	return b
}

func marshalPickle(results []*pb.FetchResponse) []byte {

	var p []map[string]interface{}

	for _, r := range results {
		values := make([]interface{}, len(r.Values))
		for i, v := range r.Values {
			if r.IsAbsent[i] {
				values[i] = pickle.None{}
			} else {
				values[i] = v
			}

		}
		p = append(p, map[string]interface{}{
			"name":   r.GetName(),
			"start":  r.GetStartTime(),
			"end":    r.GetStopTime(),
			"step":   r.GetStepTime(),
			"values": values,
		})
	}

	var buf bytes.Buffer

	penc := pickle.NewEncoder(&buf)
	penc.Encode(p)

	return buf.Bytes()
}

const (
	contentTypeJSON   = "application/json"
	contentTypeRaw    = "text/plain"
	contentTypePickle = "application/pickle"
)

func renderHandler(w http.ResponseWriter, r *http.Request) {

	Metrics.Requests.Add(1)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	targets := r.Form["target"]
	from := r.FormValue("from")
	until := r.FormValue("until")
	format := r.FormValue("format")
	useCache := r.FormValue("noCache") == ""
	jsonp := r.FormValue("jsonp")

	cacheTimeout := int32(60)

	if tstr := r.FormValue("cacheTimeout"); tstr != "" {
		t, err := strconv.Atoi(tstr)
		if err != nil {
			log.Printf("failed to parse cacheTimeout: %v: %v", tstr, err)
		} else {
			cacheTimeout = int32(t)
		}
	}

	// make sure the cache key doesn't say noCache, because it will never hit
	r.Form.Del("noCache")

	cacheKey := r.Form.Encode()

	if response, ok := queryCache.get(cacheKey); useCache && ok {
		Metrics.RequestCacheHits.Add(1)
		switch format {
		case "json":
			w.Header().Set("Content-Type", contentTypeJSON)
		case "raw":
			w.Header().Set("Content-Type", contentTypeRaw)
		case "pickle":
			w.Header().Set("Content-Type", contentTypePickle)
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")

		w.Write(response)
		return
	}

	// normalize from and until values
	// BUG(dgryski): doesn't handle timezones the same as graphite-web
	from32 := dateParamToEpoch(from, timeNow().Add(-24*time.Hour).Unix())
	until32 := dateParamToEpoch(until, timeNow().Unix())

	var results []*pb.FetchResponse
	metricMap := make(map[metricRequest][]*pb.FetchResponse)

	for _, target := range targets {

		exp, e, err := parseExpr(target)
		if err != nil || e != "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		for _, m := range exp.metrics() {

			mfetch := m
			mfetch.from += from32
			mfetch.until += until32

			if _, ok := metricMap[mfetch]; ok {
				// already fetched this metric for this request
				continue
			}

			var glob pb.GlobResponse
			var haveCacheData bool

			if response, ok := findCache.get(m.metric); useCache && ok {
				Metrics.FindCacheHits.Add(1)
				err := proto.Unmarshal(response, &glob)
				haveCacheData = err == nil
			}

			if !haveCacheData {
				var err error
				Metrics.FindRequests.Add(1)
				glob, err = Zipper.Find(m.metric)
				if err != nil {
					log.Printf("Find: %v: %v", m.metric, err)
					continue
				}
				b, err := proto.Marshal(&glob)
				if err == nil {
					findCache.set(m.metric, b, 5*60)
				}
			}

			// For each metric returned in the Find response, query Render
			// This is a conscious decision to *not* cache render data
			rch := make(chan *pb.FetchResponse, len(glob.GetMatches()))
			leaves := 0
			for _, m := range glob.GetMatches() {
				if !m.GetIsLeaf() {
					continue
				}
				Metrics.RenderRequests.Add(1)
				leaves++
				Limiter.enter()
				go func(m *pb.GlobMatch, from, until int32) {
					var rptr *pb.FetchResponse
					r, err := Zipper.Render(m.GetPath(), from, until)
					if err == nil {
						rptr = &r
					} else {
						log.Printf("Render: %v: %v", m.GetPath(), err)
					}
					rch <- rptr
					Limiter.leave()
				}(m, mfetch.from, mfetch.until)
			}

			for i := 0; i < leaves; i++ {
				r := <-rch
				if r != nil {
					metricMap[mfetch] = append(metricMap[mfetch], r)
				}
			}
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					var buf [1024]byte
					runtime.Stack(buf[:], false)
					log.Printf("panic during eval: %s: %s\n%s\n", cacheKey, r, string(buf[:]))
				}
			}()
			exprs := evalExpr(exp, from32, until32, metricMap)
			results = append(results, exprs...)
		}()
	}

	var body []byte
	var contentType string

	switch format {
	case "json":
		contentType = contentTypeJSON
		body = marshalJSON(results, jsonp)

	case "raw":

		contentType = contentTypeRaw
		body = marshalRaw(results)

	case "pickle":
		contentType = contentTypePickle
		body = marshalPickle(results)
	}

	queryCache.set(cacheKey, body, cacheTimeout)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", contentType)
	w.Write(body)
}

func lbcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ok\n"))
}

func main() {

	z := flag.String("z", "", "zipper")
	port := flag.Int("p", 8080, "port")
	l := flag.Int("l", 20, "concurrency limit")
	cacheType := flag.String("cache", "mem", "cache type to use")
	mc := flag.String("mc", "", "comma separated memcached server list")
	memsize := flag.Int("memsize", 0, "in-memory cache size in MB (0 is unlimited)")
	cpus := flag.Int("cpus", 0, "number of CPUs to use")
	tz := flag.String("tz", "", "timezone,offset to use for dates with no timezone")
	graphiteHost := flag.String("graphite", "", "graphite destination host")

	flag.Parse()

	expvar.NewString("BuildVersion").Set(BuildVersion)
	log.Println("starting carbonapi", BuildVersion)

	if p := os.Getenv("PORT"); p != "" {
		*port, _ = strconv.Atoi(p)
	}

	Limiter = make(chan struct{}, *l)

	if *z == "" {
		log.Fatal("no zipper provided")
	}

	if _, err := url.Parse(*z); err != nil {
		log.Fatal("unable to parze zipper:", err)
	}

	log.Println("using zipper", *z)
	Zipper = zipper{
		z: *z,
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: *l / 2},
		},
	}

	switch *cacheType {
	case "memcache":
		if *mc == "" {
			log.Fatal("memcache cache requested but no memcache servers provided")
		}

		servers := strings.Split(*mc, ",")
		log.Println("using memcache servers:", servers)
		queryCache = &memcachedCache{client: memcache.New(servers...)}
		findCache = &memcachedCache{client: memcache.New(servers...)}

	case "mem":
		qcache := &expireCache{cache: make(map[string]cacheElement), maxSize: uint64(*memsize * 1024 * 1024)}
		queryCache = qcache
		go queryCache.(*expireCache).cleaner()

		findCache = &expireCache{cache: make(map[string]cacheElement)}
		go findCache.(*expireCache).cleaner()

		Metrics.CacheSize = expvar.Func(func() interface{} {
			qcache.Lock()
			size := qcache.totalSize
			qcache.Unlock()
			return size
		})
		expvar.Publish("cache_size", Metrics.CacheSize)

		Metrics.CacheItems = expvar.Func(func() interface{} {
			qcache.Lock()
			size := len(qcache.keys)
			qcache.Unlock()
			return size
		})
		expvar.Publish("cache_items", Metrics.CacheItems)

	case "null":
		queryCache = &nullCache{}
		findCache = &nullCache{}
	}

	if *tz != "" {
		fields := strings.Split(*tz, ",")
		if len(fields) != 2 {
			log.Fatalf("expected two fields for tz,seconds, got %d", len(fields))
		}

		var err error
		offs, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("unable to parse seconds: %s: %s", fields[1], err)
		}

		defaultTimeZone = time.FixedZone(fields[0], offs)
		log.Printf("using fixed timezone %s, offset %d ", defaultTimeZone.String(), offs)
	}

	if *cpus != 0 {
		log.Println("using GOMAXPROCS", *cpus)
		runtime.GOMAXPROCS(*cpus)
	}

	if envhost := os.Getenv("GRAPHITEHOST") + ":" + os.Getenv("GRAPHITEPORT"); envhost != ":" || *graphiteHost != "" {

		var host string

		switch {
		case envhost != ":" && *graphiteHost != "":
			host = *graphiteHost
		case envhost != ":":
			host = envhost
		case *graphiteHost != "":
			host = *graphiteHost
		}

		log.Println("Using graphite host", host)

		// register our metrics with graphite
		graphite, err := g2g.NewGraphite(host, 60*time.Second, 10*time.Second)
		if err != nil {
			log.Fatal("unable to connect to to graphite: ", host, ":", err)
		}

		hostname, _ := os.Hostname()
		hostname = strings.Replace(hostname, ".", "_", -1)

		graphite.Register(fmt.Sprintf("carbon.api.%s.requests", hostname), Metrics.Requests)
		graphite.Register(fmt.Sprintf("carbon.api.%s.request_cache_hits", hostname), Metrics.RequestCacheHits)

		graphite.Register(fmt.Sprintf("carbon.api.%s.find_requests", hostname), Metrics.FindRequests)
		graphite.Register(fmt.Sprintf("carbon.api.%s.find_cache_hits", hostname), Metrics.FindCacheHits)

		graphite.Register(fmt.Sprintf("carbon.api.%s.render_requests", hostname), Metrics.RenderRequests)

		graphite.Register(fmt.Sprintf("carbon.api.%s.memcache_timeouts", hostname), Metrics.MemcacheTimeouts)

		if Metrics.CacheSize != nil {
			graphite.Register(fmt.Sprintf("carbon.api.%s.cache_size", hostname), Metrics.CacheSize)
			graphite.Register(fmt.Sprintf("carbon.api.%s.cache_items", hostname), Metrics.CacheItems)
		}
	}

	httputil.PublishTrackedConnections("httptrack")

	http.HandleFunc("/render/", httputil.TrackConnections(httputil.TimeHandler(renderHandler, loggercb)))
	http.HandleFunc("/render", httputil.TrackConnections(httputil.TimeHandler(renderHandler, loggercb)))

	http.HandleFunc("/lb_check", lbcheckHandler)

	log.Println("listening on port", *port)
	log.Fatalln(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}

func loggercb(r *http.Request, d time.Duration) {
	log.Println(r.RequestURI, d)
}
