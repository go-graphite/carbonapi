package movingMedian

import (
	"context"
	"math"
	"strconv"

	"github.com/JaderDias/movingmedian"
	"github.com/lomik/zapwriter"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type movingMedian struct {
	interfaces.FunctionBase

	config movingMedianConfig
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

type movingMedianConfig struct {
	ReturnNaNsIfStepMismatch *bool
}

func New(configFile string) []interfaces.FunctionMetadata {
	logger := zapwriter.Logger("functionInit").With(zap.String("function", "movingMedian"))
	res := make([]interfaces.FunctionMetadata, 0)
	f := &movingMedian{}
	functions := []string{"movingMedian"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	cfg := movingMedianConfig{}
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

// movingMedian(seriesList, windowSize)
func (f *movingMedian) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	var n int
	var err error

	var scaleByStep bool

	var argstr string

	switch e.Arg(1).Type() {
	case parser.EtConst:
		n, err = e.GetIntArg(1)
		argstr = strconv.Itoa(n)
	case parser.EtString:
		var n32 int32
		n32, err = e.GetIntervalArg(1, 1)
		n = int(n32)
		argstr = "'" + e.Arg(1).StringValue() + "'"
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

	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), start, until, values)
	if err != nil {
		return nil, err
	}

	if len(arg) == 0 {
		return nil, nil
	}

	var offset int

	if scaleByStep {
		windowSize /= int(arg[0].StepTime)
		offset = windowSize
	}

	result := make([]*types.MetricData, len(arg))
	for n, a := range arg {
		r := a.CopyName("movingMedian(" + a.Name + "," + argstr + ")")

		if windowSize == 0 {
			if *f.config.ReturnNaNsIfStepMismatch {
				r.Values = make([]float64, len(a.Values))
				for i := range a.Values {
					r.Values[i] = math.NaN()
				}
			}
		} else {
			r.Values = make([]float64, len(a.Values)-offset)
			r.StartTime = (from + r.StepTime - 1) / r.StepTime * r.StepTime // align StartTime to closest >= StepTime
			r.StopTime = r.StartTime + int64(len(r.Values))*r.StepTime

			data := movingmedian.NewMovingMedian(windowSize)

			for i, v := range a.Values {
				data.Push(v)

				if ridx := i - offset; ridx >= 0 {
					r.Values[ridx] = math.NaN()
					if i >= (windowSize - 1) {
						r.Values[ridx] = data.Median()
					}
				}
			}
		}
		r.Tags["movingMedian"] = argstr
		result[n] = r
	}
	return result, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *movingMedian) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"movingMedian": {
			Description: "Graphs the moving median of a metric (or metrics) over a fixed number of\npast points, or a time interval.\n\nTakes one metric or a wildcard seriesList followed by a number N of datapoints\nor a quoted string with a length of time like '1hour' or '5min' (See ``from /\nuntil`` in the render\\_api_ for examples of time formats), and an xFilesFactor value to specify\nhow many points in the window must be non-null for the output to be considered valid. Graphs the\nmedian of the preceeding datapoints for each point on the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=movingMedian(Server.instance01.threads.busy,10)\n  &target=movingMedian(Server.instance*.threads.idle,'5min')",
			Function:    "movingMedian(seriesList, windowSize, xFilesFactor=None)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "movingMedian",
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
