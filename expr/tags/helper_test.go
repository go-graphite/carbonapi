package tags

import (
	"testing"
)

type extractTagTestCase struct {
	TestName string
	Input    string
	Output   map[string]string
}

func TestSanitizeRegex(t *testing.T) {
	tests := []struct {
		Input  string
		Output string
	}{
		{
			Input:  "^value.*",
			Output: "value__*",
		},
		{
			Input:  "^value",
			Output: "value__*",
		},
		{
			Input:  "value$",
			Output: "__*value",
		},
		{
			Input:  "^value$",
			Output: "value",
		},
		{
			Input:  "value",
			Output: "__*value__*",
		},
		{
			Input:  "value.*a",
			Output: "__*value__*a__*",
		},
		{
			Input:  "^all__symbols__a-z____-___a_b____not_replaced$",
			Output: "all__symbols__a-z____-___a_b____not_replaced",
		},
		{
			Input:  `some_symbols_[a-z]+.?-.*{a,b}?_\.are_replaced(a|b)`,
			Output: `__*some_symbols_[a-z]+__?-__*{a,b}?_\.are_replaced(a|b)__*`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Input, func(t *testing.T) {
			out := sanitizeRegex(tt.Input)
			if tt.Output != out {
				t.Errorf("sanitizeRegex(%s) got '%s', want '%s'", tt.Input, out, tt.Output)
			}
		})
	}
}

func BenchmarkSanitizeRegex(b *testing.B) {
	benchmarks := []string{
		"_all__symbols__a-z____-___a_b___not_replaced_",
		"^some_symbols_[a-z]+.?-.*{a,b}?_are_replaced$",
	}
	for _, bm := range benchmarks {
		b.Run(bm, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				s := sanitizeRegex(bm)
				_ = s
			}
		})
	}
}

func TestExtractSeriesByTags(t *testing.T) {
	defaultName := "sumSeries"
	tests := []extractTagTestCase{
		// from aggregation functions with seriesByTag
		{
			TestName: `seriesByTag("tag2=value*", 'name=metric')`,
			Input:    `seriesByTag("tag2=value*", 'name=metric')`,
			Output:   map[string]string{"name": "metric", "tag2": "value*"},
		},
		{
			TestName: `seriesByTag('tag2=~(12|23)', "name=metric")`,
			Input:    `seriesByTag('tag2=~(12|23)', "name=metric")`,
			Output:   map[string]string{"name": "metric", "tag2": "__*(12|23)__*"},
		},
		{
			TestName: "seriesByTag('tag1=val1', 'tag2=val2', 'name=~(ab|dc|ef)', 'tag3=val3')",
			Input:    "seriesByTag('tag1=val1', 'tag2=val2', 'name=~(ab|dc|ef)', 'tag3=val3')",
			Output:   map[string]string{"name": "__*(ab|dc|ef)__*", "tag1": "val1", "tag2": "val2", "tag3": "val3"},
		},
		{
			TestName: "seriesByTag('tag2=~^value.*', 'name=metric')",
			Input:    "seriesByTag('tag2=~^value.*', 'name=metric')",
			Output:   map[string]string{"name": "metric", "tag2": "value__*"},
		},
		{
			TestName: "seriesByTag('tag2!=value21', 'name=metric.name')",
			Input:    "seriesByTag('tag2!=value21', 'name=metric.name')",
			Output:   map[string]string{"name": "metric.name"},
		},
		{
			TestName: "seriesByTag('tag2=value21')",
			Input:    "seriesByTag('tag2=value21')",
			Output:   map[string]string{"name": defaultName, "tag2": "value21"},
		},
		// brokken, from aggregation functions with seriesByTag
		{
			TestName: "seriesByTag('tag2=', 'name=metric')",
			Input:    "seriesByTag('tag2=', 'tag3', 'name=metric')",
			Output:   map[string]string{"name": "metric"},
		},
		{
			TestName: "Broken seriesByTag (missed ')",
			Input:    "seriesByTag(name=metric)",
			Output:   map[string]string{"name": defaultName},
		},
		{
			TestName: "Broken seriesByTag (missed ' at the end)",
			Input:    "seriesByTag('name=metric)",
			Output:   map[string]string{"name": defaultName},
		},
		{
			TestName: "Broken seriesByTag (missed ' at the end)",
			Input:    "seriesByTag('name=metric",
			Output:   map[string]string{"name": defaultName},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Input, func(t *testing.T) {
			res := ExtractSeriesByTags(tt.Input, defaultName)

			if len(res) != len(tt.Output) {
				t.Fatalf("result length mismatch, got %v, expected %v, %+v != %+v", len(res), len(tt.Output), res, tt.Output)
			}

			for k, v := range res {
				if expectedValue, ok := tt.Output[k]; ok {
					if v != expectedValue {
						t.Fatalf("value mismatch for key '%v': got '%v', exepcted '%v'", k, v, expectedValue)
					}
				} else {
					t.Fatalf("got unexpected key %v=%v in result", k, v)
				}
			}
		})
	}
}

