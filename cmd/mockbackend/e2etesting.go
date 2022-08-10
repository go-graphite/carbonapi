package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	ExpectedResults []ExpectedResult `yaml:"expectedResults"`
}

type ExpectedResult struct {
	SHA256  []string `yaml:"sha256"`
	Metrics []CarbonAPIResponse
}

type CarbonAPIResponse struct {
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

func isMetricsEqual(m1, m2 CarbonAPIResponse) error {
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

func doTest(logger *zap.Logger, t *Query) []error {
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

	if resp.StatusCode != t.ExpectedResponse.HttpCode {
		failures = append(failures, merry2.Errorf("unexpected status code, got %v, expected %v",
			resp.StatusCode,
			t.ExpectedResponse.HttpCode,
		),
		)
	}

	contentType = resp.Header.Get("Content-Type")
	if t.ExpectedResponse.ContentType != contentType {
		failures = append(failures,
			merry2.Errorf("unexpected content-type, got %v, expected %v",
				contentType,
				t.ExpectedResponse.ContentType,
			),
		)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = merry2.Prepend(err, "failed to read body")
		failures = append(failures, err)
		return failures
	}

	// We don't need to actually check body of response if we expect any sort of error (4xx/5xx)
	if t.ExpectedResponse.HttpCode >= 300 {
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
			failures = append(failures, merry2.Errorf("sha256 mismatch, got '%v', expected '%v', encodedBodyy: '%v'", hashStr, t.ExpectedResponse.ExpectedResults[0].SHA256, encodedBody))
			return failures
		}
	case "application/json":
		res := make([]CarbonAPIResponse, 0, 1)
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
			return failures
		}

		for i := range res {
			err := isMetricsEqual(res[i], t.ExpectedResponse.ExpectedResults[0].Metrics[i])
			if err != nil {
				err = merry2.Prependf(err, "metrics are not equal, got=`%+v`, expected=`%+v`", res[i], t.ExpectedResponse.ExpectedResults[0].Metrics[i])
				failures = append(failures, err)
			}
		}

	default:
		failures = append(failures, merry2.Errorf("unsupported content-type: got '%v'", contentType))
	}

	return failures
}

func e2eTest(logger *zap.Logger, noapp bool) bool {
	failed := false
	logger.Info("will run test",
		zap.Any("config", cfg.Test),
	)
	runningApps := make(map[string]*runner)
	if !noapp {
		for i, c := range cfg.Test.Apps {
			r := new(&cfg.Test.Apps[i], logger)
			runningApps[c.Name] = r
			go r.Run()
		}

		logger.Info("will sleep for 5 seconds to start all required apps")
		time.Sleep(5 * time.Second)
	}

	for _, t := range cfg.Test.Queries {
		failures := doTest(logger, &t)

		if len(failures) != 0 {
			failed = true
			logger.Error("test failed",
				zap.Errors("failures", failures),
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

	return failed
}
