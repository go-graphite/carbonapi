package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	merry2 "github.com/ansel1/merry"
	"go.uber.org/zap"
)

type TestSchema struct {
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
	ErrBody         string           `yaml:"errBody"`
	ErrSort         bool             `yaml:"errSort"`
	ExpectedResults []ExpectedResult `yaml:"expectedResults"`
}

type ExpectedResult struct {
	SHA256            []string `yaml:"sha256"`
	Metrics           []RenderResponse
	MetricsFind       []MetricsFindResponse `json:"metricsFind" yaml:"metricsFind"`
	TagsAutocompelete []string              `json:"tagsAutocompelete" yaml:"tagsAutocompelete"`
}

type MetricsFindResponse struct {
	AllowChildren int               `json:"allowChildren" yaml:"allowChildren"`
	Expandable    int               `json:"expandable" yaml:"expandable"`
	Leaf          int               `json:"leaf" yaml:"leaf"`
	Id            string            `json:"id" yaml:"id"`
	Text          string            `json:"text" yaml:"text"`
	Context       map[string]string `json:"context" yaml:"context"`
}

type RenderResponse struct {
	Target     string            `json:"target" yaml:"target"`
	Datapoints []Datapoint       `json:"datapoints" yaml:"datapoints"`
	Tags       map[string]string `json:"tags" yaml:"tags"`
}

type Datapoint struct {
	Timestamp int
	Value     float64
}

func (d *Datapoint) UnmarshalJSON(data []byte) error {
	pieces := strings.Split(string(data), ",")
	if len(pieces) != 2 {
		return fmt.Errorf("too many parameters in the Datapoint, got %v, expected 2", len(pieces))
	}

	var err error
	valueStr := pieces[0][1:]
	tsStr := pieces[1][:len(pieces[1])-1]

	d.Timestamp, err = strconv.Atoi(tsStr)
	if err != nil {
		return fmt.Errorf("failed to parse Timestamp: %v", err)
	}

	if valueStr == "null" || valueStr == "\"null\"" {
		d.Value = math.NaN()
		return nil
	}
	d.Value, err = strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf("failed to parse Value: %v", err)
	}

	return nil
}

func (d *Datapoint) UnmarshalYAML(unmarshal func(interface{}) error) error {
	yamlData := make([]string, 0)
	err := unmarshal(&yamlData)
	if err != nil {
		return err
	}

	if len(yamlData) != 2 {
		return fmt.Errorf("too many parameters in the Datapoint, got %v, expected 2", len(yamlData))
	}

	valueStr := yamlData[0]
	tsStr := yamlData[1]

	d.Timestamp, err = strconv.Atoi(tsStr)
	if err != nil {
		return fmt.Errorf("failed to parse Timestamp: %v", err)
	}

	if valueStr == "null" || valueStr == "\"null\"" {
		d.Value = math.NaN()
		return nil
	}
	d.Value, err = strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf("failed to parse Value: %v", err)
	}

	return nil
}

