package main

import (
	"encoding/json"
	"errors"
	_ "expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/gogoprotobuf/proto"

	pb "github.com/dgryski/carbonzipper/carbonzipperpb"
)

type zipper string

var Zipper zipper

var queryCache bytesCache
var findCache bytesCache

var timeFormats = []string{"15:04 20060102", "20060102", "01/02/06"}

// dateParamToEpoch turns a passed string parameter into an epoch we can send to the Zipper
func dateParamToEpoch(s string, d int64) string {

	if s == "" {
		// return the default if nothing was passed
		return strconv.Itoa(int(d))
	}

	// relative timestamp
	if s[0] == '-' {

		offset, err := intervalString(s[1:])
		if err != nil {
			return strconv.Itoa(int(d))
		}

		return strconv.Itoa(int(timeNow().Add(-time.Duration(offset) * time.Second).Unix()))
	}

	_, err := strconv.Atoi(s)
	if err == nil && len(s) > 8 {
		return s // We got a timestamp so returning it
	}

	if strings.Contains(s, "_") {
		s = strings.Replace(s, "_", " ", 1) // Go can't parse _ in date strings
	}

	for _, format := range timeFormats {
		t, err := time.Parse(format, s)
		if err == nil {
			return strconv.Itoa(int(t.Unix()))
		}
	}
	return strconv.Itoa(int(d))
}

func intervalString(s string) (int32, error) {

	var j int

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

	return int32(offset * units), nil
}

// FIXME(dgryski): extract the http.Get + unproto code into its own function

func (z zipper) Find(metric string) (pb.GlobResponse, error) {

	u, _ := url.Parse(string(z) + "/metrics/find/")

	u.RawQuery = url.Values{
		"query":  []string{metric},
		"format": []string{"protobuf"},
	}.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		log.Printf("Find: http.Get: %+v\n", err)
		return pb.GlobResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Find: ioutil.ReadAll: %+v\n", err)
		return pb.GlobResponse{}, err
	}

	var pbresp pb.GlobResponse

	err = proto.Unmarshal(body, &pbresp)
	if err != nil {
		log.Printf("Find: proto.Unmarshal: %+v\n", err)
		return pb.GlobResponse{}, err
	}

	return pbresp, nil
}

func (z zipper) Render(metric, from, until string) (pb.FetchResponse, error) {

	u, _ := url.Parse(string(z) + "/render/")

	u.RawQuery = url.Values{
		"target": []string{metric},
		"format": []string{"protobuf"},
		"from":   []string{from},
		"until":  []string{until},
	}.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		log.Printf("Render: http.Get: %s: %+v\n", metric, err)
		return pb.FetchResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Render: ioutil.ReadAll: %s: %+v\n", metric, err)
		return pb.FetchResponse{}, err
	}

	var pbresp pb.FetchResponse

	err = proto.Unmarshal(body, &pbresp)
	if err != nil {
		log.Printf("Render: proto.Unmarshal: %s: %+v\n", metric, err)
		return pb.FetchResponse{}, err
	}

	return pbresp, nil
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

func (g graphitePoint) MarshalJSON() ([]byte, error) {
	// TODO(dgryski): fmt.Sprintf() is slow, use strconv.Append{Float,Int}
	if math.IsNaN(g.value) {
		return []byte(fmt.Sprintf("[null,%d]", g.t)), nil
	}
	return []byte(fmt.Sprintf("[%g,%d]", g.value, g.t)), nil
}

type jsonResponse struct {
	Target     string          `json:"target"`
	Datapoints []graphitePoint `json:"datapoints"`
}

func renderHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	targets := r.Form["target"]
	from := r.FormValue("from")
	until := r.FormValue("until")
	useCache := r.FormValue("noCache") == ""

	// make sure the cache key doesn't say noCache, because it will never hit
	r.Form.Del("noCache")

	cacheKey := r.Form.Encode()

	if response, ok := queryCache.get(cacheKey); useCache && ok {
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
		return
	}

	// normalize from and until values
	from = dateParamToEpoch(from, timeNow().Add(-24*time.Hour).Unix())
	until = dateParamToEpoch(until, timeNow().Unix())

	var results []*pb.FetchResponse

	for _, target := range targets {

		exp, e, err := parseExpr(target)
		if err != nil || e != "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		metricMap := make(map[string][]*pb.FetchResponse)

		for _, metric := range exp.metrics() {

			var glob pb.GlobResponse

			if response, ok := findCache.get(metric); ok {
				proto.Unmarshal(response, &glob)
			} else {
				var err error
				glob, err = Zipper.Find(metric)
				if err != nil {
					continue
				}
				b, err := proto.Marshal(&glob)
				if err == nil {
					findCache.set(metric, b)
				}
			}

			// For each metric returned in the Find response, query Render
			rch := make(chan *pb.FetchResponse, len(glob.GetMatches()))
			leaves := 0
			for _, m := range glob.GetMatches() {
				if !m.GetIsLeaf() {
					continue
				}
				leaves++
				Limiter.enter()
				go func(m *pb.GlobMatch) {
					var rptr *pb.FetchResponse
					r, err := Zipper.Render(m.GetPath(), from, until)
					if err == nil {
						rptr = &r
					}
					rch <- rptr
					Limiter.leave()
				}(m)
			}

			for i := 0; i < leaves; i++ {
				r := <-rch
				if r != nil {
					metricMap[metric] = append(metricMap[metric], r)
				}
			}
		}

		exprs := evalExpr(exp, metricMap)
		results = append(results, exprs...)
	}

	var jresults []jsonResponse

	for _, r := range results {
		if r == nil {
			log.Println("skipping nil result")
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

	jout, err := json.Marshal(jresults)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	queryCache.set(cacheKey, jout)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jout)
}

func main() {

	z := flag.String("z", "", "zipper")
	port := flag.Int("p", 8080, "port")
	l := flag.Int("l", 20, "concurrency limit")

	flag.Parse()

	if *z == "" {
		log.Fatal("no zipper (-z) provided")
	}

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

	queryCache = &expireCache{cache: make(map[string]cacheElement)}
	go queryCache.(*expireCache).cleaner()

	findCache = &expireCache{cache: make(map[string]cacheElement)}
	go findCache.(*expireCache).cleaner()

	Zipper = zipper(*z)

	http.HandleFunc("/render/", renderHandler)

	log.Println("listening on port", *port)
	log.Fatalln(http.ListenAndServe(":"+strconv.Itoa(*port), nil))

}
