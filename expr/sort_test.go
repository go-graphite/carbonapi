package expr

import (
	"testing"

	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func TestSortMetrics(t *testing.T) {
	const (
		gold   = "a.gold.c.d"
		silver = "a.silver.c.d"
		bronze = "a.bronze.c.d"
		first  = "a.first.c.d"
		second = "a.second.c.d"
		third  = "a.third.c.d"
		fourth = "a.fourth.c.d"
	)
	tests := []struct {
		metrics []*types.MetricData
		mfetch  parser.MetricRequest
		sorted  []*types.MetricData
	}{
		{
			[]*types.MetricData{
				//NOTE(nnuss): keep these lines lexically sorted ;)
				types.MakeMetricData(bronze, []float64{}, 1, 0),
				types.MakeMetricData(first, []float64{}, 1, 0),
				types.MakeMetricData(fourth, []float64{}, 1, 0),
				types.MakeMetricData(gold, []float64{}, 1, 0),
				types.MakeMetricData(second, []float64{}, 1, 0),
				types.MakeMetricData(silver, []float64{}, 1, 0),
				types.MakeMetricData(third, []float64{}, 1, 0),
			},
			parser.MetricRequest{
				Metric: "a.{first,second,third,fourth}.c.d",
				From:   0,
				Until:  1,
			},
			[]*types.MetricData{
				//These are in the brace appearance order
				types.MakeMetricData(first, []float64{}, 1, 0),
				types.MakeMetricData(second, []float64{}, 1, 0),
				types.MakeMetricData(third, []float64{}, 1, 0),
				types.MakeMetricData(fourth, []float64{}, 1, 0),

				//These are in the slice order as above and come after
				types.MakeMetricData(bronze, []float64{}, 1, 0),
				types.MakeMetricData(gold, []float64{}, 1, 0),
				types.MakeMetricData(silver, []float64{}, 1, 0),
			},
		},
	}
	for i, test := range tests {
		if len(test.metrics) != len(test.sorted) {
			t.Skipf("Error in test %d : length mismatch %d vs. %d", i, len(test.metrics), len(test.sorted))
		}
		SortMetrics(test.metrics, test.mfetch)
		for i := range test.metrics {
			if test.metrics[i].Name != test.sorted[i].Name {
				t.Errorf("[%d] Expected %q but have %q", i, test.sorted[i].Name, test.metrics[i].Name)
			}
		}
	}
}
