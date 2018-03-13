package graphiteWeb

import (
	"encoding/json"
	"fmt"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"github.com/go-graphite/carbonzipper/limiter"
	"github.com/lomik/zapwriter"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"
)

type graphiteWeb struct {
	interfaces.FunctionBase

	working      bool
	strict       bool
	maxTries     int
	fallbackUrls []string
	proxy        *http.Client

	supportedFunctions map[string]types.FunctionDescription
	limiter            limiter.ServerLimiter

	logger         *zap.Logger
	requestCounter uint64
}

func (f *graphiteWeb) pickServer() string {
	sid := atomic.AddUint64(&f.requestCounter, 1)
	return f.fallbackUrls[sid%uint64(len(f.fallbackUrls))]
}

func GetOrder() interfaces.Order {
	return interfaces.Last
}

type graphiteWebConfig struct {
	Enabled                  bool
	FallbackUrls             []string
	Strict                   bool
	MaxConcurrentConnections int
	MaxTries                 int
	Timeout                  time.Duration
	KeepAliveInterval        time.Duration
	ForceKeep                []string
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
	logger := zapwriter.Logger("functionInit").With(zap.String("function", "graphiteWeb"))
	if configFile == "" {
		logger.Error("no config file specified",
			zap.Error(fmt.Errorf("config is required for this function")),
		)
		return nil
	}
	v := viper.New()
	v.SetConfigFile(configFile)
	err := v.ReadInConfig()
	if err != nil {
		logger.Fatal("failed to read config file",
			zap.Error(err),
		)
	}

	cfg := graphiteWebConfig{
		Enabled: false,
		Strict:  false,
		MaxConcurrentConnections: 10,
		Timeout:                  60 * time.Second,
		KeepAliveInterval:        30 * time.Second,
		MaxTries:                 3,
	}
	err = v.Unmarshal(&cfg)
	if err != nil {
		logger.Fatal("failed to parse config",
			zap.Error(err),
		)
	}

	if !cfg.Enabled {
		logger.Warn("graphiteWeb config found but graphiteWeb proxy is disabled")
		return nil
	}

	logger.Info("graphiteWeb configured",
		zap.Any("config", cfg),
		zap.String("config_file", configFile),
	)

	f := &graphiteWeb{
		limiter: limiter.NewServerLimiter(cfg.FallbackUrls, cfg.MaxConcurrentConnections),
		proxy: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: cfg.MaxConcurrentConnections,
				DialContext: (&net.Dialer{
					Timeout:   cfg.Timeout,
					KeepAlive: cfg.KeepAliveInterval,
					DualStack: true,
				}).DialContext,
			},
		},
		fallbackUrls: cfg.FallbackUrls,
		strict:       cfg.Strict,
		maxTries:     cfg.MaxTries,
		working:      false,
		logger:       zapwriter.Logger("graphiteWeb"),
		supportedFunctions: map[string]types.FunctionDescription{
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

	ok := false
	var body []byte
	for i := 0; i < len(f.fallbackUrls); i++ {
		srv := f.fallbackUrls[i]
		req, err := http.NewRequest("GET", srv+"/functions/?format=json", nil)
		if err != nil {
			logger.Warn("failed to create list of functions, will try next fallbackUrl",
				zap.String("backend", srv),
				zap.Error(err),
			)
			continue
		}

		resp, err := f.proxy.Do(req)
		if err != nil {
			logger.Warn("failed to obtain list of functions, will try next fallbackUrl",
				zap.String("backend", srv),
				zap.Error(err),
			)
			continue
		}

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Warn("failed to obtain list of functions, will try next fallbackUrl",
				zap.String("backend", srv),
				zap.Error(fmt.Errorf("return code is not 200 OK")),
				zap.Int("status_code", resp.StatusCode),
			)
			resp.Body.Close()
			continue
		}

		if resp.StatusCode != http.StatusOK {
			logger.Warn("failed to obtain list of functions, will try next fallbackUrl",
				zap.String("backend", srv),
				zap.Error(fmt.Errorf("return code is not 200 OK")),
				zap.Int("status_code", resp.StatusCode),
				zap.String("body", string(body)),
			)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()
		ok = true
		break
	}

	if !ok {
		logger.Error("failed to initialize graphiteWeb fallback function",
			zap.Error(fmt.Errorf("no more backends to try, see warnings above for more details")),
		)
		return nil
	}

	forceAdd := make(map[string]struct{})
	for _, n := range cfg.ForceAdd {
		forceAdd[n] = struct{}{}
	}

	forceKeep := make(map[string]struct{})
	for _, n := range cfg.ForceKeep {
		forceKeep[n] = struct{}{}
	}

	graphiteWebSupportedFunctions := make(map[string]types.FunctionDescription)

	err = json.Unmarshal(body, &graphiteWebSupportedFunctions)
	if err != nil {
		logger.Error("failed to parse list of functions",
			zap.Error(err),
		)
		return nil
	}

	functions := []string{"graphiteWeb"}
	metadata.FunctionMD.RLock()
	for k, v := range graphiteWebSupportedFunctions {
		var ok bool
		if _, ok = forceKeep[k]; ok {
			continue
		}

		if _, ok = forceAdd[k]; ok {
			functions = append(functions, k)
			f.supportedFunctions[k] = v
			continue
		}

		if v2, ok := metadata.FunctionMD.Descriptions[k]; ok {
			if f.strict {
				ok = paramsIsEqual(v.Params, v2.Params)
			}
			if ok {
				continue
			}
		}

		functions = append(functions, k)
		f.supportedFunctions[k] = v
	}
	metadata.FunctionMD.RUnlock()

	f.working = true

	logger.Info("will handle following functions",
		zap.Strings("functions_metadata", functions),
	)

	res := make([]interfaces.FunctionMetadata, 0, len(functions))
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f, Order: interfaces.Any})
	}
	return res
}

