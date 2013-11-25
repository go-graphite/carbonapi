package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	pickle "github.com/kisielk/og-rek"
)

var Debug = false

var Config struct {
	Backends []string
}

func multiGet(servers []string, uri string) [][]byte {

	ch := make(chan []byte)

	for _, server := range servers {
		go func(server string, ch chan<- []byte) {

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
				log.Println("got status code", resp.StatusCode, "while querying", server)
				ch <- nil
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ch <- nil
			}

			ch <- body
		}(server, ch)
	}

	var response [][]byte

	for i := 0; i < len(servers); i++ {
		response = append(response, <-ch)
	}

	return response
}

func findHandler(w http.ResponseWriter, req *http.Request) {

	responses := multiGet(Config.Backends, req.URL.RequestURI())

	seenIds := make(map[string]bool)
	var metrics []map[interface{}]interface{}
	for _, r := range responses {
		d := pickle.NewDecoder(bytes.NewReader(r))
		metric, err := d.Decode()
		if err != nil {
			log.Println("error during decode:", err)
			continue
		}

		for _, m := range metric.([]interface{}) {
			mm := m.(map[interface{}]interface{})
			name := mm["metric_path"].(string)
			if !seenIds[name] {
				seenIds[name] = true
				metrics = append(metrics, mm)
			}
		}
	}

	w.Header().Set("Content-Type", "application/pickle")

	pEnc := pickle.NewEncoder(w)
	pEnc.Encode(metrics)
}

func renderHandler(w http.ResponseWriter, req *http.Request) {

	responses := multiGet(Config.Backends, req.URL.RequestURI())

	for _, r := range responses {
		d := pickle.NewDecoder(bytes.NewReader(r))
		metric, err := d.Decode()
		if err != nil {
			log.Println("error during decode:", err)
			continue
		}

		// TODO: merge metrics here
		// something like
		/*
		   base := metric[0]
		   for i := 0; i< len(base['values']); i++ {
		       if (base['values'][i] == pickle.None{}) {
		           // find one in the other values
		           for other := 1; i< len(metric); i++ {
		               if metric[other]["values"][i] != pickle.None{} {
		                   base['values'][i] = metric[other]["values"][i]
		               }
		           }
		       }
		   }
		*/

		_ = metric
	}

	// Fake it, for now.  Just return the first one
	w.Header().Set("Content-Type", "application/pickle")
	w.Write(responses[0])
}

func main() {

	configFile := flag.String("c", "", "config file (json)")
	port := flag.Int("p", 8080, "port to listen on")

	flag.Parse()

	if *configFile == "" {
		log.Fatal("missing config file")
	}

	cfgjs, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal("unable to load config file:", err)
	}

	json.Unmarshal(cfgjs, &Config)

	http.HandleFunc("/metrics/find/", findHandler)
	http.HandleFunc("/render/", renderHandler)

	portStr := fmt.Sprintf(":%d", *port)
	log.Println("listening on", portStr)
	log.Fatal(http.ListenAndServe(portStr, nil))
}
