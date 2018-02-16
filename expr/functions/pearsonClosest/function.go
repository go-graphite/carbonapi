package pearsonClosest

import (
	"container/heap"
	"errors"
	"github.com/dgryski/go-onlinestats"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

func init() {
	metadata.RegisterFunction("pearsonClosest", &pearsonClosest{})
}

type pearsonClosest struct {
	interfaces.FunctionBase
}

// pearsonClosest(series, seriesList, n, direction=abs)
func (f *pearsonClosest) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if len(e.Args()) > 3 {
		return nil, types.ErrTooManyArguments
	}

	ref, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	if len(ref) != 1 {
		// TODO(nnuss) error("First argument must be single reference series")
		return nil, types.ErrWildcardNotAllowed
	}

	compare, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
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
	for i, v := range ref[0].IsAbsent {
		if v {
			refValues[i] = math.NaN()
		}
	}

	var mh types.MetricHeap

	for index, a := range compare {
		compareValues := make([]float64, len(a.Values))
		copy(compareValues, a.Values)
		if len(refValues) != len(compareValues) {
			// Pearson will panic if arrays are not equal length; skip
			continue
		}
		for i, v := range a.IsAbsent {
			if v {
				compareValues[i] = math.NaN()
			}
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

func (f *pearsonClosest) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"pearsonClosest": {
			Description: "The Pearson product-moment correlation coefficient (PPMCC) or the bivariate correlation,[1] is a measure of the linear correlation between two variables X and Y. It has a value between +1 and −1, where 1 is total positive linear correlation, 0 is no linear correlation, and −1 is total negative linear correlation.",
			Function:    "pearsonClosest(seriesList, seriesList, n, direction)",
			Group:       "Transform",
			Module:      "graphite.render.functions.custom",
			Name:        "pearsonClosest",
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
					Options: []string{
						"abs",
						"pos",
						"neg",
					},
					Type:     types.String,
				},
			},
		},
	}
}
