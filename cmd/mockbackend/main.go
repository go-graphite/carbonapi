package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type MainConfig struct {
	Version   string   `yaml:"version"`
	Test      *Schema  `yaml:"test"`
	Listeners []Config `yaml:"listeners"`
}

type Schema struct {
	Apps    []App
	Queries []Query
}

type App struct {
	Name   string
	Binary string
	Args   []string
}

type Query struct {
	Endpoint         string           `yaml:"endpoint"`
	Delay            int              `yaml:"delay"`
	URL              string           `yaml:"URL"`
	Type             string           `yaml:"type"`
	Body             string           `yaml:"body"`
	ExpectedResponse ExpectedResponse `yaml:"expectedResponse"`
}

type ExpectedResponse struct {
	HttpCode        int              `yaml:"httpCode"`
	ContentType     string           `yaml:"contentType"`
	ExpectedResults []ExpectedResult `yaml:"expectedResults"`
}

type ExpectedResult struct {
	SHA256 string `yaml:"sha256"`
}

type Config struct {
	Address        string              `yaml:"address"`
	Code           int                 `yaml:"httpCode"`
	ShuffleResults bool                `yaml:"shuffleResults"`
	EmptyBody      bool                `yaml:"emptyBody"`
	Expressions    map[string]Response `yaml:"expressions"`
}

var cfg = MainConfig{}

type listener struct {
	Config
	logger *zap.Logger
}

func doTest(logger *zap.Logger, t *Query) []string {
	client := http.Client{}
	failures := make([]string, 0)
	d, err := time.ParseDuration(fmt.Sprintf("%v", t.Delay) + "s")
	if err != nil {
		failures = append(failures, fmt.Sprintf("failed parse duration: %v", err))
		return failures
	}
	time.Sleep(d)
	ctx := context.Background()
	var body io.Reader
	if t.Type != "GET" {
		body = strings.NewReader(t.Body)
	}
	var resp *http.Response
	var contentType string
	u, err := url.Parse(t.URL)
	if err != nil {
		failures = append(failures, fmt.Sprintf("failed to parse URL: %v", err))
		return failures
	}
	req, err := http.NewRequestWithContext(ctx, t.Type, t.Endpoint+u.EscapedPath(), body)
	if err != nil {
		failures = append(failures, fmt.Sprintf("failed to prepare the request: %v", err))
		return failures
	}

	req.URL.RawQuery = req.URL.Query().Encode()

	resp, err = client.Do(req)
	if err != nil {
		failures = append(failures, fmt.Sprintf("failed to perform the request: %v", err))
	}

	if resp.StatusCode != t.ExpectedResponse.HttpCode {
		failures = append(failures,
			fmt.Sprintf("http code different, got %v, expected %v",
				resp.StatusCode,
				t.ExpectedResponse.HttpCode,
			),
		)
	}

	contentType = resp.Header.Get("Content-Type")
	if t.ExpectedResponse.ContentType != contentType {
		failures = append(failures,
			fmt.Sprintf("unexpected content-type, got %v, expected %v",
				resp.StatusCode,
				t.ExpectedResponse.HttpCode,
			),
		)
	}

	if contentType == "image/png" {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			failures = append(failures, fmt.Sprintf("failed to read body: %v", err))
			return failures
		}

		hash := sha256.Sum256(body)
		hashStr := fmt.Sprintf("%x", hash)
		if hashStr != t.ExpectedResponse.ExpectedResults[0].SHA256 {
			failures = append(failures, fmt.Sprintf("sha256 mismatch, got '%v', expected '%v'", hashStr, t.ExpectedResponse.ExpectedResults[0].SHA256))
			return failures
		}
	}

	return failures
}

func main() {
	config := flag.String("config", "average.yaml", "yaml where it would be possible to get data")
	flag.Parse()
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	if *config == "" {
		logger.Fatal("failed to get config, it should be non-null")
	}

	d, err := ioutil.ReadFile(*config)
	if err != nil {
		logger.Fatal("failed to read config", zap.Error(err))
	}

	err = yaml.Unmarshal(d, &cfg)
	if err != nil {
		logger.Fatal("failed to read config", zap.Error(err))
		return
	}

	logger.Info("starting mockbackend",
		zap.Any("config", cfg),
	)

	httpServers := make([]*http.Server, 0)
	wg := sync.WaitGroup{}
	for _, c := range cfg.Listeners {
		logger := logger.With(zap.String("listener", c.Address))
		listener := listener{
			Config: c,
			logger: logger,
		}

		if listener.Address == "" {
			listener.Address = ":9070"
		}

		if listener.Code == 0 {
			listener.Code = http.StatusOK
		}

		logger.Info("started",
			zap.String("listener", listener.Address),
			zap.Any("config", c),
		)

		mux := http.NewServeMux()
		mux.HandleFunc("/render", listener.renderHandler)
		mux.HandleFunc("/render/", listener.renderHandler)
		mux.HandleFunc("/metrics/find", listener.findHandler)
		mux.HandleFunc("/metrics/find/", listener.findHandler)

		wg.Add(1)
		server := &http.Server{
			Addr:    listener.Address,
			Handler: mux,
		}
		go func(h *http.Server) {
			err = h.ListenAndServe()
			if err != nil {
				logger.Error("failed to start server",
					zap.Error(err),
				)
			}
			wg.Done()
		}(server)

		httpServers = append(httpServers, server)
	}
	logger.Info("all listeners started")

	failed := false
	if cfg.Test != nil {
		logger.Info("will run test",
			zap.Any("config", cfg.Test),
		)
		runningApps := make(map[string]*runner)
		for i, c := range cfg.Test.Apps {
			r := NewRunner(&cfg.Test.Apps[i], logger)
			runningApps[c.Name] = r
			go r.Run()
		}

		logger.Info("will sleep for 5 seconds to start all required apps")
		time.Sleep(5 * time.Second)

		for _, t := range cfg.Test.Queries {
			failures := doTest(logger, &t)

			if len(failures) != 0 {
				failed = true
				logger.Error("test failed",
					zap.Strings("failures", failures),
				)
			} else {
				logger.Info("test OK")
			}
		}

		logger.Info("shutting down running application")
		for _, v := range runningApps {
			v.Finish()
		}

		if failed {
			logger.Error("tests failed")
		} else {
			logger.Info("All tests OK")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for i := range httpServers {
			// we don't care about error here
			_ = httpServers[i].Shutdown(ctx)
		}
	}

	wg.Wait()
	if failed {
		os.Exit(1)
	}
}
