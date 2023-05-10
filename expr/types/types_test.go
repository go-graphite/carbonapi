package types

import (
	"testing"

	"github.com/go-graphite/carbonapi/expr/types/config"
	"github.com/stretchr/testify/assert"
)

func TestAggregatedValuesNudgedAndHighestTimestamp(t *testing.T) {

	config.Config.NudgeStartTimeOnAggregation = true
	config.Config.UseBucketsHighestTimestampOnAggregation = true

	tests := []struct {
		name      string
		values    []float64
		step      int64
		start     int64
		mdp       int64
		want      []float64
		wantStep  int64
		wantStart int64
	}{
		{
			name:      "empty",
			values:    []float64{},
			step:      60,
			mdp:       100,
			want:      []float64{},
			wantStep:  60,
			wantStart: 0,
		},
		{
			name:      "one point",
			values:    []float64{1, 2, 3, 4},
			start:     10,
			step:      10,
			mdp:       1,
			want:      []float64{10},
			wantStep:  40,
			wantStart: 40,
		},
		{
			name:      "no nudge if few points",
			values:    []float64{1, 2, 3, 4},
			step:      10,
			start:     20,
			mdp:       1,
			want:      []float64{10},
			wantStep:  40,
			wantStart: 50,
		},

		{
			name:      "should nudge the first point",
			values:    []float64{1, 2, 3, 4, 5, 6},
			start:     20,
			step:      10,
			mdp:       3,
			want:      []float64{5, 9, 6},
			wantStep:  20,
			wantStart: 40,
		},
		{
			name:      "should be stable with previous",
			values:    []float64{2, 3, 4, 5, 6, 7},
			start:     30,
			step:      10,
			mdp:       3,
			want:      []float64{5, 9, 13},
			wantStep:  20,
			wantStart: 40,
		},
		{
			name:      "more data",
			values:    []float64{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14},
			start:     20,
			step:      10,
			mdp:       3,
			want:      []float64{40, 50},
			wantStep:  50,
			wantStart: 100,
		},
		{
			name:      "even more data",
			values:    []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10.0, 11, 12, 13, 14},
			start:     10,
			step:      10,
			mdp:       3,
			want:      []float64{15, 40, 50},
			wantStep:  50,
			wantStart: 50,
		},
		{
			name:      "skewed start time",
			values:    []float64{2, 3, 4, 5, 6, 7, 8, 9, 10},
			start:     21,
			step:      10,
			mdp:       5,
			want:      []float64{2 + 3, 4 + 5, 6 + 7, 8 + 9, 10}, // no points discarded, bucket starts at 20
			wantStep:  20,
			wantStart: 31,
		},
		{
			name:      "skewed start time 2",
			values:    []float64{2, 3, 4, 5, 6, 7, 8, 9, 10},
			start:     29,
			step:      10,
			mdp:       5,
			want:      []float64{2 + 3, 4 + 5, 6 + 7, 8 + 9, 10}, // no points discarded, bucket starts at 20
			wantStep:  20,
			wantStart: 39,
		},
		{
			name:      "skewed start time 3",
			values:    []float64{2, 3, 4, 5, 6, 7, 8, 9, 10},
			start:     31,
			step:      10,
			mdp:       5,
			want:      []float64{3 + 4, 5 + 6, 7 + 8, 9 + 10}, // 1st point discarded, it belongs to the incomplete bucket (20,40)
			wantStep:  20,
			wantStart: 51,
		},
		{
			name:      "skewed start time no aggregation",
			values:    []float64{1, 2, 3, 4},
			start:     31,
			step:      10,
			mdp:       4,
			want:      []float64{1, 2, 3, 4},
			wantStep:  10,
			wantStart: 31,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := MakeMetricData("test", tt.values, tt.step, tt.start)
			input.ConsolidationFunc = "sum"
			ConsolidateJSON(tt.mdp, []*MetricData{input})

			got := input.AggregatedValues()
			gotStep := input.AggregatedTimeStep()
			gotStart := input.AggregatedStartTime()

			assert.Equal(t, tt.want, got, "bad values")
			assert.Equal(t, tt.wantStep, gotStep, "bad step")
			assert.Equal(t, tt.wantStart, gotStart, "bad start")
		})
	}
}

func TestAggregatedValuesConfigVariants(t *testing.T) {
	const start = 20
	const step = 10
	const mdp = 3
	values := []float64{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	const expectedStep = int64(50)
	/*
		ts:                |    | 20 | 30 | 40 | 50 | 60 | 70 | 80 | 90 | 100 | 110 | 120 | 130 | 140 |
		vals:              |    | 2  | 3  | 4  | 5  | 6  | 7  | 8  | 9  | 10  | 11  | 12  | 13  | 14  |
		unaligned buckets:      |                        |                          |
		aligned buckets:   |                        |                         |
	*/

	tests := []struct {
		name             string
		nudge            bool
		highestTimestamp bool
		want             []float64
		wantStart        int64
	}{
		{
			name:             "nudge start and highest timestamp",
			nudge:            true,
			highestTimestamp: true,
			want:             []float64{40, 50},
			wantStart:        100,
		},
		{
			name:             "nudge start and not highest timestamp",
			nudge:            true,
			highestTimestamp: false,
			want:             []float64{40, 50},
			wantStart:        60,
		},
		{
			name:             "not nudge start and highest timestamp",
			nudge:            false,
			highestTimestamp: true,
			want:             []float64{20, 45, 39},
			wantStart:        60,
		},
		{
			name:             "not nudge start and not highest timestamp",
			nudge:            false,
			highestTimestamp: false,
			want:             []float64{20, 45, 39},
			wantStart:        20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Config.NudgeStartTimeOnAggregation = tt.nudge
			config.Config.UseBucketsHighestTimestampOnAggregation = tt.highestTimestamp

			input := MakeMetricData("test", values, step, start)
			input.ConsolidationFunc = "sum"
			ConsolidateJSON(mdp, []*MetricData{input})

			got := input.AggregatedValues()
			gotStep := input.AggregatedTimeStep()
			gotStart := input.AggregatedStartTime()

			assert.Equal(t, tt.want, got, "bad values")
			assert.Equal(t, expectedStep, gotStep, "bad step")
			assert.Equal(t, tt.wantStart, gotStart, "bad start")
		})
	}
}
