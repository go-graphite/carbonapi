package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseExpr(t *testing.T) {
	tests := []struct {
		s string
		e *expr
	}{
		{"metric=",
			&expr{target: "metric="},
		},
		{"metric",
			&expr{target: "metric"},
		},
		{
			"metric.foo",
			&expr{target: "metric.foo"},
		},
		{"metric.*.foo",
			&expr{target: "metric.*.foo"},
		},
		{
			"func(metric)",
			&expr{
				target:    "func",
				etype:     EtFunc,
				args:      []*expr{{target: "metric"}},
				argString: "metric",
			},
		},
		{
			"func(metric1,metric2,metric3)",
			&expr{
				target: "func",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric1"},
					{target: "metric2"},
					{target: "metric3"}},
				argString: "metric1,metric2,metric3",
			},
		},
		{
			"func1(metric1,func2(metricA, metricB),metric3)",
			&expr{
				target: "func1",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric1"},
					{target: "func2",
						etype:     EtFunc,
						args:      []*expr{{target: "metricA"}, {target: "metricB"}},
						argString: "metricA, metricB",
					},
					{target: "metric3"}},
				argString: "metric1,func2(metricA, metricB),metric3",
			},
		},

		{
			"3",
			&expr{val: 3, etype: EtConst, valStr: "3"},
		},
		{
			"3.1",
			&expr{val: 3.1, etype: EtConst, valStr: "3.1"},
		},
		{
			"func1(metric1, 3, 1e2, 2e-3)",
			&expr{
				target: "func1",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric1"},
					{val: 3, etype: EtConst, valStr: "3"},
					{val: 100, etype: EtConst, valStr: "1e2"},
					{val: 0.002, etype: EtConst, valStr: "2e-3"},
				},
				argString: "metric1, 3, 1e2, 2e-3",
			},
		},
		{
			"func1(metric1, 'stringconst')",
			&expr{
				target: "func1",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric1"},
					{valStr: "stringconst", etype: EtString},
				},
				argString: "metric1, 'stringconst'",
			},
		},
		{
			`func1(metric1, "stringconst")`,
			&expr{
				target: "func1",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric1"},
					{valStr: "stringconst", etype: EtString},
				},
				argString: `metric1, "stringconst"`,
			},
		},
		{
			"func1(metric1, -3)",
			&expr{
				target: "func1",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric1"},
					{val: -3, etype: EtConst, valStr: "-3"},
				},
				argString: "metric1, -3",
			},
		},

		{
			"func1(metric1, -3 , 'foo' )",
			&expr{
				target: "func1",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric1"},
					{val: -3, etype: EtConst, valStr: "-3"},
					{valStr: "foo", etype: EtString},
				},
				argString: "metric1, -3 , 'foo' ",
			},
		},

		{
			"func(metric, key='value')",
			&expr{
				target: "func",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric"},
				},
				namedArgs: map[string]*expr{
					"key": {etype: EtString, valStr: "value"},
				},
				argString: "metric, key='value'",
			},
		},
		{
			"func(metric, key=true)",
			&expr{
				target: "func",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric"},
				},
				namedArgs: map[string]*expr{
					"key": {etype: EtBool, target: "true", valStr: "true"},
				},
				argString: "metric, key=true",
			},
		},
		{
			"func(metric, key=1)",
			&expr{
				target: "func",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric"},
				},
				namedArgs: map[string]*expr{
					"key": {etype: EtConst, val: 1, valStr: "1"},
				},
				argString: "metric, key=1",
			},
		},
		{
			"func(metric, key=0.1)",
			&expr{
				target: "func",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric"},
				},
				namedArgs: map[string]*expr{
					"key": {etype: EtConst, val: 0.1, valStr: "0.1"},
				},
				argString: "metric, key=0.1",
			},
		},

		{
			"func(metric, 1, key='value')",
			&expr{
				target: "func",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric"},
					{etype: EtConst, val: 1, valStr: "1"},
				},
				namedArgs: map[string]*expr{
					"key": {etype: EtString, valStr: "value"},
				},
				argString: "metric, 1, key='value'",
			},
		},
		{
			"func(metric, key='value', 1)",
			&expr{
				target: "func",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric"},
					{etype: EtConst, val: 1, valStr: "1"},
				},
				namedArgs: map[string]*expr{
					"key": {etype: EtString, valStr: "value"},
				},
				argString: "metric, key='value', 1",
			},
		},
		{
			"func(metric, key1='value1', key2='value2')",
			&expr{
				target: "func",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric"},
				},
				namedArgs: map[string]*expr{
					"key1": {etype: EtString, valStr: "value1"},
					"key2": {etype: EtString, valStr: "value2"},
				},
				argString: "metric, key1='value1', key2='value2'",
			},
		},
		{
			"func(metric, key2='value2', key1='value1')",
			&expr{
				target: "func",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric"},
				},
				namedArgs: map[string]*expr{
					"key2": {etype: EtString, valStr: "value2"},
					"key1": {etype: EtString, valStr: "value1"},
				},
				argString: "metric, key2='value2', key1='value1'",
			},
		},

		{
			`foo.{bar,baz}.qux`,
			&expr{
				target: "foo.{bar,baz}.qux",
				etype:  EtName,
			},
		},
		{
			`foo.b[0-9].qux`,
			&expr{
				target: "foo.b[0-9].qux",
				etype:  EtName,
			},
		},
		{
			`foo.b[0-9]+.qux`,
			&expr{
				target: "foo.b[0-9]+.qux",
				etype:  EtName,
			},
		},
		{
			`virt.v1.*.text-match:<foo.bar.qux>`,
			&expr{
				target: "virt.v1.*.text-match:<foo.bar.qux>",
				etype:  EtName,
			},
		},
		{
			"func2(metricA, metricB)|func1(metric1,metric3)",
			&expr{
				target: "func1",
				etype:  EtFunc,
				args: []*expr{
					{target: "func2",
						etype:     EtFunc,
						args:      []*expr{{target: "metricA"}, {target: "metricB"}},
						argString: "metricA, metricB",
					},
					{target: "metric1"},
					{target: "metric3"}},
				argString: "func2(metricA, metricB),metric1,metric3",
			},
		},
		{
			`movingAverage(company.server*.applicationInstance.requestsHandled|aliasByNode(1),"5min")`,
			&expr{
				target: "movingAverage",
				etype:  EtFunc,
				args: []*expr{
					{target: "aliasByNode",
						etype: EtFunc,
						args: []*expr{
							{target: "company.server*.applicationInstance.requestsHandled"},
							{val: 1, etype: EtConst, valStr: "1"},
						},
						argString: "company.server*.applicationInstance.requestsHandled,1",
					},
					{etype: EtString, valStr: "5min"},
				},
				argString: `aliasByNode(company.server*.applicationInstance.requestsHandled,1),"5min"`,
			},
		},
		{
			`aliasByNode(company.server*.applicationInstance.requestsHandled,1)|movingAverage("5min")`,
			&expr{
				target: "movingAverage",
				etype:  EtFunc,
				args: []*expr{
					{target: "aliasByNode",
						etype: EtFunc,
						args: []*expr{
							{target: "company.server*.applicationInstance.requestsHandled"},
							{val: 1, etype: EtConst, valStr: "1"},
						},
						argString: "company.server*.applicationInstance.requestsHandled,1",
					},
					{etype: EtString, valStr: "5min"},
				},
				argString: `aliasByNode(company.server*.applicationInstance.requestsHandled,1),"5min"`,
			},
		},
		{
			`company.server*.applicationInstance.requestsHandled|aliasByNode(1)|movingAverage("5min")`,
			&expr{
				target: "movingAverage",
				etype:  EtFunc,
				args: []*expr{
					{target: "aliasByNode",
						etype: EtFunc,
						args: []*expr{
							{target: "company.server*.applicationInstance.requestsHandled"},
							{val: 1, etype: EtConst, valStr: "1"},
						},
						argString: "company.server*.applicationInstance.requestsHandled,1",
					},
					{etype: EtString, valStr: "5min"},
				},
				argString: `aliasByNode(company.server*.applicationInstance.requestsHandled,1),"5min"`,
			},
		},
		{
			`company.server*.applicationInstance.requestsHandled|keepLastValue()`,
			&expr{
				target: "keepLastValue",
				etype:  EtFunc,
				args: []*expr{
					{target: "company.server*.applicationInstance.requestsHandled"},
				},
				argString: `company.server*.applicationInstance.requestsHandled`,
			},
		},
		{"hello&world",
			&expr{target: "hello&world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			assert := assert.New(t)

			e, _, err := ParseExpr(tt.s)
			if assert.NoError(err) {
				assert.Equal(tt.e, e, tt.s)
			}
		})
	}
}

