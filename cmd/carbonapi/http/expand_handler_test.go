package http

import (
	"testing"

	pbv3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

func TestExpandEncoder(t *testing.T) {
	var tests = []struct {
		name        string
		metricIn    pbv3.MultiGlobResponse
		metricOut   string
		leavesOnly  string
		groupByExpr string
	}{
		{
			name: "test1",
			metricIn: pbv3.MultiGlobResponse{
				Metrics: []pbv3.GlobResponse{
					{
						Name: "foo.ba*",
						Matches: []pbv3.GlobMatch{
							{Path: "foo.bar", IsLeaf: false},
							{Path: "foo.bat", IsLeaf: true},
						},
					},
				},
			},
			metricOut:   "{\"results\":[\"foo.bar\",\"foo.bat\"]}",
			leavesOnly:  "0",
			groupByExpr: "0",
		},
		{
			name: "test2",
			metricIn: pbv3.MultiGlobResponse{
				Metrics: []pbv3.GlobResponse{
					{
						Name: "foo.ba*",
						Matches: []pbv3.GlobMatch{
							{Path: "foo.bar", IsLeaf: false},
							{Path: "foo.bat", IsLeaf: true},
						},
					},
				},
			},
			metricOut:   "{\"results\":[\"foo.bat\"]}",
			leavesOnly:  "1",
			groupByExpr: "0",
		},
		{
			name: "test3",
			metricIn: pbv3.MultiGlobResponse{
				Metrics: []pbv3.GlobResponse{
					{
						Name: "foo.ba*",
						Matches: []pbv3.GlobMatch{
							{Path: "foo.bar", IsLeaf: false},
							{Path: "foo.bat", IsLeaf: true},
						},
					},
				},
			},
			metricOut:   "{\"results\":{\"foo.ba*\":[\"foo.bar\",\"foo.bat\"]}}",
			leavesOnly:  "0",
			groupByExpr: "1",
		},
		{
			name: "test4",
			metricIn: pbv3.MultiGlobResponse{
				Metrics: []pbv3.GlobResponse{
					{
						Name: "foo.ba*",
						Matches: []pbv3.GlobMatch{
							{Path: "foo.bar", IsLeaf: false},
							{Path: "foo.bat", IsLeaf: true},
						},
					},
					{
						Name: "foo.ba*.*",
						Matches: []pbv3.GlobMatch{
							{Path: "foo.bar", IsLeaf: false},
							{Path: "foo.bat", IsLeaf: true},
							{Path: "foo.bar.baz", IsLeaf: true},
						},
					},
				},
			},
			metricOut:   "{\"results\":{\"foo.ba*\":[\"foo.bar\",\"foo.bat\"],\"foo.ba*.*\":[\"foo.bar.baz\"]}}",
			leavesOnly:  "0",
			groupByExpr: "1",
		},
	}
	for _, tst := range tests {
		tst := tst
		t.Run(tst.name, func(t *testing.T) {
			response, _ := expandEncoder(&tst.metricIn, tst.leavesOnly, tst.groupByExpr)
			if tst.metricOut != string(response) {
				t.Errorf("%v should be same as %v", tst.metricOut, string(response))
			}
		})
	}
}
