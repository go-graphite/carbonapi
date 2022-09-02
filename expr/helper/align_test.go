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
			"Metrics with one step",
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 3, 5, 7, 9, 11, 13, 15, 17}, 2, 4), // 4..26
				types.MakeMetricData("metric2", []float64{1, 5, 7, 9, 11, 13, 18}, 2, 5),        // 5..26
				types.MakeMetricData("metric3", []float64{1, 5, 7, 9, 11, 13}, 2, 6),            // 4..22
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 3, 5, 7, 9, 11, 13, 15, 17}, 2, 4),     // 4..26
				types.MakeMetricData("metric2", []float64{1, 5, 7, 9, 11, 13, 18, NaN}, 2, 4),       // 4..26
				types.MakeMetricData("metric3", []float64{NaN, 1, 5, 7, 9, 11, 13, NaN, NaN}, 2, 4), // 4..26
			},
		},
		{
			"Normal metrics",
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 3, 5, 7, 9, 11, 13, 15, 17}, 1, 4), // 4..13
				types.MakeMetricData("metric2", []float64{1, 5, 7, 9, 11, 13, 18}, 1, 5),        // 5..13
				types.MakeMetricData("metric3", []float64{0, 1, 2, 3, 4, 5}, 2, 2),              // 2..14
				types.MakeMetricData("metric4", []float64{1, 2, 3, 4, 5, 6}, 3, 3),              // 3..21
				types.MakeMetricData("metric5", []float64{1, 2, 3, 4}, 4, 4),                    // 4..20
				types.MakeMetricData("metric6", []float64{1, 2, 3, 4}, 4, 2),                    // 2..18
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{NaN, NaN, 1, 3, 5, 7, 9, 11, 13, 15, 17, NaN, NaN, NaN, NaN, NaN, NaN, NaN, NaN}, 1, 2),    // 2..21
				types.MakeMetricData("metric2", []float64{NaN, NaN, NaN, 1, 5, 7, 9, 11, 13, 18, NaN, NaN, NaN, NaN, NaN, NaN, NaN, NaN, NaN}, 1, 2), // 2..21
				types.MakeMetricData("metric3", []float64{0, 1, 2, 3, 4, 5, NaN, NaN, NaN}, 2, 2),                                                    // 2..20
				types.MakeMetricData("metric4", []float64{1, 2, 3, 4, 5, 6}, 3, 2),                                                                   // 2..20
				types.MakeMetricData("metric5", []float64{1, 2, 3, 4}, 4, 2),                                                                         // 2..18
				types.MakeMetricData("metric6", []float64{1, 2, 3, 4}, 4, 2),                                                                         // 2..18
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
				types.MakeMetricData("metric2", []float64{1, 3, 5, NaN}, 1, 1),
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

func BenchmarkAlignSeries(b *testing.B) {
	metrics := []*types.MetricData{
		types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 4),
		types.MakeMetricData("metric2", compare.GenerateMetrics(1023, 1.0, 9.0, 1.0), 1, 4),
		types.MakeMetricData("metric3", compare.GenerateMetrics(512, 1.0, 9.0, 1.0), 2, 2),
		types.MakeMetricData("metric4", compare.GenerateMetrics(340, 1.0, 9.0, 1.0), 3, 3),
		types.MakeMetricData("metric5", compare.GenerateMetrics(256, 1.0, 9.0, 1.0), 4, 4),
		types.MakeMetricData("metric6", compare.GenerateMetrics(510, 1.0, 9.0, 1.0), 4, 2),
	}

	for n := 0; n < b.N; n++ {
		got := AlignSeries(metrics)
		_ = got

	}
}

