package types

import "testing"

func TestExtractName(t *testing.T) {
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
			"scale(scaleToSeconds(nonNegativeDerivative(ab==.c),60),60)",
			"ab==.c",
		},
		{
			"divideSeries(metric[12])",
			"metric[12]",
		},
		{
			"average(metric{1,2}e,'sum')",
			"metric{1,2}e",
		},
		{
			"aliasByNode(alias(0.1.2.@.4, 2), 1)",
			"0.1.2.@.4",
		},
		{
			"aliasByTags(alias(0.1.2.@.4, 2), 1)",
			"0.1.2.@.4",
		},
		// non-ASCII symbols
		{
			"alias(Количество изменений)",
			"Количество изменений",
		},
		{
			"some(Количество изменений, Аргумент)",
			"Количество изменений",
		},
		{
			"seriesByTag('tag2=value*', 'name=metric')",
			"seriesByTag('tag2=value*', 'name=metric')",
		},
		{
			"sum(seriesByTag('tag2=value*', 'name=metric'))",
			"seriesByTag('tag2=value*', 'name=metric')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if m := ExtractName(tt.input); m != tt.metric {
				t.Errorf("extractMetric(%q)=%q, want %q", tt.input, m, tt.metric)
			}
		})
	}
}
