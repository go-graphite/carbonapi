package metric

import "testing"

func TestExtractMetric(t *testing.T) {
	var tests = []struct {
		input  string
		metric string
	}{
		{
			"f",
			"f",
		},
		{
			"func(f)",
			"f",
		},
		{
			"foo.bar.baz",
			"foo.bar.baz",
		},
		{
			"nonNegativeDerivative(foo.bar.baz)",
			"foo.bar.baz",
		},
		{
			"movingAverage(foo.bar.baz,10)",
			"foo.bar.baz",
		},
		{
			"scale(scaleToSeconds(nonNegativeDerivative(foo.bar.baz),60),60)",
			"foo.bar.baz",
		},
		{
			"divideSeries(foo.bar.baz,baz.qux.zot)",
			"foo.bar.baz",
		},
		{
			"{something}",
			"{something}",
		},
		{
			"ab=",
			"ab=",
		},
		{
			"ab=.c",
			"ab=.c",
		},
		{
			"ab==",
			"ab==",
		},
		{
			"ab==.c",
			"ab==.c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if m := ExtractMetric(tt.input); m != tt.metric {
				t.Errorf("extractMetric(%q)=%q, want %q", tt.input, m, tt.metric)
			}
		})
	}
}
