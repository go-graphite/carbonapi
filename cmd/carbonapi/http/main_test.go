package http

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/go-graphite/carbonapi/expr/types"
	zipperTypes "github.com/go-graphite/carbonapi/zipper/types"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/lomik/zapwriter"
	"github.com/stretchr/testify/assert"
)

type mockCarbonZipper struct{}

func newMockCarbonZipper() *mockCarbonZipper {
	return new(mockCarbonZipper)
}

func (z mockCarbonZipper) Find(ctx context.Context, request pb.MultiGlobRequest) (*pb.MultiGlobResponse, *zipperTypes.Stats, merry.Error) {
	return getGlobResponse(), nil, nil
}

func (z mockCarbonZipper) Info(ctx context.Context, metrics []string) (*pb.ZipperInfoResponse, *zipperTypes.Stats, merry.Error) {
	response := getMockInfoResponse()

	return response, nil, nil
}

func (z mockCarbonZipper) Render(ctx context.Context, request pb.MultiFetchRequest) ([]*types.MetricData, *zipperTypes.Stats, merry.Error) {
	return z.RenderCompat(ctx, []string{""}, 0, 0)
}

func (z mockCarbonZipper) RenderCompat(ctx context.Context, metrics []string, from, until int64) ([]*types.MetricData, *zipperTypes.Stats, merry.Error) {
	var result []*types.MetricData
	multiFetchResponse := getMultiFetchResponse()
	result = append(result, &types.MetricData{FetchResponse: multiFetchResponse.Metrics[0]})
	return result, nil, nil
}

func (z mockCarbonZipper) TagNames(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return []string{}, nil
}

func (z mockCarbonZipper) TagValues(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return []string{}, nil
}

func (z mockCarbonZipper) ScaleToCommonStep() bool {
	return true
}

func getGlobResponse() *pb.MultiGlobResponse {
	globMtach := pb.GlobMatch{Path: "foo.bar", IsLeaf: true}
	var matches []pb.GlobMatch
	matches = append(matches, globMtach)
	globResponse := &pb.MultiGlobResponse{
		Metrics: []pb.GlobResponse{
			{
				Name:    "foo.bar",
				Matches: matches,
			},
		},
	}
	return globResponse
}

func getMultiFetchResponse() pb.MultiFetchResponse {
	mfr := pb.FetchResponse{
		Name:           "foo.bar",
		PathExpression: "foo.bar",
		StartTime:      1510913280,
		StopTime:       1510913880,
		StepTime:       60,
		Values:         []float64{math.NaN(), 1510913759, 1510913818},
	}

	result := pb.MultiFetchResponse{Metrics: []pb.FetchResponse{mfr}}
	return result
}

func getMockInfoResponse() *pb.ZipperInfoResponse {
	r := pb.Retention{
		SecondsPerPoint: 60,
		NumberOfPoints:  43200,
	}
	decoded := &pb.ZipperInfoResponse{
		Info: map[string]pb.MultiMetricsInfoResponse{
			"http://127.0.0.1:8080": {
				Metrics: []pb.MetricsInfoResponse{{
					Name:              "foo.bar",
					ConsolidationFunc: "average",
					MaxRetention:      157680000,
					XFilesFactor:      0.5,
					Retentions:        []pb.Retention{r},
				}},
			},
		},
	}

	return decoded
}

func init() {
	cfg := config.DefaultLoggerConfig
	cfg.Level = "debug"
	zapwriter.ApplyConfig([]zapwriter.Config{cfg})
	logger := zapwriter.Logger("main")

	cfgFile := ""
	config.SetUpViper(logger, &cfgFile, "CARBONAPI_")
	config.Config.Upstreams.Backends = []string{"dummy"}
	config.SetUpConfigUpstreams(logger)
	config.SetUpConfig(logger, "(test)")
	config.Config.SetZipper(newMockCarbonZipper())
	emptyStringList := make([]string, 0)
	InitHandlers(emptyStringList, emptyStringList)
}

func setUpRequest(t *testing.T, url string) (*http.Request, *httptest.ResponseRecorder) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	return req, rr
}

func TestRenderHandler(t *testing.T) {
	req, rr := setUpRequest(t, "/render/?target=fallbackSeries(foo.bar,foo.baz)&from=-10minutes&format=json")
	renderHandler(rr, req)

	expected := `[{"target":"foo.bar","datapoints":[[null,1510913280],[1510913759,1510913340],[1510913818,1510913400]],"tags":{}}]`

	// Check the status code is what we expect.
	r := assert.Equal(t, rr.Code, http.StatusOK, "HttpStatusCode should be 200 OK.")
	if !r {
		t.Error("HttpStatusCode should be 200 OK.")
	}
	r = assert.Equal(t, expected, rr.Body.String(), "Http response should be same.")
	if !r {
		t.Error("Http response should be same.")
	}
}

func TestFindHandler(t *testing.T) {
	req, rr := setUpRequest(t, "/metrics/find/?query=foo.bar&format=json")
	findHandler(rr, req)

	body := rr.Body.String()
	expected := `[{"allowChildren":0,"expandable":0,"leaf":1,"id":"foo.bar","text":"bar","context":{}}]` + "\n"
	r := assert.Equal(t, rr.Code, http.StatusOK, "HttpStatusCode should be 200 OK.")
	if !r {
		t.Error("HttpStatusCode should be 200 OK.")
	}
	r = assert.Equal(t, expected, body, "Http response should be same.")
	if !r {
		t.Error("Http response should be same.")
	}
}

func TestInfoHandler(t *testing.T) {
	req, rr := setUpRequest(t, "/info/?target=foo.bar&format=json")
	infoHandler(rr, req)

	body := rr.Body.String()
	expected := getMockInfoResponse()
	expectedJson, err := json.Marshal(expected)
	r := assert.Nil(t, err)
	if !r {
		t.Errorf("err should be nil, %v instead", err)
	}

	r = assert.Equal(t, rr.Code, http.StatusOK, "HttpStatusCode should be 200 OK.")
	if !r {
		t.Error("Http response should be same.")
	}
	r = assert.Equal(t, string(expectedJson), body, "Http response should be same.")
	if !r {
		t.Error("Http response should be same.")
	}
}