func TestDoGetBoolVar(t *testing.T) {
	tests := []struct {
		s string
		e *expr
		r bool
	}{
		{
			"1 is true",
			&expr{val: 1, etype: EtConst, valStr: "1"},
			true,
		},
		{
			"true is true",
			&expr{etype: EtString, valStr: "true"},
			true,
		},
		{
			"True is true",
			&expr{etype: EtString, valStr: "True"},
			true,
		},
		{
			"0 is false",
			&expr{val: 0, etype: EtConst, valStr: "0"},
			false,
		},
		{
			"False is false",
			&expr{etype: EtString, valStr: "False"},
			false,
		},
		{
			"false is false",
			&expr{etype: EtString, valStr: "false"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			assert := assert.New(t)

			r, err := tt.e.doGetBoolArg()
			if assert.NoError(err) {
				assert.Equal(tt.r, r, tt.s)
			}
		})
	}
}

func TestGetIntervalNamedOrPosArgDefault(t *testing.T) {
	e, _, err := ParseExpr("func(metric, key='1min')")
	assert.NoError(t, err)

	val, err := e.GetIntervalNamedOrPosArgDefault("key", 1, -1, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(-60), val)
}

func TestDoGetFloatArg(t *testing.T) {
	tests := []struct {
		s string
		e *expr
		r float64
	}{
		{
			"parse float",
			&expr{val: 1.0, etype: EtConst, valStr: "1.0"},
			1.0,
		},
		{
			"parse string to float",
			&expr{etype: EtString, valStr: "1.0"},
			1.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			assert := assert.New(t)

			r, err := tt.e.doGetFloatArg()
			if assert.NoError(err) {
				assert.Equal(tt.r, r, tt.s)
			}
		})
	}
}

