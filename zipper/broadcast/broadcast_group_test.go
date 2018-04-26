package broadcast

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/go-graphite/carbonzipper/zipper/dummy"
	"github.com/go-graphite/carbonzipper/zipper/errors"
	"github.com/go-graphite/carbonzipper/zipper/types"

	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

var logger *zap.Logger
var timeouts types.Timeouts

func init() {
	defaultLoggerConfig := zapwriter.Config{
		Logger:           "",
		File:             "stdout",
		Level:            "debug",
		Encoding:         "json",
		EncodingTime:     "iso8601",
		EncodingDuration: "seconds",
	}

	zapwriter.ApplyConfig([]zapwriter.Config{defaultLoggerConfig})

	logger = zapwriter.Logger("test")
	timeouts = types.Timeouts{
		1000 * time.Second,
		1000 * time.Second,
		1000 * time.Second,
	}
}

func errorTextsEqual(e1, e2 []error) bool {
	e1s := make([]string, 0, len(e1))
	e2s := make([]string, 0, len(e2))
	for _, e := range e1 {
		e1s = append(e1s, e.Error())
	}

	for _, e := range e2 {
		e2s = append(e2s, e.Error())
	}
	sort.Strings(e1s)
	sort.Strings(e2s)
	for i := range e1s {
		if e1s[i] != e2s[i] {
			return false
		}
	}
	return true
}

func errorsAreEqual(e1, e2 *errors.Errors) bool {
	if e1 == nil && e2 != nil {
		return false
	}

	if e1 != nil && e2 == nil {
		return false
	}

	if e1 != nil && e2 != nil {
		if e1.HaveFatalErrors != e2.HaveFatalErrors || len(e1.Errors) != len(e2.Errors) || !errorTextsEqual(e1.Errors, e2.Errors) {
			return false
		}
	}
	return true
}

type testCaseNew struct {
	name        string
	servers     []types.ServerClient
	expectedErr *errors.Errors
}

