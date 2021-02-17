package tags

import (
	"testing"
)

type extractTagTestCase struct {
	TestName string
	Input    string
	Output   map[string]string
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
				t.Fatalf("result length mismatch, got %v, expected %v", len(res), len(tt.Output))
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
