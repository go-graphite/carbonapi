package pearsonClosest

import (
	"container/heap"
	"context"
	"errors"
	"math"

	"github.com/dgryski/go-onlinestats"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type pearsonClosest struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &pearsonClosest{}
	functions := []string{"pearsonClosest"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// pearsonClosest(series, seriesList, n, direction=abs)
func (f *pearsonClosest) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() > 3 {
		return nil, types.ErrTooManyArguments
	}

	ref, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	if len(ref) != 1 {
		// TODO(nnuss) error("First argument must be single reference series")
		return nil, types.ErrWildcardNotAllowed
	}

	compare, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(1), from, until, values)
	if err != nil {
		return nil, err
	}

	n, err := e.GetIntArg(2)
	if err != nil {
		return nil, err
	}

	direction, err := e.GetStringNamedOrPosArgDefault("direction", 3, "abs")
	if err != nil {
		return nil, err
	}
	if direction != "pos" && direction != "neg" && direction != "abs" {
		return nil, errors.New("direction must be one of: pos, neg, abs")
	}

	// NOTE: if direction == "abs" && len(compare) <= n : we'll still do the work to rank them

	refValues := make([]float64, len(ref[0].Values))
	copy(refValues, ref[0].Values)

	mh := make(types.MetricHeap, 0, len(compare))

	for index, a := range compare {
		compareValues := make([]float64, len(a.Values))
		copy(compareValues, a.Values)
		if len(refValues) != len(compareValues) {
			// Pearson will panic if arrays are not equal length; skip
			continue
		}

		value := onlinestats.Pearson(refValues, compareValues)
		// Standardize the value so sort ASC will have strongest correlation first
		switch {
		case math.IsNaN(value):
			// special case of at least one series containing all zeros which leads to div-by-zero in Pearson
			continue
		case direction == "abs":
			value = math.Abs(value) * -1
		case direction == "pos" && value >= 0:
			value = value * -1
		case direction == "neg" && value <= 0:
		default:
			continue
		}
		heap.Push(&mh, types.MetricHeapElement{Idx: index, Val: value})
	}

	if n > len(mh) {
		n = len(mh)
	}
	results := make([]*types.MetricData, n)
	for i := range results {
		v := heap.Pop(&mh).(types.MetricHeapElement)
		results[i] = compare[v.Idx]
	}

	return results, nil
}

func (f *pearsonClosest) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"pearsonClosest": {
			Description: `
Implementation of Pearson product-moment correlation coefficient (PMCC) function(s)

.. code-block:: none

	pearsonClosest( series, seriesList, n, direction="abs" )


Return the n series in seriesList with closest Pearson score to the first series argument.
An optional direction parameter may also be given:
	"abs"   - (default) Series with any Pearson score + or - [-1 .. 1].
    "pos"   - Only series with positive correlation score [0 .. 1]
    "neg"   - Only series with negative correlation score [1 .. 0]


The default is "abs" which is most correlated (in either direction)

Examples:

.. code-block:: none

	#  metrics from 'metric.forest.*'' that "look like" 'metric.abnormal'' (have closest correllation coeeficient)
	pearsonClosest( metric.abnormal , metric.forest.* , 2, direction="pos" )


.. code-block:: none

	# 2 metrics from "metric.forest.*"" that are most negatively correlated to "metric.increasing" (ie. "metric.forest.decreasing" )
	pearsonClosest( metric.increasing , metric.forest.* , 2, direction="neg" )


.. code-block:: none

    # you'd get "metric.increasing", "metric.decreasing"
    pearsonClosest( metric.increasing, group (metric.increasing, metric.decreasing, metric.flat, metric.sine), 2 )

Note:
Pearson will discard epochs where either series has a missing value.

Additionally there is a special case where a series (or window) containing only zeros leads to a division-by-zero
and will manifest as if the entire window/series had missing values.`,
			Function: "pearsonClosest(seriesList, seriesList, n, direction)",
			Group:    "Transform",
			Module:   "graphite.render.functions.custom",
			Name:     "pearsonClosest",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
				{
					Name:     "direction",
					Required: true,
					Options: types.StringsToSuggestionList([]string{
						"abs",
						"pos",
						"neg",
					}),
					Type: types.String,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
	}
}
