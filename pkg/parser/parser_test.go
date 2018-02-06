package parser

import (
	"testing"
	"reflect"

	"github.com/davecgh/go-spew/spew"
)

func TestParseExpr(t *testing.T) {

	tests := []struct {
		s string
		e *expr
	}{
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
			&expr{val: 3, etype: EtConst},
		},
		{
			"3.1",
			&expr{val: 3.1, etype: EtConst},
		},
		{
			"func1(metric1, 3, 1e2, 2e-3)",
			&expr{
				target: "func1",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric1"},
					{val: 3, etype: EtConst},
					{val: 100, etype: EtConst},
					{val: 0.002, etype: EtConst},
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
					{val: -3, etype: EtConst},
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
					{val: -3, etype: EtConst},
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
					"key": {etype: EtName, target: "true"},
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
					"key": {etype: EtConst, val: 1},
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
					"key": {etype: EtConst, val: 0.1},
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
					{etype: EtConst, val: 1},
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
					{etype: EtConst, val: 1},
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
			`virt.v1.*.text-match:<foo.bar.qux>`,
			&expr{
				target: "virt.v1.*.text-match:<foo.bar.qux>",
				etype:  EtName,
			},
		},
	}

	for _, tt := range tests {
		e, _, err := ParseExpr(tt.s)
		if err != nil {
			t.Errorf("parse for %+v failed: err=%v", tt.s, err)
			continue
		}
		if !reflect.DeepEqual(e, tt.e) {
			t.Errorf("parse for %+v failed:\ngot  %+s\nwant %+v", tt.s, spew.Sdump(e), spew.Sdump(tt.e))
		}
	}
}