package types

import (
	"bytes"
	"math"
	"math/rand"
	"testing"
)

func TestJSONResponse(t *testing.T) {

	tests := []struct {
		results []*MetricData
		out     []byte
	}{
		{
			[]*MetricData{
				MakeMetricData("metric1", []float64{1, 1.5, 2.25, math.NaN()}, 100, 100),
				MakeMetricData("metric2;foo=bar", []float64{2, 2.5, 3.25, 4, 5}, 100, 100),
			},
			[]byte(`[{"target":"metric1","datapoints":[[1,100],[1.5,200],[2.25,300],[null,400]],"tags":{"name":"metric1"}},{"target":"metric2;foo=bar","datapoints":[[2,100],[2.5,200],[3.25,300],[4,400],[5,500]],"tags":{"foo":"bar","name":"metric2"}}]`),
		},
	}

	for _, tt := range tests {
		b := MarshalJSON(tt.results, 1.0, false)
		if !bytes.Equal(b, tt.out) {
			t.Errorf("marshalJSON(%+v): got\n%+v\nwant\n%+v", tt.results, string(b), string(tt.out))
		}
	}
}

func TestJSONResponseNoNullPoints(t *testing.T) {

	tests := []struct {
		results []*MetricData
		out     []byte
	}{
		{
			[]*MetricData{
				MakeMetricData("metric1", []float64{1, 1.5, 2.25, math.NaN()}, 100, 100),
				MakeMetricData("metric2;foo=bar", []float64{math.NaN(), 2.5, 3.25, 4, 5}, 100, 100),
				MakeMetricData("metric3;foo=bar", []float64{2, math.NaN(), 3.25, 4, 5}, 100, 100),
				MakeMetricData("metric4;foo=bar", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 100, 100),
			},
			[]byte(`[{"target":"metric1","datapoints":[[1,100],[1.5,200],[2.25,300]],"tags":{"name":"metric1"}},{"target":"metric2;foo=bar","datapoints":[[2.5,200],[3.25,300],[4,400],[5,500]],"tags":{"foo":"bar","name":"metric2"}},{"target":"metric3;foo=bar","datapoints":[[2,100],[3.25,300],[4,400],[5,500]],"tags":{"foo":"bar","name":"metric3"}},{"target":"metric4;foo=bar","datapoints":[],"tags":{"foo":"bar","name":"metric4"}}]`),
		},
	}

	for _, tt := range tests {
		b := MarshalJSON(tt.results, 1.0, true)
		if !bytes.Equal(b, tt.out) {
			t.Errorf("marshalJSON(%+v): got\n%+v\nwant\n%+v", tt.results, string(b), string(tt.out))
		}
	}
}

func TestRawResponse(t *testing.T) {

	tests := []struct {
		results []*MetricData
		out     []byte
	}{
		{
			[]*MetricData{
				MakeMetricData("metric1", []float64{1, 1.5, 2.25, math.NaN()}, 100, 100),
				MakeMetricData("metric2", []float64{2, 2.5, 3.25, 4, 5}, 100, 100),
			},
			[]byte(`metric1,100,500,100|1,1.5,2.25,None` + "\n" + `metric2,100,600,100|2,2.5,3.25,4,5` + "\n"),
		},
	}

	for _, tt := range tests {
		b := MarshalRaw(tt.results)
		if !bytes.Equal(b, tt.out) {
			t.Errorf("marshalRaw(%+v): got\n%+v\nwant\n%+v", tt.results, string(b), string(tt.out))
		}
	}
}

func TestCSVResponse(t *testing.T) {

	tests := []struct {
		results []*MetricData
		out     []byte
	}{
		{
			[]*MetricData{
				MakeMetricData("metric1", []float64{1, 1.5, 2.25, math.NaN()}, 100, 100),
				MakeMetricData("metric2", []float64{2, 2.5, 3.25, 4, 5}, 100, 100),
			},
			[]byte(`"metric1",1970-01-01 00:01:40,1` + "\n" +
				`"metric1",1970-01-01 00:03:20,1.5` + "\n" +
				`"metric1",1970-01-01 00:05:00,2.25` + "\n" +
				`"metric1",1970-01-01 00:06:40,` + "\n" +
				`"metric2",1970-01-01 00:01:40,2` + "\n" +
				`"metric2",1970-01-01 00:03:20,2.5` + "\n" +
				`"metric2",1970-01-01 00:05:00,3.25` + "\n" +
				`"metric2",1970-01-01 00:06:40,4` + "\n" +
				`"metric2",1970-01-01 00:08:20,5` + "\n",
			),
		},
	}

	for _, tt := range tests {
		b := MarshalCSV(tt.results)
		if !bytes.Equal(b, tt.out) {
			t.Errorf("marshalCSV(%+v): \n%+v\nwant\n%+v", tt.results, string(b), string(tt.out))
		}
	}
}

func getData(rangeSize int) []float64 {
	var data = make([]float64, rangeSize)
	var r = rand.New(rand.NewSource(99))
	for i := range data {
		data[i] = math.Floor(1000 * r.Float64())
	}

	return data
}

