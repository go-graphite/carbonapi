package helpers

import (
	"math"
	"testing"

	th "github.com/go-graphite/carbonapi/tests"
	"github.com/go-graphite/carbonapi/zipper/protocols/prometheus/types"
)

func TestAlignValues(t *testing.T) {
	type args struct {
		startTime  int64
		stopTime   int64
		step       int64
		promValues []types.Value
	}
	tests := []struct {
		name string
		args args
		want []float64
	}{
		{
			"single value, don't miss it",
			args{
				startTime: 1617736200,
				stopTime:  1617736200,
				step:      60,
				promValues: []types.Value{
					{
						Timestamp: 1617736200,
						Value:     12,
					},
				},
			},
			[]float64{12},
		},
		{
			"multiple values, don't miss the last one",
			args{
				startTime: 1617729600,
				stopTime:  1617730200,
				step:      60,
				promValues: []types.Value{
					{
						Timestamp: 1617729600,
						Value:     823,
					},
					{
						Timestamp: 1617730200,
						Value:     743,
					},
				},
			},
			[]float64{823, math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 743},
		},
		{
			"multiple values, step is not aligned", // last value has timestamp which is not aligned according to step
			args{
				startTime: 1617729600,
				stopTime:  1617729660,
				step:      11,
				promValues: []types.Value{
					{
						Timestamp: 1617729600,
						Value:     823,
					},
					{
						Timestamp: 1617729660,
						Value:     743,
					},
				},
			},
			[]float64{823, math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AlignValues(tt.args.startTime, tt.args.stopTime, tt.args.step, tt.args.promValues); !th.NearlyEqual(got, tt.want) {
				t.Errorf("AlignValues() = %v, want %v", got, tt.want)
			}
		})
	}
}
