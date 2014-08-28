package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/gogoprotobuf/proto"

	pb "github.com/dgryski/carbonzipper/carbonzipperpb"
)

type zipper string

var Zipper zipper

var timeFormats = []string{"15:04 20060102", "20060102", "01/02/06"}

// dateParamToEpoch turns a passed string parameter into an epoch we can send to the Zipper
func dateParamToEpoch(s string, d int64) string {

	if s == "" {
		// return the default if nothing was passed
		return strconv.Itoa(int(d))
	}

	// relative timestamp
	if s[0] == '-' {

		j := 1
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		offsetStr, unitStr := s[:j], s[j:]

		var units time.Duration
		switch unitStr {
		case "s", "sec", "secs", "second", "seconds":
			units = time.Second
		case "min", "minute", "minutes":
			units = time.Minute
		case "h", "hour", "hours":
			units = time.Hour
		case "d", "day", "days":
			units = 24 * time.Hour
		case "mon", "month", "months":
			units = 30 * 24 * time.Hour
		case "y", "year", "years":
			units = 365 * 24 * time.Hour
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return strconv.Itoa(int(d))
		}

		return strconv.Itoa(int(timeNow().Add(-time.Duration(offset) * units).Unix()))

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

func renderHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	targets := r.Form["target"]
	from := r.FormValue("from")
	until := r.FormValue("until")

	// normalize from and until values
	from = dateParamToEpoch(from, timeNow().Add(-24*time.Hour).Unix())
	until = dateParamToEpoch(until, timeNow().Unix())

	var results []*pb.FetchResponse
	// query zipper for find
	for _, target := range targets {
		glob, err := Zipper.Find(target)
		if err != nil {
			continue
		}

		// for each server in find response query render
		rch := make(chan *pb.FetchResponse, len(glob.GetMatches()))
		for _, m := range glob.GetMatches() {
			go func(m *pb.GlobMatch) {
				Limiter.enter()
				if m.GetIsLeaf() {
					r, err := Zipper.Render(m.GetPath(), from, until)
					if err != nil {
						rch <- nil
					} else {
						rch <- &r
					}
				} else {
					rch <- nil
				}
				Limiter.leave()
			}(m)
		}

		for i := 0; i < len(glob.GetMatches()); i++ {
			r := <-rch
			if r != nil {
				results = append(results, r)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	jEnc := json.NewEncoder(w)
	jEnc.Encode(results)
}

func main() {

	z := flag.String("z", "", "zipper")
	port := flag.Int("p", 8080, "port")
	l := flag.Int("l", 20, "concurrency limit")

	flag.Parse()

	if *z == "" {
		log.Fatal("no zipper (-z) provided")
	}

	Limiter = make(chan struct{}, *l)

	if _, err := url.Parse(*z); err != nil {
		log.Fatal("unable to parze zipper:", err)
	}

	Zipper = zipper(*z)

	http.HandleFunc("/render/", renderHandler)

	log.Println("listening on port", *port)
	log.Fatalln(http.ListenAndServe(":"+strconv.Itoa(*port), nil))

}
