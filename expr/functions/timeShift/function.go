package timeShift

import (
	"context"
	"fmt"
	"strconv"

	"github.com/lomik/zapwriter"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type timeShift struct {
	interfaces.FunctionBase

	config timeShiftConfig
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

type timeShiftConfig struct {
	ResetEndDefaultValue *bool
}

func New(configFile string) []interfaces.FunctionMetadata {
	logger := zapwriter.Logger("functionInit").With(zap.String("function", "timeShift"))
	res := make([]interfaces.FunctionMetadata, 0)
	f := &timeShift{}
	functions := []string{"timeShift"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	cfg := timeShiftConfig{}
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

	if cfg.ResetEndDefaultValue == nil {
		// TODO(civil): Change default value in 0.15
		v := false
		f.config.ResetEndDefaultValue = &v
		logger.Warn("timeShift function in graphite-web have a default value for resetEnd set to true." +
			"carbonapi currently forces this to be false. This behavior will change in next major release (0.15)" +
			"to be compatible with graphite-web. Please change your dashboards to explicitly pass resetEnd parameter" +
			"or create a config file for this function that sets it to false." +
			"Please see https://github.com/go-graphite/carbonapi/blob/main/doc/configuration.md#example-for-timeshift")
	}

	return res
}

// timeShift(seriesList, timeShift, resetEnd=True)
func (f *timeShift) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// FIXME(civil): support alignDst
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	offs, err := e.GetIntervalArg(1, -1)
	if err != nil {
		return nil, err
	}
	offsStr := strconv.Itoa(int(offs))

	resetEnd, err := e.GetBoolArgDefault(2, *f.config.ResetEndDefaultValue)
	if err != nil {
		return nil, err
	}
	resetEndStr := strconv.FormatBool(resetEnd)

	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from+int64(offs), until+int64(offs), values)
	if err != nil {
		return nil, err
	}
	results := make([]*types.MetricData, len(arg))

	for n, a := range arg {
		r := a.CopyLink()
		r.Name = "timeShift(" + a.Name + ",'" + offsStr + "'," + resetEndStr + ")"
		r.StartTime = a.StartTime - int64(offs)
		r.StopTime = a.StopTime - int64(offs)
		if resetEnd && r.StopTime > until {
			r.StopTime = until
		}
		length := int((r.StopTime - r.StartTime) / r.StepTime)
		if length < 0 {
			continue
		}
		r.Values = r.Values[:length]

		r.Tags["timeshift"] = fmt.Sprintf("%d", offs)
		results[n] = r

	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *timeShift) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"timeShift": {
			Description: "Takes one metric or a wildcard seriesList, followed by a quoted string with the\nlength of time (See ``from / until`` in the render\\_api_ for examples of time formats).\n\nDraws the selected metrics shifted in time. If no sign is given, a minus sign ( - ) is\nimplied which will shift the metric back in time. If a plus sign ( + ) is given, the\nmetric will be shifted forward in time.\n\nWill reset the end date range automatically to the end of the base stat unless\nresetEnd is False. Example case is when you timeshift to last week and have the graph\ndate range set to include a time in the future, will limit this timeshift to pretend\nending at the current time. If resetEnd is False, will instead draw full range including\nfuture time.\n\nBecause time is shifted by a fixed number of seconds, comparing a time period with DST to\na time period without DST, and vice-versa, will result in an apparent misalignment. For\nexample, 8am might be overlaid with 7am. To compensate for this, use the alignDST option.\n\nUseful for comparing a metric against itself at a past periods or correcting data\nstored at an offset.\n\nExample:\n\n.. code-block:: none\n\n  &target=timeShift(Sales.widgets.largeBlue,\"7d\")\n  &target=timeShift(Sales.widgets.largeBlue,\"-7d\")\n  &target=timeShift(Sales.widgets.largeBlue,\"+1h\")",
			Function:    "timeShift(seriesList, timeShift, resetEnd=True, alignDST=False)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "timeShift",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "timeShift",
					Required: true,
					Suggestions: types.NewSuggestions(
						"1h",
						"6h",
						"12h",
						"1d",
						"2d",
						"7d",
						"14d",
						"30d",
					),
					Type: types.Interval,
				},
				{
					Default: types.NewSuggestion(*f.config.ResetEndDefaultValue),
					Name:    "resetEnd",
					Type:    types.Boolean,
				},
				/*
					{
						Default: types.NewSuggestion(false),
						Name:    "alignDst",
						Type:    types.Boolean,
					},
				*/
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
