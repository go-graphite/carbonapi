package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"code.google.com/p/gogoprotobuf/proto"

	cspb "github.com/grobian/carbonserver/carbonserverpb"
)

type zipper string

var Zipper zipper

// FIXME(dgryski): extract the http.Get + unproto code into its own function

func (z zipper) Find(metric string) (cspb.GlobResponse, error) {

	u, _ := url.Parse(string(z) + "/metrics/find/")

	u.RawQuery = url.Values{
		"query":  []string{metric},
		"format": []string{"protobuf"},
	}.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		log.Printf("Find: http.Get: %+v\n", err)
		return cspb.GlobResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Find: ioutil.ReadAll: %+v\n", err)
		return cspb.GlobResponse{}, err
	}

	var pbresp cspb.GlobResponse

	err = proto.Unmarshal(body, &pbresp)
	if err != nil {
		log.Printf("Find: proto.Unmarshal: %+v\n", err)
		return cspb.GlobResponse{}, err
	}

	return pbresp, nil
}

func (z zipper) Render(metric, from, until string) (cspb.FetchResponse, error) {

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
		return cspb.FetchResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Render: ioutil.ReadAll: %s: %+v\n", metric, err)
		return cspb.FetchResponse{}, err
	}

	var pbresp cspb.FetchResponse

	err = proto.Unmarshal(body, &pbresp)
	if err != nil {
		log.Printf("Render: proto.Unmarshal: %s: %+v\n", metric, err)
		return cspb.FetchResponse{}, err
	}

	return pbresp, nil
}

func renderHandler(w http.ResponseWriter, r *http.Request) {

	target := r.FormValue("target")
	from := r.FormValue("from")
	until := r.FormValue("until")

	// query zipper for find
	glob, err := Zipper.Find(target)
	if err != nil {
		return
	}

	var results []cspb.FetchResponse

	// for each server in find response query render
	// TODO(dgryski): run this in parallel
	for _, m := range glob.GetMatches() {
		if m.GetIsLeaf() {
			r, err := Zipper.Render(m.GetPath(), from, until)
			if err != nil {
				continue
			}
			results = append(results, r)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	jEnc := json.NewEncoder(w)
	jEnc.Encode(results)
}

func main() {

	z := flag.String("z", "", "zipper")
	port := flag.Int("p", 8080, "port")

	flag.Parse()

	if *z == "" {
		log.Fatal("no zipper (-z) provided")
	}

	if _, err := url.Parse(*z); err != nil {
		log.Fatal("unable to parze zipper:", err)
	}

	Zipper = zipper(*z)

	http.HandleFunc("/render/", renderHandler)

	log.Println("listening on port", *port)
	log.Fatalln(http.ListenAndServe(":"+strconv.Itoa(*port), nil))

}