func isRenderEqual(m1, m2 RenderResponse) error {
	if m1.Target != m2.Target {
		return fmt.Errorf("target mismatch, got '%v', expected '%v'", m1.Target, m2.Target)
	}

	if len(m1.Datapoints) != len(m2.Datapoints) {
		return fmt.Errorf("response have unexpected length, got '%v', expected '%v'", m1.Datapoints, m2.Datapoints)
	}

	if len(m1.Datapoints) > 1 {
		step1 := m1.Datapoints[1].Timestamp - m1.Datapoints[2].Timestamp
		step2 := m2.Datapoints[1].Timestamp - m2.Datapoints[2].Timestamp
		if step1 != step2 {
			return fmt.Errorf("series has unexpected step, got '%v', expected '%v'", step1, step2)
		}
	}
	datapointsMismatch := false
	for i := range m1.Datapoints {
		if math.IsNaN(m1.Datapoints[i].Value) && math.IsNaN(m2.Datapoints[i].Value) {
			continue
		}
		if m1.Datapoints[i].Value != m2.Datapoints[i].Value {
			datapointsMismatch = true
			break
		}
		if m1.Datapoints[i].Timestamp != m2.Datapoints[i].Timestamp {
			datapointsMismatch = true
			break
		}
	}
	if datapointsMismatch {
		return fmt.Errorf("data in response is different, got '%v', expected '%v'", m1.Datapoints, m2.Datapoints)
	}

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func resortErr(errStr string) string {
	first := strings.Index(errStr, "\n")
	if first >= 0 && first != len(errStr)-1 {
		// resort error string
		errs := strings.Split(errStr, "\n")
		if errs[len(errs)-1] == "" {
			errs = errs[:len(errs)-1]
		}
		sort.Strings(errs)
		errStr = strings.Join(errs, "\n") + "\n"
	}
	return errStr
}

func doTest(logger *zap.Logger, t *Query, verbose bool) []error {
	client := http.Client{}
	failures := make([]error, 0)
	d, err := time.ParseDuration(fmt.Sprintf("%v", t.Delay) + "s")
	if err != nil {
		err = merry2.Prepend(err, "failed parse duration")
		failures = append(failures, err)
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
	u, err := url.Parse(t.Endpoint + t.URL)
	if err != nil {
		err = merry2.Prepend(err, "failed to parse URL")
		failures = append(failures, err)
		return failures
	}

	logger.Info("sending request",
		zap.String("endpoint", t.Endpoint),
		zap.String("original_URL", t.URL),
	)

	req, err := http.NewRequestWithContext(ctx, t.Type, t.Endpoint+u.Path+"/?"+u.Query().Encode(), body)
	if err != nil {
		err = merry2.Prepend(err, "failed to prepare the request")
		failures = append(failures, err)
		return failures
	}

	resp, err = client.Do(req)
	if err != nil {
		err = merry2.Prepend(err, "failed to perform the request")
		failures = append(failures, err)
		return failures
	}

	contentType = resp.Header.Get("Content-Type")
	if t.ExpectedResponse.ContentType != contentType {
		failures = append(failures,
			merry2.Errorf("unexpected content-type, got %v (code %d), expected %v",
				contentType, resp.StatusCode,
				t.ExpectedResponse.ContentType,
			),
		)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		err = merry2.Prepend(err, "failed to read body")
		failures = append(failures, err)
		return failures
	}

	if resp.StatusCode != t.ExpectedResponse.HttpCode {
		failures = append(failures, merry2.Errorf("unexpected status code, got %v, expected %v",
			resp.StatusCode,
			t.ExpectedResponse.HttpCode,
		),
		)
	}

	// We don't need to actually check body of response if we expect any sort of error (4xx/5xx), but for check error handling do this
	if t.ExpectedResponse.HttpCode >= 300 {
		if t.ExpectedResponse.ErrBody != "" {
			errStr := string(b)
			if t.ExpectedResponse.ErrSort {
				errStr = resortErr(errStr)
			}
			if t.ExpectedResponse.ErrBody != errStr {
				failures = append(failures, merry2.Errorf("mismatch error body, got '%s', expected '%s'", string(b), t.ExpectedResponse.ErrBody))
			}
		}
		return failures
	}

	switch contentType {
	case "image/png":
	case "image/svg+xml":
		hash := sha256.Sum256(b)
		hashStr := fmt.Sprintf("%x", hash)
		sha256matched := false
		for _, sha256sum := range t.ExpectedResponse.ExpectedResults[0].SHA256 {
			if hashStr == sha256sum {
				sha256matched = true
				break
			}
		}
		if !sha256matched {
			encodedBody := base64.StdEncoding.EncodeToString(b)
			failures = append(failures, merry2.Errorf("sha256 mismatch, got '%v', expected '%v', encodedBody: '%v'", hashStr, t.ExpectedResponse.ExpectedResults[0].SHA256, encodedBody))
			return failures
		}
	case "application/json":
		if strings.HasPrefix(t.URL, "/metrics/find") {
			res := make([]MetricsFindResponse, 0, 1)
			err := json.Unmarshal(b, &res)
			if err != nil {
				err = merry2.Prepend(err, "failed to parse response")
				failures = append(failures, err)
				return failures
			}

			if len(t.ExpectedResponse.ExpectedResults) == 0 {
				return failures
			}

			if len(res) != len(t.ExpectedResponse.ExpectedResults[0].MetricsFind) {
				failures = append(failures, merry2.Errorf("unexpected amount of metrics find, got %v, expected %v",
					len(res),
					len(t.ExpectedResponse.ExpectedResults[0].MetricsFind)))
				if verbose {
					length := max(len(t.ExpectedResponse.ExpectedResults[0].MetricsFind), len(res))
					for i := 0; i < length; i++ {
						if i >= len(res) {
							err = fmt.Errorf("metrics find[%d] want=`%+v`", i, t.ExpectedResponse.ExpectedResults[0].MetricsFind[i])
							failures = append(failures, err)
						} else if i >= len(t.ExpectedResponse.ExpectedResults[0].MetricsFind) {
							err = fmt.Errorf("metrics find[%d] got unexpected=`%+v`", i, res[i])
							failures = append(failures, err)
						} else if !reflect.DeepEqual(res[i], t.ExpectedResponse.ExpectedResults[0].MetricsFind[i]) {
							err = fmt.Errorf("metrics find[%d] are not equal, got=`%+v`, expected=`%+v`", i, res[i], t.ExpectedResponse.ExpectedResults[0].MetricsFind[i])
							failures = append(failures, err)
						}
					}
				}
				return failures
			}

			for i := range res {
				if !reflect.DeepEqual(res[i], t.ExpectedResponse.ExpectedResults[0].MetricsFind[i]) {
					err = fmt.Errorf("metrics find[%d] are not equal, got=`%+v`, expected=`%+v`", i, res[i], t.ExpectedResponse.ExpectedResults[0].MetricsFind[i])
					failures = append(failures, err)
				}
			}
		} else if strings.HasPrefix(t.URL, "/tags/autoComplete/") {
			// tags/autoComplete
			res := make([]string, 0, 1)
			err := json.Unmarshal(b, &res)
			if err != nil {
				err = merry2.Prepend(err, "failed to parse response")
				failures = append(failures, err)
				return failures
			}

			if len(t.ExpectedResponse.ExpectedResults) == 0 {
				return failures
			}

			if len(res) != len(t.ExpectedResponse.ExpectedResults[0].TagsAutocompelete) {
				failures = append(failures, merry2.Errorf("unexpected amount of results, got %v, expected %v",
					len(res),
					len(t.ExpectedResponse.ExpectedResults[0].TagsAutocompelete)))
				if verbose {
					length := max(len(t.ExpectedResponse.ExpectedResults[0].TagsAutocompelete), len(res))
					for i := 0; i < length; i++ {
						if i >= len(res) {
							err = fmt.Errorf("tags[%d] want=`%+v`", i, t.ExpectedResponse.ExpectedResults[0].TagsAutocompelete[i])
							failures = append(failures, err)
						} else if i >= len(t.ExpectedResponse.ExpectedResults[0].TagsAutocompelete) {
							err = fmt.Errorf("tags[%d] got unexpected=`%+v`", i, res[i])
							failures = append(failures, err)
						} else if !reflect.DeepEqual(res[i], t.ExpectedResponse.ExpectedResults[0].TagsAutocompelete[i]) {
							err = fmt.Errorf("tags[%d] are not equal, got=`%+v`, expected=`%+v`", i, res[i], t.ExpectedResponse.ExpectedResults[0].TagsAutocompelete[i])
							failures = append(failures, err)
						}
					}
				}
				return failures
			}

			for i := range res {
				if res[i] != t.ExpectedResponse.ExpectedResults[0].TagsAutocompelete[i] {
					err = merry2.Prependf(err, "tags[%d] are not equal, got=`%+v`, expected=`%+v`", i, res[i], t.ExpectedResponse.ExpectedResults[0].TagsAutocompelete[i])
					failures = append(failures, err)
				}
			}

		} else {
			// render
			res := make([]RenderResponse, 0, 1)
			err := json.Unmarshal(b, &res)
			if err != nil {
				err = merry2.Prepend(err, "failed to parse response")
				failures = append(failures, err)
				return failures
			}

			if len(t.ExpectedResponse.ExpectedResults) == 0 {
				return failures
			}

			if len(res) != len(t.ExpectedResponse.ExpectedResults[0].Metrics) {
				failures = append(failures, merry2.Errorf("unexpected amount of results, got %v, expected %v",
					len(res),
					len(t.ExpectedResponse.ExpectedResults[0].Metrics)))
				if verbose {
					length := max(len(t.ExpectedResponse.ExpectedResults[0].Metrics), len(res))
					for i := 0; i < length; i++ {
						if i >= len(res) {
							err = fmt.Errorf("metrics[%d] want=`%+v`", i, t.ExpectedResponse.ExpectedResults[0].Metrics[i])
							failures = append(failures, err)
						} else if i >= len(t.ExpectedResponse.ExpectedResults[0].Metrics) {
							err = fmt.Errorf("metrics[%d] got unexpected=`%+v`", i, res[i])
							failures = append(failures, err)
						} else if !reflect.DeepEqual(res[i], t.ExpectedResponse.ExpectedResults[0].Metrics[i]) {
							err = fmt.Errorf("metrics[%d] are not equal, got=`%+v`, expected=`%+v`", i, res[i], t.ExpectedResponse.ExpectedResults[0].Metrics[i])
							failures = append(failures, err)
						}
					}
				}
				return failures
			}

			for i := range res {
				err := isRenderEqual(res[i], t.ExpectedResponse.ExpectedResults[0].Metrics[i])
				if err != nil {
					err = merry2.Prependf(err, "metrics are not equal, got=`%+v`, expected=`%+v`", res[i], t.ExpectedResponse.ExpectedResults[0].Metrics[i])
					failures = append(failures, err)
				}
			}
		}
	default:
		if resp.StatusCode == http.StatusOK {
			// if !strings.HasPrefix(t.URL, "/tags/autoComplete/") ||
			// 	(contentType == "text/plain; charset=utf-8" &&
			// 		resp.StatusCode == http.StatusNotFound &&
			// 		t.ExpectedResponse.HttpCode == http.StatusNotFound) {
			failures = append(failures, merry2.Errorf("unsupported content-type: got '%v'", contentType))
		}
	}

	return failures
}

func e2eTest(logger *zap.Logger, noapp, breakOnError, verbose bool) bool {
	failed := false
	logger.Info("will run test",
		zap.Any("config", cfg.Test),
	)
	runningApps := make(map[string]*runner)
	if !noapp {
		wgStart := sync.WaitGroup{}
		for i, c := range cfg.Test.Apps {
			r := new(&cfg.Test.Apps[i], logger)
			wgStart.Add(1)
			runningApps[c.Name] = r
			go func() {
				wgStart.Done()
				r.Run()
			}()
		}

		wgStart.Wait()
		logger.Info("will sleep for 1 seconds to start all required apps")
		time.Sleep(1 * time.Second)
	}

	for _, t := range cfg.Test.Queries {
		failures := doTest(logger, &t, verbose)
		if len(failures) != 0 {
			failed = true
			logger.Error("test failed",
				zap.Errors("failures", failures),
				zap.String("url", t.URL), zap.String("type", t.Type), zap.String("body", t.Body),
			)
			for _, v := range runningApps {
				if !v.IsRunning() {
					logger.Error("unexpected app crash", zap.Any("app", v))
				}
			}
			if breakOnError {
				for {
					fmt.Print("Some queries was failed, press y for continue after debug test:")
					in := bufio.NewScanner(os.Stdin)
					in.Scan()
					s := in.Text()
					if s == "y" || s == "Y" {
						break
					}
				}
			}
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
		for _, v := range runningApps {
			logger.Info("app out", zap.Any("app", v), zap.String("out", v.Out()))
		}
	} else {
		logger.Info("All tests OK")
	}

	return failed
}
