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
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/pkg/tlsconfig"

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
	exactConfig := flag.Bool("exact-config", false, "Ensure that all config params are contained in the target struct.")
	envPrefix := flag.String("envprefix", "CARBONAPI", "Prefix for environment variables override")
	if *envPrefix == "(empty)" {
		*envPrefix = ""
	}
	if *envPrefix == "" {
		logger.Warn("empty prefix is not recommended due to possible collisions with OS environment variables")
	}
	flag.Parse()
	config.SetUpViper(logger, configPath, *exactConfig, *envPrefix)
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

	if err := config.Config.SetZipper(newZipper(carbonapiHttp.ZipperStats, &config.Config.Upstreams, config.Config.IgnoreClientTimeout, zapwriter.Logger("zipper"))); err != nil {
		logger.Fatal("failed to setup zipper",
			zap.Error(err),
		)
	}

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
		httpLogger, err := zap.NewStdLogAt(zapwriter.Logger("http"), zap.WarnLevel)
		if err != nil {
			logger.Fatal("failed to set up http server logger",
				zap.Error(err),
			)
		}

		servers := make([]*http.Server, 0)

		for _, ip := range ips {
			address := (&net.TCPAddr{IP: ip, Port: port}).String()
			s := &http.Server{
				Addr:     address,
				Handler:  handler,
				ErrorLog: httpLogger,
			}
			servers = append(servers, s)
			isTLS := false
			if len(listen.ServerTLSConfig.CACertFiles) > 0 {
				tlsConfig, warns, err := tlsconfig.ParseServerTLSConfig(&listen.ServerTLSConfig, &listen.ClientTLSConfig)
				if err != nil {
					logger.Fatal("failed to initialize TLS",
						zap.Error(err),
					)
				}
				if len(warns) != 0 {
					logger.Warn("insecure ciphers are in-use",
						zap.Strings("insecure_ciphers", warns),
					)
				}
				s.TLSConfig = tlsConfig
				isTLS = true
			}

			listener, err := l.Listen(context.Background(), "tcp", address)
			if err != nil {
				logger.Fatal("failed to start http server",
					zap.Error(err),
				)
			}
			wg.Add(1)
			go func(listener net.Listener, isTLS bool) {
				if isTLS {
					err = s.ServeTLS(listener, "", "")
				} else {
					err = s.Serve(listener)
				}

				if err != nil && err != http.ErrServerClosed {
					logger.Error("failed to start http server",
						zap.Error(err),
					)
				}

				wg.Done()
			}(listener, isTLS)
		}

		go func() {
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
			<-stop
			logger.Info("stoping carbonapi")
			// initiating the shutdown
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			for _, s := range servers {
				s.Shutdown(ctx)
			}
			cancel()
		}()

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
