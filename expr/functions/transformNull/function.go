package transformNull

import (
	"context"
	"fmt"
	"math"
	"strconv"

	pbv3 "github.com/go-graphite/protocol/carbonapi_v3_pb"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type transformNull struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &transformNull{}
	functions := []string{"transformNull"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// transformNull(seriesList, default=0)
func (f *transformNull) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	defv, err := e.GetFloatNamedOrPosArgDefault("default", 1, 0)
	if err != nil {
		return nil, err
	}
	defaultOnAbsent, err := e.GetBoolNamedOrPosArgDefault("defaultOnAbsent", 2, false)
	if err != nil {
		return nil, err
	}

	_, ok := e.NamedArg("default")
	if !ok {
		ok = e.ArgsLen() > 1
	}
	var defvStr string
	if defv != 0 {
		defvStr = strconv.FormatFloat(defv, 'g', -1, 64)
	}

	var valMap []bool
	referenceSeriesExpr := e.GetNamedArg("referenceSeries")
	if !referenceSeriesExpr.IsInterfaceNil() {
		referenceSeries, err := helper.GetSeriesArg(ctx, referenceSeriesExpr, from, until, values)
		if err != nil {
			return nil, err
		}

		if len(referenceSeries) == 0 {
			return nil, fmt.Errorf("reference series is not a valid metric")
		}
		length := len(referenceSeries[0].Values)
		if length != len(arg[0].Values) {
			return nil, fmt.Errorf("length of series and reference series must be the same")
		}
		valMap = make([]bool, length)

		for _, a := range referenceSeries {
			for i, v := range a.Values {
				if !math.IsNaN(v) {
					valMap[i] = true
				}
			}
		}
	}

	results := make([]*types.MetricData, 0, len(arg)+1)
	for _, a := range arg {
		var name string
		if ok {
			name = "transformNull(" + a.Name + "," + defvStr + ")"
		} else {
			name = "transformNull(" + a.Name + ")"
		}

		r := a.CopyTag(name, a.Tags)
		r.Values = make([]float64, len(a.Values))

		for i, v := range a.Values {
			if math.IsNaN(v) {
				if len(valMap) == 0 {
					v = defv
				} else if valMap[i] {
					v = defv
				}
			}

			r.Values[i] = v
		}

		results = append(results, r)
	}
	if len(arg) == 0 && defaultOnAbsent {
		values := []float64{defv, defv}
		step := until - from
		results = append(results, &types.MetricData{
			FetchResponse: pbv3.FetchResponse{
				Name:      e.ToString(),
				StartTime: from,
				StopTime:  from + step*int64(len(values)),
				StepTime:  step,
				Values:    values,
			},
			Tags: map[string]string{"name": types.ExtractName(e.ToString())},
		})
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *transformNull) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"transformNull": {
			Description: `Takes a metric or wildcard seriesList and replaces null values with the value
  specified by 'default'.  The value 0 used if not specified.  The optional
  referenceSeries, if specified, is a metric or wildcard series list that governs
  which time intervals nulls should be replaced.  If specified, nulls are replaced
  only in intervals where a non-null is found for the same interval in any of
  referenceSeries.  This method compliments the drawNullAsZero function in
  graphical mode, but also works in text-only mode.
  defaultOnAbsent if specified, would produce a constant line if no metrics will be returned by backends.
  Example:
  .. code-block:: none
    &target=transformNull(webapp.pages.*.views,-1)
  This would take any page that didn't have values and supply negative 1 as a default.
  Any other numeric value may be used as well.
`,
			Function: "transformNull(seriesList, default=0, referenceSeries=None)",
			Group:    "Transform",
			Module:   "graphite.render.functions",
			Name:     "transformNull",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: types.NewSuggestion(0),
					Name:    "default",
					Type:    types.Float,
				},
				{
					Name: "referenceSeries",
					Type: types.SeriesList,
				},
				{
					Name: "defaultOnAbsent",
					Type: types.Boolean,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
