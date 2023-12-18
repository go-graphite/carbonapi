package removeEmptySeries

import (
	"context"
	"math"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type removeEmptySeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &removeEmptySeries{}
	functions := []string{"removeEmptySeries", "removeZeroSeries"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// removeEmptySeries(seriesLists, n), removeZeroSeries(seriesLists, n)
func (f *removeEmptySeries) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	var xFilesFactor float64

	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return []*types.MetricData{}, nil
	}

	if e.ArgsLen() == 2 {
		xFilesFactor, err = e.GetFloatArgDefault(1, float64(args[0].XFilesFactor)) // If set by setXFilesFactor, all series in a list will have the same value
		if err != nil {
			return nil, err
		}
	}

	xFilesFactorStr := strconv.FormatFloat(xFilesFactor, 'f', -1, 64)

	results := make([]*types.MetricData, 0, len(args))
	for _, arg := range args {
		nonNull := 0
		for _, v := range arg.Values {
			if !math.IsNaN(v) {
				switch e.Target() {
				case "removeEmptySeries":
					nonNull++
				case "removeZeroSeries":
					if v != 0 {
						nonNull++
					}
				}
			}
		}

		if nonNull != 0 && helper.XFilesFactor(nonNull, len(arg.Values), xFilesFactor) {
			r := arg.CopyLink()
			r.Tags[e.Target()] = xFilesFactorStr
			results = append(results, r)
		}
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *removeEmptySeries) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"removeEmptySeries": {
			Description: "Takes one metric or a wildcard seriesList.\nOut of all metrics passed, draws only the metrics with not empty data\n\nExample:\n\n.. code-block:: none\n\n  &target=removeEmptySeries(server*.instance*.threads.busy)\n\nDraws only live servers with not empty data.\n\n`xFilesFactor` follows the same semantics as in Whisper storage schemas.  Setting it to 0 (the\ndefault) means that only a single value in the series needs to be non-null for it to be\nconsidered non-empty, setting it to 1 means that all values in the series must be non-null.\nA setting of 0.5 means that at least half the values in the series must be non-null.",
			Function:    "removeEmptySeries(seriesList, xFilesFactor=None)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "removeEmptySeries",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "xFilesFactor",
					Required: true,
					Type:     types.Float,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
		"removeZeroSeries": {
			Description: "Takes one metric or a wildcard seriesList.\nOut of all metrics passed, draws only the metrics with not ZERO data\n\nExample:\n\n.. code-block:: none\n\n  &target=removeZeroSeries(server*.instance*.threads.busy)\n\nDraws only live servers with not empty data.\n\n`xFilesFactor` follows the same semantics as in Whisper storage schemas.  Setting it to 0 (the\ndefault) means that only a single value in the series needs to be non-null for it to be\nconsidered non-empty, setting it to 1 means that all values in the series must be non-null.\nA setting of 0.5 means that at least half the values in the series must be non-null.",
			Function:    "removeZeroSeries(seriesList, xFilesFactor=None)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions.custom",
			Name:        "removeZeroSeries",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "xFilesFactor",
					Required: true,
					Type:     types.Float,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
