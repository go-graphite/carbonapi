package main

import (
	"expvar"
	"flag"
	"log"
	"net/http"
	"net/http/pprof"
	_ "net/http/pprof"
	"sync"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	carbonapiHttp "github.com/go-graphite/carbonapi/cmd/carbonapi/http"
	"github.com/gorilla/handlers"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

// BuildVersion is provided to be overridden at build time. Eg. go build -ldflags -X 'main.BuildVersion=...'
var BuildVersion = "(development build)"

func main() {
	err := zapwriter.ApplyConfig([]zapwriter.Config{config.DefaultLoggerConfig})
	if err != nil {
		log.Fatal("Failed to initialize logger with default configuration")
	}
	logger := zapwriter.Logger("main")

	configPath := flag.String("config", "", "Path to the `config file`.")
	envPrefix := flag.String("envprefix", "CARBONAPI", "Prefix for environment variables override")
	if *envPrefix == "(empty)" {
		*envPrefix = ""
	}
	if *envPrefix == "" {
		logger.Warn("empty prefix is not recommended due to possible collisions with OS environment variables")
	}
	flag.Parse()
	config.SetUpViper(logger, configPath, *envPrefix)
	config.SetUpConfigUpstreams(logger)
	config.SetUpConfig(logger, BuildVersion)
	carbonapiHttp.SetupMetrics(logger)
	setupGraphiteMetrics(logger)

	config.Config.ZipperInstance = newZipper(carbonapiHttp.ZipperStats, &config.Config.Upstreams, config.Config.IgnoreClientTimeout, zapwriter.Logger("zipper"))

	r := carbonapiHttp.InitHandlers(config.Config.HeadersToPass, config.Config.HeadersToLog)
	handler := handlers.CompressHandler(r)
	handler = handlers.CORS()(handler)
	handler = handlers.ProxyHeaders(handler)

	wg := sync.WaitGroup{}
	if config.Config.Expvar.Enabled {
		if config.Config.Expvar.Listen != "" || config.Config.Expvar.Listen != config.Config.Listen {
			r := http.NewServeMux()
			r.HandleFunc(config.Config.Prefix+"/debug/vars", expvar.Handler().ServeHTTP)
			if config.Config.Expvar.PProfEnabled {
				r.HandleFunc(config.Config.Prefix+"/debug/pprof/", pprof.Index)
				r.HandleFunc(config.Config.Prefix+"/debug/pprof/cmdline", pprof.Cmdline)
				r.HandleFunc(config.Config.Prefix+"/debug/pprof/profile", pprof.Profile)
				r.HandleFunc(config.Config.Prefix+"/debug/pprof/symbol", pprof.Symbol)
				r.HandleFunc(config.Config.Prefix+"/debug/pprof/trace", pprof.Trace)
			}

			handler := handlers.CompressHandler(r)
			handler = handlers.CORS()(handler)
			handler = handlers.ProxyHeaders(handler)

			logger.Info("expvar handler will listen on a separate address/port",
				zap.String("expvar_listen", config.Config.Expvar.Listen),
				zap.Bool("pprof_enabled", config.Config.Expvar.PProfEnabled),
			)

			wg.Add(1)
			go func() {
				err = gracehttp.Serve(&http.Server{
					Addr:    config.Config.Expvar.Listen,
					Handler: handler,
				})

				if err != nil {
					logger.Fatal("failed to start http server",
						zap.Error(err),
					)
				}

				wg.Done()
			}()
		}
	}

	wg.Add(1)
	go func() {
		err = gracehttp.Serve(&http.Server{
			Addr:    config.Config.Listen,
			Handler: handler,
		})

		if err != nil {
			logger.Fatal("gracehttp failed",
				zap.Error(err),
			)
		}

		wg.Done()
	}()

	wg.Wait()
}
