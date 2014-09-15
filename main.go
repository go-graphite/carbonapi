package main

import (
	"encoding/json"
	"errors"
	"expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
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
	"github.com/peterbourgon/g2g"
)

type zipper string

var Zipper zipper

var Metrics = struct {
	Requests         *expvar.Int
	RequestCacheHits *expvar.Int

	FindRequests  *expvar.Int
	FindCacheHits *expvar.Int

	RenderRequests *expvar.Int
}{
	Requests:         expvar.NewInt("requests"),
	RequestCacheHits: expvar.NewInt("request_cache_hits"),

	FindRequests:  expvar.NewInt("find_requests"),
	FindCacheHits: expvar.NewInt("find_cache_hits"),

	RenderRequests: expvar.NewInt("render_requests"),
}

var queryCache bytesCache
var findCache bytesCache

var timeFormats = []string{"15:04 20060102", "20060102", "01/02/06"}

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
		t, err := time.Parse(format, s)
		if err == nil {
			return int32(t.Unix())
		}
	}
	return int32(d)
}

func intervalString(s string, defaultSign int) (int32, error) {

	var j int

	sign := defaultSign

	switch s[0] {
	case '-':
		sign = -1
		s = s[1:]
	case '+':
		sign = 1
		s = s[1:]
	}

	for j < len(s) && s[j] >= '0' && s[j] <= '9' {
		j++
	}
	offsetStr, unitStr := s[:j], s[j:]

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
		return 0, errors.New("unknown time units")
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return 0, err
	}

	return int32(sign * offset * units), nil
}

func (z zipper) Find(metric string) (pb.GlobResponse, error) {

	u, _ := url.Parse(string(z) + "/metrics/find/")

	u.RawQuery = url.Values{
		"query":  []string{metric},
		"format": []string{"protobuf"},
	}.Encode()

	var pbresp pb.GlobResponse

	err := z.get("Find", u, &pbresp)

	return pbresp, err
}

func (z zipper) Render(metric string, from, until int32) (pb.FetchResponse, error) {

	u, _ := url.Parse(string(z) + "/render/")

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
	resp, err := http.Get(u.String())
	if err != nil {
		log.Printf("%s: http.Get: %+v\n", who, err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%s: ioutil.ReadAll: %+v\n", who, err)
		return err
	}

	err = proto.Unmarshal(body, msg)
	if err != nil {
		log.Printf("%s: proto.Unmarshal: %+v\n", who, err)
		return err
	}

	return nil
}

type limiter chan struct{}

func (l limiter) enter() { l <- struct{}{} }
func (l limiter) leave() { <-l }

var Limiter limiter

// for testing
var timeNow = time.Now

type graphitePoint struct {
	value float64
	t     int32
}

type jsonResponse struct {
	Target     string          `json:"target"`
	Datapoints []graphitePoint `json:"datapoints"`
}

func (j jsonResponse) MarshalJSON() ([]byte, error) {
	var b []byte
	b = append(b, `{"target":`...)
	b = strconv.AppendQuoteToASCII(b, j.Target)
	b = append(b, `,"datapoints":[`...)

	var comma bool
	for _, v := range j.Datapoints {
		if comma {
			b = append(b, ',')
		}
		comma = true
		b = append(b, '[')
		if math.IsNaN(v.value) {
			b = append(b, "null"...)
		} else {
			b = strconv.AppendFloat(b, v.value, 'f', -1, 64)
		}
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(v.t), 10)
		b = append(b, ']')
	}
	b = append(b, `]}`...)
	return b, nil
}

func marshalRaw(r *pb.FetchResponse) ([]byte, error) {
	var b []byte
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
	return b, nil
}

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

	// make sure the cache key doesn't say noCache, because it will never hit
	r.Form.Del("noCache")

	cacheKey := r.Form.Encode()

	if response, ok := queryCache.get(cacheKey); useCache && ok {
		Metrics.RequestCacheHits.Add(1)
		switch format {
		case "json":
			w.Header().Set("Content-Type", "application/json")
		case "raw":
			w.Header().Set("Content-Type", "text/plain")
		}
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

		exprs := evalExpr(exp, from32, until32, metricMap)
		results = append(results, exprs...)
	}

	var body []byte
	var contentType string

	switch format {
	case "json":
		contentType = "application/json"

		var jresults []jsonResponse

		for _, r := range results {
			if r == nil {
				continue
			}
			datapoints := make([]graphitePoint, 0, len(r.Values))
			t := *r.StartTime
			for i, v := range r.Values {
				if r.IsAbsent[i] {
					v = math.NaN()
				}
				datapoints = append(datapoints, graphitePoint{value: v, t: t})
				t += *r.StepTime
			}
			jresults = append(jresults, jsonResponse{Target: r.GetName(), Datapoints: datapoints})
		}

		var err error
		body, err = json.Marshal(jresults)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

	case "raw":
		contentType = "text/plain"

		var newLine = false

		for _, j := range results {
			if newLine {
				body = append(body, '\n')
			}
			newLine = true

			rout, err := marshalRaw(j)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			body = append(body, rout...)
		}
	}

	queryCache.set(cacheKey, body, 60)

	w.Header().Set("Content-Type", contentType)
	w.Write(body)
}

func lbcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Ok\n")
}

func main() {

	z := flag.String("z", "", "zipper")
	port := flag.Int("p", 8080, "port")
	l := flag.Int("l", 20, "concurrency limit")
	mc := flag.String("mc", "", "comma separated memcached server list")
	cpus := flag.Int("cpus", 0, "number of CPUs to use")

	flag.Parse()

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

	if *mc != "" {
		servers := strings.Split(*mc, ",")
		log.Println("using memcache servers:", servers)
		queryCache = &memcachedCache{client: memcache.New(servers...)}
		findCache = &memcachedCache{client: memcache.New(servers...)}
	} else {
		queryCache = &expireCache{cache: make(map[string]cacheElement)}
		go queryCache.(*expireCache).cleaner()

		findCache = &expireCache{cache: make(map[string]cacheElement)}
		go findCache.(*expireCache).cleaner()
	}

	Zipper = zipper(*z)

	if *cpus != 0 {
		log.Println("using GOMAXPROCS", *cpus)
		runtime.GOMAXPROCS(*cpus)
	}

	if host := os.Getenv("GRAPHITEHOST") + ":" + os.Getenv("GRAPHITEPORT"); host != ":" {

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
	}

	http.HandleFunc("/render/", renderHandler)
	http.HandleFunc("/lbcheck", lbcheckHandler)

	log.Println("listening on port", *port)
	log.Fatalln(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
