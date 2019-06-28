package aboveSeries

import (
	"regexp"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type aboveSeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &aboveSeries{}
	functions := []string{"useSeriesAbove", "aboveSeries"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *aboveSeries) Do(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	max, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	rename := true
	search, err := e.GetStringArg(2)
	if err != nil {
		rename = false
	}

	replace, err := e.GetStringArg(3)
	if err != nil {
		rename = false
	}

	var rre *regexp.Regexp
	if rename {
		rre, err = regexp.Compile(search)
		if err != nil {
			return nil, err
		}
	}

	var results []*types.MetricData
	for _, a := range args {
		if consolidations.MaxValue(a.Values) > max {
			r := *a
			if rename {
				r.Name = rre.ReplaceAllString(r.Name, replace)
			}
			results = append(results, &r)
		}
	}

	return results, nil
}

func (f *aboveSeries) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"useSeriesAbove": {
			Name:        "useSeriesAbove",
			Description: "Takes a seriesList and compares the maximum of each series against the given value. If the series maximum is greater than value, the regular expression search and replace is applied against the series name to plot a related metric e.g. given useSeriesAbove(ganglia.metric1.reqs,10,’reqs’,’time’), the response time metric will be plotted only when the maximum value of the corresponding request/s metric is > 10\n\nShort form: aboveSeries()",
			Function:    "useSeriesAbove(seriesList, value, search, replace)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "value",
					Required: true,
					Type:     types.Integer,
				},
				{
					Name: "search",
					Type: types.String,
				},
				{
					Name: "replace",
					Type: types.String,
				},
			},
		},
	}
}
