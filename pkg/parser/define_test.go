package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefineExpand(t *testing.T) {
	assert := assert.New(t)

	defineMap := map[string]string{
		"constMetric": "metric.name",
		"perMinute":   "perSecond({{.argString}})|scale(60)",
		"funcAlias":   "funcOrig({{index .args 0}},{{index .args 1}})",
		"funcAlias2":  "funcOrig2({{index .args 0}},{{index .kwargs \"key\"}})",
	}

	tests := []struct {
		s string
		e *expr
	}{
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
			"func1(metric1,constMetric(metricA, metricB),metric3)",
			&expr{
				target: "func1",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric1"},
					{target: "metric.name"},
					{target: "metric3"}},
				argString: "metric1,constMetric(metricA, metricB),metric3",
			},
		},
		{
			"func1(metric1,perMinute(metricA),metric3)",
			&expr{
				target: "func1",
				etype:  EtFunc,
				args: []*expr{
					{target: "metric1"},
					{target: "scale",
						etype: EtFunc,
						args: []*expr{
							{target: "perSecond",
								etype: EtFunc,
								args: []*expr{
									{target: "metricA"},
								},
								argString: "metricA",
							},
							{etype: EtConst,
								val: 60.000000,
							},
						},
						argString: "perSecond(metricA),60",
					},
					{target: "metric3"}},
				argString: "metric1,perMinute(metricA),metric3",
			},
		},
		{
			"funcAlias(metricA,metricB)",
			&expr{
				target: "funcOrig",
				etype:  EtFunc,
				args: []*expr{
					{target: "metricA"},
					{target: "metricB"},
				},
				argString: "metricA,metricB",
			},
		},
		{
			"funcAlias2(metricA,key=\"42\")",
			&expr{
				target: "funcOrig2",
				etype:  EtFunc,
				args: []*expr{
					{target: "metricA"},
					{valStr: "42", etype: EtString},
				},
				argString: "metricA,'42'",
			},
		},
	}

	d, err := NewDefine(defineMap)
	if err != nil {
		t.Errorf("compile defines failed: err=%v", err)
		return
	}

	for _, tt := range tests {
		e, _, err := ParseExpr(tt.s)
		if err != nil {
			t.Errorf("parse for %+v failed: err=%v", tt.s, err)
			continue
		}
		e, err = d.Expand(e)
		if err != nil {
			t.Errorf("expand for %+v failed: err=%v", tt.s, err)
			continue
		}

		assert.Equal(tt.e, e, tt.s)
	}
}
