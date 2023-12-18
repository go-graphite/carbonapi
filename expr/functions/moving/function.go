package moving

import (
	"context"
	"math"
	"strconv"

	"github.com/lomik/zapwriter"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type moving struct {
	interfaces.FunctionBase

	config movingConfig
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

type movingConfig struct {
	ReturnNaNsIfStepMismatch *bool
}

func New(configFile string) []interfaces.FunctionMetadata {
	logger := zapwriter.Logger("functionInit").With(zap.String("function", "moving"))
	res := make([]interfaces.FunctionMetadata, 0)
	f := &moving{}
	functions := []string{"movingAverage", "movingMin", "movingMax", "movingSum", "movingWindow"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	cfg := movingConfig{}
	v := viper.New()
	v.SetConfigFile(configFile)
	err := v.ReadInConfig()
	if err != nil {
		logger.Info("failed to read config file, using default",
			zap.Error(err),
		)
	} else {
		err = v.Unmarshal(&cfg)
		if err != nil {
			logger.Fatal("failed to parse config",
				zap.Error(err),
			)
			return nil
		}
		f.config = cfg
	}

	if cfg.ReturnNaNsIfStepMismatch == nil {
		v := true
		f.config.ReturnNaNsIfStepMismatch = &v
	}
	return res
}

// movingXyz(seriesList, windowSize)
func (f *moving) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	var n int
	var err error

	var argstr string
	var cons string

	var xFilesFactor float64

	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	adjustedStart := from
	var refetch bool
	var windowPoints int
	var preview int64

	switch e.Arg(1).Type() {
	case parser.EtConst:
		n, err = e.GetIntArg(1)
		argstr = strconv.Itoa(n)

		arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
		if err != nil {
			return nil, err
		}
		if len(arg) == 0 {
			return arg, nil
		}

		// Find the maximum step to use for determining the altered start time
		var maxStep int64
		for _, a := range arg {
			if a.StepTime > maxStep {
				maxStep = a.StepTime
			}
		}
		preview = maxStep * int64(n)
		adjustedStart -= maxStep * int64(n)
		windowPoints = n
		if adjustedStart != from {
			refetch = true
		}
	case parser.EtString:
		var n32 int32
		n32, err = e.GetIntervalArg(1, 1)
		argstr = "'" + e.Arg(1).StringValue() + "'"
		preview = int64(math.Abs(float64(n32))) // Absolute is used in order to handle negative string intervals
		adjustedStart -= preview
	default:
		err = parser.ErrBadType
	}
	if err != nil {
		return nil, err
	}

	var targetValues map[parser.MetricRequest][]*types.MetricData
	if refetch {
		targetValues, err = f.GetEvaluator().Fetch(ctx, []parser.Expr{e.Arg(0)}, adjustedStart, until, values)
		if err != nil {
			return nil, err
		}
	} else {
		targetValues = values
	}

	adjustedArgs, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), adjustedStart, until, targetValues)
	if err != nil {
		return nil, err
	}

	if len(adjustedArgs) == 0 {
		return adjustedArgs, nil
	}

	if e.ArgsLen() >= 2 && e.Target() == "movingWindow" {
		cons, err = e.GetStringArgDefault(2, "average")
		if err != nil {
			return nil, err
		}

		if e.ArgsLen() == 4 {
			xFilesFactor, err = e.GetFloatArgDefault(3, float64(adjustedArgs[0].XFilesFactor))

			if err != nil {
				return nil, err
			}
		}
	} else if e.ArgsLen() == 3 {
		xFilesFactor, err = e.GetFloatArgDefault(2, float64(adjustedArgs[0].XFilesFactor))

		if err != nil {
			return nil, err
		}
	}

	switch e.Target() {
	case "movingAverage":
		cons = "average"
	case "movingSum":
		cons = "sum"
	case "movingMin":
		cons = "min"
	case "movingMax":
		cons = "max"
	case "movingMedian":
		cons = "median"
	}

	result := make([]*types.MetricData, len(adjustedArgs))

	for j, a := range adjustedArgs {
		r := a.CopyName(e.Target() + "(" + a.Name + "," + argstr + ")")
		r.Tags[e.Target()] = argstr

		if e.Arg(1).Type() == parser.EtString {
			windowPoints = int(preview / a.StepTime)
		}

		if windowPoints == 0 {
			if *f.config.ReturnNaNsIfStepMismatch {
				r.Values = make([]float64, len(a.Values))
				for i := range a.Values {
					r.Values[i] = math.NaN()
				}
			}
			r.StartTime += preview
			r.StopTime += preview
			result[j] = r
			continue
		}

		size := len(a.Values) - windowPoints
		if size < 0 {
			size = 0
		}
		r.Values = make([]float64, size)
		r.StartTime = a.StartTime + preview
		r.StopTime = r.StartTime + int64(len(r.Values))*r.StepTime

		w := &types.Windowed{Data: make([]float64, windowPoints)}
		for i := 1; i < len(a.Values); i++ { // ignoring the first value in the series to avoid shifting of results one step in the future
			w.Push(a.Values[i])

			if ridx := i - windowPoints; ridx >= 0 {
				if w.IsNonNull() && helper.XFilesFactorValues(w.Data, xFilesFactor) {
					switch cons {
					case "average":
						r.Values[ridx] = w.Mean()
					case "avg":
						r.Values[ridx] = w.Mean()
					case "avg_zero":
						r.Values[ridx] = w.MeanZero()
					case "sum":
						r.Values[ridx] = w.Sum()
					case "min":
						r.Values[ridx] = w.Min()
					case "max":
						r.Values[ridx] = w.Max()
					case "multiply":
						r.Values[ridx] = w.Multiply()
					case "range":
						r.Values[ridx] = w.Range()
					case "diff":
						r.Values[ridx] = w.Diff()
					case "stddev":
						r.Values[ridx] = w.Stdev()
					case "count":
						r.Values[ridx] = w.Count()
					case "last":
						r.Values[ridx] = w.Last()
					case "median":
						r.Values[ridx] = w.Median()
					}
					if i < windowPoints || math.IsNaN(r.Values[ridx]) {
						r.Values[ridx] = math.NaN()
					}
				} else {
					r.Values[ridx] = math.NaN()
				}
			}
		}
		result[j] = r
	}
	return result, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *moving) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"movingWindow": {
			Description: "Graphs a moving window function of a metric (or metrics) over a fixed number of past points, or a time interval.\n\nTakes one metric or a wildcard seriesList followed by a number N of datapoints\nor a quoted string with a length of time like '1hour' or '5min' (See ``from /\nuntil`` in the render\\_api_ for examples of time formats), and an xFilesFactor value to specify\nhow many points in the window must be non-null for the output to be considered valid. Graphs the\nsum of the preceeding datapoints for each point on the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=movingWindow(Server.instance01.threads.busy,10)\n  &target=movingWindow(Server.instance*.threads.idle,'5min','median',0.5)",
			Function:    "movingWindow(seriesList, windowSize, func='average', xFilesFactor=None)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "movingWindow",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "windowSize",
					Required: true,
					Suggestions: types.NewSuggestions(
						5,
						7,
						10,
						"1min",
						"5min",
						"10min",
						"30min",
						"1hour",
					),
					Type: types.IntOrInterval,
				},
				{
					Name: "func",
					Type: types.AggFunc,
				},
				{
					Name: "xFilesFactor",
					Type: types.Float,
				},
			},
		},
		"movingAverage": {
			Description: "Graphs the moving average of a metric (or metrics) over a fixed number of\npast points, or a time interval.\n\nTakes one metric or a wildcard seriesList followed by a number N of datapoints\nor a quoted string with a length of time like '1hour' or '5min' (See ``from /\nuntil`` in the render\\_api_ for examples of time formats), and an xFilesFactor value to specify\nhow many points in the window must be non-null for the output to be considered valid. Graphs the\naverage of the preceeding datapoints for each point on the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=movingAverage(Server.instance01.threads.busy,10)\n  &target=movingAverage(Server.instance*.threads.idle,'5min')",
			Function:    "movingAverage(seriesList, windowSize, xFilesFactor=None)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "movingAverage",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "windowSize",
					Required: true,
					Suggestions: types.NewSuggestions(
						5,
						7,
						10,
						"1min",
						"5min",
						"10min",
						"30min",
						"1hour",
					),
					Type: types.IntOrInterval,
				},
				{
					Name: "xFilesFactor",
					Type: types.Float,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
		"movingMin": {
			Description: "Graphs the moving minimum of a metric (or metrics) over a fixed number of\npast points, or a time interval.\n\nTakes one metric or a wildcard seriesList followed by a number N of datapoints\nor a quoted string with a length of time like '1hour' or '5min' (See ``from /\nuntil`` in the render\\_api_ for examples of time formats), and an xFilesFactor value to specify\nhow many points in the window must be non-null for the output to be considered valid. Graphs the\nminimum of the preceeding datapoints for each point on the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=movingMin(Server.instance01.requests,10)\n  &target=movingMin(Server.instance*.errors,'5min')",
			Function:    "movingMin(seriesList, windowSize, xFilesFactor=None)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "movingMin",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "windowSize",
					Required: true,
					Suggestions: types.NewSuggestions(
						5,
						7,
						10,
						"1min",
						"5min",
						"10min",
						"30min",
						"1hour",
					),
					Type: types.IntOrInterval,
				},
				{
					Name: "xFilesFactor",
					Type: types.Float,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
		"movingMax": {
			Description: "Graphs the moving maximum of a metric (or metrics) over a fixed number of\npast points, or a time interval.\n\nTakes one metric or a wildcard seriesList followed by a number N of datapoints\nor a quoted string with a length of time like '1hour' or '5min' (See ``from /\nuntil`` in the render\\_api_ for examples of time formats), and an xFilesFactor value to specify\nhow many points in the window must be non-null for the output to be considered valid. Graphs the\nmaximum of the preceeding datapoints for each point on the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=movingMax(Server.instance01.requests,10)\n  &target=movingMax(Server.instance*.errors,'5min')",
			Function:    "movingMax(seriesList, windowSize, xFilesFactor=None)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "movingMax",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "windowSize",
					Required: true,
					Suggestions: types.NewSuggestions(
						5,
						7,
						10,
						"1min",
						"5min",
						"10min",
						"30min",
						"1hour",
					),
					Type: types.IntOrInterval,
				},
				{
					Name: "xFilesFactor",
					Type: types.Float,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
		"movingSum": {
			Description: "Graphs the moving sum of a metric (or metrics) over a fixed number of\npast points, or a time interval.\n\nTakes one metric or a wildcard seriesList followed by a number N of datapoints\nor a quoted string with a length of time like '1hour' or '5min' (See ``from /\nuntil`` in the render\\_api_ for examples of time formats), and an xFilesFactor value to specify\nhow many points in the window must be non-null for the output to be considered valid. Graphs the\nsum of the preceeding datapoints for each point on the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=movingSum(Server.instance01.requests,10)\n  &target=movingSum(Server.instance*.errors,'5min')",
			Function:    "movingSum(seriesList, windowSize, xFilesFactor=None)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "movingSum",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "windowSize",
					Required: true,
					Suggestions: types.NewSuggestions(
						5,
						7,
						10,
						"1min",
						"5min",
						"10min",
						"30min",
						"1hour",
					),
					Type: types.IntOrInterval,
				},
				{
					Name: "xFilesFactor",
					Type: types.Float,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
