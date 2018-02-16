package transformNull

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	f := &transformNull{}
	functions := []string{"transformNull"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type transformNull struct {
	interfaces.FunctionBase
}

// transformNull(seriesList, default=0)
func (f *transformNull) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	defv, err := e.GetFloatNamedOrPosArgDefault("default", 1, 0)
	if err != nil {
		return nil, err
	}

	_, ok := e.NamedArgs()["default"]
	if !ok {
		ok = len(e.Args()) > 1
	}

	// FIXME(civil): support referenceSeries

	var results []*types.MetricData

	for _, a := range arg {

		var name string
		if ok {
			name = fmt.Sprintf("transformNull(%s,%g)", a.Name, defv)
		} else {
			name = fmt.Sprintf("transformNull(%s)", a.Name)
		}

		r := *a
		r.Name = name
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				v = defv
			}

			r.Values[i] = v
		}

		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *transformNull) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"transformNull": {
			Description: "Takes a metric or wildcard seriesList and replaces null values with the value\nspecified by `default`.  The value 0 used if not specified.  The optional\nreferenceSeries, if specified, is a metric or wildcard series list that governs\nwhich time intervals nulls should be replaced.  If specified, nulls are replaced\nonly in intervals where a non-null is found for the same interval in any of\nreferenceSeries.  This method compliments the drawNullAsZero function in\ngraphical mode, but also works in text-only mode.\n\nExample:\n\n.. code-block:: none\n\n  &target=transformNull(webapp.pages.*.views,-1)\n\nThis would take any page that didn't have values and supply negative 1 as a default.\nAny other numeric value may be used as well.",
			Function:    "transformNull(seriesList, default=0, referenceSeries=None)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "transformNull",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: "0",
					Name:    "default",
					Type:    types.Float,
				},
				/*				{
									Name: "referenceSeries",
									Type: types.SeriesList,
								},
				*/
			},
		},
	}
}