func TestNewBroadcastGroup(t *testing.T) {
	tests := []testCaseNew{
		{
			name:        "no servers",
			expectedErr: errors.Fatal("no servers specified"),
		},
		{
			name: "some servers",
			servers: []types.ServerClient{
				dummy.NewDummyClient("client1", []string{"backend1", "backend2"}, 0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewBroadcastGroup(logger, tt.name, tt.servers, 60, 500, timeouts)
			if !errorsAreEqual(err, tt.expectedErr) {
				t.Fatalf("unexpected error %v, expected %v", err, tt.expectedErr)
			}
			_ = b
		})
	}
}

type testCaseProbe struct {
	name            string
	servers         []types.ServerClient
	clientResponses map[string]dummy.ProbeResponse
	response        []string
	expectedErr     *errors.Errors
}

func TestProbeTLDs(t *testing.T) {
	tests := []testCaseProbe{
		{
			name: "two clients different data",
			servers: []types.ServerClient{
				dummy.NewDummyClient("client1", []string{"backend1", "backend2"}, 1),
				dummy.NewDummyClient("client2", []string{"backend3", "backend4"}, 1),
			},
			clientResponses: map[string]dummy.ProbeResponse{
				"client1": dummy.ProbeResponse{
					Response: []string{"a", "b", "c"},
				},
				"client2": dummy.ProbeResponse{
					Response: []string{"a", "d", "e"},
				},
			},
			response:    []string{"a", "b", "c", "d", "e"},
			expectedErr: &errors.Errors{},
		},
	}

	for _, tt := range tests {
		b, err := NewBroadcastGroup(logger, tt.name, tt.servers, 60, 500, timeouts)
		if err != nil && (err.HaveFatalErrors || len(err.Errors) > 0) {
			t.Fatalf("error while initializing group, when it shouldn't be: %v", err)
		}

		for i := range tt.servers {
			name := fmt.Sprintf("client%v", i+1)
			s := tt.servers[i].(*dummy.DummyClient)
			s.SetTLDResponse(tt.clientResponses[name])
		}

		ctx := context.Background()

		t.Run(tt.name, func(t *testing.T) {
			res, err := b.ProbeTLDs(ctx)
			if !errorsAreEqual(err, tt.expectedErr) {
				t.Fatalf("unexpected error %v, expected %v", err, tt.expectedErr)
			}

			if len(res) != len(tt.response) {
				t.Fatalf("different amount of responses %v, expected %v", res, tt.response)
			}

			sort.Strings(res)
			sort.Strings(tt.response)
			for i := range res {
				if res[i] != tt.response[i] {
					t.Errorf("got %v, expected %v", res[i], tt.response[i])
				}
			}
		})
	}
}

type testCaseFetch struct {
	name           string
	servers        []types.ServerClient
	fetchRequest   map[string][]*protov3.MultiFetchRequest
	fetchResponses map[string][]dummy.FetchResponse

	expectedErr      *errors.Errors
	expectedResponse *protov3.MultiFetchResponse
}

func mergeFetchRequests(requests []*protov3.MultiFetchRequest) *protov3.MultiFetchRequest {
	var res protov3.MultiFetchRequest

	for _, r := range requests {
		res.Metrics = append(res.Metrics, r.Metrics...)
	}

	return &res
}

func TestFetchRequests(t *testing.T) {
	tests := []testCaseFetch{
		{
			name: "two clients different data",
			servers: []types.ServerClient{
				dummy.NewDummyClient("client1", []string{"backend1", "backend2"}, 1),
				dummy.NewDummyClient("client2", []string{"backend1", "backend2"}, 1),
			},
			fetchRequest: map[string][]*protov3.MultiFetchRequest{
				"client1": []*protov3.MultiFetchRequest{
					{
						Metrics: []protov3.FetchRequest{
							{
								Name:           "foo",
								StartTime:      0,
								StopTime:       120,
								PathExpression: "foo",
							},
						},
					},
				},
				"client2": []*protov3.MultiFetchRequest{
					{
						Metrics: []protov3.FetchRequest{
							{
								Name:           "foo2",
								StartTime:      0,
								StopTime:       120,
								PathExpression: "foo2",
							},
						},
					},
				},
			},
			fetchResponses: map[string][]dummy.FetchResponse{
				"client1": []dummy.FetchResponse{
					{
						Response: &protov3.MultiFetchResponse{
							Metrics: []protov3.FetchResponse{
								{
									Name:              "foo",
									PathExpression:    "foo",
									ConsolidationFunc: "avg",
									StartTime:         0,
									StopTime:          120,
									StepTime:          60,
									XFilesFactor:      0.5,
									Values:            []float64{0, 1, 2},
								},
							},
						},
						Stats:  &types.Stats{},
						Errors: &errors.Errors{},
					},
				},
				"client2": []dummy.FetchResponse{
					{
						Response: &protov3.MultiFetchResponse{
							Metrics: []protov3.FetchResponse{
								{
									Name:              "foo2",
									PathExpression:    "foo2",
									ConsolidationFunc: "avg",
									StartTime:         0,
									StopTime:          120,
									StepTime:          60,
									XFilesFactor:      0.5,
									Values:            []float64{0, 1, 2},
								},
							},
						},
						Stats:  &types.Stats{},
						Errors: &errors.Errors{},
					},
				},
			},

			expectedResponse: &protov3.MultiFetchResponse{
				Metrics: []protov3.FetchResponse{
					{
						Name:              "foo",
						PathExpression:    "foo",
						ConsolidationFunc: "avg",
						StartTime:         0,
						StopTime:          120,
						StepTime:          60,
						XFilesFactor:      0.5,
						Values:            []float64{0, 1, 2},
					},
					{
						Name:              "foo2",
						PathExpression:    "foo2",
						ConsolidationFunc: "avg",
						StartTime:         0,
						StopTime:          120,
						StepTime:          60,
						XFilesFactor:      0.5,
						Values:            []float64{0, 1, 2},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		b, err := NewBroadcastGroup(logger, tt.name, tt.servers, 60, 500, timeouts)
		if err != nil && (err.HaveFatalErrors || len(err.Errors) > 0) {
			t.Fatalf("error while initializing group, when it shouldn't be: %v", err)
		}

		for i := range tt.servers {
			name := fmt.Sprintf("client%v", i+1)
			s := tt.servers[i].(*dummy.DummyClient)
			for i, r := range tt.fetchRequest[name] {
				resp := tt.fetchResponses[name][i]
				s.AddFetchResponse(r, resp.Response, resp.Stats, resp.Errors)
			}
		}

		ctx := context.Background()

		t.Run(tt.name, func(t *testing.T) {
			var requests []*protov3.MultiFetchRequest
			for _, r := range tt.fetchRequest {
				requests = append(requests, r...)
			}
			request := mergeFetchRequests(requests)
			res, _, err := b.Fetch(ctx, request)
			if tt.expectedErr == nil || !tt.expectedErr.HaveFatalErrors {
				if err != nil && err.HaveFatalErrors {
					t.Errorf("unexpected error %v, expected %v", err, tt.expectedErr)
				}
			} else {
				if !errorsAreEqual(err, tt.expectedErr) {
					t.Errorf("unexpected error %v, expected %v", err, tt.expectedErr)
				}
			}

			if len(res.Metrics) != len(tt.expectedResponse.Metrics) {
				t.Fatalf("different amount of responses %v, expected %v", res, tt.expectedResponse)
			}

			sort.Slice(res.Metrics, func(i, j int) bool {
				return res.Metrics[i].Name < res.Metrics[j].Name
			})
			sort.Slice(tt.expectedResponse.Metrics, func(i, j int) bool {
				return tt.expectedResponse.Metrics[i].Name < tt.expectedResponse.Metrics[j].Name
			})
			if !reflect.DeepEqual(res, tt.expectedResponse) {
				t.Errorf("got %v, expected %v", res, tt.expectedResponse)
			}
		})
	}
}
