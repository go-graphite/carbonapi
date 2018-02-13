package sort

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"sort"
)

func init() {
	functions := []string{"sortByMaxima", "sortByMinima", "sortByTotal"}
	fObj := &MathSort{}
	for _, f := range functions {
		metadata.RegisterFunction(f, fObj)
	}
	metadata.RegisterFunction("sortByName", &SortByName{})
}

type MathSort struct {
	interfaces.FunctionBase
}

// sortByMaxima(seriesList), sortByMinima(seriesList), sortByTotal(seriesList)
func (f *MathSort) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	original, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	arg := make([]*types.MetricData, len(original))
	copy(arg, original)
	vals := make([]float64, len(arg))

	for i, a := range arg {
		switch e.Target() {
		case "sortByTotal":
			vals[i] = helper.SummarizeValues("sum", a.Values)
		case "sortByMaxima":
			vals[i] = helper.SummarizeValues("max", a.Values)
		case "sortByMinima":
			vals[i] = 1 / helper.SummarizeValues("min", a.Values)
		}
	}

	sort.Sort(helper.ByVals{Vals: vals, Series: arg})

	return arg, nil
}

type SortByName struct {
	interfaces.FunctionBase
}

// sortByName(seriesList, natural=false)
func (f *SortByName) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	original, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	natSort, err := e.GetBoolNamedOrPosArgDefault("natural", 1, false)
	if err != nil {
		return nil, err
	}

	arg := make([]*types.MetricData, len(original))
	copy(arg, original)
	if natSort {
		sort.Sort(helper.ByNameNatural(arg))
	} else {
		sort.Sort(helper.ByName(arg))
	}

	return arg, nil
}
