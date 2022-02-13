package helpers

import (
	"math"
	"testing"
	"time"

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

func TestAdjustStep(t *testing.T) {
	type args struct {
		start                int64
		stop                 int64
		maxPointsPerQuery    int64
		minStep              int64
		forceMinStepInterval time.Duration
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			"(interval/step) points is less than maxPointsPerQuery (not force min_step)",
			args{
				0,
				60 * 60 * 24, // interval =  24h
				60 * 24,      // maxDataPoints =  60 points for 1 hour x 24 hours
				60,           // step = 1min
				0,            // don't force min_step
			},
			60,
		},
		{
			"(interval/step) points is more than maxPointsPerQuery (not force min_step)",
			args{
				0,
				60 * 60 * 25, // interval = 25h
				60 * 24,      // maxDataPoints = 60 points for 1 hour x 24 hours
				60,           // step = 1min
				0,            // don't force min_step
			},
			120,
		},
		{
			"(interval/step) points is much more than maxPointsPerQuery (not force min_step)",
			args{
				0,
				60 * 60 * 50, // interval = 50h
				60 * 24,      // maxDataPoints = 60 points for 1 hour x 24 hours
				60,           // step = 1min
				0,            // don't force min_step
			},
			300,
		},
		{
			"force min_step for more than interval while maxPointsPerQuery is less than (interval/step) points",
			args{
				0,
				60 * 20,          // interval = 20min
				15,               // maxDataPoints = 15 points
				60,               // step = 1min
				30 * time.Minute, // do(!) force min_step for 30min
			},
			60,
		},
		{
			"force min_step for less than interval while maxPointsPerQuery is less than (interval/step) points",
			args{
				0,
				60 * 20,          // interval = 20min
				15,               // maxDataPoints = 15 points
				60,               // step = 1min
				10 * time.Minute, // do(!) force min_step for 10min
			},
			120,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AdjustStep(tt.args.start, tt.args.stop, tt.args.maxPointsPerQuery, tt.args.minStep, tt.args.forceMinStepInterval); got != tt.want {
				t.Errorf("AdjustStep() = %v, want %v", got, tt.want)
			}
		})
	}
}
