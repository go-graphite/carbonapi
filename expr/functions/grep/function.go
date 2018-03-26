package grep

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"regexp"
)

type grep struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &grep{}
	functions := []string{"grep"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// grep(seriesList, pattern)
func (f *grep) Do(e parser.Expr, from, until uint32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	pat, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	patre, err := regexp.Compile(pat)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range arg {
		if patre.MatchString(a.Name) {
			results = append(results, a)
		}
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *grep) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"grep": {
			Description: "Takes a metric or a wildcard seriesList, followed by a regular expression\nin double quotes.  Excludes metrics that don't match the regular expression.\n\nExample:\n\n.. code-block:: none\n\n  &target=grep(servers*.instance*.threads.busy,\"server02\")",
			Function:    "grep(seriesList, pattern)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "grep",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "pattern",
					Required: true,
					Type:     types.String,
				},
			},
		},
	}
}