func BenchmarkMarshalJSON(b *testing.B) {
	data := []*MetricData{
		MakeMetricData("metric1", getData(10000), 100, 100),
		MakeMetricData("metric2", getData(100000), 100, 100),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MarshalJSON(data, 1.0, false)
	}
}

func BenchmarkMarshalJSONLong(b *testing.B) {
	data := []*MetricData{
		MakeMetricData("metric1", getData(10000), 100, 100),
		MakeMetricData("metric2", getData(100000), 100, 100),
		MakeMetricData("metric3", getData(100000), 100, 100),
		MakeMetricData("metric4", getData(100000), 100, 100),
		MakeMetricData("metric5", getData(100000), 100, 100),
		MakeMetricData("metric6", getData(100000), 100, 100),
		MakeMetricData("metric7", getData(100000), 100, 100),
		MakeMetricData("metric8", getData(100000), 100, 100),
		MakeMetricData("metric9", getData(100000), 100, 100),
		MakeMetricData("metric10", getData(100000), 100, 100),
		MakeMetricData("metric11", getData(10000), 100, 100),
		MakeMetricData("metric12", getData(100000), 100, 100),
		MakeMetricData("metric13", getData(100000), 100, 100),
		MakeMetricData("metric14", getData(100000), 100, 100),
		MakeMetricData("metric15", getData(100000), 100, 100),
		MakeMetricData("metric16", getData(100000), 100, 100),
		MakeMetricData("metric17", getData(100000), 100, 100),
		MakeMetricData("metric18", getData(100000), 100, 100),
		MakeMetricData("metric19", getData(100000), 100, 100),
		MakeMetricData("metric20", getData(100000), 100, 100),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MarshalJSON(data, 1.0, false)
	}
}

func BenchmarkMarshalRaw(b *testing.B) {
	data := []*MetricData{
		MakeMetricData("metric1", getData(10000), 100, 100),
		MakeMetricData("metric2", getData(100000), 100, 100),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MarshalRaw(data)
	}
}

func BenchmarkMarshalRawLong(b *testing.B) {
	data := []*MetricData{
		MakeMetricData("metric1", getData(10000), 100, 100),
		MakeMetricData("metric2", getData(100000), 100, 100),
		MakeMetricData("metric3", getData(100000), 100, 100),
		MakeMetricData("metric4", getData(100000), 100, 100),
		MakeMetricData("metric5", getData(100000), 100, 100),
		MakeMetricData("metric6", getData(100000), 100, 100),
		MakeMetricData("metric7", getData(100000), 100, 100),
		MakeMetricData("metric8", getData(100000), 100, 100),
		MakeMetricData("metric9", getData(100000), 100, 100),
		MakeMetricData("metric10", getData(100000), 100, 100),
		MakeMetricData("metric11", getData(10000), 100, 100),
		MakeMetricData("metric12", getData(100000), 100, 100),
		MakeMetricData("metric13", getData(100000), 100, 100),
		MakeMetricData("metric14", getData(100000), 100, 100),
		MakeMetricData("metric15", getData(100000), 100, 100),
		MakeMetricData("metric16", getData(100000), 100, 100),
		MakeMetricData("metric17", getData(100000), 100, 100),
		MakeMetricData("metric18", getData(100000), 100, 100),
		MakeMetricData("metric19", getData(100000), 100, 100),
		MakeMetricData("metric20", getData(100000), 100, 100),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MarshalRaw(data)
	}
}

func BenchmarkMarshalCSV(b *testing.B) {
	data := []*MetricData{
		MakeMetricData("metric1", getData(10000), 100, 100),
		MakeMetricData("metric2", getData(100000), 100, 100),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MarshalCSV(data)
	}
}

func BenchmarkMarshalCSVLong(b *testing.B) {
	data := []*MetricData{
		MakeMetricData("metric1", getData(10000), 100, 100),
		MakeMetricData("metric2", getData(100000), 100, 100),
		MakeMetricData("metric3", getData(100000), 100, 100),
		MakeMetricData("metric4", getData(100000), 100, 100),
		MakeMetricData("metric5", getData(100000), 100, 100),
		MakeMetricData("metric6", getData(100000), 100, 100),
		MakeMetricData("metric7", getData(100000), 100, 100),
		MakeMetricData("metric8", getData(100000), 100, 100),
		MakeMetricData("metric9", getData(100000), 100, 100),
		MakeMetricData("metric10", getData(100000), 100, 100),
		MakeMetricData("metric11", getData(10000), 100, 100),
		MakeMetricData("metric12", getData(100000), 100, 100),
		MakeMetricData("metric13", getData(100000), 100, 100),
		MakeMetricData("metric14", getData(100000), 100, 100),
		MakeMetricData("metric15", getData(100000), 100, 100),
		MakeMetricData("metric16", getData(100000), 100, 100),
		MakeMetricData("metric17", getData(100000), 100, 100),
		MakeMetricData("metric18", getData(100000), 100, 100),
		MakeMetricData("metric19", getData(100000), 100, 100),
		MakeMetricData("metric20", getData(100000), 100, 100),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MarshalCSV(data)
	}
}
