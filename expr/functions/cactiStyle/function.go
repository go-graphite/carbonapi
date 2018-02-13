package cactiStyle

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

func init() {
	f := &Function{}
	functions := []string{"cactiStyle"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type Function struct {
	interfaces.FunctionBase
}

// cactiStyle(seriesList, system=None, units=None)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// Get the series data
	original, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	// Get the arguments
	system, err := e.GetStringNamedOrPosArgDefault("system", 1, "")
	if err != nil {
		return nil, err
	}
	unit, err := e.GetStringNamedOrPosArgDefault("units", 2, "")
	if err != nil {
		return nil, err
	}

	// Deal with each of the series
	var metrics []*types.MetricData
	for _, a := range original {
		// Calculate min, max, current
		//
		// This saves calling helper.SummarizeValues 3 times and looping over
		// the metrics 3 times
		//
		// For min:
		// Ignoring any absent values and inf (if we have a value)
		// Using helper.SummarizeValues("min", ...) results in incorrect values, when absent
		// values are present
		//
		minVal := math.Inf(1)
		currentVal := math.Inf(-1)
		maxVal := math.Inf(-1)
		for i, av := range a.Values {
			if !a.IsAbsent[i] {
				minVal = math.Min(minVal, av)
				maxVal = math.Max(maxVal, av)
				currentVal = av
			}
		}

		// Format the output correctly
		min := ""
		max := ""
		current := ""
		if system == "si" {
			mv, mf := humanize.ComputeSI(minVal)
			xv, xf := humanize.ComputeSI(maxVal)
			cv, cf := humanize.ComputeSI(currentVal)

			min = fmt.Sprintf("%.0f%s", mv, mf)
			max = fmt.Sprintf("%.0f%s", xv, xf)
			current = fmt.Sprintf("%.0f%s", cv, cf)

		} else if system == "" {
			min = fmt.Sprintf("%.0f", minVal)
			max = fmt.Sprintf("%.0f", maxVal)
			current = fmt.Sprintf("%.0f", currentVal)

		} else {
			return nil, fmt.Errorf("%s is not supported for system", system)
		}

		// Append the unit if specified
		if len(unit) > 0 {
			min = fmt.Sprintf("%s %s", min, unit)
			max = fmt.Sprintf("%s %s", max, unit)
			current = fmt.Sprintf("%s %s", current, unit)
		}

		r := *a
		r.Name = fmt.Sprintf("%s Current: %s Max: %s Min: %s", a.Name, current, max, min)
		metrics = append(metrics, &r)
	}

	return metrics, nil
}
