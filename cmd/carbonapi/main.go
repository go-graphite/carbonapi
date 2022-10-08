package main

import (
	"context"
	"expvar"
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/helper"
	carbonapiHttp "github.com/go-graphite/carbonapi/cmd/carbonapi/http"
	"github.com/go-graphite/carbonapi/internal/dns"
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
	checkConfig := flag.Bool("check-config", false, "Check config file and exit.")
	envPrefix := flag.String("envprefix", "CARBONAPI", "Prefix for environment variables override")
	if *envPrefix == "(empty)" {
		*envPrefix = ""
	}
	if *envPrefix == "" {
		logger.Warn("empty prefix is not recommended due to possible collisions with OS environment variables")
	}
	flag.Parse()
	config.SetUpViper(logger, configPath, *envPrefix)
	if *checkConfig {
		os.Exit(0)
	}
	config.SetUpConfigUpstreams(logger)
	config.SetUpConfig(logger, BuildVersion)
	carbonapiHttp.SetupMetrics(logger)
	setupGraphiteMetrics(logger)

	if config.Config.UseCachingDNSResolver {
		logger.Info("will use custom caching dns resolver")
		dns.UseDNSCache(config.Config.CachingDNSRefreshTime)
	}

	config.Config.ZipperInstance = newZipper(carbonapiHttp.ZipperStats, &config.Config.Upstreams, config.Config.IgnoreClientTimeout, zapwriter.Logger("zipper"))

	wg := sync.WaitGroup{}
	serve := func(listen config.Listener, handler http.Handler) {
		l := &net.ListenConfig{Control: helper.ReusePort}
		h, p, err := net.SplitHostPort(listen.Address)
		if err != nil {
			logger.Fatal("failed to split address",
				zap.String("address", listen.Address),
				zap.Error(err),
			)
		}
		any := false
		if h == "" {
			h = "[::]"
			any = true
		}
		ips, err := net.LookupIP(h)
		if err != nil {
			// Fallback for a case where machine doesn't have ipv6 at all
			if any {
				h = "0.0.0.0"
				ips, err = net.LookupIP(h)
			}
			if err != nil {
				logger.Fatal("failed to resolve address",
					zap.String("address", h),
					zap.String("port", p),
					zap.Error(err),
				)
			}
		}
		// Resolve named ports
		port, err := net.LookupPort("tcp", p)
		if err != nil {
			logger.Fatal("failed to resolve port",
				zap.String("address", h),
				zap.String("port", p),
				zap.Error(err),
			)
		}
		for _, ip := range ips {
			address := (&net.TCPAddr{IP: ip, Port: port}).String()
			s := &http.Server{
				Addr:    address,
				Handler: handler,
			}
			listener, err := l.Listen(context.Background(), "tcp", address)
			if err != nil {
				logger.Fatal("failed to start http server",
					zap.Error(err),
				)
			}
			wg.Add(1)
			go func() {
				err = s.Serve(listener)

				if err != nil {
					logger.Fatal("failed to start http server",
						zap.Error(err),
					)
				}

				wg.Done()
			}()
		}
	}

	if config.Config.Expvar.Enabled {
		if config.Config.Expvar.Listen != "" && config.Config.Expvar.Listen != config.Config.Listeners[0].Address {
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

			listener := config.Listener{
				Address: config.Config.Expvar.Listen,
			}
			serve(listener, handler)
		}
	}

	r := carbonapiHttp.InitHandlers(config.Config.HeadersToPass, config.Config.HeadersToLog)
	handler := handlers.CompressHandler(r)
	handler = handlers.CORS()(handler)
	handler = handlers.ProxyHeaders(handler)

	for _, listener := range config.Config.Listeners {
		serve(listener, handler)
	}

	wg.Wait()

	if g != nil {
		g.Stop()
	}
	if carbonapiHttp.Gstatsd != nil {
		carbonapiHttp.Gstatsd.Close()
	}
}
