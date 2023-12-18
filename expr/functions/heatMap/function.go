package heatMap

import (
	"context"
	"math"

	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type heatMap struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{{
		F:    &heatMap{},
		Name: "heatMap",
	}}
}

func (f *heatMap) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	series, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	series = sortMetricData(series)
	seriesQty := len(series)
	result := make([]*types.MetricData, 0, seriesQty-1)

	if err = validateNeighbourSeries(series); err != nil {
		return nil, err
	}
	for i := 1; i < seriesQty; i++ {
		curr, prev := series[i], series[i-1]

		pointsQty := len(curr.Values)
		r := &types.MetricData{
			FetchResponse: pb.FetchResponse{
				Name:      "heatMap(" + curr.Name + "," + prev.Name + ")",
				Values:    make([]float64, pointsQty),
				StartTime: curr.StartTime,
				StopTime:  curr.StopTime,
				StepTime:  curr.StepTime,
			},
			Tags: curr.Tags,
		}

		for j := 0; j < pointsQty; j++ {
			if math.IsNaN(curr.Values[j]) || math.IsNaN(prev.Values[j]) {
				r.Values[j] = math.NaN()
				continue
			}
			r.Values[j] = curr.Values[j] - prev.Values[j]
		}

		result = append(result, r)
	}

	return result, nil
}

const description = `Compute heat-map like result based on a values of a metric.

All metrics are assigned weights, based on the sum of their first 5 values and then sorted based on that.

After that for the sorted metrics, diff with the previous one will be computed.

Assuming seriesList has values N series in total (sorted by sum of the first 5 values):
(a[1], a[2], ..., a[N]). Then heatMap will output N-1 series: (a[2] - a[1], a[3] - a[2], ..., a[N] - a[N-1]).

That function produce similar result to prometheus heatmaps for any list of incrementing counters and plays well with
grafana heatmap graph type.'
`

func (f *heatMap) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"heatMap": {
			Description: description,
			Function:    "heatMap(seriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "heatMap",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
