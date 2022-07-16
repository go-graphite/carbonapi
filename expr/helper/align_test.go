package helper

import (
	"math"
	"testing"

	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/tests/compare"
)

func TestAlignSeries(t *testing.T) {
	NaN := math.NaN()
	tests := []struct {
		name    string
		metrics []*types.MetricData
		want    []*types.MetricData
	}{
		{
			"Normal metrics",
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 3, 5, 7, 9, 11, 13, 15, 17}, 1, 4), // 4..13
				types.MakeMetricData("metric2", []float64{1, 5, 7, 9, 11, 13, 15, 18}, 1, 5),    // 5..13
				types.MakeMetricData("metric3", []float64{0, 1, 2, 3, 4, 5}, 2, 2),              // 2..14
				types.MakeMetricData("metric4", []float64{1, 2, 3, 4, 5, 6}, 3, 3),              // 3..21
				types.MakeMetricData("metric5", []float64{1, 2, 3, 4}, 4, 4),                    // 4..20
				types.MakeMetricData("metric6", []float64{1, 2, 3, 4}, 4, 2),                    // 2..18
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{NaN, NaN, 1, 3, 5, 7, 9, 11, 13, 15, 17, NaN, NaN, NaN, NaN, NaN, NaN, NaN, NaN}, 1, 2),   // 2..21
				types.MakeMetricData("metric2", []float64{NaN, NaN, NaN, 1, 5, 7, 9, 11, 13, 15, 18, NaN, NaN, NaN, NaN, NaN, NaN, NaN, NaN}, 1, 2), // 2..21
				types.MakeMetricData("metric3", []float64{0, 1, 2, 3, 4, 5, NaN, NaN, NaN}, 2, 2),                                                   // 2:21
				types.MakeMetricData("metric4", []float64{1, 2, 3, 4, 5, 6}, 3, 3),                                                                  // 2:21
				types.MakeMetricData("metric5", []float64{1, 2, 3, 4}, 4, 4),                                                                        // 4..18
				types.MakeMetricData("metric6", []float64{1, 2, 3, 4}, 4, 2),                                                                        // 2..18
			},
		},
		{
			"Broken StopTime",
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 5, 7, 8}, 1, 1).AppendStopTime(-1),
				types.MakeMetricData("metric2", []float64{1, 3, 5}, 1, 1),
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 5, 7, 8}, 1, 1),
				types.MakeMetricData("metric2", []float64{1, 3, 5}, 1, 1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AlignSeries(tt.metrics)
			compare.TestMetricData(t, got, tt.want)
		})
	}
}
