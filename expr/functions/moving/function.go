package moving

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/lomik/zapwriter"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/interfaces"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
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
	functions := []string{"movingAverage", "movingMin", "movingMax", "movingSum"}
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

	var scaleByStep bool

	var argstr string

	var xFilesFactor float64

	if len(e.Args()) < 2 {
		return nil, parser.ErrMissingArgument
	}

	switch e.Args()[1].Type() {
	case parser.EtConst:
		// In this case, zipper does not request additional retrospective points,
		// and leading `n` values, that used to calculate window, become NaN
		n, err = e.GetIntArg(1)
		argstr = strconv.Itoa(n)
	case parser.EtString:
		var n32 int32
		n32, err = e.GetIntervalArg(1, 1)
		argstr = fmt.Sprintf("%q", e.Args()[1].StringValue())
		n = int(n32)
		scaleByStep = true
	default:
		err = parser.ErrBadType
	}
	if err != nil {
		return nil, err
	}

	windowSize := n

	start := from
	if scaleByStep {
		start -= int64(n)
	}

	arg, err := helper.GetSeriesArg(ctx, e.Args()[0], start, until, values)
	if err != nil {
		return nil, err
	}

	if len(e.Args()) == 3 {
		xFilesFactor, err = e.GetFloatArgDefault(2, float64(arg[0].XFilesFactor))

		if err != nil {
			return nil, err
		}
	}

	var result []*types.MetricData

	if len(arg) == 0 {
		return result, nil
	}

	var offset int

	if scaleByStep {
		windowSize /= int(arg[0].StepTime)
		offset = windowSize
	}

	for _, a := range arg {
		r := a.CopyLink()
		r.Name = fmt.Sprintf("%s(%s,%s)", e.Target(), a.Name, argstr)

		if windowSize == 0 {
			if *f.config.ReturnNaNsIfStepMismatch {
				r.Values = make([]float64, len(a.Values))
				for i := range a.Values {
					r.Values[i] = math.NaN()
				}
			}
			result = append(result, r)
			continue
		}
		r.Values = make([]float64, len(a.Values)-offset)
		r.StartTime = (from + r.StepTime - 1) / r.StepTime * r.StepTime // align StartTime to closest >= StepTime
		r.StopTime = r.StartTime + int64(len(r.Values))*r.StepTime

		w := &types.Windowed{Data: make([]float64, windowSize)}
		for i, v := range a.Values {
			if ridx := i - offset; ridx >= 0 {
				if helper.XFilesFactorValues(w.Data, xFilesFactor) {
					switch e.Target() {
					case "movingAverage":
						r.Values[ridx] = w.Mean()
					case "movingSum":
						r.Values[ridx] = w.Sum()
						//TODO(cldellow): consider a linear time min/max-heap for these,
						// e.g. http://stackoverflow.com/questions/8905525/computing-a-moving-maximum/8905575#8905575
					case "movingMin":
						r.Values[ridx] = w.Min()
					case "movingMax":
						r.Values[ridx] = w.Max()
					}
					if i < windowSize || math.IsNaN(r.Values[ridx]) {
						r.Values[ridx] = math.NaN()
					}
				} else {
					r.Values[ridx] = math.NaN()
				}
			}
			w.Push(v)
		}
		r.Tags[e.Target()] = fmt.Sprintf("%d", windowSize)
		result = append(result, r)
	}
	return result, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *moving) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
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
		},
	}
}
