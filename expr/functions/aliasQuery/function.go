package aliasQuery

import (
	"context"
	"fmt"
	"regexp"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type aliasQuery struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{
		{Name: "aliasQuery", F: &aliasQuery{}},
	}
}

func (f *aliasQuery) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	seriesList, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	search, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}
	replace, err := e.GetStringArg(2)
	if err != nil {
		return nil, err
	}
	newName, err := e.GetStringArg(3)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(search)
	if err != nil {
		return nil, err
	}
	replace = helper.Backref.ReplaceAllString(replace, "$${$1}")

	fetchTargets := make([]parser.Expr, len(seriesList))
	for i, series := range seriesList {
		newTarget := re.ReplaceAllString(series.Name, replace)
		expr, _, err := parser.ParseExpr(newTarget)
		if err != nil {
			return nil, err
		}
		fetchTargets[i] = expr
	}
	targetValues, err := f.GetEvaluator().Fetch(ctx, fetchTargets, from, until, values)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, len(seriesList))

	for i, series := range seriesList {
		v, err := f.getLastValueOfSeries(ctx, fetchTargets[i], from, until, targetValues)
		if err != nil {
			return nil, err
		}

		n := fmt.Sprintf(newName, v)

		var r *types.MetricData
		if series.Name == n {
			r = series.CopyLinkTags()
			r.Tags["name"] = r.Name
		} else {
			r = series.CopyName(n)
		}

		results[i] = r
	}

	return results, nil
}

func (f *aliasQuery) getLastValueOfSeries(ctx context.Context, e parser.Expr, from, until int64, targetValues map[parser.MetricRequest][]*types.MetricData) (float64, error) {
	res, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e, from, until, targetValues)
	if err != nil {
		return 0, err
	}

	if len(res) == 0 {
		return 0, parser.ErrMissingTimeseries
	}

	if len(res[0].Values) == 0 {
		return 0, parser.ErrMissingValues
	}

	return res[0].Values[len(res[0].Values)-1], nil
}

func (f *aliasQuery) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"aliasQuery": {
			Description: "Performs a query to alias the metrics in seriesList.\nThe series in seriesList will be aliased by first translating the series names using the search & replace parameters, then using the last value of the resulting series to construct the alias using sprintf-style syntax.",
			Function:    "aliasQuery(seriesList, search, replace, newName)",
			Group:       "Alias",
			Module:      "graphite.render.functions",
			Name:        "aliasQuery",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "search",
					Required: true,
					Type:     types.String,
				},
				{
					Name:     "replace",
					Required: true,
					Type:     types.String,
				},
				{
					Name:     "newName",
					Required: true,
					Type:     types.String,
				},
			},
			NameChange: true,
			TagsChange: true,
		},
	}
}