type target string

func (t *target) UnmarshalJSON(d []byte) error {
	var res interface{}
	err := json.Unmarshal(d, &res)
	if err != nil {
		return err
	}
	switch v := res.(type) {
	case int:
		*t = target(strconv.FormatInt(int64(v), 10))
	case int32:
		*t = target(strconv.FormatInt(int64(v), 10))
	case int64:
		*t = target(strconv.FormatInt(v, 10))
	case float64:
		*t = target(strconv.FormatFloat(v, 'f', -1, 64))
	case string:
		*t = target(v)
	case bool:
		*t = target(strconv.FormatBool(v))
	default:
		return fmt.Errorf("unsupported type for target")
	}

	return nil
}

type graphiteMetric struct {
	Tags       map[string]json.RawMessage
	Target     target
	Datapoints [][2]float64
}

type graphiteError struct {
	server string
	err    error
}

func (f *graphiteWeb) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	f.logger.Info("received request",
		zap.Bool("working", f.working),
	)
	if !f.working {
		return nil, nil
	}

	var target string
	if e.Target() == "graphiteWeb" {
		target = e.RawArgs()
	} else {
		target = e.ToString()
	}

	var body []byte
	var srv string
	var request string
	var errors []graphiteError
	ok := false
	for i := 0; i < f.maxTries; i++ {
		srv = f.pickServer()
		rewrite, _ := url.Parse(srv + "/render/")
		v := url.Values{
			"target": []string{target},
			"from":   []string{strconv.FormatInt(int64(from), 10)},
			"until":  []string{strconv.FormatInt(int64(until), 10)},
			"format": []string{"json"},
		}

		rewrite.RawQuery = v.Encode()

		f.limiter.Enter(srv)

		req, err := http.NewRequest("GET", rewrite.String(), nil)
		if err != nil {
			f.limiter.Leave(srv)
			return nil, err
		}

		resp, err := f.proxy.Do(req)
		f.limiter.Leave(srv)
		if err != nil {
			errors = append(errors, graphiteError{srv, err})
			resp.Body.Close()
			continue
		}

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			errors = append(errors, graphiteError{srv, err})
			resp.Body.Close()
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			err := fmt.Errorf("return code is not 200 OK, code: %v, body: %v", resp.StatusCode, string(body))
			errors = append(errors, graphiteError{srv, err})
			continue
		}
		resp.Body.Close()
		ok = true
		request = rewrite.String()
		break
	}

	if !ok {
		f.logger.Error("failed to get response from graphite-web, max tries exceeded",
			zap.Any("errors", errors),
		)
		return nil, fmt.Errorf("max tries exceeded for request target=%v", target)
	}

	f.logger.Debug("got response",
		zap.String("request", request),
		zap.String("body", string(body)),
	)

	var tmp []graphiteMetric

	err := json.Unmarshal(body, &tmp)
	if err != nil {
		return nil, err
	}

	res := make([]*types.MetricData, len(tmp))

	for _, m := range tmp {
		stepTime := int32(60)
		if len(m.Datapoints) > 1 {
			stepTime = int32(m.Datapoints[1][0] - m.Datapoints[0][0])
		}
		pbResp := pb.FetchResponse{
			Name:      string(m.Target),
			StartTime: int32(m.Datapoints[0][0]),
			StopTime:  int32(m.Datapoints[len(m.Datapoints)-1][0]),
			StepTime:  stepTime,
			Values:    make([]float64, len(m.Datapoints)),
			IsAbsent:  make([]bool, len(m.Datapoints)),
		}
		for i, v := range m.Datapoints {
			if math.IsNaN(v[1]) {
				pbResp.Values[i] = 0
				pbResp.IsAbsent[i] = true
			} else {
				pbResp.Values[i] = v[1]
				pbResp.IsAbsent[i] = false
			}
		}
		res = append(res, &types.MetricData{
			FetchResponse: pbResp,
		})
	}

	return res, nil
}

func (f *graphiteWeb) Description() map[string]types.FunctionDescription {
	return f.supportedFunctions
}
