package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"

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
	envPrefix := flag.String("envprefix", "CARBONAPI", "Preifx for environment variables override")
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

	r := carbonapiHttp.InitHandlers()
	handler := handlers.CompressHandler(r)
	handler = handlers.CORS()(handler)
	handler = handlers.ProxyHeaders(handler)

	err = gracehttp.Serve(&http.Server{
		Addr:    config.Config.Listen,
		Handler: handler,
	})

	if err != nil {
		logger.Fatal("gracehttp failed",
			zap.Error(err),
		)
	}
}
