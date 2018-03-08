package graphiteWeb

import (
	"encoding/json"
	"fmt"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/go-graphite/carbonzipper/limiter"
	"github.com/lomik/zapwriter"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type graphiteWeb struct {
	interfaces.FunctionBase

	working     bool
	strict      bool
	fallbackUrl string
	proxy       *http.Client

	supportedFunctions map[string]*types.FunctionDescription
	limiter            limiter.ServerLimiter
}

func GetOrder() interfaces.Order {
	return interfaces.Last
}

type graphiteWebConfig struct {
	FallbackUrl              string
	Strict                   bool
	MaxConcurrentConnections int
	Timeout                  time.Duration
	KeelAliveInterval        time.Duration
	ForceRemove              []string
	ForceAdd                 []string
}

func paramsIsEqual(first, second []types.FunctionParam) bool {
	if len(first) != len(second) {
		return false
	}
	for i, p1 := range first {
		p2 := second[i]
		equal := p1.Name == p2.Name && p1.Type == p2.Type
		if !equal {
			return false
		}
	}
	return true
}

func New(configFile string) []interfaces.FunctionMetadata {
	logger := zapwriter.Logger("graphiteWeb fallback")
	v := viper.New()
	v.SetConfigFile(configFile)

	cfg := graphiteWebConfig{}
	err := v.Unmarshal(&cfg)
	if err != nil {
		logger.Fatal("failed to parse config",
			zap.Error(err),
		)
	}

	f := &graphiteWeb{
		limiter: limiter.NewServerLimiter([]string{cfg.FallbackUrl}, cfg.MaxConcurrentConnections),
		proxy: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: cfg.MaxConcurrentConnections,
				DialContext: (&net.Dialer{
					Timeout:   cfg.Timeout,
					KeepAlive: cfg.KeelAliveInterval,
					DualStack: true,
				}).DialContext,
			},
		},
		fallbackUrl: cfg.FallbackUrl,
		strict:      cfg.Strict,
		supportedFunctions: map[string]*types.FunctionDescription{
			"graphiteWeb": {
				Description: "This is special function which will pass everything inside to graphiteWeb (if configured)",
				Function:    "graphiteWeb(seriesList)",
				Group:       "Fallback",
				Module:      "graphite.render.fallback.custom",
				Name:        "example",
				Params: []types.FunctionParam{
					{
						Name:     "seriesList",
						Required: true,
						Type:     types.SeriesList,
					},
				},
			},
		},
	}

	req, err := http.NewRequest("GET", f.fallbackUrl+"/functions/?format=json", nil)
	if err != nil {
		logger.Fatal("failed to create list of functions",
			zap.Error(err),
		)
	}

	resp, err := f.proxy.Do(req)
	if err != nil {
		logger.Fatal("failed to obtain list of functions",
			zap.Error(err),
		)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to obtain list of functions",
			zap.Error(fmt.Errorf("return code is not 200 OK")),
			zap.Int("status_code", resp.StatusCode),
		)
		return []interfaces.FunctionMetadata{}
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("failed to obtain list of functions",
			zap.Error(fmt.Errorf("return code is not 200 OK")),
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return []interfaces.FunctionMetadata{}
	}

	graphiteWebSupportedFunctions := make(map[string]*types.FunctionDescription, 0)

	err = json.Unmarshal(body, &graphiteWebSupportedFunctions)
	if err != nil {
		logger.Error("failed to parse list of functions",
			zap.Error(err),
		)
		return []interfaces.FunctionMetadata{}
	}

	functions := []string{"graphiteWeb"}
	metadata.FunctionMD.RLock()
	for k, v := range graphiteWebSupportedFunctions {
		replace := false
		for _, n := range cfg.ForceAdd {
			if k == n {
				replace = true
				break
			}
		}
		if v2, ok := metadata.FunctionMD.Descriptions[k]; !replace && ok {
			equals := true
			if f.strict {
				equals = paramsIsEqual(v.Params, v2.Params)
			}
			if equals {
				continue
			}
			replace = true
		}
		for _, n := range cfg.ForceRemove {
			if k == n {
				replace = false
				break
			}
		}
		if replace {
			functions = append(functions, k)
			f.supportedFunctions[k] = v
		}
	}
	metadata.FunctionMD.RUnlock()

	logger.Info("will handle following functions",
		zap.Strings("functions_metadata", functions),
	)

	res := make([]interfaces.FunctionMetadata, 0, len(functions))
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f, Order: interfaces.Any})
	}
	return res
}

func (f *graphiteWeb) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if !f.working {
		return nil, nil
	}

	rewrite, _ := url.Parse(f.fallbackUrl + "/render/")

	v := url.Values{
		"target": []string{e.RawArgs()},
		"from":   []string{strconv.FormatInt(int64(from), 10)},
		"until":  []string{strconv.FormatInt(int64(until), 10)},
		"format": []string{"pickle"},
	}

	rewrite.RawQuery = v.Encode()

	f.limiter.Enter(f.fallbackUrl)

	req, err := http.NewRequest("GET", rewrite.String(), nil)
	if err != nil {
		f.limiter.Leave(f.fallbackUrl)
		return nil, err
	}

	resp, err := f.proxy.Do(req)
	f.limiter.Leave(f.fallbackUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("return code is not 200 OK, code: %v, body: %v", resp.StatusCode, string(body))
	}

	return nil, nil
}

func (f *graphiteWeb) Description() map[string]*types.FunctionDescription {
	return f.supportedFunctions
}