func BenchmarkTestExtractSeriesByTags(b *testing.B) {
	defaultName := "sumSeries"
	benchmarks := []string{
		"seriesByTag('tag1=val1', 'tag2=val2', 'name=metric.name', 'project=~(ab|dc|ef|Test123|Test23|Test459)', 'tag3!=val3')",
	}
	for _, bm := range benchmarks {
		b.Run(bm, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				s := ExtractSeriesByTags(bm, defaultName)
				_ = s
			}
		})
	}
}

func TestExtractTags(t *testing.T) {
	tests := []extractTagTestCase{
		{
			TestName: "NoTags",
			Input:    "metric",
			Output: map[string]string{
				"name": "metric",
			},
		},
		{
			TestName: "no tags in metric",
			Input:    "cpu.usage_idle",
			Output: map[string]string{
				"name": "cpu.usage_idle",
			},
		},
		{
			TestName: "FewTags",
			Input:    "metricWithSomeTags;tag1=v1;tag2=v2;tag3=this is value with string",
			Output: map[string]string{
				"name": "metricWithSomeTags",
				"tag1": "v1",
				"tag2": "v2",
				"tag3": "this is value with string",
			},
		},
		{
			TestName: "tagged metric",
			Input:    "cpu.usage_idle;cpu=cpu-total;host=test",
			Output: map[string]string{
				"name": "cpu.usage_idle",
				"cpu":  "cpu-total",
				"host": "test",
			},
		},
		{
			TestName: "BrokenTags",
			Input:    "metric;tag1=v1;;tag2=v2;tag3=;tag4;tag5=value=with=other=equal=signs;tag6=value=with-equal-signs-2",
			Output: map[string]string{
				"name": "metric",
				"tag1": "v1",
				"tag2": "v2",
				"tag3": "",
				"tag4": "",
				"tag5": "value=with=other=equal=signs",
				"tag6": "value=with-equal-signs-2",
			},
		},
		{
			TestName: "BrokenTags2",
			Input:    "metric;tag1=v1;",
			Output: map[string]string{
				"name": "metric",
				"tag1": "v1",
			},
		},
		{
			TestName: "BrokenTags2",
			Input:    "metric;tag1",
			Output: map[string]string{
				"name": "metric",
				"tag1": "",
			},
		},
		{
			TestName: "BrokenTags3",
			Input:    "metric;=;=",
			Output: map[string]string{
				"name": "metric",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Input, func(t *testing.T) {
			res := ExtractTags(tt.Input)

			if len(res) != len(tt.Output) {
				t.Fatalf("result length mismatch, got %v, expected %v, %+v != %+v", len(res), len(tt.Output), res, tt.Output)
			}

			for k, v := range res {
				if expectedValue, ok := tt.Output[k]; ok {
					if v != expectedValue {
						t.Fatalf("value mismatch for key '%v': got '%v', exepcted '%v'", k, v, expectedValue)
					}
				} else {
					t.Fatalf("got unexpected key %v=%v in result", k, v)
				}
			}
		})
	}
}
