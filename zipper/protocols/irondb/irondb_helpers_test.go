package irondb

import (
	"testing"
)

func TestGraphiteExprListToIronDBTagQuery(t *testing.T) {
	cases := []struct {
		desc           string
		input          []string
		expectedOutput string
	}{
		{"TestEmpty", []string{}, ""},
		{"TestNameEq", []string{"name=host"}, "and(__name:host)"},
		{"TestEq", []string{"tag=value"}, "and(tag:value)"},
		{"TestEqTilda", []string{"tag=~value.*"}, "and(tag:/value.*/)"},
		{"TestNotEq", []string{"tag!=value"}, "not(tag:value)"},
		{"TestNotEqTilda", []string{"tag!=~value.*"}, "not(tag:/value.*/)"},
		{"TestNotEqTildaWrong", []string{"tag!~value.*"}, ""},
		{"TestComplex", []string{`tag!=~value\..*`, "tag2=host", "tag!=cat3"}, `and(not(tag:/value\..*/),and(tag2:host),not(tag:cat3))`},
	}
	for _, tc := range cases {
		output := graphiteExprListToIronDBTagQuery(tc.input)
		if output != tc.expectedOutput {
			t.Fatalf("%s: expected value: %s got: %s for input: %s",
				tc.desc, tc.expectedOutput, output, tc.input)
		}
	}
}

func TestConvertNameToGraphite(t *testing.T) {
	cases := []struct {
		desc           string
		input          string
		expectedOutput string
	}{
		{"TestEmpty", "", ""},
		{"Test1", "name|ST[c1=v1,c2=v2,c3=v3]", "name;c1=v1;c2=v2;c3=v3"},
		{"Test1", "name|ST[c1=v1,c2=v2,c3=v3]|ST[a=b]", "name;c1=v1;c2=v2;c3=v3;a=b"},
		{"Test1", "name|ST[c1=v1,c2=v2,c3=v3]|MT{taag}", "name;c1=v1;c2=v2;c3=v3"},
		{"Test1", "name|MT{tag1}|ST[c1=v1,c2=v2,c3=v3]|MT{tag}", "name;c1=v1;c2=v2;c3=v3"},
		{"Test1", "name|ST[a=b]|ST[c1=v1,c2=v2,c3=v3]", "name;a=b;c1=v1;c2=v2;c3=v3"},
	}
	for _, tc := range cases {
		output := convertNameToGraphite(tc.input)
		if output != tc.expectedOutput {
			t.Fatalf("%s: expected value: %s got: %s for input: %s",
				tc.desc, tc.expectedOutput, output, tc.input)
		}
	}
}

func BenchmarkConvertNameToGraphite(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = convertNameToGraphite("name|MT{tag1}|ST[c1=v1,c2=v2,c3=v3]|MT{tag}")
	}
}

func TestAdjustStep(t *testing.T) {
	cases := []struct {
		desc                                    string
		start, stop, maxPointsPerQuery, minStep int64
		expectedOutput                          int64
	}{
		{"TestZero", 0, 600, 0, 60, 60},
		{"Test10", 0, 600, 10, 10, 60},
		{"Test60", 0, 6000, 60, 10, 120},
		{"Test600", 0, 6000, 600, 10, 10},
		{"Test7200", 0, 60000, 10, 10, 7200},
		{"Test21600", 0, 120000, 10, 60, 21600},
		{"Test86400", 0, 120000, 1, 60, 86400},
	}
	for _, tc := range cases {
		output := adjustStep(tc.start, tc.stop, tc.maxPointsPerQuery, tc.minStep)
		if output != tc.expectedOutput {
			t.Fatalf("%s: expected value: %d got: %d for input: %d %d %d %d",
				tc.desc, tc.expectedOutput, output, tc.start, tc.stop, tc.maxPointsPerQuery, tc.minStep)
		}
	}
}
