package main

import (
	"context"
	"flag"
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
	Code           int                 `yaml:"httpCode"` // global responce code
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
	verbose := flag.Bool("verbose", false, "verbose reporting")
	testonly := flag.Bool("testonly", false, "run only unit test")
	noapp := flag.Bool("noapp", false, "do not run application")
	test := flag.Bool("test", false, "run unit test if present")
	breakOnError := flag.Bool("break", false, "break and wait user response if request failed")
	flag.Parse()
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	if *config == "" {
		logger.Fatal("failed to get config, it should be non-null")
	}

	f, err := os.Open(*config)
	if err != nil {
		logger.Fatal("failed to read config", zap.Error(err))
	}

	decoder := yaml.NewDecoder(f)
	decoder.SetStrict(true)
	err = decoder.Decode(&cfg)
	if err != nil {
		logger.Fatal("failed to read config", zap.Error(err))
		return
	}

	logger.Info("starting mockbackend",
		zap.Any("config", cfg),
	)

	httpServers := make([]*http.Server, 0)
	wg := sync.WaitGroup{}
	wgStart := sync.WaitGroup{}
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
			mux.HandleFunc("/tags/autoComplete/values", listener.tagsValuesHandler)
			mux.HandleFunc("/tags/autoComplete/tags", listener.tagsNamesHandler)

			wg.Add(1)
			wgStart.Add(1)
			server := &http.Server{
				Addr:    listener.Address,
				Handler: mux,
			}
			go func(h *http.Server) {
				wgStart.Done()
				err = h.ListenAndServe()
				if err != nil && err != http.ErrServerClosed {
					logger.Error("failed to start server",
						zap.Error(err),
					)
				}
				wg.Done()
			}(server)

			wgStart.Wait()
			httpServers = append(httpServers, server)
		}
		logger.Info("all listeners started")
	}

	failed := false
	if cfg.Test != nil && (*test || *testonly) {
		failed = e2eTest(logger, *noapp, *breakOnError, *verbose)
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
