package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-graphite/carbonapi/expr"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"github.com/go-graphite/carbonapi/util"
)

var errNoMetrics = errors.New("no metrics")

type unmarshaler interface {
	Unmarshal([]byte) error
}

type zipper struct {
	z      string
	client *http.Client
}

func (z zipper) Find(ctx context.Context, metric string) (pb.GlobResponse, error) {
	u, _ := url.Parse(z.z + "/metrics/find/")

	u.RawQuery = url.Values{
		"query":  []string{metric},
		"format": []string{"protobuf3"},
	}.Encode()

	var pbresp pb.GlobResponse

	err := z.get(ctx, "Find", u, &pbresp)

	return pbresp, err
}

func (z zipper) get(ctx context.Context, who string, u *url.URL, msg unmarshaler) error {
	request, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return fmt.Errorf("http.NewRequest: %+v", err)
	}

	request = util.MarshalCtx(ctx, request)

	resp, err := z.client.Do(request.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("http.Get: %+v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll: %+v", err)
	}

	err = msg.Unmarshal(body)
	if err != nil {
		return fmt.Errorf("proto.Unmarshal: %+v", err)
	}

	return nil
}

func (z zipper) Passthrough(ctx context.Context, metric string) ([]byte, error) {

	u, _ := url.Parse(z.z + metric)

	request, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %+v", err)
	}
	request = util.MarshalCtx(ctx, request)

	resp, err := z.client.Do(request.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("http.Get: %+v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll: %+v", err)
	}

	return body, nil
}

func (z zipper) Render(ctx context.Context, metric string, from, until int32) ([]*expr.MetricData, error) {
	var result []*expr.MetricData

	u, _ := url.Parse(z.z + "/render/")

	u.RawQuery = url.Values{
		"target": []string{metric},
		"format": []string{"protobuf3"},
		"from":   []string{strconv.Itoa(int(from))},
		"until":  []string{strconv.Itoa(int(until))},
	}.Encode()

	var pbresp pb.MultiFetchResponse
	err := z.get(ctx, "Render", u, &pbresp)
	if err != nil {
		return result, err
	}

	if m := pbresp.Metrics; len(m) == 0 {
		return result, errNoMetrics
	}

	for i := range pbresp.Metrics {
		result = append(result, &expr.MetricData{FetchResponse: *pbresp.Metrics[i]})
	}

	return result, nil
}