func TestDoGetIntArg(t *testing.T) {
	tests := []struct {
		s string
		e *expr
		r int
	}{
		{
			"parse int",
			&expr{val: 5, etype: EtConst, valStr: "5"},
			5,
		},
		{
			"parse string to int",
			&expr{etype: EtString, valStr: "1"},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			assert := assert.New(t)

			r, err := tt.e.doGetIntArg()
			if assert.NoError(err) {
				assert.Equal(tt.r, r, tt.s)
			}
		})
	}
}

func TestMetrics(t *testing.T) {
	tests := []struct {
		s        string
		e        *expr
		from     int64
		to       int64
		expected []MetricRequest
	}{
		{
			"hitcount(metric1, '1h', true)",
			&expr{
				target: "hitcount",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric1"},
					{valStr: "1h", etype: EtString},
					{valStr: "true", etype: EtBool},
				},
				argString: "metric1, '1h', true",
			},
			1410346740,
			1410346865,
			[]MetricRequest{
				{
					Metric: "metric1",
					From:   1410343200,
					Until:  1410346865,
				},
			},
		},
		{
			"hitcount(metric1, '1h')",
			&expr{
				target: "hitcount",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric1"},
					{valStr: "1h", etype: EtString},
				},
				argString: "metric1, '1h'",
			},
			1410346740,
			1410346865,
			[]MetricRequest{
				{
					Metric: "metric1",
					From:   1410346740,
					Until:  1410346865,
				},
			},
		},
		{
			"hitcount(timeShift(metric1, '-1h'),'1h')",
			&expr{
				target: "hitcount",
				etype:  EtFunc,
				args: []*expr{
					{
						target: "timeShift",
						etype:  EtFunc,
						args: []*expr{
							{target: "metric1"},
							{valStr: "-1h", etype: EtString},
						},
						argString: "metric1, '-1h'",
					},
					{valStr: "1h", etype: EtString},
					{valStr: "true", etype: EtBool},
				},
				argString: "timeShift(metric1, '-1h'),'1h'",
			},
			1410346740,
			1410346865,
			[]MetricRequest{
				{
					Metric: "metric1",
					From:   1410339600,
					Until:  1410343265,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {

			r := tt.e.Metrics(tt.from, tt.to)
			assert.Equal(t, tt.expected, r)
		})
	}
}