func TestScaleSeries(t *testing.T) {
	NaN := math.NaN()
	tests := []struct {
		name    string
		metrics []*types.MetricData
		want    []*types.MetricData
	}{
		{
			"Metrics with one step (avg)",
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 3, 5, 7, 9, 11, 13, 15, 17}, 2, 4), // 4..26
				types.MakeMetricData("metric2", []float64{1, 5, 7, 9, 11, 13, 18}, 2, 5),        // 5..26
				types.MakeMetricData("metric3", []float64{1, 5, 7, 9, 11, 13}, 2, 6),            // 4..22
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 3, 5, 7, 9, 11, 13, 15, 17}, 2, 4),     // 4..26
				types.MakeMetricData("metric2", []float64{1, 5, 7, 9, 11, 13, 18, NaN, NaN}, 2, 4),  // 4..26
				types.MakeMetricData("metric3", []float64{NaN, 1, 5, 7, 9, 11, 13, NaN, NaN}, 2, 4), // 4..26
			},
		},

		// Indx     |  0   |   1  |   2  |   3  |   4  |   5  |   6  |   7  |   8  |   9  |   10  |   11  |   12  |   13  |   14  |   15  |   20  |   21  |
		// commonStep  12
		// Start  0 (2 - 2 % 12)
		//  metric1 |      |      |      |      |   1  |   3  |   5  |   7  |   9  |  11  |   13  |   15  |   17  |       |       |       |       |       |
		//  metric1 |  N   |      |      |      |      |      |      |      |      |      |       |       |   17  |       |       |       |       |       |
		//
		//  metric2 |      |      |      |      |      |   1  |   5  |   7  |   9  |  11  |   13  |   18  |       |       |       |       |       |       |
		//  metric2 | 9.142857142857143  |      |      |      |      |      |      |      |       |       |  NaN  |       |       |       |       |       |
		//
		//  metric3 |      |      |   0  |      |  1   |      |   2  |      |   3  |      |    4  |       |    5  |       |       |       |       |       |
		//  metric3 |   2  |      |      |      |      |      |      |      |      |      |       |       |    5  |       |       |       |       |       |
		//
		//  metric4 |      |      |      |   1  |      |      |   2  |      |      |   3  |       |       |    4  |       |    5  |       |       |    6  |
		//  metric4 |   2  |      |      |      |      |      |      |      |      |      |       |       |    5  |       |       |       |       |      |
		//
		//  metric5 |      |      |      |      |  1   |      |      |      |   2  |      |       |       |    3  |       |       |       |    4  |       |
		//  metric5 | 1.5  |      |      |      |      |      |      |      |      |      |       |       |  3.5  |       |       |       |       |       |
		//
		//  metric6 |      |      |   1  |      |      |      |   2  |      |      |      |    3  |       |       |       |    4  |       |       |       |
		//  metric6 |   2  |      |      |      |      |      |      |      |      |      |       |       |    4  |       |       |       |       |       |
		{
			"Normal metrics (avg)",
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 3, 5, 7, 9, 11, 13, 15, 17}, 1, 4),
				types.MakeMetricData("metric2", []float64{1, 5, 7, 9, 11, 13, 18}, 1, 5),
				types.MakeMetricData("metric3", []float64{0, 1, 2, 3, 4, 5}, 2, 2),
				types.MakeMetricData("metric4", []float64{1, 2, 3, 4, 5, 6}, 3, 3),
				types.MakeMetricData("metric5", []float64{1, 2, 3, 4}, 4, 4),
				types.MakeMetricData("metric6", []float64{1, 2, 3, 4}, 4, 2),
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{8, 17}, 12, 0),
				types.MakeMetricData("metric2", []float64{9.142857142857143, NaN}, 12, 0),
				types.MakeMetricData("metric3", []float64{2, 5}, 12, 0),
				types.MakeMetricData("metric4", []float64{2, 5}, 12, 0),
				types.MakeMetricData("metric5", []float64{1.5, 3.5}, 12, 0),
				types.MakeMetricData("metric6", []float64{2, 4}, 12, 0),
			},
		},
		{
			"Broken StopTime (avg)",
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 5, 7, 8}, 1, 1).AppendStopTime(-1),
				types.MakeMetricData("metric2", []float64{1, 3, 5}, 1, 1),
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 5, 7, 8}, 1, 1),
				types.MakeMetricData("metric2", []float64{1, 3, 5, NaN}, 1, 1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScaleSeries(tt.metrics)
			compare.TestMetricDataEqLen(t, got, tt.want)
		})
	}
}

func BenchmarkScaleSeries(b *testing.B) {
	metrics := []*types.MetricData{
		types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 4),
		types.MakeMetricData("metric2", compare.GenerateMetrics(1023, 1.0, 9.0, 1.0), 1, 4),
		types.MakeMetricData("metric3", compare.GenerateMetrics(512, 1.0, 9.0, 1.0), 2, 2),
		types.MakeMetricData("metric4", compare.GenerateMetrics(340, 1.0, 9.0, 1.0), 3, 3),
		types.MakeMetricData("metric5", compare.GenerateMetrics(256, 1.0, 9.0, 1.0), 4, 4),
		types.MakeMetricData("metric6", compare.GenerateMetrics(510, 1.0, 9.0, 1.0), 4, 2),
	}

	for n := 0; n < b.N; n++ {
		got := ScaleSeries(metrics)
		_ = got

	}
}
