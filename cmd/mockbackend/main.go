package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type MainConfig struct {
	Version   string      `yaml:"version"`
	Test      *TestSchema `yaml:"test"`
	Listeners []Listener  `yaml:"listeners"`
}

type Listener struct {
	Address        string              `yaml:"address"`
	Code           int                 `yaml:"httpCode"`
	ShuffleResults bool                `yaml:"shuffleResults"`
	EmptyBody      bool                `yaml:"emptyBody"`
	Expressions    map[string]Response `yaml:"expressions"`
}

var cfg = MainConfig{}

type listener struct {
	Listener
	logger *zap.Logger
}

func main() {
	config := flag.String("config", "average.yaml", "yaml where it would be possible to get data")
	testonly := flag.Bool("testonly", false, "run only unit test")
	noapp := flag.Bool("noapp", false, "do not run application")
	test := flag.Bool("test", false, "run unit test if present")
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
	if !*testonly {
		for _, c := range cfg.Listeners {
			logger := logger.With(zap.String("listener", c.Address))
			listener := listener{
				Listener: c,
				logger:   logger,
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
	}

	failed := false
	if cfg.Test != nil && (*test || *testonly) {
		failed = e2eTest(logger, *noapp)
	}

	if !*testonly {
		if *test {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			for i := range httpServers {
				// we don't care about error here
				_ = httpServers[i].Shutdown(ctx)
			}
			cancel()
		}

		wg.Wait()
	}

	if failed {
		// skipcq: CRT-D0011
		os.Exit(1)
	}
}
